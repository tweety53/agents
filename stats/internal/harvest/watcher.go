package harvest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DefaultTranscriptsRootEnv, when set, overrides DefaultTranscriptsRoot --
// a test's own isolated root, the same pattern
// internal/fallback.StateRoot's MYFLOW_STATE_DIR already uses for the
// state directory.
const DefaultTranscriptsRootEnv = "MYFLOW_TRANSCRIPTS_DIR"

// DefaultTranscriptsRoot returns the directory Claude Code writes session
// transcripts under: DefaultTranscriptsRootEnv when set, otherwise
// "~/.claude/projects" resolved against the current user's home
// directory.
func DefaultTranscriptsRoot() (string, error) {
	if v := os.Getenv(DefaultTranscriptsRootEnv); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("harvest: resolve home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// HarvestSink is where a batch's results -- attributed token deltas and
// the transcript's newly consumed byte offset -- are committed together.
//
// Defined here, at the consumer, per go-interface-design: internal/harvest
// never imports internal/store, so this package -- and
// TestHarvestNeedsNoDatabase in particular -- is testable with nothing
// but a fake satisfying these two methods, no PostgreSQL required. The
// daemon wires a real implementation backed directly by *store.Store,
// whose GetHarvestOffset and CommitHarvestBatch methods are written to
// match this interface exactly (structural typing, no adapter needed --
// cmd/myflowd/main.go asserts this at compile time).
//
// This interface, and the offset it makes this package's sole authority
// for, replaces an earlier design (this package's own git history) in
// which Watcher kept a local harvest-offsets.json file and committed
// metrics and the offset as two separate calls. Task 9's post-commit
// review found that design unsafe under both possible orderings of those
// two calls -- offset-then-metrics silently under-counted a batch on
// every store outage, which is the routine condition this whole change
// exists to survive, not a rare crash window; metrics-then-offset risked
// the opposite, a retried batch adding its usage twice. CommitHarvestBatch
// removes the choice by making both effects one atomic write, which is
// also what makes a purely local, uncommitted offset file redundant: the
// only offset that matters is the one Postgres holds beside the totals it
// governs.
//
// CommitHarvestBatch's expectedOffset/expectedFound parameters exist for
// a second reason, found in the same review's follow-up (F7): reading
// the offset and committing the batch are still two separate calls, so
// nothing prevents two Watchers -- two myflowd processes, one stale
// alongside a freshly started one -- from both reading the same offset,
// computing overlapping deltas, and both attempting to commit. Passing
// back exactly what GetHarvestOffset returned lets the store guard the
// commit with optimistic concurrency instead: only the caller whose
// expected state still matches actually commits.
type HarvestSink interface {
	// GetHarvestOffset returns the last committed byte offset for
	// transcriptPath, and whether any batch has ever been committed for
	// it (false, with offset 0, for a transcript never seen before).
	GetHarvestOffset(ctx context.Context, transcriptPath string) (offset int64, found bool, err error)
	// CommitHarvestBatch atomically adds deltas (additive token patches,
	// keyed by stage run id; may be empty) onto whatever is already
	// stored, and advances transcriptPath's committed offset from
	// (expectedOffset, expectedFound) -- exactly what a prior
	// GetHarvestOffset call returned -- to newOffset. Either both effects
	// land or neither does.
	//
	// applied reports whether the batch was actually committed. A false
	// return with a nil error means the offset had already moved by the
	// time this call reached the store -- a concurrent committer won the
	// race for the same transcriptPath -- which is not a failure: the
	// caller should simply re-read (GetHarvestOffset) and try again on
	// its next cycle. A non-nil error is a genuine failure.
	CommitHarvestBatch(ctx context.Context, transcriptPath string, expectedOffset int64, expectedFound bool, newOffset int64, deltas map[int64]json.RawMessage) (applied bool, err error)
}

// Pricer prices one stage run's already-committed metrics -- called only
// after a harvest batch's CommitHarvestBatch call has reported the batch
// actually applied (RunOnce, below), never from inside that commit's own
// transaction: the commit's atomicity is what makes harvesting
// exactly-once, and pricing is a pure, idempotent recomputation from
// metrics already durably stored, so it has nothing to contribute to that
// atomicity and everything to lose by extending it (a slow or failing
// pricing pass would then block the harvest it has no business blocking).
//
// Defined here, at the consumer, per go-interface-design, like
// WindowSource and HarvestSink above: internal/harvest never imports
// internal/store, so *store.Store.Price (whose signature already matches
// this interface exactly, no adapter needed) is wired in only by the
// daemon (cmd/myflowd/main.go, via WithPricer).
type Pricer interface {
	Price(ctx context.Context, stageRunID int64) error
}

// maxSessionTokenResolutionCycles bounds how many RunOnce cycles a Watcher
// keeps looking for a given session token before giving up (design.md,
// "binding is bounded, and unbinding never happens"; the spec's own
// "bound is a bound number of cycles, or a wall-clock window -- pick
// one"). This package picks cycles, not wall-clock: the bound only
// exists to cap wasted work for a harness that will never produce a
// transcript at all (Cursor, Codex -- design.md's rejected-alternatives
// section), and a cycle count is exact and trivial to test
// deterministically (drive RunOnce N times), where a wall-clock bound
// would make the same test depend on either a fake clock threaded
// through this package for no other purpose, or a real sleep. At
// cmd/myflowd's own harvestInterval (5s, main.go), 60 cycles is 5
// minutes -- ample time for a live harness's transcript line to be
// flushed and read even under load, while bounding the cost of a
// harness that never will to a few minutes of per-cycle map lookups
// rather than forever.
//
// The bound is tracked per token (task 4b), not per stage run: a run's own
// later marks never enter this bookkeeping at all once its token has
// bound (they resolve at insert time, store.Store.insertStageRunAndSupersede's
// own doc comment), so there is exactly one bounded search per session, not one
// per mark that session makes.
const maxSessionTokenResolutionCycles = 60

// SessionTokenBinder resolves the session tokens a run generates once and
// passes on every mark it makes (KAN-172, task 1; reworked from one
// correlator per mark to one per session in task 4b) into session_id
// bindings, once a harvest cycle has located the transcript that carries
// one. Defined here, at the consumer, per go-interface-design, exactly
// like WindowSource, HarvestSink and Pricer above: internal/harvest never
// imports internal/store, so this package is testable against a fake with
// no PostgreSQL required. The daemon wires a real implementation backed by
// *store.Store, whose UnresolvedSessionTokens and BindSession methods are
// written to match this interface exactly.
type SessionTokenBinder interface {
	// UnresolvedSessionTokens returns every stage run id and its session
	// token for which no session has yet been bound.
	UnresolvedSessionTokens(ctx context.Context) (map[int64]string, error)
	// BindSession binds session_id to sessionID on every stage run
	// carrying sessionToken that has not already been bound -- not just
	// the one stage run whose mark first revealed the token (task 4b's
	// "a resolving token binds every run carrying it, not just the one
	// that revealed it"). bound is how many rows this call actually
	// updated; zero, with no error, is not a failure -- see
	// store.Store.BindSession's own doc comment.
	BindSession(ctx context.Context, sessionToken string, sessionID string) (bound int64, err error)
}

// Watcher periodically scans a transcripts root for *.jsonl files and
// harvests whatever bytes are new since each one's last committed offset,
// attributing them via its Attributor and committing the result -- both
// the token deltas and the advanced offset -- through its HarvestSink in
// one atomic call per file. It never talks to PostgreSQL directly --
// only through WindowSource (via Attributor), HarvestSink, and (when
// configured) Pricer.
type Watcher struct {
	root          string
	sink          HarvestSink
	attributor    *Attributor
	logger        *slog.Logger
	pricer        Pricer
	sessionTokens SessionTokenBinder

	// tokenCycles counts, per session token, how many RunOnce cycles have
	// searched for that token without finding a unique match yet.
	// gaveUpTokens holds the tokens this Watcher has stopped looking for
	// at all -- either maxSessionTokenResolutionCycles was reached
	// (bounded give-up) or an earlier cycle found the token in more than
	// one session (ambiguity is treated as terminal too: waiting longer
	// cannot un-ambiguate two transcripts that already both carry the
	// same literal token). Both maps are keyed by token, not by stage run
	// id (task 4b): every stage run carrying a given token shares one
	// bounded search, not one each, which is the whole economy of moving
	// from one correlator per mark to one per session -- a run's second or
	// later mark, once its token has bound, resolves session_id at insert
	// time and never enters this bookkeeping at all
	// (store.Store.insertStageRunAndSupersede's own doc comment).
	//
	// Both maps are this Watcher's own in-memory state, not persisted -- a
	// daemon restart resets them, which only ever gives an abandoned token
	// a fresh bounded number of cycles rather than losing correctness
	// (UnresolvedSessionTokens is re-queried from Postgres every cycle
	// regardless), so this is a purely local cost-bounding optimisation,
	// never a source of truth.
	//
	// Neither map is guarded by a mutex: RunOnce is never called
	// concurrently with itself on one Watcher (Run's own loop calls it
	// serially; every test in this package does too), the same
	// assumption this struct's other fields already rely on.
	tokenCycles  map[string]int
	gaveUpTokens map[string]bool
}

// WatcherOption configures optional Watcher behaviour not every caller
// needs -- currently just WithPricer. Following the same functional-option
// shape internal/api.WithSPA already uses (cmd/myflowd/main.go) rather
// than growing NewWatcher's positional parameter list keeps every
// existing call site (this package's own tests included) compiling
// unchanged: a Watcher built with no options simply never prices, exactly
// as it behaved before this task.
type WatcherOption func(*Watcher)

// WithPricer configures the Watcher to price every stage run a batch
// touches, once that batch's CommitHarvestBatch call has reported it
// applied (RunOnce's own doc comment explains why after, and only after).
func WithPricer(p Pricer) WatcherOption {
	return func(w *Watcher) { w.pricer = p }
}

// WithSessionTokenBinder configures the Watcher to resolve stage runs'
// session tokens (KAN-172, task 2): each cycle, it asks binder which
// session tokens are still unresolved and looks for them among the
// transcripts that cycle is already reading, binding session_id where
// exactly one session's transcript carries a session token. A Watcher
// built with no WithSessionTokenBinder option resolves no session tokens
// at all -- every stage run stays exactly as unattributed as it was
// before this task, the same additive shape WithPricer already
// established.
func WithSessionTokenBinder(binder SessionTokenBinder) WatcherOption {
	return func(w *Watcher) { w.sessionTokens = binder }
}

// HasPricer reports whether this Watcher was configured with WithPricer.
// Exported for cmd/myflowd's own wiring test (KAN-172, task 7): asserting
// the constructed value here, rather than grepping main.go's source text
// for "WithPricer", is what still catches a refactor that keeps the call
// site's text but drops its effect.
func (w *Watcher) HasPricer() bool {
	return w.pricer != nil
}

// HasSessionTokenBinder reports whether this Watcher was configured with
// WithSessionTokenBinder. See HasPricer's doc comment for why this exists
// and why it asserts the constructed value rather than source text -- this
// is the accessor that would have caught task 7's own defect (main.go
// building a Watcher with no binder at all, so no stage run was ever
// bound despite tasks 1-6 all working).
func (w *Watcher) HasSessionTokenBinder() bool {
	return w.sessionTokens != nil
}

// NewWatcher builds a Watcher over root (scanned recursively for
// *.jsonl files), sink (where offsets are read from and results are
// committed to) and attributor (how records become deltas). logger may
// be nil. opts configures optional behaviour -- see WithPricer and
// WithSessionTokenBinder.
func NewWatcher(root string, sink HarvestSink, attributor *Attributor, logger *slog.Logger, opts ...WatcherOption) *Watcher {
	w := &Watcher{
		root:         root,
		sink:         sink,
		attributor:   attributor,
		logger:       logger,
		tokenCycles:  make(map[string]int),
		gaveUpTokens: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// RunOnce performs a single scan-and-harvest pass: every *.jsonl file
// under the watcher's root is read from its last committed offset
// (HarvestSink.GetHarvestOffset), attributed, and committed -- offset and
// deltas together, atomically -- through HarvestSink.CommitHarvestBatch.
//
// A failure at any point for one file (reading the transcript, looking
// up its committed offset, attributing its records, or committing the
// result) is logged and this loop moves on to the next file, rather than
// aborting the whole pass -- one bad file must never starve every other
// session's harvesting. Because nothing about a file's committed state
// changes until CommitHarvestBatch succeeds, a failure anywhere before
// that call leaves the file exactly as it was: the next RunOnce reads
// the same bytes from the same offset and tries again, with no risk of
// either losing that batch's usage or adding it twice.
//
// A commit that reports applied=false with a nil error (HarvestSink's own
// doc comment) is not logged as a failure at all: it means a concurrent
// harvester already advanced this file's offset first, which is the
// ordinary, correct outcome of losing that race, not an error condition
// -- the next RunOnce simply re-reads the current state and tries again.
//
// Withholding a batch that revealed a sessionToken: when a SessionTokenBinder is
// configured (WithSessionTokenBinder), a file's newly read batch is checked
// for still-pending sessionTokens (matchSessionTokens) before it is attributed or
// committed at all. A batch that matched one is *not* attributed or
// committed this cycle -- its offset is left exactly where it was, so
// the next cycle re-reads the identical bytes once resolveSessionTokens (run
// once, after every file this cycle has been read) has had a chance to
// bind the session that batch just revealed.
//
// This is load-bearing, not an optimisation: Claude Code flushes a
// turn's transcript entries together, so the `stage begin -session-token ...`
// mark and that same turn's own usage routinely arrive in the very same
// batch -- not a later one, and a mark's own turn is frequently a
// stage's largest. Attributing and committing that batch before binding
// has happened would compute it against a window that does not exist
// yet, commit a delta of nothing, advance the offset past it, and never
// get another chance: nothing re-reads bytes the offset has already
// moved past. Withholding the commit is what keeps this batch's usage
// from being lost outright rather than merely attributed one cycle
// late (TestBindMarkAndFirstUsageInSameBatchAreBothAttributed,
// watcher_test.go, is the regression test for exactly this).
//
// RunOnce returns the number of files whose newly read records were
// successfully committed by this call.
func (w *Watcher) RunOnce(ctx context.Context) (int, error) {
	files, err := discoverTranscripts(w.root)
	if err != nil {
		return 0, err
	}

	pendingSessionTokens := w.pendingSessionTokens(ctx)
	// matchedSessions accumulates, per sessionToken, the distinct session ids
	// found carrying it across every transcript this cycle reads -- not
	// just the first file that matches, since the whole point of
	// scanning every file before deciding is telling "exactly one
	// session" apart from "more than one" (design.md, "a session is
	// never guessed").
	matchedSessions := make(map[string]map[string]bool, len(pendingSessionTokens))

	touchedFiles := 0
	for _, path := range files {
		if ctx.Err() != nil {
			return touchedFiles, ctx.Err()
		}

		offset, found, err := w.sink.GetHarvestOffset(ctx, path)
		if err != nil {
			w.warn("harvest: get committed offset failed, will retry", "path", path, "error", err)
			continue
		}

		records, commands, newOffset, err := ReadNewRecords(path, offset)
		if err != nil {
			w.warn("harvest: read transcript failed, will retry", "path", path, "error", err)
			continue
		}

		matchedHere := w.matchSessionTokens(pendingSessionTokens, commands, matchedSessions)

		if newOffset == offset {
			continue // nothing new (or only a partial trailing line) since last time.
		}

		if matchedHere {
			// This batch is the one that revealed a still-pending sessionToken
			// -- withhold its commit rather than attribute and commit it
			// now (see this method's own doc comment, "withholding a
			// batch that revealed a sessionToken"). The offset does not
			// advance, so the next cycle re-reads this exact same
			// region once resolveSessionTokens (below, after every file this
			// cycle has been read) has had a chance to bind it -- at
			// which point WindowsForSession will actually have a window
			// for it, and this same batch's usage attributes correctly
			// instead of being computed against no window, committed as
			// an empty delta, and never revisited.
			continue
		}

		deltas, err := w.attributor.Attribute(ctx, records)
		if err != nil {
			w.warn("harvest: attribute failed, will retry", "path", path, "error", err)
			continue
		}

		patches, err := encodePatches(deltas)
		if err != nil {
			w.warn("harvest: encode failed, will retry", "path", path, "error", err)
			continue
		}

		applied, err := w.sink.CommitHarvestBatch(ctx, path, offset, found, newOffset, patches)
		if err != nil {
			w.warn("harvest: commit failed, will retry", "path", path, "error", err)
			continue
		}
		if !applied {
			// Lost a race with a concurrent harvester for this file --
			// benign, not a warning; the next cycle re-reads and retries.
			continue
		}
		touchedFiles++

		// Price every stage run this batch touched -- deliberately after
		// CommitHarvestBatch has already reported applied, never inside
		// it (Pricer's own doc comment explains why). A pricing failure
		// (no rate for a model, no chargeable tokens yet, or a transient
		// store error) is logged and skipped, never fatal to this pass
		// and never a reason to retry the batch itself: the metrics are
		// already durably committed either way, and the next harvest
		// cycle -- or a future re-price -- gets another chance.
		if w.pricer != nil {
			for stageRunID := range deltas {
				if err := w.pricer.Price(ctx, stageRunID); err != nil {
					w.warn("harvest: price stage run failed, will retry next cycle", "stage_run_id", stageRunID, "error", err)
				}
			}
		}
	}

	w.resolveSessionTokens(ctx, pendingSessionTokens, matchedSessions)

	return touchedFiles, nil
}

// pendingSessionTokens returns this cycle's stage-run-id -> session-token
// map to search for: whatever w.sessionTokens.UnresolvedSessionTokens
// reports, minus any run whose token this Watcher has already given up on
// (gaveUpTokens, keyed by token -- task 4b) -- so a token that has already
// been logged as abandoned is never looked for again by this process,
// which is the whole point of tracking gaveUpTokens at all. A Watcher with
// no SessionTokenBinder configured (WithSessionTokenBinder never called)
// returns nil, and every other token-related step below is a no-op over
// an empty map -- resolving session tokens is additive, exactly like
// pricing.
func (w *Watcher) pendingSessionTokens(ctx context.Context) map[int64]string {
	if w.sessionTokens == nil {
		return nil
	}
	all, err := w.sessionTokens.UnresolvedSessionTokens(ctx)
	if err != nil {
		w.warn("harvest: list unresolved session tokens failed, will retry", "error", err)
		return nil
	}
	pending := make(map[int64]string, len(all))
	for stageRunID, sessionToken := range all {
		if w.gaveUpTokens[sessionToken] {
			continue
		}
		pending[stageRunID] = sessionToken
	}
	return pending
}

// matchSessionTokens scans commands -- the Bash commands newly read from one
// transcript file this cycle -- for every sessionToken in pending, recording
// each distinct session id a match was found under into matched, and
// reports whether this file's commands matched at least one pending
// sessionToken (RunOnce uses that to withhold this batch's commit -- see its
// own doc comment on "withholding a batch that revealed a sessionToken").
//
// A sessionToken is matched only when the command is genuinely a mark
// carrying it -- isSessionMarkCommand, below -- never by the token's bare
// presence in the command text. KAN-172's final review panel (finding F4)
// caught a bare `strings.Contains(cmd.Command, sessionToken)` here: any
// command that merely MENTIONED a pending sessionToken -- a diagnostic
// grep for it, a log dump, a database query, an echo -- counted exactly
// like the real `stage begin -session-token <sessionToken> ...` invocation
// that actually identifies the session. Two matches produce a correctly
// refused ambiguity, but once the owning session's own mark bytes have
// already been read past (the offset only ever advances) the mention
// becomes the *only* remaining occurrence, and the run silently bound to
// the mentioning session instead -- reproduced live, twice: stage run 22
// bound through the dispatcher's own diagnostic greps for its token.
func (w *Watcher) matchSessionTokens(pending map[int64]string, commands []CommandRecord, matched map[string]map[string]bool) bool {
	if len(pending) == 0 || len(commands) == 0 {
		return false
	}
	matchedHere := false
	for _, cmd := range commands {
		for _, sessionToken := range pending {
			if !isSessionMarkCommand(cmd.Command, sessionToken) {
				continue
			}
			matchedHere = true
			sessions := matched[sessionToken]
			if sessions == nil {
				sessions = make(map[string]bool)
				matched[sessionToken] = sessions
			}
			sessions[cmd.SessionID] = true
		}
	}
	return matchedHere
}

// stageMarkInvocationPattern matches the `stage begin` / `stage end`
// subcommand shape as two adjacent words -- design.md's and stage.go's
// own usage string (`myflow stage begin ...` / `myflow stage end ...`),
// with whatever whitespace (including a newline, inside a multi-line
// shell block) separates them. It deliberately imposes no flag ordering
// of its own: stage.go's own flag registration imposes none on -harness,
// -session, -stage, -command or -session-token -- and, since KAN-174, no
// requirement that this subcommand lead the command text either. Real
// marks are emitted inside shell blocks carrying variable assignments,
// directory changes and other statements ahead of the invocation, on the
// same line or a later one (design.md, "recognise a mark by its
// invocation, not by its position"; isSessionMarkCommand's own doc
// comment has the fuller history).
var stageMarkInvocationPattern = regexp.MustCompile(`\bstage\s+(?:begin|end)\b`)

// isSessionMarkCommand reports whether command is genuinely a stage mark
// carrying sessionToken as the value of its own -session-token flag --
// the fix for KAN-172 finding F4 (matchSessionTokens' own doc comment
// above has the defect and its live reproduction), reworked by KAN-174
// below. It requires both of the following:
//
//  1. the command invokes `stage begin` or `stage end`, anywhere in the
//     command text (stageMarkInvocationPattern);
//  2. sessionToken is the exact value bound to -session-token in that
//     same command, whether written as two fields ("-session-token
//     TOKEN") or joined with "=" ("-session-token=TOKEN") -- both are
//     valid to the flag package cmd/myflow/stage.go builds on.
//
// Requirement 2 is field-based (strings.Fields), not a second substring
// test on the whole command: a token that merely follows the word
// "-session-token" somewhere in an unrelated position would reintroduce
// exactly the class of bug requirement 2 alone replaces -- this is what
// still keeps a `grep`, a `psql` query, a piped `cat`, or a bare mention
// of the token from matching
// (TestCommandsThatOnlyMentionTokenNeverBind): none of those bind the
// token as -session-token's own value, no matter how the position
// requirement below is decided.
//
// KAN-172 finding F5 added a third requirement here -- the command must
// begin (after stripping one leading `cd <path> &&`) with a word whose
// base name is "myflow" -- to reject a command that only PRINTS a
// mark-shaped string, e.g. `echo "myflow stage begin ...
// -session-token mf-abc123 ..."`. KAN-174 removes that requirement: real
// marks are routinely emitted after variable assignments
// (`N=kan; T=mf-x; cd /repo`) and on a later line of a multi-statement
// block, not just behind a single leading `cd ... &&`, and every one of
// those shapes was measured rejected by the F5 anchor in production --
// all 14 stage runs from the finish sequence that merged F5 went unbound
// (design.md). A position anchor loose enough to admit those shapes (e.g.
// requiring "myflow" only after a command boundary -- start of text, or
// after `;`, `&&`, `||`, `|`, a newline) was considered and rejected: the
// echoed-example text it is meant to exclude still only needs a newline
// before "myflow" to satisfy a boundary anchor too, which is common in
// exactly the multi-line blocks this defect is about, so the anchor adds
// machinery without closing the gap it targets (design.md, "recognise a
// mark by its invocation, not by its position").
//
// The residual this leaves open, admitted deliberately rather than left
// implicit: a command that reproduces a mark's text -- including inside a
// quoted `echo` -- without performing it now matches
// (TestEchoedMarkExampleIsAnAcceptedResidual). This is preferred over
// keeping a position anchor because the two failure modes are not
// symmetric. A false negative here is silent and total: a mark that fails
// to match binds nothing, and nothing else in this package or its caller
// ever notices or retries -- exactly the state this whole change repairs.
// A false positive needs a mark-shaped string carrying a session token
// that is *currently pending* (UnresolvedSessionTokens;
// pendingSessionTokens's own doc comment) -- an already-bound or
// never-pending token in printed text matches nothing here at all, since
// nothing is being searched for it -- and where it collides with a
// genuine mark's own token, or with another echoed example's,
// matchSessionTokens accumulates every distinct session id under that
// token and resolveSessionTokens's ambiguity branch refuses to bind
// rather than choosing one. So the false positive this admits is narrow,
// and, on the one path where it could actually mislead (two sessions
// truly matching), already refused rather than resolved by guessing.
//
// This comment does not repeat KAN-172's own error here: an earlier
// version of the F5 comment justified admitting this same residual with a
// claim that closing it would need parsing the command as shell syntax, a
// claim a reviewer disproved by tracing the code. The justification above
// is checked against this package's actual matching and withholding logic
// (matchSessionTokens, resolveSessionTokens), not merely asserted.
func isSessionMarkCommand(command, sessionToken string) bool {
	if !stageMarkInvocationPattern.MatchString(command) {
		return false
	}
	fields := strings.Fields(command)
	for i, field := range fields {
		switch {
		case field == "-session-token" || field == "--session-token":
			if i+1 < len(fields) && trimTokenQuotes(fields[i+1]) == sessionToken {
				return true
			}
		case strings.HasPrefix(field, "-session-token="):
			if trimTokenQuotes(field[len("-session-token="):]) == sessionToken {
				return true
			}
		case strings.HasPrefix(field, "--session-token="):
			if trimTokenQuotes(field[len("--session-token="):]) == sessionToken {
				return true
			}
		}
	}
	return false
}

// trimTokenQuotes strips a single layer of surrounding straight quotes a
// shell-quoted flag value might carry ("-session-token 'mf-abc'" or
// "-session-token=\"mf-abc\""), so isSessionMarkCommand compares the same
// literal validateSessionToken accepted, not a quoted rendering of it.
func trimTokenQuotes(s string) string {
	return strings.Trim(s, `"'`)
}

// resolveSessionTokens decides, for every distinct token this cycle
// searched for (pending, still keyed by stage run id, one entry per run
// carrying an unresolved token), what its matchedSessions say and acts on
// it -- this is where "exactly one match binds, zero stays unresolved,
// more than one is refused" (design.md, this task's own spec requirement)
// actually happens, once every transcript this cycle reads has already
// been scanned (matchSessionTokens, above), never before.
//
// The decision and the give-up/ambiguity bookkeeping are made once per
// token (task 4b), not once per stage run: BindSession itself binds every
// run sharing a token in one call (store.Store.BindSession's own doc
// comment), so calling it once per stage run here would be redundant work
// for every run beyond the first, and tokenCycles/gaveUpTokens would
// otherwise let one run's stage give up on a schedule out of step with
// another run sharing the exact same token.
func (w *Watcher) resolveSessionTokens(ctx context.Context, pending map[int64]string, matchedSessions map[string]map[string]bool) {
	tokenRuns := make(map[string][]int64, len(pending))
	for stageRunID, sessionToken := range pending {
		tokenRuns[sessionToken] = append(tokenRuns[sessionToken], stageRunID)
	}

	for sessionToken, stageRunIDs := range tokenRuns {
		sessions := matchedSessions[sessionToken]

		switch len(sessions) {
		case 0:
			w.tokenCycles[sessionToken]++
			if w.tokenCycles[sessionToken] < maxSessionTokenResolutionCycles {
				continue
			}
			w.warn("harvest: session token unresolved after the bounded window, giving up",
				"stage_run_ids", stageRunIDs, "cycles", w.tokenCycles[sessionToken])
			w.gaveUpTokens[sessionToken] = true
			delete(w.tokenCycles, sessionToken)

		case 1:
			var sessionID string
			for s := range sessions {
				sessionID = s
			}
			bound, err := w.sessionTokens.BindSession(ctx, sessionToken, sessionID)
			if err != nil {
				w.warn("harvest: bind session failed, will retry", "stage_run_ids", stageRunIDs, "error", err)
				continue
			}
			// bound == 0 means every run carrying this token was in fact
			// bound already (task 2 spec's "binding is one-way": nothing
			// here ever re-binds it) -- not an error, just stale
			// information from this cycle's own read of
			// UnresolvedSessionTokens. Either way, this Watcher has
			// nothing left to do for this token.
			_ = bound
			delete(w.tokenCycles, sessionToken)

		default:
			sessionIDs := make([]string, 0, len(sessions))
			for s := range sessions {
				sessionIDs = append(sessionIDs, s)
			}
			w.warn("harvest: session token matched more than one session, refusing to bind",
				"stage_run_ids", stageRunIDs, "sessions", sessionIDs)
			w.gaveUpTokens[sessionToken] = true
			delete(w.tokenCycles, sessionToken)
		}
	}
}

// encodePatches marshals each stage run's Delta into the json.RawMessage
// shape HarvestSink.CommitHarvestBatch's deltas parameter expects: the
// whole-run total under "tokens" (Delta.Total, unchanged meaning) and,
// when this batch carried any, the per-model breakdown under "models"
// (Delta.Models, task 22) -- omitted entirely rather than encoded as an
// empty object when a batch attributed nothing to any model, so it never
// masks a stage run's own "models" key already holding real data from an
// earlier batch (json.RawMessage(nil) participates in no merge at all,
// where an encoded "models":{} would still be a value jsonb_deep_add has
// to reconcile against).
func encodePatches(deltas map[int64]Delta) (map[int64]json.RawMessage, error) {
	patches := make(map[int64]json.RawMessage, len(deltas))
	for stageRunID, delta := range deltas {
		mp := MetricsPatch{Tokens: delta.Total, Speed: delta.Speed}
		if len(delta.Models) > 0 {
			mp.Models = make(map[string]ModelBucket, len(delta.Models))
			for model, td := range delta.Models {
				mp.Models[model] = ModelBucket{Tokens: td}
			}
		}
		patch, err := json.Marshal(mp)
		if err != nil {
			return nil, fmt.Errorf("harvest: encode metrics patch for stage run %d: %w", stageRunID, err)
		}
		patches[stageRunID] = patch
	}
	return patches, nil
}

// Run calls RunOnce every interval until ctx is done. A single failed
// pass is logged and does not stop the loop -- the daemon must keep
// harvesting future transcript growth even if one pass hit a transient
// error.
func (w *Watcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := w.RunOnce(ctx); err != nil && w.logger != nil {
				w.logger.Warn("harvest: run failed", "error", err)
			}
		}
	}
}

func (w *Watcher) warn(msg string, args ...any) {
	if w.logger != nil {
		w.logger.Warn(msg, args...)
	}
}

// discoverTranscripts walks root recursively and returns every *.jsonl
// file found, sorted for deterministic processing order. This picks up
// both a session's own top-level file
// (~/.claude/projects/<project>/<session>.jsonl) and its per-session
// subagents/agent-*.jsonl files, which carry sidechain messages under
// the same top-level sessionId -- confirmed by reading a real
// subagents/*.jsonl file directly rather than assumed. A root that does
// not exist yet is not an error: it reads as no transcripts found,
// exactly like a machine that has not run Claude Code yet.
func discoverTranscripts(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jsonl") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("harvest: scan transcripts root %s: %w", root, err)
	}
	sort.Strings(out)
	return out, nil
}
