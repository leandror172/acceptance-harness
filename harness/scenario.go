// Package harness is a domain-agnostic, file-driven BDD engine for end-to-end
// testing of CLI tools. It provides Context/Scenario/Run plus fixture and
// binary-build helpers, and carries zero knowledge of any consumer's domain.
// Domain lives in each consumer's sibling actions/ (When: command runners) and
// verify/ (Then: assertions) packages. The core is LLM-free, network-free, and
// git-free by design.
package harness

import (
	"os"
	"testing"
)

// Context holds state for a single acceptance test scenario.
type Context struct {
	T          *testing.T
	FixtureDir string            // absolute path to the scenario's fixture dir (read-only source)
	WorkDir    string            // per-scenario temp dir — write scenarios run against a copy here
	BinaryPath string            // the once-built CLI under test
	Env        map[string]string // extra environment for the command (e.g. an injected clock)
	Artifacts  map[string]string // key → absolute file path, registered by actions
	ExitCode   int
	Stdout     string
	Stderr     string
}

// Scenario defines a Given/When/Then acceptance test.
type Scenario struct {
	Name  string
	Given func(*Context)
	When  func(*Context)
	Then  []func(*Context)
}

// Run executes the scenario directly on t (no subtest), calling Given → When →
// Then[] in order. Running on the parent t flushes t.Log output in real time
// under -v. A fresh temp WorkDir is created per scenario and removed on cleanup
// unless keep-artifacts (always) or keep-on-failure (on failure) is requested
// via env var or registered flag.
func Run(t *testing.T, s Scenario) {
	t.Helper()
	t.Logf("scenario: %s", s.Name)

	workDir, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatalf("Run: create work dir: %v", err)
	}
	t.Cleanup(func() {
		if keepArtifacts || (t.Failed() && keepArtifactsOnFailure) {
			t.Logf("artifacts preserved: %s", workDir)
			return
		}
		os.RemoveAll(workDir)
	})

	ctx := &Context{
		T:         t,
		Artifacts: make(map[string]string),
		Env:       make(map[string]string),
		WorkDir:   workDir,
	}
	if s.Given != nil {
		t.Log("→ Given: setting up scenario")
		s.Given(ctx)
	}
	if s.When != nil {
		t.Log("→ When: executing command")
		s.When(ctx)
	}
	t.Logf("→ Then: checking %d assertion(s)", len(s.Then))
	for _, step := range s.Then {
		if step != nil {
			step(ctx)
		}
	}
}
