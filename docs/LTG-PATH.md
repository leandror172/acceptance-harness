# Adoption sketch — pipeline-shaped consumers (the LTG path)

*Documented, not built. This records how a corpus/pipeline tool whose
acceptance tests are currently hand-run probe records would adopt the harness.
The concrete case is the author's latent-topic-graph (LTG) project; the shape
generalizes to any trigger-a-run, inspect-the-outputs tool.*

## Why the fit is natural

LTG's acceptance tests follow a GIVEN → RUN → OBSERVE → ASSERT formalism with
named observation captures, plus an explicit `MANUAL` marker for the few
judgment-bearing steps. Its OBSERVE captures are already subprocess-shaped:
exit codes, stdout lines, files on disk, table/JSONL row counts — exactly what
`Context{ExitCode, Stdout, Artifacts}` holds. The harness's three-phase model
is the same formalism with OBSERVE folded into When (actions populate the
captures) and ASSERT spelled Then.

## The path

1. **A tiny Go module inside the consumer repo** (e.g. `acceptance/go.mod`)
   importing the harness. The consumer stays Python; only the test runner is
   Go. `actions/` shell out to the repo's blessed run wrappers (if the project
   mandates driving via shim scripts, the actions call the shims — never the
   internals directly). `verify/` asserts on exit codes, stdout, and JSONL/
   table row counts.
2. **Map the determinism vocabulary onto assertion tiers.** Structural-
   equivalence assertions (same files represented, same schema, all stages
   completed) are hard assertions. Negative resource-count checks — e.g. "an
   incremental no-op run made zero backend calls", asserted against
   instrumented counters — are `OutputContains`-class checks. Row-identity is
   deliberately NOT asserted where the underlying extraction is
   nondeterministic.
3. **Judgment steps use a `MANUAL` convention** (roadmap item, docs-first): a
   Then marker that logs "requires human judgment: <text>" and records rather
   than asserts. It keeps the automatable 90% automated without pretending the
   judgment steps are checkable.

## The alternative stays valid

A Python-native runner bound into the consumer's own inspection tooling is an
equally valid design the consumer's docs already imagine. The Go harness is an
option, not a mandate — decide in the consumer repo, not here.
