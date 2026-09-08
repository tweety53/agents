# kan-393-flow-give-guards-a-canonical-worktree-concept

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Four commits in dependency order: the lib's ref helper and its harness (task 1), the gatherer's
content-directory resolution and its harness (task 2), then the two caller files' one-argument
prose (tasks 3 and 4, independent of each other). The approved design is
`docs/superpowers/specs/2026-09-08-kan-393-flow-give-guards-a-canonical-worktree-concept-design.md`;
`design.md` beside this file carries the decisions.

**Baseline, measured before any edit:**

- `scripts/test-gather-dispatch-context.sh` prints 75 `ok:` lines and `all cases passed`;
  `scripts/test-lib-change-plan.sh` prints 24 and `all cases passed`.
  <!-- measured: scripts/test-gather-dispatch-context.sh | grep -c '^ok'; scripts/test-lib-change-plan.sh | grep -c '^ok' @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept -->
- `scripts/check-contract-budget.sh` reports 71 owned files within budget;
  `skills/flow/implement.md` is 29117 bytes against its 32500-byte row, `skills/flow/review-panel.md`
  49647 against 58732 — no budget row moves.
  <!-- measured: scripts/check-contract-budget.sh; wc -c skills/flow/implement.md skills/flow/review-panel.md; grep -E 'implement.md|review-panel.md' scripts/check-contract-budget.sh @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept -->

**Verification of the diffs below.** All four code diffs were authored for this plan, applied
together to a scratch copy of this worktree, and run there: `test-gather-dispatch-context.sh`
prints 83 `ok:` lines, `test-lib-change-plan.sh` 32, and `check-vocabulary.sh`,
`check-references.sh`, `check-guard-symlinks.sh` and `check-contract-budget.sh` all exit 0.
<!-- measured: scratch copy of the worktree with tasks 1-4 applied: scripts/test-gather-dispatch-context.sh | grep -c '^ok' -> 83; scripts/test-lib-change-plan.sh | grep -c '^ok' -> 32; the four guards' exit codes @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept -->
`check-normative-inventory.sh` output is byte-identical before and after the two skill-file
diffs — the new prose adds no `SHALL`/`MUST` sentence.
<!-- measured: scripts/check-normative-inventory.sh captured in the worktree and in the scratch copy with the skill diffs applied; diff of the two captures is empty @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept -->

---

- [x] 1. `change-plan.sh` gains `change_plan_ref`, the satellite link's `<peer>:<change-id>`

`gather-dispatch-context.sh` (task 2) labels a remotely-resolved plan section with the ref its
link carries. The ref is parsed by the library that owns `link.md`'s grammar, not re-parsed by
the caller: `change_plan_ref <worktree> <change-name>` prints the `## Part of` code span and
returns 0 when the change is a satellite by the definition every caller shares — no local
`tasks.md`, a `link.md` carrying `## Part of` — and returns 1 with no output when the plan is
local or the link is not a satellite's. The ref is deliberately not re-validated inside the
helper: a caller uses it to label a resolution `change_plan_dir` has already succeeded at, and
that call has already allowlisted both halves of the same string. It is a stdout-returning
function rather than a shared global on purpose: every existing caller consumes
`change_plan_path`/`change_plan_dir` through a command substitution, and a global set inside
one dies with that subshell.

**Files:** `scripts/lib/change-plan.sh`, `scripts/test-lib-change-plan.sh`
**Tests:** `scripts/test-lib-change-plan.sh` — case 7, case 8, case 8b, case 9
**Regression:** reverting this commit removes `change_plan_ref` — cases 7 through 9 fail with
`change_plan_ref: command not found`, and task 2's gatherer loses the source of its
`(canonical …)` labels.
**Baseline:** before=24 after=32 `ok:` lines in `scripts/test-lib-change-plan.sh`
<!-- measured: scripts/test-lib-change-plan.sh | grep -c '^ok' @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept (before, worktree as-is; after, scratch copy with this diff applied) -->
**Commit:** `feat(scripts): add change_plan_ref to change-plan.sh`
**Build:** green

  - [x] **Step 1: Apply the diff** — from the repository root, `git apply` the block below verbatim
    (it touches the two files in `**Files:**` and nothing else).

```diff verified:authored for this plan, applied to a scratch copy of the worktree together with tasks 2-4's diffs and run — test-lib-change-plan.sh 32 ok lines and all cases passed, test-gather-dispatch-context.sh 83, check-guard-symlinks.sh exit 0
--- a/scripts/lib/change-plan.sh
+++ b/scripts/lib/change-plan.sh
@@ -218,6 +218,34 @@
     return 0
   fi
   return 1
+}
+
+# change_plan_ref <worktree> <change-name>
+#
+# Prints the `<peer>:<change-id>` ref of the change's link.md and returns 0
+# when the change is a satellite by the definition every caller shares — no
+# local tasks.md, a link.md carrying `## Part of` — and returns 1 with no
+# output when the plan is local or the change carries no readable link.
+# The ref is deliberately NOT re-validated here: a caller uses it to LABEL
+# a resolution change_plan_dir has already succeeded at, and that call has
+# already applied the allowlist to both halves of this exact string.
+# gather-dispatch-context.sh (KAN-393) reads it for its
+# "(canonical <peer>:<change-id>)" section labels — the ref parsed here, in
+# the library that owns link.md's grammar, rather than re-parsed by every
+# caller that needs it.
+change_plan_ref() {
+  local worktree="$1" name="$2" spec_root link ref
+  _change_plan_name_ok "$name" || return 1
+  spec_root="$(spec_root_leaf "$worktree")"
+  if [ -f "$worktree/$spec_root/changes/$name/tasks.md" ]; then
+    return 1
+  fi
+  link="$worktree/$spec_root/changes/$name/link.md"
+  if [ ! -f "$link" ]; then
+    return 1
+  fi
+  ref="$(_change_plan_link_part_of "$link")" || return 1
+  printf '%s\n' "$ref"
 }
 
 # change_plan_path <worktree> <change-name> [canonical-worktree]
