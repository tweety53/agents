## Task 2 report: Match both plan-commit wordings in the self-review context script

**Status:** DONE

### Files changed

- `scripts/gather-self-review-context.sh`
  - Line ~427 comment: renamed the referenced commit from "plan, test guide and session records"
    to "plan and session records" (matches the post-rename subject Task 3 introduces).
  - `PLAN_SHA`'s `--grep` (was line 447): widened to the alternation
    `^chore\(${NAME_RE}\): plan(, test guide and| and) session records`, with a new comment
    explaining the alternation exists for back-compat with pre-rename changes and that it must
    stay in sync with `IMPL_SHA`'s exclusion below.
  - `IMPL_SHA`'s `grep -Ev` exclusion (was line 456): widened the same way —
    `plan(, test guide and| and) session records|^[0-9a-f]+ chore\(${NAME_RE}\): .*archive` — with
    a comment stating the specific failure mode if the two sites drift: the plan commit outranks
    the true implementation commit by recency, so an unexcluded wording makes `head -1` resolve
    `IMPL_SHA` to the plan commit itself, and the real implementation commit is never reached.

- `scripts/test-gather-self-review-context.sh`
  - Added a new section (before the "Invocation contract" section, after the F25 section) with a
    fresh fixture repo using the new-wording plan-commit subject
    (`chore(demo): plan and session records NEW-WORDING-PLAN-COMMIT-BODY`), alongside a distinct
    implementation commit and an archive commit. Asserts:
    1. exits 0
    2. plan commit content is present in the bundle
    3. **the implementation commit's own body marker is present in the bundle** — this is the
       decisive assertion. It fails, not just an absent `PLAN_SHA` match, when `IMPL_SHA`'s
       exclusion isn't widened: the plan commit shadows the true implementation commit via
       `head -1`'s recency ordering, so the implementation commit's content never appears at all.
  - Left both existing old-wording fixtures untouched exactly as instructed: `add_commits()`
    (line 96, `"chore(demo): plan, test guide and session records PLAN-COMMIT-BODY"`) and the F8
    section fixture (line 625, `"chore(demo): plan, test guide and session records"`).

### TDD evidence (RED → GREEN)

- **RED** (before widening the greps): new section's first two assertions passed by coincidence
  (plan content did appear in the bundle — but only because `IMPL_SHA` had wrongly resolved to the
  plan commit itself, printing its `git log --stat` entry once). The third assertion — the
  implementation-commit marker — correctly failed:
  `FAIL: new-wording plan commit: implementation commit not resolved as IMPL_SHA — the plan commit
  likely shadowed it: ...` — proving the fixture reproduces the exact bug described in the task
  brief (plan commit reported as both `PLAN_SHA` candidate content and as `IMPL_SHA`, with the true
  implementation commit never surfacing). Run: `1 case(s) failed`.
- **GREEN** (after widening both grep sites): `scripts/test-gather-self-review-context.sh` →
  `gather-self-review-context: all cases pass`, including both old-wording fixtures (`add_commits`
  section: "all sources present: IMPL_SHA content in bundle" / "PLAN_SHA content in bundle", and
  the F8 "archive substring in impl subject" section) and the new "new-wording plan commit"
  section.

### Verification

- `scripts/test-gather-self-review-context.sh` → all cases pass.
- Lint guards from `.myflow/project.md`'s `## lint` list — `check-vocabulary.sh`,
  `check-references.sh`, `check-plan-provenance.sh`, `check-task-build-green.sh`,
  `check-workspace-isolation.sh`, `check-contract-budget.sh` — all exit 0. None of them cover
  `scripts/*.sh` content directly (contract budget only covers `skills/myflow-contracts/*.md` and
  `skills/*/SKILL*.md`), so no budget row needed raising for this task's files.
- `shellcheck` not installed in this environment; not run.

### Plan-provenance note on the brief's snippets

- Step 1's snippet was tagged `unverified` (hypothesis about fixture-builder/assertion-helper
  names). Read the actual harness instead: fixtures are built with `new_repo` +
  `git commit -q --allow-empty -m "..."` inside `(cd "$REPO" && ...)`, and assertions use the
  file's own `pass`/`fail` + `case "$OUT" in *MARKER*) ... esac` idiom (there is no separate
  `assert_*` helper). Used that idiom, matching the F8 section's style exactly (which builds its
  three commits the same way).
- Step 3's two snippets (verified) were pasted essentially as given. Both fit their surroundings
  without adjustment — same indentation, same `-E`/`-Ev` flag usage, same `${NAME_RE}` variable
  already in scope at both sites.

### Concerns

None. No suppression markers added, no guard weakened, both back-compat fixtures left byte-for-byte
as they were, and the fixture proves the specific failure mode (implementation commit shadowed by
the plan commit) rather than only checking `PLAN_SHA` was found.
