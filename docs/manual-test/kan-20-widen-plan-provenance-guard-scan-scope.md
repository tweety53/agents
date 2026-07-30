# Manual test — kan-20-widen-plan-provenance-guard-scan-scope

**Worktree (all commands run from here):**
`/Users/tweety53/Projects/agents-worktrees/openspec-kan-20-widen-plan-provenance-guard-scan-scope`

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-20-widen-plan-provenance-guard-scan-scope
```

## There is no application to run

This repository declares no runnable application — it is the source of the myflow skills, installed
elsewhere by `setup.sh`. There is nothing to start, no port, and no URL. "Running the app" here
means running the guard and its harness, which is what the checklist below does.

## First: the automated evidence

```bash
./scripts/test-check-plan-provenance.sh
scripts/check-plan-provenance.sh
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: `All check-plan-provenance assertions passed`, then
`check-plan-provenance: 3 file(s) scanned, all provenance stated`, then two more clean guards.

**`3 file(s)` is itself the headline result.** Before this change it read `1 tasks.md file(s)`. The
three are this change's own `tasks.md`, `design.md` and `proposal.md` — the widening, observed on
the change that introduces it.

## Checklist — each item is a scenario from the delta spec

Every item below uses a scratch fixture so nothing touches the real repository. Create one once:

```bash
FIX="$(mktemp -d)"
mkdir -p "$FIX/openspec/changes/demo" "$FIX/openspec/changes/archive/old"
printf '```bash verified:by hand\necho hi\n```\n' > "$FIX/openspec/changes/demo/tasks.md"
run() { CHECK_PLAN_PROVENANCE_ROOT="$FIX" scripts/check-plan-provenance.sh; echo "exit=$?"; }
```

- [ ] **A change's `design.md` is scanned.** An untagged number there is now a violation.
  ```bash
  printf 'the suite reported 197 tests\n' > "$FIX/openspec/changes/demo/design.md"
  run   # expect: reports design.md:1, exit=1
  ```

- [ ] **A change's `proposal.md` is scanned.** Same rule, same result.
  ```bash
  rm "$FIX/openspec/changes/demo/design.md"
  printf 'the suite reported 197 tests\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports proposal.md:1, exit=1
  ```

- [ ] **A quoted number is reproduced, not claimed.** This is what lets the provenance
  documentation quote the historical invented baseline without an untruthful tag.
  ```bash
  printf 'a baseline of "194 tests" was invented, not measured\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: exit=0
  ```

- [ ] **A code span is NOT a quotation.** Double quotes — straight and curly — are the whole
  delimiter set. Inline code spans were a second delimiter class during implementation and were
  removed after failing open five times, so a number written inside backticks is an ordinary
  unattributed claim. The rationale is under **Why the delimiters are double quotes only** in
  `/Users/tweety53/Projects/agents-worktrees/openspec-kan-20-widen-plan-provenance-guard-scan-scope/skills/myflow-contracts/plan-provenance.md`.
  ```bash
  printf 'the offending line read `85 lines` with no tag\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1
  ```

- [ ] **An unmatched delimiter exempts nothing.** Fail-closed — and the guard now says so. Read the
  second line it prints: it names the veto that fired, the remedy, and the contract.
  ```bash
  printf 'it ran "and then reported 12 failures\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, plus a "quotation exemption was withdrawn
        #         ... (unbalanced quotation delimiters)" note; exit=1
  ```

- [ ] **A backslash-escaped delimiter vetoes the line.** CommonMark §2.4 makes `\"` a literal quote,
  so it opens and closes nothing. Note the veto reason names the escape specifically.
  ```bash
  printf '%s\n' 'Use \" for a literal quote; the run reported 99 tests, then printed \" done' \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, plus a "(a backslash-escaped quotation
        #         delimiter)" note; exit=1
  ```

- [ ] **The escape veto is whole-line.** A quotation that pairs perfectly still loses its exemption
  when an escaped delimiter appears elsewhere on the line. This is the accepted cost, not a bug.
  ```bash
  printf '%s\n' 'the baseline was "99 tests", and a literal quote is written \"' \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1
  ```

- [ ] **A delimiter inside raw HTML vetoes the line.** CommonMark §6.6 passes an HTML comment
  through untouched, so the quote marks below are comment text and `77 tests` stands in open prose.
  This is the reproducer that ended the code-span class.
  ```bash
  printf '%s\n' 'See <!--"--> the benchmark ran 77 tests <!--"-->.' \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, plus a "(a `<` character on the line)"
        #         note; exit=1
  ```

- [ ] **A `<` with no delimiter in it vetoes the line too — this is the behaviour CHANGE.** This item
  previously asserted exit=0, on the reasoning that the veto's trigger was a delimiter inside a
  construct. Deciding that requires knowing where the construct ENDS, and that grammar leaked six
  times, so the trigger is now the `<` itself. Every `<!-- measured: … -->` comment on a line that
  also quotes a number now withdraws that line's exemption. Measured cost across this repository:
  two claims, neither inside the guard's scan scope.
  ```bash
  printf '%s\n' 'the baseline was "194 tests" <!-- see the earlier plan -->' \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1
  ```

