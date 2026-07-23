# acceptance-harness

A domain-agnostic, file-driven Given/When/Then engine for end-to-end acceptance
testing of CLI tools, written in Go and run with plain `go test`.

You point it at a compiled binary and a fixture directory; it gives you a
three-phase scenario structure, an isolated per-scenario work dir, and a base of
composable assertions over exit codes, output, JSON, and files. Everything
specific to *your* CLI — how to invoke it, what its artifacts mean — stays in
your repo as small `actions/` and `verify/` packages.

## The three-phase model

A scenario is data, not a framework subclass:

```go
harness.Run(t, harness.Scenario{
    Name:    "tracked role is reported with its current status",
    Fixture: fixDir,                            // engine assigns ctx.FixtureDir
    Given:   roleTrackedWithStatus("applied"),  // set up fixture state
    When:    runGet("101", "--json"),           // invoke the binary once
    Then: slices.Concat(                        // composable assertions
        commandSucceeded(),
        statusReportedAs("applied"),
    ),
})
```

- **Given** `func(*Context)` — prepare state: copy a fixture tree into the work
  dir, seed files, set env. Compose several with `harness.Events(...)`.
- **When** `func(*Context)` — run the CLI under test, capture exit code,
  stdout/stderr, and named output artifacts.
- **Then** `[]func(*Context)` — assertions, each scoped to one concern,
  composed with `slices.Concat` at the call site.

Scenario plumbing belongs to the engine, not to your Givens: `Scenario.Fixture`
seeds `ctx.FixtureDir`, and `harness.UseBinary(path)` — called once in `TestMain` —
seeds `ctx.BinaryPath`. Both are optional and seeded *before* Given runs, so a
suite that wires them by hand keeps working unchanged.

The conventions that keep suites legible at scale — event-style Given names,
result-named Then helpers, no raw `verify.*` calls inside a `Then:` block — are
documented in [docs/PATTERNS.md](docs/PATTERNS.md). They are the most valuable
part of this repo; the code exists to serve them.

## What's in the box

| Package | Contents |
|---|---|
| `harness` | `Context`, `Scenario`, `Run`; scenario plumbing (`Scenario.Fixture`, `UseBinary`, `Events`, `Context.State` + `Context.BeforeWhen`); fixture plumbing (`CopyTreeToWorkDir`, `CopyFixtureToWorkDir`, `DiscoverFixtures`, `SeedFileFromFixture`, `FixtureConfig`); `BuildBinary` / `FindModuleRoot` for the build-once `TestMain` pattern |
| `verify` | Generic Then base: exit code + output (`CommandSucceeded/Failed`, `OutputContains/...`), JSON with dotted-path + array-index navigation (`JSONFieldEquals("app.items.0.status", ...)`), file state (`FileUnchanged`, `FileAbsent`, `WorkFileExists`, `FileContains/NotContains`) |
| `docs` | [PATTERNS.md](docs/PATTERNS.md) — the methodology; [ADOPTION.md](docs/ADOPTION.md) — wiring a consumer repo step by step; [LTG-PATH.md](docs/LTG-PATH.md) — adoption sketch for a pipeline-shaped consumer |

Deliberately **not** in the box: LLM gating, network probes, soft/statistical
assertions, domain file formats. The core is LLM-free, network-free, and
git-free; consumers with nondeterministic output add their own machinery on
top.

## Quick start

```go
//go:build acceptance

package acceptance_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/leandror172/acceptance-harness/harness"
)

var binaryPath string

func TestMain(m *testing.M) {
    root, err := harness.FindModuleRoot()
    if err != nil {
        panic(err)
    }
    bin, err := harness.BuildBinary(root, "./cmd/mycli", "mycli")
    if err != nil {
        panic(err)
    }
    binaryPath = bin

    // Not deferred: os.Exit skips defers.
    code := m.Run()
    os.RemoveAll(filepath.Dir(bin))
    os.Exit(code)
}
```

Whether your suite hides behind a build tag (`//go:build acceptance`) is your
call, not the library's — the harness itself carries no tag. See
[docs/ADOPTION.md](docs/ADOPTION.md) for the full consumer setup, including the
zero-copy-read vs copy-to-WorkDir-write pattern.

## Keeping artifacts for debugging

Failed-scenario work dirs are removed by default. To keep them:

```bash
HARNESS_KEEP_ON_FAILURE=1 go test -tags=acceptance ./test/...   # keep on failure
HARNESS_KEEP_ARTIFACTS=1  go test -tags=acceptance ./test/...   # keep always
```

Consumers who prefer CLI flags can call `harness.RegisterFlags()` from their
`TestMain` before `flag.Parse()` runs, which binds the same two knobs to
`--keep-on-failure` / `--keep-artifacts`.

## Versioning

Pre-1.0: the API may change between minor versions. Pin an exact version.

## License

MIT — see [LICENSE](LICENSE).
