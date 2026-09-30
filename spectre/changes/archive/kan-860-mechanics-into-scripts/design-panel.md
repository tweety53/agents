# KAN-860 — panel group (WIP: design only, no code landed)

I stopped at the coordinator's request before making any code or prompt edit. This group changes
nothing in `skills/`, `scripts/` or `stats/`. This file records what the investigation found and
the design the next session should implement.

## Findings from the investigation

- `bugbot-reviewer-prompt.md` and `security-reviewer-prompt.md` no longer exist (retired in
  ebdfdde1), and `skills/flow/experimental/` does not exist. The slot prompts still in scope are
  `primary-`, `principles-` and `failure-modes-reviewer-prompt.md`.
- D2 (the fill protocol) is already half-resolved. `review-panel.md` **The roster** carries a
  placeholder table ("Fill every template placeholder from this table — never by reading the
  template").
- D4 persists in the templates' own **Placeholders:** lists. Principles names `fix-round-N.diff`
  first, while primary and failure-modes name `slot-delta` first. The canonical definition is the
  review-panel.md table row for `[DIFF_PATH]`.
- `fix-round-N.diff` is unsectioned today: `git diff "$FIX_BASE"..HEAD`, with one FIX_BASE.
  Sectioning it needs a FIX_BASE per worktree.

## RP35 — `write-panel-diff.sh` (Go guard `write-panel-diff`, reuses paneltouchedpaths.go)

    write-panel-diff.sh final        <canonical-wt> <wt> <merge-base> [<wt> <merge-base>…]
    write-panel-diff.sh late-fix     <canonical-wt> <wt> <since-close-sha> […]
    write-panel-diff.sh fix-round <N> <canonical-wt> <wt> <fix-base> […]
    write-panel-diff.sh slot-delta <round> <slot> <canonical-wt> <wt> <merge-base> <held-sha|-> […]

**What it writes.** Two files, both written atomically, and it prints both paths:

- `<canonical-wt>/.superpowers/sdd/<file>`;
- `<file>.touched`, the `[TOUCHED_FILES]` list: the same sections, built with
  `git diff --name-status` over the same range.

**Exit codes.** 0 when written. 2 when it cannot answer: a usage error, a pair that fails
panelValidateWorktree, or a git failure.

**Ranges**, in parity with the prose recipes:

- **final:** `git diff <mb>`. The file opens with the line
  `# final-review.diff — working tree vs merge-base, unstaged changes included`, and each
  section opens with `# worktree: %s — merge base %s`.
- **slot-delta:** `git diff <held> HEAD`, with no semantics line. When the held sha is `-`, it
  falls back to `git diff <mb>` under a merge-base header.
- **late-fix:** `git diff <since-close-sha>`.
- **fix-round:** `git diff <fix-base>..HEAD`.

**Tests.** Each case builds a fixture repo, runs the hand recipe, and asserts the script's output
is byte-for-byte the same. Cases: one worktree and two, held and unheld shas, and every refusal.

## Late-fix trigger — `check-late-fix-trigger.sh`

    check-late-fix-trigger.sh <change> <tasks.md> <wt> <since-close-sha|-> <base-verdict> […]

The first triple is the canonical worktree.

**Exit codes:**

- **0, reduce:** prints `late-fix reduction: <n> changed lines since <sha>`.
- **1, full path:** prints one line per failed condition.
- **2, cannot answer:** read as full path.

**Conditions**, all of which must hold to reduce:

1. checkPanelFindingsClosed exits 0, and a since-close sha is given. A `-` means there was no
   earlier clean close.
2. Every base verdict is `CLEAR`.
3. `git diff --numstat <sha>`, summed, is at most 40 lines. Any binary entry fails this.
4. `git diff <sha> -- tasks.md` adds no line matching `^- \[[ xX]\] [0-9]+\.`. An untracked
   tasks.md fails this.
5. No path in `git diff --name-only <sha>` matches any of:
   - `skills/flow/review-panel*.md`
   - `scripts/check-panel-*.sh`
   - `skills/flow/*-reviewer-prompt.md`
   - `skills/flow/engineering-principles.md`

## Renderer — `render-slot-prompt.sh`

    render-slot-prompt.sh <skill-dir> <round> <canonical-wt> <plan-dir> <slot>[+<slot>…] \
      -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>…] [-fix-report <path>…] \
      -- <wt> …

**What it writes:** `<canonical-wt>/.superpowers/sdd/slot-prompt-<round>-<slots>.md`, in this
order.

1. The shared blockquotes, extracted by label from `<skill-dir>/review-panel.md`:
   - TOOLS, FOREGROUND BUILDS, NO DELEGATION, REPRODUCE DON'T READ and WORKTREES;
   - CITATION CHECK, once for each file that exists;
   - ENTRY CONTEXT, followed by the diff's `.touched` list;
   - INDEPENDENT PASSES, only when there is more than one pass;
   - FIX-ROUND SCOPE, from review-panel-fix-round.md, only when `-fix-report` is given.
2. One **PASS `<id>`** section per slot: the template's fenced body, dedented, with its
   placeholders filled per the review-panel.md table, followed by that slot's REPORT FILE line.

**Exit codes.** 0 when written, and it prints the path. 1 when a placeholder stays unresolved,
and it names that placeholder. 2 when it cannot answer.

**Not rendered:** `mutation`. Its brief and its throwaway copy stay typed by the parent.

**Still typed in the Agent call:**

- the baseline pointer;
- MODEL HANDSHAKE;
- CONTEXT BUNDLE;
- the relocation-comparison pointer;
- the reproducer rule;
- "read <rendered path> in full first — it is your brief".

**Guard pins.** The check-dispatch-paragraphs pins stay satisfied because the blocks remain in
review-panel.md.

**Follow-up cut.** Once the renderer lands, cut the three **Placeholders:** lists. Keep
check-installed-citations coverage non-zero: primary keeps 3 citations in its body.

## Rows

- **Done:** none.
- **Left:** RP35, the late-fix guard and the renderer, all designed above.
- **Untouched:** RP36 and RP37.
