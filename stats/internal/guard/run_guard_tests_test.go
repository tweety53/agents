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

// Every case of the runner's bash harness and cases 1–6 (with 5b) of the
// bash parallel library's harness at 2a3a0297, one subtest per ok: label,
// plus the duration ordering and a SIGINT case. The bash harnesses ran the
// runner with RUN_GUARD_TESTS_ROOT on a fixture and asserted on its exit and
// combined output; here the runner runs in-process with the same overrides
// on Env. Cases 7–9 of the library's harness tested bash traps, mktemp and
// `set -e` inside a library that no longer exists, and are not ported.

const (
	rgtAlpha = "printf 'ALPHA-OUT-MARKER\\n'\nexit 0\n"
	rgtBeta  = "printf 'BETA-OUT-MARKER\\n'\nprintf 'BETA-ERR-MARKER\\n' >&2\nexit 3\n"
	rgtGamma = "printf 'GAMMA-OUT-MARKER\\n'\nexit 0\n"
)

// rgtFixture is a test root carrying files (name: body).
func rgtFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		writeFile(t, filepath.Join(root, rel), body)
	}
	return root
}

// rgtFull is new_fixture: alpha and gamma pass, beta fails, and a non-test
// helper plus a nested harness the flat glob must never run.
func rgtFull(t *testing.T) string {
	return rgtFixture(t, map[string]string{
		"test-alpha.sh":         rgtAlpha,
		"test-beta.sh":          rgtBeta,
		"test-gamma.sh":         rgtGamma,
		"helper.sh":             "printf 'HELPER-SHOULD-NEVER-RUN-MARKER\\n'\nexit 1\n",
		"subdir/test-nested.sh": "printf 'NESTED-SHOULD-NEVER-RUN-MARKER\\n'\nexit 1\n",
	})
}

// rgtClean is new_clean_fixture: two passing harnesses.
func rgtClean(t *testing.T) string {
	return rgtFixture(t, map[string]string{"test-alpha.sh": rgtAlpha, "test-gamma.sh": rgtGamma})
}

// rgtRunIn runs the runner over root with vars on Env (a name mapped to nil
// is unset); GUARD_TESTS_REPO_ROOT defaults to a directory outside any
// repository, so no case gates on this repository's tree. It returns the
// exit code, stdout and stderr.
func rgtRunIn(t *testing.T, root string, vars map[string]*string) (int, string, string) {
	t.Helper()
	all := map[string]*string{"RUN_GUARD_TESTS_ROOT": &root, "GUARD_TESTS_REPO_ROOT": ptr(t.TempDir()), "TMPDIR": ptr(t.TempDir())}
	for k, v := range vars {
		all[k] = v
	}
	lookup := func(k string) (string, bool) {
		if v, ok := all[k]; ok {
			if v == nil {
				return "", false
			}
			return *v, true
		}
		if k == "PATH" {
			return os.LookupEnv(k)
		}
		return "", false
	}
	var out, errb bytes.Buffer
	rc := runGuardTests(nil, Env{Dir: t.TempDir(), LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}, &out, &errb)
	return rc, out.String(), errb.String()
}

func rgtRun1(t *testing.T, root string, vars map[string]*string) (int, string) {
	t.Helper()
	rc, out, errs := rgtRunIn(t, root, vars)
	return rc, out + errs
}

// rgtRepo is a git work tree with one commit.
func rgtRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")
	writeFile(t, dir+"/base.txt", "base\n")
	gitRun(t, dir, "add", "base.txt")
	gitRun(t, dir, "commit", "-qm", "base")
	return dir
}

var rgtTimes = regexp.MustCompile(`\(\d+s\)|\d+s wall`)

