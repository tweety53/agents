# Design — kan-172

## Decision: bind after the fact, by a correlator the caller writes

**Chosen.** The skill writes a literal unique nonce into the `stage begin` command. That command text
is recorded in the calling session's own transcript. The daemon later finds the nonce there and binds
the stage run's `session_id` to that transcript's `sessionId`.

**Alternatives rejected:**

- *Newest transcript in the directory.* Claude Code flushes transcript entries per **turn**, so at
  mark time the marking session's file may be a minute stale while another session's is fresh — the
  resolver would pick the wrong session, systematically, and silently. Observed on this machine with
  two concurrent sessions.
- *Match on `cwd` / `gitBranch`.* Strong once a worktree exists, useless before it does: three live
  sessions on 2026-08-14 all reported `cwd=/Users/tweety53/Projects/agents`, `gitBranch=main`.
- *Timestamp overlap alone.* Cannot separate concurrent sessions at all, and would attribute one
  session's tokens to another's stages with nothing able to detect it.
- *An environment variable.* Claude Code exposes `CLAUDECODE`, `CLAUDE_CODE_ENTRYPOINT` and a
  messaging socket, but no session id.

## Decision: one token per session, not one per mark

**Chosen.** A command generates **one** session token at the start of its run and passes it on every
mark it makes. The first mark carrying an unbound token teaches the daemon which session it belongs
to; every later mark carrying the same token binds immediately, because the binding is already known.

**Why it matters, concretely.** A batch containing an *unbound* token must be withheld from commit —
see the withholding decision below. Per mark, that is one withheld batch per stage, roughly 34 per
run. Per session it is **one, ever**: after the first bind the token is known and nothing is withheld
again.

It is also the shape the operator expected on being told what the token was for, which is worth
something on its own — a correlator that needs explaining is one that gets misused.

**Alternatives rejected:**

- *One nonce per mark.* Correct, and what this change was first built on. Every stage pays a withheld
  re-read for a fact the run established at its first mark. Strictly more work for strictly less
  clarity.
- *Reuse one token forever, across runs.* Two concurrent sessions could then carry the same token,
  which resolves to ambiguity and refuses — turning a permanent identifier into a permanent
  non-identifier. The token is per **run**, generated fresh.

**A reused token is not a nonce, and is not called one.** The flag is `-session-token`. Calling a
deliberately-repeated value a nonce would mislead every future reader about whether repetition is a
bug.

## Decision: the session token is a literal, never a shell substitution

**This is the detail the whole mechanism turns on.** The transcript records `tool_use.input.command`
— the string handed to the tool, *before* the shell runs. `-session-token "mf-$(date +%s)-$$"` is therefore
recorded verbatim, identically, by every session, and discriminates nothing.

<!-- verified: a sampled transcript entry on 2026-08-14 retained an unexpanded `~` in its recorded command, confirming the string is stored pre-expansion -->

The command emits a concrete token, once per run. The guard against regression is a test asserting
that no mark call in any skill carries a `$(`, a backtick or a `$VAR` inside its token argument.

## Decision: `/myflow-fast` reuses the chained commands' stage names

**Chosen.** `/myflow-fast`'s allowed set is the **union** of `/myflow-start`'s, `/myflow-do`'s and
`/myflow-finish`'s documented stages, composed in `names.go` and asserted by the existing
README-parsing test. A fast run records the same stage names as the equivalent
`start`→`do`→`finish` sequence, under `command = /myflow-fast`, so the two are directly comparable
stage by stage — which is the question the operator asked and the reason this decision exists.

**Alternatives rejected:**

- *List every name in `/myflow-fast`'s README row.* Duplicates the other three commands' vocabulary
  inside the one table that is canonical for it, and drifts the first time any of them changes. This
  repository's standing answer to that tension is to cite rather than copy.
- *Make validation global.* Discards a real check — per-command validation is what stops a command
  inventing a stage name — to solve a problem composition already solves.

## Decision: a declared key, and a name that may change

**Chosen.** Every stage has a **key** — short, lowercase, dotted, unique across the whole pipeline —
and a **name**, which is prose for humans. The key is what a mark passes, what the store groups by,
and what never changes. The name is what the interface shows and may be reworded freely.

Keys are namespaced by the command that *defines* the stage, not by the command that runs it:
`do.review-panel` is the review panel wherever it runs, so `/myflow-fast` reuses it rather than
minting `fast.review-panel`. That is what makes a fast run comparable, stage for stage, against a
`start`→`do`→`finish` run — the question the operator asked, answered in the data model rather than
in a join.

**Alternatives rejected:**

- *Derive the key by slugifying the name.* Defeats the purpose. The reason to have a key is that
  improving a name must not split a stage's history, and a derived key changes with the name.
- *Keep the prose name as the identity.* The status quo. The identifier then contains backticks,
  apostrophes and markdown emphasis, has to be escaped at every call site, and any wording fix is a
  silent data migration.
- *Numeric ids.* Stable and unreadable. A key appears in shell invocations, guard output and the
  dashboard's group-by; `do.review-panel` is self-describing where `17` is not.

**The table becomes the source of both.** `README.md`'s Level 1 entry stops being arrow-separated
prose and becomes a real table — key, name, and the commands that run it — which is also what makes
the `/myflow-fast` union expressible rather than a paragraph. `names.go` parses that table and its
existing README-parsing test governs it, so the vocabulary stays single-sourced.

**Uniqueness is enforced, not assumed.** A duplicate key is a defect the parsing test fails on: two
stages sharing a key would merge two different things in every statistic silently.

## Decision: binding is bounded, and unbinding never happens

A nonce unresolved after a bounded window is abandoned; the run stays recorded and unattributed. On a
harness that produces no transcript at all, every mark takes that path exactly once, cheaply.

Binding is **one-way**. A bound stage run is never re-bound, so the exactly-once harvest mechanism
still sees each message attributed to at most one run.

## Decision: the third arm of the absence distinction

The interface today separates *not recorded* from *recorded as zero*. It needs a third: *recorded,
not measured*. Without it a recording misconfiguration is indistinguishable from a quiet period —
which is precisely how this defect went unnoticed while the dashboards merely looked empty.

## What is deliberately not changed

The harvest offsets and their exactly-once semantics; the pricing path; the never-block guarantee;
the three pipeline states; and every already-recorded stage run — **no backfill**, because the only
unattributed runs are from KAN-16's own development and are worth nothing.
