package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"syscall"
)

// checkGuardSymlinks is scripts/check-guard-symlinks.sh: that script's header
// comment is the contract — six rules over skills/ and scripts/, one
// `path:line: message` line per finding, then the GUARD-SYMLINKS-OK or
// GUARD-SYMLINKS-INVALID verdict; exit 0 clean, 1 any violation, 2 cannot
// answer (a refusal writes its reason to stderr and nothing to stdout). The
// reasoning for each step, moved here from the bash body it replaced
// (d71a2327), sits beside the code it explains.
//
// THE THREE DISCIPLINES the bash carried are kept in the shape a Go read
// takes: a file is read whole as bytes, so a stray NUL cannot put a matcher
// into binary "no match" mode (grep's `-a`, which rule 4 and the sibling scan
// used); the citation, rule 3 and delegation scans ran through awk, which
// ends a line at its first NUL, so gsScanMD and gsDelegates cut there too;
// a failed read is a failure to
// look, reported as the bash reported grep's `rc > 1`, never read as an empty
// answer; and no path is ever parsed as an option, because none reaches an
// argv. The bash's scratch directory is gone with its mktemp refusal: every
// intermediate list lives in memory.
//
// ponytail: a relative CHECK_GUARD_SYMLINKS_ROOT resolves against the process
// cwd, which is env.Dir in production (flow-guard runs in the shim's cwd); the
// override is a test fixture's, and tests pass absolute roots.
func init() {
	Registry["check-guard-symlinks"] = checkGuardSymlinks
}

// gsExempt is rule 3's EXEMPT set: the five project-configured guards named in
// design.md, "Two families of guard, and only one of them ships": these are
// resolved through a project's own .flow/project.md, never invoked by a
// command directly, so prose naming them keeps its repository-relative form
// legitimately — rule 3 governs INVOKED guards only.
var gsExempt = []string{
	"check-references.sh",
	"check-vocabulary.sh",
	"check-plan-provenance.sh",
	"check-task-build-green.sh",
	"check-stage-mark-calls.sh",
}

// gsDeclaredRule6 is DECLARED_RULE6 — the pairs whose requirement genuinely
// lives outside what rule 2's classifier may scan, declared here rather than
// inferred, on the KAN-197 pattern: a statement this guard's own source
// makes, that a reviewer can read and question. Declaring is for a guard
// whose invocation is real but whose citation cannot be carried in the
// skill's own files without restating a canonical text elsewhere — a
// project's own `## lint` list (resolved through .flow/project.md, which
// flow's verify stage runs wholesale) or a step whose procedure is canonical
// in a flow-contract file rule 2 deliberately does not scan. A guard with no
// such text anywhere is not declared; it is pruned. One line per pair:
// skill, guard, reason, tab-separated.
const gsDeclaredRule6 = "flow\tcheck-visual-verification.sh\tresolved through a project's own ## visual verification and ## lint configuration, never invoked by flow's own text\n" +
	"flow\tcheck-task-records.sh\tresolved through a project's own ## lint configuration, never invoked by flow's own text\n" +
	"flow\tcheck-visual-verify-dispatched.sh\tinvoked per Run 1 (finish-contract-run1.md), which rule 2's scope deliberately does not scan\n"

var (
	// FIXED_DEPTH_ROOT_RE — THREE SPELLINGS OF THE SAME DEFECT, not one.
	// `$SCRIPT_DIR/..` was the only shape matched originally, and it is the
	// only one that was ever fixed by name — but `REPO_ROOT="$(dirname
	// "$SCRIPT_DIR")"` and `cd "$SCRIPT_DIR" && cd ..` (or `; cd ..`) answer
	// the identical question the identical wrong way, and neither is
	// textually `$SCRIPT_DIR/..`. A guard passing rule 4 by spelling the same
	// fixed-depth walk differently is rule 4 defeated by the bug it exists
	// to stop, so all three are matched here.
	gsFixedDepthRoot = regexp.MustCompile(`\$\{?SCRIPT_DIR\}?"?/\.\.|dirname([ \t]+--)?[ \t]+"?\$\{?SCRIPT_DIR\}?"?|cd[ \t]+"?\$\{?SCRIPT_DIR\}?"?"?[ \t]*(&&|;)[ \t]*cd[ \t]+\.\.`)
	gsSibling        = regexp.MustCompile(`\$\{?SCRIPT_DIR\}?"?/[A-Za-z0-9._-]+`)
	// Anchored to a full word — "bash", "sh" or "zsh" followed by whitespace
	// or end of string — never a prefix match. Unanchored, this also matched
	// "shell" and "shellsession" by accident (both start with "sh"), which
	// were never a declared member of the set this guard scans; "console" and
	// "text" are deliberately still excluded, since every real invocation in
	// this repository is written in a bash/sh/zsh fence and nowhere else.
	gsShellFence   = regexp.MustCompile(`^(bash|sh|zsh)([ \t]|$)`)
	gsRule3Fence   = regexp.MustCompile(`scripts/[A-Za-z0-9._/-]+`)
	gsRule3Prose   = regexp.MustCompile("(Run|run|Invoke|invoke|Execute|execute)[ \t]+`scripts/[A-Za-z0-9._/-]+")
	gsNameRun      = regexp.MustCompile(`[A-Za-z0-9._-]+`)
	gsPlaceholder  = regexp.MustCompile(`^[ \t]+<`)
	gsBlanks       = regexp.MustCompile(`[ \t]+`)
	gsInvokingWord = regexp.MustCompile(`(run|invoke|invocation|invoking|execute)`)
	gsBlankLine    = regexp.MustCompile(`^[ \t]*$`)
	gsDelegate     = regexp.MustCompile("`/flow[A-Za-z0-9_-]*`")
)

