package dat

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"path"
	"strings"
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

// FillDbg implements FalloutDat
func (dat *falloutDatV2) FillDbg() {
	// TODO: deleteme
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
	return path.Base(dir.GetPath())
}

// GetPath implements FalloutDir
func (dir *falloutDirV2) GetPath() string {
	// `dir.Path` always uses cleaned up *nix format, see `falloutDatV2.readDat()`

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
func (file *falloutFileV2) GetName() (fileName string) {
	fileName = strings.ReplaceAll(file.Path, `\`, "/")

	return path.Base(fileName)
}

// GetPath implements FalloutFile
func (file *falloutFileV2) GetPath() string {
	return file.GetParentDir().GetPath() + "/" + file.GetName()
}

// GetOffset implements FalloutFile
func (file *falloutFileV2) GetOffset() int64 {
	return int64(file.Offset)
}

// GetSizeReal implements FalloutFile
func (file *falloutFileV2) GetSizeReal() int64 {
	return int64(file.SizeReal)
}

// GetSizePacked implements FalloutFile
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
	return file.getBytesReal(stream, file)
}

// GetBytesPacked implements FalloutFile
func (file *falloutFileV2) GetBytesPacked(stream io.ReadSeeker) ([]byte, error) {
	return file.getBytesPacked(stream, file)
}

func (file *falloutFileV2) GetBytesUnpacked(bytesPacked []byte) (bytesUnpacked []byte, err error) {
	var bytesReader = bytes.NewReader(bytesPacked)

	var zstream io.ReadCloser
	if zstream, err = zlib.NewReader(bytesReader); err != nil {
		return nil, err
	}
	defer zstream.Close()

	if bytesUnpacked, err = io.ReadAll(zstream); err != nil {
		return nil, err
	}

	// Quickly discard obviously incorrect data
	if int64(len(bytesUnpacked)) != file.GetSizeReal() {
		return nil, fmt.Errorf("%s[2] decompressed size mismatch: have(%d) != want(%d)", errPackage, len(bytesUnpacked), file.GetSizeReal())
	}

	return bytesUnpacked, nil
}