--- a/scripts/test-lib-change-plan.sh
+++ b/scripts/test-lib-change-plan.sh
@@ -370,6 +370,45 @@
 assert_eq "case 6c: it prints nothing to stdout" "" "$OUT"
 
 # ---------------------------------------------------------------------------
+# Cases 7-9 (KAN-393): change_plan_ref — the link's `<peer>:<change-id>`
+# for a satellite (no local tasks.md, a link.md carrying ## Part of), and
+# return 1 with no output for a local plan or a link.md that is not a
+# satellite's. gather-dispatch-context.sh reads it for its
+# "(canonical <peer>:<change-id>)" section labels. Asserted here, in the
+# library's own harness, because the ref is the library's contract, not
+# any caller's.
+# ---------------------------------------------------------------------------
+set +e
+OUT="$(change_plan_ref "$PLAIN" "plain-change" 2>/dev/null)"
+RC=$?
+set -e
+assert_nonzero_rc "case 7: a local plan has no ref" "$RC"
+assert_eq "case 7: it prints nothing" "" "$OUT"
+
+set +e
+OUT="$(change_plan_ref "$SAT2" "sat-change" 2>/dev/null)"
+RC=$?
+set -e
+assert_zero_rc "case 8: a satellite's ref resolves at exit 0" "$RC"
+assert_eq "case 8: it prints the link's <peer>:<change-id>" \
+  "peerx:canon-change" "$OUT"
+
+set +e
+OUT="$(change_plan_ref "$SAT3" "sat-change" 2>/dev/null)"
+RC=$?
+set -e
+assert_zero_rc "case 8b: the peers-only satellite's ref resolves too" "$RC"
+assert_eq "case 8b: it prints that link's <peer>:<change-id>" \
+  "peery:canon-change" "$OUT"
+
+set +e
+OUT="$(change_plan_ref "$CANON5" "canon-only-satellite-dir" 2>/dev/null)"
+RC=$?
+set -e
+assert_nonzero_rc "case 9: a ## Parts-only link.md is not a satellite and has no ref" "$RC"
+assert_eq "case 9: it prints nothing" "" "$OUT"
+
+# ---------------------------------------------------------------------------
 if [ "$FAILURES" -eq 0 ]; then
   printf '\n✓ PASS\n'
   exit 0
```

  - [x] **Step 2: Run the harness** — `scripts/test-lib-change-plan.sh`; expected: 32 `ok:` lines
    and `all cases passed` (the after= count above). `scripts/test-gather-dispatch-context.sh`
    stays at its before= count: this diff touches nothing it reads.
  - [x] **Step 3: Commit** with the subject above.

- [x] 2. `gather-dispatch-context.sh` reads the plan from the content directory

The change's three plan leaves move from the argument `<change-root>` to a CONTENT DIRECTORY
resolved once before any leaf is read: the change-root itself whenever it carries a `tasks.md`
(every single-repo change — byte-for-byte identical bundles, so `SKIP-WHEN-UNCHANGED` hashes
hold), otherwise the canonical change directory `change_plan_dir` resolves through the new
optional seventh argument `[canonical-worktree]`, or — only when that argument is empty —
through `<worktree>/<spec-root>/peers`, the same resolution order KAN-363 gave
`check-unfinished-work.sh` and `check-task-commit-fields.sh`. Three consequences, each
asserted below: the leaves' `within_root` boundary moves with the resolution (the canonical
directory is the same change's tracked content, allowlisted by the lib); a satellite whose
canonical plan cannot be reached is skipped with a distinct
`(satellite plan unresolved — link.md at <path>)` label at exit 0 — the deliberate inverse of
`check-unfinished-work.sh`'s refusal for the same condition, because the bundle never gates a
run and has no verdict to give, while a change with no `tasks.md` and no `link.md` keeps the
ordinary absent-leaf handling; and resolved sections carry a `(canonical <peer>:<change-id>)`
label, composed with the existing `(scoped to task(s) …)` suffix. `project commands`,
`incidents` and the `head:` sha keep resolving from the given worktree — a satellite
implementer's bundle carries its own repository's commands, its own project's incident log,
and the canonical plan.

**Files:** `scripts/gather-dispatch-context.sh`, `scripts/test-gather-dispatch-context.sh`
**Tests:** `scripts/test-gather-dispatch-context.sh` — CASE 50, CASE 51, CASE 52, CASE 53,
CASE 54, CASE 55, CASE 56, CASE 57
**Regression:** reverting this commit — cases 50 through 53 fail (no canonical sections,
labels or bodies; case 53 reports the ordinary `(absent)` skips instead of the distinct
unresolved label), case 56 fails (the canonical dir's escaping leaf is skipped, not refused),
and the existing 75 cases still pass, which is the point: single-repo behavior is unchanged.
**Baseline:** before=75 after=83 `ok:` lines in `scripts/test-gather-dispatch-context.sh`
<!-- measured: scripts/test-gather-dispatch-context.sh | grep -c '^ok' @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept (before, worktree as-is; after, scratch copy with tasks 1-2 applied) -->
**Commit:** `feat(scripts): resolve a satellite's canonical plan in gather-dispatch-context.sh`
**Build:** green

  - [x] **Step 1: Apply the diff** — from the repository root, `git apply` the block below verbatim
    (it touches the two files in `**Files:**` and nothing else; task 1 is already committed).

