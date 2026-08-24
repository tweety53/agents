package records

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// This file is the whole of the Markdown side of the run record. Every
// record a change carries is now rows in the store, and these two
// functions are what makes the archive readable without a daemon: the
// SDD ledger and the review panel record are GENERATED, never
// hand-written, so there is one source and no second copy to disagree
// with it.
//
// RenderPanel is the load-bearing one. scripts/check-unfinished-work.sh
// is unchanged by the move into the store, so the marker contract it
// enforces -- the single `findings-total:` line, the unbroken
// `finding-status:` span, and the SEPARATE `reproducers-total:` /
// `finding-reproducer:` block that check-panel-reproducers.sh reads --
// is now this file's obligation rather than an agent's discipline.

// markerLabels are the marker labels the two panel guards read. They are
// listed here because neutraliseMarkers has to know them, not as a second
// definition of the marker format: the format itself is stated in
// skills/myflow-do/SKILL.md and enforced by the two guards, and this list
// is only the set of prefixes whose colon must not survive into free
// text.
var markerLabels = []string{
	"findings-total",
	"finding-status",
	"reproducers-total",
	"finding-reproducer",
}

// markerColonSubstitute is U+A789 MODIFIER LETTER COLON, which reads as a
// colon and is not one. See neutraliseMarkers for why the substitution
// exists at all.
const markerColonSubstitute = "꞉"

// neutraliseMarkers makes operator-authored text safe to place in a
// rendered record.
//
// SUBSTITUTION: after any of the marker labels above, an ASCII colon
// (U+003A) is replaced by U+A789 MODIFIER LETTER COLON, so that
// `finding-status: F9 fixed` written inside a note renders as
// `finding-status<U+A789> F9 fixed`. It still reads as the text the
// operator wrote and matches none of the guards' patterns.
//
// WHY IT IS NEEDED. A finding's note and location are operator text. A
// validly-formatted marker written inside prose is byte-for-byte
// indistinguishable from a real one: quoted in a table cell it makes
// check-unfinished-work.sh report a line naming `finding-status:` that is
// not a marker, and reaching the start of a line it would add a finding
// that does not exist. The record's own format rule -- do not quote the
// marker format inside the record -- is exactly this hazard, and the
// renderer is now the only writer that can honour it.
//
// WHY ON THE WAY OUT. The store keeps the operator's text verbatim. A
// note is evidence, and rewriting it on the way in would destroy the
// only copy of what the slot actually said; the rendering is derived and
// can be regenerated, so it is the safe place to alter.
//
// Line breaks are flattened to spaces in the same pass, for the same
// reason and one more: an embedded newline both breaks the Markdown table
// and can carry the second half of a note to the start of a line, and it
// would break the unbroken marker span if it appeared in a status or a
// reproducer.
func neutraliseMarkers(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	for _, label := range markerLabels {
		s = strings.ReplaceAll(s, label+":", label+markerColonSubstitute)
	}
	return s
}

// tableCell is neutraliseMarkers plus the one escape a GFM table needs: an
// unescaped pipe in a cell shifts every column after it. Neither guard
// parses the table, so this is for the reader rather than for them.
func tableCell(s string) string {
	if s == "" {
		return " "
	}
	return strings.ReplaceAll(neutraliseMarkers(s), "|", `\|`)
}

