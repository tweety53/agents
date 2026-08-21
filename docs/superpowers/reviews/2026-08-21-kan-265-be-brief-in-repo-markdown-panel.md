# Review panel — kan-265-be-brief-in-repo-markdown

findings-total: 7

**Roster:** `light` — Primary, Principles, Code review. All three required, all dispatched, in the
single affected worktree.

**Diff size.** 1917 changed lines against merge base `f09f2b1`, under the 2000 cap —
`check-panel-diff-size.sh` exit 0, "under cap". No over-cap decision was needed.

**Optional slots.** None dispatched. Recorded as *not selected by the operator*, distinct from a
trigger that never fired: the operator's standing answer for this run's optional slots is none.

**Models.** Every slot on `models.reviewPanel` = sonnet, as recorded in the state file. Implementers
ran on `models.implementation` = opus and panel fixes on `models.panelFix` = opus, both operator
overrides of this command's recorded defaults.

## A disclosure the panel was given

**The per-task combined review ran for 3 of 11 tasks only** — 1.1, 2.1 and 4.1 — and was skipped for
4.2, 4.3, 5.1, 6.1, 7.1, 7.2, 8.1 and 9.1. Those eight were verified mechanically (commit-fields
guard, normative-inventory diff, full lint and harness sets) plus the parent's spot checks, but had
no review eyes before this panel.

Each of the three per-task reviews that did run found a Major, which is the argument against the
omission rather than for it. The operator was told, and chose a whole-diff panel over back-filling
the eight — so every panel slot was dispatched knowing it is the first reviewer on those commits.

## Passes

**Pass 1** — full roster, whole-change diff `f09f2b1..HEAD`.

### Principles — 2 Minor, no Blocker

- **F1 (Minor)** — the "why one shared lib" rationale is written three times, in
  `scripts/lib/owned-corpus.sh`'s header and again in each caller's header, each in different words.
  `scripts/` is not a corpus root, so no guard this change adds would ever catch it drifting.
- **F2 (Minor)** — `check-contract-budget.sh`'s symlink sweep re-implements the lib's
  maxdepth-1-plus-scope-dirs traversal shape with `-type l`. It reuses the shared array and
  `owned_corpus_excluded`, so it is WET rather than coupled; judged defensible at three lines.

Verified under mutation rather than asserted: disabling the symlink sweep in a scratch copy turns a
dangling `.md` symlink from a per-file violation into a silent `BUDGET-OK`, so the sweep's
justification holds. Both spot-checked mutation-proof claims held — breaking
`owned_corpus_excluded`'s `node_modules` case reddened the inventory harness, and neutralising the
sweep reddened all three symlink cases in the budget harness.

**On whether this change created an eighth rotted copy.** The new spec requirement and the new
`be-brief.mdc` clause share near-verbatim text, including the never-cut list. Investigated and
rejected as a finding: `rules/commit-scope-is-the-module.mdc` and
`openspec/specs/myflow-commit-scope/spec.md` show the same shape, and
`grep -rn "rules/.*\.mdc" openspec/specs/*.md` returns nothing — no spec in this repository cites a
rule file. Rules carry the operative text that reaches a session's prompt; specs carry the archival
SHALL-form record. A pre-existing, load-bearing convention, not duplication this change introduced.

### Primary — clean

Every "still carried by X" claim in every trim commit message was checked against X's actual current
text, across `4e971bd`, `6186c7e`, `e4ead90`, `08faeb7` and `79e444f`. Every claimed holder states
the claimed content verbatim. No paraphrase; no `SHOULD`/`MAY` sentence, exit-code contract,
ordering constraint, scenario or worked example was cut.

- `8c3a00d`, the highest-risk commit, read hunk by hunk: all ~30 remove only the trailing
  instruction, leaving every citation path untouched.
