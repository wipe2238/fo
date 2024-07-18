package dat

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strings"

	"github.com/wipe2238/fo/x/dbg"
)

//
// FalloutDat implementation
//

// GetGame implements FalloutDat
func (dat *falloutDatV2) GetGame() uint8 {
	return 2
}

// GetDirs implements FalloutDat
func (dat *falloutDatV2) GetDirs() (dirs []FalloutDir) {
	dirs = make([]FalloutDir, len(dat.Dirs))

	for idx, dir := range dat.Dirs {
		dirs[idx] = dir
	}

	return dirs
}

// GetDbg implements FalloutDat
func (dat *falloutDatV2) GetDbg() dbg.Map {
	return dat.Dbg
}

// FillDbg implements FalloutDat
func (dat *falloutDatV2) FillDbg() {
	dat.Dbg["DAT2:0:FilesCount"] = dat.FilesCount
	dat.Dbg["DAT2:1:SizeTree"] = dat.SizeTree
	dat.Dbg["DAT2:2:SizeDat"] = dat.SizeDat

	for idxDir, dir := range dat.Dirs {
		dir.Dbg["Idx"] = uint16(idxDir)

		for idxFile, file := range dir.Files {
			file.Dbg["DAT2:0:Path"] = file.Path
			file.Dbg["DAT2:1:PackedMode"] = file.PackedMode
			file.Dbg["DAT2:2:SizeReal"] = file.SizeReal
			file.Dbg["DAT2:3:SizePacked"] = file.SizePacked
			file.Dbg["DAT2:4:Offset"] = file.Offset

			file.Dbg["Idx:Dir"] = uint16(idxFile)
			file.Dbg["Idx:Dat"] = file.Index

			dbgAddContentStats(file, dir, dat)
		}

		dbgCleanupContentStats(dir.Dbg)
	}

	dbgCleanupContentStats(dat.Dbg)
}

//
// FalloutDir implementation
//

// GetParentDat implements FalloutDir
func (dir *falloutDirV2) GetParentDat() FalloutDat {
	return dir.parentDat
}

// GetName implements FalloutDir
func (dir *falloutDirV2) GetName() string {
	var path = strings.Split(dir.Path, "/")

	return path[len(path)-1]
}

// GetPath implements FalloutDir
func (dir *falloutDirV2) GetPath() string {
	return dir.Path
}

// GetFiles implements FalloutDir
func (dir *falloutDirV2) GetFiles() (files []FalloutFile) {
	files = make([]FalloutFile, len(dir.Files))

	for idx, file := range dir.Files {
		files[idx] = file
	}

	return files
}

// GetDbg implements FalloutDir
func (dir *falloutDirV2) GetDbg() dbg.Map {
	return dir.Dbg
}

//
// FalloutFile implementation
//

// GetParentDat implements FalloutFile
func (file *falloutFileV2) GetParentDat() FalloutDat {
	return file.GetParentDir().GetParentDat()
}

// GetParentDir implements FalloutFile
func (file *falloutFileV2) GetParentDir() FalloutDir {
	return file.parentDir
}

// GetName implements FalloutFile
func (file *falloutFileV2) GetName() string {
	var path = strings.Split(file.Path, "/")

	return path[len(path)-1]
}

// GetPath implements FalloutFile
func (file *falloutFileV2) GetPath() string {
	return file.GetParentDir().GetPath() + "/" + file.GetName()
}

// GetOffset implements FalloutFile
func (file *falloutFileV2) GetOffset() int64 {
	return int64(file.Offset)
}

func (file *falloutFileV2) GetSizeReal() int64 {
	return int64(file.SizeReal)
}

func (file *falloutFileV2) GetSizePacked() int64 {
	return int64(file.SizePacked)
}

// GetPacked implements FalloutFile
func (file *falloutFileV2) GetPacked() bool {
	return file.PackedMode > 0
}

// GetPackedMode implements FalloutFile
func (file *falloutFileV2) GetPackedMode() uint32 {
	return uint32(file.PackedMode)
}

// GetBytesReal implements FalloutFile
func (file *falloutFileV2) GetBytesReal(stream io.ReadSeeker) (bytesReal []byte, err error) {
	file.seek(stream, file)

	// TODO: switch to different reader
	var bytesPacked []byte
	if bytesPacked, err = file.GetBytesPacked(stream); err != nil {
		return nil, err
	}
	var bytesPackedReader = bytes.NewReader(bytesPacked)

	var zstream io.ReadCloser
	if zstream, err = zlib.NewReader(bytesPackedReader); err != nil {
		return nil, err
	}
	defer zstream.Close()

	if bytesReal, err = io.ReadAll(zstream); err != nil {
		return nil, err
	}

	// Quickly discard obviously incorrect data
	if int64(len(bytesReal)) != file.GetSizeReal() {
		return nil, fmt.Errorf("%s[2] decompressed size mismatch: have(%d) != want(%d)", errPackage, len(bytesReal), file.GetSizeReal())
	}

	return bytesReal, nil
}

// GetBytesPacked implements FalloutFile
func (file *falloutFileV2) GetBytesPacked(stream io.ReadSeeker) ([]byte, error) {
	return file.getBytesPacked(stream, file)
}

// GetDbg implements FalloutFile
func (file *falloutFileV2) GetDbg() dbg.Map {
	return file.Dbg
}