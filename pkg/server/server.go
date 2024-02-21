// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var landingPage = []byte(`<html>
<head><title>Gardener Metrics Exporter</title></head>
<body>
<h1>Gardener Metrics Exporter</h1>
<p><a href='/metrics'>Metrics</a></p>
</body>
</html>
`)

// Serve start the webserver and configure gracefull shut downs.
func Serve(ctx context.Context, bindAddress string, port int, logger *logrus.Logger, stopCh chan struct{}) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type: text/html; charset=utf-8")
		if _, err := w.Write(landingPage); err != nil {
			logger.Warnf("Error writing HTTP response: %v", err)
		}
	})

	server := http.Server{
		Addr: fmt.Sprintf("%s:%d", bindAddress, port),
	}

	go func() {
		<-ctx.Done()
		logger.Info("Shutting down webserver...")

		// New requests should not keep alive connections anymore.
		server.SetKeepAlivesEnabled(false)

		// Shutdown webserver.
		ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Could not gracefully stop the webserver. %s", err.Error())
		}
		logger.Info("Webserver stopped.")
		close(stopCh)
	}()

	logger.Infof("Starting webserver on port %d...", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server starting error. %s", err.Error())
	}
}
