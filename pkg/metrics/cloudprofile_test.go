// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generateCloudProfileInfoMetrics(t *testing.T) {
	cloudProfileName := "cloudProfileName"

	cloudProfile := &gardenv1beta1.CloudProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: cloudProfileName,
		},
		Status: gardenv1beta1.CloudProfileStatus{
			Kubernetes: &gardenv1beta1.KubernetesStatus{
				Versions: []gardenv1beta1.ExpirableVersionStatus{
					{
						Version:        "1.36.5",
						Classification: gardenv1beta1.ClassificationSupported,
					},
					{
						Version:        "1.37.1",
						Classification: gardenv1beta1.ClassificationPreview,
					},
				},
			},
			MachineImages: []gardenv1beta1.MachineImageStatus{
				{
					Name: "my_os",
					Versions: []gardenv1beta1.ExpirableVersionStatus{
						{
							Version:        "v1",
							Classification: gardenv1beta1.ClassificationDeprecated,
						},
						{
							Version:        "v2",
							Classification: gardenv1beta1.ClassificationSupported,
						},
					},
				},
				{
					Name: "second_os",
					Versions: []gardenv1beta1.ExpirableVersionStatus{
						{
							Version:        "v1",
							Classification: gardenv1beta1.ClassificationSupported,
						},
					},
				},
			},
		},
	}

	var cloudProfiles []*gardenv1beta1.CloudProfile
	cloudProfiles = append(cloudProfiles, cloudProfile)

	desc := prometheus.NewDesc(
		metricGardenCloudProfileKubernetesVersionInfo,
		"Information about Kubernetes versions in a CloudProfile",
		[]string{
			"name",
			"classification",
			"version",
		},
		nil,
	)

	ch := make(chan prometheus.Metric, 2)

	generateCloudProfileKubernetesVersionInfoMetrics(cloudProfiles, desc, ch)

	metric := <-ch

	expected, _ := prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			cloudProfileName,
			"supported",
			"1.36.5",
		}...,
	)

	assert(t, expected, metric)

	metric = <-ch

	expected, _ = prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			cloudProfileName,
			"preview",
			"1.37.1",
		}...,
	)

	assert(t, expected, metric)

	desc = prometheus.NewDesc(
		metricGardenCloudProfileMachineImageVersionInfo,
		"Information about machine image versions in a CloudProfile",
		[]string{
			"name",
			"machine_image_name",
			"classification",
			"version",
		},
		nil,
	)

	ch = make(chan prometheus.Metric, 3)

	generateCloudProfileMachineImageVersionInfoMetrics(cloudProfiles, desc, ch)

	metric = <-ch

	expected, _ = prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			cloudProfileName,
			"my_os",
			"deprecated",
			"v1",
		}...,
	)

	assert(t, expected, metric)

	metric = <-ch

	expected, _ = prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			cloudProfileName,
			"my_os",
			"supported",
			"v2",
		}...,
	)

	assert(t, expected, metric)

	metric = <-ch

	expected, _ = prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			cloudProfileName,
			"second_os",
			"supported",
			"v1",
		}...,
	)

	assert(t, expected, metric)
}
