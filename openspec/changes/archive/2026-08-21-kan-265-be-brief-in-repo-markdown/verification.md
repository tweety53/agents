# Verification — kan-265-be-brief-in-repo-markdown

End-to-end proof for the whole change, run against the worktree at
`openspec/kan-265-be-brief-in-repo-markdown`, HEAD `f54889f`, merge base `f09f2b1`.

Ten commits are under test:

| Commit | Subject |
|--------|---------|
| `e0d6a33` | `feat(be-brief): state brevity for repository prose and generated artifacts` |
| `764eb77` | `feat(check-normative-inventory): inventory the corpus's normative sentences` |
| `76fbb56` | `refactor(skills): cut restatement from the myflow-do and contracts text` |
| `4e971bd` | `refactor(skills): cut restatement from the remaining skill text` |
| `6186c7e` | `refactor(skills): cut restatement from the skills README` |
| `e4ead90` | `refactor(rules): cut restatement from the rule, command and project text` |
| `08faeb7` | `refactor(openspec): cut restatement from the specs` |
| `79e444f` | `docs: cut restatement from the root documentation` |
| `8c3a00d` | `refactor: cut the per-citation do-not-restate tails` |
| `f54889f` | `feat(check-contract-budget): ratchet every owned Markdown file` |

This task modified no source file. It ran the checks below and wrote this document, and nothing else.

---

## 1. The whole-change normative inventory — proof against the base branch

### Why this check exists, and what the per-task checks did not cover

Each trim task (4.1 through 7.1) compared a fresh inventory against
`normative-baseline.txt`. That baseline was captured **inside this worktree**, on a tree that already
carried two of this change's commits. So the per-task checks prove each trim did no harm relative to
the tree the baseline was taken from — they do not, on their own, prove the baseline itself matched
the base branch. If the baseline had been captured after a trim had already landed, every subsequent
per-task check would have been comparing against a goalpost that had already moved, and a sentence
lost before the capture would be invisible to all of them.

This section closes that gap by running the guard against the base branch's own tree.

### Procedure

The base branch was materialised as a detached worktree at the merge base, and **HEAD's** guard was
run over it through the guard's `CHECK_NORMATIVE_INVENTORY_ROOT` override. Running HEAD's guard
rather than the base's is deliberate and necessary: `scripts/lib/owned-corpus.sh`, through which the
guard resolves the corpus, did not exist at `f09f2b1`. The override is the mechanism the guard
provides for exactly this — pointing one implementation of the rule at another tree — so both
inventories are produced by one implementation and differ only in the files they read.

```console
$ git worktree add --detach <scratch>/kan265-base f09f2b1
Preparing worktree (detached HEAD f09f2b1)
HEAD is now at f09f2b1 Merge branch 'chore/make-restart'

$ CHECK_NORMATIVE_INVENTORY_ROOT=<scratch>/kan265-base scripts/check-normative-inventory.sh > base-inv.txt
check-normative-inventory: 1077 normative sentence(s) from 92 file(s) under <scratch>/kan265-base
$ echo "exit=$?"
exit=0
$ wc -l < base-inv.txt
1077

$ scripts/check-normative-inventory.sh > head-inv.txt
check-normative-inventory: 1078 normative sentence(s) from 92 file(s) under <worktree>
$ echo "exit=$?"
exit=0
$ wc -l < head-inv.txt
1078

$ git worktree remove --force <scratch>/kan265-base
```

**Neither run was a silent exit-2.** An exit-2 run of this guard leaves an empty stdout, and two
empty files compare equal — a false pass that looks exactly like a real one. Three independent
signals rule it out on both runs: each exited 0, each printed a non-empty line count, and each wrote
a stderr verdict naming its own root and a non-zero sentence and file count. The two verdict lines
name different roots, so the override did reach the base tree rather than silently falling back to
the worktree.

### Result

```console
$ diff base-inv.txt head-inv.txt
486a487
> It prints every sentence in this repository's owned Markdown that carries `SHALL`, `SHALL NOT`,
> `MUST` or `MUST NOT` as a whole word — one per line, whitespace-normalised, sorted — and its exit
> codes are only `0` the inventory printed and `2` it cannot answer.
```

The diff carries **one line, and it is an addition** (`486a487`, a `>` line — no `<` line anywhere in
the output). Every one of the base branch's 1077 normative sentences is present verbatim at HEAD.

