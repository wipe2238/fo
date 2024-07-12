package maketest

import (
	"testing"

	"github.com/shoenig/test"
)

func TestRepoDir(t *testing.T) {
	var repo, found = RepoDir()

	test.StrNotEqFold(t, repo, "")
	test.FilePathValid(t, repo)

	test.True(t, found)
}

func TestMust(t *testing.T) {
	test.True(t, Must("go"))
}
