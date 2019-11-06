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

	"github.com/gardener/gardener-metrics-exporter/pkg/template"
	"github.com/gardener/gardener-metrics-exporter/pkg/utils"
	gardenv1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
)

const metricPlantPrefix = "garden_plant"

/*
	metricGardenPlantInfo:      prometheus.NewDesc(metricGardenPlantInfo, "Information about a plant.", []string{"name", "project", "provider", "region", "version"}, nil),
	metricGardenPlantCondition: prometheus.NewDesc(metricGardenPlantCondition, "Condition state of a Plant. Possible values: -1=Unknown|0=Unhealthy|1=Healthy|2=Progressing", []string{"name", "project", "condition"}, nil),
*/

var plantMetrics = []*template.MetricTemplate{
	{
		Name:   fmt.Sprintf("%s_info", metricPlantPrefix),
		Help:   "Information about a plant.",
		Labels: []string{"name", "project", "provider", "region", "version"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			plant, projectName, err := checkPlantMetricParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var (
				k8sVersion = unknown
				provider   = unknown
				region     = unknown
			)
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

			return &[]float64{1.0}, &[][]string{{plant.ObjectMeta.Name, *projectName, provider, region, k8sVersion}}, nil
		},
	},

	{
		Name:   fmt.Sprintf("%s_condition", metricPlantPrefix),
		Help:   "Condition state of a Plant. (-1=Unknown|0=Unhealthy|1=Healthy|2=Progressing)",
		Labels: []string{"name", "project", "condition"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			plant, projectName, err := checkPlantMetricParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}
			var (
				conditionCount = len(plant.Status.Conditions)
				values         = make([]float64, 0, conditionCount)
				labels         = make([][]string, 0, conditionCount)
			)
			for _, c := range plant.Status.Conditions {
				labels = append(labels, []string{plant.ObjectMeta.Name, *projectName, string(c.Type)})
				values = append(values, mapConditionStatus(c.Status))
			}
			return &values, &labels, nil
		},
	},
}

func checkPlantMetricParameters(obj interface{}, params ...interface{}) (*gardenv1alpha1.Plant, *string, error) {
	if len(params) != 1 {
		return nil, nil, fmt.Errorf("Invalid amount of parameters")
	}
	plant, ok := obj.(*gardenv1alpha1.Plant)
	if !ok {
		return nil, nil, utils.NewTypeConversionError()
	}
	projectName, ok := params[0].(*string)
	if !ok {
		return nil, nil, utils.NewTypeConversionError()
	}
	return plant, projectName, nil
}

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
		projectName, err := findProject(projects, plant.Namespace)
		if err != nil {
			continue
		}

		for _, c := range plantMetrics {
			c.Collect(ch, plant, projectName)
		}
	}
}

func registerPlantMetrics(ch chan<- *prometheus.Desc) {
	for _, c := range plantMetrics {
		c.Register(ch)
	}
}

func collectPlantMetrics(plants []*gardenv1alpha1.Plant, projects []*gardenv1alpha1.Project, ch chan<- prometheus.Metric) {
	for _, plant := range plants {
		projectName, err := findProject(projects, plant.Namespace)
		if err != nil {
			continue
		}

		for _, c := range plantMetrics {
			c.Collect(ch, plant, projectName)
		}
	}
}
