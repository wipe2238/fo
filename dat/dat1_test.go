package dat

import (
	"testing"

	"github.com/shoenig/test"
)

func TestStruct(t *testing.T) {
	var dir falloutDir_1
	test.Eq(t, len(dir.Header), 3)

	var dat falloutDat_1
	test.Eq(t, len(dat.Header), 3)
}
