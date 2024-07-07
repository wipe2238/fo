package steam

// IsSteamInstalled returns true if Steam directory can be found without errors
func IsSteamInstalled() bool {
	var _, err = GetSteamInstallPath()

	return err == nil
}

// IsSteamAppInstalled returns true if app directory can be found without errors
func IsSteamAppInstalled(appId uint64) bool {
	var _, err = GetAppPath(appId)

	return err == nil
}
