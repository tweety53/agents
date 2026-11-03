package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// The contract below moved here from the header of the deleted
// scripts/gather-dispatch-context.sh shim (KAN-1058).
//
// flow-guard gather-dispatch-context — deterministically collect one change's
// planning context into a single bundle, so every dispatch in a /flow implement
// stage (each panel slot, each implementer, each fix subagent) is given
// demonstrably identical inputs instead of separately locating and reading
// proposal.md, design.md, tasks.md and the engineering principles on its
// own.
//
// Usage: flow-guard gather-dispatch-context <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree> [<shape>]]]
//
// <task-ids> is optional: a comma-separated list of integer task ids. When
// given, the "## tasks.md" section carries only the plan's header (every
// line before the first task line) plus the named tasks' own blocks, in
// document order, instead of the whole file — a named id the plan does not
// carry is a malformed invocation (exit 2), reported like any other plan
// defect. Task-line, body-boundary and fence grammar are
// plan-dispatch-bundles's (stats/internal/guard/planparse.go) — the ids that guard
// printed are the ids this finds.
//
// <worktree> and <change-root> are absolute paths; <change-root> is expected
// to sit under <worktree> (spectre/changes/<name>/, but this script does not
// assume that literal shape — it only requires containment), or under
// <canonical-worktree> when <worktree> is an ## apps repository given a
// worktree without a spectre link (skills/flow/cross-repo-worktrees.md), which
// holds no change directory of its own. <principles-path>
// is the absolute path of engineering-principles.md, resolved by the CALLER
// (skills/flow/review-panel.md's own [PRINCIPLES_PATH] rule) and never derived
// here — it may legitimately sit OUTSIDE <worktree> entirely, e.g. under a
// global install (~/.claude/skills/flow/engineering-principles.md), so
// it is validated but never checked for containment under <worktree>.
//
// Writes the bundle to <output-path>, falling back to reusing the existing
// file unchanged when its content hash matches the freshly computed one (see
// SKIP-WHEN-UNCHANGED below): a header naming the change, the generating
// instant and the worktree's HEAD sha; a found/skipped/refused census line;
// one "refused: <src> (resolves outside the change directory)" line per source
// whose leaf resolves outside <change-root> (see LEAF VALIDATION below), one
// "skipped: <src> (absent)" line per missing source; then each found source
// under its own "## " heading, in this fixed order: proposal.md, design.md,
// tasks.md, then the principles file.
//
// SKIP-WHEN-UNCHANGED. Everything the bundle prints after the header's
// `generated:`/`head:` lines (the census plus every "## <label>" section) is
// hashed with sha256Hex (stats/internal/guard/sha256.go) and compared against a sidecar <output-path>.hash
// from the previous call. A match skips the write entirely — the existing
// <output-path> and its .hash are left untouched — because a dispatch reads
// that body, and identical bytes need not be rewritten just to refresh a
// timestamp. Either outcome is reported to stderr as
// "gather-dispatch-context: bundle unchanged — reusing <output-path>" or
// "gather-dispatch-context: bundle rebuilt — <reason>", and this path always
// exits 0.
//
// THERE IS NO SPEC SOURCE, AND ADDING ONE BACK WOULD BE A MISTAKE. This
// script used to also carry every <change-root>/specs/*/spec.md — the
// OpenSpec delta specs. Under spectre a change has no specs/ subdirectory at
// all: spec edits go straight into <project>/spectre/specs/<capability>.md on
// the change's own branch, so there was nothing left for that glob to find
// and it reported "skipped: specs/*/spec.md (absent)" on every run of every
// change. The capability specs themselves are NOT bundled in its place —
// which ones a change touches is a judgement from its proposal, and this
// script is deterministic collection, not judgement; skills/flow/implement.md
// names them for the reader instead.
//
// `## standards` entries are NEVER read or carried here — the spec's own
// Requirement forbids it: they resolve through the entry-form table and
// containment rule in skills/flow-contracts/project-configuration-standards.md,
// belong to the principles slot alone, and re-implementing that containment
// logic here would duplicate the contract it depends on. This script globs
// nothing anywhere under <change-root>: it reads three leaves by exact name,
// so any other file placed there (a standards file, a stray note) never
// reaches the bundle.
//
// Exit 2 on a malformed invocation: a missing argument, a change name outside
// the allowlist below, or any of the three paths failing validation (see
// below). Exit 0 in every other case, including every missing source — a
// change may legitimately carry no design.md, and this script performs no
// judgement about that.
//
// VALIDATION. Modelled on scripts/the retired self-review gather's (kan-526) own
// mechanism — read that script's header before touching this one. Each of
// <worktree>, <change-root> and <principles-path> is:
//   1. lexically normalized into an absolute path by pure string
//      manipulation (a component-stack walk: push each `/`-separated
//      segment, pop on `..`, never past the root, skip `.` and empty
//      segments) — touching the filesystem nowhere, so it cannot itself be
//      fooled by a symlink;
//   2. semantically resolved by following every symlink at every component
//      (the leaf, via a bounded readlink loop, and every ancestor directory,
//      via `cd -P`), the resolver always given the LEXICAL form from step 1
//      rather than the raw argument, so a bare "." or a trailing slash —
//      neither of which step 1 ever produces — can never itself masquerade
//      as a divergence;
//   3. compared for EXACT equality against the lexical form.
// For <worktree> and <change-root>, any divergence found in step 3 — an
// ancestor symlink at any depth, the leaf itself being a symlink, a `..`
// component, in any combination — means the invocation is refused, exactly
// as the retired self-review gather's (kan-526) header documents: one general
// mechanism rather than several bounded, individually-bypassable checks.
// They must additionally resolve to an existing directory (not merely pass
// the lexical/real comparison), and <change-root> must resolve INSIDE
// <worktree> — or, when the seventh argument is given, inside
// <canonical-worktree>, which must then itself resolve to an existing
// directory with no divergence — a change-root outside both is a malformed
// invocation, not a legitimate absence.
// <principles-path> is normalized and resolved the same way but is NEVER
// refused for a divergence found in step 3, and carries no existence or
// containment requirement: it may be absent (reported as a skipped source,
// exit 0) or sit outside <worktree> (the global-install case above) or be
// reached through a symlink at any depth — a global install reaches every
// skill, and therefore every principles-path, through one by construction.
// It is a trusted argument the CALLER resolves (skills/flow/review-panel.md's
// own [PRINCIPLES_PATH] rule), never attacker-influenced content the way
// <change-root>'s leaves are, so it is simply resolved and read.
//
// THE CHANGE-NAME ALLOWLIST is copied verbatim from
// the retired self-review gather (kan-526): one leading alphanumeric, then letters,
// digits, '.', '_' and '-'.
//
// LEAF VALIDATION. The three CONTENT sources (proposal.md, design.md and
// tasks.md) are never passed as arguments, so the
// exit-2 VALIDATION above does not apply to them — a change legitimately
// carries no design.md, and a symlinked leaf must not abort a whole
// dispatching stage the way a malformed invocation does. Each is instead
// resolved and boundary-checked against <change-root> the same way
// the retired self-review gather's (kan-526) check_boundary() validates its own leaves
// (LEDGER_FILE, PANEL_FILE, TASKS_FILE) before reading them: a leaf that
// resolves outside <change-root> is refused per-source — omitted from the
// bundle, reported as "refused", but the run still exits 0 — never treated
// as a missing source, so an attack is never indistinguishable from a
// legitimate absence. See addFixedSource in
// stats/internal/guard/gatherdispatch.go.
//
// THE PLAN LIVES IN ONE TREE (KAN-393). proposal.md, design.md and tasks.md
// are read from the CONTENT DIRECTORY, resolved once before any leaf is
// read: the <change-root> itself whenever it carries a tasks.md — every
// single-repo change, whose bundle is therefore byte-for-byte what this
// script always wrote — and otherwise, when <change-root> carries a
// satellite's link.md (## Part of) and no tasks.md, the canonical change
// directory stats/internal/guard/changeplan.go's rules resolve through the optional
// seventh argument <canonical-worktree>, or — only when that argument is
// empty — through <worktree>/<spec-root>/peers, the same resolution order
// flow-guard check-unfinished-work and flow-guard check-task-commit-fields have used since
// KAN-363. The caller (skills/flow/implement.md's per-bundle gather,
// skills/flow/review-panel.md's rebuild) passes the canonical worktree on
// every call: the member of the run's resolved worktree set whose own
// change directory holds tasks.md (skills/flow-contracts/pipeline.md's
// "A change's directory"), inert on a single-repo change. Three consequences, each deliberate:
//
//   1. The boundary the three leaves are checked against is the CONTENT
//      DIRECTORY, not the argument <change-root>: the canonical directory is
//      the same change's tracked content, reached through a link whose peer
//      name and change id changeplan.go allowlist-checks character for
//      character. The invocation-level containment rules above are
//      unchanged — <change-root> must still sit inside <worktree> or
//      <canonical-worktree>, exit 2.
//   2. A satellite whose canonical plan cannot be reached is NOT a refusal:
//      the three plan leaves are reported as skipped with the distinct
//      label "(satellite plan unresolved — link.md at <path>)" and the run
//      still exits 0. This is the deliberate inverse of
//      flow-guard check-unfinished-work's exit-2 refusal for the same condition — a
//      guard verdict must never silently clear, but this bundle has no
//      verdict to give: it never gates a run (skills/flow/implement.md), so
//      refusing here would abort a dispatching stage a bundle has no
//      authority to stop. A change with no tasks.md and no link.md at all
//      is not a satellite, and its absent leaves are still reported exactly
//      as before.
//   3. When the plan resolved remotely, the three plan sections' labels
//      carry a "(canonical <peer>:<change-id>)" suffix — composed with the
//      existing "(scoped to task(s) ...)" suffix where both apply — so a
//      dispatch reading the bundle can see the plan came from another
//      tree. Everything else stays worktree-local: "project commands" from
//      <worktree>/.flow/project.md, the incidents section from
//      `flow record incidents -C <worktree>`, and the header's head: sha
//      from <worktree>'s own HEAD — a satellite implementer's bundle
//      carries its own repository's commands, its own project's incident
//      log, and the canonical plan.
//
// THE HAZARDS SECTION (KAN-452). After "## incidents" the bundle renders
// "## hazards", sourced from `flow hazards -C <worktree> -shape <shape>`:
// this project's proactive warnings -- incidents' proactive sibling, one
// "- **name (applies):** body" bullet per row, part of the hashed BODY so a
// newly recorded hazard forces a rebuild. <shape> is the change's shape the
// CALLER computes (its resolved worktree set spans more than one repository
// -> cross-repo, otherwise single-repo); the empty/omitted eighth argument
// becomes the literal "all", so a caller that does not know the shape still
// fails open to the always-on rows rather than over-injecting. Anything
// outside the closed vocabulary is a malformed invocation (exit 2), the
// same contract the task-ids argument carries. The section never gates a
// run: flow absent or failing skips as "hazards (flow unavailable)", an
// empty result as "hazards (none)".
//
// <canonical-worktree>, when given, is validated nowhere beyond what
// changeplan.go's own file probes imply and the change-root containment
// above — the same trust level as
// flow-guard check-unfinished-work's third argument: a caller-supplied worktree
// path, not attacker-influenced change content, and every name concatenated
// under it (the canonical change id) is allowlist-checked inside
// changeplan.go itself.

