# CLAUDE-DRAFT.md — acceptance-harness

> Tentative CLAUDE.md content, deliberately named otherwise. Decide whether to
> `claude init` and merge this in, or rename this file to CLAUDE.md as-is.

This file provides guidance to Claude Code when working in this repository.

## Project Identity

**Public Go library** (`github.com/leandror172/acceptance-harness`, MIT) — a
domain-agnostic Given/When/Then engine for acceptance-testing CLI tools, plus a
generic assertion base. Extracted 2026-07-13 from the expense-reporter harness
via career-search's de-domained copy (the second consumer that shook out the
generic/domain boundary).

**The methodology docs are first-class content.** If a change improves code but
degrades `docs/PATTERNS.md` / `docs/ADOPTION.md`, it's a net loss.

## Build & Test

```bash
go build ./...   # library builds plain — no build tags in this repo
go test ./...    # unit tests, deterministic, ~0.3s
go vet ./...
gofmt -l .       # must print nothing
```

## Hard Design Rules (v1 — from the extraction plan)

1. **Core is LLM-free, network-free, git-free.** No Ollama gating, no soft
   assertions/accuracy floors/drift tracking, no domain file formats. Gate:
   `grep -ri ollama` over the repo returns nothing — including log strings.
2. **No `//go:build acceptance` in the library.** Build tags are consumer
   policy; a tagged library can't be vetted or tested normally.
3. **Recursive tree copy (`CopyTreeToWorkDir`) is the default; flat copy is the
   special case.** Most CLIs under test have tree fixtures.
4. **Nested JSON from day one**; numeric comparison via `assert.EqualValues`,
   never `!=`.
5. **`config.json` is optional** — actions may pass flags directly. Consumer
   fields decode from `FixtureConfig.Raw`; never add domain fields to the core
   struct.
6. **No package-level flag registration.** Retention knobs read env vars
   (`HARNESS_KEEP_ARTIFACTS`, `HARNESS_KEEP_ON_FAILURE`); consumers opt into
   flags via `RegisterFlags()`.
7. **Keep `os.MkdirTemp` + manual cleanup in `Run`** — it exists to honor
   keep-on-failure; `t.TempDir()` can't.
8. **`Context.T` stays `*testing.T`.** This is a go-test library by design; do
   not abstract it speculatively.
9. **The core's own log/error text must never name a domain dependency** (the
   "waiting for Ollama" leak class).

## Versioning & Consumers

- **v0.x: breaking changes are cheap — prefer fixing the API over compat
  shims.** Both consumers pin exact versions. No stable v1.0.0 until
  expense-reporter has migrated (its fallout is the last boundary test).
- **Consumers:** career-search `dashboard/test/` (repointed, TC-EXT closed);
  expense-reporter `test/` (migration pending — plan §5 in expenses repo,
  `.claude/plans/harness-extraction-plan.md`); latent-topic-graph (documented
  path only, `docs/LTG-PATH.md`).
- After any API change, smoke the consumers' suites before tagging.

## Public-Repo Hygiene

- Never commit real expense/career/personal data — examples use a neutral toy
  CLI only.
- Roadmap (deferred until a consumer demands it): polling `WaitFor`, bulk
  assertion executor + circuit breaker, `ignore_order_on`/status-only JSON
  options, `MANUAL` step convention, `extern/llm` package. See the extraction
  plan §6 in the expenses repo.

## Per-Folder Memories

`.memories/QUICK.md` (working memory, ≤30 lines, read first) and
`.memories/KNOWLEDGE.md` (accumulated decisions with rationale). Update QUICK
on status changes; promote stable findings to KNOWLEDGE.
