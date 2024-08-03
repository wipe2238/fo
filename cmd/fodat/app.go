package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/wipe2238/fo/cmd"
)

// All top-level args must be added via `app.AddCommand(...)` (preferably one sub-command per file),
// with `GroupID` set to `app.GroupID`, which will place them near the top of usage text
var app = &cobra.Command{
	Args:    cobra.ExactArgs(1),
	GroupID: "main",
	Version: "info\n\n" + cmd.Version(),
}

func init() {
	if executable, err := os.Executable(); err == nil {
		app.Use = filepath.Base(executable)
	} else {
		app.Use = filepath.Base(os.Args[0])
	}

	// Default group for all top-level sub-commands
	app.AddGroup(&cobra.Group{
		ID:    app.GroupID,
		Title: "Main Commands:",
	})

	app.SetErrPrefix("ERROR: ")
	app.SetOut(os.Stdout)
}

func run() error {
	// Add `cobra` group for builtin sub-commands
	// Should be done right before Execute*(), which will place them at the bottom of usage text
	//
	// Ungrouped sub-commands still will be shown below that (as `Additional Commands`),
	// which might be a good place for args which are still work in progress, added by forks, etc.
	app.SetHelpCommandGroupID("cobra")
	app.SetCompletionCommandGroupID("cobra")
	app.AddGroup(&cobra.Group{
		ID:    "cobra",
		Title: "General Commands:",
	})

	return app.Execute()
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
