package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jleagle/steam-go/steamvdf"
)

func GetAppPath(appID uint64) (appPath string, err error) {
	var (
		self           = fmt.Sprintf("GetAppPath(%d)", appID)
		librariesPaths []string
	)

	if librariesPaths, err = GetLibrariesPaths(); err != nil {
		return "", err
	}

	for _, libraryPath := range librariesPaths {
		var (
			kvTemp steamvdf.KeyValue
			ok     bool
		)

		// read <library path>/steamapps/appmanifest_<appID>.acf

		var path = filepath.Join(libraryPath, "steamapps", fmt.Sprintf("appmanifest_%d.acf", appID))

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

func GetAppFilePath(appID uint64, filename string) (gameFile string, err error) {
	var (
		self = fmt.Sprintf("GetAppFilePath(%d,%s)", appID, filename)
	)

	// prevent escaping app directory
	for _, prefix := range []string{".", "/", "\\"} {
		if strings.HasPrefix(filename, prefix) {
			return "", fmt.Errorf("%s filename cannot start with '%s'", self, prefix)
		}
	}

	if gameFile, err = GetAppPath(appID); err != nil {
		return "", err
	}

	gameFile = filepath.Join(gameFile, filename)

	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(gameFile); err != nil {
		return "", fmt.Errorf("%s file not found", self)
	}

	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("%s not a regular file", self)
	}

	return filepath.Clean(gameFile), nil
}
