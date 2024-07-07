package maketest

import (
	"os"
	"path/filepath"
)

// Must returns true if file "<repo dir>/maketest.<ext>" exists
func Must(ext string) bool {
	var tmp, err = os.Getwd()

	if err != nil {
		return false
	}

	// search for main repo directory, up to 10 parents
	for range 10 {
		var dir, _ = filepath.Split(tmp)
		dir = filepath.Clean(dir)

		// if workspace file is found, it's main repo dir
		_, err = os.Stat(filepath.Join(dir, "fo.code-workspace"))
		if err == nil {
			// if maketest.<ext> is found, test cannot be skipped
			_, err = os.Stat(filepath.Join(dir, "maketest."+ext))
			return err == nil
		} else {
			tmp = dir
		}
	}

	return false
}
