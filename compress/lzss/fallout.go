package lzss

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/wipe2238/fo/dat"
)

const (
	FalloutCompressZero = uint8(iota)
	FalloutCompressPlain
	FalloutCompressStore
	FalloutCompressLZSS
)

type FalloutLZSS struct {
	LZSS
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
// "reader" position must be set to file offset before calling this functions
// "unpackedSize"
func (fallout FalloutLZSS) Decompress(reader io.Reader, unpackedSize uint32) ([]byte, error) {
	var self = "lzss.Fallout.Decompress()"

	var packedSize int16

	// Fallout 1 data block starts with 2 bytes describing block size and  is stored as-is or compressed,
	// and compressed block size.

	if errBlock := binary.Read(reader, binary.BigEndian, &packedSize); errBlock != nil {
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
		if errBytes := binary.Read(reader, binary.BigEndian, &output); errBytes != nil {
			return nil, fmt.Errorf("%s cannot read uncompressed block: %w", self, errBytes)
		}

		return output, nil
	} else if packedSize == 0 {
		// If block size is 0, something funky happened, most likely
		return nil, fmt.Errorf("%s block size == 0", self)
	}

	// Call
	return fallout.LZSS.Decompress(reader, uint32(packedSize), unpackedSize)
}

func (fallout FalloutLZSS) DecompressFile(reader io.ReadSeeker, file dat.FalloutFile) ([]byte, error) {
	var self = "lzss.Fallout.DecompressFile()"

	if file.GetParentDir().GetParentDat().GetGame() != 1 {
		return nil, fmt.Errorf("%s can only decompress Fallout 1 files", self)
	}

	reader.Seek(int64(file.GetOffset()+file.GetSizePacked()), io.SeekStart)
	reader.Seek(int64(file.GetOffset()), io.SeekStart)

	if !file.GetPacked() {
		// Simply copy all bytes as-is, without without involving LZSS in the process
		var output = make([]byte, file.GetSizeReal())
		if errBytes := binary.Read(reader, binary.BigEndian, &output); errBytes != nil {
			return nil, fmt.Errorf("%s cannot read uncompressed block: %w", self, errBytes)
		}

		return output, nil
	}
	return fallout.Decompress(reader, uint32(file.GetSizeReal()))
}

func (fallout FalloutLZSS) CompressMethod(reader io.ReadSeeker, file dat.FalloutFile) (output uint8, err error) {
	if !file.GetPacked() {
		return FalloutCompressPlain, nil
	}

	// remember stream position
	var offset int64
	if offset, err = reader.Seek(0, io.SeekCurrent); err != nil {
		return 0xFF, err
	}

	// go to file block
	if _, err = reader.Seek(int64(file.GetOffset()), io.SeekStart); err != nil {
		return 0xFF, err
	}

	var packedSize int16
	if err = binary.Read(reader, binary.BigEndian, &packedSize); err != nil {
		return 0xFF, err
	}

	if packedSize < 0 {
		output = FalloutCompressStore
	} else if packedSize == 0 {
		output = FalloutCompressZero
	} else {
		output = FalloutCompressLZSS
	}

	// restore stream position
	if _, err = reader.Seek(offset, io.SeekStart); err != nil {
		return 0xFF, err
	}

	return output, nil
}
