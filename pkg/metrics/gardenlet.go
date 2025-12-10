// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// collectGardenletMetrics collects Gardenlet metrics.
func (c gardenMetricsCollector) collectGardenletMetrics(ch chan<- prometheus.Metric) {
	gardenlets, err := c.gardenletInformer.Lister().Gardenlets(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "gardenlets"}).Inc()
		return
	}

	for _, gardenlet := range gardenlets {
		// Export a metric for each condition of the Gardenlet.
		for _, condition := range gardenlet.Status.Conditions {
			if condition.Type == "" {
				continue
			}
			metric, err := prometheus.NewConstMetric(
				c.descs[metricGardenGardenletCondition],
				prometheus.GaugeValue,
				mapConditionStatus(condition.Status),
				[]string{
					gardenlet.Name,
					string(condition.Type),
				}...,
			)
			if err != nil {
				ScrapeFailures.With(prometheus.Labels{"kind": "gardenlets"}).Inc()
				continue
			}
			ch <- metric
		}

		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenGardenletGeneration],
			prometheus.CounterValue,
			float64(gardenlet.GetGeneration()),
			gardenlet.Name,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "gardenlets"}).Inc()
		} else {
			ch <- metric
		}

		metric, err = prometheus.NewConstMetric(
			c.descs[metricGardenGardenletObservedGeneration],
			prometheus.CounterValue,
			float64(gardenlet.Status.ObservedGeneration),
			gardenlet.Name,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "gardenlets"}).Inc()
		} else {
			ch <- metric
		}
	}
}
