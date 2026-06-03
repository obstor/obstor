# Obstor Cluster Deployment Guide

This guide covers deploying Obstor in a multi-node cluster configuration with block-level replication, node quotas, geo-distribution, and failure recovery.

## Cluster Overview

Obstor uses **block-level replication** (not erasure coding) for data durability. Each object uploaded to Obstor is:

1. Split into fixed-size, content-addressed blocks (SHA-256 hashed).
2. Replicated **N** times across drives and nodes in the cluster.

The default replication factor is **3** - every block is stored on 3 independent drives/nodes. Obstor uses **quorum-based writes and reads** to ensure consistency:

- **Write quorum**: A write succeeds only when a majority of replicas acknowledge the write.
- **Read quorum**: A read succeeds when a majority of replicas return the same block, guaranteeing you always read the latest committed data.

This model provides strong consistency (read-after-write, list-after-write) while tolerating individual drive and node failures.

> **NOTE:** Block-level replication differs from erasure coding. Replication stores full copies of each block, trading storage efficiency for simplicity and faster recovery. For erasure-coded deployments, see [Erasure Code Guide](/docs/erasure).

## Node Quotas

You can set per-node storage capacity limits using the `OBSTOR_NODE_CAPACITY` environment variable. When a node reaches its capacity threshold, Obstor automatically routes new writes to other nodes in the cluster.

```bash
export OBSTOR_NODE_CAPACITY=10T
```

Accepted units: `K` (kilobytes), `M` (megabytes), `G` (gigabytes), `T` (terabytes), `P` (petabytes).

Examples:

```bash
# 10 terabytes per node
export OBSTOR_NODE_CAPACITY=10T

# 500 gigabytes per node
export OBSTOR_NODE_CAPACITY=500G
```

When capacity is reached on a node:

- New write requests are redirected to nodes with available capacity.
- Existing data on the full node remains accessible for reads.
- Replication traffic for new objects bypasses the full node.

> **NOTE:** `OBSTOR_NODE_CAPACITY` sets a soft limit. Replication obligations may cause a node to slightly exceed its quota to satisfy write quorum requirements.

## Deploying a 3-Node Cluster

### Prerequisites

- 3 servers (physical or virtual), each with 4 dedicated drives mounted at `/mnt/data1` through `/mnt/data4`.
- Network connectivity between all nodes on port `9000` (S3 API) and port `9001` (console).
- Obstor binary installed on all nodes - Obstor Quickstart Guide.
- System clocks synchronized via NTP (nodes should be within 15 minutes of each other).

### Environment Variables

Set the following environment variables on **all nodes** before starting the server:

| Variable | Description |
|---|---|
| `OBSTOR_ROOT_USER` | Admin access key (must be identical on all nodes) |
| `OBSTOR_ROOT_PASSWORD` | Admin secret key (must be identical on all nodes) |
| `OBSTOR_ZONE` | Zone label for this node (e.g. `us-east-1`) |
| `OBSTOR_REPLICATION_FACTOR` | Number of block replicas (default: `3`) |
| `OBSTOR_BLOCK_SIZE` | Size of each content-addressed block (default: `4M`) |
| `OBSTOR_NODE_CAPACITY` | Maximum storage per node (e.g. `10T`) |
| `OBSTOR_CONSISTENCY_MODE` | Consistency level: `consistent`, `degraded`, or `dangerous` (default: `consistent`) |

### Step 1: Start Node 1 (us-east-1)

```bash
export OBSTOR_ROOT_USER=admin
export OBSTOR_ROOT_PASSWORD=secret
export OBSTOR_ZONE=us-east-1
export OBSTOR_REPLICATION_FACTOR=3
export OBSTOR_BLOCK_SIZE=4M
export OBSTOR_NODE_CAPACITY=10T
export OBSTOR_CONSISTENCY_MODE=consistent
./obstor server /mnt/data{1...4}
```

### Step 2: Start Node 2 (eu-west-1)