// gsRow is one REQUIRED_FILE row: skill requires guard, first cited at
// file:line — line "sibling" for a sibling dependency, via naming the
// delegate a delegated row came through.
type gsRow struct{ skill, guard, file, line, via string }

type gsScan struct {
	env                   Env
	root, skills, scripts string
	violations            strings.Builder
}

// violation — every finding goes through here, so the report format has
// exactly one place to drift from.
func (s *gsScan) violation(path, line, msg string) {
	fmt.Fprintf(&s.violations, "%s:%s: %s\n", path, line, msg)
}

// gsRefusal is a cannot-answer: its text goes to stderr, nothing to stdout.
type gsRefusal string

func (r gsRefusal) Error() string { return string(r) }

func gsRefuse(format string, a ...any) error {
	return gsRefusal(fmt.Sprintf("check-guard-symlinks: "+format+"\n", a...))
}

func checkGuardSymlinks(_ []string, env Env, stdout, stderr io.Writer) int {
	// The scan set is this repository's own skills/ and scripts/, from the
	// checkout the shim sits in (FLOW_GUARD_REPO_ROOT, derived as the bash
	// derived ROOT from its own location). "One level above the script" is
	// safe here: this is a project-configured guard, never itself symlinked
	// into any skill's scripts/, so it has exactly one home.
	// CHECK_GUARD_SYMLINKS_ROOT is the opt-in test override, honoured only
	// when set.
	root, set := env.LookupEnv("CHECK_GUARD_SYMLINKS_ROOT")
	switch {
	case set && root == "":
		fmt.Fprint(stderr, "check-guard-symlinks: CHECK_GUARD_SYMLINKS_ROOT is set but empty\n")
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprint(stderr, "check-guard-symlinks: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-guard-symlinks.sh, which sets it\n")
			return 2
		}
	}
	s := &gsScan{env: env, root: root, skills: root + "/skills", scripts: root + "/scripts"}
	out, err := s.run()
	if err != nil {
		fmt.Fprint(stderr, err.Error())
		return 2
	}
	fmt.Fprint(stdout, out)
	if s.violations.Len() > 0 {
		return 1
	}
	return 0
}

// gsReadableDir is `[ -d p ] && [ -r p ] && [ -x p ]`.
func gsReadableDir(p string) bool { return isDir(p) && syscall.Access(p, 4|1) == nil }

func gsReadable(p string) bool { return syscall.Access(p, 4) == nil }

func gsIsSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func gsExists(p string) bool { _, err := os.Stat(p); return err == nil }

// gsFindErr is find(1)'s own message for a directory it could not read.
func gsFindErr(p string, err error) string {
	var errno syscall.Errno
	msg := err.Error()
	if errors.As(err, &errno) {
		msg = errno.Error()
		msg = strings.ToUpper(msg[:1]) + msg[1:]
	}
	return "find: " + p + ": " + msg + "\n"
}

// gsList is `find <dir> -mindepth 1 -maxdepth 1 -print | sort`, keep
// choosing the entries: each spelled <dir>/<name> as find prints it, in the
// order of the `sort` on PATH under the caller's locale — the order the
// report's lines carry. A failure is find's, the refusal naming what.
func (s *gsScan) gsList(dir, what string, keep func(os.DirEntry) bool) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, gsRefusal(fmt.Sprintf("check-guard-symlinks: could not %s\n%s", what, gsFindErr(dir, err)))
	}
	var paths []string
	for _, e := range ents {
		if keep == nil || keep(e) {
			paths = append(paths, dir+"/"+e.Name())
		}
	}
	return s.sort(paths, false)
}

