# Review panel — optional slots

Loaded by `skills/flow/review-panel.md` only for a round whose roster carries `bugbot`,
`mutation` or an `exp-` slot. Every section name below without a path is a section of
`skills/flow/review-panel.md`.

## Experimental slot

When the decision's `panel.roster` carries an entry whose `slot` starts `exp-` — at most one, per
design.md's **The rolls** — it is dispatched once, in pass 1 alongside the rest of the roster,
exactly like any other slot in **The roster** table (`skills/flow/review-panel.md`), carrying the same paragraphs every
slot's dispatch already carries there.

Its prompt is not `principles-reviewer-prompt.md` or any other fixed template: it is the file the
roster entry names in `prompt` — `skills/flow/experimental/<name>.md` inside the agents repo, never
a path inside the project worktree — read by **absolute** path and substituted into the dispatch
prompt the same way `[PRINCIPLES_PATH]` is resolved for Principles. Confirm the file exists before
dispatching; an absent file at dispatch time (the roster was decided against a prompt that has since
moved) is reported and this slot dropped from this run, never dispatched against nothing.

Its `-slot` on the dispatch record, and every `flow record finding -slot` this slot raises, is the
full `exp-<name>` id verbatim — never shortened to `experimental` or to `<name>` alone. Its report
file is `<abs-worktree>/.superpowers/sdd/panel-report-<round>-exp-<name>.md`, the same
`panel-report-<round>-<id>` shape every slot's REPORT FILE paragraph already names with `<id>`
substituted. The rendered panel record's Slot column therefore shows the `exp-` id unchanged, so the
prefix survives into the archive.

It is a diff-reading slot like Primary, Principles and Mutation: **Panel
re-runs** governs it unchanged. **The docs-only reduction** still narrows a
docs-only branch to `primary` alone: the experimental slot is never part of that reduced roster, and
is dispatched again only if a later round's docs-only guard reclassifies the branch off the
reduction.

It runs at most once per change, whether or not the roster is `compact` — the experimental roll and
the compact roll are independent per design.md's **The rolls** — and never at all when
`REVIEW_PANEL_TOGGLE` is `default`, or when the decision recorded `experimental: none available`.

Per **Bundled dispatch**, it joins whichever group has room, last among the reading passes;
when neither group has room for a third role it is skipped and recorded with
`flow record pass -round <round> -note 'experimental: skipped — bundle cap'` rather than displacing a persistent role.

## The throwaway worktree

Bugbot and Mutation both mutate code in place to run their brief; every other slot only reads the
diff. Dispatching a mutating slot into the same worktree
a reading slot concurrently reads is the
collision — a mutation applied for one slot's test is visible to whatever a concurrently dispatched
reading slot reads from `<worktree>` at that moment. **Independent multi-slot detection is the
panel's core signal**: each of
the two defect-hunting slots — Bugbot's and Mutation's dispatch, pass 1 and
every fix-round re-run, both carrying the mutation-testing brief (Bugbot's own copy is
`bugbot-reviewer-prompt.md`'s) — therefore runs
against its own throwaway worktree, never the
shared `<worktree>` the other slots read, so every slot's findings are raised against the same
pristine snapshot and no slot's view ever contains another slot's work:

**The parent creates and removes every throwaway copy itself, in its own Bash calls — never a
subagent.** Run the sequence below once per worktree in the resolved set per slot, producing one
`<worktree>-<slot>-<round>` per repository per slot — `<slot>` is the id (`bugbot` or `mutation`),
so a roster carrying both produces two copies per repository per round.

```bash
git -C <worktree> worktree add --detach <worktree>-<slot>-<round> HEAD
git -C <worktree> diff HEAD --binary | git -C <worktree>-<slot>-<round> apply --allow-empty
git -C <worktree> status --porcelain -z | \
  while IFS= read -r -d '' entry; do
    st="${entry:0:2}"; f="${entry:3}"
    [ "$st" = "??" ] || continue
    mkdir -p "<worktree>-<slot>-<round>/$(dirname "$f")"
    cp -a "<worktree>/$f" "<worktree>-<slot>-<round>/$f"
  done
mkdir -p "<worktree>-<slot>-<round>/.superpowers"
if [ -d "<worktree>/.superpowers/sdd" ]; then
  cp -a "<worktree>/.superpowers/sdd" "<worktree>-<slot>-<round>/.superpowers/sdd"
fi
```

`git diff HEAD --binary` — against `HEAD`, not a bare `git diff --binary` — is the same "staged and
unstaged together" semantics `final-review.diff` already uses, and covers the transplant in
one diff rather than the working-tree-only diff a bare `git diff` produces: a bare `git diff` misses
anything staged, and (independently) fails to reconstruct a rename whose move is already reflected
in the index. `--allow-empty` on the `apply` side makes the sequence a no-op, not a failure, when
there is nothing to transplant — the common case, since task and fix-round work is committed and
`worktree add --detach ... HEAD` already carries every committed change on its own. The
untracked-file loop reads `git status --porcelain -z`, NUL-delimited, into `read -r -d ''` — the
plain-text `awk` form cannot survive git's quote-escaping of a filename with a space or another
special character, and silently drops that file from the copy; the `-z`/NUL form carries the literal
byte string through untouched, regardless of what the filename contains. The scaffold lines after
the loop exist because the slot dispatched into the copy resolves the bundle paths its prompt names
— `dispatch-context.md`, and the report file it writes — against its dispatched root. The `-d` test
keeps a canonical worktree with no bundle yet a no-op rather than a failure.

Dispatch each slot present in this round's roster **once**, its prompt listing every copy made for
that slot as the repository paths to mutate and test in, in place of `<worktree>`.
Remove every copy unconditionally once that slot's dispatch closes — completed, timed out
(including after the wall-clock re-dispatch), or the run stopped:

```bash
# for each copy:
for f in <worktree>-<slot>-<round>/.superpowers/sdd/panel-report-*.md \
         <worktree>-<slot>-<round>/.superpowers/sdd/reproducers/*.sh; do
  [ -s "$f" ] || continue
  b="${f#<worktree>-<slot>-<round>/.superpowers/sdd/}"
  mkdir -p "<worktree>/.superpowers/sdd/$(dirname "$b")"
  c="<worktree>/.superpowers/sdd/$b"
  [ -s "$c" ] && [ ! "$f" -nt "$c" ] && continue
  cp -a "$f" "$c"
done
git -C <worktree> worktree remove --force <worktree>-<slot>-<round>
```

The fold-back runs inside the removal step itself — on every path that removes a copy,
completed, timed out or run stopped — because a report the slot resolved onto its dispatched root
dies with the copy. It rescues the slot's reproducers beside its reports — both are recorded as
worktree-relative paths the parent later runs, and one left only in the copy dangles. A copy-side
file newer than the canonical one replaces it — the wall-clock re-dispatch's fresh report
outranking a timed-out attempt's — and the `[ -s ]` tests keep an empty copy-side file from being
copied.

Findings and reproducers are unaffected: a finding's `file:line` is repo-relative, and every
reproducer still runs against the real `<worktree>` at verification time, never against any slot's
throwaway copy. Security is **not** isolated this way — nothing in the review panel requires it to
mutate anything, so it keeps sharing `<worktree>` with the reading slots. It too is dispatched
once, its prompt naming every worktree in the resolved set.
