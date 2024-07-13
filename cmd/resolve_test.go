package cmd

import (
	"fmt"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"

	"github.com/wipe2238/fo/steam"
	"github.com/wipe2238/fo/x/maketest"
)

func TestResolveMap(t *testing.T) {
	must.NotNil(t, resolveMap)

	test.MapNotEmpty(t, resolveMap)
	test.MapNotContainsKey(t, resolveMap, "")
	test.MapNotContainsValue(t, resolveMap, nil)
}

//

func TestFilenameNil(t *testing.T) {
	var err = ResolveFilename(nil, "@")
	test.Error(t, err)
}

//

func TestPrefixEmpty(t *testing.T) {
	var filename = "filename"
	var filenameBefore = filename
	var err = ResolveFilename(&filename, "")
	test.Error(t, err)
	test.Eq(t, filename, filenameBefore)
}

func TestPrefixMissing(t *testing.T) {
	var filename = "filename"
	var filenameBefore = filename
	var err = ResolveFilename(&filename, "@")
	must.NoError(t, err)
	test.Eq(t, filename, filenameBefore)
}

func TestPrefixUnknown(t *testing.T) {
	var filename = "@unknown"
	var filenameBefore = filename
	var err = ResolveFilename(&filename, "@")
	test.Error(t, err)
	test.Eq(t, filename, filenameBefore)
}

//

func TestSteamInvalidFormat(t *testing.T) {
	var (
		err            error
		filename       string
		filenameBefore string
	)

	for idx := range 2 {
		var _, fallout, _, f = maketest.FalloutIdxData(idx)

		t.Run(fallout, func(t *testing.T) {

			// ResolveFilename()

			for _, filename = range []string{
				"@steamfallout" + f + ":MASTER.DAT",
				"@steam:",
				"@steam::",
				"@steam:fallout" + f + ":",
				"@steam::MASTER.DAT",
			} {
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {
					err = ResolveFilename(&filename, "@")
					test.Error(t, err)
					test.Eq(t, filename, filenameBefore)
				})
			}

			// resolveSteam()

			for _, filename = range []string{
				"@staem:fallout" + f + ":MASTER.DAT",
			} {
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {
					err = resolveSteam(&filename, "@steam")
					test.Error(t, err)
					test.Eq(t, filename, filenameBefore)
				})
			}

		})
	}
}

func TestSteamGame(t *testing.T) {
	var (
		err            error
		filename       string
		filenameBefore string
	)

	for idx := range 2 {
		var appID, fallout, fo, f = maketest.FalloutIdxData(idx)

		t.Run(fallout, func(t *testing.T) {
			// invalid
			for _, game := range []string{
				"failout" + filenameBefore,
				"fallout",
				"Fallout",
				"FALLOUT",
				" fallout" + f,
				"fallout" + f + " ",
			} {
				filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {

					err = ResolveFilename(&filename, "@")
					test.Error(t, err)
					test.Eq(t, filename, filenameBefore)
				})
			}
			// valid
			if !steam.IsSteamAppInstalled(appID) && !maketest.Must(fo) {
				t.Skipf("%s not installed", fallout)
			}

			for _, game := range []string{
				"fo" + f,
				"Fo" + f,
				"fO" + f,
				"FO" + f,
				"fallout" + f,
				"Fallout" + f,
				"FALLOUT" + f,
				"FaLlOuT" + f,
				"fAlLoUt" + f,
				"FALLout" + f,
			} {
				filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {
					err = ResolveFilename(&filename, "@")
					must.NoError(t, err)
					test.NotEq(t, filename, filenameBefore)
					test.FileExists(t, filename)
				})
			}

		})
	}
}

func TestSteamFile(t *testing.T) {
	var (
		err            error
		filename       string
		filenameBefore string
	)

	for idx := range 2 {
		var appID, fallout, fo, f = maketest.FalloutIdxData(idx)

		t.Run(fallout, func(t *testing.T) {
			// invalid

			for _, file := range []string{
				"MASTER.DED",
			} {
				filename = fmt.Sprintf("@steam:fallout%s:%s", f, file)
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {
					err = ResolveFilename(&filename, "@")
					test.Error(t, err)
					test.Eq(t, filename, filenameBefore)
				})
			}

			if !steam.IsSteamAppInstalled(appID) && !maketest.Must(fo) {
				t.Skipf("%s not installed", fallout)
			}

			// valid

			for _, filename := range []string{
				"MASTER.DAT",
				"CRITTER.DAT",
				"Readme.rtf",
				"Extras/ScrnSet.msg",
			} {
				filename = fmt.Sprintf("@steam:fallout%s:%s", f, filename)
				filenameBefore = filename
				t.Run(filename, func(t *testing.T) {
					err = ResolveFilename(&filename, "@")
					must.NoError(t, err)
					test.NotEq(t, filename, filenameBefore)
					test.FileExists(t, filename)
				})
			}
		})

	}
}
