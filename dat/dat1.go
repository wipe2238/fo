package dat

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/x/dbg"
)

type falloutFile_1 struct {
	Name         string // DAT1: len byte, name [len]byte
	CompressMode uint32 // DAT1
	Offset       uint32 // DAT1
	SizeReal     uint32 // DAT1
	SizePacked   uint32 // DAT1

	Dbg       dbg.Map
	parentDir *falloutDir_1
}

type falloutDir_1 struct {
	Path       string   // DAT1: len byte, name [len]byte
	FilesCount int32    // DAT1
	Header     [3]int32 // DAT1

	Files     []*falloutFile_1
	Dbg       dbg.Map
	parentDat *falloutDat_1
}

type falloutDat_1 struct {
	DirsCount int32    // DAT1
	Header    [3]int32 // DAT1

	Dirs []*falloutDir_1
	Dbg  dbg.Map
}

// readDat implements FalloutDat
func (dat *falloutDat_1) readDat(stream io.Reader) (err error) {
	dat.Dbg = make(dbg.Map)
	dat.Dbg.AddOffset("Offset:0:Info", stream)
	if err = binary.Read(stream, binary.BigEndian, &dat.DirsCount); err != nil {
		return err
	}

	for idx := range dat.Header {
		if err = binary.Read(stream, binary.BigEndian, &dat.Header[idx]); err != nil {
			return err
		}
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

	dat.Dbg.AddOffset("Offset:1:DirsNames", stream)

	dat.Dirs = make([]*falloutDir_1, dat.DirsCount)

	for idx := range dat.Dirs {
		dat.Dirs[idx] = new(falloutDir_1)
		dat.Dirs[idx].Dbg = make(dbg.Map)
		dat.Dirs[idx].parentDat = dat

		if dat.Dirs[idx].Path, err = dat.readString(stream); err != nil {
			return err
		}

		// TODO:? cache dir.Name
	}

	//
	// directories files
	//

	dat.Dbg.AddOffset("Offset:2:DirsData", stream)

	for _, dir := range dat.Dirs {
		if err = dat.readDir(stream, dir); err != nil {
			return err
		}
	}

	dat.Dbg.AddOffset("Offset:3:FilesContent", stream)

	return nil
}

func (dat *falloutDat_1) readDir(stream io.Reader, dir *falloutDir_1) (err error) {
	dir.Dbg = make(dbg.Map)
	dir.Dbg.AddOffset("Offset:0:Info", stream)

	if err = binary.Read(stream, binary.BigEndian, &dir.FilesCount); err != nil {
		return err
	}

	// Unknown[1] always 0x10, probably obsolete file entry size, which (excluding)
	for idx := range dir.Header {
		if err = binary.Read(stream, binary.BigEndian, &dir.Header[idx]); err != nil {
			return err
		}
	}

	dir.Dbg.AddOffset("Offset:1:Files", stream)

	dir.Files = make([]*falloutFile_1, dir.FilesCount)

	for idx := range dir.Files {
		dir.Files[idx] = new(falloutFile_1)
		dir.Files[idx].Dbg = make(dbg.Map)
		dir.Files[idx].parentDir = dir

		if err = dat.readFile(stream, dir.Files[idx]); err != nil {
			return err
		}
	}

	dir.Dbg.AddOffset("Offset:2:End", stream)

	return nil
}

func (dat *falloutDat_1) readFile(stream io.Reader, file *falloutFile_1) (err error) {
	file.Dbg.AddOffset("Offset:0:Name", stream)

	if file.Name, err = dat.readString(stream); err != nil {
		return err
	}

	file.Dbg.AddOffset("Offset:1:Info", stream)

	var errMsg = "dat1/readFile(" + file.Name + ") "

	if err = binary.Read(stream, binary.BigEndian, &file.CompressMode); err != nil {
		return err
	}

	switch file.CompressMode {
	case lzss.FalloutCompressStore:
		// 0x10 is used in db.c as one of values, but so far haven't seen it in wild
		return fmt.Errorf("%s compress mode [0x%X=%d] not implemented yet", errMsg, file.CompressMode, file.CompressMode)
	case lzss.FalloutCompressNone:
		//
	case lzss.FalloutCompressLZSS:
		//
	default:
		return fmt.Errorf("%s unknown compression mode [0x%X=%d]", errMsg, file.CompressMode, file.CompressMode)
	}

	if err = binary.Read(stream, binary.BigEndian, &file.Offset); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.BigEndian, &file.SizeReal); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.BigEndian, &file.SizePacked); err != nil {
		return err
	}

	file.Dbg.AddOffset("Offset:2:End", stream)

	return nil
}

func (dat *falloutDat_1) readString(stream io.Reader) (str string, err error) {
	var buff = make([]byte, 1)

	if _, err = stream.Read(buff); err != nil {
		return str, err
	}

	buff = make([]byte, buff[0])
	if _, err = stream.Read(buff); err != nil {
		return str, err
	}

	return string(buff), nil
}