- 8 budget rows spot-checked against `wc -c`: each exactly `size × 1.25`.
- Cross-task chains traced — `pipeline.md` § Model policy (cited by cuts in 4.3, 5.1 and 7.1), each
  SKILL's "load pipeline.md first" line, `agent-baseline.md`'s 10-row table: all intact. **No holder
  was cut out from under an earlier task's citation** — the failure mode a per-task reviewer cannot
  see, and the reason the whole-diff panel was the right substitute for back-filling.
- Residual disclosed, not buried: a theoretical table-cell split on `|` ahead of backtick-escaped
  pipes, with no instance in the corpus that trips it.

Guards re-run independently in the reviewer's own session rather than trusted from `verification.md`.

### Code review — 1 Major, 3 Minor

- **F3 (Major)** — a **symlinked scope directory** silently shrank the corpus in both guards, exit 0,
  no warning. `[ -d ]`/`[ -r ]`/`[ -x ]` follow symlinks, so the root passed as present; `find -type f`
  in physical mode does not descend into a symlinked start point, so the whole subtree returned zero
  files. `check-contract-budget.sh`'s own symlink sweep could not catch it either — it filters
  `-name '*.md'`, and the directory link is named `skills`.

  This contradicts the lib's own header: *"A missing scope root is refused rather than skipped —
  skipping it would hand the caller a quietly smaller corpus that still looks like a complete
  answer."* A missing root was refused; a symlinked one was not. Reproduced by the parent before
  acting: 1 sentence from 1 file with the link, 2 from 2 without, exit 0 both times.

- **F4 (Minor)** — `test-check-normative-inventory.sh:306` piped `inv()` inside an `if` condition,
  discarding the `exit 2` its own documented hard-fail contract relies on.
- **F5 (Minor)** — `check-contract-budget.sh` printed violations in raw `find` order while its
  sibling sorts; confirmed live that `z, a, m` printed as `a, z, m`.
- **F6 (Minor)** — two `printf | tr` substitutions unguarded in scripts whose headers state every
  fallible step is guarded.

`shellcheck` is not installed on this machine. The reviewer **declined to install it** — read-only,
and adding a dependency is not a reviewer's call — and said so rather than skipping silently. The
shell review was manual reading plus live fixture execution.

## Fix round 1 — all four, on `models.panelFix` = opus

Rebased `764eb77` → `92e4fd4` and `f54889f` → `83a2f0b`; commit-message stream byte-identical, all
ten `Task-Id:` trailers intact.

The mutation proof was structural rather than contrived: both harness cases were written first and
run against the **unmodified lib** — the pre-fix code is the mutant. Each fixture was built so a
skipping guard exits 0 with a *plausible smaller answer* rather than tripping the empty-corpus
refusal, which would have passed the case for the wrong reason.

**One deliberate deviation from the dispatch, and it was right.** The budget harness case was folded
into the ratchet commit, not the lib commit: at the lib commit `check-contract-budget.sh` does not
yet source the lib, so the case could not pass there and the commit would have stopped being
build-green.

## Pass 2 — fix verification, and a sibling of the Major

F4, F5, F6 confirmed fixed. F3 confirmed fixed for its own case, with no over-refusal.

**F7 (Major), found in pass 2** — the same class one level deeper: a symlinked directory *nested
inside* a scope root is dropped silently, because the `-L` check guards only the scope root itself.
It bites only when the link's target lies outside the scope root's own reach; the reviewer's first
reproduction attempt pointed inside the same root, saw no bug, and it correctly narrowed the
condition rather than reporting "cannot reproduce".

**The obvious fix would have broken the repository.** Refusing every symlinked directory under a
scope root fails on the real tree: `skills/{myflow-do,myflow-fast,myflow-finish}/scripts/lib ->
../../../scripts/lib` are deliberate, from the change that installs guard scripts beside the skills
that invoke them. They are safe because `scripts/lib` holds `.sh` and no Markdown — nothing is lost
by not descending.

## Fix round 2 — F7, on opus

