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
	"os"
	"time"

	"github.com/MarcAntoineRaymond/kintegrity/internal/helpers"
	kintegrityv1 "github.com/MarcAntoineRaymond/kintegrity/internal/webhook/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

func HandleExistingPods(mgr manager.Manager) {
	time.Sleep(5 * time.Second)

	logger := mgr.GetLogger()
	logger.Info("Start existing pods checks")

	var pods corev1.PodList
	if err := mgr.GetClient().List(context.TODO(), &pods); err != nil {
		logger.Error(err, "unable to list pods at start up")
		os.Exit(1)
	}

	for _, pod := range pods.Items {

		logger.Info("Process pod", "name", pod.Name)

		// TODO process namespace exemption better (matchSelector and objectSelector config has to be similar, need config from YAML)
		if pod.Namespace == os.Getenv("NAMESPACE") {
			logger.Info("Skip pod because namespace is exempted", "name", pod.Name)
			continue
		}

		if helpers.CONFIG["DIGEST_UPDATE_EXISTING_PODS"] {
			pod.Spec.InitContainers = kintegrityv1.AddContainerImageDigest(pod.Spec.InitContainers)
			pod.Spec.Containers = kintegrityv1.AddContainerImageDigest(pod.Spec.Containers)
			logger.Info("Updated pod with digests", "name", pod.Name)
		}

		if helpers.CONFIG["DIGEST_VALIDATE_EXISTING_PODS"] {

			_, err := kintegrityv1.ValidatePod(&pod)
			if err != nil {
				if helpers.CONFIG["DIGEST_DELETE_EXISTING_PODS"] {
					if err := mgr.GetClient().Delete(context.TODO(), &pod); err != nil {
						logger.Error(err, "unable to delete pod", "name", pod.Name)
					}
					logger.Info("Deleted pod", "name", pod.Name)
				}
			}

			logger.Info("Validated pod with digests", "name", pod.Name)
		}
	}
}
