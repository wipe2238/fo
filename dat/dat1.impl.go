package dat

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/wipe2238/fo/compress/lzss"
)

//
// FalloutDat implementation
//

// GetGame implements FalloutDat
func (dat *falloutDatV1) GetGame() byte {
	return 1
}

// GetDirs implements FalloutDat
func (dat *falloutDatV1) GetDirs() (dirs []FalloutDir) {
	dirs = make([]FalloutDir, len(dat.Dirs))

	for idx, dir := range dat.Dirs {
		dirs[idx] = dir
	}

	return dirs
}

//
// FalloutDir implementation
//

// GetParentDat implements FalloutDir
func (dir *falloutDirV1) GetParentDat() FalloutDat {
	return dir.parentDat
}

// GetName implements FalloutDir
func (dir *falloutDirV1) GetName() string {
	return path.Base(dir.GetPath())
}

// GetPath implements FalloutDir
func (dir *falloutDirV1) GetPath() (dirPath string) {
	dirPath = strings.ReplaceAll(dir.Path, `\`, "/") + "/"

	return path.Dir(dirPath)
}

// GetFiles implements FalloutDir
func (dir *falloutDirV1) GetFiles() (files []FalloutFile) {
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
func (file *falloutFileV1) GetParentDat() FalloutDat {
	return file.GetParentDir().GetParentDat()
}

// GetParentDir implements FalloutFile
func (file *falloutFileV1) GetParentDir() FalloutDir {
	return file.parentDir
}

// GetName implements FalloutFile
func (file *falloutFileV1) GetName() string {
	return file.Name
}

func (file *falloutFileV1) GetPath() string {
	return file.GetParentDir().GetPath() + "/" + file.GetName()
}

// GetOffset implements FalloutFile
func (file *falloutFileV1) GetOffset() int64 {
	return int64(file.Offset)
}

// GetOffset implements FalloutFile
func (file *falloutFileV1) GetSizeReal() int64 {
	return int64(file.SizeReal)
}

// GetSizePacked implements FalloutFile
func (file *falloutFileV1) GetSizePacked() int64 {
	if file.SizePacked == 0 {
		return file.GetSizeReal()
	}

	return int64(file.SizePacked)
}

// GetPacked implements FalloutFile
func (file *falloutFileV1) GetPacked() bool {
	return file.PackedMode != lzss.FalloutCompressNone
}

// GetPackedMode implements FalloutFile
func (file *falloutFileV1) GetPackedMode() uint32 {
	return file.PackedMode
}

// GetBytesReal implements FalloutFile
func (file *falloutFileV1) GetBytesReal(stream io.ReadSeeker) (bytesReal []byte, err error) {
	return file.getBytesReal(stream, file)
}

// GetBytesPacked implements FalloutFile
func (file *falloutFileV1) GetBytesPacked(stream io.ReadSeeker) ([]byte, error) {
	return file.getBytesPacked(stream, file)
}

func (file *falloutFileV1) GetBytesUnpacked(bytesPacked []byte) (bytesUnpacked []byte, err error) {
	var lzssFile = lzss.FalloutFile{
		Stream:       bytes.NewReader(bytesPacked),
		SizePacked:   file.GetSizePacked(),
		CompressMode: file.GetPackedMode(),
	}

	if bytesUnpacked, err = lzssFile.Decompress(); err != nil {
		return nil, err
	}

	// Quickly discard obviously incorrect data
	if int64(len(bytesUnpacked)) != file.GetSizeReal() {
		return nil, fmt.Errorf("%s[1] decompressed size mismatch: have(%d) != want(%d)", errPackage, len(bytesUnpacked), file.GetSizeReal())
	}

	return bytesUnpacked, nil
}
