## Context

- Eleven non-`big` findings from the kan-940 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task, numbered as the pass numbered the finding; none is dropped and none was
  disproved. Every premise was re-checked against this worktree at `04a3a161`:
  - 1 (Task 1): `pfLand` (`stats/internal/guard/pipelinefix.go`) ends with `git pull --ff-only` in
    the main checkout and removes the worktree; `refreshMainCheckout`
    (`stats/internal/guard/refreshmaincheckout.go`), which run 2 and `/flow-fast` call through
    `integrate-change.sh`, ends at `REFRESH-DONE:`. Neither, nor any skill, runs `make restart`
    (`grep -rn 'make restart\|stats restart' skills scripts rules hooks` finds nothing). Radius 8:
    `pipelinefix.go`, `refreshmaincheckout.go`, their two tests, the two script headers,
    `.flow/project.md`, and no skills prose.
  - 2 (Task 2): `closetask.go:79-80` prints `the dispatch end did not reach the store — continuing`
    first and the run returns 0; `TestCloseTaskExitContract`'s "a failed record never blocks" case
    pins exit 0. Radius 4: `closetask.go`, `close_task_test.go`, `scripts/close-task.sh`,
    `skills/flow/implement.md`.
  - 3 (Task 3): `skills/flow/verify-and-handoff.md:114-121` names `-role verifier -key verify -model
    … -effort … -agent-id inline` and no `-change` or `-session-token`; `recordUsage`
    (`stats/cmd/flow/record.go:253-257`) requires both. Radius 1.
  - 4 (Task 4): `finishStageIdentityFlags` (`stats/cmd/flow/stage.go:122-130`) reads `fset.NArg()`
    after one `Parse`, which stops at the first positional, so `<name> -C <wt>` leaves three
    positionals and returns "expected exactly one argument". `runTasksCount` (`tasks.go:189`) and
    `runTasksTick` (`tasks.go:121`) share the defect. `parseStateFlagsN` (`state.go:262-275`)
    already re-parses after each positional. Radius 5: `state.go`, `stage.go`, `tasks.go`,
    `stage_test.go`, `tasks_test.go`.
  - 5 (Task 5): `record.go:981` defaults `-effort` to `"default"`; `implement.md:124` and `:352`
    omit it. Nineteen `dispatch begin` call sites in `stats/cmd/flow/*_test.go` and three `-effort`
    occurrences in `record_test.go`. `loop.go:400` already passes `-effort`. Radius 3: `record.go`,
    `record_test.go`, `implement.md`.
  - 6 (Task 6): `record.go:1360` reads `-location` with no validation; `review-panel.md:727`
    prescribes `-location <file:line>`. Radius 2: `record.go`, `record_test.go`.
  - 7 (Task 7): kan-940's archived `visual-verification.md` is 126149 bytes, 108679 of them table
    rows under its four `####` matrix headings
    (`awk '/^####/{h=1;next} /^#/{h=0} h && /^\|/{n+=length($0)+1} END{print n}'`); the bundle
    copies the file whole (`archivedFiles`, `stats/internal/selfreview/git.go:454`) and runs
    `git log --stat` unfiltered (`git.go:376`, `:434`). Radius 3: `git.go`, `bundle_test.go`,
    `finish-contract-run1.md`.
  - 8 (Task 8): `checkDispatchEndFields` (`record.go:1242`), shared by `dispatch end` and
    `dispatch backfill`, accepts any well-formed `-tokens`; `EndDispatch`'s floor skips a row that
    carries one (`store/records.go:505-508`). Radius 2: `record.go`, `record_test.go`.
  - 9 (Task 9): `unfinishedwork.go:312` prints `` `flow record dispatch end -change … -key … -session-token
    … -outcome <outcome> -C …` `` and `verify-and-handoff.md:398` the same; neither carries `-ended-at` or
    `-commit`. Radius 3: `unfinishedwork.go`, `check_unfinished_work_test.go`,
    `verify-and-handoff.md`.
  - 10 (Task 10): `record.go:1090` stores `-commit` as given; `record_test.go` passes `abc1234` at
    four sites; `closetask.go:77` forwards `-end-commit` unchanged. Radius 4: `record.go`,
    `record_test.go`, `closetask.go`, `close_task_test.go`.
  - 11 (Task 11): `visualverifydispatched.go:193-210` computes, per verified HEAD, the change's own
    paths changed since it, and the `len(heads) > 0` MISSING verdict (`:269`) prints none of them;
    `verify-and-handoff.md:421-424` re-runs from the `begin` mark unscoped. Radius 4:
    `visualverifydispatched.go`, `check_visual_verify_dispatched_test.go`,
    `scripts/check-visual-verify-dispatched.sh`, `verify-and-handoff.md`.
