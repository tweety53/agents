# Self-review — kan-153-kan-108-follow-up

**Operator rating: 3 — mixed.** The outcome is right and the classes are closed, but the path was
noisy: eight panel findings, four fix rounds, two rebase retargets, and repeated corrections to the
plan by its own author. That rating is the honest one, and this report is written to explain the
noise rather than to argue the rating down.

**No Jira follow-ups were filed.** The filing question was offered per finding and no issue was
selected, so nothing was created. The findings below are the record.

## What the change delivered

Six items KAN-108 left open, with the class closed behind each rather than the instance:

- one sourced library defining the panel-record marker helpers, with both guards rewired to it and
  the digit-bound pattern reduced to a single constant;
- `scripts/check-markdown-integrity.py`, a four-signal Markdown-structure guard with 25 harness
  cases, declared under both `## lint` and `## test`;
- a reproducer that runs in its own process group, so a fast double fork is detected and its
  survivor named rather than silently missed;
- the materiality clause aligned to the requirement it restates, and the escalation ladder's size
  clause narrowed to the one signal that is not already an independent trigger.

Every lint guard and all 20 harnesses pass. The panel reached zero open findings.

## Problems, and the fixes worth making

### 1. The ledger was not maintained, and its name is not the one the tooling looks for

Model policy requires each dispatch to record its model in the SDD ledger as it happens. This run
recorded every model in its own reporting and never wrote the file, so
`scripts/preserve-session-records.sh` found no `.superpowers/sdd/tasks/progress.md` to preserve. The
ledger was compiled at finish instead and labelled as such — a reconstructed audit trail is weaker
evidence than a live one, and mislabelling it would have been worse than the gap.

Compounding it: the file was written as `docs/superpowers/ledgers/2026-08-13-<name>-ledger.md`, while
`scripts/gather-self-review-context.sh` looks for `docs/superpowers/ledgers/<name>.md`. It reported
`skipped: … (absent)` for a ledger that exists. The panel record, preserved by the script itself,
was found — so the two conventions disagree, and the one a human writes by hand is the one that
misses.

**Fix:** have the ledger written by the same mechanism that preserves it, or make the gatherer
resolve the same dated form it already resolves for the panel record. A record the tooling cannot
find is close to a record that does not exist.

### 2. Plan fields that declared counts caused a weaker test

Three `**Baseline:**` fields declared test counts that later rounds moved. One did active harm: the
field said `after=87` assertions, the implementer treated that number as a constraint, and dropped
`assert_verdict` to hit it — leaving a case that would still have passed if the guard regressed from
an `OUTSTANDING:` verdict at exit 0 to a refusal at exit 2. The per-task review caught it.

**Fix, applied during this run:** those fields now cite the harness's own output instead of a value.
A count in a plan is a measurement; when it is a guess it is worse than absent, because the plan is
read as authority.

### 3. A design decision was made against a recorded decision nobody read

Task 1 extracted a sourced library into two guards. `check-unfinished-work.sh` already carried a
comment stating these guards are single-file by design and that a sourced helper would make one
"present but unrunnable" — and two other guards cite that same rationale to justify duplicating code.
The options put to the operator never mentioned it, because it had not been read.

The extraction turned out to be safe — `setup.sh` distributes skills, rules and commands but never
`scripts/`, so a guard and its library travel together — and the operator kept it and corrected the
comment. But the decision was made in ignorance and only survived scrutiny by luck.

**Fix:** before offering options that reverse an existing arrangement, grep the code being changed
for a recorded reason it is arranged that way. A comment saying THE COPY IS DELIBERATE is exactly the
artifact that search would have found.

### 4. A truncated sweep produced a confident wrong claim

The change reported that F43's defect class no longer existed anywhere, on the strength of a
repository-wide grep. That grep's output was piped through `head -30` and the matching line was below
the cut. The instance was real, in `openspec/specs/myflow-review-panel-economics/spec.md`, and was
found later only because the delta spec restated the broken line verbatim.

**Fix:** a search whose result is used as evidence of *absence* must not be truncated. `head` on a
search for "does this exist anywhere" inverts what the command proves.

### 5. Two of the ticket's own premises were false

The ticket described a 26-digit arithmetic crash still reachable in `check-unfinished-work.sh`; the
comparison there is a string `!=` and no crash exists. It also placed F43 at a line that a later
round of KAN-108 had already reworded. Both were established by running the code rather than reading
the report, and both are recorded in the artifacts with the correction rather than quietly worked
around.

This is not a defect in the run — catching it is the run working — but it is worth stating that a
follow-up ticket written from a panel record inherits that record's errors, and its claims are
hypotheses.

## Cost

Roughly 1.2M subagent tokens across 16 dispatches: 5 implementers, 6 per-task reviewers, 3 panel
slots in pass 1 and 2 in pass 2, plus 4 fix rounds. The `light` roster and Sonnet everywhere kept
the panel cheap; the expensive part was rework, not review.

The largest single avoidable cost was the Markdown guard's calibration: 104 false positives on first
run against `skills/` and `rules/`, seven rules tightened, then two further fix rounds when the
tightenings proved too broad and then too narrow. That is inherent to writing a prose guard against
an existing corpus, but starting from a smaller signal set — the two mechanical signals, deferring
the two prose ones — would have separated "does the parser work" from "is this rule right".

## What went well

- **The review layers caught what they exist to catch.** The per-task review found a test weakened to
  satisfy a plan number. The panel found an exemption tightened until it swallowed the defect it
  guarded, a forbidden suppression form, and a comment contradicting the code beside it. None of these
  would have been visible from the diff alone.
- **Subagents disputed the plan instead of following it.** Task 6 reported that `python3 -c` does not
  strip `--` and that `os.setsid()` fails under `set -m`, each verified by experiment before writing
  code. Task 2's implementer refused the ticket's crash narrative after testing it. A dispatch that
  reports "your plan is wrong, here is what I ran" is worth more than one that complies.
- **The new guard found real damage on its first run**, which is the strongest available argument that
  it should exist.
- **Fixes that ran wider than their finding said so.** The process-group fix needed a narrowed kill
  scope beyond the reported reorder, and the agent flagged the widening rather than folding it in
  silently.

## Automation candidates

1. **A guard for baseline fields that state counts.** `check-task-commit-fields.py` already parses
   `**Baseline:**`. It could reject a field whose count is a bare number with no command beside it,
   which would have prevented finding 2 mechanically.
2. **Extend the Markdown-integrity guard to `openspec/`** — this change's recorded open question.
   Delta specs are where prose damage is introduced; the canonical file is only where it lands.
3. **A search-truncation lint for the agent's own habits** is not automatable, but a convention is:
   when a grep's result is evidence of absence, pipe it to `wc -l` alongside, so a truncated listing
   cannot be mistaken for a complete one.
