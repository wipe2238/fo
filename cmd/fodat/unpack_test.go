package main

import (
	"os"
	"testing"

	"github.com/shoenig/test"
)

func TestAppUnpack(t *testing.T) {
	const dir = "../../bin/test.unpack"
	const file = "ART/BACKGRND/BACKGRND.LST"

	test.Error(t, appExecMute("unpack"))
	test.Error(t, appExecMute("unpack", falldemo))
	test.Error(t, appExecMute("unpack", falldemo, dir))
	test.Error(t, appExecMute("unpack", falldemo, dir, "file/missing"))

	test.NoError(t, appExecMute("unpack", falldemo, dir, file))
	test.FileExists(t, (dir + "/" + file))

	os.RemoveAll(dir)
}
