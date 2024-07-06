package steam

import "fmt"

// TODO linux steam mess
func GetSteamInstallPath() (output string, err error) {
	// From quick research, it seems locating Steam on Linux is a TOTAL. MESS.
	// And by mess i mean it needs a touch of someone actually using Steam on Linux,
	// as it's hilarious disaster

	return "", fmt.Errorf("GetSteamInstallPath() not implemented for current platform")
}
