# kan-927-in-run-self-review-fixes — live verification

Run 2026-10-08 against this worktree's own flowd (`127.0.0.1:4693`, database `flow_kan_927_in_2d83`,
built from the branch tip). The dev workspace's flowd on `127.0.0.1:4173` serves the pre-change
build until the operator reloads it, so design.md's 4173 figures (96 rows, kan-916's links) are the
operator's check after that reload.

- Seeded three findings with the branch's own `flow self-review finding` (one each fixed / filed / declined).
- `GET /api/v1/stats/self-review?from=<yesterday>T00:00:00Z&to=<tomorrow>T00:00:00Z` → 200, `view: self-review`, 3 rows, newest first, `unmeasured: false`.
  - fixed `81a2ca7c` → `commitUrl` `https://github.com/tweety53/agents/commit/81a2ca7c` (origin remote mapped).
  - filed `KAN-9` and declined → empty `commitUrl`.
- Same request with `&model=opus` → 400.
- `GET /` → 200 (SPA served).

Failure look (none seen): 404 on the route, an empty table while the DB holds rows, or `fixed` rows with no link.
