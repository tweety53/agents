package main

// `flow setup <harness> [project-dir] [--register-hooks]` — the installer behind setup.sh, which
// only builds this binary from its own checkout and execs it (FLOW_SETUP_REPO_ROOT names that
// checkout, FLOW_SETUP_ARGV0 the path setup.sh was invoked as). Harnesses: claude-code | global.
//
//	./setup.sh claude-code                    # current directory
//	./setup.sh claude-code /path/to/other-project
//	./setup.sh global                         # user-level install (~/.claude)
//	./setup.sh global --register-hooks        # …and merge the hook registrations into settings.json
//
// Exit 0 is a complete install; 1 is a refusal (an "ERROR:" line) or a run that SKIPPED an item.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	claudeMDBegin = "<!-- flow:begin -->"
	claudeMDEnd   = "<!-- flow:end -->"

	// lintCommandsPlaceholder is what an opt-in rule carries where the project's declared lint
	// commands must appear. installProjectStandards resolves it from the project's own
	// .flow/project.md ## lint section before rendering; renderManagedBlock substitutes every
	// occurrence and refuses a placeholder it cannot resolve, so a managed block can never name
	// a command the project does not declare.
	lintCommandsPlaceholder = "{{lint-commands}}"

	// shimMarker is the second line of every exec shim an install writes into ~/.local/bin; only a
	// file carrying it there is ever removed or rewritten.
	shimMarker = "# flow script shim — written by setup.sh global; do not edit"
)

// setupDie is a refusal: runSetup prints it as an ERROR line and exits 1, as setup.sh's die did.
type setupDie string

type installer struct {
	repo           string
	stdout, stderr io.Writer
	skipped        int    // items that could not be linked; a skip is never a complete install
	lintCommands   string // the project's ## lint commands while its standards render
}

func (s *installer) die(format string, a ...any) { panic(setupDie(fmt.Sprintf(format, a...))) }
func (s *installer) info(format string, a ...any) {
	fmt.Fprintf(s.stdout, "→ "+format+"\n", a...)
}
func (s *installer) warn(format string, a ...any) {
	fmt.Fprintf(s.stderr, "  ⚠ "+format+"\n", a...)
}
func (s *installer) say(lines ...string) {
	for _, l := range lines {
		fmt.Fprintln(s.stdout, l)
	}
}

func runSetup(args []string, getenv func(string) string, stdout, stderr io.Writer) (code int) {
	s := &installer{repo: getenv("FLOW_SETUP_REPO_ROOT"), stdout: stdout, stderr: stderr}
	defer func() {
		if r := recover(); r != nil {
			msg, ok := r.(setupDie)
			if !ok {
				panic(r)
			}
			fmt.Fprintf(stderr, "ERROR: %s\n", msg)
			code = 1
		}
	}()
	removeBootstrapBuild(getenv)
	if s.repo == "" {
		s.die("flow setup runs through <agents repo>/setup.sh, which names the checkout to install from")
	}
	register := false
	var pos []string
	for _, a := range args {
		if a == "--register-hooks" {
			register = true
		} else {
			pos = append(pos, a)
		}
	}
	harness, projectArg := "", "."
	if len(pos) > 0 {
		harness = pos[0]
	}
	if len(pos) > 1 {
		projectArg = pos[1]
	}
	project, err := filepath.Abs(projectArg)
	if fi, serr := os.Stat(project); err != nil || serr != nil || !fi.IsDir() {
		// bash's own `cd` failure in the old installer, word for word
		fmt.Fprintf(stderr, "%s: line 23: cd: %s: No such file or directory\n", getenv("FLOW_SETUP_ARGV0"), projectArg)
		return 1
	}
	if harness == "" {
		s.die("Usage: %s <claude-code|global> [project-dir]", getenv("FLOW_SETUP_ARGV0"))
	}
	if !isDir(filepath.Join(s.repo, "skills")) {
		s.die("skills/ directory not found at %s", filepath.Join(s.repo, "skills"))
	}
	if harness == "global" {
		s.refuseLinkedWorktreeIntoRealHome(getenv)
	}
	switch harness {
	case "global":
		s.installGlobal(getenv("HOME"), register)
	case "claude-code":
		if register {
			s.die("--register-hooks applies to global only: the hooks are installed into ~/.claude by ./setup.sh global")
		}
		s.installClaudeCode(project)
		// Opt-in standards are rendered after the mode, so that the CLAUDE.md a first-time
		// `claude-code` install copies in is the file the block lands in. `global` installs no
		// project files at all, so a user-level install never writes into whatever directory it
		// happened to be run from.
		s.installProjectStandards(project)
	default:
		s.die("Unknown harness '%s'. Choose: claude-code | global", harness)
	}
	// A run that could not link something is not a successful install, and a caller (CI, a
	// Makefile, a human reading $?) must be able to tell without parsing the output.
	if s.skipped > 0 {
		return 1
	}
	return 0
}

