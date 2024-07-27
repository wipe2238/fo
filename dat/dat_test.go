package dat

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/steam"
	"github.com/wipe2238/fo/x/dbg"
	"github.com/wipe2238/fo/x/maketest"
)

var fallout = [2]func(io.ReadSeeker) (FalloutDat, error){Fallout1, Fallout2}

// DoDat searches ALL known locations where .dat files might be present and uses callbacks to
// run tests/benchmarks on each of them
//
// Current search locations:
//   - testdata/ directory
//   - Steam installation directories
func DoDat(tb testing.TB,
	callbackDat func(testing.TB, io.ReadSeeker, FalloutDat, string),
	callbackDir func(testing.TB, io.ReadSeeker, FalloutDir, string),
	callbackFile func(testing.TB, io.ReadSeeker, FalloutFile, string),
) {
	tb.Helper()

	// testdata/
	DoRunTB(tb, "testdata", func(tb testing.TB) {
		var (
			err   error
			paths []string
		)

		for idx, ext := range []string{"dat", "dat1", "dat2"} {
			DoRunTB(tb, strings.ToUpper(ext), func(tb testing.TB) {
				paths, err = filepath.Glob(filepath.Join("testdata", "*."+ext))
				must.NoError(tb, err)

				if len(paths) < 1 {
					tb.Skipf("No .%s files found", ext)
				}

				for _, filename := range paths {
					DoRunTB(tb, filepath.Base(filename), func(tb testing.TB) {
						var osFile, dat, err = DoDatOpen(filename, idx)
						must.NoError(tb, err)
						must.NotNil(tb, osFile)
						must.NotNil(tb, dat)

						DoDatFile(tb, osFile, dat, "testdata", callbackDat, callbackDir, callbackFile)
						osFile.Close()
					})
				}
			})
		}
	})

	// Steam
	DoRunTB(tb, "Steam", func(tb testing.TB) {
		for idx := range 2 {
			var appID, fallout, fo, _ = maketest.FalloutIdxData(idx)

			DoRunTB(tb, fallout, func(tb testing.TB) {
				if !steam.IsSteamAppInstalled(appID) && !maketest.Must(fo) {
					tb.Skipf("%s not installed", fallout)
				}

				for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT"} {
					DoRunTB(tb, filename, func(tb testing.TB) {
						var (
							err          error
							filenamePath string
							osFile       *os.File
							dat          FalloutDat
						)

						filenamePath, err = steam.GetAppFilePath(appID, filename)
						must.NoError(tb, err)

						osFile, dat, err = DoDatOpen(filenamePath, (idx + 1))
						must.NoError(tb, err)
						must.NotNil(tb, osFile)
						must.NotNil(tb, dat)

						DoDatFile(tb, osFile, dat, "steam", callbackDat, callbackDir, callbackFile)
						osFile.Close()
					})
				}
			})
		}
	})
}

// DoRunTB does necessary type conversion before calling testing.T.Run() or testing.B.Run()
func DoRunTB(tb testing.TB, name string, funcTB func(testing.TB)) bool {
	tb.Helper()

	var (
		tbAsT, okT = tb.(*testing.T)
		tbAsB, okB = tb.(*testing.B)
	)

	if (!okT && !okB) || (okT && okB) {
		panic(fmt.Sprintf("Both okT and okB are %t.... which is not ok", okT))
	} else if okT {
		return tbAsT.Run(name, func(t *testing.T) { funcTB(t) })
	} else if okB {
		return tbAsB.Run(name, func(b *testing.B) { funcTB(b) })
	}

	return true
}

// DoDatOpen opens a .dat file; requires to specify from which game file comes from
//
// Note: Calling function is responsible for closing os.File
func DoDatOpen(filename string, fo int) (osFile *os.File, dat FalloutDat, err error) {
	if osFile, err = os.Open(filename); err != nil {
		return nil, nil, err
	}

	if fo < 1 || fo > 2 {
		osFile.Close()

		return nil, nil, fmt.Errorf("DoDatOpen(%s) fo < 1 || fo > 2", filename)
	}

	if dat, err = fallout[fo-1](osFile); err != nil {
		osFile.Close()

		return nil, nil, err
	}

	return osFile, dat, nil
}

