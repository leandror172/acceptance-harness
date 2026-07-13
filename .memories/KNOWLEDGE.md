# acceptance-harness — Knowledge (Semantic Memory)

*Accumulated decisions. Read on demand by agents.*

## Extraction Provenance (2026-07-13)
Extracted from `expense-reporter/test/harness` (the origin) **via career-search's
de-domained copy** (`dashboard/test/harness`, the seed) — deliberately NOT from the
origin. The second consumer had already shaken out the expense-shaped leaks
(shallow-only fixture copy, LLM-typed FixtureConfig, workbook Context fields, Ollama
gate in core), so its copy encoded two consumers' needs where the original encoded one.
**Rationale:** extracting before a second consumer existed risked freezing the first
consumer's accidental shape into the "generic" lib.
**Implication:** the same discipline applies forward — new core surface requires a
consumer that needs it (see Roadmap), not an anticipated one.

## Design Decisions (v0.1.0)
- **LLM-free/network-free/git-free core.** Soft assertions, accuracy floors, drift
  tracking, liveness gates are LLM coping mechanisms — consumer- or extension-package
  material, never core. Even log strings must not name a domain dependency (the
  "waiting for Ollama" leak was scrubbed from a log line, not a file).
- **No build tag in the lib** (both consumer copies carried `//go:build acceptance`;
  stripped at seed time). A tagged library is invisible to `go vet` / `go test`.
- **Retention knobs: env vars over package-level flags.** `flag.Bool` at package init
  can collide with consumer flags; `RegisterFlags()` (sync.Once-guarded) is the opt-in.
- **`os.MkdirTemp` + manual cleanup in Run** (not `t.TempDir()`) — honors
  keep-artifacts/keep-on-failure inspection of failed scenarios.
- **`Context.T` is `*testing.T` by design.** The predecessor framework's history shows
  the CLI-runner idea dies; do not abstract T speculatively.
- **verify split cli/json/file** — the origin mixed generic + domain in one file, which
  made the boundary hard to find during extraction.

## Bugs the Lib's Own Tests Caught (2026-07-13)
`Run()` used `os.MkdirTemp("", t.Name())`; subtest names contain `/` — illegal in a
MkdirTemp pattern → crash for any consumer calling Run inside `t.Run`. Latent in BOTH
consumers' copies; fixed by sanitizing the pattern.
**Implication:** the §3.8 rule (lib has its own unit tests, consumers are not its only
safety net) is validated; keep tests for every new helper.

## Consumer Map (2026-07-13)
| Consumer | State | Notes |
|---|---|---|
| career-search `dashboard/test/` | REPOINTED (TC-EXT closed) | 65 tests green; domain verifiers at `test/tracker/`; RegisterFlags in TestMain |
| expense-reporter `test/` | PENDING (Session B) | plan §5: keeps ollama.go + comparator.go in its domain layer; Context workbook fields → Env or wrapper |
| latent-topic-graph | PATH DOCUMENTED ONLY | docs/LTG-PATH.md; Python-native runner remains a valid alternative |

## Versioning Policy
v0.x until expense-reporter migrates (the last boundary test). Breaking changes
preferred over compat shims while both consumers are same-author and pin exact
versions. Tag discipline: consumers' suites smoke-tested before any tag.

## Roadmap (deferred until a consumer demands it)
Polling `WaitFor(ctx, timeout, interval, predicate)` (async-pipeline consumers);
bulk concurrent assertion executor + circuit breaker; `ignore_order_on` /
status-only / epsilon JSON assertion options; input-line traceability;
`MANUAL` step convention (docs-first); `extern/llm` extension package (promotes
expense-reporter's ollama gate + soft assertions when a second LLM consumer appears).
Source analysis: expenses repo `.claude/plans/harness-extraction-acceptance-test-comparison.md`.
