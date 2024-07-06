package dat

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestGetGame(test *testing.T) {
	must.Eq(test, new(falloutDat_1).GetGame(), 1)
}
