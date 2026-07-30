# Widen the plan-provenance guard's scan scope — Implementation Plan

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Scan a change's `design.md` and `proposal.md` alongside `tasks.md`, after removing the two
false-positive classes that widening would otherwise expose.

**Architecture:** Three ordered changes to one script. The classifier fixes land first so the
repository never passes through a state where the widened scan reports a false hit on its own tree;
the scope widening lands second; documentation last. The numeric rule already routes through a
single method (`_check_numeric_claim`), and the scan loop already routes every file through one
containment gate, so both fixes have exactly one call site each.

**Tech Stack:** Python 3 standard library only (`/usr/bin/python3`), POSIX shell for the test
harness.

## Global Constraints

- **Python 3, standard library only.** No third-party imports, no pip, no network. The guard runs
  under `/usr/bin/python3`.
- **No suppression markers, and no weakening of any guard's configuration.** This repository's lint
  policy is fix-first; a false hit is fixed by correcting the classifier, never by silencing a line.
- **No auto-fix command exists in this repository.** The Lint Fix Priority rule's auto-fix step is
  inapplicable here rather than skipped.
- **No per-task commits.** `pipeline.md`'s git boundaries give `/myflow-do` `git add` only (no
  commits) unless a `prUrl` is already recorded. This deliberately overrides the writing-plans
  template's per-task commit step — do not add commits.
- **All three lint guards are expected to exit 0** at the end of every task.
- **Every fenced block and numeric claim added to a planning artifact carries a provenance tag.**
  After Task 3 the guard enforces this on this change's own `design.md` and `proposal.md`.

## File structure

| File | Responsibility | Tasks |
|---|---|---|
| `scripts/check-plan-provenance.py` | The guard: classifier, scan set, containment, messages | 1, 2, 3, 4 |
| `scripts/test-check-plan-provenance.sh` | The harness; one case per behaviour | 1, 2, 3 |
| `skills/myflow-contracts/plan-provenance.md` | The canonical contract prose | 4 |

---

### Task 1: Stop reading an issue key as a quantity

**Files:**
- Modify: `scripts/check-plan-provenance.py:328-330` (`CLAIM_RE`)
- Modify: `scripts/check-plan-provenance.py:97-104` (module docstring, numeric-rule paragraph)
- Test: `scripts/test-check-plan-provenance.sh`

**Interfaces:**
- Consumes: nothing.
- Produces: `CLAIM_RE`, unchanged in name and in every other property. Tasks 2 and 3 rely on it
  staying a module-level compiled pattern whose matches expose `.start()` and `.end()`.

- [ ] **Step 1: Write the failing tests**

Append to `scripts/test-check-plan-provenance.sh`, after the numeric-rule section:

```bash verified:idiom copied from the existing cases at scripts/test-check-plan-provenance.sh:148-157
# An issue key followed by a unit word is not a numeric claim.
new_fixture
printf 'catches all six KAN-6 errors in one pass\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "issue key followed by a unit word is not a claim" \
  || fail "KAN-6 adjacency: rc=$RC out=$OUT"

# Control: the same unit word after a bare digit run IS a claim.
new_fixture
printf 'the suite reported 6 errors\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "a bare digit run before a unit word is still a claim" \
  || fail "bare claim control: rc=$RC out=$OUT"
```

- [ ] **Step 2: Run the harness and watch the first case fail**

Run: `./scripts/test-check-plan-provenance.sh`
Expected: `FAIL: KAN-6 adjacency: rc=1 …` — the guard currently reads a quantity out of the issue
key itself. The control case passes already.

(That sentence originally quoted the misparsed fragment, and this guard flagged it as an untagged
claim — the exemption Task 2 adds is what will make quoting it legal. Until then the number is
removed rather than tagged, exactly as the spec requires.)

- [ ] **Step 3: Add the lookbehind**

Replace the pattern at `scripts/check-plan-provenance.py:328-330`:

```python verified:compiled and run over openspec/changes/archive/*/{design,proposal}.md during planning
CLAIM_RE = re.compile(
    rf"(?<![A-Z]-)(?<!\w)[0-9](?:,?[0-9]){{0,20}}[\s]+(?:{UNITS})(?!\w)"
)
```

`(?<![A-Z]-)` is fixed-width, which Python's `re` requires of a lookbehind. It rejects a digit run
preceded by an uppercase letter and a hyphen, which is the issue-key shape, while leaving every
other boundary decision to the existing `(?<!\w)`.

- [ ] **Step 4: Correct the docstring paragraph that describes this exclusion**

