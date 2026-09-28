# kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong

## Why

KAN-820 (flow-cost, from kan-743-step-1's deferred self-review): a fix dispatch went out on the
wrong model because the operator's model policy lives in memory and narratives — nothing
machine-readable carries it, so the dispatcher resolved a default and the mismatch cost a wasted
dispatch plus a prose correction.

Reachability checked at the resolved base `4a278320` — the finding still reproduces:

- the only machine-readable model policy the dispatcher consults is the settings store's
  harness-wide `defaultModel` (`stats/internal/store/settings.go`, resolved per run by
  **Model resolution**, `skills/flow/SKILL.md`);
- `<project>/.flow/project.md` carries a model key only for the archive-phase self-review
  (`## self review model`); no dispatcher reads a per-project model key;
- a run-scoped override still arrives as a plain-language session instruction, recorded in
  prose (`decision.json` `overrides`, dispatch context) — visible after the fact, never
  checked against anything.

## What changes

- New optional `## model` key in `<project>/.flow/project.md`: one `ValidModels` member,
  validated like `## self review model` (byte-for-byte body match, reported-by-name-and-drop).
- `DEFAULT_MODEL` resolves per run: project `## model` → settings store `defaultModel` →
  literal `opus`. Governed roles: implementer, fixer, panel dispatches, rerun pair — unchanged
  set, new first source. `VERIFY_MODEL`, `SELF_REVIEW_MODEL`, `REVIEWERS` and effort selection
  untouched.
- Resolution is surfaced and recorded: it names model + source once at resolution; the Decide
  preamble `models:` line and the run summary carry both; `decision.json` gains
  `resolved: {"model": …, "source": …}`. No enforcement gate — nothing blocks.
- The zcode harness mapping still replaces the model at dispatch; the key governs the
  pre-mapping value and the dispatch ledger keeps recording the model actually run.
- This repository's own `.flow/project.md` declares `## model opus` (equals the store default
  here — exercises the path, changes no behavior).
