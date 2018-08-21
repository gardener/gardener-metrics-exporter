#####################      builder       #####################
FROM golang:1.10.3 AS builder

WORKDIR /go/src/github.com/gardener/gardener-metrics-exporter
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/gardener-metrics-exporter cmd/main.go

#############      gardener-metrics-exporter     #############
FROM alpine:3.7 AS metrics-exporter

RUN apk add --update bash curl

COPY --from=builder /go/bin/gardener-metrics-exporter /gardener-metrics-exporter

WORKDIR /

ENTRYPOINT ["/gardener-metrics-exporter"]
