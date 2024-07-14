package dat

import (
	"path"
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
)

func dbgAddContentStats(file FalloutFile, dir FalloutDir, dat FalloutDat) {
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

	for _, dbgMap := range []dbg.Map{dir.GetDbg(), dat.GetDbg()} {
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
