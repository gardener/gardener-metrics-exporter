// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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
	metricShootsPrefix       = "garden_shoots"
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
		m.sendMergedMetric(vals, labels, ch)
	}
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

func (m *MetricTemplate) sendMergedMetric(vals []float64, labels [][]string, ch chan<- prometheus.Metric) {
	l := make(map[string]string)
	l["customization"] = strings.Replace(m.Name, fmt.Sprintf("%s_", metricShootsCustomPrefix), "", 1)
	var noLabels = len(labels) == 0

	if noLabels {
		// TODO  closer look
		if len(vals) == 0 {
			return
		}
		var desc = prometheus.NewDesc(metricShootsCustomPrefix, "Collection of all collected customization metrics.", nil, l)
		var m, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, vals[0])
		if err != nil {
			log.Error(err.Error())
			return
		}
		ch <- m

	} else {
		for i := range vals {
			l[m.Labels[0]] = labels[i][0]
			var desc = prometheus.NewDesc(metricShootsCustomPrefix, "Collection of all collected customization metrics.", nil, l)
			var m, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, vals[i])
			if err != nil {
				log.Error(err.Error())
				continue
			}
			ch <- m
		}
	}
}