func (s *gsScan) sort(lines []string, unique bool) ([]string, error) {
	sorted, err := crSort(s.env, lines, unique)
	if err != nil {
		return nil, gsRefuse("cannot sort: %v", err)
	}
	return sorted, nil
}

func (s *gsScan) run() (string, error) {
	switch {
	case !gsReadableDir(s.root):
		return "", gsRefuse("%s is not a readable directory — cannot scan", s.root)
	case !gsReadableDir(s.skills):
		return "", gsRefuse("%s is not a readable directory — cannot scan", s.skills)
	case !gsReadableDir(s.scripts):
		return "", gsRefuse("%s is not a readable directory — cannot scan", s.scripts)
	}

	// Command skills: every directory directly under skills/ except
	// flow-contracts, which is shared prose loaded by several commands and
	// SHALL NOT carry a scripts/ directory of its own (task 3's resolution
	// rule). Read from the real tree rather than hardcoded, so a new command
	// skill is covered the moment it is added.
	commandSkills, err := s.gsList(s.skills, "enumerate "+s.skills, func(e os.DirEntry) bool {
		return e.IsDir() && e.Name() != "flow-contracts"
	})
	if err != nil {
		return "", err
	}
	skillNames := make([]string, len(commandSkills))
	for i, d := range commandSkills {
		skillNames[i] = d[strings.LastIndex(d, "/")+1:]
	}
	// allSkillDirs — every directory directly under skills/, WITHOUT the
	// flow-contracts exclusion commandSkills applies. That exclusion is
	// correct for rules 1-4 and for coverage, which are about *command*
	// skills and their scripts/ directories. Rule 5's own requirement is about
	// EVERY skill directory's own top level, flow-contracts included:
	// excluding it left a symlink placed directly under skills/flow-contracts/
	// invisible to this guard entirely (pass 2, finding D). Used by rule 5
	// alone — it must never widen what rules 1-4 or coverage treat as a
	// command skill.
	allSkillDirs, err := s.gsList(s.skills, "enumerate "+s.skills, func(e os.DirEntry) bool { return e.IsDir() })
	if err != nil {
		return "", err
	}

	// The guard set — every basename directly under the repository's real
	// scripts/ directory (a file OR a directory, so "lib" counts). This is the
	// vocabulary rule 2's citation scan matches against: a guard a skill's
	// text cites but which is NOT YET shipped anywhere is exactly the
	// regression rule 2 exists to catch, so the set has to include every REAL
	// guard, not only the ones already found symlinked in somewhere.
	guards := map[string]bool{}
	ents, err := os.ReadDir(s.scripts)
	if err != nil {
		return "", gsRefuse("could not enumerate %s", s.scripts)
	}
	for _, e := range ents {
		if e.Name() != "__pycache__" {
			guards[e.Name()] = true
		}
	}

	realTargets, err := s.rule1(commandSkills, guards)
	if err != nil {
		return "", err
	}
	s.rule4(realTargets)
	if err := s.rule5(allSkillDirs); err != nil {
		return "", err
	}
	if err := s.rule3(); err != nil {
		return "", err
	}
	required, err := s.rule2(commandSkills, skillNames, guards)
	if err != nil {
		return "", err
	}
	if err := s.rule6(commandSkills, required); err != nil {
		return "", err
	}
	cov, err := s.coverage(skillNames, required)
	if err != nil {
		return "", err
	}

	if s.violations.Len() > 0 {
		v := s.violations.String()
		return fmt.Sprintf("%sGUARD-SYMLINKS-INVALID: %s — %d violation(s)\n", v, s.root, strings.Count(v, "\n")), nil
	}
	out := fmt.Sprintf("GUARD-SYMLINKS-OK: %s — %d guard(s) across %d skill(s) validated\n", s.root, len(guards), len(skillNames))
	if frag := cov.report(); frag != "" {
		out += "  " + frag + "\n"
	}
	return out, nil
}

