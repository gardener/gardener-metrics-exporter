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
	"regexp"
	"strconv"
	"strings"

	gardenv1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	constants "github.com/gardener/gardener/pkg/apis/core/v1alpha1/constants"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	shootHealthProbeResponseTimeRegExp *regexp.Regexp

	shootOperations = [3]string{
		string(gardenv1alpha1.LastOperationTypeCreate),
		string(gardenv1alpha1.LastOperationTypeReconcile),
		string(gardenv1alpha1.LastOperationTypeDelete),
	}
)

func init() {
	exp := regexp.MustCompile("^.*\\[response_time:(.*)ms\\]$")
	if exp == nil {
		panic("Could not compile regular expression.")
	}
	shootHealthProbeResponseTimeRegExp = exp
}

// collectShootMetrics collect Shoot metrics, which contain information to Shoot itself, their state and usage.
func (c gardenMetricsCollector) collectShootMetrics(ch chan<- prometheus.Metric) {
	var (
		shootOperationsCounters = make(map[string]float64)
	)

	// Fetch all Shoots.
	shoots, err := c.shootInformer.Lister().Shoots(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
		return
	}

	projects, err := c.projectInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "projects-count"}).Inc()
		return
	}

	for _, shoot := range shoots {
		// Some Shoot sanity checks.
		if shoot == nil || shoot.Spec.SeedName == nil {
			continue
		}

		var (
			isSeed  bool
			purpose string

			iaas = shoot.Spec.Provider.Type
			seed = *shoot.Spec.SeedName
		)
		isSeed = usedAsSeed(shoot)

		if shootPurpose, ok := shoot.Annotations[constants.GardenPurpose]; ok {
			purpose = shootPurpose
		}

		var projectName string
		for _, project := range projects {
			if project.Spec.Namespace != nil && *project.Spec.Namespace == shoot.Namespace {
				projectName = project.ObjectMeta.Name
				break
			}
		}
		if projectName == "" {
			c.logger.Errorf("no project found for shoot %s", shoot.Name)
		}

		// Expose a metric, which transport basic information to the Shoot cluster via the metric labels.
		metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootInfo], prometheus.GaugeValue, 0, shoot.Name, projectName, iaas, shoot.Spec.Kubernetes.Version, shoot.Spec.Region, seed)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			continue
		}
		ch <- metric

		// Collect metrics to the node count of the Shoot.
		// TODO: Use the metrics of the Machine-Controller-Manager, when available. The mcm should be able to provide the actual amount of nodes/machines.
		c.collectShootNodeMetrics(shoot, projectName, ch)

		if shoot.Status.LastOperation != nil {
			lastOperation := string(shoot.Status.LastOperation.Type)
			lastOperationState := string(shoot.Status.LastOperation.State)

			// Export a metric for any possible operation, which can be ongoing on the Shoot.
			// For currently non ongoing operations the value of the metric not will be set to 0.
			for _, operation := range shootOperations {
				var operationState float64
				var operationProgress float64
				if operation == lastOperation {
					switch shoot.Status.LastOperation.State {
					case gardenv1alpha1.LastOperationStateSucceeded:
						operationState = 1
					case gardenv1alpha1.LastOperationStateProcessing:
						operationState = 2
					case gardenv1alpha1.LastOperationStatePending:
						operationState = 3
					case gardenv1alpha1.LastOperationStateAborted:
						operationState = 4
					case gardenv1alpha1.LastOperationStateError:
						operationState = 5
					case gardenv1alpha1.LastOperationStateFailed:
						operationState = 6
					}
					operationProgress = float64(shoot.Status.LastOperation.Progress)
				}
				metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootOperationState], prometheus.GaugeValue, operationState, shoot.Name, projectName, operation)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
				metric, err = prometheus.NewConstMetric(c.descs[metricGardenShootOperationProgressPercent], prometheus.GaugeValue, operationProgress, shoot.Name, projectName, operation)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
			}

			// Export a metric for each condition of the Shoot.
			for _, condition := range shoot.Status.Conditions {
				metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootCondition], prometheus.GaugeValue, mapConditionStatus(condition.Status), shoot.Name, projectName, string(condition.Type), lastOperation, purpose, strconv.FormatBool(isSeed))
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric

				// Handle the ShootAPIServerAvailable condition special. This condition can transport a measured
				// response time of a request to the API Server. This information will be extracted if available
				// and exposed in a seperate metric.
				if condition.Type == gardenv1alpha1.ShootAPIServerAvailable {
					c.exposeAPIServerResponseTime(condition, shoot, projectName, ch)
				}
			}

			// Collect the current count of ongoing operations.
			if !isSeed {
				shootOperationsCounters[fmt.Sprintf("%s:%s:%s:%s:%s:%s", lastOperation, lastOperationState, iaas, seed, shoot.Spec.Kubernetes.Version, shoot.Spec.Region)]++
			}
		}
	}

	c.exposeShootOperations(shootOperationsCounters, ch)
}

func (c gardenMetricsCollector) collectShootNodeMetrics(shoot *gardenv1alpha1.Shoot, projectName string, ch chan<- prometheus.Metric) {
	var (
		nodeCountMax int32
		nodeCountMin int32
	)

	workers := shoot.Spec.Provider.Workers
	for _, worker := range workers {
		nodeCountMax += worker.Minimum
		nodeCountMin += worker.Maximum
	}

	// Expose metrics. Start with max node count.
	metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootNodeMaxTotal], prometheus.GaugeValue, float64(nodeCountMax), shoot.Name, projectName)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
		return
	}
	ch <- metric

	// Continue with min node count.
	metric, err = prometheus.NewConstMetric(c.descs[metricGardenShootNodeMinTotal], prometheus.GaugeValue, float64(nodeCountMin), shoot.Name, projectName)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
		return
	}
	ch <- metric
}

// exposeShootOperations is a util function which is used to transform a map
// of Shoot operations information into proper metrics and to pass them to the collector.
func (c gardenMetricsCollector) exposeShootOperations(shootOperations map[string]float64, ch chan<- prometheus.Metric) {
	for operationInfos, count := range shootOperations {
		labels := strings.Split(operationInfos, ":")
		metric, err := prometheus.NewConstMetric(c.descs[metricGardenOperationsTotal], prometheus.GaugeValue, count, labels...)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots-operations-total"}).Inc()
			continue
		}
		ch <- metric
	}
}

func (c gardenMetricsCollector) exposeAPIServerResponseTime(condition gardenv1alpha1.Condition, shoot *gardenv1alpha1.Shoot, projectName string, ch chan<- prometheus.Metric) {
	match := shootHealthProbeResponseTimeRegExp.FindAllStringSubmatch(condition.Message, -1)
	if len(match) != 1 || len(match[0]) != 2 {
		return
	}
	responseTime := match[0][1]
	if responseTime == "unknown" {
		return
	}
	responseTimeConv, err := strconv.ParseFloat(responseTime, 64)
	if err != nil {
		return
	}
	metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootResponseDuration], prometheus.GaugeValue, responseTimeConv, shoot.Name, projectName)
	if err != nil {
		return
	}
	ch <- metric
}
