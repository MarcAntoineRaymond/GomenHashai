/*
Copyright 2025 Marc-Antoine RAYMOND.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

// Config struct
type Config struct {
	// Path to the digests mapping file
	DigestsMappingFile string `yaml:"digestsMappingFile"`
	// Config for fetching digests from registry
	FetchDigests bool `yaml:"fetchDigests"`
	// Auth config to pull digests from remote registry
	RegistriesConfigFile string `yaml:"registriesConfigFile"`
	// List of images to skip, can contain regex ex: ".*redis:.*"
	Exemptions []string `yaml:"exemptions"`
	// An image without tag in the mapping will be considered default. Images with tag that do not match specific trusted digest will use this digest instead (image it is the same base image)
	ImageDefaultDigest bool `yaml:"imageDefaultDigest"`
	// Can be warn or fail (default)
	ValidationMode string `yaml:"validationMode" validate:"oneof=warn fail"`
	// Enable to not modify pods but instead logs (pods will fail validation unless you disable it or set it in warn)
	MutationDryRun bool `yaml:"mutationDryRun"`
	// Enable modifying the registry part of images with the value of MutationRegistry
	MutationRegistryEnabled bool `yaml:"mutationRegistryEnabled"`
	// The registry to inject when MutationRegistryEnabled is true
	MutationRegistry string `yaml:"mutationRegistry"`
	// Enforce image pull policy for all containers
	MutationPullPolicy string `yaml:"mutationPullPolicy" validate:"omitempty,oneof=Always IfNotPresent Never"`
	// Additional image pull secrets to add to all pods
	MutationImagePullSecrets []corev1.LocalObjectReference `yaml:"mutationImagePullSecrets"`
	// Configuration of the process that handles existing pods on init
	ExistingPods ExistingPodsConfig `yaml:"existingPods"`
	// File containing pull secret credentials to create in all namespaces
	PullSecretsCredentialsFile string `yaml:"pullSecretsCredentialsFile"`
}

type PullSecretCredential struct {
	Name      string `yaml:"name"`
	Username  string `yaml:"username"`
	Token     string `yaml:"token"`
	Registry  string `yaml:"registry"`
	DockerCfg []byte `yaml:"-"`
}

type ExistingPodsConfig struct {
	// Enable the init function that will process existing pods at startup
	Enabled bool `yaml:"enabled" envconfig:"EXISTING_PODS_ENABLED"`
	// Timeout used to wait before starting this job in seconds
	StartTimeout int `yaml:"startTimeout" validate:"gte=0" envconfig:"EXISTING_PODS_START_TIMEOUT"`
	// Timeout used to wait before retrying to process pods that failed in seconds
	RetryTimeout int `yaml:"retryTimeout" validate:"gte=0" envconfig:"EXISTING_PODS_RETRY_TIMEOUT"`
	// How many times we should retry processing pods that failed
	Retries int `yaml:"retries" validate:"gte=0" envconfig:"EXISTING_PODS_RETRIES"`
	// Replace already existing pods with output from webhook, if disabled webhook will be used with dry run to not modify pods
	UpdateEnabled bool `yaml:"updateEnabled" envconfig:"EXISTING_PODS_UPDATE_ENABLED"`
	// Allow deleting existing pods that are forbidden by webhook
	DeleteEnabled bool `yaml:"deleteEnabled" envconfig:"EXISTING_PODS_DELETE_ENABLED"`
}

const ValidationModeWarn = "warn"
const ValidationModeFail = "fail"

var CONFIG_PATH = "/etc/gomenhashai/configs/config.yaml"
var DIGEST_MAPPING = map[string]string{}
var CONFIG = defaultConfig()
var REGISTRIES_CONFIG = map[string]RegistryCredentials{}

// Create pull secrets into all namespaces
var PULL_SECRETS_CREDENTIALS = []PullSecretCredential{}

type RegistryCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func defaultConfig() Config {
	return Config{
		DigestsMappingFile:      "/etc/gomenhashai/digests/digests_mapping.yaml",
		FetchDigests:            false,
		RegistriesConfigFile:    "/etc/gomenhashai/configs/registries.yaml",
		Exemptions:              []string{},
		ImageDefaultDigest:      true,
		ValidationMode:          "fail",
		MutationDryRun:          false,
		MutationRegistryEnabled: false,
		ExistingPods: ExistingPodsConfig{
			Enabled:       true,
			StartTimeout:  5,
			RetryTimeout:  5,
			Retries:       5,
			UpdateEnabled: true,
			DeleteEnabled: true,
		},
		PullSecretsCredentialsFile: "/etc/gomenhashai/configs/pullSecretsCredentials.yaml",
	}
}

func InitConfig() error {
	cfg := defaultConfig()

	configPath := os.Getenv("GOMENHASHAI_CONFIG_PATH")
	if configPath == "" {
		configPath = CONFIG_PATH
	}

	// Read YAML from file
	if data, err := os.ReadFile(filepath.Clean(configPath)); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with environment variables
	if err := envconfig.Process("gomenhashai", &cfg); err != nil {
		return fmt.Errorf("failed to process env vars: %w", err)
	}

	// Sanatize
	cfg.MutationRegistry = strings.TrimSuffix(cfg.MutationRegistry, "/")

	// Validate config
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Load registry credentials
	if cfg.FetchDigests {
		if data, err := os.ReadFile(filepath.Clean(cfg.RegistriesConfigFile)); err == nil {
			if err := yaml.Unmarshal(data, &REGISTRIES_CONFIG); err != nil {
				return fmt.Errorf("failed to parse registries config file: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read registries config file: %w", err)
		}
	}

	// Load pull secrets credentials
	if cfg.PullSecretsCredentialsFile != "" {
		if data, err := os.ReadFile(filepath.Clean(cfg.PullSecretsCredentialsFile)); err == nil {
			if err := yaml.Unmarshal(data, &PULL_SECRETS_CREDENTIALS); err != nil {
				return fmt.Errorf("failed to parse pull secrets credentials file: %w", err)
			}
			for i, cred := range PULL_SECRETS_CREDENTIALS {
				dockerCfgJSON, err := MakeDockerConfigJson(cred.Username, cred.Token, cred.Registry)
				if err != nil {
					return fmt.Errorf("failed to build docker config json for pull secret %s: %w", cred.Name, err)
				}
				PULL_SECRETS_CREDENTIALS[i].DockerCfg = dockerCfgJSON
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read pull secrets credentials file: %w", err)
		}
	}

	CONFIG = cfg
	return nil
}

// Load Digest Mapping from file
func LoadDigestMapping() error {

	mappingPath := CONFIG.DigestsMappingFile

	data, err := os.ReadFile(filepath.Clean(mappingPath))
	if err == nil {
		if err := yaml.Unmarshal(data, &DIGEST_MAPPING); err != nil {
			return err
		}

	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func MakeDockerConfigJson(username, token, registry string) ([]byte, error) {
	// Build .dockerconfigjson content
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, token)))

	dockerCfg := map[string]any{
		"auths": map[string]any{
			registry: map[string]string{
				"username": username,
				"password": token,
				"auth":     auth,
			},
		},
	}

	dockerCfgJSON, _ := json.Marshal(dockerCfg)
	return dockerCfgJSON, nil
}
