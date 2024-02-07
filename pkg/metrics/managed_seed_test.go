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
	"reflect"
	"testing"

	constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"

	seedmanagementv1alpha1 "github.com/gardener/gardener/pkg/apis/seedmanagement/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generateManagedSeedInfoMetrics(t *testing.T) {

	managedSeedName := "managedSeedName"
	shootName := "shootName"

	managedSeed := &seedmanagementv1alpha1.ManagedSeed{
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedSeedName,
			Namespace: constants.GardenNamespace,
		},
		Spec: seedmanagementv1alpha1.ManagedSeedSpec{
			Gardenlet:    &seedmanagementv1alpha1.Gardenlet{},
			Shoot:        &seedmanagementv1alpha1.Shoot{Name: shootName},
		},
	}

	var managedSeeds []*seedmanagementv1alpha1.ManagedSeed
	managedSeeds = append(managedSeeds, managedSeed)

	desc := prometheus.NewDesc(
		metricGardenManagedSeedInfo,
		"Information about a managed seed.",
		[]string{
			"name",
			"shoot",
		},
		nil,
	)

	ch := make(chan prometheus.Metric, 1)

	generateManagedSeedInfoMetrics(managedSeeds, desc, ch)

	metric := <-ch

	expected, _ := prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		0,
		[]string{
			managedSeedName,
			shootName,
		}...,
	)

	assert(t, expected, metric)

}

func assert(t *testing.T, got interface{}, expected interface{}) {
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %+v\nwant %+v", got, expected)
	}
}
