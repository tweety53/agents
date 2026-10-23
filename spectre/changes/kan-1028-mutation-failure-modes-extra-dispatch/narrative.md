# kan-1028-mutation-failure-modes-extra-dispatch — session narrative

## 2026-10-23 — creating run

- Tasks 1–6 were implemented inline by an earlier session, which died partway through panel round 0. Its two panel dispatches had begun but never wrote reports.
- The resumed run first reused session token `mf-k1028b`, which turned out to be the dead session's token, so it collided. Both orphaned rows were closed as `aborted` under that token, and round 0 was re-dispatched under the fresh token `mf-k1028c`.
- The first full-suite run hit a dev-postgres restart in the middle of `internal/store` (SQLSTATE 57P01 "terminating connection due to administrator command"). This run did not cause it. With the container back up, the package re-ran green, and so did the whole verify list.
- Round 0 raised 2 Important findings and 6 Minor. F1 was decided without asking: the extra-dispatch role drop now covers every resolved roster, decided rosters recorded before kan-1028 included. F6, the duplicated anonymous struct in `decision.go`, was withdrawn: the plan prescribes it, and no second abstraction caller exists yet.
- The fix rewords run-loaded sentences on purpose. The `check-verbatim-moves` violations this raised were acknowledged in `verbatim-moves.txt`.
