#####################      builder       #####################
FROM golang:1.11.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-metrics-exporter
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -o /go/bin/gardener-metrics-exporter \
  -ldflags="-s -w \
    -X github.com/gardener/gardener-metrics-exporter/pkg/version.gitVersion=$(cat VERSION) \
    -X github.com/gardener/gardener-metrics-exporter/pkg/version.gitCommit=$(git rev-parse --verify HEAD) \
    -X github.com/gardener/gardener-metrics-exporter/pkg/version.buildDate=$(date --rfc-3339=seconds | sed 's/ /T/')" \
  cmd/main.go

#############      gardener-metrics-exporter     #############
FROM alpine:3.7 AS metrics-exporter

RUN apk add --update bash curl

COPY --from=builder /go/bin/gardener-metrics-exporter /gardener-metrics-exporter

WORKDIR /

ENTRYPOINT ["/gardener-metrics-exporter"]
