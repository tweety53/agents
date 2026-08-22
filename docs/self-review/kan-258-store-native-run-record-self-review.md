# Self-review — kan-258-store-native-run-record

**Rating: 3 — mixed.** Got there, but the cost and the number of corrections were higher than they
should have been for the result.

**Run shape:** 24 planned tasks (19 up front, 5 added mid-run), 17 commits reshaped to 2 at finish,
13 implementer dispatches on `opus`, 13 per-task reviews and 3 panel slots on `sonnet`, 5 per-task fix
rounds and 2 panel fix rounds on `opus`. Roster `light`, planning effort `default`.

**Defect yield:** 27 plan or spec defects corrected during implementation; 15 review-panel findings,
6 Major, 14 fixed and 1 withdrawn with reason.

## Angle 1 — Problems encountered — `myflow-fix`

- **[myflow-fix]** `check-task-commit-fields.sh` was unusable for the whole run. It refuses when more than one `tasks.md` exists under `openspec/changes/`, and kan-77's merged-but-unarchived directory sat there throughout, so all 24 task commits were hand-checked. The guard is already passed the change name; resolving by that name rather than scanning and refusing on ambiguity would close it. — declined
- **[myflow-fix]** OpenSpec's MODIFIED-block rule cost three archive attempts. Renaming a scenario inside a MODIFIED requirement silently deletes it on archive, because the archive replaces the block wholesale by header match. `openspec archive` refuses correctly, but one requirement at a time, so the same trap was hit three times running — `myflow-finish-cleanup` (three scenarios), `myflow-model-policy`, `myflow-review-panel-economics`. Running the same comparison at `/myflow-do`'s validate step would surface all of them before any commit. — declined
- **[myflow-fix]** The reproducer format was wrong for all 15 panel findings on first write. `check-panel-reproducers.sh` requires a bare path with plain arguments or the `none — <reason>` exemption, and nothing in the panel section states that shape, so every slot returned shell command lines and the dispatcher transcribed them. All 15 had to be rewritten. — declined
- **[myflow-fix]** `stage end -outcome` accepts any string. Probing for the valid values through the CLI wrote a literal `bogus` outcome into the stage journal; nothing validated it, though `-stage` is checked byte-for-byte against a documented table. A corrective `-outcome abandoned` was issued and file order wins on replay; removing the bad entry was blocked by the permission classifier and was not worked around. — declined
- **[myflow-fix]** `go vet ./...` cannot run in a fresh worktree until the SPA `dist` is built, so every task vetted only its own packages and `internal/web` silently failed to compile from task 4 until task 19 found it. Building `dist` during workspace preparation would make the whole-tree vet available from task 1. — declined

## Angle 2 — Token and time cost — `myflow-cost`

- **[myflow-cost]** The 2000-line panel cap is mis-calibrated for a 24-task change. The diff measured 9762; splitting was offered and first chosen, then reversed. Running the panel found six Major cross-commit defects that splitting would have hidden, because the seam between commits is exactly where they lived. — filed: KAN-287
- **[myflow-cost]** The dispatch-context bundle was rebuilt twice, ~150KB each, with nothing changed between. The per-stage rebuild rule is correct — a fix edits the plan mid-run — but a content hash preserves it without paying for a regeneration that cannot differ. — filed: KAN-288

## Angle 3 — What went well — `myflow-improvement`

- **[myflow-improvement]** Requiring reviewers to reproduce rather than read caught four defects that reading missed: the `created`-flag concurrency race (20 concurrent upserts, five runs), `jsonb_deep_add`'s semantics (queried against the live database), the marker-neutralisation surface (twelve encodings against the real guard), and `tokenLine`'s absent-vs-zero test. That last is the general case, and it recurred four times: a test backed by a fake or a hand-built value passes while the real integration is broken. The instruction lived in the dispatcher's prose each time, not in the slot template. — filed: KAN-289
- **[myflow-improvement]** Red/green pairing surfaced 27 plan defects during implementation, each corrected in the plan rather than worked around, so the plan is a record of what was true rather than what was guessed. Nothing measures that yield, which makes the practice's value invisible when someone proposes dropping it to save dispatches. — filed: KAN-290
- **[myflow-improvement]** Implementers reported rather than silently working around. Every one of the six Major panel findings and most of the 27 plan defects reached the dispatcher because an implementer said "this is wrong" instead of quietly adapting — the `jsonb_deep_merge` cost overwrite, the proposal artifact nothing preserved, and the panel-record path mismatch among them. — declined

## Angle 4 — Automation — `myflow-automation`

- **[myflow-automation]** Ticking checkboxes after review is mechanical and was done by hand twelve times, each with an ad-hoc snippet, on the one file that gates finishing. — filed: KAN-291
- **[myflow-automation]** Assembling the panel record by hand — table, marker block, reproducer block, `F<n>` renumbering across three slots — is exactly what this change makes store-native, so it is already closing. — declined

## Angle 5 — Go app and persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** The `Records:` count needed a whole new CLI subcommand because nothing exposed it, and every other handoff-line input is still derived by the agent reading files — task counts, panel finding counts, guard results, decision counts. `handoff-blocks.md` defines the block's shape in one place; its values have no single source. — filed: KAN-292
- **[myflow-stats-app]** The proposal-artifact copy is the last file-based preservation step, and its path protections now live in contract prose rather than in code beside the renderer's — the same rules twice, in two languages. Its source-containment check was dropped in the retirement and restored only because a panel slot found it. — filed: KAN-293

## Outcome

Seven issues filed: KAN-287, KAN-288, KAN-289, KAN-290, KAN-291, KAN-292, KAN-293. Angle 1's five were
declined by the operator and stay here.

**The most serious defect was the dispatcher's own.** `design.md` recorded that one `myflow record
dispatch` call at close was sufficient, reasoning that a begin/end pair "would double the call count
and buy nothing". That was false: the harvester commits its transcript offset every 5s and never
re-reads past it, so a row that does not exist until close means every tick elapsing inside the
dispatch finds nothing to attribute to. Per-dispatch cost — one of the four records this change exists
to make trustworthy — would have been quietly wrong. The review panel found it; no per-task review
could have. The decision is superseded by `dispatch-begin-and-end` rather than deleted, with the
reasoning error recorded beside it.

**The operator overrode a standing rule at the landing question**, choosing merge-and-push into `main`
against `no-direct-pushes-to-main`, having been shown what that meant. Recorded because the rule is
agent-enforced and its override should not be invisible.
