// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"regexp"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
)

var userServiceAccountRegExp *regexp.Regexp

func init() {
	exp := regexp.MustCompile("^system:serviceaccount.*$")
	if exp == nil {
		panic("Could not compile regular expression.")
	}
	userServiceAccountRegExp = exp
}

// collectProjectMetrics collect Project metrics.
func (c gardenMetricsCollector) collectProjectMetrics(ch chan<- prometheus.Metric) {
	projects, err := c.projectInformer.Lister().List(labels.Everything())
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "projects-count"}).Inc()
		return
	}

	var status float64
	for _, project := range projects {
		switch project.Status.Phase {
		case gardenv1beta1.ProjectPending:
			status = 1
		case gardenv1beta1.ProjectReady:
			status = 0
		case gardenv1beta1.ProjectFailed:
			status = -1
		case gardenv1beta1.ProjectTerminating:
			status = 2
		}
		metric, err := prometheus.NewConstMetric(
			c.descs[metricGardenProjectsStatus],
			prometheus.GaugeValue,
			status,
			[]string{
				project.Name,
				string(project.Status.Phase),
			}...,
		)
		if err != nil {
			ScrapeFailures.With(prometheus.Labels{"kind": "projects-status"}).Inc()
			return
		}
		ch <- metric
	}

	// Determine user counts.
	var (
		metric prometheus.Metric

		users          = make(map[string]bool)
		groups         = make(map[string]bool)
		technicalUsers = make(map[string]bool)
	)

	for _, project := range projects {
		for _, user := range project.Spec.Members {

			// Check if user is of kind service account.
			if match := userServiceAccountRegExp.FindString(user.Name); match != "" {
				technicalUsers[user.Name] = true
				continue
			}

			switch user.Kind {
			case "User":
				if _, exists := users[user.Name]; !exists {
					users[user.Name] = true
				}

			case "Group":
				if _, exists := groups[user.Name]; !exists {
					groups[user.Name] = true
				}

			case "ServiceAccount":
				if _, exists := technicalUsers[user.Name]; !exists {
					technicalUsers[user.Name] = true
				}
			}
		}
	}

	metric, err = prometheus.NewConstMetric(c.descs[metricGardenUsersSum], prometheus.GaugeValue, float64(len(users)), "users")
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "users-count"}).Inc()
		return
	}
	ch <- metric

	metric, err = prometheus.NewConstMetric(c.descs[metricGardenUsersSum], prometheus.GaugeValue, float64(len(groups)), "group")
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "users-count"}).Inc()
		return
	}
	ch <- metric

	metric, err = prometheus.NewConstMetric(c.descs[metricGardenUsersSum], prometheus.GaugeValue, float64(len(technicalUsers)), "technical")
	if err != nil {
		ScrapeFailures.With(prometheus.Labels{"kind": "users-count"}).Inc()
		return
	}
	ch <- metric
}
