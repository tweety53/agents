package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// checkInstalledRules is scripts/check-installed-rules.sh: that script's
// header comment is the contract -- fail when this repository's always-on
// rules and what `setup.sh global` last installed have drifted apart. Exit 0
// on the OK, NONE and WORKTREE verdicts, 1 on the STALE verdict and on every
// refusal: this guard has no exit 2.
func init() {
	Registry["check-installed-rules"] = checkInstalledRules
}

const cirName = "check-installed-rules"

// cirMarker is rule 4's `grep -o '<!-- rule: [^ ]*\.mdc -->'`: a marker
// cannot contain a space, so the match is unambiguous per line.
var cirMarker = regexp.MustCompile(`<!-- rule: ([^ \n]*\.mdc) -->`)

// cirAlways is the always-on test's `/^alwaysApply:[[:space:]]*true[[:space:]]*$/`.
var cirAlways = regexp.MustCompile(`^alwaysApply:[[:space:]]*true[[:space:]]*$`)

func checkInstalledRules(_ []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, cirName+": "+format+"\n", a...)
		return 1
	}
	// The rule set is this repository's own rules/, resolved from the
	// script's own location. The binary lives in a cache, so the shim
	// exports that location's parent as FLOW_GUARD_REPO_ROOT.
	repoRoot := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if repoRoot == "" {
		return refuse("FLOW_GUARD_REPO_ROOT is unset — run scripts/check-installed-rules.sh, which sets it")
	}
	rulesSrc := repoRoot + "/rules"
	homeDir := env.Getenv("CHECK_INSTALLED_RULES_HOME")
	override := homeDir != ""
	if !override {
		// `${CHECK_INSTALLED_RULES_HOME:-$HOME}` under set -u: the bash died
		// on an unset HOME, so the port refuses rather than read it as empty.
		h, ok := env.LookupEnv("HOME")
		if !ok {
			return refuse("HOME is unset and CHECK_INSTALLED_RULES_HOME is not set — cannot locate the install")
		}
		homeDir = h
	}
	rulesDir := homeDir + "/.claude/rules"

	violations := 0
	violation := func(format string, a ...any) {
		fmt.Fprintf(stderr, cirName+": "+format+"\n", a...)
		violations++
	}

	if !isDir(rulesSrc) {
		return refuse("%s does not exist — this is not an agents checkout", rulesSrc)
	}
	alwaysOn, err := cirAlwaysOnRules(env, rulesSrc)
	if err != nil {
		return refuse("%v", err)
	}
	if len(alwaysOn) == 0 {
		return refuse("%s declares no always-on rule — refusing to report an empty install as clean", rulesSrc)
	}

	if !override && isFile(repoRoot+"/.git") {
		fmt.Fprintf(stdout, "INSTALLED-RULES-WORKTREE: %s is a linked worktree — the install tracks the main checkout, nothing to check here\n", repoRoot)
		return 0
	}

	if !isDir(rulesDir) {
		fmt.Fprintf(stdout, "INSTALLED-RULES-NONE: %s — no global install found, nothing to check\n", homeDir)
		return 0
	}

	// Rule 1 — every always-on rule is linked, resolves, and points back here.
	for _, ruleName := range alwaysOn {
		installed := rulesDir + "/" + strings.TrimSuffix(ruleName, ".mdc") + ".md"
		want := rulesSrc + "/" + ruleName
		fi, err := os.Lstat(installed)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			if err == nil {
				violation("%s is not a symlink — a copy goes stale the next time the rule is edited", installed)
			} else {
				violation("%s is missing — rules/%s is always-on but was never installed", installed, ruleName)
			}
			continue
		}
		target, _ := os.Readlink(installed)
		if _, err := os.Stat(installed); err != nil {
			violation("%s is a dangling symlink — it points at %s, which does not exist", installed, target)
			continue
		}
		// The link's target directory as `cd <link dir> && cd <target dir>
		// && pwd` names it: one readlink, `..` resolved logically.
		dir := filepath.Dir(target)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(installed), dir)
		}
		if got := filepath.Clean(dir) + "/" + filepath.Base(target); got != want {
			violation("%s points at %s, not this checkout's %s", installed, got, want)
		}
	}

	// Rule 2 — nothing installed that is no longer an always-on rule.
	installedMDs, err := cirGlob(env, rulesDir, ".md")
	if err != nil {
		return refuse("%v", err)
	}
	for _, installed := range installedMDs {
		base := filepath.Base(installed)
		if base == "agent-baseline.md" {
			continue
		}
		if !slices.Contains(alwaysOn, strings.TrimSuffix(base, ".md")+".mdc") {
			violation("%s is installed but rules/%s.mdc is not an always-on rule here — a stale link still reads as installed", installed, strings.TrimSuffix(base, ".md"))
		}
	}

	// Rule 3 — the baseline every dispatched agent is told to read.
	baseline := rulesDir + "/agent-baseline.md"
	if _, err := os.Stat(baseline); err != nil {
		violation("%s is missing or dangling — every subagent dispatch points at it", baseline)
	}

	// Rule 4 — the managed block in each installed harness file renders each
	// rule. The file set is setup.sh's own `managed_files` declaration, parsed
	// live (kan-585) — see the header's rule 4 for why a hardcoded copy here
	// was already wrong once. CHECK_INSTALLED_RULES_SETUP_SH defaults to this
	// checkout's setup.sh.
	setupSh := env.Getenv("CHECK_INSTALLED_RULES_SETUP_SH")
	if setupSh == "" {
		setupSh = repoRoot + "/setup.sh"
	}
	setup, err := os.ReadFile(setupSh)
	if err != nil {
		return refuse("%s is unreadable — cannot resolve the managed-block targets", setupSh)
	}
	harnessFiles, code := cirManagedFiles(string(setup), setupSh, refuse)
	if harnessFiles == nil {
		return code
	}
	for _, rel := range harnessFiles {
		harness := homeDir + "/" + rel
		if !isFile(harness) {
			fmt.Fprintf(stderr, "%s: %s does not exist, not scanned\n", cirName, harness)
			continue
		}
		b, _ := os.ReadFile(harness)
		body := string(b)
		for _, ruleName := range alwaysOn {
			if !strings.Contains(body, "<!-- rule: "+ruleName+" -->") {
				violation("%s carries no managed-block marker for %s — no session reads that rule", harness, ruleName)
			}
		}
		var markers []string
		for _, m := range cirMarker.FindAllStringSubmatch(body, -1) {
			markers = append(markers, m[1])
		}
		// `sort -u`, exec'd so the order is the caller's collation (crSort).
		markers, err := crSort(env, markers, true)
		if err != nil {
			return refuse("%v", err)
		}
		for _, marker := range markers {
			if !slices.Contains(alwaysOn, marker) {
				violation("%s renders %s, which is not an always-on rule here", harness, marker)
			}
		}
	}

	if violations != 0 {
		fmt.Fprintf(stdout, "INSTALLED-RULES-STALE: %s — %d violation(s); re-run ./setup.sh global from %s\n", homeDir, violations, repoRoot)
		return 1
	}
	fmt.Fprintf(stdout, "INSTALLED-RULES-OK: %s — %d always-on rule(s) installed and rendered\n", homeDir, len(alwaysOn))
	return 0
}

