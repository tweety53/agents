// Package selfreview assembles the self-review context bundle a change's
// reasoning pass reads — the sources run 1's bundle step judges a change
// by, served from the store and the change's own worktree instead of
// gathered from files a Bash script had to be pointed at.
//
// The change directory is gitignored and never committed (KAN-1073), so its
// files are read from disk in the worktree the caller names, and the
// finish-run commits are that worktree's first-parent merge-base..HEAD
// range; no path any caller supplies is resolved against this process's
// working directory, and the bundle's section labels carry no worktree path.
package selfreview

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// changeBranchPrefix names a /flow change's branch, spectre/<name>: its
// recorded base (branch.<branch>.flowBase) bounds the finish range, and a
// peer's branch of that name carries its share of a cross-repo change.
const changeBranchPrefix = "spectre/"

// liveDir is the change directories' parent under the repository's spectre
// tree — the label root of every file the bundle serves from the change. The Bash gather
// resolved the spec root leaf per repository (the since-retired
// scripts/lib/spec-root.sh, twinned in stats/internal/guard/specroot.go);
// the app pins `spectre`, this repository's leaf — a second spec root would
// be a store schema question first, not a string to probe for.
const liveDir = "spectre/changes"

// Runner is the git access the assembler needs. Every call is
// `git -C <repo> <args...>`; output is the command's stdout, an error its
// non-zero exit.
type Runner interface {
	Output(repo string, args ...string) ([]byte, error)
}

// DefaultGitBound bounds one git invocation: a hung repository must cost
// the handler its bound, not an unbounded goroutine and a git process that
// outlives the request. It bounds one call only; BundleBudget bounds the
// bundle's calls together. ExecRunner's Bound field overrides it per
// runner — the seam the bound's own test drives a sleeping git through.
// DefaultGitBound is the bound an ExecRunner with a zero Bound applies.
// Exported for the api package's budget test.
const DefaultGitBound = 10 * time.Second

// waitAfterKill bounds how long Wait may stay blocked on pipes a killed
// git's descendants still hold: without it, a textconv filter or credential
// helper that ignores the signal defeats the bound entirely.
const waitAfterKill = 5 * time.Second

// BundleBudget bounds every git call of one bundle together. One bundle
// runs some thirty git calls in series, so DefaultGitBound alone would let
// a slow repository — a loaded machine where each call takes a couple of
// seconds — run the bundle far past the server's writeTimeout, which cuts
// the connection with an empty reply instead of serving the degraded
// bundle. Once the budget is spent, every remaining call fails at once and
// its source is reported skipped. Exported for the api package's budget
// test: BundleBudget plus DefaultGitBound must sit inside writeTimeout.
const BundleBudget = 15 * time.Second

// ExecRunner runs the real git binary. A zero Bound means DefaultGitBound.
// A non-zero Deadline caps every call: a call starting past it fails
// without running git, and a call running into it is killed there.
type ExecRunner struct {
	Bound    time.Duration
	Deadline time.Time
}

// Output runs git in repo and returns its stdout. The command is bound —
// the daemon serves local repositories, where git either answers in well
// under a second or is not going to answer at all — and its wait is
// itself bounded past the kill, so no pipe-holding descendant can hang the
// call past the bound.
func (r ExecRunner) Output(repo string, args ...string) ([]byte, error) {
	bound := r.Bound
	if bound == 0 {
		bound = DefaultGitBound
	}
	if !r.Deadline.IsZero() {
		left := time.Until(r.Deadline)
		if left <= 0 {
			return nil, fmt.Errorf("git %s: the bundle's git budget is spent", strings.Join(args, " "))
		}
		bound = min(bound, left)
	}
	ctx, cancel := context.WithTimeout(context.Background(), bound)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	cmd.WaitDelay = waitAfterKill
	return cmd.Output()
}