```diff verified:authored for this plan, applied to a scratch copy of the worktree together with tasks 1, 3 and 4's diffs and run — test-gather-dispatch-context.sh 83 ok lines and all cases passed, test-lib-change-plan.sh 32, check-vocabulary.sh, check-references.sh and check-contract-budget.sh all exit 0
--- a/scripts/gather-dispatch-context.sh
+++ b/scripts/gather-dispatch-context.sh
@@ -6,7 +6,7 @@
 # proposal.md, design.md, tasks.md and the engineering principles on its
 # own.
 #
-# Usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids>]
+# Usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree>]]
 #
 # <task-ids> is optional: a comma-separated list of integer task ids. When
 # given, the "## tasks.md" section carries only the plan's header (every
@@ -123,6 +123,57 @@
 # bundle, reported as "refused", but the run still exits 0 — never treated
 # as a missing source, so an attack is never indistinguishable from a
 # legitimate absence. See add_fixed_source() below.
+#
+# THE PLAN LIVES IN ONE TREE (KAN-393). proposal.md, design.md and tasks.md
+# are read from the CONTENT DIRECTORY, resolved once before any leaf is
+# read: the <change-root> itself whenever it carries a tasks.md — every
+# single-repo change, whose bundle is therefore byte-for-byte what this
+# script always wrote — and otherwise, when <change-root> carries a
+# satellite's link.md (## Part of) and no tasks.md, the canonical change
+# directory scripts/lib/change-plan.sh resolves through the optional
+# seventh argument <canonical-worktree>, or — only when that argument is
+# empty — through <worktree>/<spec-root>/peers, the same resolution order
+# check-unfinished-work.sh and check-task-commit-fields.sh have used since
+# KAN-363. The caller (skills/flow/implement.md's per-bundle gather,
+# skills/flow/review-panel.md's rebuild) passes the canonical worktree on
+# every call: the member of the run's resolved worktree set whose own
+# <project>/<spec-root>/changes/<name>/tasks.md exists, inert on a
+# single-repo change. Three consequences, each deliberate:
+#
+#   1. The boundary the three leaves are checked against is the CONTENT
+#      DIRECTORY, not the argument <change-root>: the canonical directory is
+#      the same change's tracked content, reached through a link whose peer
+#      name and change id change-plan.sh allowlist-checks character for
+#      character. The invocation-level containment rules above are
+#      unchanged — <change-root> must still sit inside <worktree>, exit 2.
+#   2. A satellite whose canonical plan cannot be reached is NOT a refusal:
+#      the three plan leaves are reported as skipped with the distinct
+#      label "(satellite plan unresolved — link.md at <path>)" and the run
+#      still exits 0. This is the deliberate inverse of
+#      check-unfinished-work.sh's exit-2 refusal for the same condition — a
+#      guard verdict must never silently clear, but this bundle has no
+#      verdict to give: it never gates a run (skills/flow/implement.md), so
+#      refusing here would abort a dispatching stage a bundle has no
+#      authority to stop. A change with no tasks.md and no link.md at all
+#      is not a satellite, and its absent leaves are still reported exactly
+#      as before.
+#   3. When the plan resolved remotely, the three plan sections' labels
+#      carry a "(canonical <peer>:<change-id>)" suffix — composed with the
+#      existing "(scoped to task(s) ...)" suffix where both apply — so a
+#      dispatch reading the bundle can see the plan came from another
+#      tree. Everything else stays worktree-local: "project commands" from
+#      <worktree>/.flow/project.md, the incidents section from
+#      `flow record incidents -C <worktree>`, and the header's head: sha
+#      from <worktree>'s own HEAD — a satellite implementer's bundle
+#      carries its own repository's commands, its own project's incident
+#      log, and the canonical plan.
+#
+# <canonical-worktree>, when given, is validated nowhere beyond what
+# change-plan.sh's own [ -f ] probes imply — the same trust level as
+# check-unfinished-work.sh's third argument: a caller-supplied worktree
+# path, not attacker-influenced change content, and every name concatenated
+# under it (the canonical change id) is allowlist-checked inside
+# change-plan.sh itself.
 set -euo pipefail
 
 # resolve_file <path> -> the path's resolved PHYSICAL location, following
@@ -143,6 +194,13 @@
 source "$SCRIPT_DIR/lib/lexical-normalize.sh"
 source "$SCRIPT_DIR/lib/sha256-hex.sh"
 source "$SCRIPT_DIR/lib/project-section.sh"
+# change_plan_dir <worktree> <change-name> [canonical-worktree] -> the
+# canonical plan's directory, and change_plan_ref <worktree> <change-name>
+# -> the link's <peer>:<change-id> when the change is a satellite (KAN-393)
+# — the label's content, parsed in the library that owns link.md's grammar.
+# The same lib check-unfinished-work.sh and check-task-commit-fields.sh
+# source; sourcing it pulls in spec-root.sh in turn, by its own design.
+source "$SCRIPT_DIR/lib/change-plan.sh"
 
 WORKTREE="${1:-}"
 CHANGE_ROOT="${2:-}"
@@ -150,9 +208,10 @@
 PRINCIPLES_PATH="${4:-}"
 OUTPUT_PATH="${5:-}"
 TASK_IDS="${6:-}"
+CANONICAL_WORKTREE="${7:-}"
 
 if [ -z "$WORKTREE" ] || [ -z "$CHANGE_ROOT" ] || [ -z "$NAME" ] || [ -z "$PRINCIPLES_PATH" ] || [ -z "$OUTPUT_PATH" ]; then
-  echo "usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids>]" >&2
+  echo "usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree>]]" >&2
   exit 2
 fi
 
@@ -276,10 +335,39 @@
 # The bundle itself
 # ===========================================================================
 
-PROPOSAL_FILE="$CHANGE_ROOT_REAL/proposal.md"
-DESIGN_FILE="$CHANGE_ROOT_REAL/design.md"
-TASKS_FILE="$CHANGE_ROOT_REAL/tasks.md"
+# --- the content directory (KAN-393): where this change's plan leaves
+# actually live. The <change-root> itself whenever it carries a tasks.md —
+# the single-repo case, byte-for-byte unchanged — and otherwise, when it
+# carries a satellite's link.md, the canonical change directory
+# change_plan_dir resolves through the seventh argument or, only when that
+# argument is empty, through <worktree>/<spec-root>/peers. When even that
+# fails, the change IS a satellite (its link.md carries ## Part of — the
+# same definition check-unfinished-work.sh applies) and the three plan
+# leaves below are reported as skips with the distinct unresolved label;
+# a change with neither file is not a satellite, and its leaves keep the
+# ordinary absent-leaf handling. CANONICAL_REF is change_plan_ref's answer:
+# the <peer>:<change-id> ref of the link the successful resolution already
+# validated, feeding the section label.
+CONTENT_DIR_REAL="$CHANGE_ROOT_REAL"
+CANONICAL_SUFFIX=""
+CANONICAL_LINK=""
+if [ ! -f "$CHANGE_ROOT_REAL/tasks.md" ]; then
+  if PLAN_DIR="$(change_plan_dir "$WORKTREE_REAL" "$NAME" "$CANONICAL_WORKTREE" 2>/dev/null)" && [ -n "$PLAN_DIR" ]; then
+    CONTENT_DIR_REAL="$PLAN_DIR"
+    CANONICAL_REF="$(change_plan_ref "$WORKTREE_REAL" "$NAME" 2>/dev/null)" || CANONICAL_REF=""
+    if [ -n "$CANONICAL_REF" ]; then
+      CANONICAL_SUFFIX=" (canonical ${CANONICAL_REF})"
+    fi
+  elif [ -f "$CHANGE_ROOT_REAL/link.md" ] && grep -qxF '## Part of' "$CHANGE_ROOT_REAL/link.md" 2>/dev/null; then
+    CONTENT_DIR_REAL=""
+    CANONICAL_LINK="$CHANGE_ROOT_REAL/link.md"
+  fi
+fi
 
+PROPOSAL_FILE="$CONTENT_DIR_REAL/proposal.md"
+DESIGN_FILE="$CONTENT_DIR_REAL/design.md"
+TASKS_FILE="$CONTENT_DIR_REAL/tasks.md"
+
 FOUND_LABELS=()
 FOUND_PATHS=()
 SKIPPED_LABELS=()
@@ -302,7 +390,7 @@
 # still let the run exit 0 — only the printed diagnostic differs.
 REFUSED_REASONS=()
 
-# add_fixed_source <path> <label> — unlike <worktree>/<change-root>/
+# add_fixed_source <path> <label> <root> — unlike <worktree>/<change-root>/
 # <principles-path> above, these three content sources are never passed as
 # arguments, so validate_path()'s exit-2-on-mismatch contract does not apply
 # to them: a change legitimately carries no design.md, and "this leaf is a
@@ -310,19 +398,20 @@
 # gather-self-review-context.sh's check_boundary() + is_refused() pair: `[ -f
 # "$path" ]` first (a missing target, including a dangling symlink, is a
 # plain absence — skipped, not refused); then resolve_file() (leaf and every
-# ancestor symlink) and a within_root() boundary check against
-# CHANGE_ROOT_REAL, the same containment root that script uses for its own
-# change-relative leaf (tasks.md). A leaf that resolves inside change-root —
-# including a symlink to another file inside it — is read normally. A leaf
-# that resolve_file() cannot even walk (REFUSED_REASONS' own "unresolvable"),
-# or that resolves OUTSIDE change-root ("outside"), is refused per-source:
-# omitted from the bundle, counted separately from "skipped", and the run
-# still exits 0 — the spec's "a missing bundle never stops a run"
-# requirement, and the same disposition gather-self-review-context.sh's
+# ancestor symlink) and a within_root() boundary check against the third
+# argument — the CONTENT DIRECTORY (KAN-393): the change-root itself for a
+# local plan, the canonical change directory for a satellite, so a leaf can
+# never resolve outside the change whose content it is. A leaf that resolves
+# inside that root — including a symlink to another file inside it — is read
+# normally. A leaf that resolve_file() cannot even walk (REFUSED_REASONS' own
+# "unresolvable"), or that resolves OUTSIDE it ("outside"), is refused
+# per-source: omitted from the bundle, counted separately from "skipped",
+# and the run still exits 0 — the spec's "a missing bundle never stops a
+# run" requirement, and the same disposition gather-self-review-context.sh's
 # header documents for exactly this shape of problem (a per-source
 # trust-boundary violation is not a malformed invocation).
 add_fixed_source() {
-  local path="$1" label="$2" resolved
+  local path="$1" label="$2" root="$3" resolved
   if [ ! -f "$path" ]; then
     SKIPPED_LABELS+=("$label (absent)")
     return 0
@@ -332,7 +421,7 @@
     REFUSED_REASONS+=("unresolvable")
     return 0
   }
-  if within_root "$resolved" "$CHANGE_ROOT_REAL"; then
+  if within_root "$resolved" "$root"; then
     FOUND_LABELS+=("$label")
     FOUND_PATHS+=("$resolved")
   else
@@ -376,19 +465,29 @@
   ' "$1"
 }
 
