# Design: collapse the awk→bash boundary

## Decisions

### The guard becomes all-bash; `awk` is deleted rather than promoted

**ID:** all-bash-target
**Status:** active
**Chosen:** move the line walk into the existing bash loop — no `awk`, no record protocol, no
cross-language value passing at all. The classification bodies already live in bash and are correct;
`scan_report`'s `awk` was only walking lines and re-encoding them for bash to decode again.

```bash unverified:confirm against the implemented loop in scripts/check-self-review-report.sh
lineno=0; cur=""
while IFS= read -r line || [[ -n "$line" ]]; do
  lineno=$((lineno + 1))
  line="${line%$'\r'}"
  # ## heading with a label -> cur=<label>; any other #+ heading -> cur=""
  # body under a recognized section -> classify; finding-shaped before any
  # heading -> orphan violation
done < "$f"
```

**Considered:**

- *One `awk` pass emitting finished verdict lines, bash only recording coverage.* Rejected: POSIX
  awk has no capture groups and macOS ships the one-true-awk rather than gawk, so extracting a
  finding's label and text from `FINDING_LINE_RE` becomes hand-rolled `index`/`substr` surgery. It
  trades an encoding seam for string surgery in the weakest of the three languages.
- *Python classifier plus a thin `.sh` wrapper*, matching `check-plan-provenance` and
  `check-installed-citations`. Rejected despite being the closest repository precedent: it removes
  all three maintainer rules, but it **reintroduces a two-language crossing** — its own
  tab-delimited `member`/`count` sentinel lines — which is the same bug class one field-type away.
  KAN-211 names the class as "a value has to survive a hand-written encoding across two languages";
  the all-bash target removes the crossing outright, and a narrower crossing is not the same as
  none. It also adds a `python3` availability probe and a second file for a guard whose logic is
  already written and already correct in bash.

The bash-only target keeps `scripts/lib/coverage.sh` sourced in-process — no marshalling of
`(member, count)` across any boundary — which is the concrete reason bash beats Python here
specifically, whatever the merits elsewhere in this repository.

### The `0x1F` rule is retired with the protocol that created it

**ID:** retire-ctrlbyte-rule
**Status:** active
**Chosen:** delete the `CTRLBYTE` record and its violation. `0x1F` was the record protocol's field
delimiter; with no protocol, a `0x1F` byte in a report is ordinary content and there is nothing to
reject. Harness case twenty-one keeps its fixture and its exit-1 assertion; its message assertion
becomes `neither a finding line nor the none-marker`, because a none-marker line carrying a trailing
`0x1F` no longer equals the none-marker string and its section is flagged for that reason instead.
The case keeps its real purpose — proving a trailing control byte does not silently forgive a
non-compliant line — and joins cases nineteen and twenty as byte-fidelity tests.

**Considered:** reimplementing the check in bash as a standalone content prohibition, so the case
stays byte-for-byte green. Rejected: it costs nothing at runtime but keeps a rule whose only
justification has been removed, and leaves a maintainer still holding "report content may not
contain `0x1F`" — one of the three incidental rules this change exists to delete.

This is a deliberate, recorded deviation from KAN-211's "any implementation must keep every existing
case green": the exit code stays green, one assertion string does not.

### Retiring the `0x1F` rule needs its own regression assertion

**ID:** ctrlbyte-retirement-needs-its-own-assertion
**Status:** active
**Chosen:** case 21 additionally asserts that `unsupported control byte` is **absent** from the
guard's output. Found during task 1's RED step and recorded here rather than left in a commit
message: retargeting case 21's message assertion alone does not detect a revert, because the
pre-change guard reports the section-level violation **as well as** the control-byte one — `CTRLBYTE`
consumed the none-marker line, so that section failed the post-loop neither-marker-nor-finding check
too. The retargeted assertion therefore passes in both states. Asserting the retired violation's
absence is the only thing in the harness that makes the revert loud.

**Considered:** accepting the gap, on the grounds that `retire-ctrlbyte-rule` is an intended
behaviour change and `structural-proof-not-new-cases` had already declined new cases. Rejected: that
decision declines cases for G2/G3/H2, which have no protocol left to revert to. This retirement
*is* revertible, and a harness that cannot tell the two states apart is the "checked nothing" shape
`scripts/lib/coverage.sh` exists across this repository to make loud.

This supersedes nothing; it fills a gap `structural-proof-not-new-cases` does not cover. It moves
the plan's recorded baseline from forty-three assertions to forty-four.

### G2, G3 and H2 are proved gone structurally, not by new cases

**ID:** structural-proof-not-new-cases
**Status:** active
**Chosen:** argue their absence from the code's shape and add no new harness cases for them. Cases
nineteen, twenty and twenty-one exist as regressions against a *revert of the protocol*; with no
protocol there is nothing to revert to, so a case guarding that revert cannot be written. What the
three fixtures still prove — that a tab or a control byte in report content reaches the classifier
byte for byte — remains exactly what they assert.

