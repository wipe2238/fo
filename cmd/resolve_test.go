package cmd

import (
	"fmt"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"

	"github.com/wipe2238/fo/steam"
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
	t.Logf("ResolveFilename: '%s'", filename)
	var err = ResolveFilename(&filename, "")
	test.Error(t, err)
	test.Eq(t, filename, "filename")
}

func TestPrefixMissing(t *testing.T) {
	var filename = "filename"
	t.Logf("ResolveFilename: '%s'", filename)
	var err = ResolveFilename(&filename, "@")
	test.NoError(t, err)
	test.Eq(t, filename, "filename")
}

func TestPrefixUnknown(t *testing.T) {
	var filename = "@unknown"
	t.Logf("ResolveFilename: '%s'", filename)
	var err = ResolveFilename(&filename, "@")
	test.Error(t, err)
	test.Eq(t, filename, "@unknown")
}

//

func SkipIfAppNotInstalled(t *testing.T, app uint64, canSkip bool) {
	var _, err = steam.GetAppPath(steam.AppId.Fallout1)
	if err != nil && canSkip {
		t.Skipf("Fallout1 not installed")
	}
	must.NoError(t, err)
}

func TestSteamInvalidFormat(t *testing.T) {
	var (
		err            error
		canSkip        = true // TODO opt-in false
		filename       string
		filenameBefore string
	)

	SkipIfAppNotInstalled(t, steam.AppId.Fallout1, canSkip)

	// ResolveFilename()

	for _, filename = range []string{
		"@steamfallout1:MASTER.DAT",
		"@steam:",
		"@steam::",
		"@steam:fallout1:",
		"@steam::MASTER.DAT",
	} {
		t.Logf("ResolveFilename: '%s'", filename)
		filenameBefore = filename

		err = ResolveFilename(&filename, "@")
		test.Error(t, err)
		test.Eq(t, filename, filenameBefore)
	}

	// resolveSteam()

	for _, filename = range []string{
		"@staem:fallout1:MASTER.DAT",
	} {
		var filenameBefore = filename
		t.Logf("ResolveFilename: '%s'", filename)

		err = resolveSteam(&filename, "@steam")
		test.Error(t, err)
		test.Eq(t, filename, filenameBefore)
	}

}

func TestSteamGame(t *testing.T) {
	var (
		err            error
		canSkip        = true // TODO opt-in false
		filename       string
		filenameBefore string
	)

	SkipIfAppNotInstalled(t, steam.AppId.Fallout1, canSkip)

	// invalid

	for _, game := range []string{
		"failout1",
		"fallout",
		"Fallout",
		"FALLOUT",
		" fallout1",
		"fallout2 ",
	} {
		filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
		t.Logf("ResolveFilename: '%s'", filename)
		filenameBefore = filename

		err = ResolveFilename(&filename, "@")
		test.Error(t, err)
		test.Eq(t, filename, filenameBefore)
	}

	// valid

	for _, game := range []string{"fo1", "Fo1", "FO1", "fallout1", "Fallout1", "FALLOUT1", "FaLlOuT1", "fAlLoUt1"} {
		filename = fmt.Sprintf("@steam:%s:MASTER.DAT", game)
		t.Logf("ResolveFilename: '%s'", filename)
		filenameBefore = filename

		err = ResolveFilename(&filename, "@")
		test.NoError(t, err)
		test.NotEq(t, filename, filenameBefore)
		test.FilePathValid(t, filename)
		test.FileExists(t, filename)
	}
}

func TestSteamFile(t *testing.T) {
	var (
		err            error
		canSkip        = true // TODO opt-in false
		filename       string
		filenameBefore string
	)

	SkipIfAppNotInstalled(t, steam.AppId.Fallout1, canSkip)

	// invalid

	for _, file := range []string{
		"MASTER.DED",
	} {
		filename = fmt.Sprintf("@steam:fallout1:%s", file)
		t.Logf("ResolveFilename: '%s'", filename)
		filenameBefore = filename

		err = ResolveFilename(&filename, "@")
		test.Error(t, err)
		test.Eq(t, filename, filenameBefore)
	}

	// valid

	for _, filename := range []string{
		"MASTER.DAT",
		"CRITTER.DAT",
		"FALLOUTW.EXE",
		"falloutwHR.exe",
		"Manual/MANUAL.PDF",
		"Manual/fallout_refcard.pdf",
	} {
		filename = fmt.Sprintf("@steam:fallout1:%s", filename)
		t.Logf("ResolveFilename: '%s'", filename)
		filenameBefore = filename

		err = ResolveFilename(&filename, "@")
		test.NoError(t, err)
		test.NotEq(t, filename, filenameBefore)
		test.FilePathValid(t, filename)
		test.FileExists(t, filename)
	}
}
