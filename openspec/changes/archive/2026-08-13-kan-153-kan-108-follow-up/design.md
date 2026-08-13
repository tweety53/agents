# KAN-153 — KAN-108 follow-up

**Date:** 2026-08-13
**Jira:** KAN-153
**Change:** `kan-153-kan-108-follow-up`

## Context

KAN-108 closed two structural cost problems in `/myflow-do` and left six items behind, recorded on
KAN-153 at the operator's direction. Four are findings its review panel raised and the operator
chose not to fix in that round (F25, F43, F46, F54); two are observations from the same run that
never became findings.

The items are independent of one another. What they share is a single theme: in every case a guard
existed but could not see the defect class, either because the logic it needed lived in a sibling
script that had drifted, or because no guard read that kind of content at all.

### The six items, and what the repository actually holds today

**F25 — duplicated marker machinery.** `scripts/check-panel-reproducers.sh` reimplements
`scripts/check-unfinished-work.sh`'s `count_matching` and `ids_of`. The duplication is real, and the
drift is measurable: `check-unfinished-work.sh:347` matches `findings-total: (0|[1-9][0-9]*)` with
no bound on the digit run, while `check-panel-reproducers.sh` bounds its equivalent pattern at
`{0,14}` — one record format, two guards disagreeing on which totals are well-formed.

**The ticket's crash claim is withdrawn, established by running the unpatched guard.** KAN-153
describes the 26-digit arithmetic crash as still reachable here. It is not: this guard compares the
declared total to the marker count with a string `!=`, and a 26-digit total produces an ordinary
`OUTSTANDING:` verdict at exit 0. The divergence is the whole defect, and it is enough.

The two helper sets are **not** identical, which the ticket's "near-verbatim" understates and which
the extraction has to account for:

| Helper | `check-unfinished-work.sh` | `check-panel-reproducers.sh` |
|--------|----------------------------|------------------------------|
| `count_matching` | takes `<file> <ere>` | takes `<ere>`, reads a `$PANEL` global; adds `--` before the path |
| `ids_of` | returns bare digits, `sort` **with duplicates kept** — `repeated_ids` depends on that | returns `F<n>`, `sort -u` |
| `grep_lines_of` | absent | present, with `match`/`line` modes |

**F43 — a broken Markdown code span.** The panel recorded it at `skills/myflow-do/SKILL.md:537`. A
later fix round of the same KAN-108 run reworded that passage, so the span at *that location* is
gone.

**A live instance of the class does survive elsewhere, and an earlier claim in this document that
none did was wrong.** The first sweep behind that claim truncated its own output and missed
`openspec/specs/myflow-review-panel-economics/spec.md:504`, which carries the defect exactly as F43
describes it: a single-backtick span holding a backslash-escaped backtick, so the span closes early
and half the banned-character list renders as literal text. It was found while restating that same
requirement into this change's delta spec, which inherited the broken line verbatim. Both copies are
repaired with double-backtick delimiters, where a literal backtick needs no escape.

That instance sits in `openspec/`, outside the scope chosen for the new guard, which is why the
guard does not see it — recorded here because it is the sharpest available argument about that
scope, not as a defect in the guard.

**F46 — the materiality clause.** `skills/myflow-do/SKILL.md:574` reads that a fix whose diff
"touches only the reproducer's own target" is not a fix. Taken literally that disqualifies the
ordinary guard-script case, where the reproducer's target *is* the path the finding named. The
capability spec is already unambiguous — its scenario reads "touches only the reproducer's own
target **and no path the finding named**" — so the defect is that the skill's prose dropped the
qualifier, not that the requirement is wrong.

**F54 — the double-fork survivor gap.** `scripts/run-reproducer.sh` walks the descendant tree
breadth-first each poll and kills deepest-first, which catches a detached child whose intermediate
lives long enough to be observed. A grandchild whose intermediate exits within milliseconds
re-parents to launchd before the first poll, and no parentage-based technique can see it afterwards.
`ps -o sid=` and `ps -o sess=` are both unusable on Darwin, leaving `pgrep -P` as the only tool the
current implementation has.

**The escalation ladder's size clause.** KAN-108 narrowed one vacuous trigger out of five. Two fix
rounds during that change escalated to Full on `the fix diff exceeds approximately 150 changed
lines` instead, at seven slots each. Size alone carries no risk signal — a mechanical rename is
large and harmless — so the clause now selects nearly as indiscriminately as the one that was
narrowed.