// The port of gather-dispatch-context.sh, whose header (the comment above) is canonical
// for the contract: arguments, VALIDATION, LEAF VALIDATION, the content
// directory, the incidents and hazards sections, and SKIP-WHEN-UNCHANGED.
// The helpers it sourced -- the since-retired lib/resolve-file.sh, twinned in resolvefile.go;
// the section reader now in projectsection.go; and
// the since-deleted lib/within-root.sh and lib/lexical-normalize.sh -- are
// ported below as far as this guard uses them.
//
// Two bash dependencies are gone: the hash comes from crypto/sha256, so the
// "no hash tool available" rebuild cannot happen, and flow's JSON is parsed
// here, so jq is not required for the incidents and hazards sections. The
// reasoning for each branch, moved here from the bash body it replaced
// (0747740), sits beside the code it explains. The bash sourced its helpers
// from scripts/lib/ because it shipped through the skills/*/scripts/ symlink
// farm — the criterion each library's own header states for when a guard may
// source a sibling rather than carry its own copy; the port carries Go
// copies of those helpers instead, and no longer sources any of them.

func init() { Registry["gather-dispatch-context"] = gatherDispatchContext }

const gdcUsage = "usage: flow-guard gather-dispatch-context <worktree> <change-root> <name> <principles-path> <output-path> [<task-ids> [<canonical-worktree>]]\n"

