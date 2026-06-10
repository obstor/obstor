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

package cmd

import (
	"context"
	"io"
	"sync"

	"github.com/obstor/obstor/cmd/logger"
)

// Write erasure-coded data to multiple disks in parallel.
type parallelWriter struct {
	writers     []io.Writer
	writeQuorum int
	errs        []error
}

// Writes data to writers in parallel.
func (p *parallelWriter) Write(ctx context.Context, blocks [][]byte) error {
	var wg sync.WaitGroup

	for i := range p.writers {
		if p.writers[i] == nil {
			p.errs[i] = errDiskNotFound
			continue
		}
		if p.errs[i] != nil {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, p.errs[i] = p.writers[i].Write(blocks[i])
			if p.errs[i] != nil {
				if wc, ok := p.writers[i].(io.Closer); ok {
					_ = wc.Close()
				}
				p.writers[i] = nil
			}
		}(i)
	}
	wg.Wait()

	return reduceWriteQuorumErrs(ctx, p.errs, objectOpIgnoredErrs, p.writeQuorum)
}

// Read from src, erasure encode and write to writers
func (e *Erasure) Encode(ctx context.Context, src io.Reader, writers []io.Writer, buf []byte, quorum int) (total int64, err error) {
	writer := &parallelWriter{
		writers:     writers,
		writeQuorum: quorum,
		errs:        make([]error, len(writers)),
	}

	for {
		var blocks [][]byte
		n, rerr := io.ReadFull(src, buf)
		if n > 0 {
			blocks, err = e.EncodeData(ctx, buf[:n])
			if err != nil {
				logger.LogIf(ctx, err)
				return 0, err
			}

			if err = writer.Write(ctx, blocks); err != nil {
				logger.LogIf(ctx, err)
				return 0, err
			}

			total += int64(n)
		}
		if rerr == io.EOF || rerr == io.ErrUnexpectedEOF {
			break
		}
		if rerr != nil {
			logger.LogIf(ctx, rerr)
			return 0, rerr
		}
	}
	return total, nil
}

// WriteBlock writes the same block data to multiple target storage nodes
// in parallel. It returns nil if at least writeQuorum targets succeed.
func (br *BlockReplicator) WriteBlock(ctx context.Context, blockHash string, data []byte, targets []StorageAPI, writeQuorum int) error {
	errs := make([]error, len(targets))
	var wg sync.WaitGroup

	for i, disk := range targets {
		if disk == nil {
			errs[i] = errDiskNotFound
			continue
		}
		wg.Add(1)
		go func(i int, disk StorageAPI) {
			defer wg.Done()
			errs[i] = disk.WriteBlock(ctx, blockHash, data)
		}(i, disk)
	}
	wg.Wait()

	var successes int
	for _, err := range errs {
		if err == nil {
			successes++
		}
	}
	if successes >= writeQuorum {
		return nil
	}
	return reduceWriteQuorumErrs(ctx, errs, objectOpIgnoredErrs, writeQuorum)
}
