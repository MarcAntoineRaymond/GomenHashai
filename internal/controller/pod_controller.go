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
	"fmt"
	"time"

	"github.com/MarcAntoineRaymond/gomenhashai/internal/helpers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=core,resources=pods,verbs=create;get;list;watch;update;patch;delete

type PodInitializer struct {
	Client client.Client
	Logger logr.Logger
}

func (r *PodInitializer) Start(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	r.Logger.Info("[🐾IntegrityPatrol] investigate existing pods 🔍")

	var podList corev1.PodList
	if err := r.Client.List(ctx, &podList); err != nil {
		return err
	}
	pods := podList.Items
	retries := 0
	maxRetries := 5
	r.Logger.Info("[🐾IntegrityPatrol] has rounded up all existing pods and is ready to bite 🐶")

	// Loop until list is empty as error can occur we may need to retry deleting/updating pods on unexpected failure
	for len(pods) > 0 && retries < maxRetries {

		var remaining []corev1.Pod
		for _, pod := range pods {
			r.Logger.Info("Process pod", "name", pod.Name)

			updateOpts := &client.UpdateOptions{
				FieldManager: "gomenhashai",
			}
			if !helpers.CONFIG["DIGEST_UPDATE_EXISTING_PODS"] {
				updateOpts.DryRun = []string{"All"}
			}

			if err := r.Client.Update(ctx, &pod, updateOpts); err != nil {
				// If err is API forbidden
				if apierrors.IsInvalid(err) || apierrors.IsForbidden(err) {
					r.Logger.Info("[🍣GomenHashai!] this pod is forbidden and will be gently offboarded ☁️✂️ Sayonara, pod-san.", "name", pod.Name)
					if helpers.CONFIG["DIGEST_DELETE_EXISTING_PODS"] {
						if err := r.Client.Delete(ctx, &pod); err != nil {
							r.Logger.Error(err, "[🐾IntegrityPatrol] is embarrassed, an error occured when deleting pod 😶", "name", pod.Name)
							remaining = append(remaining, pod)
							continue
						}
					}
				} else {
					r.Logger.Error(err, "[🐾IntegrityPatrol] unexpected error occured when updating pod, even samurai stumble sometimes ⛩️", "name", pod.Name)
					remaining = append(remaining, pod)
					continue
				}
			} else {
				r.Logger.Info("[🍣GomenHashai] nods respectfully. Pod integrity confirmed.", "name", pod.Name)
			}

			r.Logger.Info("[🐾IntegrityPatrol] finished processing pod", "name", pod.Name)
		}

		pods = remaining
		retries++
		time.Sleep(5 * time.Second)
	}

	if len(pods) > 0 {
		podNames := []string{}
		for _, pod := range pods {
			podNames = append(podNames, pod.Name)
		}
		return fmt.Errorf("[🐾IntegrityPatrol] failed to process some pods: %v", podNames)
	}

	r.Logger.Info("[🐾IntegrityPatrol] existing pods investigation complete 🍜")
	return nil
}
