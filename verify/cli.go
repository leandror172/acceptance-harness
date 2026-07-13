// Package verify holds the generic Then assertions for acceptance suites.
// cli.go covers exit status, output substrings, and artifact existence;
// json.go adds JSON-output assertions with dotted-path navigation; file.go
// adds work-dir file-state assertions. Consumer-specific verifiers (domain
// file formats, semantic checks) belong in the consumer's own verify package.
package verify

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"

	"github.com/leandror172/acceptance-harness/harness"
)

// CommandSucceeded asserts the command exited with code 0.
func CommandSucceeded() func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		assert.Zero(ctx.T, ctx.ExitCode,
			"command should succeed (exit 0)\nstdout: %s\nstderr: %s", ctx.Stdout, ctx.Stderr)
	}
}

// CommandFailed asserts the command exited with a non-zero code.
func CommandFailed() func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		assert.NotZero(ctx.T, ctx.ExitCode,
			"command should fail (non-zero exit)\nstdout: %s\nstderr: %s", ctx.Stdout, ctx.Stderr)
	}
}

// OutputContains asserts stdout+stderr contains substr.
func OutputContains(substr string, msgAndArgs ...interface{}) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		output := ctx.Stdout + ctx.Stderr
		msg := fmt.Sprintf("%q not found in command output\nstdout: %s\nstderr: %s",
			substr, ctx.Stdout, ctx.Stderr)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%s — %v\n%s", substr, msgAndArgs[0], msg)
		}
		assert.Contains(ctx.T, output, substr, msg)
	}
}

// OutputNotContains asserts stdout+stderr does NOT contain substr.
func OutputNotContains(substr string, msgAndArgs ...interface{}) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		output := ctx.Stdout + ctx.Stderr
		msg := fmt.Sprintf("%q unexpectedly found in command output\nstdout: %s\nstderr: %s",
			substr, ctx.Stdout, ctx.Stderr)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf("%s — %v\n%s", substr, msgAndArgs[0], msg)
		}
		assert.NotContains(ctx.T, output, substr, msg)
	}
}

// OutputFileExists asserts the registered artifact key maps to an existing file.
func OutputFileExists(artifactKey string) func(*harness.Context) {
	return func(ctx *harness.Context) {
		ctx.T.Helper()
		path, ok := ctx.Artifacts[artifactKey]
		if !assert.True(ctx.T, ok, "artifact %q not registered", artifactKey) {
			return
		}
		_, err := os.Stat(path)
		assert.NoError(ctx.T, err, "output file %q should exist at %s", artifactKey, path)
	}
}
