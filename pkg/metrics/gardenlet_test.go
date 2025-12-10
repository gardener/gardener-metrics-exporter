package metrics

import (
	"fmt"
	"testing"

	gardencorev1 "github.com/gardener/gardener/pkg/apis/core/v1"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	seedv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func setupGardenletResource(generation int64, observedGeneration int64, conditions []gardencorev1beta1.Condition) *seedv1alpha1.Gardenlet {
	gardenlet := &seedv1alpha1.Gardenlet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "test",
			Generation: generation,
		},
		Spec: seedv1alpha1.GardenletSpec{
			Deployment: seedv1alpha1.GardenletSelfDeployment{
				GardenletDeployment: seedv1alpha1.GardenletDeployment{
					ReplicaCount: ptr.To(int32(3)),
				},
				Helm: seedv1alpha1.GardenletHelm{
					OCIRepository: gardencorev1.OCIRepository{
						Ref: ptr.To("test"),
					},
				},
			},
		},
		Status: seedv1alpha1.GardenletStatus{
			ObservedGeneration: observedGeneration,
		},
	}

	if conditions != nil && len(conditions) == 0 {
		gardenlet.Status.Conditions = conditions
	}

	return gardenlet
}

func setupExpectations(gardenlet *seedv1alpha1.Gardenlet, descs map[string]*prometheus.Desc) ([]prometheus.Metric, error) {
	expectations := []prometheus.Metric{}
	if gardenlet.Status.Conditions != nil {
		expectedConditionMetric, err := prometheus.NewConstMetric(
			descs[metricGardenGardenletCondition],
			prometheus.GaugeValue,
			mapConditionStatus(gardenlet.Status.Conditions[0].Status),
			[]string{
				gardenlet.Name,
				string(gardenlet.Status.Conditions[0].Type),
			}...,
		)
		if err != nil {
			return nil, fmt.Errorf("could not create expected condition metric: %v", err)
		}
		expectations = append(expectations, expectedConditionMetric)
	}
	expectedGenerationMetric, err := prometheus.NewConstMetric(
		descs[metricGardenGardenletGeneration],
		prometheus.CounterValue,
		float64(gardenlet.Generation),
		gardenlet.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create expected generation metric: %v", err)
	}
	expectations = append(expectations, expectedGenerationMetric)
	expectedObservedGenerationMetric, err := prometheus.NewConstMetric(
		descs[metricGardenGardenletObservedGeneration],
		prometheus.CounterValue,
		float64(gardenlet.Status.ObservedGeneration),
		gardenlet.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create expected observed generation metric: %v", err)
	}
	expectations = append(expectations, expectedObservedGenerationMetric)

	return expectations, nil
}

func TestGenerateGardenletMetrics(t *testing.T) {
	expectedGeneration := int64(10)
	expectedObservedGeneration := int64(10)
	conditions := []gardencorev1beta1.Condition{
		{
			Type:   seedv1alpha1.GardenletReconciled,
			Status: gardencorev1beta1.ConditionTrue,
		},
	}
	gardenlet := setupGardenletResource(expectedGeneration, expectedObservedGeneration, conditions)
	gardenlets := []*seedv1alpha1.Gardenlet{
		gardenlet,
	}

	descs := getGardenMetricsDefinitions()

	ch := make(chan prometheus.Metric, 3)
	generateGardenletMetrics(gardenlets, descs, ch)
	close(ch)

	expectations, err := setupExpectations(gardenlet, descs)
	if err != nil {
		t.Logf("could not setup expectations: %v", err)
	}

	metrics_count := 0
	for {
		metric, ok := <-ch
		if !ok && metrics_count == len(expectations) {
			break
		}
		assert(t, metric, expectations[metrics_count])
		metrics_count += 1
	}
}

func TestGenerateGardenletMetricsWithNoStatusConditions(t *testing.T) {
	expectedGeneration := int64(10)
	expectedObservedGeneration := int64(10)
	gardenlet := setupGardenletResource(expectedGeneration, expectedObservedGeneration, nil)
	gardenlets := []*seedv1alpha1.Gardenlet{
		gardenlet,
	}

	descs := getGardenMetricsDefinitions()

	ch := make(chan prometheus.Metric, 3)
	generateGardenletMetrics(gardenlets, descs, ch)
	close(ch)

	expectations, err := setupExpectations(gardenlet, descs)
	if err != nil {
		t.Logf("could not setup expectations: %v", err)
	}

	metrics_count := 0
	for {
		metric, ok := <-ch
		if !ok && metrics_count == len(expectations) {
			break
		}
		assert(t, metric, expectations[metrics_count])
		metrics_count += 1
	}
}

func TestGardenletMetricsGenerationWithNoGenerations(t *testing.T) {
	gardenlet := &seedv1alpha1.Gardenlet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: seedv1alpha1.GardenletSpec{
			Deployment: seedv1alpha1.GardenletSelfDeployment{
				GardenletDeployment: seedv1alpha1.GardenletDeployment{
					ReplicaCount: ptr.To(int32(3)),
				},
				Helm: seedv1alpha1.GardenletHelm{
					OCIRepository: gardencorev1.OCIRepository{
						Ref: ptr.To("test"),
					},
				},
			},
		},
		Status: seedv1alpha1.GardenletStatus{
			Conditions: []gardencorev1beta1.Condition{
				{
					Type:   seedv1alpha1.GardenletReconciled,
					Status: gardencorev1beta1.ConditionTrue,
				},
			},
		},
	}
	gardenlets := []*seedv1alpha1.Gardenlet{
		gardenlet,
	}

	descs := getGardenMetricsDefinitions()

	ch := make(chan prometheus.Metric, 3)
	generateGardenletMetrics(gardenlets, descs, ch)
	close(ch)

	scrapeFailuresCh := make(chan prometheus.Metric, 1)
	ScrapeFailures.With(prometheus.Labels{"kind": "gardenlets"}).Collect(scrapeFailuresCh)
	scrapeFailuresMetric := <-scrapeFailuresCh

	dtoMetric := &dto.Metric{}
	err := scrapeFailuresMetric.Write(dtoMetric)
	if err != nil {
		t.Logf("could not get failure scrape metric: %v", err)
		t.FailNow()
	}

	if dtoMetric.Counter == nil {
		t.Logf("could not get failure scrape counter: %v", err)
		t.FailNow()
	}
	value := dtoMetric.Counter.Value
	expectedFailureCounter := float64(0)
	if *value != expectedFailureCounter {
		t.Logf("failure scrape counter, want: %f, got: %v", expectedFailureCounter, *value)
		t.FailNow()
	}

	expectations, err := setupExpectations(gardenlet, descs)
	if err != nil {
		t.Logf("could not setup expectations: %v", err)
	}

	metrics_count := 0
	for {
		metric, ok := <-ch
		if !ok && metrics_count == len(expectations) {
			break
		}
		assert(t, metric, expectations[metrics_count])
		metrics_count += 1
	}
}