**No normative sentence was dropped, altered, or reworded by this change.** That is the whole claim
this section makes, and the diff above is its proof: a reworded sentence would appear as a `<` and a
`>` pair, and a deleted one as a bare `<`. Neither shape occurs.

The single added sentence was introduced by `764eb77`, the guard commit, in the `.myflow/project.md`
paragraph documenting `check-normative-inventory.sh`'s own exit-code contract:

```console
$ git log --oneline -S 'its exit codes are only' f09f2b1..HEAD
764eb77 feat(check-normative-inventory): inventory the corpus's normative sentences
```

Both runs report **92 files**, unchanged. No file entered or left the corpus.

### `normative-baseline.txt` against the base branch

```console
$ wc -l < openspec/changes/kan-265-be-brief-in-repo-markdown/normative-baseline.txt
1078
$ diff base-inv.txt normative-baseline.txt
486a487
> It prints every sentence in this repository's owned Markdown that carries `SHALL`, `SHALL NOT`,
> `MUST` or `MUST NOT` as a whole word — …

$ diff normative-baseline.txt head-inv.txt
$ echo "exit=$?"
exit=0
```

`normative-baseline.txt` is **not** byte-identical to the base branch's inventory, and this is the
expected, benign case rather than the failure the check was looking for. The single difference is the
same addition traced above — the sentence `764eb77` added. The baseline was therefore captured after
`e0d6a33` and `764eb77` had landed and **before any trim commit**, which is exactly what task 2.2
specifies ("before any file in tasks 4.1–7.1 is edited") and the only order physically available: the
guard that produces the baseline is itself created by `764eb77`.

**The goalpost did not move for any trim task.** The failure mode this check exists to catch is a
baseline captured *after* a trim, which would hide a sentence deleted before the capture. That would
show as a `<` line in `diff base-inv.txt normative-baseline.txt` — the base branch holding a sentence
the baseline lacks. There is no such line. The two enabling commits added one sentence and removed
none, and `normative-baseline.txt` equals the HEAD inventory exactly, so the per-task checks were all
measured against a baseline that is a strict superset of the base branch's normative content.

---

## 2. Byte totals — measured, per area

The corpus is resolved by `scripts/lib/owned-corpus.sh` — the enumerated scope roots (`skills/`,
`rules/`, `openspec/specs/`, `commands/`, `commands-claude/`, `.myflow/`) plus the `.md`/`.mdc` files
sitting directly at the repository root, with the structural exclusions applied. This is **not** a
whole-tree sweep: a sweep finds 114 files where the library finds 92, the difference being
`docs/self-review/*.md` and `stats/README.md`. The library's answer governs, because it is the corpus
both guards and the normative inventory agree on.

Sizes were taken from each commit's git tree, filtered through the library's own scope-root and
exclusion rules, and cross-checked against `wc -c` over the two real working trees; the two methods
agree exactly. The worktree at HEAD is clean with respect to the corpus (`git diff HEAD` over the
scope roots is empty, and no untracked `.md` or `.mdc` sits in any of them), so the on-disk figures
are the commit's figures.

| Area | Files | `f09f2b1` bytes | `f54889f` bytes | Delta |
|------|------:|----------------:|----------------:|------:|
| `skills/` | 33 | 603,656 | 589,609 | −14,047 |
| repository root (`README.md`, `AGENTS.md`, `CLAUDE.md`) | 3 | 71,092 | 64,920 | −6,172 |
| `rules/` | 13 | 45,444 | 44,890 | −554 |
| `commands/` | 6 | 18,649 | 18,188 | −461 |
| `openspec/specs/` | 31 | 461,845 | 461,600 | −245 |
| `commands-claude/` | 5 | 9,114 | 9,114 | 0 |
| `.myflow/` | 1 | 14,532 | 15,477 | **+945** |
| **Total** | **92** | **1,224,332** | **1,203,798** | **−20,534 (1.677%)** |

<!-- measured: git ls-tree -r -l <commit>, filtered by scripts/lib/owned-corpus.sh's scope roots and
     exclusions, cross-checked against owned_corpus_files + wc -c on both trees @ task 9.1 -->

`.myflow/` grew rather than shrank, and that is intended: `764eb77` and `f54889f` between them added
the paragraphs documenting `check-normative-inventory.sh` and the widened
`check-contract-budget.sh` to `project.md`. The trim of that file was smaller than the documentation
the change owed it.

