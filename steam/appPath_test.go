package steam

import (
	"os"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"github.com/wipe2238/fo/x/maketest"
)

func DoTestSteamFallout(t *testing.T, appId uint64, ext string) {
	var err error

	// skip current test if app not found
	if !IsSteamAppInstalled(appId) || !maketest.Must(ext) {
		t.Skipf("App %d not installed", appId)
	}
	must.NoError(t, err)

	for _, filename := range []string{"./MASTER.DAT", `\CRITTER.DAT`, "/MASTER.DAT"} {
		_, err = GetAppFilePath(appId, filename)
		test.Error(t, err)
	}

	for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT", "Manual/../MASTER.DAT"} {
		var path string
		path, err = GetAppFilePath(appId, filename)
		must.NoError(t, err)

		_, err = os.Stat(path)
		must.NoError(t, err)
		test.StrNotContains(t, path, "..")
		test.FilePathValid(t, path)
		test.FileExists(t, path)
	}
}

func TestUnknownApp(t *testing.T) {
	if IsSteamInstalled() || maketest.Must("steam") {
		var _, err = GetAppPath(1207)
		test.Error(t, err)
	}
}

func TestSteamFallout1(test *testing.T) {
	DoTestSteamFallout(test, AppId.Fallout1, "fo1")
}

func TestSteamFallout2(test *testing.T) {
	DoTestSteamFallout(test, AppId.Fallout2, "fo2")
}
