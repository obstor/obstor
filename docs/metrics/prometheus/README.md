# How to monitor Obstor server with Prometheus

[Prometheus](https://prometheus.io) is a cloud-native monitoring platform.

Prometheus offers a multi-dimensional data model with time series data identified by metric name and key/value pairs. The data collection happens via a pull model over HTTP/HTTPS.

Obstor exports Prometheus compatible data by default as an authorized endpoint at `/obstor/v2/metrics/cluster`. Users looking to monitor their Obstor instances can point Prometheus configuration to scrape data from this endpoint. This document explains how to setup Prometheus and configure it to scrape data from Obstor servers.

**Table of Contents**

- [Prerequisites](#prerequisites)
    - [1. Download Prometheus](#1-download-prometheus)
    - [2. Configure authentication type for Prometheus metrics](#2-configure-authentication-type-for-prometheus-metrics)
    - [3. Configuring Prometheus](#3-configuring-prometheus)
        - [3.1 Authenticated Prometheus config](#31-authenticated-prometheus-config)
        - [3.2 Public Prometheus config](#32-public-prometheus-config)
    - [4. Update `scrape_configs` section in prometheus.yml](#4-update-scrapeconfigs-section-in-prometheusyml)
    - [5. Start Prometheus](#5-start-prometheus)
    - [6. Configure Grafana](#6-configure-grafana)
- [List of metrics exposed by Obstor](#list-of-metrics-exposed-by-obstor)

## Prerequisites
To get started with Obstor, refer Obstor QuickStart Document.
Follow below steps to get started with Obstor monitoring using Prometheus.

### 1. Download Prometheus

[Download the latest release](https://prometheus.io/download) of Prometheus for your platform, then extract it

```bash
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

Prometheus server is a single binary called `prometheus` (or `prometheus.exe` on Microsoft Windows). Run the binary and pass `--help` flag to see available options

```bash
./prometheus --help
usage: prometheus [<flags>]

The Prometheus monitoring server

. . .
```

Refer [Prometheus documentation](https://prometheus.io/docs/introduction/first_steps/) for more details.

### 2. Configure authentication type for Prometheus metrics

Obstor supports two authentication modes for Prometheus either `jwt` or `public`, by default Obstor runs in `jwt` mode. To allow public access without authentication for prometheus metrics set environment as follows.

```bash
export OBSTOR_PROMETHEUS_AUTH_TYPE="public"
obstor server ~/test
```

### 3. Configuring Prometheus

#### 3.1 Authenticated Prometheus config

> If Obstor is configured to expose metrics without authentication, you don't need to use `mc` to generate prometheus config. You can skip reading further and move to 3.2 section.

The Prometheus endpoint in Obstor requires authentication by default. Prometheus supports a bearer token approach to authenticate prometheus scrape requests, override the default Prometheus config with the one generated using mc. To generate a Prometheus config for an alias, use mc as follows `mc admin prometheus generate <alias>`.

The command will generate the `scrape_configs` section of the prometheus.yml as follows:

```yaml
scrape_configs:
- job_name: obstor-job
  bearer_token: <secret>
  metrics_path: /obstor/v2/metrics/cluster
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

#### 3.2 Public Prometheus config

If Prometheus endpoint authentication type is set to `public`. Following prometheus config is sufficient to start scraping metrics data from Obstor.
This can be collected from any server once per collection.

##### Cluster
```yaml
scrape_configs:
- job_name: obstor-job
  metrics_path: /obstor/v2/metrics/cluster
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

##### Node (optional)
Optionally you can also collect per node metrics. This needs to be done on a per server instance.
```yaml
scrape_configs:
- job_name: obstor-job
  metrics_path: /obstor/v2/metrics/node
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

### 4. Update `scrape_configs` section in prometheus.yml

To authorize every scrape request, copy and paste the generated `scrape_configs` section in the prometheus.yml and restart the Prometheus service.

### 5. Start Prometheus

Start (or) Restart Prometheus service by running

```bash
./prometheus --config.file=prometheus.yml
```

Here `prometheus.yml` is the name of configuration file. You can now see Obstor metrics in Prometheus dashboard. By default Prometheus dashboard is accessible at `http://localhost:9090`.

### 6. Configure Grafana

After Prometheus is configured, you can use Grafana to visualize Obstor metrics.
Refer the [document here to setup Grafana with Obstor prometheus metrics](/docs/metrics/prometheus/grafana).

## List of metrics exposed by Obstor

Obstor server exposes the following metrics on `/obstor/v2/metrics/cluster` endpoint. All of these can be accessed via Prometheus dashboard. A sample list of exposed metrics along with their definition is available in the demo server at

```bash
curl https://play.obstor.net/obstor/v2/metrics/cluster
```

### List of metrics reported

[The list of metrics reported can be here](/docs/metrics/prometheus/list)
