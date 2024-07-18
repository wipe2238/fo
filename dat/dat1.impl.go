package dat

import (
	"fmt"
	"io"
	"strings"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/x/dbg"
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

func (dat *falloutDatV1) GetDbg() dbg.Map {
	return dat.Dbg
}

func (dat *falloutDatV1) FillDbg() {
	const strHeader = "DAT1:1:Header"

	dat.Dbg["DAT1:0:DirsCount"] = dat.DirsCount
	dat.Dbg[strHeader] = dat.Header

	if dat.Dbg.KeysMaxLen("Offset:") > 0 {
		dat.Dbg.AddSize("Size:Tree:Info", "Offset:0:Info", "Offset:1:DirsNames")
		dat.Dbg.AddSize("Size:Tree:DirsNames", "Offset:1:DirsNames", "Offset:2:DirsData")
		dat.Dbg.AddSize("Size:Tree:DirsData", "Offset:2:DirsData", "Offset:3:FilesContent")
		dat.Dbg.AddSize("Size:Tree:Total", "Offset:0:Info", "Offset:3:FilesContent")
	}

	for idxDir, dir := range dat.Dirs {
		dir.Dbg["DAT1:0:FilesCount"] = dir.FilesCount
		dir.Dbg[strHeader] = dir.Header

		dir.Dbg["Idx"] = uint16(idxDir)

		if dir.Dbg.KeysMaxLen("Offset:") > 0 {
			dir.Dbg.AddSize("Size:DirEntry:Info", "Offset:0:Info", "Offset:1:Files")
			dir.Dbg.AddSize("Size:DirEntry:FilesData", "Offset:1:Files", "Offset:2:End")
			dir.Dbg.AddSize("Size:DirEntry:Total", "Offset:0:Info", "Offset:2:End")
		}

		for idxFile, file := range dir.Files {
			file.Dbg["DAT1:1:Name"] = file.Name
			file.Dbg["DAT1:2:PackedMode"] = file.PackedMode
			file.Dbg["DAT1:3:Offset"] = file.Offset
			file.Dbg["DAT1:4:SizeReal"] = file.SizeReal
			file.Dbg["DAT1:5:SizePacked"] = file.SizePacked

			if file.Dbg.KeysMaxLen("Offset:") > 0 {
				file.Dbg.AddSize("Size:FileEntry:Name", "Offset:0:Name", "Offset:1:Info")
				file.Dbg.AddSize("Size:FileEntry:Info", "Offset:1:Info", "Offset:2:End")
				file.Dbg.AddSize("Size:FileEntry:Total", "Offset:0:Name", "Offset:2:End")
			}

			//

			dbgAddContentStats(file, dir, dat)

			file.Dbg["Idx:Dir"] = uint16(idxFile)
			file.Dbg["Idx:Dat"] = uint16(dat.Dbg[(dbgContent+dbgAll+dbgCount)].(int64) - 1)
		}

		dbgCleanupContentStats(dir.Dbg)
	}

	dbgCleanupContentStats(dat.Dbg)
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
	var path = strings.Split(dir.Path, "\\")

	return path[len(path)-1]
}

// GetPath implements FalloutDir
func (dir *falloutDirV1) GetPath() string {
	return strings.ReplaceAll(dir.Path, `\`, "/")
}

// GetFiles implements FalloutDir
func (dir *falloutDirV1) GetFiles() (files []FalloutFile) {
	files = make([]FalloutFile, len(dir.Files))

	for idx, file := range dir.Files {
		files[idx] = file
	}

	return files
}

func (dir *falloutDirV1) GetDbg() dbg.Map {
	return dir.Dbg
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
	return strings.ReplaceAll(file.GetParentDir().GetPath()+"/"+file.GetName(), "\\", "/")
}

// GetOffset implements FalloutFile
func (file *falloutFileV1) GetOffset() int64 {
	return int64(file.Offset)
}

// GetOffset implements FalloutFile
func (file *falloutFileV1) GetSizeReal() int64 {
	return int64(file.SizeReal)
}

// GetSizePacked returns size of file block
//
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

// LZSS returns file data converted to struct used by `compress/lzss` package
func (file *falloutFileV1) LZSS(stream io.ReadSeeker) lzss.FalloutFile {
	return lzss.FalloutFile{
		Stream:       stream,
		Offset:       file.GetOffset(),
		SizePacked:   file.GetSizePacked(),
		CompressMode: file.GetPackedMode(),
	}
}

// GetBytesReal implements FalloutFile
func (file *falloutFileV1) GetBytesReal(stream io.ReadSeeker) (bytesReal []byte, err error) {
	if bytesReal, err = file.LZSS(stream).Decompress(); err != nil {
		return nil, err
	}

	// Quickly discard obviously incorrect data
	if int64(len(bytesReal)) != file.GetSizeReal() {
		return nil, fmt.Errorf("%s[1] decompressed size mismatch: have(%d) != want(%d)", errPackage, len(bytesReal), file.GetSizeReal())
	}

	return bytesReal, nil
}

func (file *falloutFileV1) GetBytesPacked(stream io.ReadSeeker) ([]byte, error) {
	return file.getBytesPacked(stream, file)
}

func (file *falloutFileV1) GetDbg() dbg.Map {
	return file.Dbg
}
