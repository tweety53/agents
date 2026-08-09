# Self-review — kan-111-myflow-fast

**Rating:** 4/5

## Problems, and the pipeline change that would avoid them

Three real defects surfaced during the run, each caught before it could reach `main` broken, but
each also a symptom of a gap this pipeline could close on its own next time:

1. **Name collision, discovered late.** `/myflow-fast` collided with a retired five-state-pipeline
   command of the same name, only found when `scripts/check-vocabulary.sh` flagged it mid-implementation
   — after the name was already committed to the Jira issue, the brainstorming, and every planning
   artifact. Filed as **KAN-115**: check a proposed name against the retired-vocabulary list at
   `/myflow-start`, before the design assumes it.
2. **`openspec archive` failure from an undocumented convention.** A `MODIFIED` requirement's
   heading must match the capability's *current live* heading, never the change's new title — this
   change's delta spec used the new title, and the archive step failed until I read the CLI's own
   source to find the `RENAMED` `FROM:`/`TO:` syntax. Filed as **KAN-116**: document this rule and
   syntax in `/myflow-start`'s delta-spec instructions.
3. **Main-checkout cross-session contamination, twice.** A bare `git commit` after a pathspec-scoped
   `git add` still commits the *whole* index — and the main checkout is shared with a concurrently
   running peer session (`kan-100`, in its own worktree) that had files staged there before this run
   began. Caught and reconciled by hand both times, not by the pipeline. Filed as **KAN-117**: use
   explicit pathspecs on every main-checkout commit.

## Cost

This was an expensive run for what it produced: brainstorming reused an already-approved design
(cheap), but implementation dispatched roughly 15 subagents (7 planned tasks + Task 1a + 2 fix
rounds + a 3-slot final panel + a 2-slot targeted re-panel), across two `/myflow-do`-equivalent
sub-runs chained by `/myflow-fast` conceptually, plus this `/myflow-finish` run's own investigation
into the `openspec archive` failure. Two of the per-task review rounds (Task 2's fix, the final
panel's fix) were necessary — both found real defects. The `openspec archive` detour (reading CLI
source) was pure overhead that KAN-116 would eliminate for the next change touching a renamed
requirement. Sonnet-everywhere (this change's own recorded default) kept per-dispatch cost down
relative to the pipeline's Opus-implementer default, which is presumably why the operator chose it
for exactly this kind of change.

## What went well, and how to reproduce it

- Reusing the already-approved brainstorming design directly, on explicit operator instruction
  ("use the spec above"), skipped a redundant re-derivation entirely — worth doing again whenever a
  design was already worked out earlier in the same conversation.
- The `light` panel + Sonnet-everywhere combination caught four real findings (2 per-task, 2
  whole-branch) at a fraction of the cost a `full`/Opus panel would have run, and none of the
  findings were false positives — the roster did its job.
- The whole-branch panel caught two cross-file consistency findings (F1: model-recommendation
  contradiction between two files; F2: a citation pointing at guidance that didn't exist yet) that
  no per-task review could have seen, because each was individually correct within its own task's
  diff and only wrong once read together with another task's output. This is the concrete case for
  running the integration-level panel at all, not skipping straight from per-task review to handoff.
- Root-causing the `openspec archive` failure by reading the CLI's actual parser rather than
  guessing at delta-spec syntax avoided a second failed attempt.

## What could be automated or moved to a script

- A pre-flight check (could live in `check-vocabulary.sh` itself, or a new small guard) that scans
  a proposed change/command name against the retired-vocabulary list — this is exactly the kind of
  mechanical check a script should own rather than a person discovering it by running the guard and
  reading the failure.
- `openspec archive`'s heading-mismatch failure is already validated by the CLI itself (it refused
  cleanly with a clear message) — the gap is purely documentation, not tooling; KAN-116 covers it.
- The main-checkout pathspec issue (KAN-117) is a mechanical fix to the commit invocation itself,
  not a process change — a script-level guard could assert `git diff --cached --name-only` before
  each guarded commit matches only the paths that command's own staging pass just added.

## Findings filed

| Finding | Filed |
|---|---|
| Check proposed names against the retired-vocabulary list at `/myflow-start` | KAN-115 |
| Document the MODIFIED-heading / RENAMED-syntax rule in `/myflow-start`'s delta-spec instructions | KAN-116 |
| Use explicit pathspecs on main-checkout commits to avoid cross-session contamination | KAN-117 |

All three linked to KAN-111.
