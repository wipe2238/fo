package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

const falldemo = "../../dat/testdata/falldemo.dat1"

func appExec(mute bool, args ...string) (err error) {
	func(...any) {}(appExecMute, appExecLoud)

	var oldOut = app.OutOrStdout()
	var oldErr = app.ErrOrStderr()

	if mute {
		app.SetOut(io.Discard)
		app.SetErr(io.Discard)
	}

	var osArgs = os.Args

	os.Args = slices.Insert(args, 0, "APP")

	fmt.Printf("Execute: %s\n", strings.Join(os.Args, " "))
	err = run()

	os.Args = osArgs

	if mute {
		app.SetOut(oldOut)
		app.SetErr(oldErr)
	}

	return err
}

func appExecMute(args ...string) (err error) {
	return appExec(true, args...)
}

func appExecLoud(args ...string) (err error) {
	return appExec(false, args...)
}

func TestApp(t *testing.T) {
	//test.Error(t, appExecMute("invalid-sub-command"))
}
