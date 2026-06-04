# Obstor Multi-Tenant Deployment Guide

This topic provides commands to set up different configurations of hosts, nodes, and drives. The examples provided here can be used as a starting point for other configurations.

1. [Standalone Deployment](#standalone-deployment)
2. [Distributed Deployment](#distributed-deployment)
3. [Cloud Scale Deployment](#cloud-scale-deployment)

## <a name="standalone-deployment"></a>1. Standalone Deployment

To host multiple tenants on a single machine, run one Obstor Server per tenant with a dedicated HTTPS port, configuration, and data directory.

### 1.1 Host Multiple Tenants on a Single Drive

Use the following commands to host 3 tenants on a single drive:

```sh
obstor server --web-address :9001 /data/tenant1
obstor server --web-address :9002 /data/tenant2
obstor server --web-address :9003 /data/tenant3
```

![Example-1](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/multi-tenant-single-drive.svg)

### 1.2 Host Multiple Tenants on Multiple Drives (Erasure Code)

Use the following commands to host 3 tenants on multiple drives:

```sh
obstor server --web-address :9001 /disk{1...4}/data/tenant1
obstor server --web-address :9002 /disk{1...4}/data/tenant2
obstor server --web-address :9003 /disk{1...4}/data/tenant3
```

![Example-2](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/multi-tenant-multiple-drives.svg)

## <a name="distributed-deployment"></a>2. Distributed Deployment

To host multiple tenants in a distributed environment, run several distributed Obstor Server instances concurrently.

### 2.1 Host Multiple Tenants on Multiple Drives (Erasure Code)

Use the following commands to host 3 tenants on a 4-node distributed configuration:

```sh
export OBSTOR_ROOT_USER=<TENANT1_ACCESS_KEY>
export OBSTOR_ROOT_PASSWORD=<TENANT1_SECRET_KEY>
obstor server --web-address :9001 http://192.168.10.1{1...4}/data/tenant1

export OBSTOR_ROOT_USER=<TENANT2_ACCESS_KEY>
export OBSTOR_ROOT_PASSWORD=<TENANT2_SECRET_KEY>
obstor server --web-address :9002 http://192.168.10.1{1...4}/data/tenant2

export OBSTOR_ROOT_USER=<TENANT3_ACCESS_KEY>
export OBSTOR_ROOT_PASSWORD=<TENANT3_SECRET_KEY>
obstor server --web-address :9003 http://192.168.10.1{1...4}/data/tenant3
```

**Note:** Execute the commands on all 4 nodes.

![Example-3](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/multi-tenant-distributed.svg)

**Note**: On distributed systems, credentials must be defined and exported using the `OBSTOR_ROOT_USER` and  `OBSTOR_ROOT_PASSWORD` environment variables. If a domain is required, it must be specified by defining and exporting the `OBSTOR_DOMAIN` environment variable.

## <a name="cloud-scale-deployment"></a>Cloud Scale Deployment

A container orchestration platform (e.g. Kubernetes) is recommended for large-scale, multi-tenant Obstor deployments. See the Obstor Deployment Quickstart Guide to get started with Obstor on orchestration platforms.
