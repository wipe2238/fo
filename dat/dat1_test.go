package dat

import (
	"testing"

	"github.com/shoenig/test"
)

func TestStructV1(t *testing.T) {
	var dir falloutDirV1
	test.Eq(t, len(dir.Header), 3)

	var dat falloutDatV1
	test.Eq(t, len(dat.Header), 3)
}