`commands-claude/` is unchanged: no file in it was touched by any commit in this change.

### Correction to the figure recorded in `f54889f`'s commit message

`f54889f`'s message states "92 files, 1,203,716 bytes … 20,616 bytes cut, 1.68%". **The file count
and the percentage are right; the two byte figures are 82 bytes stale.** 1,203,716 is the corpus size
at `8c3a00d`, the commit before it — the measurement was taken before that commit's own
`.myflow/project.md` edit (+82 bytes) was included in the tree it describes:

| Commit | Corpus bytes |
|--------|-------------:|
| `f09f2b1` (base) | 1,224,332 |
| `e0d6a33` | 1,225,603 |
| `764eb77` | 1,226,719 |
| `76fbb56` | 1,220,693 |
| `4e971bd` | 1,215,777 |
| `6186c7e` | 1,213,770 |
| `e4ead90` | 1,211,266 |
| `08faeb7` | 1,211,164 |
| `79e444f` | 1,205,632 |
| `8c3a00d` | 1,203,716 |
| `f54889f` (HEAD) | 1,203,798 |

The correct figures for the change as a whole are **1,224,332 → 1,203,798, a cut of 20,534 bytes
(1.677%)**. The discrepancy is confined to that one prose sentence in a commit message; the guard's
own `budgets()` row for `.myflow/project.md` is 19346, which is 15,477 × 1.25 truncated, so the
ratchet table was written against the file's real final size and is correct.

The corpus rises above the base for the first two commits before falling: this change adds
documentation for what it builds and only then cuts.

---

## 3. Every guard and every harness

All commands were run in the foreground from the worktree root. The test set was split across
several invocations, per the runtime note in `.myflow/project.md`.

`stats/internal/web/dist` was already present, so **`make build` in `stats/` was not needed** and was
not run. The `pattern all:dist: no matching files found` trap did not fire.

### `## lint` — every entry exit 0

| Command | Exit | Verdict |
|---------|-----:|---------|
| `scripts/check-vocabulary.sh` | 0 | scanned skills 93 · rules 13 · commands 6 · commands-claude 5 · scripts 74 · root 3 |
| `scripts/check-references.sh` | 0 | all references resolve |
| `scripts/check-plan-provenance.sh` | 0 | `check-plan-provenance: 3 file(s) scanned, all provenance stated` |
| `scripts/check-task-build-green.sh` | 0 | clean |
| `scripts/check-workspace-isolation.sh` | 0 | `ISOLATION-OK` — 3 resource rows and 3 command rows |
| `scripts/check-uitest-overrides.sh` | 0 | `UITEST-OVERRIDES-OK` |
| `scripts/check-contract-budget.sh` | 0 | `BUDGET-OK: 92 owned Markdown file(s) within budget` |
| `scripts/check-markdown-integrity.py` | 0 | clean |
| `scripts/check-stage-mark-calls.sh` | 0 | clean |
| `scripts/check-guard-symlinks.sh` | 0 | clean |
| `scripts/check-self-review-report.sh` | 0 | clean |
| `scripts/check-installed-citations.sh` | 0 | clean |
| `scripts/check-normative-inventory.sh` | 0 | `1078 normative sentence(s) from 92 file(s)` |
| `cd stats && gofmt -l .` | 0 | no files listed |
| `cd stats && go vet ./...` | 0 | no output |
| `cd stats/web && npx tsc -b` | 0 | no output |

`check-contract-budget.sh` reporting **92** owned files is the widened ratchet from `f54889f` doing
its job: the pre-widening guard covered 13.

### `## test` — every entry exit 0

