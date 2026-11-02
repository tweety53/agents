package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The contract below moved here from the header of the deleted
// scripts/plan-class.sh shim (KAN-1058).
//
// flow-guard plan-class — the planner's mechanical classifier and rolls.
//
// Usage: flow-guard plan-class [-class <class>] <tasks.md> <repos> [<worktree> <merge-base>]
//
// Prints exactly nine lines to stdout:
//   inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no
//   class: micro|small|regular|big
//   rolls: compact N · experimental N · bundle N · effort N
//   tree: class <c> · execution inline|sdd · implementer skipped — inline|chosen
//   panel: default|compact|full · roster <slot; …> · rerun delta
//   grouping: static|free · dispatches primary+principles[+exp-<name>]
//   experimental: no slot|none available|exp-<name> · skills/flow/experimental/<name>.md · <description>[ · skipped — bundle cap]
//   extra: test-sanity+call-efficiency · roll N · sonnet|opus low
//   extra: mutation+failure-modes · roll N · sonnet|opus low
//
// The first three lines are the mechanical answer. The four tree lines are
// brainstorm-planner.md's Decide tree for the effective class — the -class
// value when given, else class_mechanical — and the rolls: compact when
// compact < 90, static grouping when bundle < 30, an experimental slot when
// experimental < 30. The execution is `sdd`, implementer `chosen`, on every
// class but micro, which is `inline` (decision sdd-below-micro). The slot is
// sorted(skills/flow/experimental/*.md) [experimental mod count] under FLOW_GUARD_REPO_ROOT, with its line-1
// `description:` value; an absent directory or no *.md file is `none
// available`. The slot joins the first dispatch of a full regular/big
// panel and is otherwise `skipped — bundle cap`. On
// micro, panel is `default` and grouping/experimental read `not
// consulted — micro`.
// The extra line is printed on every class, micro included: the always-on test-sanity+call-efficiency dispatch's pair, sonnet when roll(<name>extra) < 50, else opus, always at low (kan-959).
// The second extra line is the always-on mutation+failure-modes dispatch's pair, rolled the same way on roll(<name>extramf) (kan-1028).
//
// -class may equal class_mechanical or sit one step above it
// (micro→small→regular→big) — the planner's recorded raise; the class: line
// stays class_mechanical.
//
// Exit 0 on a printed answer, exit 2 on a missing <tasks.md>, a
// non-integer <repos>, an argument count that is neither 2 nor 4, a -class
// that is unknown, two or more steps up or any step down, an experimental
// candidate whose line 1 is not a `description:` line, or — with
// the optional arguments — a <worktree>/<merge-base> that cannot be answered
// (not a directory, not a git worktree, an unresolving merge base, or a
// failed git invocation). Rules from design.md's "Inputs and the class" and
// "The rolls" (kan-472-flow-dynamic-review-panel-roster-repo-scoped); the
// small/big thresholds were raised by kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop
// so more changes classify small/regular and roll compact, then lowered by
// roughly 25% so a plan is classed up one step sooner:
//
//   tasks     = count of column-0 `- [ ] <n>.` / `- [x] <n>.` lines
//   files     = size of the union of every task's `**Files:**` backticked
//               paths
//   repos     = the <repos> argument, verbatim — the number of distinct
//               repository roots the plan's **Files:** fall under, which the
//               planner counts (brainstorm-planner.md's Decide); this script
//               only classifies it
//   migration = any **Files:** path under stats/internal/store/migrations/
//               or ending .sql
//   spec      = any **Files:** path under spectre/specs/
//   red       = any task tagged `**Build:** red`
//   unverified= any `unverified:` provenance tag anywhere in the plan
//
//   small:   tasks<=8 and files<=19 and repos=1 and not migration and not spec
//   big:     tasks>=22 or files>=60 or (repos>1 and tasks>=11)
//            or (migration and tasks>=11)
//   regular: everything else
//
//   micro:   the small thresholds met with tasks<=2, every **Files:** path
//            documentation (.md/.mdc), no `**Build:** red` tag, and — when
//            <worktree> <merge-base> are passed — this change's own touched
//            paths (paneltouchedpaths.go's union of
//            committed-since-merge-base, staged and unstaged) entirely
//            documentation and at most planclass.go's pcMicroLineCap (20)
//            changed lines; an empty
//            touched surface passes vacuously. Both are measured from
//            `git merge-base HEAD <merge-base>`, never from <merge-base>
//            itself, so a base ref that moved after the worktree was cut
//            counts none of its own new commits (KAN-885); a <merge-base>
//            sharing no history with HEAD is exit 2. Without the optional arguments
//            micro never fires, so the two-argument form classifies exactly as
//            before it existed. Added by
//            kan-617-flow-cost-the-decide-record-ceremony-runs-full so a
//            two-line prose change stops paying the full decide ceremony
//            (brainstorm-planner.md's Decide collapses to a recorded default
//            decision on this class).
//
// The planner may raise this class one step with a recorded override — the
// judgment and its reason live in brainstorm-planner.md; this script only
// bounds the raise (-class), so `override` never appears in this script's own
// output (stats/internal/guard/plan_class_test.go asserts exactly that).
//
// Rolls are reproducible per change name (basename of the directory holding
// <tasks.md>): compact_roll = sha256("<name>") mod 100, experimental_roll =
// sha256("<name>exp") mod 100, bundle_roll = sha256("<name>bundle") mod 100,
// effort_roll = sha256("<name>effort") mod 100 (added by
// medium-effort-default), each over the first 8 hex digits of the digest read as an integer —
// hashed by stats/internal/guard/sha256.go's sha256Hex.
//
// The logic is the Go port in stats/internal/guard/planclass.go. FLOW_GUARD_REPO_ROOT, which scripts/flow-guard exports, is the root the experimental directory is read under.

