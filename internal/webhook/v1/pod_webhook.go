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
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

var DIGEST_MAPPING = map[string]string{
	"busybox:latest":                   "sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f",
	"busybox":                          "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
	"library/busybox":                  "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
	"docker.io/library/busybox":        "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
	"docker.io/library/busybox:stable": "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
	"busybox:stable":                   "sha256:e246aa22ad2cbdfbd19e2a6ca2b275e26245a21920e2b2d0666324cee3f15549",
	"nginx/nginx-ingress:5.0.0-alpine": "sha256:a6c4d7c7270f03a3abb1ff38973f5db98d8660832364561990c4d0ef8b1477af",
}

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithValidator(&PodCustomValidator{}).
		WithDefaulter(&PodCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)

	if !ok {
		return fmt.Errorf("expected an Pod object but got %T", obj)
	}

	podlog.Info("Mutating digests", "name", pod.GetName())

	pod.Spec.InitContainers = addContainerImageDigest(pod.Spec.InitContainers)
	pod.Spec.Containers = addContainerImageDigest(pod.Spec.Containers)

	return nil
}

// Loop container list and append digest to images, return errors for every container images not matching mapping list
func addContainerImageDigest(containers []corev1.Container) []corev1.Container {
	for i, container := range containers {
		image := container.Image
		// Remove digest if already present in image field
		digest := getDigest(image)
		if digest != "" {
			image = strings.TrimSuffix(image, "@"+digest)
		}
		// Append digest from mapping or send error if no mapping
		trustedDigest := getTrustedDigest(image)
		if trustedDigest != "" {
			image = image + "@" + trustedDigest
			podlog.Info("Add digest to image", "name", container.Name, "image", container.Image, "digest", trustedDigest)
		} else {
			podlog.Info("No valid digest for image", "name", container.Name, "image", container.Image)
		}
		container.Image = image
		containers[i] = container
	}
	return containers
}

// +kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomValidator struct is responsible for validating the Pod resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PodCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &PodCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return v.ValidatePod(ctx, obj)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return v.ValidatePod(ctx, newObj)
}

func (v *PodCustomValidator) ValidatePod(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object for the obj but got %T", obj)
	}
	podlog.Info("Validating digests", "name", pod.GetName())

	containersList := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for i, container := range containersList {
		image := container.Image
		digest := getDigest(image)
		if digest == "" {
			podlog.Error(nil, "No digest found", "name", pod.GetName(), "image", image)
			return nil, apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image does not contain a digest",
				),
			)
		}
		podlog.Info("Digest found", "name", pod.GetName(), "digest", digest)
		image = strings.TrimSuffix(image, "@"+digest)
		// Append digest from mapping or send error if no mapping
		trustedDigest := getTrustedDigest(image)
		if trustedDigest == "" {
			return nil, apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image does not have any trusted digest",
				),
			)
		}
		if trustedDigest != digest {
			return nil, apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image contains an untrusted digest",
				),
			)
		}
		podlog.Info("Digest is trusted", "name", pod.GetName(), "image", image, "digest", digest)
	}
	podlog.Info("Completed digests validation", "name", pod.GetName())
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// Do nothing on delete
	return nil, nil
}

// getDigest from container image or return empty
func getDigest(image string) string {
	re := regexp.MustCompile(`@sha256:[a-fA-F0-9]{64}`)
	match := re.FindString(image)
	if match != "" {
		return match[1:]
	}
	return ""
}

// Return digest from mapping for this image or empty string
func getTrustedDigest(image string) string {
	// TODO handle tag or not tag mechanic, if image has tag and digest exist for base image without tag, use this one
	if digest, ok := DIGEST_MAPPING[image]; ok {
		return digest
	}
	return ""
}
