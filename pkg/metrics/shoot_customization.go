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

package metrics

import (
	"fmt"
	"sync"

	"github.com/gardener/gardener-metrics-exporter/pkg/template"
	"github.com/gardener/gardener-metrics-exporter/pkg/utils"
	gardenv1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricShootsCustomPrefix = "garden_shoots_custom"
	metricShootsPrefix       = "garden_shoots"
)

var shootCustomizationMetrics = []*template.MetricTemplate{
	// General customization.
	{
		Name:   fmt.Sprintf("%s_privileged_containers_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which allow privileged containers.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.AllowPrivilegedContainers != nil && *s.Spec.Kubernetes.AllowPrivilegedContainers {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	{
		Name:   fmt.Sprintf("%s_extensions_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have an extension(s) configured.",
		Labels: []string{"extension"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var extensionCounter = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Extensions != nil {
					for _, extension := range s.Spec.Extensions {
						extensionCounter[extension.Type]++
					}
				}
			}
			values, labels := mapLabelAndValues(&extensionCounter)
			return values, labels, nil
		},
	},

	// Kube API Server customization.
	{
		Name:   fmt.Sprintf("%s_apiserver_basicauth_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have basic auth enabled on the kube apiserver.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeAPIServer != nil && s.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication != nil && *s.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_apiserver_auditpolicy_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have an audit log policy configured for the kube apiserver.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeAPIServer != nil && s.Spec.Kubernetes.KubeAPIServer.AuditConfig != nil && s.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy != nil && s.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_apiserver_oidcconfig_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have an open id connect coniguration for the kube apiserver.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeAPIServer != nil && s.Spec.Kubernetes.KubeAPIServer.OIDCConfig != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_apiserver_featuregates_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with enabled kube apiserver feature gates.",
		Labels: []string{"featuregate"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var featureGateCounters = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeAPIServer != nil {
					for featureGate, active := range s.Spec.Kubernetes.KubeAPIServer.FeatureGates {
						if active {
							featureGateCounters[featureGate]++
						}
					}
				}
			}
			values, labels := mapLabelAndValues(&featureGateCounters)
			return values, labels, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_apiserver_admissionplugins_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with enabled kube apiserver admission plugins.",
		Labels: []string{"admissionplugin"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var admissionpluginCounters = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeAPIServer != nil {
					for _, admissionPlugin := range s.Spec.Kubernetes.KubeAPIServer.AdmissionPlugins {
						admissionpluginCounters[admissionPlugin.Name]++
					}
				}
			}
			values, labels := mapLabelAndValues(&admissionpluginCounters)
			return values, labels, nil
		},
	},

	// Kube Controller Manager customization.
	{
		Name:   fmt.Sprintf("%s_kcm_nodecidrmasksize_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have node cidr mask size configured on the kube controller manager.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeControllerManager != nil && s.Spec.Kubernetes.KubeControllerManager.NodeCIDRMaskSize != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_kcm_horizontalpodautoscale_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with horizontal pod autoscaling for the kube controller manager.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeControllerManager != nil && s.Spec.Kubernetes.KubeControllerManager.HorizontalPodAutoscalerConfig != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_kcm_featuregates_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with enabled kube controller manager feature gates.",
		Labels: []string{"featuregate"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var featureGateCounters = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeControllerManager != nil {
					for featureGate, active := range s.Spec.Kubernetes.KubeControllerManager.FeatureGates {
						if active {
							featureGateCounters[featureGate]++
						}
					}
				}
			}
			var (
				timeSeriesCount = len(featureGateCounters)
				values          = make([]float64, 0, timeSeriesCount)
				labelSets       = make([][]string, 0, timeSeriesCount)
			)
			for i, k := range featureGateCounters {
				labelSets = append(labelSets, []string{i})
				values = append(values, k)
			}
			return &values, &labelSets, nil
		},
	},

	// Kube Scheduler customization.
	{
		Name:   fmt.Sprintf("%s_scheduler_featuregates_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with enabled kube scheduler feature gates.",
		Labels: []string{"featuregate"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var featureGateCounters = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeScheduler != nil {
					for featureGate, active := range s.Spec.Kubernetes.KubeScheduler.FeatureGates {
						if active {
							featureGateCounters[featureGate]++
						}
					}
				}
			}
			values, labels := mapLabelAndValues(&featureGateCounters)
			return values, labels, nil
		},
	},

	// Kubelet customization.
	{
		Name:   fmt.Sprintf("%s_kubelet_podpidlimit_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have a pod pid limit configured for the kubelet(s).",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Kubernetes.Kubelet != nil && s.Spec.Kubernetes.Kubelet.PodPIDsLimit != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	// Kube Proxy customization.
	{
		Name:   fmt.Sprintf("%s_proxy_mode_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which by proxy mode configuration for the kube proxy.",
		Labels: []string{"mode"},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}

			var modeCounters = map[string]float64{}
			for _, s := range shoots {
				if s.Spec.Kubernetes.KubeProxy != nil && s.Spec.Kubernetes.KubeProxy.Mode != nil {
					modeCounters[string(*s.Spec.Kubernetes.KubeProxy.Mode)]++
				} else {
					modeCounters[unknown]++
				}
			}
			var (
				timeSeriesCount = len(modeCounters)
				values          = make([]float64, 0, timeSeriesCount)
				labelSets       = make([][]string, 0, timeSeriesCount)
			)
			for i, k := range modeCounters {
				labelSets = append(labelSets, []string{i})
				values = append(values, k)
			}
			return &values, &labelSets, nil
		},
	},

	// Worker pool customization.
	{
		Name:   fmt.Sprintf("%s_worker_multiplepools_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with multiple worker pools.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if len(s.Spec.Provider.Workers) > 1 {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_worker_multizones_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with multi zone worker pools.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				for _, w := range s.Spec.Provider.Workers {
					if len(w.Zones) > 1 {
						counter[0]++
						break
					}
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_worker_taints_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with worker pool taints.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				for _, w := range s.Spec.Provider.Workers {
					if len(w.Taints) > 0 {
						counter[0]++
						break
					}
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_worker_labels_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with worker pool labels.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				for _, w := range s.Spec.Provider.Workers {
					if len(w.Labels) > 0 {
						counter[0]++
						break
					}
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_worker_annotations_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots with worker pool annotations.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				for _, w := range s.Spec.Provider.Workers {
					if len(w.Annotations) > 0 {
						counter[0]++
						break
					}
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	// Network customization.
	{
		Name:   fmt.Sprintf("%s_network_customdomain_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which use a custom dns domain.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.DNS != nil && len(s.Spec.DNS.Providers) > 0 {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	// Addons customization.
	{
		Name:   fmt.Sprintf("%s_addon_nginxingress_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have nginx ingress conroller addon enabled.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Addons != nil && s.Spec.Addons.NginxIngress != nil && s.Spec.Addons.NginxIngress.Enabled {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_addon_kubedashboard_total", metricShootsCustomPrefix),
		Help:   "Count of Shoots which have kubernetes dashboard addon enabled.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Addons != nil && s.Spec.Addons.KubernetesDashboard != nil && s.Spec.Addons.KubernetesDashboard.Enabled {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	// TODO: Hibernation/Maintenance configurations shouldn't be treated as Shoot customization.
	// Therefore the metrics should move perspectively to another location.
	// Hibernation.
	{
		Name:   fmt.Sprintf("%s_hibernation_enabled_total", metricShootsPrefix),
		Help:   "Count of Shoots which have hibernation enabled.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Hibernation != nil && s.Spec.Hibernation.Enabled != nil && *s.Spec.Hibernation.Enabled {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_hibernation_schedule_total", metricShootsPrefix),
		Help:   "Count of Shoots which have a hibernation schedule configured.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Hibernation != nil && len(s.Spec.Hibernation.Schedules) > 0 {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},

	// Maintenance.
	{
		Name:   fmt.Sprintf("%s_maintenance_window_total", metricShootsPrefix),
		Help:   "Count of Shoots which have a maintenance window configured.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Maintenance != nil && s.Spec.Maintenance.TimeWindow != nil {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_maintenance_autoupdate_k8sversion_total", metricShootsPrefix),
		Help:   "Count of Shoots which have autoupdate for kubernetes versions configured.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Maintenance != nil && s.Spec.Maintenance.AutoUpdate != nil && s.Spec.Maintenance.AutoUpdate.KubernetesVersion {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
	{
		Name:   fmt.Sprintf("%s_maintenance_autoupdate_imageversion_total", metricShootsPrefix),
		Help:   "Count of Shoots which have autoupdate for machine image versions configured.",
		Labels: []string{},
		Type:   template.Gauge,
		CollectFunc: func(obj interface{}, params ...interface{}) (*[]float64, *[][]string, error) {
			shoots, ok := obj.([]*gardenv1alpha1.Shoot)
			if !ok {
				return nil, nil, utils.NewTypeConversionError()
			}
			var counter = make([]float64, 1, 1)
			for _, s := range shoots {
				if s.Spec.Maintenance != nil && s.Spec.Maintenance.AutoUpdate != nil && s.Spec.Maintenance.AutoUpdate.MachineImageVersion {
					counter[0]++
				}
			}
			return &counter, &[][]string{}, nil
		},
	},
}

func registerShootCustomizationMetrics(ch chan<- *prometheus.Desc) {
	for _, c := range shootCustomizationMetrics {
		c.Register(ch)
	}
}

func collectShootCustomizationMetrics(shoots []*gardenv1alpha1.Shoot, ch chan<- prometheus.Metric) {
	var (
		wg  sync.WaitGroup
		run = func(c *template.MetricTemplate) {
			c.Collect(ch, shoots)
			wg.Done()
		}
	)
	wg.Add(len(shootCustomizationMetrics))
	for _, c := range shootCustomizationMetrics {
		go run(c)
	}
	wg.Wait()
}

func mapLabelAndValues(list *map[string]float64) (*[]float64, *[][]string) {
	var (
		timeSeriesCount = len(*list)
		values          = make([]float64, 0, timeSeriesCount)
		labelSets       = make([][]string, 0, timeSeriesCount)
	)
	for i, k := range *list {
		labelSets = append(labelSets, []string{i})
		values = append(values, k)
	}
	return &values, &labelSets
}
