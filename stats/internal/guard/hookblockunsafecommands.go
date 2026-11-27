package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/shlex"
)

// flow-guard hook block-unsafe-commands — PreToolUse: keep agents from writing into a repository's
// main checkout while it sits on the default branch.
//
// Every pipeline run works in a worktree. The main checkout on main/master/develop is the landing
// target and a place to read from, never a workbench. A session that edits, stages, commits,
// stashes or resets there leaves residue no later session owns. This hook denies the tools that
// create such residue:
//
//   - Edit / Write / MultiEdit / NotebookEdit on a path inside such a checkout;
//   - Bash commands that run a mutating git verb there (`git commit`, `git -C <main> add`,
//     `cd <main> && git reset ...`), or write into it with `sed -i`, `tee`, a `>`/`>>`
//     redirect, `cp` destinations, `mv`/`rm` sources and destinations, or `trash` of a path git
//     tracks there; removing a whole clone of a local working checkout is not a write to a main
//     checkout. A `cd` inside `( … )` or `$( … )` ends at its close, as bash scopes it. A heredoc
//     body is scanned like any other line, except where the whole command is one plain data
//     write — `cat > ABS <<EOF`, `cat >> ABS <<EOF` or `tee [-a] ABS <<EOF` to a literal absolute
//     path, its body and closing line, nothing after — where only ABS is judged — and except the
//     body of any `cat`/`tee` heredoc inside a longer command, which is that writer's data, never
//     commands. A `flow` heredoc body (`flow record summary -file - <<'EOF'`) is data when its
//     opener writes no file and nothing else in the command runs code. A body no shell runs — a
//     `python`/`python3`/`node` program on stdin, or a file the rest of the command runs only
//     through one of those or `npx playwright test` — is that language's text: only the git
//     invocations its lines name are judged.
//
// "Main checkout" is the worktree whose git dir IS the common dir — a linked worktree's
// `--git-dir` is `<common>/worktrees/<name>`, so `<repo>-worktrees/<change>` beside the main
// checkout passes, as does a pre-migration `.worktrees/<change>` inside it. A path into a worktree
// that does not exist yet passes too, when its nearest existing ancestor is either root. The deny
// reason suggests the sibling `<dirname top>/<basename top>-worktrees/<change>` only. "Default
// branch" is what `origin/HEAD` points at; when the remote head is unknown, a checkout of main,
// master or develop counts. A main checkout on a feature branch is not protected — leaving the
// default branch is how work gets a branch of its own.
//
// It also denies, in a linked worktree, a `git checkout … -- <path>` or `git restore <path>`
// naming a path with unstaged edits: the restore reverts it to its index state and an uncommitted
// fix dies with it. The reason points at flow-guard break-and-prove, which restores from a byte
// snapshot. A worktree listed in its change's state record (`flow state get`) is exempt from that
// denial.
//
// It denies `git stash pop`, `apply`, `drop` and `branch` unless the stash entry they act on was
// made on the current branch. A stash list is shared by every worktree of a repository, so a bare
// `git stash pop` after a failed `git stash push` pops another change's work into this worktree. A
// call with no explicit `stash@{N}` is denied outright; one naming an entry is allowed only when
// `git stash list` shows "On <current branch>:" for it. The reason tells the agent to stop and
// report BLOCKED.
//
// It denies a bare `cat > <file>` with no input of its own, and a `go test` over a whole heavy
// package of the agents repo with no `-run` (bucBareCatReason, bucFullGoTestReason).
//
// Allowed everywhere: reads, `git pull --ff-only`, `git fetch`, `git checkout -b`,
// `git switch -c`, `git worktree add`, `git push`, and the landing script `flow-guard
// refresh-main-checkout` — it is not a `git` verb in the command text, and owns its own
// main-checkout rules.
//
// Denies, never rewrites a write: the deny reason names the worktree command to use instead. The
// one rewrite is a call it lets through: a subagent's `flow`/`flow-guard` call gets
// `-caller-agent <agent_id>` (hookcalleragent.go).
//
// Fails open: any unexpected input, parse error, git failure or internal fault exits 0 and lets
// the call through. A guardrail on hygiene must never be the reason work stops.
//
// Byte-identical to hooks/block-unsafe-commands.py, the Python original it replaced, on the golden
// corpus testdata/block-unsafe-commands-golden.json, save case 145's `-caller-agent` rewrite, which
// the Python hook never made: every regular expression keeps Python's
// Unicode `\s`, `\w`, `\d` and `\b` (pyposix.go), and a lookahead the RE2 syntax lacks is a check
// in code after the match. Where Python raised an exception the hook did not catch, this panics,
// and hook()'s recover makes it the same silent exit.
//
// ponytail: Bash coverage is by token scan — a heredoc piped into python, a script that edits a
// file, or a `find -exec` still writes. Widen the verb and writer lists when a new shape shows up
// in a repo's reflog, rather than trying to model the shell. Variable expansion is the same
// approximation, one step further: `NAME=value` assignments collected positionally from the
// command text only, values kept literal, within-line use before an assignment left unexpanded —
// never environment state, which a one-shot hook process cannot know. The bare-`cat` check is a
// line scan: a heredoc body line opening with `cat > <file>` is denied too.
func init() { hooks["block-unsafe-commands"] = withCallerAgent(blockUnsafeCommands) }

var bucEditTools = map[string]bool{"Edit": true, "Write": true, "MultiEdit": true, "NotebookEdit": true}

var bucAssignmentRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// bucDeniedGitVerbs are the git verbs that change the index, the worktree, a ref or the stash list.
var bucDeniedGitVerbs = bucSet("add", "am", "apply", "cherry-pick", "clean", "commit", "merge", "mv", "rebase",
	"reset", "restore", "revert", "rm", "stash", "update-index", "update-ref")

var bucFallbackDefaultBranches = bucSet("main", "master", "develop")

// bucBareCatRE is a bare `cat > <file>` with no input of its own: it reads stdin, which a Bash
// call never has, so it waits until the call times out. Python's trailing
// `\s*(?=$|[;&)]|\|\|)` (multiline) is bareCatEnds.
var bucBareCatRE = regexp.MustCompile(`(?m)(?:^|[;&(]|\|\|)[` + pyReS + `]*cat[` + pyReS + `]*>>?[` + pyReS + `]*[^` + pyReS + `<>|;&()]+`)

const bucBareCatReason = "a bare `cat > <file>` reads stdin, which a Bash call never has, so it waits until the " +
	"call times out — write the file with the Write tool, or give cat a heredoc " +
	"(`cat > <file> <<'EOF'`)"

// A `go test` over a whole heavy package of the agents repo — the guard package (~5 minutes), the
// flow command or `./...` — with no `-run`. The operator's own `!` run is not a tool call, so it is
// never judged; a session names the tests it touched with `-run`. Python's `(?=\s|$)` after the
// package is a check in fullGoTest.
var (
	bucFullGoTestRE = regexp.MustCompile(`(?:^|[` + pyReS + `])\./(?:internal/guard|cmd/flow|\.\.\.)/?`)
	bucGoTestSplit  = regexp.MustCompile(`[;&|\n]+`)
	bucRunFlagRE    = regexp.MustCompile(`(?:^|[` + pyReS + `])-run[` + pyReS + `=]`)
)

const bucFullGoTestReason = "`go test` over a whole heavy package (./internal/guard, ./cmd/flow or ./...) takes minutes — name " +
	"the tests the change touches with `-run '<TestName|OtherTest>'`. A full package run is the " +
	"operator's to ask for, and runs from their own `!` prompt."

const (
	bucLiteralAbs = `(/[^` + pyReS + `'"$` + "`" + `\\;&|<>(){}*?\[\]~!#]+)`
	bucDelim      = `(?:'([` + pyReW + `]+)'|"([` + pyReW + `]+)"|([` + pyReW + `]+))`
)

