# Final review panel — kan-73-install-guard-scripts-alongside-skills

Roster: `light` (recorded). Required slots: Primary, Principles, Code review (low).
Panel model: sonnet (`models.reviewPanel`). Merge base: 7922242. Diff: 54 files, +1346/-35.
Optional slots: none — no trigger fired that the light preset escalates.

Panel diff size: `scripts/check-panel-diff-size.sh <worktree> 7922242` reported **1381, under cap**,
exit 0. No operator prompt was required. Recorded here per the contract's every-run rule.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Critical | `scripts/check-guard-symlinks.sh:129-142`, `scripts/plan-dispatch-bundles.sh:70-83`, `scripts/check-workspace-isolation.sh:117-129` | `resolve_file` duplicated; this diff takes the repo from 2 copies to 5 |
| F2 | Primary + Code review (low) | Major | `skills/myflow-do/SKILL.md:438,450,454`, `skills/myflow-do/scripts/check-unfinished-work.sh` | `check-unfinished-work.sh` is shipped to `myflow-do` but its presence-check list omits it, and no site names it as an invocation |
| F3 | Primary | Minor | `scripts/check-guard-symlinks.sh:311-346` | rule 3 would false-fail a future `Run \`scripts/check-vocabulary.sh\`` — the six project-configured guards are exempt by `pipeline.md` but the classifier does not know it |
| F4 | Code review (low) | Major | `scripts/check-guard-symlinks.sh:456` | rule 2 checks nothing for `myflow-fast` — its SKILL.md cites no guard by basename, so its 17 symlinks are unverified |
| F5 | Code review (low) | Major | `scripts/check-guard-symlinks.sh:269` | rule 4's regex matches only the literal `$SCRIPT_DIR/..`; `dirname "$SCRIPT_DIR"` is the same bug class and passes |
| F6 | Code review (low) | Minor | `scripts/check-guard-symlinks.sh:415` | fence scanner reads only a line's leading token, so `./guard.sh` and `bash guard.sh` are never recorded as required |
| F7 | Code review (low) | Minor | `scripts/check-guard-symlinks.sh:420` | positional backtick pairing mis-pairs after a stray backtick, dropping a real citation |
| F8 | Code review (low) | Minor | `scripts/check-guard-symlinks.sh:335` | rule 3 skips ```console and ```text fences, and matches `shell`/`shellsession` only loosely via unanchored `^sh` |
| F9 | Code review (low) | Minor | `scripts/check-guard-symlinks.sh:228` | rule 1's absolute-target branch lacks the `continue` its siblings have, feeding an off-repo path into rule 4 |
| F10 | Code review (low) | Minor | `scripts/check-guard-symlinks.sh:535` | the sibling-closure `grep` omits `-a` and the `rc >= 2` refusal the file's own header requires |
| F11 | Code review (low) | Minor | `scripts/check-workspace-isolation.sh:131` | `REPO_ROOT` resolves unconditionally though only the no-args branch reads it, so it can abort the guard on the production path |

## Slot notes

**Principles (slot 2)** — one Critical, no Important. Confirmed clean, with evidence: the
guard-to-skill map and sibling derivation in `check-guard-symlinks.sh` are genuinely derived from
the tree and from each guard's own source, not hardcoded; `test-setup.sh`'s `GUARD_SKILLS` is
likewise derived; the resolution rule is stated exactly once in `pipeline.md` with no second copy
anywhere. Judged the presence check as earning its keep, and `check-guard-symlinks.sh`'s size as
explained by four independent concerns rather than sprawl (explicitly not raised as a violation).