| Harness | Exit | Verdict |
|---------|-----:|---------|
| `scripts/test-setup.sh` | 0 | `✓ PASS — 472 assertions, 0 failures` |
| `scripts/test-check-references.sh` | 0 | all assertions passed |
| `scripts/test-check-plan-provenance.sh` | 0 | all assertions passed |
| `scripts/test-check-finish-preflight.sh` | 0 | all cases pass |
| `scripts/test-commit-split.sh` | 0 | all cases pass |
| `scripts/test-prepare-workspace.sh` | 0 | all cases pass |
| `scripts/test-preserve-session-records.sh` | 0 | all cases pass |
| `scripts/test-check-unfinished-work.sh` | 0 | all cases pass |
| `scripts/test-check-cleanup-complete.sh` | 0 | all cases pass |
| `scripts/test-gather-self-review-context.sh` | 0 | all cases pass |
| `scripts/test-gather-dispatch-context.sh` | 0 | all cases pass |
| `scripts/test-check-task-build-green.sh` | 0 | all cases passed |
| `scripts/test-check-task-commit-fields.sh` | 0 | all cases passed |
| `scripts/test-check-workspace-isolation.sh` | 0 | all cases pass |
| `scripts/test-workspace.sh` | 0 | `21 passed, 0 failed` |
| `scripts/test-check-contract-budget.sh` | 0 | all checks passed |
| `scripts/test-check-vocabulary.sh` | 0 | all cases passed |
| `scripts/test-check-panel-diff-size.sh` | 0 | all cases pass |
| `scripts/test-plan-dispatch-bundles.sh` | 0 | all cases passed |
| `scripts/test-check-panel-reproducers.sh` | 0 | all 38 cases plus the metacharacter loop pass |
| `scripts/test-run-reproducer.sh` | 0 | all 18 cases plus the metacharacter loop and its control pass |
| `scripts/test-check-markdown-integrity.sh` | 0 | all 25 cases pass |
| `scripts/test-check-uitest-overrides.sh` | 0 | all cases pass |
| `scripts/test-check-guard-symlinks.sh` | 0 | `✓ PASS` |
| `scripts/test-resolve-base-branch.sh` | 0 | all cases pass |
| `scripts/test-prepare-archive-branch.sh` | 0 | all cases pass |
| `scripts/test-check-stage-mark-calls.sh` | 0 | all cases passed |
| `scripts/test-check-self-review-report.sh` | 0 | all cases passed |
| `scripts/test-lib-coverage.sh` | 0 | `✓ PASS` |
| `scripts/test-check-installed-citations.sh` | 0 | all cases pass |
| `scripts/test-check-normative-inventory.sh` | 0 | all checks passed |
| `cd stats && go test ./... -race -count=1` | 0 | 14 packages `ok`, none failed |
| `cd stats/web && npm test` | 0 | `Test Files 9 passed (9)`, `Tests 163 passed (163)` |

That is **31 Bash/Python harnesses plus the Go and SPA suites** — the full `## test` set as declared,
with `test-check-normative-inventory.sh` the entry this change added.

### OpenSpec

```console
$ openspec validate --strict --all
Totals: 32 passed, 0 failed (32 items)
$ echo "exit=$?"
exit=0
```

---

## 4. No new lint suppression anywhere in the change

```console
$ git diff f09f2b1..HEAD | grep -nE "noqa|shellcheck disable|eslint-disable|@Suppress|nolint|# type: ignore"
$ echo "exit=$?"
exit=1
```

**The command printed nothing.** `grep` exit 1 is "no lines matched", which is the passing result
here. The change adds no suppression marker of any kind, in any language, and weakens no lint or
formatter configuration.

---

## 5. Install coherence

The change edits always-on rules, which `setup.sh` renders into the managed block of
`~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`. Both were checked, in a throwaway `HOME` — **never
the operator's real home** — which was deleted afterwards.

```console
$ scripts/test-setup.sh
✓ PASS — 472 assertions, 0 failures
exit=0

$ SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global
exit=0
```

| Property | Value | Task 5.1 recorded | Agrees |
|----------|------:|------------------:|--------|
| Rule blocks (`^# ` headings) in the rendered `~/.claude/CLAUDE.md` | 11 | 11 | yes |
| `Full rule:` pointers | 10 | 10 | yes |
| Symlinks in `~/.claude/rules/` | 12 | 12 | yes |
| Dangling symlinks anywhere under the sandbox `HOME` | 0 | — | — |
| Total symlinks created across `.claude`, `.codex`, `.cursor` | 56 | — | — |

The eleventh block carrying no `Full rule:` pointer is `# myflow — start → do → finish`, which points
at the pipeline contract files rather than at a single rule file — by design, not a missing pointer.
The twelfth symlink in `~/.claude/rules/` beyond the ten pointed-at rules is
`myflow-manual-review.md`, likewise not a rule with a rendered block.

`~/.codex/AGENTS.md` is byte-identical to `~/.claude/CLAUDE.md` (`diff` silent), so the reworded
be-brief rule reaches both harnesses from the one source.

