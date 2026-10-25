package repotest

// TestProtectMainCheckout ports the retired shell harness, the
// harness for hooks/protect-main-checkout.py.
//
// It builds one bare origin with a clone whose main checkout sits on `main`
// (origin/HEAD set), a linked worktree of it on a feature branch, and one repo
// with no origin at all, then feeds the hook PreToolUse payloads and asserts
// deny/allow on each. The cases share one fixture that later cases mutate
// (archiving a change, adding worktrees, moving the main checkout off main),
// so they run in the harness's order, one subtest per harness case; the case
// numbers are the harness's own. In command templates {ROOT}, {MAIN}, {WT},
// {WT2}, {SIB}, {LWT}, {LSHA}, {CWT} and {RWT} stand for the fixture paths;
// every other `$NAME` is literal shell text the hook sees unexpanded.
//
// The loop-landing checker (cases 77-101) and the untasked-commits checker
// (cases 104-107) are stubs passed through LOOP_LANDING_CHECK and
// UNTASKED_COMMITS_CHECK; cases 108-109 run the real untasked checker.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var denyRe = regexp.MustCompile(`"permissionDecision": *"deny"`)

type pmFixture struct {
	t    *testing.T
	home string
	vars map[string]string
	env  []string // env the hook sees on every run, on top of runHook's
}

