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

import "github.com/obstor/obstor/cmd/config"

// Help template for SFTP feature.
var Help = config.HelpKVS{
	config.HelpKV{
		Key:         config.Enable,
		Description: `enable or disable SFTP server, defaults to "off"`,
		Type:        "on|off",
	},
	config.HelpKV{
		Key:         Address,
		Description: `SFTP server listen address e.g. ":9002"`,
		Optional:    true,
		Type:        "address",
	},
	config.HelpKV{
		Key:         HostKey,
		Description: `path to SSH host private key file, auto-generated if empty`,
		Optional:    true,
		Type:        "path",
	},
	config.HelpKV{
		Key:         config.Comment,
		Description: config.DefaultComment,
		Optional:    true,
		Type:        "sentence",
	},
}
