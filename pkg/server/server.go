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

package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
func Serve(bindAddress string, port int, logger *logrus.Logger, closeCh chan os.Signal, stopCh chan struct{}) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type: text/html; charset=utf-8")
		w.Write(landingPage)
	})

	server := http.Server{
		Addr: fmt.Sprintf("%s:%d", bindAddress, port),
	}
	go func() {
		<-closeCh
		logger.Info("Received interupt signal. Shutting down webserver...")

		// New requests should not keep alive connections anymore.
		server.SetKeepAlivesEnabled(false)
		server.RegisterOnShutdown(func() {
			logger.Info("Webserver shut down.")
			close(stopCh)
		})

		// Shutdown webserver.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Could not gracefully stop the webserver. %s", err.Error())
		}
	}()

	logger.Infof("Starting webserver on port %d...", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Errorf("Server starting error. %s", err.Error())
	}
	close(closeCh)
}
