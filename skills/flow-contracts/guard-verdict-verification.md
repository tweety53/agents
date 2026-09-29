# Hand-verifying a guard verdict

Loaded from **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`) only when a gate guard's verdict fires.

**When a gate guard's verdict fires while the situation contradicts the pipeline's own structural
conventions, the run relays how to hand-verify that verdict, and the operator verifies before
choosing a course.** The gate guards are `check-foreign-staged.sh` (`STAGED-FOREIGN`),
`check-finish-preflight.sh` (`REFUSE`),
`check-unfinished-work.sh` (`OUTSTANDING`), `check-base-moved.sh` (`MOVED`) and
`check-cleanup-complete.sh` (`LEFTOVER`). The contradiction is structural, not evidential: a
cross-repo change whose plan resolves only in the canonical tree, a recorded merge base a rebase
made stale, a worktree the state file keeps because its removal failed. The procedure the run
relays is the guard's own — each gate guard's header carries its hand-verification paragraph,
canonical for it as the header is for the guard's every other semantic — shown alongside the
breakdown, before the prompt's courses are chosen, never instead of them.

Two failure directions are named so neither is taken. Trusting the verdict blindly blocks
verified-complete work, or teaches the operator to click past a gate that does sometimes speak
truly. Overriding it blindly silences a guard that may have seen real unfinished work. Hand
verification is the middle path: it recomputes the guard's own signals from the primary records —
the plan, the store, git — and never rewrites the verdict line or the exit code, never bypasses
the courses the gate offers. A verification that upholds the verdict leaves the operator exactly
where they were. A verdict the verification confirms structural is recorded where the call site
provides a recording — today, `flow record verdict false-positive` at the unfinished-work gate in
`skills/flow/integrate.md`, with the operator's reason, verbatim.