// rule1 — every entry under skills/*/scripts/ is a symlink, it resolves, and
// its target is relative; `__pycache__` is skipped (Python writes it beside a
// .py it imports through the symlinked path, and it is gitignored). Adds every
// entry's basename to guards, and returns the unique resolved sources rule 4
// reads, in `sort -u` order.
func (s *gsScan) rule1(commandSkills []string, guards map[string]bool) ([]string, error) {
	var realTargets []string
	for _, skillDir := range commandSkills {
		scriptsDir := skillDir + "/scripts"
		if !gsExists(scriptsDir) {
			continue
		}
		if !gsReadableDir(scriptsDir) {
			return nil, gsRefuse("%s is not a readable directory — cannot scan", scriptsDir)
		}
		entries, err := s.gsList(scriptsDir, "list "+scriptsDir, nil)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			base := entry[strings.LastIndex(entry, "/")+1:]
			if base == "__pycache__" {
				continue
			}
			guards[base] = true
			if !gsIsSymlink(entry) {
				s.violation(entry, "0", "not a symlink — every entry under skills/*/scripts/ must be a relative symlink into the repository's scripts/ directory (rule 1)")
				continue
			}
			target, _ := os.Readlink(entry)
			if strings.HasPrefix(target, "/") {
				// Reported once, here, and not fed on to rule 4: an absolute
				// (off-repo) target is ALREADY a rule 1 violation on its own,
				// and falling through would have rule 4 read a file outside
				// the repository this guard was asked to scan (F9).
				s.violation(entry, "0", "symlink target is absolute ('"+target+"') — every entry under skills/*/scripts/ must be a RELATIVE symlink, or it bakes this machine's checkout path into the repository (rule 1)")
				continue
			}
			if !gsExists(entry) {
				s.violation(entry, "0", "symlink does not resolve (target '"+target+"') (rule 1)")
				continue
			}
			real, ok := resolveFile(entry)
			if !ok {
				s.violation(entry, "0", "cannot resolve this symlink's real physical location (rule 1)")
				continue
			}
			if isFile(real) {
				realTargets = append(realTargets, real)
			}
		}
	}
	return s.sort(realTargets, true)
}

// rule4 — no guard actually shipped derives a repository root as a fixed
// number of levels above $SCRIPT_DIR. Scoped to rule 1's resolved targets,
// what is ACTUALLY symlinked in — never a hardcoded guard list — so a
// project-configured guard that legitimately keeps the old `$SCRIPT_DIR/..`
// form (it is never reachable from anywhere but its own scripts/ directory)
// is correctly left alone.
func (s *gsScan) rule4(realTargets []string) {
	for _, real := range realTargets {
		if !gsReadable(real) {
			s.violation(real, "0", "cannot read this guard's real source to check for a fixed-depth root derivation (rule 4)")
			continue
		}
		b, err := os.ReadFile(real)
		if err != nil {
			s.violation(real, "0", "grep exited 2 while scanning for a fixed-depth root derivation — a failure to look, not an absence (rule 4)")
			continue
		}
		for i, line := range lines(b) {
			if gsFixedDepthRoot.MatchString(line) {
				s.violation(real, fmt.Sprint(i+1), "derives a repository root as $SCRIPT_DIR/.. — this guard is reachable from more than one directory once shipped, and a fixed number of levels above itself silently answers a skill directory instead of the repository root (rule 4)")
			}
		}
	}
}

// rule5 — no skill directory carries a symlink directly at its own top level.
// Scanned over EVERY skill directory, flow-contracts included (pass 2,
// finding D): the requirement is about placement inside "the only place a
// skill's symlinks may live" for every skill directory. Only the skill
// directory's own top level is listed — an entry named "scripts" is always a
// directory, never a symlink, and anything under scripts/ stays rule 1's.
func (s *gsScan) rule5(allSkillDirs []string) error {
	for _, skillDir := range allSkillDirs {
		entries, err := s.gsList(skillDir, "list "+skillDir, nil)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if gsIsSymlink(entry) {
				target, _ := os.Readlink(entry)
				s.violation(entry, "0", "symlink directly under a skill directory (target '"+target+"') — every symlink a skill carries must sit under skills/<skill>/scripts/, never at the skill directory's own top level (rule 5)")
			}
		}
	}
	return nil
}

// gsScanMD walks a markdown file's lines through the fence state both
// classifiers share: a line opening a ``` fence records its length and
// language; inside a fence, a line of at least as many backticks and nothing
// else closes it, and a line of a bash/sh/zsh fence goes to fenced; outside
// every fence, a line goes to prose. Both get the raw line and its 1-based
// number.
func gsScanMD(b []byte, prose, fenced func(raw string, n int)) {
	fenceLen, fenceLang := 0, ""
	for i, raw := range lines(b) {
		raw, _, _ = strings.Cut(raw, "\x00") // awk's record ends at a NUL
		line := strings.TrimLeft(raw, " \t")
		ticks := len(line) - len(strings.TrimLeft(line, "`"))
		switch {
		case fenceLen == 0 && ticks >= 3:
			fenceLen, fenceLang = ticks, strings.Trim(line[ticks:], " \t")
		case fenceLen == 0:
			prose(raw, i+1)
		case ticks >= fenceLen && strings.TrimRight(line[ticks:], " \t") == "":
			fenceLen, fenceLang = 0, ""
		case gsShellFence.MatchString(fenceLang):
			fenced(raw, i+1)
		}
	}
}

