# kan-887-flow-fix-delete-the-retained-bash-twins-and

## Why

The bash → Go guard ports left their bash sources in the tree as a live second copy of what the Go
port carries, with no decision recorded. A retained copy invites drift between the two — the
failure `helper-parity-tests` exists to prevent for the copies that must stay (KAN-887, from
kan-778's panel finding F9, widened by the 2026-10-07 migration inventory).

Reachability, checked on the base `ef6f87cd`: every file the ticket names is still present, and
for each twin below the only references left are comments and Go parity tests reading the live
bash file — no bash or Python script sources it.

## What changes

- Deleted bash twins: `scripts/reproducer-metachars.sh`, `scripts/lib/change-plan.sh` with
  `scripts/test-lib-change-plan.sh`, `scripts/lib/project-section.sh`,
  `scripts/lib/panel-touched-paths.sh`, and `scripts/lib/test-git-shim.sh` with
  `scripts/test-lib-test-git-shim.sh`.
- Deleted bash harnesses whose every case has a Go counterpart (missing ones added to Go first):
  `scripts/test-aside-planning-artifacts.sh`, `scripts/test-check-done-when-paths.sh`,
  `scripts/test-check-hand-notes-in-step.sh`, `scripts/test-project-get.sh`,
  `scripts/test-check-fast-route-record.sh`.
- `scripts/lib/spec-root.sh` stays: four live scripts still source it; it is KAN-923's.
- Go parity tests that read a deleted file read its copy at a pinned historical commit instead;
  each Go twin's header records that its bash twin was deleted by KAN-887.
