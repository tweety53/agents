# Finding themes

Canonical for the theme a self-review finding is recorded under, so two findings that share a cause
can be found together. Loaded by the pass (step 2 of `skills/flow-self-review/SKILL.md`) and by the
session that records its findings (step 5).

## The theme list

**The list is closed.** The pass gives every finding at most one theme, the tag below whose meaning
names the finding's cause — the cause, never the file, the guard or the angle it surfaced in. A
theme names a subject, not a verdict: a `flow-improvement` finding about something that went well
takes the same tag a `flow-fix` finding about its absence would.

**A finding that fits no tag is recorded with no theme, never under an invented one**, and never
under the nearest tag that does not name its cause. The pass's hand-back lists each such finding
as a theme miss, with the tag it would have wanted and that tag's one-line meaning, and step 7's
report carries those lines, so the list grows by an edit to this file and never by a pass.

**The theme is stored as a prefix of the note**: `-note '[theme: <tag>] <the finding, one line>'`,
ahead of everything else the note carries, an angle-3 note's trace included. A finding with no
theme carries no prefix. Read a theme's findings back with
`flow self-review findings -change <name>` and a match on that prefix.

| Tag | Meaning — the finding's cause is … |
|-----|-----------------------------------|
| `async-scope-lifetime` | async work run in a scope that ends before the work does — a cancelled coroutine, a disposed view, an affordance that left the screen |
| `async-ordering-race` | an async read or write not ordered against another one in flight, so the later result is overwritten by the earlier |
| `stale-derived-value` | a stored value derived from other data that not every writer of that data recomputes — a cached total, a fingerprint, a copied field |
| `writer-reader-mismatch` | a write and its reader that disagree on shape, key or presence — an input read that no code path produces, a field written that nothing reads |
| `error-path-conflated` | two distinct failures handled as one — a fallback for an unreachable dependency that also fires on its error answer, a timeout read as a refusal |
| `port-loses-behaviour` | a port, move or rewrite that drops a behaviour its original had, usually a failure path its copied tests never exercised |
| `class-fixed-at-instances` | a class of defect fixed only at the instances a finding cited, its other instances left in place |
| `test-separating-inputs` | a test whose inputs cannot tell the asserted rule from its nearest wrong alternative — a surviving mutant, an untested branch, a mutation slot's catch |
| `text-output-parsing` | a value read out of human-readable text — a trimmed prefix, unquoted paths, a shell expansion that mangles an argument |
| `guard-false-refusal` | a guard, hook or check that refuses correct work, or two that contradict each other or the contract they enforce |
| `guard-false-pass` | a guard, gate or stage mark that admits what it exists to refuse — a gate satisfied before its check ran, a case the check never reads |
| `contract-prose-drift` | skill, contract or brief prose that contradicts the code, another file or itself — a wrong key shape, path, command or a rule stated twice differently |
| `brief-rule-not-carried` | a rule that exists in one prompt or contract but never reaches the agent that needs it — a brief, slot prompt or fix-task template that omits it |
| `dispatch-recording` | a dispatch row opened, closed, keyed or credited wrongly — under the wrong change, left open, unmeasured, or begun by hand |
| `store-row-integrity` | store rows collapsed, overwritten, duplicated or recorded without a required field |
| `plan-premise-measurement` | a plan premise, step or diff measured — or not — before the plan gate: an unreachable step, a wrong cwd, a diff applied and linted on a scratch copy |
| `plan-scope-gap` | work the plan left untasked or mis-scoped — a file, fixture, parent action or test path no task names |
| `capture-state-coverage` | a visual capture that misses the state, region, viewport or fixture data the defect or the mockup lives in |
| `visual-baseline-drift` | a visual baseline or record gone stale or rewritten by something other than the change it belongs to |
| `visual-pass-sizing` | how a visual-verify pass is sized or scoped — sharding, the frame list a fix task names, a scoped pass over too many frames |
| `verdict-without-evidence` | a reviewer or verifier verdict raised, reversed or withdrawn without a reproduction or measurement — a false positive, an unmeasured departure |
| `reviewer-late-miss` | a defect on an unchanged hunk first raised in a later round or by a later slot — a delta-only re-review, a missing reviewer instruction, variance |
| `dispatch-grouping` | how dispatches are grouped, bundled or run in parallel — wave size, one group at a time, reviewers bundled by implementation group |
| `dispatch-budget` | a dispatch or slot past its time, cost or scope ceiling, or one given no budget at all |
| `load-sensitive-timing` | a wall-clock bound, timeout or unbounded call that fails or hangs under machine load or a slow dependency |
| `shared-state-leak` | a test or tool that reads or writes state it does not own — the live store, the installed binary, a shared worktree, another session's files, a leftover process |
| `worktree-sync-loss` | content lost or mis-resolved moving between trees — a sync conflict, a rebase, a cherry-pick, a throwaway copy removed with its files |
| `live-check-evidence` | a live check against a store copy or the running app, and what its before/after evidence proved or could not |
