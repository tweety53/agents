package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// The integrate-change cases are its former bash harness's cases, each run
// as its own bash session over that harness's own fixture, `g`, `to_land`,
// `peer_fix_case`, `ui_a` and `flag`, kept verbatim so every case's setup
// and every `check` expression is the harness's own. Only `sut` changes: it
// hands its argv and environment back to this test, which runs
// Registry["integrate-change"] in process and returns OUT, RC and the
// stderr file — so the cases exercise the Go under `go test -cover`, and
// each `check` label is a subtest of its case.

// icDriver is the session's own protocol: a marker line on stdout per
// `sut` call, `check` verdict and `halt`.
const icDriver = `set -u
SCRIPT_DIR="$IC_SCRIPTS"
SUT="$SCRIPT_DIR/integrate-change.sh"
OUT=""; RC=0
check() { if eval "$2"; then printf '\n__IC_PASS__ %s\n' "$1"; else printf '\n__IC_FAIL__ %s\n' "$1"; fi; }
halt() { printf '%s\n%s\n' "$OUT" "$(cat "$WORK/stderr")" >&2; printf '\n__IC_HALT__ %s\n' "$1"; exit 1; }
sut() {
  PATH="$WORK/bin:$PATH" GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.invalid \
    GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.invalid \
    FLOW_GUARD_SCRIPT_DIR="$(dirname "$SUT")" env -0 >"$IC_TMP/sut.env"
  printf '%s\0' "$@" --harness test --session-token mf-test >"$IC_TMP/sut.args"
  printf '\n__IC_SUT__\n'
  read -r _ <"$IC_TMP/sut.ack"
  OUT="$(cat "$IC_TMP/sut.out")"; RC="$(cat "$IC_TMP/sut.rc")"
  cp "$IC_TMP/sut.err" "$WORK/stderr"
}
`

