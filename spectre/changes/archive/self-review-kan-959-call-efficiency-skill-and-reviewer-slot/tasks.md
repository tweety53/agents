# self-review-kan-959-call-efficiency-skill-and-reviewer-slot

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.
- No new run-loaded prose: the only non-Go edit is a comment in `scripts/check-verbatim-moves.sh`, outside the verbatim-moves corpus.

- [x] 1. `render-slot-prompt.sh` accepts any slot whose template exists
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the test of step 2 first and run it: it fails with `brand-new: exit 2 stderr "render-slot-prompt: no reviewer template for slot \"brand-new\"\n"` — the defect: a slot with a template on disk is refused because it is not in `rspSlots`.
  - [x] **Step 2: Write the failing test.** Append to `stats/internal/guard/render_slot_prompt_test.go`, after `TestRenderSlotPromptAlwaysOnRoles`:

```go verified:applied with step 3's code at 1a422665 — fails without the change (exit 2, no reviewer template), passes with it; dropping step 3's `/` check fails it on the ../ slot; go vet and gofmt -l clean
// TestRenderSlotPromptSlotFromTemplate pins the kan-959 self-review fix: a
// slot is valid when <skill-dir>/<id>-reviewer-prompt.md exists, so a slot
// added to the skill directory renders without a rebuild of the guard; an id
// with no template, or carrying a `/`, is still refused.
func TestRenderSlotPromptSlotFromTemplate(t *testing.T) {
	t.Parallel()
	real := rspSkillDir(t)
	skill := t.TempDir()
	for _, f := range []string{"review-panel.md", "reviewer-calibration.md", "engineering-principles.md"} {
		writeFile(t, skill+"/"+f, rbRead(real+"/"+f))
	}
	writeFile(t, skill+"/brand-new-reviewer-prompt.md", rbRead(real+"/primary-reviewer-prompt.md"))
	fx := rspNewFx(t)
	r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "brand-new", "-diff", "final", "--", fx.canon)
	if r.rc != 0 || !strings.Contains(got, "## PASS brand-new") {
		t.Fatalf("brand-new: exit %d stderr %q; want 0 and a PASS brand-new section", r.rc, r.err)
	}
	for _, slot := range []string{"no-such-slot", "../" + filepath.Base(skill) + "/brand-new"} {
		if r, _ := fx.run(t, skill, skill, "0", fx.canon, fx.plan, slot, "-diff", "final", "--", fx.canon); r.rc != 2 || !strings.Contains(r.err, "no reviewer template") {
			t.Errorf("%s: exit %d stderr %q; want 2 naming no reviewer template", slot, r.rc, r.err)
		}
	}
}
```

    Run `cd stats && go test ./internal/guard/ -run 'TestRenderSlotPromptSlotFromTemplate' -count=1` and confirm it FAILS as step 1 states.
  - [x] **Step 3: Derive slot validity from the template.** In `stats/internal/guard/renderslotprompt.go`, delete the `rspSlots` comment and declaration from the `var (…)` block, leaving `rspUnfilled` and its comment as the block's only entry:

```go verified:applied at 1a422665 — go vet ./internal/guard/ clean
	// rspSlots are the slots with a template -- test-sanity and
	// call-efficiency are the always-on extra dispatch's roles (kan-959);
	// mutation's brief stays typed, and an exp-<name> slot's template is
	// experimental/<name>.md.
	rspSlots = map[string]bool{"primary": true, "principles": true, "failure-modes": true, "test-sanity": true, "call-efficiency": true}
```

    Then, in `renderSlotPrompt`'s slot switch, replace the line

```go verified:read at 1a422665 in renderSlotPrompt's slot switch
		case rspSlots[s], len(s) > len("exp-") && strings.HasPrefix(s, "exp-") && !strings.Contains(s, "/"):
```

    with these lines, keeping `case s == "mutation":` above them and the `roles = append(roles, s)` and `default:` refusal below them unchanged:

```go verified:applied with step 2's test at 1a422665 — the test passes, the whole internal/guard package passes, go vet and gofmt -l clean
		// Any other slot is valid when its template exists -- <id>-reviewer-prompt.md,
		// or experimental/<name>.md for exp-<name> -- so a slot added to the skill
		// directory renders without a rebuild of this guard.
		case !strings.Contains(s, "/") && (len(s) > len("exp-") && strings.HasPrefix(s, "exp-") || isFile(skill+"/"+s+"-reviewer-prompt.md")):
```

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** `TestRenderSlotPromptSlotFromTemplate`
**Regression:** reverting the commit refuses a slot whose template exists: `brand-new` exits 2 with `no reviewer template`.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/render_slot_prompt_test.go @ 1a422665 -->
**Commit:** fix(guard): render-slot-prompt accepts any slot whose template exists
**After:** none

**Decision:** slot-from-template

