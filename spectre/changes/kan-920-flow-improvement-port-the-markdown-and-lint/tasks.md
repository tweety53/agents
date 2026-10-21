# kan-920-flow-improvement-port-the-markdown-and-lint

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports three bash guards into `flow-guard` (tasks 1–3; task 1 also moves the owned-corpus walk into
one shared Go file and deletes `scripts/lib/owned-corpus.sh`), sweeps the citations of the deleted
files (task 4) and records the after parity counts (task 5). `design.md` is canonical for every
decision; each task cites its entry by ID.

**The script at `75261910` is each port's specification.** Its header comment states the
contract; its body is the behaviour. A port reproduces arguments, environment overrides, every
output line byte for byte, the stdout/stderr split, every side effect on the filesystem, and
every exit code. Where the source and its header disagree, stop and report — never pick one
silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:**
carry-prior-port-decisions — KAN-760's `parity-by-case-count`).

**Every port task (1–3) follows the same five steps:** port the harness's cases to a Go table test
first (red — the guard is unregistered), port the script, run the test green, replace the script's
body with the shim, `git rm` the harness. Each Go subtest is named after the harness case's `ok`
label, so parity is a `--- PASS` count. The script body's comments move into the Go file beside the
code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without check- and without ->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the only name
`scripts/run-guard-tests.sh`'s companion rule accepts for a `check-*` shim. Each port registers its
basename in `guard.Registry` from its own file's `init()`, as `basemoved.go` does.

