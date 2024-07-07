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
	t.Logf("ResolveFilename: <nil>")

	var err = ResolveFilename(nil, "@")
	test.Error(t, err)
}

//

func TestPrefixEmpty(t *testing.T) {
	var filename = "filename"
	var filenameBefore = filename
	t.Logf("ResolveFilename: '%s'", filename)

	var err = ResolveFilename(&filename, "")
	test.Error(t, err)
	test.Eq(t, filename, filenameBefore)
}

func TestPrefixMissing(t *testing.T) {
	var filename = "filename"
	var filenameBefore = filename
	t.Logf("ResolveFilename: '%s'", filename)

	var err = ResolveFilename(&filename, "@")
	must.NoError(t, err)
	test.Eq(t, filename, filenameBefore)
}

func TestPrefixUnknown(t *testing.T) {
	var filename = "@unknown"
	var filenameBefore = filename
	t.Logf("ResolveFilename: '%s'", filename)

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
		var part = fmt.Sprintf("%d", idx+1)

		// ResolveFilename()

		for _, filename = range []string{
			"@steamfallout" + part + ":MASTER.DAT",
			"@steam:",
			"@steam::",
			"@steam:fallout" + part + ":",
			"@steam::MASTER.DAT",
		} {
			filenameBefore = filename
			t.Logf("ResolveFilename: '%s'", filename)

			err = ResolveFilename(&filename, "@")
			test.Error(t, err)
			test.Eq(t, filename, filenameBefore)
		}

		// resolveSteam()

		for _, filename = range []string{
			"@staem:fallout" + part + ":MASTER.DAT",
		} {
			filenameBefore = filename
			t.Logf("ResolveFilename: '%s'", filename)

			err = resolveSteam(&filename, "@steam")
			test.Error(t, err)
			test.Eq(t, filename, filenameBefore)
		}
	}
}

func TestSteamGame(t *testing.T) {
	var (
		err            error
		filename       string
		filenameBefore string
	)

	for idx := range 2 {
		var part = fmt.Sprintf("%d", idx+1)
		var appId = steam.AppId.Fallout1 + uint64((idx * 10))

		// invalid

		for _, game := range []string{
			"failout" + part,
			"fallout",
			"Fallout",
			"FALLOUT",
			" fallout" + part,
			"fallout" + part + " ",
		} {
			filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
			filenameBefore = filename
			t.Logf("ResolveFilename: '%s'", filename)

			err = ResolveFilename(&filename, "@")
			test.Error(t, err)
			test.Eq(t, filename, filenameBefore)
		}

		// valid

		if steam.IsSteamAppInstalled(appId) || maketest.Must("fo"+part) {
			for _, game := range []string{
				"fo" + part,
				"Fo" + part,
				"FO" + part,
				"fallout" + part,
				"Fallout" + part,
				"FALLOUT" + part,
				"FaLlOuT" + part,
				"fAlLoUt" + part,
			} {
				filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
				filenameBefore = filename
				t.Logf("ResolveFilename: '%s'", filename)

				err = ResolveFilename(&filename, "@")
				must.NoError(t, err)
				test.NotEq(t, filename, filenameBefore)
				test.FilePathValid(t, filename)
				test.FileExists(t, filename)
			}
		}
	}
}

func TestSteamFile(t *testing.T) {
	var (
		err            error
		filename       string
		filenameBefore string
	)

	for idx := range 2 {

		// invalid

		for _, file := range []string{
			"MASTER.DED",
		} {
			filename = fmt.Sprintf("@steam:fallout%d:%s", idx+1, file)
			filenameBefore = filename
			t.Logf("ResolveFilename: '%s'", filename)

			err = ResolveFilename(&filename, "@")
			test.Error(t, err)
			test.Eq(t, filename, filenameBefore)
		}
	}

	if steam.IsSteamAppInstalled(steam.AppId.Fallout1) || maketest.Must("fo1") {

		// valid

		for _, filename := range []string{
			"MASTER.DAT",
			"CRITTER.DAT",
			"Manual/MANUAL.PDF",
			"Manual/fallout_refcard.pdf",
		} {
			filename = fmt.Sprintf("@steam:fallout%d:%s", 1, filename)
			filenameBefore = filename
			t.Logf("ResolveFilename: '%s'", filename)

			err = ResolveFilename(&filename, "@")
			must.NoError(t, err)
			test.NotEq(t, filename, filenameBefore)
			test.FilePathValid(t, filename)
			test.FileExists(t, filename)
		}
	}
}
