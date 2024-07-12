package dat

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"

	"github.com/wipe2238/fo/compress/lzss"
	"github.com/wipe2238/fo/steam"
	"github.com/wipe2238/fo/x/dbg"
	"github.com/wipe2238/fo/x/maketest"
)

// DoDat searchs ALL known locations where .dat files might bepresent and uses callbacks to
// run tests/benchmarks on each of them
//
// Current search locations:
//   - testdata/ directory (WIP)
//   - Steam installation directories
func DoDat(tb testing.TB,
	callbackDat func(testing.TB, io.ReadSeeker, FalloutDat),
	callbackDir func(testing.TB, io.ReadSeeker, FalloutDir),
	callbackFile func(testing.TB, io.ReadSeeker, FalloutFile),
) {
	// testdata/
	DoRunTB(tb, "testdata", func(tb0 testing.TB) {
		var (
			err   error
			paths []string
		)

		for idx, ext := range []string{"dat", "dat1", "dat2"} {
			DoRunTB(tb0, strings.ToUpper(ext), func(tb1 testing.TB) {
				if idx == 0 {
					tb1.Skipf("Guessing DAT version not implemented yet")
				}

				paths, err = filepath.Glob(filepath.Join("testdata", "*."+strings.ToLower(ext)))
				must.NoError(tb1, err)

				if len(paths) < 1 {
					tb1.Skipf("No .%s files found", strings.ToLower(ext))
				}

				for _, filename := range paths {
					DoRunTB(tb1, filepath.Base(filename), func(tb2 testing.TB) {
						var osFile, dat, err = DoDatOpen(filename, idx)
						must.NoError(tb2, err)
						must.NotNil(tb2, osFile)
						must.NotNil(tb2, dat)

						DoDatFile(tb2, osFile, dat, callbackDat, callbackDir, callbackFile)
						osFile.Close()
					})
				}
			})
		}
	})

	// Steam
	DoRunTB(tb, "Steam", func(tb0 testing.TB) {
		for idx := range 2 {
			var appId, fallout, fo, _ = maketest.FalloutIdxData(idx)

			DoRunTB(tb0, fallout, func(tb1 testing.TB) {
				if !steam.IsSteamAppInstalled(appId) && !maketest.Must(fo) {
					tb1.Skipf("%s not installed", fallout)
				}

				for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT"} {
					DoRunTB(tb1, filename, func(tb2 testing.TB) {
						var (
							err          error
							filenamePath string
							osFile       *os.File
							dat          FalloutDat
						)
						filenamePath, err = steam.GetAppFilePath(appId, filename)
						must.NoError(tb2, err)

						osFile, dat, err = DoDatOpen(filenamePath, (idx + 1))
						must.NoError(tb2, err)
						must.NotNil(tb2, osFile)
						must.NotNil(tb2, dat)

						DoDatFile(tb2, osFile, dat, callbackDat, callbackDir, callbackFile)
						osFile.Close()
					})
				}
			})
		}
	})
}

// DoRunTB does necessary type conversion before calling testing.T.Run() or testing.B.Run()
func DoRunTB(tb testing.TB, name string, funcTB func(testing.TB)) bool {
	var (
		tbAsT, okT = tb.(*testing.T)
		tbAsB, okB = tb.(*testing.B)
	)

	if !okT && !okB {
		panic("Both okT and okB are false.... which is not ok")
	} else if okT && okB {
		panic("Both okT and okB are true.... which is not ok")
	} else if okT {
		return tbAsT.Run(name, func(t *testing.T) { funcTB(t) })
	} else if okB {
		return tbAsB.Run(name, func(b *testing.B) { funcTB(b) })
	}

	return true
}

// DoDatOpen opens a .dat file; currently requires to specify from which game file comes from
//
// Note: Calling functions is responsible for closing returned *os.File
func DoDatOpen(filename string, fo int) (osFile *os.File, dat FalloutDat, err error) {
	var fn = [2]func(io.Reader, *FalloutDatOptions) (FalloutDat, error){Fallout1, Fallout2}

	if osFile, err = os.Open(filename); err != nil {
		return nil, nil, err
	}

	if fo < 1 || fo > 2 {
		osFile.Close()
		return nil, nil, fmt.Errorf("DoDatOpen(%s) fo < 1 || fo > 2", filename)
	}

	if dat, err = fn[fo-1](osFile, DefaultOptions()); err != nil {
		osFile.Close()
		return nil, nil, err
	}

	return osFile, dat, nil
}

// DoDatFile creates sub-tests/sub-benchmarks for each object within parsed .dat file
func DoDatFile(tb testing.TB, osFile *os.File, dat FalloutDat,
	callbackDat func(testing.TB, io.ReadSeeker, FalloutDat),
	callbackDir func(testing.TB, io.ReadSeeker, FalloutDir),
	callbackFile func(testing.TB, io.ReadSeeker, FalloutFile),
) {
	if callbackDat != nil {
		callbackDat(tb, osFile, dat)
	}

	for _, dir := range dat.GetDirs() {
		DoRunTB(tb, dir.GetPath(), func(tb1 testing.TB) {
			if callbackDir != nil {
				callbackDir(tb1, osFile, dir)
			}
			for _, file := range dir.GetFiles() {
				DoRunTB(tb1, file.GetName(), func(tb2 testing.TB) {
					if callbackFile != nil {
						callbackFile(tb2, osFile, file)
					}
				})
			}
		})
	}
}

