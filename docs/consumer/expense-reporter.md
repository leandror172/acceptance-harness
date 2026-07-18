# Consumer report — expense-reporter

**Date:** 2026-07-17 · **Module:** v0.1.1 · **Suite:** 16 files, 50 test functions, ~49 scenarios

expense-reporter is the module's origin repo and, after career-search, its second consumer.
It is a Go CLI that classifies bank-export expenses with a local LLM (Ollama) and writes an
Excel budget; its acceptance suite is **LLM-backed and partly non-deterministic**, which makes
it the module's first real test of the "supports a nondeterministic CLI" claim. This report is
the expense lens on what the module provides — what's load-bearing, what shipped but isn't
touched, and the one package plus three patterns worth promoting. For the full post-v1 roadmap
see the origin plan `harness-extraction-plan.md` §6 (additions) and §7 (LTG path); this doc
does not duplicate it.

All counts below are from the migrated suite (`expense-reporter/test/`), not estimates.

---

## 1. What expense-reporter uses

### Engine (`harness`) — all load-bearing
| API | Sites | Note |
|---|---|---|
| `Context` (type) | 266 | threaded through every action/verifier signature |
| `Run` | 51 | one per scenario |
| `Scenario` | 49 | |
| `FindModuleRoot`, `BuildBinary` | 2, 2 | `TestMain` build-once |
| `RegisterFlags` | 3 | **required, not optional** — see §4 |
| `CopyFixtureToWorkDir` | 25 | the copy path expenses actually uses |
| `SeedFileFromFixture` | 7 | seed one prior-state file (e.g. a prewritten log) |
| `LoadFixtureConfig` / `FixtureConfig` | 1 / 1 | wrapped by a domain loader (`Raw` decode) |
| `ctx.Env` | (new) | carries `DATA_DIR`/`WORKBOOK_PATH`; forwarded to the subprocess |

### Generic assertions (`verify`) — the flat exit/output/JSON core
| API | Sites |
|---|---|
| `CommandSucceeded` / `CommandFailed` | 14 / 2 |
| `OutputContains` / `OutputNotContains` | 8 / 4 |
| `OutputFileExists` | 7 |
| `OutputJSONHasKey` / `OutputJSONHasValue` / `OutputIsValidJSON` | 9 / 3 / 1 |

That is the whole of `verify` that expenses touches. Everything else it needs is domain and
lives in its own `test/expect/` (25 verifiers over JSONL logs, CSV structure, review HTML, and
generated-workbook dumps) — correctly, since those parse expense file formats.

---

## 2. Shipped but unused (the part worth the author's attention)

A consumer's tested surface is smaller than the module's shipped surface. Verified by grep
against `test/`:

| Module API | expense-reporter sites | career-search |
|---|---|---|
| `CopyTreeToWorkDir` (recursive copy — a headline v1 feature) | **0** | ? |
| `DiscoverFixtures` | **0** | ? |
| `JSONFieldEquals` / `JSONArrayLen` / `JSONArrayEvery` (nested JSON) | **0 / 0 / 0** | ? |
| `verify/file.go` — `FileUnchanged` / `FileAbsent` / `WorkFileExists` / `FileContains` / `FileNotContains` | **0 / 0 / 0 / 0 / 0** | ? |

Two of these are just "not needed here" (expenses fixtures are flat; it uses `CopyFixtureToWorkDir`
+ `SeedFileFromFixture`, and asserts flat JSON). But one is a **fit signal, not an absence**:

### `FileUnchanged` doesn't fit a real "nothing was written" check
expense-reporter's dry-run test (`TestApply_DryRunWritesNothing`) needs to prove a command
wrote nothing to a log. It does **not** use `verify.FileUnchanged`; it uses a domain verifier
(`expect.ClassificationsMatch(seed)`). The reason is structural, and it generalizes to any
consumer whose artifacts are non-deterministic:

`FileUnchanged(relFile)` compares `WorkDir/relFile` against `FixtureDir/relFile` — **byte-identical,
same filename**. Expenses' logs (a) carry per-run `id`/`timestamp` fields, so byte-equality is
the wrong relation, and (b) are seeded under a different name than they're written
(`seed-classifications.jsonl` → `classifications.jsonl`), so the same-name assumption doesn't
hold either. The check it actually needs is *"semantically equal to this named baseline, ignoring
these volatile fields."*

