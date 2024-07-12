package dat

import (
	"io"
	"path"
	"strings"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/x/dbg"
)

//
// FalloutDat implementation
//

// GetGame implements FalloutDat
func (dat *falloutDat_1) GetGame() byte {
	return 1
}

// GetDirs implements FalloutDat
func (dat *falloutDat_1) GetDirs() (dirs []FalloutDir) {
	dirs = make([]FalloutDir, len(dat.Dirs))

	for idx, dir := range dat.Dirs {
		dirs[idx] = dir
	}

	return dirs
}

func (dat *falloutDat_1) GetDbg() dbg.Map {
	return dat.Dbg
}

func (dat *falloutDat_1) FillDbg() {
	const (
		strHeader      = "DAT1:1:Header"
		strContent     = "Size:FilesContent"
		strContentMean = "Size:FilesContentMean"
		strAll         = ":All"
		strCount       = ":Count"
		strReal        = ":Real"
		strPacked      = ":Packed"
		strFilesCount  = "Stats:FilesCount"
	)

	var contentAdd = func(file *falloutFile_1, dir *falloutDir_1, dat *falloutDat_1) {
		var ext = path.Ext(file.Name)
		if ext == "" {
			ext = ".?"
		}

		var (
			prefixOne     = strContent + ":" + ext
			prefixMeanOne = strContentMean + ":" + ext
		)

		var update = func(dbgMap dbg.Map, prefixNorm string, prefixMean string) {
			var count, real, packed int64

			if val, ok := dbgMap[(prefixNorm + strCount)].(int64); ok {
				count = val
			}

			if val, ok := dbgMap[(prefixNorm + strReal)].(int64); ok {
				real = val
			}

			if val, ok := dbgMap[(prefixNorm + strPacked)].(int64); ok {
				packed = val
			}

			count++
			real += file.GetSizeReal()
			packed += file.GetSizePacked()

			dbgMap[(prefixNorm + strCount)] = count
			dbgMap[(prefixNorm + strReal)] = real
			dbgMap[(prefixNorm + strPacked)] = packed

			dbgMap[(prefixMean + strReal)] = real / count
			dbgMap[(prefixMean + strPacked)] = packed / count
		}

		for _, dbgMap := range []dbg.Map{dir.Dbg, dat.Dbg} {
			update(dbgMap, prefixOne, prefixMeanOne)
			update(dbgMap, (strContent + strAll), (strContentMean + strAll))
		}
		//update(dat.Dbg, prefixOne, prefixMeanOne)
		//update(dat.Dbg, (strContent + strAll), (strContentMean + strAll))
	}

	var contentCleanup = func(dbgMap dbg.Map) {
		var keys = dbgMap.Keys((strContent + ":."))
		if len(keys) == 3 {
			// there's only one file extension in directory

			// delete `Size:FilesContent:.EXT:Count`
			// duplicate of `DAT1:0:FilesCount`
			delete(dbgMap, keys[0])

			// delete `Size:FilesContent:All:*` and `Size:FilesContentMean:All:*`
			// duplicates of `Size:FilesContent:.EXT:*`
			for _, prefix := range []string{strContent, strContentMean} {
				for _, key := range dbgMap.Keys(prefix + strAll) {
					delete(dbgMap, key)
				}
			}
		} else {
			// there's at least two different file extensions in directory

			for _, keyOld := range keys {
				// `Size:FilesContent:.EXT:Count` (int64) -> `Stats:FilesCount:.EXT` (uint32)
				if strings.HasSuffix(keyOld, strCount) {
					var key = strings.Replace(keyOld, strContent, strFilesCount, 1)
					key, _ = strings.CutSuffix(key, strCount)

					dbgMap[key] = uint32(dbgMap[keyOld].(int64))
					delete(dbgMap, keyOld)
				}
			}

			// delete `Size::FilesContent:All:Count`
			// duplicate of `DAT1:0:FilesCount`
			delete(dbgMap, (strContent + strAll + strCount))
		}
	}

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

			contentAdd(file, dir, dat)

			file.Dbg["Idx:Dir"] = uint16(idxFile)
			file.Dbg["Idx:Dat"] = uint16(dat.Dbg[strContent+strAll+strCount].(int64) - 1)
		}

		contentCleanup(dir.Dbg)
	}

	contentCleanup(dat.Dbg)
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
func (dir *falloutDir_1) GetFiles() (files []FalloutFile) {
	files = make([]FalloutFile, len(dir.Files))

	for idx, file := range dir.Files {
		files[idx] = file
	}
	return files
}

func (dir *falloutDir_1) GetDbg() dbg.Map {
	return dir.Dbg
}

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

// GetOffset implements FalloutFile
func (file *falloutFile_1) GetOffset() int64 {
	return int64(file.Offset)
}

// GetCompressMode implements FalloutFile
func (file *falloutFile_1) GetCompressMode() uint32 {
	return file.CompressMode
}

// GetOffset implements FalloutFile
func (file *falloutFile_1) GetSizeReal() int64 {
	return int64(file.SizeReal)
}

// GetSizePacked returns size of file block
//
// GetSizePacked implements FalloutFile
func (file *falloutFile_1) GetSizePacked() int64 {
	if file.SizePacked == 0 {
		return file.GetSizeReal()
	}

	return int64(file.SizePacked)
}

// GetFileLZSS returns converted file data, which can be passed to `lzss.FalloutLZSS“ functions
func (file *falloutFile_1) getFileLZSS() lzss.FalloutFileLZSS {
	return lzss.FalloutFileLZSS{
		Offset:       file.Offset,
		CompressMode: file.CompressMode,
		SizeReal:     file.SizeReal,
		SizePacked:   file.SizePacked,
	}
}

// GetBytesReal implements FalloutFile
func (file *falloutFile_1) GetBytesReal(stream io.ReadSeeker) (out []byte, err error) {
	if out, err = lzss.Fallout1.DecompressFile(stream, file.getFileLZSS()); err != nil {
		return nil, err
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
	if _, err = stream.Read(out); err != nil {
		return nil, err
	}

	return out, nil
}

func (file *falloutFile_1) GetDbg() dbg.Map {
	return file.Dbg
}
