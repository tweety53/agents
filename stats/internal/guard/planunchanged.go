package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkPlanUnchanged is scripts/check-plan-unchanged.sh: that script's
// header comment is the contract -- snapshot writes the plan tree's git
// status to a file, verify recomputes it and compares byte for byte; exit 0
// written or unchanged, 1 changed (each difference on stderr), 2 cannot
// answer. The reasoning for each step, moved here from the bash body it
// replaced (c9b8fde8), sits beside the code it explains.

func init() {
	Registry["check-plan-unchanged"] = checkPlanUnchanged
}

func checkPlanUnchanged(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	mode, worktree, name, snap := arg(0), arg(1), arg(2), arg(3)
	usage := func() int {
		fmt.Fprintln(stderr, "check-plan-unchanged: usage: check-plan-unchanged.sh snapshot|verify <worktree> <change-name> <snapshot-file>")
		return 2
	}
	if mode != "snapshot" && mode != "verify" {
		return usage()
	}
	if worktree == "" || !isDir(smcAbs(env, worktree)) {
		shown := worktree
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "check-plan-unchanged: not a directory: %s\n", shown)
		return 2
	}
	if name == "" || snap == "" {
		return usage()
	}

	// CONTAINMENT: the name lands in a git pathspec, so `../escape` and a
	// glob metacharacter are hazards here exactly as they are in a
	// state-file-derived change name elsewhere. plainChangeName is the
	// package's one copy of the allowlist (changename.go is canonical for
	// why it is enumerated).
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "check-plan-unchanged: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return 2
	}

	// Canonicalise before git runs against it (`cd && pwd -P`), and refuse
	// rather than proceed if it vanished between the directory check above
	// and here.
	wt, err := filepath.EvalSymlinks(smcAbs(env, worktree))
	if err == nil {
		wt, err = filepath.Abs(wt)
	}
	if err != nil {
		fmt.Fprintf(stderr, "check-plan-unchanged: worktree vanished: %s\n", worktree)
		return 2
	}

	// The bash exported LC_ALL=C for its whole run; git and diff inherit it.
	git := envGit(env, "LC_ALL=C")
	planStatus := func(out io.Writer) bool {
		cmd := git("-C", wt, "status", "--porcelain", "--untracked-files=all", "--",
			"spectre/changes/"+name, "spectre/changes/archive/"+name)
		cmd.Stdout, cmd.Stderr = out, stderr
		return cmd.Run() == nil
	}
	gitFailed := func() int {
		fmt.Fprintf(stderr, "check-plan-unchanged: git status failed in %s -- cannot answer (is it a git repository?)\n", wt)
		return 2
	}
	snapPath := smcAbs(env, snap)

	if mode == "snapshot" {
		// `plan_status > "$SNAP"`: the file is opened (truncated) before git
		// runs, so an unwritable path never reaches git, and any failure
		// removes the file rather than leave a partial snapshot behind.
		f, err := os.Create(snapPath)
		if err != nil {
			fmt.Fprintf(stderr, "check-plan-unchanged: %s: %v\n", snap, err)
			_ = os.Remove(snapPath)
			return gitFailed()
		}
		ok := planStatus(f)
		if cerr := f.Close(); cerr != nil {
			ok = false
		}
		if !ok {
			_ = os.Remove(snapPath)
			return gitFailed()
		}
		fmt.Fprintf(stdout, "PLAN-SNAPSHOT-OK: spectre/changes/%s status pinned at %s\n", name, snap)
		return 0
	}

	// THE SNAPSHOT IS THE EVIDENCE: a missing or unreadable one is cannot
	// answer -- snapshot runs before the dispatch, never after it.
	want, err := os.ReadFile(snapPath)
	if fi, serr := os.Stat(snapPath); serr != nil || !fi.Mode().IsRegular() || err != nil {
		fmt.Fprintf(stderr, "check-plan-unchanged: no readable snapshot at %s -- cannot answer; snapshot runs before the dispatch, never after it\n", snap)
		return 2
	}
	var got bytes.Buffer
	if !planStatus(&got) {
		return gitFailed()
	}
	if bytes.Equal(want, got.Bytes()) {
		fmt.Fprintf(stdout, "PLAN-UNCHANGED-OK: spectre/changes/%s matches its pre-dispatch snapshot\n", name)
		return 0
	}
	fmt.Fprintf(stderr, "check-plan-unchanged: spectre/changes/%s changed across the review dispatch -- the reviewer's own clean-or-restored claim is not evidence; stop and recover by hand from the pre-dispatch commit\n", name)
	// `diff "$SNAP" "$current" | sed 's/^/check-plan-unchanged: /' >&2`:
	// diff(1)'s normal format names each difference; the current status
	// goes through a temp file, as the bash's mktemp did.
	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	cur, err := os.CreateTemp(tmpdir, "check-plan-unchanged-verify.")
	if err != nil {
		fmt.Fprintln(stderr, "check-plan-unchanged: mktemp failed -- cannot answer")
		return 2
	}
	defer os.Remove(cur.Name())
	_, werr := cur.Write(got.Bytes())
	if cerr := cur.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		fmt.Fprintln(stderr, "check-plan-unchanged: mktemp failed -- cannot answer")
		return 2
	}
	diffBin := "diff"
	if p, ok := lookPath(env, "diff"); ok {
		diffBin = p
	}
	d := exec.Command(diffBin, snapPath, cur.Name())
	d.Env = append(os.Environ(), "LC_ALL=C")
	out, _ := d.Output()
	for _, l := range strings.SplitAfter(string(out), "\n") {
		if l != "" {
			fmt.Fprint(stderr, "check-plan-unchanged: "+l)
		}
	}
	return 1
}