// icHelpers is the harness's fixture and shared case helpers, fixture's
// temp directory under the test's own.
const icHelpers = `g() { GIT_AUTHOR_DATE=2026-01-01T00:00:00Z GIT_COMMITTER_DATE=2026-01-01T00:00:00Z git -c user.name=Test -c user.email=test@example.invalid -c commit.gpgsign=false "$@"; }
fixture() {
  WORK="$(cd "$(mktemp -d "$IC_TMP/work.XXXXXX")" && pwd -P)"
  local r
  for r in a b; do
    g init -q --bare -b main "$WORK/$r-origin.git"
    g init -q -b main "$WORK/$r"
    printf 'base\n' >"$WORK/$r/base.txt"
    if [ "$r" = a ]; then
      mkdir -p "$WORK/a/.flow" "$WORK/a/spectre/changes"
      printf '# project\n\n## default landing route\n\nmerge and push\n' >"$WORK/a/.flow/project.md"
      printf 'b ../b\n' >"$WORK/a/spectre/peers"
    fi
    g -C "$WORK/$r" add -A && g -C "$WORK/$r" commit -qm base
    g -C "$WORK/$r" remote add origin "$WORK/$r-origin.git"
    g -C "$WORK/$r" push -q -u origin main
    g -C "$WORK/$r" remote set-head origin main
    g -C "$WORK/$r" worktree add -q -b spectre/demo "$WORK/$r-worktrees/demo" main
    printf 'work in %s\n' "$r" >"$WORK/$r-worktrees/demo/work-$r.txt"
    g -C "$WORK/$r-worktrees/demo" add -A
    g -C "$WORK/$r-worktrees/demo" commit -qm "feat($r): task 1" -m "Task-Id: 1"
  done
  local c="$WORK/a-worktrees/demo/spectre/changes/demo"
  mkdir -p "$c"
  printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
  printf '# demo\n\n## Parts\n\n` + "`" + `b:demo` + "`" + `\n\n## Branch\n\nmain\n\n## Merge order\n\n1. ` + "`" + `b` + "`" + `\n2. ` + "`" + `.` + "`" + `\n' >"$c/link.md"
  g -C "$WORK/a-worktrees/demo" add -A
  g -C "$WORK/a-worktrees/demo" commit -qm "chore(spectre): plan-gate demo"

  mkdir -p "$WORK/bin" "$WORK/statedir"
  cat >"$WORK/bin/flow" <<EOF
#!/usr/bin/env bash
case "\$1 \$2" in
  "state get") cat "$WORK/state.json" ;;
  "state set") cat >"$WORK/state.json" ;;
  "state dir") echo "$WORK/statedir" ;;
  "state list") echo '{"source":"store","complete":true,"records":[{"name":"demo"}]}' ;;
  "stage begin"|"stage end") echo "\$*" >>"$WORK/stages.log" ;;
  "stage sessions") [ -f "$WORK/fail-sessions" ] && exit 1; if [ "\$6" = flow.document-fix ]; then cat "$WORK/fix-sessions" 2>/dev/null || echo 0; else echo 0; fi ;;
  "record findings") [ -f "$WORK/fail-findings" ] && exit 1; echo '[]' ;;
  "self-review findings") cat "$WORK/sr-findings.json" 2>/dev/null || echo '[]' ;;
  "self-review bundle") printf 'found: 1 of 1 sources; skipped: 0 of 1 sources\n' ;;
  "self-review finding") echo "\$*" >>"$WORK/fixed.log" ;;
  "record dispatches") [ -f "$WORK/loop-rows" ] && exec cat "$WORK/loop-rows"; cat "$WORK/dispatches.json" 2>/dev/null || echo '[{"key":"task-1-implementer","outcome":"completed"},{"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-02T00:00:00Z"}]' ;;
  "record decisions") echo '[]' ;;
  "record passes") echo '[{"id":1,"round":0,"note":"roster: full"}]' ;;
  "record verdicts") cat "$WORK/verdicts.json" 2>/dev/null || echo '[]' ;;
  "workspace-id "*) echo demo-id ;;
  *) : ;;
esac
EOF
  # Run 1 no longer calls spectre: any call fails the run.
  printf '#!/usr/bin/env bash\nexit 1\n' >"$WORK/bin/spectre"
  chmod +x "$WORK/bin/flow" "$WORK/bin/spectre"
  printf '{"state":"IN_PROGRESS","branch":"spectre/demo","worktrees":{"%s":"%s","%s":"%s"},"jiraIssue":"KAN-1","prUrl":null,"updatedBy":"/flow"}\n' \
    "$WORK/a-worktrees/demo" "$(git -C "$WORK/a" rev-parse main)" \
    "$WORK/b-worktrees/demo" "$(git -C "$WORK/b" rev-parse main)" >"$WORK/state.json"
}
to_land() {
  sut prepare "$WORK/a" demo
  check "${1}prepare exits 0" '[ "$RC" -eq 0 ]'
  [ "$RC" -eq 0 ] || halt prepare
  printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
  sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
  check "${1}commit exits 0" '[ "$RC" -eq 0 ]'
  [ "$RC" -eq 0 ] || halt commit
  [ "${2:-}" = no-report ] && return 0
  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
  printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
}
peer_fix_case() {
  to_land "$1" no-report
  printf 'fix\n' >"$WORK/b-worktrees/demo/fix-b.txt"
  g -C "$WORK/b-worktrees/demo" add fix-b.txt && g -C "$WORK/b-worktrees/demo" commit -qm "fix(b): the peer fix"
  old="$(git -C "$WORK/b-worktrees/demo" rev-parse --short HEAD)"
  mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
  printf '# demo self-review\n\n- **[flow-fix]** the peer fix — fixed: %s\n' "$old" >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
  g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
  g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"
  printf 'moved\n' >"$WORK/b/moved.txt"
  g -C "$WORK/b" add moved.txt && g -C "$WORK/b" commit -qm "base: moved" && g -C "$WORK/b" push -q origin main
  sut land "$WORK/a" demo
  new="$(git --git-dir="$WORK/b-origin.git" log --format=%h --grep="^fix(b): the peer fix$" -1 main)"
  check "${1}land exits 0" '[ "$RC" -eq 0 ]'
  check "${1}the re-sync rewrote b's fix sha" '[ -n "$new" ] && [ "$new" != "$old" ]'
  check "${1}the report on a's origin main names the rewritten sha" \
    'git --git-dir="$WORK/a-origin.git" show main:docs/self-review/demo-self-review.md | grep -q "fixed: $new$"'
  check "${1}a fixed row is recorded for b's fix" \
    'grep -q -- "-disposition fixed -ref $new " "$WORK/fixed.log"'
}
ui_a() {
  printf '\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| ` + "`" + `ui paths` + "`" + ` | ` + "`" + `work-a.txt` + "`" + ` |\n' >>"$WORK/a-worktrees/demo/.flow/project.md"
  printf '# Visual verification — demo\n' >"$WORK/a-worktrees/demo/spectre/changes/demo/visual-verification.md"
  g -C "$WORK/a-worktrees/demo" add spectre/changes/demo/visual-verification.md
  g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
}
flag() {
  printf '[{"change":"%s","guard":"check-visual-verify-dispatched","worktree":"%s","verdict":"VISUAL-VERIFY-MISSING: no verifier — HEAD %s","falsePositive":true,"falsePositiveReason":"hand-verified","flaggedAt":"2026-01-01T00:00:00Z"}]' \
    "${3:-demo}" "$(cd "$1" && pwd -P)" "$2" >"$WORK/verdicts.json"
}`

