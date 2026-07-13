package harness

import (
	"flag"
	"os"
	"strings"
	"sync"
)

// keepArtifacts controls whether to preserve work directory contents after every test (pass or fail).
// It is initialized from the HARNESS_KEEP_ARTIFACTS environment variable.
var keepArtifacts bool

// keepArtifactsOnFailure controls whether to preserve work directory contents when a test fails.
// It is initialized from the HARNESS_KEEP_ON_FAILURE environment variable.
var keepArtifactsOnFailure bool

var registerFlagsOnce sync.Once

func init() {
	keepArtifacts = envBool("HARNESS_KEEP_ARTIFACTS")
	keepArtifactsOnFailure = envBool("HARNESS_KEEP_ON_FAILURE")
}

// RegisterFlags binds CLI flags --keep-artifacts and --keep-on-failure to the package-level
// keepArtifacts and keepArtifactsOnFailure variables, using their environment-resolved values as defaults.
// It is safe to call from TestMain before flag.Parse, and safe to call more than once.
func RegisterFlags() {
	registerFlagsOnce.Do(func() {
		flag.BoolVar(&keepArtifacts, "keep-artifacts", keepArtifacts, "preserve work dir contents after every test (pass or fail)")
		flag.BoolVar(&keepArtifactsOnFailure, "keep-on-failure", keepArtifactsOnFailure, "preserve work dir contents when a test fails (for inspection)")
	})
}

// envBool returns true if the environment variable is set to a truthy value.
// Truthy values are any non-empty string except "0" or "false" (case-insensitive).
func envBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		return false
	}
	return !strings.EqualFold(value, "false") && value != "0"
}