var (
	bucDataWriteRE = regexp.MustCompile(`^[ \t]*(?:cat[ \t]+>>?[ \t]*` + bucLiteralAbs + `|tee(?:[ \t]+-a)?[ \t]+` +
		bucLiteralAbs + `(?:[ \t]+>[ \t]*/dev/null)?)[ \t]+<<[ \t]*` + bucDelim + `[ \t]*\r?$`)
	bucHeredocOpenRE  = regexp.MustCompile(`<<-?[ \t]*` + bucDelim)
	bucSegmentSplitRE = regexp.MustCompile(`&&|\|\||[;|]`)
	bucCatTargetRE    = regexp.MustCompile(`>>?[ \t]*([^` + pyReS + `;&|<>]+)`)
	bucSubstRE        = regexp.MustCompile(`[<>]\(|\$\(|` + "`")
	bucCommandCutRE   = regexp.MustCompile(`&&|\|\||[\n;|&(){}]`)
	// A `flow` opener's later words must be plain tokens and none a shell or interpreter name
	// (`flow run csh <<EOF`), or its heredoc body stays scanned.
	bucFlowPlainWordRE = regexp.MustCompile(`^[A-Za-z0-9_.,=-]+$`)
	bucCheckScriptRE   = regexp.MustCompile(`^check-[` + pyReW + `.-]*\.(?:sh|py)$`)
	// A guard run through the bootstrap: `flow-guard check-…`, bare or as `…/scripts/flow-guard`.
	bucCheckGuardRE   = regexp.MustCompile(`^check-[` + pyReW + `.-]*$`)
	bucDollarParenRE  = regexp.MustCompile(`\$\(([^()]*)\)`)
	bucProgramGitRE   = regexp.MustCompile(`(?:^|[` + pyReS + `])(git[` + pyReS + `].*)$`)
	bucProgramPunctRE = regexp.MustCompile(`["'()\[\],<>|;&{}]`)
	// A `$(...)` without a nested paren, or a backtick pair: the spans of an unquoted heredoc body
	// that bash runs.
	bucSubstSpanRE = regexp.MustCompile(`\$\([^()]*\)|` + "`[^`]*`")
	// A redirect glued to its target in an unbalanced-quote fallback word: `>f`, `2>>f`, `&>f`;
	// `>&1` duplicates a descriptor and writes no file.
	bucAttachedRedirectRE = regexp.MustCompile(`^\p{Nd}*(&>>|&>|>>|>\||>)([^&].*)$`)
	// A heredoc opener's segment cut into words so `x=$(bash` yields `bash`.
	bucOpenerWordSplitRE = regexp.MustCompile(`[` + pyReS + `=$()` + "`" + `'"]+`)
	bucPunctWordRE       = regexp.MustCompile(`^[();&|<>]+$`)
	bucVarRE             = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
	// Python's `\bgit((?:\s+-[Cc]\s+\S+)*)\s+stash\s+(pop|apply|drop|branch)\b([^;&|\n)]*)`, matched
	// at each `git` behind a word boundary; the `\b` after the verb is a check in stashDenial.
	bucStashRE = regexp.MustCompile(`^git((?:[` + pyReS + `]+-[Cc][` + pyReS + `]+[^` + pyReS + `]+)*)[` + pyReS +
		`]+stash[` + pyReS + `]+(pop|apply|drop|branch)([^;&|\n)]*)`)
	bucStashRefRE  = regexp.MustCompile(`^stash@\{(\p{Nd}+)\}$`)
	bucDigitsRE    = regexp.MustCompile(`^\p{Nd}+$`)
	bucStashListRE = regexp.MustCompile(`^stash@\{(\p{Nd}+)\}: (?:WIP on|On) ([^:]+):`)
)

var bucShellKeywords = bucSet("do", "then", "else", "elif", "if", "while", "until", "!")

var bucFlowShellWords = bucSet("sh", "bash", "zsh", "dash", "ksh", "csh", "tcsh", "mksh", "yash", "fish", "ash",
	"busybox", "su", "ssh", "doas", "sudo", "env", "script", "parallel", "xargs", "python", "python3", "perl",
	"ruby", "node")

var bucRunnerWords = bucSet("sh", "bash", "zsh", "dash", "ksh", "fish", "source", ".", "eval", "exec", "xargs",
	"env", "command", "builtin", "nohup", "timeout", "nice", "time", "sudo", "python", "python3", "perl", "ruby",
	"node", "osascript", "awk", "gawk", "mawk", "nawk", "make", "gmake", "watch", "crontab", "at", "batch")

// bucReaderWords are commands known not to execute a file they are given: the heredoc target
// named only as an argument of one of these is data. Anything else naming it may run it, so the
// body stays scanned.
var bucReaderWords = bucSet("cat", "tee", "head", "tail", "wc", "grep", "diff", "cmp", "ls", "test", "[", "stat",
	"cd", "echo", "printf")

// bucLeadingKeywords are shell keywords that open a command without being one: `do x`'s command is `x`.
var bucLeadingKeywords = bucSet("do", "then", "else", "elif", "!")

var bucGitReadSubcommands = bucSet("diff", "add", "show", "log", "status")

// bucProgramRunners are interpreters whose program text is not shell: a body they run is scanned
// for git verbs only.
var bucProgramRunners = bucSet("python", "python3", "node")

// bucRedirectOps are the operators that write a file: shlex yields each as a word of its own, so
// a word that merely starts with `>` came out of quotes and is data (`'>Save<'`).
var bucRedirectOps = bucSet(">", ">>", ">|", "&>", "&>>")

var bucSeparators = bucSet("&&", "||", ";", "|")

const bucStashBlocked = " Never use git stash pop/apply/drop/branch: the stash list is shared by every worktree of the " +
	"repository, so it can hold another change's work. If a worktree looks wrong, stop and report " +
	"BLOCKED with the exact `git status` output."

const bucHeredocNote = " A heredoc body the hook could not read as data is scanned as shell — an " +
	"`export`, a `NAME=value` prefix or a code-running word elsewhere in the command " +
	"keeps it scanned. Write the file with the Write tool, or run the heredoc as a " +
	"command of its own."

// bucBadCwd stands for a payload `cwd` that is truthy but no string: Python's hook raised on the
// first use of it, which here is a panic in resolve, git or pyRealpath (its NUL).
const bucBadCwd = "\x00cwd"

func bucSet(words ...string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}

// bucLast is Python's s.rsplit("/", 1)[-1].
func bucLast(s string) string { return pyBasename(s) }

type buc struct {
	env   Env
	cache map[string]bucGitOut
}

type bucGitOut struct {
	out string
	ok  bool
}

// bucHit is a path a Bash command writes; link marks an `rm` or `mv` source, which acts on a
// symlink itself, not on its target, so a source that is a symlink is judged by the directory
// holding it.
type bucHit struct {
	path string
	link bool
}

func blockUnsafeCommands(env Env, payload map[string]any, stdout, _ io.Writer) {
	if payload == nil {
		return
	}
	tool, _ := payload["tool_name"].(string)
	input, ok := payload["tool_input"].(map[string]any)
	if !ok {
		return
	}
	cwd := env.Dir
	if c := payload["cwd"]; pyTruthy(c) {
		if cwd, ok = c.(string); !ok {
			cwd = bucBadCwd
		}
	}
	b := &buc{env: env, cache: map[string]bucGitOut{}}

	var candidates []bucHit
	var command string
	switch {
	case bucEditTools[tool]:
		path := input["file_path"]
		if !pyTruthy(path) {
			path = input["notebook_path"]
		}
		if p, ok := path.(string); ok && p != "" {
			candidates = append(candidates, bucHit{path: b.resolve(p, cwd)})
		}
	case tool == "Bash":
		command, _ = input["command"].(string)
		if command == "" {
			break
		}
		var landings []string
		if bareCat(command) {
			landings = append(landings, bucBareCatReason)
		}
		if fullGoTest(command) {
			landings = append(landings, bucFullGoTestReason)
		}
		if stash := b.stashDenial(command, cwd); stash != "" {
			landings = append(landings, stash)
		}
		candidates = b.bashHits(command, cwd, &landings)
		if len(landings) > 0 {
			io.WriteString(stdout, hookDenyJSON(landings[0]))
			return
		}
	default:
		return
	}

	for _, c := range candidates {
		if c.path == "" {
			continue
		}
		top, branch, ok := b.protected(c)
		if !ok {
			continue
		}
		reason := bucDenyReason(top, branch)
		if tool == "Bash" {
			reason += "\n\nThis command reads as writing " + c.path + "."
			if bucHeredocOpenRE.MatchString(command) {
				reason += bucHeredocNote
			}
		}
		io.WriteString(stdout, hookDenyJSON(reason))
		return
	}
}

