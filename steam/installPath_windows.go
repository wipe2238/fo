package steam

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func registryPaths() []string {
	return []string{
		// HKLM
		`SOFTWARE\Wow6432Node\Valve\Steam\InstallPath`,
		`SOFTWARE\Valve\Steam\InstallPath`,
	}
}

func GetSteamInstallPath() (output string, err error) {
	for _, registryPath := range registryPaths() {
		var (
			key     registry.Key
			valtype uint32
			path    string
		)

		if key, err = registry.OpenKey(registry.LOCAL_MACHINE, filepath.Dir(registryPath), registry.QUERY_VALUE); err != nil {
			continue
		}
		defer key.Close()

		if path, valtype, err = key.GetStringValue(filepath.Base(registryPath)); err != nil {
			continue
		}

		if valtype == registry.EXPAND_SZ {
			if path, err = registry.ExpandString(path); err != nil {
				continue
			}
		}

		path = filepath.Clean(path)

		// double check if it really is valid install directory
		for _, steamFile := range []string{"Steam.exe", "GameOverlayUI.exe"} {
			if _, err = os.Stat(filepath.Join(path, steamFile)); err != nil {
				continue
			}
		}

		return path, nil

	}

	return "", fmt.Errorf("GetSteamInstallPath() Steam not found")
}