func (f *pmFixture) x(s string) string {
	for k, v := range f.vars {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}

func (f *pmFixture) git(args ...string) string {
	f.t.Helper()
	return sh(f.t, "", []string{"HOME=" + f.home}, "git",
		append([]string{"-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)...)
}

func (f *pmFixture) write(path, body string) {
	f.t.Helper()
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
	out, stderr, rc := runHook(t, "protect-main-checkout.py", buf.Bytes(), append(append([]string{}, f.env...), env...)...)
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

// check runs one case as a subtest.
func (f *pmFixture) check(name string, fn func(t *testing.T)) {
	f.t.Helper()
	f.t.Run(name, fn)
}

func (f *pmFixture) expect(name, want, cwd, tool string, input map[string]string, env ...string) {
	f.t.Helper()
	f.check(name, func(t *testing.T) {
		if got := decision(f.run(t, f.x(cwd), tool, input, env...)); got != want {
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

func (f *pmFixture) readLog(path string) string {
	b, _ := os.ReadFile(f.x(path))
	return strings.ReplaceAll(string(b), "\n", " ")
}

func TestProtectMainCheckout(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	main := root + "/main"
	wt := main + "/.worktrees/change"
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
		out, _, rc := runHook(t, "protect-main-checkout.py", []byte("not json"))
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

	// 38-41: a landing push must not carry its own change folder open
	f.write(wt+"/spectre/changes/change/proposal.md", "x\n")
	f.git("-C", wt, "add", "spectre")
	f.git("-C", wt, "commit", "-q", "-m", "plan")
	f.bash("38 HEAD:main push with the change open", "deny", "{ROOT}", `git -C {WT} push origin HEAD:main`)
	f.bash("39 change:main push with the change open", "deny", "{WT}", `git push origin change:main`)
	f.bash("41 branch push with the change open", "allow", "{ROOT}", `git -C {WT} push origin change`)
	wt2 := main + "/.worktrees/flowchg"
	f.vars["WT2"] = wt2
	f.git("-C", main, "worktree", "add", wt2, "-b", "spectre/flowchg")
	f.write(wt2+"/spectre/changes/flowchg/proposal.md", "x\n")
	f.git("-C", wt2, "add", "spectre")
	f.git("-C", wt2, "commit", "-q", "-m", "plan")
	f.bash("42 spectre/<name>:main push with the change open", "deny", "{ROOT}", `git -C {WT2} push origin spectre/flowchg:main`)
	if err := os.MkdirAll(wt+"/spectre/changes/archive", 0o755); err != nil {
		t.Fatal(err)
	}
	f.git("-C", wt, "mv", "spectre/changes/change", "spectre/changes/archive/change")
	f.git("-C", wt, "commit", "-q", "-m", "archive")
	f.bash("40 HEAD:main push with the change archived", "allow", "{ROOT}", `git -C {WT} push origin HEAD:main`)

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

	// 77-81: a fix loop lands only once check-loop-landing.sh admits it; the
	// checker is a stub here (the guard is covered by check_loop_landing_test.go).
	loopLog, loopRC := root+"/loop-args", root+"/loop-rc"
	f.write(root+"/loopcheck", "#!/usr/bin/env bash\nprintf '%s\\n' \"$@\" >\""+loopLog+"\"\necho \"LOOP-LANDING-REFUSED: REVIEW — stub\"\nexit \"$(cat \""+loopRC+"\")\"\n")
	f.env = []string{"LOOP_LANDING_CHECK=" + root + "/loopcheck"}
	lwt := root + "/main-worktrees/pipeline-fix-x-6"
	f.vars["LWT"] = lwt
	f.git("-C", main, "worktree", "add", lwt, "-b", "pipeline-fix-x-6", "origin/main")
	f.git("-C", main, "worktree", "add", root+"/main-worktrees/self-review-x", "-b", "self-review-x", "origin/main")
	for _, w := range []string{lwt, root + "/main-worktrees/self-review-x"} {
		f.write(w+"/loop.txt", w+"\n")
		f.git("-C", w, "add", "loop.txt")
		f.git("-C", w, "commit", "-q", "-m", "fix: loop")
	}
	setRC := func(rc string) { f.write(loopRC, rc+"\n") }
	setRC("1")
	f.check("77 loop push the checker refuses", func(t *testing.T) {
		out := f.run(t, root, "Bash", map[string]string{"command": f.x(`git -C {LWT} push origin pipeline-fix-x-6:main`)})
		if !denyRe.MatchString(out) || !strings.Contains(out, "LOOP-LANDING-REFUSED: REVIEW") {
			t.Errorf("%s", out)
		}
	})
	setRC("0")
	f.bash("78 loop push the checker admits", "allow", "{ROOT}", `git -C {LWT} push origin pipeline-fix-x-6:main`)
	f.check("78 checker args", func(t *testing.T) {
		if got := f.readLog(loopLog); got != lwt+" pipeline-fix-x-6 " {
			t.Errorf("checker args: %s", got)
		}
	})
	setRC("1")
	f.bash("79 self-review HEAD:main push the checker refuses", "deny", "{ROOT}", `git -C {ROOT}/main-worktrees/self-review-x push origin HEAD:main`)
	f.bash("80 loop push with no runnable checker", "deny", "{ROOT}", `git -C {LWT} push origin pipeline-fix-x-6:main`, "LOOP_LANDING_CHECK="+root+"/missing")
	_ = os.Remove(loopLog)
	f.bash("81 a non-loop landing", "allow", "{ROOT}", `git -C {WT} push origin change:main`)
	f.check("81 loop checker not run", func(t *testing.T) {
		if _, err := os.Stat(loopLog); err == nil {
			t.Error("loop checker ran")
		}
	})
	setRC("1")
	f.vars["LSHA"] = f.git("-C", lwt, "rev-parse", "HEAD")
	for i, c := range []string{"git -C {LWT} push origin @:main", "git -C {LWT} push origin {LSHA}:main",
		"git -C {LWT} push origin HEAD~0:main", "git -C {LWT} -c x=y push origin pipeline-fix-x-6:main",
		"git -C {LWT} branch tmp && git -C {LWT} push origin tmp:main"} {
		f.bash(fmt.Sprintf("%d loop landed as: %s", 82+i, c), "deny", "{ROOT}", c)
	}
	f.git("-C", main, "merge", "-q", "--ff-only", "pipeline-fix-x-6")
	f.bash("87 a colon-less push of the default branch the loop was pulled into", "deny", "{ROOT}", `git -C {MAIN} push origin main`)
	f.git("-C", main, "reset", "-q", "--hard", "origin/main")
	f.git("-C", main, "branch", "-q", "spectre/chg", f.vars["LSHA"])
	f.git("-C", main, "update-ref", "refs/heads/pipeline-fix-x-6", "origin/main")
	_ = os.Remove(loopLog)
	f.bash("88 a change branch the checker refuses", "deny", "{ROOT}", `git -C {LWT} push origin spectre/chg:main`)
	f.check("88 checker handed the change branch", func(t *testing.T) {
		b, _ := os.ReadFile(loopLog)
		if !strings.Contains("\n"+string(b), "\nspectre/chg\n") {
			t.Errorf("checker args: %s", b)
		}
	})
	f.write(root+"/slowcheck", "#!/usr/bin/env bash\nsleep 1.5\n")
	f.bash("89 two landing pushes share one checker budget", "deny", "{ROOT}", `git -C {LWT} push origin spectre/chg:main; git -C {LWT} push origin spectre/chg:main`,
		"LOOP_LANDING_CHECK="+root+"/slowcheck", "LOOP_LANDING_TIMEOUT=2")
	// 90-97: a push named through an unexpanded shell expansion cannot be checked;
	// the stub admits, so only the unreadable-landing check can deny.
	setRC("0")
	for i, c := range []string{"git -C $W push origin $B:main", "git -C {WT} push origin \"${B}:main\"",
		"git -C {WT} push origin $(git branch --show-current):main",
		"git -C {WT} push origin `git branch --show-current`:main"} {
		f.bash(fmt.Sprintf("%d unreadable landing: %s", 90+i, c), "deny", "{ROOT}", c)
	}
	f.bash("94 cd to an unexpanded path, then push HEAD", "deny", "{ROOT}", `cd $W && git push origin HEAD`)
	f.bash("95 push of an unexpanded branch", "deny", "{ROOT}", `git -C $W push -u origin $B`)
	f.bash("96 unexpanded repository, literal non-default destination", "allow", "{ROOT}", `git -C $W push origin spectre/x`)
	f.bash("97 a branch assigned in the command itself", "allow", "{ROOT}", `B=change; git -C {WT} push origin $B`)
	f.bash("98 a prefix assignment is not seen by its own command's words", "deny", "{ROOT}", `B=x git -C {WT} push origin $B:main`)
	f.bash("99 --all with an unexpanded remote pushes the default branch too", "deny", "{ROOT}", `git -C {WT} push --all $R`)
	f.bash("100 an expanded git option hides the repository", "deny", "{ROOT}", `git ${OPTS} push origin HEAD`)
	f.bash("101 an expanded git command word", "deny", "{ROOT}", `$GIT push origin HEAD`)
	f.env = nil
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
	f.bash("118 a data heredoc's target named again as an argument keeps its body as data", "allow", "{MAIN}", planBody+"check-plan-shape.sh $C/t.md")
	f.bash("119 a path-named command naming another file keeps the body as data", "allow", "{MAIN}", planBody+"~/.claude/skills/flow/scripts/check-plan-shape.sh /x/o.md")
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

	// 104-107: a push of a change branch runs check-untasked-commits over its
	// unpushed commits (kan-941); a stub stands in for the guard.
	untaskedRC, untaskedLog := root+"/untasked-rc", root+"/untasked-log"
	f.write(root+"/untaskedcheck", "#!/usr/bin/env bash\nprintf '%s\\n' \"$@\" >\""+untaskedLog+"\"\necho \"UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer\"\nexit \"$(cat \""+untaskedRC+"\")\"\n")
	f.env = []string{"UNTASKED_COMMITS_CHECK=" + root + "/untaskedcheck"}
	cwt := root + "/main-worktrees/chg2"
	f.vars["CWT"] = cwt
	f.git("-C", main, "worktree", "add", cwt, "-b", "spectre/chg2", "origin/main")
	f.git("-C", cwt, "push", "-q", "-u", "origin", "spectre/chg2")
	f.write(cwt+"/m.txt", "m\n")
	f.git("-C", cwt, "add", "m.txt")
	f.git("-C", cwt, "commit", "-q", "-m", "fix(app): review Minors")
	f.write(untaskedRC, "1\n")
	f.bash("104 a change-branch push carrying an untasked commit", "deny", "{ROOT}", `git -C {CWT} push origin spectre/chg2`)
	f.check("104 checker args", func(t *testing.T) {
		if got := f.readLog(untaskedLog); got != cwt+" refs/remotes/origin/spectre/chg2 --not-remotes " {
			t.Errorf("checker args: %s", got)
		}
	})
	f.bash("160 a bare change-branch push with its output redirected and piped", "deny", "{ROOT}", `git -C {CWT} push -q 2>&1 | tail -1`)
	f.write(untaskedRC, "0\n")
	f.bash("105 a change-branch push the checker clears", "allow", "{ROOT}", `cd {CWT} && git push`)
	f.write(untaskedRC, "1\n")
	_ = os.Remove(untaskedLog)
	f.git("-C", cwt, "checkout", "-q", "-b", "spectre/chg3")
	f.bash("106 a change branch origin does not hold yet", "allow", "{ROOT}", `git -C {CWT} push -u origin spectre/chg3`)
	f.check("106 checker not run", func(t *testing.T) {
		if _, err := os.Stat(untaskedLog); err == nil {
			t.Error("checker ran")
		}
	})
	f.bash("107 a checker that cannot run fails open", "allow", "{ROOT}", `git -C {CWT} push origin HEAD:spectre/chg2`, "UNTASKED_COMMITS_CHECK="+root+"/missing")
	f.env = nil

	// 143-148: a commit and its push in one command (kan-1003's self-review: the design handoff
	// 0497d943 was committed and pushed in one call, before the push check could see it).
	f.write(cwt+"/h.txt", "h\n")
	f.bash("143 commit and push in one command, no trailer", "deny", "{ROOT}", `cd {CWT} && git add h.txt && git commit -qm 'docs(design): add the handoff' && git push -q`)
	f.bash("144 the same with a Flow-Key trailer", "allow", "{ROOT}", `cd {CWT} && git commit -q -m 'docs(design): add the handoff' -m 'Flow-Key: flow.brainstorm' && git push -q`)
	f.bash("145 a planning commit pathspec-limited to spectre/changes/", "allow", "{ROOT}", `git -C {CWT} commit -q -m 'docs(spectre): record' -- spectre/changes/chg && git -C {CWT} push -q origin spectre/chg3`)
	f.bash("150 a planning commit's pathspec given without '--'", "allow", "{ROOT}", `git -C {CWT} commit -q -m 'docs(spectre): record' spectre/changes/chg && git -C {CWT} push -q origin spectre/chg3`)
	f.bash("146 a chore subject", "allow", "{ROOT}", `cd {CWT} && git commit -q -m 'chore(spectre): plan chg — plan gate' && git push -q`)
	f.bash("147 a message the command does not state", "allow", "{ROOT}", `cd {CWT} && git commit -q -F msg.txt && git push -q`)
	f.bash("161 commit and push in one command, both redirected, no trailer", "deny", "{ROOT}", `cd {CWT} && git commit -qm 'docs(design): add the handoff' 2>&1 && git push -q >/dev/null 2>&1`)
	f.bash("162 a planning commit pathspec-limited to spectre/changes/, redirected", "allow", "{ROOT}", `git -C {CWT} commit -q -m 'docs(spectre): record' -- spectre/changes/chg 2>&1 && git -C {CWT} push -q origin spectre/chg3`)
	f.bash("148 a trailer-less commit on a change branch, pushed to a branch that is no change branch, is denied at the commit", "deny", "{ROOT}", `cd {CWT} && git commit -qm 'docs(x): y' && git push -q origin HEAD:feature/x`)

	// 108-109: the real checker after a rebase onto a moved base — the base's own
	// trailer-less commits are on origin already, never the change's to judge.
	rwt := root + "/main-worktrees/chg4"
	f.vars["RWT"] = rwt
	f.git("-C", main, "worktree", "add", rwt, "-b", "spectre/chg4", "origin/main")
	f.git("-C", rwt, "push", "-q", "-u", "origin", "spectre/chg4")
	f.write(rwt+"/b.txt", "b\n")
	f.git("-C", rwt, "add", "b.txt")
	f.git("-C", rwt, "commit", "-q", "-m", "fix(build): someone else's work")
	f.git("-C", rwt, "push", "-q", "origin", "HEAD:refs/heads/main")
	f.write(rwt+"/t.txt", "t\n")
	f.git("-C", rwt, "add", "t.txt")
	f.git("-C", rwt, "commit", "-q", "-m", "feat(app): task one", "-m", "Task-Id: 1")
	f.bash("108 a push after a rebase onto a moved base", "allow", "{ROOT}", `git -C {RWT} push --force-with-lease origin spectre/chg4`)
	f.write(rwt+"/u.txt", "u\n")
	f.git("-C", rwt, "add", "u.txt")
	f.git("-C", rwt, "commit", "-q", "-m", "fix(app): review Minors")
	f.bash("109 the real checker still lists the change's own untasked commit", "deny", "{ROOT}", `git -C {RWT} push --force-with-lease origin spectre/chg4`)

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
	f.bash("166 a report heredoc then render-slot-prompt.sh -fix-report it, cwd = main", "allow", "{MAIN}", "D={SIB}; SK=$(readlink -f ~/.claude/skills/flow); cat > $D/r.md <<'EOF'\nfrom `x-<n>.png`\nEOF\nfor s in a b; do render-slot-prompt.sh $SK 4 -fix-report $D/r.md; done")
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
	f.git("-C", wt, "checkout", "HEAD", "--", "f.txt")
	f.bash("155 checkout -- of a clean path", "allow", "{ROOT}", `git -C {WT} checkout -- f.txt`)

	// 3 last: moving the main checkout off main lifts the protection
	f.git("-C", main, "checkout", "-q", "-b", "feature")
	f.writeTool("3 Write in main checkout on a feature branch", "allow", "{MAIN}/new.txt")
}

// TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit pins the change-branch commit
// check: a commit the command states with no trailer and no chore subject, in a worktree on
// spectre/<name>, is denied; a trailered, chore, -F or --amend commit, and any commit on another
// branch, passes.
func TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
	main := root + "/main"
	f.vars["SWT"], f.vars["FWT"] = root+"/main-worktrees/c1", root+"/main-worktrees/fc1"
	f.git("init", "-q", "--bare", root+"/origin.git")
	f.git("clone", "-q", root+"/origin.git", main)
	f.write(main+"/f.txt", "a\n")
	f.git("-C", main, "add", "f.txt")
	f.git("-C", main, "commit", "-q", "-m", "init")
	f.git("-C", main, "branch", "-M", "main")
	f.git("-C", main, "push", "-q", "-u", "origin", "main")
	f.git("-C", main, "worktree", "add", f.vars["SWT"], "-b", "spectre/c1")
	f.git("-C", main, "worktree", "add", f.vars["FWT"], "-b", "feature/c1")

	f.check("1 untrailered commit on spectre/c1", func(t *testing.T) {
		out := f.run(t, root, "Bash", map[string]string{"command": f.x(`git -C {SWT} commit -m "fix(x): y"`)})
		if decision(out) != "deny" || !strings.Contains(out, "Flow-Key:") {
			t.Errorf("out=%s", out)
		}
	})
	f.bash("2 Task-Id trailer", "allow", "{ROOT}", `git -C {SWT} commit -m "fix(x): y" -m "Task-Id: 3"`)
	f.bash("3 chore subject", "allow", "{ROOT}", `git -C {SWT} commit -m "chore(spectre): plan c1"`)
	f.bash("4 message from a file", "allow", "{ROOT}", `git -C {SWT} commit -F msg.txt`)
	f.bash("5 amend", "allow", "{ROOT}", `git -C {SWT} commit --amend --no-edit`)
	f.bash("6 another branch", "allow", "{ROOT}", `git -C {FWT} commit -m "fix(x): y"`)
}

// TestProtectMainCheckoutDeniesABareCatRedirect pins kan-1033's finding 4.2:
// a bare `cat > <file>` reads stdin, which a Bash call never has, so it
// hangs until the call times out; a cat with input of its own passes.
func TestProtectMainCheckoutDeniesABareCatRedirect(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
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
