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

package v1

import (
	"github.com/MarcAntoineRaymond/gomenhashai/internal/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Pod Webhook", func() {
	var (
		containersTrusted    []corev1.Container
		containersNotTrusted []corev1.Container
		containersExempted   []corev1.Container
		validator            PodCustomValidator
		defaulter            PodCustomDefaulter
		mutatedContainers    []corev1.Container
		containers           []corev1.Container
	)

	BeforeEach(func() {
		containersTrusted = []corev1.Container{
			{
				Name:  "app",
				Image: "docker.io/library/busybox:stable",
			},
			{
				Name:  "sidecar",
				Image: "busybox",
			},
		}
		containersNotTrusted = []corev1.Container{
			{
				Name:  "app",
				Image: "busybox",
			},
			{
				Name:  "sidecar",
				Image: "curlimages/curl:7",
			},
		}
		containersExempted = []corev1.Container{
			{
				Name:  "app",
				Image: "test/redis:test",
			},
		}
		validator = PodCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		defaulter = PodCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
	})

	Describe("Function mutation", func() {
		BeforeEach(func() {
			containers = make([]corev1.Container, len(containersTrusted))
			copy(containers, containersTrusted)
			mutatedContainers = AddContainerImageDigest(containers, "test")
		})
		It("should not modify orignal", func() {
			Expect(mutatedContainers).To(Not(Equal(containersTrusted)))
			Expect(containers).To(Equal(containersTrusted))
		})

		It("should return a new array", func() {
			Expect(mutatedContainers).To(Not(Equal(containers)))
		})
	})

	Describe("Common use case pod without digests", func() {
		Context("Container using trusted images", func() {
			BeforeEach(func() {
				mutatedContainers = AddContainerImageDigest(containersTrusted, "test")
			})

			It("Should have trusted digests", func() {
				Expect(mutatedContainers).To(HaveLen(len(containersTrusted)))
				for i, container := range containersTrusted {
					Expect(helpers.GetDigest(mutatedContainers[i].Image)).To(Not(BeEmpty()))
					Expect(helpers.GetDigest(mutatedContainers[i].Image)).To(Equal(helpers.GetTrustedDigest(container.Image)))
				}
			})
			It("Should be Allowed", func() {
				pod := corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: mutatedContainers,
					},
				}
				warn, err := ValidatePod(&pod)
				Expect(warn).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})
		Context("Container using NOT trusted images", func() {
			BeforeEach(func() {
				mutatedContainers = AddContainerImageDigest(containersNotTrusted, "test")
			})
			It("Should have trusted digest on trusted image and nothing on not trusted", func() {
				Expect(mutatedContainers).To(HaveLen(len(containersNotTrusted)))
				Expect(helpers.GetDigest(mutatedContainers[0].Image)).To(Equal(helpers.GetTrustedDigest(containersNotTrusted[0].Image)))
				Expect(helpers.GetDigest(mutatedContainers[1].Image)).To(BeEmpty())
			})
			It("Should be denied", func() {
				pod := corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: mutatedContainers,
					},
				}
				warn, err := ValidatePod(&pod)
				Expect(warn).To(BeEmpty())
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsForbidden(err)).To(BeTrue())
			})
		})
		Context("Container using exempted images", func() {
			BeforeEach(func() {
				mutatedContainers = AddContainerImageDigest(containersExempted, "test")
			})
			It("Should not be modified", func() {
				Expect(mutatedContainers).To(Equal(containersExempted))
			})
			It("Should be Allowed", func() {
				pod := corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: mutatedContainers,
					},
				}
				warn, err := ValidatePod(&pod)
				Expect(warn).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})
		Context("Containers empty list", func() {
			BeforeEach(func() {
				mutatedContainers = AddContainerImageDigest([]corev1.Container{}, "test")
			})
			It("Should not be modified", func() {
				Expect(mutatedContainers).To(Equal([]corev1.Container{}))
			})
			It("Should be Allowed", func() {
				pod := corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: mutatedContainers,
					},
				}
				warn, err := ValidatePod(&pod)
				Expect(warn).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})
	})
})
