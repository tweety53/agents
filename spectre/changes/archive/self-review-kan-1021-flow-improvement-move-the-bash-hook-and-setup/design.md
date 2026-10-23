## Context

Self-review fix change for `kan-1021-flow-improvement-move-the-bash-hook-and-setup`, planned on that
change's own branch `spectre/kan-1021-flow-improvement-move-the-bash-hook-and-setup` in its
worktree. Three non-`big` findings, each carried by one task in `tasks.md`; all three reproduce at
`f85971db` and `git log --since=2026-10-22` over the touched paths shows no earlier fix. Every
question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). Every reworded or added run-loaded
sentence is listed in this change's `verbatim-moves.txt`, checked with `check-verbatim-moves.sh`
(exit 0) with all three tasks' diffs applied.

## Decisions

### The citation-check fence is handled in prose

**ID:** citation-fence-prose
**Status:** active
**Chosen:** one clause in `skills/flow/review-panel.md`'s "run its command" sentence — the command is the lines between the fence markers `project-get.sh` prints around it, never an `eval` of the printed body — where the run reads it.
**Considered:** a fence-strip mode in `stats/internal/guard/projectget.go` — six files for a behaviour every other fenced key (`## lint`, `## test`, `## worktree setup`) already handles by its "one command per line inside the fence" wording; editing the `## review panel citation check` row of `skills/flow-contracts/project-configuration.md` as well — a second statement of the same rule; `.flow/project.md` — its section is only the fenced block, nothing in it is wrong.

### A flow-* agent reads the sub-skill from its installed file

**ID:** subskill-file-path
**Status:** active
**Chosen:** the two REQUIRED SUB-SKILL paragraphs and the panel-fix sentence say "read and follow … from its file — on Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/<skill>/SKILL.md`", the pattern `skills/flow-self-review/SKILL.md` step 3 item 1 already uses for the planner; the inline route, where the parent applies the paragraphs itself, reads the same file.
**Considered:** granting the Skill tool in `agents/flow-*.md` — widens every flow agent's tool set and re-prices its context for two skills; dropping the sub-skill instruction — loses the TDD and debugging discipline the paragraphs exist for.

### The sub-skill paragraphs get a dispatch-paragraphs row

**ID:** subskill-guard-row
**Status:** active
**Chosen:** a `subskill` entry in `stats/internal/guard/dispatchparagraphs.go` — label `**REQUIRED SUB-SKILL:**`, shared phrases `from its file` and the plugin-cache path prefix, variants `tdd` and `debugging` naming each skill's `SKILL.md` — required twice in `skills/flow/implement.md`; with it the guard fails the current tree (6 violations) and passes the fixed one (34 sites).
**Considered:** no guard — a later prose edit could drop the path silently, the failure the guard exists for; a row for `review-panel-fix-round.md`'s sentence — it is prose, not a labelled dispatch paragraph, and the guard matches labels.

### flow-fast and the shim comment stay as they are

**ID:** flow-fast-and-shim-unchanged
**Status:** active
**Chosen:** no edit to `skills/flow-fast/SKILL.md` or `scripts/check-dispatch-paragraphs.sh`. flow-fast's `superpowers:test-driven-development` sentence is read by the session itself, which carries the Skill tool, and its `sdd` route dispatches through `skills/flow/implement.md`'s paragraphs; the shim's line-75 comment ("beside the TDD and systematic-debugging REQUIRED SUB-SKILL paragraphs") stays true, and its documented table already stops at READ-ONLY REVIEW, with the later `fixround` and `proofruns` rows living in the Go table alone.
**Considered:** editing both, as the finding's blast radius counted them — a change to text that is not wrong.

### The parent corrects an off-by-line citation in place

**ID:** parent-line-correction
**Status:** active
**Chosen:** in `skills/flow/review-panel-fix-round.md`'s audit exit-1 dispositions, an unresolving demonstrates citation whose cited path holds the claimed content verbatim on another line is the one exception, never bounced: the parent corrects the line number in the reproducer, re-runs the audit, and records old and new line in that round's `flow record pass`.
**Considered:** keeping the bounce — a slot round-trip for a one-line offset the parent can check mechanically; letting the parent correct any unresolving citation — a content mismatch is the reviewer's claim being wrong, which only the reviewer can restate.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes the panel's audit contract, and Task 2 adds a guard row every lint run evaluates plus fixture edits across seven existing subtests.
**Considered:** `flow-low` — every task is small and prose or guard-pinned, but a misread exception in Task 3 lets a wrong citation through unbounced.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1021-flow-improvement-move-the-bash-hook-and-setup-fix`, Tasks 1, 2 and 3 in order; Tasks 2 and 3 share `skills/flow/review-panel-fix-round.md`.
**Considered:** one fixer per task — three dispatches for three small diffs, two of them on the same file.

## Open questions

## Live check

None — the change edits skill prose and a lint guard's table; no running service or persistent state is touched, and `check-dispatch-paragraphs.sh` over the real tree is the guard's own check.
