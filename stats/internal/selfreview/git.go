// Package selfreview assembles the self-review context bundle a finished
// change's reasoning pass reads — the six sources run 2 step 9 judges a
// change by, served from the store and the change's own repository instead
// of gathered from files a Bash script had to be pointed at.
//
// Everything this package reads is read through git (`git -C <repo>`) or
// rendered from the run record; no path any caller supplies is resolved
// against this process's working directory, and the archived change is
// read out of the `chore/archive-<name>` branch's committed tree rather
// than out of any worktree's working files — which is what removes the
// landing-worktree path coupling the Bash gather carried, both in its
// invocation (the caller passed the worktree path in) and in its output
// (the bundle's own section labels quoted that absolute path back).
package selfreview

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// archiveBranch is the branch run 2's archive step commits the archived
// change onto (prepare-archive-branch.sh's own name). It exists as a local
// branch in the repository's shared object store while step 9 runs — the
// landing worktree that checked it out is not removed until step 11 — so
// reading the archived files through `git show <branch>:<path>` needs no
// worktree path at all, and keeps working after cleanup for as long as the
// branch lives.
const archiveBranchPrefix = "chore/archive-"

// probeFile is the archived file whose presence decides that a repository
// carries THIS change's archive, not merely a same-named branch: spectre
// scaffolds tasks.md for every change, so an archived change always has
// one.
const probeFile = "tasks.md"

// liveDir is the working change's directory under the repository's spectre
// tree — the pathspec the planning-commit query searches.
const liveDir = "spectre/changes"

// archiveDir is the archived change's directory under the repository's
// spectre tree. The Bash gather resolved the spec root leaf per repository
// (scripts/lib/spec-root.sh); the app pins `spectre`, this repository's
// leaf — a second spec root would be a store schema question first, not a
// string to probe for.
const archiveDir = "spectre/changes/archive"

// Runner is the git access the assembler needs. Every call is
// `git -C <repo> <args...>`; output is the command's stdout, an error its
// non-zero exit.
type Runner interface {
	Output(repo string, args ...string) ([]byte, error)
}

// defaultGitBound bounds one git invocation: a hung repository must cost
// the handler its bound, not an unbounded goroutine and a git process that
// outlives the request. It sits comfortably inside the API server's
// writeTimeout, so a bundle degraded by a bound-exceeded call still gets
// served. ExecRunner's Bound field overrides it per runner — the seam the
// bound's own test drives a sleeping git through.
// DefaultGitBound is the bound an ExecRunner with a zero Bound applies.
// Exported for the api package's budget test: the bound must sit
// comfortably inside that server's writeTimeout, so a bundle degraded by a
// bound-exceeded call is still delivered.
const DefaultGitBound = 10 * time.Second

// waitAfterKill bounds how long Wait may stay blocked on pipes a killed
// git's descendants still hold: without it, a textconv filter or credential
// helper that ignores the signal defeats the bound entirely.
const waitAfterKill = 5 * time.Second

// ExecRunner runs the real git binary. A zero Bound means defaultGitBound.
type ExecRunner struct {
	Bound time.Duration
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
	ctx, cancel := context.WithTimeout(context.Background(), bound)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	cmd.WaitDelay = waitAfterKill
	return cmd.Output()
}

// finishCommits is the three-commit spine of the git-log source: the
// implementation commit and the planning commit (finish run 1's own
// two-commit chain), plus run 2's archive commit. Any member may be empty —
// a change finished without one of them keeps the others rather than
// failing the source.
type finishCommits struct {
	impl    string
	plan    string
	archive string
}

