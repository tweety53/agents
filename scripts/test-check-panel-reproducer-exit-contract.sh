#!/usr/bin/env bash
# Assertion harness for check-panel-reproducer-exit-contract.sh.
#
# THE GUARD UNDER TEST IS BEHAVIOURAL, so every case runs it against a
# sandbox where both of its dependencies are controlled: a stub `flow` on
# PATH answers `record findings -change <name>` with a canned JSON array
# (the shape stats/cmd/flow/record.go's `record findings` prints -- one
# object per finding with `ref`, `status`, `reproducer`), and a stub
# `run-reproducer.sh` sitting BESIDE A COPY OF THE GUARD answers with a
# canned verdict code. The copy is what puts the stub inside the guard's
# own SCRIPT_DIR resolution -- the guard execs "$SCRIPT_DIR/run-reproducer.sh",
# never a PATH lookup -- so the stub must sit there, not in bin/.
#
# Two cases (16 and 17) wire the REAL scripts/run-reproducer.sh in
# against a real reproducer script instead, positive and inverted, so the
# guard's wiring to the runner's actual exit-code vocabulary is proven and
# not only assumed. The cases after them exercise the guard's instrument
# audit (KAN-606): the `# demonstrates:` citation every runnable open
# finding's reproducer must carry, and the failure classes of checking it
# against the tree before the runner is ever invoked.
#
# set -euo pipefail, with set +e brackets around every command that is
# SUPPOSED to fail -- the same shape as test-check-panel-reproducers.sh,
# whose header records why the bracket exists.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-reproducer-exit-contract.sh"
REAL_RUNNER="$SCRIPT_DIR/run-reproducer.sh"
FAILED=0

WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# findings_json <ref> <status> <reproducer> [<ref> <status> <reproducer> ...]
# -- a compact JSON array of finding objects, built via `jq -n --args` so a
# reproducer value carrying a quote or a backslash is escaped by jq, not by
# this harness. Same helper, same reasoning, as test-check-panel-reproducers.sh.
findings_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 3)
     | {ref: $a[. * 3], status: $a[. * 3 + 1], reproducer: $a[. * 3 + 2]}]
  ' --args -- "$@"
}

# make_stub_sandbox <json> <runner-exit-code> -- a worktree-shaped sandbox
# whose stub runner exits <runner-exit-code> with the matching canned
# verdict line, and which records the argument count it was last invoked
# with into runner/argc.txt. Prints the sandbox path.
#
# Every sandbox also carries an AUDITABLE reproducer script and the file its
# declaration cites (KAN-606): the guard audits the `# demonstrates:`
# citation BEFORE it ever invokes the runner, so the default recorded
# reproducer `repro.sh` must carry a resolvable declaration and
# target.txt:2 must carry the cited content -- otherwise every verdict case
# below would die in the audit and never reach the runner it exists to
# stub. Cases that exercise the audit itself rewrite repro.sh after this
# helper runs.
make_stub_sandbox() {
  local wt json="$1" code="$2" line
  case "$code" in
    0) line="run-reproducer: defect demonstrated — 'repro.sh' exited 9" ;;
    1) line="run-reproducer: defect not demonstrated — 'repro.sh' exited 0" ;;
    2) line="run-reproducer: refused — shape check failed — never executed" ;;
    3) line="run-reproducer: unverifiable — 'repro.sh' was still running at the bound and was killed" ;;
    4) line="run-reproducer: cannot answer — bad usage" ;;
    *) printf 'make_stub_sandbox: unknown runner exit code %s\n' "$code" >&2; exit 1 ;;
  esac
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")" || {
    printf 'make_stub_sandbox: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin" "$wt/runner"
  printf '%s' "$json" > "$wt/bin/findings.json"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
  chmod +x "$wt/bin/flow"
  cat > "$wt/runner/run-reproducer.sh" <<STUB
#!/usr/bin/env bash
printf '%s' "\$#" > "$(printf '%s' "$wt")/runner/argc.txt"
echo "$line" >&2
exit "$code"
STUB
  chmod +x "$wt/runner/run-reproducer.sh"
  printf '%s\n' 'line one' 'defect present here' 'line three' > "$wt/target.txt"
  printf '%s\n' '#!/usr/bin/env bash' '# demonstrates: target.txt:2:defect present here' 'exit 9' > "$wt/repro.sh"
  cp "$GUARD" "$wt/runner/check-panel-reproducer-exit-contract.sh"
  printf '%s' "$wt"
}

