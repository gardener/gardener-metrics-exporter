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
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
)

// collectProjectMetrics collects the number of projects within a Garden cluster.
func (c gardenMetricsCollector) collectProjectMetrics(ch chan<- prometheus.Metric) {

	var status float64 = 0
	projects, err := c.projectInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "projects-count"}).Inc()
		return
	}

	for _, project := range projects {
		switch project.Status.Phase {
		case v1beta1.ProjectPending:
			status = 1
		case v1beta1.ProjectReady:
			status = 0
		case v1beta1.ProjectFailed:
			status = -1
		case v1beta1.ProjectTerminating:
			status = 2
		}
		metric, err := prometheus.NewConstMetric(c.descs[metricGardenProjectsStatus], prometheus.GaugeValue, status, project.ObjectMeta.Name, project.ObjectMeta.ClusterName, string(project.Status.Phase))
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "projects-status"}).Inc()
			return
		}
		ch <- metric
	}
}
