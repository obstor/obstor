/*
 * MinIO Cloud Storage, (C) 2016-2020 MinIO, Inc.
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

package s3

import obstor "github.com/obstor/obstor/cmd"

// Handlers embeds obstor.ObjectAPIHandlers to inherit all S3 handler methods.
// The handler methods (GetObjectHandler, PutObjectHandler, etc.) are defined
// on ObjectAPIHandlers in the cmd package and are automatically available
// on Handlers via embedding.
type Handlers struct {
	obstor.ObjectAPIHandlers
}