// gsFindMD is `find <dir> -type f -name '*.md'`: every regular .md file
// under dir, no symlink followed, each spelled as find prints it; an
// unreadable directory is find's message, collected in errs.
func gsFindMD(dir string, depth1 bool, files *[]string, errs *strings.Builder) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		errs.WriteString(gsFindErr(dir, err))
	}
	for _, e := range ents {
		p := dir + "/" + e.Name()
		switch {
		case e.IsDir() && !depth1:
			gsFindMD(p, false, files, errs)
		case e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".md"):
			*files = append(*files, p)
		}
	}
}

// rule3 — no skill text carries a repository-relative scripts/<name> path in
// an invoking position. Scanned across every .md file under skills/,
// fence-aware: a non-comment line inside a bash/sh/zsh fence, or a
// "Run"/"Invoke"/"Execute" immediately before the backtick, is an
// invocation; everything else — including a plain fenced example with no such
// language tag, and a sentence describing what a guard reads or does — is
// prose and is left alone. The offending line's own text is never echoed into
// the report: this repository's skill text is tracked and editable in any
// pull request, so the report names path and line only, never the
// attacker-influenced content.
func (s *gsScan) rule3() error {
	var files []string
	var errs strings.Builder
	gsFindMD(s.skills, false, &files, &errs)
	if errs.Len() > 0 {
		return gsRefusal("check-guard-symlinks: could not enumerate .md files under " + s.skills + "\n" + errs.String())
	}
	files, err := s.sort(files, false)
	if err != nil {
		return err
	}
	// isExempt keys on the guard name after `scripts/`, up to the next `/`.
	isExempt := func(span string) bool {
		name, _, _ := strings.Cut(strings.TrimPrefix(span, "scripts/"), "/")
		return slices.Contains(gsExempt, name)
	}
	for _, md := range files {
		b, err := gsReadMD(md, "rule 3")
		if err != nil {
			return err
		}
		var hits []int
		gsScanMD(b, func(raw string, n int) {
			if m := gsRule3Prose.FindString(raw); m != "" && !isExempt(m[strings.LastIndex(m, "`")+1:]) {
				hits = append(hits, n)
			}
		}, func(raw string, n int) {
			trimmed := strings.TrimLeft(raw, " \t")
			if strings.HasPrefix(trimmed, "#") {
				return
			}
			if m := gsRule3Fence.FindString(trimmed); m != "" && !isExempt(m) {
				hits = append(hits, n)
			}
		})
		for _, n := range hits {
			s.violation(md, fmt.Sprint(n), "a repository-relative scripts/<name> path appears in an invoking position (a bash-fenced command, or an imperative Run/Invoke/Execute) — name the guard by basename instead, per the resolution rule in skills/flow-contracts/pipeline.md (rule 3)")
		}
	}
	return nil
}

// gsReadMD reads a skill .md whole, refusing when it cannot.
func gsReadMD(md, rule string) ([]byte, error) {
	if !gsReadable(md) {
		return nil, gsRefuse("cannot read %s (%s)", md, rule)
	}
	b, err := os.ReadFile(md)
	if err != nil {
		return nil, gsRefuse("cannot read %s (%s)", md, rule)
	}
	return b, nil
}

