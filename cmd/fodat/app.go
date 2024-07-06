package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// All top-level args must be added via app.AddCommand(...) (preferably one arg per file),
// with `GroupID` set to `app.GroupID`, which will place them near the top of usage text
var app = &cobra.Command{
	Args:    cobra.ExactArgs(1),
	GroupID: "main",
}

func init() {
	if executable, err := os.Executable(); err == nil {
		app.Use = filepath.Base(executable)
	} else {
		app.Use = filepath.Base(os.Args[0])
	}

	// Default group for all top-level args
	app.AddGroup(&cobra.Group{
		ID:    app.GroupID,
		Title: "Main Commands:",
	})

	app.SetErrPrefix("ERROR: ")

}

func main() {
	// Add `cobra` group for builtin args
	// Should be done right before Execute*(), which will place them at the bottom of usage text
	//
	// Ungrouped args still will be shown below that, in `Additional Commands`, which might be a
	// good place for args which are still work in progress, or those added by forks (if any)
	app.SetHelpCommandGroupID("cobra")
	app.SetCompletionCommandGroupID("cobra")
	app.AddGroup(&cobra.Group{
		ID:    "cobra",
		Title: "General Commands:",
	})

	if err := app.Execute(); err != nil {
		// fmt.Printf("ERR: %#v\n", err)
		os.Exit(1)
	}
}