// bareCat is BARE_CAT_RE.search: a match whose trailing whitespace reaches a line end, the end of
// the command, `;`, `&`, `)` or `||`.
func bareCat(command string) bool {
	for _, m := range bucBareCatRE.FindAllStringIndex(command, -1) {
		rest := command[m[1]:]
		ws := strings.TrimLeftFunc(rest, pyIsSpace)
		if strings.Contains(rest[:len(rest)-len(ws)], "\n") || ws == "" ||
			strings.ContainsAny(ws[:1], ";&)") || strings.HasPrefix(ws, "||") {
			return true
		}
	}
	return false
}

func fullGoTest(command string) bool {
	for _, segment := range bucGoTestSplit.Split(command, -1) {
		if bucGoTestWord(segment) && bucWholeHeavyPackage(segment) && !bucRunFlagRE.MatchString(segment) {
			return true
		}
	}
	return false
}

// bucGoTestWord is Python's re.search(r"\bgo\s+test\b", segment).
func bucGoTestWord(s string) bool {
	for i := 0; ; i++ {
		k := strings.Index(s[i:], "go")
		if k < 0 {
			return false
		}
		i += k
		if r, _ := utf8.DecodeLastRuneInString(s[:i]); i > 0 && pyIsWord(r) {
			continue
		}
		rest := s[i+2:]
		ws := strings.TrimLeftFunc(rest, pyIsSpace)
		if len(ws) == len(rest) || !strings.HasPrefix(ws, "test") {
			continue
		}
		if r, _ := utf8.DecodeRuneInString(ws[4:]); len(ws) == 4 || !pyIsWord(r) {
			return true
		}
	}
}

// bucWholeHeavyPackage is FULL_GO_TEST_RE.search, its `(?=\s|$)` checked after the match.
func bucWholeHeavyPackage(s string) bool {
	for _, m := range bucFullGoTestRE.FindAllStringIndex(s, -1) {
		if r, _ := utf8.DecodeRuneInString(s[m[1]:]); m[1] == len(s) || pyIsSpace(r) {
			return true
		}
	}
	return false
}

// git is the stripped stdout of a git call in cwd, ok false on any failure. A cwd of None (""),
// a NUL in an argument or stdout that is not UTF-8 panics: Python's subprocess raised there.
func (b *buc) git(cwd string, args ...string) (string, bool) {
	key := fmt.Sprintf("%q %q", cwd, args)
	if r, ok := b.cache[key]; ok {
		return r.out, r.ok
	}
	if cwd == "" || strings.ContainsRune(cwd, 0) || slices.ContainsFunc(args, func(a string) bool { return strings.ContainsRune(a, 0) }) {
		panic("git: no directory, or a NUL in an argument")
	}
	out, rc, err := hookRun(b.env, hookBound(b.env, 5*time.Second), b.env.Dir, "git", append([]string{"-C", cwd}, args...)...)
	var r bucGitOut
	if err == nil {
		if !utf8.Valid(out) {
			panic("git: stdout is not UTF-8")
		}
		if rc == 0 {
			r = bucGitOut{out: pyStrip(bucNewlines(string(out))), ok: true}
		}
	}
	b.cache[key] = r
	return r.out, r.ok
}

// bucNewlines is Python's text-mode newline translation: `\r\n` and `\r` read as `\n`.
func bucNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// existingDir is the nearest existing directory at or above path (a Write may create the file),
// "" when none.
func (b *buc) existingDir(path string) string {
	p := pyAbspath(path, b.env.Dir)
	for !pyIsDir(p) {
		parent := pyDirname(p)
		if parent == p {
			return ""
		}
		p = parent
	}
	return p
}

// isMainCheckout is true when p is the toplevel of a main checkout — not a subdirectory, not a
// linked worktree.
func (b *buc) isMainCheckout(p string) bool {
	top, _ := b.git(p, "rev-parse", "--show-toplevel")
	gitDir, _ := b.git(p, "rev-parse", "--absolute-git-dir")
	common, _ := b.git(p, "rev-parse", "--git-common-dir")
	if top == "" || gitDir == "" || common == "" {
		return false
	}
	if !strings.HasPrefix(common, "/") {
		common = pyJoin(p, common)
	}
	return b.realpath(top) == b.realpath(p) && b.realpath(gitDir) == b.realpath(common)
}

func (b *buc) realpath(p string) string { return pyRealpath(p, b.env.Dir) }

// localClone is true when p is the toplevel of a clone whose origin is a local non-bare checkout —
// a scratch copy, whose whole removal writes no main checkout.
func (b *buc) localClone(p string) bool {
	if !pyIsDir(p) || !b.isMainCheckout(p) {
		return false
	}
	origin, _ := b.git(p, "remote", "get-url", "origin")
	origin = strings.TrimPrefix(origin, "file://")
	if !strings.HasPrefix(origin, "/") || !pyIsDir(origin) {
		return false
	}
	bare, ok := b.git(origin, "rev-parse", "--is-bare-repository")
	return ok && bare == "false"
}

// worktreeRoot is true when d is a worktree root: `<main>-worktrees` beside a main checkout
// `<main>`, or the pre-migration `<main>/.worktrees` the hook no longer suggests.
func (b *buc) worktreeRoot(d string) bool {
	name := pyBasename(d)
	if name == ".worktrees" {
		return true
	}
	main, ok := strings.CutSuffix(name, "-worktrees")
	return ok && main != "" && b.isMainCheckout(pyJoin(pyDirname(d), main))
}

// protected is the (toplevel, branch) when the hit lies in a main checkout on its default branch.
// The path is judged after its symlinks resolve, as the write lands: `abspath` would collapse
// `<link>/..` lexically, and a link into a checkout's subdirectory would hide the checkout.
func (b *buc) protected(h bucHit) (string, string, bool) {
	path := h.path
	if h.link && pyIsLink(path) {
		path = b.realpath(pyDirname(path)) // the link's own directory is what changes
	} else {
		path = b.realpath(path)
	}
	d := b.existingDir(path)
	if d == "" {
		return "", "", false
	}
	if b.worktreeRoot(d) && pyAbspath(path, b.env.Dir) != d {
		// the nearest existing ancestor is a worktree root, so a missing component below it is a
		// future worktree — never main-checkout content. The root itself stays protected.
		return "", "", false
	}
	top, _ := b.git(d, "rev-parse", "--show-toplevel")
	if top == "" {
		return "", "", false
	}
	gitDir, _ := b.git(d, "rev-parse", "--absolute-git-dir")
	common, _ := b.git(d, "rev-parse", "--git-common-dir")
	if gitDir == "" || common == "" {
		return "", "", false
	}
	if !strings.HasPrefix(common, "/") {
		common = pyJoin(d, common)
	}
	if b.realpath(gitDir) != b.realpath(common) {
		return "", "", false // a linked worktree
	}
	branch, _ := b.git(d, "branch", "--show-current")
	if branch == "" {
		return "", "", false // detached
	}
	def, ok := b.defaultBranch(d)
	if !ok && bucFallbackDefaultBranches[branch] {
		def, ok = branch, true
	}
	if !ok || branch != def {
		return "", "", false
	}
	return top, branch, true
}

// resolve is path made absolute against cwd, or "" (None) when the shell would expand it (`$VAR`,
// a backtick) or cwd itself is unknown — a path the hook cannot know is let through, never
// guessed at.
func (b *buc) resolve(path, cwd string) string {
	if strings.ContainsAny(path, "$`") {
		return ""
	}
	path = pyExpandUser(b.env, path)
	if strings.HasPrefix(path, "/") {
		return path
	}
	if cwd == bucBadCwd {
		panic("resolve: cwd is not a string")
	}
	if cwd == "" {
		return ""
	}
	return pyJoin(cwd, path)
}