**The rule:** a symlinked directory nested inside a scope root is refused when following it reaches
any `.md`/`.mdc` file, and skipped otherwise. The only harm is a lost file, so "does following it
reveal Markdown at all" is exactly the question — no set comparison, no identity resolution.

Both new harness cases are non-vacuous in both directions: the refuse case run against pre-fix HEAD
gives `exit 0, wanted 2`; the skip case run against a blanket-refusal mutant goes red too.

Rebased to `e0d6a33 2f5080a 05ad893 61756eb f25dfef 1852c65 9060296 7556983 b0ba285 654ad2b`.
Subjects unchanged, `Task-Id:` count 10.

## Closure

**The F7 fix was verified by the parent, not by a third review pass** — recorded as a limit of this
panel rather than left implicit. The parent reproduced both directions independently (nested link
holding Markdown → exit 2; link holding none → exit 0), confirmed the three real `scripts/lib` links
still pass, and re-ran the complete lint and test sets in its own session.

Final state, parent-verified: 12 lint guards + markdown integrity green, `openspec validate --strict
--all` 32/32, all 31 harnesses green, inventory 1078 sentences from 92 files byte-identical to
`normative-baseline.txt`, `BUDGET-OK: 92`, no new lint suppression anywhere in the diff.

**Panel verdict: clean.** Two Majors found and fixed, three Minors found and fixed, no finding left
open.

## Findings

| ID | Slot | Severity | Location | Finding |
|----|------|----------|----------|---------|
| F1 | Principles | Minor | `scripts/lib/owned-corpus.sh:1`, both callers' headers | The "why one shared lib, not two disagreeing guards" rationale is written independently three times, each in different words. `scripts/` is not a corpus root, so no guard this change adds would catch it drifting. |
| F2 | Principles | Minor | `scripts/check-contract-budget.sh:278` | The symlink sweep re-implements the lib's maxdepth-1-plus-scope-dirs traversal with `-type l`. Reuses the shared array and `owned_corpus_excluded`, so WET rather than coupled; judged defensible at three lines. |
| F3 | Code review | Major | `scripts/lib/owned-corpus.sh:88` | A **symlinked scope directory** silently shrank the corpus in both guards, exit 0, no warning: `-d`/`-r`/`-x` follow symlinks so the root passed as present, while `find -type f` in physical mode does not descend into a symlinked start point. Contradicts the lib's own header guarantee that a scope root is refused rather than skipped. Reproduced by the parent. |
| F4 | Code review | Minor | `scripts/test-check-normative-inventory.sh:306` | One `inv()` call site piped its output inside an `if` condition, discarding the `exit 2` its own documented hard-fail contract relies on. Would degrade a harness abort into a misleading generic FAIL. |
| F5 | Code review | Minor | `scripts/check-contract-budget.sh` | Violation lines printed in raw `find` order while the sibling guard sorts; confirmed live that three files created `z, a, m` printed as `a, z, m`. Presentation only — counts and exit codes are order-independent. |
| F6 | Code review | Minor | `scripts/check-normative-inventory.sh:262`, `scripts/check-contract-budget.sh:246` | Two `printf \| tr` substitutions unguarded, in scripts whose own headers state every fallible step is guarded so its status cannot escape as the guard's. |
| F7 | Code review (pass 2) | Major | `scripts/lib/owned-corpus.sh` | The F3 class one level deeper: a symlinked directory **nested inside** a scope root is dropped silently, because the `-L` check guards only the scope root itself. Bites only when the link's target lies outside the scope root's own reach. Reproduced by the parent. |

## Finding status

finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

F1 and F2 are the Principles slot's Minors on comment-level DRY in the two new shell files, both
closed in fix round 1 alongside the Code review slot's findings. F3 (Major, symlinked scope root)
and F4–F6 (Minors) came from the Code review slot in pass 1. F7 (Major, nested symlinked directory)
was found by pass 2 and closed in fix round 2.

Every finding is `fixed`; none was withdrawn, and none is open.
