// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
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

	generateManagedSeedInfoMetrics(managedSeeds, c.descs[metricGardenManagedSeedInfo], ch)
}

func generateManagedSeedInfoMetrics(managedSeeds []*v1alpha1.ManagedSeed, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	for _, ms := range managedSeeds {
		// Some sanity checks.
		if ms == nil || ms.Spec.Shoot == nil {
			continue
		}

		// Expose a metric that exposes information about a managed seed.
		metric, err := prometheus.NewConstMetric(
			desc,
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