# rewrite_repro <sandbox> -- replaces the sandbox's auditable repro.sh with
# the given body, so an audit case decides exactly what the guard reads.
rewrite_repro() {
  local wt="$1"
  cat > "$wt/repro.sh"
}

# runner_never_invoked <label> <sandbox> -- the stub writes runner/argc.txt
# the moment the guard invokes it, so a missing file is the proof that an
# audit failure skipped the run: the guard never spends the runner's verdict
# on an instrument it could not resolve.
runner_never_invoked() {
  local label="$1" wt="$2"
  if [ -f "$wt/runner/argc.txt" ]; then
    printf 'FAIL %s: the runner was invoked despite the audit failure\n' "$label"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
}

# make_store_unreachable_sandbox -- a sandbox whose stub `flow` exits
# non-zero, simulating a store `record findings` could not reach. The stub
# runner is still present, so the case exercises the store failure and not
# a missing-runner cannot-answer the guard's own readability check would
# otherwise answer first.
make_store_unreachable_sandbox() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")" || {
    printf 'make_store_unreachable_sandbox: mktemp failed -- aborting suite\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin" "$wt/runner"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: connect: connection refused" >&2
exit 1
STUB
  chmod +x "$wt/bin/flow"
  cat > "$wt/runner/run-reproducer.sh" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
  chmod +x "$wt/runner/run-reproducer.sh"
  cp "$GUARD" "$wt/runner/check-panel-reproducer-exit-contract.sh"
  printf '%s' "$wt"
}

# make_real_sandbox <repro-exit-code> -- a sandbox with the REAL
# run-reproducer.sh beside the guard copy and a real executable reproducer
# script inside the worktree, exiting <repro-exit-code>. An open finding's
# recorded reproducer names that script, so the runner's verdict is
# demonstrated (non-zero script exit) or not demonstrated (0) for real.
# The runner sources scripts/reproducer-metachars.sh from its OWN directory
# (its readability check resolves $SCRIPT_DIR), so the shared metachar file
# is copied beside it -- its absence is the runner's exit-4 cannot-answer,
# the very verdict this sandbox must not trip over.
make_real_sandbox() {
  local wt code="$1" json
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")" || {
    printf 'make_real_sandbox: mktemp failed -- aborting suite\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin" "$wt/runner"
  json="$(findings_json F1 open demo-repro.sh)"
  printf '%s' "$json" > "$wt/bin/findings.json"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s\n' 'line one' 'defect present here' 'line three' > "$wt/target.txt"
  printf '%s\n' '#!/usr/bin/env bash' '# demonstrates: target.txt:2:defect present here' 'exit '"$code" > "$wt/demo-repro.sh"
  chmod +x "$wt/demo-repro.sh"
  cp "$REAL_RUNNER" "$wt/runner/run-reproducer.sh"
  cp "$SCRIPT_DIR/reproducer-metachars.sh" "$wt/runner/reproducer-metachars.sh"
  mkdir -p "$wt/runner/lib"
  cp "$SCRIPT_DIR/lib/sha256-hex.sh" "$wt/runner/lib/sha256-hex.sh"
  cp "$GUARD" "$wt/runner/check-panel-reproducer-exit-contract.sh"
  printf '%s' "$wt"
}

# run_guard <sandbox> [change-name] -- runs the sandbox's copy of the guard
# with the sandbox's bin/ (the stub flow) first on PATH. The default name is
# applied only when the argument is ABSENT, never when it is an empty
# string: the missing-name usage case passes "" and must reach the guard as
# "".
run_guard() {
  local wt="$1" name="demo"
  [ "$#" -ge 2 ] && name="$2"
  PATH="$wt/bin:$PATH" "$wt/runner/check-panel-reproducer-exit-contract.sh" "$wt" "$name"
}

expect_exit() {
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
}

expect_exit_and_names() {
  local label="$1" want="$2" needle="$3"; shift 3
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
    return
  fi
  case "$out" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *)
      printf 'FAIL %s: expected output to name %s, got:\n%s\n' "$label" "$needle" "$out"
      FAILED=1
      ;;
  esac
}

