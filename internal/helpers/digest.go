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
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

// Return trusted digests from file or registry depending on conf
func GetTrustedDigest(image string) (string, error) {
	if CONFIG.FetchDigests {
		return GetDigestFromRegistry(image)
	} else {
		return GetTrustedDigestFromMapping(image), nil
	}
}

// Return digest from mapping for this image or empty string
func GetTrustedDigestFromMapping(image string) string {
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
		imageWithoutRegistry := GetImageWithoutRegistry(image)
		if imageWithoutRegistry != image {
			return GetTrustedDigestFromMapping(imageWithoutRegistry)
		}
	}
	return ""
}

// Return digest from registry for this image or empty string
func GetDigestFromRegistry(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("failed to parse image reference: %v", err)
	}

	// Use DefaultKeychain (works with K8s service account/imagePullSecrets)
	desc, err := remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", fmt.Errorf("failed to get image from registry: %v", err)
	}

	return desc.Digest.String(), nil
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
