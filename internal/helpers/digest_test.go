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
	var registry string
	var digest string
	var imageDigest string
	var invalidDigest string
	var imageInvalidDigest string
	var imageNoDigest string

	BeforeEach(func() {
		registry = "docker.io"
		baseImage = "library/busybox"
		imageNoDigest = registry + "/" + baseImage
		tagLatest = "latest"
		tag = "8.13.0"
		imageWithTrustedTag = "curlimages/curl" + ":" + tag
		imageWithTag = imageNoDigest + tagLatest
		digest = "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549"
		imageDigest = imageNoDigest + "@" + digest
		invalidDigest = "sh56:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e"
		imageInvalidDigest = imageNoDigest + "@" + invalidDigest
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
		Context("with untrusted tag", func() {
			It("should be empty", func() {
				Expect(helpers.GetTrustedDigest(imageWithTag)).To(Equal(""))
			})
		})
		Context("with trusted tag", func() {
			It("should NOT be empty", func() {
				Expect(helpers.GetTrustedDigest(imageWithTrustedTag)).To(Not(Equal("")))
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
