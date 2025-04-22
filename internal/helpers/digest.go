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

// getDigest from container image or return empty, invalid digests are ignored
func GetDigest(image string) string {
	re := regexp.MustCompile(`@sha256:[a-fA-F0-9]{64}$`)
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
		// Check for base image without tag in mapping this will be default
		if CONFIG.ImageDefaultDigest && strings.Contains(image, ":") {
			imageWithoutTag := strings.Split(image, ":")[0]
			if digest, ok := DIGEST_MAPPING[imageWithoutTag]; ok {
				return digest
			}
		}
		// Try to find digest without registry part if it exist
		if imageWithoutRegistry := GetImageWithoutRegistry(image); imageWithoutRegistry != image {
			return GetTrustedDigest(imageWithoutRegistry)
		}
	}
	return ""
}

// Return image without registry part if present at the beginning of image
func GetImageWithoutRegistry(image string) string {
	if strings.Contains(strings.Split(image, "/")[0], ".") {
		imageWithoutRegistry := strings.Join(strings.Split(image, "/")[1:], "/")
		return imageWithoutRegistry
	}
	return image
}

// Return if the image match an entry in the exempt list which can contain regex
func IsImageExempt(image string) bool {
	if len(CONFIG.Exemptions) > 0 {
		for _, exemption := range CONFIG.Exemptions {
			if image == exemption {
				return true
			}
			re, err := regexp.Compile(exemption)
			if err == nil {
				match := re.FindString(image)
				if match != "" {
					return true
				}
			}
		}
	}
	return false
}
