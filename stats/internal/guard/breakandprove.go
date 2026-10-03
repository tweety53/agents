package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

// breakAndProve is scripts/break-and-prove.sh: that script's header comment
// is the contract — mutate <file> by --sed or --patch, run the test command
// (after --clean, when given), assert it fails, restore <file> from its
// pre-mutation bytes, run it again, assert it passes, and print both runs'
// output; exit 0 proof held, 1 it did not, 2 refused before mutating anything
// or post-restore drift, 3 could not restore, 4 cannot answer. The reasoning
// for each step, moved here from the bash body it replaced (9cd35da8), sits
// beside the code it explains.
//
// THE RESTORE IS A PRE-MUTATION SNAPSHOT of the file's bytes, never
// `git checkout --`: one of the observed failures this guard exists to
// prevent is an agent whose `git checkout --` restore also reverted the
// uncommitted edits it had made on the same file. So uncommitted edits on
// <file> survive the loop, the guard deliberately does NOT refuse a dirty
// <file> the way mutate-and-verify does, and no checkout call appears in this
// file, structurally (break_and_prove_test.go case 5).
func init() {
	Registry["break-and-prove"] = breakAndProve
}

const bpUsageLine = "break-and-prove: usage: break-and-prove.sh [--clean <command>] <file> " +
	"(--sed <expr> | --patch <patchfile>) -- <test-command> [<args>...]"