-add_fixed_source "$PROPOSAL_FILE" "proposal.md"
-add_fixed_source "$DESIGN_FILE" "design.md"
-add_fixed_source "$TASKS_FILE" "tasks.md"
+if [ -n "$CONTENT_DIR_REAL" ]; then
+  add_fixed_source "$PROPOSAL_FILE" "proposal.md${CANONICAL_SUFFIX}" "$CONTENT_DIR_REAL"
+  add_fixed_source "$DESIGN_FILE" "design.md${CANONICAL_SUFFIX}" "$CONTENT_DIR_REAL"
+  add_fixed_source "$TASKS_FILE" "tasks.md${CANONICAL_SUFFIX}" "$CONTENT_DIR_REAL"
+else
+  # A satellite whose canonical plan could not be reached: the bundle is
+  # advisory and never gates a run, so this is a skip, not a refusal —
+  # with its own label, so "the plan is elsewhere and unreachable" never
+  # reads like a plain change's legitimately-absent design.md (KAN-393).
+  for label in proposal.md design.md tasks.md; do
+    SKIPPED_LABELS+=("$label (satellite plan unresolved — link.md at $CANONICAL_LINK)")
+  done
+fi
 
 TASKS_SCOPED_BODY=""
 if [ -n "$TASK_IDS" ]; then
   last=$(( ${#FOUND_LABELS[@]} - 1 ))
-  if [ "$last" -lt 0 ] || [ "${FOUND_LABELS[$last]}" != "tasks.md" ]; then
+  if [ "$last" -lt 0 ] || [ "${FOUND_LABELS[$last]}" != "tasks.md${CANONICAL_SUFFIX}" ]; then
     echo "gather-dispatch-context: task ids given but tasks.md is absent or refused" >&2
     exit 2
   fi
   TASKS_SCOPED_BODY="$(scope_tasks "${FOUND_PATHS[$last]}" "$TASK_IDS")" || exit 2
-  FOUND_LABELS[$last]="tasks.md (scoped to task(s) $TASK_IDS)"
+  FOUND_LABELS[$last]="tasks.md${CANONICAL_SUFFIX} (scoped to task(s) $TASK_IDS)"
   FOUND_PATHS[$last]="@scoped-tasks"
 fi
 
--- a/scripts/test-gather-dispatch-context.sh
+++ b/scripts/test-gather-dispatch-context.sh
@@ -187,6 +187,47 @@
 exit 1
 STUB
   chmod +x "$dest/flow"
+}
+
+# capture7 <task-ids> <canonical-worktree> -> RC, ERR, OUT — the seven-
+# argument invocation (kan-393). The optional sixth argument scopes the
+# bundle's ## tasks.md section; the optional seventh names the canonical
+# worktree a satellite's plan resolves through.
+capture7() {
+  set +e
+  ERR="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" demo "$PRINCIPLES" "$OUTPUT_PATH" "$1" "$2" 2>&1 1>/dev/null)"
+  RC=$?
+  set -e
+  if [ -f "$OUTPUT_PATH" ]; then
+    OUT="$(cat "$OUTPUT_PATH")
+$ERR"
+  else
+    OUT="$ERR"
+  fi
+}
+
+# new_satellite_pair -> REPO becomes a SATELLITE worktree (its change dir
+# carries ONLY link.md — KAN-363's pointer tree), CANON_REPO the canonical
+# one carrying the real plan under the same change id, and the satellite's
+# peers file declares the canonical at a path that does NOT resolve, so the
+# canonical-ARGUMENT branch is the one under test and never an accidental
+# peers hit. CANON_CHANGE_ROOT is the canonical change directory.
+# OUTPUT_PATH (new_repo's) already points inside the satellite — the
+# production shape: a satellite implementer's bundle.
+new_satellite_pair() {
+  new_repo
+  rm -f "$CHANGE_ROOT/proposal.md" "$CHANGE_ROOT/design.md" "$CHANGE_ROOT/tasks.md"
+  printf '## Part of\n\n`canon:demo`\n' > "$CHANGE_ROOT/link.md"
+  CANON_REPO="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-canon.XXXXXX")"
+  CANON_REPO="$(cd -P "$CANON_REPO" && pwd -P)"
+  TREES+=("$CANON_REPO")
+  CANON_CHANGE_ROOT="$CANON_REPO/spectre/changes/demo"
+  mkdir -p "$CANON_CHANGE_ROOT"
+  printf 'CANON-PROPOSAL-BODY\n' > "$CANON_CHANGE_ROOT/proposal.md"
+  printf 'CANON-DESIGN-BODY\n' > "$CANON_CHANGE_ROOT/design.md"
+  printf 'CANON-TASKS-BODY\n' > "$CANON_CHANGE_ROOT/tasks.md"
+  mkdir -p "$REPO/spectre"
+  printf 'canon ../canon-nowhere\n' > "$REPO/spectre/peers"
 }
 
 # ===========================================================================
