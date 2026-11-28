package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// The integrate-change cases are its former bash harness's cases, each run
// as its own bash session over that harness's own fixture, `g`, `to_land`
// and `peer_fix_case`, kept verbatim so every case's setup
// and every `check` expression is the harness's own. Only `sut` changes: it
// hands its argv and environment back to this test, which runs
// Registry["integrate-change"] in process and returns OUT, RC and the
// stderr file — so the cases exercise the Go under `go test -cover`, and
// each `check` label is a subtest of its case.

// icDriver is the session's own protocol: a marker line on stdout per
// `sut` call, `check` verdict and `halt`.
const icDriver = `set -u
SCRIPT_DIR="$IC_SCRIPTS"
SUT="$SCRIPT_DIR/flow-guard"
OUT=""; RC=0
check() { if eval "$2"; then printf '\n__IC_PASS__ %s\n' "$1"; else printf '\n__IC_FAIL__ %s\n' "$1"; fi; }
halt() { printf '%s\n%s\n' "$OUT" "$(cat "$WORK/stderr")" >&2; printf '\n__IC_HALT__ %s\n' "$1"; exit 1; }
sut() {
  PATH="$WORK/bin:$PATH" GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.invalid \
    GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.invalid \
    FLOW_GUARD_REPO_ROOT="$(dirname "$(dirname "$SUT")")" env -0 >"$IC_TMP/sut.env"
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
  printf '\n__IC_FIXTURE__ %s\n' "$WORK"
  read -r _ <"$IC_TMP/sut.ack"
  ln "$IC_STUBS/flow" "$WORK/bin/flow"
}
fixture_build() {
  WORK="$(cd "$(mktemp -d "$IC_TMP/work.XXXXXX")" && pwd -P)"
  local r
  for r in a b; do
    g init -q --bare -b main "$WORK/$r-origin.git"
    g init -q -b main "$WORK/$r"
    printf 'base\n' >"$WORK/$r/base.txt"
    printf '.superpowers/\nspectre/changes/\n' >"$WORK/$r/.gitignore"
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

  mkdir -p "$WORK/bin" "$WORK/statedir"
  # Run 1 no longer calls spectre: any call fails the run.
  printf '#!/usr/bin/env bash\nexit 1\n' >"$WORK/bin/spectre"
  chmod +x "$WORK/bin/spectre"
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
  touch "$WORK/sr-completed"
}`

// icFlowStub is the fixture's fake flow CLI; it finds its fixture from its
// own path, so one body serves every case and WriteExec links each copy to
// one assessed inode.
const icFlowStub = `#!/usr/bin/env bash
W="${0%/bin/*}"
case "$1 $2" in
  "state get") cat "$W/state.json" ;;
  "state set") cat >"$W/state.json" ;;
  "state dir") echo "$W/statedir" ;;
  "state list") echo '{"source":"store","complete":true,"records":[{"name":"demo"}]}' ;;
  "stage begin"|"stage end") echo "$*" >>"$W/stages.log" ;;
  "stage completed") [ -f "$W/sr-unreachable" ] && exit 2; [ -f "$W/sr-completed" ] ;;
  "record findings") [ -f "$W/fail-findings" ] && exit 1; echo '[]' ;;
  "self-review findings") cat "$W/sr-findings.json" 2>/dev/null || echo '[]' ;;
  "self-review bundle") printf 'found: 1 of 1 sources; skipped: 0 of 1 sources\n' ;;
  "self-review finding") echo "$*" >>"$W/fixed.log" ;;
  "record dispatch") echo "$*" >>"$W/dispatch.log" ;;
  "record dispatches") cat "$W/dispatches.json" 2>/dev/null || echo '[{"key":"task-1-implementer","outcome":"completed"},{"key":"self-review-demo-pass","outcome":"completed"},{"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-02T00:00:00Z"}]' ;;
  "record decisions") echo '[]' ;;
  "record passes") echo '[{"id":1,"round":0,"note":"roster: full"}]' ;;
  "change save") echo "$*" >>"$W/saves.log"; [ -f "$W/fail-save" ] && { echo "missing: tasks.md"; exit 1; }; : ;;
  "workspace-id "*) echo demo-id ;;
  *) : ;;
esac
`

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
	guardtest.WriteExec(t, tmp+"/stubs/flow", icFlowStub)
	if err := syscall.Mkfifo(tmp+"/sut.ack", 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", icDriver+icHelpers+"\n"+script)
	cmd.Dir = tmp
	// HOME is the test's own, so no guard's commit reads the operator's
	// git config; the Go caches stay where the toolchain keeps them.
	cmd.Env = append(os.Environ(), "IC_TMP="+tmp, "IC_STUBS="+tmp+"/stubs", "IC_SCRIPTS="+scripts,
		"FLOW_GUARD_CACHE_DIR="+guardtest.GuardCache(t), "FLOW_GUARD_TELEMETRY=off",
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
		case strings.HasPrefix(l, "__IC_FIXTURE__ "):
			icFixture(t, tmp, strings.TrimPrefix(l, "__IC_FIXTURE__ "))
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
	v, err := icGoEnvOnce()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

var icGoEnvOnce = sync.OnceValues(func() ([]string, error) {
	out, err := exec.Command("go", "env", "GOCACHE", "GOMODCACHE", "GOPATH").Output()
	if err != nil {
		return nil, err
	}
	v := strings.Split(strings.TrimSpace(string(out)), "\n")
	return []string{"GOCACHE=" + v[0], "GOMODCACHE=" + v[1], "GOPATH=" + v[2]}, nil
})

// icTemplate is the harness's fixture, built once per package run by its own
// fixture_build under ExecFixtures.dir and the environment a case runs under;
// every case's fixture is a copy of it (icFixture). It returns the template's
// WORK.
var icTemplate = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp(guardtest.ExecFixtures.Dir, "ic-template")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("bash", "-c", icHelpers+"\nfixture_build && printf %s \"$WORK\"")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "IC_TMP="+dir, "HOME="+dir+"/home", "GIT_CONFIG_NOSYSTEM=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("fixture_build: %v\n%s", err, stderr.String())
	}
	return string(out), nil
})

