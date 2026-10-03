package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of scripts/test-break-and-prove.sh at 9cd35da8, one subtest per
// ok: label, nested under the harness's own case. The harness built a
// throwaway repository per case carrying a committed config.yaml and
// README.md and ran the script from inside it; here each case copies one
// repository built once, and the guard runs in-process from the same
// directory. Every case that ran the script also carries a "matches the bash
// at base" subtest: the bash script and the port run on separate copies of
// the same fixture with the same arguments, and stdout, stderr and exit must
// be the same once each copy's own path reads <repo>. The bash is read at
// d71a2327 (bashAtBase), whose break-and-prove.sh differs from 9cd35da8's in
// one comment line only. The cases after the harness's eighteen are exit
// paths the harness never reached, and the real shim.

const bpConfig = "mode: strict\nretries: 3\n"

// bpFlipPatch is write_flip_patch: a hand-written diff flipping mode: strict
// to mode: lax.
const bpFlipPatch = "--- a/config.yaml\n+++ b/config.yaml\n@@ -1,2 +1,2 @@\n-mode: strict\n+mode: lax\n retries: 3\n"

// bpSed and bpTest are the harness's usual mutation and test command.
var (
	bpSed  = []string{"--sed", "s/mode: strict/mode: lax/", "config.yaml"}
	bpTest = []string{"--", "grep", "-q", "mode: strict", "config.yaml"}
)

func bpArgs(parts ...[]string) []string {
	var args []string
	for _, p := range parts {
		args = append(args, p...)
	}
	return args
}

// bpBuild is new_fixture_repo, built once: config.yaml (2 lines) and
// README.md (1 line) committed.
func bpBuild(t *testing.T) string {
	t.Helper()
	repo := t.TempDir() + "/base"
	gitRun(t, "", "init", "-q", "-b", "main", "--template=", repo)
	writeFile(t, repo+"/config.yaml", bpConfig)
	writeFile(t, repo+"/README.md", "readme v1\n")
	gitRun(t, repo, "add", "config.yaml", "README.md")
	gitRun(t, repo, "commit", "-qm", "fixture")
	return repo
}

// bpRepo is a copy of the base repository in its own temporary directory.
func bpRepo(t *testing.T, base string) string {
	t.Helper()
	repo := t.TempDir() + "/repo"
	mvCopyTree(t, base, repo)
	return repo
}

// bpProve is run_proof: the guard in-process from inside repo.
func bpProve(repo string, args ...string) mvRes {
	var out, errb bytes.Buffer
	rc := breakAndProve(args, crEnv(repo, nil), &out, &errb)
	return mvRes{rc, out.String(), errb.String()}
}

func bpConfigOf(repo string) string {
	b, _ := os.ReadFile(repo + "/config.yaml")
	return string(b)
}

// bpBashDiag is the location bash puts before its own diagnostics: the
// script's path and line. The port prints "break-and-prove: " there instead.
var bpBashDiag = regexp.MustCompile(`(?m)^\S*break-and-prove\.sh: line [0-9]+: `)

// bpParity runs the bash at base and the port on separate copies of the
// fixture prep sets up, with the arguments prep returns, and asserts stdout,
// stderr and exit are the same.
func bpParity(t *testing.T, base, bashDir string, prep func(t *testing.T, repo string) []string) {
	t.Helper()
	run := func(bash bool) mvRes {
		repo := bpRepo(t, base)
		args := prep(t, repo)
		var r mvRes
		if bash {
			cmd := exec.Command("bash", append([]string{bashDir + "/break-and-prove.sh"}, args...)...)
			cmd.Dir = repo
			var out, errb bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errb
			_ = cmd.Run()
			r = mvRes{cmd.ProcessState.ExitCode(),
				bpBashDiag.ReplaceAllString(out.String(), "break-and-prove: "),
				bpBashDiag.ReplaceAllString(errb.String(), "break-and-prove: ")}
		} else {
			r = bpProve(repo, args...)
		}
		_ = os.Chmod(repo+"/config.yaml", 0o644)
		phys, err := filepath.EvalSymlinks(repo)
		if err != nil {
			t.Fatal(err)
		}
		norm := strings.NewReplacer(phys, "<repo>", repo, "<repo>")
		return mvRes{r.rc, norm.Replace(r.out), norm.Replace(r.err)}
	}
	want, got := run(true), run(false)
	mvCheck(t, "matches the bash at base", got == want, "got:\n%+v\nwant:\n%+v", got, want)
}

