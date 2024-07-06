package dat

import (
	"io"
	"strings"
)

//
// FalloutFile implementation
//

// GetParentDir implements FalloutFile
func (file *falloutFile_1) GetParentDir() FalloutDir {
	return file.parentDir
}

// GetName implements FalloutFile
func (file *falloutFile_1) GetName() string {
	return file.Name
}

func (file *falloutFile_1) GetPath() string {
	return strings.ReplaceAll(file.GetParentDir().GetPath()+"/"+file.GetName(), "\\", "/")
}

// GetPacked implements FalloutFile
func (file *falloutFile_1) GetPacked() bool {
	return file.Packed
}

// GetOffset implements FalloutFile
func (file *falloutFile_1) GetOffset() int32 {
	return file.Offset
}

// GetOffset implements FalloutFile
func (file *falloutFile_1) GetSizeReal() int32 {
	return file.SizeReal
}

// GetOffset implements FalloutFile
func (file *falloutFile_1) GetSizePacked() int32 {
	return file.SizePacked
}

// GetBytesReal implements FalloutFile
//
// TODO replace with `compress/lzss“
func (file *falloutFile_1) GetBytesReal(stream io.ReadSeeker) (out []byte, err error) {
	// NOTE: for now, everything lives in single function, except read*()
	//       maybe it should be moved somewhere else and split when it's proven to work correctly

	if !file.GetPacked() {
		return file.GetBytesPacked(stream)
	}

	// seekFile() sets stream position to file offset, plus some additional checks
	// called early to not waste time initializing stuff for stream which can't be used
	if err = seekFile(stream, file); err != nil {
		return nil, err
	}

	// using falloutDat_1 for read*() functions
	var dat = file.GetParentDir().GetParentDat().(*falloutDat_1)

	//
	// LZSS decompression
	// Based on Ghosthack's UndatUI
	//
	// https://github.com/ghost2238/
	// https://github.com/rotators/Fo1in2/blob/master/Tools/UndatUI/src/dat.cs
	//

	// constants and variables

	const (
		DICT_SIZE int16 = 4096
		MIN_MATCH int16 = 3
		MAX_MATCH int16 = 18
	)

	var (
		N  int16 // number of bytes to read
		NR int16 // bytes read from last block header
		DO int16 // dictionary offset - for reading
		DI int16 // dictionary index - for writing
		OI int32 // output index, used for writing to the output array
		L  int32 // match length
		FL byte  // @Flags indicating the compression status of up to 8 next bytes

		NRx int16 // bytes read from stream which don't increase NR
	)

	// output arrays

	out = make([]byte, file.GetSizeReal())
	var dictionary = make([]byte, DICT_SIZE)

	var lastByte = func() bool {
		return int32((NR + NRx)) == file.GetSizePacked()
	}

	var readByte = func() (data byte, err error) {
		NR++
		if data, err = dat.readByte(); err != nil {
			return 0, err
		}

		return data, nil
	}

	var writeByte = func(b byte) {
		out[OI] = b
		OI++
		dictionary[(DI % DICT_SIZE)] = b
		DI++
	}

	var readBlock = func() (err error) {
		NR = 0
		if N < 0 {
			panic("TODO: n < 0 -> uncompressed block")
		} else { // n > 0 only; n == 0 must be handled in caller
			// clear dictionary
			for idx := range dictionary {
				dictionary[idx] = 0x20 // ASCII ' '
			}
			DI = DICT_SIZE - MAX_MATCH

			// @Flag
			for {
				if NR >= N || lastByte() {
					return nil // Go to @Start
				}

				// Read flag byte
				if FL, err = readByte(); err != nil {
					return err
				}

				if NR >= N || lastByte() {
					return nil // Go to @Start
				}

				for range 8 { // XXX magic number
					var tmp byte

					// @FLodd, normal byte
					if (FL % 2) == 1 {
						// Read byte from stream and put it in the output buffer and dictionary

						if tmp, err = readByte(); err == nil {
							writeByte(tmp)
						} else {
							return err
						}

						if NR >= N {
							return nil
						}
						// @FLeven, encoded dictionary offset
					} else {
						if NR >= N {
							return nil
						}

						// Read dictionary offset byte
						if tmp, err = readByte(); err == nil {
							DO = int16(tmp)
						} else {
							return err
						}

						if NR >= N {
							return nil
						}

						// LB, length byte
						if tmp, err = readByte(); err != nil {
							return err
						}

						DO |= int16((tmp & 0xF0)) << 4             // Prepend the high-nibble (first 4 bits) from LB to DO
						L = int32(int16((tmp & 0x0F)) + MIN_MATCH) // and remove it from LB and add MIN_MATCH

						for range L {
							// Read a byte from the dictionary at DO, increment index and write to output and dictionary at DI
							writeByte(dictionary[(DO % DICT_SIZE)])
							DO++
						}
					}

					// @FlagNext
					FL = byte(FL >> 1)
					if lastByte() {
						return nil
					}
				}

			} // loop
		}
	}

	// func Decompress() (out []byte, err error)
	for !lastByte() {
		// @Start
		if N, err = dat.readInt16(); err == nil {
			NRx = 2 // sizeof(int16), used by lastByte()
		} else {
			return nil, err
		}

		if N == 0 {
			break
		} else if err = readBlock(); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (file *falloutFile_1) GetBytesPacked(stream io.ReadSeeker) (out []byte, err error) {
	// seekFile() sets stream position to file offset, plus some additional checks
	// called early to not waste time initializing stuff for stream which can't be used
	if err = seekFile(stream, file); err != nil {
		return nil, err
	}

	out = make([]byte, file.GetSizePacked())

	for idx := range out {
		if out[idx], err = file.GetParentDir().GetParentDat().(*falloutDat_1).readUInt8(); err != nil {
			return nil, err
		}
	}

	return out, nil
}

//
// FalloutDir implementation
//

// GetParentDat implements FalloutDir
func (dir *falloutDir_1) GetParentDat() FalloutDat {
	return dir.parentDat
}

// GetName implements FalloutDir
func (dir *falloutDir_1) GetName() string {
	var path = strings.Split(dir.Path, "\\")

	return path[len(path)-1]
}

// GetPath implements FalloutDir
func (dir *falloutDir_1) GetPath() string {
	return strings.ReplaceAll(dir.Path, "\\", "/")
}

// GetFiles implements FalloutDir
func (dir *falloutDir_1) GetFiles() (files []FalloutFile, err error) {
	files = make([]FalloutFile, 0)

	for idx := range dir.Files {
		files = append(files, dir.Files[idx])
	}

	return files, nil
}

//
// FalloutDat implementation
//

// Reset implements FalloutDat
func (dat *falloutDat_1) Reset() {
	dat.stream = nil

	clear(dat.Header)
	clear(dat.Dirs)

	dat.DirsCount = 0
}

// GetGame implements FalloutDat
func (dat *falloutDat_1) GetGame() byte {
	return 1
}

// GetDirs implements FalloutDat
func (dat *falloutDat_1) GetDirs() (dirs []FalloutDir, err error) {
	dirs = make([]FalloutDir, len(dat.Dirs))

	for idx, dir := range dat.Dirs {
		dirs[idx] = dir //append(dirs, dir) //dat.Dirs[idx])
	}

	return dirs, nil
}
