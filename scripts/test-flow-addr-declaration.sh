#!/usr/bin/env bash
# Assertion harness pinning the flow CLI's store-address resolution against
# the project's own declaration of it (.flow/project.md's `## workspace
# isolation` section). KAN-616: which commands resolve FLOW_RECORDS_ADDR was
# one fact stated four ways, none matching the code, until
# registerRecordConnFlags became the one seam — this harness is what keeps
# the declared set and the wired set from drifting apart again.
#
# WHAT IT ASSERTS, against a root directory (default: the repository this
# script sits in; a positional argument overrides, and the mutation cases
# below always pass a sandbox copy):
#
#   1. The record-family sentence under `## workspace isolation` parses to a
#      non-empty set of backticked command names, and the isolation table
#      carries the `FLOW_ADDR` and `FLOW_RECORDS_ADDR` rows the sentence
#      speaks of.
#   2. Exactly one `-addr` registration under stats/cmd/flow resolves
#      FLOW_RECORDS_ADDR, and it is the one inside registerRecordConnFlags.
#   3. The files holding registerRecordConnFlags call sites are exactly the
#      files the declared commands map to — the CLI's own dispatch
#      convention: the command's subcommand word, hyphens dropped, is the
#      source file's stem (`flow record` -> record.go, `flow self-review
#      bundle` -> selfreview.go). Equality checks both directions at once:
#      no call site outside the declared family, no declared file without
#      one.
#   4. No registerConnFlags call site — the seam whose address follows
#      FLOW_ADDR alone — sits in a declared record-family file, and no
#      inline `-addr` registration through resolveDefaultAddr() sits in one
#      either, outside the registerConnFlags definition itself: a family
#      verb registering through its own helper (the registerJiraConnFlags
#      shape) resolves FLOW_ADDR exactly where the declaration says the
#      whole family resolves FLOW_RECORDS_ADDR.
#   5. Every `-addr` registration under stats/cmd/flow takes its default
#      from resolveDefaultAddr() or resolveRecordsAddr(), never a literal
#      and never a third resolver — so every store-touching command honors
#      FLOW_ADDR, and only the record family consults FLOW_RECORDS_ADDR.
#
# THE SIX CASES run the assertion against the real root and the unmutated
# copy (both must pass) and against one drift mutation each (every one must
# fail). The mutation cases make the harness self-pinning: an assertion
# loosened into vacuity stops failing the mutations and the harness reports
# FAIL on its own next run. Test sources are read only; the real tree is
# never written.
#
# OUTPUT: one `ok <case>` or `FAIL <case>: <why>` line per case.
#
# EXIT CODES: 0 every case passed; 1 at least one case failed; 2 the
# harness could not answer (root layout wrong, sandbox setup failed).

set -u

REAL_ROOT="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
FAILURES=0
REASON=""

ok() { printf 'ok %s\n' "$1"; }
bad() { printf 'FAIL %s: %s\n' "$1" "$2"; FAILURES=$((FAILURES + 1)); }

