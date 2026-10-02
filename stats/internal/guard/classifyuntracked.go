package guard

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// classifyUntracked is scripts/classify-untracked.sh: that script's header
// comment is the contract -- one stdout line per action taken or finding
// reported (exactly `CLEAN` for a clean worktree, nothing for an absent
// one), exit 0 whenever it could answer, exit 2 when it cannot. It has no
// exit 1: it never refuses, it only settles and reports; refusing stays
// prepare-archive-branch's job.
func init() {
	Registry["classify-untracked"] = classifyUntracked
}

func classifyUntracked(args []string, env Env, stdout, stderr io.Writer) int {
	landing := ""
	if len(args) > 0 {
		landing = args[0]
	}
	if landing == "" {
		fmt.Fprintln(stderr, "usage: classify-untracked.sh <landing-worktree>")
		return 2
	}
	abs := pcAbs(env, landing)
	if _, err := os.Stat(abs); err != nil {
		// Nothing to classify; prepare-archive-branch will create it fresh.
		return 0
	}
	if !isDir(abs) {
		fmt.Fprintf(stderr, "classify-untracked: %s is not a directory\n", landing)
		return 2
	}
	// The bash exported LC_ALL=C for its whole run; git inherits it here too.
	git := envGit(env, "LC_ALL=C")
	if git("-C", landing, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "classify-untracked: %s is not a git worktree\n", landing)
		return 2
	}
	common, ok := capture(git("-C", landing, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	if !ok {
		fmt.Fprintf(stderr, "classify-untracked: cannot resolve the common git directory of %s\n", landing)
		return 2
	}
	scratch := filepath.Dir(common) + "/.worktrees/_scratchpad"

	// `--directory` collapses an untracked directory to one `<dir>/` entry,
	// so a directory is classified whole -- by its own name for `.claude`,
	// otherwise as an asset, never by a file inside it: no whole directory
	// is moved on the strength of one file's extension. `-z` reads each name
	// raw: the bash read the quoted form core.quotePath prints, so a
	// non-ASCII screenshot's name ended in `"`, missed the capture class and
	// was reported as an asset under its quoted spelling. A failed listing is
	// cannot-answer, never `CLEAN` -- the bash read it inside a heredoc,
	// where `set -e` does not reach, and reported a corrupt index as clean.
	ls := git("-C", landing, "ls-files", "-z", "--others", "--exclude-standard", "--directory")
	ls.Stderr = stderr
	out, err := ls.Output()
	if err != nil {
		fmt.Fprintf(stderr, "classify-untracked: cannot list the untracked entries of %s\n", landing)
		return 2
	}
	var entries []string
	for _, e := range strings.Split(string(out), "\x00") {
		if e != "" {
			entries = append(entries, e)
		}
	}
	if len(entries) == 0 {
		fmt.Fprintln(stdout, "CLEAN")
		return 0
	}

	for _, entry := range entries {
		switch {
		case entry == ".claude" || entry == ".claude/":
			// The checkout's LOCAL exclude file, shared by every worktree of
			// it -- never a committed .gitignore (the header's `config`).
			exclude := common + "/info/exclude"
			if err := cuExclude(exclude, entry); err != nil {
				fmt.Fprintf(stderr, "classify-untracked: cannot add %s to %s: %v\n", entry, exclude, err)
				return 2
			}
			fmt.Fprintf(stdout, "IGNORED: %s (local exclude)\n", entry)
		case cuCapture(entry):
			if err := os.MkdirAll(scratch, 0o777); err != nil {
				fmt.Fprintf(stderr, "classify-untracked: cannot create %s: %v\n", scratch, err)
				return 2
			}
			// A same-named file already in the scratchpad is never clobbered:
			// the incoming copy gains a timestamp prefix.
			base := path.Base(entry)
			dest := scratch + "/" + base
			if _, err := os.Stat(dest); err == nil {
				stamp := time.Now().Format("20060102-150405")
				dest = scratch + "/" + stamp + "-" + base
				// Two same-named captures in one second: count up, never clobber.
				for n := 2; isFile(dest); n++ {
					dest = fmt.Sprintf("%s/%s-%d-%s", scratch, stamp, n, base)
				}
			}
			if err := os.Rename(pcAbs(env, landing+"/"+entry), dest); err != nil {
				fmt.Fprintf(stderr, "classify-untracked: cannot move %s to %s: %v\n", entry, dest, err)
				return 2
			}
			fmt.Fprintf(stdout, "CAPTURED: %s -> %s\n", entry, dest)
		default:
			fmt.Fprintf(stdout, "ASSET: %s (operator decides — commit deliberately or delete)\n", entry)
		}
	}
	return 0
}

// cuCapture is the bash's `*.png | *.PNG | *.jpg | *.JPG | *.jpeg | *.JPEG`.
func cuCapture(entry string) bool {
	for _, ext := range []string{".png", ".PNG", ".jpg", ".JPG", ".jpeg", ".JPEG"} {
		if strings.HasSuffix(entry, ext) {
			return true
		}
	}
	return false
}

// cuExclude appends entry to exclude unless a line already equals it
// (`grep -qxF entry || printf '%s\n' entry >>`), creating the file and its
// directory as `mkdir -p` and `>>` did.
func cuExclude(exclude, entry string) error {
	if err := os.MkdirAll(filepath.Dir(exclude), 0o777); err != nil {
		return err
	}
	if body, err := os.ReadFile(exclude); err == nil {
		for _, l := range strings.Split(strings.TrimSuffix(string(body), "\n"), "\n") {
			if l == entry {
				return nil
			}
		}
	}
	f, err := os.OpenFile(exclude, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, entry+"\n")
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