// tokenize is the shell words with `;`, `&&`, `|` and redirects split out even where they touch a
// word — `cd <worktree>; git commit` must not make `<worktree>;` the directory.
func tokenize(command string) []string {
	split, err := shlex.Split(command)
	if err != nil {
		// Unbalanced quotes: raw words, a redirect glued to its target split off it, so `>f` and
		// `2>>f` read as shlex would have split them; a quoted word keeps its quote and never
		// matches.
		var words []string
		for _, w := range pySplit(command) {
			words = append(words, bucSplitRedirect(w)...)
		}
		return words
	}
	// A run of punctuation is one shlex token: split each paren off it, so `)>` and `;)` still
	// close their group.
	var words []string
	for _, w := range split {
		if !bucPunctWordRE.MatchString(w) {
			if w != "" {
				words = append(words, w)
			}
			continue
		}
		start := 0
		for i := 0; i < len(w); i++ {
			if w[i] == '(' || w[i] == ')' {
				if i > start {
					words = append(words, w[start:i])
				}
				words = append(words, w[i:i+1])
				start = i + 1
			}
		}
		if start < len(w) {
			words = append(words, w[start:])
		}
	}
	// A backtick span, and a `$(...)` inside a quoted word, still run: each one's command is
	// scanned again as a command of its own, after the line's words.
	var spans []string
	for _, w := range words {
		for _, sp := range bucSubstSpanRE.FindAllString(w, -1) {
			if strings.HasPrefix(sp, "$(") {
				spans = append(spans, sp)
			}
		}
	}
	for _, sp := range bucSubstSpanRE.FindAllString(strings.ReplaceAll(command, "\\`", ""), -1) {
		if strings.HasPrefix(sp, "`") {
			spans = append(spans, sp)
		}
	}
	// Each is wrapped in `( … )`: a `cd` inside a substitution ends at its close.
	for _, sp := range spans {
		inner := sp[2 : len(sp)-1]
		if strings.HasPrefix(sp, "`") {
			inner = sp[1 : len(sp)-1]
		}
		words = append(append(append(words, ";", "("), tokenize(inner)...), ")", ";")
	}
	return words
}

// dataWriteTarget is ABS when the whole command is one plain heredoc data write — `cat > ABS
// <<DELIM`, `cat >> ABS <<DELIM` or `tee [-a] ABS [>/dev/null] <<DELIM` with ABS a literal
// absolute path, its body, a closing line equal to DELIM (a trailing `\r` tolerated), then only
// blank lines — else "". Such a body is stdin data, so only ABS is judged; any other command is
// scanned whole, body included, so the hook is never weaker than a plain token scan. An unquoted
// delimiter's body runs `$(...)` and backticks, so one holding either is no data.
func dataWriteTarget(command string) string {
	lines := strings.Split(command, "\n")
	m := bucDataWriteRE.FindStringSubmatch(lines[0])
	if m == nil {
		return ""
	}
	target := bucOr(m[1], m[2])
	delim := bucOr(m[3], m[4], m[5])
	for k := 1; k < len(lines); k++ {
		if strings.TrimRight(lines[k], "\r") != delim {
			continue
		}
		body := strings.Join(lines[1:k], "\n")
		if m[5] != "" && bucRunsSubst(body) {
			return ""
		}
		for _, rest := range lines[k+1:] {
			if pyStrip(rest) != "" {
				return ""
			}
		}
		return target
	}
	return ""
}

// bucOr is Python's `a or b or …` over strings.
func bucOr(s ...string) string {
	for _, x := range s {
		if x != "" {
			return x
		}
	}
	return ""
}

// namesTarget is true when a command word may name the heredoc's target: the same spelling, the
// same file name under another path (`"$D"/x.sh` for `$SIB/x.sh`), or a file name an expansion
// hides.
func namesTarget(word, target string) bool {
	word = strings.NewReplacer(`"`, "", "'", "").Replace(word)
	name := bucLast(word)
	return word == target || name == bucLast(target) || strings.Contains(name, "$")
}

// readsOnly is true when the command words are one known not to execute its file arguments: a
// pipeline guard `flow-guard check-*`, `check-*.sh`/`check-*.py` other than the target itself (a
// written check script run afterwards runs the body), a reader in bucReaderWords (`rg` is not
// one: its `--pre` runs a command), `git` with a read subcommand, or `spectre validate`.
func readsOnly(words []string, target string) bool {
	name := bucLast(words[0])
	if bucCheckScriptRE.MatchString(name) && name != bucLast(target) {
		return true
	}
	// The bootstrap runs a compiled guard, never a file it is given; written as the target it
	// runs that body.
	if name == "flow-guard" && name != bucLast(target) && len(words) > 1 {
		return bucCheckGuardRE.MatchString(words[1])
	}
	if bucReaderWords[name] {
		return true
	}
	if name == "git" {
		return len(words) > 1 && bucGitReadSubcommands[words[1]]
	}
	if name == "spectre" {
		return len(words) > 1 && words[1] == "validate"
	}
	return false
}

// outsideSingleQuotes is text with every single-quoted span — `'…'`, and `$'…'` with its
// backslash escapes — blanked, since bash reads a backtick or `$(` inside one as a plain
// character; a `'` inside `"…"` or behind a backslash is literal. ok false when a quote never
// closes, or when a single-quoted span crosses a newline — an apostrophe in a `#` comment or
// another heredoc's body opens no quote in bash — so the caller reads the text as before.
func outsideSingleQuotes(text string) (string, bool) {
	rs := []rune(text)
	var out []rune
	quote := ""
	for i := 0; i < len(rs); {
		c := rs[i]
		if quote == "'" || quote == "$'" {
			if quote == "$'" && c == '\\' {
				i += 2
				continue
			}
			if c == '\n' {
				return "", false
			}
			if c == '\'' {
				quote = ""
			}
			i++
			continue
		}
		if c == '\\' {
			out = append(out, rs[i:min(i+2, len(rs))]...)
			i += 2
			continue
		}
		switch {
		case quote == `"`:
			if c == '"' {
				quote = ""
			}
		case c == '"':
			quote = `"`
		case c == '\'':
			quote = "'"
			if i > 0 && rs[i-1] == '$' {
				quote = "$'"
			}
			out = append(out, ' ')
			i++
			continue
		}
		out = append(out, c)
		i++
	}
	if quote != "" {
		return "", false
	}
	return string(out), true
}

type bucSegment struct {
	words    []string
	assigned bool
}

// commandSegments is (words, assigned) per command of text — quotes stripped, a leading
// `do`/`then`/`!` dropped, `assigned` when NAME=value words led it — or ok false when text holds a
// substitution this cannot read: a process substitution, a backtick, a nested `$(`. A plain
// `$(...)` is a command of its own. A literal NAME=value expands later words, so `$D/r.md` reads
// as the file D names; a name no literal assigns stays an expansion namesTarget reads as hiding
// any.
func commandSegments(text string) ([]bucSegment, bool) {
	var spans []string
	for _, m := range bucDollarParenRE.FindAllStringSubmatch(text, -1) {
		spans = append(spans, m[1])
	}
	flat := bucDollarParenRE.ReplaceAllLiteralString(text, " ")
	scan := flat
	if bare, ok := outsideSingleQuotes(flat); ok {
		scan = bare
	}
	if bucSubstRE.MatchString(scan) {
		return nil, false
	}
	env := map[string]string{}
	var out []bucSegment
	for _, seg := range bucCommandCutRE.Split(flat+"\n"+strings.Join(spans, "\n"), -1) {
		var words []string
		for _, w := range pySplit(seg) {
			words = append(words, strings.Trim(w, `'"`))
		}
		for len(words) > 0 && bucLeadingKeywords[words[0]] {
			words = words[1:]
		}
		assigned := false
		for len(words) > 0 && bucAssignmentRE.MatchString(words[0]) {
			name, value, _ := strings.Cut(words[0], "=")
			words = words[1:]
			value = expandVars(value, env)
			if value != "" && !strings.Contains(value, "$") {
				env[name] = value
			} else {
				delete(env, name)
			}
			assigned = true
		}
		expanded := make([]string, len(words))
		for i, w := range words {
			expanded[i] = expandVars(w, env)
		}
		out = append(out, bucSegment{words: expanded, assigned: assigned})
	}
	return out, true
}

