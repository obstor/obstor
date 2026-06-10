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

package sftp

import (
	"github.com/obstor/obstor/cmd/config"
	"github.com/obstor/obstor/pkg/env"
)

const (
	Address = "address"
	HostKey = "host_key"

	EnvSFTPAddress = "OBSTOR_SFTP_ADDRESS"
	EnvSFTPHostKey = "OBSTOR_SFTP_HOST_KEY"
	EnvSFTPEnable  = "OBSTOR_SFTP_ENABLE"

	DefaultAddress = ":9002"
)

// DefaultKVS - default KV settings for SFTP.
var DefaultKVS = config.KVS{
	config.KV{
		Key:   config.Enable,
		Value: config.EnableOff,
	},
	config.KV{
		Key:   Address,
		Value: DefaultAddress,
	},
	config.KV{
		Key:   HostKey,
		Value: "",
	},
}

// Enabled returns if SFTP is enabled.
func Enabled(kvs config.KVS) bool {
	enabled := env.Get(EnvSFTPEnable, kvs.Get(config.Enable))
	return enabled == config.EnableOn
}

// LookupConfig - extracts SFTP configuration.
func LookupConfig(kvs config.KVS) (Config, error) {
	cfg := Config{}
	if err := config.CheckValidKeys(config.SftpSubSys, kvs, DefaultKVS); err != nil {
		return cfg, err
	}

	enabled := env.Get(EnvSFTPEnable, kvs.Get(config.Enable))
	if enabled != config.EnableOn {
		return cfg, nil
	}

	cfg.Enabled = true
	cfg.Address = env.Get(EnvSFTPAddress, kvs.Get(Address))
	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}

	cfg.HostKeyPath = env.Get(EnvSFTPHostKey, kvs.Get(HostKey))

	return cfg, nil
}