**Repository root.** Each port reads its root from `Env`: the script's own `*_ROOT` override when
set (set-but-empty refused with the bash's exact message and exit 2), else `FLOW_GUARD_REPO_ROOT`,
exactly as `dispatchparagraphs.go` does for `CHECK_DISPATCH_PARAGRAPHS_ROOT`. The message the bash
printed when the root is not a directory stays byte-identical.

**Shim template** — the header comment block kept verbatim, corrected only where the port makes a
statement false (bash-specific plumbing, "resolved from this script's own location" now done by the
shim), then:

```bash verified:copied from the tail of scripts/check-dispatch-paragraphs.sh @ 75261910
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec <name> 2 "<name>:" "$@"
```

`<name>` is the basename without `.sh`, written literally.

**Git config pins.** The one git call (`check-python-suppressions`' `git ls-files`) carries the
pins R1–R3 of `scripts/test-git-config-pins.sh` require (its case 7 audits
`stats/internal/guard/*.go`) and runs through `gitExec` (`postmutationcheck.go`).

**Ordering.** `LC_ALL=C sort` is byte order: Go's `sort.Strings` on the raw lines. The bash
`sort -u` in `check-mutation-reproducer-pin`'s near-miss scan ran under the caller's locale; its
lines are ASCII-only file:line rows from one file, so byte order matches every locale the harness
runs under — the port uses `sort.Strings` plus adjacent-duplicate removal.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees are built per case in
`t.TempDir()` with the existing helpers (`helpers_test.go`, `writeFile`, `mkdir`); the root is
passed on `Env`, never through the process environment. No test reads or writes the operator's real
`$HOME`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real as
its step 5 names. The full package and `scripts/run-guard-tests.sh` run in `flow.verify`.

## Review Focus

- A shim run through a symlink (`~/.claude/skills/*/scripts/<name>.sh`) must resolve the same
  repository root the bash did — each port's step 5 runs the shim through a symlink in a temp
  directory.
- `check-normative-inventory`'s awk sentence splitter: the abbreviation carry-forward, the
  terminator class `[.!?][]"'`*_)]* +`, blockquote stripping, table-cell splitting and fence
  toggling per file (`FNR == 1` resets `in_fence`) must produce the identical inventory on the real
  corpus, not only on fixtures.
- `check-mutation-reproducer-pin`'s joined two-line scan reports only a match that spans the join
  boundary (KAN-624 F5) — an off-by-one there reintroduces phantom rows.
- `check-python-suppressions` with no arguments scans tracked `*.py` from the repository root,
  skipping symlinks; an enumeration failure is exit 2, never an empty scan set.

---

- [ ] 1. Port check-normative-inventory and share the owned-corpus walk

**Files:** `stats/internal/guard/normativeinventory.go`, `stats/internal/guard/check_normative_inventory_test.go`, `stats/internal/guard/ownedcorpus.go`, `stats/internal/guard/asksilence.go`, `stats/internal/guard/check_ask_silence_test.go`, `scripts/check-normative-inventory.sh`, `scripts/test-check-normative-inventory.sh`, `scripts/lib/owned-corpus.sh`
**Tests:** `TestCheckNormativeInventory`
**Regression:** fails if any of the harness's 20 `ok` behaviours regress; the existing
`TestAskSilenceCorpusParity`, now run against `owned-corpus.sh` at `libTwinBase`, fails if the
shared walk drifts from the bash library; the existing `TestCheckAskSilence` fails if its messages
lose their `check-ask-silence:` prefix.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_normative_inventory_test.go 2>/dev/null | grep -cE '^func Test' @ 75261910 -->
**After:** none
**Commit:** `feat(guard): port check-normative-inventory to Go`
**Build:** green

**Decision:** scope-three-guards-and-owned-corpus
**Decision:** shared-owned-corpus-walk

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-normative-inventory.sh`,
    one subtest per `ok` label (the abbreviation, reflow-invariance, duplicate-kept, empty-corpus
    and set-but-empty-override cases included). Run — expect failure.
  - [ ] **Step 2: Share the walk.** Move `asCorpus` and `askSilenceScopes` from `asksilence.go`
    into `ownedcorpus.go` as `ownedCorpus(root, prefix string, stderr io.Writer) ([]string, int)`
    and `ownedCorpusScopes`, the refusal prefix a parameter; `checkAskSilence` passes
    `askSilenceName`; the `THE CORPUS mirrors scripts/lib/owned-corpus.sh` comment moves with it,
    rewritten to say this file is now the one definition and the bash library was deleted.
    `TestAskSilenceCorpusParity`'s `bashLib(t, "owned-corpus.sh", …)` becomes
    `bashLibAt(t, nil, []string{"owned-corpus.sh"}, "owned_corpus_files", root)`.
  - [ ] **Step 3: Port**, registering `check-normative-inventory`: corpus via
    `ownedCorpus(root, "owned_corpus_files", stderr)`; the awk program's line classification,
    `emit`, `is_abbrev` and `has_keyword` reproduced; sentences sorted bytewise, duplicates kept;
    the count line on stderr byte-identical; every exit-2 message byte-identical.
  - [ ] **Step 4: Green.** `cd stats && go test ./internal/guard/ -run
    '^(TestCheckNormativeInventory|TestAskSilenceCorpusParity|TestCheckAskSilence)$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckNormativeInventory/'` — at least 20, and the run
    exits 0.
  - [ ] **Step 5: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-normative-inventory.sh scripts/lib/owned-corpus.sh`.
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-normative-inventory.sh` on this worktree prints stdout byte-identical to the
    bash at `75261910` (`git show 75261910:scripts/check-normative-inventory.sh` and
    `git show 75261910:scripts/lib/owned-corpus.sh` written into a temp `scripts/` +
    `scripts/lib/` beside a copy of the corpus, or run with `CHECK_NORMATIVE_INVENTORY_ROOT`
    pointing at this worktree) and the same exit code, directly and through a symlink to it in a
    temp directory.

- [ ] 2. Port check-mutation-reproducer-pin

