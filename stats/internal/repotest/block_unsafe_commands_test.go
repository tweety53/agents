package repotest

// TestBlockUnsafeCommands ports the retired shell harness, the
// harness for hooks/block-unsafe-commands.py.
//
// It builds one bare origin with a clone whose main checkout sits on `main`
// (origin/HEAD set), a linked worktree of it on a feature branch, and one repo
// with no origin at all, then feeds the hook PreToolUse payloads and asserts
// deny/allow on each. The cases share one fixture the test body mutates
// between them (adding worktrees, flipping a file, moving the main checkout
// off main); a hook call never mutates it, so the cases between two mutations
// run concurrently and each mutation waits for them (pmFixture.batch), one
// subtest per harness case; the case numbers are the harness's own. In command templates {ROOT}, {MAIN},
// {SIB} and {WT} stand for the fixture paths;
// every other `$NAME` is literal shell text the hook sees unexpanded.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

var denyRe = regexp.MustCompile(`"permissionDecision": *"deny"`)

type pmFixture struct {
	t    *testing.T
	home string
	vars map[string]string
	// pending holds the cases check queued since the last flush; a test
	// defers f.flush() so the last ones run before it returns.
	pending []pmCase
}

type pmCase struct {
	name string
	fn   func(t *testing.T)
}

// flush runs the queued cases concurrently, as flat subtests, and returns once
// all have finished: the barrier every fixture mutation sits behind.
func (f *pmFixture) flush() {
	var wg sync.WaitGroup
	for _, c := range f.pending {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.t.Run(c.name, c.fn)
		}()
	}
	wg.Wait()
	f.pending = nil
}