- Spiked at plan time, every edit reverted: the six skills diffs the tasks carry, applied together,
  pass `check-references.sh`, `check-markdown-integrity.py`, `check-installed-citations.sh`,
  `check-ask-silence.sh`, `check-normative-inventory.sh`, `check-hand-notes-in-step.sh` and
  `check-dispatch-paragraphs.sh`, and `check-verbatim-moves.sh` exits 0 with this change's
  `verbatim-moves.txt`, which lists exactly the twelve lines it printed as FAIL before.
- `.flow/project.md`'s protected-services table forbids stopping flowd "to make a later step
  succeed". Task 1's restart runs after the landing push, from the code that landed, and is the one
  sanctioned exception the task records there; the operator restarts flowd by hand the same way
  today.

## Decisions

### A landing that moved stats/ restarts flowd

**ID:** flowd-restart-on-landing
**Status:** active
**Chosen:** one helper, `restartFlowdIfStatsMoved(env Env, main, old, new string, stdout, stderr io.Writer) int`
in `pipelinefix.go`, called by `pfLand` after its `pull --ff-only` succeeds and by
`refreshMainCheckout` after `REFRESH-DONE:` — `git diff --name-only <old> <new>` in the main
checkout; any `stats/` path, with `<main>/stats/Makefile` present, runs `make -C <main>/stats
restart` (which rebuilds and reinstalls the CLI too) and prints `FLOWD-RESTARTED: <main>/stats`,
or `FLOWD-RESTART-FAILED: make -C <main>/stats restart exited <n> — run it by hand` with its output,
returning 1 then.
**Considered:** a `## after landing` section in `.flow/project.md` — configuration for one
repository's one command; restarting in `pfLand` alone — run 2 and `/flow-fast` land agents
changes through `refresh-main-checkout.sh` and would stay stale; printing a reminder instead —
the reminder is what was missing in kan-940 and is the swallowed-warning shape finding 2 names;
restarting before the push — the daemon would serve unlanded code.

### close-task exits 3 on an unrecorded end

**ID:** close-task-unrecorded-exit-3
**Status:** active
**Chosen:** a failed dispatch end keeps its first warning and adds the last stderr line
`close-task: UNRECORDED — the dispatch end for <key> did not reach the store; run: flow record
dispatch end -change <name> -C <canonical> -key <key> -session-token <token> [-commit <sha>]
-outcome completed`, printed on every exit path once the end failed; the exit is 3 only where it
would otherwise be 0 — ticks and pushes still happen, so "a record never blocks" holds.
**Considered:** exit 1 or 2 — both already mean "nothing ticked or pushed", which would be false;
exit 0 with the last line alone — a filter on exit status still drops it; retrying the end inside
close-task — a store that refused once refuses the retry the same way.

### Inline verify's pair is spelled in full

**ID:** verify-recording-full-pair
**Status:** active
**Chosen:** a fenced `begin`/`end` pair after the **Recording** paragraph, the paragraph unchanged.
**Considered:** rewording the paragraph to name `-change` and `-session-token` — rewords a
normative sentence, where a block only adds.

### Flags may follow the change name

