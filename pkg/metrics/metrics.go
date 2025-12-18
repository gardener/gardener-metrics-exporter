// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	gardencoreinformers "github.com/gardener/gardener/pkg/client/core/informers/externalversions/core/v1beta1"
	gardensecurityinformers "github.com/gardener/gardener/pkg/client/security/informers/externalversions/security/v1alpha1"
	gardenmanagedseedinformers "github.com/gardener/gardener/pkg/client/seedmanagement/informers/externalversions/seedmanagement/v1alpha1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func getGardenMetricsDefinitions() map[string]*prometheus.Desc {
	return map[string]*prometheus.Desc{
		metricGardenManagedSeedInfo: prometheus.NewDesc(
			metricGardenManagedSeedInfo,
			"Information about a managed seed.",
			[]string{
				"name",
				"shoot",
			},
			nil,
		),

		metricGardenOperationsTotal: prometheus.NewDesc(
			metricGardenOperationsTotal,
			"Count of ongoing operations.",
			[]string{
				"operation",
				"state",
				"iaas",
				"seed",
				"version",
				"region",
			},
			nil,
		),

		metricGardenProjectsStatus: prometheus.NewDesc(
			metricGardenProjectsStatus,
			"Status of projects. Possible Values: -1=Failed|0=Ready|1=Pending|2=Terminating",
			[]string{
				"name",
				"phase",
			},
			nil,
		),

		metricGardenSeedCondition: prometheus.NewDesc(
			metricGardenSeedCondition,
			"Condition state of a Seed. Possible values: -1=Unknown|0=Unhealthy|1=Healthy|2=Progressing",
			[]string{
				"name",
				"condition",
			},
			nil,
		),

		metricGardenSeedInfo: prometheus.NewDesc(
			metricGardenSeedInfo,
			"Information about a Seed.",
			[]string{
				"name",
				"namespace",
				"iaas",
				"region",
				"visible",
				"protected",
				"version",
			},
			nil,
		),

		metricGardenSeedCapacity: prometheus.NewDesc(
			metricGardenSeedCapacity,
			"Seed capacity.",
			[]string{
				"name",
				"namespace",
				"iaas",
				"region",
				"visible",
				"protected",
				"resource",
			},
			nil,
		),

		metricGardenSeedUsage: prometheus.NewDesc(
			metricGardenSeedUsage,
			"Seed usage.",
			[]string{
				"name",
				"namespace",
				"iaas",
				"region",
				"visible",
				"protected",
				"resource",
			},
			nil,
		),

		metricGardenShootCondition: prometheus.NewDesc(
			metricGardenShootCondition,
			"Condition state of a Shoot. Possible values: -1=Unknown|0=Unhealthy|1=Healthy|2=Progressing",
			[]string{
				"name",
				"project",
				"condition",
				"operation",
				"purpose",
				"is_seed",
				"iaas",
				"seed",
				"seed_iaas",
				"seed_region",
				"uid",
				"technical_id",
				"has_user_errors",
				"is_compliant",
			},
			nil,
		),

		metricGardenShootCreation: prometheus.NewDesc(
			metricGardenShootCreation,
			"Timestamp of the shoot creation.",
			[]string{
				"name",
				"project",
				"uid",
				"technical_id",
			},
			nil,
		),

		metricGardenShootHibernated: prometheus.NewDesc(
			metricGardenShootHibernated,
			"Hibernation status of a shoot.",
			[]string{
				"name",
				"project",
				"uid",
				"technical_id",
			},
			nil,
		),

		metricGardenShootInfo: prometheus.NewDesc(
			metricGardenShootInfo,
			"Information about a Shoot.",
			[]string{
				"name",
				"project",
				"iaas",
				"version",
				"region",
				"seed",
				"is_seed",
				"seed_iaas",
				"seed_region",
				"shoot_uid",
				"cost_object",
				"cost_object_type",
				"cost_object_owner",
				"failure_tolerance",
				"technical_id",
				"gardener_version",
				"is_workerless",
				"is_hibernated",
			},
			nil,
		),

		metricGardenShootNodeMaxTotal: prometheus.NewDesc(
			metricGardenShootNodeMaxTotal,
			"Max node count of a Shoot.",
			[]string{"name",
				"project",
				"technical_id",
			},
			nil,
		),

		metricGardenShootNodeMinTotal: prometheus.NewDesc(
			metricGardenShootNodeMinTotal,
			"Min node count of a Shoot.",
			[]string{"name",
				"project",
				"technical_id",
			},
			nil,
		),

		metricGardenShootWorkerNodeMaxTotal: prometheus.NewDesc(
			metricGardenShootWorkerNodeMaxTotal,
			"Max node count of a worker Shoot.",
			[]string{
				"name",
				"project",
				"worker_group",
				"worker_machine_type",
				"technical_id",
			},
			nil,
		),

		metricGardenShootWorkerNodeMinTotal: prometheus.NewDesc(
			metricGardenShootWorkerNodeMinTotal,
			"Min node count of a worker Shoot.",
			[]string{
				"name",
				"project",
				"worker_group",
				"worker_machine_type",
				"technical_id",
			},
			nil,
		),

		metricGardenShootNodeInfo: prometheus.NewDesc(
			metricGardenShootNodeInfo,
			"Information about the nodes in a Shoot.",
			[]string{
				"name",
				"project",
				"worker_group",
				"image",
				"version",
				"cri",
				"container_runtimes",
				"architecture",
				"technical_id",
			},
			nil,
		),

		metricGardenShootOperationProgressPercent: prometheus.NewDesc(
			metricGardenShootOperationProgressPercent,
			"Operation progress percent of a Shoot.",
			[]string{
				"name",
				"project",
				"operation",
				"technical_id",
			},
			nil,
		),

		metricGardenShootOperationState: prometheus.NewDesc(
			metricGardenShootOperationState,
			"Operation state of a Shoot. Possible values: 1=Succeeded|2=Processing|3=Pending|4=Aborted|5=Error|6=Failed. Available operations: 'Create'|'Reconcile'|'Delete'|'Restore'|'Migrate'.",
			[]string{"name",
				"project",
				"operation",
				"technical_id",
			},
			nil,
		),

		metricGardenUsersSum: prometheus.NewDesc(
			metricGardenUsersSum,
			"Count of users.",
			[]string{
				"kind",
			},
			nil,
		),
	}
}

