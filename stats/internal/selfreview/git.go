// Package selfreview assembles the self-review context bundle a change's
// reasoning pass reads — the sources run 1's bundle step judges a change
// by, served from the store and the change's own repository instead of
// gathered from files a Bash script had to be pointed at.
//
// Everything this package reads is read through git (`git -C <repo>`) or
// rendered from the run record; no path any caller supplies is resolved
// against this process's working directory, and the dropped change is
// read out of history — the parent of run 1's `chore(spectre): drop <name>`
// commit (stats/internal/landed), found on the change branch `spectre/<name>`
// while it exists, the base branch once it has landed — rather than out of
// any worktree's working files, so neither the invocation nor the bundle's
// own section labels carry a worktree path.
package selfreview

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/landed"
	"github.com/tweety53/agents/stats/internal/records"
)

// changeBranchPrefix names the branch run 1 drops the change directory on,
// in its apply worktree. The branch lives in the repository's shared ref store, so
// `git show <branch>:<path>` from the main checkout reads it with no
// worktree path at all.
const changeBranchPrefix = "spectre/"

// liveDir is the change directories' parent under the repository's spectre
// tree — the pathspec the planning-commit query searches and the label root
// of every file the bundle serves from the dropped change. The Bash gather
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

// deriveFinishCommits resolves every implementation, planning and drop
// commit of change name out of repo, oldest first, under the same rules
// the retired Bash gather stated for one triple of them. Every query
// starts from rev — the revision dropRev found the drop commit on —
// never HEAD: naming it explicitly is what lets the app see the same
// commits from any checkout of the repository. A change carries more than
// one implementation-and-planning pair, and more than one drop commit, once
// a fix run after the drop (skills/flow-contracts/finish-contract-run1.md)
// committed its own, and self-review covers the whole change, so every
// pair is listed.
//
//   - a drop commit is a commit whose subject is exactly
//     `chore(spectre): drop <name>`;
//   - a planning commit is a commit that BOTH carries a plan-commit
//     subject — a prefix match on `chore(spectre): plan`, which covers the
//     bare literal of changes landed before 2026-10-05 and the
//     `chore(spectre): plan <name>[ — <boundary>]` subjects since, or either
//     pre-rename wording scoped to the change — AND touched the change's
//     spectre/changes/<name>/ path. Path alone is not commit-specific (the
//     drop commit itself touches it); subject alone is not change-specific
//     (the bare literal is identical across changes);
//   - each planning commit's implementation commit is its first parent,
//     accepted only when it sits above the change's merge base
//     (changeMergeBase), is a non-merge commit, matches none of the four
//     reserved subject shapes, and touches at least one path outside
//     spectre/changes/, docs/research/ and docs/superpowers/ — anything
//     else resolves NOTHING rather than a confident wrong answer — and
//     listed directly before its planning commit. The merge-base gate is
//     what keeps the oldest planning commit's parent out: before
//     integrate's reshape the first planning commit sits on the merge
//     base — and changes landed from 2026-09-24 to 2026-10-05 kept it
//     there — so that parent is the base tip, a commit of some other change; the bundle
//     subject is what keeps a re-run's first planning commit's parent out,
//     the previous run's self-review bundle commit.
//
// Every git failure degrades to an empty list, or a pair without its
// implementation commit, the gather's `|| true` semantics: a missing
// source is never fatal to the bundle.
func deriveFinishCommits(g Runner, repo, rev, name string) []string {
	nameRe := regexp.QuoteMeta(name)
	shapes := reservedShapes(nameRe)
	dropRe := regexp.MustCompile(shapes.drop)

	// The pathspec finds run 1's planning commit even after the drop
	// commit deleted the directory, since `git log -- <path>` filters each
	// commit by its own historical tree; the drop commit deletes under it,
	// and a fix run's restoring planning commit adds under it, so the same
	// pathspec carries both.
	found := trimmedOutput(g.Output(repo, "log", rev, "--reverse", "-E",
		"--grep="+shapes.planNew, "--grep="+shapes.planOld, "--grep="+shapes.drop,
		"--format=%H %s",
		"--", liveDir+"/"+name))

	bound := changeMergeBase(g, repo, rev, name)
	var out []string
	for _, line := range strings.Split(found, "\n") {
		sha, subject, _ := strings.Cut(line, " ")
		if sha == "" {
			continue
		}
		if !dropRe.MatchString(subject) {
			impl := trimmedOutput(g.Output(repo, "rev-parse", sha+"^"))
			if !atOrBelow(g, repo, impl, bound) && isRealImplCommit(g, repo, impl, nameRe) {
				out = append(out, impl)
			}
		}
		out = append(out, sha)
	}
	return out
}

