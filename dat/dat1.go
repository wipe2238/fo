package dat

import (
	"encoding/binary"
	"fmt"
	"io"
)

type falloutFile_1 struct {
	Name string

	PackMethod int32
	Offset     int32
	SizeReal   int32
	SizePacked int32

	//

	parentDir *falloutDir_1
	Packed    bool
}

type falloutDir_1 struct {
	FilesCount int32
	Unknown    []int32

	//

	parentDat *falloutDat_1
	Path      string
	Files     []*falloutFile_1
}

type falloutDat_1 struct {
	DirsCount int32
	Header    []int32

	//

	stream io.Reader // TODO remove io.Reader
	Dirs   []*falloutDir_1
}

// readStream implements FalloutDat
func (dat *falloutDat_1) readStream(stream io.Reader, options *FalloutDatOptions) (err error) {
	dat.Reset()
	dat.stream = stream

	if dat.DirsCount, err = dat.readInt32(); err != nil {
		return err
	}

	if dat.Header, err = dat.readInt32array(Const().HeaderLen); err != nil {
		return err
	}

	// header validation

	switch dat.Header[0] {
	// known values
	case 0x5E: // master.dat
		break
	case 0x0A: // critter.dat
		break
	case 0x2E: // falldemo.dat
		break
	// errors
	case 0:
		return fmt.Errorf("dat1/read: header[0] = 0 %+v", dat.Header)
	default:
		return fmt.Errorf("dat1/read: unknown header[0] [0x%X 0x%X 0x%X] %+v", dat.Header[0], dat.Header[1], dat.Header[2], dat.Header)
	}

	if dat.Header[1] != 0 {
		return fmt.Errorf("dat1/read: invalid header[1] 0x%X %#v", dat.Header[1], dat.Header)
	}

	//
	// directories list
	//

	dat.Dirs = make([]*falloutDir_1, dat.DirsCount)

	for idx := range dat.Dirs {
		dat.Dirs[idx] = new(falloutDir_1)
		dat.Dirs[idx].parentDat = dat

		if dat.Dirs[idx].Path, err = dat.readString(); err != nil {
			return err
		}

		// TODO? cache dir.Name
	}
	//
	// directories files
	//
	for _, dir := range dat.Dirs {
		if err = dat.readDir(dir); err != nil {
			return err
		}
	}

	return nil
}

func (dat *falloutDat_1) readDir(dir *falloutDir_1) (err error) {
	// probably pointless to keep that in struct
	if dir.FilesCount, err = dat.readInt32(); err != nil {
		return err
	}

	// same as above, plus it', well, unknown
	if dir.Unknown, err = dat.readInt32array(Const().HeaderLen); err != nil {
		return err
	}

	dir.Files = make([]*falloutFile_1, dir.FilesCount)

	for idx := range dir.Files {
		dir.Files[idx] = new(falloutFile_1)
		dir.Files[idx].parentDir = dir

		if err = dat.readFile(dir.Files[idx]); err != nil {
			return err
		}
	}

	return nil
}

func (dat *falloutDat_1) readFile(file *falloutFile_1) (err error) {
	if file.Name, err = dat.readString(); err != nil {
		return err
	}

	var errMsg = "dat1/readFile(" + file.Name + ") "

	if file.PackMethod, err = dat.readInt32(); err == nil {
		switch file.PackMethod {
		case 0x10:
			// FalloutCompressStore???
			return fmt.Errorf("%s compression mode [0x%X=%d] not implemented", errMsg, file.PackMethod, file.PackMethod)
		case 0x20:
			// FalloutCompressPlain
		case 0x40:
			// FalloutCompressLZSS
			file.Packed = true
		default:
			return fmt.Errorf("%s unknown compression mode [0x%X=%d]", errMsg, file.PackMethod, file.PackMethod)
		}
	} else {
		return err
	}

	if file.Offset, err = dat.readInt32(); err != nil {
		return err
	}

	if file.SizeReal, err = dat.readInt32(); err != nil {
		return err
	}

	if file.SizePacked, err = dat.readInt32(); err != nil {
		return err
	}

	return nil
}

// TODO? generics

func (dat *falloutDat_1) readInt16() (out int16, err error) {
	if err = binary.Read(dat.stream, binary.BigEndian, &out); err != nil {
		return 0, err
	}

	return out, nil
}

func (dat *falloutDat_1) readInt32() (out int32, err error) {
	if err = binary.Read(dat.stream, binary.BigEndian, &out); err != nil {
		return 0, err
	}

	return out, nil
}

func (dat *falloutDat_1) readUInt8() (out uint8, err error) {
	if err = binary.Read(dat.stream, binary.BigEndian, &out); err != nil {
		return 0, err
	}

	return out, nil
}

func (dat *falloutDat_1) readByte() (out byte, err error) {
	if err = binary.Read(dat.stream, binary.BigEndian, &out); err != nil {
		return 0, err
	}

	return out, nil
}

/*
func (dat *falloutDat_1) readUInt8AsInt16() (out int16, err error) {
	var tmp uint8
	if err = binary.Read(dat.stream, binary.BigEndian, &tmp); err != nil {
		return 0, err
	}

	return int16(tmp), nil
}
*/

func (dat *falloutDat_1) readInt32array(size uint8) (out []int32, err error) {
	out = make([]int32, size)
	for idx := range out {
		if out[idx], err = dat.readInt32(); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (dat *falloutDat_1) readString() (str string, err error) {
	var len uint8
	if len, err = dat.readUInt8(); err != nil {
		return str, err
	}

	var buff = make([]byte, len)
	if err = binary.Read(dat.stream, binary.BigEndian, &buff); err != nil {
		return str, err
	}

	return string(buff), nil
}
