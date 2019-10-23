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
	"strconv"

	gardenv1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
)

// collectProjectMetrics collect Seed metrics.
func (c gardenMetricsCollector) collectSeedMetrics(ch chan<- prometheus.Metric) {
	seeds, err := c.seedInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "seeds"}).Inc()
		return
	}

	for _, seed := range seeds {
		var (
			protected bool
			visible   = true
		)
		for _, t := range seed.Spec.Taints {
			if t.Key == gardenv1alpha1.SeedTaintProtected {
				protected = true
			}
			if t.Key == gardenv1alpha1.SeedTaintInvisible {
				visible = false
			}
		}

		metric, err := prometheus.NewConstMetric(c.descs[metricGardenSeedInfo], prometheus.GaugeValue, 0, seed.ObjectMeta.Name, seed.ObjectMeta.Namespace, seed.Spec.Provider.Type, seed.Spec.Provider.Region, strconv.FormatBool(visible), strconv.FormatBool(protected))
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			continue
		}
		ch <- metric
		// Export a metric for each condition of the Seed.
		for _, condition := range seed.Status.Conditions {
			metric, err := prometheus.NewConstMetric(c.descs[metricGardenSeedCondition], prometheus.GaugeValue, mapConditionStatus(condition.Status), seed.Name, string(condition.Type))
			if err != nil {
				ScrapeFailures.With(prometheus.Labels{"kind": "seeds"}).Inc()
				continue
			}
			ch <- metric
		}
	}
}