```bash
export OBSTOR_ROOT_USER=admin
export OBSTOR_ROOT_PASSWORD=secret
export OBSTOR_ZONE=eu-west-1
export OBSTOR_REPLICATION_FACTOR=3
export OBSTOR_BLOCK_SIZE=4M
export OBSTOR_NODE_CAPACITY=10T
export OBSTOR_CONSISTENCY_MODE=consistent
./obstor server /mnt/data{1...4}
```

### Step 3: Start Node 3 (ap-southeast-1)

```bash
export OBSTOR_ROOT_USER=admin
export OBSTOR_ROOT_PASSWORD=secret
export OBSTOR_ZONE=ap-southeast-1
export OBSTOR_REPLICATION_FACTOR=3
export OBSTOR_BLOCK_SIZE=4M
export OBSTOR_NODE_CAPACITY=10T
export OBSTOR_CONSISTENCY_MODE=consistent
./obstor server /mnt/data{1...4}
```

### Distributed Mode (Recommended)

Instead of starting nodes independently, use the distributed mode syntax to connect all nodes into a single cluster. Run this command on **every node**:

```bash
OBSTOR_ROOT_USER=admin \
OBSTOR_ROOT_PASSWORD=secret \
OBSTOR_ZONE=us-east-1 \
OBSTOR_REPLICATION_FACTOR=3 \
./obstor server http://node{1...3}/mnt/data{1...4}
```

> **NOTE:** `{1...3}` uses 3 dots (ellipsis syntax). Using 2 dots `{1..3}` will be interpreted by your shell and will not be passed to Obstor, resulting in incorrect cluster formation.

Set `OBSTOR_ZONE` to the appropriate zone for each node:

| Node | Zone | Drives |
|---|---|---|
| `node1` | `us-east-1` | `/mnt/data1` ... `/mnt/data4` |
| `node2` | `eu-west-1` | `/mnt/data1` ... `/mnt/data4` |
| `node3` | `ap-southeast-1` | `/mnt/data1` ... `/mnt/data4` |

This gives you a 3-node, 12-drive cluster with replication factor 3 spread across 3 geographic zones.

## Geo-Distribution

Obstor is designed for geo-distributed deployments. Each bucket's objects may be spread across multiple nodes and locations based on the replication policy.

### Zone-Aware Replica Placement

The `OBSTOR_ZONE` tag controls how Obstor places block replicas:

- Obstor **always** attempts to spread replicas across different zones for maximum durability.
- With replication factor 3 and 3 zones, each block replica is placed in a **different zone**.
- At least **2 copies** of every block are guaranteed to exist in different zones, even if zone imbalance occurs.

```
Object: report.pdf (3 blocks)
├── Block A (SHA-256: 8f14e4...)
│   ├── Replica 1 → node1 (us-east-1)
│   ├── Replica 2 → node2 (eu-west-1)
│   └── Replica 3 → node3 (ap-southeast-1)
├── Block B (SHA-256: 2c624b...)
│   ├── Replica 1 → node2 (eu-west-1)
│   ├── Replica 2 → node3 (ap-southeast-1)
│   └── Replica 3 → node1 (us-east-1)
└── Block C (SHA-256: 9a3b7f...)
    ├── Replica 1 → node3 (ap-southeast-1)
    ├── Replica 2 → node1 (us-east-1)
    └── Replica 3 → node2 (eu-west-1)
```

### Benefits

- **Disaster recovery**: A full zone outage does not cause data loss.
- **Read locality**: Clients can be routed to the nearest zone for lower latency reads.
- **Regulatory compliance**: Zone tags can align with data residency requirements.

> **NOTE:** Cross-zone replication incurs network transfer costs. Place nodes in the same region for lower latency, or across regions for maximum durability.

## Recovering from Failures

### Single Drive Failure

Data survives on remaining drives. Obstor detects the missing blocks and initiates automatic re-replication from surviving copies.

**Recovery steps:**