# assert_root <dir>: run every declaration assertion against one root.
# Prints nothing; returns 0 when all hold, else sets REASON to the first
# violated assertion's description.
assert_root() {
  REASON=""
  local md="$1/.flow/project.md" cmd="$1/stats/cmd/flow"
  [ -f "$md" ] || { REASON="no .flow/project.md under $1"; return 1; }
  [ -d "$cmd" ] || { REASON="no stats/cmd/flow under $1"; return 1; }

  # 1. the declaration parses, and the rows it speaks of are in the table.
  # The anchor is the sentence's opening through its verb on one physical
  # line: the file hard-wraps its prose, and a reflow that moves the paren
  # group or the verb off that line fails here — the safe direction, forcing
  # whoever rewraps the sentence to keep it greppable or move the parse.
  local sentences
  sentences="$(grep -c 'The record family (.*) resolves' "$md")"
  [ "$sentences" -eq 1 ] ||
    { REASON="expected exactly one record-family sentence in .flow/project.md, found $sentences"; return 1; }
  local declared
  declared="$(sed -n 's/.*The record family (\(.*\)) resolves.*/\1/p' "$md" |
    grep -o '`[^`]*`' | tr -d '`' | sort)"
  [ -n "$declared" ] ||
    { REASON="the record-family sentence names no backticked commands"; return 1; }
  grep -q '| `FLOW_ADDR` |' "$md" ||
    { REASON="the isolation table has no FLOW_ADDR row"; return 1; }
  grep -q '| `FLOW_RECORDS_ADDR` |' "$md" ||
    { REASON="the isolation table has no FLOW_RECORDS_ADDR row"; return 1; }

  # the file each declared command maps to, per the dispatch convention
  local declared_files
  declared_files="$(printf '%s\n' "$declared" | awk '{print $2}' | tr -d '-' | sed 's/$/.go/' | sort -u)"

  local srcs
  srcs="$(find "$cmd" -name '*.go' ! -name '*_test.go' | sort)"

  # 2. exactly one FLOW_RECORDS_ADDR registration, inside registerRecordConnFlags
  local total inbody
  total="$(grep -h '"addr", resolveRecordsAddr()' $srcs | wc -l | tr -d ' ')"
  [ "$total" -eq 1 ] ||
    { REASON="expected exactly one FLOW_RECORDS_ADDR -addr registration under stats/cmd/flow, found $total"; return 1; }
  inbody="$(sed -n '/^func registerRecordConnFlags(/,/^func /p' "$cmd/record.go" |
    grep -c '"addr", resolveRecordsAddr()')"
  [ "$inbody" -eq 1 ] ||
    { REASON="the FLOW_RECORDS_ADDR -addr registration is not the one inside registerRecordConnFlags"; return 1; }

  # 3. record-seam call sites live in exactly the declared commands' files
  local site_files
  site_files="$(grep -l "^$(printf '\t')registerRecordConnFlags(" $srcs | xargs -n1 basename | sort -u)"
  [ "$site_files" = "$declared_files" ] || {
    REASON="registerRecordConnFlags call-site files ($(printf '%s' "$site_files" | tr '\n' ' ')) are not exactly the declared record family's files ($(printf '%s' "$declared_files" | tr '\n' ' '))"
    return 1
  }

  # 4. the FLOW_ADDR-only seam is never called from inside the record family,
  #    and the family's files never carry an inline FLOW_ADDR-only -addr
  #    registration either — a new record-family verb registering through its
  #    own helper (the registerJiraConnFlags shape) with resolveDefaultAddr()
  #    resolves FLOW_ADDR where the declaration says the family resolves
  #    FLOW_RECORDS_ADDR, and only the derived file list sees it
  local fam_files=()
  while IFS= read -r f; do fam_files+=("$cmd/$f"); done < <(printf '%s\n' "$declared_files")
  local leaked inline_total in_seam
  leaked="$(grep -l "^$(printf '\t')registerConnFlags(" "${fam_files[@]}" 2>/dev/null)"
  [ -z "$leaked" ] ||
    { REASON="record-family files call the FLOW_ADDR-only seam: $(printf '%s' "$leaked" | tr '\n' ' ')"; return 1; }
  # an inline resolveDefaultAddr() registration inside a family file counts
  # as drift only outside the FLOW_ADDR-only seam's own definition — that
  # helper lives in record.go (registerConnFlags) and legitimately registers
  # through the resolver for the verbs that call it. The pattern is
  # variable-agnostic on purpose: a helper need not register into an `f
  # .addr` field (StringVar(new(string), ...) is the same registration), so
  # only the flag name plus resolver default identifies one.
  inline_total="$(grep -h '"addr", resolveDefaultAddr()' "${fam_files[@]}" | wc -l | tr -d ' ')"
  in_seam="$(sed -n '/^func registerConnFlags(/,/^func /p' "$cmd/record.go" |
    grep -c '"addr", resolveDefaultAddr()')"
  [ "$inline_total" -eq "$in_seam" ] ||
    { REASON="declared record-family files carry $((inline_total - in_seam)) inline -addr registration(s) through resolveDefaultAddr() outside the registerConnFlags definition itself"; return 1; }

  # 5. every -addr registration takes a resolver default, never a literal —
  # the same variable-agnostic shape, so a registration into any target
  # field is counted and attributed
  local regs dres rres
  regs="$(grep -h 'StringVar(.*"addr", ' $srcs | wc -l | tr -d ' ')"
  dres="$(grep -h '"addr", resolveDefaultAddr()' $srcs | wc -l | tr -d ' ')"
  rres="$(grep -h '"addr", resolveRecordsAddr()' $srcs | wc -l | tr -d ' ')"
  [ "$regs" -eq "$((dres + rres))" ] ||
    { REASON="$regs -addr registrations but only $((dres + rres)) take a resolver default — a literal or third resolver is wired somewhere"; return 1; }
  [ "$regs" -eq "$((dres + 1))" ] ||
    { REASON="$regs -addr registrations but only $((dres + 1)) take a resolver default — a literal or third resolver is wired somewhere"; return 1; }

  return 0
}

