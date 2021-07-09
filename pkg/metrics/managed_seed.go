// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// collectManagedSeedMetrics collect managed seed metrics.
func (c gardenMetricsCollector) collectManagedSeedMetrics(ch chan<- prometheus.Metric) {

	managedSeeds, err := c.managedSeedInformer.Lister().ManagedSeeds(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "managedSeeds"}).Inc()
		return
	}

	for _, ms := range managedSeeds {
		// Some sanity checks.
		if ms == nil || ms.Spec.Shoot == nil {
			continue
		}

		// Expose a metric that exposes information about a managed seed.
		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenManagedSeedInfo],
			prometheus.GaugeValue,
			0,
			[]string{
				ms.Name,
				ms.Spec.Shoot.Name,
			}...,
		)

		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "managedSeeds"}).Inc()
			return
		}

		ch <- metric
	}
}
