//go:build !windows
// +build !windows

/*
 * MinIO Cloud Storage, (C) 2019-2020 MinIO, Inc.
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

package cmd

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/obstor/obstor/pkg/atime"
)

// Return error if Atime is disabled on the O/S
func checkAtimeSupport(dir string) (err error) {
	file, err := os.CreateTemp(dir, "prefix")
	if err != nil {
		return
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer file.Close()
	finfo1, err := os.Stat(file.Name())
	if err != nil {
		return
	}
	// Add a sleep to ensure atime change is detected
	time.Sleep(10 * time.Millisecond)

	if _, err = io.Copy(io.Discard, file); err != nil {
		return
	}

	finfo2, err := os.Stat(file.Name())

	if atime.Get(finfo2).Equal(atime.Get(finfo1)) {
		return errors.New("atime not supported")
	}
	return
}
