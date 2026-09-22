# Self-review context bundle for kan-358-put-a-pinned-clock-and-viewport-in-the-capture

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-358-put-a-pinned-clock-and-viewport-in-the-capture, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-358-put-a-pinned-clock-and-viewport-in-the-capture.md (absent)
skipped: .superpowers/sdd/reviews/kan-358-put-a-pinned-clock-and-viewport-in-the-capture-panel.md (absent)
skipped: spectre/changes/archive/kan-358-put-a-pinned-clock-and-viewport-in-the-capture/tasks.md (absent)
skipped: spectre/changes/archive/kan-358-put-a-pinned-clock-and-viewport-in-the-capture/design.md (absent)
skipped: spectre/changes/archive/kan-358-put-a-pinned-clock-and-viewport-in-the-capture/narrative.md (absent)

## git log --stat

commit 2b50579d05a53ffbbcccef073518ed82b6fb410d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 22:30:19 2026 +0300

    docs(flow): pin the clock and viewport in every authored capture spec

 skills/flow/verify-and-handoff.md | 16 ++++++++++++++++
 1 file changed, 16 insertions(+)

## Session narrative

A `/flow-fast` run invoked from the gymie checkout for a ticket that names the flow pipeline's own visual-verification stage; the run had to decide which repository the change belonged to before anything else. It read gymie-playwright first and found the viewport already pinned in `playwright.config.ts`, a `pinClock`/`PINNED_TODAY` helper already present, twelve specs using it and four fidelity specs deliberately not, because the server derives their windows from its own clock — so a blanket pin in the shared fixture would have broken them, and the missing piece was the guidance in the stage that authors each new spec, which lives here. The operator, asked once, left the repository choice to the session and moved KAN-358 out of its unrecognised `Backlog Flow` status. The change itself is one bolded rule in step 8 of `skills/flow/verify-and-handoff.md`: clock installed before the first navigation, fixture dates from the same instant in the app's timezone, viewport stated in the spec, and the server-clock exception named. The plan classed `micro`, so no panel ran — the bundle's RECORDS LOSS note reads a completed `flow.review-panel` mark as a lost panel, where in fact `/flow-fast` marks that stage through on a `default` panel. The only friction was the main-checkout hook rejecting a heredoc whose path went through a shell variable; the literal path passed.