// literalExport is true when every export argument is `NAME=value` whose value can run nothing:
// no `$` or backtick, and not naming target (namesTarget) — with no target, holding no `/`.
func literalExport(args []string, target string) bool {
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		if !bucAssignmentRE.MatchString(arg) {
			return false
		}
		_, value, _ := strings.Cut(arg, "=")
		if strings.ContainsAny(value, "$`") {
			return false
		}
		if target != "" && namesTarget(value, target) || target == "" && strings.Contains(value, "/") {
			return false
		}
	}
	return true
}

func bucAnyRunner(words []string) bool {
	return slices.ContainsFunc(words, func(w string) bool { return bucRunnerWords[bucLast(w)] })
}

// mayRunCode is true when text — the command outside one heredoc body — could run that body or
// the file target it writes. Fails closed: a process or command substitution (`>(sh)`, `bash <(`,
// `eval "$(`), a word that runs code (bucRunnerWords, matched by file name so `/bin/bash` counts
// as `bash`) anywhere, or any command whose words may name target (namesTarget — the same
// spelling, the same file name, a bare `x.sh`, an expansion) unless that command is one readsOnly
// allows. A command naming another file (`lint.sh /x/o.md`, `./gradlew`) leaves the body as data;
// `php $T`, `setsid $T`, `find . -exec`, `ssh h $T` keep it scanned, as does any command behind a
// leading `NAME=value` or run under `env`/`export` — save an `export` whose every argument is
// `NAME=value` with a literal value: no `$` or backtick, and one that cannot name target (with no
// target, no `/`). With no target no command names it, so substitutions, runner words, env
// assignments and a command word holding `/` or `$` count. Over-matching only keeps the body
// scanned.
func mayRunCode(text, target string) bool {
	segments, ok := commandSegments(text)
	if !ok {
		return true
	}
	for _, s := range segments {
		words := s.words
		if bucAnyRunner(words) {
			return true
		}
		// An env assignment can make any command run a file (GIT_EXTERNAL_DIFF, PAGER,
		// LD_PRELOAD, …): a command carrying one, or run under `env`/`export`, is never a reader.
		if len(words) > 0 && !s.assigned && words[0] == "export" && literalExport(words[1:], target) {
			continue
		}
		if len(words) > 0 && (s.assigned || bucLast(words[0]) == "env" || bucLast(words[0]) == "export") {
			return true
		}
		// With no target, a command word holding `/` or `$` may run what the opener wrote or read
		// (`/tmp/zz.sh`, `$l`).
		lead := words
		for len(lead) > 0 && bucShellKeywords[lead[0]] {
			lead = lead[1:]
		}
		if target == "" && len(lead) > 0 && strings.ContainsAny(lead[0], "/$") {
			return true
		}
		if target != "" && len(words) > 0 && slices.ContainsFunc(words, func(w string) bool { return namesTarget(w, target) }) &&
			!readsOnly(words, target) {
			return true
		}
	}
	return false
}

// runsAsProgram is true when text — the command outside one heredoc body — runs target only as a
// non-shell program: no substitution, and every command naming target is a reader (readsOnly), a
// bucProgramRunners interpreter given it (`python3 x.py`, `node x.js`) or `npx playwright test`,
// with no other runner word anywhere. Such a body is that language's text, never shell.
func runsAsProgram(text, target string) bool {
	segments, ok := commandSegments(text)
	if !ok {
		return false
	}
	for _, s := range segments {
		words := s.words
		if len(words) == 0 {
			continue
		}
		if bucProgramRunners[bucLast(words[0])] || len(words) >= 3 && words[0] == "npx" && words[1] == "playwright" && words[2] == "test" {
			if bucAnyRunner(words[1:]) {
				return false
			}
			continue
		}
		if bucAnyRunner(words) {
			return false
		}
		if s.assigned || bucLast(words[0]) == "env" || bucLast(words[0]) == "export" {
			return false
		}
		if slices.ContainsFunc(words, func(w string) bool { return namesTarget(w, target) }) && !readsOnly(words, target) {
			return false
		}
	}
	return true
}

// programGitLines is the git invocations a non-shell program body may make, one scan line each:
// quotes and brackets become spaces, so `subprocess.run(["git", "-C", d, "commit"])` reads
// `git -C d commit`, and every other word — a `d>0` comparison, an `rm` in a string — is dropped.
func programGitLines(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		if m := bucProgramGitRE.FindStringSubmatch(bucProgramPunctRE.ReplaceAllLiteralString(line, " ")); m != nil {
			out = append(out, m[1])
		}
	}
	return out
}

// substSpans is the `$(...)` and backtick spans an unquoted heredoc body runs, a backtick span
// rewritten as `$(...)`, or ok false when a span cannot be closed. body has its escaped backticks
// removed.
func substSpans(body string) ([]string, bool) {
	spans := bucSubstSpanRE.FindAllString(body, -1)
	dollar, closes := 0, 0
	for _, sp := range spans {
		if strings.HasPrefix(sp, "$(") {
			dollar++
			closes += strings.Count(sp, ")")
		}
	}
	// `$(case a in a) rm …;; esac)`: the span closes early and drops the rest.
	if strings.Count(body, "$(") != dollar || strings.Count(body, "`")%2 != 0 || strings.Count(body, ")") != closes {
		return nil, false
	}
	out := make([]string, len(spans))
	for i, sp := range spans {
		if strings.HasPrefix(sp, "`") {
			sp = "$(" + sp[1:len(sp)-1] + ")"
		}
		out[i] = sp
	}
	return out, true
}

func bucSplitRedirect(word string) []string {
	if m := bucAttachedRedirectRE.FindStringSubmatch(word); m != nil {
		return []string{m[1], m[2]}
	}
	return []string{word}
}

// bucHeredocEnd is the index of the line closing a heredoc opened before lines[k], -1 when none:
// a line equal to delim, a trailing `\r` tolerated, leading tabs too after `<<-`.
func bucHeredocEnd(lines []string, k int, delim string, tabs bool) int {
	for j := k; j < len(lines); j++ {
		l := lines[j]
		if tabs {
			l = strings.TrimLeft(l, "\t")
		}
		if strings.TrimRight(l, "\r") == delim {
			return j
		}
	}
	return -1
}

// bucOpener is a heredoc opener found in line: its delimiter, whether the delimiter is unquoted,
// whether it is `<<-`, and its match bounds.
type bucOpener struct {
	delim      string
	unquoted   bool
	tabs       bool
	start, end int
}

func bucFindOpener(line string) (bucOpener, bool) {
	m := bucHeredocOpenRE.FindStringSubmatchIndex(line)
	if m == nil {
		return bucOpener{}, false
	}
	group := func(g int) string {
		if m[2*g] < 0 {
			return ""
		}
		return line[m[2*g]:m[2*g+1]]
	}
	return bucOpener{
		delim:    bucOr(group(1), group(2), group(3)),
		unquoted: group(3) != "",
		tabs:     strings.HasPrefix(line[m[0]:], "<<-"),
		start:    m[0],
		end:      m[1],
	}, true
}

func bucRunsSubst(body string) bool {
	return strings.Contains(body, "$(") || strings.Contains(body, "`")
}