The numeric-rule paragraph at `scripts/check-plan-provenance.py:97-104` claims the rule "deliberately
does NOT match … `IU-262`, `KAN-14`". That held only because those keys are not followed by a unit
word. Reword it to state what the code now does: a digit run preceded by an uppercase letter and a
hyphen is an issue key, not a quantity, and is excluded by the leading lookbehind.

- [ ] **Step 5: Run the harness and the lint guards**

Run: `./scripts/test-check-plan-provenance.sh && scripts/check-plan-provenance.sh && scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: `All check-plan-provenance assertions passed`, then all three guards exit 0.

---

### Task 2: A quoted number is reproduced, not claimed

**Files:**
- Modify: `scripts/check-plan-provenance.py` — add `_is_quoted` beside the module-level patterns;
  change the early return in `_check_numeric_claim` (`scripts/check-plan-provenance.py:1739-1740`)
- Test: `scripts/test-check-plan-provenance.sh`

**Interfaces:**
- Consumes: `CLAIM_RE` from Task 1.
- Produces: `_is_quoted(text: str, start: int, end: int) -> bool` — module-level, pure, no state.
  Task 3 does not call it; only `_check_numeric_claim` does.

> **Superseded by the pass-15 fix wave, recorded here rather than left to drift.** The predicate as
> planned recomputed the line's delimiter regions on every call, and `_check_numeric_claim` calls it
> once per `CLAIM_RE` match — quadratic in the number of code spans on one line, which is a denial
> of service on a guard that runs as a CI gate. The shipped split is
> `quotation_regions(text) -> Optional[list]`, computed once per line, and
> `_is_quoted(start, end, classes) -> bool`, which never reads the text. Both are still module-level
> and pure: the regions are threaded in as a parameter, never cached inside a function. The
> `Optional` return is the other half of the same wave — `None` means the line's code-span state is
> indeterminate and nothing on it may be exempted, which is a different fact from the empty list a
> line with no code spans returns. See `_code_span_regions` and `_region_at` in the guard.

> **Superseded again by the pass-17 fix wave — the code-span delimiter class is GONE.** Everything
> below that names inline code spans as a delimiter describes a design that shipped and was then
> withdrawn: the class failed open five times, all five in code-span handling, and every measured
> false positive the exemption exists for is double-quote delimited. The shipped delimiter set is
> matched double quotes only, straight and curly; a number inside backticks is an ordinary
> unattributed claim and is reported. `_code_span_regions` no longer exists, and the `Optional`
> return above is now a `QuotationScan` carrying the regions plus the reason — if any — that the
> exemption was withdrawn, which the guard prints so an author is not told to tag a number they had
> already quoted correctly. See design decision `quotation-quotes-only`, which supersedes
> `quotation-syntactic`, and the module docstring's "WHY THE QUOTATION EXEMPTION IS QUOTES-ONLY".
> The instruction text below is kept unchanged because it is the record of what was tried.

- [ ] **Step 1: Write the failing tests**

```bash verified:idiom copied from the existing cases at scripts/test-check-plan-provenance.sh:148-157
# A number inside straight double quotes is a quotation, not a claim.
new_fixture
printf 'a baseline of "194 tests" was invented, not measured\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "straight-quoted number is not a claim" \
  || fail "straight quotes: rc=$RC out=$OUT"

# The same holds for an inline code span.
new_fixture
printf 'the offending line read `85 lines` with no tag\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "code-span number is not a claim" \
  || fail "code span: rc=$RC out=$OUT"

# And for curly quotes.
new_fixture
printf 'quoting \342\200\234194 tests\342\200\235 from the earlier plan\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "curly-quoted number is not a claim" \
  || fail "curly quotes: rc=$RC out=$OUT"

# Fail closed: one unmatched delimiter creates no quotation.
new_fixture
printf 'it ran "and then reported 12 failures\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "unmatched delimiter does not exempt" \
  || fail "fail-closed: rc=$RC out=$OUT"

# An apostrophe is not a delimiter.
new_fixture
printf "the guard's own suite reported 12 failures\n" \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "apostrophe is not a quotation delimiter" \
  || fail "apostrophe: rc=$RC out=$OUT"

# A bare claim on the same line as a quoted one is still reported.
new_fixture
printf 'we quoted "194 tests" but then asserted 12 failures\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "a bare claim beside a quoted one is still reported" \
  || fail "mixed line: rc=$RC out=$OUT"
