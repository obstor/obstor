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
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/obstor/obstor/cmd/logger"
)

// Close a reader if it has io.Closer.
func releaseReader(r io.ReaderAt) {
	if r == nil {
		return
	}
	if rc, ok := r.(io.Closer); ok {
		_ = rc.Close()
	}
}

// Reads in parallel from readers.
type parallelReader struct {
	readers       []io.ReaderAt
	orgReaders    []io.ReaderAt
	dataBlocks    int
	errs          []error
	offset        int64
	shardSize     int64
	shardFileSize int64
	buf           [][]byte
	readerToBuf   []int
}

// newParallelReader returns parallelReader.
func newParallelReader(readers []io.ReaderAt, e Erasure, offset, totalLength int64) *parallelReader {
	r2b := make([]int, len(readers))
	for i := range r2b {
		r2b[i] = i
	}
	return &parallelReader{
		readers:       readers,
		orgReaders:    readers,
		errs:          make([]error, len(readers)),
		dataBlocks:    e.dataBlocks,
		offset:        (offset / e.blockSize) * e.ShardSize(),
		shardSize:     e.ShardSize(),
		shardFileSize: e.ShardFileSize(totalLength),
		buf:           make([][]byte, len(readers)),
		readerToBuf:   r2b,
	}
}

// Free buffers held by the parallel reader
func (p *parallelReader) Release() {
	for i := range p.buf {
		p.buf[i] = nil
	}
	p.buf = nil
}

// preferReaders can mark readers as preferred.
// These will be chosen before others.
func (p *parallelReader) preferReaders(prefer []bool) {
	if len(prefer) != len(p.orgReaders) {
		return
	}
	tmp := make([]io.ReaderAt, len(p.orgReaders))
	copy(tmp, p.orgReaders)
	p.readers = tmp
	next := 0
	for i, ok := range prefer {
		if !ok || p.readers[i] == nil {
			continue
		}
		if i == next {
			next++
			continue
		}
		p.readers[next], p.readers[i] = p.readers[i], p.readers[next]
		p.readerToBuf[next] = i
		p.readerToBuf[i] = next
		next++
	}
}

// Returns if buf can be erasure decoded.
func (p *parallelReader) canDecode(buf [][]byte) bool {
	bufCount := 0
	for _, b := range buf {
		if len(b) > 0 {
			bufCount++
		}
	}
	return bufCount >= p.dataBlocks
}

// Returns p.dataBlocks number of bufs
func (p *parallelReader) Read(dst [][]byte) ([][]byte, error) {
	newBuf := dst
	if len(dst) != len(p.readers) {
		newBuf = make([][]byte, len(p.readers))
	} else {
		for i := range newBuf {
			newBuf[i] = newBuf[i][:0]
		}
	}
	var newBufLK sync.RWMutex

	if p.offset+p.shardSize > p.shardFileSize {
		p.shardSize = p.shardFileSize - p.offset
	}
	if p.shardSize == 0 {
		return newBuf, nil
	}

	var (
		bitrotHeal       int32
		missingPartsHeal int32
		offlineCount     int32
		successCount     int32
		wg               sync.WaitGroup
		spawnMu          sync.Mutex
	)

	// Initial batch of readers
	nextReader := 0
	var spawnOne func() bool
	spawnOne = func() bool {
		spawnMu.Lock()
		defer spawnMu.Unlock()
		for nextReader < len(p.readers) {
			i := nextReader
			nextReader++
			rr := p.readers[i]
			if rr == nil {
				atomic.AddInt32(&offlineCount, 1)
				continue
			}
			wg.Add(1)
			go func(idx int, reader io.ReaderAt) {
				defer wg.Done()
				bufIdx := p.readerToBuf[idx]
				if p.buf[bufIdx] == nil {
					p.buf[bufIdx] = make([]byte, p.shardSize)
				}
				p.buf[bufIdx] = p.buf[bufIdx][:p.shardSize]
				n, err := reader.ReadAt(p.buf[bufIdx], p.offset)
				if err != nil {
					if errors.Is(err, errFileNotFound) {
						atomic.StoreInt32(&missingPartsHeal, 1)
					} else if errors.Is(err, errFileCorrupt) {
						atomic.StoreInt32(&bitrotHeal, 1)
					}
					releaseReader(p.orgReaders[bufIdx])
					p.orgReaders[bufIdx] = nil
					p.readers[idx] = nil
					p.errs[idx] = err
					// Try to spawn a replacement reader
					spawnOne()
					return
				}
				newBufLK.Lock()
				newBuf[bufIdx] = p.buf[bufIdx][:n]
				newBufLK.Unlock()
				atomic.AddInt32(&successCount, 1)
			}(i, rr)
			return true
		}
		return false
	}

	// Launch initial readers
	for launched := 0; launched < p.dataBlocks; launched++ {
		if !spawnOne() {
			break
		}
	}
	wg.Wait()

	if p.canDecode(newBuf) {
		p.offset += p.shardSize
		if atomic.LoadInt32(&missingPartsHeal) == 1 {
			return newBuf, errFileNotFound
		} else if atomic.LoadInt32(&bitrotHeal) == 1 {
			return newBuf, errFileCorrupt
		}
		return newBuf, nil
	}

	return nil, reduceReadQuorumErrs(context.Background(), p.errs, objectOpIgnoredErrs, p.dataBlocks)
}