**ID:** flags-after-positionals
**Status:** active
**Chosen:** move `parseStateFlagsN`'s re-parse loop into `parseInterspersed(fset *flag.FlagSet, args []string) ([]string, error)`
in `state.go`, returning the positionals in order; `parseStateFlagsN`, the three stage callers of
`finishStageIdentityFlags` (which then takes those positionals) and `runTasksCount`/`runTasksTick`
use it. `flow stage begin … <name> -C <wt>` then succeeds.
**Considered:** a clearer error naming the misplaced flag (the pass's fix) — the command still
fails, where `flow state` already accepts the same shape; fixing `stage.go` alone — `flow tasks
count <name> -C <wt>` fails the same way.

### -effort is required on begin

**ID:** effort-required-on-begin
**Status:** active
**Chosen:** `-effort` defaults to `""` and `requireRecordFlags` names it; `default` stays an
accepted explicit value; every test call site lacking it gains `"-effort", "default"`.
**Considered:** keeping the default and warning — a warning is the swallowed shape again;
requiring it on `backfill` too — backfill already lists `effort` as a begin-only field it takes,
and no finding names a backfill row with the wrong effort.

### -location names a line

**ID:** location-needs-line
**Status:** active
**Chosen:** a non-empty `-location` must match `:[0-9]+` (`path:12`, `path:12-20` pass); empty stays
allowed for a finding that names no file.
**Considered:** requiring `-location` — a whole-branch finding has no line; a full
`^[^:\s]+:[0-9]+` match — a path with spaces is legitimate in another project.

### The bundle filters text, not git pathspecs

**ID:** bundle-text-filters
**Status:** active
**Chosen:** `elideMatrixRows(md string) string` drops every `|` line between a `####` heading and
the next heading, leaving one `_(<n> matrix rows elided)_` line per run; `dropPNGStatLines(s
string) string` drops every `git log --stat` line whose path ends `.png`. Both apply to the
rendered text.
**Considered:** a `-- . ':(exclude)*.png'` pathspec — `git log -1 <sha> -- <paths>` drops a
PNG-only commit and shows an older one instead; dropping `visual-verification.md` whole — the
flow-eval angle's visual candidates read its departures and frame headings.

### -tokens is refused under Claude Code

**ID:** tokens-refused-with-transcript
**Status:** active
**Chosen:** `checkDispatchEndFields` refuses a non-empty `-tokens` when `CLAUDE_CODE_SESSION_ID` is
set, exit 2, naming the harvester as the source; existing `-tokens` tests set it empty.
**Considered:** refusing on `MUSE_SESSION_ID` too — no evidence Muse writes a transcript the
harvester reads; letting the report add to harvested cost — the double count the finding names.

### The close command carries -ended-at

**ID:** close-command-ended-at
**Status:** active
**Chosen:** each printed command appends `-ended-at <when its work ended>` after `-C <anchor>`, and
` -commit <its task commit>` for a row whose role is `implementer`; the skills sentence says what
those are.
**Considered:** computing the instant in the guard — it holds neither the report path nor the
task's commit; appending inside the command before `-C` — breaks the pinned substring of the
existing case.

### -commit is a full sha

**ID:** commit-full-sha
**Status:** active
**Chosen:** `dispatch end` and `backfill` refuse a non-empty `-commit` not matching
`^[0-9a-f]{40}$`, exit 2; `close-task` expands `-end-commit` with `git -C <canonical> rev-parse
--verify --quiet <sha>^{commit}` first, passing it unchanged when that fails.
**Considered:** the CLI expanding it — the CLI has no repository to resolve in when `-C` names
another; accepting 7+ hex — the store then joins rows to commits by prefix.

### MISSING names the paths since the verified HEAD

**ID:** missing-names-changed-paths
**Status:** active
**Chosen:** the guard keeps the newest resolvable verified HEAD and its changed own paths; the
`len(heads) > 0` verdict appends ` — paths changed since verified HEAD <sha>: <p1>, <p2>` before the
` — HEAD <sha>` suffix, which stays last; the skills sentence scopes the re-run to that HEAD as a
round base.
**Considered:** naming only UI paths — `ask` answers for a set, and per-path calls multiply the
trigger runs; a new verdict prefix — integrate and `integrate-change.sh` key on
`VISUAL-VERIFY-MISSING:`.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — eleven behaviour changes to CLI and guard contracts other runs call,
two of them exit codes.
**Considered:** flow-low — fits applied diffs, not eleven tasks whose Go the fixer writes.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** three fixer dispatches in series — `fix-g1`: Tasks 4, 5, 6, 8, 10 (`stats/cmd/flow`);
`fix-g2`: Tasks 1, 2 (close-task and landing, after Task 10's `closetask.go` edit); `fix-g3`:
Tasks 3, 7, 9, 11 (guards and `verify-and-handoff.md`).
**Considered:** one fixer — eleven tasks over three packages in one context; one per task — eleven
dispatches, five of them editing `record.go` in sequence.

## Open questions

## Live check

After the loop lands: the landing's own output carries `FLOWD-RESTARTED:` (the range touches
`stats/`), and `lsof -ti tcp:4173 -sTCP:LISTEN` names a pid other than the one before the landing.
Against that live daemon, `flow record dispatch begin -change x -role verifier -model opus -key k
-session-token mf-x` exits 2 naming `-effort`, and `flow stage keys` still answers. Failure looks
like: no `FLOWD-RESTARTED:` line, the same pid, or the begin recorded with `effort: default`.