//------------------------------------------------------------------------------------------------//

func TestFunc(t *testing.T) {
	var err error
	var fn = [2]func(io.Reader, *FalloutDatOptions) (FalloutDat, error){Fallout1, Fallout2}

	for idx := range 2 {
		var appId, fallout, fo, _ = maketest.FalloutIdxData(idx)

		t.Run(fallout, func(t *testing.T) {

			if !steam.IsSteamAppInstalled(appId) && !maketest.Must(fo) {
				t.Skipf("%s not installed", fallout)
			}

			for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT"} {
				t.Run(filename, func(t *testing.T) {
					var osFile *os.File

					filename, err = steam.GetAppFilePath(appId, filename)
					must.NoError(t, err)

					osFile, err = os.Open(filename)
					must.NoError(t, err)
					defer osFile.Close()

					var dat FalloutDat
					dat, err = fn[idx](osFile, DefaultOptions())
					must.NoError(t, err)

					test.Eq(t, int(dat.GetGame()), idx+1)
				})
			}
		})
	}
}

func TestDbg(t *testing.T) {
	var dbg = func(tb testing.TB, dbgMap dbg.Map) {
		DoRunTB(tb, "DbgMap", func(tb0 testing.TB) {
			must.NotNil(tb0, dbgMap)
			test.MapNotContainsKey(tb0, dbgMap, "")

			dbgMap.Dump("", "", func(key string, val any, left string, right string) {
				test.NotEq(tb0, key, "")
				test.NotNil(tb0, val)
				test.NotEq(tb0, left, "")
				test.NotEq(tb0, right, "")
			})
		})
	}

	DoDat(t,
		func(tb testing.TB, _ io.ReadSeeker, dat FalloutDat) {
			dat.FillDbg()
			dbg(tb, dat.GetDbg())
		},
		func(tb testing.TB, _ io.ReadSeeker, dir FalloutDir) {
			dbg(tb, dir.GetDbg())
		},
		func(tb testing.TB, _ io.ReadSeeker, file FalloutFile) {
			dbg(tb, file.GetDbg())
		},
	)
}

func TestImpl(t *testing.T) {
	DoDat(t,
		func(t testing.TB, _ io.ReadSeeker, dat FalloutDat) {
			var (
				dbg  = dat.GetDbg()
				dirs = dat.GetDirs()
				game = dat.GetGame()
			)
			test.NotNil(t, dbg)

			test.NotNil(t, dirs)

			must.GreaterEq(t, 1, game)
			must.LessEq(t, 2, game)

		},
		func(t testing.TB, _ io.ReadSeeker, dir FalloutDir) {
			var (
				dbg    = dir.GetDbg()
				files  = dir.GetFiles()
				name   = dir.GetName()
				parent = dir.GetParentDat()
				path   = dir.GetPath()
			)
			test.NotNil(t, dbg)

			test.NotNil(t, files)

			must.StrNotEqFold(t, name, "")
			test.StrNotContains(t, name, "/")
			test.StrNotContains(t, name, `\`)

			test.NotNil(t, parent)

			must.StrNotEqFold(t, path, "")
			test.StrNotContains(t, path, `\`)
			test.StrNotHasSuffix(t, path, "/")
			if path != "." {
				test.StrNotHasPrefix(t, ".", path)
			}
		},
		func(t testing.TB, _ io.ReadSeeker, file FalloutFile) {
			var (
				parentDat FalloutDat
				parentDir = file.GetParentDir()

				compressMode = file.GetCompressMode()
				dbg          = file.GetDbg()
				name         = file.GetName()
				offset       = file.GetOffset()
				path         = file.GetPath()
				sizePacked   = file.GetSizePacked()
				sizeReal     = file.GetSizeReal()
			)

			if compressMode == lzss.FalloutCompressNone {
				test.Eq(t, sizeReal, sizePacked)
			} else if compressMode == lzss.FalloutCompressLZSS {
				test.Less(t, sizeReal, sizePacked)
			} else {
				// detect lzss.FalloutCompressStore
				t.Fatalf("Unknown compressMode 0x%X %d", compressMode, compressMode)
			}

			// always first

			must.NotNil(t, parentDir)
			parentDat = parentDir.GetParentDat()
			must.NotNil(t, parentDat)

			//

			test.NotNil(t, dbg)

			must.StrNotEqFold(t, name, "")
			test.StrNotContains(t, name, "/")
			test.StrNotContains(t, name, `\`)

			// HACK dbg.Map in tests
			var minOffset = parentDat.GetDbg()["Offset:3:FilesContent"].(int64)
			test.GreaterEq(t, minOffset, offset)

			must.StrNotEqFold(t, path, "")
			test.StrNotContains(t, path, `\`)
			test.StrHasPrefix(t, parentDir.GetPath(), path)
			test.Eq(t, path, parentDir.GetPath()+"/"+name)

			if sizePacked == 0 {
				test.EqOp(t, sizeReal, sizePacked)
			}

		},
	)
}