// reservedShapes is the one declaration of the five reserved subject
// shapes — the current plan-commit literal (a prefix match, no `$`, exactly
// as the gather anchored it), the pre-rename wording scoped to the change,
// the exact drop subject, the exact self-review bundle subject a run
// before KAN-927 committed, and the exact self-review report subject run 1
// commits since. deriveFinishCommits greps with the first three and
// isRealImplCommit rejects against all five; one declaration is what keeps the gate and the
// query from drifting apart.
func reservedShapes(nameRe string) struct{ planNew, planOld, drop, bundle, report string } {
	return struct{ planNew, planOld, drop, bundle, report string }{
		planNew: `^chore\(spectre\): plan`,
		planOld: `^chore\(` + nameRe + `\): plan(, test guide and| and) session records`,
		drop:    `^chore\(spectre\): drop ` + nameRe + `$`,
		bundle:  `^docs\(self-review\): ` + nameRe + ` self-review context bundle$`,
		report:  `^docs\(self-review\): ` + nameRe + ` self-review report$`,
	}
}

// changeMergeBase returns the commit the change branch forked from: the
// merge base of rev and the change's base — origin/<base> for the base
// recorded on spectre/<name>, else what origin/HEAD points at. "" when rev
// is not the change branch or is already contained in its base (the change
// has landed, and the fork point is no longer visible in the topology), or
// when no base resolves; every parent is then judged by isRealImplCommit
// alone.
//
// ponytail: a landed change's bundle has no merge-base gate; it was saved,
// gated, at run 1 before landing, so only a hand re-assembly after landing
// can list the base tip below a reshaped first planning commit.
func changeMergeBase(g Runner, repo, rev, name string) string {
	if rev != changeBranchPrefix+name {
		return ""
	}
	base := baseRef(g, repo, rev)
	if base == "" {
		return ""
	}
	mb := trimmedOutput(g.Output(repo, "merge-base", rev, base))
	if mb == "" || mb == trimmedOutput(g.Output(repo, "rev-parse", rev)) {
		return ""
	}
	return mb
}

