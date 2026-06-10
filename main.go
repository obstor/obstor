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

/*
 * Below main package has canonical imports for 'go get' and 'go build'
 * to work with all other clones of github.com/obstor/obstor repository. For
 * more information refer https://golang.org/doc/go1.4#canonicalimports
 */

package main // import "github.com/obstor/obstor"

import (
	"os"

	obstor "github.com/obstor/obstor/cmd"

	// Import backends
	_ "github.com/obstor/obstor/cmd/backend"

	// Import protocols
	_ "github.com/obstor/obstor/cmd/protocols/s3"
	_ "github.com/obstor/obstor/cmd/protocols/sftp"
)

func main() {
	obstor.Main(os.Args)
}
