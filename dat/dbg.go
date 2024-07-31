package dat

import (
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/wipe2238/fo/x/dbg"
)

const (
	dbgContent     = "Size:FilesContent"
	dbgContentMean = "Size:FilesContentMean"
	dbgFilesCount  = "Stats:FilesCount"
	dbgAll         = ":All"
	dbgCount       = ":Count"
	dbgReal        = ":Real"
	dbgPacked      = ":Packed"

	dbgOffsetInfo0 = "Offset:0:Info"
	dbgOffsetInfo1 = "Offset:1:Info"

	dbgHeader = "DAT1:1:Header"
)

func dbgAddContentStats(file FalloutFile) {
	var ext = strings.ToUpper(path.Ext(file.GetName()))
	if ext == "" {
		ext = ".?"
	}

	var (
		prefixOne     = dbgContent + ":" + ext
		prefixMeanOne = dbgContentMean + ":" + ext
	)

	var update = func(dbgMap dbg.Map, prefixNorm string, prefixMean string) {
		var count, sizeReal, sizePacked int64

		if val, ok := dbgMap[(prefixNorm + dbgCount)].(int64); ok {
			count = val
		}

		if val, ok := dbgMap[(prefixNorm + dbgReal)].(int64); ok {
			sizeReal = val
		}

		if val, ok := dbgMap[(prefixNorm + dbgPacked)].(int64); ok {
			sizePacked = val
		}

		count++
		sizeReal += file.GetSizeReal()
		sizePacked += file.GetSizePacked()

		dbgMap[(prefixNorm + dbgCount)] = count
		dbgMap[(prefixNorm + dbgReal)] = sizeReal
		dbgMap[(prefixNorm + dbgPacked)] = sizePacked

		dbgMap[(prefixMean + dbgReal)] = sizeReal / count
		dbgMap[(prefixMean + dbgPacked)] = sizePacked / count
	}

	for _, dbgMap := range []dbg.Map{file.GetParentDir().GetDbg(), file.GetParentDat().GetDbg()} {
		update(dbgMap, prefixOne, prefixMeanOne)
		update(dbgMap, (dbgContent + dbgAll), (dbgContentMean + dbgAll))
	}
}

func dbgCleanupContentStats(dbgMap dbg.Map) {
	var keys = dbgMap.Keys((dbgContent + ":."))
	if len(keys) == 3 {
		// there's only one file extension in directory

		// delete `Size:FilesContent:.EXT:Count`
		// duplicate of `DAT1:0:FilesCount` (FalloutDirV1)
		delete(dbgMap, keys[0])

		// delete `Size:FilesContent:All:*` and `Size:FilesContentMean:All:*`
		// duplicates of `Size:FilesContent:.EXT:*`
		for _, prefix := range []string{dbgContent, dbgContentMean} {
			for _, key := range dbgMap.Keys(prefix + dbgAll) {
				delete(dbgMap, key)
			}
		}
	} else {
		// there's at least two different file extensions in directory

		for _, keyOld := range keys {
			// `Size:FilesContent:.EXT:Count` (int64) -> `Stats:FilesCount:.EXT` (uint32)
			if strings.HasSuffix(keyOld, dbgCount) {
				var key = strings.Replace(keyOld, dbgContent, dbgFilesCount, 1)
				key, _ = strings.CutSuffix(key, dbgCount)

				dbgMap[key] = uint32(dbgMap[keyOld].(int64))
				delete(dbgMap, keyOld)
			}
		}

		// delete `Size::FilesContent:All:Count`
		// duplicate of `DAT1:0:FilesCount` (FalloutDirV1)
		// duplicate of `DAT2:0:FilesCount` (FalloutDatV2)
		delete(dbgMap, (dbgContent + dbgAll + dbgCount))
	}
}

//
// Fallout*V1
//

// GetDbg implements FalloutDat
func (dat *falloutDatV1) GetDbg() dbg.Map {
	return dat.Dbg
}