// baseRef is the remote-tracking ref a branch lands on: origin/<the base
// recorded as branch.<branch>.flowBase>, as resolve-base-branch.sh reads
// it, else what origin/HEAD points at; "" when neither resolves.
func baseRef(g Runner, repo, branch string) string {
	if rec := trimmedOutput(g.Output(repo, "config", "--get", "branch."+branch+".flowBase")); rec != "" {
		return "origin/" + rec
	}
	return trimmedOutput(g.Output(repo, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
}

// atOrBelow reports whether commit is bound or one of its ancestors —
// outside the change's own range. An empty bound or commit bounds nothing.
func atOrBelow(g Runner, repo, commit, bound string) bool {
	if commit == "" || bound == "" {
		return false
	}
	_, err := g.Output(repo, "merge-base", "--is-ancestor", commit, bound)
	return err == nil
}

// isRealImplCommit judges plan's first parent by the four conditions the
// gather named. The merge gate and the plan and drop subject rejections
// are deliberately dead today — commit-split.sh's commits always carry one
// parent and a plan-commit subject never sits one commit below another
// reserved subject — and kept anyway: each guards this function's own git
// queries against a future edit to commit-split.sh's staging shape, which
// no test of this function can see. The bundle and report subject
// rejections are live: a re-run's first planning commit sits on
// the previous run's bundle or report commit until the re-run's reshape.
func isRealImplCommit(g Runner, repo, impl, nameRe string) bool {
	if impl == "" {
		return false
	}

	// A second parent resolving at all means a merge commit.
	if _, err := g.Output(repo, "rev-parse", "--verify", "-q", impl+"^2"); err == nil {
		return false
	}

	subject := trimmedOutput(g.Output(repo, "log", "-1", "--format=%s", impl))
	shapes := reservedShapes(nameRe)
	for _, re := range []string{shapes.planNew, shapes.planOld, shapes.drop, shapes.bundle, shapes.report} {
		if matched, err := regexp.MatchString(re, subject); err == nil && matched {
			return false
		}
	}

	// The exclusion pathspec tracks commit-split.sh's planning-tree list
	// exactly: spectre/changes/, not spectre/ — a capability spec under
	// spectre/specs/ is implementation, and widening the exclusion would
	// filter a spec-only implementation commit's only path away.
	outside, err := g.Output(repo, "diff-tree", "--no-commit-id", "--name-only", "-r", "--root", impl)
	if err != nil {
		return false
	}
	excl := regexp.MustCompile(`^(` + liveDir + `/|docs/superpowers/|docs/research/)`)
	for _, line := range strings.Split(strings.TrimSpace(string(outside)), "\n") {
		if line != "" && !excl.MatchString(line) {
			return true
		}
	}
	return false
}

// show reads one path out of rev's committed tree in repo. The caller
// treats an error as "absent": a missing source is reported skipped inside
// the bundle, never fatal.
func show(g Runner, repo, rev, path string) ([]byte, error) {
	return g.Output(repo, "show", rev+":"+path)
}

// droppedFile reads file of the dropped change name out of repo's history,
// through landed.Read on g: the drop commit's parent. ok false when absent.
func droppedFile(g Runner, repo, name, file string) ([]byte, bool) {
	b, ok, err := landed.Read(func(args ...string) ([]byte, error) { return g.Output(repo, args...) }, name, file)
	return b, ok && err == nil
}

// dropRepo returns the first recorded repository, and the revision in it,
// carrying this change's drop commit whose parent held its tasks.md — the
// tasks.md probe is what tells THIS change's drop from a same-named branch
// with none. Repos are probed in the order the caller listed them; a change
// with no repository carrying it yields "", "".
func dropRepo(g Runner, repos []string, name string) (string, string) {
	for _, repo := range repos {
		if rev := dropRev(g, repo, name); rev != "" {
			if _, ok := droppedFile(g, repo, name, "tasks.md"); ok {
				return repo, rev
			}
		}
	}
	return "", ""
}

// dropRev returns the first revision in repo whose history holds the
// change's drop commit: the change branch spectre/<name>, where run 1 drops
// it, while that branch exists; then origin/<base> and <base> — the base
// being what origin/HEAD points at — once the branch has landed and been
// removed, origin's tip first because the local base may lag it. "" when
// none holds it.
func dropRev(g Runner, repo, name string) string {
	revs := []string{changeBranchPrefix + name}
	if remote := trimmedOutput(g.Output(repo, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")); strings.HasPrefix(remote, "origin/") {
		revs = append(revs, remote, strings.TrimPrefix(remote, "origin/"))
	}
	subject := "chore(spectre): drop " + name
	for _, rev := range revs {
		log, err := g.Output(repo, "log", rev, "--format=%s", "--fixed-strings", "--grep="+subject)
		if err == nil && slices.Contains(strings.Split(strings.TrimSpace(string(log)), "\n"), subject) {
			return rev
		}
	}
	return ""
}

// unreadableRepos names every supplied repository git cannot read at all —
// the environmental failure class that must never wear the same
// "skipped (absent)" wording a legitimately never-dropped change wears.
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
// all — for a change no repository drops, this is the main checkout the
// caller resolved, and the only repository the change-branch fallback
// reads.
func firstReadableRepo(g Runner, repos []string) string {
	for _, repo := range repos {
		if _, err := g.Output(repo, "rev-parse", "--git-dir"); err == nil {
			return repo
		}
	}
	return ""
}

// changeBranchLog renders the change branch's own commits against the
// remote default branch — the git-log source for a change with no drop
// commit: a /flow-fast change's implementation commits live on the branch
// named after the change, where deriveFinishCommits's planning-and-drop query
// has nothing to find. The base is the one recorded on the branch
// (branch.<change>.flowBase, as resolve-base-branch.sh reads it), else what
// origin/HEAD points at, the same notion the flow-fast run's own Branch log
// printed; a repository whose
// origin/HEAD is unset, and a branch that resolves nothing — landed and
// deleted, or never created — degrade to an empty string, the same "a
// missing source is never fatal" rule deriveFinishCommits follows.
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
// request's `### Request <n>: <title> — <state>` heading, then for a landed
// one its base..commit range's `git log --stat`. A range git cannot read
// contributes its heading alone.
func requestsLog(g Runner, repo string, requests []records.Request) string {
	sorted := slices.Clone(requests)
	slices.SortFunc(sorted, func(a, b records.Request) int { return a.N - b.N })
	var b strings.Builder
	for _, r := range sorted {
		fmt.Fprintf(&b, "### Request %d: %s — %s\n\n", r.N, r.Title, r.State)
		if r.Outcome == "landed" && r.BaseSHA != "" && r.CommitSHA != "" {
			if log := dropPNGStatLines(trimmedOutput(g.Output(repo, "log", "--stat", r.BaseSHA+".."+r.CommitSHA))); log != "" {
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
// `### <peer>` heading, in the dropped link.md's `## Merge order`. Peers
// resolve through rev's spectre/peers (`<name> <path>`, a relative path
// joined to repo). A change with no link.md, a peer the file does not
// declare and a branch that resolves nothing all contribute nothing: a
// missing source is never fatal.
func peerLogs(g Runner, repo, rev, name string) string {
	link, ok := droppedFile(g, repo, name, "link.md")
	if !ok {
		return ""
	}
	peersFile, _ := show(g, repo, rev, "spectre/peers")
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
			p = filepath.Join(repo, p)
		}
		if log := changeBranchLog(g, p, changeBranchPrefix+name); log != "" {
			fmt.Fprintf(&b, "### %s\n\n%s\n\n", m[1], log)
		}
	}
	return b.String()
}

// gitLogSection renders the git-log source's content: `git log --stat -1`
// per resolved sha, in deriveFinishCommits's oldest-first order.
func gitLogSection(g Runner, repo string, shas []string) string {
	var b strings.Builder
	for _, sha := range shas {
		if logStat, err := g.Output(repo, "log", "--stat", "-1", sha); err == nil {
			b.WriteString(dropPNGStatLines(string(logStat)))
			b.WriteString("\n")
		}
	}
	return b.String()
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

// droppedFiles are the dropped change's files the bundle serves, in
// order. visual-verification.md is the record the flow-eval angle's visual
// candidates are read from (skills/flow-self-review/SKILL.md, angle 7), and
// live-verification.md the live check's before/after figures
// (skills/flow/verify-and-handoff.md); a change with neither reports each
// skipped like any absent source.
var droppedFiles = []string{"tasks.md", "design.md", "narrative.md", "visual-verification.md", "live-verification.md"}

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
// empty records nobody wrote; no committed copy exists to fall back to, the
// drop commit carrying none; repos are the candidate repository roots the
// caller supplied, probed for the change's drop commit in order; panelRan reports whether a
// flow.review-panel stage run for the change completed in this store —
// the one fact separating a records-source loss from a panel that never
// ran (KAN-621): when it holds while the run carries no dispatch rows,
// the bundle's notes name the loss loudly, because the plain
// "skipped (absent)" wording would read as the panel never having
// existed; g reads everything git has to
// answer for. A change no supplied repository drops — a /flow-fast
// change, never dropped by design — still gets its git-log source, read
// off the change branch in the first readable supplied repository, the
// main checkout the caller resolved (KAN-583); the droppedFiles sources
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

	// The history-derived sources all come from one repository: the first
	// supplied one carrying the change's drop commit. No drop anywhere skips
	// all of them together — the store renders still stand. A repository
	// git cannot read at all is named in a note rather than folded into
	// "absent".
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
	repo, rev := dropRepo(g, repos, change)
	if repo == "" {
		for _, broken := range unreadableRepos(g, repos) {
			notes = append(notes, "note: repository "+broken+" could not be read — its history sources are reported skipped for that reason, not because the change was never dropped")
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

	for _, file := range droppedFiles {
		label := liveDir + "/" + change + "/" + file
		var content []byte
		ok := false
		if repo != "" {
			content, ok = droppedFile(g, repo, change, file)
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
		gitLog = gitLogSection(g, repo, deriveFinishCommits(g, repo, rev, change)) + peerLogs(g, repo, rev, change)
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
