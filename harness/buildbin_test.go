package harness

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeThrowawayModule(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "throwaway-module-*")
	require.NoError(t, err)

	modFile := filepath.Join(tmpDir, "go.mod")
	err = os.WriteFile(modFile, []byte("module throwaway\n\ngo 1.24\n"), 0644)
	require.NoError(t, err)

	return tmpDir
}

func TestFindModuleRoot_FromModuleDir(t *testing.T) {
	tmpDir := writeThrowawayModule(t)
	defer os.RemoveAll(tmpDir)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	root, err := FindModuleRoot()
	require.NoError(t, err)

	evalRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	evalTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, evalTmpDir, evalRoot)
}

func TestFindModuleRoot_WalksUpFromSubdir(t *testing.T) {
	tmpDir := writeThrowawayModule(t)
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "sub", "sub2")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)

	err = os.Chdir(subDir)
	require.NoError(t, err)

	root, err := FindModuleRoot()
	require.NoError(t, err)

	evalRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	evalTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, evalTmpDir, evalRoot)
}

func TestBuildBinary_CompilesAndReturnsPath(t *testing.T) {
	tmpDir := writeThrowawayModule(t)
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

import "fmt"

func main() {
	fmt.Println("ok")
}`), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	binPath, err := BuildBinary(tmpDir, ".", "throwaway")
	require.NoError(t, err)
	assert.FileExists(t, binPath)
	defer os.RemoveAll(filepath.Dir(binPath))

	output, err := exec.Command(binPath).CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "ok")
}

func TestBuildBinary_BuildFailureReturnsError(t *testing.T) {
	tmpDir := writeThrowawayModule(t)
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(`package main

import "fmt"

func main() {
	fmt.Println("ok"
}`), 0644)
	require.NoError(t, err)

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	binPath, err := BuildBinary(tmpDir, ".", "throwaway")
	assert.Error(t, err)
	assert.Empty(t, binPath)
	assert.True(t, strings.Contains(err.Error(), "build"))
}