// Decode reads from readers, reconstructs data if needed and writes the data to the writer.
func (e Erasure) Decode(ctx context.Context, writer io.Writer, readers []io.ReaderAt, offset, length, totalLength int64, prefer []bool) (written int64, derr error) {
	if offset < 0 || length < 0 {
		logger.LogIf(ctx, errInvalidArgument)
		return -1, errInvalidArgument
	}
	if offset+length > totalLength {
		logger.LogIf(ctx, errInvalidArgument)
		return -1, errInvalidArgument
	}

	if length == 0 {
		return 0, nil
	}

	reader := newParallelReader(readers, e, offset, totalLength)
	defer reader.Release()
	if len(prefer) == len(readers) {
		reader.preferReaders(prefer)
	}

	startBlock := offset / e.blockSize
	endBlock := (offset + length) / e.blockSize

	var bytesWritten int64
	var bufs [][]byte
	for block := startBlock; block <= endBlock; block++ {
		var blockOffset, blockLength int64
		switch {
		case startBlock == endBlock:
			blockOffset = offset % e.blockSize
			blockLength = length
		case block == startBlock:
			blockOffset = offset % e.blockSize
			blockLength = e.blockSize - blockOffset
		case block == endBlock:
			blockOffset = 0
			blockLength = (offset + length) % e.blockSize
		default:
			blockOffset = 0
			blockLength = e.blockSize
		}
		if blockLength == 0 {
			break
		}

		var err error
		bufs, err = reader.Read(bufs)
		if len(bufs) > 0 {
			if errors.Is(err, errFileNotFound) || errors.Is(err, errFileCorrupt) {
				if derr == nil {
					derr = err
				}
			}
		} else if err != nil {
			return -1, err
		}

		if err = e.DecodeDataBlocks(bufs); err != nil {
			logger.LogIf(ctx, err)
			return -1, err
		}

		n, err := writeDataBlocks(ctx, writer, bufs, e.dataBlocks, blockOffset, blockLength)
		if err != nil {
			return -1, err
		}

		bytesWritten += n
	}

	if bytesWritten != length {
		logger.LogIf(ctx, errLessData)
		return bytesWritten, errLessData
	}

	return bytesWritten, derr
}

// ReadBlock reads a block from the first available source node.
// It tries each source in order (prefer local disks first) and
// returns on the first successful read. No reconstruction is needed
// since every copy is identical.
func (br *BlockReplicator) ReadBlock(ctx context.Context, blockHash string, sources []StorageAPI) ([]byte, error) {
	var lastErr error
	for _, disk := range sources {
		if disk == nil {
			continue
		}
		data, err := disk.ReadBlock(ctx, blockHash)
		if err != nil {
			lastErr = err
			continue
		}
		// Verify integrity.
		if err := verifyBlockHash(data, blockHash); err != nil {
			lastErr = err
			continue
		}
		return data, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("block %s: no available source disks", blockHash)
}

// HasBlock checks if at least one source has the block.
func (br *BlockReplicator) HasBlock(ctx context.Context, blockHash string, sources []StorageAPI) bool {
	for _, disk := range sources {
		if disk == nil {
			continue
		}
		has, err := disk.HasBlock(ctx, blockHash)
		if err == nil && has {
			return true
		}
	}
	return false
}

// CountBlockCopies returns how many sources have the block.
func (br *BlockReplicator) CountBlockCopies(ctx context.Context, blockHash string, sources []StorageAPI) int {
	count := 0
	for _, disk := range sources {
		if disk == nil {
			continue
		}
		has, err := disk.HasBlock(ctx, blockHash)
		if err == nil && has {
			count++
		}
	}
	return count
}