var gdcTaskIDs = regexp.MustCompile(`^[0-9]+(,[0-9]+)*$`)

// gdcUnsplitTaskIDsHint is gather-dispatch-context's own remedy: its task ids
// are one comma-separated argument, so splitting them into several would shift
// the later positionals.
const gdcUnsplitTaskIDsHint = "looks like a list passed as one argument — zsh does not word-split an unquoted $VAR; join the task ids with commas, no spaces"

func gatherDispatchContext(args []string, env Env, _, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, changeRoot, name, principles, output := arg(0), arg(1), arg(2), arg(3), arg(4)
	taskIDs, canonical, shape := arg(5), arg(6), arg(7)

	if worktree == "" || changeRoot == "" || name == "" || principles == "" || output == "" {
		fmt.Fprint(stderr, gdcUsage)
		return 2
	}
	if taskIDs != "" && !gdcTaskIDs.MatchString(taskIDs) {
		hint := ""
		if strings.ContainsAny(taskIDs, " \t\n") {
			hint = " — " + gdcUnsplitTaskIDsHint
		}
		fmt.Fprintf(stderr, "gather-dispatch-context: task ids '%s' must be a comma-separated list of integers%s\n", taskIDs, hint)
		return 2
	}
	switch shape {
	case "", "all", "cross-repo", "single-repo":
	default:
		fmt.Fprintf(stderr, "gather-dispatch-context: shape '%s' must be one of all, cross-repo, single-repo\n", shape)
		return 2
	}
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "gather-dispatch-context: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return 2
	}

	// --- worktree: must resolve to an existing, non-symlinked directory ---
	worktreeReal, mismatch := validatePath(env, worktree)
	if worktreeReal == "" || !isDir(worktreeReal) {
		fmt.Fprintf(stderr, "gather-dispatch-context: worktree '%s' does not resolve to an existing directory\n", worktree)
		return 2
	}
	if mismatch {
		fmt.Fprintf(stderr, "gather-dispatch-context: worktree '%s' resolves through a symlink somewhere between the repository root and the leaf\n", worktree)
		return 2
	}
	// --- change-root: must resolve to an existing, non-symlinked directory
	// inside the (already-validated) worktree ---
	changeRootReal, mismatch := validatePath(env, changeRoot)
	if changeRootReal == "" || !isDir(changeRootReal) {
		fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' does not resolve to an existing directory\n", changeRoot)
		return 2
	}
	if mismatch {
		fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' resolves through a symlink somewhere between the repository root and the leaf\n", changeRoot)
		return 2
	}
	// An ## apps worktree created without a spectre link
	// (skills/flow/cross-repo-worktrees.md) holds no change directory, so
	// its change-root is the canonical worktree's: containment is then
	// checked against <canonical-worktree>, validated like <worktree>.
	inCanonical := false
	if canonical != "" {
		real, mismatch := validatePath(env, canonical)
		inCanonical = real != "" && !mismatch && isDir(real) && withinRoot(changeRootReal, real)
	}
	if !withinRoot(changeRootReal, worktreeReal) && !inCanonical {
		if canonical != "" {
			fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' resolves outside the worktree '%s' and the canonical worktree '%s'\n", changeRoot, worktree, canonical)
		} else {
			fmt.Fprintf(stderr, "gather-dispatch-context: change-root '%s' resolves outside the worktree '%s'\n", changeRoot, worktree)
		}
		return 2
	}
	// --- principles-path: normalized and resolved the same way, but never
	// required to exist, never checked for containment under the worktree (it
	// may be a global install path outside it entirely), and NEVER refused for
	// a symlink divergence (F1). setup.sh global installs every skill as a
	// symlink (~/.claude/skills/flow -> the repository's skills/flow/), so
	// [PRINCIPLES_PATH] — resolved by the CALLER from its own installed skill
	// directory (skills/flow/review-panel.md's own rule), never derived here —
	// ALWAYS diverges lexically from its real path in that install shape. The
	// lexical/real divergence refusal exists to stop a change directory's own
	// (repo-tracked, pull-request-editable, attacker-influenced) content
	// escaping <change-root>; principles-path is not that — it is a trusted
	// argument the dispatching skill resolves for itself, so it is simply
	// resolved (following every symlink, leaf and ancestor alike) and read.
	// Containment and divergence refusal are UNCHANGED for <worktree> and
	// <change-root> above, and for the leaf content sources below. ---
	principlesReal, _ := validatePath(env, principles)

	// The bundle itself.
	b := &gdcBundle{}

	// The content directory (KAN-393): where this change's plan leaves
	// actually live. The <change-root> itself whenever it carries a tasks.md —
	// the single-repo case, byte-for-byte unchanged — and otherwise, when it
	// carries a satellite's link.md, the canonical change directory
	// changePlanDir resolves through the seventh argument or, only when that
	// argument is empty, through <worktree>/<spec-root>/peers. When even that
	// fails, the change IS a satellite (its link.md carries ## Part of — the
	// same definition check-unfinished-work.sh applies) and the three plan
	// leaves below are reported as skips with the distinct unresolved label;
	// a change with neither file is not a satellite, and its leaves keep the
	// ordinary absent-leaf handling. changePlanRef's answer is the
	// <peer>:<change-id> ref of the link the successful resolution already
	// validated, feeding the section label. Both live in changeplan.go, the
	// owner of link.md's grammar.
	contentDir, suffix, satelliteLink := changeRootReal, "", ""
	if !isFile(changeRootReal + "/tasks.md") {
		// A relative canonical stays relative, as in bash: changePlanDir
		// probes it against the process cwd, and every leaf found under it
		// is then refused, its resolved path never within the relative root.
		if dir, rc := changePlanDir(env, io.Discard, worktreeReal, name, canonical); rc == 0 && dir != "" {
			contentDir = dir
			if ref, ok := changePlanRef(io.Discard, worktreeReal, name); ok {
				suffix = " (canonical " + ref + ")"
			}
		} else if link := changeRootReal + "/link.md"; isFile(link) && gdcHasPartOf(link) {
			contentDir, satelliteLink = "", link
		}
	}

	if contentDir != "" {
		for _, leaf := range []string{"proposal.md", "design.md", "tasks.md"} {
			b.addFixedSource(contentDir+"/"+leaf, leaf+suffix, contentDir)
		}
	} else {
		// A satellite whose canonical plan could not be reached: the bundle is
		// advisory and never gates a run, so this is a skip, not a refusal —
		// with its own label, so "the plan is elsewhere and unreachable" never
		// reads like a plain change's legitimately-absent design.md (KAN-393).
		for _, leaf := range []string{"proposal.md", "design.md", "tasks.md"} {
			b.skipped = append(b.skipped, leaf+" (satellite plan unresolved — link.md at "+satelliteLink+")")
		}
	}

	if taskIDs != "" {
		last := len(b.found) - 1
		if last < 0 || b.found[last].label != "tasks.md"+suffix {
			fmt.Fprintln(stderr, "gather-dispatch-context: task ids given but tasks.md is absent or refused")
			return 2
		}
		body, ok := scopeTasks(b.found[last].path, taskIDs, stderr)
		if !ok {
			return 2
		}
		b.found[last] = gdcSource{label: "tasks.md" + suffix + " (scoped to task(s) " + taskIDs + ")", text: body + "\n"}
	}

	if principlesReal != "" && isFile(principlesReal) {
		b.found = append(b.found, gdcSource{label: principles, path: principlesReal})
	} else {
		b.skipped = append(b.skipped, principles+" (absent)")
	}

	b.addProjectCommands(worktreeReal)
	b.addIncidents(env, worktreeReal)
	// The change's shape, computed by the caller from its resolved worktree
	// set. Empty means the caller does not know it; the hazards section then
	// filters to the always-on rows (the literal "all"), never to a guess.
	if shape == "" {
		shape = "all"
	}
	b.addHazards(env, worktreeReal, shape)

	body := strings.TrimRight(b.render(), "\n") // `BODY="$(render_body)"`
	// The hashing primitive SKIP-WHEN-UNCHANGED compares the body against:
	// sha256Hex (sha256.go), the repository's one helper, shared with
	// check-cleanup-complete (F1, this change's own review panel) — once a
	// second caller existed, rules/build-the-simplest-thing.mdc's "no
	// abstraction until a second caller exists" called for extracting the
	// shared helper rather than leaving two copies free to drift apart,
	// exactly the failure within_root suffered as two bash copies before
	// its extraction.
	newHash := sha256Hex(body)
	outputAbs := output
	if !strings.HasPrefix(outputAbs, "/") {
		outputAbs = env.Dir + "/" + outputAbs
	}
	hashPath := outputAbs + ".hash"

	reason := ""
	if !isFile(hashPath) || !isFile(outputAbs) {
		reason = "no cached bundle"
	} else if old, _ := os.ReadFile(hashPath); strings.TrimRight(string(old), "\n") != newHash {
		reason = "inputs changed"
	}
	if reason == "" {
		fmt.Fprintf(stderr, "gather-dispatch-context: bundle unchanged — reusing %s\n", output)
		return 0
	}

	head, ok := gitOut(env, "-C", worktreeReal, "rev-parse", "--short", "HEAD")
	if !ok {
		head = "unknown"
	}
	// The body's own trailing newline(s) are stripped above, as the bash's
	// `$(render_body)` stripped them (F2, this change's own review panel) —
	// the hash comparison is unaffected, since newHash is derived from that
	// same stripped body on every call, but the file this writes must still
	// end with a newline the way the pre-kan-288 stdout-redirected script
	// always did. The final "\n" restores exactly that: one trailing newline,
	// appended after the body's own content rather than baked into what gets
	// hashed.
	file := "# Dispatch context bundle for " + name + "\n" +
		"generated: " + time.Now().UTC().Format("2006-01-02T15:04:05Z") + "\n" +
		"head: " + head + "\n\n" + body + "\n"
	if err := os.MkdirAll(gdcDirname(outputAbs), 0o777); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outputAbs, []byte(file), 0o666); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	if err := os.WriteFile(hashPath, []byte(newHash), 0o666); err != nil {
		fmt.Fprintf(stderr, "gather-dispatch-context: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "gather-dispatch-context: bundle rebuilt — %s\n", reason)
	return 0
}

// gdcSource is one "## <label>" section: a file's bytes (path) or text
// already rendered, printed verbatim.
type gdcSource struct{ label, path, text string }

// gdcRefusal carries a reason per refused label — "outside_change" (or
// "outside_worktree" for project.md) for a genuine withinRoot containment
// breach, "unresolvable" for a resolveFile failure (a symlink loop past its
// own 40-hop cap, or a component that could not be walked, e.g. a race or a
// permissions failure) (F36, pass 7 of this change's own review panel).
// Before this, both cases were folded into one bucket and every one was
// reported with the same "(resolves outside the change directory)" text —
// true for the first case, and simply wrong for the second: a source
// resolveFile could not even walk was never compared against the change root
// at all, so nothing about it says "outside". Both dispositions still omit
// the source from the bundle and still let the run exit 0 — only the printed
// diagnostic differs.
type gdcRefusal struct{ label, reason string }

// gdcBundle's every skipped entry must already be the complete "skipped: "
// line's payload — either self-describing (e.g. "incidents (none)") or a bare
// identifier with " (absent)" already appended. render only echoes each entry
// verbatim; it does no further formatting.
type gdcBundle struct {
	found   []gdcSource
	skipped []string // each the complete payload of its "skipped: " line
	refused []gdcRefusal
}

// addFixedSource is add_fixed_source: absent when not a file, refused when
// its path cannot be walked or resolves outside root, found otherwise.
//
// Unlike <worktree>/<change-root>/<principles-path>, these content sources are
// never passed as arguments, so validatePath's exit-2-on-mismatch contract
// does not apply to them: a change legitimately carries no design.md, and
// "this leaf is a symlink" must not abort a whole dispatching stage. Instead
// this mirrors the retired self-review gather's (kan-526) check_boundary() +
// is_refused() pair: a regular-file test first (a missing target, including a
// dangling symlink, is a plain absence — skipped, not refused); then
// resolveFile (leaf and every ancestor symlink) and a withinRoot boundary
// check against root — the CONTENT DIRECTORY (KAN-393): the change-root itself
// for a local plan, the canonical change directory for a satellite, so a leaf
// can never resolve outside the change whose content it is. A leaf that
// resolves inside that root — including a symlink to another file inside it —
// is read normally. A leaf that resolveFile cannot even walk ("unresolvable"),
// or that resolves OUTSIDE it ("outside_change"), is refused per-source:
// omitted from the bundle, counted separately from "skipped", and the run
// still exits 0 — the spec's "a missing bundle never stops a run"
// requirement, and the same disposition the retired self-review gather's
// (kan-526) header documented for exactly this shape of problem (a per-source
// trust-boundary violation is not a malformed invocation).
func (b *gdcBundle) addFixedSource(path, label, root string) {
	if !isFile(path) {
		b.skipped = append(b.skipped, label+" (absent)")
		return
	}
	resolved, ok := resolveFile(path)
	switch {
	case !ok:
		b.refused = append(b.refused, gdcRefusal{label, "unresolvable"})
	case withinRoot(resolved, root):
		b.found = append(b.found, gdcSource{label: label, path: resolved})
	default:
		b.refused = append(b.refused, gdcRefusal{label, "outside_change"})
	}
}

// addProjectCommands carries the ## lint, ## test and ## run sections of
// <worktree>/.flow/project.md, so a dispatched subagent already carries this
// project's lint/test/run commands and never needs to open project.md itself
// (CLAUDE.md's lint-fix-priority rule sends every subagent there). Extracted,
// not the whole 20+ KB file, per design.md's
// scoped-dispatch-bundle-not-full-project-md decision. Appended after the
// principles file section, as the last "found" entry. projectSection is in
// projectsection.go.
func (b *gdcBundle) addProjectCommands(worktree string) {
	file := worktree + "/.flow/project.md"
	if !isFile(file) {
		b.skipped = append(b.skipped, "project commands (absent)")
		return
	}
	resolved, ok := resolveFile(file)
	if !ok {
		b.refused = append(b.refused, gdcRefusal{"project commands", "unresolvable"})
		return
	}
	if !withinRoot(resolved, worktree) {
		b.refused = append(b.refused, gdcRefusal{"project commands", "outside_worktree"})
		return
	}
	var text strings.Builder
	for _, key := range []string{"lint", "test", "run"} {
		section := projectSection(resolved, key)
		if strings.Trim(section, gdcSpace+"\n") != "" {
			text.WriteString("### " + key + "\n\n" + section + "\n\n")
		}
	}
	if text.Len() == 0 {
		b.skipped = append(b.skipped, "project commands (absent)")
		return
	}
	b.found = append(b.found, gdcSource{label: "project commands", text: text.String()})
}

// gdcIncident is one row of `flow record incidents`. minutesLost keeps its
// JSON literal, as jq's tostring prints it; change is null when unset.
type gdcIncident struct {
	OccurredAt  string      `json:"occurredAt"`
	Guard       string      `json:"guard"`
	Symptom     string      `json:"symptom"`
	Recovery    string      `json:"recovery"`
	MinutesLost json.Number `json:"minutesLost"`
	Change      *string     `json:"change"`
}

// addIncidents is this project's guard-incident log, from
// `flow record incidents -C <worktree>`, so every dispatch already carries the
// hazards prior runs hit instead of relying on someone re-reading memory
// before dispatching (KAN-451). Three outcomes: `flow` absent or the call
// failing skips as "incidents (flow unavailable)"; a `[]` result skips as
// "incidents (none)"; anything else renders the "## incidents" section, a
// six-column table. The section is part of the hashed body, so a newly
// recorded incident forces a rebuild.
func (b *gdcBundle) addIncidents(env Env, worktree string) {
	var rows []gdcIncident
	if !gdcFlowJSON(env, &rows, "record", "incidents", "-C", worktree) {
		b.skipped = append(b.skipped, "incidents (flow unavailable)")
		return
	}
	if len(rows) == 0 {
		b.skipped = append(b.skipped, "incidents (none)")
		return
	}
	pipe := strings.NewReplacer("|", `\|`)
	var text strings.Builder
	text.WriteString("| when | guard | symptom | recovery | minutes lost | change |\n|------|-------|---------|----------|--------------|--------|\n")
	for _, r := range rows {
		when := r.OccurredAt
		if len(when) > 10 {
			when = when[:10]
		}
		change := "—"
		if r.Change != nil {
			change = *r.Change
		}
		text.WriteString("| " + strings.Join([]string{when, r.Guard, pipe.Replace(r.Symptom), pipe.Replace(r.Recovery), r.MinutesLost.String(), change}, " | ") + " |\n")
	}
	b.found = append(b.found, gdcSource{label: "incidents", text: strings.TrimRight(text.String(), "\n") + "\n"})
}

// addHazards is this project's proactive warnings, from
// `flow hazards -C <worktree> -shape <shape>` (KAN-452) — the incidents
// section's three-outcome shape: `flow` absent or the call failing skips as
// "hazards (flow unavailable)"; a `[]` result skips as "hazards (none)";
// anything else renders the "## hazards" section, one bullet per row. The
// section is part of the hashed body, so a newly recorded hazard forces a
// rebuild. An empty shape arrives here as the literal "all" — fail-open to
// the always-on rows, per the header's HAZARDS paragraph.
func (b *gdcBundle) addHazards(env Env, worktree, shape string) {
	var rows []struct{ Name, Applies, Body string }
	if !gdcFlowJSON(env, &rows, "hazards", "-C", worktree, "-shape", shape) {
		b.skipped = append(b.skipped, "hazards (flow unavailable)")
		return
	}
	if len(rows) == 0 {
		b.skipped = append(b.skipped, "hazards (none)")
		return
	}
	var text strings.Builder
	for _, r := range rows {
		text.WriteString("- **" + r.Name + " (" + r.Applies + "):** " + r.Body + "\n")
	}
	b.found = append(b.found, gdcSource{label: "hazards", text: strings.TrimRight(text.String(), "\n") + "\n"})
}

// gdcFlowJSON runs `flow <args>` from env's PATH and decodes its stdout into
// v: false when flow is absent, fails, or prints no JSON array. JSON null is
// jq's length 0, an empty result.
func gdcFlowJSON(env Env, v any, args ...string) bool {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return false
	}
	cmd := exec.Command(flow, args...)
	cmd.Dir = env.Dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return json.Unmarshal(out, v) == nil
}

