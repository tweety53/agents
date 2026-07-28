# myflow Economical Updates Implementation Plan

> **Execution:** `/myflow-do` runs Basic Workflow #2–#6 via `openspec-apply-superpowers` (#7 runs later, in `/myflow-review`). Mark each checkbox when its task passes spec + quality review (SDD #6).

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Cut myflow's fixed per-session token cost by moving four contract sections out of the always-on rule layer, making the pure-state-write path mechanical, and moving one review panel slot to the economy tier.

**Architecture:** Four sections leave `rules/myflow-manual-review.mdc` for a new on-demand `myflow-contracts` skill, each leaving a discoverable stub. A new `state-advance.sh` handles the mechanical state write and escalates to the existing skill by exit code. A new `check-references.sh` guard fails when a cross-referenced section no longer exists, protecting the move.

**Tech Stack:** Bash 3.2 (macOS system bash), `jq`, `git`, Markdown. No build system, no package manager.

## Global Constraints

- **No commits.** `/myflow-do` stages only. Every task ends with `git add`, never `git commit`, `git push`, or a PR. This overrides the commit steps in the writing-plans template.
- **Bash 3.2 compatible.** macOS ships bash 3.2; no associative arrays (`declare -A`), no `mapfile`/`readarray`, no `${var^^}`.
- **All new scripts:** `#!/usr/bin/env bash`, `set -euo pipefail`, executable mode `755`, argument-free, own their scan set in a single `DEFAULT_TARGETS` definition, exit non-zero on any failure, report failures as `file:line`.
- **Never weaken a guard to make it pass.** Per the Lint Fix Priority rule, fix the input, not the assertion.
- **Contract text moves verbatim.** When relocating a section, copy its body byte-for-byte; the only permitted edits are the canonical-authority rewording and cross-reference paths. Do not reword, summarize, or "improve" contract prose in this change.
- **Repo root:** `/Users/tweety53/Projects/agents` (the apply worktree once one exists — resolve from `git worktree list`, never the main checkout).

---

## File Structure

| File | Responsibility |
|------|----------------|
| `scripts/check-references.sh` | Guard: every referenced section still exists in the file it is referenced from |
| `scripts/test-check-references.sh` | Assertion harness for the above, against sandboxed fixtures |
| `scripts/test-state-advance.sh` | Assertion harness for `state-advance.sh`, against sandboxed state files |
| `skills/myflow-contracts/SKILL.md` | Index naming the four contract files and when each is needed |
| `skills/myflow-contracts/state-file.md` | State file contract (moved) |
| `skills/myflow-contracts/state-self-heal.md` | State self-heal contract (moved) + the script's narrowing |
| `skills/myflow-contracts/project-configuration.md` | Project configuration contract (moved) |
| `skills/myflow-contracts/jira-integration.md` | Jira integration contract (moved) |
| `skills/myflow-state-advance/state-advance.sh` | Mechanical state write; escalates by exit code |
| `rules/myflow-manual-review.mdc` | Always-on judgment only; four stubs pointing at the contracts |
| `.myflow/project.md` | This repo's apps/test/lint/standards/jira for myflow |

---

## Task 1: Reference guard

**Files:**
- Create: `scripts/check-references.sh`
- Create: `scripts/test-check-references.sh`
- Modify: any file needing a `refs-guard:allow` marker (discovered in Step 6)

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `scripts/check-references.sh`, argument-free, exit `0` clean / non-zero with `file:line` lines. Task 2 relies on it to prove the contract move left no stale pointer; Task 6 wires it into `/myflow-review`.

- [x] **Step 1: Write the failing test harness**

Create `scripts/test-check-references.sh`:

```bash
#!/usr/bin/env bash
# Assertion harness for check-references.sh. Builds fixture trees under a
# sandboxed TMPDIR and asserts the guard's exit status and output. Never
# touches the real repository tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-references.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <fixture_dir> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$(cd "$1" && "$GUARD" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/refs-test.XXXXXX")"
  mkdir -p "$FIXTURE/rules" "$FIXTURE/skills/demo"
}

# 1. A live reference passes.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
printf 'Resolve it per **State file** in `rules/contract.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "live reference passes" || fail "live reference: rc=$RC out=$OUT"

# 2. A moved section fails and is reported as file:line.
new_fixture
printf '## Something else\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
printf 'Resolve it per **State file** in `rules/contract.mdc`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "moved section fails" || fail "moved section: expected non-zero"
case "$OUT" in
  *"skills/demo/SKILL.md:1"*) pass "reports file:line" ;;
  *) fail "reports file:line: out=$OUT" ;;
esac

# 3. All three phrasing variants are checked identically.
new_fixture
printf '## State file\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
{
  printf 'see **State file** in `rules/contract.mdc`\n'
  printf 'per **State file** in `rules/contract.mdc`\n'
  printf 'defined once under **State file** in `rules/contract.mdc`\n'
} > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "phrasing variants pass" || fail "phrasing variants: out=$OUT"

new_fixture
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
{
  printf 'see **State file** in `rules/contract.mdc`\n'
  printf 'per **State file** in `rules/contract.mdc`\n'
  printf 'defined once under **State file** in `rules/contract.mdc`\n'
} > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "phrasing variants all fail when absent" || fail "variants: expected non-zero"
for n in 1 2 3; do
  case "$OUT" in
    *"SKILL.md:$n"*) ;;
    *) fail "variant on line $n not reported: out=$OUT" ;;
  esac
done

# 4. Multi-section lines pass when at least one bold token resolves.
new_fixture
printf '## Stage transitions\n\nbody\n\n## State file\n\nbody\n' \
  > "$FIXTURE/rules/contract.mdc"