// RenderPanel renders a change's findings as the review panel record
// /myflow-do writes to .superpowers/sdd/final-review-panel.md and the two
// panel guards read.
//
// The layout is fixed by what those guards require, and each part of it
// is load-bearing:
//
//   - the findings table, one row per finding, each beginning `| F<n> |`
//     so that check-unfinished-work.sh's anchored row pattern names the
//     same identifiers the marker block does;
//   - one `findings-total: <n>` line, appearing exactly once;
//   - the `finding-status:` markers on CONSECUTIVE lines, one unbroken
//     span, so that a marker written anywhere else in the record cannot
//     stand in for a missing one;
//   - a blank line, and then the reproducer block kept SEPARATE, because
//     interleaving the two would break that span.
//
// Findings render in the order the store returned them, which is the
// order their refs' digits spell; the guards compare identifiers as
// sorted lists, so the order is for the reader alone.
//
// A finding with no recorded reproducer renders the exemption form
// check-panel-reproducers.sh defines, `none — <reason>`, rather than no
// line at all: that guard requires exactly one reproducer line per
// finding, and an omitted line would be reported as a missing one.
func RenderPanel(r Run) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Review panel — %s\n\n", neutraliseMarkers(r.Change))
	b.WriteString("Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.\n\n")

	b.WriteString("| ID | Slot | Severity | Location | Note |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
			tableCell(f.Ref), tableCell(f.Slot), tableCell(f.Severity),
			tableCell(f.Location), tableCell(f.Note))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "findings-total: %d\n", len(r.Findings))
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "finding-status: %s %s\n", neutraliseMarkers(f.Ref), neutraliseMarkers(f.Status))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "reproducers-total: %d\n", len(r.Findings))
	for _, f := range r.Findings {
		repro := neutraliseMarkers(f.Reproducer)
		if strings.TrimSpace(repro) == "" {
			repro = "none — no reproducer was recorded for this finding"
		}
		fmt.Fprintf(&b, "finding-reproducer: %s %s\n", neutraliseMarkers(f.Ref), repro)
	}

	return b.String()
}

// dispatchMetrics is the part of a dispatch's metrics bag the ledger
// reads: the token figures internal/harvest's second attribution pass
// merges in, under the bag's top-level "tokens" key.
//
// It is declared here rather than shared with internal/harvest's
// MetricsPatch on purpose. This package is the wire shape every layer
// depends on and imports nothing but encoding/json and time; importing
// the harvester into it to reuse a struct would invert the dependency for
// four field names. The duplication is deliberate and narrow: these are
// the keys as READ, and a harvester that renamed one would show up here as
// a figure that stopped rendering rather than as a compile error.
//
// Tokens is a POINTER, and that is the whole of how an unmeasured dispatch
// is told from one measured at zero. Unattributed is a POINTER for the
// same reason: a bag with no "unattributed" key at all -- the ordinary
// shape of everything this task's producers have not stamped -- must stay
// distinguishable from an explicit, empty one. See tokenLine.
type dispatchMetrics struct {
	Tokens       *tokenTotals  `json:"tokens"`
	Unattributed *unattributed `json:"unattributed"`
}

// unattributed is the bag's "unattributed" object: the reason
// internal/store's MarkDispatchesUnattributed /
// MarkDispatchesUnattributedByID and internal/harvest/watcher.go's
// resolveSessionTokens / attributeDispatches recorded for a dispatch whose
// cost could not be attributed, plus the candidate count for the two
// reasons that are ambiguities. Candidates is meaningless, and left at its
// zero value, for reasonSessionNeverBound, which is not an ambiguity.
type unattributed struct {
	Reason     string `json:"reason"`
	Candidates int    `json:"candidates"`
}

// The three reason strings a producer writes, mirrored from
// internal/harvest/watcher.go's unexported constants of the same name.
// tokenLine keys its wording off these exact strings -- never off
// Candidates > 0 -- because reasonSessionAmbiguous and
// reasonDispatchAmbiguous both carry a nonzero Candidates and count
// different things: one counts sessions a token matched, the other counts
// dispatches a record could not be told apart from. Rendering a session
// count behind the words "concurrent dispatches" would misreport the
// measurement rather than abbreviate it.
const (
	reasonSessionNeverBound = "session never bound"
	reasonSessionAmbiguous  = "matched more than one session"
	reasonDispatchAmbiguous = "matched more than one dispatch"
)

// tokenTotals is the bag's "tokens" object: the harvester's two buckets.
type tokenTotals struct {
	Main      tokenBucket `json:"main"`
	Sidechain tokenBucket `json:"sidechain"`
}

type tokenBucket struct {
	Input         int64 `json:"input"`
	Output        int64 `json:"output"`
	CacheCreation int64 `json:"cache_creation"`
	CacheRead     int64 `json:"cache_read"`
}

