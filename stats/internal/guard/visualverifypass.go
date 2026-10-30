package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// visualVerifyPass is scripts/visual-verify-pass.sh: that script's header is
// the contract for every argument, step and closing line of its subcommands.
// It runs the parent's side of `flow.visual-verify` (decision
// thin-parent-prepare-close) — the existing guards in process, the `flow`
// CLI from PATH — and closes on one `NEXT:` (exit 0) or `STOP:` (exit 1)
// line; exit 2 is a usage error or an unreadable input.
func init() { Registry["visual-verify-pass"] = visualVerifyPass }

const vvpUsage = "usage: visual-verify-pass.sh prepare <canonical-worktree> <name> <key> --harness <h> --session-token mf-<t> --merge-base <worktree>=<sha>... [--worktree <worktree>]... [--since <worktree>=<sha>]... [--round-base <worktree>=<sha>]... [--app <worktree>:<app-root>=<url>]... [--miss] [--continue [--started]]\n" +
	"       visual-verify-pass.sh close <canonical-worktree> <name> <key> <motions named> --session-token mf-<t>"

// vvpExit unwinds a step to visualVerifyPass with its exit code.
type vvpExit int

// vvp is one call's state: its arguments and whether it began the stage.
type vvp struct {
	env            Env
	stdout, stderr io.Writer

	canon, name, key, harness, token string
	motions                          int
	worktrees                        []string          // canonical first, then each --worktree
	mergeBase, since, roundBase      map[string]string // by worktree
	apps                             map[string][]string
	miss, cont, started              bool
	siblings                         map[string][]string // a worktree's siblings whose diff its `ui paths` matched
	setSize                          int

	stageOpen bool
}

func visualVerifyPass(args []string, env Env, stdout, stderr io.Writer) (code int) {
	v := &vvp{env: env, stdout: stdout, stderr: stderr,
		mergeBase: map[string]string{}, since: map[string]string{}, roundBase: map[string]string{}, apps: map[string][]string{}}
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(vvpExit)
			if !ok {
				panic(r)
			}
			code = int(e)
		}
	}()
	if len(args) == 0 {
		v.usage("no subcommand")
	}
	switch args[0] {
	case "prepare":
		v.parse(args[1:], true)
		v.prepare()
	case "close":
		v.parse(args[1:], false)
		v.close()
	default:
		v.usage("unknown subcommand '" + args[0] + "'")
	}
	return 0
}

func (v *vvp) usage(why string) {
	fmt.Fprintf(v.stderr, "visual-verify-pass: %s\n%s\n", why, vvpUsage)
	panic(vvpExit(2))
}

func (v *vvp) die(format string, a ...any) {
	fmt.Fprintf(v.stderr, "visual-verify-pass: "+format+"\n", a...)
	panic(vvpExit(2))
}

func (v *vvp) say(format string, a ...any) { fmt.Fprintf(v.stdout, format+"\n", a...) }

// next prints the closing `NEXT:` line and exits 0.
func (v *vvp) next(line string) {
	v.say("NEXT: %s", line)
	panic(vvpExit(0))
}

// stop ends an open stage with outcome, prints the closing `STOP:` line and
// exits 1.
func (v *vvp) stop(outcome, line string) {
	if v.stageOpen {
		v.markEnd(outcome)
	}
	v.say("STOP: %s", line)
	panic(vvpExit(1))
}

// abs resolves p against env.Dir, as every guard here does.
func (v *vvp) abs(p string) string { return pcAbs(v.env, p) }

func (v *vvp) parse(args []string, prepare bool) {
	pos := 3
	if !prepare {
		pos = 4
	}
	if len(args) < pos {
		v.usage("missing arguments")
	}
	v.canon, v.name, v.key = v.abs(args[0]), args[1], args[2]
	if !prepare {
		n, err := fmt.Sscan(args[3], &v.motions)
		if n != 1 || err != nil || v.motions < 0 {
			v.usage("<motions named> '" + args[3] + "' is not a count")
		}
	}
	if !plainChangeName(v.name) || v.key == "" {
		v.usage("change name '" + v.name + "' or key '" + v.key + "' is not plain")
	}
	v.worktrees = []string{v.canon}
	rest := args[pos:]
	val := func(i int) string {
		if i+1 >= len(rest) {
			v.usage(rest[i] + " takes a value")
		}
		return rest[i+1]
	}
	pair := func(flag, s string) (string, string) {
		w, sha, ok := strings.Cut(s, "=")
		if !ok || w == "" || sha == "" {
			v.usage(flag + " '" + s + "' is not <worktree>=<sha>")
		}
		return v.abs(w), sha
	}
	for i := 0; i < len(rest); i++ {
		switch f := rest[i]; {
		case f == "--session-token":
			v.token = val(i)
			i++
		case prepare && f == "--harness":
			v.harness = val(i)
			i++
		case prepare && f == "--worktree":
			v.worktrees = append(v.worktrees, v.abs(val(i)))
			i++
		case prepare && (f == "--merge-base" || f == "--since" || f == "--round-base"):
			w, sha := pair(f, val(i))
			map[string]map[string]string{"--merge-base": v.mergeBase, "--since": v.since, "--round-base": v.roundBase}[f][w] = sha
			i++
		case prepare && f == "--app":
			a := val(i)
			w, app, ok := strings.Cut(a, ":")
			root, url, ok2 := strings.Cut(app, "=")
			if !ok || !ok2 || w == "" || root == "" || url == "" {
				v.usage("--app '" + a + "' is not <worktree>:<app-root>=<url>")
			}
			v.apps[v.abs(w)] = append(v.apps[v.abs(w)], root+"="+url)
			i++
		case prepare && f == "--miss":
			v.miss = true
		case prepare && f == "--continue":
			v.cont = true
		case prepare && f == "--started":
			v.started = true
		default:
			v.usage("unknown argument '" + f + "'")
		}
	}
	if !strings.HasPrefix(v.token, "mf-") || len(v.token) == 3 {
		v.usage("--session-token mf-<literal-token> is required")
	}
	if prepare && v.harness == "" {
		v.usage("--harness is required")
	}
	if v.started && !v.cont {
		v.usage("--started is given with --continue only")
	}
	if prepare && !v.cont {
		for _, w := range v.worktrees {
			if v.mergeBase[w] == "" {
				v.usage("no --merge-base for " + w)
			}
		}
	}
	for _, w := range v.worktrees {
		if !isDir(w) {
			v.die("%s is not a directory", w)
		}
	}
}