printf 'Follow `rules/contract.mdc` — sections **Stage transitions**, **State file**.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "multi-section line passes" || fail "multi-section: out=$OUT"

# 5. An allow marker suppresses a line.
new_fixture
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
printf '**Do not** copy from `rules/contract.mdc` <!-- refs-guard:allow -->\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "allow marker suppresses" || fail "allow marker: out=$OUT"

# 6. Fenced code blocks are skipped.
new_fixture
printf '## Gone\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
{
  printf '```\n'
  printf 'see **State file** in `rules/contract.mdc`\n'
  printf '```\n'
} > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "fenced block skipped" || fail "fenced block: out=$OUT"

# 7. A path that does not resolve to a file is skipped, not failed.
new_fixture
printf 'see **Whatever** in `docs/manual-test/<name>.md`\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "unresolvable path skipped" || fail "unresolvable path: out=$OUT"

# 8. Paths relative to the referring file resolve.
new_fixture
mkdir -p "$FIXTURE/skills/other"
printf '## Panel re-runs\n\nbody\n' > "$FIXTURE/skills/other/SKILL.md"
printf 'Follow **Panel re-runs** in `../other/SKILL.md`.\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "relative path resolves" || fail "relative path: out=$OUT"

# 9. Headings match case-insensitively and ignore backticks in the heading.
new_fixture
printf '## The `automerge` flag\n\nbody\n' > "$FIXTURE/rules/contract.mdc"
printf 'see **The automerge flag** in `rules/contract.mdc`\n' \
  > "$FIXTURE/skills/demo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "heading normalization" || fail "heading normalization: out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '\n%d assertion(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '\nAll check-references assertions passed\n'
```

```bash
chmod 755 scripts/test-check-references.sh
```

- [x] **Step 2: Run the harness to verify it fails**

Run: `scripts/test-check-references.sh`
Expected: FAIL — `check-references.sh` does not exist yet, so the harness errors on the missing `$GUARD`.

- [x] **Step 3: Write the guard**

Create `scripts/check-references.sh`:

```bash
#!/usr/bin/env bash
# check-references.sh — fail when a cross-referenced section no longer exists
# in the file it is referenced from.
#
# Rule: for every line carrying BOTH a **bold token** AND a backticked .md/.mdc
# path that resolves to a real file, at least one bold token on that line must
# match a `##` or `###` heading in that file.
#
# This is the companion guard to check-vocabulary.sh. That one greps for
# known-retired literals; this one catches a section that MOVED, which no
# literal list can know about in advance.
#
# Takes no arguments: the scan set lives here, in one place, so no call site
# can narrow it. Lines carrying `refs-guard:allow` are skipped — use it for a
# line whose bold text is emphasis rather than a section name.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DEFAULT_TARGETS=(
  "rules"
  "skills"
  "commands"
  "commands-claude"
  "README.md"
  "AGENTS.md"
  "CLAUDE.md"
)

FAILURES=0

# headings_of <file> — every ## / ### heading, normalized: markers stripped,
# backticks and emphasis removed, trimmed, lowercased.
headings_of() {
  sed -n 's/^###\{0,1\} //p' "$1" \
    | tr -d '`*' \
    | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' \
    | tr '[:upper:]' '[:lower:]'
}

normalize_token() {
  printf '%s' "$1" | tr -d '`*' \
    | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' \
    | tr '[:upper:]' '[:lower:]'
}