- [ ] **A `>` inside a single-quoted attribute value cannot hide a delimiter.** The sixth fail-open,
  pinned. §6.6 lets this tag run past its first `>`; the old per-construct regex stopped there, found
  no delimiter in the truncated prefix, cleared the line, and let the trailing `"` pair with the one
  inside the attribute — exit 0 on a bare `77 tests`.
  ```bash
  printf '%s\n' "<a b='>\"'> the benchmark ran 77 tests \"" \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1
  ```

- [ ] **A veto note names something you can see on your own line.** No quotation delimiter appears
  anywhere below, so the old note — "a quotation delimiter inside raw HTML" — sent you hunting for
  something provably absent. The note must name the `<`.
  ```bash
  printf '%s\n' 'the range a <b and c covers 5 tests' \
    > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, plus a "(a `<` character on the line)"
        #         note and "remove the `<` or reword the line"; exit=1
  ```

- [ ] **An apostrophe is not a quotation delimiter.**
  ```bash
  printf "the guard's own suite reported 12 failures\n" > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1
  ```

- [ ] **A bare claim beside a quoted one is still reported.** The scan looks at every match on the
  line, not just the first.
  ```bash
  printf 'we quoted "194 tests" but then asserted 12 failures\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports ONLY the 12 failures claim, exit=1
  ```

- [ ] **An issue key is not a quantity.**
  ```bash
  printf 'catches the KAN-6 errors in one pass\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: exit=0
  printf 'the suite reported 6 errors\n' > "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: reports the claim, exit=1   (the control — the rule is not toothless)
  ```

- [ ] **A change's `specs/` directory is NOT scanned.** Spec text legislates rather than describes.
  ```bash
  rm "$FIX/openspec/changes/demo/proposal.md"
  mkdir -p "$FIX/openspec/changes/demo/specs/cap"
  printf 'a baseline of 197 tests\n' > "$FIX/openspec/changes/demo/specs/cap/spec.md"
  run   # expect: exit=0
  ```

- [ ] **Archived changes are not scanned**, for the new file types too.
  ```bash
  printf 'the suite reported 197 tests\n' > "$FIX/openspec/changes/archive/old/design.md"
  run   # expect: exit=0
  ```

- [ ] **A change with a `design.md` but no `tasks.md` is scanned normally** — it must not trip the
  "broken glob" scan-integrity failure.
  ```bash
  rm -f "$FIX/openspec/changes/demo/tasks.md" "$FIX/openspec/changes/archive/old/design.md"
  printf '```bash verified:by hand\necho hi\n```\n' > "$FIX/openspec/changes/demo/design.md"
  run   # expect: "1 file(s) scanned", exit=0
  ```

- [ ] **Containment covers every scanned file, not just `tasks.md`.** A symlinked candidate is
  refused before it is opened. This is the security-relevant behaviour.
  ```bash
  ln -s /etc/hosts "$FIX/openspec/changes/demo/proposal.md"
  run   # expect: "proposal.md is a symlink — refusing to open it", exit=3
  ```

- [ ] **A containment refusal does not abandon the scan.** The refusal outranks the violation in the
  exit code, but the violation is still reported.
  ```bash
  printf 'the suite reported 197 tests\n' > "$FIX/openspec/changes/demo/design.md"
  run   # expect: BOTH the symlink refusal AND the design.md:1 violation printed; exit=3
  ```

Clean up when finished:

```bash
rm -rf "$FIX"
```

## What this change deliberately does NOT do

- [ ] **Confirm the guard still refuses to check whether provenance is *true*.** A tag naming a
  command that does not exist, or a ref that does not resolve, still passes. That boundary is
  unchanged and is documented under "What the guard does not do" in
  `/Users/tweety53/Projects/agents-worktrees/openspec-kan-20-widen-plan-provenance-guard-scan-scope/skills/myflow-contracts/plan-provenance.md`.
  Verifying the `@ <ref>` half is a separate, unfiled piece of work — and it is the one that would
  have caught the defect KAN-20 cites most often.

- [ ] **Read the three veto notes** in that same contract file — **The class-wide veto**, **The
  escape veto** and **The angle-bracket veto**. A genuinely closed quotation loses its exemption if a
  stray, escaped, or HTML-embedded delimiter appears anywhere on the line. Deliberate and
  fail-closed: the cost is a loud false positive you fix by rewording, in place of a silent false
  negative. Confirm the exit-1 banner offers rewording as a remedy — an author who tags a number
  that was already correctly quoted has written a `measured:` comment that is not true, which is
  worse than the finding it silenced.