// deriveFinishCommits resolves the three shas for change name out of repo,
// under the same rules the retired Bash gather stated for the same query.
// Every query starts from the archive branch, not HEAD: at step 9 the
// implementation and planning commits are already on the default branch,
// but the archive commit lives on chore/archive-<name> alone — the gather
// saw it only because its process cwd sat in the landing worktree whose
// HEAD was that branch, and naming the branch explicitly is what lets the
// app see the same three commits from any checkout of the repository.
//
//   - the archive commit is the most recent commit whose subject is exactly
//     `chore(spectre): archive <name>`;
//   - the planning commit is the most recent commit that BOTH carries a
//     plan-commit subject — the current fixed literal `chore(spectre):
//     plan`, or either pre-rename wording scoped to the change — AND
//     touched the change's live spectre/changes/<name>/ path. Path alone is
//     not commit-specific (a later typo fix to the archived directory would
//     outrank the real planning commit by recency); subject alone is not
//     change-specific (the fixed literal is identical across changes); the
//     live pathspec alone still finds the commit after run 2's `git mv`
//     because `git log -- <path>` filters each commit by its own
//     historical tree;
//   - the implementation commit is the planning commit's first parent,
//     accepted only when it is a non-merge commit, matches none of the
//     three reserved subject shapes, and touches at least one path outside
//     spectre/changes/, docs/research/ and docs/superpowers/ — anything
//     else resolves NOTHING rather than a confident wrong answer.
//
// Every git failure degrades to an empty sha, the gather's `|| true`
// semantics: a missing source is never fatal to the bundle.
func deriveFinishCommits(g Runner, repo, name string) finishCommits {
	nameRe := regexp.QuoteMeta(name)

	shapes := reservedShapes(nameRe)

	var out finishCommits

	branch := archiveBranchPrefix + name

	out.archive = trimmedOutput(g.Output(repo, "log", branch, "-E",
		"--grep="+shapes.archive, "--max-count=1", "--format=%H"))

	// Only the LIVE pathspec is searched, never the archived location:
	// `git log -- <path>` filters each commit by its own historical tree,
	// so the planning commit resolves even after run 2's `git mv` renamed
	// the directory into the archive.
	out.plan = trimmedOutput(g.Output(repo, "log", branch, "-E",
		"--grep="+shapes.planNew, "--grep="+shapes.planOld,
		"--max-count=1", "--format=%H",
		"--", liveDir+"/"+name))

	if out.plan == "" {
		return out
	}

	parent, err := g.Output(repo, "rev-parse", out.plan+"^")
	if err != nil {
		return out
	}
	impl := strings.TrimSpace(string(parent))
	if isRealImplCommit(g, repo, impl, nameRe) {
		out.impl = impl
	}
	return out
}

// reservedShapes is the one declaration of the three reserved subject
// shapes — the current plan-commit literal (a prefix match, no `$`, exactly
// as the gather anchored it), the pre-rename wording scoped to the change,
// and the exact archive subject. deriveFinishCommits greps with them and
// isRealImplCommit rejects against them; one declaration is what keeps the
// gate and the query from drifting apart.
func reservedShapes(nameRe string) struct{ planNew, planOld, archive string } {
	return struct{ planNew, planOld, archive string }{
		planNew: `^chore\(spectre\): plan`,
		planOld: `^chore\(` + nameRe + `\): plan(, test guide and| and) session records`,
		archive: `^chore\(spectre\): archive ` + nameRe + `$`,
	}
}

