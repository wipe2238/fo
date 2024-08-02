package dat

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/x/dbg"
)

type falloutDatV1 struct {
	DirsCount int32    // DAT1
	Header    [3]int32 // DAT1

	Dirs []*falloutDirV1
	Dbg  dbg.Map
}

type falloutDirV1 struct {
	Path       string   // DAT1: len byte, name [len]byte
	FilesCount int32    // DAT1
	Header     [3]int32 // DAT1

	Files     []*falloutFileV1
	Dbg       dbg.Map
	parentDat *falloutDatV1
}

type falloutFileV1 struct {
	falloutShared

	Name       string // DAT1: len byte, name [len]byte
	PackedMode uint32 // DAT1
	Offset     uint32 // DAT1
	SizeReal   uint32 // DAT1
	SizePacked uint32 // DAT1

	Dbg       dbg.Map
	parentDir *falloutDirV1
}

// readDat implements FalloutDat
func (dat *falloutDatV1) readDat(stream io.ReadSeeker) (err error) {
	const errPrefix = errPackage + "readDat(1)"

	if err = binary.Read(stream, binary.BigEndian, &dat.DirsCount); err != nil {
		return err
	}

	if dat.DirsCount < 1 {
		return fmt.Errorf("%s DirsCount(%d) < 1", errPrefix, dat.DirsCount)
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
		return fmt.Errorf("%s header[0] = 0, %s", errPrefix, dbg.Fmt(" ", dat.Header))
	default:
		// ignore?
		return fmt.Errorf("%s header[0] = unknown, %s", errPrefix, dbg.Fmt(" ", dat.Header))
	}

	if dat.Header[1] != 0 {
		return fmt.Errorf("%s header[1] != 0, %s", errPrefix, dbg.Fmt(" ", dat.Header))
	}

	//
	// directories list
	//

	dat.Dirs = make([]*falloutDirV1, dat.DirsCount)

	for idx := range dat.Dirs {
		dat.Dirs[idx] = new(falloutDirV1)
		dat.Dirs[idx].parentDat = dat

		// do not introduce anything here which could change length of dir.Path
		// SetDbg() relies on dir.Path length for offset jumping
		if dat.Dirs[idx].Path, err = dat.readString(stream); err != nil {
			return err
		}
	}

	//
	// directories files
	//

	for _, dir := range dat.Dirs {
		if err = dat.readDir(stream, dir); err != nil {
			return err
		}
	}

	return nil
}

func (dat *falloutDatV1) readDir(stream io.ReadSeeker, dir *falloutDirV1) (err error) {
	//const errPrefix = errPackage + "readDir(1)"

	if err = binary.Read(stream, binary.BigEndian, &dir.FilesCount); err != nil {
		return err
	}

	// Unknown[1] always 0x10
	for idx := range dir.Header {
		if err = binary.Read(stream, binary.BigEndian, &dir.Header[idx]); err != nil {
			return err
		}
	}

	dir.Files = make([]*falloutFileV1, dir.FilesCount)

	for idx := range dir.Files {
		dir.Files[idx] = new(falloutFileV1)
		dir.Files[idx].parentDir = dir

		if err = dat.readFile(stream, dir.Files[idx]); err != nil {
			return err
		}
	}

	return nil
}

func (dat *falloutDatV1) readFile(stream io.ReadSeeker, file *falloutFileV1) (err error) {
	const errPrefix = errPackage + "readFile(1)"

	// do not introduce anything here which could change length of file.Name
	// SetDbg() relies on file.Name length for offset jumping
	if file.Name, err = dat.readString(stream); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.BigEndian, &file.PackedMode); err != nil {
		return err
	}

	switch file.PackedMode {
	case lzss.FalloutCompressStore:
		// 0x10 is used in db.c as one of values, but so far haven't seen it in wild
		return fmt.Errorf("%s PackedMode not implemented (yet?) [0x%X=%d]", errPrefix, file.PackedMode, file.PackedMode)
	case lzss.FalloutCompressNone:
		//
	case lzss.FalloutCompressLZSS:
		//
	default:
		return fmt.Errorf("%s unknown PackedMode [0x%X=%d]", errPrefix, file.PackedMode, file.PackedMode)
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

	return nil
}

func (dat *falloutDatV1) readString(stream io.Reader) (str string, err error) {
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