**Torn prose from a moving edit.** A trim during KAN-108 moved reasoning into
`skills/myflow-do/SKILL-rationale.md` and tore sentences apart: a paragraph promising a reason that
never arrived, a sentence ending mid-clause, a blockquote whose first line lost its marker.
`scripts/check-references.sh` passed throughout, because it verifies that headings resolve, not that
sentences are whole. A reviewer found the damage two panel passes later.

## Goals

- Close the *class* behind F25, not the instance: one definition of the marker helpers, and the
  unbounded digit run bounded.
- Give the repository a guard that reads its own Markdown for structural damage, closing both F43's
  class and the torn-prose class.
- Make the escalation ladder's size clause discriminate.
- Close F54's residual gap with a mechanism that does not depend on parentage.
- Align the materiality prose with the requirement it already has.

## Non-goals

- Rewriting `run-reproducer.sh`, which is already long. The process-group work is an addition at
  the exec point and the cleanup path, not a restructure.
- Prose linting in general. The new guard checks Markdown *structure*, not style, grammar, or
  wording.
- Any change to the zero-open-findings handoff bar, to roster presets, or to what a Full re-run
  costs once it fires.

## Decisions

### D1 — One sourced library for the marker helpers

**ID:** shared-panel-record-lib
**Status:** active
**Chosen:** a new `scripts/lib/panel-record.sh`, sourced by both guards, with the callers adapted to
its signatures — because the drift already happened once and only a single definition prevents the
next one.
**Considered:**
- *Share `count_matching` only, leave the two `ids_of` bodies separate* — honest about the semantic
  difference, but leaves half the drift surface intact, which is the outcome the ticket names.
- *Bound the regex and record why the helpers legitimately differ* — cheapest, but closes the
  instance and not the class.

The library exports:

```bash unverified:signatures proposed by this design; the library does not exist yet
count_matching <file> <ere>            # line count; grep -a; rc>1 propagates
grep_lines_of <file> <mode> <ere>      # mode: match | line
ids_of <file> <ere> <shape>            # shape: digits | ids-unique
```

`digits` reproduces `check-unfinished-work.sh`'s existing behaviour exactly — bare digits, `sort`,
duplicates retained, which `repeated_ids` reads. `ids-unique` reproduces
`check-panel-reproducers.sh`'s — `F<n>`, `sort -u`. Neither caller's observable behaviour changes.

Every body keeps the disciplines stated once in the library header rather than in each guard: `-a`
on every `grep` so a stray NUL byte cannot turn a file into a silent "no match"; the `rc > 1` split
that distinguishes grep's "no match" (an answer) from a real error (a refusal); and `--` before
every path so a record path beginning with `-` is never read as an option.

The guards resolve the library from `${BASH_SOURCE[0]}`, so a guard invoked from any working
directory finds it, and a missing library is a refusal (exit 2), never a guard that silently checks
less.

### D2 — One Markdown-integrity guard, in Python

**ID:** markdown-integrity-guard
**Status:** active
**Chosen:** `scripts/check-markdown-integrity.py`, Python 3 standard library only, following the
precedent `scripts/check-plan-provenance.py` set — a real block-structure parser rather than a
hand-rolled Bash regex allowlist.
**Considered:**
- *Bash* — stays inside the repository's mostly-Bash toolchain, but a fence-and-container parser in
  Bash is precisely what `check-plan-provenance` was rewritten out of, after five review passes and
  seven fix waves.
- *Two separate guards* — a span check and a prose check, each with its own harness and exit-code
  contract. Rejected: they read the same files with the same parser, and splitting doubles the
  block-classification code, which is the part that is hard to get right.

**Scope:** `skills/**/*.md` and `rules/*.mdc` — the files every `/myflow-*` run loads, and where a
torn sentence changes an agent's behaviour rather than merely reading badly.

**Signals.** Each reports `file:line: <signal>: <what is wrong>`:

1. *broken code span* — a single-backtick span containing a backslash-escaped backtick (backslash is
   not an escape inside a code span under CommonMark, so the span closes early), or an odd backtick
   count on a line outside a fence.
2. *orphaned blockquote body* — a `> ` line whose immediately preceding non-blank line is un-marked
   prose that does not end a sentence, i.e. a blockquote whose first line lost its marker.
3. *unterminated paragraph* — a paragraph, outside fences, tables, headings and list items, whose
   last line ends with none of `.` `:` `?` `!` `)` `|` `"` `'` or a closing backtick.
