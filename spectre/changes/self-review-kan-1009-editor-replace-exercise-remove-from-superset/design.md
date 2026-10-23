## Context

Self-review pass over gymie's `kan-1009-editor-replace-exercise-remove-from-superset` run
(session token `mf-k1009b`, parent transcript
`~/.claude/projects/-Users-tweety53-Projects-gymie/9fe58c09-fded-4f35-bacc-6270f79bd1a9.jsonl`,
its subagents under `…/9fe58c09-fded-4f35-bacc-6270f79bd1a9/subagents/`). Six findings, none
`big`; every one is planned below and none is disproved. Store rows are read with
`flow record dispatches -change kan-1009-editor-replace-exercise-remove-from-superset -C /Users/tweety53/Projects/gymie-worktrees/kan-1009-editor-replace-exercise-remove-from-superset`.

### 1. Verify rows keyed with an empty repository suffix

Trace: the parent's first verify (11:11:53Z) ran
`for w in B F; do flow record dispatch begin … -key verify-$(basename $(dirname ${!w})) …; done`
in zsh, where `${!w}` is a bash-only indirection: zsh printed `bad substitution`, the key became
`verify-`, and both iterations recorded the same row (`recorded: dispatch 16` twice). Its ends,
keyed `verify-gymie-worktrees` and `verify-gymie-frontend-worktrees`, found no row; the parent
closed `verify-` blocked. Every later round then typed `verify-fix-<k>-` literally, because
`check-verify-green.sh` groups rows by key with `-fix-<k>` removed and only a `verify-fix-<k>-`
row could supersede the blocked `verify-` group. `stats/cmd/flow/record.go`'s begin validates
panel, visual-verify and implementer keys but never a trailing `-`.

Rows the defect wrote, and the key each should carry — derived from the worktree each row's
command ran its list in (parent transcript, Bash calls at the times shown):

| id | key now | ran | key after repair |
|----|---------|-----|------------------|
| 7290 | `verify-` | 11:11:53Z, both lists, first iteration `B` | `verify-gymie` |
| 7308 | `verify-fix-1-` | 13:15:07Z, `$F test` — a re-run after 7306 | `verify-fix-1-2-gymie-frontend` |
| 7315 | `verify-fix-2-` | 14:19:57Z, `$F lint` + `$F test` | `verify-fix-2-gymie-frontend` |
| 7321 | `verify-fix-3-` | 15:04:02Z, `$F lint` + `$F test` | `verify-fix-3-gymie-frontend` |
| 7334 | `verify-fix-4-` | 16:21:28Z, `$B lint` + `$B test` + one playwright spec | `verify-fix-4-gymie` |

No repair command exists for a key (`flowd` carries `restamp-dispatch`, `retime-dispatch`,
`fold-dispatch`, `unflag-verdict`), so one is added.

### 2. The full-app-suite zip written into the caller's cwd