- [x] 2. `check-verbatim-moves.sh` FAIL lines print the acknowledgement line itself
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the test of step 2 first and run it: the FAIL lines print `## Old heading`, `## New heading` and `\*Literal\* asterisks stay.` unescaped, and the re-run after copying them into `verbatim-moves.txt` still exits 1 — the defect.
  - [x] **Step 2: Write the failing test.** In `stats/internal/guard/check_verbatim_moves_test.go`, in `TestCheckVerbatimMoves`'s case `an unescaped heading in the list is a comment, not an acknowledgement`, change the `want` entry `"FAIL new run-loaded text (paraphrase?) :: ## Transition table"` to `"FAIL new run-loaded text (paraphrase?) :: \\## Transition table"`. Then insert, directly before `// TestCheckVerbatimMovesSymlinkSkipped`:

```go verified:applied with step 3's code at 1a422665 — fails without the change (FAIL lines unescaped, re-run exit 1), passes with it; dropping step 3's `\` case fails it on the literal sentence; go vet and gofmt -l clean
// TestCheckVerbatimMovesFailLineCopiesIntoAcks pins the kan-959 self-review
// fix: each FAIL line's text after "::" is the acknowledgement line itself --
// a heading, and a sentence opening with a backslash, print escaped -- so
// copying every FAIL line's text into verbatim-moves.txt as printed clears
// the run.
func TestCheckVerbatimMovesFailLineCopiesIntoAcks(t *testing.T) {
	t.Parallel()
	root := vmRepo(t, map[string]string{vmPipe: "# Pipeline\n\n## Old heading\n\n" + vmWhy + "\n", vmRat: vmBase[vmRat]})
	writeFile(t, filepath.Join(root, vmPipe), "# Pipeline\n\n## New heading\n\n\\*Literal\\* asterisks stay.\n\n"+vmWhy+"\n")
	rc, out, _ := vmRun(root)
	if rc != 1 {
		t.Fatalf("rc = %d, want 1\n%s", rc, out)
	}
	var acks []string
	for _, l := range strings.Split(out, "\n") {
		if _, text, ok := strings.Cut(l, " :: "); ok && strings.HasPrefix(l, "FAIL ") {
			acks = append(acks, text)
		}
	}
	for _, want := range []string{`\## Old heading`, `\## New heading`, `\\*Literal\* asterisks stay.`} {
		if !strings.Contains(strings.Join(acks, "\n"), want) {
			t.Errorf("FAIL lines lack %q:\n%s", want, out)
		}
	}
	writeFile(t, filepath.Join(root, "spectre/changes/demo/verbatim-moves.txt"), strings.Join(acks, "\n")+"\n")
	if rc, out, _ := vmRun(root); rc != 0 {
		t.Errorf("rc = %d after acknowledging every FAIL line as printed, want 0\n%s", rc, out)
	}
}

```

    Run `cd stats && go test ./internal/guard/ -run 'TestCheckVerbatimMoves' -count=1` and confirm both the new test and the changed case FAIL.
  - [x] **Step 3: Escape the FAIL lines.** In `stats/internal/guard/verbatimmoves.go`, change the two FAIL `Fprintf` calls' final argument from `s` to `vmAckLine(s)` — `"FAIL deleted or reworded :: %s\n"` and `"FAIL new run-loaded text (paraphrase?) :: %s\n"`; no other line. Insert directly before the `// vmShort is the prototype's` comment:

```go verified:applied with step 2's test at 1a422665 — every TestCheckVerbatimMoves* passes, go vet and gofmt -l clean
// vmAckLine is s as verbatim-moves.txt lists it, so a FAIL line's text after
// "::" copies into the list as printed: a leading `#` would read as a comment
// and a leading `\` is dropped, so either gets one `\` before it.
func vmAckLine(s string) string {
	if strings.HasPrefix(s, "#") || strings.HasPrefix(s, `\`) {
		return `\` + s
	}
	return s
}

```

  - [x] **Step 4: Update the script header.** In `scripts/check-verbatim-moves.sh`, replace the two header lines

```bash verified:read at 1a422665 in the header's acknowledgement paragraph
# `#` lines are comments and a leading `\` is dropped, so a heading is listed
# as `\## …`.
```

    with:

```bash verified:bash -n scripts/check-verbatim-moves.sh clean with these lines applied, spiked at 1a422665
# `#` lines are comments and a leading `\` is dropped. The FAIL line already
# prints a sentence opening with `#` or `\` behind one `\`, so a heading is
# listed as `\## …`, exactly as printed.
```

  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVerbatimMoves' -count=1`, then `bash -n scripts/check-verbatim-moves.sh && scripts/check-verbatim-moves.sh` from the repository root — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/verbatimmoves.go`, `stats/internal/guard/check_verbatim_moves_test.go`, `scripts/check-verbatim-moves.sh`
**Tests:** `TestCheckVerbatimMovesFailLineCopiesIntoAcks`
**Regression:** reverting the commit prints `## …` headings and `\`-leading sentences raw on FAIL lines; copied as printed they acknowledge nothing and the re-run exits 1.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/check_verbatim_moves_test.go @ 1a422665 -->
**Commit:** fix(guard): verbatim-moves FAIL lines print the acknowledgement line itself
**After:** none

**Decision:** fail-line-is-ack-line
