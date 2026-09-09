// Package records holds the wire shape of a change's derived run record --
// every subagent dispatch and every review-panel finding -- shared by the
// store, the API, the client and (from this change's renderer) the Markdown
// rendering. It holds data and rendering, nothing else, and imports only
// encoding/json and time so that every layer above can depend on it without
// dragging a database or an HTTP client with it.
package records

import (
	"bytes"
	"encoding/json"
	"time"
)

// Dispatch is one recorded subagent dispatch: the SDD ledger's row, the
// task -> commit -> model record and the per-dispatch cost bag, which are
// three views of the same row rather than three records.
//
// ID is the row's own identifier, allocated by the store and returned on
// the recorded row. It is what MergeDispatchMetrics is keyed by, so the
// harvester's attribution pass can name the row it is merging into without
// a query of its own; a caller that has not recorded the dispatch yet
// leaves it zero.
//
// Seq is the dispatch's append order within its change, allocated by the
// store rather than supplied by the caller; a value sent on the way in is
// ignored and the allocated one comes back. Seq, not ID, is the identifier
// a rendered record and a finding's own DispatchSeq refer to: it is stable
// per change and readable, where the row id is the store's private
// bookkeeping.
//
// Key is the dispatcher's own label for this dispatch, unique within the
// run its SessionToken names. It is what makes recording a dispatch
// idempotent and what DispatchEnd closes the row by, and both properties
// need the same thing of it: a value a *replay* reproduces exactly. The
// caller writes it as a literal, and the journalled request carries that
// literal, so a write replayed after a lost response arrives carrying the
// identical key and updates the row the first attempt inserted rather
// than inserting a second row for one logical dispatch. Seq cannot serve
// this purpose -- the store allocates it, so a caller whose write was
// journalled has never seen one.
//
// An empty Key is "not labelled": the store's uniqueness is enforced over
// (change, session token, key), and SQL's NULLs-distinct rule means an
// unlabelled row conflicts with nothing, including another unlabelled
// row. Such a row can be inserted but never deduplicated and never closed
// by DispatchEnd, which is why `flow record dispatch begin` requires
// the flag.
//
// AgentID is the harness's own identifier for the subagent this dispatch
// dispatched, where the harness exposes one. It is what separates two
// dispatches whose attribution intervals overlap -- a review panel's
// slots, dispatched at once against one session -- since the interval
// alone cannot say which of them a record inside the overlap belongs to.
//
// It is optional, and its absence is ordinary rather than degraded:
// Cursor and Codex expose no such identifier at all, so on two of the
// three supported harnesses every dispatch is recorded without one and
// attribution falls back to the interval rule, which remains correct for
// dispatches that do not overlap. An absent value is "" and means "not
// reported"; it never matches another absent value.
//
// DiffBase is the sha the diff this dispatch was given was computed
// from -- the answer to "what has this slot already read". A Full panel
// re-run hands each slot only its own delta, and that delta is anchored
// at the slot's own last read rather than at the current round, because
// a Targeted round re-runs slot 0 and the slots that raised findings, so
// a slot arriving at a Full pass may have missed rounds in between.
//
// It is optional and its absence is ordinary: every implementer dispatch
// and the primary reviewer's own dispatch read the whole diff and record
// none. An absent value is "" and means "this dispatch read no delta",
// and a slot for which no base is held is given the whole diff rather
// than a delta anchored at nothing -- an unrecorded base is not a small
// delta.
//
// Model is recorded intent, written by the dispatcher -- the literal
// `unknown (agent-defined)` where a slot resolves its own model from an
// agent definition the dispatcher cannot read, never a plausible-looking
// slug. Metrics is derived instead: the harvester attributes it from the
// harness transcript afterwards, so no agent is ever asked to report its
// own token consumption.
type Dispatch struct {
	ID         int64  `json:"id"`
	AgentID    string `json:"agentId,omitempty"`
	Seq        int    `json:"seq"`
	Key        string `json:"key,omitempty"`
	StageRunID *int64 `json:"stageRunId,omitempty"`
	TaskID     string `json:"taskId,omitempty"`
	Role       string `json:"role"`
	Slot       string `json:"slot,omitempty"`
	Model      string `json:"model"`
	CommitSHA  string `json:"commitSha,omitempty"`
	DiffBase   string `json:"diffBase,omitempty"`
	// CostUSD is the dispatch's derived cost, lifted from the stage-run
	// metrics bag's dispatches.<agentId>.cost_usd bucket -- the figure
	// store.Price wrote through the one pricing path every other cost
	// figure uses. It is resolved on reads that join the stage run
	// (RunRecord); a row returned by a write path carries nil. A POINTER,
	// because absence is not zero: nil means no priced bucket was joinable
	// -- no stage run, no agent id, or a bag without the key -- never
	// "measured at zero".
	CostUSD      *float64        `json:"costUsd,omitempty"`
	Outcome      string          `json:"outcome,omitempty"`
	SessionToken string          `json:"sessionToken,omitempty"`
	StartedAt    time.Time       `json:"startedAt"`
	EndedAt      *time.Time      `json:"endedAt,omitempty"`
	Metrics      json.RawMessage `json:"metrics,omitempty"`
	Notes        string          `json:"notes,omitempty"`
}