// DoDatFile creates sub-tests/sub-benchmarks for each object within parsed .dat file
func DoDatFile(tb testing.TB, osFile *os.File, dat FalloutDat, source string,
	callbackDat func(testing.TB, io.ReadSeeker, FalloutDat, string),
	callbackDir func(testing.TB, io.ReadSeeker, FalloutDir, string),
	callbackFile func(testing.TB, io.ReadSeeker, FalloutFile, string),
) {
	tb.Helper()

	if callbackDat != nil {
		callbackDat(tb, osFile, dat, source)
	}

	for _, dir := range dat.GetDirs() {
		DoRunTB(tb, dir.GetPath(), func(tb testing.TB) {
			if callbackDir != nil {
				callbackDir(tb, osFile, dir, source)
			}
			for _, file := range dir.GetFiles() {
				DoRunTB(tb, file.GetName(), func(tb testing.TB) {
					if callbackFile != nil {
						callbackFile(tb, osFile, file, source)
					}
				})
			}
		})
	}
}

//------------------------------------------------------------------------------------------------//

func TestFunc(t *testing.T) {
	var err error
	var fn = [2]func(io.ReadSeeker) (FalloutDat, error){Fallout1, Fallout2}

	for idx := range 2 {
		var appID, fallout, fo, _ = maketest.FalloutIdxData(idx)

		t.Run(fallout, func(t *testing.T) {

			if !steam.IsSteamAppInstalled(appID) && !maketest.Must(fo) {
				t.Skipf("%s not installed", fallout)
			}

			for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT"} {
				t.Run(filename, func(t *testing.T) {
					var osFile *os.File

					filename, err = steam.GetAppFilePath(appID, filename)
					must.NoError(t, err)

					osFile, err = os.Open(filename)
					must.NoError(t, err)
					defer osFile.Close()

					var dat FalloutDat
					dat, err = fn[idx](osFile)
					must.NoError(t, err)

					test.EqOp(t, int(dat.GetGame()), idx+1)
				})
			}
		})
	}
}

