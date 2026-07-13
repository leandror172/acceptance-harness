package verify

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

// FileUnchanged asserts a WorkDir file is byte-identical to its pristine fixture
// copy — the dry-run "nothing written" proof.
func FileUnchanged(relFile string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		work, err1 := os.ReadFile(filepath.Join(ctx.WorkDir, relFile))
		orig, err2 := os.ReadFile(filepath.Join(ctx.FixtureDir, relFile))
		if !assert.NoError(ctx.T, err1, "read work %s", relFile) || !assert.NoError(ctx.T, err2, "read fixture %s", relFile) {
			return
		}
		assert.Equal(ctx.T, string(orig), string(work),
			"%s should be byte-identical to the fixture (no write)", relFile)
	}
}

// FileAbsent asserts a WorkDir file does not exist (e.g. a dry-run created no side file).
func FileAbsent(relFile string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		_, err := os.Stat(filepath.Join(ctx.WorkDir, relFile))
		assert.True(ctx.T, os.IsNotExist(err), "%s should not exist", relFile)
	}
}

// WorkFileExists asserts a WorkDir file exists (e.g. the command created its output).
func WorkFileExists(relFile string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		_, err := os.Stat(filepath.Join(ctx.WorkDir, relFile))
		assert.NoError(ctx.T, err, "%s should exist", relFile)
	}
}

// FileContains asserts a WorkDir file contains substr.
func FileContains(relFile, substr string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		b, err := os.ReadFile(filepath.Join(ctx.WorkDir, relFile))
		if !assert.NoError(ctx.T, err, "read %s", relFile) {
			return
		}
		assert.Contains(ctx.T, string(b), substr, "%s should contain %q", relFile, substr)
	}
}

// FileNotContains asserts a WorkDir file does NOT contain substr — used to prove
// a generated artifact omits a forbidden token.
func FileNotContains(relFile, substr string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		b, err := os.ReadFile(filepath.Join(ctx.WorkDir, relFile))
		if !assert.NoError(ctx.T, err, "read %s", relFile) {
			return
		}
		assert.NotContains(ctx.T, string(b), substr, "%s should not contain %q", relFile, substr)
	}
}