// render is render_body: the census, the refused and skipped lines, then
// every found section — everything printed after the header's
// `generated:`/`head:` lines. Rendered once and hashed, so the header's own
// per-call metadata (the timestamp, the HEAD sha) never enters the comparison
// — see SKIP-WHEN-UNCHANGED in the header and design.md's
// header-excluded-from-hash decision.
func (b *gdcBundle) render() string {
	var s strings.Builder
	fmt.Fprintf(&s, "found: %d source(s); skipped: %d source(s); refused: %d source(s)\n", len(b.found), len(b.skipped), len(b.refused))
	for _, r := range b.refused {
		switch r.reason {
		case "outside_change":
			s.WriteString("refused: " + r.label + " (resolves outside the change directory)\n")
		case "outside_worktree":
			s.WriteString("refused: " + r.label + " (resolves outside the worktree)\n")
		default:
			s.WriteString("refused: " + r.label + " (could not be resolved: a symlink loop or other failure walking its path)\n")
		}
	}
	for _, l := range b.skipped {
		s.WriteString("skipped: " + l + "\n")
	}
	s.WriteString("\n")
	for _, src := range b.found {
		s.WriteString("## " + src.label + "\n\n")
		if src.path != "" {
			content, _ := os.ReadFile(src.path)
			s.Write(content)
		} else {
			s.WriteString(src.text)
		}
		s.WriteString("\n")
	}
	return s.String()
}

