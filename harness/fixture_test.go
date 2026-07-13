package harness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFixtureConfig(t *testing.T) {
	t.Run("reads config.json", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
			"command": "test",
			"extra_args": ["--flag", "--verbose"]
		}`), 0644))

		cfg, err := LoadFixtureConfig(dir)
		require.NoError(t, err)
		assert.Equal(t, "test", cfg.Command)
		assert.Equal(t, []string{"--flag", "--verbose"}, cfg.ExtraArgs)
		assert.NotNil(t, cfg.Raw)
	})

	t.Run("captures all keys in Raw", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
			"command": "test",
			"extra_args": ["--flag"],
			"accuracy_floor": 0.95,
			"model_name": "gpt-4"
		}`), 0644))

		cfg, err := LoadFixtureConfig(dir)
		require.NoError(t, err)
		assert.Equal(t, "test", cfg.Command)
		assert.Equal(t, []string{"--flag"}, cfg.ExtraArgs)

		// Check that all keys are in Raw
		assert.Contains(t, cfg.Raw, "command")
		assert.Contains(t, cfg.Raw, "extra_args")
		assert.Contains(t, cfg.Raw, "accuracy_floor")
		assert.Contains(t, cfg.Raw, "model_name")
	})

	t.Run("returns error for missing config.json", func(t *testing.T) {
		dir := t.TempDir()
		_, err := LoadFixtureConfig(dir)
		assert.Error(t, err)
	})

	t.Run("returns error for malformed JSON", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"command": "test"`), 0644))
		_, err := LoadFixtureConfig(dir)
		assert.Error(t, err)
	})
}

func TestCopyFixtureToWorkDir_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	fixtureDir := filepath.Join(dir, "fixture")
	require.NoError(t, os.MkdirAll(fixtureDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(fixtureDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "subdir", "nested.txt"), []byte("nested content"), 0644))

	ctx := &Context{WorkDir: t.TempDir()}
	err := CopyFixtureToWorkDir(ctx, fixtureDir)
	require.NoError(t, err)

	// Check that only top-level files were copied
	assert.FileExists(t, filepath.Join(ctx.WorkDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(ctx.WorkDir, "file2.txt"))
	assert.NoFileExists(t, filepath.Join(ctx.WorkDir, "subdir"))
}

func TestCopyTreeToWorkDir_Recursive(t *testing.T) {
	dir := t.TempDir()
	fixtureDir := filepath.Join(dir, "fixture")
	require.NoError(t, os.MkdirAll(fixtureDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(fixtureDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "subdir", "nested.txt"), []byte("nested content"), 0644))

	ctx := &Context{WorkDir: t.TempDir()}
	err := CopyTreeToWorkDir(ctx, fixtureDir)
	require.NoError(t, err)

	// Check that tree structure is preserved
	assert.FileExists(t, filepath.Join(ctx.WorkDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(ctx.WorkDir, "file2.txt"))
	assert.DirExists(t, filepath.Join(ctx.WorkDir, "subdir"))
	assert.FileExists(t, filepath.Join(ctx.WorkDir, "subdir", "nested.txt"))

	// Check file contents
	content1, err := os.ReadFile(filepath.Join(ctx.WorkDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	nestedContent, err := os.ReadFile(filepath.Join(ctx.WorkDir, "subdir", "nested.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(nestedContent))
}

func TestSeedFileFromFixture(t *testing.T) {
	t.Run("copies file with default name", func(t *testing.T) {
		dir := t.TempDir()
		fixtureDir := filepath.Join(dir, "fixture")
		require.NoError(t, os.MkdirAll(fixtureDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "source.txt"), []byte("file content"), 0644))

		ctx := &Context{WorkDir: t.TempDir()}
		err := SeedFileFromFixture(ctx, fixtureDir, "source.txt", "")
		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(ctx.WorkDir, "source.txt"))
		content, err := os.ReadFile(filepath.Join(ctx.WorkDir, "source.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file content", string(content))
	})

	t.Run("copies file with custom name", func(t *testing.T) {
		dir := t.TempDir()
		fixtureDir := filepath.Join(dir, "fixture")
		require.NoError(t, os.MkdirAll(fixtureDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(fixtureDir, "source.txt"), []byte("file content"), 0644))

		ctx := &Context{WorkDir: t.TempDir()}
		err := SeedFileFromFixture(ctx, fixtureDir, "source.txt", "dest.txt")
		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(ctx.WorkDir, "dest.txt"))
		content, err := os.ReadFile(filepath.Join(ctx.WorkDir, "dest.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file content", string(content))
	})

	t.Run("returns error for missing source file", func(t *testing.T) {
		dir := t.TempDir()
		fixtureDir := filepath.Join(dir, "fixture")
		require.NoError(t, os.MkdirAll(fixtureDir, 0755))

		ctx := &Context{WorkDir: t.TempDir()}
		err := SeedFileFromFixture(ctx, fixtureDir, "missing.txt", "")
		assert.Error(t, err)
	})
}

func TestDiscoverFixtures(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644))

	subdir1 := filepath.Join(dir, "subdir1")
	require.NoError(t, os.MkdirAll(subdir1, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir1, "config.json"), []byte(`{"command": "test"}`), 0644))

	subdir2 := filepath.Join(dir, "subdir2")
	require.NoError(t, os.MkdirAll(subdir2, 0755))

	// subdir3 has no config.json
	subdir3 := filepath.Join(dir, "subdir3")
	require.NoError(t, os.MkdirAll(subdir3, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir3, "somefile.txt"), []byte("content"), 0644))

	fixtures, err := DiscoverFixtures(dir)
	require.NoError(t, err)
	assert.Len(t, fixtures, 1)
	assert.Contains(t, fixtures, subdir1)
}
