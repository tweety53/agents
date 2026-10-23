## Context

- Self-review fix change for gymie's `kan-1026-stored-data-first-on-every-open`, planned per item 1
  of **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`) on branch
  `self-review-kan-1026-stored-data-first-on-every-open` off agents `origin/main` (`c1282bc4`), in
  worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1026-stored-data-first-on-every-open`.
  Seven non-`big` findings, each carried by one task in `tasks.md`. The brainstorm design is
  `.superpowers/sdd/2026-10-23-self-review-kan-1026-stored-data-first-on-every-open-design.md` in
  this worktree. Every question was auto-resolved to its recommended option (**Auto-resolution**,
  `skills/flow-contracts/operator-prompts-auto-resolution.md`).
- No path below changed on `origin/main` since the run. `spectre/changes/archive/` records one
  earlier meeting with A1-1's shape: the kan-1000 self-review chose description-key stamping
  (`launch-stamp-by-description-key` there), and the description rule already tells every
  non-task row to open with its key; nothing enforced it, which Task 1 adds.
- The kan-1026 rows live in the dev flow store (`flow-postgres`, `localhost:5433`), read with
  `flow record dispatches -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open`.
  Row numbers below are the rows' `seq`; the repair commands take the row's `id`. The run's parent
  transcript is
  `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/50d27d94-f31d-4308-90ff-eef33cdbaa77.jsonl`,
  its agents under that path without `.jsonl`, in `subagents/`.
- Every reworded or added run-loaded sentence is listed in this change's `verbatim-moves.txt`;
  `scripts/check-verbatim-moves.sh` exits 0 with all seven tasks' diffs applied.
- A1-1's pass report counted 15 wrong rows; the store holds 12 rows with no agent and 2 carrying
  another dispatch's, all 14 repaired by Task 1.
  <!-- measured: flow record dispatches -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open (rows with empty agentId: seqs 37 38 52 57 58 61 62 64 68 69 73 74) @ live store, 2026-10-23 (not re-runnable once repaired) -->

## Decisions

### Launch descriptions are enforced, not re-paired

**ID:** enforce-launch-description-key
**Status:** active
**Chosen:** `hooks/enforce-agent-baseline.py` denies, in a session whose transcript records a flow command, an Agent call whose `description` opens neither `Task <x>/<n>` / `Tasks <a+b>/<n>` nor a key-shaped word (lowercase, no spaces); a prompt carrying a `Key: self-review-` / `Key: pipeline-fix-` line is exempt, being stamped by that line. The harvester's keyed stamp then names every non-task row, a daemon restart included (round 6: flowd restarted between the round's begins and its launches, and the watcher's pending begins live in memory only). `skills/flow/implement.md`'s description paragraph names the hook; `skills/flow/review-panel.md` cites the rule where it states the stamping.
**Considered:** a phantom begin — a `$`-key begin claims its launch and stamps nothing — spiked: it fails three kan-925/kan-964 fixtures (`unresolvable template key claims no launch`, `backtick substitution is not a literal session token`, `for inside quotes is prose beside a real begin`) that pin the opposite rule, and with keyed descriptions enforced an orphan launch is stamped before time pairing runs; persisting the watcher's pending begins across restarts — state for a fallback path the keyed stamp makes rare; checking `skills/flow` in the prompt to detect a flow session — 29 of kan-1026's 41 launch prompts name no flow skill.
<!-- measured: python3 over the parent transcript's Agent tool_use prompts, counting those without "skills/flow" @ machine-local transcript, 2026-10-23 -->

### Time pairing keeps its launch-first order

**ID:** pairing-order-unchanged
**Status:** active
**Chosen:** no change to `stampDispatchAgents`' candidate sort. Cause (d) is not a defect of the order: a nearest-first sort was spiked and fails `TestWatcherRetiresBeginsOfHookDeniedLaunches` (kan-995's recorded case, where the begin's own launch is the earlier one) and `TestWatcherPairsPanelRoundBeginsRecordedAfterTheLaunches`; row 53 took row 52's launch only because row 52's `$R` begin left that launch unpaired, which `enforce-launch-description-key` removes.
**Considered:** nearest-first ordering, as the pass proposed — the two recorded cases above break.

### A lone begin pairs at its row's started_at

