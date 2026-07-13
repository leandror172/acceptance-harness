package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindModuleRoot walks upward from the current working directory until it finds a
// go.mod, returning that directory.
func FindModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// BuildBinary compiles the package at pkgPath (relative to moduleRoot) into a
// fresh temp dir and returns the binary's path. The caller owns cleanup of the
// returned binary's parent dir (e.g. defer os.RemoveAll(filepath.Dir(path))).
// A .exe suffix is added on Windows.
func BuildBinary(moduleRoot, pkgPath, binName string) (string, error) {
	binDir, err := os.MkdirTemp("", "acceptance-bin-*")
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)
	cmd := exec.Command("go", "build", "-o", binPath, pkgPath)
	cmd.Dir = moduleRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(binDir)
		return "", fmt.Errorf("build %s: %w", pkgPath, err)
	}
	return binPath, nil
}
