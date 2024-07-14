package dat

import (
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
			file.Dbg["DAT1:2:CompressMode"] = file.CompressMode
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

// GetCompressMode implements FalloutFile
func (file *falloutFileV1) GetCompressMode() uint32 {
	return file.CompressMode
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

// GetFileLZSS returns converted file data, which can be passed to `lzss.FalloutLZSS“ functions
func (file *falloutFileV1) getFileLZSS() lzss.FalloutFileLZSS {
	return lzss.FalloutFileLZSS{
		Offset:       file.Offset,
		CompressMode: file.CompressMode,
		SizeReal:     file.SizeReal,
		SizePacked:   file.SizePacked,
	}
}

// GetBytesReal implements FalloutFile
func (file *falloutFileV1) GetBytesReal(stream io.ReadSeeker) (out []byte, err error) {
	if out, err = lzss.Fallout1.DecompressFile(stream, file.getFileLZSS()); err != nil {
		return nil, err
	}

	return out, nil
}

func (file *falloutFileV1) GetBytesPacked(stream io.ReadSeeker) (out []byte, err error) {
	// seekFile() sets stream position to file offset, plus some additional checks
	// called early to not waste time initializing stuff for stream which can't be used
	if err = seekFile(stream, file); err != nil {
		return nil, err
	}

	out = make([]byte, file.GetSizePacked())
	if _, err = stream.Read(out); err != nil {
		return nil, err
	}

	return out, nil
}

func (file *falloutFileV1) GetDbg() dbg.Map {
	return file.Dbg
}