# ===========================================================================
# 1. An open finding whose reproducer is demonstrated -- exit 0, and the
#    runner was invoked with EXACTLY two arguments: the dispatch-time bare
#    run, never a --pre-fix-verdict form, whose ambiguity refusal would turn
#    every first verdict into a cannot-answer.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
expect_exit 'case 1: a demonstrated open finding exits 0' 0 run_guard "$wt"
argc="$(cat "$wt/runner/argc.txt")"
if [ "$argc" = "2" ]; then
  printf 'ok: %s\n' 'case 1b: the runner ran bare (two arguments, no --pre-fix-verdict)'
else
  printf 'FAIL case 1b: the runner was invoked with %s arguments, not 2\n' "$argc"
  FAILED=1
fi

# ===========================================================================
# 2. THE CLASS THIS GUARD EXISTS FOR (KAN-554): an open finding whose
#    reproducer reads *defect not demonstrated* on the tree under review --
#    the inverted exit-code condition. Exit 1, naming the ref.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 1)"
expect_exit_and_names 'case 2: a not-demonstrated verdict on an open finding exits 1' 1 'F1' run_guard "$wt"

# ===========================================================================
# 3. An open finding whose reproducer the runner refused -- unusable, the
#    claim cannot hold. Exit 1 (a violation), naming the ref.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 2)"
expect_exit_and_names 'case 3: a refused reproducer exits 1' 1 'F1' run_guard "$wt"

# ===========================================================================
# 4. An open finding whose reproducer could not be verdicted (timeout /
#    survivor, runner exit 3) -- no verdict exists, so the guard cannot
#    answer: exit 2, naming the ref.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 3)"
expect_exit_and_names 'case 4: an unverifiable reproducer is cannot-answer' 2 'F1' run_guard "$wt"

# ===========================================================================
# 5. Runner exit 4 (cannot answer at all) -- likewise the guard's exit 2.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 4)"
expect_exit_and_names 'case 5: a cannot-answer runner verdict is the guard exit 2' 2 'F1' run_guard "$wt"

# ===========================================================================
# 6. Findings that claim nothing about the current tree are skipped,
#    whatever reproducer they carry: fixed, deferred, withdrawn. Exit 0
#    even though the (stubbed) runner would have refused, because the
#    runner is never invoked for them.
# ===========================================================================
for st in fixed "deferred covered by the suite run" "withdrawn operator said so"; do
  wt="$(make_stub_sandbox "$(findings_json F1 "$st" repro.sh)" 2)"
  expect_exit "case 6: a $st finding is skipped, runner never invoked" 0 run_guard "$wt"
done

# ===========================================================================
# 7. The exemption form on an open finding -- it claims nothing runnable.
#    Exit 0, runner never invoked.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open 'none — prose-only, no runnable check')" 2)"
expect_exit 'case 7: the none — reason exemption is skipped' 0 run_guard "$wt"

# ===========================================================================
# 8. A bare `none` on an open finding is the lexical guard's violation, not
#    this guard's: skipped here, exit 0.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open none)" 2)"
expect_exit 'case 8: a bare none is the lexical guard'"'"'s subject, skipped here' 0 run_guard "$wt"

# ===========================================================================
# 9. No findings at all -- exit 0.
# ===========================================================================
wt="$(make_stub_sandbox '[]' 2)"
expect_exit 'case 9: zero findings exits 0' 0 run_guard "$wt"