**Files:** `stats/internal/guard/mutationreproducerpin.go`, `stats/internal/guard/check_mutation_reproducer_pin_test.go`, `scripts/check-mutation-reproducer-pin.sh`, `scripts/test-check-mutation-reproducer-pin.sh`
**Tests:** `TestCheckMutationReproducerPin`
**Regression:** fails if any of the harness's 20 `ok` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_mutation_reproducer_pin_test.go 2>/dev/null | grep -cE '^func Test' @ 75261910 -->
**After:** none
**Commit:** `feat(guard): port check-mutation-reproducer-pin to Go`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Failing test.** Port every case of
    `scripts/test-check-mutation-reproducer-pin.sh`, one subtest per `ok` label (the wrapped-window
    near-miss, the boundary-spanning-only joined scan, the absent/ambiguous declaration exit 2 and
    the symlinked-site exit 2 included). Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-mutation-reproducer-pin`: the two declaration lines of
    `runreproducer.go` found by the same patterns (exactly one each, else exit 2), marker and window
    extracted by the same token patterns, presence checks with anchor lines, the near-miss scan per
    file sorted and de-duplicated; every message byte-identical.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    '^TestCheckMutationReproducerPin$' -count=1 -race -v | grep -c -- '--- PASS:
    TestCheckMutationReproducerPin/'` — at least 20.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-mutation-reproducer-pin.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-mutation-reproducer-pin.sh` on this worktree prints the
    `MUTATION-REPRODUCER-PIN-OK` line the bash printed at `75261910` and exits 0, directly and
    through a symlink.

- [ ] 3. Port check-python-suppressions

**Files:** `stats/internal/guard/pythonsuppressions.go`, `stats/internal/guard/check_python_suppressions_test.go`, `scripts/check-python-suppressions.sh`, `scripts/test-check-python-suppressions.sh`
**Tests:** `TestCheckPythonSuppressions`
**Regression:** fails if any of the harness's 19 `ok` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_python_suppressions_test.go 2>/dev/null | grep -cE '^func Test' @ 75261910 -->
**After:** none
**Commit:** `feat(guard): port check-python-suppressions to Go`
**Build:** green

**Decision:** python-suppressions-binary-skip

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-check-python-suppressions.sh`,
    one subtest per `ok` label (every marker spelling, case-insensitivity, the symlink skip, the
    missing-path exit 2 and the no-argument tracked-files scan included). Run — expect failure.
  - [ ] **Step 2: Port**, registering `check-python-suppressions`: no arguments → `git ls-files -z
    -- '*.py'` at the root through `gitExec`, symlinks skipped, paths absolute; arguments → files
    (`*.py`, non-symlink) and directories walked for regular `*.py` files, symlinks skipped; hits
    printed `path:line:text` on stdout in argument then walk order, the `suppressions-found` count
    line on stderr; every message byte-identical.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run
    '^TestCheckPythonSuppressions$' -count=1 -race -v | grep -c -- '--- PASS:
    TestCheckPythonSuppressions/'` — at least 19.
  - [ ] **Step 4: Shim and delete** — shim template, code 2;
    `git rm scripts/test-check-python-suppressions.sh`.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/`;
    `scripts/check-python-suppressions.sh` with no arguments, and with `hooks/` as its one argument,
    prints what the bash printed at `75261910` and exits the same, directly and through a symlink.

- [ ] 4. Repoint citations of the deleted harnesses and library

**Files:** `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/*.py`, `hooks/*.py`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`
**Tests:** none — citation sweep; the lint guards are the check
**Regression:** none — prose and comments only
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3
**Commit:** `docs(scripts): repoint citations of the ported guards to their Go sources`
**Build:** green

**Decision:** carry-prior-port-decisions

  - [ ] **Step 1: Find.** `grep -rlF -e test-check-normative-inventory.sh -e
    test-check-mutation-reproducer-pin.sh -e test-check-python-suppressions.sh -e owned-corpus.sh
    --exclude-dir=archive --exclude-dir=.worktrees --exclude-dir=node_modules --exclude-dir=.git
    --exclude-dir=self-review .`
    `unverified: the file set is known only after tasks 1–3 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [ ] **Step 2: Repoint** each citation to the Go file or Go test that now holds what it cites; a
    sentence describing bash plumbing that no longer exists is corrected, not repointed —
    `.flow/project.md`'s "resolves the corpus through `scripts/lib/owned-corpus.sh`" names
    `stats/internal/guard/ownedcorpus.go`, `scripts/lib/parallel.sh`'s trap-neutral comment drops
    the deleted library, and `.flow/project.md`'s paragraph naming which guards are Go lists the
    three new ones.
  - [ ] **Step 3: Verify.** Step 1's grep returns only `spectre/changes/kan-920-*`,
    `stats/internal/guard/libtwins_test.go`/`check_ask_silence_test.go`'s historical
    `owned-corpus.sh` reference and the Go ports' own history comments; every guard in
    `.flow/project.md`'s `## lint` exits 0.

- [ ] 5. Live verification: parity counts

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4
**Build:** green

**Decision:** no-suite-timings

  - [ ] **Step 1: Parity.** `cd stats && go test ./internal/guard/ -run '^Test<Name>$' -count=1 -v
    | grep -c -- '--- PASS: Test<Name>/'` per port against its floor in `design.md`.
  - [ ] **Step 2: Record** an **After** table under design.md's parity floors — Go test, `--- PASS`
    subtests, floor — tagged `measured:` with the command and
    `@ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint`, replacing the
    "recorded by task 5" line, in a planning commit.
  - [ ] **Step 3: Judge.** Failure looks like: any port's `--- PASS` count below
    its floor. It is reported, not recorded as success.