// DispatchEnd is the closing half of a dispatch's record: the three facts
// that are knowable only once the dispatch has finished, plus the identity
// of the row they close.
//
// A dispatch is recorded in two calls, not one. The row has to exist while
// the dispatch is still running, because the harvester commits its
// transcript offset every few seconds and never re-reads past it: usage
// harvested before the row existed has no window to be attributed to, and
// is dropped or credited to an unrelated earlier dispatch. So `begin`
// writes what is known at the start and `end` writes what is known at the
// close -- see the `dispatch-begin-and-end` decision in this change's
// design.md.
//
// The row is named by SessionToken and Key rather than by Seq, for the
// reason Dispatch.Key gives: a `begin` whose response was lost never
// returned a seq to the caller, and an end that could not name its row
// would leave the window open forever.
//
// EndedAt is required. It is the whole point of the call: an open window
// (a NULL ended_at) goes on claiming later usage, so a dispatch that never
// closes keeps stealing its successors' tokens.
//
// AgentID may be recorded here instead of at Dispatch.AgentID's own opening
// call: on Claude Code the harness reports a subagent's identifier only
// once the dispatch has actually launched, so `begin` cannot always carry
// it. It is optional for the same reason it is optional on Dispatch, and an
// EMPTY AgentID means "not given here", never "clear the one begin
// recorded" -- ApplyDispatchEnd's implementations set the column only when
// this is non-empty, so an end that omits it leaves whatever begin already
// captured untouched. A NON-EMPTY AgentID that names a different identifier
// than begin recorded overwrites it: end's value wins on conflict.
type DispatchEnd struct {
	SessionToken string    `json:"sessionToken"`
	Key          string    `json:"key"`
	CommitSHA    string    `json:"commitSha,omitempty"`
	Outcome      string    `json:"outcome,omitempty"`
	EndedAt      time.Time `json:"endedAt"`
	AgentID      string    `json:"agentId,omitempty"`
}

// Finding is one review-panel finding. Ref is unique per change, not per
// round: a fix round updates the row's Status in place, so a change's
// findings never accumulate a second row for the same reference.
//
// DispatchSeq names the dispatch that raised the finding, by that
// dispatch's Seq within the same change -- the identifier a caller has,
// where the store's own row id is not. It is nil for a finding no single
// dispatch raised, which is a legitimate case rather than a missing value.
type Finding struct {
	Ref         string `json:"ref"`
	DispatchSeq *int   `json:"dispatchSeq,omitempty"`
	Round       int    `json:"round"`
	Slot        string `json:"slot"`
	Severity    string `json:"severity"`
	Location    string `json:"location,omitempty"`
	Note        string `json:"note"`
	Status      string `json:"status"`
	Reproducer  string `json:"reproducer,omitempty"`
}

// Verdict is one recorded outcome a guard reached against one worktree, at
// one point in time -- the record `check-unfinished-work.sh` and any guard
// like it leaves behind so that "this guard tripped here before" stops
// being a fact only memory or chat holds.
//
// Change is the change name on the read side (joined from `changes`, the
// way ListVerdicts and RunRecord report it) and is left empty on the write
// side, where the URL already names the change.
//
// Verdict carries the guard's whole verdict line verbatim -- e.g.
// "OUTSTANDING: <worktree> -- <breakdown>" -- rather than a structured
// breakdown of its own, so the reason the guard gave survives exactly as an
// operator saw it, without a second schema to keep in sync with the first.
//
// FlaggedAt is a pointer because a verdict starts unflagged: nil means "no
// operator has said this was a false positive", not "flagged at the zero
// time", and it is only ever set together with FalsePositiveReason.
type Verdict struct {
	ID                  int64      `json:"id"`
	Change              string     `json:"change,omitempty"`
	Guard               string     `json:"guard"`
	Worktree            string     `json:"worktree"`
	Verdict             string     `json:"verdict"`
	RecordedAt          time.Time  `json:"recordedAt"`
	FalsePositive       bool       `json:"falsePositive"`
	FalsePositiveReason string     `json:"falsePositiveReason,omitempty"`
	FlaggedAt           *time.Time `json:"flaggedAt,omitempty"`
}