# ===========================================================================
# 10. The worktree argument is not a directory -- exit 2.
# ===========================================================================
not_a_dir="$(mktemp "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")"
expect_exit_and_names 'case 10: non-directory argument exits 2 and names it' 2 'not a directory' "$GUARD" "$not_a_dir" demo
rm -f "$not_a_dir"

# ===========================================================================
# 11. The change name is containment-checked the same way its sibling guard
#     checks it, and a missing name is usage.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
for bad_name in "../../../planted/clear" "demo*" "demo/../demo" ".hidden" "demo?x"; do
  expect_exit_and_names "case 11: change name '$bad_name' is rejected" 2 'is not a plain change name' run_guard "$wt" "$bad_name"
done
expect_exit_and_names 'case 11: a missing change name is rejected' 2 'usage:' run_guard "$wt" ""

# ===========================================================================
# 12. An unreachable store is cannot-answer, never clean and never
#     violations-found.
# ===========================================================================
wt="$(make_store_unreachable_sandbox)"
expect_exit_and_names 'case 12: an unreachable store exits 2' 2 'cannot determine anything' run_guard "$wt"

# ===========================================================================
# 13. An open finding carrying NO reproducer field at all -- cannot answer
#     (exit 2), the same reading its sibling guard gives the shape.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json F1 open '')" 0)"
expect_exit_and_names 'case 13: a null reproducer on an open finding is cannot-answer' 2 'no reproducer field' run_guard "$wt"

# ===========================================================================
# 14. One array, three findings: F1 open and demonstrated, F2 open and
#     not-demonstrated, F3 fixed and runnable (skipped). Exit 1 naming F2
#     only -- F1 holds and F3 claims nothing about this tree.
# ===========================================================================
wt="$(make_stub_sandbox "$(findings_json \
  F1 open repro-good.sh \
  F2 open repro-inverted.sh \
  F3 fixed repro-old.sh)" 0)"
# The stub answers 0 for every run, which would make F2 pass -- so this
# case needs a runner that answers per command. Rebuild it by hand: same
# sandbox shape, runner keyed on $2.
wt14="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")"
WORKTREES+=("$wt14")
mkdir -p "$wt14/bin" "$wt14/runner"
findings_json \
  F1 open repro-good.sh \
  F2 open repro-inverted.sh \
  F3 fixed repro-old.sh > "$wt14/bin/findings.json"
printf '%s\n' 'line one' 'defect present here' 'line three' > "$wt14/target.txt"
for script in repro-good.sh repro-inverted.sh repro-old.sh; do
  printf '%s\n' '#!/usr/bin/env bash' '# demonstrates: target.txt:2:defect present here' 'exit 9' > "$wt14/$script"
done
cat > "$wt14/bin/flow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
chmod +x "$wt14/bin/flow"
cat > "$wt14/runner/run-reproducer.sh" <<'STUB'
#!/usr/bin/env bash
case "${2:-}" in
  *repro-inverted*) echo "run-reproducer: defect not demonstrated — '$2' exited 0" >&2; exit 1 ;;
  *) echo "run-reproducer: defect demonstrated — '$2' exited 9" >&2; exit 0 ;;
esac
STUB
chmod +x "$wt14/runner/run-reproducer.sh"
cp "$GUARD" "$wt14/runner/check-panel-reproducer-exit-contract.sh"
expect_exit_and_names 'case 14: mixed verdicts exit 1 naming only the contradicted ref' 1 'F2' \
  run_guard "$wt14" demo

# ===========================================================================
# 15. Precedence: a violation (F1 not demonstrated) beside a run that could
#     not be verdicted (F2 unverifiable). Cannot-answer outranks: exit 2.
# ===========================================================================
wt17="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")"
WORKTREES+=("$wt17")
mkdir -p "$wt17/bin" "$wt17/runner"
findings_json F1 open repro-a.sh F2 open repro-b.sh > "$wt17/bin/findings.json"
printf '%s\n' 'line one' 'defect present here' 'line three' > "$wt17/target.txt"
for script in repro-a.sh repro-b.sh; do
  printf '%s\n' '#!/usr/bin/env bash' '# demonstrates: target.txt:2:defect present here' 'exit 9' > "$wt17/$script"
