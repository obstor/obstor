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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"

	"github.com/obstor/obstor/cmd/logger"
)

// Content-addressed block reference.
type BlockRef struct {
	Hash  string `json:"h" msg:"h"` // SHA-256 hex digest
	Size  int64  `json:"s" msg:"s"` // block data length
	Index int    `json:"i" msg:"i"` // ordinal position in object
}

// Chunk data into content-addressed blocks and replicates N copies.
type BlockReplicator struct {
	blockSize    int64
	replicaCount int
}

// Block replicator.
func NewBlockReplicator(blockSize int64, replicaCount int) *BlockReplicator {
	return &BlockReplicator{
		blockSize:    blockSize,
		replicaCount: replicaCount,
	}
}

// Split src into blocks, returning refs and raw data.
func (br *BlockReplicator) ChunkAndHash(src io.Reader) ([]BlockRef, [][]byte, error) {
	var refs []BlockRef
	var blocks [][]byte
	buf := make([]byte, br.blockSize)
	idx := 0

	for {
		n, err := io.ReadFull(src, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, nil, err
		}
		if n == 0 {
			break
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		hash := sha256.Sum256(data)
		hashHex := hex.EncodeToString(hash[:])

		refs = append(refs, BlockRef{
			Hash:  hashHex,
			Size:  int64(n),
			Index: idx,
		})
		blocks = append(blocks, data)
		idx++

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}

	return refs, blocks, nil
}

// Return blocks needed for totalSize.
func (br *BlockReplicator) BlockCount(totalSize int64) int {
	if totalSize <= 0 {
		return 0
	}
	n := totalSize / br.blockSize
	if totalSize%br.blockSize != 0 {
		n++
	}
	return int(n)
}

// Return block indices overlapping [offset, offset+length).
func (br *BlockReplicator) BlockRange(offset, length, totalSize int64) (startBlock, endBlock int) {
	if length <= 0 || totalSize <= 0 {
		return 0, 0
	}
	startBlock = int(offset / br.blockSize)
	end := offset + length - 1
	if end >= totalSize {
		end = totalSize - 1
	}
	endBlock = int(end / br.blockSize)
	return startBlock, endBlock
}

// Return the two-level on-disk path for a block hash.
func isValidBlockHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for i := 0; i < len(hash); i++ {
		c := hash[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func blockStoragePath(hash string) string {
	if len(hash) < 4 {
		return path.Join("blocks", hash)
	}
	return path.Join("blocks", hash[:2], hash[2:4], hash)
}

// Return the SHA-256 hex digest of data.
func hashBlockData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Check data against an expected hash.
func verifyBlockHash(data []byte, expectedHash string) error {
	actual := hashBlockData(data)
	if actual != expectedHash {
		return fmt.Errorf("block hash mismatch: expected %s, got %s", expectedHash, actual)
	}
	return nil
}

// Copy a block from any source to targets missing it.
func (br *BlockReplicator) HealBlock(ctx context.Context, blockHash string, sources []StorageAPI, targets []StorageAPI) (int, error) {
	data, err := br.ReadBlock(ctx, blockHash, sources)
	if err != nil {
		return 0, err
	}

	healed := 0
	for _, disk := range targets {
		if disk == nil {
			continue
		}
		if err := disk.WriteBlock(ctx, blockHash, data); err != nil {
			logger.LogIf(ctx, err)
			continue
		}
		healed++
	}
	if healed == 0 {
		attempted := 0
		for _, d := range targets {
			if d != nil {
				attempted++
			}
		}
		if attempted > 0 {
			return 0, fmt.Errorf("heal block %s: all %d target writes failed", blockHash, attempted)
		}
	}
	return healed, nil
}

// Write data blocks to dst.
func writeDataBlocks(ctx context.Context, dst io.Writer, enBlocks [][]byte, dataBlocks int, offset int64, length int64) (int64, error) {
	if offset < 0 || length < 0 {
		logger.LogIf(ctx, errUnexpected)
		return 0, errUnexpected
	}
	if len(enBlocks) < dataBlocks {
		return 0, fmt.Errorf("too few blocks: have %d, need %d", len(enBlocks), dataBlocks)
	}

	write := length
	var totalWritten int64
	for _, block := range enBlocks[:dataBlocks] {
		if offset >= int64(len(block)) {
			offset -= int64(len(block))
			continue
		} else {
			block = block[offset:]
			offset = 0
		}
		if write < int64(len(block)) {
			n, err := io.Copy(dst, bytes.NewReader(block[:write]))
			if err != nil {
				return 0, err
			}
			totalWritten += n
			break
		}
		n, err := io.Copy(dst, bytes.NewReader(block))
		if err != nil {
			return 0, err
		}
		write -= n
		totalWritten += n
	}
	return totalWritten, nil
}
