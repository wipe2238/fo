package dbg

import (
	"testing"

	"github.com/shoenig/test"
)

func TestAddSize(t *testing.T) {
	var dbgMap Map = make(Map)

	dbgMap["Begin"] = int64(77)
	dbgMap["End"] = int64(127)

	dbgMap.AddSize("Size", "Begin", "End")

	test.EqOp(t, dbgMap["Size"].(int64), int64(127-77))
}