check_file() {
  local file="$1" lineno=0 in_fence=0 line
  while IFS= read -r line || [ -n "$line" ]; do
    lineno=$((lineno + 1))

    case "$line" in
      '```'*|'~~~'*) in_fence=$((1 - in_fence)); continue ;;
    esac
    [ "$in_fence" -eq 1 ] && continue
    case "$line" in *refs-guard:allow*) continue ;; esac
    case "$line" in *'**'*) ;; *) continue ;; esac

    # Every backticked token ending in .md or .mdc on this line.
    local paths
    paths="$(printf '%s\n' "$line" \
      | grep -oE '`[^`]+\.mdc?`' \
      | tr -d '`' || true)"
    [ -n "$paths" ] || continue

    local bold
    bold="$(printf '%s\n' "$line" \
      | grep -oE '\*\*[^*]+\*\*' \
      | sed 's/^\*\*//; s/\*\*$//' || true)"
    [ -n "$bold" ] || continue

    local path resolved
    while IFS= read -r path; do
      [ -n "$path" ] || continue
      resolved=""
      if [ -f "$REPO_ROOT/$path" ]; then
        resolved="$REPO_ROOT/$path"
      elif [ -f "$(dirname "$file")/$path" ]; then
        resolved="$(dirname "$file")/$path"
      fi
      # A path that resolves to nothing is out of scope: templated paths like
      # docs/manual-test/<name>.md are legitimate and must not fail the guard.
      [ -n "$resolved" ] || continue

      local heads matched token
      heads="$(headings_of "$resolved")"
      [ -n "$heads" ] || continue
      matched=0
      while IFS= read -r token; do
        [ -n "$token" ] || continue
        token="$(normalize_token "$token")"
        # The path itself is often bolded; that is a file reference, not a
        # section name, so it never counts as a match.
        case "$token" in */*) continue ;; esac
        if printf '%s\n' "$heads" | grep -qxF "$token"; then
          matched=1
          break
        fi
      done <<EOF
$bold
EOF
      if [ "$matched" -eq 0 ]; then
        printf '%s:%d: no bold token resolves to a heading in %s\n' \
          "${file#"$REPO_ROOT"/}" "$lineno" "$path"
        FAILURES=$((FAILURES + 1))
      fi
    done <<EOF
$paths
EOF
  done < "$file"
}

main() {
  local target
  for target in "${DEFAULT_TARGETS[@]}"; do
    local full="$REPO_ROOT/$target"
    [ -e "$full" ] || continue
    if [ -d "$full" ]; then
      while IFS= read -r f; do
        check_file "$f"
      done < <(find "$full" -type f \( -name '*.md' -o -name '*.mdc' \) | sort)
    else
      check_file "$full"
    fi
  done

  if [ "$FAILURES" -ne 0 ]; then
    printf '\n%d stale reference(s) found.\n' "$FAILURES" >&2
    printf 'Fix the reference, or mark the line refs-guard:allow if the bold text is emphasis.\n' >&2
    exit 1
  fi
  printf 'check-references: all referenced sections resolve\n'
}

main "$@"
```

```bash
chmod 755 scripts/check-references.sh
```

- [x] **Step 4: Run the harness to verify it passes**

Run: `scripts/test-check-references.sh`
Expected: PASS — every `ok:` line printed, `All check-references assertions passed`, exit `0`.

- [x] **Step 5: Verify the guard is cwd-independent**

Run: `cd /tmp && /Users/tweety53/Projects/agents/scripts/check-references.sh; echo "rc=$?"`
Expected: it scans the repo (not `/tmp`) because `REPO_ROOT` derives from `BASH_SOURCE`.

- [x] **Step 6: Baseline the current tree and resolve pre-existing failures**

Run: `scripts/check-references.sh; echo "rc=$?"`

If it exits non-zero **before** any contract has moved, each hit is pre-existing drift or a false
positive. For each reported `file:line`, decide and act:

- The bold text names a section that genuinely moved or was renamed → **fix the reference**.
- The bold text is emphasis, not a section name (e.g. `**Do not** copy from …`) → append
  `<!-- refs-guard:allow -->` to that line.

Re-run until it exits `0`. Record in the progress ledger how many hits were fixed versus marked.

- [x] **Step 7: Verify the other guards still pass**

Run: `scripts/check-vocabulary.sh && scripts/test-setup.sh`
Expected: both exit `0`.

- [x] **Step 8: Stage**

```bash
git add scripts/check-references.sh scripts/test-check-references.sh
git add -A
```

---

## Task 2: Extract the four contracts

**Files:**
- Create: `skills/myflow-contracts/SKILL.md`
- Create: `skills/myflow-contracts/state-file.md`
- Create: `skills/myflow-contracts/state-self-heal.md`
- Create: `skills/myflow-contracts/project-configuration.md`
- Create: `skills/myflow-contracts/jira-integration.md`
- Modify: `rules/myflow-manual-review.mdc` (remove four bodies, add four stubs)
- Modify: `scripts/test-setup.sh` (additive assertions)
- Modify: every file referencing one of the four moved sections

**Interfaces:**
- Consumes: `scripts/check-references.sh` from Task 1 — it is the proof this task's sweep is complete.
- Produces: the five files above at `skills/myflow-contracts/`. Task 4 references `state-self-heal.md`; Task 6's `.myflow/project.md` is governed by `project-configuration.md`.

- [x] **Step 1: Write the failing installer assertions**

In `scripts/test-setup.sh`, add to the global-install case assertions that the new skill is
installed into all three harness skill directories and that the managed block does not carry the
contract bodies. Follow the file's existing assertion style and its sandboxed `HOME`; add these as
new assertions, never by loosening an existing one:

```bash
# myflow-contracts must install like any other skill, through install_skills,
# and must NOT be inlined into the managed block — that is the whole point of
# extracting it from the always-on rule.
assert_exists "$SANDBOX_HOME/.claude/skills/myflow-contracts/SKILL.md"
assert_exists "$SANDBOX_HOME/.cursor/skills/myflow-contracts/SKILL.md"
assert_exists "$SANDBOX_HOME/.codex/skills/myflow-contracts/SKILL.md"
assert_exists "$SANDBOX_HOME/.claude/skills/myflow-contracts/state-file.md"
assert_not_contains "$SANDBOX_HOME/.claude/CLAUDE.md" 'PROJECT_KEY="$(basename'
assert_not_contains "$SANDBOX_HOME/.codex/AGENTS.md" 'PROJECT_KEY="$(basename'
```

Use the helper names `test-setup.sh` already defines. If it has no `assert_not_contains`, add one
next to its existing helpers:

```bash
assert_not_contains() {
  if grep -qF "$2" "$1" 2>/dev/null; then
    fail "$1 must not contain: $2"
  else
    pass "$1 does not contain: $2"
  fi
}
```

- [x] **Step 2: Run the harness to verify it fails**

Run: `scripts/test-setup.sh`
Expected: FAIL — `myflow-contracts/SKILL.md` does not exist, and the managed block still contains
the state-file write template.

- [x] **Step 3: Move the four sections verbatim**

Create `skills/myflow-contracts/` and move each section's body **byte-for-byte** out of
`rules/myflow-manual-review.mdc`:

| Source section | Destination |
|----------------|-------------|
| `## State file` | `skills/myflow-contracts/state-file.md` |
| `## State self-heal` | `skills/myflow-contracts/state-self-heal.md` |
| `## Project configuration` | `skills/myflow-contracts/project-configuration.md` |
| `## Jira integration` | `skills/myflow-contracts/jira-integration.md` |

In each destination file, promote the section heading to `#` (the file's title) and keep every
subheading at its current level. Make exactly two kinds of edit to the moved text:

1. **Canonical authority.** Where the section declared itself canonical over skills, reword so the
   contracts file is canonical — e.g. "**This section is the canonical definition**" becomes
   "**This file is the canonical definition**; if a skill and this file ever disagree, this file
   wins."
2. **Cross-reference paths.** A moved section referring to another moved section now points at its
   sibling file (e.g. `state-file.md`), and one referring to a section still in the rule points at
   `rules/myflow-manual-review.mdc`.

Change nothing else — no rewording, no summarizing, no reformatting.

- [x] **Step 4: Write the index**

Create `skills/myflow-contracts/SKILL.md`:

```markdown
---
name: myflow-contracts
description: The myflow pipeline's contract definitions — state file shape, state self-heal, project configuration, and Jira integration. Load the one file you need; each is canonical for its own contract. Referenced by the stubs in rules/myflow-manual-review.mdc.
allowed-tools: Bash(jq:*), Bash(git:*)
license: MIT
metadata:
  author: gymie
  version: "1.0"
---

myflow's contract definitions, split out of `rules/myflow-manual-review.mdc` so the always-on rule
layer carries only the judgment an agent needs without being asked to load anything.

**Load the one file you need — not this whole directory.**

| File | Load it when you need to |
|------|--------------------------|
| [state-file.md](state-file.md) | Read or write a change's state file: its path, its full shape, monotonic gates, carry-forward |
| [state-self-heal.md](state-self-heal.md) | Validate a state file against on-disk artifacts, or handle a missing/contradicted one |
| [project-configuration.md](project-configuration.md) | Resolve `.myflow/project.md` — apps, run, test, lint, standards, jira — including standards-entry resolution and containment |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins.

The `##` headings inside these files keep the names they had in the rule, so an existing reference
to a section by name still resolves — now to the file that holds it.
```

- [x] **Step 5: Leave a stub at each vacated location**

In `rules/myflow-manual-review.mdc`, at each section's original position, keep the original `##`
heading and add one sentence plus the file path. Example for Jira:

```markdown
## Jira integration

The contract governing how a change is linked to a Jira issue, transitioned through the pipeline,
and has its description synced — including that Jira is never a gate and never blocks a stage
write. **Load `skills/myflow-contracts/jira-integration.md` before any Jira-related step.**
```

Write the equivalent stub for `## State file`, `## State self-heal`, and `## Project
configuration`. A stub must not declare itself canonical over text it no longer holds.

- [x] **Step 6: Sweep every reference to the moved sections**

Find every reference:

```bash
grep -rn "State file\|State self-heal\|Project configuration\|Jira integration" \
  rules skills commands commands-claude README.md AGENTS.md CLAUDE.md
```

For each hit that points a reader at the rule file for one of those sections, repoint it at the
contracts file. Keep the section name in bold — `check-references.sh` matches on it. Expect hits
across `myflow-state-advance`, `openspec-apply-superpowers`, `openspec-propose-superpowers`,
`openspec-review-superpowers`, `openspec-archive-superpowers`, `openspec-fast-path-superpowers`,
`myflow-status`, `myflow-info`, both `commands/` trees, and the three root docs.

- [x] **Step 7: Verify the guards**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/test-setup.sh`
Expected: all three exit `0`. A `check-references.sh` failure here names exactly the file and line
whose pointer the sweep missed — fix it and re-run.

- [x] **Step 8: Verify the size target**

Run:

```bash
wc -c rules/myflow-manual-review.mdc skills/myflow-contracts/*.md
```

Expected: `rules/myflow-manual-review.mdc` is at most 32768 bytes. Record the before (58137) and
after byte counts in the progress ledger as evidence.

- [x] **Step 9: Verify a real global install**

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
wc -c "$SANDBOX/.claude/CLAUDE.md"
ls "$SANDBOX/.claude/skills/myflow-contracts/"
grep -c "myflow:begin" "$SANDBOX/.claude/CLAUDE.md"
```

Expected: `CLAUDE.md` at most ~33 KB, the five contract files present, exactly one begin delimiter.

- [x] **Step 10: Stage**

```bash
git add -A
```

---

## Task 3: state-advance.sh

**Files:**
- Create: `skills/myflow-state-advance/state-advance.sh`
- Create: `scripts/test-state-advance.sh`

**Interfaces:**
- Consumes: the state-file shape from `skills/myflow-contracts/state-file.md` (Task 2).
- Produces: `state-advance.sh --name <n> --target <stage|originStage> --accepted <csv> --by <cmd>`; exit `0` on write, `3` unreadable state or stale worktree, `4` stage mismatch, `5` null `originStage`, `6` corrupt `originStage`. Task 4 wires the commands to these exact codes.

- [x] **Step 1: Write the failing test harness**

Create `scripts/test-state-advance.sh`:

```bash
#!/usr/bin/env bash
# Assertion harness for state-advance.sh. Every case runs against a state file
# in a sandboxed directory; the real state tree under ~/Agents/myflow is never
# read or written.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ADVANCE="$SCRIPT_DIR/../skills/myflow-state-advance/state-advance.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

new_state() {
  SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/state-test.XXXXXX")"
  STATE_FILE="$SANDBOX/demo.json"
  cat > "$STATE_FILE" <<'JSON'
{
  "stage": "awaiting-do-review",
  "gates": { "reviewed": true, "tested": "skipped", "prOpened": null, "prMerged": null },
  "worktree": null,
  "branch": "openspec/demo",
  "originStage": null,
  "artifactUrl": "https://example.invalid/a",
  "jiraIssue": "KAN-10",
  "fastPath": true,
  "REVIEWED_TREE": null,
  "MERGE_BASE": { "/tmp/wt": "abc123" },
  "updatedAt": "2026-07-01T00:00:00Z",
  "updatedBy": "/myflow-do"
}
JSON
}

run_advance() {
  set +e
  OUT="$(MYFLOW_STATE_FILE="$STATE_FILE" "$ADVANCE" "$@" 2>&1)"
  RC=$?
  set -e
}

# 1. Happy path writes stage, updatedAt, updatedBy.
new_state
run_advance --name demo --target do-done \
  --accepted awaiting-do-review,do-review-started --by /myflow-do-done
[ "$RC" -eq 0 ] && pass "happy path exits 0" || fail "happy path: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "do-done" ] \
  && pass "stage written" || fail "stage not written"
[ "$(jq -r '.updatedBy' "$STATE_FILE")" = "/myflow-do-done" ] \
  && pass "updatedBy written" || fail "updatedBy not written"
[ "$(jq -r '.updatedAt' "$STATE_FILE")" != "2026-07-01T00:00:00Z" ] \
  && pass "updatedAt refreshed" || fail "updatedAt not refreshed"

# 2. Unowned fields are carried forward verbatim.
[ "$(jq -r '.gates.tested' "$STATE_FILE")" = "skipped" ] \
  && pass "gates.tested carried" || fail "gates.tested lost"
[ "$(jq -r '.gates.reviewed' "$STATE_FILE")" = "true" ] \
  && pass "gates.reviewed carried" || fail "gates.reviewed lost"
[ "$(jq -r '.jiraIssue' "$STATE_FILE")" = "KAN-10" ] \
  && pass "jiraIssue carried" || fail "jiraIssue lost"
[ "$(jq -r '.artifactUrl' "$STATE_FILE")" = "https://example.invalid/a" ] \
  && pass "artifactUrl carried" || fail "artifactUrl lost"
[ "$(jq -r '.fastPath' "$STATE_FILE")" = "true" ] \
  && pass "fastPath carried" || fail "fastPath lost"
[ "$(jq -r '.MERGE_BASE["/tmp/wt"]' "$STATE_FILE")" = "abc123" ] \
  && pass "MERGE_BASE carried" || fail "MERGE_BASE lost"

# 3. Stage mismatch exits 4 and writes nothing.
new_state
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target review-done --accepted awaiting-pr-review --by /myflow-review-done
[ "$RC" -eq 4 ] && pass "mismatch exits 4" || fail "mismatch: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "mismatch writes nothing" || fail "mismatch mutated the file"

# 4. Missing state file exits 3.
new_state
rm -f "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "missing state exits 3" || fail "missing state: rc=$RC"
[ ! -f "$STATE_FILE" ] && pass "missing state creates nothing" || fail "file was created"

# 5. Unparseable state file exits 3.
new_state
printf 'not json' > "$STATE_FILE"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "unparseable exits 3" || fail "unparseable: rc=$RC"

# 6. Stale worktree exits 3 and writes nothing.
new_state
jq '.worktree = "/nonexistent/worktree"' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
[ "$RC" -eq 3 ] && pass "stale worktree exits 3" || fail "stale worktree: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "stale worktree writes nothing" || fail "stale worktree mutated the file"

# 7. Dynamic target: do-review-started resolves to awaiting-do-review, clears originStage.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "do-review-started"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage \
  --accepted awaiting-fix-review,fix-review-started --by /myflow-do-fix-done
[ "$RC" -eq 0 ] && pass "dynamic target exits 0" || fail "dynamic: rc=$RC out=$OUT"
[ "$(jq -r '.stage' "$STATE_FILE")" = "awaiting-do-review" ] \
  && pass "do-review-started resolves" || fail "dynamic resolution wrong"
[ "$(jq -r '.originStage' "$STATE_FILE")" = "null" ] \
  && pass "originStage cleared" || fail "originStage not cleared"

# 8. Dynamic target: the other five origins target themselves.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "manual-test-done"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$(jq -r '.stage' "$STATE_FILE")" = "manual-test-done" ] \
  && pass "self-targeting origin" || fail "self-targeting origin wrong"

# 9. Null originStage exits 5.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = null' "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$RC" -eq 5 ] && pass "null originStage exits 5" || fail "null originStage: rc=$RC"

# 10. Corrupt originStage exits 6 and is never repaired.
new_state
jq '.stage = "awaiting-fix-review" | .originStage = "awaiting-fix-review"' \
  "$STATE_FILE" > "$STATE_FILE.tmp"
mv "$STATE_FILE.tmp" "$STATE_FILE"
BEFORE="$(cat "$STATE_FILE")"
run_advance --name demo --target originStage --accepted awaiting-fix-review --by /myflow-do-fix-done
[ "$RC" -eq 6 ] && pass "corrupt originStage exits 6" || fail "corrupt originStage: rc=$RC"
[ "$(cat "$STATE_FILE")" = "$BEFORE" ] \
  && pass "corrupt originStage writes nothing" || fail "corrupt originStage mutated the file"

# 11. The written file is valid JSON with every original key present.
new_state
run_advance --name demo --target do-done --accepted awaiting-do-review --by /myflow-do-done
jq -e . "$STATE_FILE" >/dev/null 2>&1 \
  && pass "output is valid JSON" || fail "output is not valid JSON"
for key in stage gates worktree branch originStage artifactUrl jiraIssue fastPath \
           REVIEWED_TREE MERGE_BASE updatedAt updatedBy; do
  jq -e "has(\"$key\")" "$STATE_FILE" >/dev/null 2>&1 \
    || fail "key dropped: $key"
done
pass "all keys retained"

if [ "$FAILURES" -ne 0 ]; then
  printf '\n%d assertion(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '\nAll state-advance assertions passed\n'
```

```bash
chmod 755 scripts/test-state-advance.sh
```

- [x] **Step 2: Run the harness to verify it fails**

Run: `scripts/test-state-advance.sh`
Expected: FAIL — `state-advance.sh` does not exist.

- [x] **Step 3: Write the script**

Create `skills/myflow-state-advance/state-advance.sh`:

```bash
#!/usr/bin/env bash
# state-advance.sh — the mechanical half of myflow's pure state write.
#
# Handles the happy path: state file readable, current stage in --accepted,
# worktree (if any) still present. Anything needing judgment — resolving a name
# across multiple candidates, the stage-mismatch override prompt, artifact-based
# self-heal — is NOT done here. The script exits with a distinct code and the
# calling command loads the myflow-state-advance skill, which behaves exactly as
# it always has.
#
#   0  wrote the new stage
#   2  usage error
#   3  state file missing/unparseable, or a non-null worktree that git no longer knows
#   4  current stage not in --accepted
#   5  --target originStage but originStage is null/missing
#   6  --target originStage but originStage is outside the six legal origins
#
# MYFLOW_STATE_FILE overrides path resolution (used by the test harness).
set -euo pipefail

NAME=""; TARGET=""; ACCEPTED=""; BY=""

while [ $# -gt 0 ]; do
  case "$1" in
    --name)     NAME="${2:-}"; shift 2 ;;
    --target)   TARGET="${2:-}"; shift 2 ;;
    --accepted) ACCEPTED="${2:-}"; shift 2 ;;
    --by)       BY="${2:-}"; shift 2 ;;
    *) printf 'usage error: unknown argument %s\n' "$1" >&2; exit 2 ;;
  esac
done

for required in NAME TARGET ACCEPTED BY; do
  eval "value=\$$required"
  if [ -z "$value" ]; then
    printf 'usage error: --%s is required\n' "$(printf '%s' "$required" | tr '[:upper:]' '[:lower:]')" >&2
    exit 2
  fi
done

command -v jq >/dev/null 2>&1 || { printf 'jq is required\n' >&2; exit 2; }

# Resolve the state file exactly as the State file contract specifies: via
# --git-common-dir, so a worktree and the main checkout agree on one path.
if [ -n "${MYFLOW_STATE_FILE:-}" ]; then
  STATE_FILE="$MYFLOW_STATE_FILE"
else
  MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
  PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
  STATE_FILE="$HOME/Agents/myflow/state/$PROJECT_KEY/$NAME.json"
fi

if [ ! -f "$STATE_FILE" ]; then
  printf 'escalate: no state file at %s\n' "$STATE_FILE" >&2
  exit 3
fi
if ! jq -e . "$STATE_FILE" >/dev/null 2>&1; then
  printf 'escalate: state file is not valid JSON: %s\n' "$STATE_FILE" >&2
  exit 3
fi

CURRENT="$(jq -r '.stage // empty' "$STATE_FILE")"
if [ -z "$CURRENT" ]; then
  printf 'escalate: state file has no stage: %s\n' "$STATE_FILE" >&2
  exit 3
fi

# A worktree the state file names but git no longer knows about is exactly the
# contradiction State self-heal exists for — hand it to the skill.
WORKTREE="$(jq -r '.worktree // empty' "$STATE_FILE")"
if [ -n "$WORKTREE" ] && [ "$WORKTREE" != "null" ]; then
  if ! git worktree list --porcelain 2>/dev/null | grep -qxF "worktree $WORKTREE"; then
    printf 'escalate: state file names a worktree git does not list: %s\n' "$WORKTREE" >&2
    exit 3
  fi
fi

stage_accepted=0
saved_ifs="$IFS"; IFS=','
for candidate in $ACCEPTED; do
  [ "$candidate" = "$CURRENT" ] && stage_accepted=1
done
IFS="$saved_ifs"
if [ "$stage_accepted" -eq 0 ]; then
  printf 'escalate: stage %s is not in accepted set %s\n' "$CURRENT" "$ACCEPTED" >&2
  exit 4
fi

CLEAR_ORIGIN=0
RESOLVED_TARGET="$TARGET"
if [ "$TARGET" = "originStage" ]; then
  ORIGIN="$(jq -r '.originStage // empty' "$STATE_FILE")"
  if [ -z "$ORIGIN" ] || [ "$ORIGIN" = "null" ]; then
    printf 'escalate: --target originStage but originStage is null/missing\n' >&2
    exit 5
  fi
  case "$ORIGIN" in
    awaiting-do-review|do-review-started|do-done|awaiting-manual-test|manual-test-done|awaiting-pr-review) ;;
    *) printf 'escalate: originStage %s is not one of the six legal origins\n' "$ORIGIN" >&2
       exit 6 ;;
  esac
  if [ "$ORIGIN" = "do-review-started" ]; then
    RESOLVED_TARGET="awaiting-do-review"
  else
    RESOLVED_TARGET="$ORIGIN"
  fi
  CLEAR_ORIGIN=1
fi

NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TMP="$STATE_FILE.tmp.$$"

if [ "$CLEAR_ORIGIN" -eq 1 ]; then
  jq --arg s "$RESOLVED_TARGET" --arg t "$NOW" --arg b "$BY" \
    '.stage = $s | .originStage = null | .updatedAt = $t | .updatedBy = $b' \
    "$STATE_FILE" > "$TMP"
else
  jq --arg s "$RESOLVED_TARGET" --arg t "$NOW" --arg b "$BY" \
    '.stage = $s | .updatedAt = $t | .updatedBy = $b' \
    "$STATE_FILE" > "$TMP"
fi
mv "$TMP" "$STATE_FILE"

printf '## Stage advanced\n\n'
printf '**Change:** %s\n' "$NAME"
printf '**Stage:** %s → %s\n' "$CURRENT" "$RESOLVED_TARGET"
printf '**State:** %s\n' "$STATE_FILE"
```

```bash
chmod 755 skills/myflow-state-advance/state-advance.sh
```

Note the `jq` assignment form — it edits three fields of the object it read, so every other field
is carried forward by construction rather than by re-listing it. That is what makes the
carry-forward assertions in Step 1 pass without enumerating each key.

- [x] **Step 4: Run the harness to verify it passes**

Run: `scripts/test-state-advance.sh`
Expected: PASS — `All state-advance assertions passed`, exit `0`.

- [x] **Step 5: Verify no gate can be written**

Run: `grep -n "gates" skills/myflow-state-advance/state-advance.sh`
Expected: no match. The script must have no code path that reads or writes a gate.

- [x] **Step 6: Verify Jira is never contacted**

Run: `grep -niE "jira|atlassian|curl|http" skills/myflow-state-advance/state-advance.sh`
Expected: no match outside the header comment.

- [x] **Step 7: Stage**

```bash
git add -A
```

---

## Task 4: Wire the commands to the script

**Files:**
- Modify: `commands/myflow-start-done.md`, `myflow-do-manual-review.md`, `myflow-do-done.md`, `myflow-do-fix-manual-review.md`, `myflow-do-fix-done.md`, `myflow-manual-test-done.md`, `myflow-review-done.md`
- Modify: the same seven files under `commands-claude/`
- Modify: `skills/myflow-state-advance/SKILL.md`
- Modify: `skills/myflow-contracts/state-self-heal.md`

**Interfaces:**
- Consumes: `state-advance.sh` and its exit codes from Task 3; `state-self-heal.md` from Task 2.
- Produces: no new interface. This task makes the seven commands use the script.

- [x] **Step 1: Add the script-first block to one command and verify the wording**

In `commands-claude/myflow-do-done.md`, insert before the existing "Use the
**myflow-state-advance** skill" line:

```markdown
**Run the script first.** It handles the mechanical write; the skill is only for what needs
judgment.

```bash
~/.claude/skills/myflow-state-advance/state-advance.sh \
  --name <name> --target do-done \
  --accepted awaiting-do-review,do-review-started --by /myflow-do-done
```

- **Exit 0** — print its output and stop. The stage is written; nothing further is needed.
- **Exit 2** — a usage error in this command file. Report it; do not work around it.
- **Exit 3, 4, 5, or 6** — do **not** retry and do **not** hand-edit the state file. Load the
  **myflow-state-advance** skill and follow it from step 1; it owns name resolution, the
  stage-mismatch override prompt, and self-heal.
- **`<name>` omitted** — skip the script and use the skill, which resolves the name first.
```

- [x] **Step 2: Verify the script path resolves in a real install**

Run: `ls -l ~/.claude/skills/myflow-state-advance/state-advance.sh`
Expected: a symlink into the agents checkout, executable. If the symlink is missing, re-run
`./setup.sh global` and re-check — the file must be reachable at the path the command names.

- [x] **Step 3: Apply the same block to the other thirteen command files**

Repeat Step 1 for the remaining six `commands-claude/` files and all seven `commands/` files,
substituting each file's own `--target` and `--accepted` values, which must match its existing
`TARGET_STAGE` / `ACCEPTED_STAGES` exactly:

| Command | `--target` | `--accepted` |
|---------|-----------|--------------|
| `myflow-start-done` | `proposal-done` | `awaiting-proposal-review` |
| `myflow-do-manual-review` | `do-review-started` | `awaiting-do-review` |
| `myflow-do-done` | `do-done` | `awaiting-do-review,do-review-started` |
| `myflow-do-fix-manual-review` | `fix-review-started` | `awaiting-fix-review` |
| `myflow-do-fix-done` | `originStage` | `awaiting-fix-review,fix-review-started` |
| `myflow-manual-test-done` | `manual-test-done` | `awaiting-manual-test` |
| `myflow-review-done` | `review-done` | `awaiting-pr-review` |

- [x] **Step 4: Verify every command file agrees with its own stage table**

Run:

```bash
for f in commands/myflow-*-done.md commands/myflow-*-manual-review.md \
         commands-claude/myflow-*-done.md commands-claude/myflow-*-manual-review.md; do
  printf '=== %s\n' "$f"
  grep -E "TARGET_STAGE|ACCEPTED_STAGES|--target|--accepted" "$f"
done
```

Expected: in each file the `--target`/`--accepted` values match its `TARGET_STAGE`/
`ACCEPTED_STAGES` lines exactly. A mismatch here is the defect this step exists to catch.

- [x] **Step 5: Record the escalation contract in the skill**

In `skills/myflow-state-advance/SKILL.md`, add a section stating that the invoking command runs
`state-advance.sh` first, listing the four escalation codes and what the skill does for each, and
stating that the skill's own workflow is unchanged when it is reached.

- [x] **Step 6: Record the narrowing in the contract**

In `skills/myflow-contracts/state-self-heal.md`, add a subsection stating exactly which checks the
script performs (state file parses; a named `worktree` is still listed by git), which it does not
(every artifact-contradiction check in the table above it), that this applies only to the seven
pure-state-write commands on their happy path, and why it is acceptable — those stages are pure
human-confirmation writes, and `/myflow-review` and `/myflow-finish` verify gate and merge state
independently.

- [x] **Step 7: Verify the guards**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/test-setup.sh && scripts/test-state-advance.sh`
Expected: all four exit `0`.

- [x] **Step 8: Stage**

```bash
git add -A
```

---

## Task 5: Adversarial slot to the economy tier

**Files:**
- Modify: `skills/openspec-apply-superpowers/SKILL.md` (panel table row 4; the economic mapping heading and body; the guardrail near the "Do not pass a `model` override" line)
- Modify: `skills/openspec-apply-fix-superpowers/SKILL.md:43,194`
- Modify: `skills/openspec-fast-path-superpowers/SKILL.md:296`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: nothing later tasks depend on.

- [x] **Step 1: Move slot 4 to the economy tier**

In the strict review panel table, change row `4` (Adversarial) from `parent (**omit** `model`)` to
`**economy** (mapping below)`, and add to its "How to spawn" cell that `model` **must** be set,
matching how row `5+` reads.

- [x] **Step 2: Correct the mapping's scope**

Change the heading `##### Economic model mapping (slots 5+ only)` to
`##### Economic model mapping (slots 4 and 5+)`, and rewrite its opening sentence:

```markdown
**This mapping applies to the Adversarial reviewer (slot 4) and the conditional extra lens
reviewers (slots 5+).** The **required** principles reviewer (slot 2) and the primary reviewer
(slot 0) inherit the **parent model** — pass **no** `model` override at all. Weighing principle
tradeoffs and judging plan alignment is judgment work that degrades on a weaker agent; a narrowed
breadth pass over a bounded diff does not.
```

Also update the closing sentence so it reads "Never skip a slot 4 or 5+ reviewer because of model
selection".

- [x] **Step 3: Extend the guardrail**

Change the guardrail line to:

```markdown
- Do not pass a `model` override to the required Principles reviewer (slot 2) or the primary
  reviewer (slot 0) — both inherit the parent model. Do not omit `model` on the Adversarial
  reviewer (slot 4) or a slot 5+ lens reviewer — resolve it from the economic model mapping.
```

- [x] **Step 4: Sweep the deferring skills**

Run: `grep -n "slots 5+\|slot 5+" skills/openspec-apply-fix-superpowers/SKILL.md skills/openspec-fast-path-superpowers/SKILL.md`

For each hit describing the economic mapping's scope, change it to name slots 4 and 5+. Leave
intact any sentence about which slots the fast path *dispatches* — the fast path still runs only
the three required slots, and slot 4 is not among them.

- [x] **Step 5: Verify no stale scope wording survives**

Run: `grep -rn "slots 5+ only" skills/`
Expected: no match.

- [x] **Step 6: Verify the guards**

Run: `scripts/check-references.sh && scripts/check-vocabulary.sh`
Expected: both exit `0`.

- [x] **Step 7: Stage**

```bash
git add -A
```

---

## Task 6: This repo as a first-class myflow project

**Files:**
- Create: `.myflow/project.md`
- Modify: `skills/openspec-review-superpowers/SKILL.md` (verification gate runs the third guard)

**Interfaces:**
- Consumes: `scripts/check-references.sh` (Task 1); the `.myflow/project.md` format defined by `skills/myflow-contracts/project-configuration.md` (Task 2).
- Produces: `.myflow/project.md`, which every later myflow run in this repository reads.

- [x] **Step 1: Write the project configuration**

Create `.myflow/project.md`:

```markdown
# myflow project configuration — agents

Read by globally installed myflow skills. Every key is optional; anything absent is
auto-detected from the repository instead.

## apps

This repository has **no runnable application**. It is the source of the myflow skills, commands,
and rules, installed elsewhere by `setup.sh`. There is nothing to start, no port, and no URL.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| myflow sources | `/Users/tweety53/Projects/agents` | Bash + Markdown | — | The only repo in scope. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |

## run

There is no service to run. To exercise the installer without touching the real home directory:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

## test

```bash
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-state-advance.sh
```

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
```

**There is no auto-fix command in this repository.** The Lint Fix Priority rule's "run the
auto-fix command first" step is therefore inapplicable here — not skipped. Both guards report
`file:line` and are fixed by editing the offending line, never by weakening the guard or adding a
suppression marker to silence a real hit.

## standards

- `CLAUDE.md`
- `AGENTS.md`

## jira

`KAN`
```

- [x] **Step 2: Verify the declared commands actually run**

Run each command in the `## test` and `## lint` blocks exactly as written.
Expected: all five exit `0`. A command in this file that does not run is worse than an absent key,
because myflow will believe it.

- [x] **Step 3: Wire the third guard into the review gate**

In `skills/openspec-review-superpowers/SKILL.md`, find the verification section that already runs
`scripts/check-vocabulary.sh` and `scripts/test-setup.sh` when the agents repo is among the
affected worktrees. Add `scripts/check-references.sh` alongside them, with one sentence on what it
guards: that a section referenced by name still exists in the file it is referenced from, which no
list of retired literals can catch.

- [x] **Step 4: Verify the review skill names all three guards**

Run: `grep -n "check-vocabulary\|test-setup\|check-references" skills/openspec-review-superpowers/SKILL.md`
Expected: all three appear in the same verification section.

- [x] **Step 5: Full verification sweep**

Run:

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-state-advance.sh
wc -c rules/myflow-manual-review.mdc
```

Expected: all five scripts exit `0`; the rule file is at most 32768 bytes.

- [x] **Step 6: Stage**

```bash
git add -A
git status
```

Confirm every change is staged and nothing is committed.