Dangling symlinks were checked with `find "$SANDBOX" -type l ! -exec test -e {} \; -print`, which
printed nothing. Every one of the 56 links resolves.

The rendered be-brief block was read in full and carries this change's new clauses — the
completeness-and-non-repetition sentence, the two named subjects, the explicit carve-out for a
consuming project's pre-existing documentation, and the "Cut, never paraphrase" paragraph with its
list of what must never be cut.

---

## 6. What this verification did NOT check

A verification document that claims more than it checked is worse than one that admits its gaps.
These are the gaps.

**Meaning preservation beyond normative sentences is unverified by any automation.** The inventory in
§1 covers `SHALL`, `SHALL NOT`, `MUST` and `MUST NOT` as whole words, and nothing else. `SHOULD` and
`MAY` sentences, exit-code contracts stated without a keyword, ordering constraints, worked examples,
scenarios, and recorded reasons a rejected alternative was rejected are all things `rules/be-brief.mdc`
now forbids cutting — and **no check run here would notice if one had been cut**. The trims were
cuts-only by construction and were reviewed per task, but that is a review claim, not a measurement.
This is the largest residual risk in the change, and it is irreducible with the tooling that exists.

**A paraphrase inside a non-normative paragraph is invisible.** The inventory catches a reworded
normative sentence, because rewording changes the string. It says nothing about a reworded paragraph
that contains no keyword.

**No `/myflow-*` run was executed against the trimmed skills.** The skill text was verified as text —
budget, references, citations, vocabulary, symlinks, normative content — and as an install. It was
not verified as an executed procedure: no `/myflow-start`, `/myflow-do`, `/myflow-finish` or
`/myflow-fast` run was performed on the trimmed sources. Whether an agent following the trimmed
`skills/myflow-do/SKILL.md` behaves as one following the original does is not something this change
measured, and no harness in this repository can measure it.

**The 22-file difference between the library's corpus and a whole-tree sweep was not inventoried.**
`docs/self-review/*.md` and `stats/README.md` sit outside the scope roots and are neither ratcheted
nor inventoried. This is by design (see `f54889f`'s message), and it is not a gap *for this change*:
`git diff --stat f09f2b1..HEAD` shows every Markdown file touched lies inside the 92-file corpus, so
nothing was edited outside the inventory's reach. It remains a gap for future changes.

**The guard's own correctness is assumed, not independently proved here.** §1 runs one
implementation over two trees; if `check-normative-inventory.sh` systematically missed a class of
sentence, it would miss it identically on both sides and the diff would still be clean. What covers
that is `scripts/test-check-normative-inventory.sh` (exit 0 above), whose cases were each proved to
fail against a deliberately broken guard before being trusted — a self-test, and named as such.

**Cursor's rules are symlinked, not rendered.** The install check confirms the twelve `~/.cursor/rules/*.mdc`
links resolve to the source files. Cursor has no managed block to render, so there is no rendered
Cursor artifact to compare against `CLAUDE.md`; nothing beyond link resolution was checked there.

**No token-cost measurement was taken.** The change's motivation is bytes loaded per `/myflow-*` run.
§2 measures bytes on disk. Nobody measured the actual token count of a real run before and after, so
the practical saving is inferred from file size, not observed.

**The base-branch worktree measured the merge base, not `main`'s current tip.** `f09f2b1` is this
branch's merge base and is the correct comparison point for the change's own delta. If `main` has
moved since, this document says nothing about that divergence.

---

## 7. Status of this file

This file lives under `openspec/changes/`, which `COMMIT-PER-TASK` forbids an implementer from
committing. **It is left unstaged**; `/myflow-finish` commits the change directory with the rest of
`openspec/`. Task 9.1 makes no commit, and modified no file other than this one.

## Amendment after this document was first written

The ratchet commit's message carried a stale byte figure — 1,203,716 / 20,616, measured at
`8c3a00d` before that commit's own `.myflow/project.md` edit (+82 bytes) landed. It was amended to
the correct 1,203,798 / 20,534, moving the commit from `4c927f6` to `f54889f`. Subject and
`Task-Id: 8.1` trailer unchanged; `check-task-commit-fields.sh` re-run clean against the new sha,
and `check-contract-budget.sh` still reports BUDGET-OK for 92 files.

The correction was made rather than deferred because this change exists to stop stale figures rotting
in documentation, and its own permanent record is not exempt.
