package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jleagle/steam-go/steamvdf"
)

// TODO extract into GetLibrariesPath()
func GetAppPath(appId uint64) (appPath string, err error) {
	const (
		librariesVdf string = "libraryfolders.vdf"
	)
	var (
		self           string = fmt.Sprintf("GetAppPath(%d)", appId)
		steamDir       string
		libraryFolders steamvdf.KeyValue
	)

	// get <steam path>
	if steamDir, err = GetSteamInstallPath(); err != nil {
		return "", err
	}

	// read <steam path>/steamapps/libraryfolders.vdf
	if libraryFolders, err = steamvdf.ReadFile(filepath.Join(steamDir, "steamapps", librariesVdf)); err != nil {
		return "", fmt.Errorf("%s cannot read %s", self, librariesVdf)
	}

	libraryFolders.SortChildren()

	// iterate all known libraries by index

	for idx := range libraryFolders.GetChildrenAsMap() {
		var (
			kvTemp steamvdf.KeyValue
			ok     bool
		)

		if kvTemp, ok = libraryFolders.GetChild(idx); !ok {
			continue
		}

		// extract <library path>

		if kvTemp, ok = kvTemp.GetChild("path"); !ok {
			continue
		}

		// read <library path>/steamapps/appmanifest_<appId>.acf

		var path = filepath.Clean(kvTemp.Value)
		path = filepath.Join(path, "steamapps", fmt.Sprintf("appmanifest_%d.acf", appId))

		if _, err = os.Stat(path); err != nil {
			continue
		}

		if kvTemp, err = steamvdf.ReadFile(path); err != nil {
			continue
		}

		// extract <app installdir> from appmanifest

		if kvTemp, ok = kvTemp.GetChild("installdir"); !ok {
			continue
		}

		// app path = <library path>/steamapps/common/<app installdir>
		appPath = filepath.Join(filepath.Dir(path), "common", kvTemp.Value)
		return filepath.Clean(appPath), nil

	}

	// app path not found
	return "", fmt.Errorf("%s cannot locate app directory", self)
}

// TODO test relative/path/to/file; useless for Fallouts, useful for code reusability
func GetAppFilePath(appId uint64, filename string) (gameFile string, err error) {
	var (
		self = fmt.Sprintf("GetAppFilePath(%d,%s)", appId, filename)
	)

	// prevent escaping app directory
	for _, prefix := range []string{".", "/", "\\"} {
		if strings.HasPrefix(filename, prefix) {
			return "", fmt.Errorf("%s filename cannot start with '%s'", self, prefix)
		}
	}

	if gameFile, err = GetAppPath(appId); err != nil {
		return "", err
	}

	gameFile = filepath.Join(gameFile, filename)

	if _, err = os.Stat(gameFile); err != nil {
		return "", fmt.Errorf("%s file not found", self)
	}

	return filepath.Clean(gameFile), nil
}
