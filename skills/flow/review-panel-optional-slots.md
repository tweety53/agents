# Review panel — optional slots

Loaded by `skills/flow/review-panel.md` only for a round whose roster carries `mutation`. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## The mutation slot's prompt

The mutating
role (`mutation`) is always the last pass of a bundle and still works in its
throwaway copy (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); the reading passes before them read the shared
`<worktree>`.

**The mutation slot's dispatch prompt also carries the MUTATION ENTRY CONTEXT paragraph**, every
test the plan's `**Tests:**` fields name and every test file in the touched-files list inlined
directly beneath it:

> **MUTATION ENTRY CONTEXT:** `final-review.diff`, at the merge-base sha each `# worktree:`
> header carries, is your diff base; the tests beneath this paragraph are the tests your
> mutations target. Begin the brief's search for a behaviour's covering test from those two —
> never a whole-tree sweep. Read a test outside the list only when the behaviour you are mutating
> cannot otherwise be judged caught or survived, and name that read in your report.

## The throwaway worktree

**Independent multi-slot detection is the
panel's core signal**: Mutation's
dispatch, pass 1 and every fix-round re-run, carrying the mutation-testing brief, therefore runs
against its own throwaway worktree, never the
shared `<worktree>` the other slots read, so every slot's findings are raised against the same
pristine snapshot and no slot's view ever contains another slot's work:

**The parent creates and removes every throwaway copy itself, in its own Bash calls — never a
subagent.** Run the sequence below once per worktree in the resolved set per slot, producing one
`<worktree>-<slot>-<round>` per repository per slot — `<slot>` is the id (`mutation`).

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
dies with the copy.

Findings and reproducers are unaffected: a finding's `file:line` is repo-relative, and every
reproducer still runs against the real `<worktree>` at verification time, never against any slot's
throwaway copy.
