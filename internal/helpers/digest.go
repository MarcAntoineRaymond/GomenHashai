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
	"regexp"
	"strings"
)

const DEFAULT_DIGEST_MAPPING_PATH = "/etc/gomenhashai/digests_mapping.yaml"

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
	if digest, ok := DIGEST_MAPPING[image]; ok {
		return digest
	} else {
		if CONFIG.ImageDefaultDigest && strings.Contains(image, ":") {
			imageWithoutTag := strings.Split(image, ":")[0]
			if digest, ok := DIGEST_MAPPING[imageWithoutTag]; ok {
				return digest
			}
		}
	}
	return ""
}