// planClass is `flow-guard plan-class`: the comment above is the contract -- the inputs/class/rolls lines and the four tree lines on
// stdout, exit 0 on a printed answer, exit 2 on a refusal -- and states
// every threshold below.
// The reasoning for each step, moved here from the bash body it replaced
// (c5379c0a), sits beside the code it explains.
func init() {
	Registry["plan-class"] = planClass
}

// pcMicroLineCap is MICRO_LINE_CAP: the micro class's changed-line cap on
// the change's own touched surface.
const pcMicroLineCap = 20

var (
	pcTaskLine  = regexp.MustCompile(`^- \[[ xX]\] [0-9]+\.`)
	pcBacktick  = regexp.MustCompile("`[^`]+`")
	pcMigration = regexp.MustCompile(`^stats/internal/store/migrations/|\.sql$`)
)

// pcRoll is one reproducible roll: the first 8 hex digits of sha256(s) read
// as an integer, mod 100. plan-agent-pair rolls with it too.
func pcRoll(s string) uint64 {
	v, _ := strconv.ParseUint(sha256Hex(s)[:8], 16, 64)
	return v % 100
}

// pcRed is grep -E '^\*\*Build:\*\* red\b' in the caller's locale: under
// UTF-8 a non-ASCII letter or digit right after "red" continues the word, as
// grep's locale-aware \b has it, and so does a byte that is not UTF-8 (grep
// at c5379c0a, measured); otherwise only an ASCII one does.
func pcRed(env Env, line string) bool {
	rest, ok := strings.CutPrefix(line, "**Build:** red")
	if !ok || rest == "" {
		return ok
	}
	r, size := utf8.DecodeRuneInString(rest)
	if r >= utf8.RuneSelf {
		if !smcUTF8(env) {
			return true
		}
		if r == utf8.RuneError && size == 1 {
			return false
		}
	}
	return !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r))
}

// isDocPath is `\.mdc?$`: a documentation path.
func isDocPath(p string) bool {
	return strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdc")
}