type gardenMetricsCollector struct {
	managedSeedInformer        gardenmanagedseedinformers.ManagedSeedInformer
	shootInformer              gardencoreinformers.ShootInformer
	seedInformer               gardencoreinformers.SeedInformer
	projectInformer            gardencoreinformers.ProjectInformer
	secretBindingInformer      gardencoreinformers.SecretBindingInformer
	credentialsBindingInformer gardensecurityinformers.CredentialsBindingInformer
	descs                      map[string]*prometheus.Desc
	logger                     *logrus.Logger
}

// Describe implements the prometheus.Describe interface, which intends the gardenMetricsCollector to be a Prometheus collector.
func (c *gardenMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
	registerShootCustomizationMetrics(ch)
}

// Collect implements the prometheus.Collect interface, which intends the gardenMetricsCollector to be a Prometheus collector.
// TODO Can we run the collectors in parallel?
func (c *gardenMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectManagedSeedMetrics(ch)
	c.collectProjectMetrics(ch)
	c.collectShootMetrics(ch)
	c.collectSeedMetrics(ch)
}

// SetupMetricsCollector takes informers to configure the metrics collectors.
func SetupMetricsCollector(shootInformer gardencoreinformers.ShootInformer, seedInformer gardencoreinformers.SeedInformer, projectInformer gardencoreinformers.ProjectInformer, managedSeedInformer gardenmanagedseedinformers.ManagedSeedInformer, secretBindingInformer gardencoreinformers.SecretBindingInformer, credentialsBindingInformer gardensecurityinformers.CredentialsBindingInformer, logger *logrus.Logger) {
	metricsCollector := gardenMetricsCollector{
		managedSeedInformer:        managedSeedInformer,
		shootInformer:              shootInformer,
		seedInformer:               seedInformer,
		projectInformer:            projectInformer,
		secretBindingInformer:      secretBindingInformer,
		credentialsBindingInformer: credentialsBindingInformer,
		descs:                      getGardenMetricsDefinitions(),
		logger:                     logger,
	}
	prometheus.MustRegister(&metricsCollector)
	prometheus.MustRegister(ScrapeFailures)
}