// gsCitations is the rule 2 classifier over one .md: every known guard it
// invokes, and every invoking-position .sh-shaped name matching no known
// guard (KAN-530 F6), each with its line, in line order.
func gsCitations(b []byte, guards map[string]bool) (cited, unknown [][2]string) {
	gsScanMD(b, func(raw string, n int) {
		// A backtick span whose leading name-run is followed by a
		// `<placeholder>` usage argument, or which a "run"/"invoke"/
		// "invocation"/"invoking"/"execute" in the four words before it
		// names, is an invocation.
		pos := 0
		for {
			m := strings.IndexByte(raw[pos:], '`')
			if m < 0 {
				break
			}
			start := pos + m + 1
			m2 := strings.IndexByte(raw[start:], '`')
			if m2 < 0 {
				break
			}
			span := raw[start : start+m2]
			if loc := gsNameRun.FindStringIndex(span); loc != nil && loc[0] == 0 {
				base := span[:loc[1]]
				matched := gsPlaceholder.MatchString(span[loc[1]:])
				if !matched {
					words := gsBlanks.Split(raw[:start-1], -1)
					tail := ""
					for _, w := range words[max(0, len(words)-4):] {
						tail += " " + gsASCIILower(w)
					}
					matched = gsInvokingWord.MatchString(tail)
				}
				switch {
				case matched && guards[base]:
					cited = append(cited, [2]string{base, fmt.Sprint(n)})
				case matched && strings.HasSuffix(base, ".sh"):
					// KAN-530 F6: the same flag the fence scan carries — an
					// invoking-position name matching no known guard is a
					// typoed basename the required set used to lose silently.
					// The .sh shape keeps ordinary invoking prose naming no
					// guard at all (Run `flow stage begin`, Run `git push`) in
					// the prose bucket, where it always lived.
					unknown = append(unknown, [2]string{base, fmt.Sprint(n)})
				}
			}
			// Advance to just past the OPENING backtick of THIS attempt, not
			// past both. A stray, unpaired backtick earlier in the line makes
			// `start` land on that stray mark and `m2` land on the NEXT real
			// backtick — the one that was meant to OPEN the following
			// citation. Skipping past both would consume that backtick as a
			// CLOSER and drop the real citation silently; resyncing to `start`
			// retries from the very next backtick as a fresh opening
			// candidate, so one stray mark costs at most one empty non-match
			// rather than a real citation (F7).
			pos = start
		}
	}, func(raw string, n int) {
		trimmed := strings.TrimLeft(raw, " \t")
		if strings.HasPrefix(trimmed, "#") {
			return
		}
		// A guard named as ANY token of a fenced command line is an
		// invocation shape — the leading-token-only scan missed three real
		// ones: a pipeline segment (`git diff ... | check-visual-trigger.sh
		// <worktree>`), an `&&` continuation line whose first token is the
		// operator, and command substitution (`BASE="$(resolve-base-branch.sh
		// <worktree>)"`), where the basename is glued to the assignment and
		// only a mid-token match sees it. Each name-character run is
		// therefore tried against the guard set, and EVERY guard a line names
		// is emitted. KAN-530 F6: a name-run matching no known guard is
		// flagged when .sh-shaped; anything not .sh-shaped (git, flow, echo)
		// stays prose. A run BEGINNING with a dot is a path fragment, never a
		// basename — the `.` of `./guard.sh` and the `.sh` of a `*.sh` glob
		// are not guard names — so only dot-led runs are spared the flag.
		for _, w := range gsNameRun.FindAllString(trimmed, -1) {
			switch {
			case guards[w]:
				cited = append(cited, [2]string{w, fmt.Sprint(n)})
			case strings.HasSuffix(w, ".sh") && !strings.HasPrefix(w, "."):
				unknown = append(unknown, [2]string{w, fmt.Sprint(n)})
			}
		}
	})
	return cited, unknown
}

// gsASCIILower is awk's tolower in the byte-oriented awk the bash ran.
func gsASCIILower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

// gsDelegates is DELEGATE_AWK: the slash-commands a "**Check guard
// presence.**" paragraph names. For a skill that invokes no guard of its own
// and instead CHAINS another command's stages verbatim (the shape flow-fast
// had when this was written), its required set has to come from somewhere
// other than its own text, or rule 2 checks nothing for it at all. A bare
// citation of another command is not a safe delegation signal — command
// skills cite other skills' sections constantly without invoking anything in
// them (flow-status cites `/flow` by slash-command repeatedly, purely in
// explanatory prose). What IS safe: every command skill that invokes at
// least one guard of its own states so in a paragraph beginning "**Check
// guard presence.**"; a delegating skill's OWN copy of that paragraph is the
// one place in its text that names OTHER commands' presence checks by
// slash-command, so the scan is scoped to THAT paragraph, which ends at the
// first blank line. A command with no such paragraph names no delegate.
func gsDelegates(b []byte) []string {
	var out []string
	inPara := false
	for _, line := range lines(b) {
		line, _, _ = strings.Cut(line, "\x00") // awk's record ends at a NUL
		if strings.HasPrefix(line, "**Check guard presence.**") {
			inPara = true
		}
		if !inPara {
			continue
		}
		if gsBlankLine.MatchString(line) {
			inPara = false
			continue
		}
		for _, m := range gsDelegate.FindAllString(line, -1) {
			out = append(out, m[2:len(m)-1])
		}
	}
	return out
}

