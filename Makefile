# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

IMAGE_REPOSITORY := europe-docker.pkg.dev/gardener-project/public/gardener/metrics-exporter
IMAGE_VERSION    := $(shell cat VERSION)
WORKDIR          := $(shell pwd)
HOMEDIR          := $(shell echo ${HOME})
LDFLAGS          := "-s -w -X github.com/gardener/gardener-metrics-exporter/pkg/version.gitVersion=$(IMAGE_VERSION) -X github.com/gardener/gardener-metrics-exporter/pkg/version.gitCommit=$(shell git rev-parse --verify HEAD) -X github.com/gardener/gardener-metrics-exporter/pkg/version.buildDate=$(shell date --rfc-3339=seconds | sed 's/ /T/')"

.PHONY: start
start:
	@go run $(WORKDIR)/cmd/main.go --kubeconfig $(HOMEDIR)/.kube/config

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-o $(WORKDIR)/bin/gardener-metrics-exporter \
		-ldflags $(LDFLAGS) \
		$(WORKDIR)/cmd/main.go

.PHONY: build-local
build-local:
	@go build \
		-o $(WORKDIR)/bin/gardener-metrics-exporter \
		-ldflags $(LDFLAGS) \
		$(WORKDIR)/cmd/main.go

.PHONY: docker-build
docker-build:
	@docker build \
		-t $(IMAGE_REPOSITORY):$(IMAGE_VERSION) \
		-t $(IMAGE_REPOSITORY):latest \
		-f $(WORKDIR)/Dockerfile .

.PHONY: docker-push
docker-push: docker-build
	@echo "Login to image registry ..."
	@gcloud auth activate-service-account --key-file $(WORKDIR)/dev/gcr-readwrite.json || echo "Login to registry failed with exit code $$?"; exit 1
	@echo "Push image to registry ..."
	@gcloud docker -- push $(IMAGE_REPOSITORY):$(IMAGE_TAG)
	@gcloud docker -- push $(IMAGE_REPOSITORY):latest

.PHONY: clean
clean:
	@ rm -rf $(WORKDIR)/bin

.PHONY: check
check:
	@$(WORKDIR)/.ci/check