func TestRunGuardTests(t *testing.T) {
	t.Parallel()
	has := func(t *testing.T, out, s string) {
		t.Helper()
		if !strings.Contains(out, s) {
			t.Errorf("output lacks %q:\n%s", s, out)
		}
	}
	lacks := func(t *testing.T, out, s string) {
		t.Helper()
		if strings.Contains(out, s) {
			t.Errorf("output carries %q:\n%s", s, out)
		}
	}
	line := func(t *testing.T, out, re string, want bool) {
		t.Helper()
		if regexp.MustCompile(`(?m)`+re).MatchString(out) != want {
			t.Errorf("line /%s/ present = %v, want %v:\n%s", re, !want, want, out)
		}
	}
	code := func(t *testing.T, rc, want int, out string) {
		t.Helper()
		if rc != want {
			t.Errorf("exit = %d, want %d:\n%s", rc, want, out)
		}
	}

	t.Run("case 1/discovery", func(t *testing.T) {
		t.Parallel()
		_, out := rgtRun1(t, rgtFull(t), nil)
		for _, c := range []struct{ name, s string }{
			{"discovers test-alpha.sh", "test-alpha.sh"},
			{"discovers test-beta.sh", "test-beta.sh"},
			{"discovers test-gamma.sh", "test-gamma.sh"},
		} {
			t.Run(c.name, func(t *testing.T) { has(t, out, c.s) })
		}
		t.Run("does not run helper.sh", func(t *testing.T) { lacks(t, out, "HELPER-SHOULD-NEVER-RUN-MARKER") })
		t.Run("does not descend into subdir", func(t *testing.T) { lacks(t, out, "NESTED-SHOULD-NEVER-RUN-MARKER") })
	})

	t.Run("case 2/ok and FAIL lines and counts", func(t *testing.T) {
		t.Parallel()
		_, out := rgtRun1(t, rgtFull(t), nil)
		t.Run("ok line names test-alpha.sh", func(t *testing.T) { line(t, out, `^ok:.*test-alpha\.sh`, true) })
		t.Run("ok line names test-gamma.sh", func(t *testing.T) { line(t, out, `^ok:.*test-gamma\.sh`, true) })
		t.Run("FAIL line names test-beta.sh", func(t *testing.T) { line(t, out, `^FAIL:.*test-beta\.sh`, true) })
		t.Run("failing harness never ok", func(t *testing.T) { line(t, out, `^ok:.*test-beta\.sh`, false) })
		t.Run("summary names the total", func(t *testing.T) { has(t, out, "3 harnesses") })
		t.Run("summary names passed", func(t *testing.T) { has(t, out, "2 passed") })
		t.Run("summary names failed", func(t *testing.T) { has(t, out, "1 failed") })
	})

	t.Run("case 2 mutation/all passing", func(t *testing.T) {
		t.Parallel()
		_, out := rgtRun1(t, rgtClean(t), nil)
		t.Run("no FAIL line", func(t *testing.T) { line(t, out, `^FAIL:`, false) })
		t.Run("summary names 0 failed", func(t *testing.T) { has(t, out, "0 failed") })
	})

	t.Run("case 3/a failing harness exits non-zero", func(t *testing.T) {
		t.Parallel()
		rc, out := rgtRun1(t, rgtFull(t), nil)
		code(t, rc, 1, out)
	})
	t.Run("case 3/an all-passing fixture exits 0", func(t *testing.T) {
		t.Parallel()
		rc, out := rgtRun1(t, rgtClean(t), nil)
		code(t, rc, 0, out)
	})

	t.Run("case 4/RUN_GUARD_TESTS_ROOT set but empty exits 2", func(t *testing.T) {
		t.Parallel()
		rc, out := rgtRun1(t, "", nil)
		code(t, rc, 2, out)
		has(t, out, "RUN_GUARD_TESTS_ROOT is set but empty")
	})
	t.Run("case 4 mutation/a non-empty root is honoured", func(t *testing.T) {
		t.Parallel()
		rc, out := rgtRun1(t, rgtClean(t), nil)
		code(t, rc, 0, out)
		lacks(t, out, "set but empty")
	})

	t.Run("case 5/clean run replays nothing", func(t *testing.T) {
		t.Parallel()
		_, out := rgtRun1(t, rgtClean(t), nil)
		t.Run("not alpha", func(t *testing.T) { lacks(t, out, "ALPHA-OUT-MARKER") })
		t.Run("not gamma", func(t *testing.T) { lacks(t, out, "GAMMA-OUT-MARKER") })
	})
	t.Run("case 5/failing run replays the failure only", func(t *testing.T) {
		t.Parallel()
		_, out := rgtRun1(t, rgtFull(t), nil)
		t.Run("replays stdout", func(t *testing.T) { has(t, out, "BETA-OUT-MARKER") })
		t.Run("replays stderr", func(t *testing.T) { has(t, out, "BETA-ERR-MARKER") })
		t.Run("not a passing neighbour", func(t *testing.T) { lacks(t, out, "ALPHA-OUT-MARKER") })
		t.Run("names the failing harness", func(t *testing.T) { has(t, out, "---- test-beta.sh ----") })
	})

	t.Run("case 6a/an uncompanioned guard is refused before anything runs", func(t *testing.T) {
		t.Parallel()
		root := rgtClean(t)
		writeFile(t, root+"/check-lonely.sh", "printf 'LONELY-SHOULD-NEVER-RUN-MARKER\\n'\n")
		rc, out, errs := rgtRunIn(t, root, nil)
		code(t, rc, 1, out+errs)
		if out != "" {
			t.Errorf("stdout = %q, want empty", out)
		}
		if want := "run-guard-tests: check-lonely.sh has no companion test-check-lonely.sh\n" +
			"run-guard-tests: every check-*.sh guard gets a test-check-*.sh companion — add the missing harnesses\n"; !strings.HasSuffix(errs, want) {
			t.Errorf("stderr = %q, want suffix %q", errs, want)
		}
		lacks(t, errs, "ALPHA-OUT-MARKER")
	})
	t.Run("case 6b/a balanced fixture passes", func(t *testing.T) {
		t.Parallel()
		root := rgtClean(t)
		writeFile(t, root+"/check-paired.sh", "printf 'PAIRED-GUARD\\n'\n")
		writeFile(t, root+"/test-check-paired.sh", "exit 0\n")
		rc, out := rgtRun1(t, root, nil)
		code(t, rc, 0, out)
		lacks(t, out, "companion")
	})

	t.Run("case 7a/a suite that dirties the watched repository", func(t *testing.T) {
		t.Parallel()
		watched := rgtRepo(t)
		writeFile(t, watched+"/pre.txt", "untracked before the suite\n") // not part of the delta
		root := rgtFixture(t, map[string]string{"test-dirty.sh": "printf 'mutated by a harness case\\n' > \"$GUARD_TESTS_REPO_ROOT/dirt.txt\"\nexit 0\n"})
		rc, out := rgtRun1(t, root, map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 1, out)
		has(t, out, "\nrun-guard-tests: the watched repository tree changed during the suite ("+watched+"):\n?? dirt.txt\n")
		lacks(t, out, "pre.txt")
		line(t, out, `^ok:   test-dirty\.sh`, true)
	})
	t.Run("case 7a/a pure deletion prints the removed-entries line", func(t *testing.T) {
		t.Parallel()
		watched := rgtRepo(t)
		writeFile(t, watched+"/stray.txt", "x\n")
		root := rgtFixture(t, map[string]string{"test-clean-up.sh": "rm \"$GUARD_TESTS_REPO_ROOT/stray.txt\"\n"})
		rc, out := rgtRun1(t, root, map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 1, out)
		has(t, out, "during the suite ("+watched+"):\n  (entries removed relative to the pre-suite snapshot)\n")
	})
	t.Run("case 7b/a sandbox-only suite passes the gate", func(t *testing.T) {
		t.Parallel()
		watched := rgtRepo(t)
		rc, out := rgtRun1(t, rgtClean(t), map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 0, out)
		lacks(t, out, "tree changed during the suite")
	})
	t.Run("case 7c/a non-git watched root skips the gate", func(t *testing.T) {
		t.Parallel()
		watched := t.TempDir()
		root := rgtFixture(t, map[string]string{"test-dirty.sh": "printf x > \"$GUARD_TESTS_REPO_ROOT/dirt.txt\"\n"})
		rc, out, errs := rgtRunIn(t, root, map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 0, out+errs)
		if want := "run-guard-tests: " + watched + " is not a git work tree — skipping the tree gate\n"; errs != want {
			t.Errorf("stderr = %q, want %q", errs, want)
		}
	})
	t.Run("case 7d/GUARD_TESTS_REPO_ROOT set but empty exits 2", func(t *testing.T) {
		t.Parallel()
		rc, out := rgtRun1(t, rgtClean(t), map[string]*string{"GUARD_TESTS_REPO_ROOT": ptr("")})
		code(t, rc, 2, out)
		has(t, out, "GUARD_TESTS_REPO_ROOT is set but empty")
	})
	t.Run("case 7e/a bare watched repository skips the gate", func(t *testing.T) {
		t.Parallel()
		watched := t.TempDir()
		gitRun(t, "", "init", "-q", "--bare", watched)
		rc, out := rgtRun1(t, rgtClean(t), map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 0, out)
		has(t, out, "skipping the tree gate")
	})

	// shim is new_shim_fixture: <parent>/scripts holding a ported guard's shim.
	shim := func(t *testing.T) (string, string) {
		parent := t.TempDir()
		root := parent + "/scripts"
		writeFile(t, root+"/check-x.sh", "flow_guard_exec check-x 2 \"check-x:\" \"$@\"\n")
		writeFile(t, root+"/test-alpha.sh", "exit 0\n")
		return parent, root
	}
	t.Run("case 8/a shim guard with its Go test is companioned", func(t *testing.T) {
		t.Parallel()
		parent, root := shim(t)
		writeFile(t, parent+"/stats/internal/guard/check_x_test.go", "package guard\n")
		rc, out := rgtRun1(t, root, nil)
		code(t, rc, 0, out)
		lacks(t, out, "companion")
	})
	t.Run("case 9/a shim guard with neither companion is refused", func(t *testing.T) {
		t.Parallel()
		_, root := shim(t)
		rc, out := rgtRun1(t, root, nil)
		code(t, rc, 1, out)
		has(t, out, "check-x.sh has no companion")
	})
	t.Run("case 10/harness environment", func(t *testing.T) {
		t.Parallel()
		parent := t.TempDir()
		tmp := parent + "/tmp"
		mkdir(t, tmp)
		root := rgtFixture(t, map[string]string{"test-which.sh": "printf '%s' \"${FLOW_GUARD_CACHE_DIR:-}\" > " + parent + "/which.out\n" +
			"printf '%s' \"${FLOW_GUARD_TELEMETRY:-}\" > " + parent + "/telemetry.out\n"})
		rc, out := rgtRun1(t, root, map[string]*string{"TMPDIR": &tmp})
		code(t, rc, 0, out)
		which, _ := os.ReadFile(parent + "/which.out")
		telemetry, _ := os.ReadFile(parent + "/telemetry.out")
		t.Run("flow-guard cache in the runner's temp dir", func(t *testing.T) {
			if !strings.HasPrefix(string(which), tmp+"/") {
				t.Errorf("FLOW_GUARD_CACHE_DIR = %q, want under %s", which, tmp)
			}
		})
		t.Run("telemetry off", func(t *testing.T) {
			if string(telemetry) != "off" {
				t.Errorf("FLOW_GUARD_TELEMETRY = %q, want off", telemetry)
			}
		})
		t.Run("temp dir removed after the run", func(t *testing.T) {
			if left, _ := os.ReadDir(tmp); len(left) != 0 {
				t.Errorf("%s still holds %v", tmp, left)
			}
		})
	})

	// The parallel library harness's cases, at the runner's level.
	t.Run("parallel case 1/one failure among passes exits 1", func(t *testing.T) {
		t.Parallel()
		root := rgtFixture(t, map[string]string{"test-a.sh": "exit 0\n", "test-b.sh": "exit 1\n", "test-c.sh": "true\n"})
		rc, out := rgtRun1(t, root, nil)
		code(t, rc, 1, out)
	})
	t.Run("parallel case 1/a clean run replays nothing", func(t *testing.T) {
		t.Parallel()
		root := rgtFixture(t, map[string]string{"test-a.sh": "exit 0\n", "test-b.sh": "printf ok\n", "test-c.sh": "true\n"})
		rc, out := rgtRun1(t, root, nil)
		code(t, rc, 0, out)
		lacks(t, out, "----")
		lacks(t, out, "FAILED:")
	})
	t.Run("parallel case 3/replay follows list order under shuffled completion", func(t *testing.T) {
		t.Parallel()
		order := t.TempDir() + "/order"
		job := func(d, m string) string {
			return "sleep " + d + "; printf 'MARK-" + m + "\\n'; printf '" + m + "' >> " + order + "; exit 1\n"
		}
		root := rgtFixture(t, map[string]string{"test-0.sh": job("0.6", "0"), "test-1.sh": job("0.1", "1"), "test-2.sh": job("0.3", "2")})
		_, out := rgtRun1(t, root, map[string]*string{"JOBS": ptr("3")})
		i0, i1, i2 := strings.Index(out, "MARK-0"), strings.Index(out, "MARK-1"), strings.Index(out, "MARK-2")
		if i0 < 0 || i0 > i1 || i1 > i2 {
			t.Errorf("replay not in list order: 0@%d 1@%d 2@%d\n%s", i0, i1, i2, out)
		}
		t.Run("mutation proof: completion order differs", func(t *testing.T) {
			if got, _ := os.ReadFile(order); string(got) == "012" {
				t.Errorf("completion order came back 012 — the shuffle did not shuffle")
			}
		})
	})
	t.Run("parallel case 4/each failure's output is contiguous", func(t *testing.T) {
		t.Parallel()
		job := func(l string) string {
			return "for i in 1 2 3 4 5; do printf '" + l + "-%s\\n' \"$i\"; sleep 0.05; done; exit 1\n"
		}
		root := rgtFixture(t, map[string]string{"test-a.sh": job("A"), "test-b.sh": job("B")})
		_, out, _ := rgtRunIn(t, root, nil)
		want := "---- test-a.sh ----\nA-1\nA-2\nA-3\nA-4\nA-5\n\n---- test-b.sh ----\nB-1\nB-2\nB-3\nB-4\nB-5\n\n"
		if !strings.HasSuffix(out, want) {
			t.Errorf("replay is not two contiguous blocks; want suffix %q:\n%s", want, out)
		}
	})
	t.Run("parallel case 5/JOBS=1 and JOBS=4 are identical", func(t *testing.T) {
		t.Parallel()
		root := rgtFixture(t, map[string]string{
			"test-j1.sh": "printf 'J1-out\\n'; printf 'J1-err\\n' >&2; exit 3\n",
			"test-j2.sh": "printf 'J2-out\\n'; exit 0\n",
		})
		rc1, out1, _ := rgtRunIn(t, root, map[string]*string{"JOBS": ptr("1")})
		rc4, out4, _ := rgtRunIn(t, root, map[string]*string{"JOBS": ptr("4")})
		if rc1 != rc4 || rgtTimes.ReplaceAllString(out1, "") != rgtTimes.ReplaceAllString(out4, "") {
			t.Errorf("JOBS=1 (%d):\n%s\nJOBS=4 (%d):\n%s", rc1, out1, rc4, out4)
		}
	})
	t.Run("parallel case 5b/labelled headers", func(t *testing.T) {
		t.Parallel()
		root := rgtFixture(t, map[string]string{"test-alpha.sh": "exit 0\n", "test-beta.sh": "exit 1\n", "test-gamma.sh": "exit 1\n"})
		_, out := rgtRun1(t, root, nil)
		has(t, out, "---- test-beta.sh ----\n\n---- test-gamma.sh ----\n\n")
		lacks(t, out, "---- test-alpha.sh ----")
	})
	for _, c := range []struct{ jobs, msg string }{
		{"abc", "got 'abc'"}, {"0", "got 0"}, {"-1", "got '-1'"}, {"3.5", "got '3.5'"}, {"", "got ''"},
	} {
		t.Run("parallel case 6/JOBS="+c.jobs+" is refused", func(t *testing.T) {
			t.Parallel()
			rc, out, errs := rgtRunIn(t, rgtClean(t), map[string]*string{"JOBS": ptr(c.jobs)})
			code(t, rc, 2, out+errs)
			if want := "parallel_job_count: JOBS must be a positive integer, " + c.msg + "\nrun-guard-tests: could not run the harnesses (see above)\n"; !strings.HasSuffix(errs, want) || out != "" {
				t.Errorf("stdout %q stderr %q, want stderr suffix %q", out, errs, want)
			}
		})
	}
	t.Run("parallel case 6/JOBS=5 and unset JOBS run", func(t *testing.T) {
		t.Parallel()
		for _, v := range []*string{ptr("5"), nil} {
			rc, out := rgtRun1(t, rgtClean(t), map[string]*string{"JOBS": v})
			code(t, rc, 0, out)
		}
	})

	t.Run("durations/longest first, unrecorded first of all, rewritten", func(t *testing.T) {
		t.Parallel()
		watched := rgtRepo(t)
		durations := watched + "/.git/guard-test-durations"
		writeFile(t, durations, "1\ttest-a.sh\n9\ttest-c.sh\n")
		root := rgtFixture(t, map[string]string{"test-a.sh": "exit 0\n", "test-b.sh": "exit 0\n", "test-c.sh": "exit 0\n"})
		rc, out := rgtRun1(t, root, map[string]*string{"GUARD_TESTS_REPO_ROOT": &watched})
		code(t, rc, 0, out)
		if !regexp.MustCompile(`^ok:   test-b\.sh \(\d+s\)\nok:   test-c\.sh \(\d+s\)\nok:   test-a\.sh \(\d+s\)\n`).MatchString(out) {
			t.Errorf("want b (unrecorded), c (9), a (1):\n%s", out)
		}
		// A harness straddling a second boundary records 1s: whole Unix
		// seconds, per the elapsed-unix-seconds decision.
		if got, _ := os.ReadFile(durations); !regexp.MustCompile(`^[01]\ttest-b\.sh\n[01]\ttest-c\.sh\n[01]\ttest-a\.sh\n$`).Match(got) {
			t.Errorf("durations = %q", got)
		}
	})

	// The SIGINT case runs the real binary, so the signal reaches only it.
	t.Run("SIGINT removes the runner's temp dir", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		root := rgtFixture(t, map[string]string{"test-slow.sh": "sleep 3\n"})
		cmd := exec.Command(guardBinary(t), "run-guard-tests")
		cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "RUN_GUARD_TESTS_ROOT="+root, "GUARD_TESTS_REPO_ROOT="+t.TempDir(), "FLOW_GUARD_TELEMETRY=off")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(10 * time.Second)
		for {
			if left, _ := os.ReadDir(tmp); len(left) > 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("the runner's temp dir never appeared")
			}
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(200 * time.Millisecond) // the harness is running
		_ = cmd.Process.Signal(syscall.SIGINT)
		err := cmd.Wait()
		if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); !ok || !ws.Signaled() || ws.Signal() != syscall.SIGINT {
			t.Errorf("runner did not die of SIGINT: %v", err)
		}
		if left, _ := os.ReadDir(tmp); len(left) != 0 {
			t.Errorf("%s still holds %v", tmp, left)
		}
	})
}
