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
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	usersNamespaceSelector labels.Selector
)

func init() {
	// Define a selector to find rolebindings, which contain bindings for Garden users.
	req, err := labels.NewRequirement(common.GardenRole, selection.Equals, []string{common.GardenRoleMembers})
	if err != nil {
		panic("Could not construct a requirement to select Garden projects.")
	}
	usersNamespaceSelector = labels.NewSelector().Add(*req)
}

func (c *gardenMetricsCollector) collectUserMetrics(ch chan<- prometheus.Metric) {
	roleBindings, err := c.rolebindingInformer.Lister().List(usersNamespaceSelector)
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "users-count"}).Inc()
		return
	}
	users := sets.NewString()
	for _, rb := range roleBindings {
		if rb == nil {
			continue
		}
		for _, subject := range rb.Subjects {
			if subject.Kind == "User" && !users.Has(subject.Name) {
				if !users.Has(subject.Name) {
					users.Insert(subject.Name)
				}
			}
		}
	}

	metric, err := prometheus.NewConstMetric(c.descs[metricGardenUsersSum], prometheus.GaugeValue, float64(users.Len()))
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "users-count"}).Inc()
		return
	}
	ch <- metric
}
