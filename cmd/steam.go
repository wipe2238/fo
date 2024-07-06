package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/wipe2238/fo/steam"
)

// SteamGameFile converts string in "<prefix>:<game>:<filename>" format to absolute path.
func SteamGameFile(filename string, prefix string) (output string, err error) {
	var (
		usage     = "usage:  " + prefix + ":<game>:<filename>"
		fileparts = strings.Split(filename, ":")
		appId     uint64
	)

	// no error if prefix is not present
	if !strings.HasPrefix(filename, prefix) {
		return filename, nil
	} else if !strings.HasPrefix(filename, prefix+":") {
		return "", fmt.Errorf(usage)
	} else if len(fileparts) != 3 {
		return "", fmt.Errorf(usage)
	}

	switch strings.ToLower(fileparts[1]) {
	case "fo1", "fallout1":
		appId = steam.AppId.Fallout1
	case "fo2", "fallout2":
		appId = steam.AppId.Fallout2
	default:
		fmt.Printf("%s unknown game '%s' allowed values: fo1/fallout1/fo2/fallout2\n", usage, fileparts[1])
		os.Exit(1)
	}

	if filename, err = steam.GetAppFilePath(appId, fileparts[2]); err != nil {
		return "", err
	}

	return filename, nil
}