// VerdictFlag is the payload `flow record verdict false-positive` sends: an
// operator's judgment that the most recent verdict a named guard reached
// for a change was wrong, and why. It names no verdict id, deliberately --
// see FlagVerdictFalsePositive's own doc comment for why "the latest one
// for this change and guard" is the row it always means.
type VerdictFlag struct {
	Guard  string `json:"guard"`
	Reason string `json:"reason"`
}

// Incident is one per-project record of a guard actually costing time: what
// guard, what went wrong, what recovery was taken and how many minutes it
// cost. Unlike Verdict, an incident is written by hand -- see this change's
// design.md decision "incident-written-by-hand" -- because most verdicts
// are unremarkable and only a fraction ever become an incident worth
// keeping.
//
// Change is optional on both sides: an incident can be recorded against a
// project with no change in flight, or after the change that produced it
// has archived, so a NULL/empty Change is a legitimate case rather than a
// missing value.
type Incident struct {
	ID          int64     `json:"id"`
	Change      string    `json:"change,omitempty"`
	Guard       string    `json:"guard"`
	Symptom     string    `json:"symptom"`
	Recovery    string    `json:"recovery"`
	MinutesLost int       `json:"minutesLost"`
	OccurredAt  time.Time `json:"occurredAt"`
}

// Run is one change's whole derived record: its dispatches in seq order
// and its findings in ref order.
type Run struct {
	Change     string     `json:"change"`
	Dispatches []Dispatch `json:"dispatches"`
	Findings   []Finding  `json:"findings"`
}

// CostStatus is the wire shape GET .../cost-status answers: how many of a
// change's dispatches carry no cost figure, and, for those that do not,
// why -- one count per raw reason a producer stamped
// (internal/harvest/watcher.go's resolveSessionTokens and
// internal/store's MarkDispatchesUnattributed /
// MarkDispatchesUnattributedByID). It exists so `flow record cost-status`
// can state a change's cost honestly, at a handoff, without deriving the
// figure by hand -- see that command's own doc comment.
//
// Reasons is keyed by the reason string verbatim, the same value
// tokenLine (render.go) switches on -- never its rendered prose -- since
// this is a machine-readable count, not a line meant for the ledger.
type CostStatus struct {
	Unattributed int            `json:"unattributed"`
	Reasons      map[string]int `json:"reasons"`
}

// CostStatusOf summarises r's dispatches into a CostStatus, applying
// exactly the precedence tokenLine (render.go) applies to one dispatch at
// a time: a dispatch carrying a real "tokens" figure is measured, never
// unattributed, even where it also carries a stale "unattributed" stamp --
// MarkDispatchesUnattributed's own doc comment explains why that
// contradiction is possible and why the measurement wins. Only a dispatch
// with no "tokens" key and a non-empty "unattributed.reason" counts.
//
// It reuses dispatchMetrics and unattributed (render.go) rather than
// re-declaring the bag's shape a second time here: the two read the same
// wire fact -- does this dispatch have a cost figure, and if not, why --
// for two different purposes, and a second declaration of it would be
// exactly the drift Single Source of Truth exists to prevent.
func CostStatusOf(r Run) CostStatus {
	reasons := map[string]int{}
	unattributedCount := 0
	for _, d := range r.Dispatches {
		if len(bytes.TrimSpace(d.Metrics)) == 0 {
			continue
		}
		var m dispatchMetrics
		if err := json.Unmarshal(d.Metrics, &m); err != nil {
			continue
		}
		if m.Tokens != nil {
			continue
		}
		if m.Unattributed == nil || m.Unattributed.Reason == "" {
			continue
		}
		unattributedCount++
		reasons[m.Unattributed.Reason]++
	}
	return CostStatus{Unattributed: unattributedCount, Reasons: reasons}
}