**ID:** anchor-lone-begin-at-row
**Status:** active
**Chosen:** `stampDispatchAgents` reads each lone begin's row `started_at` through a new `DispatchStartedAt(sessionToken, key)` on `harvest.DispatchAgentStamper` (store: the latest row under that pair) and pairs at it; a begin sharing its Bash call with others keeps the call's instant, which is what pairs a round's begins with its launches in order; a row not yet written keeps the call's instant and is read again next batch. Row 76's begin closed a three-minute call and took a launch made before it; the new watcher test stamps exactly that orphan without the anchor.
**Considered:** anchoring every begin at its row — begins of one call get distinct instants milliseconds apart, and the launch-first sort then hands the round's first launch to its last begin; the Bash call's result instant — the transcript reader carries no result timestamp per command, a parser change for the same effect.

### panel-round-begins.sh records a round's begins

**ID:** panel-round-begins-script
**Status:** active
**Chosen:** a flow-guard command `panel-round-begins` (`stats/internal/guard/panelroundbegins.go`, shim `scripts/panel-round-begins.sh`, skill symlink `skills/flow/scripts/panel-round-begins.sh`) taking `<worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...`, refusing a non-numeric round, and running one `flow record dispatch begin` per slot under `panel-<round>-<slot>`; `skills/flow/review-panel.md` names it where the round's begins are recorded.
**Considered:** one script opening the whole round — snapshots, markers, prompt rendering and begins — the begins run after the launch message and the snapshots before it, so it would be two modes of one script, and each of the other steps is already one script call; a bash-only script — every scripts/ helper is a flow-guard shim with a Go test.

### A stamped row ends at its agent's last write

**ID:** end-from-agent-transcript
**Status:** active
**Chosen:** flowd's end handler, given an end with no `endedAt` that closes a row carrying a launched agent's id, closes it again at that agent transcript's latest record instant when that instant falls inside the row's window (`api.WithAgentEnds`, wired in `cmd/flowd` over `harvest.AgentTranscriptEnd` on the transcripts root); `skills/flow/review-panel.md` drops `-ended-at "$(date -u -r <report-file> …)"`. The rows' existing ends came from report mtimes, close to the agents' finishes, so no row is retimed.
**Considered:** the CLI reading the transcript — it holds no agent id until the end's answer, so two round trips from the caller; a store-side lookup — the store knows no transcripts root.

### A worktree's visual-verify commit rides Branch backup

**ID:** worktree-commit-rides-branch-backup
**Status:** active
**Chosen:** `skills/flow/visual-verify.md` and `skills/flow-contracts/project-configuration-visual.md` say a commit in a resolved worktree of the `regression checkout` is on `spectre/<name>` and is pushed under **Branch backup**; only a main-checkout commit prints the hand push. `scripts/check-visual-verification.sh`'s comment states the identity assertion only and stays as is; `skills/flow-contracts/git-boundaries.md` already pushes every commit the run makes on the branch and stays as is.
**Considered:** pushing the regression checkout from visual-verify — `no-automatic-push` (kan-171) rejected any push a repository file authorises.

### A non-shell heredoc body is judged by its git invocations

**ID:** program-body-git-only
**Status:** active
**Chosen:** in `hooks/protect-main-checkout.py`, a heredoc fed to `python`, `python3` or `node`, or written to a file the rest of the command runs only through one of those or `npx playwright test`, is scanned only for the git invocations its lines name (quotes and brackets read as spaces, so `subprocess.run(["git", "-C", d, "commit"])` is still caught); the may-run-code test reads a plain `$(...)` as a command, expands the command's literal assignments, drops a leading `do`/`then`/`else`/`elif`/`!`, and reads `cd`, `echo`, `printf` and `render-slot-prompt.sh` as commands that never run a file. A process substitution, a backtick or a nested `$(` still fails closed, and a file written then run by a shell, by any launcher, or as a command keeps its body scanned. The hook's docstring states the rule; no skill sentence restates the heredoc rule, so none changes.
**Considered:** dropping the redirect scan for every heredoc body — a body a shell later runs would write unseen; a body-language guess from the target's extension — a file with no shebang that a launcher executes runs under `sh` whatever its name.

### The store refuses a token report on a harvested session

