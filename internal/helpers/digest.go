package helpers

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const defaultDigestMappingPath = "/etc/kintegrity/digests_mapping.yaml"

var DIGEST_MAPPING = map[string]string{}

// Load Digest Mapping from file, filepath can be set from env DIGEST_MAPPING_PATH
func LoadDigestMapping() map[string]string {
	filepath := os.Getenv("DIGEST_MAPPING_PATH")
	if filepath == "" {
		filepath = defaultDigestMappingPath
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	digestMapping := make(map[string]string)
	if err := yaml.Unmarshal(data, &digestMapping); err != nil {
		panic(err)
	}
	return digestMapping
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
	}
	return ""
}