Parent verification of F1's factual claims, each checked independently rather than taken on trust:
  - 5 copies of `resolve_file` confirmed: check-guard-symlinks.sh, gather-self-review-context.sh,
    plan-dispatch-bundles.sh, check-workspace-isolation.sh, preserve-session-records.sh.
  - `check-guard-symlinks.sh` is NOT symlinked into any skill -- confirmed, no matches under
    skills/*/scripts/. So the "single-file by design, copied into projects one at a time" defence
    genuinely does not cover it.
  - `lib` IS already symlinked beside `check-workspace-isolation.sh` in myflow-do/scripts, so a
    shared `scripts/lib/resolve-file.sh` would ship through the mechanism already in use, with no
    new install step. Confirmed by `ls -l`.
  - The surviving counter-point is narrower than F1 allows: check-workspace-isolation.sh's header
    asserts a MANUAL-COPY distribution mode, which the symlink farm supersedes for the shipped
    guards but does not erase. Task 1's reviewer reached the opposite conclusion on exactly this
    file, recommending its copy stay inline.

F1 is OPEN. A withdrawal is the operator's decision at handback, never one this run may write for
itself, so it is recorded open pending either a fix or the operator's call.

## Marker block

findings-total: 11
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed

**Primary (slot 0)** — no Critical, no Major; two Minor (F2, F3). Verified clean with evidence: all
7 commits present in plan order with the exact subjects their `Commit:` fields name; all symlinks
mode 120000, relative, resolving; `myflow-status` correctly carries none; every guard and harness
passes against the real tree; no `openspec/` or `docs/superpowers/` edits; bash 3.2 and `/bin/sh`
portability clean across every new and modified script; and the mid-run corrections are reflected
consistently in design.md, tasks.md, the SKILL diffs and the symlink farm alike.

Primary independently corroborated F1 and materially strengthened it: the fifth copy has **already
diverged** -- `check-guard-symlinks.sh`'s `resolve_file` adds `--` before every `readlink`, `dirname`
and `cd` argument; the other four do not. Parent confirmed this by direct diff. Drift inside a single
change is exactly what `scripts/lib/panel-record.sh`'s header warns duplication causes, so F1 is no
longer a hypothetical risk. Primary also noted the copy-paste was plan-MANDATED (tasks 1 and 5 both
say "reuse/port ... rather than writing a third"), so the defect originates in the parent's plan.

Parent verification of F2: both mentions at :438 and :450 are prose about what the guard reads, but
:454's "Run the guard and read its own output" is an imperative aimed at whoever is debugging a
rejected record. So the guard IS reachable from a `/myflow-do` run, the shipped symlink is correct,
and what is wrong is that task 4's presence-check enumeration omits it and no site names it in a form
rule 2 can see. Resolving toward "it is invoked" rather than removing the symlink.

Reproducers supplied: F1 a `diff` of two function bodies; F3 a runnable fixture-tree command that
trips the classifier on legitimately exempt prose. F2's is a `grep` over the file.

**Code review (low), slot 3** — 3 Major, 7 Minor, and the most productive slot of the three despite
the lowest effort setting. Every finding carried a runnable reproducer. It found two gaps neither
other slot saw (F4, F5), and independently reached F2, which Primary had rated Minor and it rates
Major -- recorded at the higher severity, since the two agree on the substance and the higher rating
is the safer reading.

Parent verification of the two most consequential, neither taken on trust:
  - F4 CONFIRMED by mutation: deleting skills/myflow-fast/scripts/plan-dispatch-bundles.sh leaves the
    guard reporting GUARD-SYMLINKS-OK at exit 0. The skill carrying the MOST symlinks (17) is the one
    rule 2 does not cover at all, because myflow-fast delegates entirely by citation and names no
    guard in a form the classifier can see. Symlink restored; tree clean.
  - F5 CONFIRMED by reading the regex at :269 -- `\$\{?SCRIPT_DIR\}?"?/\.\.` matches only that one
    spelling. `REPO_ROOT="$(dirname "$SCRIPT_DIR")"` is the identical defect differently written and
    passes rule 4 clean. This matters more than an ordinary gap because rule 4 exists specifically to
    stop KAN-73's own bug class from returning.

Slot 3 also reported the inverse of F1's divergence: FOUR of the five resolve_file copies omit the
`--` terminator that the fifth copy's own header calls a required discipline. The two readings agree
-- the copies are out of step with each other, and which one is "right" is itself now ambiguous.

No slot found a defect in quoting, word-splitting, `set -euo pipefail` interaction, or macOS `awk`
handling. Every finding against check-guard-symlinks.sh is a gap in what it DETECTS, not a failure it
reports: all 426 test-setup assertions and all 24 harnesses pass.

## Fix round — all 11 closed

Applied on the panel-fix model (sonnet). Every fix folded into the commit that introduced it via
`git commit --fixup` plus one `git rebase --autosquash`. Post-fix shas: task 1 `0f2b62f`,
task 2 `3cf9527`, task 3 `175b42d`, task 4 `27e6946`, task 7 `5ef9f71`, task 5 `f210dec`,
task 6 `9355b92`.

F1's judgement call went the other way from task 1's reviewer: `check-workspace-isolation.sh` was
converted to source the shared library rather than keeping its inline copy. The reasoning given is
that its "single-file, copied into projects one at a time" header describes guards hand-copied into
other projects' own tooling, whereas this guard's actual distribution is exclusively the symlink farm
this change builds, where `lib` travels beside it unconditionally. Copies went from five to two; the
two that remain are pre-existing and out of scope.

**PARENT VERIFICATION by mutation, not by report:**
  - F4: deleting skills/myflow-fast/scripts/plan-dispatch-bundles.sh now exits 1 and names the
    delegation chain ("myflow-fast delegates to myflow-do's own guard presence check"). Before the
    fix this same mutation returned GUARD-SYMLINKS-OK at exit 0.
  - F5: appending `BOGUS_ROOT="$(dirname "$SCRIPT_DIR")"` to a shipped guard now exits 1 and cites
    rule 4. Before the fix it passed clean.
  - F1: `grep -ln '^resolve_file() {' scripts/*.sh` returns 2 files, both pre-existing; three guards
    now source scripts/lib/resolve-file.sh; `ls skills/myflow-do/scripts/lib/` shows resolve-file.sh
    shipping through the existing symlink; a guard invoked through the skill-dir path still exits 0.
  Tree restored clean after every mutation.

**Honest limitation, reported by the fix agent rather than papered over:** F10's `rc >= 2` refusal
path has no black-box reproducer. `seen_file` lives in the guard's own `mktemp -d` scratch directory
and its contents are always guard-generated clean ASCII, so neither an unreadable file nor a NUL byte
can be injected without white-box hooks the guard does not expose. The added fixture pins the
happy-path dedup behaviour instead, and passes both before and after the fix. The `-a` and the rc
split were still added, because the file's own header requires them of every grep.

**Verification after the fix round:** all 10 lint guards exit 0; all 24 test harnesses pass;
`test-setup.sh` holds at 426 assertions; `check-guard-symlinks.sh` exits 0 against the real tree;
every one of the seven commits is individually budget-clean. Two pre-existing fixtures needed
`scripts/lib/` added to their scratch trees after the F1 extraction, which the fix agent did.

**Panel clean: zero open findings at any severity.**
