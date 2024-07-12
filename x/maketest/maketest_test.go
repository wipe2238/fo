package maketest

import (
	"path/filepath"
	"testing"

	"github.com/shoenig/test"
)

func TestRepoDir(t *testing.T) {
	var repo, found = RepoDir()

	t.Logf("repo=[%s] found=[%t]", repo, found)
	test.StrNotEqFold(t, repo, "")

	for _, file := range []string{"go.mod", "go.sum", "makefile.go", "maketest.go"} {
		t.Logf(filepath.Join(repo, file))
		test.FileExists(t, filepath.Join(repo, file))
	}

	test.True(t, found)
}

func TestMust(t *testing.T) {
	test.True(t, Must("go"))
}
