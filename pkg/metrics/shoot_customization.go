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

	gardenv1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

const metricShootCustomPrefix = "garden_shoot_custom"
const unknown = "unknown"

var shootCustomizationMetrics = []*metricDefintion{
	// Kube API Server customisation.
	{
		name:      fmt.Sprintf("%s_kubeapiserver_basicauth", metricShootCustomPrefix),
		help:      "Shoot Kube apiserver basic auth enabled. (0=disabled|1=enabled)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var basicAuthEnabled float64
			if shoot.Spec.Kubernetes.KubeAPIServer != nil && shoot.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication != nil && *shoot.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication {
				basicAuthEnabled = 1.0
			}
			return &basicAuthEnabled, &[]string{shoot.Name, *projectName}, nil
		},
	},
	{
		name:      fmt.Sprintf("%s_kubeapiserver_auditpolicy", metricShootCustomPrefix),
		help:      "Shoot Kube apiserver audit policy configured. (0=disabled|1=enabled)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var auditLogPolicyConfigured float64
			if shoot.Spec.Kubernetes.KubeAPIServer != nil && shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig != nil && shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy != nil && shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef != nil {
				auditLogPolicyConfigured = 1.0
			}
			return &auditLogPolicyConfigured, &[]string{shoot.Name, *projectName}, nil
		},
	},

	// Kube Controller Manager customisation.
	{
		name:      fmt.Sprintf("%s_kubecontrollermanager_nodecidrmasksize", metricShootCustomPrefix),
		help:      "The node cidr size configured on Kube controller manager of a Shoot. (0=no cidr mask size configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var nodeCidrMaskSize float64
			if shoot.Spec.Kubernetes.KubeControllerManager != nil && shoot.Spec.Kubernetes.KubeControllerManager.NodeCIDRMaskSize != nil {
				nodeCidrMaskSize = float64(*shoot.Spec.Kubernetes.KubeControllerManager.NodeCIDRMaskSize)
			}
			return &nodeCidrMaskSize, &[]string{shoot.Name, *projectName}, nil
		},
	},

	// Kubelet customisation.
	{
		name:      fmt.Sprintf("%s_kubelet_podpidlimit", metricShootCustomPrefix),
		help:      "The pod pid limit configured on the kubelet(s) of a Shoot. (0=no pod pid limit configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var podPidLimit float64
			if shoot.Spec.Kubernetes.Kubelet != nil && shoot.Spec.Kubernetes.Kubelet.PodPIDsLimit != nil {
				podPidLimit = float64(*shoot.Spec.Kubernetes.Kubelet.PodPIDsLimit)
			}
			return &podPidLimit, &[]string{shoot.Name, *projectName}, nil
		},
	},

	// Kube Proxy customisation.
	{
		name:      fmt.Sprintf("%s_kubeproxy_mode", metricShootCustomPrefix),
		help:      "Proxy mode which the Kube Proxy(ies) of the Shoot use.",
		labels:    []string{"name", "project", "mode"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var (
				mode    = unknown
				enabled float64
			)
			if shoot.Spec.Kubernetes.KubeProxy != nil && shoot.Spec.Kubernetes.KubeProxy.Mode != nil {
				mode = string(*shoot.Spec.Kubernetes.KubeProxy.Mode)
			}
			return &enabled, &[]string{shoot.Name, *projectName, mode}, nil
		},
	},

	// Worker pool customisation.
	{
		name:      fmt.Sprintf("%s_worker_poolscount", metricShootCustomPrefix),
		help:      "Count of workerpools configured for a Shoot.",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var workerPoolCount = float64(len(shoot.Spec.Provider.Workers))
			return &workerPoolCount, &[]string{shoot.Name, *projectName}, nil
		},
	},
	{
		name:      fmt.Sprintf("%s_worker_taints", metricShootCustomPrefix),
		help:      "Shoot worker pool have taints conifgured. (0=not configured|1=configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var taintsConfigured float64
			for _, w := range shoot.Spec.Provider.Workers {
				if len(w.Taints) > 0 {
					taintsConfigured = 1.0
				}
			}
			return &taintsConfigured, &[]string{shoot.Name, *projectName}, nil
		},
	},
	{
		name:      fmt.Sprintf("%s_worker_labels", metricShootCustomPrefix),
		help:      "Shoot worker pool have labels conifgured. (0=not configured|1=configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var labelsConfigured float64
			for _, w := range shoot.Spec.Provider.Workers {
				if len(w.Labels) > 0 {
					labelsConfigured = 1.0
				}
			}
			return &labelsConfigured, &[]string{shoot.Name, *projectName}, nil
		},
	},

	// Network customisation.
	{
		name:      fmt.Sprintf("%s_network_customdomain", metricShootCustomPrefix),
		help:      "Shoot use custom dns domains. (0=not configured|1=configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			var useCustomDNSDomains float64
			if shoot.Spec.DNS != nil && len(shoot.Spec.DNS.Providers) > 0 {
				useCustomDNSDomains = 1.0
			}
			return &useCustomDNSDomains, &[]string{shoot.Name, *projectName}, nil
		},
	},
	{
		name:      fmt.Sprintf("%s_network_existingnetwork", metricShootCustomPrefix),
		help:      "Shoot cluster is deployed to existing network. (0=not configured|1=configured)",
		labels:    []string{"name", "project"},
		valueType: prometheus.GaugeValue,
		collectFunc: func(obj interface{}, params ...interface{}) (*float64, *[]string, error) {
			shoot, projectName, err := checkShootCustomizationParameters(obj, params...)
			if err != nil {
				return nil, nil, err
			}

			// TODO See: https://github.wdf.sap.corp/kubernetes/gardener-adoption/blob/master/stats/miner.py#L156
			var useExistingNetwork float64
			return &useExistingNetwork, &[]string{shoot.Name, *projectName}, nil
		},
	},
}

func registerShootCustomizationMetrics(ch chan<- *prometheus.Desc) {
	for _, c := range shootCustomizationMetrics {
		c.desc = prometheus.NewDesc(c.name, c.help, c.labels, nil)
		ch <- c.desc
	}
}

func collectShootCustomizationMetrics(shoot *gardenv1alpha1.Shoot, projectName *string, ch chan<- prometheus.Metric) {
	for _, c := range shootCustomizationMetrics {
		c.collect(ch, shoot, projectName)
	}
}

func checkShootCustomizationParameters(obj interface{}, params ...interface{}) (*gardenv1alpha1.Shoot, *string, error) {
	if len(params) != 1 {
		return nil, nil, fmt.Errorf("Invalid amount of parameters")
	}
	shoot, ok := obj.(*gardenv1alpha1.Shoot)
	if !ok {
		return nil, nil, fmt.Errorf("Type conversion error")
	}
	projectName, ok := params[0].(*string)
	if !ok {
		return nil, nil, fmt.Errorf("Type conversion error")
	}
	return shoot, projectName, nil
}
