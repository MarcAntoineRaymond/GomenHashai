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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	GomenhashaiValidationTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_validation_total",
			Help: "Number of pods processed by GomenHashai's validating webhook",
		},
	)
	GomenhashaiMutationTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_mutation_total",
			Help: "Number of pods processed by GomenHashai's mutation webhook",
		},
	)
	GomenhashaiAllowed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_allowed_count",
			Help: "Number of pods Allowed by GomenHashai",
		},
	)
	GomenhashaiDenied = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_denied_count",
			Help: "Number of pods Denied by GomenHashai",
		},
	)
	GomenhashaiWarnings = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_warnings_count",
			Help: "Number of pods processed with Warnings by GomenHashai",
		},
	)
	GomenhashaiMutationExempted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_mutation_exempted_count",
			Help: "Number of pods Exempted processed by GomenHashai during mutation",
		},
	)
	GomenhashaiValidationExempted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_validation_exempted_count",
			Help: "Number of pods Exempted processed by GomenHashai during validation",
		},
	)
	GomenhashaiDeleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gomenhashai_deleted_count",
			Help: "Number of pods Deleted by GomenHashai",
		},
	)
)

func Init() {
	metrics.Registry.MustRegister(GomenhashaiValidationTotal, GomenhashaiMutationTotal, GomenhashaiAllowed, GomenhashaiDenied, GomenhashaiWarnings, GomenhashaiMutationExempted, GomenhashaiValidationExempted, GomenhashaiDeleted)
}