func breakAndProve(args []string, env Env, stdout, stderr io.Writer) int {
	usageFail := func() int {
		fmt.Fprintln(stderr, bpUsageLine)
		return 4
	}
	cannotAnswer := func(msg string) int {
		fmt.Fprintf(stderr, "break-and-prove: cannot answer — %s\n", msg)
		return 4
	}
	refuse := func(msg string) int {
		fmt.Fprintf(stderr, "break-and-prove: refused — %s — nothing was mutated\n", msg)
		return 2
	}

	// The flags may come in any order before `--`; any other word is <file>,
	// and each of the three may be given once.
	var cleanCmd, fileArg, kind, value string
	var testCmd []string
	for i := 0; i < len(args); {
		switch a := args[i]; a {
		case "--clean":
			if i+1 >= len(args) || cleanCmd != "" {
				return usageFail()
			}
			cleanCmd = args[i+1]
			i += 2
		case "--sed", "--patch":
			if i+1 >= len(args) || kind != "" {
				return usageFail()
			}
			kind, value = strings.TrimPrefix(a, "--"), args[i+1]
			i += 2
		case "--":
			testCmd = args[i+1:]
			i = len(args)
		default:
			if fileArg != "" {
				return usageFail()
			}
			fileArg = a
			i++
		}
	}
	if fileArg == "" || kind == "" || len(testCmd) == 0 {
		return usageFail()
	}

	// The file and patch arguments resolve to absolute paths against the
	// caller's own cwd BEFORE anything runs from the repository root — a
	// relative argument must keep meaning what the caller meant by it.
	// Joined as bash joined them, "$ORIG_PWD/$1", uncleaned. toAbs and the
	// usageFail/cannotAnswer/refuse trio deliberately repeat
	// mutateandverify.go's rather than moving into a shared helper: tiny,
	// program-name-parameterized, and an extraction would be the wrong
	// abstraction for helpers this small.
	origPwd := env.Dir
	toAbs := func(p string) string {
		if strings.HasPrefix(p, "/") {
			return p
		}
		return origPwd + "/" + p
	}
	file := toAbs(fileArg)
	patch := ""
	if kind == "patch" {
		patch = toAbs(value)
	}

	// Step 1: resolve the repository root; the clean command, the test
	// command and every git call below run from it.
	top, rc := gitExec(origPwd, nil, io.Discard, "rev-parse", "--show-toplevel")
	if rc != 0 {
		return cannotAnswer("not inside a git worktree: " + origPwd)
	}
	root := strings.TrimRight(top, "\n")

	// Step 2: the target file must exist and be readable — the loop mutates
	// an existing config file; creating one is not part of the proof.
	if !isFile(file) || syscall.Access(file, 4) != nil {
		return refuse("target file does not exist or is not readable: " + fileArg)
	}

	if patch != "" {
		// Step 3: the patch, when given, must exist and be readable.
		if !isFile(patch) || syscall.Access(patch, 4) != nil {
			return cannotAnswer("cannot read patch file: " + value)
		}

		// Step 4: git apply --check — refuse (exit 2) without touching
		// anything.
		check := exec.Command("git", "apply", "--check", "--whitespace=nowarn", patch)
		check.Dir = root
		if out, err := check.CombinedOutput(); err != nil {
			fmt.Fprintln(stderr, "break-and-prove: git apply --check failed:")
			fmt.Fprintln(stderr, strings.TrimRight(string(out), "\n"))
			return refuse("patch does not apply cleanly: " + value)
		}

		// Step 5: the patch must touch exactly <file> — the operator named
		// the one file being proven, and the snapshot/restore contract covers
		// exactly that file, so a patch reaching further is refused rather
		// than silently narrowed. The touched file is numstat's third column
		// (cut -f3: a line without a tab is kept whole).
		numstat, _ := gitExec(root, nil, stderr, "apply", "--numstat", patch)
		count, touched := 0, ""
		for _, line := range strings.Split(numstat, "\n") {
			if line == "" {
				continue
			}
			count++
			touched = line
			if f := strings.Split(line, "\t"); len(f) >= 3 {
				touched = f[2]
			} else if len(f) == 2 {
				touched = ""
			}
		}
		if count != 1 {
			return refuse(fmt.Sprintf("patch touches %d files — it must touch exactly %s", count, fileArg))
		}
		// Compare canonicalized paths: a caller cwd reached through a
		// symlink (macOS's /tmp → /private/tmp) yields an absolute path that
		// shares no string prefix with git's physical toplevel. Cleaned
		// first, as bash's logical `cd` read a `..`, then resolved, as
		// `pwd -P` printed it.
		dir, _ := filepath.EvalSymlinks(filepath.Clean(filepath.Dir(file)))
		repoCanon, _ := filepath.EvalSymlinks(root)
		rel, inside := strings.CutPrefix(dir+"/"+filepath.Base(file), repoCanon+"/")
		if !inside {
			return refuse("target file is outside the repository: " + fileArg)
		}
		if touched != rel {
			return refuse(fmt.Sprintf("patch touches %s, not %s", touched, rel))
		}
	}

	// Step 6: the pre-mutation byte snapshot — the restore source, never
	// `git checkout --`.
	snap, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(stderr, "break-and-prove: %s\n", bpErrText(file, err))
		return 1
	}

	// Step 7: snapshot the tree the mutation starts from — status and stash
	// list — so the exit path can tell post-restore residue from the state
	// the run actually left.
	tree, rc := snapshotTreeState(root, stderr)
	if rc != 0 {
		return rc
	}

	// Step 8: from here every exit goes through r.exit, the bash's EXIT
	// trap, installed before anything mutates — a signal included: restore
	// and report as for a run whose last status was 0, then die of it.
	r := &bpRun{file: file, fileArg: fileArg, snap: snap, root: root, tree: tree, stdout: stdout, stderr: stderr}
	defer trapExitSignals(func() {
		r.dying.Store(true)
		r.exit(0)
	})()
	// A write to a closed stdout halts the main flow, so no leg starts while
	// the trap restores; the trap's own report goes through r.stdout.
	rc = r.main(kind, value, patch, cleanCmd, testCmd, env, mvHaltOnEPIPE{stdout}, cannotAnswer, refuse)
	rc = r.exit(rc)
	if r.dying.Load() {
		select {} // the signal goroutine ends the process
	}
	return rc
}

// bpRun is the state the bash's EXIT trap read: whether the mutation is
// applied, whether the main flow already restored and verified it, and the
// pre-mutation snapshots.
type bpRun struct {
	file, fileArg, root, tree string
	snap                      []byte
	stdout, stderr            io.Writer

	mu                sync.Mutex // held across each write to file, so the trap's restore never interleaves one
	applied, restored bool
	once              sync.Once
	rc                int
	dying             atomic.Bool // a signal is ending the process
}

