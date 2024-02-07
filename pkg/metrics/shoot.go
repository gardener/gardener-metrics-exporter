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
	"sort"
	"strconv"
	"strings"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
)

var (
	shootHealthProbeResponseTimeRegExp *regexp.Regexp

	shootOperations = [5]string{
		string(gardenv1beta1.LastOperationTypeCreate),
		string(gardenv1beta1.LastOperationTypeReconcile),
		string(gardenv1beta1.LastOperationTypeDelete),
		string(gardenv1beta1.LastOperationTypeMigrate),
		string(gardenv1beta1.LastOperationTypeRestore),
	}

	userErrorCodes = map[gardenv1beta1.ErrorCode]bool{
		gardenv1beta1.ErrorInfraUnauthenticated:          true,
		gardenv1beta1.ErrorInfraUnauthorized:             true,
		gardenv1beta1.ErrorInfraQuotaExceeded:            true,
		gardenv1beta1.ErrorInfraRateLimitsExceeded:       false,
		gardenv1beta1.ErrorInfraDependencies:             true,
		gardenv1beta1.ErrorRetryableInfraDependencies:    false,
		gardenv1beta1.ErrorInfraResourcesDepleted:        true,
		gardenv1beta1.ErrorCleanupClusterResources:       true,
		gardenv1beta1.ErrorConfigurationProblem:          true,
		gardenv1beta1.ErrorRetryableConfigurationProblem: true,
		gardenv1beta1.ErrorProblematicWebhook:            true,
	}
)

func init() {
	exp := regexp.MustCompile("^.*\\[response_time:(.*)ms\\]$")
	if exp == nil {
		panic("Could not compile regular expression.")
	}
	shootHealthProbeResponseTimeRegExp = exp
}

// collectShootMetrics collect Shoot metrics.
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

	managedSeeds, err := c.managedSeedInformer.Lister().ManagedSeeds(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "managedSeeds"}).Inc()
		return
	}

	secretBindings, err := c.secretBindingInformer.Lister().SecretBindings(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "secretBindings"}).Inc()
		return
	}

	seeds := c.getSeeds()

	collectShootCustomizationMetrics(shoots, ch)

	secretBindingMap := make(map[string]*gardenv1beta1.SecretBinding)
	for _, secretBinding := range secretBindings {
		secretBindingMap[fmt.Sprintf("%s/%s", secretBinding.Namespace, secretBinding.Name)] = secretBinding
	}

	projectMap := make(map[string]*gardenv1beta1.Project)
	for _, project := range projects {
		projectMap[*project.Spec.Namespace] = project
	}

	for _, shoot := range shoots {
		var costObject, costObjectOwner string

		if secretBinding, ok := secretBindingMap[fmt.Sprintf("%s/%s", shoot.Namespace, shoot.Spec.SecretBindingName)]; ok {
			if project, ok := projectMap[secretBinding.SecretRef.Namespace]; ok {
				costObject = project.GetObjectMeta().GetAnnotations()["billing.gardener.cloud/costObject"]
				costObjectOwner = project.Spec.Owner.Name
			}
		}

		var failureTolerance string
		if shoot.Spec.ControlPlane != nil && shoot.Spec.ControlPlane.HighAvailability != nil {
			failureTolerance = string(shoot.Spec.ControlPlane.HighAvailability.FailureTolerance.Type)
		}

		var (
			isSeed       bool
			purpose, uid string

			iaas = shoot.Spec.Provider.Type
			seed = pointer.StringDeref(shoot.Spec.SeedName, "")
		)
		isSeed = usedAsSeed(shoot, managedSeeds)

		if shoot.Spec.Purpose != nil {
			purpose = string(*shoot.Spec.Purpose)
		}

		if shoot.ObjectMeta.Labels != nil {
			for k, v := range shoot.ObjectMeta.Labels {
				if k == "business-critical" && v == "true" {
					purpose = "business-critical"
				}
			}
		}

		projectName, err := findProject(projects, shoot.Namespace)
		if err != nil {
			c.logger.Error(err.Error())
			continue
		}

		var seedProviderType, seedRegion string
		if seed != "" {
			seedProviderType = seeds[seed].Spec.Provider.Type
			seedRegion = seeds[seed].Spec.Provider.Region
		}

		// Expose a metric, which transport basic information to the Shoot cluster via the metric labels.
		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenShootInfo],
			prometheus.GaugeValue,
			0,
			[]string{
				shoot.Name,
				*projectName,
				iaas,
				shoot.Spec.Kubernetes.Version,
				shoot.Spec.Region,
				seed,
				strconv.FormatBool(isSeed),
				seedProviderType,
				seedRegion,
				string(shoot.UID),
				costObject,
				costObjectOwner,
				failureTolerance,
			}...,
		)

		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			continue
		}
		ch <- metric

		hibernatedVal := 0

		if shoot.Status.IsHibernated {
			hibernatedVal = 1
		}

		uid = string(shoot.UID)

		labels := []string{
			shoot.Name,
			*projectName,
			uid,
		}

		metric, err = prometheus.NewConstMetric(
			c.descs[metricGardenShootHibernated],
			prometheus.GaugeValue,
			float64(hibernatedVal),
			labels...,
		)

		ch <- metric

		shootCreation := shoot.CreationTimestamp
		metric, err = prometheus.NewConstMetric(
			c.descs[metricGardenShootCreation],
			prometheus.GaugeValue,
			float64(shootCreation.Unix()),
			labels...,
		)

		ch <- metric

		// Collect metrics to the node count of the Shoot.
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
					case gardenv1beta1.LastOperationStateSucceeded:
						operationState = 1
					case gardenv1beta1.LastOperationStateProcessing:
						operationState = 2
					case gardenv1beta1.LastOperationStatePending:
						operationState = 3
					case gardenv1beta1.LastOperationStateAborted:
						operationState = 4
					case gardenv1beta1.LastOperationStateError:
						operationState = 5
					case gardenv1beta1.LastOperationStateFailed:
						operationState = 6
					}
					operationProgress = float64(shoot.Status.LastOperation.Progress)
				}

				metric, err := prometheus.NewConstMetric(
					c.descs[metricGardenShootOperationState],
					prometheus.GaugeValue,
					operationState,
					[]string{
						shoot.Name,
						*projectName,
						operation,
					}...,
				)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
				metric, err = prometheus.NewConstMetric(
					c.descs[metricGardenShootOperationProgressPercent],
					prometheus.GaugeValue,
					operationProgress,
					[]string{
						shoot.Name,
						*projectName,
						operation,
					}...,
				)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
			}

			// Export a metric for each condition of the Shoot.
			for _, condition := range shoot.Status.Conditions {
				if condition.Type == "" {
					continue
				}

				metric, err := prometheus.NewConstMetric(
					c.descs[metricGardenShootCondition],
					prometheus.GaugeValue,
					mapConditionStatus(condition.Status),
					[]string{
						shoot.Name,
						*projectName,
						string(condition.Type),
						lastOperation,
						purpose,
						strconv.FormatBool(isSeed),
						iaas,
						seed,
						seedProviderType,
						seedRegion,
						uid,
						strconv.FormatBool(hasUserErrors(shoot.Status.LastErrors)),
						shootIsCompliant(shoot.Status.Constraints),
					}...,
				)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
					continue
				}
				ch <- metric
			}

			// Collect the current count of ongoing operations.
			if !isSeed {
				shootOperationsCounters[fmt.Sprintf(
					"%s:%s:%s:%s:%s:%s",
					lastOperation,
					lastOperationState,
					iaas,
					seed,
					shoot.Spec.Kubernetes.Version,
					shoot.Spec.Region,
				)]++
			}
		}
	}

	c.exposeShootOperations(shootOperationsCounters, ch)
}

