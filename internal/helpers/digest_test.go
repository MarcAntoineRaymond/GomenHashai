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

package helpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/MarcAntoineRaymond/gomenhashai/internal/helpers"
)

var _ = Describe("Digest", func() {
	var tag string
	var tagLatest string
	var imageWithTag string
	var imageWithTrustedTag string
	var baseImage string
	var digest string
	var imageDigest string
	var invalidDigest string
	var imageInvalidDigest string
	var imageNoDigest string

	helpers.DIGEST_MAPPING = map[string]string{
		"busybox:latest":                   "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f",
		"busybox":                          "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
		"library/busybox":                  "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
		"docker.io/library/busybox":        "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
		"docker.io/library/busybox:stable": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
		"busybox:stable":                   "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
		"nginx/nginx-ingress:5.0.0-alpine": "sha256:a6c4d7c7270f03a3abb1ff38973f5db98d8660832364561990c4d0ef8b1477af",
		"curlimages/curl:8.13.0":           "sha256:d43bdb28bae0be0998f3be83199bfb2b81e0a30b034b6d7586ce7e05de34c3fd",
	}

	BeforeEach(func() {
		baseImage = "library/busybox"
		imageNoDigest = "docker.io/library/busybox"
		tagLatest = "latest"
		tag = "8.13.0"
		imageWithTrustedTag = "curlimages/curl" + ":" + tag
		imageWithTag = "docker.io/library/busybox" + ":" + tagLatest
		digest = "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
		imageDigest = "docker.io/library/busybox" + "@" + digest
		invalidDigest = "sh56:aae246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e"
		imageInvalidDigest = "docker.io/library/busybox" + "@" + invalidDigest
	})

	// Test GetDigest()
	Describe("Extract Digest from image", func() {
		Context("with digest", func() {
			It("should be digest value", func() {
				Expect(helpers.GetDigest(imageDigest)).To(Equal(digest))
			})
		})
		Context("without digest", func() {
			It("should be empty", func() {
				Expect(helpers.GetDigest(imageNoDigest)).To(Equal(""))
			})
		})
		Context("with tag", func() {
			It("should be empty", func() {
				Expect(helpers.GetDigest(imageWithTag)).To(Equal(""))
			})
		})
		Context("with invalid digest", func() {
			It("should be empty", func() {
				Expect(helpers.GetDigest(imageInvalidDigest)).To(Equal(""))
			})
		})
	})

	// Test GetTrustedDigest()
	Describe("Get trusted digest from mapping", func() {
		Context("with digest", func() {
			It("should be empty", func() {
				Expect(helpers.GetTrustedDigest(imageDigest)).To(Equal(""))
			})
		})
		Context("with untrusted tag but default base image", func() {
			It("should be digest same as default image", func() {
				Expect(helpers.GetTrustedDigest(imageWithTag)).To(Equal(helpers.GetTrustedDigest(imageNoDigest)))
			})
		})
		Context("with trusted tag", func() {
			It("should be digest from mapping", func() {
				Expect(helpers.GetTrustedDigest(imageWithTrustedTag)).To(Equal(helpers.DIGEST_MAPPING[imageWithTrustedTag]))
			})
		})
		Context("without tag", func() {
			It("should be digest from mapping", func() {
				Expect(helpers.GetTrustedDigest(baseImage)).To(Equal(helpers.DIGEST_MAPPING[baseImage]))
			})
		})
	})

	// Test GetImageWithoutRegistry()
	Describe("Get the image without the registry part of image name", func() {
		Context("with registry", func() {
			It("should be base image", func() {
				Expect(helpers.GetImageWithoutRegistry(imageNoDigest)).To(Equal(baseImage))
			})
		})
		Context("without registry", func() {
			It("should be same image", func() {
				Expect(helpers.GetImageWithoutRegistry(imageWithTrustedTag)).To(Equal(imageWithTrustedTag))
			})
		})
		Context("without registry with digest", func() {
			It("should be same image", func() {
				Expect(helpers.GetImageWithoutRegistry(imageWithTrustedTag + "@" + digest)).To(Equal(imageWithTrustedTag + "@" + digest))
			})
		})
	})

	// Test IsImageExempt()
	Describe("Is the image exempted by the exemption config", func() {
		Context("with registry", func() {
			It("should be exempted", func() {
				Expect(helpers.IsImageExempt(imageDigest)).To(BeTrue())
			})
		})
		Context("with redis", func() {
			It("should be exempted", func() {
				Expect(helpers.IsImageExempt("lib/redis:6")).To(BeTrue())
			})
		})
		Context("with curl", func() {
			It("should  NOT be exempted", func() {
				Expect(helpers.IsImageExempt(imageWithTrustedTag)).To(BeFalse())
			})
		})
	})
})