// isRealImplCommit judges plan's first parent by the four conditions the
// gather named. The three subject rejections and the merge gate are
// deliberately dead today — commit-split.sh's commits always carry one
// parent and a plan-commit subject never sits one commit below another
// reserved subject — and kept anyway: each guards this function's own git
// queries against a future edit to commit-split.sh's staging shape, which
// no test of this function can see.
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
	for _, re := range []string{shapes.planNew, shapes.planOld, shapes.archive} {
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

// archiveRepo returns the first recorded repository whose
// chore/archive-<name> branch carries this change's own archived directory
// — probed on the archived tasks.md's content, not on branch existence
// alone: a stale or same-prefixed branch that happens to exist is never
// the repository the archived change lives in. Repos are probed in the
// order the caller listed them; a change with no repository carrying the
// archived directory yields "".
func archiveRepo(g Runner, repos []string, name string) string {
	branch := archiveBranchPrefix + name
	for _, repo := range repos {
		if _, err := show(g, repo, branch, archiveDir+"/"+name+"/"+probeFile); err == nil {
			return repo
		}
	}
	return ""
}

// unreadableRepos names every supplied repository git cannot read at all —
// the environmental failure class that must never wear the same
// "skipped (absent)" wording a legitimately never-archived change wears.
func unreadableRepos(g Runner, repos []string) []string {
	var broken []string
	for _, repo := range repos {
		if _, err := g.Output(repo, "rev-parse", "--git-dir"); err != nil {
			broken = append(broken, repo)
		}
	}
	return broken
}

// gitLogSection renders the git-log source's content: `git log --stat -1`
// per resolved sha, implementation, planning, archive in that order — the
// same three commits in the same order the gather printed.
func gitLogSection(g Runner, repo string, fc finishCommits) string {
	var b strings.Builder
	for _, sha := range []string{fc.impl, fc.plan, fc.archive} {
		if sha == "" {
			continue
		}
		if logStat, err := g.Output(repo, "log", "--stat", "-1", sha); err == nil {
			b.Write(logStat)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func trimmedOutput(b []byte, err error) string {
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// Bundle assembles the whole self-review context bundle for change as one
// Markdown document: the header, the found/skipped summary line, one
// `skipped: <label> (absent)` line per absent source, and one `## <label>`
// section per found source. run renders the ledger and panel sources —
// each present only when the run holds rows of its kind, so a change the
// store has never heard of reports both skipped rather than rendering
// empty records nobody wrote; repos are the candidate repository roots the
// caller supplied, probed for the change's archived directory in order; g
// reads everything git has to answer for. A repository git cannot read at
// all is reported in a `note:` line — environmental failure keeps a
// different wording from legitimate absence. An invalid change name is the
// one error: the same allowlist records.Destination enforces, checked
// before the name builds a label or a ref.
func Bundle(change string, run records.Run, repos []string, g Runner) (string, error) {
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

	// The store renders: the ledger needs dispatch rows (RenderKind's own
	// rule); the panel needs any row of the run at all, so an unknown
	// change's empty run never renders a findings-total: 0 record nobody
	// wrote. records.Run.HasRows is the one home of that row set.
	ledger, ledgerOK := records.RenderKind("ledger", run)
	add(".superpowers/sdd/ledgers/"+change+".md", ledger, ledgerOK)
	panel, _ := records.RenderKind("panel", run)
	add(".superpowers/sdd/reviews/"+change+"-panel.md", panel, run.HasRows())

	// The archive-derived sources all come from one repository: the first
	// supplied one carrying the change's archived directory. No archive
	// anywhere skips all of them together — the store renders above still
	// stand. A repository git cannot read at all is named in a note rather
	// than folded into "absent".
	var notes []string
	repo := archiveRepo(g, repos, change)
	if repo == "" {
		for _, broken := range unreadableRepos(g, repos) {
			notes = append(notes, "note: repository "+broken+" could not be read — its archive sources are reported skipped for that reason, not because the change was never archived")
		}
	}
	branch := archiveBranchPrefix + change
	if repo != "" {
		for _, file := range []string{"tasks.md", "design.md", "narrative.md"} {
			label := archiveDir + "/" + change + "/" + file
			content, err := show(g, repo, branch, label)
			add(label, string(content), err == nil)
		}
	} else {
		for _, file := range []string{"tasks.md", "design.md", "narrative.md"} {
			add(archiveDir+"/"+change+"/"+file, "", false)
		}
	}

	gitLog := ""
	if repo != "" {
		gitLog = gitLogSection(g, repo, deriveFinishCommits(g, repo, change))
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
