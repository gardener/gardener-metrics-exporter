# Gardener Metrics Exporter
The `gardener-metrics-exporter` is a [Prometheus](https://prometheus.io/) exporter for [Gardener](https://github.com/gardener/gardener) service-level related metrics.

**This application requires Go 1.9 or later.**

## Metrics
|Metic|Description|Scope|Type|
|-----|-----------|-----|----|
|garden_shoot_operation_states|Operation state of a Shoot|Shoot|Gauge|
|garden_shoot_info|Information to a Shoot|Shoot|Gauge|
|garden_shoot_conditions|Condition state of Shoot|Shoot|Gauge|
|garden_shoot_node_min_total|Min node count of a Shoot|Shoot|Gauge|
|garden_shoot_node_max_total|Max node count of a Shoot|Shoot|Gauge|
|garden_shoot_operations_total|Count of ongoing operations|Shoot|Gauge|
|garden_projects_total|Count of projects|Project|Gauge|
|garden_users_total|Count of users|Users|Gauge|
|garden_scrape_failure_total|Total count of scraping failures, grouped by kind/group of metric(s)|App|Counter|

## Grafana Dashboards
Some [Grafana](https://grafana.com/) dashboards are included in the `dashboards/` folder. Simply import them and make sure you have your prometheus data source set to `cluster-prometheus`.

## Usage
First, clone the repo into your `$GOPATH`.
```sh
mkdir -p "$GOPATH/src/github.com/gardener"
git clone https://github.com/gardener/gardener-metrics-exporter.git "$GOPATH/src/github.com/gardener/gardener-metrics-exporter"

cd "$GOPATH/src/github.com/gardener/gardener-metrics-exporter"
```

### Local
Build the metrics-exporter locally and run it by passing a kubeconfig. Such an enviroment can be setup by following the instructions here: https://github.com/gardener/gardener/blob/master/docs/development/local_setup.md

Let's build the app locally.
```sh
go build -o gardener-metrics-exporter cmd/main.go
```

Run the binary and pass a kubeconfig. The kubeconfig must point to a Kubernetes cluster which the Gardener is deployed to.
The user in the config needs permissions to ``GET, LIST, WATCH`` the resources ``shoot(garden.sapcloud.io/v1beta1), namespace(v1), rolebindings (rbac.authorization.k8s.io)`` in all namespaces of the cluster.
```sh
./gardener-metrics-exporter --kubeconfig=<path-to-kubeconfig-file>
```
Verify that everything works by calling the `/metrics` endpoint of the app.
```sh
curl http://localhost:2718/metrics
```
Run a local [Prometheus](https://prometheus.io/download/) instance and add the following scrape config snippet to your config.
```yaml
scrape_configs:
- job_name: 'gardener-metrics-exporter'
  static_configs:
    - targets:
      - localhost:2718
  metric_relabel_configs:
   - source_labels: [ __name__ ]
     regex: '^garden_.*$'
     action: keep
```
Now the metrics should be collected by Prometheus. Open the Prometheus console and query for ``garden_*`` metrics.

### In Cluster
Deploy the metrics-exporter to a kubernetes cluster via helm.
```sh
helm upgrade gardener-metrics-exporter charts/gardener-metrics-exporter --install --namespace=<your-namespace> --values=<path-to-your-values.yaml>
```
For example, the scrape config for your Prometheus could look like this:
```yaml
scrape_configs:
- job_name: 'gardener-metrics-exporter'
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - <helm-deployment-namespace>
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name]
    regex: gardener-metrics-exporter
    action: keep
  metric_relabel_configs:
   - source_labels: [ __name__ ]
     regex: '^garden_.*$'
     action: keep
```