done
cat > "$wt17/bin/flow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
chmod +x "$wt17/bin/flow"
cat > "$wt17/runner/run-reproducer.sh" <<'STUB'
#!/usr/bin/env bash
case "${2:-}" in
  *repro-a*) echo "run-reproducer: defect not demonstrated — '$2' exited 0" >&2; exit 1 ;;
  *) echo "run-reproducer: unverifiable — '$2' was still running at the bound and was killed" >&2; exit 3 ;;
esac
STUB
chmod +x "$wt17/runner/run-reproducer.sh"
cp "$GUARD" "$wt17/runner/check-panel-reproducer-exit-contract.sh"
expect_exit_and_names 'case 15: cannot-answer outranks violations' 2 'F2' \
  run_guard "$wt17" demo

# ===========================================================================
# 16. END-TO-END with the REAL runner: an open finding whose reproducer
#     script exits 7 is demonstrated -- exit 0.
# ===========================================================================
wt="$(make_real_sandbox 7)"
expect_exit 'case 16: real runner, non-zero reproducer, exit 0' 0 run_guard "$wt"

# ===========================================================================
# 17. END-TO-END with the REAL runner: an open finding whose reproducer
#     script exits 0 is the inverted class -- exit 1.
# ===========================================================================
wt="$(make_real_sandbox 0)"
expect_exit_and_names 'case 17: real runner, zero-exit reproducer, exit 1' 1 'F1' run_guard "$wt"

# ===========================================================================
# THE INSTRUMENT AUDIT (KAN-606). Before the guard invokes the runner it
# audits the reproducer's `# demonstrates: <path>:<line>:<content>`
# declaration against the tree: the class KAN-554's guard cannot see, a
# reproducer whose greps captured mismatched text and cited the wrong
# location, still flips green when its exit code happens to be non-zero.
# Every audit failure is a violation (exit 1), and the runner is never
# invoked for an instrument that failed the audit: its verdict would answer
# a question the audit already settled.
# ===========================================================================

# 18. No `# demonstrates:` declaration at all -- unaudited instrument,
#     exit 1, runner never invoked even though the stub would answer 0.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
exit 9
REPRO
expect_exit_and_names 'case 18: no demonstrates declaration is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 18b: no declaration means the runner never runs' "$wt"

# 19. A declaration outside the first 10 lines is no declaration -- the
#     window is the mutation-reproducer convention's own.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
{
  printf '%s\n' '#!/usr/bin/env bash'
  for _ in $(seq 10); do printf '%s\n' '# padding'; done
  printf '%s\n' '# demonstrates: target.txt:2:defect present here' 'exit 9'
} > "$wt/repro.sh"
expect_exit_and_names 'case 19: a declaration past line 10 is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 19b: the audit failure means the runner never runs' "$wt"

# 20. An absolute declared path cannot resolve inside the worktree.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: /etc/passwd:1:root
exit 9
REPRO
expect_exit_and_names 'case 20: an absolute declared path is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 20b: the audit failure means the runner never runs' "$wt"

# 21. A `..` segment in the declared path walks out of the worktree.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: ../outside.txt:1:content
exit 9
REPRO
expect_exit_and_names 'case 21: a .. declared path is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 21b: the audit failure means the runner never runs' "$wt"

# 22. The declared file does not exist on the tree under review -- the
#     mismatched-grep class itself.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: missing.txt:2:defect present here
exit 9
REPRO
expect_exit_and_names 'case 22: a declared file the tree does not carry is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 22b: the audit failure means the runner never runs' "$wt"

# 23. The declared line number is past the end of the declared file.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: target.txt:99:defect present here
exit 9
REPRO
expect_exit_and_names 'case 23: a declared line past the end of file is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 23b: the audit failure means the runner never runs' "$wt"

# 24. The declared content is absent from the declared line -- the wrong
#     test's assertion, cited.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: target.txt:2:some other assertion entirely
exit 9
REPRO
expect_exit_and_names 'case 24: content absent from the declared line is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 24b: the audit failure means the runner never runs' "$wt"

