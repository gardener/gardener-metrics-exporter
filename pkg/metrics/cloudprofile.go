// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
)

func (c gardenMetricsCollector) collectCloudProfileMetrics(ch chan<- prometheus.Metric) {
	cloudProfiles, err := c.cloudProfileInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "cloudProfiles"}).Inc()
		return
	}

	generateCloudProfileKubernetesVersionInfoMetrics(cloudProfiles, c.descs[metricGardenCloudProfileKubernetesVersionInfo], ch)
	generateCloudProfileMachineImageVersionInfoMetrics(cloudProfiles, c.descs[metricGardenCloudProfileMachineImageVersionInfo], ch)
}

func generateCloudProfileKubernetesVersionInfoMetrics(cloudProfiles []*gardenv1beta1.CloudProfile, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	for _, cloudProfile := range cloudProfiles {
		for _, kubernetes := range cloudProfile.Status.Kubernetes.Versions {
			metric, err := prometheus.NewConstMetric(
				desc,
				prometheus.GaugeValue,
				0,
				[]string{
					cloudProfile.Name,
					string(kubernetes.Classification),
					kubernetes.Version,
				}...,
			)
			if err != nil {
				ScrapeFailures.With(prometheus.Labels{"kind": "cloudProfiles"}).Inc()
				continue
			}
			ch <- metric
		}
	}
}

func generateCloudProfileMachineImageVersionInfoMetrics(cloudProfiles []*gardenv1beta1.CloudProfile, desc *prometheus.Desc, ch chan<- prometheus.Metric) {
	for _, cloudProfile := range cloudProfiles {
		for _, machineImage := range cloudProfile.Status.MachineImages {
			for _, version := range machineImage.Versions {
				metric, err := prometheus.NewConstMetric(
					desc,
					prometheus.GaugeValue,
					0,
					[]string{
						cloudProfile.Name,
						machineImage.Name,
						string(version.Classification),
						version.Version,
					}...,
				)
				if err != nil {
					ScrapeFailures.With(prometheus.Labels{"kind": "cloudProfiles"}).Inc()
					continue
				}
				ch <- metric
			}
		}
	}
}