// SetDbg implements FalloutDat
func (dat *falloutDatV1) SetDbg(stream io.Seeker) (err error) {
	// TODO: dat.Dbg = make(dbg.Map)

	var streamPos int64
	if streamPos, err = stream.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("%s cannot store stream position; %w", errPackage, err)
	}

	//
	// DAT1
	//

	dat.Dbg["DAT1:0:DirsCount"] = dat.DirsCount
	dat.Dbg[dbgHeader] = dat.Header

	//
	// Offset, Size
	//

	if err = dat.Dbg.AddOffset(dbgOffsetInfo0, stream); err != nil {
		return err
	}

	// DirsCount = 4
	// Header    = 4 * 3
	if err = dat.Dbg.AddOffsetSeek("Offset:1:DirsNames", stream, 16, io.SeekCurrent); err != nil {
		return err
	}

	for _, dir := range dat.Dirs {
		if _, err = stream.Seek(int64(len(dir.Path)+1), io.SeekCurrent); err != nil {
			return err
		}
	}

	if err = dat.Dbg.AddOffset("Offset:2:DirsData", stream); err != nil {
		return err
	}

	for idx, dir := range dat.Dirs {
		if err = dir.SetDbg(stream); err != nil {
			return err
		}
		dir.Dbg["Idx"] = uint16(idx)
	}

	if err = dat.Dbg.AddOffset("Offset:3:FilesContent", stream); err != nil {
		return err
	}

	if err = dat.Dbg.AddSize("Size:Tree:Info", dbgOffsetInfo0, "Offset:1:DirsNames"); err != nil {
		return err
	}

	if err = dat.Dbg.AddSize("Size:Tree:DirsNames", "Offset:1:DirsNames", "Offset:2:DirsData"); err != nil {
		return err
	}

	if err = dat.Dbg.AddSize("Size:Tree:DirsData", "Offset:2:DirsData", "Offset:3:FilesContent"); err != nil {
		return err
	}

	if err = dat.Dbg.AddSize("Size:Tree:Total", dbgOffsetInfo0, "Offset:3:FilesContent"); err != nil {
		return err
	}

	//
	// Cleanup
	//

	dbgCleanupContentStats(dat.Dbg)

	if _, err = stream.Seek(streamPos, io.SeekStart); err != nil {
		return fmt.Errorf("%s cannot restore stream position", errPackage)
	}

	return nil
}

// GetDbg implements FalloutDir
func (dir *falloutDirV1) GetDbg() dbg.Map {
	return dir.Dbg
}

// SetDbg implements FalloutDir
func (dir *falloutDirV1) SetDbg(stream io.Seeker) (err error) {
	// TODO: dir.Dbg = make(dbg.Map)

	//
	// DAT1
	//

	dir.Dbg["DAT1:0:FilesCount"] = dir.FilesCount
	dir.Dbg[dbgHeader] = dir.Header

	//
	// Offset, Size
	//

	if err = dir.Dbg.AddOffset(dbgOffsetInfo0, stream); err != nil {
		return err
	}

	// FilesCount = 4
	// Header     = 4 * 3
	if err = dir.Dbg.AddOffsetSeek("Offset:1:Files", stream, 16, io.SeekCurrent); err != nil {
		return err
	}

	for idx, file := range dir.Files {
		if err = file.SetDbg(stream); err != nil {
			return err
		}

		file.Dbg["Idx:Dir"] = uint16(idx)
	}

	if err = dir.Dbg.AddOffset("Offset:2:End", stream); err != nil {
		return err
	}

	if err = dir.Dbg.AddSize("Size:DirEntry:Info", dbgOffsetInfo0, "Offset:1:Files"); err != nil {
		return err
	}

	if err = dir.Dbg.AddSize("Size:DirEntry:FilesData", "Offset:1:Files", "Offset:2:End"); err != nil {
		return err
	}

	if err = dir.Dbg.AddSize("Size:DirEntry:Total", dbgOffsetInfo0, "Offset:2:End"); err != nil {
		return err
	}

	//
	// Cleanup, always last
	//

	dbgCleanupContentStats(dir.Dbg)

	return nil
}

// GetDbg implements FalloutFile
func (file *falloutFileV1) GetDbg() dbg.Map {
	return file.Dbg
}

// SetDbg implements FalloutFile
func (file *falloutFileV1) SetDbg(stream io.Seeker) (err error) {
	// TODO: file.Dbg = make(dbg.Map)

	dbgAddContentStats(file)

	//
	// DAT1
	//

	file.Dbg["DAT1:0:NameLength"] = byte(len(file.Name))
	file.Dbg["DAT1:1:Name"] = file.Name
	file.Dbg["DAT1:2:PackedMode"] = file.PackedMode
	file.Dbg["DAT1:3:Offset"] = file.Offset
	file.Dbg["DAT1:4:SizeReal"] = file.SizeReal
	file.Dbg["DAT1:5:SizePacked"] = file.SizePacked

	//
	// Misc
	//

	file.Dbg["Idx:Dat"] = uint16(file.parentDir.parentDat.Dbg[(dbgContent+dbgAll+dbgCount)].(int64) - 1)

	//
	// Offset, Size
	//

	if err = file.Dbg.AddOffset("Offset:0:Name", stream); err != nil {
		return err
	}

	// NameLenght  = 1
	// Name       <- NameLength
	if err = file.Dbg.AddOffsetSeek(dbgOffsetInfo1, stream, int64((len(file.Name) + 1)), io.SeekCurrent); err != nil {
		return err
	}

	// PackedMode = 4
	// Offset     = 4
	// SizeReal   = 4
	// SizePacked = 4
	if err = file.Dbg.AddOffsetSeek("Offset:2:End", stream, 16, io.SeekCurrent); err != nil {
		return err
	}

	if err = file.Dbg.AddSize("Size:FileEntry:Name", "Offset:0:Name", dbgOffsetInfo1); err != nil {
		return err
	}

	if err = file.Dbg.AddSize("Size:FileEntry:Info", dbgOffsetInfo1, "Offset:2:End"); err != nil {
		return err
	}

	if err = file.Dbg.AddSize("Size:FileEntry:Total", "Offset:0:Name", "Offset:2:End"); err != nil {
		return err
	}

	return nil
}