// removeBootstrapBuild deletes the temporary directory setup.sh built this binary into — first
// thing, so a refusal leaves nothing behind either; a running binary may be unlinked. Only a
// directory named flow-setup.* holding this very executable is ever removed.
func removeBootstrapBuild(getenv func(string) string) {
	if getenv("FLOW_SETUP_REPO_ROOT") == "" {
		return
	}
	exe, err := os.Executable()
	if err != nil || filepath.Base(exe) != "flow" || !strings.HasPrefix(filepath.Base(filepath.Dir(exe)), "flow-setup.") {
		return
	}
	_ = os.RemoveAll(filepath.Dir(exe))
}

// refuseLinkedWorktreeIntoRealHome: a global install into the operator's own home links it into
// this checkout, so from a linked worktree every session's rules break the moment cleanup removes
// it. A sandboxed $HOME — every test and sandbox-diff run — is not the operator's and installs as
// before. The home is the password database's, SETUP_REAL_HOME standing in for it in a test.
func (s *installer) refuseLinkedWorktreeIntoRealHome(getenv func(string) string) {
	realHome := getenv("SETUP_REAL_HOME")
	if realHome == "" {
		if u, err := user.Current(); err == nil {
			realHome = u.HomeDir
		}
	}
	if getenv("HOME") != realHome {
		return
	}
	common, err := exec.Command("git", "-C", s.repo, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return
	}
	gitDir, _ := exec.Command("git", "-C", s.repo, "rev-parse", "--absolute-git-dir").Output()
	c := strings.TrimSpace(string(common))
	if strings.TrimSpace(string(gitDir)) != c {
		s.die("%s is a linked worktree — a global install from it breaks when it is removed. Run from the main checkout: %s/setup.sh global", s.repo, filepath.Dir(c))
	}
}

// --- filesystem predicates, each the bash test it replaces ---

func isDir(p string) bool     { fi, err := os.Stat(p); return err == nil && fi.IsDir() }            // [[ -d ]]
func isRegular(p string) bool { fi, err := os.Stat(p); return err == nil && fi.Mode().IsRegular() } // [[ -f ]]
func exists(p string) bool    { _, err := os.Stat(p); return err == nil }                           // [[ -e ]]
func isLink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}
func nonEmpty(p string) bool { fi, err := os.Stat(p); return err == nil && fi.Size() > 0 } // [[ -s ]]
func readable(p string) bool { return syscall.Access(p, 0x4) == nil }

