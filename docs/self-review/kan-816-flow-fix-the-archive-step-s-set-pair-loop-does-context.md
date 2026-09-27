# Self-review context bundle for kan-816-flow-fix-the-archive-step-s-set-pair-loop-does

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-816-flow-fix-the-archive-step-s-set-pair-loop-does, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-816-flow-fix-the-archive-step-s-set-pair-loop-does.md (absent)
skipped: .superpowers/sdd/reviews/kan-816-flow-fix-the-archive-step-s-set-pair-loop-does-panel.md (absent)
skipped: spectre/changes/archive/kan-816-flow-fix-the-archive-step-s-set-pair-loop-does/tasks.md (absent)
skipped: spectre/changes/archive/kan-816-flow-fix-the-archive-step-s-set-pair-loop-does/design.md (absent)
skipped: spectre/changes/archive/kan-816-flow-fix-the-archive-step-s-set-pair-loop-does/narrative.md (absent)

## git log --stat

commit 1f0682404575ea34003b61329a89ced2e47bd2ea
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Sep 27 22:01:09 2026 +0300

    fix(flow): pin the archive record-preservation loop to bash

 skills/flow/archive.md | 10 +++++++---
 1 file changed, 7 insertions(+), 3 deletions(-)

## Session narrative

A `/flow-fast` micro run (inline, no panel — the RECORDS LOSS note above is that no-panel case,
not a lost workspace store: no dispatch rows were ever meant to exist) fixing KAN-816: the
`/flow` archive step's record-preservation loop used `set -- $pair`, which zsh does not
word-split, so under the operator's zsh the archive commit silently omitted `ledger.md` and
`panel.md`. The defect was reproduced first in a temp tree (zsh: `$1` holds the whole pair, `$2`
empty, nothing copied; `bash -c`-wrapped: both records copied), then fixed per the issue's stated
preference by pinning the loop's invocation through `bash -c` inside the
`skills/flow/archive.md` snippet — the rest of the step-4 block is shell-neutral and stays
unwrapped, and `finish-contract-run2.md` step 4, which states the requirement in prose only,
needed no edit. A scan of every fenced snippet across the flow skills found no other
unquoted-splitting reliance (`implement.md`'s `$(seq 1 48)` is command substitution, which zsh
splits by default), making this loop the whole defect. Where the run struggled: the plan failed
its first `check-plan-shape.sh` pass because the field family was indented two columns —
`FIELD_RE` anchors at column 0 — and the Derive-the-name step is judgment, not mechanics: the
mechanical 48-char summary truncation yields a dangling "does" (as kan-841's yields "to"), and the
descriptive-slug alternative was weighed before settling on the mechanical form for precedent.
Full `## lint` green in the worktree after the fresh-checkout `make web-build` prerequisite; the
26 script guards, the normative-inventory diff against the pre-edit tree (unchanged), and
gofmt/vet/tsc all pass.
