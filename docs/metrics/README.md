## Obstor Monitoring Guide

Obstor server exposes monitoring data over endpoints. Monitoring tools can pick the data from these endpoints. This document lists the monitoring endpoints and relevant documentation.

### Healthcheck Probe

Obstor server has two healthcheck related un-authenticated endpoints, a liveness probe to indicate if server is responding, cluster probe to check if server can be taken down for maintenance.

- Liveness probe available at `/obstor/health/live`
- Cluster probe available at `/obstor/health/cluster`

Read more on how to use these endpoints in [Obstor healthcheck guide](/docs/metrics/healthcheck).

### Prometheus Probe

Obstor allows reading metrics for the entire cluster from any single node. This allows for metrics collection for a Obstor instance across all servers. Thus, metrics collection for instances behind a load balancer can be done without any knowledge of the individual node addresses. The cluster wide metrics can be read at
`<Address for Obstor Service>/obstor/v2/metrics/cluster`.

The additional node specific metrics which include additional go metrics or process metrics are exposed at
`<Address for Obstor Node>/obstor/v2/metrics/node`.

To use this endpoint, setup Prometheus to scrape data from this endpoint. Read more on how to configure and use Prometheus to monitor Obstor server in [How to monitor Obstor server with Prometheus](/docs/metrics/prometheus).

**Deprecated metrics monitoring**

- Prometheus' data available at `/obstor/prometheus/metrics` is deprecated

