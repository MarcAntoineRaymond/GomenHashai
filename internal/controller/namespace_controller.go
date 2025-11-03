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

package controller

import (
	"context"

	"github.com/GomenHashai/gomenhashai/internal/helpers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciler struct {
	client.Client
	Logger logr.Logger
}

func (r *NamespaceReconciler) Start(ctx context.Context) error {
	nsList := &corev1.NamespaceList{}
	if err := r.List(ctx, nsList); err != nil {
		return err
	}

	for _, ns := range nsList.Items {
		// Trigger a reconciliation manually
		_, err := r.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Name: ns.Name},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ns := &corev1.Namespace{}
	// Skip excluded namespaces
	for _, excluded := range helpers.CONFIG.PullSecretsExemptedNamespaces {
		if req.Name == excluded {
			r.Logger.Info("[🐾IntegrityPatrol] Skipping excluded namespace", "namespace", req.Name)
			return ctrl.Result{}, nil
		}
	}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil // namespace deleted
		}
		return ctrl.Result{}, err
	}

	// Check label selector
	if !helpers.CONFIG.PullSecretsNamespaceSelectorLabels.Matches(labels.Set(ns.Labels)) {
		r.Logger.Info("[🐾IntegrityPatrol] Namespace does not match selector; skipping", "namespace", req.Name)
		return ctrl.Result{}, nil
	}

	for _, cred := range helpers.PULL_SECRETS_CREDENTIALS {
		secretName := cred.Name
		secret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ns.Name}, secret)

		if errors.IsNotFound(err) {
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: ns.Name,
				},
				Type: corev1.SecretTypeDockerConfigJson,
				Data: map[string][]byte{
					".dockerconfigjson": cred.DockerCfg,
				},
			}

			if err := r.Create(ctx, newSecret); err != nil {
				r.Logger.Error(err, "[🐾IntegrityPatrol] failed to create Secret", "namespace", ns.Name, "secret", secretName)
				return ctrl.Result{}, err
			}
			r.Logger.Info("[🐾IntegrityPatrol] Created Docker pull Secret", "namespace", ns.Name, "secret", secretName)
		} else if err != nil {
			return ctrl.Result{}, err
		} else {
			// Check if contents differ (for updates)
			if string(secret.Data[".dockerconfigjson"]) != string(cred.DockerCfg) {
				secret.Data = map[string][]byte{
					".dockerconfigjson": cred.DockerCfg,
				}
				if err := r.Update(ctx, secret); err != nil {
					r.Logger.Error(err, "[🐾IntegrityPatrol] failed to update Secret", "namespace", ns.Name, "secret", secretName)
					return ctrl.Result{}, err
				}
				r.Logger.Info("[🐾IntegrityPatrol] Updated Docker pull Secret", "namespace", ns.Name, "secret", secretName)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
