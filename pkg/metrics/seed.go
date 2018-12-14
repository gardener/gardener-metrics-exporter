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
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
	"strconv"
)

// collectProjectMetrics collects the number of projects within a Garden cluster.
func (c gardenMetricsCollector) collectSeedMetrics(ch chan<- prometheus.Metric) {

	seeds, err := c.seedInformer.Lister().List(labels.Everything())
	if err != nil {
		fmt.Printf("seed informer failure")
		return
	}
	metric, err := prometheus.NewConstMetric(c.descs[metricGardenSeedsSum], prometheus.GaugeValue, float64(len(seeds)))
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "seeds-count"}).Inc()
		return
	}
	ch <- metric

	for _, seed := range seeds {

		metric, err := prometheus.NewConstMetric(c.descs[metricGardenSeedInfo], prometheus.GaugeValue, 0,
			seed.ObjectMeta.Name, seed.ObjectMeta.Namespace, seed.Spec.Cloud.Profile, seed.Spec.Cloud.Region,
			strconv.FormatBool(*seed.Spec.Visible), strconv.FormatBool(*seed.Spec.Protected))
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			continue
		}
		ch <- metric
	}
}