**Takeaway for v1.0.0:** the battle-tested core is genuinely small — `Context`/`Scenario`/`Run`,
shallow copy + seed, flat exit/output/JSON asserts. `CopyTreeToWorkDir`, the nested-JSON walker,
and `verify/file.go` are so-far-unexercised by the only real LLM consumer. That's not "delete
them," but it is "don't freeze their signatures as if two consumers had shaken them out — one
did, and it didn't touch them." A byte-`FileUnchanged` in particular may want a companion that
snapshots-then-compares against a named baseline with a field-ignore list before it's canon.

---

## 3. The one addition expense-reporter concretely justifies: `extern/llm`

The roadmap (§6) already names an optional `extern/llm` package "once a second LLM consumer
appears." expense-reporter is the *first and only* LLM consumer today (career-search is
deterministic), so everything LLM-shaped it had to keep in its own `test/extern/` + `test/expect/`
is exactly that package's spec, written from real use rather than speculation:

1. **Liveness gate** — `RequireOllama(t, url)`: a 3s `GET /api/tags`; `t.Skipf` on any non-200 so
   a suite without the backend skips cleanly instead of failing. (In `test/extern/ollama.go`.)
2. **Soft assertion with an accuracy floor** — `ClassificationAccuracyAtLeast(artifact, expected,
   floor, resultsDir)`: compare against an expected file, compute a percentage, fail only below
   `floor`. Non-deterministic model output makes hard assertions flaky; a floor catches
   regressions without demanding reproducibility. (In `test/expect/accuracy.go`.)
3. **Drift tracking** — the same verifier writes a JSON report per run to a results dir, so
   accuracy trend across model/prompt changes is inspectable. (`test/results/`, gitignored.)

These three are a coherent package. **Gate:** it should still wait for a second LLM consumer
(LTG is the candidate — §7) before promotion, so the honest framing is *"here is the spec, drawn
from expense-reporter's real API, for when you build it"* — not "build it now." Promoting on a
single consumer risks freezing expense-specific choices (a 3s timeout, a semicolon-CSV accuracy
comparator) as the general shape.

---

## 4. Patterns worth lifting into the methodology docs (not code)

These emerged from the expense-reporter migration and are absent from `docs/PATTERNS.md`. Each is
free — a docs paragraph — and each is exactly what the *next* nondeterministic consumer needs.

- **Two-bucket verification for behavior-preserving changes to a nondeterministic suite.** When
  refactoring under the suite (as this migration did), "still green" is not proof — model drift
  can flip a soft test independently of your change. Split the roster: **deterministic tests must
  match pass/fail exactly**; **soft tests must stay above floor** (a dip that holds the floor is
  noise, not a break). Also compare *failure reasons*, not just verdicts. This is what let the
  migration assert behavior-preservation across a 44-pass/5-fail suite with confidence.

- **Baseline-roster-before-refactor.** Capture the full roster *before* touching anything, so
  "it went red" is attributable. Under model nondeterminism a red at the end of a 14-minute run
  is ambiguous between "my change" and "pre-existing / drift" unless you have the before-state.
  Cheap discipline, disproportionate payoff.

- **`ctx.Env` as the carrier for "flag defaults that look like state."** Values a Given sets and
  an action reads as CLI flags (here: data-dir, workbook path) are *not* scenario state and don't
  belong as forked `Context` fields — they belong in `ctx.Env`, with domain accessors, forwarded
  to the subprocess. This keeps the module `Context` domain-free while giving consumers a typed
  seam. (A domain wrapper struct around `Context` is structurally blocked — `Run` hands
  `*harness.Context` to the callbacks, so a wrapper can't be the parameter type.) Worth a note in
  ADOPTION.md §5 beside the existing `ctx.Env` mention.

---

## 5. What expense-reporter does NOT need (so the report stays honest)

- **Concurrent assertion executor / circuit breaker** (§6, from the acceptance-test quarry). At
  ~50 tests the suite's cost is Ollama latency, not assertion dispatch; a worker pool buys
  nothing here. This is a scale feature for a tens-of-thousands-of-assertions consumer.
- **`.ljson` line-numbered expected files** (§6). Per-file expected fixtures are sufficient at
  this size.
- **Order-insensitive JSON / status-only assertions** (§6). No expense assertion needs array
  order normalization today.

If a future addition can't name an expense-reporter (or career-search) use-case, it's a
speculative-surface candidate, not a gap.

---

*Scope note: this report deals in test mechanics and API usage only. It contains no expense
descriptions, taxonomy contents, or other domain data — those stay in the consumer repo.*
