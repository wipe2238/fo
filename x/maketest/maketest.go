package maketest

import (
	"fmt"
	"os"
	"path/filepath"
)

func RepoDir() (string, bool) {
	var tmp, err = os.Getwd()

	if err != nil {
		return "", false
	}

	// search for main repo directory, up to 10 parents
	for range 10 {
		var dir, _ = filepath.Split(tmp)
		dir = filepath.Clean(dir)

		// if workspace file is found, it's main repo dir
		_, err = os.Stat(filepath.Join(dir, "fo.code-workspace"))
		if err == nil {
			return dir, true
		}

		tmp = dir
	}

	return "", false
}

// Must returns true if file "<repo dir>/maketest.<ext>" exists
func Must(ext string) bool {
	if dir, found := RepoDir(); found {
		var _, err = os.Stat(filepath.Join(dir, "maketest."+ext))
		return err == nil
	}

	return false
}

func FalloutIdxData(idx int) (appID uint64, nameLong string, nameShort string, n string) {
	appID = 38400 + uint64((idx * 10))
	nameLong = fmt.Sprintf("Fallout%d", idx+1)
	nameShort = fmt.Sprintf("fo%d", idx+1)
	n = fmt.Sprintf("%d", idx+1)

	return appID, nameLong, nameShort, n
}
