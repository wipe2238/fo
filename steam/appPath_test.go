package steam

import (
	"os"
	"testing"

	"github.com/shoenig/test/must"
)

func DoTestSteamFallout(t *testing.T, appId uint64) {
	var (
		err     error
		canSkip = true
	)

	// skip current test if steam not found
	if _, err = GetSteamInstallPath(); err != nil && canSkip {
		t.Skipf("Steam not installed")
	}
	must.NoError(t, err)

	// skip current test if app not found
	if _, err = GetAppPath(appId); err != nil && canSkip {
		t.Skipf("App %d not installed", appId)
	}
	must.NoError(t, err)

	for _, filename := range []string{"MASTER.DAT", "CRITTER.DAT"} {
		var path string
		path, err = GetAppFilePath(appId, filename)
		must.NoError(t, err)

		_, err = os.Stat(path)
		must.NoError(t, err)
	}
}

func TestSteamFallout1(test *testing.T) {
	DoTestSteamFallout(test, AppId.Fallout1)
}

func TestSteamFallout2(test *testing.T) {
	DoTestSteamFallout(test, AppId.Fallout2)
}
