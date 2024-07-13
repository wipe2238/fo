package cmd

import (
	"fmt"
	"strings"

	"github.com/wipe2238/fo/steam"
)

var resolveMap = map[string]func(*string, string) error{
	"steam": resolveSteam,
}

// ResolveFilename converts pseudo-filename with given prefix to absolute path.
func ResolveFilename(filename *string, prefix string) (err error) {
	if filename == nil {
		return fmt.Errorf("ResolveFilename() nil filename")
	} else if len(prefix) < 1 {
		return fmt.Errorf("ResolveFilename() empty prefix")
	}

	// Quickly filter out regular filenames
	if !strings.HasPrefix(*filename, prefix) {
		return nil
	}

	// Find matching resolver and call it
	for id, resolver := range resolveMap {
		if !strings.HasPrefix(*filename, prefix+id) {
			continue
		}

		if err = resolver(filename, prefix+id); err != nil {
			return fmt.Errorf("ResolveFilename(%s) cannot resolve '%s' : %w", id, *filename, err)
		}

		return nil
	}

	// None of known resolvers has been called
	return fmt.Errorf("ResolveFilename() cannot resolve '%s'", *filename)
}

func resolveSteam(filename *string, prefix string) (err error) {
	var (
		fileparts = strings.Split(*filename, ":")
		appID     uint64
	)

	if len(fileparts) != 3 {
		return fmt.Errorf("invalid format")
	} else if fileparts[0] != prefix {
		return fmt.Errorf("invalid format (prefix)")
	} else if len(fileparts[1]) < 1 {
		return fmt.Errorf("invalid format (game)")
	} else if len(fileparts[2]) < 1 {
		return fmt.Errorf("invalid format (filename)")
	}

	switch strings.ToLower(fileparts[1]) {
	case "fo1", "fallout1":
		appID = steam.AppID.Fallout1
	case "fo2", "fallout2":
		appID = steam.AppID.Fallout2
	default:
		return fmt.Errorf("invalid <game> value")
	}

	var result string
	if result, err = steam.GetAppFilePath(appID, fileparts[2]); err != nil {
		return err
	}

	*filename = result

	return nil
}
