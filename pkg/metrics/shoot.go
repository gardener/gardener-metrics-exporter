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

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1/helper"
	"github.com/gardener/gardener/pkg/operation/common"
	operationshoot "github.com/gardener/gardener/pkg/operation/shoot"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	shootHealthProbeResponseTimeRegExp *regexp.Regexp

	shootOperations = [3]string{
		string(gardenv1beta1.ShootLastOperationTypeCreate),
		string(gardenv1beta1.ShootLastOperationTypeReconcile),
		string(gardenv1beta1.ShootLastOperationTypeDelete),
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

	for _, shoot := range shoots {
		// Some Shoot sanity checks.
		if shoot == nil || shoot.Spec.Cloud.Seed == nil {
			continue
		}

		// Determine the cloud provider.
		cloudProvider, err := helper.DetermineCloudProviderInShoot(shoot.Spec.Cloud)
		if err != nil {
			return
		}

		var (
			isSeed  bool
			purpose string

			iaas = string(cloudProvider)
			seed = *(shoot.Spec.Cloud.Seed)
		)
		isSeed = usedAsSeed(shoot)

		if shootPurpose, ok := shoot.Annotations[common.GardenPurpose]; ok {
			purpose = shootPurpose
		}

		// Expose a metric, which transport basic information to the Shoot cluster via the metric labels.
		metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootInfo], prometheus.GaugeValue, 0, shoot.Name, shoot.Namespace, iaas, shoot.Spec.Kubernetes.Version, shoot.Spec.Cloud.Region, seed)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			continue
		}
		ch <- metric

		// Collect metrics to the node count of the Shoot.
		// TODO: Use the metrics of the Machine-Controller-Manager, when available. The mcm should be able to provide the actual amount of nodes/machines.
		c.collectShootNodeMetrics(shoot, cloudProvider, ch)

		if shoot.Status.LastOperation != nil {
			lastOperation := string(shoot.Status.LastOperation.Type)
			lastOperationState := string(shoot.Status.LastOperation.State)

			// Export a metric for any possible operation, which can be ongoing on the Shoot.
			// For currently non ongoing operations the value of the metric not will be set to 0.
			for _, operation := range shootOperations {
				var operationState float64
				if operation == lastOperation {
					switch shoot.Status.LastOperation.State {
					case gardenv1beta1.ShootLastOperationStateSucceeded:
						operationState = 1
					case gardenv1beta1.ShootLastOperationStateProcessing:
						operationState = 2
					case gardenv1beta1.ShootLastOperationStatePending:
						operationState = 3
					case gardenv1beta1.ShootLastOperationStateError:
						operationState = 4
					case gardenv1beta1.ShootLastOperationStateFailed:
						operationState = 5
					}
				}
				metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootOperationState], prometheus.GaugeValue, operationState, shoot.Name, shoot.Namespace, operation)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
			}

			// Export a metric for each condition of the Shoot.
			for _, condition := range shoot.Status.Conditions {
				metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootCondition], prometheus.GaugeValue, mapConditionStatus(condition.Status), shoot.Name, shoot.Namespace, string(condition.Type), lastOperation, purpose, strconv.FormatBool(isSeed))
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric

				// Handle the ShootAPIServerAvailable condition special. This condition can transport a measured
				// response time of a request to the API Server. This information will be extracted if available
				// and exposed in a seperate metric.
				if condition.Type == gardenv1beta1.ShootAPIServerAvailable {
					c.exposeAPIServerResponseTime(condition, shoot, ch)
				}
			}

			// Collect the current count of ongoing operations.
			if !isSeed {
				shootOperationsCounters[fmt.Sprintf("%s:%s:%s:%s:%s:%s", lastOperation, lastOperationState, iaas, seed, shoot.Spec.Kubernetes.Version, shoot.Spec.Cloud.Region)]++
			}
		}
	}

	c.exposeShootOperations(shootOperationsCounters, ch)
}

func (c gardenMetricsCollector) collectShootNodeMetrics(shoot *gardenv1beta1.Shoot, cloudProvider gardenv1beta1.CloudProvider, ch chan<- prometheus.Metric) {
	var (
		nodeCountMax int
		nodeCountMin int
	)

	operationShoot := operationshoot.Shoot{
		Info:          shoot,
		CloudProvider: cloudProvider,
	}

	workers := operationShoot.GetWorkers()
	for _, worker := range workers {
		nodeCountMax += worker.AutoScalerMax
		nodeCountMin += worker.AutoScalerMin
	}

	// Expose metrics. Start with max node count.
	metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootNodeMaxTotal], prometheus.GaugeValue, float64(nodeCountMax), shoot.Name, shoot.Namespace)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
		return
	}
	ch <- metric

	// Continue with min node count.
	metric, err = prometheus.NewConstMetric(c.descs[metricGardenShootNodeMinTotal], prometheus.GaugeValue, float64(nodeCountMin), shoot.Name, shoot.Namespace)
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

func (c gardenMetricsCollector) exposeAPIServerResponseTime(condition gardenv1beta1.Condition, shoot *gardenv1beta1.Shoot, ch chan<- prometheus.Metric) {
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
	metric, err := prometheus.NewConstMetric(c.descs[metricGardenShootResponseDuration], prometheus.GaugeValue, responseTimeConv, shoot.Name, shoot.Namespace)
	if err != nil {
		return
	}
	ch <- metric
}