// cirManagedFiles parses setup.sh's `local managed_files=(...)` into
// $home_dir-relative paths; nil means it refused, with refuse's exit code.
//
// Anchored at line start (optional leading whitespace only): a commented-out
// stale declaration must never serve (F2/F5, round 0). The trailing-`)` check
// refuses a declaration this one-line parse cannot fully see — a multi-line
// declaration would otherwise be silently truncated to its first line, the
// guard scanning a subset and reporting green over an unscanned file's broken
// block (F1/F4, round 0).
func cirManagedFiles(setup, setupSh string, refuse func(string, ...any) int) ([]string, int) {
	decl := ""
	for _, line := range strings.Split(setup, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t\n\v\f\r"), "local managed_files=(") {
			decl = line
			break
		}
	}
	if decl == "" {
		return nil, refuse("no 'local managed_files=(' declaration in %s — cannot resolve the managed-block targets", setupSh)
	}
	if !strings.HasSuffix(decl, ")") {
		return nil, refuse("the managed_files declaration in %s does not close on its own line — this guard parses a single-line declaration only, and a truncated set must be refused, never silently served", setupSh)
	}
	_, rest, _ := strings.Cut(decl, "managed_files=(")
	files := []string{}
	// The bash word-split the rest unquoted: IFS's space, tab and newline.
	for _, el := range strings.FieldsFunc(rest, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }) {
		el = strings.TrimSuffix(el, ")")
		// The bare `)` of an empty `managed_files=()` is not an element:
		// skipping it lets the empty-declaration refusal below fire, where it
		// names the real problem, instead of this loop misreporting it as
		// unresolvable.
		if el == "" {
			continue
		}
		el = strings.TrimPrefix(strings.TrimSuffix(el, `"`), `"`)
		rel, ok := strings.CutPrefix(el, "$home_dir/")
		if !ok {
			return nil, refuse("cannot resolve managed_files element '%s' in %s — expected a \"$home_dir/-relative path", el, setupSh)
		}
		files = append(files, rel)
	}
	if len(files) == 0 {
		return nil, refuse("setup.sh's managed_files declaration is empty in %s — refusing to scan nothing and call it clean", setupSh)
	}
	return files, 0
}

// cirAlwaysOnRules is always_on_rules — the same frontmatter test setup.sh's
// own always_on_rules() applies, kept identical on purpose: a rule this guard
// calls always-on that the installer does not would demand a link setup.sh
// never creates, and the two would disagree forever with no way to satisfy
// both. It is a port of setup.sh's awk because setup.sh is not sourceable —
// it runs an install on load.
func cirAlwaysOnRules(env Env, rulesSrc string) ([]string, error) {
	files, err := cirGlob(env, rulesSrc, ".mdc")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, f := range files {
		if !isFile(f) {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if cirIsAlwaysOn(string(b)) {
			out = append(out, filepath.Base(f))
		}
	}
	return out, nil
}

// cirIsAlwaysOn is the awk: a `---` first line, an `alwaysApply: true` line
// and a closing `---`, all within the first 20 lines; CRs stripped.
func cirIsAlwaysOn(body string) bool {
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	if body == "" {
		lines = nil
	}
	always := false
	for i, line := range lines {
		nr := i + 1
		line = strings.TrimSuffix(line, "\r")
		switch {
		case nr == 1 && line != "---":
			return false
		case nr > 1 && line == "---":
			return always
		case nr > 20:
			return false
		case nr > 1 && cirAlways.MatchString(line):
			always = true
		}
	}
	return false
}

// cirGlob is bash's `"$dir"/*<ext>`: non-hidden names, in the glob's order
// — the locale's collation, which an exec'd `sort` reproduces (crSort).
func cirGlob(env Env, dir, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var paths []string
	for _, e := range entries {
		if n := e.Name(); !strings.HasPrefix(n, ".") && strings.HasSuffix(n, ext) {
			paths = append(paths, dir+"/"+n)
		}
	}
	return crSort(env, paths, false)
}