@@ -1307,7 +1348,176 @@
   *"skipped: $PRINCIPLES_49B (absent)"*) pass "principles-path shaped 'foo(bar)': absent suffix appended" ;;
   *) fail "principles-path shaped 'foo(bar)': absent suffix missing: $OUT" ;;
 esac
+
+# ===========================================================================
+# CASES 50-57 (kan-393): the plan lives in one tree. A satellite worktree's
+# bundle carries the canonical proposal/design/tasks (labeled), while
+# project commands, incidents and the head sha stay the satellite's own.
+# ===========================================================================
+
+# CASE 50: the canonical-ARGUMENT branch. A satellite bundle carries the
+# canonical plan under labeled sections, and the satellite's OWN project
+# commands and head sha — never the canonical repository's.
+new_satellite_pair
+mkdir -p "$REPO/.flow" "$CANON_REPO/.flow"
+printf '# sat project\n\n## lint\n\nSAT-LINT-MARKER\n' > "$REPO/.flow/project.md"
+printf '# canon project\n\n## lint\n\nCANON-LINT-MARKER\n' > "$CANON_REPO/.flow/project.md"
+SAT_SHA_50="$(git -C "$REPO" rev-parse --short HEAD)"
+capture7 "" "$CANON_REPO"
+if [ "$RC" -ne 0 ]; then
+  fail "satellite via argument: exited $RC: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF '## proposal.md (canonical canon:demo)'; then
+  fail "satellite via argument: proposal section unlabeled or missing: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF '## design.md (canonical canon:demo)'; then
+  fail "satellite via argument: design section unlabeled or missing: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF '## tasks.md (canonical canon:demo)'; then
+  fail "satellite via argument: tasks section unlabeled or missing: $OUT"
+elif ! printf '%s' "$OUT" | grep -q 'CANON-PROPOSAL-BODY' \
+  || ! printf '%s' "$OUT" | grep -q 'CANON-DESIGN-BODY' \
+  || ! printf '%s' "$OUT" | grep -q 'CANON-TASKS-BODY'; then
+  fail "satellite via argument: a canonical plan body is missing: $OUT"
+elif ! printf '%s' "$OUT" | grep -q 'SAT-LINT-MARKER'; then
+  fail "satellite via argument: the satellite's own project commands are missing: $OUT"
+elif printf '%s' "$OUT" | grep -q 'CANON-LINT-MARKER'; then
+  fail "satellite via argument: the CANONICAL project commands leaked in: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF "head: $SAT_SHA_50"; then
+  fail "satellite via argument: head sha is not the satellite's own: $OUT"
+else
+  pass "satellite via argument: canonical plan labeled in, satellite commands and head kept"
+fi
+
+# CASE 51: the canonical and scoped labels compose. The plan fixture lives
+# in the CANONICAL change dir here (a satellite has no tasks.md of its own).
+new_satellite_pair
+FENCE_51='```'
+printf '%s\n' \
+  '# demo plan' '' '> **Execution:** header line' '' \
+  '- [ ] 1. First' '**Files:** `a`' '  - [ ] **Step 1: one**' '' \
+  '- [ ] 2. Second' '**Files:** `b`' > "$CANON_CHANGE_ROOT/tasks.md"
+capture7 "1" "$CANON_REPO"
+if [ "$RC" -ne 0 ]; then
+  fail "satellite scoped: exited $RC: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF '## tasks.md (canonical canon:demo) (scoped to task(s) 1)'; then
+  fail "satellite scoped: the two labels do not compose: $OUT"
+elif ! printf '%s' "$OUT" | grep -q '^- \[ \] 1\. First$' \
+  || printf '%s' "$OUT" | grep -q 'Second'; then
+  fail "satellite scoped: scoping broke under the canonical label: $OUT"
+else
+  pass "satellite scoped: canonical and scoped labels compose"
+fi
+
+# CASE 52: the PEERS branch — no canonical-worktree argument, the peer path
+# declared in the satellite's peers file really resolves on disk (the
+# main-checkout shape, where ../peer exists).
+new_repo
+rm -f "$CHANGE_ROOT/proposal.md" "$CHANGE_ROOT/design.md" "$CHANGE_ROOT/tasks.md"
+printf '## Part of\n\n`peerc:demo`\n' > "$CHANGE_ROOT/link.md"
+mkdir -p "$REPO/spectre"
+printf 'peerc ../gather-canon-peer\n' > "$REPO/spectre/peers"
+CANON_PEER_52="$REPO/../gather-canon-peer"
+mkdir -p "$CANON_PEER_52/spectre/changes/demo"
+printf 'PEER-PROPOSAL-BODY\n' > "$CANON_PEER_52/spectre/changes/demo/proposal.md"
+printf 'PEER-DESIGN-BODY\n' > "$CANON_PEER_52/spectre/changes/demo/design.md"
+printf 'PEER-TASKS-BODY\n' > "$CANON_PEER_52/spectre/changes/demo/tasks.md"
+TREES+=("$CANON_PEER_52")
+capture7 "" ""
+if [ "$RC" -ne 0 ]; then
+  fail "satellite via peers: exited $RC: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF '## tasks.md (canonical peerc:demo)'; then
+  fail "satellite via peers: tasks section unlabeled or missing: $OUT"
+elif ! printf '%s' "$OUT" | grep -q 'PEER-TASKS-BODY'; then
+  fail "satellite via peers: the peer plan body is missing: $OUT"
+else
+  pass "satellite via peers: resolves and labels without the argument"
+fi
 
