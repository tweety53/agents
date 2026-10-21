## Context

The seven findings of `proposal.md` are planned as one self-review change on branch
`self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without`, based at origin/main
`75261910`. Every non-trivial fix was drafted and checked in a scratch clone of this worktree at
`75261910` before planning; `tasks.md` carries the measured results.

**A's measurement.** A `--single-branch` clone of gymie `develop` (no feature refs) exits 1 with 8
`DONE-WHEN-PATH` lines. A throwaway scan of every `## Done when` token in gymie `develop`, in this
repository and in gymie-frontend found 8 dotless tokens whose first segment is no tracked
directory — exactly those 8, all branch names — and no other dotless token at all. With
**done-when-branch-shape** applied the same clone exits 0, this repository's own tree stays
`DONE-WHEN-OK`, and the new test fails on the unpatched guard.

**C's trace.** The planning session's transcript
(`~/.claude/projects/-Users-tweety53-Projects-gymie/5ed2d808-0c8b-4133-b196-5df6f3654f8d.jsonl`,
10:52:12Z) ran `… git add docs/design/one-block && git commit -qm "docs(design): add the one-block
handoff and its OB mockup frames" && git push -q` as one Bash call. The hook's push check
(`untasked_push`) runs before the command, while HEAD still lacked the commit, so it passed; the
next commit (0fa50521, 10:57) carried `Flow-Key: flow.brainstorm`. The finding also names two
wrong `false-positive` verdict rows. The store holds one: verdict 646 (kan-1003, a `CLEAR` row of
gymie-playwright's worktree, flagged 2026-10-22 00:03 with "The untasked commit 0497d943 is the
operator's design handoff…"). **No kan-994 row is flagged**: `flow record verdicts
-false-positive` over gymie, gymie-frontend, gymie-playwright and this repository lists only 646 and
kan-941's 436, and no kan-994 record write waits in gymie's journal — that half of the repair is
disproved and planned no work.

**D's measurement.** shlex with `punctuation_chars` yields every unquoted redirect operator as a
word of its own (`>`, `>>`, `>|`, `&>`, `&>>`, `>&`), so a word that merely starts with `>` came
out of quotes. The hook's second redirect branch (`tok.startswith(">")`) is what denies `'>Save<'`;
it also misses `>|` and `&>`, which the scratch harness proved (cases 140, 141 fail on the
unpatched hook).

**H's measurement.** The kan-1003 inline rows (dev store, session token `mf-k1003b`) include 17
whose instants are not the work's: tasks 1 and 2 (exactly 2700 s and 1800 s — back-filled guesses),
full-suite-fix-1 and -2 (exactly 180 s and 600 s), and the 13 `task-<10…22>-implementer-fix-<k>`
rows opened after their commits (3–53 s, 12 of them with no commit). The table in Task 8 was
derived by **inline-backfill-derives**'s rule over the four kan-1003 worktrees' HEAD and their
`refs/flow/pre-reshape/<change>` refs, which still hold the task commits the reshape collapsed.

## Live check

Against the dev store (`flowd` on 127.0.0.1:4173, database `flow`), after Tasks 4 and 8 land their
repairs:

- `flow record verdicts -C /Users/tweety53/Projects/gymie -guard check-unfinished-work -false-positive`
  — before: two rows, 646 (kan-1003) and 436 (kan-941); after: 436 alone. Still listing 646 means
  the unflag never reached the store.
- `docker exec -i flow-postgres psql -U flow -d flow -At -c "select id, started_at, ended_at, left(commit_sha,8) from dispatches where id in (6799,6804,6817,6819,6836,6842,6847,6859,6875,6885,6886,6887,6888,6889,6890,6902,6903) order by id"`
  — before: 6799 spans 2700 s and 12 rows carry no commit; after: every row's instants and commit
  equal Task 8's table. Any row still 3–53 s wide, or with an empty commit, means the retime missed it.
- A `flow record dispatch end` of a fresh inline task row with no `-commit` on a throwaway change
  is refused; the same row ended with `-commit` of a commit made after its begin is accepted. An
  accepted no-commit end means the refusal never deployed — the live flowd still runs the old
  build until `make -C <main>/stats restart` after the landing.

## Decisions

### A dotless name under no tracked directory is a branch

**ID:** done-when-branch-shape
**Status:** active
**Chosen:** `check-done-when-paths` skips a token whose last segment carries no dot, which neither
starts with `/` nor climbs with `../`, and whose first segment — read relative to the naming file's
directory or any ancestor up to the root — is no tracked directory; the existing ref exemption
stays. Measured to exempt exactly the 8 failing tokens (**Context**).
**Considered:** reading merged branch names from merge-commit subjects — a squash merge or a branch
never merged leaves nothing, and a clone's history decides the verdict; matching the words around
the token ("Branch", "Merged into") — tickets phrase it a dozen ways; exempting every dotless token
— `src/Makefile` under a tracked `src/` would never be judged again.

### A Done-when a decision overrules is amended in its own brainstorm commit

**ID:** done-when-amended-by-decision
**Status:** active
**Chosen:** a rule in section B of `skills/flow/brainstorm-planner.md`: when a decision departs from
a handoff ticket's `## Done when`, the planner rewrites that line to what the decision chose, citing
the decision's ID and date, in its own commit carrying `Flow-Key: flow.brainstorm`, before the
plan-gate commit.
**Considered:** riding the plan-gate planning commit — that commit is pathspec-limited to
`spectre/changes/<name>/` (**Planning commits**, `skills/flow-contracts/git-boundaries.md`) and the
ticket lives under `docs/`; teaching integrate to read decisions against Done-when lines — a second
reader of the same contradiction, and the ticket would still say the wrong thing.

### The KAN-1003 ticket repair is its own task, in gymie

**ID:** ticket-repair-own-task
**Status:** active
**Chosen:** Task 3 rewrites gymie's `docs/design/one-block/tickets/KAN-1003-step-1.md` Done-when
line to cite `lands-on-develop` of 2026-10-21, committed as `docs(design): …` in
`/Users/tweety53/Projects/gymie-worktrees/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`
(branch `spectre/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`, not yet landed), never
touching its `spectre/` directory. Finding B is carried by Tasks 2 and 3, the repair split out at
the dispatch's instruction because its commit lands in another repository.
**Considered:** folding the repair into Task 2 — one commit cannot span two repositories.

### A commit and its push in one command are judged together

**ID:** commit-push-one-command
**Status:** active
**Chosen:** `hooks/protect-main-checkout.py` remembers each `git commit` a command runs and, at a
later `git push` of the same repository to a `spectre/` destination, denies the command when that
commit's message as the command states it (`-m`, `-qm`, `--message`) carries neither a `Task-Id:`
nor a `Flow-Key:` trailer, its subject is no `chore`, and its pathspec is not limited to
`<spec-root>/changes/`. A message the command does not state (`-F`, `-C`, `--amend`, an unexpanded
`$VAR`) is not judged — the existing push check's fail-open stance.
**Considered:** running `check-untasked-commits.sh` at the plan gate — it would catch only
brainstorm commits, after they were pushed, needing a reword and a force push; a git `pre-push`
hook — installed per repository, and gymie's `.githooks` is the project's own; denying every
commit-and-push command — kan-1003's own transcript makes dozens with a `Task-Id:` trailer.

### An UNTASKED listing is never flagged as a verdict false positive

**ID:** untasked-never-flagged
**Status:** active
**Chosen:** `skills/flow/unfinished-work-gate.md` states that `flow record verdict
false-positive` is never run for an `UNTASKED` listing, since `check-untasked-commits.sh` records
no verdict and the flag lands on `check-unfinished-work`'s newest row, and that the relay never
calls a listed commit the operator's own.
**Considered:** making `check-untasked-commits.sh` record verdicts — a new store write for a
listing whose remedy is a reword, not a judgment.

### flowd unflag-verdict repairs a wrongly flagged verdict

**ID:** flowd-unflag-verdict
**Status:** active
**Chosen:** `flowd unflag-verdict <verdict-id>`, beside `flowd restamp-dispatch`: a direct database
write through a new `store.Store.UnflagVerdict`, clearing `false_positive`,
`false_positive_reason` and `flagged_at`. Task 4 runs it once, on verdict 646.
**Considered:** a `flow record verdict unflag` CLI verb and API route — a store, route and client
surface for a repair an operator runs once; editing the row with `psql` — no test pins it and the
next repair repeats the hand SQL.

### The hook judges only standalone redirect operators

**ID:** redirect-operators-only
**Status:** active
**Chosen:** a redirect is a word in `REDIRECT_OPS` (`>`, `>>`, `>|`, `&>`, `&>>`) followed by its
target; the branch judging a word that merely starts with `>` is removed. The unbalanced-quote
fallback (`command.split()`) splits a redirect glued to its target off the word — `>f`, `2>>f`,
`&>f`, never `>&1` — so it still reads as shlex would have split it.
**Considered:** tracking which shlex words were quoted — shlex does not report it, and the
operator set already says it; keeping the prefix branch for non-quoted words only — the same
question with no answer from shlex.

### A departure's expected value cites the handoff source

**ID:** expected-cites-source
**Status:** active
**Chosen:** one rule in step 10 of `skills/flow/visual-verify-verifier.md`, beside "never a bare
`departure`": where the frame was drawn from HTML or CSS in the tree, the departure line names the
declaration its expected value comes from; a value no declaration states is named as read from the
pixels.
**Considered:** a guard parsing the verifier's report for declarations — a departure about a
frame with no source in the tree states none, so the guard could not tell a missing citation from
a legitimate one.

### One script rebuilds the full app suite's zip

**ID:** zip-rebuild-script
**Status:** active
**Chosen:** `scripts/rebuild-full-app-suite-zip.sh <worktree> <screenshots root>`, the step-8
inline bash moved verbatim into a script (old zip removed, PNG mtimes pinned to 1980-01-01,
`zip -X -j`), linked from `skills/flow/scripts/`, with `scripts/test-rebuild-full-app-suite-zip.sh`.
The verifier's step 8, the in-run fix loop's step 2 and the base sync's recapture step 3 call it.
**Considered:** the script resolving `<screenshots root>` itself — the resolver prints PNG paths,
not the root, and every caller already holds the root; a guard checking the zip's freshness at
handoff — it would find the staleness after the round that caused it, where the call prevents it.

### The store refuses an inline row begun after its commit

**ID:** inline-commit-checks
**Status:** active
**Chosen:** `store.EndDispatch` refuses a `completed` inline row (a) on a task key
(`task-<…>-implementer…`) whose end carries no commit — `ErrInlineCommitMissing` — and (b) whose
end carries a `commitAt` earlier than the row's `started_at` — `ErrInlineBegunAfterCommit`; both
answer 409 and are retired on replay, as `ErrInlineWindowTooShort` is. `flow record dispatch end`
sends `commitAt` — the `-commit`'s committer time, looked up in `-C`'s repository, else in each
worktree of the change's state record — and sends none when no worktree knows the sha.
**Considered:** deriving the commit on `end` from `Task-Id:` trailers — `end` cannot tell an inline
row from a dispatched one, and every existing end test on a task key would gain a lookup; refusing
on the CLI — only the store knows a row is inline.

### An inline back-fill with a commit derives its instants

**ID:** inline-backfill-derives
**Status:** active
**Chosen:** `flow record dispatch backfill -agent-id inline -commit <sha>` refuses `-started-at`
and `-ended-at` and derives them: `-ended-at` is the commit's committer time; `-started-at` is the
newest committer time strictly before it among the commits `git log -1 --before` finds from HEAD in
each of the change's worktrees (`-C`'s repository and the state record's `worktrees`) — the end of
the previous unit of work anywhere in the change. Without `-commit`, or for another agent id, both
instants stay required as today.
**Considered:** the commit's first parent as the start — measured over kan-1003, task 2's frontend
commit's parent is a `develop` commit from the night before (a 13-hour window) and task 22's parent
was made in the same second; accepting explicit instants beside a commit — the guess the finding
names stays possible.

### kan-1003's rows are retimed, then restamped

**ID:** retime-then-restamp
**Status:** active
**Chosen:** `flowd retime-dispatch <id> <started-at> <ended-at> <commit>` (a direct database write,
`store.Store.RetimeDispatch`) sets each of the 17 rows to Task 8's measured table; then the
existing `flowd restamp-dispatch <id> <parent transcript>` recomputes each inline row's token share
over its new window from session `85208c79-6c1b-4be3-9f2a-5b8150f79a68`'s transcript, the one
carrying session token `mf-k1003b`.
**Considered:** deleting the rows and back-filling them anew — row ids and seqs other records
cite would change; leaving the windows and restamping alone — the attributor reads the windows,
so the tokens would stay wrong.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change moves a store refusal on the dispatch write path, a CLI
derivation across worktrees and the hook's command parsing, each with live-data repairs.
**Considered:** flow-low — fits a prose-only change, not new refusals on a hot write path.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** three fixer dispatches in series: g1 Tasks 1, 2, 3, 6, 7 (the guard, the prose rules,
the gymie ticket and the zip script — no shared file with the other groups); g2 Tasks 4, 5 (both
edit `hooks/protect-main-checkout.py` and its harness, Task 5 after Task 4); g3 Task 8 (the
inline-row refusals and repair, after Task 4's edits to `stats/cmd/flowd/main.go` and
`stats/internal/store/records.go`).
**Considered:** one dispatch for all eight — Task 8 alone carries a store, CLI and live repair
large enough to want a fresh context; a dispatch per task — eight cold starts for five prose-sized
tasks.

## Open questions
