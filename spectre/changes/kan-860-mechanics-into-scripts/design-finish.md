# KAN-860 — group `finish` — design notes (WIP: nothing implemented)

Status: stopped at the coordinator's request before any code or prompt edit. No script, verb, test
or prompt change landed from this group; every row below is still open. What follows is the
investigation result and the chosen shapes, so a follow-up run can implement without re-deriving.

## Findings that change the audit rows

- **MI1/MA2** — archive.md no longer carries the `## self review` enum parse (cut by KAN-855..859);
  skills/flow/SKILL.md's model block has none either. The remaining hand parses are
  `skills/flow/integrate.md` ("Run `project-get.sh <main-checkout> "default landing route"` (exit 1:
  absent), take the body's first non-blank line — trimmed, backticks removed …") and
  `skills/flow-fast/SKILL.md` §6 ("The route is `## default landing route` when the project declares
  one …", plus §4's `## handoff` read). `scripts/project-get.sh` is **bash** (sources
  scripts/lib/project-section.sh), so it must first be ported to Go byte-for-byte
  (`stats/internal/guard/projectget.go`; the parity test runs the bash via
  `git show ae805186:scripts/project-get.sh` + lib/project-section.sh + lib/strip-bom.sh, like
  `bashAtBase` in mutate_and_verify_test.go), shim via `flow_guard_exec project-get 2 …`.
  Chosen enum shape: `project-get.sh <root> <key> --enum <literal>...` — prints the matched literal,
  exit 0; exit 1 key/file absent; exit 3 declared but the head (first non-blank line, trimmed,
  surrounding backticks stripped, trimmed again) matches no literal byte-for-byte — one stderr line
  quoting the head; the caller reports it and resolves as absent; exit 2 unchanged.
  Contract text: skills/flow-contracts/project-configuration.md "A single-line-literal key's value…".
- **MA1** — A04/A05/A07 prose was already cut from archive.md by KAN-859 (A05 sits in
  skills/flow/SKILL-rationale.md:939). The block to replace is archive.md step 4's
  `bash -c 'for pair in …'` + branch assert + `git add -A` + `check-archive-scope.sh` + commit.
  Plan: `stats/internal/guard/commitarchive.go` (`commit-archive <landing> <canonical-wt> <name>`),
  calling sibling check-archive-scope.sh via guardSelfDir (the shim exports FLOW_GUARD_SELF and
  asserts `$SCRIPT_DIR/check-archive-scope.sh` exists, so check-guard-symlinks rule 2 sees the
  dependency). Verdicts: `ARCHIVE-COMMITTED: <sha>` / `ARCHIVE-NOTHING-STAGED` exit 0;
  `ARCHIVE-WRONG-BRANCH: <found>` / scope-violation lines exit 1; exit 2 cannot answer. Move the
  old recipe verbatim into finish-hand-fallbacks.md under a `commit-archive.sh — Run 2, step 4`
  section.
- **MR2** — reusable pieces: cleanupcomplete.go's porcelain loop (factor into a helper returning
  apply worktrees on `refs/heads/spectre/<name>` + detached `ccWaveGroupCopy` copies),
  `ccRunSurvivors` (bash -o pipefail, own process group, SIGTERM, grace, SIGKILL) for check 5's
  60-second bound, `resolveBaseBranch` in-process for check 3, sibling
  `check-worktree-processes.sh` for check 6. Planned interface:
  `remove-change-worktrees.sh <repo> <name> <merge-base|-> [--proceed]`. Without `--proceed`, a
  non-empty unclassified bucket or any wave-group copy stops before removal with DISCLOSE lines
  (exit 3), so the run relays, judges irreplaceable/preserved, asks the one ask, then calls again
  with `--proceed`. Check 5 reads the `## stop` body from `<repo>`. A body with no fenced command
  block counts as "declares no command" and the check is skipped. This repo's `## stop` is prose
  only (CLAUDE.md: never stop flowd, flow-postgres or the `flow` DB). The script runs only what the
  key's fence declares, never anything of its own. The remote delete tells "remote ref does not
  exist" apart by git's message, then runs `fetch --prune`. Verdict lines: REMOVED / HELD /
  REGENERATABLE / UNCLASSIFIED / REFUSED / REMOTE-*. Move the six-check recipe, the bucket list and
  the remove/remote blocks verbatim into finish-hand-fallbacks.md; run2 keeps the gates-vs-disclosure
  rules, the ask and the relay.
- **FF1** — chosen verb: `flow stage mark -command '/flow-fast' -stages flow.a,flow.b
  -harness <harness> -session-token ff-<literal-token> <name>` (begin, then end `completed`, per
  key, in order; every key validated by stages.Validate before any store call; the same journal
  fallback as begin/end). Also needed: `stats/internal/harvest/watcher.go`'s
  `stageMarkInvocationPattern` must also match `stage mark`, or session-token binding breaks where
  CLAUDE_CODE_SESSION_ID is unset. `stagemarkcalls.go` must detect `flow stage mark` lines, apply
  the begin rules (token, placeholder harness, no guess) and check every comma-separated `-stages`
  key against the served vocabulary. TestStageKeysMatchFlowFastSkillTable reads only the Stage keys
  table, which stays untouched.

## Rows left

MI1/MA2, MA1, MR2 and FF1 are all still open: none was started before the usage limit. MR1, MA3 and
the land steps were out of scope by the task.
