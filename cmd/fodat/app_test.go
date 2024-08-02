package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const falldemo = "../../dat/testdata/falldemo.dat1"

func appExec(mute bool, name string, args ...string) (err error) {
	func(...any) {}(appExecMute, appExecLoud)

	var cmdRun *cobra.Command

	for _, cmdFind := range app.Commands() {
		if cmdFind.Name() == name {
			cmdRun = cmdFind
			break
		}
	}

	if cmdRun == nil {
		return fmt.Errorf("test: command '%s' not found", name)
	}

	var oldOut = app.OutOrStdout()
	var oldErr = app.ErrOrStderr()

	if mute {
		app.SetOut(io.Discard)
		app.SetErr(io.Discard)
	}

	var osArgs = os.Args

	os.Args = slices.Insert(args, 0, "APP", name)

	fmt.Printf("Execute: %s\n", strings.Join(os.Args, " "))
	err = app.Execute()

	os.Args = osArgs

	if mute {
		app.SetOut(oldOut)
		app.SetErr(oldErr)
	}

	return err
}

func appExecMute(name string, args ...string) (err error) {
	return appExec(true, name, args...)
}

func appExecLoud(name string, args ...string) (err error) {
	return appExec(false, name, args...)
}