// dropDataBodies is lines with the body of every heredoc whose text the shell runs nothing of
// removed — a quoted delimiter's, or an unquoted one holding no `$(` or backtick — keeping its
// opener and delimiter lines, so a `bash <<'E'` opener still reads as a runner. stripDataHeredocs
// judges one body's rest through this: another heredoc's prose (a lone `.`, an unpaired backtick)
// is no command of the rest. A body naming target is kept: it may be a script, or a list, that
// runs the target.
func dropDataBodies(lines []string, target string) []string {
	var out []string
	for k := 0; k < len(lines); {
		line := lines[k]
		out = append(out, line)
		k++
		o, ok := bucFindOpener(line)
		if !ok {
			continue
		}
		end := bucHeredocEnd(lines, k, o.delim, o.tabs)
		if end < 0 {
			continue
		}
		body := strings.Join(lines[k:end], "\n")
		if target != "" && strings.Contains(body, target) || o.unquoted && bucRunsSubst(body) {
			continue
		}
		out = append(out, lines[end])
		k = end + 1
	}
	return out
}

func bucJoinRest(lines []string, k, end int, target string) string {
	return strings.Join(append(dropDataBodies(lines[:k], target), dropDataBodies(lines[end+1:], target)...), "\n")
}

// stripDataHeredocs is command with the body of a `cat`/`tee` heredoc to a named file removed —
// that body is data for a writer whose target the opening line already names, so a report line
// such as `a > the` must not read as a redirect. A body stays scanned when the writer has no
// target, its output is piped on, the target is a device or a process substitution
// (`/dev/stdout`, `>(sh)`), the rest of the command may run code (mayRunCode — an enclosing
// `bash <(` or `eval "$(`, a later `sh "$D"/x.sh` or `"$D"/x.sh`: written, then run; the target
// named again only as an argument is data), the opener is not a plain `cat`/`tee`
// (`python3 - <<EOF`, `x=$(cat <<EOF`). An unquoted-delimiter body that runs `$(...)` or backticks
// keeps only those spans scanned, a backtick span as `$(...)` — bash runs them and nothing else of
// the body — or the whole body when a span cannot be closed. Any other opener's body is data only
// when the opener is `flow` (an allowlist), its line is not continued, it redirects no output, and
// the rest of the command runs no code (mayRunCode).
func stripDataHeredocs(command string) string {
	lines := strings.Split(command, "\n")
	var out []string
	for k := 0; k < len(lines); {
		line := lines[k]
		out = append(out, line)
		k++
		o, ok := bucFindOpener(line)
		if !ok || strings.Contains(line[o.end:], "|") {
			continue
		}
		parts := bucSegmentSplitRE.Split(line[:o.start], -1)
		seg := parts[len(parts)-1] + " " + line[o.end:]
		words := pySplit(seg)
		for len(words) > 0 && bucAssignmentRE.MatchString(words[0]) {
			words = words[1:]
		}
		end := bucHeredocEnd(lines, k, o.delim, o.tabs)
		if len(words) > 0 && bucProgramRunners[bucLast(words[0])] && end >= 0 {
			// A python/node program on stdin: its text is not shell.
			body := strings.Join(lines[k:end], "\n")
			if o.unquoted && bucRunsSubst(body) {
				continue
			}
			out = append(out, programGitLines(body)...)
			k = end + 1
			continue
		}
		pre := line[:o.start]
		if end >= 0 &&
			// A `<<` in quotes, a `${…}` or a comment opens no heredoc: its "body" is shell.
			strings.Count(pre, "'")%2 == 0 &&
			strings.Count(pre, `"`)%2 == 0 &&
			strings.Count(pre, "${") <= strings.Count(pre, "}") &&
			!strings.Contains(pre, "#") &&
			len(words) > 0 &&
			// An allowlist, not a denylist: any other opener may be a stdin interpreter (`csh`,
			// `ssh h`, `sqlite3`), so only `flow` reads its body as data.
			strings.TrimLeft(words[0], `\`) == "flow" &&
			!slices.ContainsFunc(words[1:], func(w string) bool { return !bucFlowPlainWordRE.MatchString(w) || bucFlowShellWords[w] }) &&
			!strings.HasSuffix(strings.TrimRightFunc(line, pyIsSpace), `\`) && // a continued opener line hides the rest
			!strings.Contains(seg, ">") && // output written to a file may be run later
			!bucAnyRunner(bucOpenerWordSplitRE.Split(seg, -1)) &&
			!mayRunCode(bucJoinRest(lines, k, end, ""), "") {
			body := strings.ReplaceAll(strings.Join(lines[k:end], "\n"), "\\`", "")
			if o.unquoted && bucRunsSubst(body) {
				spans, ok := substSpans(body)
				if !ok {
					continue
				}
				out = append(out, spans...)
			}
			k = end + 1
			continue
		}
		if len(words) == 0 || words[0] != "cat" && words[0] != "tee" {
			continue
		}
		var target string
		if words[0] == "cat" {
			if t := bucCatTargetRE.FindStringSubmatch(seg); t != nil {
				target = t[1]
			}
		} else {
			for _, w := range words[1:] {
				if !strings.HasPrefix(w, "-") && !strings.HasPrefix(w, ">") {
					target = w
					break
				}
			}
		}
		if target == "" || end < 0 {
			continue
		}
		if strings.HasPrefix(target, "/dev/") || strings.Contains(target, "(") {
			continue
		}
		rest := bucJoinRest(lines, k, end, target)
		if mayRunCode(rest, target) {
			if runsAsProgram(rest, target) && !(o.unquoted && bucRunsSubst(strings.Join(lines[k:end], "\n"))) {
				out = append(out, programGitLines(strings.Join(lines[k:end], "\n"))...)
				k = end + 1
			}
			continue
		}
		// An escaped backtick is a literal in an unquoted body; dropped, it cannot pair with a real one.
		body := strings.ReplaceAll(strings.Join(lines[k:end], "\n"), "\\`", "")
		if o.unquoted && bucRunsSubst(body) {
			// Only the substitutions of an unquoted body run; its other text is data, so the spans
			// alone are kept scanned. A span this cannot close keeps the whole body.
			spans, ok := substSpans(body)
			if !ok {
				continue
			}
			out = append(out, spans...)
		}
		k = end + 1
	}
	return strings.Join(out, "\n")
}

// logicalLines is the command split into logical lines: a newline outside quotes cuts, a
// backslash-newline outside quotes joins, and a quoted newline stays inside its line — bash's own
// line discipline, close enough for a token scan.
func logicalLines(command string) []string {
	rs := []rune(command)
	var lines []string
	var buf []rune
	var quote rune // 0, '\'' or '"' while inside that quote
	n := len(rs)
	for i := 0; i < n; {
		ch := rs[i]
		if quote == 0 && ch == '\\' && i+1 < n && rs[i+1] == '\n' {
			i += 2 // continuation: the two physical lines are one command
			continue
		}
		if quote == 0 && ch == '\\' && i+1 < n ||
			quote == '"' && ch == '\\' && i+1 < n && strings.ContainsRune("\"\\$`", rs[i+1]) {
			buf = append(buf, ch, rs[i+1])
			i += 2
			continue
		}
		switch {
		case quote == 0 && (ch == '\'' || ch == '"'):
			quote = ch
		case quote == ch:
			quote = 0
		case quote == 0 && ch == '\n':
			lines = append(lines, string(buf))
			buf = nil
			i++
			continue
		}
		buf = append(buf, ch)
		i++
	}
	return append(lines, string(buf))
}

// expandVars is $NAME and ${NAME} replaced where NAME is set in env; the name match is greedy, so
// `$Dy` looks up `Dy`, never `D` — an unset name is left for resolve to let through, never
// guessed at.
func expandVars(token string, env map[string]string) string {
	if !strings.Contains(token, "$") || len(env) == 0 {
		return token
	}
	var b strings.Builder
	last := 0
	for _, m := range bucVarRE.FindAllStringSubmatchIndex(token, -1) {
		g := 2 // ${NAME}, else $NAME
		if m[2] < 0 {
			g = 4
		}
		name := token[m[g]:m[g+1]]
		b.WriteString(token[last:m[0]])
		if v, ok := env[name]; ok {
			b.WriteString(v)
		} else {
			b.WriteString(token[m[0]:m[1]])
		}
		last = m[1]
	}
	b.WriteString(token[last:])
	return b.String()
}

