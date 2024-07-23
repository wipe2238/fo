package lzss

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

const (
	FalloutCompressStore uint32 = 0x10 // needs research, see fallout1-re/src/plib/db/db.c, de.flags 16
	FalloutCompressNone  uint32 = 0x20 // as-is, DAT1 sets SizePacked to 0
	FalloutCompressLZSS  uint32 = 0x40
)

// FalloutFile is a small wrapper around LZSS holding minimal required data about DAT1 file entry
type FalloutFile struct {
	// Stream containing compressed file data
	//
	// All `FalloutFile` functions keep `Stream` position unchanged,
	// as long there were no errors
	Stream io.ReadSeeker

	// File data offset;
	// same as in DAT1 file entry
	Offset int64

	// Size of compressed data in bytes
	//
	// DAT1 file entry sets it to 0 for uncompressed files;
	// in that case, uncompressed size should be used instead
	SizePacked int64

	// Compression mode
	//
	// Must be set to one of `FalloutCompress*` constants;
	// same as in DAT1 file entry
	CompressMode uint32
}

// Decompress returns slice of bytes
//
// Note that this function can handle uncompressed data
func (file FalloutFile) Decompress() (bytes []byte, err error) {
	var streamPos int64

	// store current position, will be used when function finishes without errors
	if streamPos, err = file.Stream.Seek(0, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("%s cannot store stream position", errPackage)
	}

	// try to detect unexpected EOF before decompression

	if _, err = file.Stream.Seek((file.Offset + file.SizePacked), io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s eof detection (1/2): %w", errPackage, err)
	}

	if _, err = file.Stream.Seek(file.Offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s eof detection (2/2): %w", errPackage, err)
	}

	// all cases must either fill `bytes` without returning, or return error
	switch file.CompressMode {
	case FalloutCompressStore:
		return nil, fmt.Errorf("%s FalloutCompressStore", errPackage)

	case FalloutCompressNone:
		// Uncompressed file
		// Simply copy all bytes as-is, without involving LZSS in the process

		bytes = make([]byte, file.SizePacked)
		if _, err = file.Stream.Read(bytes); err != nil {
			return nil, fmt.Errorf("%s cannot read uncompressed file: %w", errPackage, err)
		}

	case FalloutCompressLZSS:
		var blocks []int64

		// TODO(?) add optional `Blocks` to `FalloutFile`
		if blocks, err = file.ReadBlocksSize(); err != nil {
			return nil, err
		} else if len(blocks) < 1 {
			return nil, fmt.Errorf("%s error reading blocks size", errPackage)
		}

		for _, sizeBlock := range blocks {
			var bytesBlock []byte

			// Skip already known block size
			if _, err = file.Stream.Seek(2 /* sizeof(int16) */, io.SeekCurrent); err != nil {
				return nil, err
			}

			if sizeBlock < 0 {
				// Uncompressed block
				// Simply copy all bytes as-is, without involving LZSS in the process

				bytesBlock = make([]byte, -sizeBlock)
				if _, err = file.Stream.Read(bytesBlock); err != nil {
					return nil, err
				}
			} else if sizeBlock > 0 {
				// Compressed block
				// LZSS configuration is hidden from user for simplicity; there's hardly any reason
				// to change it, except intentionally breaking whole point of parent struct

				var dat1 = LZSS{DictionarySize: 4096, MinMatch: 3, MaxMatch: 18}

				if bytesBlock, err = dat1.Decompress(file.Stream, sizeBlock); err != nil {
					return nil, err
				}
			} else {
				// None of MASTER.DAT / CRITTER.DAT / FALLDEMO.DAT contains file with sizeBlock=0
				if true {
					panic("sizeBlock == 0")
				}
				break
			}

			bytes = append(bytes, bytesBlock...)
		}

	default:
		return nil, fmt.Errorf("%s unknown compress mode 0x%X = %d", errPackage, file.CompressMode, file.CompressMode)
	}

	// Restore previous position
	if _, err = file.Stream.Seek(streamPos, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s cannot restore stream position", errPackage)
	}

	return bytes, nil
}

// ReadBlocks returns compressed file split into block defined by DAT1
func (file FalloutFile) ReadBlocks() (blocksBytes [][]byte, err error) {
	// Note that this function is mostly for convenience, when user wants detailed info about files
	// It could easily be removed, as DecompressFile() doesn't use it at all

	var streamPos int64

	// Store current position, will be used when function finishes without errors
	if streamPos, err = file.Stream.Seek(0, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("%s cannot store stream position", errPackage)
	}

	if _, err = file.Stream.Seek(file.Offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s cannot set stream position", errPackage)
	}

	var blocks []int64
	if blocks, err = file.ReadBlocksSize(); err != nil {
		return nil, err
	}

	blocksBytes = make([][]byte, len(blocks))

	for idx, blockSize := range blocks {
		// Skip already known block size
		if _, err = file.Stream.Seek(2 /* sizeof(int16) */, io.SeekCurrent); err != nil {
			return nil, err
		}

		blocksBytes[idx] = make([]byte, blockSize)

		if _, err = file.Stream.Read(blocksBytes[idx]); err != nil {
			return nil, err
		}
	}

	// Restore previous position
	if _, err = file.Stream.Seek(streamPos, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s cannot restore stream position", errPackage)
	}

	return blocksBytes, nil
}

// ReadBlocksSize returns a slice containing compressed file blocks sizes
func (file FalloutFile) ReadBlocksSize() (blocks []int64, err error) {
	// Note that this function is mostly for convenience, when user wants detailed info about files
	// It could easily be merged with DecompressFile() and simplified

	var streamPos int64

	// Store current position, will be used when function finishes without errors
	if streamPos, err = file.Stream.Seek(0, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("%s cannot store stream position", errPackage)
	}

	blocks = make([]int64, 0)

	var sizePacked = file.SizePacked
	for sizePacked > 0 {
		// Compressed Fallout1 file is split into one or more blocks (chunks/parts/...)
		// Each blocks starts with int16 describing it's size; if the value is negative,
		// block is not compressed

		var sizeBlock int16
		if err = binary.Read(file.Stream, binary.BigEndian, &sizeBlock); err == nil {
			sizePacked -= 2 /* sizeof(int16) */
		} else {
			return nil, fmt.Errorf("%s cannot read block size", errPackage)
		}

		// Last block's size CAN and WILL be incorrect for many multi-block files (such as .ACM),
		// and must be clamped aganst remaining data size before decompression/copying starts
		// Difference can vary from couple 100s bytes to ~30 kilobytes

		var sizeBlockReal = min(int64(math.Abs(float64(sizeBlock))), sizePacked)

		// Skip block data
		if _, err = file.Stream.Seek(sizeBlockReal, io.SeekCurrent); err == nil {
			sizePacked -= sizeBlockReal
		} else {
			return nil, err
		}

		// Restore sign before appending, if needed
		if sizeBlock < 0 {
			sizeBlockReal = -sizeBlockReal
		}

		blocks = append(blocks, sizeBlockReal)
	}

	if sizePacked != 0 {
		// JIC, should never happen
		return nil, fmt.Errorf("%s error reading blocks size: sizePacked(%d) != 0, blocks = %v, streamPos = %d", errPackage, sizePacked, blocks, streamPos)
	}

	// Restore previous position
	if _, err = file.Stream.Seek(streamPos, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s cannot restore stream position", errPackage)
	}

	return blocks, nil
}
