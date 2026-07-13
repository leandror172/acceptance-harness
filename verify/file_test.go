package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

func workContext(t *testing.T) *harness.Context {
	t.Helper()
	return &harness.Context{
		T:          t,
		WorkDir:    t.TempDir(),
		FixtureDir: t.TempDir(),
	}
}

func writeWorkFile(t *testing.T, ctx *harness.Context, relPath, content string) {
	t.Helper()
	dir := filepath.Dir(relPath)
	if dir != "." {
		err := os.MkdirAll(filepath.Join(ctx.WorkDir, dir), 0755)
		assert.NoError(t, err)
	}
	err := os.WriteFile(filepath.Join(ctx.WorkDir, relPath), []byte(content), 0644)
	assert.NoError(t, err)
}

func TestFileUnchanged_PassesWhenIdentical(t *testing.T) {
	ctx := workContext(t)
	content := "hello world"
	err := os.WriteFile(filepath.Join(ctx.FixtureDir, "data.txt"), []byte(content), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(ctx.WorkDir, "data.txt"), []byte(content), 0644)
	assert.NoError(t, err)
	FileUnchanged("data.txt")(ctx)
}

func TestFileAbsent_PassesWhenMissing(t *testing.T) {
	ctx := workContext(t)
	FileAbsent("ghost.txt")(ctx)
}

func TestWorkFileExists_PassesWhenPresent(t *testing.T) {
	ctx := workContext(t)
	writeWorkFile(t, ctx, "out.txt", "output")
	WorkFileExists("out.txt")(ctx)
}

func TestFileContains_FindsSubstring(t *testing.T) {
	ctx := workContext(t)
	writeWorkFile(t, ctx, "file.txt", "hello world")
	FileContains("file.txt", "world")(ctx)
}

func TestFileNotContains_PassesWhenAbsent(t *testing.T) {
	ctx := workContext(t)
	writeWorkFile(t, ctx, "file.txt", "hello world")
	FileNotContains("file.txt", "goodbye")(ctx)
}

func TestWorkFileExists_ResolvesRelativeSubdirPath(t *testing.T) {
	ctx := workContext(t)
	writeWorkFile(t, ctx, "sub/nested.txt", "content")
	WorkFileExists("sub/nested.txt")(ctx)
}
