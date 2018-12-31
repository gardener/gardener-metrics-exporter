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

package app

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/gardener/gardener-metrics-exporter/pkg/metrics"
	"github.com/gardener/gardener-metrics-exporter/pkg/server"
	"github.com/gardener/gardener-metrics-exporter/pkg/version"
	clientset "github.com/gardener/gardener/pkg/client/garden/clientset/versioned"
	gardeninformers "github.com/gardener/gardener/pkg/client/garden/informers/externalversions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kubeinformers "k8s.io/client-go/informers"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var log *logrus.Logger

type options struct {
	bindAddress    string
	port           int
	kubeconfigPath string
}

func (o *options) validate() bool {
	// Validate only if the kubeconfig file exits, when a path is given.
	if o.kubeconfigPath != "" {
		if _, err := os.Stat(o.kubeconfigPath); os.IsNotExist(err) {
			log.Errorf("kubeconfig does not exits on path %s", o.kubeconfigPath)
			return false
		}
	}

	// Validate if passed ip is a valid ip.
	if ip := net.ParseIP(o.bindAddress); ip == nil {
		log.Errorf("bind-address is not a valid ip %s", o.bindAddress)
		return false
	}

	// Validate if port is in range.
	if o.port < 0 || o.port > 65535 {
		log.Errorf("port is out of range: %d", o.port)
		return false
	}
	return true
}

// NewStartGardenMetricsExporter creates a new GardenMetricsExporter command.
func NewStartGardenMetricsExporter(logger *logrus.Logger, closeCh chan os.Signal) *cobra.Command {
	log = logger
	options := options{}
	cmd := &cobra.Command{
		Use:  "gardener-metrics-exporter",
		Long: "A Prometheus exporter for Gardener related metrics.",
		Run: func(cmd *cobra.Command, args []string) {
			if !options.validate() {
				os.Exit(1)
			}
			if err := run(&options, closeCh); err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
		},
	}
	cmd.AddCommand(version.GetVersionCmd())
	cmd.Flags().StringVar(&options.bindAddress, "bind-address", "0.0.0.0", "bind address for the webserver")
	cmd.Flags().IntVar(&options.port, "port", 2718, "port for the webserver")
	cmd.Flags().StringVar(&options.kubeconfigPath, "kubeconfig", "", "path to kubeconfig file for a Garden cluster")
	return cmd
}

func run(o *options, closeCh chan os.Signal) error {
	stopCh := make(chan struct{})

	// Create informer factories to create informers.
	gardenInformerFactory, kubeInformerFactory, err := setupInformerFactories(o.kubeconfigPath, stopCh)
	if err != nil {
		return err
	}

	// Create informers.
	var (
		shootInformer       = gardenInformerFactory.Garden().V1beta1().Shoots().Informer()
		seedInformer        = gardenInformerFactory.Garden().V1beta1().Seeds().Informer()
		projectInformer     = gardenInformerFactory.Garden().V1beta1().Projects().Informer()
		rolebindingInformer = kubeInformerFactory.Rbac().V1().RoleBindings().Informer()
	)

	// Start the factories and wait until the creates informes has synce
	gardenInformerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(make(<-chan struct{}), shootInformer.HasSynced, seedInformer.HasSynced, projectInformer.HasSynced) {
		return errors.New("Timed out waiting for Garden caches to sync")
	}

	kubeInformerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(make(<-chan struct{}), rolebindingInformer.HasSynced) {
		return errors.New("Timed out waiting for Kube caches to sync")
	}
	// Start the metrics collector
	metrics.SetupMetricsCollector(
		gardenInformerFactory.Garden().V1beta1().Shoots(),
		gardenInformerFactory.Garden().V1beta1().Seeds(),
		gardenInformerFactory.Garden().V1beta1().Projects(),
		kubeInformerFactory.Rbac().V1().RoleBindings(), log)

	// Start the webserver.
	go server.Serve(o.bindAddress, o.port, log, closeCh, stopCh)

	<-stopCh
	log.Info("App shut down.")
	return nil
}

// newClientConfig returns rest config to create a k8s clients. In case that
// kubeconfigPath is empty it tries to create in cluster configuration.
func newClientConfig(kubeconfigPath string) (*rest.Config, error) {
	// In cluster configuration
	if kubeconfigPath == "" {
		log.Info("Use in cluster configuration. This might not work.")
		return rest.InClusterConfig()
	}

	// Kubeconfig based configuration
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	configObj, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, err
	}
	if configObj == nil {
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*configObj, &clientcmd.ConfigOverrides{})
	client, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("ClientConfig is nil")
	}
	return client, nil
}

func setupInformerFactories(kubeconfigPath string, stopCh <-chan struct{}) (gardeninformers.SharedInformerFactory, kubeinformers.SharedInformerFactory, error) {
	restConfig, err := newClientConfig(kubeconfigPath)
	if err != nil {
		return nil, nil, err
	}
	if restConfig == nil {
		return nil, nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	if k8sClient == nil {
		return nil, nil, errors.New("k8sClient is nil")
	}
	gardenClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	if gardenClient == nil {
		return nil, nil, errors.New("gardenClient is nil")
	}
	gardenInformerFactory := gardeninformers.NewSharedInformerFactory(gardenClient, 0)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(k8sClient, 30*time.Second)

	return gardenInformerFactory, kubeInformerFactory, nil
}
