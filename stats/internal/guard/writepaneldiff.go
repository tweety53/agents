package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

// writePanelDiff is scripts/write-panel-diff.sh: that script's header
// comment is the contract -- one panel diff file and its `.touched` list
// (the `[TOUCHED_FILES]` list), both sectioned per worktree and written
// atomically into <canonical-wt>/.superpowers/sdd/, both paths printed;
// exit 0 written, 2 cannot answer with nothing written.
func init() {
	Registry["write-panel-diff"] = writePanelDiff
}

const wpdUsage = "usage: write-panel-diff.sh final <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]\n" +
	"       write-panel-diff.sh late-fix <canonical-wt> <wt> <since-close-sha> [<wt> <since-close-sha>...]\n" +
	"       write-panel-diff.sh fix-round <N> <canonical-wt> <wt> <fix-base> [<wt> <fix-base>...]\n" +
	"       write-panel-diff.sh slot-delta <round> <slot> <canonical-wt> <wt> <merge-base> <held-sha|-> [...]\n"

var (
	wpdNumber = regexp.MustCompile(`^[0-9]+$`)
	wpdSlot   = regexp.MustCompile(`^[A-Za-z0-9._+-]+$`)
)

func writePanelDiff(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func() int {
		fmt.Fprint(stderr, wpdUsage)
		return 2
	}
	if len(args) == 0 {
		return usage()
	}
	kind, args := args[0], args[1:]
	var file string
	width := 2 // arguments per worktree
	switch kind {
	case "final":
		file = "final-review.diff"
	case "late-fix":
		file = "late-fix.diff"
	case "fix-round":
		if len(args) == 0 || !wpdNumber.MatchString(args[0]) {
			return usage()
		}
		file, args = "fix-round-"+args[0]+".diff", args[1:]
	case "slot-delta":
		if len(args) < 2 || !wpdNumber.MatchString(args[0]) || !wpdSlot.MatchString(args[1]) || args[1] == "." || args[1] == ".." {
			return usage()
		}
		file, args, width = "slot-delta-"+args[0]+"-"+args[1]+".diff", args[2:], 3
	default:
		return usage()
	}
	if len(args) < 1+width || (len(args)-1)%width != 0 {
		return usage()
	}
	canon, rest := args[0], args[1:]
	if !isDir(smcAbs(env, canon)) {
		fmt.Fprintf(stderr, "write-panel-diff: %s is not a directory — cannot determine anything\n", canon)
		return 2
	}
	gitBin, ok := panelResolveGit(env, "write-panel-diff", stderr)
	if !ok {
		return 2
	}

	var diff, touched bytes.Buffer
	if kind == "final" {
		diff.WriteString("# final-review.diff — working tree vs merge-base, unstaged changes included\n")
	}
	for i := 0; i < len(rest); i += width {
		wt, base := rest[i], rest[i+1]
		if !panelValidateWorktree(env, "write-panel-diff", wt, base, gitBin, stderr) {
			return 2
		}
		// The range each kind reads, as the prose recipes spelled it.
		var rng []string
		switch kind {
		case "final", "late-fix":
			rng = []string{base}
		case "fix-round":
			rng = []string{base + "..HEAD"}
		case "slot-delta":
			rng = []string{base}
			if held := rest[i+2]; held != "-" {
				if !panelValidateWorktree(env, "write-panel-diff", wt, held, gitBin, stderr) {
					return 2
				}
				rng = []string{held, "HEAD"}
			}
		}
		header := fmt.Sprintf("# worktree: %s — merge base %s\n", wt, base)
		for _, o := range []struct {
			buf   *bytes.Buffer
			flags []string
		}{{&diff, nil}, {&touched, []string{"--name-status"}}} {
			out, err := panelGit(env, gitBin, wt, append(append([]string{"diff"}, o.flags...), rng...)...).Output()
			if err != nil {
				fmt.Fprintf(stderr, "write-panel-diff: git diff %v failed in %s\n", append(o.flags, rng...), wt)
				return 2
			}
			o.buf.WriteString(header)
			o.buf.Write(out)
		}
	}

	dir := smcAbs(env, canon) + "/.superpowers/sdd"
	path := canon + "/.superpowers/sdd/" + file
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
		return 2
	}
	for _, f := range []struct {
		name string
		body []byte
	}{{file, diff.Bytes()}, {file + ".touched", touched.Bytes()}} {
		if err := wpdWriteAtomic(filepath.Join(dir, f.name), f.body); err != nil {
			fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
			return 2
		}
	}
	// A final write opens a full pass 1, so a late-fix.diff left by an
	// earlier run is stale by construction: removed, so no slot can read it
	// in place of final-review.diff.
	if kind == "final" {
		for _, f := range []string{"late-fix.diff", "late-fix.diff.touched"} {
			if err := os.Remove(filepath.Join(dir, f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
				return 2
			}
		}
	}
	fmt.Fprintf(stdout, "%s\n%s.touched\n", path, path)
	return 0
}

// wpdWriteAtomic writes body to a temp file beside path and renames it over
// path, so a reader never sees a half-written diff.
func wpdWriteAtomic(path string, body []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	// CreateTemp's 0600 would differ from the recipe's `>` redirection.
	err = f.Chmod(0o644)
	if err == nil {
		_, err = f.Write(body)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		_ = os.Remove(f.Name())
	}
	return err
}
