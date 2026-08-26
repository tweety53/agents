# Collapse check-self-review-report.sh's awk→bash boundary

## Why

`scripts/check-self-review-report.sh` classifies a report in one `awk` pass, serializes the result
as `US`-delimited records, and re-parses those records in `bash`. Three of KAN-200's findings — G2,
G3 and H2 — were taxes on that one seam, each a manifestation of the same fact: a value has to
survive a hand-written encoding across two languages with different escaping and splitting rules.
H2 reopened the class one fix round after G3 supposedly closed it, which is why the Principles slot
filed this as an observation rather than patching a fourth manifestation.

To touch this file safely a maintainer currently has to hold three unrelated rules at once: bash's
quote-removal behaviour for a regex used after `[[ =~ ]]`, awk's `-v` escape-stripping, and the fact
that record fields must avoid `\037`, `\t` and `\r`. Only the first is about self-review reports;
the other two are incidental to the seam.

Filed as KAN-211 from KAN-200's self-review, `myflow-automation` angle, and explicitly ruled out of
scope for that change's third fix round so it would not become a rewrite.

## What changes

- **`scan_report()` and the record protocol are deleted.** The per-file loop reads the report
  directly with `while IFS= read -r line`, classifying each line at the point it is read. The four
  record kinds the protocol encoded become four branches in that loop.
- **`FINDING_SHAPE_RE_AWK` is deleted** — the backslash-doubled relay of `FINDING_SHAPE_RE` through
  awk's `-v`. One pattern variable, one language, referenced twice.
- **The `CTRLBYTE` rule is retired.** `0x1F` was the field delimiter; with no protocol it is
  ordinary report content.
- **The header's seam documentation is replaced** by a short record of why the boundary existed,
  which three defects it produced, and that it was collapsed rather than patched — so a future
  reader does not reintroduce it.

**Not changing: what the guard checks.** Every rule, every message string and all three exit codes
are preserved, with the single exception of the retired `0x1F` violation. The harness keeps every
case and every fixture; one assertion string moves, recorded under `retire-ctrlbyte-rule` in
`design.md`.
