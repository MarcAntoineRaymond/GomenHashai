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

	"strings"

	"github.com/MarcAntoineRaymond/kintegrity/internal/helpers"
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
		digest := helpers.GetDigest(image)
		if digest != "" {
			image = strings.TrimSuffix(image, "@"+digest)
		}
		// Append digest from mapping or send error if no mapping
		trustedDigest := helpers.GetTrustedDigest(image)
		if trustedDigest != "" {
			image = image + "@" + trustedDigest
			// Only modify image in incoming pod if there is a trusted digest
			if !helpers.DIGEST_MUTATION_DRYRUN {
				container.Image = image
				containers[i] = container
			}
			podlog.Info("Add digest to image", "name", container.Name, "image", container.Image, "digest", trustedDigest)
		} else {
			podlog.Info("No trusted digest for image", "name", container.Name, "image", container.Image)
		}
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

	warnings := admission.Warnings{}

	containersList := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for i, container := range containersList {
		image := container.Image
		digest := helpers.GetDigest(image)
		if digest == "" {
			podlog.Info("Deny pod - no digest found", "name", pod.GetName(), "image", image)
			err := apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image is not using a digest",
				),
			)
			if !helpers.DIGEST_VALIDATION_WARNING {
				return nil, err
			} else {
				warnings = append(warnings, err.Error())
			}
		}
		podlog.Info("Digest found", "name", pod.GetName(), "digest", digest)
		image = strings.TrimSuffix(image, "@"+digest)
		// Get trusted dige
		trustedDigest := helpers.GetTrustedDigest(image)
		// Check if image has a mapping with a trusted digest
		if trustedDigest == "" {
			podlog.Info("Deny pod - image has no trusted digest", "name", pod.GetName(), "image", image, "digest", digest)
			err := apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image does not have a trusted digest",
				),
			)
			if !helpers.DIGEST_VALIDATION_WARNING {
				return nil, err
			} else {
				warnings = append(warnings, err.Error())
			}
		}
		// Check if the image is using the trusted digest
		if trustedDigest != digest {
			podlog.Info("Deny pod - digest is not trusted", "name", pod.GetName(), "image", image, "digest", digest)
			err := apierrors.NewForbidden(
				schema.GroupResource{Group: pod.GroupVersionKind().Group, Resource: pod.Kind},
				pod.Name,
				field.Forbidden(
					field.NewPath("spec").Child("containers").Index(i).Child("image"),
					"image use an untrusted digest",
				),
			)
			if !helpers.DIGEST_VALIDATION_WARNING {
				return nil, err
			} else {
				warnings = append(warnings, err.Error())
			}
		} else {
			podlog.Info("Digest is trusted", "name", pod.GetName(), "image", image, "digest", digest)
		}
	}
	podlog.Info("Allow Pod - Completed digests validation", "name", pod.GetName())
	return warnings, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// Do nothing on delete
	return nil, nil
}
