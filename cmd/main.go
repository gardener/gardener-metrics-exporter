// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
