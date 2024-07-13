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
	DoRunTB(tb, "testdata", func(tb testing.TB) {
		var (
			err   error
			paths []string
		)

		for idx, ext := range []string{"dat", "dat1", "dat2"} {
			DoRunTB(tb, strings.ToUpper(ext), func(tb testing.TB) {
				if idx == 0 {
					tb.Skipf("Guessing DAT version not implemented yet")
				}

				paths, err = filepath.Glob(filepath.Join("testdata", "*."+strings.ToLower(ext)))
				must.NoError(tb, err)

				if len(paths) < 1 {
					tb.Skipf("No .%s files found", strings.ToLower(ext))
				}

				for _, filename := range paths {
					DoRunTB(tb, filepath.Base(filename), func(tb testing.TB) {
						// TODO: remove when DAT2 implementation starts
						if idx == 2 {
							tb.Skipf("DAT2 reading not implemented yet")
						}

						var osFile, dat, err = DoDatOpen(filename, idx)
						must.NoError(tb, err)
						must.NotNil(tb, osFile)
						must.NotNil(tb, dat)

						DoDatFile(tb, osFile, dat, callbackDat, callbackDir, callbackFile)
						osFile.Close()
					})
				}
			})
		}
	})

	// Steam
	DoRunTB(tb, "Steam", func(tb testing.TB) {
		for idx := range 2 {
			var appId, fallout, fo, _ = maketest.FalloutIdxData(idx)

			DoRunTB(tb, fallout, func(tb testing.TB) {
				if !steam.IsSteamAppInstalled(appId) && !maketest.Must(fo) {
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
						filenamePath, err = steam.GetAppFilePath(appId, filename)
						must.NoError(tb, err)

						// TODO: remove when DAT2 implementation starts
						if idx == 1 {
							tb.Skipf("DAT2 reading not implemented yet")
						}
						osFile, dat, err = DoDatOpen(filenamePath, (idx + 1))
						must.NoError(tb, err)
						must.NotNil(tb, osFile)
						must.NotNil(tb, dat)

						DoDatFile(tb, osFile, dat, callbackDat, callbackDir, callbackFile)
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

	if (!okT && !okB) || (okT && okB) {
		panic(fmt.Sprintf("Both okT and okB are %t.... which is not ok", okT))
	} else if okT {
		return tbAsT.Run(name, func(t *testing.T) { funcTB(t) })
	} else if okB {
		return tbAsB.Run(name, func(b *testing.B) { funcTB(b) })
	}

	return true
}

// DoDatOpen opens a .dat file; currently requires to specify from which game file comes from
//
// Note: Calling function is responsible for closing os.File
func DoDatOpen(filename string, fo int) (osFile *os.File, dat FalloutDat, err error) {
	var fn = [2]func(io.Reader) (FalloutDat, error){Fallout1, Fallout2}

	if osFile, err = os.Open(filename); err != nil {
		return nil, nil, err
	}

	if fo < 1 || fo > 2 {
		osFile.Close()
		return nil, nil, fmt.Errorf("DoDatOpen(%s) fo < 1 || fo > 2", filename)
	}

	if dat, err = fn[fo-1](osFile); err != nil {
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
		DoRunTB(tb, dir.GetPath(), func(tb testing.TB) {
			if callbackDir != nil {
				callbackDir(tb, osFile, dir)
			}
			for _, file := range dir.GetFiles() {
				DoRunTB(tb, file.GetName(), func(tb testing.TB) {
					if callbackFile != nil {
						callbackFile(tb, osFile, file)
					}
				})
			}
		})
	}
}

//------------------------------------------------------------------------------------------------//

func TestFunc(t *testing.T) {
	var err error
	var fn = [2]func(io.Reader) (FalloutDat, error){Fallout1, Fallout2}

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

					// TODO: remove when DAT2 implementation starts
					if idx == 1 {
						t.Skipf("DAT2 reading not implemented yet")
					}

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

func TestDbg(testingT *testing.T) {
	var dbg = func(tb testing.TB, dbgMap dbg.Map) {
		DoRunTB(tb, "DbgMap", func(tb testing.TB) {
			must.NotNil(tb, dbgMap)
			test.MapNotContainsKey(tb, dbgMap, "")

			dbgMap.Dump("", "", func(key string, val any, left string, right string) {
				test.NotEq(tb, key, "")
				test.NotNil(tb, val)
				test.NotEq(tb, left, "")
				test.NotEq(tb, right, "")
			})
		})
	}

	DoDat(testingT,
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

func TestImpl(testingT *testing.T) {
	DoDat(testingT,
		func(tb testing.TB, _ io.ReadSeeker, dat FalloutDat) {
			var (
				dbg  = dat.GetDbg()
				dirs = dat.GetDirs()
				game = dat.GetGame()
			)
			test.NotNil(tb, dbg)

			test.NotNil(tb, dirs)

			must.GreaterEq(tb, 1, game)
			must.LessEq(tb, 2, game)

		},
		func(tb testing.TB, _ io.ReadSeeker, dir FalloutDir) {
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
			test.StrNotContains(tb, name, "/")
			test.StrNotContains(tb, name, `\`)

			test.NotNil(tb, parent)

			must.StrNotEqFold(tb, path, "")
			test.StrNotContains(tb, path, `\`)
			test.StrNotHasSuffix(tb, path, "/")
			if path != "." {
				test.StrNotHasPrefix(tb, ".", path)
			}
		},
		func(tb testing.TB, _ io.ReadSeeker, file FalloutFile) {
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

			// always first

			must.NotNil(tb, parentDir)
			parentDat = parentDir.GetParentDat()
			must.NotNil(tb, parentDat)

			//

			if compressMode == lzss.FalloutCompressNone {
				test.EqOp(tb, sizeReal, sizePacked)
			} else if compressMode == lzss.FalloutCompressLZSS {
				test.Less(tb, sizeReal, sizePacked)
			} else {
				// detect lzss.FalloutCompressStore
				tb.Fatalf("Unknown compressMode 0x%X %d", compressMode, compressMode)
			}

			//

			test.NotNil(tb, dbg)

			must.StrNotEqFold(tb, name, "")
			test.StrNotContains(tb, name, "/")
			test.StrNotContains(tb, name, `\`)

			// HACK: dbg.Map in tests
			test.GreaterEq(tb, parentDat.GetDbg()["Offset:3:FilesContent"].(int64), offset)

			must.StrNotEqFold(tb, path, "")
			test.StrNotContains(tb, path, `\`)
			test.StrHasPrefix(tb, parentDir.GetPath(), path)
			test.EqOp(tb, path, parentDir.GetPath()+"/"+name)

			if sizePacked == 0 {
				test.EqOp(tb, sizeReal, sizePacked)
			}
		},
	)
}
