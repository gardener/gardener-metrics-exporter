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
	"context"
	"errors"
	"net"
	"os"

	"github.com/gardener/gardener-metrics-exporter/pkg/metrics"
	"github.com/gardener/gardener-metrics-exporter/pkg/server"
	"github.com/gardener/gardener-metrics-exporter/pkg/version"
	clientset "github.com/gardener/gardener/pkg/client/core/clientset/versioned"
	gardencoreinformers "github.com/gardener/gardener/pkg/client/core/informers/externalversions"
	seedmanagementclientset "github.com/gardener/gardener/pkg/client/seedmanagement/clientset/versioned"
	gardenmanagedseedinformers "github.com/gardener/gardener/pkg/client/seedmanagement/informers/externalversions"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	// Validate whether a kubeconfig file exists under the given path.
	if o.kubeconfigPath != "" {
		if _, err := os.Stat(o.kubeconfigPath); os.IsNotExist(err) {
			log.Errorf("kubeconfig does not exist on path %s", o.kubeconfigPath)
			return false
		}
	}

	// Validate whether the passed IP address is valid.
	if ip := net.ParseIP(o.bindAddress); ip == nil {
		log.Errorf("bind-address is not a valid ip %s", o.bindAddress)
		return false
	}

	// Validate if the port is in range.
	if o.port < 0 || o.port > 65535 {
		log.Errorf("port is out of range: %d", o.port)
		return false
	}
	return true
}

// NewStartGardenMetricsExporter creates a new GardenMetricsExporter command.
func NewStartGardenMetricsExporter(ctx context.Context, logger *logrus.Logger) *cobra.Command {
	log = logger
	options := options{}
	cmd := &cobra.Command{
		Use:  "gardener-metrics-exporter",
		Long: "A Prometheus exporter for Gardener related metrics.",
		Run: func(cmd *cobra.Command, args []string) {
			if !options.validate() {
				os.Exit(1)
			}
			if err := run(ctx, &options); err != nil {
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

func run(ctx context.Context, o *options) error {
	stopCh := make(chan struct{})

	// Create informer factories to create informers.
	gardenInformerFactory, gardenManagedSeedInformerFactory, err := setupInformerFactories(o.kubeconfigPath, stopCh)
	if err != nil {
		return err
	}

	// Create informers.
	var (
		managedSeedInformer   = gardenManagedSeedInformerFactory.Seedmanagement().V1alpha1().ManagedSeeds().Informer()
		shootInformer         = gardenInformerFactory.Core().V1beta1().Shoots().Informer()
		seedInformer          = gardenInformerFactory.Core().V1beta1().Seeds().Informer()
		projectInformer       = gardenInformerFactory.Core().V1beta1().Projects().Informer()
		secretBindingInformer = gardenInformerFactory.Core().V1beta1().SecretBindings().Informer()
	)

	// Start the factories and wait until the informers have synced.
	gardenInformerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(ctx.Done(), shootInformer.HasSynced, seedInformer.HasSynced, projectInformer.HasSynced, secretBindingInformer.HasSynced) {
		return errors.New("Timed out waiting for Garden caches to sync")
	}

	gardenManagedSeedInformerFactory.Start((stopCh))
	if !cache.WaitForCacheSync(ctx.Done(), managedSeedInformer.HasSynced) {
		return errors.New("Timed out waiting for Managed Seed caches to sync")
	}

	// Start the metrics collector
	metrics.SetupMetricsCollector(
		gardenInformerFactory.Core().V1beta1().Shoots(),
		gardenInformerFactory.Core().V1beta1().Seeds(),
		gardenInformerFactory.Core().V1beta1().Projects(),
		gardenManagedSeedInformerFactory.Seedmanagement().V1alpha1().ManagedSeeds(),
		gardenInformerFactory.Core().V1beta1().SecretBindings(),
		log,
	)

	// Start the webserver.
	go server.Serve(ctx, o.bindAddress, o.port, log, stopCh)

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
	kubeconfig, err := os.ReadFile(kubeconfigPath)
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

func setupInformerFactories(kubeconfigPath string, stopCh <-chan struct{}) (gardencoreinformers.SharedInformerFactory, gardenmanagedseedinformers.SharedInformerFactory, error) {
	restConfig, err := newClientConfig(kubeconfigPath)
	if err != nil {
		return nil, nil, err
	}
	if restConfig == nil {
		return nil, nil, err
	}
	gardenClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	if gardenClient == nil {
		return nil, nil, errors.New("gardenClient is nil")
	}
	var gardenManagedSeedClient *seedmanagementclientset.Clientset
	if gardenManagedSeedClient, err = seedmanagementclientset.NewForConfig(restConfig); err != nil {
		return nil, nil, err
	}
	if gardenManagedSeedClient == nil {
		return nil, nil, errors.New("gardenManagedSeedClient is nil")
	}

	gardenInformerFactory := gardencoreinformers.NewSharedInformerFactory(gardenClient, 0)
	gardenManagedSeedInformerFactory := gardenmanagedseedinformers.NewSharedInformerFactory(gardenManagedSeedClient, 0)
	return gardenInformerFactory, gardenManagedSeedInformerFactory, nil
}