// pcClasses is the class order -class may step up by one.
var pcClasses = []string{"micro", "small", "regular", "big"}

// pcExp is one experimental-prompt candidate: its basename without .md and
// its line-1 description: value.
type pcExp struct{ name, desc string }

// pcExtra is an always-on extra dispatch's line — test-sanity+call-efficiency
// (kan-959) or mutation+failure-modes (kan-1028): its bundle, its roll and
// the pair it picks from pairs, the bundle's own row of the allowed-pairs
// table — its sonnet pair below 50, its opus pair otherwise, either one when
// the row names only the other. Printed on every class, micro included.
func pcExtra(bundle string, roll uint64, pairs []string) string {
	want := "opus/"
	if roll < 50 {
		want = "sonnet/"
	}
	pick := pairs[0]
	for _, p := range pairs {
		if strings.HasPrefix(p, want) {
			pick = p
			break
		}
	}
	model, typ, _ := strings.Cut(pick, "/")
	return fmt.Sprintf("extra: %s · roll %d · %s %s\n", bundle, roll, model, lpEffort(typ))
}

// pcTree is the decision tree brainstorm-planner.md's Decide reads, for the
// effective class and the three rolls: execution and implementer, the panel
// roster, its grouping and dispatches, and the experimental slot. Compact
// when compact < 90, static grouping when bundle < 30, an experimental slot
// when exp < 30, picked at index exp mod len(cands) of the byte-ordered
// candidates. The roster is primary and principles alone — failure-modes and
// mutation run on an always-on extra dispatch — so the slot joins the first
// dispatch of a full regular/big panel and is otherwise skipped for the
// bundle cap. Every class but micro executes sdd with an implementer
// chosen; micro alone is inline, records defaults and consults no roll.
func pcTree(class string, compact, exp, bundle uint64, cands []pcExp) string {
	if class == "micro" {
		return "tree: class micro · execution inline · implementer skipped — inline\n" +
			"panel: default\n" +
			"grouping: not consulted — micro\n" +
			"experimental: not consulted — micro\n"
	}
	execution, implementer := "sdd", "chosen"
	shape, roster, dispatches, room := "compact", "primary; principles", "primary+principles", false
	if compact >= 90 {
		shape = "full"
		room = class != "small"
	}
	grouping := "free"
	if bundle < 30 {
		grouping = "static"
	}
	experimental := "no slot"
	if exp < 30 && len(cands) == 0 {
		experimental = "none available"
	} else if exp < 30 {
		c := cands[exp%uint64(len(cands))]
		experimental = "exp-" + c.name + " · skills/flow/experimental/" + c.name + ".md · " + c.desc
		if room {
			roster += "; exp-" + c.name
			dispatches += "+exp-" + c.name
		} else {
			experimental += " · skipped — bundle cap"
		}
	}
	return "tree: class " + class + " · execution " + execution + " · implementer " + implementer + "\n" +
		"panel: " + shape + " · roster " + roster + " · rerun delta\n" +
		"grouping: " + grouping + " · dispatches " + dispatches + "\n" +
		"experimental: " + experimental + "\n"
}

// pcCandidates reads <root>/skills/flow/experimental/*.md in byte order: an
// absent directory is no candidates; a candidate whose line 1 is not a
// description: line cannot be answered.
func pcCandidates(root string) ([]pcExp, error) {
	dir := filepath.Join(root, "skills", "flow", "experimental")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() && strings.HasSuffix(n, ".md") && !strings.HasPrefix(n, ".") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	var cands []pcExp
	for _, n := range names {
		p := filepath.Join(dir, n)
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		line, _, _ := strings.Cut(string(b), "\n")
		desc, ok := strings.CutPrefix(line, "description:")
		if !ok {
			return nil, fmt.Errorf("%s: line 1 is not a description: line", p)
		}
		cands = append(cands, pcExp{strings.TrimSuffix(n, ".md"), strings.TrimSpace(desc)})
	}
	return cands, nil
}

