# Review panel — optional slots

Loaded by `skills/flow/review-panel.md` only for a round that dispatches `mutation`, always a bundle alone. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## The mutation slot's prompt

The mutating
role (`mutation`) works in its
throwaway copy (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); every other bundle reads the shared
`<worktree>`.

**The mutation slot's brief also carries the MUTATION ENTRY CONTEXT paragraph**, the tests
`build-slot-prompt`'s `-tests` names inlined directly beneath it:

> **MUTATION ENTRY CONTEXT:** `final-review.diff`, at the merge-base sha each `# worktree:`
> header carries, is your diff base; the tests beneath this paragraph are the tests your
> mutations target. Begin the brief's search for a behaviour's covering test from those two —
> never a whole-tree sweep. Read a test outside the list only when the behaviour you are mutating
> cannot otherwise be judged caught or survived, and name that read in your report.

## The throwaway worktree

**Independent multi-slot detection is the
panel's core signal**: Mutation's
dispatch, carrying the mutation-testing brief, therefore runs
against its own throwaway worktree, never the
shared `<worktree>` the other slots read, so every slot's findings are raised against the same
pristine snapshot and no slot's view ever contains another slot's work:

**The parent creates and removes every throwaway copy itself, in its own Bash calls — never a
subagent.** Run `flow-guard throwaway-worktree create <worktree> <worktree>-<slot>-<round> --sdd` once per
worktree in the resolved set per slot, producing one `<worktree>-<slot>-<round>` per repository per
slot — `<slot>` is the id (`mutation`). Exit 2 means a step failed, named on stderr: that slot is
not dispatched against that copy.

Never symlink a shared `node_modules` into a copy: the copy's own `npm ci` installs one, and over a symlink it empties the shared directory — `<agents repo>/stats/Makefile` refuses to run it there.

Dispatch each slot present in this round's panel **once**, its brief built by `flow-guard
build-slot-prompt` (**Bundled dispatch**, `skills/flow/review-panel.md`) from the shared resolved set, never from a copy: its
`-copies` names each copy, `<worktree>-<slot>-<round>`, as the repository path to mutate and test
in, and its `-worktrees` never lists a worktree whose name already ends `-mutation-<round>`.
Remove every copy unconditionally once that slot's dispatch closes — completed, aborted or the run stopped — with
`flow-guard throwaway-worktree remove <worktree> <worktree>-<slot>-<round> --fold-back`, run as its own Bash call with literal absolute paths, never chained into a compound command, which first copies
each non-empty `<worktree>-<slot>-<round>/.superpowers/sdd/panel-report-*.md`,
`<worktree>-<slot>-<round>/.superpowers/sdd/reproducers/*.sh`,
`<worktree>-<slot>-<round>/.superpowers/sdd/dispatch-context-group-*.md` and
`<worktree>-<slot>-<round>/.superpowers/sdd/implementer-report-*.md` back to the same path under `<worktree>`, unless `<worktree>` already holds a non-empty file the copy's is not newer than.
It also folds a non-empty `panel-report-*.md` found at the root of the copy or of `<worktree>` — a report a slot wrote by a relative path after a `cd` — into the directory the copy's own reports fold back to, by the same rule, moving the one at `<worktree>`'s root, so a misplaced report is never moved by hand.

The fold-back runs inside the removal step itself — on every path that removes a copy,
completed, timed out or run stopped — because a report the slot resolved onto its dispatched root
dies with the copy.

Findings are unaffected: a finding's `file:line` is repo-relative.

## The mutation-testing brief

Wherever the panel dispatches Mutation, the dispatch prompt carries a mutation-testing
brief: for each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a
boundary, remove a branch, move an interaction off its target, overlay an earlier commit's tree and
run the tests — and establish whether an existing test fails. A mutation only counts once its edit
landed: every mutation must confirm its edit landed before the tests run — the target changed
where the slot intended, not nowhere and not somewhere else. An edit that never applied must be
redone with a working mechanism: it is a refusal, never a **surviving mutant**, and it never buys
a test. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless it is
withdrawn with a reason under the fix round's handback. A reproducer never `cd`s — it runs each tool against its directory with `go -C`, `make -C` or `npm --prefix` and names every file as `$PWD/<path>` — and makes its edit through a patch file or a fixed-string replacement, never through shell quoting nested inside another quoted string. A reproducer adds no file under `$PWD`: a Go test it adds is written under `mktemp -d` and passed to `go test -overlay <json>`, whose `Replace` map names the test's path in the package, a Go mutation goes through the same overlay, its `Replace` map naming the mutated source's path and its mutated copy under `mktemp -d`, and every other file it writes lives in its own `mktemp -d` directory. Its test command filters by the name prefix the mutated code's test family
shares (`-run '^TestCheckPanel'`), never by a full test name, so a test the fix adds under that
prefix still kills the mutant and the run's cost does not grow with the package; only when the
family shares no prefix does it run every test of the package, file or module the mutated code
belongs to.
Every test run of a mutant passes the test runner's no-cache flag — `-count=1` for `go test` — because a cached result never sees a mutated file a test reads through a subprocess.
The slot's own mutant runs are scoped before the first mutation too: it measures one run of the mutated code's whole package, and when that run's seconds times the number of mutants it plans passes 15 minutes, every mutant run filters by the name prefix the mutated code's test family shares (`-run '^TestCheckPanel'`), never a full test name (group-allowed-pairs-roles self-review: the slot ran the whole guard package per mutant and timed out at 15 minutes; the prefix-filtered retry finished in 88 s).
**A class that survived twice is not mutated again.** A mutant's class is its mutation operator together with the code shape it hits — a mapping branch (a `case`, map entry or `if` arm that translates one value into another) with its mutation left unasserted, a dropped guard of one kind, a boundary of one comparison form. The first surviving mutant of a class does not stop it; once two mutants of the same class have survived, the slot runs no further mutant of that class: it finds the remaining sites of the same shape in the diff by reading (`grep` or the source), and reports them in one finding that names both surviving mutants' reproducers and every other site as `path:line`, each marked `by reading`. Those sites are unproven by a run, so that finding asserts only that the shape recurs and that the two shown sites' test gaps are shown. The mutant runs go to classes not yet seen. A class whose mutant was killed is not closed: its other sites are still mutated as the brief requires. Every distinct class the diff contains is still mutated at least once.

The slot never forces a full rebuild of its copy — no `--rerun-tasks`, `clean`, cache-busting flag or deleting build output — whatever the build tool; it measures one scoped run before its first mutation; and it runs one build at a time in its copy, never starting another while an earlier one, a backgrounded one included, still runs.
A Gradle run that fails with `Timeout waiting to lock journal` is waited out and retried with backoff, a few minutes in total — the lock sits in the shared Gradle home, held by another worktree's daemon; never run `gradle --stop` or kill a daemon another worktree owns, and if the lock is still held the mutant is reported `unproven (journal lock held by <pid from the message>)`, never skipped silently.
Never symlink a shared `node_modules` into your copy: run `npm ci` inside the copy, which installs its own — over a symlink it empties the shared directory.
Never run `git worktree add`: you work only in the copies your prompt lists, which the parent creates and removes (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); a comparison against the base reads `git show <base>:<path>` or `git diff <base> -- <path>`, never a checkout of your own.