+# CASE 53: an UNRESOLVABLE satellite — the peer is declared but absent and
+# no argument is passed. The three plan leaves are skipped with the distinct
+# label, the run still exits 0, and the bundle is still written: the bundle
+# never gates a run, so this is the inverse of check-unfinished-work.sh's
+# refusal, deliberately.
+new_satellite_pair
+capture7 "" ""
+if [ "$RC" -ne 0 ]; then
+  fail "unresolvable satellite: exited $RC (expected 0): $OUT"
+elif ! printf '%s' "$OUT" | grep -qxF "skipped: proposal.md (satellite plan unresolved — link.md at $CHANGE_ROOT/link.md)" \
+  || ! printf '%s' "$OUT" | grep -qxF "skipped: design.md (satellite plan unresolved — link.md at $CHANGE_ROOT/link.md)" \
+  || ! printf '%s' "$OUT" | grep -qxF "skipped: tasks.md (satellite plan unresolved — link.md at $CHANGE_ROOT/link.md)"; then
+  fail "unresolvable satellite: the three distinct skip labels are missing: $OUT"
+elif printf '%s' "$OUT" | grep -q 'CANON-'; then
+  fail "unresolvable satellite: canonical content reached the bundle anyway: $OUT"
+else
+  pass "unresolvable satellite: three distinct skips, exit 0, bundle written"
+fi
+
+# CASE 54: task ids given while the satellite is unresolvable — still exit 2
+# naming the absence, exactly as when tasks.md was merely missing.
+new_satellite_pair
+capture7 "1" ""
+if [ "$RC" -ne 2 ]; then
+  fail "unresolvable satellite with ids: expected exit 2, got $RC: $OUT"
+elif ! printf '%s' "$ERR" | grep -q 'task ids given but tasks.md is absent or refused'; then
+  fail "unresolvable satellite with ids: stderr does not name the absence: $ERR"
+else
+  pass "unresolvable satellite with ids: exit 2 naming it"
+fi
+
+# CASE 55: a PLAIN change with the canonical-worktree argument passed anyway
+# (the uniform caller rule) is byte-identical in its labels — local wins
+# before the link is ever consulted, and no canonical suffix appears.
+new_repo
+capture7 "" "$REPO"
+if [ "$RC" -ne 0 ]; then
+  fail "plain change with argument: exited $RC: $OUT"
+elif printf '%s' "$OUT" | grep -q '(canonical'; then
+  fail "plain change with argument: a canonical suffix appeared on a local plan: $OUT"
+elif ! printf '%s' "$OUT" | grep -qxF '## proposal.md' \
+  || ! printf '%s' "$OUT" | grep -qxF '## design.md' \
+  || ! printf '%s' "$OUT" | grep -qxF '## tasks.md'; then
+  fail "plain change with argument: a section label changed shape: $OUT"
+else
+  pass "plain change with the argument: labels byte-identical, no suffix"
+fi
+
+# CASE 56: the boundary moved with the resolution. The CANONICAL change
+# dir's proposal.md is a symlink resolving outside the CANONICAL change
+# dir: refused per-source against the content directory, content omitted,
+# exit 0 — the same disposition, against the moved root.
+new_satellite_pair
+OUTSIDE_56="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-outside56.XXXXXX")"
+TREES+=("$OUTSIDE_56")
+printf 'TOP-SECRET-56\n' > "$OUTSIDE_56/secret.md"
+rm -f "$CANON_CHANGE_ROOT/proposal.md"
+ln -s "$OUTSIDE_56/secret.md" "$CANON_CHANGE_ROOT/proposal.md"
+capture7 "" "$CANON_REPO"
+if [ "$RC" -ne 0 ]; then
+  fail "satellite leaf escapes content dir: exited $RC (expected 0): $OUT"
+elif printf '%s' "$OUT" | grep -q 'TOP-SECRET-56'; then
+  fail "satellite leaf escapes content dir: target content leaked: $OUT"
+elif ! printf '%s' "$OUT" | grep -qF 'refused: proposal.md (canonical canon:demo) (resolves outside the change directory)'; then
+  fail "satellite leaf escapes content dir: not reported refused: $OUT"
+else
+  pass "a satellite's canonical leaf symlink outside the content dir is refused, exit 0"
+fi
+
+# CASE 57: a link.md carrying ## Parts but no ## Part of is the canonical
+# side's own copy — NOT a satellite by the definition every guard shares —
+# so a missing tasks.md there is the ordinary absence, never the satellite
+# label.
+new_repo
+rm -f "$CHANGE_ROOT/tasks.md"
+printf '## Parts\n\n`peerz:some-part`\n' > "$CHANGE_ROOT/link.md"
+run_it
+if [ "$RC" -ne 0 ]; then
+  fail "## Parts-only link.md: exited $RC: $OUT"
+elif ! printf '%s' "$OUT" | grep -qxF 'skipped: tasks.md (absent)'; then
+  fail "## Parts-only link.md: tasks.md not reported as the ordinary absence: $OUT"
+elif printf '%s' "$OUT" | grep -q 'satellite plan unresolved'; then
+  fail "## Parts-only link.md: wrongly treated as a satellite: $OUT"
+else
+  pass "a ## Parts-only link.md is not a satellite: ordinary absent-leaf handling"
+fi
+
 if [ "$FAILURES" -ne 0 ]; then
   printf '%s case(s) failed\n' "$FAILURES" >&2
   exit 1
