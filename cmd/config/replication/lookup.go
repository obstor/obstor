/*
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package replication

import (
	"fmt"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/obstor/obstor/cmd/config"
	"github.com/obstor/obstor/pkg/env"
)

const (
	ReplicationFactor = "replication_factor"
	Consistency       = "consistency_mode"
	BlockSize         = "block_size"
	Zone              = "zone"
	NodeCapacity      = "node_capacity"
	SyncQueueSize     = "sync_queue_size"
	SyncWorkers       = "sync_workers"
	MaxSyncRetries    = "max_sync_retries"

	EnvReplicationFactor = "OBSTOR_REPLICATION_FACTOR"
	EnvConsistencyMode   = "OBSTOR_CONSISTENCY_MODE"
	EnvBlockSize         = "OBSTOR_BLOCK_SIZE"
	EnvZone              = "OBSTOR_ZONE"
	EnvNodeCapacity      = "OBSTOR_NODE_CAPACITY"
	EnvSyncQueueSize     = "OBSTOR_SYNC_QUEUE_SIZE"
	EnvSyncWorkers       = "OBSTOR_SYNC_WORKERS"
	EnvMaxSyncRetries    = "OBSTOR_MAX_SYNC_RETRIES"
)

// DefaultKVS - default KV settings for replication.
var DefaultKVS = config.KVS{
	config.KV{Key: ReplicationFactor, Value: "3"},
	config.KV{Key: Consistency, Value: "consistent"},
	config.KV{Key: BlockSize, Value: "1048576"},
	config.KV{Key: Zone, Value: ""},
	config.KV{Key: NodeCapacity, Value: ""},
	config.KV{Key: SyncQueueSize, Value: "100000"},
	config.KV{Key: SyncWorkers, Value: "4"},
	config.KV{Key: MaxSyncRetries, Value: "10"},
}

// LookupConfig builds a Config from KVS + env overrides.
func LookupConfig(kvs config.KVS) (cfg Config, err error) {
	if err = config.CheckValidKeys(config.ReplicationSubSys, kvs, DefaultKVS); err != nil {
		return cfg, err
	}

	rfStr := env.Get(EnvReplicationFactor, kvs.Get(ReplicationFactor))
	if rfStr == "" {
		rfStr = "3"
	}
	cfg.ReplicationFactor, err = strconv.Atoi(rfStr)
	if err != nil {
		return cfg, fmt.Errorf("replication: replication_factor invalid: %w", err)
	}
	if cfg.ReplicationFactor < 1 {
		return cfg, fmt.Errorf("replication: replication_factor must be >= 1, got %d", cfg.ReplicationFactor)
	}

	consistencyStr := env.Get(EnvConsistencyMode, kvs.Get(Consistency))
	if consistencyStr == "" {
		consistencyStr = "consistent"
	}
	switch ConsistencyMode(consistencyStr) {
	case ConsistencyModeConsistent, ConsistencyModeDegraded, ConsistencyModeDangerous:
		cfg.Consistency = ConsistencyMode(consistencyStr)
	default:
		return cfg, fmt.Errorf("replication: consistency_mode must be consistent|degraded|dangerous, got %q", consistencyStr)
	}

	blockSizeStr := env.Get(EnvBlockSize, kvs.Get(BlockSize))
	if blockSizeStr == "" {
		blockSizeStr = "1048576"
	}
	blockBytes, parseErr := humanize.ParseBytes(blockSizeStr)
	if parseErr != nil {
		bs, err2 := strconv.ParseInt(blockSizeStr, 10, 64)
		if err2 != nil {
			return cfg, fmt.Errorf("replication: block_size invalid: %w", parseErr)
		}
		blockBytes = uint64(bs)
	}
	cfg.BlockSize = int64(blockBytes)
	if cfg.BlockSize < 64*1024 {
		return cfg, fmt.Errorf("replication: block_size must be >= 64KiB, got %d", cfg.BlockSize)
	}
	if cfg.BlockSize > 64*1024*1024 {
		return cfg, fmt.Errorf("replication: block_size must be <= 64MiB, got %d", cfg.BlockSize)
	}

	cfg.Zone = env.Get(EnvZone, kvs.Get(Zone))

	capStr := env.Get(EnvNodeCapacity, kvs.Get(NodeCapacity))
	if capStr != "" {
		capBytes, err := humanize.ParseBytes(capStr)
		if err != nil {
			return cfg, fmt.Errorf("replication: node_capacity invalid: %w", err)
		}
		cfg.NodeCapacity = int64(capBytes)
	}

	queueStr := env.Get(EnvSyncQueueSize, kvs.Get(SyncQueueSize))
	if queueStr == "" {
		queueStr = "100000"
	}
	cfg.SyncQueueSize, err = strconv.Atoi(queueStr)
	if err != nil || cfg.SyncQueueSize < 1 {
		return cfg, fmt.Errorf("replication: sync_queue_size must be positive, got %q", queueStr)
	}

	workersStr := env.Get(EnvSyncWorkers, kvs.Get(SyncWorkers))
	if workersStr == "" {
		workersStr = "4"
	}
	cfg.SyncWorkers, err = strconv.Atoi(workersStr)
	if err != nil || cfg.SyncWorkers < 1 {
		return cfg, fmt.Errorf("replication: sync_workers must be positive, got %q", workersStr)
	}

	retriesStr := env.Get(EnvMaxSyncRetries, kvs.Get(MaxSyncRetries))
	if retriesStr == "" {
		retriesStr = "10"
	}
	cfg.MaxSyncRetries, err = strconv.Atoi(retriesStr)
	if err != nil || cfg.MaxSyncRetries < 0 {
		return cfg, fmt.Errorf("replication: max_sync_retries must be >= 0, got %q", retriesStr)
	}

	return cfg, nil
}