// flow runs one `flow` call from PATH, its combined output and exit code.
func (v *vvp) flow(args ...string) (string, int) {
	out, _, rc := pcFlow(v.env, true, args...)
	return strings.TrimSpace(string(out)), rc
}

func (v *vvp) markBegin() {
	if out, rc := v.flow("stage", "begin", "-C", v.canon, "-command", "/flow", "-stage", "flow.visual-verify",
		"-harness", v.harness, "-session-token", v.token, v.name); rc != 0 {
		fmt.Fprintf(v.stderr, "visual-verify-pass: warning: flow stage begin failed: %s\n", out)
	}
	v.stageOpen = true
}

func (v *vvp) markEnd(outcome string) {
	if out, rc := v.flow("stage", "end", "-C", v.canon, "-command", "/flow", "-stage", "flow.visual-verify",
		"-outcome", outcome, v.name); rc != 0 {
		fmt.Fprintf(v.stderr, "visual-verify-pass: warning: flow stage end failed: %s\n", out)
	}
	v.stageOpen = false
}

// guard runs a registered guard in process: its exit, stdout and stderr.
func (v *vvp) guard(name string, stdin io.Reader, args ...string) (int, string, string) {
	env := v.env
	env.Stdin = stdin
	var out, errb bytes.Buffer
	rc := Registry[name](args, env, &out, &errb)
	return rc, strings.TrimSpace(out.String()), strings.TrimSpace(errb.String())
}

func (v *vvp) git(args ...string) (string, bool) { return capture(envGit(v.env)(args...)) }

// repoName is a worktree's repository name: its main checkout's basename,
// the name `check-visual-trigger.sh --repo` takes.
func (v *vvp) repoName(w string) string {
	if common, ok := v.git("-C", w, "rev-parse", "--path-format=absolute", "--git-common-dir"); ok {
		return filepath.Base(filepath.Dir(common))
	}
	return filepath.Base(w)
}

// The stage file, <canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt,
// is the only state a later call reads (decision stage-file-names-its-rows):
//
//	token <session token>
//	begin <RFC 3339 instant>
//	set <worktrees in the resolved set>
//	worktree <path> <repository name> <merge base> <round base or ->[ <sibling>,…]
//	row <key> <worktree>
//	verdict <key> <guard verdict line>
//	blocked <key>
//	closed <key> <outcome>
func (v *vvp) stagePath() string {
	return filepath.Join(v.canon, ".superpowers", "sdd", "visual-verify-stage.txt")
}

func (v *vvp) writeStage(survivors []string) {
	var b strings.Builder
	fmt.Fprintf(&b, "token %s\nbegin %s\nset %d\n", v.token, time.Now().UTC().Format(time.RFC3339), len(v.worktrees))
	for _, w := range survivors {
		rb := v.roundBase[w]
		if rb == "" {
			rb = "-"
		}
		fmt.Fprintf(&b, "worktree %s %s %s %s", w, v.repoName(w), v.mergeBase[w], rb)
		if sib := v.siblings[w]; len(sib) > 0 {
			fmt.Fprintf(&b, " %s", strings.Join(sib, ","))
		}
		b.WriteString("\n")
	}
	if err := os.MkdirAll(filepath.Dir(v.stagePath()), 0o755); err != nil {
		v.die("cannot write the stage file: %v", err)
	}
	if err := os.WriteFile(v.stagePath(), []byte(b.String()), 0o644); err != nil {
		v.die("cannot write the stage file: %v", err)
	}
}

func (v *vvp) appendStage(line string) {
	f, err := os.OpenFile(v.stagePath(), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		v.die("cannot append to the stage file: %v", err)
	}
	defer f.Close()
	fmt.Fprintln(f, line)
}

// readStage is the stage file's lines, split into their leading word and
// the rest; exit 2 when it cannot be read.
func (v *vvp) readStage() [][2]string {
	b, err := os.ReadFile(v.stagePath())
	if err != nil {
		v.die("cannot read the stage file %s — run prepare first: %v", v.stagePath(), err)
	}
	var out [][2]string
	for _, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		k, rest, _ := strings.Cut(l, " ")
		out = append(out, [2]string{k, rest})
	}
	return out
}

// cells is a worktree's `## visual verification` two-cell rows.
func (v *vvp) cells(w string) map[string]string {
	b, _ := os.ReadFile(filepath.Join(w, ".flow", "project.md"))
	return vvCells(string(b))
}

// cellPath is a cell's path value, backticks dropped, resolved against w.
func cellPath(w, cell string) string {
	cell = strings.Trim(cell, "`")
	if cell == "" || filepath.IsAbs(cell) {
		return cell
	}
	return filepath.Join(w, cell)
}