```

  - [x] **Step 2: Run the harness and the sibling guard** — `scripts/test-gather-dispatch-context.sh`;
    expected: 83 `ok:` lines and `all cases passed` (the after= count above), every pre-existing
    case among them. Then `scripts/check-guard-symlinks.sh`; expected: exit 0 — this diff adds
    the `$SCRIPT_DIR/lib/change-plan.sh` source, whose sibling the rule derives from the source
    grep, and the shipped farm carries the script through the single `lib` symlink.
  - [x] **Step 3: Commit** with the subject above.

- [x] 3. `implement.md`'s gather step passes the canonical worktree

The per-bundle gather call gains the seventh argument and one sentence defining it: the member
of this run's resolved worktree set whose own `<project>/<spec-root>/changes/<name>/tasks.md`
exists — the same argument `check-unfinished-work.sh` takes at the integrate gate, passed on
every call and inert on a single-repo change. Without it, a satellite worktree's bundle is back
to `skipped: tasks.md (absent)` and the satellite implementer is dispatched with no plan.

**Files:** `skills/flow/implement.md`
**Tests:** **none** — prose; `scripts/check-contract-budget.sh` checks the file's row,
`scripts/check-vocabulary.sh` and `scripts/check-references.sh` the text, and
`check-normative-inventory.sh` output stays byte-identical.
**Regression:** reverting this commit leaves the skill telling the conductor to gather without
the canonical worktree — a satellite worktree's dispatch bundle loses the plan sections again,
the defect this change closes; no guard fires on the revert, which is why the added sentence
names `scripts/gather-dispatch-context.sh`'s header as canonical for the resolution.
**Baseline:** before=71 after=71 owned files within budget in `check-contract-budget.sh`
<!-- measured: scripts/check-contract-budget.sh @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept (before, worktree as-is; after, scratch copy with all four tasks applied — implement.md 29726 bytes against its 32500-byte row) -->
**Commit:** `docs(flow): pass the canonical worktree in the bundle gather step`
**Build:** green

  - [x] **Step 1: Edit the gather step's call and add the defining sentence.** In the paragraph
    **Gather one bundle per dispatch bundle, immediately before that bundle's implementer goes
    out.**, the fenced call's second line currently ends `<id>[,<id>…]`; append the seventh
    argument so it reads:

```sh verified:authored for this plan and applied verbatim to a scratch copy of the worktree together with tasks 1, 2 and 4's edits; check-contract-budget.sh, check-vocabulary.sh and check-references.sh exit 0 there
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-bundle-<k>.md <id>[,<id>…] <canonical-worktree>
```

    Then, immediately after the sentence ending "`skills/flow/`, always." (before the "A
    non-zero exit" sentence), insert this paragraph:

```markdown verified:applied verbatim to the same scratch copy; this task's two edits add 609 bytes to the file in total, inside its 32500-byte budget row, measured there
`<canonical-worktree>` — the seventh argument, passed on every call — is
the member of this run's resolved worktree set whose own
`<project>/<spec-root>/changes/<name>/tasks.md` exists, the same argument
`check-unfinished-work.sh` takes at the integrate gate; on a single-repo change that member is
this worktree and the argument is inert, while on a satellite worktree's bundle it carries the
canonical plan under labeled sections while keeping this worktree's own project commands,
incidents and HEAD (`scripts/gather-dispatch-context.sh`'s header is canonical for the
resolution).
```

    Keep every other sentence of the step verbatim.

  - [x] **Step 2: Run the prose guards** — capture `scripts/check-normative-inventory.sh` output
    before and after the edit and diff the two captures (expected: empty); then
    `scripts/check-contract-budget.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`; expected: all exit 0.
  - [x] **Step 3: Commit** with the subject above.

- [x] 4. `review-panel.md`'s rebuild step passes the canonical worktree

The review panel rebuilds its dispatch-context bundle at the start of every round with the
five-argument call; the same seventh argument is appended and the same one-sentence definition
added, so a round reviewing a satellite's diff reads a bundle that carries the canonical plan.
The panel itself already runs in the canonical worktree — the argument is inert there today —
so this task is uniformity with `implement.md`'s rule, not a behavior change on its own.

**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — prose; same guard set as task 3.
**Regression:** reverting this commit leaves the panel's rebuild call without the canonical
worktree, so the two skill files disagree about the calling convention and a panel run against
a satellite-affecting change resumed from a non-canonical worktree is back to a plan-less
bundle.
**Baseline:** before=71 after=71 owned files within budget in `check-contract-budget.sh`
<!-- measured: scripts/check-contract-budget.sh @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept (before, worktree as-is; after, scratch copy with all four tasks applied — review-panel.md 50032 bytes against its 58732-byte row) -->
**Commit:** `docs(flow): pass the canonical worktree in the bundle rebuild step`
**Build:** green

  - [x] **Step 1: Edit the rebuild step's call and add the defining sentence.** In **Rebuild the
    dispatch context bundle at the start of this stage too**, the fenced call's second line
    currently ends `dispatch-context.md`; append an empty sixth argument (the panel passes no
    task ids — the script's sixth position is `[<task-ids>]`, and a path there would be scoped
    as task ids and exit 2) and the seventh argument so it reads:

```sh verified:authored for this plan and applied verbatim to a scratch copy of the worktree together with tasks 1-3's edits; check-contract-budget.sh, check-vocabulary.sh and check-references.sh exit 0 there; empty-sixth-argument form repaired by the run's conductor at flow.load-context — the originally authored fence placed <canonical-worktree> in the script's sixth position, where scope_tasks would refuse it
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context.md "" <canonical-worktree>
```

    Then, between the fenced call and the "Report the script's stderr line" sentence, insert
    this paragraph:

```markdown verified:applied verbatim to the same scratch copy; this task's two edits add 385 bytes to the file in total, inside its 58732-byte budget row, measured there
`<canonical-worktree>` is the member of the run's resolved worktree set whose own
`<project>/<spec-root>/changes/<name>/tasks.md` exists — the same argument
`check-unfinished-work.sh` takes, passed on every call and inert when that member is this
worktree (`scripts/gather-dispatch-context.sh`'s header is canonical for what a satellite's
bundle then carries).
```

  - [x] **Step 2: Run the prose guards** — same three guards and the same
    `check-normative-inventory.sh` before/after capture diff as task 3, against
    `skills/flow/review-panel.md`; expected: all exit 0, captures identical.
  - [x] **Step 3: Commit** with the subject above.

- [x] 5. The two new `gather-dispatch-context.sh` citations gain the `<agents repo>/` root (fix 1)

`scripts/check-installed-citations.sh` exits 1 on the prose tasks 3 and 4 wrote: the backticked
citation `scripts/gather-dispatch-context.sh` names no root, once in each file. Both citations
are prefixed `<agents repo>/` — the root the guard accepts and sibling prose in both files
already uses. No behavior change; the citations' targets are unchanged.

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`
**Tests:** **none** — prose; `scripts/check-installed-citations.sh` is the failing check this
task closes, and `check-contract-budget.sh`, `check-vocabulary.sh` and `check-references.sh`
must stay at exit 0.
**Regression:** reverting this commit puts the two root-less citations back and
`check-installed-citations.sh` exits 1 again on the same two lines.
**Baseline:** before=2 after=0 violations in `scripts/check-installed-citations.sh` (exit 1 → 0)
<!-- measured: scripts/check-installed-citations.sh @ branch spectre/kan-393-flow-give-guards-a-canonical-worktree-concept (before, worktree as-is — exit 1, the two violations below; after, measured post-edit by the implementer) -->
**Commit:** `docs(flow): root the gather-dispatch-context citations`
**Build:** green

  - [x] **Step 1: Prefix both citations.** In `skills/flow/implement.md` line 323 and
    `skills/flow/review-panel.md` line 107, change the backticked citation
    `` `scripts/gather-dispatch-context.sh` `` to
    `` `<agents repo>/scripts/gather-dispatch-context.sh` `` — the parenthetical "…'s header is
    canonical" sentence each file gained in tasks 3 and 4. Touch nothing else.
  - [x] **Step 2: Run the guards** — `scripts/check-installed-citations.sh` (expected: exit 0),
    then `scripts/check-contract-budget.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh` (expected: all exit 0).
  - [x] **Step 3: Commit** with the subject above.