// icFixture is one `fixture` call: the template copied to work, every
// absolute path to the template's WORK in it -- the origin URLs, the
// worktrees' gitdir links, state.json -- rewritten to work, then the
// session acknowledged. Objects are compressed and name no path.
func icFixture(t *testing.T, tmp, work string) {
	t.Helper()
	tpl, err := icTemplate()
	if err == nil {
		err = guardtest.CopyTree(tpl, work)
	}
	if err == nil {
		from, to := []byte(tpl), []byte(work)
		err = filepath.WalkDir(work, func(p string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case d.IsDir() && d.Name() == "objects":
				return filepath.SkipDir
			case !d.Type().IsRegular():
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil || !bytes.Contains(b, from) {
				return err
			}
			return os.WriteFile(p, bytes.ReplaceAll(b, from, to), 0)
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	icAck(t, tmp)
}

// icAck releases the session's `read` on sut.ack.
func icAck(t *testing.T, tmp string) {
	t.Helper()
	ack, err := os.OpenFile(tmp+"/sut.ack", os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	ack.WriteString("\n")
	ack.Close()
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
	env := core.Env{
		Getenv:    func(k string) string { return vars[k] },
		LookupEnv: func(k string) (string, bool) { v, ok := vars[k]; return v, ok },
		Environ:   func() []string { return environ },
		Dir:       "/",
	}
	var out, errOut bytes.Buffer
	guard := vars["IC_GUARD"] // a case driving another phase driver names it
	if guard == "" {
		guard = "integrate-change"
	}
	rc := Registry[guard](args, env, &out, &errOut)
	for name, b := range map[string][]byte{"sut.out": out.Bytes(), "sut.err": errOut.Bytes(), "sut.rc": []byte(strconv.Itoa(rc))} {
		if err := os.WriteFile(tmp+"/"+name, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	icAck(t, tmp)
}

// The fixture's flow stub is a link to the one inode WriteExec assessed, never
// a file written afresh: macOS assesses each new executable on its first exec,
// serialised machine-wide, and every case would pay it.
func TestIntegrateChangeFixtureLinksTheFlowStub(t *testing.T) {
	t.Parallel()
	icRun(t, `fixture
check "the fixture's flow stub is the assessed inode" '[ "$WORK/bin/flow" -ef "$IC_STUBS/flow" ]'`)
}

// Cost: 5.88 s — the two-repository prepare, commit and land the test proves, one chain no split can shorten.
// smoke: core — prepare, commit, land and cleanup over two repositories
func TestIntegrateChangeMarksLandAndCleanupOnly(t *testing.T) {
	icRunCases(t, []icCaseDef{
		{"happy_path", `# --- happy_path ---
fixture
sut prepare "$WORK/a" demo
check "prepare exits 0" '[ "$RC" -eq 0 ]'
check "prepare names the configured route" 'printf "%s" "$OUT" | grep -q "^ROUTE: merge and push — from this project"'
check "prepare prints each worktree's reshape base" '[ "$(printf "%s\n" "$OUT" | grep -c "^WORKTREE: ")" -eq 2 ]'
check "prepare ends on NEXT: commit" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: flow-guard integrate-change commit "'
[ "$RC" -eq 0 ] || halt prepare

printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "commit exits 0" '[ "$RC" -eq 0 ]'
check "commit ends on NEXT: land" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: flow-guard integrate-change land"'
[ "$RC" -eq 0 ] || halt commit
touch "$WORK/sr-completed"

sut land "$WORK/a" demo
check "land exits 0" '[ "$RC" -eq 0 ]'
[ "$RC" -eq 0 ] || halt land
check "land stops at the self-review NEXT" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run the self-review pass .*then flow-guard integrate-change cleanup"'
check "prepare, commit and land begin one flow.land and end it completed, then land begins flow.self-review" \
  '[ "$(grep -c "stage begin" "$WORK/stages.log")" = 2 ] && grep -q "stage begin .*-stage flow.self-review" "$WORK/stages.log" && grep -q "stage begin .*-stage flow.land" "$WORK/stages.log" && grep -q "stage end .*-stage flow.land -outcome completed" "$WORK/stages.log"'
check "b was pushed before a (merge order)" \
  'printf "%s\n" "$OUT" | grep "^PUSHED: " | head -1 | grep -q "/b-worktrees/demo"'
sut cleanup "$WORK/a" demo
check "cleanup exits 0" '[ "$RC" -eq 0 ]'
check "cleanup saved the change from its canonical worktree" 'grep -qx "change save -C $WORK/a-worktrees/demo demo" "$WORK/saves.log"'
check "a's origin main carries the implementation commit alone" \
  '[ "$(git --git-dir="$WORK/a-origin.git" log --format=%s main -2 | tr "\n" "|")" = "feat(a): do the work|base|" ]'
check "a's origin main holds no change directory" \
  '[ -z "$(git --git-dir="$WORK/a-origin.git" ls-tree -d main spectre/changes)" ]'
check "b's origin main carries its implementation commit" \
  '[ "$(git --git-dir="$WORK/b-origin.git" log --format=%s main -1)" = "feat(b): do the work" ]'
check "both worktrees are removed" '[ ! -d "$WORK/a-worktrees/demo" ] && [ ! -d "$WORK/b-worktrees/demo" ]'
check "the record is FINISHED with an empty worktrees map" \
  '[ "$(jq -c "[.state, .worktrees]" "$WORK/state.json")" = "[\"FINISHED\",{}]" ]'
check "Jira Done is handed to the parent" \
  '[ "$(printf "%s\n" "$OUT" | grep "^JIRA: " | tr "\n" "|")" = "JIRA: transition KAN-1 to Done|" ]'
check "cleanup begins one flow.cleanup after land's flow.land and ends it completed" \
  '[ "$(grep -c "stage begin" "$WORK/stages.log")" = 3 ] && grep -q "stage begin .*-stage flow.cleanup" "$WORK/stages.log" && tail -1 "$WORK/stages.log" | grep -q "stage end .*-stage flow.cleanup -outcome completed"'
check "cleanup ends on NEXT: Finished handoff" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: print the Finished handoff"'
check "the NEXT line names the template and no line the template lacks" 'printf "%s" "$OUT" | tail -1 | grep -qxF "NEXT: print the Finished handoff (skills/flow/cleanup.md)"'`},
		{"run2_preflight", `# --- run2_preflight ---
fixture
for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q origin HEAD:main; done
sut prepare "$WORK/a" demo
check "run2_preflight: a merged branch sends prepare to run 2" 'printf "%s" "$OUT" | tail -1 | grep -q "^NEXT: run 2"'
check "run2_preflight: run 2's prepare begins no mark" '! grep -q "stage begin" "$WORK/stages.log" 2>/dev/null'`},
	})
}

// The bootstrap run outside sut, from inside the canonical
// worktree, with a relative main checkout.
func TestIntegrateChangeShim(t *testing.T) {
	smoke.Heavy(t)
	icRunCases(t, []icCaseDef{
		{"shim_caller_cwd", `# --- shim_caller_cwd ---
fixture
for r in a b; do
  g -C "$WORK/$r-worktrees/demo" push -q origin HEAD:main
  g -C "$WORK/$r" fetch -q origin
done
touch "$WORK/sr-completed"
OUT="$(cd "$WORK/a-worktrees/demo" && PATH="$WORK/bin:$PATH" exec "$SUT" integrate-change cleanup ../../a demo --harness test --session-token mf-test 2>"$WORK/stderr")"; RC=$?
check "shim_caller_cwd: cleanup run from inside a worktree is not HELD by its own cwd" '! printf "%s\n" "$OUT" | grep -q "^HELD:"'
check "shim_caller_cwd: a relative main checkout resolves against the caller's directory" \
  '[ "$RC" -eq 0 ] && [ ! -d "$WORK/a-worktrees/demo" ] && [ "$(jq -r .state "$WORK/state.json")" = FINISHED ]'`},
	})
}
