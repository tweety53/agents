package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The helpers run 2's worktree cleanup (remove-change-worktrees), the
// workspace guards and prepare-workspace share: the bound on a
// project-supplied command, the workspace id, and the change's worktrees.

const (
	// ccDefaultTimeout is the bound on the project's `## stop` command under
	// **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`).
	//
	// ccDefaultGrace is the grace between the SIGTERM and the SIGKILL
	// ccRunSurvivors sends, spent AFTER the bound has fired: two seconds is
	// enough for a client that closes a socket on SIGTERM.
	ccDefaultTimeout = 60 * time.Second
	ccDefaultGrace   = 2 * time.Second
)

// ccHeading is the `## workspace isolation` heading rule, written once.
var (
	ccUnknownToken = regexp.MustCompile(`<[A-Za-z0-9_:.-]+>`)
	ccHeading      = regexp.MustCompile(`^##[ \t\n\v\f\r]+workspace isolation[ \t\n\v\f\r]*$`)
	ccAnyHeading   = regexp.MustCompile(`^#+[ \t\n\v\f\r]`)
)

// ccWorkspaceID is the derivation skills/flow-contracts/workspace-isolation.md
// carries as a runnable block, which is canonical: ASCII-lowercase, every
// byte outside [a-z0-9] to `-`, drop trailing segments while longer than 12
// and a `-` remains, cut to 12, strip every trailing `-`, then `-` and the
// first four hex digits of the SHA-256 of the original name.
//
// Derived from the change name exactly as workspace-isolation.md derives it,
// and from nothing else. This is a second implementation of that block; the
// bash copy DID drift, in the three ways the three paragraphs below each
// name.
//
// THE NORMALISATION RUNS FIRST, and the order is the whole of step 2 and step
// 3 being separate steps. From there on the string is pure ASCII, so its
// length and its first 12 bytes count the same thing in every locale — a
// C-locale shell counts bytes and a UTF-8 one counts characters, and on an
// ASCII string those are one number. Normalising LAST, as "truncate then
// clean up" would have it, leaves that arithmetic running over the raw name
// where the two disagree, and makes `.` and `_` opaque characters inside a
// segment rather than the segment boundaries they are: `KAN-99-Fix.Thing`
// keeps `kan-99-fix` under this order and `kan-99` under the other.
//
// THE CASE FOLD IS A-Z ONLY (ASCIILower), which is the whole of that step's
// locale independence. The bash needed `LC_ALL=C` on both `tr` calls, since
// without it `tr` consults the ambient locale's case table and does — the
// canonical file records `İstanbul-test` yielding two different prefixes from
// one script on one machine — and explicit `A-Z`/`a-z` ranges rather than
// `[:upper:]`/`[:lower:]`, a character class being exactly the construct that
// starts meaning something else in another locale. The digest is still taken
// over the ORIGINAL name's bytes, with no trailing newline, so normalizing
// can never merge two distinct names into one id.
//
// EVERY trailing `-` is removed, not one. Stripping a single one is
// indistinguishable from this on every name whose normalised prefix ends in
// exactly one separator — and differs on the first one that ends in two,
// `Trailing.__` deriving `trailing---…` instead of `trailing-…`. The digest
// comes from the repository's one sha256 helper (sha256.go).
func ccWorkspaceID(name string) string {
	b := []byte(core.ASCIILower(name))
	for i, c := range b {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			b[i] = '-'
		}
	}
	prefix := string(b)
	for len(prefix) > 12 && strings.Contains(prefix, "-") {
		prefix = prefix[:strings.LastIndex(prefix, "-")]
	}
	prefix = strings.TrimRight(prefix[:min(12, len(prefix))], "-")
	return prefix + "-" + core.Sha256Hex(name)[:4]
}