// baseRef is the remote-tracking ref a branch lands on: origin/<the base
// recorded as branch.<branch>.flowBase>, as flow-guard resolve-base-branch reads
// it, else what origin/HEAD points at; "" when neither resolves.
func baseRef(g Runner, repo, branch string) string {
	if rec := trimmedOutput(g.Output(repo, "config", "--get", "branch."+branch+".flowBase")); rec != "" {
		return "origin/" + rec
	}
	return trimmedOutput(g.Output(repo, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
}

// unreadableRepos names every supplied repository git cannot read at all —
// the environmental failure class that must never wear the same
// "skipped (absent)" wording a change with no directory on disk wears.
func unreadableRepos(g Runner, repos []string) []string {
	var broken []string
	for _, repo := range repos {
		if _, err := g.Output(repo, "rev-parse", "--git-dir"); err != nil {
			broken = append(broken, repo)
		}
	}
	return broken
}

// firstReadableRepo returns the first supplied repository git can read at
// all — for a change with no directory on disk, the checkout the caller
// resolved, and the only repository the change-branch fallback reads.
func firstReadableRepo(g Runner, repos []string) string {
	for _, repo := range repos {
		if _, err := g.Output(repo, "rev-parse", "--git-dir"); err == nil {
			return repo
		}
	}
	return ""
}

// changeBranchLog renders the change branch's own commits against the
// remote default branch — the git-log source for a change with no change
// directory: a /flow-fast change's implementation commits live on the branch
// named after the change. The base is the one recorded on the branch
// (branch.<change>.flowBase, as flow-guard resolve-base-branch reads it), else what
// origin/HEAD points at, the same notion the flow-fast run's own Branch log
// printed; a repository whose
// origin/HEAD is unset, and a branch that resolves nothing — landed and
// deleted, or never created — degrade to an empty string: a missing source
// is never fatal.
func changeBranchLog(g Runner, repo, change string) string {
	base := baseRef(g, repo, change)
	if base == "" {
		return ""
	}
	return dropPNGStatLines(trimmedOutput(g.Output(repo, "log", "--stat", base+".."+change)))
}

// requestsLog renders a /flow-interactive session's requests in n order --
// the git-log source for a change with requests (design.md's Explicit stop
// step 2), whose session branch's own log is empty by stop time: each
// request's `### Request <n>: <title> — <state>` heading (an empty title
// reads `(untitled)`), then for a landed one each commit of its base..commit
// range: its `git log -1 --stat` the first time any request lists it, and
// `commit <sha12> — landed with Request <m>` after, so ranges that share
// commits print each once (kan-1081 self-review, F6). A range git cannot
// read contributes its heading alone.
func requestsLog(g Runner, repo string, requests []records.Request) string {
	sorted := slices.Clone(requests)
	slices.SortFunc(sorted, func(a, b records.Request) int { return a.N - b.N })
	printedUnder := map[string]int{}
	var b strings.Builder
	for _, r := range sorted {
		title := r.Title
		if title == "" {
			title = "(untitled)"
		}
		fmt.Fprintf(&b, "### Request %d: %s — %s\n\n", r.N, title, r.State)
		if r.Outcome != "landed" || r.BaseSHA == "" || r.CommitSHA == "" {
			continue
		}
		for _, sha := range strings.Fields(trimmedOutput(g.Output(repo, "log", "--format=%H", r.BaseSHA+".."+r.CommitSHA))) {
			if m, ok := printedUnder[sha]; ok {
				fmt.Fprintf(&b, "commit %s — landed with Request %d\n\n", sha[:min(12, len(sha))], m)
				continue
			}
			printedUnder[sha] = r.N
			if log := dropPNGStatLines(trimmedOutput(g.Output(repo, "log", "-1", "--stat", sha))); log != "" {
				b.WriteString(log + "\n\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// mergeOrderEntry is one `## Merge order` line of a link.md: a numbered
// entry naming `.` or a peer in backticks — the syntax spectre validate
// accepts.
var mergeOrderEntry = regexp.MustCompile("^[0-9]+\\. `([^`]+)`$")

// peerLogs renders each satellite repository's change-branch log — a
// cross-repo change's implementation commits outside repo — under a
// `### <peer>` heading, in link.md's `## Merge order`. Peers resolve through
// repo's spectre/peers (`<name> <path>`, a relative path joined to repo's
// main checkout, the directory the file's paths are written against). A
// change with no link.md, a peer the file does not declare and a branch
// that resolves nothing all contribute nothing: a missing source is never
// fatal.
func peerLogs(g Runner, repo, name string) string {
	link, ok := changeFile(repo, name, "link.md")
	if !ok {
		return ""
	}
	peersFile, _ := os.ReadFile(filepath.Join(repo, "spectre", "peers"))
	main := repo
	if common := trimmedOutput(g.Output(repo, "rev-parse", "--path-format=absolute", "--git-common-dir")); common != "" {
		main = filepath.Dir(common)
	}
	paths := map[string]string{}
	for l := range strings.Lines(string(peersFile)) {
		if f := strings.Fields(l); len(f) == 2 {
			paths[f[0]] = f[1]
		}
	}
	var b strings.Builder
	inOrder := false
	for l := range strings.Lines(string(link)) {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "## ") {
			inOrder = l == "## Merge order"
			continue
		}
		m := mergeOrderEntry.FindStringSubmatch(l)
		if !inOrder || m == nil || m[1] == "." {
			continue
		}
		p, ok := paths[m[1]]
		if !ok {
			continue
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(main, p)
		}
		if log := changeBranchLog(g, p, changeBranchPrefix+name); log != "" {
			fmt.Fprintf(&b, "### %s\n\n%s\n\n", m[1], log)
		}
	}
	return b.String()
}

// finishLog renders the change's finish-run commits in repo, the worktree on
// spectre/<name>: the first-parent `git log --stat` of merge-base..HEAD
// against the branch's base (baseRef). No base, or nothing above it, is "".
func finishLog(g Runner, repo, name string) string {
	base := baseRef(g, repo, changeBranchPrefix+name)
	if base == "" {
		return ""
	}
	mb := trimmedOutput(g.Output(repo, "merge-base", "HEAD", base))
	if mb == "" {
		return ""
	}
	return dropPNGStatLines(trimmedOutput(g.Output(repo, "log", "--first-parent", "--stat", mb+"..HEAD")))
}

// changeFile reads file of change name from repo's working tree on disk —
// spectre/changes/<name>/<file>, gitignored and never committed. ok false
// when absent.
func changeFile(repo, name, file string) ([]byte, bool) {
	b, err := os.ReadFile(filepath.Join(repo, liveDir, name, file))
	return b, err == nil
}

// changeRepo returns the first supplied repository whose working tree holds
// change name's directory, "" when none does.
func changeRepo(repos []string, name string) string {
	for _, repo := range repos {
		if fi, err := os.Stat(filepath.Join(repo, liveDir, name)); err == nil && fi.IsDir() {
			return repo
		}
	}
	return ""
}

// pngStatLine matches a `git log --stat` file line whose path ends `.png`.
var pngStatLine = regexp.MustCompile(`^ \S.*\.png\s+\|`)

// dropPNGStatLines drops every `git log --stat` line naming a `*.png` path:
// screenshot churn the reasoning pass cannot read.
func dropPNGStatLines(s string) string {
	var b strings.Builder
	for l := range strings.Lines(s) {
		if !pngStatLine.MatchString(l) {
			b.WriteString(l)
		}
	}
	return b.String()
}

// elideMatrixRows drops every table row between a heading naming a matrix
// (its text holds `matrix`, any case, at any level) and the next heading,
// leaving one `_(<n> matrix rows elided)_` line per run. A line inside a ```
// fence is never a heading or a row.
func elideMatrixRows(md string) string {
	var b strings.Builder
	inMatrix, inFence, run := false, false, 0
	flush := func() {
		if run > 0 {
			fmt.Fprintf(&b, "_(%d matrix rows elided)_\n", run)
			run = 0
		}
	}
	for l := range strings.Lines(md) {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			inFence = !inFence
			flush()
		} else if inFence {
			flush()
		} else if strings.HasPrefix(l, "#") {
			flush()
			inMatrix = strings.Contains(strings.ToLower(l), "matrix")
		} else if inMatrix && strings.HasPrefix(strings.TrimSpace(l), "|") {
			run++
			continue
		} else {
			flush()
		}
		b.WriteString(l)
	}
	flush()
	return b.String()
}

// measurementHeading names a verification record's raw measurement
// section by its heading's first word: the bands and seams summaries and
// the ink runs the verifier writes beside its verdicts (kan-1026's
// self-review: about 200 of the bundle's lines, read by no angle). A heading
// naming one later, `### Q1 sweeps (rows, ink)`, is a verdict section.
var measurementHeading = regexp.MustCompile(`(?i)^#+\s+(bands|seams|ink)\b`)

// elideMeasurementSections drops every section whose heading names a raw
// measurement (measurementHeading), from that heading through the line
// before the next heading of its level or higher, leaving one `_(<n> lines
// elided: <heading> — in <label>)_` line per section. The header, the
// verify output and every verdict section stay whole, and so does a matrix
// section, whose rows elideMatrixRows already dropped. A line inside a ```
// fence is never a heading.
func elideMeasurementSections(md, label string) string {
	var b strings.Builder
	inFence, level, n, heading := false, 0, 0, ""
	flush := func() {
		if level > 0 {
			fmt.Fprintf(&b, "_(%d lines elided: %s — in %s)_\n\n", n, heading, label)
		}
		level, n = 0, 0
	}
	for l := range strings.Lines(md) {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			inFence = !inFence
		}
		if !inFence && strings.HasPrefix(l, "#") {
			depth := len(l) - len(strings.TrimLeft(l, "#"))
			if level > 0 && depth <= level {
				flush()
			}
			if level == 0 && measurementHeading.MatchString(l) {
				level, heading = depth, strings.TrimSpace(strings.TrimLeft(l, "# "))
			}
		}
		if level > 0 {
			n++
			continue
		}
		b.WriteString(l)
	}
	flush()
	return b.String()
}

func trimmedOutput(b []byte, err error) string {
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// changeFiles are the change directory's files the bundle serves, in
// order. visual-verification.md is the record the flow-eval angle's visual
// candidates are read from (skills/flow-self-review/SKILL.md, angle 7), and
// live-verification.md the live check's before/after figures
// (skills/flow/verify-and-handoff.md); a change with neither reports each
// skipped like any absent source.
var changeFiles = []string{"tasks.md", "design.md", "narrative.md", "visual-verification.md", "live-verification.md"}

// Bundle assembles the whole self-review context bundle for change as one
// Markdown document: the header, the found/skipped summary line, one
// `skipped: <label> (absent)` line per absent source, and one `## <label>`
// section per found source. summary and summaryFound carry the change's
// recorded summary — the run's own handoff report, stored verbatim in the
// store — which is the bundle's FIRST source when found: the change's own
// statement of what it did and why is the central content a deferred
// reasoning pass reads, and it must not sit behind the row renders it
// explains. Absent, it reports skipped like any other missing source.
// run renders the ledger and panel sources —
// each present only when the run holds rows of its kind, so a change the
// store has never heard of reports both skipped rather than rendering
// empty records nobody wrote; repos are the candidate worktree roots the
// caller supplied, probed in order for the change's directory on disk;
// panelRan reports whether a
// flow.review-panel stage run for the change completed in this store —
// the one fact separating a records-source loss from a panel that never
// ran (KAN-621): when it holds while the run carries no dispatch rows,
// the bundle's notes name the loss loudly, because the plain
// "skipped (absent)" wording would read as the panel never having
// existed; g reads everything git has to
// answer for. The worktree holding the directory serves the changeFiles
// sources and its finish range (finishLog) as the git-log source. A change
// no supplied worktree holds a directory for — a /flow-fast change, which
// has none — still gets its git-log source, read off the change branch in
// the first readable supplied repository (KAN-583); the changeFiles sources
// stay skipped for it. requests, when any, are a /flow-interactive
// session's (KAN-1061): its git-log source is then requestsLog's, read in
// that same repository, in place of every other git-log derivation. The git-log source also carries each merge-order
// peer's change-branch log under `### <peer>`. A repository git cannot read at all is reported in
// a `note:`
// line — environmental failure keeps a different wording from legitimate
// absence. An invalid change name is the one error: the same allowlist
// records.Destination enforces, checked before the name builds a label or
// a ref. build names the serving daemon's own build, the note leading the
// bundle's notes when non-empty: a source reported skipped that the tree
// holds then reads as a daemon older than the source, never as an absence
// (KAN-771). The gated per-task reviewers' findings, which the panel record
// leaves out, follow the panel as their own source.
func Bundle(change string, run records.Run, summary string, summaryFound bool, panelRan bool, build string, repos []string, requests []records.Request, g Runner) (string, error) {
	if !records.ValidChangeName(change) {
		return "", fmt.Errorf("change name %q is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'", change)
	}

	type source struct {
		label   string
		content string
		found   bool
	}

	var sources []source
	add := func(label, content string, found bool) {
		sources = append(sources, source{label: label, content: content, found: found})
	}

	// The recorded summary comes first, ahead of every row render and
	// history probe: it is the change's own statement of what it did and
	// why, and the source a deferred reasoning pass reads before any of
	// the evidence below it. The content is served verbatim -- the run
	// wrote Markdown and the store keeps it byte for byte -- never
	// reflowed or re-derived from rows.
	add("change summary", summary, summaryFound)

	// The change-directory sources all come from one worktree: the first
	// supplied one holding the directory on disk. None skips all of them
	// together — the store renders still stand. A repository git cannot read
	// at all is named in a note rather than folded into "absent".
	var notes []string
	if build != "" {
		notes = append(notes, "note: served by flowd "+build+" — a source reported skipped that the tree holds means this daemon predates it, not that it is absent")
	}
	// The records-source loss note leads the notes after the build: a panel whose stage
	// demonstrably completed while the store holds no dispatch rows means
	// the rows were lost, not skipped, and every "skipped (absent)" line
	// below would otherwise read as the panel never having run (KAN-621).
	if panelRan && len(run.Dispatches) == 0 {
		notes = append(notes, "note: RECORDS LOSS — a flow.review-panel stage run completed for "+change+", but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.")
	}
	repo := changeRepo(repos, change)
	if repo == "" {
		for _, broken := range unreadableRepos(g, repos) {
			notes = append(notes, "note: repository "+broken+" could not be read — its sources are reported skipped for that reason, not because the change has no directory")
		}
	}

	// The store renders: the ledger needs dispatch rows (RenderKind's own
	// rule); the panel needs any row of the run at all, so an unknown
	// change's empty run never renders a findings-total: 0 record nobody
	// wrote. records.Run.HasRows is the one home of that row set.
	ledger, ledgerOK := records.RenderKind("ledger", run)
	add(".superpowers/sdd/ledgers/"+change+".md", ledger, ledgerOK)
	panel, _ := records.RenderKind("panel", run)
	add(".superpowers/sdd/reviews/"+change+"-panel.md", panel, run.HasRows())
	gated, ok := records.RenderGatedFindings(run)
	add("gated per-task reviewer findings", gated, ok)

	for _, file := range changeFiles {
		label := liveDir + "/" + change + "/" + file
		var content []byte
		ok := false
		if repo != "" {
			content, ok = changeFile(repo, change, file)
		}
		text := string(content)
		if file == "visual-verification.md" {
			text = elideMeasurementSections(elideMatrixRows(text), label)
		}
		add(label, text, ok)
	}

	gitLog := ""
	if len(requests) > 0 {
		if mainRepo := firstReadableRepo(g, repos); mainRepo != "" {
			gitLog = requestsLog(g, mainRepo, requests)
		}
	} else if repo != "" {
		gitLog = strings.TrimSpace(finishLog(g, repo, change) + "\n\n" + peerLogs(g, repo, change))
	} else if mainRepo := firstReadableRepo(g, repos); mainRepo != "" {
		gitLog = changeBranchLog(g, mainRepo, change)
	}
	add("git log --stat", gitLog, gitLog != "")

	var found, skipped []string
	for _, s := range sources {
		if s.found {
			found = append(found, s.label)
		} else {
			skipped = append(skipped, s.label)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Self-review context bundle for %s\n\n", change)
	fmt.Fprintf(&b, "found: %d of %d sources; skipped: %d of %d sources\n",
		len(found), len(sources), len(skipped), len(sources))
	for _, note := range notes {
		b.WriteString(note)
		b.WriteString("\n")
	}
	for _, label := range skipped {
		fmt.Fprintf(&b, "skipped: %s (absent)\n", label)
	}
	b.WriteString("\n")
	for _, s := range sources {
		if !s.found {
			continue
		}
		fmt.Fprintf(&b, "## %s\n\n%s\n", s.label, strings.TrimRight(s.content, "\n"))
	}
	return b.String(), nil
}