func TestImpl(testingT *testing.T) {
	DoDat(testingT,
		func(tb testing.TB, stream io.ReadSeeker, dat FalloutDat, _ string) {
			var (
				dbg  = dat.GetDbg()
				dirs = dat.GetDirs()
				game = dat.GetGame()
			)
			test.NotNil(tb, dbg)

			test.NotNil(tb, dirs)

			must.GreaterEq(tb, 1, game)
			must.LessEq(tb, 2, game)

			stream.Seek(0, io.SeekStart)
			must.NoError(tb, dat.SetDbg(stream))
			dat.FillDbg()

			// TODO: dat.SetDbg(nil)
			// TODO: SetDbg(nil) / SetDbg(stream) mixing

			CheckDbgMap(tb, dat.GetDbg())
		},
		func(tb testing.TB, _ io.ReadSeeker, dir FalloutDir, _ string) {
			var (
				parentDat = dir.GetParentDat()

				dbg    = dir.GetDbg()
				files  = dir.GetFiles()
				name   = dir.GetName()
				parent = dir.GetParentDat()
				path   = dir.GetPath()
			)
			// always first

			must.NotNil(tb, parentDat)

			//

			test.NotNil(tb, dbg)

			test.NotNil(tb, files)

			must.StrNotEqFold(tb, name, "")
			test.StrNotContains(tb, name, `\`)
			test.StrNotContains(tb, name, "/")

			test.NotNil(tb, parent)

			must.StrNotEqFold(tb, path, "")
			must.StrNotEqFold(tb, path, "/")
			test.StrNotContains(tb, path, `\`)
			test.StrNotHasPrefix(tb, path, "/")
			test.StrNotHasSuffix(tb, path, "/")
			if path != "." {
				test.StrNotHasPrefix(tb, ".", path)
			}

			CheckDbgMap(tb, dir.GetDbg())
		},
		func(tb testing.TB, stream io.ReadSeeker, file FalloutFile, fileSource string) {
			var (
				parentDat = file.GetParentDat()
				parentDir = file.GetParentDir()

				dbg        = file.GetDbg()
				name       = file.GetName()
				offset     = file.GetOffset()
				_          = file.GetPacked()
				packedMode = file.GetPackedMode()
				path       = file.GetPath()
				sizePacked = file.GetSizePacked()
				sizeReal   = file.GetSizeReal()
			)

			tb.Logf("name = %s", name)
			tb.Logf("path = %s", path)

			// always first

			must.NotNil(tb, parentDat)
			must.NotNil(tb, parentDir)

			//

			test.NotNil(tb, dbg)

			must.StrNotEqFold(tb, name, "")
			test.StrNotContains(tb, name, `\`)
			test.StrNotContains(tb, name, "/")

			if parentDat.GetGame() == 1 {
				// HACK: dbg.Map in tests
				test.GreaterEq(tb, parentDat.GetDbg()["Offset:3:FilesContent"].(int64), offset)
			}

			if parentDat.GetGame() == 1 {
				if packedMode == lzss.FalloutCompressNone {
					test.EqOp(tb, sizeReal, sizePacked)
				} else if packedMode == lzss.FalloutCompressLZSS {
					test.Less(tb, sizeReal, sizePacked)
				} else {
					// detect lzss.FalloutCompressStore
					tb.Fatalf("Unknown packedMode 0x%X %d", packedMode, packedMode)
				}
			} else if parentDat.GetGame() == 2 {
				test.LessEq(tb, 1, packedMode)
				if packedMode == 1 {
					test.Less(tb, sizeReal, sizePacked)
				}
			}

			must.StrNotEqFold(tb, path, "")
			test.StrNotContains(tb, path, `\`)
			test.StrNotHasPrefix(tb, path, "/")
			test.StrNotHasSuffix(tb, path, "/")
			test.StrHasPrefix(tb, parentDir.GetPath(), path)
			test.EqOp(tb, path, parentDir.GetPath()+"/"+name)

			if sizePacked == 0 {
				test.EqOp(tb, sizeReal, sizePacked)
			}

			//

			CheckDbgMap(tb, dbg)
			CheckExtract(tb, stream, file, fileSource, "testdata/extracted")
			CheckExtracted(tb, stream, file, fileSource, "testdata/extracted")
			CheckUndatUI(tb, stream, file, fileSource, "testdata/undatui/data")
		},
	)
}

func CheckDbgMap(tb testing.TB, dbgMap dbg.Map) {
	DoRunTB(tb, "DbgMap", func(tb testing.TB) {
		must.NotNil(tb, dbgMap)
		test.MapNotContainsKey(tb, dbgMap, "")
		test.MapNotContainsValue(tb, dbgMap, nil)

		dbgMap.Dump("", "", func(key string, val any, left string, right string) {
			test.NotEq(tb, key, "")
			test.NotNil(tb, val)
			test.NotEq(tb, left, "")
			test.NotEq(tb, right, "")
		})

		// TODO: remove after FillDbg() -> SetDbg() transition

		for _, prefix := range []string{"Offset", "Size", "Idx"} {
			for _, keyOld := range dbgMap.Keys(prefix + "OLD:") {
				var keyNew = strings.Replace(keyOld, prefix+"OLD:", prefix+":", 1)

				tb.Logf("%s -> %s | %d -> %d", keyOld, keyNew, dbgMap[keyOld], dbgMap[keyNew])

				must.MapContainsKey(tb, dbgMap, keyNew)
				if dbgMap[keyNew] != dbgMap[keyOld] {
					tb.Fatalf("%s(%d) != %s(%d)", keyNew, dbgMap[keyNew], keyOld, dbgMap[keyOld])
					//defer os.Exit(1)
					//break
				}
				must.EqOp(tb, dbgMap[keyOld], dbgMap[keyNew])
			}
		}
	})
}

// CheckExtract saves selected list of files
//   - If local file doesn't exists, it's extacted to `<dir>/fallout<version>/<file path>/`
//   - If local file exists, nothing happens
func CheckExtract(tb testing.TB, stream io.ReadSeeker, file FalloutFile, fileSource string, dir string) {
	// check with only one source
	if fileSource != "steam" {
		return
	}

	if strings.EqualFold(filepath.Ext(file.GetPath()), ".lst") {
		if strings.EqualFold(file.GetName(), "new.lst") {
			return
		}
	} else {
		return
	}

	if !file.GetPacked() {
		tb.Fatalf("%s is not packed, cannot be used with CheckExtract", file.GetPath())
		os.Exit(1)
	}

	var extractPath = fmt.Sprintf("%s/fallout%d/%s", dir, file.GetParentDat().GetGame(), file.GetPath())
	extractPath = strings.ToLower(extractPath)

	var err error

	if _, err = os.Stat(extractPath); err != nil && errors.Is(err, fs.ErrNotExist) {
		DoRunTB(tb, "Extract", func(tb testing.TB) {
			// dat file path -> filesystem directory
			err = os.MkdirAll(filepath.Clean(extractPath), 0755)
			must.NoError(tb, err)

			var bytes []byte

			bytes, err = file.GetBytesReal(stream)
			must.NoError(tb, err)
			must.EqOp(tb, int64(len(bytes)), file.GetSizeReal())

			err = os.WriteFile(filepath.Clean((extractPath + "/Real")), bytes, 0644)
			must.NoError(tb, err)

			bytes, err = file.GetBytesPacked(stream)
			must.NoError(tb, err)
			must.EqOp(tb, int64(len(bytes)), file.GetSizePacked())

			err = os.WriteFile(filepath.Clean((extractPath + "/Packed")), bytes, 0644)
			must.NoError(tb, err)

			/*
				if packedMode := file.GetPackedMode(); file.GetParentDat().GetGame() == 1 && packedMode == lzss.FalloutCompressLZSS {
					var fileLZSS = lzss.FalloutFile{Stream: stream, Offset: file.GetOffset(), SizePacked: file.GetSizePacked(), CompressMode: packedMode}

					var blocksSize []int64
					blocksSize, err = fileLZSS.ReadBlocksSize()
					must.NoError(tb, err)

					var blocksBytes [][]byte
					blocksBytes, err = fileLZSS.ReadBlocks()
					must.NoError(tb, err)

					must.SliceLen(tb, len(blocksSize), blocksBytes)

					for idx, block := range blocksBytes {
						err = os.WriteFile(filepath.Clean(fmt.Sprintf("%s/Block.%d", extractPath, idx)), block, 0644)
						must.NoError(tb, err)
					}
				}
			*/
		})
	}
}

func TestExtracted(t *testing.T) {
	var dir = "testdata/extracted"

	var packed = make([]string, 0)
	filepath.WalkDir(filepath.Clean(dir), func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() || dirEntry.Name() != "Packed" {
			return nil
		}

		packed = append(packed, path)

		return nil
	})

	for _, filePacked := range packed {
		var testName, _ = strings.CutPrefix(filepath.ToSlash(filePacked), dir+"/")
		testName, _ = strings.CutSuffix(testName, "/Packed")

		t.Run(testName, func(t *testing.T) {
			// TODO: simplify when/if .dat editing is possible

			var (
				err          error
				fileReal     = filepath.Join(filepath.Dir(filePacked), "Real")
				bytesDatReal []byte
				isFallout1   = strings.Contains(filePacked, "fallout1")
				loader       *falloutFileLoader
			)

			if isFallout1 {
				loader = newLoader(1)
				loader.FalloutFile.(*falloutFileV1).PackedMode = lzss.FalloutCompressLZSS
			} else {
				loader = newLoader(2)
				loader.FalloutFile.(*falloutFileV2).PackedMode = 1
			}

			// both files are needed to set inner `FalloutFile.Size*`
			err = loader.loadFiles(fileReal, filePacked)
			must.NoError(t, err)

			bytesDatReal, err = loader.GetBytesReal(bytes.NewReader(loader.DataPacked))
			test.NoError(t, err)

			test.SliceEqFunc(t, loader.DataReal, bytesDatReal, func(a, b byte) bool { return a == b })
		})
	}
}

// CheckExtracted
//   - Both `Real` and `Packed` files must exist in directory
func CheckExtracted(tb testing.TB, stream io.ReadSeeker, file FalloutFile, fileSource string, dir string) {
	if fileSource == "testdata" {
		return
	}

	var (
		err           error
		extractedPath = strings.ToLower(fmt.Sprintf("%s/fallout%d/%s", dir, file.GetParentDat().GetGame(), file.GetPath()))
	)

	for _, filename := range []string{"Real", "Packed"} {
		if _, err = os.Stat(filepath.Clean(extractedPath + "/" + filename)); err != nil {
			return
		}
	}

	DoRunTB(tb, "Extracted", func(tb testing.TB) {
		var fn = [2]func(FalloutFile, io.ReadSeeker) ([]byte, error){FalloutFile.GetBytesReal, FalloutFile.GetBytesPacked}

		for idx, filename := range []string{"Real", "Packed"} {
			var bytesOs, bytesDat []byte

			bytesOs, err = os.ReadFile(filepath.Clean(extractedPath + "/" + filename))
			must.NoError(tb, err)

			bytesDat, err = fn[idx](file, stream)
			test.NoError(tb, err)

			test.SliceEqFunc(tb, bytesOs, bytesDat, func(a, b byte) bool { return a == b })
		}
	})
}

func CheckUndatUI(tb testing.TB, stream io.ReadSeeker, file FalloutFile, fileSource string, dir string) {
	if fileSource != "steam" {
		return
	}

	if file.GetParentDat().GetGame() != 1 {
		return
	}

	var (
		err         error
		undatuiPath = filepath.Clean(fmt.Sprintf("%s/%s", dir, file.GetPath()))
	)

	if _, err = os.Stat(filepath.Clean(undatuiPath)); err != nil {
		return
	}

	DoRunTB(tb, "UndatUI", func(tb testing.TB) {
		var bytesOs, bytesDat []byte

		bytesOs, err = os.ReadFile(filepath.Clean(undatuiPath))
		must.NoError(tb, err)

		bytesDat, err = file.GetBytesReal(stream)
		test.NoError(tb, err)

		if !slices.Equal(bytesOs, bytesDat) {
			tb.Fatalf("Output mismatch")
		}
	})
}
