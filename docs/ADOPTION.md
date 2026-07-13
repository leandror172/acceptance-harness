# Adoption — wiring the harness into a consumer repo

Step-by-step for adding an acceptance suite to a Go CLI project. The result is
a `test/` directory in your repo where the library provides the engine and the
generic assertions, and you provide three small domain pieces: actions (how to
invoke *your* CLI), domain verifiers (what *your* artifacts mean), and
fixtures.

## 1. Add the dependency

```bash
go get github.com/leandror172/acceptance-harness@v0.1.0
```

Pin an exact version — pre-1.0 the API may change between minors.

## 2. Layout

```
your-repo/
  cmd/mycli/            # the CLI under test
  test/
    setup_test.go       # TestMain: build the binary once
    <feature>_test.go   # scenarios + their then*/given* helpers
    actions/            # When: how to run your CLI
    verify/             # Then: domain assertions (yours)
    fixtures/           # per-scenario fixture dirs
```

`harness` and the generic `verify` come from the library; the three domain
directories stay yours. The dividing line: *"how to drive and assert a CLI
generically"* is library; *"what this CLI's bytes mean"* is yours. Domain
verifiers that parse your file formats should import your internal packages
and parse them exactly as the CLI does — never re-implement the parsing.

## 3. The build tag is yours, not the library's

The library carries no build tag, so it vets and tests like any dependency.
Whether your suite hides behind one is consumer policy. If your suite is slow
or needs external resources, tag your own test files:

```go
//go:build acceptance
```

and run with `go test -tags=acceptance ./test/...`. If your suite is fast and
hermetic, skip the tag and let it run with `go test ./...`. Remember the trap
either way: a tagged suite is invisible to plain `go test ./...` and must be
run explicitly after contract changes.

## 4. TestMain: build the binary once

```go
//go:build acceptance

package acceptance_test

import (
    "fmt"
    "os"
    "path/filepath"
    "testing"

    "github.com/leandror172/acceptance-harness/harness"
)

var binaryPath string

func TestMain(m *testing.M) {
    root, err := harness.FindModuleRoot()
    if err != nil {
        fmt.Fprintln(os.Stderr, "TestMain: find module root:", err)
        os.Exit(1)
    }
    bin, err := harness.BuildBinary(root, "./cmd/mycli", "mycli")
    if err != nil {
        fmt.Fprintln(os.Stderr, "TestMain: build:", err)
        os.Exit(1)
    }
    defer os.RemoveAll(filepath.Dir(bin))
    binaryPath = bin
    os.Exit(m.Run())
}
```

Every test file in the package shares `binaryPath`; each scenario's Given
assigns it to `ctx.BinaryPath`.

## 5. Actions: your When functions

An action runs the binary once and captures everything observable into the
`Context`. A typical exec wrapper:

```go
func runCLI(ctx *harness.Context, args ...string) {
    cmd := exec.Command(ctx.BinaryPath, args...)
    cmd.Dir = ctx.WorkDir
    cmd.Env = os.Environ()
    for k, v := range ctx.Env {
        cmd.Env = append(cmd.Env, k+"="+v)
    }
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    err := cmd.Run()
    ctx.Stdout, ctx.Stderr = stdout.String(), stderr.String()
    ctx.ExitCode = exitCodeOf(err)
}
```

Register any output files the command produced as named artifacts
(`ctx.Artifacts["report"] = filepath.Join(ctx.WorkDir, "report.csv")`) so
generic verifiers can find them by name.

`ctx.Env` is the injection seam for anything your binary reads from the
environment — an overridable clock, a config path, a feature flag. Givens set
it; the exec wrapper forwards it.

## 6. Fixtures: reads run zero-copy, writes run against a copy

- **Read-only scenario:** skip copying; pass `ctx.FixtureDir` to the binary
  (e.g. as a `--path` flag). Fast and obviously safe.
- **Mutating scenario:** `harness.CopyTreeToWorkDir(ctx, fixtureDir)` in the
  Given, then run against `ctx.WorkDir`.
- **Seeding single files:** `harness.SeedFileFromFixture` copies one named
  state file without dragging the whole fixture along.

A fixture `config.json` is optional. If you use one, the library decodes
`command` and `extra_args`; your domain fields ride in `Raw`:

```go
type myFixtureConfig struct {
    harness.FixtureConfig
    AccuracyFloor float64
}

func loadMyFixtureConfig(dir string) (myFixtureConfig, error) {
    base, err := harness.LoadFixtureConfig(dir)
    if err != nil {
        return myFixtureConfig{}, err
    }
    out := myFixtureConfig{FixtureConfig: base}
    if raw, ok := base.Raw["accuracy_floor"]; ok {
        if err := json.Unmarshal(raw, &out.AccuracyFloor); err != nil {
            return myFixtureConfig{}, err
        }
    }
    return out, nil
}
```

`harness.DiscoverFixtures(baseDir)` lists fixture dirs that contain a
config.json, for table-driven suites.

## 7. Debugging: keep the work dir

```bash
HARNESS_KEEP_ON_FAILURE=1 go test -tags=acceptance ./test/...
```

leaves each failed scenario's WorkDir on disk (the path is logged). Set
`HARNESS_KEEP_ARTIFACTS=1` to keep every WorkDir. Consumers who prefer flags
call `harness.RegisterFlags()` in `TestMain` before flags are parsed, which
binds `--keep-on-failure` / `--keep-artifacts` to the same knobs.

## 8. Nondeterministic CLIs (LLM-backed and friends)

The library core is deliberately deterministic-only: no liveness gates for
external services, no soft/statistical assertions, no accuracy floors, no
drift tracking. If your CLI's output varies run-to-run, build that machinery
in your own `verify/` (or an internal package): a `t.Skip` gate for the
backing service, soft assertions that compute an accuracy percentage against
a floor, expected files that omit the nondeterministic fields. Two patterns
that carry over well:

- **Canonical anchor inputs** — inputs empirically stable across model/config
  versions — for structural tests; diverse inputs only in soft-assertion tests.
- **Deterministic-subset expected files** — assert only the fields the
  contract guarantees.

## 9. Conventions

Read [PATTERNS.md](PATTERNS.md) before writing the first scenario. The
naming discipline (event-style Givens, result-named Then helpers, no raw
`verify.*` in `Then:` blocks) is what keeps the suite legible as it grows —
it costs nothing to adopt on day one and is painful to retrofit.