`stats/internal/guard/rebuildfullappsuitezip.go` builds `zipPath := args[1] + "/full-app-suite.zip"`
and runs `zip` with `Dir: env.Dir`, so a relative `<screenshots root>` lands under the caller's
working directory. Every caller's contract already says the root is "`screenshots` resolved the
way `resolve-visual-screenshots.sh` resolves it" (`skills/flow/visual-verify-verifier.md` step 8),
an absolute path; gymie declares `screenshots` as `.`, and a call passing it verbatim from the
gymie main checkout left `/Users/tweety53/Projects/gymie/full-app-suite.zip` (untracked, in the
main checkout's status). That stray file is the operator's to remove — the main checkout is off
limits to this pass.

### 3. Frame-list guard over- and under-declares

Reproduced against the kan-1009 design and gymie's mockups root
(`scripts/check-frame-list.sh <design.md> <gymie wt>/docs/design/mockups <a report with no frames>`):
the guard declared 4 frames — `LM-R1`, `LM-R10`, `LM-R3`, `LM-R9` — though the design names the ten
`LM-R1 … LM-R10`; the range `docs/design/mockups/LM-R1.png … docs/design/mockups/LM-R10.png` is
read as its two ends. The run's plan-time text of decision `frame-ids-lm-r`'s `**Considered:**` line
read "`R1.png`/`R2.png` already exist in the mockups root for another feature", which declared the
unrelated frames `R1` and `R2`; the parent then reworded it to `R<n>.png` to get past the guard.
<!-- measured: scripts/check-frame-list.sh on the archived kan-1009 design.md and docs/design/mockups @ gymie fce43a04 → "below 4 declared" -->

### 4. Unmeasured dispatches from #17 on

`flow record cost-status` reads `24 unattributed — not measured: 24`.
<!-- measured: flow record cost-status -change kan-1009-editor-replace-exercise-remove-from-superset @ store read 2026-10-23 -->
Trace: from 11:20:06Z every begin of a launched agent carried an invented `-agent-id` —
`task-11`, `task-11-rev`, `panel-2-extra`, `visual-verify`, `visual-verify-fix-3-2` and the like
(parent transcript's Bash calls). `-agent-id` is "recorded intent the daemon never overwrites"
(`skills/flow/implement.md`), and the harvester stamps a launch only onto a row whose agent id is
empty (`StampDispatchAgent`), so 21 subagent rows never met their transcripts. The other three
unmeasured rows are inline (7307, 7308, 7334): each began and ended inside one Bash call, padded
with `sleep 2`/`sleep 3`, so no assistant record falls in its window. The harvest commits
`72088308`, `1fb982b0`, `892e9916` landing at 11:33Z are not the cause: the invented ids precede
them (11:20:06Z) and every later row is unstamped for the same reason. The CLI checks
`-agent-id` only against the store's charset and placeholder words (`validateAgentID`), which a
dispatch label passes.

Rows 10, 11, 14 and 15 (`output not measured`) are the harness, not the harvester: their agent
transcripts write most messages only with `"stop_reason": null` and a first-chunk output count
(e.g. `agent-aaa9fbb57bed1a56d.jsonl`), which the harvester counts as unmeasured by design
(kan-920, `Usage.OutputUnmeasured`); the parent's task notification carries only a total
(`<subagent_tokens>`), no output figure. Nothing derives their output; they are left as they are.

Repair map, each subagent row to its agent transcript (meta `description` and launch time):

| id | key | agent |
|----|-----|-------|
| 7294 | task-11-implementer | a30f48a5e14b11392 |
| 7295 | task-11-reviewer | a721416b5c2a6fcf4 |
| 7297 | panel-2-primary | ac48ed3052e52ccec |
| 7298 | panel-2-test-sanity+call-efficiency | a343f5bc00248dd18 |
| 7309 | visual-verify-gymie-playwright | ae9f09c3949fc0d17 |
| 7310 | visual-verify-rest-gymie-playwright | ae00ec334d58fb969 |
| 7311 | task-12-implementer | af31af1511bec0a75 |
| 7312 | task-12-reviewer | aec2404d48dd3b55c |
| 7313 | panel-3-primary | a9f4ea45f2739a94a |
| 7314 | panel-3-test-sanity+call-efficiency | a395d3563d7f75f48 |
| 7316 | visual-verify-fix-2-gymie-playwright | af12aad0d787e4961 |
| 7317 | task-13-implementer | aa4821d3bcd77b88a |
| 7318 | task-13-reviewer | a14c5b9ed9017da4f |
| 7319 | panel-4-primary | a6c0c76621b4fb31c |
| 7320 | panel-4-test-sanity+call-efficiency | a594930fdc3fb33eb |
| 7322 | visual-verify-fix-3-gymie-playwright | aa572b9ce0042dff8 |
| 7323 | visual-verify-fix-3-2-gymie-playwright | aef6a703059daa4e2 |
| 7324 | visual-verify-fix-3-full-gymie-playwright | a8dff3d66a557a7a3 |
| 7331 | task-14-implementer | ac4f440c2eb1cf217 |
| 7332 | panel-5-primary | af4ae6359a567c2c2 |
| 7333 | panel-5-test-sanity+call-efficiency | a165814e413169e5d |

Every agent id the machine's Claude Code transcripts hold is `a` and sixteen hex digits; ZCode's
is `agent_<uuid>` (`harvest.AgentIDFromRolloutPath`).
<!-- measured: ls ~/.claude/projects/-Users-tweety53-Projects-gymie/*/subagents/ | agent-*.jsonl names against ^a[0-9a-f]{16}$ → 2668 match, 0 other @ 2026-10-23 -->

### 5. The first visual verifier left step 10

Trace (`subagents/agent-ae9f09c3949fc0d17.jsonl`): the verifier ran verify (94 mapped specs,
14.9 min) and capture, wrote its report, ran `check-verify-report.sh` on it at 13:37:15Z — exit 1,
`10 not done` — and handed back 33 s later with "This is left work, not a blocker; it needs a
remainder dispatch". Nothing failed and nothing ran out: no error, 14.8M tokens left. Its prompt's
REPORT FORMAT paragraph (`skills/flow/visual-verify.md`) says "a report it rejects costs a
remainder dispatch" — the wording the handback echoes — while the rule against handing back
undone steps (`skills/flow/visual-verify-verifier.md`, **No required step is left undone for want
of time or context**, landed `b9d1709b` before the run) sits only in the file the prompt points at.
The remainder (`visual-verify-rest-gymie-playwright`) ran 9m34s. Finding 6 lengthened the verify
step that preceded the stop.

### 6. specs-for-diff falls back to every spec

Reproduced in the playwright worktree: `node scripts/specs-for-diff.mjs <frontend wt> a6b3e80e`
prints 94 lines. A traced copy of the walker (scratch, not committed) names each dead end:
<!-- measured: node scripts/specs-for-diff.mjs /Users/tweety53/Projects/gymie-frontend-worktrees/kan-1009-editor-replace-exercise-remove-from-superset a6b3e80e | wc -l @ gymie-playwright 4c62dc9 → 94 -->

- Comments are walked as code: KDoc in `ui/friends/FriendsState.kt` names `flashes`, so
  `DraftStructureEdits.kt` reaches the friends screens; line comments in `ActiveExerciseCard.kt`
  name `ActiveExerciseIconControls`.
- `declared()` reads `fun List<BlockAction>.hairlineIndex()` (`BlockHeader.kt`) as declaring
  `List`, and does not know `suspend`, so `Prefetch.kt`'s `suspend fun prefetchOwnData` has no
  caller though `ui/AppScreens.kt` calls it.
- A caller file is followed through every declaration it holds, not only the ones that use the
  callee: `WorkoutEditorSelectors.kt` reaches `BroadsheetButton.kt` through `FOCUS_RING_WIDTH`
  and from there every screen.
- A file no source calls (`ActiveExerciseAnnotations.kt`, `ReorderArrows.kt`, `BroadsheetTag.kt`)
  is read as a branch that "reaches specs nothing here can name" and returns every spec, though it
  draws nothing.

With all four fixed in the traced copy the same diff maps to 51 specs. Left out one at a time:
without comment stripping, the generic-receiver read or the dead-code rule a touched file still
dead-ends and the script prints all 94; with a file-level walk it maps to 89. The `suspend`
modifier changes nothing on this diff once comments are stripped, and is fixed because a touched
file whose only declarations are `suspend fun`s would otherwise find no caller at all.
<!-- measured: traced copies of specs-for-diff.mjs with one rule removed each, frontend diff a6b3e80e @ gymie-playwright 4c62dc9 → 51 all rules; dead end without stripping, receiver or dead-code rule; 89 file-level -->
<!-- measured: t9.mjs without suspend in the modifier list @ gymie-playwright 4c62dc9 → 51 -->

The file lives outside `<agents repo>`: per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) its fix is committed on the change's own branch
`spectre/kan-1009-editor-replace-exercise-remove-from-superset` in the playwright worktree
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1009-editor-replace-exercise-remove-from-superset`,
where it rides kan-1009's landing route.

## Decisions

### Where the trailing-dash refusal lives

**ID:** key-trailing-dash-refused-at-begin
**Status:** active
**Chosen:** `runRecordDispatchBegin` refuses a `-key` ending in `-` with exit 2, naming the key;
`dispatch backfill` routes through begin and inherits it, and `end` needs none since no row can
carry such a key.
**Considered:** refusing in the store's `RecordDispatch` — the API and reconcile paths would then
need an error type and HTTP mapping for one shape the CLI already sees first; refusing any key
segment that is empty (`verify--gymie`) — no run has written one, a speculative case.

### How the five verify rows are repaired

**ID:** rekey-dispatch-command
**Status:** active
**Chosen:** `flowd rekey-dispatch <dispatch-id> <key>` (`store.Store.RekeyDispatch`), beside the
existing repair commands, run once over the five rows of **Context §1**; row 7308 takes
`verify-fix-1-2-gymie-frontend`, the second-dispatch-in-a-round shape `visual-verify-fix-3-2-…`
already uses, because `verify-fix-1-gymie-frontend` is row 7306's and the key is unique per run.
**Considered:** leaving the rows — the rule that a fix repairs what its defect wrote forbids it;
folding 7308 into 7306 — `fold-dispatch` moves a diff base between verifier rows and would need
widening for a row that has none.

### Where an invented agent id is refused

**ID:** agent-id-shape-at-cli
**Status:** active
**Chosen:** `flow record dispatch begin` and `end` refuse an `-agent-id` that is not `inline`,
`none`, a Claude Code id (`^a[0-9a-f]{16}$`) or a ZCode id
(`^agent_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), exit 2 with a message
telling the caller to omit the flag and let the daemon stamp the launch.
**Considered:** tightening `store.validateAgentID` to the same shapes — the store's tests write
`a1`, `agent-x`, `agent-one` and the like across six test files, and `StampDispatchAgent` stores
harness ids unvalidated by design, so the shared function is the wider change for no extra case
caught; refusing only an id equal to a prefix of the key — `panel-2-extra` and `task-11-rev`
are not prefixes of their keys.

### Repairing the unmeasured rows

**ID:** restamp-21-rows
**Status:** active
**Chosen:** `flowd restamp-dispatch <id> <agent transcript>` over the 21 pairs of **Context §4**,
and `flowd restamp-dispatch <id> <parent transcript>` over the inline rows 7307, 7308 and 7334;
an `ErrNoInlineShare` there is the expected result for a window inside one Bash call, reported and
left. Rows 10, 11, 14 and 15 are left: no source holds their output.
**Considered:** re-harvesting the whole session — the watcher would still skip the rows whose
agent id is set.

### Relative screenshots root

**ID:** relative-screenshots-root-refused
**Status:** active
**Chosen:** `rebuild-full-app-suite-zip` exits 2 on a relative `<screenshots root>`, naming it and
the resolution the callers' contract requires; nothing is removed or written.
**Considered:** resolving a relative root against `<worktree>` — `screenshots` resolves against
the project's `regression checkout`, not necessarily `<worktree>`, so the guess could write to a
second wrong place.

### Frame ranges and rejected alternatives

**ID:** frame-list-ranges-and-considered
**Status:** active
**Chosen:** before token matching, drop every design line opening (after whitespace) with
`**Considered:**`, and add to the declared set every `<prefix><n>` for `<n>` from `<a>` to `<b>`
whose `<prefix><n>.png` is in the mockups dir, for each `<prefix><a>` and `<prefix><b>` with the
same prefix joined by `…`, `–` or `...` — each end optionally a path ending in `/`, a `.png`
suffix and backticks.
**Considered:** declaring frames only from a dedicated `## Frames` list — every existing design
cites frames in prose, so it would reopen every archived design's reconciliation.

### The verifier's prompt carries the no-undone-step rule

**ID:** report-format-forbids-early-handback
**Status:** active
**Chosen:** rewrite the REPORT FORMAT paragraph of `skills/flow/visual-verify.md` to tell the
verifier to run `check-verify-report.sh` on its own report before handing back, finish every step
an exit 1 names in the same dispatch, and hand back incomplete only for a named blocker it cannot
cure — time, context and a found departure being none; the "costs a remainder dispatch" clause
goes.
**Considered:** a `PreToolUse` hook refusing the verifier's handback while its report is
incomplete — it would have to recover the report path from the dispatch prompt, and the handback
tool is one harness's; a mechanical answer the prompt fix makes unnecessary until a verifier
ignores the rule stated in its own prompt.

### specs-for-diff walks declarations, not files

**ID:** specs-walk-declarations
**Status:** active
**Chosen:** in `scripts/specs-for-diff.mjs`: strip `/* … */` and `//` comments from every Kotlin
text before it is matched (a `//` preceded by `:` stays, so URLs survive); `declared()` accepts
the `suspend` modifier and a generic receiver (`fun List<BlockAction>.hairlineIndex()` declares
`hairlineIndex`); a caller is followed only through
its top-level declarations whose text uses the callee's names; and a UI file whose names no
frontend Kotlin source outside the test sources references is dead code, its branch ending with
no specs, while a UI file referenced only from outside the UI tree still selects every spec.
`specsFor` takes the non-UI sources as a new trailing argument; the CLI passes
`git ls-files '*.kt'` minus the UI tree and test sources.
**Considered:** a Kotlin parser — a dependency for a walker the four rules bring from 94 to 51;
stripping comments alone, as the finding first proposed — the traced copy still falls back to
every spec.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** two groups, in series. `g1`: Tasks 1–5, the `<agents repo>` worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-1009-editor-replace-exercise-remove-from-superset`.
`g2`: Task 6, the playwright worktree
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1009-editor-replace-exercise-remove-from-superset`,
on branch `spectre/kan-1009-editor-replace-exercise-remove-from-superset`.
**Considered:** one fixer for all six — its commits would span two repositories and two branches;
one per task — Tasks 1 and 2 share `record.go` and must run in series anyway.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the pass changes a CLI refusal, a store write and a dependency walker
whose misses select fewer specs than a diff reaches, and it repairs live store rows.
**Considered:** flow-low — a walker that under-selects specs fails silently, worth the deeper read.

## Open questions

## Live check

Against the running flowd, from this worktree's `stats/`, before landing:

- `go run ./cmd/flow record dispatch begin -change kan-1009-editor-replace-exercise-remove-from-superset -C <gymie wt> -role verifier -model opus -effort low -agent-id inline -key verify- -session-token mf-livecheck`
  exits 2 and `flow record dispatches` for the change shows no row with session `mf-livecheck`.
  Not working: exit 0 and a new row.
- The same with `-key verify-gymie -agent-id task-11`: exit 2, no row. Not working: a row whose
  `agentId` is `task-11`.
- After Task 1's and Task 2's repair steps: no row of the change has a key ending in `-`, and
  `flow record cost-status -change kan-1009-editor-replace-exercise-remove-from-superset` drops
  from `24 unattributed` to at most 3 (the inline rows). Not working: 24 still.
- `scripts/rebuild-full-app-suite-zip.sh <playwright wt> .` run from a scratch directory exits 2
  and leaves no `full-app-suite.zip` there.