// bpCase is one harness case: prep sets up its copy of the fixture and
// returns the arguments, check asserts on the port's run against it.
type bpCase struct {
	name  string
	prep  func(t *testing.T, repo string) []string
	check func(t *testing.T, repo string, r mvRes)
}

// bpPatch writes body beside repo (outside its tree, where a capture inside
// it would be untracked drift) and returns its path.
func bpPatch(t *testing.T, repo, body string) string {
	t.Helper()
	p := repo + ".parent.patch"
	writeFile(t, p, body)
	return p
}

func TestBreakAndProve(t *testing.T) {
	t.Parallel()
	base := bpBuild(t)
	bashDir := bashAtBase(t, "break-and-prove.sh", "lib/post-mutation-check.sh")
	restored := func(t *testing.T, label, repo string) {
		t.Helper()
		mvCheck(t, label, bpConfigOf(repo) == bpConfig, "config.yaml is %q", bpConfigOf(repo))
	}
	rcIs := func(t *testing.T, label string, r mvRes, want int) {
		t.Helper()
		mvCheck(t, label, r.rc == want, "expected exit %d, got %d\n%s", want, r.rc, r.all())
	}
	contains := func(t *testing.T, label string, r mvRes, needle string) {
		t.Helper()
		mvCheck(t, label, strings.Contains(r.all(), needle), "output has no line matching: %s\n%s", needle, r.all())
	}
	usage := func(t *testing.T, label string, r mvRes) {
		t.Helper()
		rcIs(t, label, r, 4)
	}
	cases := []bpCase{
		// 1. proof held — sed mutation.
		{"1 proof held — sed mutation", func(t *testing.T, repo string) []string {
			writeFile(t, repo+"/.config-before", bpConfig)
			return bpArgs(bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "proof held (sed mutation) — exit code", r, 0)
			contains(t, "proof held (sed mutation) — run 1 reported", r, "run 1 exited")
			restored(t, "proof held (sed mutation) — config restored byte-exact", repo)
		}},
		// 2. proof held — patch mutation.
		{"2 proof held — patch mutation", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"--patch", bpPatch(t, repo, bpFlipPatch), "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "proof held (patch mutation) — exit code", r, 0)
			restored(t, "proof held (patch mutation) — config restored byte-exact", repo)
		}},
		// 3. uncommitted edit on the target file survives — the observed
		// KAN-329 failure: `git checkout --` destroyed such edits; the
		// snapshot must not.
		{"3 uncommitted edit on the target file survives", func(t *testing.T, repo string) []string {
			writeFile(t, repo+"/config.yaml", bpConfig+"# tuned by hand\n")
			return bpArgs(bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "uncommitted edit on the target file survives — exit code", r, 0)
			mvCheck(t, "uncommitted edit on the target file survives — edit and original content both present",
				bpConfigOf(repo) == bpConfig+"# tuned by hand\n", "config.yaml is %q", bpConfigOf(repo))
		}},
		// 4. uncommitted edit on a second file survives.
		{"4 uncommitted edit on a second file survives", func(t *testing.T, repo string) []string {
			writeFile(t, repo+"/README.md", "readme v1\nuncommitted note\n")
			return bpArgs(bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "uncommitted edit on a second file survives — exit code", r, 0)
			b, _ := os.ReadFile(repo + "/README.md")
			mvCheck(t, "uncommitted edit on a second file survives — edit still present",
				string(b) == "readme v1\nuncommitted note\n", "README.md is %q", b)
		}},
		// 6. mutated run passed — the proof did not hold (surviving mutant).
		{"6 mutated run passed", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "grep", "-q", "mode:", "config.yaml"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "mutated run passed — proof did not hold", r, 1)
			contains(t, "mutated run passed — leg named", r, "run 1 exited 0")
			restored(t, "mutated run passed — config still restored", repo)
		}},
		// 7. restored run failed — the proof did not hold.
		{"7 restored run failed", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "grep", "-q", "impossible-token", "config.yaml"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "restored run failed — proof did not hold", r, 1)
			contains(t, "restored run failed — leg named", r, "run 2")
		}},
		// 8. patch not applying — refused before mutating anything.
		{"8 patch not applying", func(t *testing.T, repo string) []string {
			p := bpPatch(t, repo, strings.Replace(bpFlipPatch, " retries: 3", " retries: 9", 1))
			return bpArgs([]string{"--patch", p, "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "patch not applying refused", r, 2)
			restored(t, "patch not applying refused — config untouched", repo)
		}},
		// 9. sed matching nothing — refused.
		{"9 sed matching nothing", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"--sed", "s/zzz-no-match/yyy/", "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "sed matching nothing refused", r, 2)
			restored(t, "sed matching nothing refused — config untouched", repo)
		}},
		// 10. patch touching other files — refused.
		{"10 patch touching other files", func(t *testing.T, repo string) []string {
			p := bpPatch(t, repo, bpFlipPatch+"--- a/README.md\n+++ b/README.md\n@@ -1 +1 @@\n-readme v1\n+readme v2\n")
			return bpArgs([]string{"--patch", p, "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "patch touching other files refused", r, 2)
		}},
		// 11. missing target file — refused.
		{"11 missing target file", func(t *testing.T, repo string) []string {
			return []string{"--sed", "s/a/b/", "no-such-file.yaml", "--", "grep", "-q", "x", "no-such-file.yaml"}
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "missing target file refused", r, 2)
		}},
		// 12. the clean command runs before each leg — the Gradle cleanTest
		// trap, handled once by the script instead of remembered every time.
		{"12 clean command runs before each leg", func(t *testing.T, repo string) []string {
			log := repo + ".parent.log"
			return bpArgs([]string{"--clean", "echo clean >> '" + log + "'"}, bpSed,
				[]string{"--", "sh", "-c", "echo test >> '" + log + "'; grep -q 'mode: strict' config.yaml"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "clean command runs before each leg — exit code", r, 0)
			b, _ := os.ReadFile(repo + ".parent.log")
			mvCheck(t, "clean command runs before each leg — clean,test,clean,test",
				string(b) == "clean\ntest\nclean\ntest\n", "log was: %q", b)
		}},
		// 13. restore failure — exit 3, mutated content may remain.
		{"13 restore failure", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "sh", "-c", "chmod 444 config.yaml; exit 7"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "restore failure exits 3", r, 3)
			contains(t, "restore failure exits 3 — named in the report", r, "could not restore")
			_ = os.Chmod(repo+"/config.yaml", 0o644)
		}},
		// 14. tree drift — an untracked file the test command leaves behind
		// is reported and forces exit 2.
		{"14 tree drift", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "sh", "-c", "touch stray.txt; exit 5"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "tree drift exits 2", r, 2)
			contains(t, "tree drift exits 2 — stray file named", r, "stray.txt")
		}},
		// 15. usage errors — exit 4.
		{"15 usage errors — no mutation flag", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "usage errors exit 4 — no mutation flag", r) }},
		{"15 usage errors — no test command", func(t *testing.T, repo string) []string {
			return []string{"--sed", "s/a/b/", "config.yaml"}
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "usage errors exit 4 — no test command", r) }},
		{"15 usage errors — unreadable patch file", func(t *testing.T, repo string) []string {
			return []string{"--patch", repo + "/no-such.patch", "config.yaml", "--", "grep", "-q", "x", "config.yaml"}
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "usage errors exit 4 — unreadable patch file", r) }},
		// 16. not inside a git worktree — exit 4.
		{"16 not a git worktree", func(t *testing.T, repo string) []string {
			if err := os.RemoveAll(repo + "/.git"); err != nil {
				t.Fatal(err)
			}
			return []string{"--sed", "s/a/b/", "whatever.yaml", "--", "true"}
		}, func(t *testing.T, repo string, r mvRes) { rcIs(t, "not a git worktree exits 4", r, 4) }},
		// 17. test command that cannot exec (127) — exit 4, never a false
		// verdict.
		{"17 test command 127", func(t *testing.T, repo string) []string {
			writeFile(t, repo+"/.config-before", bpConfig)
			return bpArgs(bpSed, []string{"--", "./no-such-command-anywhere"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "test command 127 exits 4", r, 4)
			contains(t, "test command 127 — the 126/127 message, not a proof verdict", r, "not usable as evidence")
			restored(t, "test command 127 — config restored", repo)
		}},
		// 18. test command that exists but is not executable (126) — exit 4
		// too.
		{"18 test command 126", func(t *testing.T, repo string) []string {
			writeFile(t, repo+"/nonexec-cmd", "#!/usr/bin/env bash\necho should-not-run\n")
			return bpArgs(bpSed, []string{"--", "./nonexec-cmd"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "test command 126 exits 4", r, 4)
			contains(t, "test command 126 — the 126/127 message, not a proof verdict", r, "not usable as evidence")
			restored(t, "test command 126 — config restored", repo)
		}},

		// Exit paths the harness never reached, each beside the bash.
		{"port: a target the mutation cannot write ends the run unmutated", func(t *testing.T, repo string) []string {
			if err := os.Chmod(repo+"/config.yaml", 0o444); err != nil {
				t.Fatal(err)
			}
			return bpArgs(bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "unwritable target: exits 1", r, 1)
			restored(t, "unwritable target: config untouched", repo)
			_ = os.Chmod(repo+"/config.yaml", 0o644)
		}},
		{"port: an executable without #! runs as a bash script", func(t *testing.T, repo string) []string {
			writeExec(t, repo+"/no-shebang", "grep -q 'mode: strict' config.yaml\n")
			return bpArgs(bpSed, []string{"--", "./no-shebang"})
		}, func(t *testing.T, repo string, r mvRes) { rcIs(t, "no #!: proof held", r, 0) }},
		{"port: a test command not on PATH", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "no-such-command-kan873"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "not on PATH: exits 4", r, 4)
			restored(t, "not on PATH: config restored", repo)
		}},
		{"port: a failing sed is refused", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"--sed", "s/unterminated", "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "failing sed: exits 2", r, 2)
			contains(t, "failing sed: named", r, "sed failed with exit")
		}},
		{"port: a patch target outside the repository is refused", func(t *testing.T, repo string) []string {
			writeFile(t, filepath.Dir(repo)+"/outside/config.yaml", bpConfig)
			return bpArgs([]string{"--patch", bpPatch(t, repo, bpFlipPatch), "../outside/config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "outside target: exits 2", r, 2)
			contains(t, "outside target: named", r, "target file is outside the repository: ../outside/config.yaml")
		}},
		{"port: a patch touching another file is refused", func(t *testing.T, repo string) []string {
			p := bpPatch(t, repo, "--- a/README.md\n+++ b/README.md\n@@ -1 +1 @@\n-readme v1\n+readme v2\n")
			return bpArgs([]string{"--patch", p, "config.yaml"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "another file: exits 2", r, 2)
			contains(t, "another file: named", r, "patch touches README.md, not config.yaml")
		}},
		{"port: a failing clean command cannot answer", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"--clean", "echo 'it''s'; exit 3"}, bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "failing clean: exits 4", r, 4)
			contains(t, "failing clean: named", r, "clean command exited 3")
			restored(t, "failing clean: config restored by the exit path", repo)
		}},
		{"port: run 2 leaving the file changed exits 3", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "sh", "-c",
				"grep -q 'mode: strict' config.yaml || exit 1; echo extra >> config.yaml"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "changed by run 2: exits 3", r, 3)
			contains(t, "changed by run 2: drift named", r, "post-restore drift detected")
		}},
		{"port: usage — no arguments", func(t *testing.T, repo string) []string {
			return nil
		}, func(t *testing.T, repo string, r mvRes) {
			mvCheck(t, "no arguments: the usage line", r.rc == 4 && r.out == "" && r.err == bpUsage+"\n", "%+v", r)
		}},
		{"port: usage — --clean twice", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"--clean", "true", "--clean", "true"}, bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "--clean twice: exits 4", r) }},
		{"port: usage — --sed and --patch", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--patch", "p"}, bpTest)
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "--sed and --patch: exits 4", r) }},
		{"port: usage — two files", func(t *testing.T, repo string) []string {
			return bpArgs([]string{"README.md"}, bpSed, bpTest)
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "two files: exits 4", r) }},
		{"port: usage — --clean without its command", func(t *testing.T, repo string) []string {
			return []string{"--clean"}
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "--clean alone: exits 4", r) }},
		{"port: usage — nothing after --", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--"})
		}, func(t *testing.T, repo string, r mvRes) { usage(t, "nothing after --: exits 4", r) }},
		// A bash builtin as the test command ran as the builtin at base.
		{"port: a bash builtin test command runs as the builtin", func(t *testing.T, repo string) []string {
			return bpArgs(bpSed, []string{"--", "eval", "grep -q 'mode: strict' config.yaml"})
		}, func(t *testing.T, repo string, r mvRes) {
			rcIs(t, "eval test command: proof held, exit 0", r, 0)
			restored(t, "eval test command: config restored byte-exact", repo)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			repo := bpRepo(t, base)
			c.check(t, repo, bpProve(repo, c.prep(t, repo)...))
			bpParity(t, base, bashDir, c.prep)
		})
	}

	// 5. no `git checkout` call in the port — the anti-pattern is banned
	// structurally, not just in the happy path. Comment lines are stripped
	// first: the comments document the ban and may name the command.
	t.Run("5 no git checkout in the script", func(t *testing.T) {
		t.Parallel()
		src, err := os.ReadFile("breakandprove.go")
		if err != nil {
			t.Fatal(err)
		}
		var hits []string
		for _, l := range strings.Split(string(src), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(l), "//") && strings.Contains(l, "checkout") {
				hits = append(hits, l)
			}
		}
		mvCheck(t, "no git checkout in the script", len(hits) == 0, "found a git checkout call: %q", hits)
	})

	// The real shim, directly and through a symlink in a temporary
	// directory that links lib/ beside it, as an installed skill's scripts/
	// reaches it.
	t.Run("shim", func(t *testing.T) {
		t.Parallel()
		shim := tcfScriptsDir(t) + "/break-and-prove.sh"
		linkDir := t.TempDir()
		link := linkDir + "/break-and-prove.sh"
		symlink(t, shim, link)
		symlink(t, tcfScriptsDir(t)+"/lib", linkDir+"/lib")
		for label, path := range map[string]string{"shim: no arguments exits 4 with the usage line": shim,
			"shim through a symlink: no arguments exits 4 with the usage line": link} {
			cmd := exec.Command("/bin/bash", path)
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
			var out, errb bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errb
			_ = cmd.Run()
			mvCheck(t, label, cmd.ProcessState.ExitCode() == 4 && out.Len() == 0 && errb.String() == bpUsage+"\n",
				"rc=%d stdout=%q stderr=%q", cmd.ProcessState.ExitCode(), out.String(), errb.String())
		}
	})

	// A SIGTERM while a leg runs restores the file, as the bash EXIT trap
	// did, and the process then dies of the signal.
	t.Run("signal mid-run", func(t *testing.T) {
		t.Parallel()
		repo := bpRepo(t, base)
		dir := t.TempDir()
		ready, release := dir+"/ready", dir+"/release"
		for _, p := range []string{ready, release} {
			if err := syscall.Mkfifo(p, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command("/bin/bash", append([]string{tcfScriptsDir(t) + "/break-and-prove.sh"}, bpArgs(bpSed,
			[]string{"--", "sh", "-c", `echo ready > "$BP_READY"; read -r _ < "$BP_RELEASE"; exit 1`})...)...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "BP_READY="+ready, "BP_RELEASE="+release)
		// Files, not buffers: Wait then waits for the process alone, never
		// for a pipe the orphaned test command still holds.
		out := mvCreate(t, dir+"/stdout")
		cmd.Stdout, cmd.Stderr = out, mvCreate(t, dir+"/stderr")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.ReadFile(ready); err != nil { // blocks until the test command writes
			t.Fatal(err)
		}
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatal(err)
		}
		_ = cmd.Wait()
		if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil { // unblock the orphaned test command
			f.Close()
		}
		ws, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
		mvCheck(t, "signal mid-run: dies of SIGTERM", ws.Signaled() && ws.Signal() == syscall.SIGTERM, "status %v", cmd.ProcessState)
		restored(t, "signal mid-run: config restored", repo)
		o, _ := os.ReadFile(dir + "/stdout")
		mvCheck(t, "signal mid-run: the restore reported", strings.Contains(string(o),
			"break-and-prove: config.yaml restored byte-exact (restored by the exit trap)"), "%s", o)
	})
}

