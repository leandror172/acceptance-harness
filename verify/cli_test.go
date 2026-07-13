package verify

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

func TestCommandSucceeded_PassesOnExitZero(t *testing.T) {
	t.Parallel()
	ctx := &harness.Context{T: t, ExitCode: 0}
	CommandSucceeded()(ctx)
}

func TestCommandFailed_PassesOnNonZeroExit(t *testing.T) {
	t.Parallel()
	ctx := &harness.Context{T: t, ExitCode: 1}
	CommandFailed()(ctx)
}

func TestOutputContains_SearchesStdoutAndStderr(t *testing.T) {
	t.Parallel()
	ctx := &harness.Context{
		T:      t,
		Stdout: "alpha",
		Stderr: "beta",
	}
	OutputContains("alpha")(ctx)
	OutputContains("beta")(ctx)
}

func TestOutputNotContains_PassesWhenAbsent(t *testing.T) {
	t.Parallel()
	ctx := &harness.Context{
		T:      t,
		Stdout: "alpha",
	}
	OutputNotContains("gamma")(ctx)
}

func TestOutputFileExists_ChecksRegisteredArtifact(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := tmpDir + "/report.txt"
	err := os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)

	ctx := &harness.Context{
		T:         t,
		Artifacts: map[string]string{"report": filePath},
	}
	OutputFileExists("report")(ctx)
}
