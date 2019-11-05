# Gardener Metrics Exporter
The `gardener-metrics-exporter` is a [Prometheus](https://prometheus.io/) metrics exporter for [Gardener](https://github.com/gardener/gardener) service-level related metrics.

**This application requires Go 1.9 or later.**

## Metrics
|Metric|Description|Scope|Type|
|-----|-----------|-----|----|
|garden_shoot_operation_states|Operation state of a Shoot|Shoot|Gauge|
|garden_shoot_info|Information to a Shoot|Shoot|Gauge|
|garden_shoot_condition|Condition state of a Shoot|Shoot|Gauge|
|garden_shoot_node_min_total|Min node count of a Shoot|Shoot|Gauge|
|garden_shoot_node_max_total|Max node count of a Shoot|Shoot|Gauge|
|garden_shoot_operations_total|Count of ongoing operations|Shoot|Gauge|
|garden_shoot_operation_states|Operation State of a Shoot|Shoot|Gauge|
|garden_shoot_operation_progress_percent|Operation Percentage of a Shoot|Shoot|Gauge|
|garden_shoot_response_duration_milliseconds| Response time of the Shoot API server (Not provided when not reachable)|Shoot|Gauge|
|garden_seed_info|Information to a Seed|Seed|Gauge|
|garden_seed_condition|Condition State of a Seed|Seed|Gauge|
|garden_projects_status|Status of Garden Projects|Projects|Gauge|
|garden_users_total|Count of users|Users|Gauge|
|garden_scrape_failure_total|Total count of scraping failures, grouped by kind/group of metric(s)|App|Counter|

## Grafana Dashboards
Some [Grafana](https://grafana.com/) dashboards are included in the `dashboards` folder. Simply import them and make sure you have your Prometheus data source named to `cluster-prometheus`.

## Usage
First, clone the repo into your `$GOPATH`.
```sh
mkdir -p "$GOPATH/src/github.com/gardener"
git clone https://github.com/gardener/gardener-metrics-exporter.git "$GOPATH/src/github.com/gardener/gardener-metrics-exporter"

cd "$GOPATH/src/github.com/gardener/gardener-metrics-exporter"
```

### Local
The metrics exporter needs to run against a Gardener enviroment (Kubernetes cluster extendend with `core.gardener.cloud/v1alpha1` api group). Such an enviroment can be created by following the instructions here: https://github.com/gardener/gardener/blob/master/docs/development/local_setup.md

If the current-context of your `$HOME/.kube/config` point to a Gardener enviroment then you can simply run:
```sh
make start
```

If you plan to pass a specific kubeconfig then you need to build the app locally and pass a kubeconfig to the binary.
Let's build the app for your enviroment locally and run it. The binary will be located in the `./bin` directory of the repository.
```sh
# Build
make build-local

# Run
./bin/gardener-metrics-exporter --kubeconfig=<path-to-kubeconfig-file>
```

**Be aware:** The user in the kubeconfig needs permissions to ``GET, LIST, WATCH`` the resources ``Shoot, Seed, Project, Plant (core.gardener.cloud/v1alpha1)`` in all namespaces of the cluster.

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
