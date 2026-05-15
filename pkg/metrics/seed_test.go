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
		ObjectMeta: metav1.ObjectMeta{Name: "test-seed"},
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

	expected, _ := prometheus.NewConstMetric(desc, prometheus.GaugeValue, 1,
		"test-seed", "GardenletReady", "aws", "eu-west-1",
	)
	assert(t, expected, <-ch)
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

func Test_generateSeedOperationStateMetrics_nilLastOperation(t *testing.T) {
	seed := &gardenv1beta1.Seed{
		ObjectMeta: metav1.ObjectMeta{Name: "test-seed"},
		Spec: gardenv1beta1.SeedSpec{
			Provider: gardenv1beta1.SeedProvider{Type: "aws", Region: "eu-west-1"},
		},
	}

	ch := make(chan prometheus.Metric, 1)
	generateSeedOperationStateMetrics(seed, nil, ch)

	if len(ch) != 0 {
		t.Errorf("expected no metrics for nil LastOperation, got %d", len(ch))
	}
}

func Test_generateSeedOperationStateMetrics_activeOperation(t *testing.T) {
	seed := &gardenv1beta1.Seed{
		ObjectMeta: metav1.ObjectMeta{Name: "test-seed"},
		Spec: gardenv1beta1.SeedSpec{
			Provider: gardenv1beta1.SeedProvider{Type: "aws", Region: "eu-west-1"},
		},
		Status: gardenv1beta1.SeedStatus{
			LastOperation: &gardenv1beta1.LastOperation{
				Type:  gardenv1beta1.LastOperationTypeReconcile,
				State: gardenv1beta1.LastOperationStateSucceeded,
			},
		},
	}

	desc := prometheus.NewDesc(
		metricGardenSeedOperationState,
		"Operation state of a Seed.",
		[]string{"name", "operation"},
		nil,
	)

	ch := make(chan prometheus.Metric, len(seedOperations))
	generateSeedOperationStateMetrics(seed, desc, ch)

	if len(ch) != len(seedOperations) {
		t.Fatalf("expected %d metrics, got %d", len(seedOperations), len(ch))
	}

	// seedOperations iteration order is deterministic; build expected values accordingly.
	expectedValues := map[string]float64{
		string(gardenv1beta1.LastOperationTypeCreate):    0,
		string(gardenv1beta1.LastOperationTypeReconcile): 1, // Succeeded
		string(gardenv1beta1.LastOperationTypeDelete):    0,
		string(gardenv1beta1.LastOperationTypeMigrate):   0,
		string(gardenv1beta1.LastOperationTypeRestore):   0,
	}

	for _, operation := range seedOperations {
		expected, _ := prometheus.NewConstMetric(desc, prometheus.GaugeValue,
			expectedValues[string(operation)],
			"test-seed", string(operation),
		)
		assert(t, expected, <-ch)
	}
}

func Test_generateSeedOperationStateMetrics_stateMapping(t *testing.T) {
	cases := []struct {
		state    gardenv1beta1.LastOperationState
		expected float64
	}{
		{gardenv1beta1.LastOperationStateSucceeded, 1},
		{gardenv1beta1.LastOperationStateProcessing, 2},
		{gardenv1beta1.LastOperationStatePending, 3},
		{gardenv1beta1.LastOperationStateAborted, 4},
		{gardenv1beta1.LastOperationStateError, 5},
		{gardenv1beta1.LastOperationStateFailed, 6},
	}

	desc := prometheus.NewDesc(
		metricGardenSeedOperationState,
		"Operation state of a Seed.",
		[]string{"name", "operation"},
		nil,
	)

	for _, tc := range cases {
		seed := &gardenv1beta1.Seed{
			ObjectMeta: metav1.ObjectMeta{Name: "test-seed"},
			Spec: gardenv1beta1.SeedSpec{
				Provider: gardenv1beta1.SeedProvider{Type: "aws", Region: "eu-west-1"},
			},
			Status: gardenv1beta1.SeedStatus{
				LastOperation: &gardenv1beta1.LastOperation{
					Type:  gardenv1beta1.LastOperationTypeReconcile,
					State: tc.state,
				},
			},
		}

		ch := make(chan prometheus.Metric, len(seedOperations))
		generateSeedOperationStateMetrics(seed, desc, ch)

		// Drain until we find the Reconcile metric (second in seedOperations order).
		var reconcileMetric prometheus.Metric
		for _, op := range seedOperations {
			m := <-ch
			if string(op) == string(gardenv1beta1.LastOperationTypeReconcile) {
				reconcileMetric = m
			}
		}

		expected, _ := prometheus.NewConstMetric(desc, prometheus.GaugeValue,
			tc.expected,
			"test-seed", string(gardenv1beta1.LastOperationTypeReconcile),
		)
		assert(t, expected, reconcileMetric)
	}
}