```

- [ ] **Step 2: Run the harness and watch the exemption cases fail**

Run: `./scripts/test-check-plan-provenance.sh`
Expected: the straight-quote, code-span and curly-quote cases FAIL with `rc=1`; the three
fail-closed cases already pass.

- [ ] **Step 3: Add the predicate**

Add beside the other module-level patterns, after `PROVENANCE_RE`:

> **Corrected during implementation — read this before writing code.** An earlier revision of this
> step carried a concrete implementation that counted delimiter *parity* across the line, and
> tagged it `verified:` on the strength of a run over the archived changes. That corpus run was
> real, and it was not proof: parity counting **fails open**. A stray unpaired delimiter earlier in
> the line, combined with an unrelated complete pair later, manufactures a bogus region and exempts
> a bare asserted number — “a "quote" and a stray " mark, then 99 tests ran, "done"” exempted
> “99 tests”. That is the exact defect the guard exists to catch, so the code block was withdrawn
> and replaced by the contract below. The lesson is the contract's own: `verified:` must name a
> check that actually settles the claim, and a friendly-corpus run does not settle correctness.

Add beside the other module-level patterns, after `PROVENANCE_RE`, a `_QUOTE_PAIRS` table and an
`_is_quoted(text, start, end)` predicate satisfying **all** of:

- **Delimiters:** matched straight double quotes, matched curly double quotes (`U+201C`/`U+201D`),
  and CommonMark inline code spans (backtick runs). Single quotes and apostrophes are excluded,
  because prose apostrophes are unpaired by nature.
- **Nearest enclosing pair, not parity.** Scan the line's delimiters in order and track open/close
  state, so the predicate answers *"is this offset inside a genuinely matched pair?"* An
  implementation that counts occurrences before the match is the withdrawn defect above.
- **Fail closed.** An unclosed opener at end of line exempts nothing after it. The predicate may
  only ever REMOVE a finding, and only where the number is demonstrably enclosed.
- **Line-scoped.** A pair spanning a line break creates no region, keeping this a pure predicate
  rather than parser state that would have to agree with the fence tracker.
- State the rule chosen for a code span inside a quoted region, and vice versa, in the report.

- [ ] **Step 4: Route the claim check through it**

Replace the early return at `scripts/check-plan-provenance.py:1739-1740`:

```python verified:the surrounding method body is quoted from scripts/check-plan-provenance.py:1739-1751 @ c1f5aa6
        claim = None
        for candidate in CLAIM_RE.finditer(text):
            if not _is_quoted(text, candidate.start(), candidate.end()):
                claim = candidate
                break
        if claim is None:
            return
```

The rest of the method — the two-line lookahead for `PROVENANCE_RE`, the message, the failure
counter — is unchanged. Scanning every match rather than only the first is what makes the mixed-line
case work: a quoted number no longer masks a bare one beside it.

- [ ] **Step 5: Run the harness and the lint guards**

Run: `./scripts/test-check-plan-provenance.sh && scripts/check-plan-provenance.sh && scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: `All check-plan-provenance assertions passed`, then all three guards exit 0.

---

### Task 3: Scan design.md and proposal.md

**Files:**
- Modify: `scripts/check-plan-provenance.py` — the scan loop at
  `scripts/check-plan-provenance.py:2380-2460`, and the containment gate then named
  `verify_tasks_md_containment` (`scripts/check-plan-provenance.py:1598`), now
  `verify_scanned_file_containment`
- Modify: `scripts/check-plan-provenance.py:88-95` (module docstring, scope paragraph)
- Test: `scripts/test-check-plan-provenance.sh` — invert case 8 (`:148-157`); update the three
  success-line assertions at `:1017`, `:1525`, `:2091`

**Interfaces:**
- Consumes: `CLAIM_RE` (Task 1), `_is_quoted` via `_check_numeric_claim` (Task 2).
- Produces: `SCANNED_FILENAMES: tuple[str, ...]` and
  `verify_scanned_file_containment(self, path: str, changes_dir: str, filename: str) -> bool`,
  replacing `verify_tasks_md_containment`. Task 4 documents both.

- [ ] **Step 1: Invert the test that asserts the old scope**

Case 8 at `scripts/test-check-plan-provenance.sh:148-157` currently asserts
`proposal.md is out of scope`. Replace its assertion and label:

```bash verified:the fixture setup is the existing case 8, quoted from scripts/test-check-plan-provenance.sh:148-157
# 8. An untagged block in proposal.md IS in scope.
new_fixture
clean_tasks_md
{
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/proposal.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "proposal.md is in scope" || fail "proposal.md: rc=$RC out=$OUT"
```

- [ ] **Step 2: Add the remaining scope cases**