# 25. A malformed declaration -- no line number to read.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: target.txt:defect present here
exit 9
REPRO
expect_exit_and_names 'case 25: a malformed declaration is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 25b: the audit failure means the runner never runs' "$wt"

# 26. A mutation-declared reproducer is EXEMPT: its instrument is audited by
#     the KAN-568 sha-pin machinery, and the content it demonstrates is the
#     mutated tree it builds at run time, not a location on this tree. No
#     declaration needed; the runner is invoked and its verdict decides.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# mutation-reproducer
exit 0
REPRO
expect_exit 'case 26: a mutation-declared reproducer skips the audit' 0 run_guard "$wt"
argc="$(cat "$wt/runner/argc.txt")"
if [ "$argc" = "2" ]; then
  printf 'ok: %s\n' 'case 26b: the exempt reproducer still ran'
else
  printf 'FAIL case 26b: the exempt reproducer was invoked with %s arguments, not 2\n' "$argc"
  FAILED=1
fi

# 27. The recorded reproducer names a script the worktree does not carry at
#     all -- nothing to audit, so nothing to run.
wt="$(make_stub_sandbox "$(findings_json F1 open absent.sh)" 0)"
expect_exit_and_names 'case 27: an unreadable reproducer script is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 27b: an unreadable script means the runner never runs' "$wt"

# 28. A TAB-separated recorded reproducer is legal everywhere else in the
#     pipeline — the lexical guard and the runner both tokenize on space
#     AND tab — so the audit must derive its path token the same way and
#     reach the runner, not bounce the finding as unreadable (round-0 F1).
wt="$(make_stub_sandbox "$(findings_json F1 open "$(printf 'repro.sh\t--strict')")" 0)"
expect_exit 'case 28: a tab-separated reproducer reaches the runner' 0 run_guard "$wt"
argc="$(cat "$wt/runner/argc.txt")"
if [ "$argc" = "2" ]; then
  printf 'ok: %s\n' 'case 28: the tab-separated reproducer was invoked bare'
else
  printf 'FAIL case 28: the tab-separated reproducer was invoked with %s arguments, not 2\n' "$argc"
  FAILED=1
fi

# 29. A declared citation path that is a symlink INSIDE the worktree but
#     points OUTSIDE it passes every lexical check — the audit resolves the
#     declared path physically and refuses the escape, runner never invoked
#     (round-0 F4).
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 0)"
outside="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducer-exit-contract-test.XXXXXX")"
WORKTREES+=("$outside")
printf '%s\n' 'outside content' > "$outside/external.txt"
ln -s "$outside/external.txt" "$wt/escape.txt"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# demonstrates: escape.txt:1:outside content
exit 9
REPRO
expect_exit_and_names 'case 29: a symlink-escape citation is a violation' 1 'F1' run_guard "$wt"
runner_never_invoked 'case 29b: a symlink escape means the runner never runs' "$wt"

# 30. KAN-622: the not-demonstrated message names BOTH exit-code
#     conventions. A declared mutation-reproducer demonstrates with exit 0
#     (the build succeeds with the mutation landed), so a message telling
#     the author the reproducer "must exit non-zero here" inverts the
#     instruction for exactly the convention KAN-568 added. Exit 1 naming
#     the ref, and the message carries the mutation mapping.
wt="$(make_stub_sandbox "$(findings_json F1 open repro.sh)" 1)"
rewrite_repro "$wt" <<'REPRO'
#!/usr/bin/env bash
# mutation-reproducer
exit 7
REPRO
expect_exit_and_names 'case 30: the not-demonstrated message names the mutation convention' 1 'mutation-reproducer' run_guard "$wt"

if [ "$FAILED" -ne 0 ]; then
  printf 'check-panel-reproducer-exit-contract-test: one or more cases failed\n' >&2
  exit 1
fi
printf 'check-panel-reproducer-exit-contract-test: every case passes\n'
