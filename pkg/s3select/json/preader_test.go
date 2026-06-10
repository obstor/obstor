/*
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
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

package json

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/obstor/obstor/pkg/s3select/sql"
)

func TestNewPReader(t *testing.T) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		name := filepath.Base(file.Name())
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			r := NewPReader(f, &ReaderArgs{})
			var record sql.Record
			for {
				record, err = r.Read(record)
				if err != nil {
					break
				}
			}
			_ = r.Close()
			if err != io.EOF {
				t.Fatalf("Reading failed with %s, %s", err, name)
			}
		})

		t.Run(name+"-close", func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			r := NewPReader(f, &ReaderArgs{})
			_ = r.Close()
			var record sql.Record
			for {
				record, err = r.Read(record)
				if err != nil {
					break
				}
			}
			if err != io.EOF {
				t.Fatalf("Reading failed with %s, %s", err, name)
			}
		})
	}
}

func BenchmarkPReader(b *testing.B) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		b.Fatal(err)
	}
	for _, file := range files {
		name := filepath.Base(file.Name())
		b.Run(name, func(b *testing.B) {
			f, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				b.Fatal(err)
			}
			b.SetBytes(int64(len(f)))
			b.ReportAllocs()
			b.ResetTimer()
			var record sql.Record
			for i := 0; i < b.N; i++ {
				r := NewPReader(io.NopCloser(bytes.NewBuffer(f)), &ReaderArgs{})
				for {
					record, err = r.Read(record)
					if err != nil {
						break
					}
				}
				_ = r.Close()
				if err != io.EOF {
					b.Fatalf("Reading failed with %s, %s", err, name)
				}
			}
		})
	}
}
