package harness

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// FixtureConfig is the generic fixture descriptor. The core carries only the
// command and extra args; any consumer-specific fields are decoded from Raw,
// which captures every key of config.json for a second, consumer-owned
// unmarshal pass. A config.json is optional — actions may pass flags directly.
type FixtureConfig struct {
	Command   string                     `json:"command"`
	ExtraArgs []string                   `json:"extra_args"`
	Raw       map[string]json.RawMessage `json:"-"` // all keys, for consumer-specific decode
}

// LoadFixtureConfig reads config.json from dir. Absent fields stay zero-valued;
// the core applies no domain defaults.
func LoadFixtureConfig(dir string) (FixtureConfig, error) {
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		return FixtureConfig{}, err
	}
	var cfg FixtureConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FixtureConfig{}, err
	}
	_ = json.Unmarshal(data, &cfg.Raw) // capture all keys for consumer-specific decode
	return cfg, nil
}

// CopyFixtureToWorkDir copies the immediate files of fixtureDir into ctx.WorkDir
// (shallow — subdirectories are skipped). Use CopyTreeToWorkDir for fixture trees.
func CopyFixtureToWorkDir(ctx *Context, fixtureDir string) error {
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(fixtureDir, entry.Name())
		dst := filepath.Join(ctx.WorkDir, entry.Name())
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// CopyTreeToWorkDir recursively copies the entire fixture tree (files and
// subdirectories, structure preserved) into ctx.WorkDir. Most CLIs under test
// have tree-shaped fixtures, so write scenarios use this to get an isolated
// mutable copy and run the binary against ctx.WorkDir; read-only scenarios can
// skip the copy entirely and point the binary at ctx.FixtureDir.
func CopyTreeToWorkDir(ctx *Context, fixtureDir string) error {
	return copyTree(fixtureDir, ctx.WorkDir)
}

func copyTree(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())
		if entry.IsDir() {
			if err := copyTree(src, dst); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// DiscoverFixtures returns all subdirectories under baseDir that contain a config.json.
func DiscoverFixtures(baseDir string) ([]string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	var fixtures []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(baseDir, entry.Name())
		if _, err := os.Stat(filepath.Join(dir, "config.json")); err == nil {
			fixtures = append(fixtures, dir)
		}
	}
	return fixtures, nil
}

// SeedFileFromFixture copies a single named file from fixtureDir into ctx.WorkDir.
// If destName is empty, the source name is used. Use this to seed pre-existing
// state files without copying the whole fixture.
func SeedFileFromFixture(ctx *Context, fixtureDir, sourceName, destName string) error {
	if destName == "" {
		destName = sourceName
	}
	src := filepath.Join(fixtureDir, sourceName)
	dst := filepath.Join(ctx.WorkDir, destName)
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
