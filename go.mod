module github.com/gardener/gardener-metrics-exporter

go 1.13

require (
	github.com/gardener/gardener v1.4.0
	github.com/prometheus/client_golang v1.3.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace (
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
	github.com/gardener/external-dns-management => github.com/gardener/external-dns-management v0.0.0-20190927090840-6659f5a46d13
	k8s.io/api => k8s.io/api v0.16.8 // 1.16.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.8 // 1.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8 // 1.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8 // 1.16.8
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)