```bash verified:idiom copied from the existing cases at scripts/test-check-plan-provenance.sh:148-157
# An untagged numeric claim in design.md is in scope.
new_fixture
clean_tasks_md
printf 'the suite reported 197 tests\n' \
  > "$FIXTURE/openspec/changes/demo-change/design.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "design.md numeric claim is in scope" \
  || fail "design.md: rc=$RC out=$OUT"

# A change's specs/ directory is NOT scanned.
new_fixture
clean_tasks_md
mkdir -p "$FIXTURE/openspec/changes/demo-change/specs/some-capability"
printf 'a baseline of 197 tests\n' \
  > "$FIXTURE/openspec/changes/demo-change/specs/some-capability/spec.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "change specs/ is not scanned" || fail "specs/: rc=$RC out=$OUT"

# An archived change's design.md is not scanned.
new_fixture
clean_tasks_md
printf 'the suite reported 197 tests\n' \
  > "$FIXTURE/openspec/changes/archive/old-change/design.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "archived design.md is not scanned" \
  || fail "archived design.md: rc=$RC out=$OUT"

# A change directory with design.md but no tasks.md is scanned, not a scan-integrity failure.
new_fixture
rm -f "$FIXTURE/openspec/changes/demo-change/tasks.md"
printf '```bash verified:ran it locally\necho hi\n```\n' \
  > "$FIXTURE/openspec/changes/demo-change/design.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "design.md alone is a scanned change, not a broken glob" \
  || fail "design.md alone: rc=$RC out=$OUT"

# Containment applies to every scanned file, not only tasks.md.
new_fixture
clean_tasks_md
ln -s /etc/hosts "$FIXTURE/openspec/changes/demo-change/design.md"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "a symlinked design.md is a containment refusal" \
  || fail "design.md symlink: rc=$RC out=$OUT"
```

- [ ] **Step 3: Run the harness and watch the new cases fail**

Run: `./scripts/test-check-plan-provenance.sh`
Expected: the `proposal.md`, `design.md`, `design.md alone` and symlink cases FAIL; the `specs/` and
archived cases already pass (nothing outside `tasks.md` is read yet).

- [ ] **Step 4: Declare the scan set**

Add beside the other module-level constants:

```python unverified:confirm no other module-level name in this file already reads SCANNED_FILENAMES
# The planning artifacts a change is scanned for. specs/ is deliberately
# absent: spec text legislates rather than describes, and a requirement
# forbidding a shape must be able to name that shape.
SCANNED_FILENAMES = ("tasks.md", "design.md", "proposal.md")
```

- [ ] **Step 5: Generalise the containment gate**

Rename `verify_tasks_md_containment` to `verify_scanned_file_containment` and give it a third
parameter, `filename`, replacing every internal literal `"tasks.md"` with it. The expected path it
confirms becomes `<changes_dir>/<name>/<filename>`. Every refusal message keeps its shape, naming
the actual filename instead of the literal. Behaviour, exit code and rank are otherwise unchanged.

- [ ] **Step 6: Widen the scan loop**

Replace the single-candidate body of the `for name in change_dirs:` loop at
`scripts/check-plan-provenance.py:2381-2433` with a nested loop:

```python verified:the inner body is quoted from scripts/check-plan-provenance.py:2382-2433 @ c1f5aa6
    for name in change_dirs:
        for filename in SCANNED_FILENAMES:
            candidate = os.path.join(changes_dir, name, filename)
            try:
                os.lstat(candidate)
            except (FileNotFoundError, NotADirectoryError):
                continue
            except OSError as exc:
                msg = (
                    f"{os.path.relpath(candidate, repo_root)}: cannot "
                    f"determine whether {filename} exists ({exc.strerror or exc}) "
                    "— refusing to report a clean run"
                )
                print(msg, file=sys.stderr)
                file_errors.append(msg)
                continue

            if not guard.verify_scanned_file_containment(
                candidate, changes_dir, filename
            ):
                continue
            err = guard.check_file(candidate, repo_root)
            if err is not None:
                file_errors.append(err)
            scanned += 1
```

Each candidate keeps the existing per-file semantics: an absent file is an ordinary skip, an
undeterminable one is a scan-integrity failure recorded in `file_errors`, and a containment refusal
skips that one candidate while the scan continues.

- [ ] **Step 7: Restate the scan-integrity invariant and the success line**

The failure at `scripts/check-plan-provenance.py:2435-2455` currently means "change directories
exist but none produced a `tasks.md`". Restate it as "none produced any of `SCANNED_FILENAMES`", and
reword its message accordingly — a change mid-planning that has a `design.md` and no `tasks.md` must
not trip it. Then reword the success line at `scripts/check-plan-provenance.py:2574` to drop the
`tasks.md` specificity:

```python unverified:confirm this is the exact final print in main() after the edits above
    print(f"check-plan-provenance: {scanned} file(s) scanned, all provenance stated")
```

- [ ] **Step 8: Update the three assertions coupled to the old success line**

`scripts/test-check-plan-provenance.sh:1017`, `:1525` and `:2091` each compare against
`check-plan-provenance: 1 tasks.md file(s) scanned, all provenance stated`. Update all three to the
new wording. They are exact string comparisons, so a missed one fails loudly rather than silently.

- [ ] **Step 9: Correct the docstring's scope paragraph**

The paragraph at `scripts/check-plan-provenance.py:88-95` states the scope as
`openspec/changes/*/tasks.md` only. Restate it for the three-file set, keeping the over-firing
rationale but re-pointing it: it now explains why `specs/` is excluded and why the widening had to
remove its own false-positive classes first.

- [ ] **Step 10: Run the harness and the lint guards**

Run: `./scripts/test-check-plan-provenance.sh && scripts/check-plan-provenance.sh && scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: `All check-plan-provenance assertions passed`, then all three guards exit 0 — including
over this change's own `design.md` and `proposal.md`, which are now in scope. A hit there is a real
finding in this plan's own artifacts, to be fixed by stating provenance, never by narrowing the
guard.

---

### Task 4: Bring the contract into line

**Files:**
- Modify: `skills/myflow-contracts/plan-provenance.md` — the section "The guard's scope, and why it
  is narrow"

**Interfaces:**
- Consumes: the behaviour established by Tasks 1-3.
- Produces: nothing executable.

- [ ] **Step 1: Rewrite the scope section**

"The guard's scope, and why it is narrow" states that the guard reads only a change's own
`tasks.md`. Rewrite it for `tasks.md`, `design.md` and `proposal.md`, and state that `specs/` is
excluded because spec text legislates rather than describes. Keep the `check-references.sh`
over-firing history — it is the reason the exclusion and the exemption both exist — but re-point it
so it justifies those two things rather than the old narrower scope.

- [ ] **Step 2: Document the quotation exemption**

> **Superseded by the pass-17 fix wave, in the same way Task 2's step 3 is, and amended by the
> pass-18 wave.** The inline code span named below is no longer a delimiter, so what the contract
> documents is: matched double quotes as the whole delimiter set, why the code-span class was
> removed, and the three vetoes (class-wide, escape, angle-bracket) with their measured cost. The
> third veto was a raw-HTML construct scanner until pass 18 deleted its grammar in favour of "any
> `<` on the line withdraws the exemption" — decision `quotation-angle-bracket-veto`. The
> instruction text is kept as the record.

Add a subsection recording that a number enclosed in matched double quotes (straight or curly) or an
inline code span is a quotation and needs no tag; that the exemption is line-scoped and fails closed;
that single quotes are excluded because prose apostrophes are unpaired; and that its purpose is to
let this contract's own documentation reproduce the invented baseline it exists to explain.

- [ ] **Step 3: Record what did not change**

State plainly that the guard still checks only that provenance is stated, never that it is true, and
that verifying a `measured:` tag's `@ <ref>` remains out of scope. The section "What the guard does
not do" already carries this; add the ref-verification point to it so a reader of the widened scope
does not infer that widening bought truth-checking.

- [ ] **Step 4: Run the lint guards**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-plan-provenance.sh`
Expected: all three exit 0. `check-references.sh` is the one that catches a heading this rewrite
renamed but left referenced elsewhere.

---

## Self-review

**Spec coverage.** Every requirement in the delta spec maps to a task: the quotation carve-out and
its fail-closed and apostrophe clauses to Task 2; the three-file scope, the `specs/` exclusion and
the generalised containment to Task 3; the issue-key clause to Task 1; the scenarios asserting that
archived plans and out-of-plan documentation stay unscanned are covered by existing harness cases
plus the new archived-`design.md` case in Task 3. The requirement that the guard checks statement
and never truth is unchanged, and Task 4 Step 3 keeps it explicit.

**Placeholder scan.** No step defers work. Every code step carries the code; every run step carries
the command and the expected result. The two `unverified:` tags are hypotheses about the file's
current contents at the moment of editing, and each names exactly what to confirm.

**Type consistency.** `SCANNED_FILENAMES` and `verify_scanned_file_containment` are introduced in
Task 3 Steps 4-5 and used in Step 6 under those exact names. `_is_quoted` is defined in Task 2
Step 3 with the signature Step 4 calls. `CLAIM_RE` keeps its name throughout.