// defaultBranch is the repository's default branch: origin/HEAD's target.
func (b *buc) defaultBranch(cwd string) (string, bool) {
	head, _ := b.git(cwd, "symbolic-ref", "-q", "--short", "refs/remotes/origin/HEAD")
	_, after, ok := strings.Cut(head, "/")
	return after, ok
}

// inChangeState is true when target's worktree is a key of `worktrees` in the state record of the
// change named by its directory (`<project>-worktrees/<change>`), read with `flow state get`. A
// change spans sibling repos' worktrees (`<other>-worktrees/<change>`) but its record lives under
// one project, so each sibling checkout is asked in turn. Fails closed.
func (b *buc) inChangeState(target string) bool {
	top, _ := b.git(target, "rev-parse", "--show-toplevel")
	if top == "" {
		return false
	}
	name := pyBasename(top)
	top = b.realpath(top)
	for _, c := range append([]string{top}, bucGlobSiblings(pyDirname(pyDirname(top)), name)...) {
		keys, ok := b.stateWorktrees(c, name)
		if !ok {
			continue
		}
		found := false
		for _, k := range keys {
			if b.realpath(k) == top {
				found = true
			}
		}
		if found {
			return true
		}
	}
	return false
}

// bucGlobSiblings is Python's sorted(glob.glob(os.path.join(parent, "*-worktrees", name))): the
// non-hidden directories of parent ending in `-worktrees` that hold an entry name.
// ponytail: parent and name are taken literally, never as glob patterns of their own.
func bucGlobSiblings(parent, name string) []string {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, ".") || !strings.HasSuffix(n, "-worktrees") || !pyIsDir(pyJoin(parent, n)) {
			continue
		}
		p := pyJoin(parent, n, name)
		if _, err := os.Lstat(p); err == nil {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// stateWorktrees is the keys of the `worktrees` of `flow state get -C c name`'s record, ok false
// where Python's read raised an error it caught. A shape Python could not iterate panics, as it
// raised outside its try.
func (b *buc) stateWorktrees(c, name string) ([]string, bool) {
	if strings.ContainsRune(c, 0) || strings.ContainsRune(name, 0) {
		return nil, false
	}
	out, _, err := hookRun(b.env, hookBound(b.env, 5*time.Second), b.env.Dir, "flow", "state", "get", "-C", c, name)
	if err != nil || !utf8.Valid(out) {
		return nil, false
	}
	text := bucNewlines(string(out))
	i := strings.Index(text, "{")
	if i < 0 {
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(text[i:]))
	dec.UseNumber()
	var record any
	if dec.Decode(&record) != nil || dec.InputOffset() != int64(len(strings.TrimRight(text[i:], " \t\n\r"))) {
		return nil, false
	}
	obj, ok := record.(map[string]any)
	if !ok {
		return nil, false
	}
	switch w := obj["worktrees"].(type) {
	case map[string]any:
		keys := make([]string, 0, len(w))
		for k := range w {
			keys = append(keys, k)
		}
		return keys, true
	case []any:
		keys := make([]string, len(w))
		for i, k := range w {
			s, ok := k.(string)
			if !ok {
				panic("flow state: a worktrees entry is no string")
			}
			keys[i] = s
		}
		return keys, true
	case string:
		return strings.Split(w, ""), true
	default:
		if pyTruthy(w) {
			panic("flow state: worktrees is not iterable")
		}
		return nil, true
	}
}

// discardsUnstaged is the deny reason when `git checkout … -- <path>` or `git restore <path>` in
// a linked worktree names a path carrying unstaged edits, else "". Either restore reverts the file
// to its index state, so an uncommitted fix dies with the mutation-proof flip it was meant to
// undo. `git restore --staged` touches only the index and passes.
//
// ponytail: a `git checkout <path>` without `--` is not judged — telling a path from a branch
// there needs the ref namespace; add it when that spelling shows up in a reflog.
func (b *buc) discardsUnstaged(verb string, args []string, target string) string {
	if target == "" {
		return ""
	}
	var paths []string
	treeIsh := false
	switch verb {
	case "checkout":
		dd := slices.Index(args, "--")
		if dd < 0 {
			return ""
		}
		paths = args[dd+1:]
		treeIsh = slices.ContainsFunc(args[:dd], func(a string) bool { return !strings.HasPrefix(a, "-") })
	case "restore":
		if (slices.Contains(args, "--staged") || slices.Contains(args, "-S")) && !slices.Contains(args, "--worktree") && !slices.Contains(args, "-W") {
			return ""
		}
		treeIsh = slices.ContainsFunc(args, func(a string) bool {
			return a == "-s" || a == "--source" || strings.HasPrefix(a, "--source=")
		})
		if dd := slices.Index(args, "--"); dd >= 0 {
			paths = args[dd+1:]
		} else {
			skip := false
			for _, a := range args {
				switch {
				case skip:
					skip = false
				case a == "-s" || a == "--source":
					skip = true
				case !strings.HasPrefix(a, "-"):
					paths = append(paths, a)
				}
			}
		}
	default:
		return ""
	}
	if len(paths) == 0 {
		return ""
	}
	gitDir, _ := b.git(target, "rev-parse", "--absolute-git-dir")
	common, _ := b.git(target, "rev-parse", "--git-common-dir")
	if gitDir == "" || common == "" {
		return ""
	}
	if !strings.HasPrefix(common, "/") {
		common = pyJoin(target, common)
	}
	if b.realpath(gitDir) == b.realpath(common) {
		return "" // not a linked worktree
	}
	if b.inChangeState(target) {
		return "" // a change's own worktree: its fixer restores stray baselines
	}
	// a tree-ish or --source overwrites the index too, so staged edits are lost as well
	diff := []string{"diff", "--name-only"}
	if treeIsh {
		diff = append(diff, "HEAD")
	}
	dirty, _ := b.git(target, append(append(diff, "--"), paths...)...)
	if dirty == "" {
		return ""
	}
	files := pySplitLines(dirty)
	if !slices.ContainsFunc(files, func(d string) bool { return !strings.HasSuffix(strings.ToLower(d), ".png") }) {
		return "" // drifted visual-capture baselines: regenerable bytes, restored to base
	}
	return "Blocked: this `git " + verb + "` would discard the uncommitted edits on " + strings.Join(files, ", ") + " — it reverts the " +
		"file to its index state, and an uncommitted fix on it is lost with the flip a mutation " +
		"proof made (**Restoring a path with unstaged edits**, skills/flow-contracts/git-boundaries.md). " +
		"Restore a flipped file with `flow-guard break-and-prove`, which restores " +
		"from a pre-mutation byte snapshot; edits that must go are committed or stashed first."
}

// gitArgs is a git verb's own arguments: the words from start up to the first separator, `)` or
// redirect — `2>&1` and `>/dev/null` are the shell's, never a remote, refspec or pathspec.
func gitArgs(tokens []string, start int) []string {
	var out []string
	for _, t := range tokens[min(start, len(tokens)):] {
		if bucSeparators[t] || t == ")" {
			break
		}
		if strings.HasPrefix(t, "<") || strings.HasPrefix(t, ">") || strings.HasPrefix(t, "&>") {
			if len(out) > 0 && pyIsDigit(out[len(out)-1]) {
				out = out[:len(out)-1] // the fd of `2>&1`
			}
			break
		}
		out = append(out, t)
	}
	return out
}

// bashHits is the paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.
//
// The command is scanned one logical line at a time — a newline ends an argument scan the way `;`
// does, so a multi-line script's `cp` cannot swallow a later line's arguments — while `cd` state
// carries across lines, as it does in bash. Assignments the command text itself makes are expanded
// before path resolution, positionally: a token sees the assignments that came before it, in this
// line and earlier lines, never the ones after — and an assignment's value is kept literal,
// unexpanded.
func (b *buc) bashHits(command, cwd string, landings *[]string) []bucHit {
	if target := dataWriteTarget(command); target != "" {
		return []bucHit{{path: target}}
	}
	var hits []bucHit
	add := func(path string, link bool) { hits = append(hits, bucHit{path: path, link: link}) }
	cur := cwd
	var saved []string // the `cur` each open `(` will restore at its `)`
	env := map[string]string{}
	for _, line := range logicalLines(stripDataHeredocs(command)) {
		raw := tokenize(line)
		tokens := make([]string, 0, len(raw))
		for k, r := range raw {
			if bucAssignmentRE.MatchString(r) {
				// A run of assignments followed by a command word is that command's prefix: bash
				// does not use it when expanding the command's own words.
				next := "" // tokenize yields no empty word, so "" is none
				for _, t := range raw[k+1:] {
					if !bucAssignmentRE.MatchString(t) {
						next = t
						break
					}
				}
				if next == "" || bucSeparators[next] {
					name, value, _ := strings.Cut(r, "=")
					env[name] = value
				}
			}
			tokens = append(tokens, expandVars(r, env))
		}
		n := len(tokens)
		for i := 0; i < n; {
			tok := tokens[i]
			if tok == "(" {
				saved = append(saved, cur)
			} else if tok == ")" && len(saved) > 0 {
				cur = saved[len(saved)-1] // a `)` with nothing open is a `case` pattern's
				saved = saved[:len(saved)-1]
			}
			if tok == "cd" && i+1 < n {
				cur = b.resolve(tokens[i+1], cur)
				i += 2
				continue
			}
			expanded := strings.ContainsAny(tok, "$`")
			if tok == "git" || expanded && i+1 < n && tokens[i+1] == "push" {
				// `$GIT push` or `git $OPTS push`: the repository cannot be read.
				target := cur
				if expanded {
					target = ""
				}
				j := i + 1
				for j < n && (strings.HasPrefix(tokens[j], "-") || strings.ContainsAny(tokens[j], "$`")) {
					switch {
					case !strings.HasPrefix(tokens[j], "-"):
						target = ""
						j++
					case tokens[j] == "-C" && j+1 < n:
						target = b.resolve(tokens[j+1], cur)
						j += 2
					case tokens[j] == "-c" && j+1 < n:
						j += 2
					default:
						j++
					}
				}
				if j < n && bucDeniedGitVerbs[tokens[j]] {
					add(target, false)
				}
				if j < n && (tokens[j] == "checkout" || tokens[j] == "restore") {
					if discard := b.discardsUnstaged(tokens[j], gitArgs(tokens, j+1), target); discard != "" {
						*landings = append(*landings, discard)
					}
				}
				i = j + 1
				continue
			}
			if tok == "sed" && slices.ContainsFunc(tokens[i+1:min(i+4, n)], func(t string) bool { return strings.HasPrefix(t, "-i") }) {
				for _, t := range tokens[i+1:] {
					if bucSeparators[t] {
						break
					}
					// an empty word is BSD sed's `-i ''` suffix, never a path
					if t == "" || strings.HasPrefix(t, "-") {
						continue
					}
					if path := b.resolve(t, cur); path != "" && pyExists(path) {
						add(path, false)
					}
				}
				i++
				continue
			}
			if tok == "tee" {
				for _, t := range tokens[i+1:] {
					if bucSeparators[t] {
						break
					}
					if !strings.HasPrefix(t, "-") {
						add(b.resolve(t, cur), false)
					}
				}
				i++
				continue
			}
			// `>& f` writes stdout and stderr to f; `>& 2` and `>& -` only duplicate or close
			if i+1 < n && (bucRedirectOps[tok] || tok == ">&" && !pyIsDigit(tokens[i+1]) && tokens[i+1] != "-") {
				add(b.resolve(tokens[i+1], cur), false)
				i += 2
				continue
			}
			if tok == "cp" || tok == "mv" || tok == "rm" {
				var args []string
				for _, t := range tokens[i+1:] {
					if bucSeparators[t] || t == ")" {
						break
					}
					if !strings.HasPrefix(t, "-") {
						args = append(args, t)
					}
				}
				if tok == "rm" || tok == "mv" {
					// mv removes its sources too; neither follows a symlink it is given
					for _, a := range args {
						if r := b.resolve(a, cur); r != "" && !(tok == "rm" && b.localClone(r)) {
							add(r, true)
						}
					}
				} else if len(args) > 0 {
					add(b.resolve(args[len(args)-1], cur), false)
				}
				i++
				continue
			}
			if tok == "trash" {
				// a move to the Trash is recoverable, so only a path git tracks is a write
				for _, t := range tokens[i+1:] {
					if bucSeparators[t] || t == ")" {
						break
					}
					if strings.HasPrefix(t, "-") {
						continue
					}
					path := b.resolve(t, cur)
					if path == "" {
						continue
					}
					if _, ok := b.git(bucOr(pyDirname(path), "/"), "ls-files", "--error-unmatch", pyBasename(path)); ok {
						add(path, false)
					}
				}
				i++
				continue
			}
			i++
		}
	}
	return hits
}

// stashDenial is the reason a `git stash pop|apply|drop|branch` in command acts on a stash not
// made on the current branch (or names none), else "".
func (b *buc) stashDenial(command, cwd string) string {
	for p := 0; p < len(command); {
		k := strings.Index(command[p:], "git")
		if k < 0 {
			return ""
		}
		start := p + k
		p = start + 1
		if r, _ := utf8.DecodeLastRuneInString(command[:start]); start > 0 && pyIsWord(r) {
			continue
		}
		m := bucStashRE.FindStringSubmatchIndex(command[start:])
		if m == nil {
			continue
		}
		if r, _ := utf8.DecodeRuneInString(command[start+m[5]:]); start+m[5] < len(command) && pyIsWord(r) {
			continue // the verb runs on into a word: no `\b`
		}
		group := func(g int) string { return command[start+m[2*g] : start+m[2*g+1]] }
		p = start + m[1]
		verb, rest := group(2), pySplit(group(3))
		target := cwd
		flags := pySplit(group(1))
		for i, f := range flags {
			if f == "-C" && i+1 < len(flags) {
				target = b.resolve(flags[i+1], cwd)
			}
		}
		ref := ""
		for _, a := range rest {
			if r := bucStashRefRE.FindStringSubmatch(a); r != nil {
				ref = r[1]
				break
			}
		}
		if ref == "" && verb != "branch" {
			for _, a := range rest {
				if bucDigitsRE.MatchString(a) {
					ref = a
					break
				}
			}
		}
		if ref == "" {
			return "Blocked: `git stash " + verb + "` names no stash entry, so it acts on stash@{0}, which may belong to another change." + bucStashBlocked
		}
		listing, _ := b.git(target, "stash", "list")
		current, ok := b.git(target, "rev-parse", "--abbrev-ref", "HEAD")
		if !ok {
			current = "None"
		}
		madeOn := map[string]string{}
		for _, l := range pySplitLines(listing) {
			if x := bucStashListRE.FindStringSubmatch(l); x != nil {
				madeOn[pyIntDigits(x[1])] = x[2]
			}
		}
		branch, ok := madeOn[pyIntDigits(ref)]
		if !ok || branch != current {
			reason := "is not in the stash list."
			if ok {
				reason = "was made on `" + branch + "`, not on `" + current + "`."
			}
			return "Blocked: `git stash " + verb + " stash@{" + ref + "}` — that entry " + reason + bucStashBlocked
		}
	}
	return ""
}

func bucDenyReason(top, branch string) string {
	name := pyBasename(top)
	return "Blocked: " + top + " is the main checkout of " + name + ", and it is on `" + branch + "`, the default " +
		"branch. Agents never edit, stage, commit, stash or reset there — it is the landing " +
		"target every pipeline run pushes to, and anything left behind becomes residue no later " +
		"session owns.\n\n" +
		"Work in a worktree on a branch of its own instead:\n\n" +
		"  git -C " + top + " worktree add " + top + "-worktrees/<change> -b <change> origin/" + branch + "\n\n" +
		"then edit, commit and push there, and land with a push of `<change>:" +
		branch + "` or a pull request. To bring this checkout up to date afterwards, " +
		"`git -C " + top + " pull --ff-only` is allowed. Reads are always allowed. If the user " +
		"themselves must act here, hand them the exact command to run with the `!` prefix."
}
