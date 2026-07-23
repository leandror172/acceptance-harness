// Package harness is a domain-agnostic, file-driven BDD engine for end-to-end
// testing of CLI tools. It provides Context/Scenario/Run plus fixture and
// binary-build helpers, and carries zero knowledge of any consumer's domain.
// Domain lives in each consumer's sibling actions/ (When: command runners) and
// verify/ (Then: assertions) packages. The core is LLM-free, network-free, and
// git-free by design.
package harness

import (
	"os"
	"strings"
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
	// State is scenario-scoped scratch space for consumer setup code that must
	// accumulate across several Given events — for example config keys contributed
	// by independent events and flushed once via BeforeWhen. The engine never reads
	// it. Key it with a package-private constant to avoid collisions between a
	// consumer's own layers.
	State    map[string]any
	ExitCode int
	Stdout   string
	Stderr   string

	beforeWhen []func()
}

// BeforeWhen registers fn to run after every Given event and before When.
//
// Use it to flush state accumulated across composed Givens exactly once. Setup
// that writes a whole file per contributing event — a config file being the usual
// case — makes the last writer win and silently drops the other events' keys,
// which is what makes composed Givens unsafe. Accumulate in Context.State,
// register one BeforeWhen, write once.
func (c *Context) BeforeWhen(fn func()) {
	if fn == nil {
		return
	}
	c.beforeWhen = append(c.beforeWhen, fn)
}

// Scenario defines a Given/When/Then acceptance test.
type Scenario struct {
	Name string
	// Fixture is the scenario's fixture directory. When set, Run assigns it to
	// ctx.FixtureDir before Given runs, so Given helpers need neither take nor
	// wire a fixture path. Optional — leave it empty and a Given may assign
	// ctx.FixtureDir itself, as it always could.
	Fixture string
	// Binary overrides the package default set by UseBinary, for suites that
	// exercise more than one binary. Optional.
	Binary string
	Given  func(*Context)
	When   func(*Context)
	Then   []func(*Context)
}

// defaultBinary is the binary Run assigns to every scenario, set by UseBinary.
var defaultBinary string

// UseBinary sets the binary Run assigns to ctx.BinaryPath for every scenario.
// Call it once from TestMain after BuildBinary; a Scenario.Binary overrides it,
// and a Given assigning ctx.BinaryPath still wins because Given runs later.
//
// This is package-level mutable state, deliberately unguarded: it is written once
// during suite setup, before any scenario runs. Mutating it from a running
// parallel test is not supported.
func UseBinary(path string) {
	defaultBinary = path
}

// Events folds a sequence of setup events into the single function Scenario.Given
// takes, so a Given can be written as the sequence of events it actually is:
//
//	Given: harness.Events(binaryBuilt(), taxonomyPublished(fix), logsConfigured()),
//
// Events are applied in order; nil events are skipped. Composed events must not
// each rewrite the same file — see BeforeWhen.
func Events(events ...func(*Context)) func(*Context) {
	return func(ctx *Context) {
		for _, event := range events {
			if event != nil {
				event(ctx)
			}
		}
	}
}

// Run executes the scenario directly on t (no subtest), calling Given →
// BeforeWhen hooks → When → Then[] in order. Running on the parent t flushes
// t.Log output in real time under -v. A fresh temp WorkDir is created per
// scenario and removed on cleanup unless keep-artifacts (always) or
// keep-on-failure (on failure) is requested via env var or registered flag.
//
// Before Given runs, Run seeds the context with the scenario's plumbing:
// BinaryPath from UseBinary (or Scenario.Binary) and FixtureDir from
// Scenario.Fixture. Both are seeded first, so a Given that assigns them itself
// still wins — suites written before these fields existed are unaffected.
func Run(t *testing.T, s Scenario) {
	t.Helper()
	t.Logf("scenario: %s", s.Name)

	// Subtest names contain "/", which is illegal in a MkdirTemp pattern.
	workDir, err := os.MkdirTemp("", strings.ReplaceAll(t.Name(), "/", "_"))
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
		T:          t,
		Artifacts:  make(map[string]string),
		Env:        make(map[string]string),
		State:      make(map[string]any),
		WorkDir:    workDir,
		BinaryPath: defaultBinary,
		FixtureDir: s.Fixture,
	}
	if s.Binary != "" {
		ctx.BinaryPath = s.Binary
	}
	if s.Given != nil {
		t.Log("→ Given: setting up scenario")
		s.Given(ctx)
	}
	for _, flush := range ctx.beforeWhen {
		flush()
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
