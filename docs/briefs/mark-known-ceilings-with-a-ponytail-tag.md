# Mark known ceilings with a searchable tag at the site

When a change ships with a known, accepted limitation — a ceiling the code
deliberately does not go past — mark it in the code at the exact site that has
it, with a `ponytail:` comment naming what the code does not cover. The
limitation lives beside the code that has it, and the next reader — and the
next grep for `ponytail` — finds the ceiling before re-deriving it as a bug.

## Why

kan-692's review left the wasm dev server's freshness check with a known
ceiling: its build-input allowlist is an allowlist, not the build's real input
set, and a deletion alone is not seen. The fix marked that ceiling in a
`ponytail:` comment at the site rather than in a ticket or a narrative file,
where a future run re-derives it as a bug; the deferred self-review pass that
filed this promotion found the practice by grepping the repo, and no durable
statement of it anywhere.

## What belongs in the mark

- The literal tag `ponytail:` — one greppable token across every project.
- What the code does not cover, in one or two lines — the ceiling itself, not
  its history.
- What replaces it if it ever bites, where one exists ("read Gradle's task
  inputs if either ever bites").

A `ponytail:` mark records an accepted limitation, never a defect: a defect
gets a failing test and a fix. It replaces neither a ticket nor a contract
sentence — it is the pointer the next reader standing at the site needs.

Recorded from kan-692 (`kan-692-flow-fix-the-wasm-dev-server-never-hot-reloads-a`),
whose deferred self-review pass named the practice; filed as KAN-840.