// rule2 — every guard invoked in a skill's own text has a symlink in that
// skill's own scripts/ directory, sibling dependencies included. Level-0
// citations come from the skill's own *.md files; a required guard's
// siblings are then derived from THAT GUARD'S OWN SOURCE (a $SCRIPT_DIR/<name>
// match), never from a hardcoded table — task 2's map, checked rather than
// trusted. Returns REQUIRED_UNIQUE: one row per (skill, guard), at its
// first-discovered citation.
//
// WHY RULE 2 IS SCOPED TO A SKILL'S OWN DIRECTORY is the script header's; in
// short, scanning every contract a skill's text cites was tried and produced
// false positives on this repository's own correct tree.
func (s *gsScan) rule2(commandSkills, skillNames []string, guards map[string]bool) ([]gsRow, error) {
	var required []gsRow
	var delegates []string
	for i, skillDir := range commandSkills {
		skill := skillNames[i]
		var mds []string
		var errs strings.Builder
		gsFindMD(skillDir, true, &mds, &errs)
		if errs.Len() > 0 {
			return nil, gsRefusal("check-guard-symlinks: could not enumerate .md files under " + skillDir + "\n" + errs.String())
		}
		mds, err := s.sort(mds, false)
		if err != nil {
			return nil, err
		}
		for _, md := range mds {
			b, err := gsReadMD(md, "rule 2")
			if err != nil {
				return nil, err
			}
			cited, unknown := gsCitations(b, guards)
			for _, c := range cited {
				required = append(required, gsRow{skill: skill, guard: c[0], file: md, line: c[1]})
			}
			// KAN-530 F6 — an invoked basename matching no known guard is a
			// violation at its citing line, through the ordinary violation
			// channel and the ordinary exit 1, exactly like a missing symlink
			// for a known one: the point is that a typo'd guard name fails
			// the run instead of quietly shrinking this skill's coverage
			// count. Each row is reported once, at its own citing file — the
			// bash's per-file drain of its scratch file, here a per-file slice.
			for _, u := range unknown {
				s.violation(md, u[1], u[0]+" is invoked here, but no guard named "+u[0]+" exists in this repository's scripts/ — a typo'd guard basename silently drops out of the required set; fix the name or ship the guard (rule 2)")
			}
			for _, d := range gsDelegates(b) {
				if d != skill {
					delegates = append(delegates, skill+"\t"+d)
				}
			}
		}
	}
	delegates, err := s.sort(delegates, true)
	if err != nil {
		return nil, err
	}

	// Resolve one level of delegation. required holds ONLY level-0 (directly
	// cited) rows here, which is exactly what a delegate's "own" required set
	// means — a transitive delegate-of-a-delegate does not exist among
	// today's command skills, and a second round would be speculative
	// generality for a case nothing exercises.
	var delegated []gsRow
	for _, pair := range delegates {
		skill, delegate, _ := strings.Cut(pair, "\t")
		if !isDir(s.skills + "/" + delegate) {
			continue
		}
		for _, r := range required {
			if r.skill == delegate {
				delegated = append(delegated, gsRow{skill, r.guard, r.file, r.line, delegate})
			}
		}
	}
	required = append(required, delegated...)

	// Sibling closure — read from each required guard's OWN source, one round
	// per frontier level, capped so a cycle in $SCRIPT_DIR references cannot
	// spin this guard forever.
	for _, skill := range skillNames {
		var frontier []string
		for _, r := range required {
			if r.skill == skill {
				frontier = append(frontier, r.guard)
			}
		}
		frontier, err := s.sort(frontier, true)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, g := range frontier {
			seen[g] = true
		}
		for round := 1; len(frontier) > 0; round++ {
			if round > 40 {
				return nil, gsRefuse("sibling-dependency closure for %s did not terminate within 40 rounds — a cycle in $SCRIPT_DIR references", skill)
			}
			var next []string
			for _, g := range frontier {
				real := s.scripts + "/" + g
				if !isFile(real) {
					continue
				}
				if !gsReadable(real) {
					return nil, gsRefuse("cannot read %s to derive sibling dependencies (rule 2)", real)
				}
				b, err := os.ReadFile(real)
				if err != nil {
					return nil, gsRefuse("grep exited 2 scanning %s for sibling dependencies (rule 2)", real)
				}
				for _, m := range gsSibling.FindAllString(string(b), -1) {
					sib := m[strings.LastIndex(m, "/")+1:]
					if !seen[sib] {
						seen[sib] = true
						next = append(next, sib)
						required = append(required, gsRow{skill, sib, real, "sibling", ""})
					}
				}
			}
			frontier = next
		}
	}

	// One violation per (skill, guard) pair, at its first-discovered citation.
	var unique []gsRow
	pairs := map[[2]string]bool{}
	for _, r := range required {
		if k := [2]string{r.skill, r.guard}; !pairs[k] {
			pairs[k] = true
			unique = append(unique, r)
		}
	}
	for _, r := range unique {
		if _, err := os.Lstat(s.skills + "/" + r.skill + "/scripts/" + r.guard); err == nil {
			continue
		}
		carry := "carries no symlink at skills/" + r.skill + "/scripts/" + r.guard + " (rule 2)"
		switch {
		case r.line == "sibling":
			s.violation(r.file, "0", r.guard+" is a sibling dependency this guard resolves from its own directory, but skill \""+r.skill+"\" "+carry)
		case r.via != "":
			s.violation(r.file, r.line, "invokes "+r.guard+" here, and skill \""+r.skill+"\" delegates to \""+r.via+"\"'s own guard presence check (task 4's convention), but "+carry)
		default:
			s.violation(r.file, r.line, "invokes "+r.guard+" here, but skill \""+r.skill+"\" "+carry)
		}
	}
	return unique, nil
}