4. *dangling promise* — a paragraph that is the **last block before a heading or end of file**,
   whose final sentence **ends in a colon**, with no list, fence, table or paragraph following. A
   colon promises what comes next; nothing following is a broken promise, mechanically.

**Exit codes:** 0 clean, 1 violations found (each with its line), 2 cannot answer at all — an
unreadable file, a bad argument, a scope root that does not exist. It takes an optional project root
and defaults to this repository, so it runs against a bare tree and joins `## lint`.

### D3 — Signal 4 is deliberately narrow

**ID:** dangling-promise-narrow
**Status:** active
**Chosen:** the colon-and-nothing-follows rule above.
**Considered:** *a phrase-based rule* keying on "see X for why", "the reason is", "because:". Ruled
out: this repository contains hundreds of legitimate `See **X** (`file.md`) for why …` citations
that correctly end a section, so a phrase rule would fire constantly, and the lint policy forbids
adding suppression markers to quiet it. The operator selected this signal after the false-positive
risk was stated; the narrow rule is how it is delivered without a suppression story.

### D4 — Process group via a `python3` exec shim

**ID:** reproducer-process-group
**Status:** active
**Chosen:** wrap the reproducer's exec in
`python3 -c 'import os,sys; os.setsid(); os.execvp(sys.argv[1], sys.argv[1:])' <path> <args…>`. The argv
vector passes through untouched, no shell ever sees the line, and the reproducer and every
descendant share a process group that survives re-parenting.
**Considered:**
- *`bash set -m` job control* — pure Bash, but monitor mode in a non-interactive script is fragile
  and interacts with the existing poll loop.
- *A pre/post `ps` snapshot diff* — works without parentage, but is racy on a busy machine and can
  attribute an unrelated process to the reproducer.
- *`perl -e 'setpgrp(0,0); exec @ARGV'`* — same shape, but adds a third language to the toolchain
  where Python is already established here.

**Two things this sketch got wrong, each established by running it during implementation.**
`python3 -c` does not strip a `--` separator — with one, `sys.argv` becomes `['-c', '--', …]` and
`os.execvp` execs a file named `--` — so no separator is passed; the resolved path is always
absolute and can never be read as a flag. And the exec point sat inside a `set -m` backgrounded
subshell, which makes that subshell a process group leader, and `os.setsid()` raises `EPERM` for a
caller that is already a leader. `set -m` is removed, so the background job inherits its parent's
group, is not a leader, and the call succeeds.

Darwin ships no `setsid` binary, which is why a shim is needed at all. `python3` is already a
declared dependency of this repository (`/usr/bin/python3`, standard library only) for
`check-plan-provenance.py`, so this widens nothing.

Detection gains a `pgrep -g <pgid>` pass beside the existing breadth-first descendant walk; cleanup
kills the group (`kill -- -<pgid>`) first and keeps the deepest-first walk as a backstop. A missing
`python3` is a refusal — exit 2, recorded unverifiable — never a silent fall back to the ungrouped
exec, since an ungrouped exec is exactly the condition the gap lives in.

**What remains open after this:** a reproducer that calls `setsid` itself leaves the group
deliberately. That is documented beside the code as the residual limit, replacing the parentage
limit it supersedes.

### D5 — Size becomes an amplifier, not a trigger

**ID:** escalation-size-amplifier
**Status:** active
**Chosen:** `the fix diff exceeds approximately 150 changed lines` stops being an independent
trigger. It fires only in conjunction with a risk signal: a new file, a delta spec, a migration, a
guard's behaviour, or a file outside the set named in the findings.
**Considered:**
- *Raise the number to a measured value* — a one-line edit, but a large mechanical fix still forces
  seven slots, and no number distinguishes a rename from a rewrite.
- *Escalate breadth proportionally* — add the conditional slots whose own triggers fire rather than
  going to Full. Attractive, and close to what the pipeline already does elsewhere, but it changes
  what escalation *means* rather than fixing which fixes reach it.
- *Cap what a Full re-run costs* — attacks cost rather than the trigger; KAN-109 already collected
  most of that saving.

The other four triggers are unchanged. The requirement that every trigger discriminate, added by
KAN-108, is what this change is applying to its own remaining clause.

