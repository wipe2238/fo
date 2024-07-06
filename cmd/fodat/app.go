package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

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

	// app should have at least one group, whe
	// which will separate cobra's `completion` and `help`
	app.AddGroup(&cobra.Group{
		ID:    app.GroupID,
		Title: "Main Commands:",
	})

	app.SetErrPrefix("ERROR: ")

}

func main() {
	// Add `cobra` group for builtin args
	// Should be right before Execute*(), which will place them at the bottom of usage text
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