// rule6 — the reverse of rule 2: a command skill must not carry a *.sh
// symlink its required set does not name. Rule 2 only ever flags a citation
// with no symlink; this is what turns a skill's prose guard list into a
// checked one: flow-plan's invoking paragraph named four guards nothing
// cited, their symlinks were carried on memory alone, and dropping a basename
// moved only an informational coverage count (KAN-532, F10). Scoped to
// entries rule 1 PASSES: a non-symlink, a dangling symlink and an
// absolute-target symlink are rule 1's findings alone, and reporting them
// here too would blur which rule to fix them under. Only *.sh entries are
// judged — the runnable guards every invoking paragraph names; the lib/
// directory symlink and the .py twins are rule 1's to validate and no prose
// list's to declare.
func (s *gsScan) rule6(commandSkills []string, required []gsRow) error {
	pairs := map[string]bool{}
	for _, r := range required {
		pairs[r.skill+"\t"+r.guard] = true
	}
	for _, skillDir := range commandSkills {
		skill := skillDir[strings.LastIndex(skillDir, "/")+1:]
		scriptsDir := skillDir + "/scripts"
		if !gsExists(scriptsDir) {
			continue
		}
		entries, err := s.gsList(scriptsDir, "list "+scriptsDir+" for rule 6", func(e os.DirEntry) bool {
			return strings.HasSuffix(e.Name(), ".sh")
		})
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !gsIsSymlink(entry) {
				continue
			}
			if target, _ := os.Readlink(entry); strings.HasPrefix(target, "/") || !gsExists(entry) {
				continue
			}
			base := entry[strings.LastIndex(entry, "/")+1:]
			// A substring match on the declaration line, as the bash's
			// `grep -F` of the pair was.
			if pair := skill + "\t" + base; strings.Contains(gsDeclaredRule6, pair) || pairs[pair] {
				continue
			}
			s.violation(entry, "0", "carried but required by nothing — no citation in this skill's own text, no delegation and no sibling dependency names "+base+", so this symlink is dead weight the next cleanup will prune (rule 6)")
		}
	}
	return nil
}

// coverage — per-skill count of what rule 2's required set actually holds
// (coverage.go). This is the layer that makes KAN-73's own defect impossible
// to ship silently again: a skill whose required set resolves to empty —
// whether it legitimately invokes no guard, or a future citation shape the
// classifier cannot see — is named on a PASSING run instead of reading as
// nothing to check; an undeclared zero is a violation.
//
// Declared here, never inferred from the tree: a member's declaration is a
// statement this guard's own source makes, that a reviewer can read and
// question. declare_if_present: flow-settings is declared ONLY when it is a
// member of THIS run's corpus — a sandboxed CHECK_GUARD_SYMLINKS_ROOT fixture
// without it would otherwise report a KAN-197 F3 "declared but never
// recorded" violation that is only a mismatch between this guard's real-repo
// names and a smaller tree.
func (s *gsScan) coverage(skillNames []string, required []gsRow) (*coverage, error) {
	var cov coverage
	for _, skill := range skillNames {
		n := 0
		for _, r := range required {
			if r.skill == skill {
				n++
			}
		}
		if err := cov.record(skill, n); err != nil {
			return nil, gsRefusal(fmt.Sprintf("%v\ncheck-guard-symlinks: coverage_record failed for skill '%s' (see stderr above)\n", err, skill))
		}
	}
	if slices.Contains(skillNames, "flow-settings") {
		if err := cov.declare("flow-settings", "invokes no guard — a standalone settings command that only calls the flow CLI, with no implementation or verification stage"); err != nil {
			return nil, gsRefusal(fmt.Sprintf("%v\ncheck-guard-symlinks: coverage_declare failed for 'flow-settings' (see stderr above)\n", err))
		}
	}
	for _, line := range cov.verdict() {
		member, _, _ := strings.Cut(line, ":")
		msg := line
		if _, after, ok := strings.Cut(line, ": "); ok {
			msg = after
		}
		s.violation(member, "0", msg)
	}
	return &cov, nil
}
