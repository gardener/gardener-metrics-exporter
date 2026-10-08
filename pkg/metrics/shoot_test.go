// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_generateShootConstraintMetrics(t *testing.T) {
	shoot := &gardenv1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-shoot",
			UID:  types.UID("shoot-uid-123"),
		},
		Status: gardenv1beta1.ShootStatus{
			TechnicalID: "shoot--test-project--test-shoot",
			Constraints: []gardenv1beta1.Condition{
				{
					Type:   "ManagedResourcesHonored",
					Status: gardenv1beta1.ConditionFalse,
				},
			},
		},
	}

	descs := getGardenMetricsDefinitions()
	collector := gardenMetricsCollector{
		descs: descs,
	}

	projectName := "test-project"
	ch := make(chan prometheus.Metric, 1)
	collector.generateShootConstraintMetrics(shoot, &projectName, ch)

	expected, _ := prometheus.NewConstMetric(
		descs[metricGardenShootConstraint],
		prometheus.GaugeValue,
		0,
		"test-shoot",
		"test-project",
		"shoot-uid-123",
		"shoot--test-project--test-shoot",
		"ManagedResourcesHonored",
	)
	assert(t, expected, <-ch)
}

func Test_generateShootConstraintMetrics_multipleConstraints(t *testing.T) {
	shoot := &gardenv1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-shoot",
			UID:  types.UID("shoot-uid-123"),
		},
		Status: gardenv1beta1.ShootStatus{
			TechnicalID: "shoot--test-project--test-shoot",
			Constraints: []gardenv1beta1.Condition{
				{
					Type:   "ManagedResourcesHonored",
					Status: gardenv1beta1.ConditionFalse,
				},
				{
					Type:   "ShootMaintenancePreconditionsSatisfied",
					Status: gardenv1beta1.ConditionTrue,
				},
			},
		},
	}

	descs := getGardenMetricsDefinitions()
	collector := gardenMetricsCollector{
		descs: descs,
	}

	projectName := "test-project"
	ch := make(chan prometheus.Metric, 2)
	collector.generateShootConstraintMetrics(shoot, &projectName, ch)

	expected1, _ := prometheus.NewConstMetric(
		descs[metricGardenShootConstraint],
		prometheus.GaugeValue,
		0,
		"test-shoot",
		"test-project",
		"shoot-uid-123",
		"shoot--test-project--test-shoot",
		"ManagedResourcesHonored",
	)
	expected2, _ := prometheus.NewConstMetric(
		descs[metricGardenShootConstraint],
		prometheus.GaugeValue,
		1,
		"test-shoot",
		"test-project",
		"shoot-uid-123",
		"shoot--test-project--test-shoot",
		"ShootMaintenancePreconditionsSatisfied",
	)
	assert(t, expected1, <-ch)
	assert(t, expected2, <-ch)
}

func Test_generateShootConstraintMetrics_skipsEmptyConstraintType(t *testing.T) {
	shoot := &gardenv1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-shoot",
			UID:  types.UID("shoot-uid-123"),
		},
		Status: gardenv1beta1.ShootStatus{
			TechnicalID: "shoot--test-project--test-shoot",
			Constraints: []gardenv1beta1.Condition{
				{Type: "", Status: gardenv1beta1.ConditionTrue},
			},
		},
	}

	collector := gardenMetricsCollector{
		descs: getGardenMetricsDefinitions(),
	}

	projectName := "test-project"
	ch := make(chan prometheus.Metric, 1)
	collector.generateShootConstraintMetrics(shoot, &projectName, ch)

	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty constraint type, got %d", len(ch))
	}
}

func Test_generateShootConstraintMetrics_skipsEmptyConstraints(t *testing.T) {
	shoot := &gardenv1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-shoot",
			UID:  types.UID("shoot-uid-123"),
		},
		Status: gardenv1beta1.ShootStatus{
			TechnicalID: "shoot--test-project--test-shoot",
		},
	}

	collector := gardenMetricsCollector{
		descs: getGardenMetricsDefinitions(),
	}

	projectName := "test-project"
	ch := make(chan prometheus.Metric, 1)
	collector.generateShootConstraintMetrics(shoot, &projectName, ch)

	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty constraints field, got %d", len(ch))
	}
}
