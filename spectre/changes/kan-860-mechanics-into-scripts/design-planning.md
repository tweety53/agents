# Design — planning group (KAN-860) — WIP, nothing implemented yet

Stopped on the coordinator's instruction (usage limits) after the investigation, before any code or
prompt edit. **No row landed.** `brainstorm.md`, `brainstorm-planner.md` and `implement.md` are unchanged
(8,986 / 41,783 / 62,337 bytes). No `verbatim-moves.txt` lines. What follows is the design reached so
far, so the next session can start from it.

## Rows and the design for each

### BP-21 + BP-22 — extend `plan-class` (`stats/internal/guard/planclass.go`, `scripts/plan-class.sh`)

- Add an optional class-override flag: `plan-class.sh [-class <class>] <tasks.md> <repos> [<worktree> <merge-base>]`.
  `-class` must equal `class_mechanical` or sit one step above it (micro→small→regular→big). Any other
  value is exit 2, so the prose's "raise one step, never lower" rule is enforced.
- Keep the three existing lines byte for byte (`inputs:`, `class:` = mechanical, `rolls:`) and append
  tree lines computed for the effective class (the override, or the mechanical class when there is none):
  - `tree: class <c> · execution <inline|sdd> · implementer <skipped — inline|chosen>`
  - `panel: <default|compact|full> · roster <slot; …> · rerun delta` (compact when `compact_roll < 90`)
  - `grouping: <static|free> · dispatches <primary+principles[ · failure-modes+mutation[+exp-<name>]]>`
    (static when `bundle_roll < 30`)
  - `experimental: <no slot|none available|exp-<name> · skills/flow/experimental/<name>.md · <line-1 description>>[ · skipped — bundle cap]`
    — `sorted(*.md)[experimental_roll mod count]`, rolled when `experimental_roll < 30`. `count = 0`
    or an absent directory gives `none available`, never a division by zero. The slot joins the
    second dispatch only when one exists and has room; otherwise it is `skipped — bundle cap`.
  - micro: `panel: default`, and `grouping:`/`experimental:` read `not consulted — micro`.
- The experimental directory is `$FLOW_GUARD_REPO_ROOT/skills/flow/experimental`. The shim exports
  `FLOW_GUARD_REPO_ROOT` (precedent: `scripts/check-installed-citations.sh`). Order is byte order
  (C-locale `ls`). The directory does not exist today, so every true roll records `none available`.
- Tests: factor a pure `pcTree(class, compact, exp, bundle, candidates)` and table-test it across every
  class × roll side × candidate state. Keep the existing byte pins on the first three lines (compare a
  prefix) and add pins for the new lines. Keep the "override never appears" test.
- Consumers that say "three lines" need rewording: `brainstorm-planner.md` Decide (owned).
  `flow-fast/SKILL.md:82` and `flow-plan/SKILL.md:183` only call the script, so they still read.
- Keep the prose that the micro row never skips the decision record, the verify stage or the self-review.

### BP-23 — `flow decision render -file <decision.json> -session-model <m> -reviewers <REVIEWERS>`

- Chosen over a print from `flow record decision`: rendering is a pure function with no store I/O and
  can be tested without fakes. It is a new `decision` command in `stats/cmd/flow/main.go` with its
  own `decision.go` and `decision_test.go`.
- The inputs are what the current text names. KAN-854 dropped DEFAULT_MODEL/MODEL_SOURCE. The
  preamble is `planning:  inline, this session (<session model>)` and `reviewers: <REVIEWERS>`, then
  the two tables exactly as the worked format in brainstorm-planner.md's Decide prescribes (one fact per
  row, `↳` sub-row rules, micro `not consulted — micro`, `default` panel, string implementer, free
  grouping row, split-only groups suffix, bundle-cap suffix). The worked format moves into
  `decision_test.go` as golden cases, and the prompt keeps one call line plus "print its output verbatim".
- Exit codes: 0 printed, 2 usage error, unreadable file or a decision.json missing a required field.

### BR-M1 — `kickoff-worktree.sh <project> <name>` (Go guard `kickoff-worktree`)

- One Go guard runs steps 1–5 in order:
  1. exec `check-worktree-location.sh`, exported by the shim as `$SCRIPT_DIR/…` so check-guard-symlinks
     rule 2 sees the sibling. Precedent: `check-finish-preflight.sh`'s `FLOW_GUARD_WORKTREE_LOCATION`.
  2. `git check-ignore -q .worktrees`, appending `<project>/.worktrees/` to `info/exclude` when it exits
     non-zero.
  3. `git fetch origin`, then `rev-parse -q --verify origin/spectre/<name>`, then `worktree add` in one
     of its two forms. The default branch comes from `refs/remotes/origin/HEAD`. After the add it runs
     `flow state add-worktree` (MX4) with `rev-parse HEAD`, before anything else runs.
  4. exec `project-get.sh <worktree> "worktree setup"` and run the fenced lines in order in the
     foreground. On a non-zero command it stops, naming the command and its output.
  5. `git push -u origin spectre/<name>`.
- Before step 1 it checks that `flow` is on PATH, so it never creates a worktree it cannot persist.
- Exit contract: 0 all five steps done (prints `worktree:` and `merge-base:` lines), 1 stop (a step
  failed, with its lines relayed; the worktree is already persisted if step 3 got that far), 2 usage or
  cannot answer. The prompt keeps the stop rule and the persist-before-return sentence.
- **Cross-file impact for the integrator:** `/flow-plan` runs "the kickoff steps 1–5" of brainstorm.md
  against its own scripts dir (**Guard resolution**). It needs a `skills/flow-plan/scripts/kickoff-worktree.sh`
  symlink and a one-line change to its guard list (`flow-plan/SKILL.md` ~line 202), or it breaks.
  brainstorm.md should keep the phrase "kickoff steps 1–5" so flow-plan's citation still resolves. The
  additional-worktree recipe now lives in `cross-repo-worktrees.md` (implement.md §2 only points
  there). That file is not owned by this group, so it is noted and left.

### MX4 — `flow state add-worktree [-addr] [-timeout] [-C dir] <name> <abs-path> <merge-base>`

- Read-merge-write in `stats/cmd/flow/state.go`: read with `getChange` (exact name, falling back to
  the on-disk file when the store is unreachable), then merge `worktrees[abs]=sha` into a
  `map[string]json.RawMessage`. Every other field and every peer entry is kept, and `state` is never
  touched. Then write through the same stamp → validate → put → journal-fallback path as `state set`
  (factor its tail into a helper).
- Refusals, which write nothing:
  - exit 2: a relative path, or a merge base that is not a 40-hex sha.
  - exit 1: no record, or a synthetic-only record (the STARTED write has not happened).
- Tests reuse `state_test.go`'s fakes (`genuineDaemon`, `deadPortAddr`, `isolatedStateRoot`). They
  cover peer preservation, state preservation, the fallback read and write, and the refusals.
- The prompt passages that change: brainstorm.md step 3 (within the guard) and implement.md's
  "Read the current record with flow state get, merge in this…" sentence, which becomes one call line.
  The timing sentence stays.

## Rows left

All four rows are left in this pass. Nothing blocks them technically: the work stopped on the
coordinator's instruction before implementation.