// lock takes mu for a write to file; once a signal is being handled the
// trap has restored, or is about to, so the main flow never writes again.
func (r *bpRun) lock() {
	r.mu.Lock()
	if r.dying.Load() {
		r.mu.Unlock()
		select {} // the signal goroutine ends the process
	}
}

// main is steps 9-12: the mutation and the two legs. It returns the exit
// code the trap starts from.
func (r *bpRun) main(kind, value, patch, cleanCmd string, testCmd []string, env Env, stdout io.Writer,
	cannotAnswer, refuse func(string) int) int {
	// Step 9: apply the mutation.
	fmt.Fprintf(stdout, "break-and-prove: proving %s\n", r.fileArg)
	if kind == "sed" {
		var mutated bytes.Buffer
		sed := exec.Command("sed", "-e", value, r.file)
		sed.Dir, sed.Stdout, sed.Stderr = r.root, &mutated, r.stderr
		if rc := exitCode(sed.Run()); rc != 0 {
			return refuse(fmt.Sprintf("sed failed with exit %d: %s", rc, value))
		}
		if bytes.Equal(mutated.Bytes(), r.snap) {
			return refuse("sed expression changes nothing — the mutation must alter the file: " + value)
		}
		fmt.Fprintf(stdout, "break-and-prove: mutating via sed: %s\n", value)
		r.lock()
		f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		if err != nil {
			r.mu.Unlock()
			fmt.Fprintf(r.stderr, "break-and-prove: %s\n", bpErrText(r.file, err))
			return 1
		}
		// Applied from the truncate on, not once the write succeeds as the
		// bash's APPLIED=1 did: a write that fails part-way leaves the file
		// neither original nor mutated, and only the restore repairs it.
		r.applied = true
		_, werr := f.Write(mutated.Bytes())
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		r.mu.Unlock()
		if werr != nil {
			fmt.Fprintf(r.stderr, "break-and-prove: %s\n", bpErrText(r.file, werr))
			return 1
		}
	} else {
		fmt.Fprintf(stdout, "break-and-prove: mutating via patch: %s\n", value)
		r.lock()
		cmd := exec.Command("git", "apply", "--whitespace=nowarn", patch)
		cmd.Dir, cmd.Stdout, cmd.Stderr = r.root, r.stdout, r.stderr
		rc := exitCode(cmd.Run())
		r.applied = rc == 0
		r.mu.Unlock()
		if rc != 0 {
			return rc
		}
	}

	// Step 10: run 1, against the mutated file — it must fail. The 126/127
	// guard appears once per leg, on purpose (WET): whichever leg the bad
	// exit lands on, the answer is the same cannot-answer.
	rc, ok := r.leg("run 1 (mutated)", cleanCmd, testCmd, env, stdout)
	if !ok {
		return cannotAnswer(fmt.Sprintf("clean command exited %d — the proof environment is broken", rc))
	}
	fmt.Fprintf(stdout, "break-and-prove: run 1 exited %d\n", rc)
	if rc == 126 || rc == 127 {
		return cannotAnswer(fmt.Sprintf("test command exited %d (126/127) — an exec failure or its own such exit, not usable as evidence", rc))
	}
	if rc == 0 {
		fmt.Fprintln(r.stderr, "break-and-prove: run 1 exited 0 — the proof did not hold: the mutation did not break the test")
		return 1
	}

	// Step 11: restore from the pre-mutation snapshot, verified byte-exact
	// before run 2 is trusted.
	r.lock()
	r.restore()
	if !r.matches() {
		r.mu.Unlock()
		return 3
	}
	r.restored = true
	r.mu.Unlock()
	fmt.Fprintf(stdout, "break-and-prove: %s restored byte-exact\n", r.fileArg)

	// Step 12: run 2, against the restored file — it must pass.
	rc, ok = r.leg("run 2 (restored)", cleanCmd, testCmd, env, stdout)
	if !ok {
		return cannotAnswer(fmt.Sprintf("clean command exited %d — the proof environment is broken", rc))
	}
	fmt.Fprintf(stdout, "break-and-prove: run 2 exited %d\n", rc)
	if rc == 126 || rc == 127 {
		return cannotAnswer(fmt.Sprintf("test command exited %d (126/127) — an exec failure or its own such exit, not usable as evidence", rc))
	}
	if rc != 0 {
		fmt.Fprintf(r.stderr, "break-and-prove: run 2 exited %d — the proof did not hold: the test fails on the restored file\n", rc)
		return 1
	}

	fmt.Fprintf(stdout, "break-and-prove: proof held — mutated run failed (exit was non-zero), restored run passed, %s restored byte-exact\n", r.fileArg)
	return 0
}

