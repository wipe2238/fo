package dbg

import (
	"testing"

	"github.com/shoenig/test"
)

func TestFmts(t *testing.T) {
	t.Run("Number", func(t *testing.T) {
		test.EqOp(t, Fmts(1207)[0], "0x4B7")
		test.EqOp(t, Fmts(1207)[1], "1207")
	})
}

func TestFmt(t *testing.T) {
	t.Run("Number", func(t *testing.T) {
		test.EqOp(t, Fmt("", int32(1207)), "0x4B71207")
		test.EqOp(t, Fmt(",", int32(0x1207)), "0x1207,4615")
	})

	t.Run("Slice", func(t *testing.T) {
		test.EqOp(t, Fmt("", []uint16{0, 1207, 0x1207}), "[0x0 0x4B7 0x1207][0 1207 4615]")
		test.EqOp(t, Fmt("=", []int16{0, -0x1207, 1207}), "[0x0 -0x1207 0x4B7]=[0 -4615 1207]")
	})

	t.Run("Other", func(t *testing.T) {
		test.EqOp(t, Fmt("", "testing"), `"testing"`)
	})
}