**Considered:** adding cases that inject each byte at every field position. Rejected: field
positions no longer exist, so such cases would test nothing and would read to a future maintainer as
if a protocol were still present.

The mutation proof KAN-211 credits the harness with is re-established the same way it was earned:
each classification branch is reverted one at a time and the case covering it must fail.

### The two branches the mutation proof found uncovered get cases here

**ID:** cover-the-two-uncovered-branches
**Status:** active
**Chosen:** add a harness case for each of the two branches task 2's mutation proof measured as
uncovered — the arm reporting a disposition that is neither `filed: <KEY>` nor `declined`, and the
loose-shape arm reporting a finding-shaped line as malformed. Both are live, reachable code, proved
so with probe fixtures rather than inferred; deleting either leaves all forty-four assertions green.
The second is KAN-200's own G2 fix, added precisely so a malformed finding-shaped line would stop
being silently dropped — it has been untested since.

**This falsifies a premise of KAN-211**, which states the harness is "correct and mutation-proved."
It is correct; it was not fully mutation-proved, and nothing had measured that before task 2.
Recorded here rather than quietly fixed, because the ticket's own framing rests on it.

**Considered:** filing a Jira follow-up and leaving the matrix note alone — rejected on the
operator's decision: the gap is two small fixtures, the branches are already written and already
correct, and covering them changes nothing about *what* the guard checks, so it stays inside
KAN-211's "not proposed: changing what the guard checks."

**This does not supersede `structural-proof-not-new-cases`.** That decision declines new cases for
G2, G3 and H2 — protocol regressions with no protocol left to revert to, so no case can be written
for them. These two branches are ordinary classification arms that a fixture can reach directly, and
that decision never covered them.

**A trap worth writing down for whoever implements it:** the loose-shape case's malformed line
leaves its section with no recognized finding and no none-marker, so the post-loop
neither-marker-nor-finding check fires on the same fixture and holds the exit code at 1 even when
the loose-shape branch is deleted. The case must therefore assert the malformed-line message itself,
never the exit code alone.

## Known limits

**Bash `read` silently drops `NUL` bytes**; `awk`'s handling of them was implementation-defined.
Neither is exercised by the harness and no behaviour depends on either, so this is recorded as a
limit of the new shape rather than treated as a regression.

**Bash's `[[ =~ ]]` quote-removal hazard stays.** It is why `FINDING_SHAPE_RE` and its siblings are
kept in variables and referenced unquoted, and that comment stays in the header. It is a real,
independent bash quirk, not part of the seam — of the three rules KAN-211 says a maintainer must
hold, this change removes two and leaves this one.

**A read that fails partway through a file it opened successfully is not detected.** The
pre-collapse guard ran each report through `scan_report "$f"`, whose `awk` reported a non-zero exit
status for *any* failure of its own input — a failed open and a mid-stream read error alike — and
the caller died on it. The all-bash loop's exit status covers only the first of those two.

A failed `< "$f"` redirection aborts the compound command before the body runs once, so the `while`
exits 1 and `read_rc` catches it; that is the whole of what the post-loop check closes. A read that
fails *after* a successful open cannot reach it: `read` returns non-zero for an I/O error and for
end-of-file indistinguishably, the loop's `|| [[ -n "$line" ]]` guard consumes whatever partial line
was buffered, the loop then terminates through its normal condition, and the `while` exits with the
status of the last command its body ran — 0, deliberately, since a deterministic body status is what
makes the failed-open signal readable at all. Measured, not reasoned: reading from a file descriptor
whose `read` errors immediately but whose `open()` succeeded — a directory, the one such target
available in this environment — the exact loop condition this guard uses exits 0 on bash 3.2.57 and
on bash 5.3.15, with `read: read error: Is a directory` on stderr and no non-zero status anywhere.

The exposure is narrow. `find -maxdepth 1 -type f` enumerates only regular files, identically before
and after the collapse, so directories, FIFOs, device nodes and broken symlinks never reach the loop
— what remains is a genuine hardware or filesystem I/O error on a regular file mid-read. Such a
report is recorded with whatever count its truncated content produced, so it fails as a content
violation at exit 1 rather than as the "cannot answer" exit 2 it deserves. No workaround was built:
every candidate costs a second process or a second read of every file to cover a failure mode that
cannot be simulated here and so could not be shown to work.

**Considered:** `exec 3<"$f"` … `done <&3` … `exec 3<&-`, the fixed-descriptor form, which is
bash-3.2-safe unlike the dynamic `exec {fd}<` one. Rejected: it is a genuine `open()` test, which is
strictly better than the `[[ -r "$f" ]]` precheck it would have replaced, but the post-loop status
already tests the same open with no descriptor to leak on an early `die` — and neither form detects
a mid-read failure, which is the limit recorded here.

## Baseline

The guard runs clean over the real corpus in under a fifth of a second, across twenty-nine reports.
<!-- measured: time scripts/check-self-review-report.sh @ branch main, before any edit -->

Whether reading those lines in bash rather than awk changes that materially is measured during
implementation, not predicted here.

## Open questions

None.