**ID:** tokens-refused-on-harvested-session
**Status:** active
**Chosen:** `store.EndDispatch` refuses an end carrying `-tokens` when any stage run under its session token ran on `claude-code` or `zcode` — the harnesses the watcher harvests — with `ErrTokensOnHarvestedSession` naming `flow record dispatch backfill -agent-id inline -commit <sha>`; the API answers 409 and journal replay retires it, as for the inline-window refusals; the CLI's environment check stays as the fast path. The two kan-525 store fixtures that report tokens now begin on `muse`, a harness no source harvests. Row 9's caller-reported figures were genuine and its window zero-width, so it is not re-recorded.
**Considered:** refusing on a bound `session_id` — `flow stage begin` binds it from `CLAUDE_CODE_SESSION_ID` or `MUSE_SESSION_ID` itself, and Muse writes no transcript the watcher reads.

### The bundle elides bands, seams and ink

**ID:** bundle-elides-raw-measurements
**Status:** active
**Chosen:** `elideMeasurementSections` replaces each section of `visual-verification.md` whose heading names bands, seams or ink (any level, any case), through the next heading of its level or higher, with `_(<n> lines elided: <heading> — in <record path>)_`, after `elideMatrixRows`; the header, `## Verify`, every verdict section and every matrix section's headings stay. kan-1026's record falls from 402 bundle lines to 202. `skills/flow-contracts/finish-contract-run1.md`'s bundle-shape sentence states it.
<!-- measured: a throwaway Go test calling elideMatrixRows then elideMeasurementSections on kan-1026's archived visual-verification.md, printing line counts @ branch self-review-kan-1026-stored-data-first-on-every-open -->
**Considered:** whitelisting the header, `## Verify` and verdict sections — a verifier names its verdict sections freely, and a departure under an unlisted heading would vanish; dropping whole matrix sections — `TestBundleElidesMatrixRowsUnderAnyMatrixHeading` keeps a departures table under a `## Matrix` heading, the case a section drop loses.

### The committed kan-1026 ledger is the parent's to refresh

**ID:** gymie-ledger-refresh-by-parent
**Status:** active
**Chosen:** Task 1 repairs the store rows; the archived `ledger.md` committed on gymie's `spectre/kan-1026-stored-data-first-on-every-open` is a gymie-branch write outside this change's repository, left to the parent once Task 1 lands: `flow record render -C <gymie worktree> -change kan-1026-stored-data-first-on-every-open -kind ledger -repo <gymie worktree>`, then the rendered ledger copied over that archive's `ledger.md`.
**Considered:** the fixer committing into gymie's worktree — the loop's branch is this repository's alone.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 1, 3 and 6 change the daemon's stamping, end and refusal paths over the shared store, and Tasks 1 and 5 change two hooks every session runs.
**Considered:** `flow-low` — every task carries a test that fails without it, but a misread of the anchoring or the hook's fail-closed rules mis-stamps rows or opens a write path unseen.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups, in series — `self-review-kan-1026-stored-data-first-on-every-open-fix-g1` carries Tasks 1, 2 and 3 (dispatch records; all three edit `skills/flow/review-panel.md`), then `-fix-g2` carries Tasks 4, 5, 6 and 7 (Task 6 follows Tasks 1 and 3 on `stats/internal/store/records.go` and `stats/internal/api/server.go`).
**Considered:** one fixer for all seven — one context over every touched file and both slow Go suites (`internal/guard`, `internal/store`); one per task — seven dispatches, three of them on one skill file.

## Open questions

## Live check

- The store repair is the live part, run by Task 1's last step against the dev store: before it,
  `flow record cost-status -C /Users/tweety53/Projects/gymie -change kan-1026-stored-data-first-on-every-open`
  prints `15 unattributed — not measured: 15`; after it the line reads 3 unattributed, the three
  inline verifier rows the pass recorded `noted`, and each of the 14 repaired ids carries its true
  agent and a `tokens` bag. Still 15, or any `FAILED <id>` line, is the repair not working.
  <!-- measured: flow record cost-status -change kan-1026-stored-data-first-on-every-open @ live store, 2026-10-23 (not re-runnable once repaired) -->
  <!-- predicted: the same command after Task 1's repair step -->
- The daemon paths — keyed stamping under the enforced descriptions, the lone-begin anchor, the
  agent-transcript end, the token refusal — run in the protected flowd only once the branch lands
  and `pipeline-fix.sh land` restarts it, which no step before landing may do. Their stand-in is
  each task's test against the real Postgres stack (`internal/store`'s tests open a throwaway
  database on it) and the hooks' tests feeding real payloads; a store test that skips for want of
  the stack is a null result, not a pass.
