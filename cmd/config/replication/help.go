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

import "github.com/obstor/obstor/cmd/config"

// Help template for replication configuration.
var Help = config.HelpKVS{
	config.HelpKV{
		Key:         ReplicationFactor,
		Description: `number of copies to maintain across zones (default: 3)`,
		Optional:    true,
		Type:        "number",
	},
	config.HelpKV{
		Key:         Consistency,
		Description: `consistency mode: "consistent", "degraded", or "dangerous" (default: consistent)`,
		Optional:    true,
		Type:        "string",
	},
	config.HelpKV{
		Key:         BlockSize,
		Description: `data block size for chunked storage, e.g. "1M", "10M" (default: 1M)`,
		Optional:    true,
		Type:        "string",
	},
	config.HelpKV{
		Key:         Zone,
		Description: `availability zone name for this node, e.g. "us-east-1a"`,
		Optional:    true,
		Type:        "string",
	},
	config.HelpKV{
		Key:         NodeCapacity,
		Description: `advertised storage capacity for this node, e.g. "1T", "500G"`,
		Optional:    true,
		Type:        "string",
	},
	config.HelpKV{
		Key:         SyncQueueSize,
		Description: `max pending block sync operations (default: 100000)`,
		Optional:    true,
		Type:        "number",
	},
	config.HelpKV{
		Key:         SyncWorkers,
		Description: `number of parallel block sync goroutines (default: 4)`,
		Optional:    true,
		Type:        "number",
	},
	config.HelpKV{
		Key:         MaxSyncRetries,
		Description: `max retries for a failed block sync before dropping (default: 10)`,
		Optional:    true,
		Type:        "number",
	},
	config.HelpKV{
		Key:         config.Comment,
		Description: config.DefaultComment,
		Optional:    true,
		Type:        "sentence",
	},
}
