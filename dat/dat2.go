package dat

import (
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/wipe2238/fo/x/dbg"
)

type falloutDatV2 struct {
	FilesCount uint32 // DAT2: offset: EOF - SizeTree - 8
	SizeTree   uint32 // DAT2: offset: EOF - 8
	SizeDat    uint32 // DAT2: offset: EOF - 4

	Dirs []*falloutDirV2
	Dbg  dbg.Map
}

type falloutDirV2 struct {
	// DAT2 does not have directories, it's all made up!

	Path      string
	Files     []*falloutFileV2
	Dbg       dbg.Map
	parentDat *falloutDatV2
}

type falloutFileV2 struct {
	falloutShared

	Path       string // DAT2: len uint32, name [len]byte
	PackedMode uint8  // DAT2
	Index      uint16
	SizeReal   uint32 // DAT2
	SizePacked uint32 // DAT2
	Offset     uint32 // DAT2

	Dbg       dbg.Map
	parentDir *falloutDirV2
}

// readDat implements FalloutDat
func (dat *falloutDatV2) readDat(stream io.ReadSeeker) (err error) {
	const errPrefix = errPackage + "readDat(2)"

	dat.Dirs = make([]*falloutDirV2, 0)

	var streamLength int64

	if streamLength, err = stream.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	if streamLength < 12 {
		return fmt.Errorf("%s stream truncated, length < 12", errPrefix)
	}

	if _, err = stream.Seek(-8, io.SeekCurrent); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &dat.SizeTree); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &dat.SizeDat); err != nil {
		return err
	}

	if int64(dat.SizeDat) != streamLength {
		return fmt.Errorf("%s stream truncated, length != saved length [0x%X != 0x%x, %d != %d]", errPrefix, streamLength, dat.SizeDat, streamLength, dat.SizeDat)
	}

	if _, err = stream.Seek(int64((dat.SizeDat - dat.SizeTree - 8)), io.SeekStart); err != nil {
		return err
	}

	// 0x5A64 = 23140 = MASTER.DAT, Steam
	// 0x1BD0 = 7120  = CRITTER.DAT, Steam
	if err = binary.Read(stream, binary.LittleEndian, &dat.FilesCount); err != nil {
		return err
	}

	// DAT2 does not have a concept of directories, so it needs to be invented
	// While reading, we use map[string]* for quick lookup, plus []string for keys/directories names
	// which will be sorted later

	var (
		dirsMap     = make(map[string]*falloutDirV2)
		dirsMapKeys = make([]string, 0)
	)

	for idxFile := range dat.FilesCount {
		var file = new(falloutFileV2)

		if err = dat.readFile(stream, file); err != nil {
			return err
		}

		// First file in newly discovered directory decides if it's lowercased, uppercased, or mixed
		// That's to prevent badly packed .dat files to either create duplicated FalloutDir objects,
		// or creating N directories during extraction on case-sensitive operating systems

		// Other than that, directory part of file.Path is completely ignored,
		// and will always use parent directory path (see falloutFileV2.GetPath())

		var dirPath = path.Dir(strings.ReplaceAll(file.Path, `\`, "/"))
		var dirPathUpper = strings.ToUpper(dirPath)

		var dir, ok = dirsMap[dirPathUpper]
		if !ok {
			dir = new(falloutDirV2)
			dir.parentDat = dat
			dir.Path = dirPath // unknown/original case

			dirsMap[dirPathUpper] = dir
			dirsMapKeys = append(dirsMapKeys, dirPathUpper)
		}

		file.Index = uint16(idxFile)
		file.parentDir = dir
		dir.Files = append(dir.Files, file)
	}

	// Sort keys/directories names and fill dat.Dirs

	slices.Sort(dirsMapKeys)
	dat.Dirs = make([]*falloutDirV2, len(dirsMapKeys))
	for idx, dirNameUpper := range dirsMapKeys {
		dat.Dirs[idx] = dirsMap[dirNameUpper]
	}

	return nil
}

func (dat *falloutDatV2) readFile(stream io.ReadSeeker, file *falloutFileV2) (err error) {
	//const errPrefix = errPackage + "readFile(2)"

	if file.Path, err = dat.readString(stream); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &file.PackedMode); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &file.SizeReal); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &file.SizePacked); err != nil {
		return err
	}

	if err = binary.Read(stream, binary.LittleEndian, &file.Offset); err != nil {
		return err
	}

	return nil
}

func (dat *falloutDatV2) readString(stream io.ReadSeeker) (str string, err error) {
	var lenght uint32

	if err = binary.Read(stream, binary.LittleEndian, &lenght); err != nil {
		return str, err
	}

	var buff = make([]byte, lenght)
	if _, err = stream.Read(buff); err != nil {
		return str, err
	}

	return string(buff), nil
}
