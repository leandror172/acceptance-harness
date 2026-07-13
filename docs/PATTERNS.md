# Patterns — writing acceptance scenarios that stay legible

The harness is ~500 lines of code; these conventions are what make a suite of
50+ scenarios readable years later. They were developed across two real
consumer suites and are the primary content of this repo. Examples use a toy
`todo` CLI (`todo add`, `todo done`, `todo list --json`).

## The three-phase model

```go
harness.Run(t, harness.Scenario{
    Name:  "completing a task moves it out of the pending list",
    Given: taskAddedToList("buy milk"),
    When:  runDone("1"),
    Then: slices.Concat(
        commandSucceeded(),
        pendingListIsEmpty(),
    ),
})
```

- **Given** prepares state (copy a fixture, seed files, set env).
- **When** runs the binary exactly once and captures everything observable.
- **Then** is a flat list of assertions, each scoped to one concern.

One scenario, one `When`. If a behavior needs several commands in sequence,
either fold the earlier commands into the Given (they are setup — events that
already happened) or split into a chain of scenarios where each test runs one
command and later Givens re-create the accumulated state.

## Given naming: past-tense domain events

Given helpers use Event-Modeling style: they name an **event that happened in
the domain**, not the technical setup that simulates it.

| Good (event) | Bad (mechanism / state) |
|---|---|
| `taskAddedToList("buy milk")` | `setupTasksFile(fixDir)` |
| `paymentReceivedForOrder(orderID)` | `withPaymentConfig` |
| `listArchivedThenRestored()` | `archivedListExists(fixDir)` |

**Exception:** the absence of any event is a state, not an event — name it
pragmatically, e.g. `noTasksRecorded()`.

Internal plumbing helpers (`withConfig`, `copyFixture`) are fine as private
functions *called by* Given helpers — they just must never appear as the
`Given:` value themselves.

**Why:** Given/When/Then reads as a story of what happened in the domain. A
system evolves as a sequence of recorded events; event-named Givens make the
test description match that reality.

## Then composition: `slices.Concat` of one-concern helpers

Then helpers return `[]func(*harness.Context)` and are combined at the call
site:

```go
Then: slices.Concat(
    commandSucceeded(),
    taskRecordedAsCompleted("1"),
    completionTimestampWritten("1"),
)
```

Rules that keep this readable:

1. **No raw `verify.*` calls inside a `Then:` block.** Library verifiers are
   building material for helper *bodies*; the scenario itself lists only named
   `then*`-style helpers. This keeps intent at the scenario level and detail
   encapsulated.
2. **One concern per helper.** Never add assertions to an existing helper
   unless they are truly part of the same concern — compose instead.
3. **Name helpers by the expected RESULT, not the artifact or mechanism.** The
   reader should infer the scenario's behavior from names alone, without
   opening the helper or the fixture. Prefer
   `recurringTaskExpandedToNDatedEntries(fixDir)` over
   `outputFileMatchesExpected(fixDir)`; prefer
   `completedTaskNotShownInPendingList()` over an inline
   `verify.OutputNotContains("buy milk")`.
4. **Split on variance.** Keep genuinely *invariant* concerns generic
   (`commandSucceeded()`); only the scenario-varying concern needs the
   outcome-describing name. This reconciles "one concern per helper" with
   "name the result".
5. **Per-scenario wrappers are worth their one-line bodies.** When several
   scenarios share a verifier, wrap it per scenario:
   `func urgentTaskListedFirst(fix string) { return outputMatchesExpected(fix) }`.
   The wrapper's value IS its name — the differing fixture encodes the
   differing result, and the name documents what that result means.

## Reads run zero-copy; writes run against a copy

`Context` carries both `FixtureDir` and `WorkDir` so both patterns are
first-class:

- **Read-only scenarios** point the binary straight at `ctx.FixtureDir` —
  no copy, faster, and obviously safe.
- **Mutating scenarios** call `harness.CopyTreeToWorkDir` in the Given and run
  the binary against `ctx.WorkDir`, so every scenario mutates an isolated copy.

The dry-run proof composes from both: run the "mutating" command with its
dry-run flag against a copy, then assert `verify.FileUnchanged(...)` — the
work-dir file is still byte-identical to the pristine fixture.

## Artifacts are the OBSERVE register

A scenario is really GIVEN → RUN → **OBSERVE** → ASSERT. The observe step is
`Context`: exit code, stdout/stderr, and `Artifacts map[string]string` — named
captures registered by the When action (`ctx.Artifacts["report"] = path`) and
consumed by Then predicates by name (`verify.OutputFileExists("report")`).
Actions decide *what is observable*; verifiers only look at what was captured.
Keeping that seam explicit is what lets the same Then helpers serve many
scenarios.

## Oracle-freezing: expected output must come from a trusted producer

When a scenario asserts against a committed expected file (a golden/frozen
expectation), that file must be produced by something **independently
trusted** — a hand-verified run, a previous system, a converged prototype —
never by the system under test itself. Freezing a program's own output and
asserting the program still emits it is circular: it pins behavior, including
the bugs.

Corollary: when the contract deliberately changes, re-freeze and **manually
review the diff** between old and new expected files before committing. A
frozen-oracle test cannot distinguish "both fixed" from "both broken".

## A write scenario earns a paired inverse scenario

Every mutating behavior should have a companion scenario asserting the
mutation can be undone or is truly gone (delete after create, restore after
archive, idempotent re-run). The inverse pair is where state-tracking bugs
live; suites that only test the forward direction miss them.

## Traps observed in real suites

- **Build tags hide breakage.** If your suite lives behind
  `//go:build acceptance`, `go test ./...` stays green while the suite rots.
  After any change to a contract the suite depends on (config schema, CLI
  flags), run the tagged suite explicitly.
- **Dry-run flags mask the write path.** A green scenario whose fixture passes
  `--dry-run` proves output production only — it never exercises the append/
  write path. Check fixture `extra_args` before trusting coverage claims.
- **Skip guards mask absent coverage.** A helper that `t.Skip`s when an
  optional resource is missing can silently skip forever. Audit what actually
  *ran*, not what passed.
- **Now()-dependent fixtures are time bombs.** If the CLI fills in the current
  year/date for partial inputs, fixtures with partial dates produce
  expectations that expire at the next year boundary. Pin absolute dates in
  fixtures and expected files.
- **Nondeterministic fields don't belong in expected files.** Generated IDs,
  timestamps, and any model/heuristic-dependent values should be omitted from
  expected files; assert the deterministic contract subset. (Consumers testing
  LLM-backed CLIs typically add soft/statistical assertion machinery in their
  own repo — deliberately not part of this library.)