// bpUsage is the usage line break-and-prove.sh printed at 9cd35da8.
const bpUsage = "break-and-prove: usage: break-and-prove.sh [--clean <command>] <file> " +
	"(--sed <expr> | --patch <patchfile>) -- <test-command> [<args>...]"

// A test command found through a relative PATH entry ran at base; Go's
// exec.ErrDot refusal must not turn it into a cannot-answer. Not parallel:
// it sets the process PATH the port's exec resolves against.
func TestBreakAndProveRelativePath(t *testing.T) {
	repo := bpRepo(t, bpBuild(t))
	writeExec(t, repo+"/chk", "#!/bin/sh\ngrep -q 'mode: strict' config.yaml\n")
	t.Setenv("PATH", ".:"+os.Getenv("PATH"))
	r := bpProve(repo, bpArgs(bpSed, []string{"--", "chk"})...)
	if r.rc != 0 || bpConfigOf(repo) != bpConfig {
		t.Errorf("rc=%d config=%q\n%s", r.rc, bpConfigOf(repo), r.all())
	}
}

// A --clean command that backgrounds a child returns as soon as `sh` exits,
// as the bash's `sh -c` on its inherited stdout did: the clean command gets
// the guard's own stdout file, never a pipe Run() would wait on until every
// descendant holding it exited. Each backgrounded sleep records its pid so the
// test kills it whatever the outcome; a run that waits on them outlasts the
// deadline long before the sleeps end.
func TestBreakAndProveCleanBackgroundChild(t *testing.T) {
	t.Parallel()
	repo := bpRepo(t, bpBuild(t))
	pids := repo + ".pids"
	t.Cleanup(func() {
		b, _ := os.ReadFile(pids)
		for _, p := range strings.Fields(string(b)) {
			_ = exec.Command("kill", p).Run()
		}
	})
	dir := t.TempDir()
	stdout, stderr := mvCreate(t, dir+"/stdout"), mvCreate(t, dir+"/stderr")
	done := make(chan int, 1)
	go func() {
		done <- breakAndProve(bpArgs([]string{"--clean", "sleep 60 & echo $! >> '" + pids + "'"}, bpSed, bpTest),
			crEnv(repo, nil), stdout, stderr)
	}()
	select {
	case rc := <-done:
		mvCheck(t, "clean backgrounding a child: proof held, exit 0", rc == 0, "rc=%d", rc)
	case <-time.After(20 * time.Second):
		t.Fatal("clean backgrounding a child: the run waited on the clean command's backgrounded child")
	}
}
