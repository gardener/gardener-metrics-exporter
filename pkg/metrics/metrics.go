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
	gardeninformers "github.com/gardener/gardener/pkg/client/garden/informers/externalversions/garden/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	rbacinformers "k8s.io/client-go/informers/rbac/v1"
)

const (
	metricGardenProjectsSum    = "garden_projects_total"
	metricGardenSeedsSum       = "garden_seeds_total"
	metricGardenProjectsStatus = "garden_projects_status"
	metricGardenUsersSum       = "garden_users_total"

	// Seed metric
	metricGardenSeedInfo = "garden_seed_info"

	// Shoot metric (available also for Shoots which act as Seed)
	metricGardenShootInfo              = "garden_shoot_info"
	metricGardenShootCondition         = "garden_shoot_condition"
	metricGardenShootOperationState    = "garden_shoot_operation_states"
	metricGardenShootOperationProgress = "garden_shoot_operation_progress"
	metricGardenShootNodeMaxTotal      = "garden_shoot_node_max_total"
	metricGardenShootNodeMinTotal      = "garden_shoot_node_min_total"
	metricGardenShootResponseDuration  = "garden_shoot_response_duration_milliseconds"

	// Aggregated Shoot metrics (exclude Shoots which act as Seed)
	metricGardenOperationsTotal = "garden_shoot_operations_total"
)

func getGardenMetricsDefinitions() map[string]*prometheus.Desc {
	return map[string]*prometheus.Desc{
		metricGardenProjectsSum:    prometheus.NewDesc(metricGardenProjectsSum, "Count of projects.", nil, nil),
		metricGardenSeedsSum:       prometheus.NewDesc(metricGardenSeedsSum, "Count of seeds.", nil, nil),
		metricGardenProjectsStatus: prometheus.NewDesc(metricGardenProjectsStatus, "Status of projects.", []string{"name", "cluster", "phase"}, nil),
		metricGardenUsersSum:       prometheus.NewDesc(metricGardenUsersSum, "Count of users.", nil, nil),

		metricGardenSeedInfo: prometheus.NewDesc(metricGardenSeedInfo, "Information to a Seed.", []string{"name", "namespace", "iaas", "region", "visible", "protected"}, nil),

		metricGardenShootInfo:              prometheus.NewDesc(metricGardenShootInfo, "Information to a Shoot.", []string{"name", "project", "iaas", "version", "region", "seed"}, nil),
		metricGardenShootOperationState:    prometheus.NewDesc(metricGardenShootOperationState, "Operation state of a Shoot.", []string{"name", "project", "operation"}, nil),
		metricGardenShootOperationProgress: prometheus.NewDesc(metricGardenShootOperationProgress, "Operation progress of a Shoot.", []string{"name", "project", "operation"}, nil),
		metricGardenShootCondition:         prometheus.NewDesc(metricGardenShootCondition, "Condition state of Shoot.", []string{"name", "project", "condition", "operation", "purpose", "is_seed"}, nil),
		metricGardenShootResponseDuration:  prometheus.NewDesc(metricGardenShootResponseDuration, "Response time of the Shoot API server. Not provided when not reachable.", []string{"name", "project"}, nil),

		metricGardenShootNodeMaxTotal: prometheus.NewDesc(metricGardenShootNodeMaxTotal, "Max node count of a Shoot.", []string{"name", "project"}, nil),
		metricGardenShootNodeMinTotal: prometheus.NewDesc(metricGardenShootNodeMinTotal, "Min node count of a Shoot.", []string{"name", "project"}, nil),

		metricGardenOperationsTotal: prometheus.NewDesc(metricGardenOperationsTotal, "Count of ongoing operations.", []string{"operation", "state", "iaas", "seed", "version", "region"}, nil),
	}
}

type gardenMetricsCollector struct {
	shootInformer       gardeninformers.ShootInformer
	seedInformer        gardeninformers.SeedInformer
	projectInformer     gardeninformers.ProjectInformer
	rolebindingInformer rbacinformers.RoleBindingInformer
	descs               map[string]*prometheus.Desc
	logger              *logrus.Logger
}

// Describe implements the prometheus.Describe interface, which intends the gardenMetricsCollector to be a Prometheus collector.
func (c *gardenMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
}

// Collect implements the prometheus.Collect interface, which intends the gardenMetricsCollector to be a Prometheus collector.
func (c *gardenMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectProjectMetrics(ch)
	c.collectShootMetrics(ch)
	c.collectSeedMetrics(ch)
	c.collectUserMetrics(ch)
}

// SetupMetricsCollector takes informers to configure the metrics collectors.
func SetupMetricsCollector(shootInformer gardeninformers.ShootInformer, seedInformer gardeninformers.SeedInformer,
	projectInformer gardeninformers.ProjectInformer, rolebindingInformer rbacinformers.RoleBindingInformer, logger *logrus.Logger) {
	metricsCollector := gardenMetricsCollector{
		shootInformer:       shootInformer,
		seedInformer:        seedInformer,
		projectInformer:     projectInformer,
		rolebindingInformer: rolebindingInformer,
		descs:               getGardenMetricsDefinitions(),
		logger:              logger,
	}
	prometheus.MustRegister(&metricsCollector)
	prometheus.MustRegister(ScrapeFailures)
}