# make_root <dest>: copy the two inputs the assertion reads from the real
# root into a sandbox layout, so a mutation never touches the real tree.
make_root() {
  local dest="$1"
  mkdir -p "$dest/.flow" "$dest/stats/cmd/flow" || return 1
  cp "$REAL_ROOT/.flow/project.md" "$dest/.flow/project.md" || return 1
  find "$REAL_ROOT/stats/cmd/flow" -name '*.go' ! -name '*_test.go' \
    -exec cp {} "$dest/stats/cmd/flow/" \; || return 1
}

# run_case <label> <pass|fail> <root>: assert_root must reach exactly the
# expected verdict on that root.
run_case() {
  local label="$1" expected="$2" root="$3" got
  if assert_root "$root"; then got=pass; else got=fail; fi
  if [ "$got" = "$expected" ]; then
    ok "$label"
  else
    bad "$label" "expected $expected, got $got${REASON:+ — $REASON}"
  fi
}

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/flow-addr-decl.XXXXXX")" || exit 2
trap 'rm -rf "$SANDBOX"' EXIT

[ -f "$REAL_ROOT/.flow/project.md" ] && [ -d "$REAL_ROOT/stats/cmd/flow" ] ||
  { echo "flow-addr-decl: $REAL_ROOT is not a repository root with .flow/ and stats/cmd/flow" >&2; exit 2; }

# pass-unmutated: the real root and its untouched copy both assert clean
make_root "$SANDBOX/pass" || { echo "flow-addr-decl: sandbox copy failed" >&2; exit 2; }
if assert_root "$REAL_ROOT" && assert_root "$SANDBOX/pass"; then
  ok pass-unmutated
else
  bad pass-unmutated "unmutated tree asserts dirty${REASON:+ — $REASON}"
fi

# mutate-record-seam-outside-family: a verb outside the record family joins
# the record seam
make_root "$SANDBOX/m-outside" || exit 2
sed -i.bak 's/registerConnFlags(/registerRecordConnFlags(/' \
  "$SANDBOX/m-outside/stats/cmd/flow/suite.go" &&
  rm -f "$SANDBOX/m-outside/stats/cmd/flow/suite.go.bak"
run_case mutate-record-seam-outside-family fail "$SANDBOX/m-outside"

# mutate-record-seam-default: registerRecordConnFlags stops resolving
# FLOW_RECORDS_ADDR
make_root "$SANDBOX/m-default" || exit 2
sed -i.bak 's/"addr", resolveRecordsAddr()/"addr", resolveDefaultAddr()/' \
  "$SANDBOX/m-default/stats/cmd/flow/record.go" &&
  rm -f "$SANDBOX/m-default/stats/cmd/flow/record.go.bak"
run_case mutate-record-seam-default fail "$SANDBOX/m-default"

# mutate-family-verb-switches-seam: a record-family verb falls back to the
# FLOW_ADDR-only seam
make_root "$SANDBOX/m-switch" || exit 2
sed -i.bak 's/registerRecordConnFlags(fset, &f)/registerConnFlags(fset, \&f)/' \
  "$SANDBOX/m-switch/stats/cmd/flow/selfreview.go" &&
  rm -f "$SANDBOX/m-switch/stats/cmd/flow/selfreview.go.bak"
run_case mutate-family-verb-switches-seam fail "$SANDBOX/m-switch"

# mutate-declaration-drops-verb: the declared set loses a command the code
# still wires. The sed keeps the closing paren, so the mutated sentence
# still parses and the case fails through assertion 3's set comparison —
# the drift the case name claims — not through the sentence-parse guard.
make_root "$SANDBOX/m-decl" || exit 2
sed -i.bak 's/, `flow self-review bundle`)/)/' "$SANDBOX/m-decl/.flow/project.md" &&
  rm -f "$SANDBOX/m-decl/.flow/project.md.bak"
run_case mutate-declaration-drops-verb fail "$SANDBOX/m-decl"

# mutate-literal-addr-registration: a registration takes a hardcoded default
make_root "$SANDBOX/m-literal" || exit 2
sed -i.bak 's/"addr", resolveDefaultAddr()/"addr", "http:\/\/127.0.0.1:9999"/' \
  "$SANDBOX/m-literal/stats/cmd/flow/jira.go" &&
  rm -f "$SANDBOX/m-literal/stats/cmd/flow/jira.go.bak"
run_case mutate-literal-addr-registration fail "$SANDBOX/m-literal"

if [ "$FAILURES" -gt 0 ]; then
  exit 1
fi
exit 0
