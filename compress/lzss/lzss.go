package lzss

import (
	"fmt"
	"io"
)

//

type LZSS struct {
	DictionarySize uint16
	MaxMatch       byte
	MinMatch       byte
}

var Default = LZSS{
	DictionarySize: 4096,
	MinMatch:       2,
	MaxMatch:       18,
}

// Decompress
func (lzss LZSS) Decompress(reader io.Reader, packedSize uint32, unpackedSize uint32) (output []byte, err error) {
	const self = "lzss.LZSS.Decompress"

	// sanity check
	if lzss.DictionarySize == 0 {
		return nil, fmt.Errorf("%s() DictionarySize == 0", self)
	} else if lzss.MinMatch == 0 {
		return nil, fmt.Errorf("%s() MinMatch == 0", self)
	} else if lzss.MaxMatch == 0 {
		return nil, fmt.Errorf("%s() MaxMatch == 0", self)
	} else if (lzss.DictionarySize % 2) != 0 {
		return nil, fmt.Errorf("%s DictionarySize %% 2 != 0", self)
	}

	//

	var (
		outputIdx     uint32
		dictionary    []byte = make([]byte, lzss.DictionarySize)
		dictionaryIdx uint16 = lzss.DictionarySize - uint16(lzss.MaxMatch)
		tmp           []byte = make([]byte, 1)
		flags         uint16
	)

	output = make([]byte, unpackedSize)

	for idx := range dictionary {
		dictionary[idx] = ' '
	}

	// Ported from code by @ghost2238 and @mattseabrook
	// https://github.com/rotators/Fo1in2/blob/master/Tools/UndatUI/src/dat.cs
	// https://github.com/mattseabrook/LZSS/blob/main/2023/lzss.cpp

	for packedSize > 0 {
		// @FlagNext
		flags >>= 1

		if (flags & uint16(0x0100)) == 0 {
			// @Flag
			// Read FL byte on very first loop, and every 9th loop after that
			// No need for `packedSize` check here, it's done naturally by `for`
			if _, err = reader.Read(tmp); err == nil {
				packedSize--
			} else {
				return nil, err
			}

			flags = uint16(tmp[0]) | 0xFF00
		}

		if (flags % 2) == 1 {
			// @FlagOdd
			// Read raw byte and copy without changes

			if packedSize == 0 {
				return nil, fmt.Errorf("%s(@FlagOdd) cannot read raw byte", self)
			} else if _, err = reader.Read(tmp); err == nil {
				packedSize--
			} else {
				return nil, err
			}

			output[outputIdx] = tmp[0]
			outputIdx++

			dictionary[(dictionaryIdx % lzss.DictionarySize)] = tmp[0]
			dictionaryIdx++
		} else {
			// @FlagEven
			// Read 2 bytes describing dictionary offset and length of data to copy

			var (
				dictionaryOffset uint16
				length           uint32
			)

			// DO, dictionary offset
			if packedSize == 0 {
				return nil, fmt.Errorf("%s(@FlagEven) cannot read DO byte", self)
			} else if _, err = reader.Read(tmp); err == nil {
				packedSize--
			} else {
				return nil, err
			}

			dictionaryOffset = uint16(tmp[0])

			// LB, length
			if packedSize == 0 {
				return nil, fmt.Errorf("%s(@FlagEven) cannot read LB byte", self)
			} else if _, err = reader.Read(tmp); err == nil {
				packedSize--
			} else {
				return nil, err
			}

			dictionaryOffset |= uint16((tmp[0] & 0xF0)) << 4
			length = uint32((tmp[0] & 0x0F)) + uint32(lzss.MinMatch)

			for idx := range length {
				tmp[0] = dictionary[((dictionaryOffset + uint16(idx)) % lzss.DictionarySize)]

				output[outputIdx] = tmp[0]
				outputIdx++

				dictionary[(dictionaryIdx % lzss.DictionarySize)] = tmp[0]
				dictionaryIdx++
			}
		}
	}

	return output, nil
}
