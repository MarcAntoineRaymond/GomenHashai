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
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const DEFAULT_DIGEST_MAPPING_PATH = "/etc/kintegrity/digests_mapping.yaml"

var DIGEST_MAPPING = map[string]string{}

// Digest mapping without tag will be used for images using tag not present in digest mappings
var CAN_FORCE_DIGEST = true

func InitConfig() error {
	forceMapping := os.Getenv("DIGEST_MAPPING_FORCE")
	if forceMapping != "" {
		boolValue, err := strconv.ParseBool(forceMapping)
		if err != nil {
			return err
		}
		CAN_FORCE_DIGEST = boolValue
	}
	return nil
}

// Load Digest Mapping from file, filepath can be set from env DIGEST_MAPPING_PATH
func LoadDigestMapping() error {
	filepath := os.Getenv("DIGEST_MAPPING_PATH")
	if filepath == "" {
		filepath = DEFAULT_DIGEST_MAPPING_PATH
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &DIGEST_MAPPING); err != nil {
		return err
	}
	return nil
}

// getDigest from container image or return empty
func GetDigest(image string) string {
	re := regexp.MustCompile(`@sha256:[a-fA-F0-9]{64}`)
	match := re.FindString(image)
	if match != "" {
		return match[1:]
	}
	return ""
}

// Return digest from mapping for this image or empty string
func GetTrustedDigest(image string) string {
	// TODO handle tag or not tag mechanic, if image has tag and digest exist for base image without tag, use this one
	if digest, ok := DIGEST_MAPPING[image]; ok {
		return digest
	} else {
		if CAN_FORCE_DIGEST && strings.Contains(image, ":") {
			imageWithoutTag := strings.Split(image, ":")[0]
			if digest, ok := DIGEST_MAPPING[imageWithoutTag]; ok {
				return digest
			}
		}
	}
	return ""
}