func planClass(args []string, env Env, stdout, stderr io.Writer) int {
	const usage = "usage: flow-guard plan-class [-class <class>] <tasks.md> <repos> [<worktree> <merge-base>]\n"
	override := ""
	if len(args) > 0 && args[0] == "-class" {
		if len(args) < 2 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		override, args = args[1], args[2:]
		if !slices.Contains(pcClasses, override) {
			fmt.Fprintf(stderr, "plan-class.sh: unknown class: %s\n", override)
			return 2
		}
	}
	if len(args) != 2 && len(args) != 4 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	tasksFile, repos := args[0], args[1]
	if !isFile(smcAbs(env, tasksFile)) {
		fmt.Fprintf(stderr, "plan-class.sh: no such file: %s\n", tasksFile)
		return 2
	}
	if repos == "" || strings.Trim(repos, "0123456789") != "" {
		fmt.Fprintf(stderr, "plan-class.sh: <repos> must be a non-negative integer, got: %s\n", repos)
		return 2
	}
	b, err := os.ReadFile(smcAbs(env, tasksFile))
	if err != nil {
		fmt.Fprintf(stderr, "plan-class.sh: cannot read %s: %v\n", tasksFile, err)
		return 2
	}

	tasks, red := 0, false
	seen := map[string]bool{}
	// Every backticked path on a **Files:** line, deduplicated across the
	// plan. The bash's `sort -u` ordered them too; only their count and
	// kinds are printed, so no order is kept.
	var files []string
	for _, l := range lines(b) {
		if pcTaskLine.MatchString(l) {
			tasks++
		}
		if pcRed(env, l) {
			red = true
		}
		if !strings.HasPrefix(l, "**Files:**") {
			continue
		}
		for _, m := range pcBacktick.FindAllString(l, -1) {
			if f := strings.Trim(m, "`"); !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	}
	migration, spec, allDocs := false, false, len(files) > 0
	for _, f := range files {
		migration = migration || pcMigration.MatchString(f)
		spec = spec || strings.HasPrefix(f, "spectre/specs/")
		allDocs = allDocs && isDocPath(f)
	}
	unverified := strings.Contains(string(b), "unverified:")

	// `[ "$REPOS" -eq 1 ]` / `-gt 1`: a value past int64 is no integer to
	// bash's test, so it is neither — the error test prints is not.
	n, err := strconv.ParseInt(repos, 10, 64)
	if err != nil {
		n = -1
	}
	class := "regular"
	if tasks <= 8 && len(files) <= 19 && n == 1 && !migration && !spec {
		class = "small"
	} else if tasks >= 22 || len(files) >= 60 || (n > 1 && tasks >= 11) || (migration && tasks >= 11) {
		class = "big"
	}

	// The change's own diff side: with the optional arguments it is
	// answered on every plan, micro-eligible or not — an unanswerable
	// <worktree>/<merge-base> is exit 2 unconditionally, exactly as the
	// header states. One numstat of the merge base against the working tree
	// covers committed, staged and unstaged together, so churn on one file
	// counts once. A numstat entry git cannot count (binary) is
	// fail-closed: never micro.
	if len(args) == 4 {
		worktree, mergebase := args[2], args[3]
		gitBin, ok := panelResolveGit(env, "plan-class", stderr)
		if !ok || !panelValidateWorktree(env, "plan-class", worktree, mergebase, gitBin, stderr) {
			return 2
		}
		// The diff side is measured from HEAD's fork point with the given
		// base, never from the base itself (KAN-885): a ref the caller
		// resolved after the base moved — origin/main fetched mid-run — is
		// not an ancestor of HEAD, and diffing against its tip counted the
		// base's own new commits as this change's surface. A recorded
		// worktree sha, already an ancestor, is its own fork point.
		fork, ok := capture(panelGit(env, gitBin, worktree, "merge-base", "HEAD", mergebase))
		if !ok {
			fmt.Fprintf(stderr, "plan-class: HEAD and '%s' share no merge base in %s\n", mergebase, worktree)
			return 2
		}
		mergebase = fork
		surface, ok := panelTouchedPaths(env, "plan-class", worktree, mergebase, gitBin, stderr)
		if !ok {
			return 2
		}
		cmd := panelGit(env, gitBin, worktree, append([]string{"diff", "--no-renames", "--numstat", "--end-of-options", mergebase}, wpdExcludes(smcAbs(env, worktree))...)...)
		cmd.Stderr = stderr
		numstat, err := cmd.Output()
		if err != nil {
			return 2
		}
		changed := 0
		for _, l := range lines(numstat) {
			f := strings.Fields(l)
			if len(f) < 2 {
				continue
			}
			if f[0] == "-" || f[1] == "-" {
				changed = pcMicroLineCap + 1
				break
			}
			a, _ := strconv.Atoi(f[0])
			d, _ := strconv.Atoi(f[1])
			changed += a + d
		}
		// The micro class: the plan side is decidable from the tasks.md
		// alone; the diff side must have been answered too. An empty
		// touched surface (a creating run, nothing implemented yet) passes
		// vacuously; a non-empty one must itself be docs-only and within
		// the cap.
		surfaceDocs := true
		for _, p := range surface {
			surfaceDocs = surfaceDocs && isDocPath(p)
		}
		if class == "small" && tasks <= 2 && !red && allDocs && surfaceDocs && changed <= pcMicroLineCap {
			class = "micro"
		}
	}

	effective := class
	if override != "" {
		step := slices.Index(pcClasses, override) - slices.Index(pcClasses, class)
		if step != 0 && step != 1 {
			fmt.Fprintf(stderr, "plan-class.sh: -class %s must be %s or one step above it\n", override, class)
			return 2
		}
		effective = override
	}
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		fmt.Fprint(stderr, "plan-class.sh: FLOW_GUARD_REPO_ROOT is unset — run flow-guard plan-class, which sets it\n")
		return 2
	}
	cands, err := pcCandidates(root)
	if err != nil {
		fmt.Fprintf(stderr, "plan-class.sh: %v\n", err)
		return 2
	}
	policy, err := policyTable(env, root)
	for _, role := range []string{"panel-sanity-efficiency", "panel-mutation-failure"} {
		if err != nil || len(policy[role]) == 0 {
			fmt.Fprintf(stderr, "plan-class.sh: the allowed-pairs table names no %s row (%v)\n", role, err)
			return 2
		}
	}

	// Rolls are reproducible per change name (the directory holding
	// <tasks.md>): each the first 8 hex digits of a SHA-256, mod 100.
	name := uwBasename(gdcDirname(tasksFile))
	roll := pcRoll
	yn := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	fmt.Fprintf(stdout, "inputs: tasks=%d files=%d repos=%s migration=%s spec=%s red=%s unverified=%s\n",
		tasks, len(files), repos, yn(migration), yn(spec), yn(red), yn(unverified))
	fmt.Fprintf(stdout, "class: %s\n", class)
	fmt.Fprintf(stdout, "rolls: compact %d · experimental %d · bundle %d · effort %d\n", roll(name), roll(name+"exp"), roll(name+"bundle"), roll(name+"effort"))
	fmt.Fprint(stdout, pcTree(effective, roll(name), roll(name+"exp"), roll(name+"bundle"), cands))
	fmt.Fprint(stdout, pcExtra("test-sanity+call-efficiency", roll(name+"extra"), policy["panel-sanity-efficiency"][0].Pairs))
	fmt.Fprint(stdout, pcExtra("mutation+failure-modes", roll(name+"extramf"), policy["panel-mutation-failure"][0].Pairs))
	return 0
}