// scopeTasks is scope_tasks: the plan's header plus the named tasks' blocks
// in document order, trailing newlines stripped as `$(...)` strips them.
// Each named id with no task line is reported on stderr, and ok is false.
// Task-line, body-boundary and fence grammar are planparse.go's
// (pdbParseTasks), so the ids plan-dispatch-bundles printed are the ids this finds.
func scopeTasks(path, ids string, stderr io.Writer) (string, bool) {
	content, _ := os.ReadFile(path)
	want := strings.Split(ids, ",")
	wanted := map[string]bool{}
	for _, id := range want {
		wanted[id] = true
	}
	seen := map[string]bool{}
	fence, header, cur := false, true, ""
	var out []string
	for _, line := range lines(content) {
		if gdcFence.MatchString(line) {
			fence = !fence
			if header || cur != "" {
				out = append(out, line)
			}
			continue
		}
		if m := gdcTaskLine.FindStringSubmatch(line); !fence && m != nil {
			header = false
			cur = ""
			if wanted[m[1]] {
				cur = m[1]
				seen[cur] = true
			}
		} else if !fence && !header && gdcSectionHeading.MatchString(line) {
			cur = ""
		}
		if header || cur != "" {
			out = append(out, line)
		}
	}
	ok := true
	for _, id := range want {
		if !seen[id] {
			fmt.Fprintf(stderr, "gather-dispatch-context: task %s not found in tasks.md\n", id)
			ok = false
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n"), ok
}

var (
	gdcFence          = regexp.MustCompile("^ {0,3}(```|~~~)")
	gdcTaskLine       = regexp.MustCompile(`^- \[[ x]\] ([0-9]+)\. `)
	gdcSectionHeading = regexp.MustCompile(`^##(#)?([ \t]|$)`)
)

// gdcSpace is awk's and tr's [:space:] less the newline lines never carry.
const gdcSpace = " \t\v\f\r"

// gdcHasPartOf is `grep -qxF '## Part of'`.
func gdcHasPartOf(file string) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	for _, line := range lines(content) {
		if line == "## Part of" {
			return true
		}
	}
	return false
}

// validatePath is the header's VALIDATION mechanism: the lexical form of
// path (made absolute against env.Dir), its fully resolved form ("" when
// resolution fails — a plain absence), and whether the two differ, i.e. a
// symlink sits somewhere between an existing ancestor and the leaf. The one
// mechanism the header documents: normalize, resolve, compare.
func validatePath(env Env, path string) (real string, mismatch bool) {
	if !strings.HasPrefix(path, "/") {
		path = env.Dir + "/" + path
	}
	lexical := lexicallyCollapse(path)
	// resolveFile is given the already lexically-normalized form, not the raw
	// argument (F2): the bash resolve_file's dirname/basename split mishandled
	// a bare "." (the result gained a spurious trailing "/.") and a
	// trailing-slash path (a spurious trailing "/") — neither of which
	// lexicallyCollapse ever produces, so either form was wrongly compared as
	// a "divergence" against a clean lexical path even with no symlink
	// anywhere in the chain. The lexical form is always absolute with no "."
	// segment and no trailing slash, so resolveFile never sees either shape
	// here. resolveFile itself also normalizes the trailing slash on its own
	// input (F9, the retired lib/resolve-file.sh's own "NORMALIZES ITS OWN INPUT"
	// comment), so this pre-normalization is belt-and-braces, not the only
	// thing standing between this function and the bug.
	real, ok := resolveFile(lexical)
	if !ok {
		return "", false
	}
	return real, real != lexical
}

// lexicallyCollapse is the port of the deleted lib/lexical-normalize.sh's
// lexically_collapse: drop
// empty and "." segments, pop on "..", never past the root — pure string
// manipulation with no filesystem access. validatePath makes a relative path
// absolute against the process's own directory first; only that "make it
// absolute" half stays local to each caller, the collapse being shared (F34,
// this change's own review panel), so it cannot be fooled by a symlink.
func lexicallyCollapse(abs string) string {
	var stack []string
	for _, part := range strings.Split(abs, "/") {
		switch part {
		case "", ".":
		case "..":
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		default:
			stack = append(stack, part)
		}
	}
	return "/" + strings.Join(stack, "/")
}

// withinRoot is the port of the deleted lib/within-root.sh's within_root: a
// path BOUNDARY test, not a string prefix test (a bare prefix would wrongly
// accept "/foo/bar-evil" against "/foo/bar").
func withinRoot(p, root string) bool {
	return p == root || strings.HasPrefix(p, root+"/")
}
