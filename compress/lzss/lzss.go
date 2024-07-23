package lzss

import (
	"fmt"
	"io"
)

//

type LZSS struct {
	DictionarySize uint16
	MinMatch       byte
	MaxMatch       byte
}

var Default = LZSS{
	DictionarySize: 4096,
	MinMatch:       2,
	MaxMatch:       18,
}

const errPackage = "fo/compress/lzss:"

// Decompress
func (lzss LZSS) Decompress(reader io.Reader, size int64) (output []byte, err error) {
	// sanity check
	if lzss.DictionarySize == 0 {
		return nil, fmt.Errorf("%s DictionarySize == 0", errPackage)
	} else if lzss.MinMatch == 0 {
		return nil, fmt.Errorf("%s MinMatch == 0", errPackage)
	} else if lzss.MaxMatch == 0 {
		return nil, fmt.Errorf("%s MaxMatch == 0", errPackage)
	} else if (lzss.DictionarySize % 2) != 0 {
		return nil, fmt.Errorf("%s DictionarySize %% 2 != 0", errPackage)
	} else if size < 0 {
		return nil, fmt.Errorf("%s size(%d) < 0", errPackage, size)
	}

	output = make([]byte, 0)

	// thank you for your hard work
	if size == 0 {
		return output, nil
	}

	var (
		dictionary    = make([]byte, lzss.DictionarySize)
		dictionaryIdx = lzss.DictionarySize - uint16(lzss.MaxMatch)
		tmp           = make([]byte, 1)
		flags         uint16
	)

	for idx := range dictionary {
		dictionary[idx] = ' '
	}

	// Ported from code by @ghost2238 and @mattseabrook
	// https://github.com/rotators/Fo1in2/blob/master/Tools/UndatUI/src/dat.cs
	// https://github.com/mattseabrook/LZSS/blob/main/2023/lzss.cpp

	for size > 0 {
		// @FlagNext
		flags >>= 1

		if (flags & uint16(0x0100)) == 0 {
			// @Flag
			// Read FL on very first loop, and every 9th loop after that

			// FL, flags
			// No need for `size` check here, it's done naturally by `for`
			if _, err = reader.Read(tmp); err == nil {
				size--
			} else {
				return nil, err
			}

			flags = uint16(tmp[0]) | 0xFF00
		}

		if (flags % 2) == 1 {
			// @FlagOdd
			// Read raw byte and copy without changes, fill dictionay

			if size == 0 {
				return nil, fmt.Errorf("%s @FlagOdd cannot read raw byte", errPackage)
			} else if _, err = reader.Read(tmp); err == nil {
				size--
			} else {
				return nil, err
			}

			output = append(output, tmp[0])

			dictionary[(dictionaryIdx % lzss.DictionarySize)] = tmp[0]
			dictionaryIdx++
		} else {
			// @FlagEven
			// Read 2 bytes describing dictionary offset and length of data to copy

			var (
				dictionaryOffset uint16
				length           uint16
			)

			// DO, dictionary offset
			if size == 0 {
				return nil, fmt.Errorf("%s @FlagEven cannot read DO byte", errPackage)
			} else if _, err = reader.Read(tmp); err == nil {
				size--
			} else {
				return nil, err
			}

			dictionaryOffset = uint16(tmp[0])

			// LB, length
			if size == 0 {
				return nil, fmt.Errorf("%s @FlagEven cannot read LB byte", errPackage)
			} else if _, err = reader.Read(tmp); err == nil {
				size--
			} else {
				return nil, err
			}

			dictionaryOffset |= uint16((tmp[0] & 0xF0)) << 4
			length = uint16((tmp[0] & 0x0F)) + uint16(lzss.MinMatch)

			for idx := range length {
				tmp[0] = dictionary[((dictionaryOffset + idx) % lzss.DictionarySize)]

				output = append(output, tmp[0])

				dictionary[(dictionaryIdx % lzss.DictionarySize)] = tmp[0]
				dictionaryIdx++
			}
		}
	}

	return output, nil
}
