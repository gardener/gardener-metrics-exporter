// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generateSeedConditionMetrics(t *testing.T) {
	seed := &gardenv1beta1.Seed{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-seed",
		},
		Spec: gardenv1beta1.SeedSpec{
			Provider: gardenv1beta1.SeedProvider{
				Type:   "aws",
				Region: "eu-west-1",
			},
		},
		Status: gardenv1beta1.SeedStatus{
			Conditions: []gardenv1beta1.Condition{
				{
					Type:   "GardenletReady",
					Status: gardenv1beta1.ConditionTrue,
				},
			},
		},
	}

	desc := prometheus.NewDesc(
		metricGardenSeedCondition,
		"Condition state of a Seed. Possible values: -1=Unknown|0=Unhealthy|1=Healthy|2=Progressing",
		[]string{"name", "condition", "iaas", "region"},
		nil,
	)

	ch := make(chan prometheus.Metric, 1)
	generateSeedConditionMetrics(seed, desc, ch)

	metric := <-ch

	expected, _ := prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		1, // ConditionTrue → Healthy
		"test-seed", "GardenletReady", "aws", "eu-west-1",
	)

	assert(t, expected, metric)
}

func Test_generateSeedConditionMetrics_skipsEmptyConditionType(t *testing.T) {
	seed := &gardenv1beta1.Seed{
		ObjectMeta: metav1.ObjectMeta{Name: "test-seed"},
		Spec: gardenv1beta1.SeedSpec{
			Provider: gardenv1beta1.SeedProvider{Type: "gcp", Region: "us-east1"},
		},
		Status: gardenv1beta1.SeedStatus{
			Conditions: []gardenv1beta1.Condition{
				{Type: "", Status: gardenv1beta1.ConditionTrue},
			},
		},
	}

	ch := make(chan prometheus.Metric, 1)
	generateSeedConditionMetrics(seed, nil, ch)

	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty condition type, got %d", len(ch))
	}
}