// tokenLine renders a dispatch's token figures, or the words that say why
// there are none.
//
// A dispatch nothing measured renders `not measured`, NEVER zero: the
// harvester attributes a dispatch's cost from the harness transcript
// afterwards, and two of the three supported harnesses write no
// transcript at all. Zero is a measurement, and reporting it for a
// dispatch nothing measured would be a figure the reader could not tell
// from a real one.
//
// "NOTHING MEASURED" IS THE ABSENCE OF THE `tokens` KEY, NOT THE ABSENCE
// OF BYTES. internal/store's insertDispatch defaults an empty Metrics to
// the JSON object `{}` on the way in -- never zero-length, never SQL NULL
// -- so `{}` is the shape of every dispatch between `myflow record
// dispatch` and the harvester running, and the PERMANENT shape of every
// dispatch on Cursor and Codex, which write no transcript at all. It
// unmarshals silently into a zero-valued struct, so a `len(raw) == 0` test
// would report `input 0, output 0, ...` for the majority of real rows
// while agreeing with every hand-built Dispatch{} in a test. Hence the
// pointer on dispatchMetrics.Tokens: a bag with no `tokens` key, an
// explicit `"tokens": null`, `{}`, `null` and whitespace all leave it nil.
//
// A `tokens` object that IS present renders whatever it carries, zeros
// included. An explicit zero is a real figure and is not hidden.
//
// TOKENS OUTRANKS UNATTRIBUTED. A dispatch can carry both: a session
// marked given up after a dispatch under it was already measured is a
// contradiction the store resolves by keeping the measurement (see
// internal/store's MarkDispatchesUnattributed), and a real figure outranks
// a stale stamp here too. Only once Tokens is nil does an Unattributed
// stamp get a say.
//
// THE FOUR STATES. Where Unattributed is present, its Reason selects the
// wording: reasonSessionNeverBound renders with no count,
// reasonSessionAmbiguous and reasonDispatchAmbiguous each render with
// Candidates but under DIFFERENT wordings, because they count different
// things -- see the constants' own comment. A reason none of the three
// matches is not `not measured`: it is still a producer's statement that
// attribution failed, so it renders verbatim (through neutraliseMarkers,
// like every other externally-sourced string in this file) rather than
// being collapsed into the state that means nothing was ever attempted. An
// Unattributed bag with no recognised key at all -- Reason "" -- falls
// back to `not measured`, the same as no bag.
//
// Main and sidechain are summed. A dispatch's own work is the sidechain
// bucket, but the row is the dispatcher's record of the whole dispatch,
// and splitting the two here would invite a reader to add them up wrongly.
func tokenLine(raw json.RawMessage) string {
	if len(bytes.TrimSpace(raw)) == 0 {
		return "not measured"
	}
	var m dispatchMetrics
	if err := json.Unmarshal(raw, &m); err != nil {
		return "not measured (metrics bag unreadable)"
	}
	if m.Tokens != nil {
		t := *m.Tokens
		sum := func(a, b int64) int64 { return a + b }
		return fmt.Sprintf("input %d, output %d, cache read %d, cache creation %d",
			sum(t.Main.Input, t.Sidechain.Input),
			sum(t.Main.Output, t.Sidechain.Output),
			sum(t.Main.CacheRead, t.Sidechain.CacheRead),
			sum(t.Main.CacheCreation, t.Sidechain.CacheCreation))
	}
	if m.Unattributed != nil && m.Unattributed.Reason != "" {
		switch m.Unattributed.Reason {
		case reasonSessionNeverBound:
			return "cost unattributed — session never bound"
		case reasonSessionAmbiguous:
			return fmt.Sprintf("cost unattributed — the session token matched %d sessions", m.Unattributed.Candidates)
		case reasonDispatchAmbiguous:
			return fmt.Sprintf("cost unattributed — indistinguishable from %d concurrent dispatches", m.Unattributed.Candidates)
		default:
			return "cost unattributed — " + neutraliseMarkers(m.Unattributed.Reason)
		}
	}
	return "not measured"
}

// orElse is what an unrecorded optional field renders as. A dispatch that
// ran against no task, or produced no commit, is ordinary rather than
// broken, and the ledger says so in words instead of leaving a cell the
// reader has to interpret.
func orElse(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return neutraliseMarkers(s)
}

