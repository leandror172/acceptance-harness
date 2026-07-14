# acceptance-harness — Quick Memory

*Working memory for the repo. Injected into agents. Keep under 30 lines.*

## Status
v0.1.1 published (2026-07-13, public GitHub, MIT; v0.1.0 + go directive 1.24→1.25).
Seeded from career-search's de-domained copy; 30 unit tests green, vet/gofmt clean.
career-search repointed (TC-EXT closed, 65 tests green, on v0.1.1 / go 1.25.5).
**Next consumer: expense-reporter migration**
(Session B — plan §5 in expenses repo `.claude/plans/harness-extraction-plan.md`).
No v1.0.0 until that migration shakes the API.

## Structure
```
harness/   # engine: Context, Scenario, Run; fixtures; BuildBinary; retention knobs
verify/    # generic Then base: cli.go, json.go (dotted-path), file.go
docs/      # PATTERNS.md (methodology), ADOPTION.md (consumer wiring), LTG-PATH.md
```

## Key Rules
- **Core is LLM/network/git-free** — gate: `grep -ri ollama` returns nothing
- **No build tags in the lib** — tags are consumer policy
- **Tree copy is the default**, flat copy the special case
- **Domain fields decode from FixtureConfig.Raw** — never extend the core struct
- **Retention via env** (HARNESS_KEEP_ARTIFACTS/_ON_FAILURE) + opt-in RegisterFlags()
- **v0.x: fix the API, no compat shims** — consumers pin exact versions
- **Docs are first-class** — PATTERNS/ADOPTION regressions are net losses

## Deeper Memory → KNOWLEDGE.md
Extraction provenance · design decisions with rationale · consumer map · roadmap