// leg is run_leg: the clean re-run first, then the test command's argv from
// the repository root, its stdout and stderr captured together and printed
// in a labeled, quotable block. ok is false, with the clean command's status,
// when the clean command failed: neither leg is evidence then.
//
// THE CLEAN RE-RUN IS THE GUARD'S JOB, not the caller's memory: Gradle
// silently skips a re-run when only a compose or YAML file changed, and an
// agent that forgot `:app:cleanTest` recorded a stale green as evidence. The
// clean command is the operator's own, printed verbatim before it runs, and
// is the one string this guard evaluates through a shell; the test command
// is an argv vector, never a string through a shell.
func (r *bpRun) leg(label, cleanCmd string, testCmd []string, env Env, stdout io.Writer) (int, bool) {
	if cleanCmd != "" {
		fmt.Fprintf(stdout, "break-and-prove: clean — sh -c %s\n", smcQuote(cleanCmd, smcUTF8(env)))
		clean := exec.Command("sh", "-c", cleanCmd)
		// r.stdout, not the EPIPE wrapper: an *os.File is inherited as the
		// bash's fd was, so Run() returns once `sh` exits, never waiting on a
		// child the clean command backgrounded that still holds a pipe.
		clean.Dir, clean.Stdin, clean.Stdout, clean.Stderr = r.root, env.Stdin, r.stdout, r.stderr
		if rc := exitCode(clean.Run()); rc != 0 {
			return rc, false
		}
	}
	fmt.Fprintf(stdout, "break-and-prove: --- begin %s output ---\n", label)
	var out bytes.Buffer
	rc := bpRunArgv(testCmd, r.root, env.Stdin, &out)
	// $(...) dropped every NUL byte and every trailing newline (bash 5 also
	// warned "ignored null byte in input" on stderr, the one byte of its
	// output not reproduced here).
	if s := strings.TrimRight(strings.ReplaceAll(out.String(), "\x00", ""), "\n"); s != "" {
		fmt.Fprintln(stdout, s)
	} else {
		fmt.Fprintln(stdout, "(no output)")
	}
	fmt.Fprintf(stdout, "break-and-prove: --- end %s output ---\n", label)
	return rc, true
}

// bpRunArgv runs argv from dir with its stdout and stderr on out and returns
// the `$?` bash would see: 127 with bash's diagnostic when the command is
// missing, 126 when it cannot be executed, and a file with no #! line run
// as a bash script, as bash ran it. Bash put its own diagnostic after its
// script's path and line; this guard puts "break-and-prove: " there.
func bpRunArgv(argv []string, dir string, stdin io.Reader, out io.Writer) int {
	// A bash builtin with no binary of that name (`eval`, `exit`, `:`) ran
	// as the builtin in the bash; it runs in a bash here too.
	if !strings.Contains(argv[0], "/") {
		if _, err := exec.LookPath(argv[0]); errors.Is(err, exec.ErrNotFound) && bpIsBuiltin(argv[0]) {
			sh := exec.Command("bash", append([]string{"-c", `"$@"`, "break-and-prove"}, argv...)...)
			sh.Dir, sh.Stdin, sh.Stdout, sh.Stderr = dir, stdin, out, out
			return exitCode(sh.Run())
		}
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, stdin, out, out
	// Bash runs a command a relative PATH entry finds, relative to its own
	// cwd — dir here. Go resolves a relative entry against the process cwd
	// and refuses the hit with ErrDot, so the PATH walk is redone from dir.
	if p, ok := bpRelativePathHit(argv[0], dir); ok {
		cmd.Path, cmd.Err = p, nil
	}
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil || errors.As(err, &ee):
		return exitCode(err)
	case errors.Is(err, exec.ErrNotFound):
		fmt.Fprintf(out, "break-and-prove: %s: command not found\n", argv[0])
		return 127
	case errors.Is(err, syscall.ENOEXEC):
		sh := exec.Command("bash", argv...)
		sh.Dir, sh.Stdin, sh.Stdout, sh.Stderr = dir, stdin, out, out
		return exitCode(sh.Run())
	case errors.Is(err, fs.ErrNotExist):
		fmt.Fprintf(out, "break-and-prove: %s\n", bpErrText(argv[0], err))
		return 127
	}
	fmt.Fprintf(out, "break-and-prove: %s\n", bpErrText(argv[0], err))
	return 126
}