**The ladder fired on this change's own fix round, which is the first evidence about it.** Fix round
1 repaired a guard's behaviour, so the `a guard's behaviour` clause fired and required escalation to
Full without asking. The operator overrode it and took a Targeted re-run of the two slots that raised
the findings. Two things are worth recording for whoever tunes this next. First, the clause fired
correctly and on a real signal — a guard's behaviour genuinely changed. Second, it fired on a
**158-line fix in a change whose entire subject is guards**, which is the shape KAN-108 found the
`public contract` clause failing on: in a repository whose product is guards, `altered a guard's
behaviour` may discriminate no better there than `altered a public contract` did. One data point is
not a finding, and this change does not narrow that clause on the strength of it. But the next person
to look at the ladder should know it fired here, that the operator judged Full disproportionate, and
that the Targeted re-run found what it needed to.

**The size clause did not fire**, and correctly so: 158 changed lines is under the threshold, and the
round added no new file. It is recorded because a clause that never fires is the failure mode this
change was written to remove, so the first occasion it stayed silent is worth one sentence.

### D6 — F43's recorded location is clear; a live instance elsewhere is repaired

**ID:** f43-verified-absent
**Status:** superseded by f43-instance-found

### D6b — F43's class is repaired where it actually survives

**ID:** f43-instance-found
**Status:** active
**Chosen:** repair `openspec/specs/myflow-review-panel-economics/spec.md:504` and the delta spec that
restated it, and close the class going forward with D2's signal 1.
**Considered:**
- *Leave it, since the recorded location is clean* — that was the earlier decision, and it rested on
  a sweep that silently truncated its own output. The instance is real and renders wrongly today.
- *Widen the guard to `openspec/` so it catches this one too* — a scope decision, deliberately left
  to the operator rather than taken here; the repair does not depend on it.

### D7 — F46 aligns the skill to the spec it already has

**ID:** materiality-align-to-spec
**Status:** active
**Chosen:** restore the qualifier the capability spec already carries — a fix fails materiality when
its diff touches only the reproducer's own target **and no path the finding named** — and add a
scenario pinning the ordinary case where the two are the same path.
**Considered:** *rewrite the materiality rule* — rejected; the requirement is right, only the skill's
restatement of it is lossy, and the ticket itself asks for a wording pass rather than a rewrite.

## Open questions

### Should the Markdown-integrity guard also scan `openspec/`?

**ID:** markdown-guard-openspec-scope
**Status:** open
**Why it is open:** the scope was chosen as `skills/` + `rules/` before it was known that a live
instance of F43's own defect class sits in `openspec/specs/myflow-review-panel-economics/spec.md`,
outside it. The operator was shown that evidence and chose to keep the scope and record the gap
rather than widen it in this change: the guard has been calibrated against `skills/` and `rules/`
alone, and plan prose is a body of text it has never been run against, so widening here would mean a
fresh round of false-positive tightening inside a change that has already had two.
**What it affects:** whether damage in a delta spec is caught while the change is in flight or only
once it lands in `openspec/specs/` at archive time — and, if widened, how much recalibration the
guard needs against plan prose. The two known instances are repaired either way; this is about the
next one.

## Risks and trade-offs

- **Extracting shared helpers can change guard behaviour silently** → both guards' existing harnesses
  must pass unchanged, and the extraction lands before any behaviour edit, so a harness failure
  during that task is unambiguously the extraction's fault.
- **Signal 3 (unterminated paragraph) has the widest false-positive surface** of the four → it is
  scoped out of fences, tables, headings and list items, and the harness carries a near-miss case per
  signal that must not fire. Any real hit is fixed by completing the sentence, never by narrowing the
  guard.
- **The process group changes how the reproducer is launched** → the existing containment,
  argv-vector and timeout behaviour must be re-verified by the existing harness cases, which do not
  change; only the survivor cases are added to.
- **`skills/myflow-do/SKILL.md` is under a contract budget** → both prose edits must leave
  `check-contract-budget.sh` green; the materiality edit adds a clause and the escalation edit
  restructures one, so the net growth is small, but it is checked rather than assumed.

## Testing

- `scripts/test-check-unfinished-work.sh` and `scripts/test-check-panel-reproducers.sh` pass
  unchanged after the extraction, plus a new case pinning the bounded `findings-total` digit run.
- `scripts/test-check-markdown-integrity.sh` — one firing case and one near-miss per signal, plus a
  mutation case per signal proving the harness catches that signal's removal.
- `scripts/test-run-reproducer.sh` — a fast double-fork case that returns "defect not demonstrated"
  today and must report a named surviving process after the change.
- `.myflow/project.md` gains the new guard under `## lint` and its harness under `## test`.
