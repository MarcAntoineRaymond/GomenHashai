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
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

// Config struct
type Config struct {
	// Path to the digests mapping file
	DigestsMappingFile string `yaml:"digestMappingFile"`
	// List of images to skip, can contain regex
	Exemptions []string `yaml:"exemptions"`
	// An image without tag in the mapping will be considered defaul. Images with tag that do not match specific trusted digest will use this digest instead (image it is the same base image)
	ImageDefaultDigest bool `yaml:"imageDefaultDigest"`
	// Can be warn or fail (default)
	ValidationMode string `yaml:"validationMode" validate:"oneof warn fail"`
	// Enable to not modify pods but instead logs (pods will fail validation unless you disable it or set it in warn)
	MutationDryRun bool `yaml:"mutationDryRun"`
	// Enable modifying the registry part of images with the value of MutationRegistry
	MutationRegistryEnabled bool `yaml:"mutationRegistryEnabled"`
	// The registry to inject when MutationRegistryEnabled is true
	MutationRegistry string `yaml:"mutationRegistry"`
	// Configuration of the process that handles existing pods on init
	ExistingPods ExistingPodsConfig `yaml:"existingPods"`
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
	// Replace already existing pods with output from webhook, if disbaled webhook will be used with dry run to not modify pods
	UpdateEnabled bool `yaml:"updateEnabled" envconfig:"EXISTING_PODS_UPDATE_ENABLED"`
	// Allow deleting existing pods that are forbidden by webhook
	DeleteEnabled bool `yaml:"deleteEnabled" envconfig:"EXISTING_PODS_DELETE_ENABLED"`
}

var CONFIG_PATH = "/etc/gomenhashai/configs/config.yaml"
var DIGEST_MAPPING = map[string]string{}
var CONFIG = defaultConfig()

func defaultConfig() Config {
	return Config{
		DigestsMappingFile:      "/etc/gomenhashai/digests/digests_mapping.yaml",
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
	}
}

func InitConfig() error {
	cfg := defaultConfig()

	filepath := os.Getenv("GOMENHASHAI_CONFIG_PATH")
	if filepath == "" {
		filepath = CONFIG_PATH
	}

	// Read YAML from file
	if data, err := os.ReadFile(filepath); err == nil {
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

	CONFIG = cfg
	return nil
}

// Load Digest Mapping from file
func LoadDigestMapping() error {

	filepath := CONFIG.DigestsMappingFile

	data, err := os.ReadFile(filepath)
	if err == nil {
		if err := yaml.Unmarshal(data, &DIGEST_MAPPING); err != nil {
			return err
		}

	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