// RenderLedger renders a change's dispatches as the SDD ledger: one
// section per dispatch, in seq order, naming the task it ran against, the
// role, the model, the commit it produced, the base of any diff it was
// given, how it ended and what it cost.
//
// The model renders VERBATIM. A slot that resolves its own model from an
// agent definition the dispatcher cannot read is recorded as the literal
// `unknown (agent-defined)`, and the ledger is where a reader would
// notice a plausible-looking slug substituted for it -- which is exactly
// why the run record's own requirement forbids one.
//
// Dispatches arrive in seq order from the store; the order is restated
// here rather than re-sorted, so that a store returning them out of order
// shows up as a ledger out of order rather than being silently corrected.
func RenderLedger(r Run) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# SDD ledger — %s\n\n", neutraliseMarkers(r.Change))
	b.WriteString("Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.\n\n")

	if len(r.Dispatches) == 0 {
		b.WriteString("No dispatches were recorded for this change.\n")
		return b.String()
	}

	for _, d := range r.Dispatches {
		fmt.Fprintf(&b, "## Dispatch %d — %s\n\n", d.Seq, orElse(d.Role, "role not recorded"))
		fmt.Fprintf(&b, "- Task: %s\n", orElse(d.TaskID, "no task"))
		fmt.Fprintf(&b, "- Role: %s\n", orElse(d.Role, "not recorded"))
		if strings.TrimSpace(d.Slot) != "" {
			fmt.Fprintf(&b, "- Slot: %s\n", neutraliseMarkers(d.Slot))
		}
		fmt.Fprintf(&b, "- Model: %s\n", orElse(d.Model, "not recorded"))
		fmt.Fprintf(&b, "- Commit: %s\n", orElse(d.CommitSHA, "no commit"))
		// Conditional, in the shape the Slot and Notes lines use, rather
		// than orElse-defaulted: every implementer dispatch and the primary
		// reviewer's own dispatch read the whole diff and legitimately
		// record no base, and a `- Diff base: no base` line on each of them
		// is noise that says nothing.
		if strings.TrimSpace(d.DiffBase) != "" {
			fmt.Fprintf(&b, "- Diff base: %s\n", neutraliseMarkers(d.DiffBase))
		}
		fmt.Fprintf(&b, "- Outcome: %s\n", orElse(d.Outcome, "not recorded"))
		started := "not recorded"
		if !d.StartedAt.IsZero() {
			started = d.StartedAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(&b, "- Started: %s\n", started)
		fmt.Fprintf(&b, "- Tokens: %s\n", tokenLine(d.Metrics))
		if strings.TrimSpace(d.Notes) != "" {
			fmt.Fprintf(&b, "- Notes: %s\n", neutraliseMarkers(d.Notes))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// changeNameAllowed is the allowlist a change name must match: one
// leading alphanumeric, then letters, digits, '.', '_' and '-'. It is
// preserve-session-records.sh's Protection 1, which this change retires,
// carried into Go character for character. Change names are a lowercased
// Jira key plus a kebab-case slug, or the slug alone, so it costs a real
// name nothing.
var changeNameAllowed = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// datedFile matches a rendered file's date prefix digit by digit. A bare
// `*-<change><suffix>` would also match a DIFFERENT change whose name ends
// in this one -- `2020-01-01-other-demo.md` for the change `demo` -- and
// reusing that path would overwrite that change's record. The retired
// script's `find` was anchored for exactly this reason.
const datedFilePrefix = `^[0-9]{4}-[0-9]{2}-[0-9]{2}-`

// renderKinds maps a kind to the directory it renders into and the
// filename suffix it takes.
//
// The suffixes are the ones docs/superpowers/ ALREADY HOLDS, taken from
// the arguments the retired preserve-session-records.sh passed for each
// directory: every file archived under reviews/ is named
// `<date>-<change>-panel.md`, and a renderer writing a second convention
// into that directory would be exactly the drift this change exists to
// remove.
var renderKinds = map[string]struct{ dir, suffix string }{
	"ledger": {filepath.Join("docs", "superpowers", "ledgers"), ".md"},
	"panel":  {filepath.Join("docs", "superpowers", "reviews"), "-panel.md"},
}

// Kinds are the render kinds, in the order a caller rendering all of them
// should render them.
func Kinds() []string { return []string{"ledger", "panel"} }

// Destination resolves where kind's record for change renders to under
// repoRoot, and refuses rather than returning a path it cannot vouch for.
//
// BOTH PATH PROTECTIONS ARE EXPLICIT, never an incidental consequence of
// how the path is assembled. They are inherited from
// preserve-session-records.sh, which this change retires, and each has
// already cost this repository an incident:
//
//  1. THE CHANGE NAME IS VALIDATED against changeNameAllowed. A name
//     carrying `/` was previously blocked only by an accident of string
//     concatenation, which a refactor could have removed without anyone
//     noticing it was load-bearing; a name carrying a glob metacharacter
//     defeated the dated-file scan below, and `*` once matched and
//     overwrote a DIFFERENT change's preserved record.
//
//  2. THE RESOLVED DESTINATION MUST BE CONTAINED WITHIN repoRoot after
//     filepath.EvalSymlinks. The directories under docs/superpowers/ are
//     ordinary tracked repository paths, editable in any pull request: if
//     one is a symlink, joining the path and writing through it puts the
//     record outside the repository entirely. The check resolves as far
//     as the path exists, so a symlinked ANCESTOR cannot smuggle a
//     not-yet-created leaf out either.
//
// The date is fixed at the FIRST render for a change: an existing dated
// file is reused, so a fix round overwrites in place rather than leaving
// one dated duplicate per round. That rule is carried over verbatim from
// the retired script and is the only thing about it that survives.
//
// Destination creates nothing. The caller makes the directory and writes
// the file, so a render that reports MISSING leaves no empty directory
// behind.
func Destination(repoRoot, kind, change string, today time.Time) (string, error) {
	spec, ok := renderKinds[kind]
	if !ok {
		return "", fmt.Errorf("unknown record kind %q: it is one of %s", kind, strings.Join(Kinds(), ", "))
	}
	if !changeNameAllowed.MatchString(change) {
		return "", fmt.Errorf("change name %q is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'", change)
	}

	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root %q: %w", repoRoot, err)
	}

	dir, err := resolveAsFarAsItExists(filepath.Join(root, spec.dir))
	if err != nil {
		return "", fmt.Errorf("resolve destination directory: %w", err)
	}
	if !containedIn(root, dir) {
		return "", fmt.Errorf("destination %q resolves outside the repository root %q — refusing to write through it", dir, root)
	}

	if existing := existingDatedFile(dir, change, spec.suffix); existing != "" {
		return existing, nil
	}
	return filepath.Join(dir, today.UTC().Format("2006-01-02")+"-"+change+spec.suffix), nil
}

// resolveAsFarAsItExists resolves every symlink in p that exists, and
// appends the components that do not yet. Resolving only p itself would
// leave a symlinked ancestor unresolved whenever the leaf is absent,
// which is precisely the case a first render is in.
func resolveAsFarAsItExists(p string) (string, error) {
	resolved, err := filepath.EvalSymlinks(p)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	parent := filepath.Dir(p)
	if parent == p {
		return "", err
	}
	resolvedParent, err := resolveAsFarAsItExists(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(p)), nil
}

// containedIn reports whether path sits strictly inside root. Both are
// already resolved by the time this is called, so the comparison is over
// real paths rather than over the arguments a caller supplied.
func containedIn(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// existingDatedFile is the change's already-rendered file in dir, or "".
// Sorted, so that the choice among several is deterministic rather than
// whatever order the filesystem happened to answer in.
func existingDatedFile(dir, change, suffix string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	pattern := regexp.MustCompile(datedFilePrefix + regexp.QuoteMeta(change+suffix) + `$`)
	var found []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if pattern.MatchString(e.Name()) {
			found = append(found, e.Name())
		}
	}
	if len(found) == 0 {
		return ""
	}
	sort.Strings(found)
	return filepath.Join(dir, found[0])
}