//
// Fallout*V2
//

// GetDbg implements FalloutDat
func (dat *falloutDatV2) GetDbg() dbg.Map {
	return dat.Dbg
}

// SetDbg implements FalloutDat
func (dat *falloutDatV2) SetDbg(stream io.Seeker) (err error) {
	// TODO: dat.Dbg = make(dbg.Map)

	var streamPos int64
	if streamPos, err = stream.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("%s cannot store stream position", errPackage)
	}

	//
	// DAT2
	//

	dat.Dbg["DAT2:0:FilesCount"] = dat.FilesCount
	dat.Dbg["DAT2:1:SizeTree"] = dat.SizeTree
	dat.Dbg["DAT2:2:SizeDat"] = dat.SizeDat

	//
	// Offset, Size
	//

	if err = dat.Dbg.AddOffset("Offset:0:Begin", stream); err != nil {
		return err
	}

	if err = dat.Dbg.AddOffsetSeek("Offset:3:End", stream, 0, io.SeekEnd); err != nil {
		return err
	}

	if err = dat.Dbg.AddOffsetSeek("Offset:2:SizeData", stream, -8, io.SeekCurrent); err != nil {
		return err
	}

	if err = dat.Dbg.AddOffsetSeek("Offset:1:Tree", stream, -int64(dat.SizeTree), io.SeekCurrent); err != nil {
		return err
	}

	// FilesCount = 4
	if _, err = stream.Seek(4, io.SeekCurrent); err != nil {
		return err
	}

	// This implementation is a messsy one; DAT2 don't know what directory is, and there is no promise
	// files entries will be sorted in any way. Due to that, as long `file.Dbg` keeps offsets data,
	// order used in DAT2 must be restored

	var files = make([]*falloutFileV2, 0, dat.FilesCount)

	for _, dir := range dat.Dirs {
		files = append(files, dir.Files...)
	}

	slices.SortFunc(files, func(a *falloutFileV2, b *falloutFileV2) int {
		return int(a.Index) - int(b.Index)
	})

	for _, file := range files {
		file.SetDbg(stream)
	}

	for idxDir, dir := range dat.Dirs {
		dir.SetDbg(stream)
		dir.Dbg["Idx"] = uint16(idxDir)
	}

	//
	// Cleanup
	//

	dbgCleanupContentStats(dat.Dbg)

	if _, err = stream.Seek(streamPos, io.SeekStart); err != nil {
		return fmt.Errorf("%s cannot restore stream position", errPackage)
	}

	return nil
}

// GetDbg implements FalloutDir
func (dir *falloutDirV2) GetDbg() dbg.Map {
	return dir.Dbg
}

// SetDbg implements FalloutDir
func (dir *falloutDirV2) SetDbg(stream io.Seeker) error {
	dir.Dbg = make(dbg.Map)

	for idxFile, file := range dir.Files {
		file.Dbg["Idx:Dir"] = uint16(idxFile)
	}

	dbgCleanupContentStats(dir.Dbg)

	return nil
}

// GetDbg implements FalloutFile
func (file *falloutFileV2) GetDbg() dbg.Map {
	return file.Dbg
}

// SetDbg implements FalloutFile
func (file *falloutFileV2) SetDbg(stream io.Seeker) (err error) {
	// TODO: file.Dbg = make(dbg.Map)

	dbgAddContentStats(file)

	//
	// DAT2
	//

	file.Dbg["DAT2:0:PathLength"] = uint32(len(file.Path) + 1)
	file.Dbg["DAT2:1:Path"] = strings.ReplaceAll(file.Path, "/", `\`)
	file.Dbg["DAT2:2:PackedMode"] = file.PackedMode
	file.Dbg["DAT2:3:SizeReal"] = file.SizeReal
	file.Dbg["DAT2:4:SizePacked"] = file.SizePacked
	file.Dbg["DAT2:5:Offset"] = file.Offset

	//
	// Misc
	//

	file.Dbg["Idx:Dat"] = file.Index

	//
	// Offset, Size
	//

	if err = file.Dbg.AddOffset("Offset:0:Path", stream); err != nil {
		return err
	}

	// PathLength  = 4
	// Path       <- PathLength
	if err = file.Dbg.AddOffsetSeek("Offset:1:Info", stream, int64(len(file.Path)+4), io.SeekCurrent); err != nil {
		return err
	}

	// PackedMode = 1
	// SizeReal   = 4
	// SizePacked = 4
	// Offset     = 4
	if err = file.Dbg.AddOffsetSeek("Offset:2:End", stream, 13, io.SeekCurrent); err != nil {
		return err
	}

	if err = file.Dbg.AddSize("Size:FileEntry:Total", "Offset:0:Path", "Offset:2:End"); err != nil {
		return err
	}

	return nil
}
