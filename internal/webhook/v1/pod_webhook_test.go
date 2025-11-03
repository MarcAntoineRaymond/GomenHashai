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
	"context"

	"github.com/GomenHashai/gomenhashai/internal/helpers"
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
		pod                  corev1.Pod
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
			Expect(mutatedContainers).ToNot(Equal(containersTrusted))
			Expect(containers).To(Equal(containersTrusted))
		})

		It("should return a new array", func() {
			Expect(mutatedContainers).ToNot(Equal(containers))
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
					Expect(helpers.GetDigest(mutatedContainers[i].Image)).ToNot(BeEmpty())
					digest, err := helpers.GetTrustedDigest(container.Image)
					Expect(err).ToNot(HaveOccurred())
					Expect(helpers.GetDigest(mutatedContainers[i].Image)).To(Equal(digest))
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
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("Container using NOT trusted images", func() {
			BeforeEach(func() {
				mutatedContainers = AddContainerImageDigest(containersNotTrusted, "test")
			})
			It("Should have trusted digest on trusted image and nothing on not trusted", func() {
				Expect(mutatedContainers).To(HaveLen(len(containersNotTrusted)))
				digest, err := helpers.GetTrustedDigest(containersNotTrusted[0].Image)
				Expect(err).ToNot(HaveOccurred())
				Expect(helpers.GetDigest(mutatedContainers[0].Image)).To(Equal(digest))
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
				Expect(err).ToNot(HaveOccurred())
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
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Registry mutation feature", func() {
		BeforeEach(func() {
			helpers.CONFIG.MutationRegistryEnabled = true
		})
		Context("Registry is myregistry", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationRegistry = "myregistry.test"
			})

			It("Should add registry prefix to all images", func() {
				mutatedContainers = AddContainerImageDigest(containersTrusted, "test")
				Expect(mutatedContainers).To(HaveEach(HaveField("Image", HavePrefix("myregistry.test"))))
			})

			It("Should add registry prefix to all images", func() {
				mutatedContainers = AddContainerImageDigest(containersNotTrusted, "test")
				Expect(mutatedContainers).To(HaveEach(HaveField("Image", HavePrefix("myregistry.test"))))
			})

			It("Should do nothing cause all images already have registry", func() {
				containersRegistry := []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "myregistry.test/image",
					},
					corev1.Container{
						Name:  "sidecar",
						Image: "myregistry.test/anotherimage",
					},
				}
				mutatedContainers = AddContainerImageDigest(containersRegistry, "test")
				Expect(mutatedContainers).To(Equal(containersRegistry))
			})
		})
		Context("Registry is empty", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationRegistry = ""
			})

			It("Should remove registry prefix in image", func() {
				mutatedContainers = AddContainerImageDigest(containersTrusted, "test")
				Expect(mutatedContainers).To(HaveEach(HaveField("Image", Not(HavePrefix("docker.io")))))
			})

			It("Should do nothing as images have no registry", func() {
				containersRegistry := []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "repo/image",
					},
					corev1.Container{
						Name:  "sidecar",
						Image: "repo/anotherimage",
					},
				}
				mutatedContainers = AddContainerImageDigest(containersRegistry, "test")
				Expect(mutatedContainers).To(Equal(containersRegistry))
			})
		})
		AfterEach(func() {
			helpers.CONFIG.MutationRegistryEnabled = false
		})
	})

	Describe("PullPolicy mutation feature", func() {
		Context("PullPolicy is Never", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationPullPolicy = "Never"
			})

			It("Should add/patch pullPolicy 'Never' to all containers", func() {
				containersRegistry := []corev1.Container{
					corev1.Container{
						Name:  "app",
						Image: "repo/image",
					},
					corev1.Container{
						Name:            "sidecar",
						Image:           "repo/anotherimage",
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				}
				mutatedContainers = AddContainerImageDigest(containersRegistry, "test")
				Expect(mutatedContainers).To((HaveEach(HaveField("ImagePullPolicy", Equal(corev1.PullNever)))))
			})

			It("Should do nothing cause all containers already have Never PullPolicy", func() {
				containersRegistry := []corev1.Container{
					corev1.Container{
						Name:            "app",
						Image:           "repo/image",
						ImagePullPolicy: corev1.PullNever,
					},
					corev1.Container{
						Name:            "sidecar",
						Image:           "repo/anotherimage",
						ImagePullPolicy: corev1.PullNever,
					},
				}
				mutatedContainers = AddContainerImageDigest(containersRegistry, "test")
				Expect(mutatedContainers).To((HaveEach(HaveField("ImagePullPolicy", Equal(corev1.PullNever)))))
			})
		})
		Context("PullPolicy is empty", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationPullPolicy = ""
			})

			It("Should do nothing", func() {
				containersRegistry := []corev1.Container{
					corev1.Container{
						Name:            "app",
						Image:           "repo/image",
						ImagePullPolicy: corev1.PullAlways,
					},
					corev1.Container{
						Name:            "sidecar",
						Image:           "repo/anotherimage",
						ImagePullPolicy: corev1.PullAlways,
					},
				}
				mutatedContainers = AddContainerImageDigest(containersRegistry, "test")
				Expect(mutatedContainers).To((HaveEach(HaveField("ImagePullPolicy", Equal(corev1.PullAlways)))))
			})

		})
		AfterEach(func() {
			helpers.CONFIG.MutationPullPolicy = ""
		})
	})

	Describe("ImagePullSecrets mutation feature", func() {
		Context("ImagePullSecrets contains secrets", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationImagePullSecrets = []corev1.LocalObjectReference{
					{Name: "my-secret1"},
					{Name: "my-secret2"},
				}
				pod = corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: containersTrusted,
					},
				}
			})

			It("Should add new ImagePullSecrets to pod", func() {
				err := (&defaulter).Default(context.TODO(), &pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pod).To(
					HaveField("Spec",
						HaveField("ImagePullSecrets",
							Equal([]corev1.LocalObjectReference{
								{Name: "my-secret1"},
								{Name: "my-secret2"},
							}),
						),
					),
				)
			})

			It("Should add ImagePullSecrets to pod over existing ones", func() {
				pod = corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: containersTrusted,
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "existing-secret"},
						},
					},
				}
				err := (&defaulter).Default(context.TODO(), &pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pod).To(
					HaveField("Spec",
						HaveField("ImagePullSecrets",
							ContainElements([]corev1.LocalObjectReference{
								{Name: "existing-secret"},
								{Name: "my-secret1"},
								{Name: "my-secret2"},
							}),
						),
					),
				)
			})

			It("Should do nothing cause pod ImagePullSecrets secrets already has correct value", func() {
				pod = corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: containersTrusted,
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "my-secret1"},
							{Name: "existing-secret"},
							{Name: "my-secret2"},
						},
					},
				}
				err := (&defaulter).Default(context.TODO(), &pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pod).To(
					HaveField("Spec",
						HaveField("ImagePullSecrets",
							Equal([]corev1.LocalObjectReference{
								{Name: "my-secret1"},
								{Name: "existing-secret"},
								{Name: "my-secret2"},
							}),
						),
					),
				)
			})
		})
		Context("ImagePullSecrets is empty", func() {
			BeforeEach(func() {
				helpers.CONFIG.MutationImagePullSecrets = []corev1.LocalObjectReference{}
			})

			It("Should do nothing on pod with empty pullSecrets", func() {
				pod = corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: containersTrusted,
					},
				}
				err := (&defaulter).Default(context.TODO(), &pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pod).To(HaveField("Spec", HaveField("ImagePullSecrets", BeEmpty())))
			})

			It("Should do nothing on pod with existing pullSecrets", func() {
				pod = corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "test",
					},
					Spec: corev1.PodSpec{
						Containers: containersTrusted,
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "my-secret1"},
						},
					},
				}
				err := (&defaulter).Default(context.TODO(), &pod)
				Expect(err).ToNot(HaveOccurred())
				Expect(pod).To(HaveField("Spec",
					HaveField("ImagePullSecrets",
						Equal([]corev1.LocalObjectReference{
							{Name: "my-secret1"},
						}),
					),
				))
			})

		})
		AfterEach(func() {
			helpers.CONFIG.MutationImagePullSecrets = []corev1.LocalObjectReference{}
		})
	})

	Describe("Dry run and warn", func() {
		BeforeEach(func() {
			mutatedContainers = AddContainerImageDigest(containersTrusted, "test")
			helpers.CONFIG.MutationDryRun = true
			helpers.CONFIG.ValidationMode = helpers.ValidationModeWarn
		})
		It("Should not modify containers", func() {
			mutatedContainers = AddContainerImageDigest(containersTrusted, "test")
			Expect(mutatedContainers).To(Equal(containersTrusted))
		})
		It("Should not deny and not send warnings for trusted containers", func() {
			pod = corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: mutatedContainers,
				},
			}
			warn, err := ValidatePod(&pod)
			Expect(warn).To(BeEmpty())
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should return warnings for not trusted but not deny", func() {
			pod = corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: containersNotTrusted,
				},
			}
			warn, err := ValidatePod(&pod)
			Expect(warn).To(HaveEach(HavePrefix("forbidden:")))
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			helpers.CONFIG.MutationDryRun = false
			helpers.CONFIG.ValidationMode = helpers.ValidationModeFail
		})
	})

	Describe("On trusted pod update", func() {
		BeforeEach(func() {
			pod = corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers:     containersTrusted,
					InitContainers: containersTrusted,
				},
			}
			err := (&defaulter).Default(context.TODO(), &pod)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should update images in containers", func() {
			Expect(pod).To(HaveField("Spec", And(
				HaveField("Containers", HaveEach(HaveField("Image", MatchRegexp("@sha256:[a-fA-F0-9]{64}$")))),
				HaveField("InitContainers", HaveEach(HaveField("Image", MatchRegexp("@sha256:[a-fA-F0-9]{64}$")))),
			)))
		})
		It("Should allow the pod and not modify it", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateCreate(context.TODO(), tmpPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
		It("On Update Should deny the pod", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateUpdate(context.TODO(), nil, tmpPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
		It("On Delete Should do nothing", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateDelete(context.TODO(), tmpPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
	})
	Describe("On untrusted pod update", func() {
		BeforeEach(func() {
			pod = corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers:     containersNotTrusted,
					InitContainers: containersTrusted,
				},
			}
			err := (&defaulter).Default(context.TODO(), &pod)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should update some images in containers", func() {
			Expect(pod).To(HaveField("Spec", HaveField("InitContainers", HaveEach(HaveField("Image", MatchRegexp("@sha256:[a-fA-F0-9]{64}$"))))))
			Expect(pod.Spec.Containers[0]).To(HaveField("Image", MatchRegexp("@sha256:[a-fA-F0-9]{64}$")))
		})
		It("On Create Should deny the pod", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateCreate(context.TODO(), tmpPod)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
		It("On Update Should deny the pod", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateUpdate(context.TODO(), nil, tmpPod)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
		It("On Delete Should do nothing", func() {
			tmpPod := pod.DeepCopy()
			warn, err := (&validator).ValidateDelete(context.TODO(), tmpPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(warn).To(BeEmpty())
			Expect(*tmpPod).To(Equal(pod))
		})
	})
})
