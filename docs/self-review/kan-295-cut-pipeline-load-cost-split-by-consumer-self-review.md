# Self-review — kan-295-cut-pipeline-load-cost-split-by-consumer

**Jira:** KAN-295 · **Run:** `/myflow-fast` · **Rating:** not given — the operator answered the
per-angle filing prompts and supplied no 1-5 rating.

Fifteen findings across five angles, all fifteen filed. Every angle produced findings.

## Angle 1 — Problems encountered — `myflow-fix`

- **[myflow-fix]** `/myflow-fast`'s handoff prints `/myflow-fast <name>` as its `Next:` line, while the skill reads any argument at `IN_PROGRESS` as fix instructions and only a bare invocation integrates. Pasting the command the handoff itself printed therefore reads as a fix round whose instructions are the change's own name. The run had to stop and ask, immediately before push and merge. — filed: KAN-296
- **[myflow-fix]** Nothing marks a panel finding fixed, and `myflow record status` accepts values outside `open`/`fixed`/`withdrawn` without complaint. Both surfaced at the same place: the unfinished-work gate blocked run 1 on 30 findings that were fixed and confirmed clean by three panel slots, plus two more written `resolved` and counted as unrecognised. All 32 had to be closed by hand. — filed: KAN-297
- **[myflow-fix]** `check-finish-preflight.sh` accepts a bare local branch name where the finish contract requires `origin/$BASE`. The contract explains why — the fetch refreshes only remote-tracking refs, so a stale local branch silently feeds the RUN1/RUN2/REFUSE decision — but the guard's usage line says only `<base-ref>`, and the natural composition from `resolve-base-branch.sh`'s output is the wrong one. Made that mistake here; the verdict happened to agree. — filed: KAN-298
- **[myflow-fix]** Worktree cleanup check 5 runs the project's `## stop` command with no scoping to the worktree. Compose derives its project name from the directory basename — `stats` in the main checkout and in every worktree — and `container_name` is pinned, so running it from a worktree stops the main checkout's Postgres. That is the store this run still needed, and the operator's own database. — filed: KAN-299
- **[myflow-fix]** `git rebase --autosquash` silently reverted three already-applied fixes through a clean 3-way auto-merge with no conflict marker, twice in the same change. It surfaced only because the dispatcher had stopped accepting claims without pasted command output. A fix round should re-verify each repair by content after any rebase, never by the commit's presence. — filed: KAN-310

## Angle 2 — Token and time cost — `myflow-cost`

- **[myflow-cost]** Reviewing a byte-for-byte relocation through a unified diff cost 3 panel rounds, 3 fix rounds and ~2.4M subagent tokens. A diff shows lines removed here and added there; proving they are the same sentence needs cross-file comparison the format hides, so every reviewer reconstructed ~44 extraction points by hand. Round 1 is the evidence it does not work: one slot sampled 5 and found 0, another sampled 4 and found 4. — filed: KAN-300
- **[myflow-cost]** Hard-wrapped Markdown defeats line-based `grep` — a citation spans a line break, so a search for the section name and the path together returns zero for text that is present. Eight stale citations survived three full review rounds that way, and it produced two false conclusions. One unwrapped scan closed the entire tail. — filed: KAN-301
- **[myflow-cost]** The panel's Code review slot hung for 4 hours with no result, because the `code-review` skill forks a background agent whose completion the parent never observes; the stall watchdog never fired because the slot kept emitting output while waiting. Re-dispatched on the documented fallback, it returned clean in 2.5 minutes. — filed: KAN-302
- **[myflow-cost]** A numeric byte target in a fix prompt, on a change whose binding constraint was "change no wording", induced sentence-splitting at punctuation joints and three Critical findings. The prompt carried the no-paraphrase rule; the number was the stronger signal. The figure it produced was then reverted, so the target was never real. — filed: KAN-303

## Angle 3 — What went well — `myflow-improvement`

- **[myflow-improvement]** The three-slot panel produced unusually clean evidence for its own slot count: one slot found 4 defects in 4 samples where another found 0 in 5, and all three of the change's own automated verification layers passed green on those same defects, because each detects loss and a sentence cut in half is still present. The slots also disagreed on a checkable fact, which is what prompted direct verification — a single-slot panel produces no disagreement to notice. — filed: KAN-307

## Angle 4 — What could be automated — `myflow-automation`

- **[myflow-automation]** Finding citations that name a moved section but still point at the old file is fully mechanical, and doing it by hand cost three review rounds of sampling. The scan that finally closed it — unwrap, then match section name against a following path — found the complete tail in one pass, correctly excluding 44 historical records under `docs/superpowers/` and `openspec/changes/archive/`. — filed: KAN-304
- **[myflow-automation]** Closing a fix round's findings is one `myflow record status` call per finding, by hand, and remembering to do it before `/myflow-finish`. It was forgotten here and the gate caught it. Closing them as part of closing the round removes the class. — filed: KAN-305
- **[myflow-automation]** `resolve-base-branch.sh` prints a bare name and `check-finish-preflight.sh` needs a remote-tracking ref; the caller composes `origin/` between them, or silently does not. The guard should compose it, removing the opportunity rather than detecting the mistake. — filed: KAN-306

## Angle 5 — Go app and persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** The store writes whatever `-status` value it is handed, so an invalid one is accepted at write time and surfaces later as an outstanding-work refusal at the integration gate. `myflow stage` already rejects an undocumented stage key and names the alternatives; the same treatment and a `close-round` operation both belong in the store, because the findings are rows there. — filed: KAN-308
- **[myflow-stats-app]** The per-move ledger `myflow-contract-economy` requires was hand-maintained and wrong in both directions twice — a row quoting a passage never removed, four genuine extractions with no row, and a self-reported total off by one in two places. Three of its four columns are derivable from the diff; only "pointer left" needs judgment. Deriving it is not the mechanical no-loss check that spec forbids substituting for it. — filed: KAN-309

## Disclosed deviations in this run

- **Cleanup check 5 skipped**, with the operator's explicit approval, on the evidence that the
  container it would stop belongs to the main checkout — its own compose labels read
  `project=stats`, `working_dir=/Users/tweety53/Projects/agents/stats` — and that check 6 returned
  `CLEAR` for this worktree. Filed as KAN-299.
- **Merge-and-push to `main`**, overriding the operator's own no-direct-pushes-to-`main` rule. The
  conflict was raised before the route was offered, and the operator approved it explicitly.