func hasUserErrors(lastErrors []gardenv1beta1.LastError) bool {
	for _, lastError := range lastErrors {
		for _, code := range lastError.Codes {
			if userError := userErrorCodes[code]; userError {
				return true
			}
		}
	}
	return false
}

func (c gardenMetricsCollector) collectShootNodeMetrics(shoot *gardenv1beta1.Shoot, projectName *string, ch chan<- prometheus.Metric) {
	var (
		nodeCountMax int32
		nodeCountMin int32
	)

	workers := shoot.Spec.Provider.Workers
	for _, worker := range workers {
		// Expose metrics. Start with max worker node count.
		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenShootWorkerNodeMaxTotal],
			prometheus.GaugeValue,
			float64(worker.Maximum),
			[]string{
				shoot.Name,
				*projectName,
				worker.Name,
				worker.Machine.Type,
			}...,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			return
		}
		ch <- metric

		// Continue with min worker node count.
		metric, err = prometheus.NewConstMetric(
			c.descs[metricGardenShootWorkerNodeMinTotal],
			prometheus.GaugeValue,
			float64(worker.Minimum),
			[]string{
				shoot.Name,
				*projectName,
				worker.Name,
				worker.Machine.Type,
			}...,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			return
		}
		ch <- metric

		nodeCountMax += worker.Maximum
		nodeCountMin += worker.Minimum

		var criName string
		var containerRuntimes []string

		if worker.CRI == nil {
			criName = "docker (default)"
		} else {
			criName = string(worker.CRI.Name)
			for _, runtime := range worker.CRI.ContainerRuntimes {
				containerRuntimes = append(containerRuntimes, runtime.Type)
			}
			sort.Strings(containerRuntimes)
		}

		// Expose metrics about the Shoot's nodes.
		metric, err = prometheus.NewConstMetric(
			c.descs[metricGardenShootNodeInfo],
			prometheus.GaugeValue,
			0,
			[]string{
				shoot.Name,
				*projectName,
				worker.Name,
				worker.Machine.Image.Name,
				*worker.Machine.Image.Version,
				criName,
				strings.Join(containerRuntimes, ", "),
				*worker.Machine.Architecture,
			}...,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
			return
		}
		ch <- metric
	}

	// Expose metrics. Start with max node count.
	metric, err := prometheus.NewConstMetric(
		c.descs[metricGardenShootNodeMaxTotal],
		prometheus.GaugeValue,
		float64(nodeCountMax),
		[]string{
			shoot.Name,
			*projectName,
		}...,
	)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "shoots"}).Inc()
		return
	}
	ch <- metric

	// Continue with min node count.
	metric, err = prometheus.NewConstMetric(
		c.descs[metricGardenShootNodeMinTotal],
		prometheus.GaugeValue,
		float64(nodeCountMin),
		[]string{
			shoot.Name,
			*projectName,
		}...,
	)
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
		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenOperationsTotal],
			prometheus.GaugeValue,
			count,
			labels...,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "shoots-operations-total"}).Inc()
			continue
		}
		ch <- metric
	}
}

func (c gardenMetricsCollector) getSeeds() map[string]*gardenv1beta1.Seed {
	s, err := c.seedInformer.Lister().List(labels.Everything())
	seeds := make(map[string]*gardenv1beta1.Seed)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "seeds"}).Inc()
	}

	for _, seed := range s {
		if seed == nil {
			continue
		}
		seeds[seed.Name] = seed
	}
	return seeds
}

func shootIsCompliant(constraints []gardenv1beta1.Condition) string {
	for _, constraint := range constraints {
		if constraint.Type == gardenv1beta1.ShootMaintenancePreconditionsSatisfied {
			return string(constraint.Status)
		}
	}
	return "Unknown"
}
