package steam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jleagle/steam-go/steamvdf"
)

// GetLibrariesPaths returns list of all valid Steam libraries
func GetLibrariesPaths() (output []string, err error) {
	const (
		self                = "GetLibrariesPaths()"
		librariesVdf string = "libraryfolders.vdf"
	)
	var (
		steamDir       string
		libraryFolders steamvdf.KeyValue
	)

	output = make([]string, 0)

	// get <steam path>
	if steamDir, err = GetSteamInstallPath(); err != nil {
		return nil, err
	}

	// read <steam path>/steamapps/libraryfolders.vdf
	if libraryFolders, err = steamvdf.ReadFile(filepath.Join(steamDir, "steamapps", librariesVdf)); err != nil {
		return nil, fmt.Errorf("%s cannot read %s", self, librariesVdf)
	}

	libraryFolders.SortChildren()

	// iterate all known libraries by index

	for idxString := range libraryFolders.GetChildrenAsMap() {
		var (
			kvTemp steamvdf.KeyValue
			ok     bool
		)

		if kvTemp, ok = libraryFolders.GetChild(idxString); !ok {
			continue
		}

		// extract <library path>

		if kvTemp, ok = kvTemp.GetChild("path"); !ok {
			continue
		}

		var path = filepath.Clean(kvTemp.Value)

		// validate

		var fileInfo os.FileInfo
		if fileInfo, err = os.Stat(path); err != nil {
			continue
		}

		if !fileInfo.Mode().IsDir() {
			continue
		}

		output = append(output, path)
	}

	if len(output) < 1 {
		return nil, fmt.Errorf("%s cannot find valid Steam library", self)
	}

	return output, nil
}