// bpErrText is bash's "<path>: <strerror>" for err: the errno's text with
// its first letter capitalised, as strerror spells it.
func bpErrText(path string, err error) string {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return path + ": " + err.Error()
	}
	msg := errno.Error()
	return path + ": " + strings.ToUpper(msg[:1]) + msg[1:]
}

// restore is `cat "$SNAP" > "$FILE_ABS" 2>/dev/null || true`: a write that
// fails is ignored, but a file that cannot be opened for it is reported, as
// bash reported a failed redirection before the 2>/dev/null took effect.
func (r *bpRun) restore() {
	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		fmt.Fprintf(r.stderr, "break-and-prove: %s\n", bpErrText(r.file, err))
		return
	}
	_, _ = f.Write(r.snap)
	_ = f.Close()
}

// matches is `cmp -s` of the file against its snapshot: a file that cannot
// be read differs.
func (r *bpRun) matches() bool {
	b, err := os.ReadFile(r.file)
	return err == nil && bytes.Equal(b, r.snap)
}

// exit is the bash's EXIT trap, run once. If the mutation was applied and
// the main flow has not yet verified the restore, restore from the snapshot;
// a file that still differs forces exit 3 whatever the run was carrying, so
// a mutated file is never left behind unreported. Then diff the whole tree
// against the step 7 snapshot (postmutationcheck.go): any new stash entry or
// unexpected status line is reported and forces exit 2, unless exit 3
// already took precedence — the precedence mutate-and-verify applies.
func (r *bpRun) exit(rc int) int {
	r.once.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.rc = rc
		if !r.applied {
			return
		}
		if !r.restored {
			r.restore()
		}
		if !r.matches() {
			fmt.Fprintf(r.stderr, "break-and-prove: could not restore %s byte-exact — residual content differs\n", r.fileArg)
			r.rc = 3
		} else if !r.restored {
			fmt.Fprintf(r.stdout, "break-and-prove: %s restored byte-exact (restored by the exit trap)\n", r.fileArg)
		}
		if drift, _ := checkTreeRestored(r.root, r.tree, r.stderr); len(drift) > 0 {
			fmt.Fprintln(r.stderr, "break-and-prove: post-restore drift detected:")
			for _, d := range drift {
				fmt.Fprintln(r.stderr, d)
			}
			if r.rc != 3 {
				r.rc = 2
			}
		}
	})
	return r.rc
}

// bpIsBuiltin reports whether bash's `type -t` names name a builtin.
func bpIsBuiltin(name string) bool {
	out, err := exec.Command("bash", "-c", `type -t -- "$1"`, "break-and-prove", name).Output()
	return err == nil && strings.TrimSpace(string(out)) == "builtin"
}

// bpRelativePathHit walks PATH as bash does for a name without a slash and
// reports the first executable hit when it lies under a relative entry ("" is
// "."), joined onto dir. An earlier hit under an absolute entry wins, as it
// does in bash, and reports false: exec's own lookup already finds it.
func bpRelativePathHit(name, dir string) (string, bool) {
	if strings.Contains(name, "/") {
		return "", false
	}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if d == "" {
			d = "."
		}
		p := filepath.Join(d, name)
		if !filepath.IsAbs(d) {
			p = filepath.Join(dir, p)
		}
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() && fi.Mode()&0o111 != 0 {
			return p, !filepath.IsAbs(d)
		}
	}
	return "", false
}