// ccRunSurvivors runs a project-supplied command: `bash -o pipefail -c <cmd>`
// from dir, stdin /dev/null, stdout into an unlinked-after-read scratch file
// under TMPDIR, stderr inherited, in a process group of its own. On the bound,
// SIGTERM to the group, the grace, then SIGKILL to the group. started is false
// only when the scratch file could not be created.
//
// NO `timeout(1)`: it is absent from stock macOS. THE KILL GOES TO THE PROCESS
// GROUP, not to the one pid, since `bash -c` forks for a pipeline; it reaches
// nothing that left the group (a `docker exec` runs in another PID namespace),
// but this function still returns, because Wait names the direct child alone
// and WaitDelay stops it waiting on a stderr pipe an escaped descendant holds.
// `-o pipefail` makes a failing earlier stage of a pipeline visible.
func ccRunSurvivors(env core.Env, repo, cmdText string, timeout, grace time.Duration, stderr io.Writer) (out []byte, rc int, timedOut, started bool) {
	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	// The scratch file the command's stdout is captured in, removed on every
	// return. It lives under TMPDIR and NEVER inside the worktree. An in-memory capture
	// through a pipe cannot be used here, though — reading it to EOF blocks
	// until every holder of the write end exits, which is precisely the wait
	// that has to be bounded.
	scratch, err := os.CreateTemp(tmpdir, "flow-survivors.*")
	if err != nil {
		return nil, 0, false, false
	}
	defer os.Remove(scratch.Name())
	defer scratch.Close()

	bash := "bash"
	if p, ok := core.LookPath(env, "bash"); ok {
		bash = p
	}
	cmd := exec.Command(bash, "-o", "pipefail", "-c", cmdText)
	cmd.Args[0] = "bash" // $0 as the bash guard's `bash -c` gave it, in the command's own messages
	cmd.Dir = repo
	// stdin is /dev/null (a nil cmd.Stdin) because this guard is
	// non-interactive: a client that prompts for a password fails immediately
	// instead of waiting on a terminal nobody is watching. That is a NARROWER
	// guarantee than the bound and not a substitute for it. stderr is
	// deliberately NOT swallowed: the command's own message is the only thing
	// that says why it failed.
	cmd.Stdout, cmd.Stderr = scratch, stderr
	// A non-file stderr is a pipe Go drains; a descendant that escaped the
	// group still holds it, so Wait stops waiting for it after the grace.
	cmd.WaitDelay = grace
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		// The bash subshell's exec failing: exit 127, its message on stderr.
		fmt.Fprintf(stderr, "flow-guard: %v\n", err)
		return nil, 127, false, true
	}
	pid := cmd.Process.Pid
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()

	expire := env.SurvivorsExpire
	if expire == nil {
		bound := time.NewTimer(timeout)
		defer bound.Stop()
		expire = bound.C
	}
	select {
	case <-done:
	case <-expire:
		timedOut = true
		if syscall.Kill(-pid, syscall.SIGTERM) != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		select {
		case <-done:
		case <-time.After(grace):
		}
		if syscall.Kill(-pid, syscall.SIGKILL) != nil {
			_ = cmd.Process.Kill()
		}
		<-done
	}
	out, _ = os.ReadFile(scratch.Name())
	return out, core.WaitExitCode(cmd.ProcessState), timedOut, true
}

// changeWorktrees reads `git worktree list --porcelain` output: the
// worktrees checked out on refs/heads/spectre/<name>, and the detached or
// flow-copy/-branched mutation copies of that change's apply worktree (ccMutationCopy), each
// in git's listed order. remove-change-worktrees removes them.
//
// The path is the whole rest of the `worktree ` line, never a
// whitespace-split field: `worktree list --porcelain` emits the path raw, so
// a worktree under a TMPDIR or a home directory containing a space is
// truncated at the first one by a field reference, and the operator is then
// sent to a path that does not exist. The branch is a ref name and cannot
// contain a space, so its first field is right for it — and the comparison
// is for EQUALITY, so a neighbouring change's spectre/<name>-something is
// not taken for this change's worktree.
func changeWorktrees(porcelain []byte, name string) (wts, copies []string) {
	wt := ""
	for _, l := range strings.Split(string(porcelain), "\n") {
		if p, ok := strings.CutPrefix(l, "worktree "); ok {
			wt = p
		} else if b, ok := strings.CutPrefix(l, "branch "); ok {
			if f := strings.Fields(b); len(f) > 0 && f[0] == "refs/heads/spectre/"+name {
				wts = append(wts, wt)
			} else if len(f) > 0 && f[0] == "refs/heads/flow-copy/"+wt[strings.LastIndex(wt, "/")+1:] && ccMutationCopy(wt, name) {
				copies = append(copies, wt)
			}
		} else if l == "detached" && ccMutationCopy(wt, name) {
			copies = append(copies, wt)
		}
	}
	return wts, copies
}

// ccMutationCopy reports whether path is a throwaway copy of change name's
// apply worktree: review-panel-optional-slots.md's
// `<worktree>-mutation-<round>` (kan-1037), matched on the path's last
// element as `<name>-mutation-` followed by digits and nothing else, or
// implement.md's parallel-group copy `<worktree>-group-<g>` the same way
// (flow-guard parallel-groups). The caller only
// asks for an entry that is DETACHED or on `refs/heads/flow-copy/<its path's
// last element>` -- a copy is created on that branch (kan-1090 self-review,
// F14), one made before that detached -- so a change actually named
// `<name>-mutation-3`, whose apply worktree carries a spectre/ branch, is
// never read as this change's copy. The path is read from git's
// own worktree list, never guessed from a layout.
func ccMutationCopy(path, name string) bool {
	base := path[strings.LastIndex(path, "/")+1:]
	g, ok := strings.CutPrefix(base, name+"-mutation-")
	if !ok {
		g, ok = strings.CutPrefix(base, name+"-group-")
	}
	return ok && g != "" && strings.Trim(g, "0123456789") == ""
}