1. Replace the failed drive.
2. Mount the new drive at the same path (e.g. `/mnt/data2`).
3. Restart the Obstor server on the affected node.

```bash
# After replacing /mnt/data2 on node1
./obstor server http://node{1...3}/mnt/data{1...4}
```

Obstor detects the empty drive, identifies missing blocks by comparing with the cluster manifest, and re-replicates them from surviving copies on other drives/nodes.

### Single Node Failure

As long as **read quorum** is met, reads continue to succeed. With replication factor 3, read quorum is 2 - so the cluster tolerates 1 node being completely offline.

**Recovery steps:**

1. Replace or repair the failed node.
2. Start Obstor with fresh drives on the replacement node.
3. Obstor automatically syncs missing blocks from the remaining nodes.

```bash
# On the replacement node3
export OBSTOR_ROOT_USER=admin
export OBSTOR_ROOT_PASSWORD=secret
export OBSTOR_ZONE=ap-southeast-1
./obstor server http://node{1...3}/mnt/data{1...4}
```

The cluster detects the fresh node and begins background re-replication to restore the target replication factor.

### Consistency Modes

Obstor provides three consistency modes controlled by `OBSTOR_CONSISTENCY_MODE`:

| Mode | Read Quorum | Write Quorum | Description |
|---|---|---|---|
| `consistent` (default) | N/2 + 1 | N/2 + 1 | Full quorum for both reads and writes. Strongest consistency guarantees. |
| `degraded` | 1 | N/2 + 1 | Reads succeed from a single replica; writes still require quorum. Use when read availability is prioritized over strict consistency. |
| `dangerous` | 1 | 1 | Single-copy reads and writes. **Risk of data loss and stale reads.** Only use for non-critical data or testing. |

### Quorum Math

With replication factor **N = 3**:

| Operation | Formula | Value |
|---|---|---|
| Read quorum | N/2 + 1 | **2** |
| Write quorum | N/2 + 1 | **2** |

This means:

- **Writes** succeed when at least **2 of 3** replicas acknowledge the write.
- **Reads** succeed when at least **2 of 3** replicas return consistent data.
- The cluster tolerates **1** simultaneous node/drive failure without data loss or unavailability (in `consistent` mode).

With replication factor **N = 5**:

| Operation | Formula | Value |
|---|---|---|
| Read quorum | N/2 + 1 | **3** |
| Write quorum | N/2 + 1 | **3** |

> **NOTE:** Increasing the replication factor improves fault tolerance but increases storage usage linearly. A replication factor of 3 is recommended for most production deployments.

## Configuration Reference

| Environment Variable | Default | Description |
|---|---|---|
| `OBSTOR_ROOT_USER` | *(required)* | Admin access key. Must be identical across all nodes in the cluster. |
| `OBSTOR_ROOT_PASSWORD` | *(required)* | Admin secret key. Must be identical across all nodes in the cluster. |
| `OBSTOR_ZONE` | `""` | Zone label for replica placement (e.g. `us-east-1`). Controls geo-distribution of block replicas. |
| `OBSTOR_REPLICATION_FACTOR` | `3` | Number of copies stored for each block. Minimum: `1`, recommended: `3`. |
| `OBSTOR_BLOCK_SIZE` | `4M` | Size of each content-addressed block. Accepts `K`, `M`, `G` suffixes. Larger blocks reduce metadata overhead; smaller blocks improve deduplication. |
| `OBSTOR_NODE_CAPACITY` | `""` (unlimited) | Maximum storage capacity per node. Accepts `K`, `M`, `G`, `T`, `P` suffixes. When reached, new writes are routed to other nodes. |
| `OBSTOR_CONSISTENCY_MODE` | `consistent` | Consistency level: `consistent` (full quorum), `degraded` (relaxed reads), `dangerous` (single-copy I/O). |
| `OBSTOR_DOMAIN` | `""` | Domain name for virtual-host-style bucket access. |
| `OBSTOR_REGION_NAME` | `""` | Region label for the server location (e.g. `us-west-rack2`). |
