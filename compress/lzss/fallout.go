package lzss

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	FalloutCompressStore uint32 = 0x10 // needs reasearch, see fallout1-re/src/plib/db/db.c, de.flags 16
	FalloutCompressNone  uint32 = 0x20 // as-is, SizePacked is 0
	FalloutCompressLZSS  uint32 = 0x40
)

type FalloutLZSS struct {
	LZSS
}

// FalloutFileLZSS holds minimal required data about DAT1 file entry
type FalloutFileLZSS struct {
	CompressMode uint32
	Offset       uint32
	SizeReal     uint32
	SizePacked   uint32
}

// Fallout1 is a small wrapper around LZSS, which helps reading data stored in .dat files
var Fallout1 = FalloutLZSS{
	LZSS{
		DictionarySize: 4096,
		MinMatch:       3,
		MaxMatch:       18,
	},
}

// Decompress prepares reader
//
// `stream` position must be set to data offset before calling this functions
func (fallout FalloutLZSS) Decompress(stream io.Reader, unpackedSize uint32) ([]byte, error) {
	const self = "FalloutLZSS.Decompress()"

	var packedSize int16

	// Fallout 1 data block starts with 2 bytes describing block size and  is stored as-is or compressed,
	// and compressed block size.

	if errBlock := binary.Read(stream, binary.BigEndian, &packedSize); errBlock != nil {
		return nil, fmt.Errorf("%s cannot read block size: %w", self, errBlock)
	}

	if packedSize < 0 {
		fmt.Printf("%s packedSize = %d\n", self, packedSize)

		// If block size is negative, file is stored without compression

		if uint32(-packedSize) != unpackedSize-2 {
			return nil, fmt.Errorf("%s size mismatch %d != %d", self, -packedSize, unpackedSize-2)
		}

		// Simply copy all bytes as-is, without without involving LZSS in the process
		var output = make([]byte, unpackedSize)
		if _, err := stream.Read(output); err != nil {
			return nil, fmt.Errorf("%s cannot read uncompressed block: %w", self, err)
		}

		return output, nil
	} else if packedSize == 0 {
		// If block size is 0, something funky happened, most likely
		return nil, fmt.Errorf("%s block size == 0", self)
	}

	// Call
	return fallout.LZSS.Decompress(stream, uint32(packedSize), unpackedSize)
}

func (fallout FalloutLZSS) DecompressFile(stream io.ReadSeeker, file FalloutFileLZSS) (output []byte, err error) {
	const self = "FalloutLZSS.DecompressFile()"

	// Try to detect unexpected EOF before decompression

	if _, err = stream.Seek(int64((file.Offset + file.SizePacked)), io.SeekStart); err != nil {
		return nil, err
	}

	if _, err = stream.Seek(int64(file.Offset), io.SeekStart); err != nil {
		return nil, err
	}

	if file.CompressMode == FalloutCompressNone {
		// Simply copy all bytes as-is, without without involving LZSS in the process

		var output = make([]byte, file.SizeReal)
		if _, err = stream.Read(output); err != nil {
			return nil, fmt.Errorf("%s cannot read uncompressed block: %w", self, err)
		}

		return output, nil
	}

	return fallout.Decompress(stream, file.SizeReal)
}
