// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package template

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Type define a metric type for a Prometheus metric.
type Type string

var (
	// Gauge is a type which refers to Prometheus Gauge value.
	Gauge Type = "gauge"

	// Counter is a type which refers to Prometheus Counter value.
	Counter Type = "counter"
)

const (
	metricShootsCustomPrefix = "garden_shoots_custom"
	metricShootsPrefix = "garden_shoots"
)

// MetricTemplate define a template for metrics of same kind. It holds all necessary
// information about the metric and instructions how to collect metric samples.
type MetricTemplate struct {
	Name        string
	Help        string
	Labels      []string
	Type        Type
	desc        *prometheus.Desc
	CollectFunc func(interface{}, ...interface{}) (*[]float64, *[][]string, error)
}

// Register registers the MetricTemplate to the Prometheus Gatherer to allow
// the collection of metric samples which are created based on the template.
func (m *MetricTemplate) Register(ch chan<- *prometheus.Desc) {
	m.desc = prometheus.NewDesc(m.Name, m.Help, m.Labels, nil)
	ch <- m.desc
}

// Collect triggers the collection of metric samples which are created based on
// template by executing the internal CollectFunc.
func (m *MetricTemplate) Collect(ch chan<- prometheus.Metric, obj interface{}, parameters ...interface{}) {
	values, labelValues, err := m.CollectFunc(obj, parameters...)
	if err != nil {
		log.Error(err.Error())
		return
	}

	var (
		vals     = *values
		labels   = *labelValues
		noLabels = len(labels) == 0
	)

	if !noLabels && len(vals) != len(labels) {
		log.Error("Amount of values does not fit with amount of labelsets")
		return
	}

	for i := range vals {
		var metric prometheus.Metric
		if noLabels {
			metric, err = prometheus.NewConstMetric(m.desc, mapType(m.Type), vals[i])
		} else {
			metric, err = prometheus.NewConstMetric(m.desc, mapType(m.Type), vals[i], labels[i]...)
		}

		if err != nil {
			log.Error(err.Error())
			return
		}
			ch <- metric
	}

	// build and send merged metric for customization metrics
	if strings.Contains(m.Name, metricShootsCustomPrefix) {
		mm := m.buildMergedMetric(vals, labels)
		if mm != nil {
			ch <- mm
		}
	}

	return
}

func mapType(t Type) prometheus.ValueType {
	switch t {
	case Gauge:
		return prometheus.GaugeValue
	case Counter:
		return prometheus.CounterValue
	default:
		return prometheus.UntypedValue
	}
}

func (m *MetricTemplate) buildMergedMetric(vals []float64, labels [][]string) prometheus.Metric {
	l := make(map[string]string)
	l["customizations"] = strings.Replace(m.Name, fmt.Sprintf("%s_", metricShootsCustomPrefix), "", 1)
	var noLabels = len(labels) == 0

	if noLabels {
		// TODO  closer look
		if len(vals) == 0 {
			return nil
		}
		var desc = prometheus.NewDesc(fmt.Sprintf("%s_merged", metricShootsCustomPrefix), "Collection of all collected customization metrics.", nil, l)
		var m, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, vals[0])
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		return m

	} else {
		for i := range vals {
			l[m.Labels[0]] = labels[i][0]
			var desc = prometheus.NewDesc(fmt.Sprintf("%s_merged", metricShootsCustomPrefix), "Collection of all collected customization metrics.", nil, l)
			var m, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, vals[i])
			if err != nil {
				log.Error(err.Error())
				return nil
			}
			return m
		}
	}
	return nil
}