func (f *pmFixture) x(s string) string {
	for k, v := range f.vars {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}

func (f *pmFixture) git(args ...string) string {
	f.t.Helper()
	f.flush()
	return sh(f.t, "", []string{"HOME=" + f.home}, "git",
		append([]string{"-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)...)
}

func (f *pmFixture) write(path, body string) {
	f.t.Helper()
	f.flush()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		f.t.Fatal(err)
	}
}

// run feeds the hook one payload and returns its stdout.
func (f *pmFixture) run(t *testing.T, cwd, tool string, input map[string]string, env ...string) string {
	t.Helper()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(map[string]any{"tool_name": tool, "tool_input": input, "cwd": cwd}); err != nil {
		t.Fatal(err)
	}
	out, stderr, rc := runHook(t, "block-unsafe-commands.py", buf.Bytes(), env...)
	if rc != 0 {
		// a crashing hook prints nothing, which would read as allow
		t.Fatalf("hook exited %d: %s", rc, stderr)
	}
	return out
}

func decision(out string) string {
	if denyRe.MatchString(out) {
		return "deny"
	}
	return "allow"
}

// check queues one case, run as a subtest by the next flush.
func (f *pmFixture) check(name string, fn func(t *testing.T)) {
	f.pending = append(f.pending, pmCase{name, fn})
}

func (f *pmFixture) expect(name, want, cwd, tool string, input map[string]string, env ...string) {
	f.t.Helper()
	cwd = f.x(cwd)
	f.check(name, func(t *testing.T) {
		if got := decision(f.run(t, cwd, tool, input, env...)); got != want {
			t.Errorf("wanted %s, got %s", want, got)
		}
	})
}

func (f *pmFixture) bash(name, want, cwd, command string, env ...string) {
	f.t.Helper()
	f.expect(name, want, cwd, "Bash", map[string]string{"command": f.x(command)}, env...)
}

func (f *pmFixture) writeTool(name, want, path string) {
	f.t.Helper()
	f.expect(name, want, "{ROOT}", "Write", map[string]string{"file_path": f.x(path), "content": "x"})
}

func (f *pmFixture) editTool(name, want, path string) {
	f.t.Helper()
	f.expect(name, want, "{ROOT}", "Edit", map[string]string{"file_path": f.x(path), "old_string": "a", "new_string": "b"})
}

// newPMFixture builds one bare origin with a clone {MAIN} on `main` (origin/HEAD set,
// f.txt tracked) and a linked worktree {WT} of it on `change`.
func newPMFixture(t *testing.T) (f *pmFixture, root, main, wt string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f = &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	main = root + "/main"
	wt = main + "/.worktrees/change"
	f.vars["MAIN"], f.vars["WT"] = main, wt

	// origin + main checkout on main
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")
	f.git("-C", main, "worktree", "add", wt, "-b", "change")
	return f, root, main, wt
}

// TestBlockUnsafeCommandsScopesACdToItsGroup: a `cd` inside `( … )` or `$( … )` ends at its
// close, so a redirect after the group lands in the cwd before it; and `trash` of a file git
// tracks in a main checkout is a write, of an untracked one is not (kan-1083 self-review, F2).
func TestBlockUnsafeCommandsScopesACdToItsGroup(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f, _, main, _ := newPMFixture(t)
	defer f.flush()
	f.write(main+"/scratch.out", "x\n")
	f.bash("redirect after a cd group lands in main", "deny", "{MAIN}", `(cd {WT} && true) > c56.out 2> c56.out.fix`)
	f.bash("redirect after a cd command substitution lands in main", "deny", "{MAIN}", `x=$(cd {WT} && pwd) && echo "$x" > f.txt`)
	f.bash("a cd in a quoted substitution ends at its close", "deny", "{MAIN}", "echo \"$(cd {WT}; pwd)\"\necho x > f.txt")
	f.bash("redirect inside the cd group lands in the worktree", "allow", "{MAIN}", `(cd {WT} && echo x > f.txt)`)
	f.bash("redirect glued to a group's close lands in main", "deny", "{MAIN}", `(cd {WT} && true)>c.out`)
	f.bash("append glued to a group's close lands in main", "deny", "{MAIN}", `(cd {WT} && true)>>c.out`)
	f.bash("&> glued to a group's close lands in main", "deny", "{MAIN}", `(cd {WT} && true)&>c.out`)
	f.bash("&& glued to a group's close ends the cd", "deny", "{MAIN}", `(cd {WT} && true)&& echo > c.out`)
	f.bash("| glued to a group's close ends the cd", "deny", "{MAIN}", `(cd {WT} && true)| tee c.out`)
	f.bash(";) closes the group", "deny", "{MAIN}", `(cd {WT}; true;) > c.out`)
	f.bash("a cd before a group still holds after it", "allow", "{ROOT}", `cd {WT} && (true) > f.txt`)
	f.bash("trash of a tracked file in main", "deny", "{ROOT}", `trash {MAIN}/f.txt`)
	f.bash("trash of an untracked file in main", "allow", "{ROOT}", `trash {MAIN}/scratch.out`)
}

func TestBlockUnsafeCommands(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f, root, main, wt := newPMFixture(t)
	defer f.flush()
	// repo with no origin, on develop
	f.git("init", "-q", "-b", "develop", root+"/noorigin")
	f.write(root+"/noorigin/f.txt", "a\n")
	f.git("-C", root+"/noorigin", "add", "f.txt")
	f.git("-C", root+"/noorigin", "commit", "-q", "-m", "init")

	f.writeTool("1 Write inside main checkout on main", "deny", "{MAIN}/new.txt")
	f.editTool("2 Edit inside linked worktree", "allow", "{WT}/f.txt")
	f.bash("4 Bash git commit with cwd main", "deny", "{MAIN}", `git commit -m x`)
	f.bash("5 Bash git -C main add from elsewhere", "deny", "{ROOT}", `git -C {MAIN} add .`)
	f.bash("6 Bash cd main && git reset", "deny", "{ROOT}", `cd {MAIN} && git reset HEAD~1`)
	f.bash("7 Bash git -C main pull --ff-only", "allow", "{ROOT}", `git -C {MAIN} pull --ff-only`)
	f.bash("8 Bash git -C main log", "allow", "{ROOT}", `git -C {MAIN} log --oneline -3`)
	f.bash("9 Bash git -C worktree commit", "allow", "{ROOT}", `git -C {WT} commit -m x`)
	f.bash("10 Bash sed -i into main", "deny", "{ROOT}", `sed -i '' s/a/b/ {MAIN}/f.txt`)
	f.bash("11 Bash redirect into main", "deny", "{ROOT}", `echo x > {MAIN}/f.txt`)
	f.bash("12 Bash cat from main", "allow", "{ROOT}", `cat {MAIN}/f.txt`)
	f.writeTool("13 Write outside any repo", "allow", "{ROOT}/loose.txt")
	f.check("14 malformed JSON passes through", func(t *testing.T) {
		out, _, rc := runHook(t, "block-unsafe-commands.py", []byte("not json"))
		if rc != 0 || strings.TrimRight(out, "\n") != "" {
			t.Errorf("rc=%d out=%s", rc, out)
		}
	})
	f.writeTool("15 no-origin repo on develop", "deny", "{ROOT}/noorigin/g.txt")
	f.bash("16 Bash git worktree add from main", "allow", "{ROOT}", `git -C {MAIN} worktree add {MAIN}/.worktrees/other -b other origin/main`)
	f.bash("17 Bash git -C main stash", "deny", "{ROOT}", `git -C {MAIN} stash`)
	f.bash("18 Bash tee into main", "deny", "{ROOT}", `echo x | tee {MAIN}/f.txt`)
	f.bash("19 Bash rm in main", "deny", "{ROOT}", `rm {MAIN}/f.txt`)
	f.bash("20 landing script named, not a git verb", "allow", "{ROOT}", `land-self-review-report.sh {MAIN} main 'docs: x' docs/x.md --push main`)

	f.write(root+"/elsewhere/x", "")
	f.bash("21 tilde destination is not cwd-relative", "allow", "{MAIN}", `mv {ROOT}/elsewhere/x ~/.Trash/`, "HOME="+root)
	f.bash("22 mv out of main checkout", "deny", "{ROOT}", `mv {MAIN}/f.txt {ROOT}/`)

	f.bash("23 cd worktree; git commit", "allow", "{ROOT}", `cd {WT}; git commit -m x`)
	f.bash("24 cd worktree; redirect to an unexpanded var", "allow", "{MAIN}", `S=/tmp/x; cd {WT}; npx playwright test a.spec.ts >$S/cap.log 2>&1; echo cap=$?`)
	f.bash("25 cd main;git reset with separators touching", "deny", "{ROOT}", `cd {MAIN};git reset HEAD~1`)
	f.bash("26 multi-line cp cannot swallow a later line's target", "allow", "{ROOT}", "cp notes.md draft.md\necho done\ngit add {MAIN}/README.md")
	f.bash("27 multi-line second line judged on its own", "deny", "{ROOT}", "echo start\ncd {MAIN} && git reset HEAD~1")
	f.bash("28 assignment-set redirect into main resolves and denies", "deny", "{ROOT}", `C={MAIN}; echo x > $C/f.txt`)
	f.bash("29 nested variable stays let-through", "allow", "{ROOT}", `B={MAIN}; A=$B; echo x > $A/f.txt`)
	f.bash("30 cp into a not-yet-existing worktree path", "allow", "{ROOT}", `cp /tmp/a.md {MAIN}/.worktrees/new-landing/spectre/foo.md`)
	f.writeTool("31 loose file directly in .worktrees", "allow", "{MAIN}/.worktrees/loose.txt")
	f.bash("32 cd state threads across lines without a separator", "deny", "{ROOT}", "cd {MAIN}\ngit reset HEAD~1")
	f.bash("33 longest-name expansion wins the collision", "deny", "{ROOT}", `CA={MAIN}; C=/tmp; echo x > $CA/f.txt`)
	f.bash("34 braced expansion resolves the variable", "deny", "{ROOT}", `C={MAIN}; echo x > ${C}/f.txt`)
	f.bash("35 rm of the .worktrees root itself stays denied", "deny", "{ROOT}", `rm -rf {MAIN}/.worktrees`)
	f.bash("36 unset near-name variable is not expanded greedily", "allow", "{ROOT}", `D={MAIN}/docs; echo x > $Dy/f.txt`)
	f.bash("37 use before set is not expanded", "allow", "{ROOT}", "echo x > $C/f.txt\nC={MAIN}")

	f.write(root+"/outside.txt", "x\n")
	f.bash("43 BSD sed -i with an empty suffix, file outside", "allow", "{MAIN}", `sed -i '' 's/x/y/' {ROOT}/outside.txt`)

	// 44-47: the sibling layout, <dirname main>/<basename main>-worktrees/<change>
	f.git("-C", main, "worktree", "add", root+"/main-worktrees/sib", "-b", "sib")
	f.editTool("44 sibling layout: Edit in an existing sibling worktree", "allow", "{ROOT}/main-worktrees/sib/f.txt")
	f.git("init", "-q", "-b", "develop", root+"/noorigin/inner")
	for _, d := range []string{"/noorigin/inner-worktrees", "/noorigin/lib", "/noorigin/lib-worktrees"} {
		if err := os.MkdirAll(root+d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	f.writeTool("45 sibling layout: Write into a not-yet-existing sibling worktree", "allow", "{ROOT}/noorigin/inner-worktrees/new/x")
	f.writeTool("47 sibling layout: <X>-worktrees beside a plain directory is main-checkout content", "deny", "{ROOT}/noorigin/lib-worktrees/new/x")
	f.check("46 sibling layout: deny reason suggests the sibling path", func(t *testing.T) {
		reason := f.run(t, root, "Write", map[string]string{"file_path": main + "/new.txt", "content": "x"})
		if !strings.Contains(reason, "worktree add "+root+"/main-worktrees/<change>") || strings.Contains(reason, ".worktrees") {
			t.Errorf("deny reason does not suggest %s/main-worktrees/<change>: %s", root, reason)
		}
	})

	f.vars["SIB"] = root + "/main-worktrees/sib"
	f.bash("48 heredoc append into a sibling worktree, body naming rm and >", "allow", "{MAIN}", "cat >> {SIB}/f.txt <<'EOF'\nthen rm the stale file and cp it > there\nEOF")
	f.bash("49 rm of an absolute sibling-worktree path from main", "allow", "{MAIN}", `rm {SIB}/tmp`)
	f.bash("50 relative rm with cwd main", "deny", "{MAIN}", `rm f.txt`)
	f.bash("51 heredoc write into main", "deny", "{ROOT}", "cat > {MAIN}/f.txt <<EOF\nx\nEOF")
	f.bash("52 a line after the heredoc is scanned again", "deny", "{MAIN}", "cat >> {SIB}/f.txt <<-EOF\n\tx\n\tEOF\nrm f.txt")
	f.bash("53 heredoc fed to a shell stays scanned", "deny", "{ROOT}", "bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")

	f.bash("54 a quoted << opens no heredoc", "deny", "{MAIN}", "echo \"<<EOF\"\nrm f.txt\nEOF")
	f.bash("55 a commented << opens no heredoc", "deny", "{MAIN}", "true # <<EOF\nrm f.txt\nEOF")
	f.bash("56 an arithmetic shift opens no heredoc", "deny", "{MAIN}", "echo $((1<<N))\nrm f.txt")
	f.bash("57 a partly quoted delimiter is unquoted whole", "deny", "{MAIN}", "cat <<E\"OF\" >/dev/null\nx\nEOF\nrm f.txt")
	f.bash("58 a shell named by path keeps its body scanned", "deny", "{MAIN}", "/bin/bash <<EOF\nrm f.txt\nEOF")
	f.bash("59 a body sourced by . stays scanned", "deny", "{MAIN}", ". /dev/stdin <<EOF\nrm f.txt\nEOF")
	f.bash("60 dash fed a heredoc keeps its body scanned", "deny", "{MAIN}", "dash <<EOF\nrm f.txt\nEOF")
	f.bash("61 a heredoc piped into a shell keeps its body scanned", "deny", "{MAIN}", "cat <<EOF | dash\nrm f.txt\nEOF")
	f.bash("62 $SHELL fed a heredoc keeps its body scanned", "deny", "{MAIN}", "$SHELL <<EOF\nrm f.txt\nEOF")
	f.bash("63 a heredoc captured then eval'd keeps its body scanned", "deny", "{MAIN}", "x=$(cat <<EOF\nrm f.txt\nEOF\n)\neval \"$x\"")
	f.bash("64 a heredoc read then eval'd keeps its body scanned", "deny", "{MAIN}", "read -r -d '' c <<EOF\nrm f.txt\nEOF\neval \"$c\"")
	f.bash("65 a heredoc written to a script then run keeps its body scanned", "deny", "{MAIN}", "cat <<EOF > {SIB}/s.sh\nrm f.txt\nEOF\nsh {SIB}/s.sh")
	f.bash("66 a CRLF closing line does not end the scan", "deny", "{MAIN}", "cat >> {SIB}/f.txt <<EOF\nx\nEOF\r\nrm f.txt")
	f.bash("67 a << inside a parameter expansion opens no heredoc", "deny", "{MAIN}", "echo ${x:-<<EOF}\nrm f.txt\nEOF")
	f.bash("68 tee -a heredoc into a sibling worktree, body naming rm", "allow", "{MAIN}", "tee -a {SIB}/f.txt <<'EOF'\nthen rm the stale file\nEOF")
	f.bash("69 heredoc append into main", "deny", "{ROOT}", "cat >> {MAIN}/f.txt <<EOF\nx\nEOF")
	f.bash("70 an unquoted heredoc body's command substitution stays scanned", "deny", "{MAIN}", "cat >> {SIB}/f.txt <<EOF\n$(rm f.txt)\nEOF")
	f.bash("71 a report heredoc in a longer command keeps its body as data", "allow", "{MAIN}", "D={SIB}; cat > $D/r.md <<'EOF'\nthe row reads > the old one\nEOF\necho written")
	f.bash("72 a report heredoc in a longer command still judges its target", "deny", "{ROOT}", "D={MAIN}; cat > $D/r.md <<'EOF'\nx\nEOF\necho written")
	f.bash("73 a heredoc into a process substitution keeps its body scanned", "deny", "{ROOT}", "echo start; cat > >(sh) <<'EOF'\nrm {MAIN}/f.txt\nEOF\necho done")
	f.bash("74 a tee heredoc feeding a process substitution keeps its body scanned", "deny", "{ROOT}", "echo start; tee /dev/null >(bash) <<'EOF'\nrm {MAIN}/f.txt\nEOF\necho done")
	f.bash("75 a heredoc inside bash <( keeps its body scanned", "deny", "{ROOT}", "bash <(\ncat > /dev/stdout <<'EOF'\nrm {MAIN}/f.txt\nEOF\n)")
	f.bash("76 a heredoc written then run under another spelling keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\nsh \"{SIB}\"/x.sh")

	f.flush()
	if err := os.MkdirAll(main+"/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, l := range []string{"/elsewhere/lsub", "/elsewhere/lsub2"} {
		if err := os.Symlink(main+"/sub", root+l); err != nil {
			t.Fatal(err)
		}
	}
	f.bash("102 a symlink into the checkout, then .., resolves before judging", "deny", "{ROOT}", `cat > {ROOT}/elsewhere/lsub/../f.txt`)
	f.bash("103 rm and mv of a symlink into the checkout act on the link, outside it", "allow", "{ROOT}", `rm {ROOT}/elsewhere/lsub2 && mv {ROOT}/elsewhere/lsub {ROOT}/elsewhere/l3`)
	f.bash("110 an unquoted data heredoc's literal redirect beside a backtick span is data", "allow", "{MAIN}", "cat > {SIB}/r.md <<EOF\nsee `a` > b\nEOF")
	f.bash("111 a backtick span in an unquoted data heredoc body is still scanned", "deny", "{MAIN}", "cat > {SIB}/r.md <<EOF\nsee `rm {MAIN}/x` here\nEOF")
	f.bash("112 a backtick span is scanned as a command", "deny", "{ROOT}", "echo `rm {MAIN}/x`")
	f.bash("113 a command substitution inside a double-quoted word is scanned", "deny", "{ROOT}", `echo "$(rm {MAIN}/x)"`)
	f.bash("114 a backtick cp in an unquoted data heredoc body is scanned", "deny", "{ROOT}", "cat > {SIB}/r.md <<EOF\nsee `cp a {MAIN}/y` here\nEOF")
	f.bash("115 a cp inside an unquoted $(...) is scanned", "deny", "{ROOT}", `echo $(cp a {MAIN}/y)`)
	f.bash("116 a cp inside a subshell is scanned", "deny", "{ROOT}", `(cp a {MAIN}/y)`)
	f.bash("117 an escaped backtick does not pair with a real span", "deny", "{MAIN}", "cat > {SIB}/r.md <<EOF\n\\` `rm {MAIN}/x` \\`\nEOF")
	planBody := "C={SIB}; cat > $C/t.md <<'EOF'\n> **Execution:** a > b; rm x\nEOF\n"
	f.bash("118 a data heredoc's target named again as an argument keeps its body as data", "allow", "{MAIN}", planBody+"check-references.sh $C/t.md")
	f.bash("119 a path-named command naming another file keeps the body as data", "allow", "{MAIN}", planBody+"~/.claude/skills/flow/scripts/lint.sh /x/o.md")
	f.bash("120 a data heredoc's target run as a command keeps its body scanned", "deny", "{MAIN}", planBody+"chmod +x $C/t.md && $C/t.md")
	xsh := "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n"
	f.bash("121 a heredoc's file name run under another path keeps its body scanned", "deny", "{ROOT}", xsh+"/elsewhere/x.sh")
	for _, c := range []struct{ n, run string }{
		{"122", "/bin/bash {SIB}/x.sh"}, {"123", "/bin/sh {SIB}/x.sh"}, {"124", "cd {SIB} && x.sh"},
		{"125", "awk -f {SIB}/x.sh"}, {"126", "make -f {SIB}/x.sh"}, {"127", "watch {SIB}/x.sh"},
		{"128", "crontab {SIB}/x.sh"}, {"129", "php {SIB}/x.sh"}, {"130", "setsid {SIB}/x.sh"},
		{"131", "find {SIB} -name x.sh -exec {} ;"}, {"132", "ssh localhost {SIB}/x.sh"},
		{"133", "psql -f {SIB}/x.sh"}, {"134", "git -c core.pager=x show {SIB}/x.sh"},
		{"135", "rg --pre {SIB}/x.sh x f"},
	} {
		f.bash(c.n+" a heredoc body then '"+c.run+"' keeps its body scanned", "deny", "{ROOT}", xsh+c.run)
	}
	f.bash("136 a written check script run afterwards keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/check-a.sh <<EOF\nrm {MAIN}/f.txt\nEOF\n{SIB}/check-a.sh")
	f.bash("137 a heredoc body then 'GIT_EXTERNAL_DIFF={SIB}/x.sh git diff' keeps its body scanned", "deny", "{ROOT}", xsh+"GIT_EXTERNAL_DIFF={SIB}/x.sh git diff --ext-diff a b")

	// 110-113: a `go test` over a whole heavy package is the operator's to ask for; a named test run is not.
	f.bash("110 go test over the whole guard package", "deny", "{ROOT}", `cd {SIB} && go test ./internal/guard/`)
	f.bash("111 go test over ./... with -count", "deny", "{ROOT}", `cd {SIB}/stats && go test -count=1 ./...`)
	f.bash("112 go test of the guard package naming its tests", "allow", "{ROOT}", `cd {SIB}/stats && go test ./internal/guard/ -run 'PairJudge|PjSample'`)
	f.bash("113 go test of a light package", "allow", "{ROOT}", `cd {SIB}/stats && go test ./internal/modelpolicy/`)

	// 143-146: the main session runs the list itself, the recording form included (no runner agent).
	rec := `{SIB}/scripts/flow-guard run-verify-list {SIB} lint kan-1 verify mf-abc123 haiku low`
	f.bash("143 run-verify-list recording form from the main session", "allow", "{ROOT}", rec)
	f.bash("144 run-verify-list bare form from the main session", "allow", "{ROOT}", `{SIB}/scripts/flow-guard run-verify-list {SIB} lint`)
	for _, c := range []struct{ n, key, val string }{{"145 recording form from a subagent (agent_id)", "agent_id", "a1b2c3"}, {"146 recording form from a subagent (transcript)", "transcript_path", "/x/s/subagents/agent-a1.jsonl"}} {
		f.check(c.n, func(t *testing.T) {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": f.x(rec)}, "cwd": f.x("{ROOT}"), c.key: c.val}); err != nil {
				t.Fatal(err)
			}
			out, stderr, rc := runHook(t, "block-unsafe-commands.py", buf.Bytes())
			if rc != 0 || decision(out) != "allow" {
				t.Errorf("wanted allow, got rc %d out %q %s", rc, out, stderr)
			}
		})
	}

	// 138-142: a `>` inside quotes is data, never a redirect (kan-1003's self-review); every
	// operator that writes a file is still judged.
	f.bash("138 a quoted '>Save<' argument, cwd = main", "allow", "{MAIN}", `python3 grab.py "{SIB}/a.html" '>Save<' 1 1200`)
	f.bash("139 a python heredoc body finding '>Save<', cwd = main", "allow", "{MAIN}", "python3 - <<'EOF'\nprint(open('{SIB}/a.html').read().find('>Save<'))\nEOF")
	f.bash("140 '>|' into the main checkout", "deny", "{ROOT}", `echo x >| {MAIN}/f.txt`)
	f.bash("141 '&>' into the main checkout", "deny", "{ROOT}", `echo x &> {MAIN}/f.txt`)
	f.bash("142 an unbalanced quote, then '>' glued to a main-checkout path", "deny", "{ROOT}", `echo it's >{MAIN}/f.txt`)
	f.bash("149 '>&' into the main checkout", "deny", "{ROOT}", `echo x >& {MAIN}/f.txt`)

	// 163-170: a body no shell runs is not shell text (kan-1026's self-review: eleven worktree-only
	// commands denied); its git invocations are still judged.
	f.bash("163 a python heredoc body comparing d>0, cwd = main", "allow", "{MAIN}", "python3 -I - x.png <<'EOF'\nif v<230 or d>0: print(y)\nEOF")
	f.bash("164 a python heredoc body running git commit in main", "deny", "{ROOT}", "python3 - <<'EOF'\nsubprocess.run([\"git\", \"-C\", \"{MAIN}\", \"commit\", \"-m\", \"x\"])\nEOF")
	f.bash("165 a report heredoc then echo of an unassigned name, cwd = main", "allow", "{MAIN}", "D={SIB}; cat > $D/r.md <<'EOF'\ncrop from `x-<n>.png` here\nEOF\necho $N")
	f.bash("167 a python file written then run by python3, cwd = main", "allow", "{MAIN}", "O={SIB}; cat > $O/t.py <<'EOF'\nif a > 0: open(p, 'w')\nEOF\npython3 $O/t.py {SIB}")
	f.bash("168 a spec written then run by npx playwright test, cwd = main", "allow", "{MAIN}", "cat > {SIB}/a.spec.ts <<'EOF'\nconst k = (x) => x > 0\nEOF\ncd {SIB} && U=http://x npx playwright test a.spec.ts")
	f.bash("169 a python file written then run by bash keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/t.py <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/t.py")
	f.bash("170 a report heredoc then a substitution running it keeps its body scanned", "deny", "{ROOT}", "cat > {SIB}/r.md <<'EOF'\nrm {MAIN}/f.txt\nEOF\nX=$({SIB}/r.md)")

	// 151-155: a restore that would discard unstaged edits in a linked worktree (kan-1006's self-review)
	f.write(wt+"/f.txt", "flipped\n")
	f.bash("151 checkout -- of a path with unstaged edits", "deny", "{ROOT}", `git -C {WT} checkout -- f.txt`)
	f.bash("152 restore of a path with unstaged edits", "deny", "{ROOT}", `git -C {WT} restore f.txt`)
	f.bash("153 checkout HEAD -- after a cd", "deny", "{ROOT}", `cd {WT} && git checkout HEAD -- f.txt`)
	f.bash("154 restore --staged touches the index only", "allow", "{ROOT}", `git -C {WT} restore --staged f.txt`)
	f.bash("156 restore with stderr sent to /dev/null", "deny", "{ROOT}", `git -C {WT} restore f.txt 2>/dev/null`)
	f.bash("157 checkout -- with stdout sent to /dev/null", "deny", "{ROOT}", `git -C {WT} checkout -- f.txt >/dev/null`)
	f.git("-C", wt, "add", "f.txt")
	f.bash("158 checkout HEAD -- of a staged-only edit", "deny", "{ROOT}", `git -C {WT} checkout HEAD -- f.txt`)
	f.bash("159 restore --source=HEAD --staged --worktree of a staged-only edit", "deny", "{ROOT}", `git -C {WT} restore --source=HEAD --staged --worktree f.txt`)
	f.git("-C", wt, "reset", "-q", "f.txt")
	// kan-1155: a worktree the change's state record lists may restore its tracked paths
	stub := root + "/bin"
	f.write(stub+"/flow", "#!/bin/sh\necho '{\"worktrees\":{\""+wt+"\":null}}'\n")
	f.bash("160 restore in a worktree the change state lists", "allow", "{ROOT}", `git -C {WT} restore f.txt`, "PATH="+stub+":"+os.Getenv("PATH"))
	f.bash("161 restore in the main checkout stays denied", "deny", "{ROOT}", `git -C {MAIN} restore f.txt`, "PATH="+stub+":"+os.Getenv("PATH"))
	// a change spans sibling repos: its record lives under another project, asked via that project's worktree
	f.write(root+"/other-worktrees/sib/x", "x\n")
	f.write(root+"/main-worktrees/sib/f.txt", "flipped\n")
	f.write(stub+"/flow", "#!/bin/sh\n[ \"$4\" = \""+root+"/other-worktrees/sib\" ] && echo '{\"worktrees\":{\""+root+"/main-worktrees/sib\":null}}'\n")
	f.bash("162 restore in a sibling worktree the change state lists under another project", "allow", "{ROOT}", `git -C {SIB} restore f.txt`, "PATH="+stub+":"+os.Getenv("PATH"))
	f.git("-C", wt, "checkout", "HEAD", "--", "f.txt")
	// a drifted PNG capture baseline restores to base even outside the change's state
	f.write(wt+"/b.png", "base\n")
	f.git("-C", wt, "add", "b.png")
	f.git("-C", wt, "commit", "-q", "-m", "baseline")
	f.write(wt+"/b.png", "drifted\n")
	f.bash("162 restore of a drifted PNG baseline", "allow", "{ROOT}", `git -C {WT} restore b.png`)
	f.bash("155 checkout -- of a clean path", "allow", "{ROOT}", `git -C {WT} checkout -- f.txt`)

	// 3 last: moving the main checkout off main lifts the protection
	f.git("-C", main, "checkout", "-q", "-b", "feature")
	f.writeTool("3 Write in main checkout on a feature branch", "allow", "{MAIN}/new.txt")
}

// TestBlockUnsafeCommandsDeniesABareCatRedirect pins kan-1033's finding 4.2:
// a bare `cat > <file>` reads stdin, which a Bash call never has, so it
// hangs until the call times out; a cat with input of its own passes.
func TestBlockUnsafeCommandsDeniesABareCatRedirect(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	defer f.flush()
	for _, c := range []string{"cat > /tmp/x.go", "cat >> f", "cat >f; echo", "a && cat > f", "false || cat > f", "(cat > f)"} {
		f.check("deny "+c, func(t *testing.T) {
			out := f.run(t, root, "Bash", map[string]string{"command": c})
			if decision(out) != "deny" || !strings.Contains(out, "Write tool") || !strings.Contains(out, "heredoc") {
				t.Errorf("out=%s", out)
			}
		})
	}
	for _, c := range []string{"cat > f <<'EOF'\nbody\nEOF", "cat > f < in", "x | cat > f", "cat a > f", "cat -n f > g", "echo cat > f", "concat > f"} {
		f.bash("allow "+c, "allow", "{ROOT}", c)
	}
}

// TestBlockUnsafeCommandsReadsADataHeredocBodyAsData pins kan-1025: a heredoc
// body fed to a command that runs no code (`flow record summary -file -`) is
// data, not shell — while a body handed to a shell, a later runner word, a pipe
// into `sh`, or an unquoted body's `$(…)` keeps being scanned.
func TestBlockUnsafeCommandsReadsADataHeredocBodyAsData(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	defer f.flush()
	main := root + "/main"
	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")

	f.bash("data body from main", "allow", "{MAIN}", "flow record summary -change x -file - <<'EOF'\nstep 1 -> step 2; rm {MAIN}/f.txt > {MAIN}/g.txt\nEOF")
	f.bash("flow opener names a shell", "deny", "{MAIN}", "flow run csh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("bash body", "deny", "{ROOT}", "bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("sudo sh body", "deny", "{ROOT}", "sudo sh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("substituted sh body", "deny", "{ROOT}", "x=$(sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n)")
	f.bash("unquoted body substitution", "deny", "{ROOT}", "flow record summary -change x -file - <<EOF\n$(rm {MAIN}/f.txt)\nEOF")
	f.bash("data opener then bash script", "deny", "{ROOT}", "flow record summary -change x -file - <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/x.sh")
	f.bash("quoted << opens no heredoc", "deny", "{ROOT}", "echo '<<EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("data opener piped to sh", "deny", "{ROOT}", "flow record summary -change x -file - <<'EOF' | sh\nrm {MAIN}/f.txt\nEOF")
	f.bash("expanded shell opener", "deny", "{ROOT}", "$SHELL <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("commented << opens no heredoc", "deny", "{ROOT}", "x=1 # <<EOF\nrm {MAIN}/f.txt\nEOF")
	f.bash("<< inside ${} opens no heredoc", "deny", "{ROOT}", "echo \"${x:-<<EOF}\"\nrm {MAIN}/f.txt\nEOF")
	f.bash("escaped bash opener", "deny", "{ROOT}", "\\bash <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("csh opener", "deny", "{ROOT}", "csh <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("written then run by path", "deny", "{ROOT}", "envsubst <<'EOF' > /tmp/zz.sh && /tmp/zz.sh\nrm {MAIN}/f.txt\nEOF")
	f.bash("flow output run by path", "deny", "{ROOT}", "flow x <<'EOF' > /tmp/zz.sh && /tmp/zz.sh\nrm {MAIN}/f.txt\nEOF")
	f.bash("lines run as commands", "deny", "{ROOT}", "while read l; do $l; done <<'EOF'\nrm {MAIN}/f.txt\nEOF")
	f.bash("continued flow opener", "deny", "{ROOT}", "flow x <<'EOF' \\\n  | sh\nrm {MAIN}/f.txt\nEOF")
	f.bash("unbalanced ) in body", "deny", "{ROOT}", "flow x <<EOF\n$(case a in a) rm {MAIN}/f.txt;; esac)\nEOF")
}

// TestBlockUnsafeCommandsReadsTwoDataHeredocsAsData pins the kan-1041
// self-review: each data heredoc's rest is judged without the other's body,
// so a lone `.` in one and an unpaired backtick in the other no longer keep
// both scanned; an unquoted body that runs code still counts.
func TestBlockUnsafeCommandsReadsTwoDataHeredocsAsData(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	defer f.flush()
	main := root + "/main"
	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")

	f.bash("two report heredocs, a lone dot and an unpaired backtick", "allow", "{MAIN}", "cat > {SIB}/a.md <<'E'\n.\nE\ncat > {SIB}/b.md <<'E'\n`> .\nE")
	f.bash("a second body that runs the first's file keeps it scanned", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/y <<B\n$(bash {SIB}/x.sh)\nB")
	f.bash("a second heredoc fed to a shell keeps the first scanned", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\nbash <<'B'\n{SIB}/x.sh\nB")
	f.bash("a second quoted body running the first's file keeps it scanned", "deny", "{MAIN}", "cat > {SIB}/a.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/b.sh <<'B'\n{SIB}/a.sh\nB\nchmod +x {SIB}/*.sh; {SIB}/b.sh")
	f.bash("a second quoted body substituted back keeps the first scanned", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/y <<'B'\n{SIB}/x.sh\nB\n$(cat {SIB}/y)")
}

// TestBlockUnsafeCommandsReadsASingleQuotedBacktickAsLiteral pins the
// kan-1039 self-review: a backtick inside single quotes after a data
// heredoc is a plain character, so the heredoc body stays data; an
// unquoted or double-quoted backtick still keeps the body scanned.
func TestBlockUnsafeCommandsReadsASingleQuotedBacktickAsLiteral(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	defer f.flush()
	main := root + "/main"
	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")

	f.bash("single-quoted backtick after a data heredoc", "allow", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`'")
	f.bash("ANSI-quoted backtick after a data heredoc", "allow", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho $'\\'`'")
	f.bash("unquoted backtick keeps the body scanned", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho `true`")
	f.bash("double-quoted backtick keeps the body scanned", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"`true`\"")
	f.bash("an apostrophe in double quotes opens no span", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho \"it's `true`\"")
	f.bash("an unclosed quote reads as before", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\necho '`")
	f.bash("an apostrophe in a comment opens no span", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n: # it's\necho `bash {SIB}/x.sh` # '")
	f.bash("an apostrophe in another heredoc body opens no span", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\ncat > {SIB}/n <<'EOF2'\nit's\nEOF2\necho `bash {SIB}/x.sh` > {SIB}/o # '")
	f.bash("a quote span crossing a newline reads as before", "deny", "{MAIN}", "cat >> {SIB}/x <<'EOF'\na > b\nEOF\n# don't\necho `true` '")
}

// TestBlockUnsafeCommandsAdmitsRemovingALocalClone: removing a whole clone of a
// local non-bare working checkout is not a write to a main checkout; a path
// inside it, a clone of a network origin and a checkout cloned from a local
// bare repository stay denied (kan-1107 self-review, F5).
func TestBlockUnsafeCommandsAdmitsRemovingALocalClone(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f, root, main, _ := newPMFixture(t)
	defer f.flush()
	f.git("clone", "-q", "--local", main, root+"/clone")
	f.git("clone", "-q", "--local", main, root+"/netclone")
	f.git("-C", root+"/netclone", "remote", "set-url", "origin", "https://example.invalid/x.git")
	f.bash("rm -rf of a whole local clone", "allow", "{ROOT}", `rm -rf {ROOT}/clone`)
	f.bash("rm of a file inside a local clone", "deny", "{ROOT}", `rm {ROOT}/clone/f.txt`)
	f.bash("rm -rf of a clone of a network origin", "deny", "{ROOT}", `rm -rf {ROOT}/netclone`)
	f.bash("rm -rf of a checkout cloned from a local bare repository", "deny", "{ROOT}", `rm -rf {MAIN}`)
}

// TestBlockUnsafeCommandsReadsALiteralExportAsData pins decisions
// literal-export-is-data and deny-names-path-and-remedy (kan-1129 self-review
// F2): an `export NAME=<literal>` beside a data heredoc keeps no body scanned,
// while an export that may name the target, expands, or carries no value
// still does — and a Bash deny names the path it read and the remedy.
func TestBlockUnsafeCommandsReadsALiteralExportAsData(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	defer f.flush()
	main := root + "/main"
	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "remote", "set-head", "origin", "main")

	f.bash("export literal", "allow", "{MAIN}", "CR={SIB}\nexport FLOWD_PORT=4174\ncat > $CR/live-verification.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	f.bash("export names target", "deny", "{MAIN}", "export PAGER={SIB}/r.md\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	expansion := "export X=$Y\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF"
	f.bash("export expansion", "deny", "{MAIN}", expansion)
	f.bash("export without value", "deny", "{MAIN}", "export PAGER\ncat > {SIB}/r.md <<'EOF'\nwrite f<i>.txt here\nEOF")
	out := f.run(t, f.x("{MAIN}"), "Bash", map[string]string{"command": f.x(expansion)})
	for _, want := range []string{f.x("{MAIN}/.txt"), "heredoc", "Write tool"} {
		if !strings.Contains(out, want) {
			t.Errorf("(e) deny reason lacks %q: %s", want, out)
		}
	}
}

// TestBlockUnsafeCommandsDeniesAStashNotMadeOnTheCurrentBranch: the stash list is shared by every
// worktree, so a pop, apply, drop or branch is allowed only for an entry made on the current
// branch and named by its ref; a bare pop/apply (which acts on stash@{0}) is denied whatever it holds.
func TestBlockUnsafeCommandsDeniesAStashNotMadeOnTheCurrentBranch(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f, _, main, wt := newPMFixture(t)
	other := main + "/.worktrees/other"
	f.git("-C", main, "worktree", "add", other, "-b", "other")
	f.write(wt+"/f.txt", "change wip\n")
	f.git("-C", wt, "stash", "push", "-q", "-m", "mine")
	f.write(other+"/f.txt", "other wip\n")
	f.git("-C", other, "stash", "push", "-q", "-m", "theirs")
	// stash@{0} is the one made on `other`, stash@{1} the one made on `change`.
	defer f.flush()
	for _, c := range []string{"git stash pop", "git stash apply", "git stash drop", "git stash pop --index", "git stash push -- x && git stash pop", "git stash pop stash@{0}", "git stash apply stash@{0}", "git stash drop stash@{0}", "git stash branch nb stash@{0}", "git stash pop stash@{9}", "git stash pop 0"} {
		f.bash("deny "+c, "deny", "{WT}", c)
	}
	for _, c := range []string{"git stash pop stash@{1}", "git stash apply stash@{1}", "git stash drop stash@{1}", "git stash branch nb stash@{1}", "git stash list", "git stash push -- f.txt", "git stash show stash@{0}"} {
		f.bash("allow "+c, "allow", "{WT}", c)
	}
	f.check("the reason tells the agent to report BLOCKED", func(t *testing.T) {
		out := f.run(t, f.x("{WT}"), "Bash", map[string]string{"command": "git stash pop"})
		if !strings.Contains(out, "BLOCKED") || !strings.Contains(out, "git status") {
			t.Errorf("out=%s", out)
		}
	})
}
