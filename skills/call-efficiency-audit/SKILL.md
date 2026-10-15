---
name: call-efficiency-audit
description: Measure a running dev stack's call efficiency — browser request log over cold opens, per-endpoint SQL statement counts, cache MONITOR — and report repeated reads, repeated queries within a request and per-item queries, with a findings table and a before/after comparison. Standalone, not a pipeline stage. Use for /call-efficiency-audit, "audit call efficiency", "duplicate requests", "N+1".
license: MIT
---

# Call-efficiency audit

Measure how many calls a running app makes — client to gateway, gateway to backend, backend to
database and cache — and report every call it does not need, with measured evidence. The method
is the KAN-924 audit's, generalised for any project's dev stack.

## Never production

The stack is a dev or worktree stack, never production (`rules/never-touch-production.mdc`). Every
setting the audit changes — statement logging above all — is set on that stack's own database only
and reset afterwards. Name the stack in the report: its URLs, its database and its cache index.

## Scope

Cover each layer the app has: the client app, the gateway, the database and the cache. Name what
is out of scope in the report, and why (an operator-only admin frontend with low traffic, for one).

## Browser request log

1. Open a fresh browser context — no site data, an empty offline store — sign in by API and load
   the app. Record every request to the backend for a fixed window, each with its start and end
   time from the start of the run.
2. Do three cold opens, each in a fresh context.
3. Then walk a fixed route across the main screens, the same dwell on each, and record each step's
   requests.
4. Write a per-path count table for one run, ids folded to `{id}`, with the count of distinct ids.
5. Write the repeats table: the read, its calls per run, whether the second call started before the
   first answered (overlapping), and the callers found in code for each call.
6. Log one more cold open through CDP's `Network.requestWillBeSent` to count CORS preflights, which
   a page-level request log does not show.

## Per-endpoint statement count

1. Turn statement logging on for the stack's database only —
   `ALTER DATABASE <db> SET log_statement = 'all'` — and reset it afterwards.
2. Call each endpoint once, warm, through the gateway, and count the statements logged in a short
   fixed window after the call.
3. Do not count `BEGIN`, `COMMIT` or `SET`; count transactions from `BEGIN`.
4. Write a table of statements and transactions per endpoint.
5. Check index use with `EXPLAIN` on the same database for every query a finding rests on.

## Cache

Run `redis-cli MONITOR` (or the cache's equivalent) over one full walk. Look for repeated key reads,
keys with no TTL, `KEYS` or `SCAN` on hot paths, reads that could be one `MGET` or a pipeline, and
the per-request cost of a rate limiter.

## Checklists

One table per layer, columns `Check | Verdict | Evidence`. A verdict is `finding <id>`, `none`,
`by design`, `not seen`, `not measured` or `not in scope`; evidence is a measurement or a
`path:line`.

**Frontend**

| Check | Verdict | Evidence |
|-------|---------|----------|
| Repeated reads on open | | |
| Waterfalls — a read waiting on another it does not need | | |
| Refresh stampede — concurrent token refreshes | | |
| Uncancelled reads, or reads cancelled and sent again | | |
| Polling while hidden | | |
| Swallowed failures | | |
| Retries without backoff | | |
| Local cache or change-queue write amplification | | |
| Effects keyed on unstable values | | |
| Prefetch of data never opened | | |
| Dead reads — a result nothing uses | | |
| Wide reads on every visit | | |

**Database**

| Check | Verdict | Evidence |
|-------|---------|----------|
| Lazy relations touched in lists | | |
| Entity loaded twice per request | | |
| One transaction per repository call | | |
| Per-owner queries in a loop | | |
| Repeated selects inside one read | | |
| Missing foreign-key indexes | | |
| Save-in-loop, read-before-update | | |
| Transactions spanning cache, storage, mail or HTTP calls | | |
| `OFFSET` / `COUNT(*)` paging | | |
| `FOR UPDATE` ordering | | |
| Scheduled jobs scanning tables | | |

**Cache**

| Check | Verdict | Evidence |
|-------|---------|----------|
| Repeated key reads | | |
| Keys with no TTL | | |
| `KEYS` / `SCAN` on hot paths | | |
| `MGET` or pipeline candidates | | |
| Stampede on expiry | | |
| Per-request rate-limiter cost | | |

**Gateway and other**

| Check | Verdict | Evidence |
|-------|---------|----------|
| CORS preflights, and the CORS block's `maxAge` | | |
| User or identity resolved in both gateway and backend | | |
| Token parsed per filter | | |
| Presigned URL generated per row | | |
| Outbound mail or HTTP inside a transaction | | |
| Per-request logging and metrics | | |
| Health probe on open | | |

## Findings

Every finding carries measured evidence, in one table:

| ID | Area | Evidence | Severity | Action |
|----|------|----------|----------|--------|

Then a **Prevention** table grouping the finding classes that recur — two or more instances — with
their count and finding ids. Three classes recurred in KAN-924, and the `call-efficiency` panel slot
(`skills/flow/call-efficiency-reviewer-prompt.md`) asks about each in every diff: the same read sent
by two independent callers, the same row or query repeated within one request, and one query per
item of a list.

## Before/after

After a fix, measure again the same way on the same stack and write the before/after table: each
request's count before and after, and the total.

## Commit nothing measuring

The measuring tooling — scripts, log captures, browser drivers — is throwaway. Only the report is
kept.
