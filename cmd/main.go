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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	app "github.com/gardener/gardener-metrics-exporter/cmd/gardener-metrics-exporter"
	"github.com/gardener/gardener-metrics-exporter/pkg/utils"
)

func main() {
	logger := utils.NewLogger()

	// Setup signal handler.
	signalCh := make(chan os.Signal, 2)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	go func() {
		<-signalCh
		logger.Info("Received interrupt signal.")
		signal.Stop(signalCh)
		cancel()
	}()

	// Init and run app.
	command := app.NewStartGardenMetricsExporter(ctx, logger)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