// glob lists dir's non-hidden entries in name order, as the shell's `dir/*` does.
func glob(dir string) []string {
	ents, _ := os.ReadDir(dir)
	var out []string
	for _, e := range ents {
		if !strings.HasPrefix(e.Name(), ".") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// lines splits text into awk/grep records: a final newline ends the last record, never opens one.
func lines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// --- links ---

// linkInto symlinks src at dest. An existing symlink is replaced. Anything else already at dest — a
// real file or a stale copied directory — is moved aside with a warning first: a link over a
// directory would nest inside it and leave the stale copy shadowing the install, and over a file it
// would destroy hand-written content. Success is only ever reported for a link actually created.
//
// Where "aside" is depends on what is being displaced:
//
//   - A FILE goes to `<dest>.bak`, safe because the harness globs `commands/*.md` and
//     `rules/*.mdc`, and `.bak` falls outside both globs.
//   - A DIRECTORY goes to a sibling `<parent>-backup/<name>-<timestamp>/`, OUTSIDE the scanned
//     tree: a skills directory is discovered by walking the tree for SKILL.md, so a
//     `flow-start.bak/` would stay live alongside the symlink — two loadable skills declaring the
//     same name.
func (s *installer) linkInto(src, dest, label string) {
	if isLink(dest) {
		if err := os.Remove(dest); err != nil {
			s.die("rm: %s: %v", dest, err)
		}
	} else if _, err := os.Lstat(dest); err == nil {
		var backup string
		if isDir(dest) {
			backup = filepath.Join(filepath.Dir(dest)+"-backup", filepath.Base(dest)+"-"+time.Now().Format("20060102-150405"))
			if err := os.MkdirAll(filepath.Dir(backup), 0o777); err != nil {
				s.die("mkdir: %s: %v", filepath.Dir(backup), err)
			}
		} else {
			backup = dest + ".bak"
		}
		if exists(backup) {
			s.warn("%s already exists and %s is taken — SKIPPED, nothing linked; move one aside and re-run", dest, backup)
			s.skipped++
			return
		}
		if err := os.Rename(dest, backup); err != nil {
			s.die("mv: %s: %v", dest, err)
		}
		s.warn("%s already existed — moved to %s before symlinking", dest, backup)
	}
	if err := os.Symlink(src, dest); err != nil {
		s.die("ln: %s: %v", dest, err)
	}
	fmt.Fprintf(s.stdout, "  ✓ symlinked %s\n", label)
}

// finishBanner closes an install mode. A run that skipped an item is not a complete install and
// must not look like one; the count is taken since the baseline so each mode reports its own skips.
func (s *installer) finishBanner(label string, before int) {
	if n := s.skipped - before; n > 0 {
		fmt.Fprintf(s.stdout, "⚠ %s setup completed with %d skipped — see the warnings above. Those items are NOT installed.\n", label, n)
	} else {
		fmt.Fprintf(s.stdout, "✅ %s setup complete.\n", label)
	}
}

var alwaysApplyTrue = regexp.MustCompile(`^alwaysApply:[ \t\n\v\f\r]*true[ \t\n\v\f\r]*$`)

// alwaysOnRules names the rules that apply to every project: those whose leading frontmatter
// declares `alwaysApply: true`. Every other rule is opt-in: never installed as a file by any path,
// and reaching a project only when its own .flow/project.md names it (installProjectStandards).
//
// A trailing \r is dropped first (a CRLF checkout would otherwise drop every always-on rule); the
// value is anchored at the tail, so `alwaysApply: true_for_kotlin_only` is not `true`; and the
// frontmatter must CLOSE, within 20 lines, for the declaration to count — an unclosed `---` is not
// frontmatter, and a prose line documenting `alwaysApply: true` must not install a rule globally.
func (s *installer) alwaysOnRules() []string {
	var out []string
	for _, f := range glob(filepath.Join(s.repo, "rules")) {
		if !strings.HasSuffix(f, ".mdc") || !isRegular(f) {
			continue
		}
		b, _ := os.ReadFile(f)
		always, closed := false, false
		for i, l := range lines(string(b)) {
			l = strings.TrimSuffix(l, "\r")
			nr := i + 1
			if nr == 1 && l != "---" {
				break
			}
			if nr > 1 && l == "---" {
				closed = true
				break
			}
			if nr > 20 {
				break
			}
			if nr > 1 && alwaysApplyTrue.MatchString(l) {
				always = true
			}
		}
		if always && closed {
			out = append(out, filepath.Base(f))
		}
	}
	return out
}

// pruneStaleLinks removes symlinks in dir whose target no longer exists and whose stored target
// lies inside this repo. Installs are additive, so an entry deleted from this repo would otherwise
// leave its link behind — still matched by the harness's globs, offered and invoked. A link the
// user made to anywhere else is left alone however broken it looks: broken "at this moment" may be
// a volume that is not mounted. readlink, not realpath: the target does not exist to resolve.
func (s *installer) pruneStaleLinks(dir string) {
	if !isDir(dir) {
		return
	}
	if !readable(dir) {
		s.warn("cannot read %s — skipping the stale-link prune (fix its permissions and re-run)", dir)
		return
	}
	// `dir/*` then `dir/.[!.]*`: the visible entries first, then the hidden ones.
	ents, _ := os.ReadDir(dir)
	var visible, hidden []string
	for _, e := range ents {
		switch n := e.Name(); {
		case !strings.HasPrefix(n, "."):
			visible = append(visible, n)
		case !strings.HasPrefix(n, ".."):
			hidden = append(hidden, n)
		}
	}
	removed := 0
	for _, name := range append(visible, hidden...) {
		p := filepath.Join(dir, name)
		if !isLink(p) || exists(p) {
			continue
		}
		target, _ := os.Readlink(p)
		if !strings.HasPrefix(target, s.repo+"/") {
			continue
		}
		_ = os.Remove(p)
		removed++
		s.info("Pruned stale link %s (its source in this repo no longer exists)", name)
	}
	if removed > 0 {
		s.info("Pruned %d stale link(s) from %s", removed, dir)
	}
}

func (s *installer) mkdirAll(dir string) {
	if err := os.MkdirAll(dir, 0o777); err != nil {
		s.die("mkdir: %s: %v", dir, err)
	}
}

// linkEntries links every entry of srcDir accepted by keep into dest, the shell glob's order;
// a directory's link target carries the glob's trailing slash, as `dir/*/` yields it.
func (s *installer) linkEntries(srcDir, dest string, keep func(string) bool, dirs bool) {
	paths := glob(srcDir)
	if dirs { // `dir/*/` sorts its matches with the slash: "flow-x/" before "flow/"
		sort.Slice(paths, func(i, j int) bool { return paths[i]+"/" < paths[j]+"/" })
	}
	for _, p := range paths {
		if dirs {
			if !isDir(p) {
				continue
			}
			s.linkInto(p+"/", filepath.Join(dest, filepath.Base(p)), filepath.Base(p))
		} else if keep(p) && isRegular(p) {
			s.linkInto(p, filepath.Join(dest, filepath.Base(p)), filepath.Base(p))
		}
	}
}

func isMD(p string) bool { return strings.HasSuffix(p, ".md") }

func (s *installer) installSkills(dest string) {
	s.info("Installing project skills into %s", dest)
	s.mkdirAll(dest)
	s.pruneStaleLinks(dest)
	s.linkEntries(filepath.Join(s.repo, "skills"), dest, nil, true)
}

// installMods links each mods/<name>/ — a Claude Code plugin of function hooks — into the skills
// folder, from which Claude Code auto-loads a plugin folder; it runs after installSkills on the
// same folder, whose prune already dropped links to deleted mods.
func (s *installer) installMods(dest string) {
	if !isDir(filepath.Join(s.repo, "mods")) {
		return
	}
	s.info("Installing mods into %s", dest)
	s.mkdirAll(dest)
	s.linkEntries(filepath.Join(s.repo, "mods"), dest, nil, true)
}

func (s *installer) installCommands(dest string) {
	src := filepath.Join(s.repo, "commands-claude")
	if !isDir(src) {
		return
	}
	s.info("Installing slash commands into %s", dest)
	s.mkdirAll(dest)
	s.pruneStaleLinks(dest)
	s.linkEntries(src, dest, isMD, false)
}

func (s *installer) installAgents(dest string) {
	src := filepath.Join(s.repo, "agents")
	if !isDir(src) {
		return
	}
	s.info("Installing agent definitions into %s", dest)
	s.mkdirAll(dest)
	s.pruneStaleLinks(dest)
	s.linkEntries(src, dest, isMD, false)
}

// installRulesClaude links the FULL text of every always-on rule as <dest>/<name>.md — the path
// each core excerpt's `Full rule:` pointer in the managed block names — plus agent-baseline.md, the
// one file a dispatched agent is told to read. Links, never copies: the core and the full text are
// one file. The `.mdc` source is linked under a `.md` name, which Claude Code associates.
func (s *installer) installRulesClaude(dest string) {
	rules := filepath.Join(s.repo, "rules")
	if !isDir(rules) {
		return
	}
	s.info("Installing full rule texts into %s", dest)
	s.mkdirAll(dest)
	s.pruneStaleLinks(dest)
	for _, name := range s.alwaysOnRules() {
		md := strings.TrimSuffix(name, ".mdc") + ".md"
		s.linkInto(filepath.Join(rules, name), filepath.Join(dest, md), md)
	}
	baseline := filepath.Join(rules, "agent-baseline.md")
	if isRegular(baseline) {
		s.linkInto(baseline, filepath.Join(dest, "agent-baseline.md"), "agent-baseline.md")
	} else {
		s.warn("%s is missing — subagent dispatches point at a file that does not exist", baseline)
		s.skipped++
	}
}

func (s *installer) installClaudeCode(project string) {
	before := s.skipped
	s.info("Setting up for Claude Code in %s", project)
	s.installSkills(filepath.Join(project, ".claude/skills"))
	s.installCommands(filepath.Join(project, ".claude/commands"))
	s.installAgents(filepath.Join(project, ".claude/agents"))
	if !isRegular(filepath.Join(project, "CLAUDE.md")) {
		s.copyFile(filepath.Join(s.repo, "templates/CLAUDE.md"), filepath.Join(project, "CLAUDE.md"), false)
		s.info("Copied templates/CLAUDE.md to project root")
	} else {
		s.info("CLAUDE.md already exists — skipping copy (diff manually if needed)")
	}
	s.say("")
	s.finishBanner("Claude Code", before)
	s.say("   Skills → .claude/skills/  Commands → .claude/commands/  Agents → .claude/agents/",
		"   Next: in a Claude Code session, run /plugin install prime-radiant-inc/superpowers")
}

// copyFile is `cp` (a new file takes the source's mode under the umask) or, with preserve,
// `cp -p` (mode and modification time kept exactly).
func (s *installer) copyFile(src, dst string, preserve bool) {
	b, err := os.ReadFile(src)
	if err != nil {
		s.die("cp: %s: %v", src, err)
	}
	fi, _ := os.Stat(src)
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode().Perm())
	if err == nil {
		_, err = f.Write(b)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}
	if err == nil && preserve {
		if err = os.Chmod(dst, fi.Mode().Perm()); err == nil {
			err = os.Chtimes(dst, time.Now(), fi.ModTime())
		}
	}
	if err != nil {
		s.die("cp: %s: %v", dst, err)
	}
}

// --- the managed block ---

// stripFrontmatter drops a leading `---` … `---` block and every line's trailing \r, and drops
// trailing newlines as a shell command substitution does.
func stripFrontmatter(text string) string {
	var out []string
	fm := false
	for i, l := range lines(text) {
		l = strings.TrimSuffix(l, "\r")
		if i == 0 && l == "---" {
			fm = true
			continue
		}
		if fm && l == "---" {
			fm = false
			continue
		}
		if !fm {
			out = append(out, l)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func hasLine(body, line string) bool {
	for _, l := range strings.Split(body, "\n") {
		if l == line {
			return true
		}
	}
	return false
}

// renderManagedBlock renders rule bodies as the managed block's content. With no rules the set is
// the always-on rules — the global layer, the only one that renders a rule's core excerpt plus a
// `Full rule:` pointer; with rules it is exactly those, each body whole, which is how a project's
// opted-in standards render into its own CLAUDE.md (a pointer there would name a path no install
// creates for an opt-in rule). One renderer for both, so the delimiter guard, the frontmatter
// stripping and the `<!-- rule: … -->` labels cannot drift apart.
//
// It refuses a rule body carrying a delimiter on a line of its own: inlined, it would make the next
// run count two begins and die on a file the installer itself corrupted. It refuses an unbalanced
// core marker, which would render the wrong amount of text into every block, and a
// {{lint-commands}} it has no project's commands for.
func (s *installer) renderManagedBlock(rules ...string) string {
	global := len(rules) == 0
	if global {
		rules = s.alwaysOnRules()
	}
	var b strings.Builder
	b.WriteString(claudeMDBegin + "\n")
	// The provenance line names the command that reproduces THIS block.
	if global {
		b.WriteString("<!-- Generated by agents/setup.sh global. Do not edit inside this block: it is\n" +
			"     rewritten on every install. Put your own notes outside the delimiters. -->\n")
	} else {
		b.WriteString("<!-- Generated by agents/setup.sh from this project's .flow/project.md\n" +
			"     ## standards section. Do not edit inside this block: it is rewritten on\n" +
			"     every install. Put your own notes outside the delimiters. -->\n")
	}
	for _, name := range rules {
		file := filepath.Join(s.repo, "rules", name)
		if !isRegular(file) {
			continue
		}
		raw, _ := os.ReadFile(file)
		body := stripFrontmatter(string(raw))
		if strings.Contains(body, lintCommandsPlaceholder) {
			if s.lintCommands == "" {
				s.die("rule %s carries %s but is being rendered without a\n"+
					"  project's ## lint commands. The placeholder resolves only for a project's opted-in\n"+
					"  standards; an always-on rule must not carry it. Fix the rule, then re-run.", name, lintCommandsPlaceholder)
			}
			body = strings.ReplaceAll(body, lintCommandsPlaceholder, s.lintCommands)
		}
		if hasLine(body, claudeMDBegin) || hasLine(body, claudeMDEnd) {
			s.die("rule %s contains a flow block delimiter on a line of its own.\n"+
				"  Inlining it would put a second delimiter inside the managed block and make every\n"+
				"  later run die on a file this installer wrote. Indent or fence that line in\n"+
				"  %s so it is no longer a bare delimiter, then re-run.", name, file)
		}
		open, close := hasLine(body, "<!-- core -->"), hasLine(body, "<!-- /core -->")
		if open != close {
			s.die("rule %s has an unbalanced core marker. <!-- core --> and <!-- /core -->\n"+
				"  must both be present, each alone on its own line, or neither. One without the other\n"+
				"  silently renders the wrong amount of text into every managed block this installer writes.\n"+
				"  Fix %s, then re-run.", name, file)
		}
		fmt.Fprintf(&b, "\n<!-- rule: %s -->\n\n", name)
		for _, l := range strings.Split(body, "\n") {
			l = strings.TrimSuffix(l, "\r")
			if global && open && l == "<!-- /core -->" {
				break
			}
			if l == "<!-- core -->" || l == "<!-- /core -->" {
				continue
			}
			b.WriteString(l + "\n")
		}
		if global && open {
			fmt.Fprintf(&b, "Full rule: `~/.claude/rules/%s.md`.\n", strings.TrimSuffix(name, ".mdc"))
		}
	}
	b.WriteString("\n" + claudeMDEnd + "\n")
	return b.String()
}

// backupOnce keeps a copy of file as it was before flow first modified it, written once and never
// again: re-copying would replace the user's own pre-install file with the previous generated
// state. Mode and time are kept, so a 600 file of personal instructions never gains a wider copy.
func (s *installer) backupOnce(file string) {
	if exists(file + ".flow.bak") {
		s.info("Keeping the existing pre-install copy at %s.flow.bak", file)
	} else {
		s.copyFile(file, file+".flow.bak", true)
	}
}

// preflightManagedBlock refuses, with context, a parent that is not a directory and a target that
// is a directory, before any step of the run writes.
func (s *installer) preflightManagedBlock(target string) {
	dir := filepath.Dir(target)
	if exists(dir) && !isDir(dir) {
		s.die("%s exists but is not a directory, so the managed\n"+
			"  flow block for %s cannot be written. That path must be the harness's\n"+
			"  configuration directory. Move or remove the file at %s, then re-run.", dir, target, dir)
	}
	if isDir(target) {
		s.die("%s is a directory, but it must be the harness's\n"+
			"  agent instruction FILE — that is where the managed flow block goes. Move the directory\n"+
			"  aside, then re-run.", target)
	}
}

// installManagedBlock rewrites ONLY the delimited block of an agent instruction file. Exactly three
// outcomes: no delimiters → append a fresh block (after a .flow.bak copy); one begin before one end →
// rewrite that span; anything else → die. Appending in the third case could never converge (1/0
// becomes 2/1 becomes 3/2), and the stray delimiter may sit in content the user wants to keep.
func (s *installer) installManagedBlock(target string, rules ...string) {
	s.preflightManagedBlock(target)
	s.mkdirAll(filepath.Dir(target))
	block := s.renderManagedBlock(rules...)
	if !nonEmpty(target) {
		if err := os.WriteFile(target, []byte(block), 0o666); err != nil {
			s.die("cannot write %s: %v", target, err)
		}
		s.info("Wrote managed flow block to %s", target)
		return
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		s.die("cannot read %s (%v).\n"+
			"  Its content decides whether a managed flow block already exists, and guessing\n"+
			"  'no block' here would append a second one that no later run could reconcile.", target, err)
	}
	ls := lines(string(raw))
	// Every count below matches a marker as a whole LF-terminated line, so a CRLF file would count
	// 0 and 0 and take the append branch, orphaning its block: stop instead.
	var begins, ends []string
	crlf := 0
	for i, l := range ls {
		switch l {
		case claudeMDBegin + "\r", claudeMDEnd + "\r":
			crlf++
		case claudeMDBegin:
			begins = append(begins, strconv.Itoa(i+1))
		case claudeMDEnd:
			ends = append(ends, strconv.Itoa(i+1))
		}
	}
	if crlf > 0 {
		s.die("%s has CRLF line endings on its flow delimiters (%d marker line(s)).\n"+
			"  This installer matches the markers as LF-terminated lines, so it would not find the\n"+
			"  existing block and would append a second one. Convert the file to LF endings\n"+
			"  (e.g. `perl -pi -e 's/\\r\\n/\\n/' \"%s\"`) and re-run.", target, crlf, target)
	}
	if len(begins) == 0 && len(ends) == 0 {
		// The append is the first modification of a file this installer did not create, so the
		// backup belongs here too.
		s.backupOnce(target)
		f, err := os.OpenFile(target, os.O_WRONLY|os.O_APPEND, 0)
		if err == nil {
			_, err = f.WriteString("\n" + block)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		if err != nil {
			s.die("cannot append to %s: %v", target, err)
		}
		s.info("Appended managed flow block to %s (pre-install copy: %s.flow.bak)", target, target)
		return
	}
	// A single pair only counts if the begin precedes the end: reversed delimiters also count 1
	// and 1, and rewriting them deletes everything after the begin marker.
	if len(begins) == 1 && len(ends) == 1 {
		b, _ := strconv.Atoi(begins[0])
		e, _ := strconv.Atoi(ends[0])
		if b < e {
			var provenance string
			// Wholly generated (the block is the file, as wc -l counts it): nothing to back up.
			if b == 1 && e == bytes.Count(raw, []byte("\n")) {
				provenance = "no pre-install copy: this installer created the file"
			} else {
				s.backupOnce(target)
				provenance = "pre-install copy: " + target + ".flow.bak"
			}
			var out strings.Builder
			in := false
			for _, l := range ls {
				switch {
				case l == claudeMDBegin:
					in = true
					out.WriteString(block)
				case l == claudeMDEnd:
					in = false
				case !in:
					out.WriteString(l + "\n")
				}
			}
			s.swapInto(target, out.String())
			s.info("Refreshed managed flow block in %s (%s)", target, provenance)
			return
		}
	}
	none := func(v []string) string {
		if len(v) == 0 {
			return "none"
		}
		return strings.Join(v, ",")
	}
	s.die("flow delimiters in %s are not one begin followed by one end\n"+
		"  %s at line(s): %s\n"+
		"  %s at line(s): %s\n"+
		"  Rewriting could delete content between them and appending would never converge.\n"+
		"  Reconcile the file by hand — leave exactly one begin above exactly one end, or\n"+
		"  remove both markers — then re-run this installer.", target, claudeMDBegin, none(begins), claudeMDEnd, none(ends))
}

// swapInto replaces target by a same-directory rename, never an in-place truncate a concurrent
// reader could see torn. The swap file takes the target's mode, so the rename changes only the
// content; a failure before the rename leaves the target untouched.
func (s *installer) swapInto(target, content string) {
	fi, err := os.Stat(target)
	if err != nil {
		s.die("cannot stat %s: %v", target, err)
	}
	swap := filepath.Join(filepath.Dir(target), fmt.Sprintf(".%s.swap.%d", filepath.Base(target), os.Getpid()))
	defer os.Remove(swap)
	f, err := os.OpenFile(swap, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode().Perm())
	if err == nil {
		_, err = f.WriteString(content)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}
	if err == nil {
		err = os.Rename(swap, target)
	}
	if err != nil {
		s.die("cannot rewrite %s: %v", target, err)
	}
}

// --- project standards ---

var (
	standardsHeading = regexp.MustCompile(`^##[ \t\n\v\f\r]+standards[ \t\n\v\f\r]*$`)
	listItem         = regexp.MustCompile(`^[ \t\n\v\f\r]*[-*][ \t\n\v\f\r]+`)
	backtickSpan     = regexp.MustCompile("^`[^`]+`")
	plainRuleName    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*\.mdc$`)
)

// projectStandardsEntries lists `## standards` entries in file order: a top-level list item's first
// backticked span, or its first word when it has none. Only a `## ` heading ends the section, and a
// fenced block is illustration, never an entry.
func projectStandardsEntries(text string) []string {
	var out []string
	fence, in := false, false
	for _, l := range lines(text) {
		l = strings.TrimSuffix(l, "\r")
		switch {
		case strings.HasPrefix(l, "```"):
			fence = !fence
			continue
		case fence:
			continue
		case len(l) > 2 && strings.HasPrefix(l, "##") && l[2] != '#':
			in = standardsHeading.MatchString(l)
			continue
		case !in:
			continue
		}
		loc := listItem.FindStringIndex(l)
		if loc == nil {
			continue
		}
		item := l[loc[1]:]
		var entry string
		if m := backtickSpan.FindString(item); m != "" {
			entry = m[1 : len(m)-1]
		} else if i := strings.IndexAny(item, " \t\n\v\f\r"); i >= 0 {
			entry = item[:i]
		} else {
			entry = item
		}
		if entry != "" {
			out = append(out, entry)
		}
	}
	return out
}

// projectLintCommands is the project's declared lint commands, read from `.flow/project.md`'s ## lint
// through `scripts/flow-guard project-get` — the reader every /flow phase uses, so the rendered block
// and the commands the pipeline runs cannot drift. Dies when the section is absent (exit 1),
// unreadable (any other exit), or carries no command line inside a fence.
func (s *installer) projectLintCommands(project, rule string) string {
	cmd := exec.Command(filepath.Join(s.repo, "scripts/flow-guard"), "project-get", project, "lint")
	cmd.Stderr = s.stderr
	out, err := cmd.Output()
	rc := 0
	if err != nil {
		rc = 127
		if ee, ok := err.(*exec.ExitError); ok {
			rc = ee.ExitCode()
		}
	}
	if rc == 1 {
		s.die("the project's resolved .flow/project.md carries no ## lint section for the placeholder\n"+
			"  in %s to render from. flow-guard project-get resolves HEAD's copy when the project\n"+
			"  sits in a git work tree, so a section that exists only uncommitted does not count:\n"+
			"  commit the ## lint section, or drop the rule from ## standards, then re-run.", rule)
	}
	if rc != 0 {
		s.die("cannot read the project's ## lint section (flow-guard project-get exit %d): the placeholder\n"+
			"  in %s has nothing safe to render from.", rc, rule)
	}
	var fenced []string
	fence := false
	for _, l := range lines(strings.TrimRight(string(out), "\n")) {
		if strings.HasPrefix(l, "```") {
			if fence {
				break
			}
			fence = true
			continue
		}
		if fence {
			fenced = append(fenced, l)
		}
	}
	blank := func(l string) bool { return strings.TrimLeft(l, " \t\n\v\f\r") == "" }
	for len(fenced) > 0 && blank(fenced[0]) {
		fenced = fenced[1:]
	}
	for len(fenced) > 0 && blank(fenced[len(fenced)-1]) {
		fenced = fenced[:len(fenced)-1]
	}
	commands := strings.TrimRight(strings.Join(fenced, "\n"), "\n")
	if commands == "" {
		s.die("the project's ## lint section carries no command lines inside a\n"+
			"  fenced block, so %s in %s has nothing to render.", lintCommandsPlaceholder, rule)
	}
	return commands
}

// installProjectStandards renders the shared rules this project opted into — its .flow/project.md
// `## standards` entries — into a managed block in <project>/CLAUDE.md. Opt-in rules are installed
// nowhere by path (the Kotlin standard's globs would match every Kotlin repo on the machine), so this
// generated block is how a project loads one. Only a bare `.mdc` filename of the shared library is
// rendered (rules/flow-manual-review.mdc's resolution table); an always-on rule is listed and
// dropped, since it already reaches every session through the global block.
func (s *installer) installProjectStandards(project string) {
	projectMD := filepath.Join(project, ".flow/project.md")
	if !isRegular(projectMD) {
		return
	}
	before := s.skipped
	always := map[string]bool{}
	for _, r := range s.alwaysOnRules() {
		always[r] = true
	}
	raw, _ := os.ReadFile(projectMD)
	var selected, alreadyGlobal []string
	seen := map[string]bool{}
	for _, entry := range projectStandardsEntries(string(raw)) {
		if !strings.HasSuffix(entry, ".mdc") || strings.Contains(entry, "/") {
			continue
		}
		if !plainRuleName.MatchString(entry) {
			s.warn("%s names '%s' under ## standards, which is not a plain rule filename — SKIPPED", projectMD, entry)
			s.skipped++
			continue
		}
		if always[entry] {
			alreadyGlobal = append(alreadyGlobal, entry)
			continue
		}
		rules := filepath.Join(s.repo, "rules")
		if !isRegular(filepath.Join(rules, entry)) {
			s.warn("%s names '%s' under ## standards, but %s/%s does not exist — SKIPPED, not rendered", projectMD, entry, rules, entry)
			s.skipped++
			continue
		}
		if !seen[entry] {
			seen[entry] = true
			selected = append(selected, entry)
		}
	}
	if len(selected) == 0 {
		// Writing an empty block would create a CLAUDE.md in a project that never asked for one.
		if s.skipped != before {
			s.finishBanner("Project standards", before)
		}
		return
	}
	s.say("")
	s.info("Rendering opt-in standards into %s/CLAUDE.md (from .flow/project.md)", project)
	for _, r := range selected {
		fmt.Fprintf(s.stdout, "  ✓ %s\n", r)
	}
	for _, r := range alreadyGlobal {
		fmt.Fprintf(s.stdout, "  – %s is always-on — it already reaches every session through the global block\n", r)
	}
	// All-or-nothing: the lint payload, every rule body and the target are checked before the
	// first write.
	s.lintCommands = ""
	for _, r := range selected {
		if b, _ := os.ReadFile(filepath.Join(s.repo, "rules", r)); bytes.Contains(b, []byte(lintCommandsPlaceholder)) {
			s.lintCommands = s.projectLintCommands(project, r)
			break
		}
	}
	s.renderManagedBlock(selected...)
	s.preflightManagedBlock(filepath.Join(project, "CLAUDE.md"))
	s.installManagedBlock(filepath.Join(project, "CLAUDE.md"), selected...)
	s.finishBanner("Project standards", before)
}

// --- global ---

// installScriptShims writes the one exec shim, flow-guard, into bin (~/.local/bin, already on PATH),
// so the `flow-guard <name>` the skills cite resolves onto scripts/flow-guard, the bootstrap every
// guard runs through. A shim, not a symlink: exec keeps the bootstrap's own path, which it derives
// its checkout from. Every marked shim is removed first (so a per-name .sh shim an earlier install
// wrote goes), and a user's own file in the shim's place is SKIPPED, never touched.
func (s *installer) installScriptShims(bin string) {
	// ponytail: the exec target is single-quoted, so a checkout path holding ' is refused rather
	// than escaped; escape it if such a checkout ever matters.
	if strings.Contains(s.repo, "'") {
		s.die("SCRIPT_DIR contains a single quote (%s) — cannot write script shims", s.repo)
	}
	s.mkdirAll(bin)
	for _, f := range glob(bin) {
		if fi, err := os.Lstat(f); err != nil || !fi.Mode().IsRegular() {
			continue
		}
		if b, err := os.ReadFile(f); err == nil && bytes.HasPrefix(b, []byte("#!/bin/sh\n"+shimMarker+"\n")) {
			_ = os.Remove(f)
		}
	}
	shim := filepath.Join(bin, "flow-guard")
	if _, err := os.Lstat(shim); err == nil {
		s.warn("%s already exists and is not a flow shim — SKIPPED, no shim written", shim)
		s.skipped++
		return
	}
	content := fmt.Sprintf("#!/bin/sh\n%s\nexec '%s' \"$@\"\n", shimMarker, filepath.Join(s.repo, "scripts/flow-guard"))
	if err := os.WriteFile(shim, []byte(content), 0o666); err != nil {
		s.die("cannot write %s: %v", shim, err)
	}
	if err := os.Chmod(shim, 0o755); err != nil {
		s.die("chmod: %s: %v", shim, err)
	}
}

func (s *installer) installGlobal(home string, register bool) {
	if home == "" {
		s.die("HOME must be set")
	}
	before := s.skipped
	managed := filepath.Join(home, ".claude/CLAUDE.md")
	settings := filepath.Join(home, ".claude/settings.json")
	s.info("Setting up flow globally under %s", home)

	// PREFLIGHT — everything that can refuse this run refuses BEFORE the first symlink, so the
	// harness never picks up a rule whose managed block was refused.
	s.renderManagedBlock()
	s.preflightManagedBlock(managed)
	if register {
		if err := checkSettings(settings); err != nil {
			s.die("%v", err)
		}
	}

	s.installSkills(filepath.Join(home, ".claude/skills"))
	s.installMods(filepath.Join(home, ".claude/skills"))
	s.installCommands(filepath.Join(home, ".claude/commands"))
	s.installAgents(filepath.Join(home, ".claude/agents"))
	// The links go in BEFORE the block, so a `Full rule:` pointer is never written to a path that
	// does not exist yet.
	s.installRulesClaude(filepath.Join(home, ".claude/rules"))
	s.installHooks(filepath.Join(home, ".claude"), register)
	s.installScriptShims(filepath.Join(home, ".local/bin"))
	s.installManagedBlock(managed)
	s.say("")
	s.finishBanner("Global", before)
	s.say(
		"   Skills   → "+home+"/.claude/skills/",
		"   Mods     → "+home+"/.claude/skills/ (Claude Code function-hooks plugins)",
		"   Commands → "+home+"/.claude/commands/",
		"   Agents   → "+home+"/.claude/agents/",
		"   Scripts  → "+home+"/.local/bin/",
		"   Rules    → the managed block in "+home+"/.claude/CLAUDE.md",
		"              Full texts → "+home+"/.claude/rules/, which the core excerpts in the",
		"              managed block point at. Same source, one link each, no copies.",
		"   Hooks    → "+home+"/.claude/hooks/ — installed, but registered only if",
		"              settings.json already names them (see any warning above).",
		"   Subagents inherit NO rules. Every dispatch must carry the two-sentence pointer to",
		"   "+home+"/.claude/rules/agent-baseline.md — that file is the whole rule set for a",
		"   dispatched agent, and it propagates itself to any depth.",
		"   Opt-in rules (e.g. kotlin-backend-development-standard) are NOT installed globally.",
		"   A project adopts one by naming it under ## standards in its own .flow/project.md;",
		"   run './setup.sh <harness> /path/to/that/project' to render it into that project's",
		"   own CLAUDE.md. This mode never writes project files.",
		"",
		"   Source   → "+s.repo,
		"   Every link above points back into that directory, so global agent behavior",
		"   follows whatever is checked out there. Install from a stable main checkout —",
		"   from a git worktree or feature branch the links break when it is removed.",
	)
}
