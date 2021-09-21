module github.com/gardener/gardener-metrics-exporter

go 1.17

require (
	github.com/gardener/gardener v1.31.3
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
)

replace (
	github.com/gardener/gardener-resource-manager/api => github.com/gardener/gardener-resource-manager/api v0.25.0
	k8s.io/client-go => k8s.io/client-go v0.21.2
)
