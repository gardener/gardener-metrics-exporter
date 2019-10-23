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
	"k8s.io/apimachinery/pkg/labels"
)

// collectPlantMetrics collects Plant metrics.
func (c gardenMetricsCollector) collectPlantMetrics(ch chan<- prometheus.Metric) {
	plants, err := c.plantInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "plants"}).Inc()
		return
	}
	projects, err := c.projectInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "projects-count"}).Inc()
		return
	}

	for _, plant := range plants {
		var (
			k8sVersion = "unknown"
			provider   = "unknown"
			region     = "unknown"
		)

		projectName, err := findProject(projects, plant.Namespace)
		if err != nil {
			c.logger.Error(err.Error())
			continue
		}

		if plant.Status.ClusterInfo != nil {
			if plant.Status.ClusterInfo.Kubernetes.Version != "" {
				k8sVersion = plant.Status.ClusterInfo.Kubernetes.Version
			}
			if plant.Status.ClusterInfo.Cloud.Type != "" {
				provider = plant.Status.ClusterInfo.Cloud.Type
			}
			if plant.Status.ClusterInfo.Cloud.Region != "" {
				region = plant.Status.ClusterInfo.Cloud.Region
			}
		}

		metric, err := prometheus.NewConstMetric(c.descs[metricGardenPlantInfo], prometheus.GaugeValue, 0, plant.ObjectMeta.Name, *projectName, provider, region, k8sVersion)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "plants"}).Inc()
			continue
		}
		ch <- metric

		// Export a metric for each condition of the Plant.
		for _, condition := range plant.Status.Conditions {
			metric, err := prometheus.NewConstMetric(c.descs[metricGardenPlantCondition], prometheus.GaugeValue, mapConditionStatus(condition.Status), plant.Name, *projectName, string(condition.Type))
			if err != nil {
				ScrapeFailures.With(prometheus.Labels{"kind": "plants"}).Inc()
				continue
			}
			ch <- metric
		}
	}
}