type icCaseDef struct{ name, script string }

// icRunCases runs each case on its own fixture, in parallel.
func icRunCases(t *testing.T, cases []icCaseDef) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			icRun(t, tc.script)
		})
	}
}

func icRun(t *testing.T, script string) {
	tmp := t.TempDir()
	scripts, err := filepath.Abs("../../../scripts")
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(tmp+"/sut.ack", 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", icDriver+icHelpers+"\n"+script)
	cmd.Dir = tmp
	// HOME is the test's own, so no guard's commit reads the operator's
	// git config; the Go caches stay where the toolchain keeps them.
	cmd.Env = append(os.Environ(), "IC_TMP="+tmp, "IC_SCRIPTS="+scripts,
		"FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_TELEMETRY=off",
		"HOME="+tmp+"/home", "GIT_CONFIG_NOSYSTEM=1")
	cmd.Env = append(cmd.Env, icGoEnv(t)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	last := func() string {
		out, _ := os.ReadFile(tmp + "/sut.out")
		errOut, _ := os.ReadFile(tmp + "/sut.err")
		return fmt.Sprintf("last sut stdout:\n%s\nlast sut stderr:\n%s", out, errOut)
	}
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		l := sc.Text()
		switch {
		case l == "__IC_SUT__":
			icServe(t, tmp)
		case strings.HasPrefix(l, "__IC_PASS__ "):
			t.Run(strings.TrimPrefix(l, "__IC_PASS__ "), func(t *testing.T) {})
		case strings.HasPrefix(l, "__IC_FAIL__ "):
			t.Run(strings.TrimPrefix(l, "__IC_FAIL__ "), func(t *testing.T) { t.Error(last()) })
		case strings.HasPrefix(l, "__IC_HALT__ "):
			t.Errorf("%s did not complete; the case stops here\n%s\n%s", strings.TrimPrefix(l, "__IC_HALT__ "), last(), stderr.String())
		}
	}
	if err := cmd.Wait(); err != nil && !t.Failed() {
		t.Errorf("the case's session failed: %v\n%s", err, stderr.String())
	}
}

// icGoEnv pins the Go caches a moved HOME would otherwise relocate.
func icGoEnv(t *testing.T) []string {
	out, err := exec.Command("go", "env", "GOCACHE", "GOMODCACHE", "GOPATH").Output()
	if err != nil {
		t.Fatal(err)
	}
	v := strings.Split(strings.TrimSpace(string(out)), "\n")
	return []string{"GOCACHE=" + v[0], "GOMODCACHE=" + v[1], "GOPATH=" + v[2]}
}

// icServe is one `sut` call: the session's argv and environment, run in
// process, its stdout, stderr and exit code handed back.
func icServe(t *testing.T, tmp string) {
	t.Helper()
	vars := map[string]string{}
	var environ []string
	raw, _ := os.ReadFile(tmp + "/sut.env")
	for _, kv := range strings.Split(string(raw), "\x00") {
		if k, v, ok := strings.Cut(kv, "="); ok {
			vars[k] = v
			environ = append(environ, kv)
		}
	}
	argv, _ := os.ReadFile(tmp + "/sut.args")
	args := strings.Split(strings.TrimSuffix(string(argv), "\x00"), "\x00")
	env := Env{
		Getenv:    func(k string) string { return vars[k] },
		LookupEnv: func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		Environ:   func() []string { return environ },
		Dir:       "/",
	}
	var out, errOut bytes.Buffer
	rc := Registry["integrate-change"](args, env, &out, &errOut)
	for name, b := range map[string][]byte{"sut.out": out.Bytes(), "sut.err": errOut.Bytes(), "sut.rc": []byte(strconv.Itoa(rc))} {
		if err := os.WriteFile(tmp+"/"+name, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ack, err := os.OpenFile(tmp+"/sut.ack", os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	ack.WriteString("\n")
	ack.Close()
}

func TestIntegrateChangeHappyPath(t *testing.T) {
	icRunCases(t, []icCaseDef{
		{"happy_path", `# --- happy_path ---
fixture
sut prepare "$WORK/a" demo
check "prepare exits 0" '[ "$RC" -eq 0 ]'
check "prepare names the configured route" 'printf "%s" "$OUT" | grep -q "^ROUTE: merge and push — from this project"'
check "prepare prints each worktree's reshape base" '[ "$(printf "%s\n" "$OUT" | grep -c "^WORKTREE: ")" -eq 2 ]'
check "prepare ends on NEXT: commit" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: append this run.s narrative"'
[ "$RC" -eq 0 ] || halt prepare

printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "commit exits 0" '[ "$RC" -eq 0 ]'
check "commit ends on NEXT: self-review" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass"'
[ "$RC" -eq 0 ] || halt commit
mkdir -p "$WORK/a-worktrees/demo/docs/self-review"
printf '# demo self-review\n' >"$WORK/a-worktrees/demo/docs/self-review/demo-self-review.md"
g -C "$WORK/a-worktrees/demo" add docs/self-review/demo-self-review.md
g -C "$WORK/a-worktrees/demo" commit -qm "docs(self-review): demo self-review report"

sut land "$WORK/a" demo
check "land exits 0" '[ "$RC" -eq 0 ]'
[ "$RC" -eq 0 ] || halt land
check "a's origin main carries implementation, plan, drop and report commits" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -4 | tr "\n" "|")" = "docs(self-review): demo self-review report|chore(spectre): drop demo|chore(spectre): plan demo|feat(a): do the work|" ]'
check "a's origin main holds neither the change directory nor an archive" \
  '[ -z "$(git --git-dir="$WORK/a-origin.git" ls-tree -d main spectre/changes/demo spectre/changes/archive)" ]'
check "b's origin main carries its implementation commit" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = "feat(b): do the work" ]'
check "b was pushed before a (merge order)" \
  'printf "%s\n" "$OUT" | grep "^PUSHED: " | head -1 | grep -q "/b-worktrees/demo"'
check "both worktrees are removed" '[ ! -d "$WORK/a-worktrees/demo" ] && [ ! -d "$WORK/b-worktrees/demo" ]'
check "the record is FINISHED with an empty worktrees map" \
  '[ "$(jq -c "[.state, .worktrees]" "$WORK/state.json")" = "[\"FINISHED\",{}]" ]'
check "Jira In Review then Done are handed to the parent" \
  '[ "$(printf "%s\n" "$OUT" | grep "^JIRA: " | tr "\n" "|")" = "JIRA: transition KAN-1 to In Review|JIRA: transition KAN-1 to Done|" ]'
check "the stage log runs preflight through refresh-main-checkout" \
  'head -1 "$WORK/stages.log" | grep -q "stage begin .*-stage flow.preflight" && tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.refresh-main-checkout -outcome completed"'
check "run 2 keeps a's pre-reshape pin, locally and on origin" \
  'git -C "$WORK/a" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null && git --git-dir="$WORK/a-origin.git" rev-parse -q --verify refs/flow/pre-reshape/demo >/dev/null'
check "land ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
check "the NEXT line names the template and its never-omitted lines" 'printf "%s" "$OUT" | tail -1 | grep -qF "(skills/flow/cleanup.md) — its Learning: and Prompt tuning: lines included"'`},
	})
}

// TestIntegrateChangeDropsSatelliteLink pins the satellite half of run 1's
// drop: a worktree holding the change's link.md directory drops it in a
// drop commit of its own, and nothing is archived.
func TestIntegrateChangeDropsSatelliteLink(t *testing.T) {
	icRunCases(t, []icCaseDef{
		{"satellite_drop", `# --- satellite_drop ---
fixture
mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
printf '## Part of\n\n` + "`" + `a:demo` + "`" + `\n' >"$WORK/b-worktrees/demo/spectre/changes/demo/link.md"
g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): link demo"
to_land "satellite_drop: " no-report
check "satellite_drop: b's change directory is gone from disk and HEAD" \
  '[ ! -d "$WORK/b-worktrees/demo/spectre/changes/demo" ] && [ -z "$(git -C "$WORK/b-worktrees/demo" ls-tree -d HEAD spectre/changes/demo spectre/changes/archive)" ]'
check "satellite_drop: b's branch ends on its own drop commit" \
  '[ "$(git -C "$WORK/b-worktrees/demo" log -1 --format=%s)" = "chore(spectre): drop demo" ]'`},
	})
}

// The shim's own `cd /`: run outside sut, from inside the canonical
// worktree, with a relative main checkout.
func TestIntegrateChangeShim(t *testing.T) {
	icRunCases(t, []icCaseDef{
		{"shim_caller_cwd", `# --- shim_caller_cwd ---
fixture
for r in a b; do
  g -C "$WORK/$r-worktrees/demo" push -q origin HEAD:main
  g -C "$WORK/$r" fetch -q origin
done
OUT="$(cd "$WORK/a-worktrees/demo" && PATH="$WORK/bin:$PATH" exec "$SUT" cleanup ../../a demo --harness test --session-token mf-test 2>"$WORK/stderr")"; RC=$?
check "shim_caller_cwd: cleanup run from inside a worktree is not HELD by its own cwd" '! printf "%s\n" "$OUT" | grep -q "^HELD:"'
check "shim_caller_cwd: a relative main checkout resolves against the caller's directory" \
  '[ "$RC" -eq 0 ] && [ ! -d "$WORK/a-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
	})
}
