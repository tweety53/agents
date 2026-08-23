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
	"strconv"
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

// DispatchMetricsSink is where the second attribution pass's per-dispatch
// deltas are merged -- one call per dispatch a batch touched, keyed by the
// dispatch row's own id (DispatchAttributor.Attribute returns exactly that
// key).
//
// The patch is additive: the harvester sends a batch's *delta*, never a
// cumulative total, so an implementation must add numeric leaves onto
// whatever it already holds rather than replacing them, exactly as
// CommitHarvestBatch's stage-run patches do. store.Store.MergeDispatchMetrics
// (whose signature already matches this interface exactly, no adapter
// needed) merges through jsonb_deep_add for that reason.
//
// Defined here, at the consumer, per go-interface-design, like every other
// interface in this file: internal/harvest never imports internal/store.
type DispatchMetricsSink interface {
	MergeDispatchMetrics(ctx context.Context, dispatchID int64, patch json.RawMessage) error
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

// maxDispatchMetaBackfillCycles bounds how many RunOnce cycles
// maybeBackfillDispatchMeta keeps re-reading one path's sidecar before
// giving up on it (F31, pass 7 of this change's own review panel) --
// the same bound value and the same rationale as
// maxSessionTokenResolutionCycles above: a permanently missing or
// corrupt sidecar must not make this re-read the same path forever.
const maxDispatchMetaBackfillCycles = maxSessionTokenResolutionCycles

// SessionTokenBinder resolves the session tokens a run generates once and
// passes on every mark it makes (KAN-172, task 1; reworked from one
// correlator per mark to one per session in task 4b) into session_id
// bindings, once a harvest cycle has located the transcript that carries
// one. Defined here, at the consumer, per go-interface-design, exactly
// like WindowSource, HarvestSink and Pricer above: internal/harvest never
// imports internal/store, so this package is testable against a fake with
// no PostgreSQL required. The daemon wires a real implementation backed by
// *store.Store, whose UnresolvedSessionTokens, BindSession,
// RecordSessionTokenGiveUp, PersistedGiveUps and
// MarkDispatchesUnattributedByID methods are written to match this
// interface exactly -- widened by task 6 (tasks.md,
// kan-212-persist-per-dispatch-cost-tokens-model-and-role) to carry the
// give-up half of that change alongside the binding half it already
// carried.
//
// GiveUp is declared here, in internal/harvest, rather than in
// internal/store -- store.Store.PersistedGiveUps returns it directly, with
// no adapter, exactly as store.Store.DispatchWindowsForSession already
// returns DispatchWindow. The dependency runs store -> harvest, never the
// reverse, which is what keeps TestHarvestNeedsNoDatabase true even as
// this interface grows.
//
// Widened again by task 6.1 (tasks.md) to add MarkDispatchesUnattributed,
// the token form: task 6 stamped a dispatch-grain ambiguity by id
// (MarkDispatchesUnattributedByID) but left the give-up itself stamping
// nothing, so the "session never bound" state task 8 renders had no
// producer.
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
	// RecordSessionTokenGiveUp persists that this Watcher has stopped
	// searching for token, with reason naming why and at the instant it
	// gave up -- resolveSessionTokens' own two give-up branches call this
	// before setting w.gaveUpTokens, so a token abandoned by this process
	// is still discoverable, and re-attemptable, by a later one
	// (design.md, "The abandoned-token set is persisted, and retried on
	// restart"). See store.Store.RecordSessionTokenGiveUp's own doc
	// comment for the upsert semantics: a token recorded a second time
	// updates rather than duplicates, and its retries count rises.
	RecordSessionTokenGiveUp(ctx context.Context, token, reason string, at time.Time) error
	// PersistedGiveUps returns every session token this Watcher's store
	// has ever given up on, across every process that has ever run
	// against it. seedPersistedGiveUps reads this exactly once, on this
	// Watcher's first cycle, so a token a prior process abandoned is
	// searched for again -- kan-302's recovery path: a transcript that
	// now carries the mark a prior process never found binds on this
	// process's own bounded window.
	PersistedGiveUps(ctx context.Context) ([]GiveUp, error)
	// MarkDispatchesUnattributedByID stamps exactly the dispatches named
	// by ids with the reason their cost could not be attributed, and how
	// many candidates could not be told apart. attributeDispatches calls
	// this for the dispatch-grain second pass's own ambiguous outcome
	// (DispatchAttributor.Attribute's AmbiguousDispatch) -- by dispatch
	// id, not by session token, because the candidates are specific rows
	// and stamping every dispatch under their shared session would also
	// stamp siblings that attributed correctly.
	MarkDispatchesUnattributedByID(ctx context.Context, ids []int64, reason string, candidates int) error
	// MarkDispatchesUnattributed stamps every dispatch recorded under
	// token with the reason its session's cost could not be attributed,
	// and, when positive, how many candidates an ambiguous match could
	// not tell apart -- resolveSessionTokens' own two give-up branches
	// call this immediately after RecordSessionTokenGiveUp, with that
	// branch's own reason (task 6.1, tasks.md, "stamp the dispatches of
	// a session that never bound"). Distinct from
	// MarkDispatchesUnattributedByID above, and not a substitute for it:
	// a session that never bound leaves every one of its dispatches
	// uncosted, which this token form expresses by stamping the whole
	// session at once; a dispatch-grain ambiguity concerns specific
	// rows, which the id form expresses by naming exactly those rows and
	// none of their attributed siblings.
	MarkDispatchesUnattributed(ctx context.Context, token, reason string, candidates int) error
}

// GiveUp is one persisted record of a session token this Watcher (or an
// earlier process sharing its store) searched for and could not resolve --
// the token, why, and how many times the search has come back to it.
// store.Store.PersistedGiveUps and store.Store.RecordSessionTokenGiveUp
// are written to return and accept this shape exactly (its own doc
// comment on 0013_session_token_giveups.sql has the schema).
type GiveUp struct {
	Token   string
	Reason  string
	Retries int
}

// reasonSessionNeverBound and reasonSessionAmbiguous are the two distinct
// reasons resolveSessionTokens persists a session-token give-up under
// (task 6, tasks.md, "each with its own distinct reason") -- case 0, the
// bounded window exhausted with no match at all, and the default branch,
// the token matched more than one session. Naming them apart is what lets
// PersistedGiveUps' caller -- and, eventually, a future reader of
// session_token_giveups.reason -- tell a token that never appeared from
// one that appeared twice, which are different failures kan-302's own
// investigation needs told apart.
const (
	reasonSessionNeverBound = "session never bound"
	reasonSessionAmbiguous  = "matched more than one session"
	// reasonDispatchAmbiguous is what attributeDispatches persists for a
	// dispatch-grain ambiguity (DispatchAttributor.Attribute's
	// AmbiguousDispatch) -- distinct from either session-token reason
	// above, since this failure is not about a session ever binding at
	// all: it is two or more already-bound dispatches that a record's
	// agent id or timestamp could not tell apart (bestDispatchWindow's
	// own doc comment, attribute.go).
	reasonDispatchAmbiguous = "matched more than one dispatch"
)

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

	// dispatchAttributor and dispatchMetrics are the second attribution
	// pass (design.md, "Cost attribution"), configured together by
	// WithDispatchAttribution or not at all: a Watcher built without that
	// option attributes stage runs exactly as it did before this pass
	// existed, the same additive shape WithPricer established.
	dispatchAttributor *DispatchAttributor
	dispatchMetrics    DispatchMetricsSink

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

	// seededGiveUps reports whether seedPersistedGiveUps has already run
	// (task 6, tasks.md, "seed the pending set from PersistedGiveUps at
	// start"): it must read PersistedGiveUps exactly once, on this
	// Watcher's first cycle, never on every cycle -- TestRetryStillBounded
	// and TestPersistedGiveUpIsRetriedOnStart both pin the call count at
	// exactly 1 across many RunOnce calls. A failed read leaves this
	// false, so the next cycle tries again rather than abandoning
	// recovery for this process's whole lifetime -- the same "log and
	// retry next cycle" discipline every other store call in this file
	// follows.
	seededGiveUps bool

	// retriedTokens holds every token seedPersistedGiveUps' single read of
	// PersistedGiveUps ever returned -- populated there, alongside
	// seededGiveUps, and never anywhere else (task 6.2, tasks.md, "bind a
	// retried token by scanning, not by waiting for new bytes"). It is
	// what scanRetriedTokens (below) gates its own extra, whole-file
	// reads on: a token this process is trying for the first time is
	// still being written to a live transcript, so waiting for new bytes
	// (matchSessionTokens, the ordinary path) is both correct and cheap
	// for it; only a *retried* give-up's own marks are provably behind an
	// offset that will never move for them again (harvest_offsets already
	// sits at that transcript's own EOF, by definition of having given up
	// once already), which is the one case worth paying for a read
	// proportional to every transcript on disk. Membership here is never
	// revoked once a token binds or gives up again -- the set stays tiny
	// (bounded by how many give-ups this store has ever persisted) and
	// pendingSessionTokens already stops offering a bound or re-given-up
	// token to scanRetriedTokens at all, so a stale entry here costs
	// nothing.
	retriedTokens map[string]bool

	// pendingDispatchMeta remembers, per subagent transcript path whose
	// most recently committed batch carried dispatch tokens but no
	// descriptors (ReadDispatchMeta found no sidecar at commit time), the
	// stage-run-id -> agentId pairs that batch attributed tokens to (F4,
	// pass 1 of this change's own review panel). A sidecar written after
	// its transcript has already been fully harvested -- the meta file is
	// flushed on a slightly different schedule than the transcript line
	// it describes, so this is not a rare race -- would otherwise never
	// get another chance: newOffset == offset on every later cycle for
	// that path (nothing new to read), which skipped ReadDispatchMeta
	// entirely before this map existed, permanently committing
	// hasMeta=false. maybeBackfillDispatchMeta (below) is what spends
	// this map: on exactly that "nothing new" branch, it re-reads the
	// sidecar and, once found, commits a descriptors-only patch (zero
	// TokenDelta, so nothing sums twice) to the remembered stage run ids,
	// never re-attributing the transcript's own already-committed tokens.
	//
	// Like tokenCycles and gaveUpTokens above, this is purely local,
	// in-memory bookkeeping, not persisted: a daemon restart loses
	// pending entries, which only means a sidecar that arrived late
	// during the previous process's lifetime and had not yet been
	// backfilled goes unbackfilled after a restart -- not a correctness hazard, since
	// the tokens themselves were already committed durably either way,
	// only the descriptors would stay absent. Not guarded by a mutex, for
	// the same reason those two maps are not: RunOnce is never called
	// concurrently with itself on one Watcher.
	pendingDispatchMeta map[string]map[int64]string

	// dispatchMetaCycles and gaveUpDispatchMeta bound maybeBackfillDispatchMeta
	// the same way tokenCycles and gaveUpTokens above bound the session-token
	// retry path (F31, pass 7 of this change's own review panel): before
	// these existed, a permanently missing or corrupt sidecar made
	// maybeBackfillDispatchMeta re-read the same path on every idle cycle
	// for this process's whole lifetime. dispatchMetaCycles counts, per
	// path, how many cycles have found still no sidecar; once it reaches
	// maxDispatchMetaBackfillCycles, gaveUpDispatchMeta[path] is set,
	// pendingDispatchMeta[path] is dropped, and RunOnce's own "else"
	// branch above stops re-adding entries for that path -- a give-up
	// that actually stops looking, not merely stops logging, exactly like
	// gaveUpTokens. Neither map is persisted or mutex-guarded, for the
	// same reasons tokenCycles and gaveUpTokens are not.
	dispatchMetaCycles map[string]int
	gaveUpDispatchMeta map[string]bool
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

// WithDispatchAttribution configures the Watcher to run the second,
// dispatch-grain attribution pass beside the first: the same batch of
// records the stage pass reads is attributed again against a's dispatch
// windows, and each touched dispatch's delta is merged into sink.
//
// The attributor and the sink are one option rather than two because
// neither is any use without the other -- an attributor with nowhere to
// write computes deltas that are discarded, and a sink with nothing
// computing for it is never called -- so a Watcher can never be
// half-configured for this pass.
func WithDispatchAttribution(a *DispatchAttributor, sink DispatchMetricsSink) WatcherOption {
	return func(w *Watcher) {
		w.dispatchAttributor = a
		w.dispatchMetrics = sink
	}
}

// HasDispatchAttribution reports whether this Watcher was configured with
// WithDispatchAttribution. See HasPricer's doc comment for why this exists
// and why cmd/myflowd's wiring test asserts the constructed value rather
// than main.go's source text.
func (w *Watcher) HasDispatchAttribution() bool {
	return w.dispatchAttributor != nil && w.dispatchMetrics != nil
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
		root:                root,
		sink:                sink,
		attributor:          attributor,
		logger:              logger,
		tokenCycles:         make(map[string]int),
		gaveUpTokens:        make(map[string]bool),
		retriedTokens:       make(map[string]bool),
		pendingDispatchMeta: make(map[string]map[int64]string),
		dispatchMetaCycles:  make(map[string]int),
		gaveUpDispatchMeta:  make(map[string]bool),
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
			// Nothing new (or only a partial trailing line) since last
			// time -- but a sidecar this path's earlier batch could not
			// find may have landed since then (F4), so give it one
			// chance before moving on rather than skipping this path
			// outright.
			w.maybeBackfillDispatchMeta(ctx, path)
			continue
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

		// ReadDispatchMeta reads path's own sidecar, once per file, right
		// here -- this loop already holds path, which Attribute never
		// sees (it works from records alone). A subagent transcript's
		// every record shares this same file's own agentId (KAN-201's
		// tasks.md, "Facts this plan rests on"), so the single meta value
		// read here applies to every dispatch entry any of this batch's
		// deltas carry. hasMeta is false, and meta is the zero value, for
		// a main-session transcript or a subagent transcript with no
		// sidecar -- encodePatches then omits the descriptors entirely
		// rather than inventing them (DispatchBucket's own doc comment,
		// attribute.go).
		meta, hasMeta := ReadDispatchMeta(path)

		patches, err := encodePatches(deltas, meta, hasMeta)
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

		// The second, dispatch-grain attribution pass over the very same
		// records, in the same batch -- deliberately here, after
		// CommitHarvestBatch has reported the batch applied, never before
		// it. A batch withheld, refused or lost to a concurrent harvester
		// is one this cycle will read again from the same offset, and
		// merging its dispatch deltas now would add them a second time
		// when it does.
		w.attributeDispatches(ctx, records, path)

		// Once this batch has actually committed, decide whether this
		// path needs a future backfill visit (F4): hasMeta true means
		// descriptors just landed for real, real content, so any
		// earlier pending entry is now stale and cleared; hasMeta false
		// means every dispatch this batch touched still has no
		// descriptors, so each is (re-)recorded for
		// maybeBackfillDispatchMeta to revisit once nothing new remains
		// to read from this path.
		if hasMeta {
			// Clear only the pending entries this batch actually delivered
			// descriptors for (F29, pass 7 of this change's own review
			// panel), never the whole per-path map. deltas can legitimately
			// be empty here -- a batch can have newOffset != offset while
			// parsing zero assistant records (interleaved tool_use/
			// tool_result lines do this routinely), and CommitHarvestBatch
			// still applies such a batch (its own doc comment says deltas
			// "may be empty"). Before this fix, an entry a strictly earlier
			// batch left pending for this same path -- while hasMeta was
			// false -- was wiped by any later batch that happened to find
			// hasMeta true, even one that delivered no descriptors at all
			// for that stageRunID, and maybeBackfillDispatchMeta never
			// revisits an entry it no longer holds: those descriptors were
			// lost permanently. A stageRunID only clears here when this
			// batch's own deltas actually carried dispatch entries for it,
			// which is exactly when encodePatches (above) just attached
			// this batch's meta to every one of them.
			if pending, ok := w.pendingDispatchMeta[path]; ok {
				for stageRunID, delta := range deltas {
					if len(delta.Dispatches) == 0 {
						continue
					}
					delete(pending, stageRunID)
				}
				if len(pending) == 0 {
					delete(w.pendingDispatchMeta, path)
				}
			}
		} else if !w.gaveUpDispatchMeta[path] {
			// Never re-add an entry for a path this Watcher has already
			// given up backfilling (F31, pass 7 of this change's own
			// review panel) -- mirrors pendingSessionTokens' own
			// gaveUpTokens filter above: a give-up must actually stop this
			// Watcher from looking, not just stop it from logging.
			for stageRunID, delta := range deltas {
				for agentID := range delta.Dispatches {
					if w.pendingDispatchMeta[path] == nil {
						w.pendingDispatchMeta[path] = make(map[int64]string)
					}
					w.pendingDispatchMeta[path][stageRunID] = agentID
				}
			}
		}

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

	w.scanRetriedTokens(files, pendingSessionTokens, matchedSessions)
	w.resolveSessionTokens(ctx, pendingSessionTokens, matchedSessions)

	return touchedFiles, nil
}

// scanRetriedTokens looks for a persisted give-up's own mark invocation
// in bytes harvest_offsets has already read past -- task 6.2's fix for
// the gap task 12's own live restart measured (tasks.md, "bind a retried
// token by scanning, not by waiting for new bytes"): matchSessionTokens
// (above) only ever sees the Bash commands newly read from a transcript
// this cycle, and a give-up's own transcript has, by definition, already
// been fully harvested by the time it gave up -- kan-302's own measured
// state, harvest_offsets already sitting at that file's full 3,147,539
// bytes. A plain retry that only waits for new bytes searches a stream
// that will never grow and gives up again, permanently, for exactly the
// transcripts that matter: every one belonging to a run that has already
// finished.
//
// Gated on w.retriedTokens -- populated once, at seedPersistedGiveUps,
// from exactly the tokens PersistedGiveUps returned -- rather than on
// every still-pending token: an ordinary, never-before-given-up token's
// transcript is still being written, so waiting for new bytes is both
// correct and cheap for it, and paying for a whole-file read of every
// transcript on disk on its behalf would buy nothing. Only a retried
// give-up's marks are provably behind an offset that will never move for
// them again, which is the one case this extra read is worth its own
// cost -- "the scan is work proportional to the transcripts on disk, so
// gate it on there actually being pending tokens the newly-read batches
// did not match" (tasks.md, task 6.2, step 4), narrowed here to the
// tokens that can structurally never resolve any other way. filtering
// out a token matched already (len(matched[token]) > 0) keeps a token the
// ordinary path already found this same cycle from paying for this scan
// too.
//
// Reads with ReadAllCommands (transcript.go), never ReadNewRecords or
// w.attributor.Attribute: no Record is ever produced from these bytes,
// so there is nothing here that could be attributed, and harvest_offsets
// for path is neither read nor written by this call -- property 1,
// "never re-attribute usage" (tasks.md, task 6.2's own "non-negotiable
// properties").
//
// The per-token cycle bound (maxSessionTokenResolutionCycles,
// resolveSessionTokens) is untouched by this method: a retried token
// this scan still cannot find still counts a cycle in w.tokenCycles on
// every RunOnce, exactly as before, and still gives up again once that
// bound is reached -- a token whose marks genuinely are not on disk
// anywhere rescans on a bounded schedule, not forever.
//
// isSessionMarkCommand -- reached the same way the ordinary path reaches
// it, through matchSessionTokens -- is what keeps the
// mention-versus-invocation distinction (matchSessionTokens' own doc
// comment, KAN-172 finding F4) and the ambiguity rule identical on this
// path: a diagnostic grep for a retried token sitting in these very
// already-consumed bytes still does not bind it
// (TestPersistedGiveUpBindsFromAFullyConsumedTranscript's own "mention
// only" subtest), and a token whose genuine marks turn out to sit in two
// transcripts still refuses to bind rather than picking one (that test's
// own "ambiguous" subtest) -- this new path reads exactly the
// already-read-past region F4's live reproduction did, so admitting
// anything looser than a real mark invocation here would walk straight
// back into that bug.
func (w *Watcher) scanRetriedTokens(files []string, pending map[int64]string, matched map[string]map[string]bool) {
	if len(w.retriedTokens) == 0 {
		return
	}
	toScan := make(map[int64]string, len(pending))
	for stageRunID, token := range pending {
		if !w.retriedTokens[token] {
			continue
		}
		if len(matched[token]) > 0 {
			continue // already found via this cycle's newly-read bytes
		}
		toScan[stageRunID] = token
	}
	if len(toScan) == 0 {
		return
	}

	for _, path := range files {
		commands, err := ReadAllCommands(path)
		if err != nil {
			w.warn("harvest: scan transcript for a retried session token failed, will retry", "path", path, "error", err)
			continue
		}
		w.matchSessionTokens(toScan, commands, matched)
	}
}

// attributeDispatches runs the second attribution pass over one batch's
// records and merges each touched dispatch's delta into the configured
// sink. A Watcher built without WithDispatchAttribution does nothing here.
//
// Every failure is logged and stepped over, never returned: the first
// pass has already committed this batch atomically by the time this runs,
// and there is nothing left for a failure here to protect. Its cost is
// real and worth stating plainly -- a merge that fails loses that batch's
// dispatch figures for good, because the offset has already moved past
// the records they were computed from. That asymmetry is accepted rather
// than overlooked: stage-run attribution is the figure every existing
// aggregation reads, and making its commit wait on a second write would
// put the established grain at the mercy of the new one. A dispatch's
// bag is refilled by the next batch inside the same window in the
// ordinary case, and a dispatch that ends before one arrives is left
// understated rather than wrong -- a whole stage run's usage would be
// neither.
//
// Where DispatchAttributor.Attribute reports an ambiguity -- a record
// whose agent id or timestamp matched more than one dispatch -- this
// stamps every candidate as unattributed (task 6, tasks.md, "do not
// stamp a dispatch that attributed"), through
// SessionTokenBinder.MarkDispatchesUnattributedByID rather than
// dispatchMetrics: the stamp is store bookkeeping about a dispatch's
// metrics bag, the same shape RecordSessionTokenGiveUp and
// PersistedGiveUps already are, not a token delta. A Watcher configured
// with WithDispatchAttribution but no WithSessionTokenBinder has nowhere
// to write the stamp and skips it -- the ambiguity is still correctly
// left uncredited either way, only the reason goes unrecorded.
func (w *Watcher) attributeDispatches(ctx context.Context, records []Record, path string) {
	if !w.HasDispatchAttribution() {
		return
	}

	deltas, ambiguous, err := w.dispatchAttributor.Attribute(ctx, records)
	if err != nil {
		w.warn("harvest: attribute dispatch windows failed, this batch's dispatch figures are lost", "path", path, "error", err)
		return
	}

	for dispatchID, tokens := range deltas {
		patch, err := json.Marshal(MetricsPatch{Tokens: tokens})
		if err != nil {
			w.warn("harvest: encode dispatch metrics failed", "path", path, "dispatch_id", dispatchID, "error", err)
			continue
		}
		if err := w.dispatchMetrics.MergeDispatchMetrics(ctx, dispatchID, patch); err != nil {
			w.warn("harvest: merge dispatch metrics failed, this batch's figures for it are lost", "path", path, "dispatch_id", dispatchID, "error", err)
		}
	}

	if w.sessionTokens == nil {
		return
	}
	for _, a := range ambiguous {
		// Filter out any candidate this very call already merged real
		// tokens for above: DispatchAttributor.Attribute computes deltas
		// and ambiguous independently, per record, so the same batch can
		// legitimately attribute one record to a dispatch cleanly while a
		// different record's agent id remains ambiguous between that same
		// dispatch and another one (bestDispatchWindow's identity pass
		// matches on AgentID alone, ignoring the record's own timestamp,
		// so two windows sharing an AgentID make every id-carrying record
		// ambiguous across both regardless of which one it actually falls
		// in). Stamping an attributed dispatch unattributed here would
		// contradict task 6 step 4's own words, "Do not stamp a dispatch
		// that attributed" -- so a candidate already present in deltas is
		// dropped from the stamp, never merely overwritten by it.
		//
		// candidates is still the ambiguity's full, unfiltered size
		// (len(a.DispatchIDs)), not len(ids): what a stamped dispatch
		// could not be told apart from is a fact about that record's
		// identity match, unchanged by whether a sibling candidate
		// happened to attribute something else in this same batch.
		var ids []int64
		for _, id := range a.DispatchIDs {
			if _, attributed := deltas[id]; attributed {
				continue
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			continue
		}
		if err := w.sessionTokens.MarkDispatchesUnattributedByID(ctx, ids, reasonDispatchAmbiguous, len(a.DispatchIDs)); err != nil {
			w.warn("harvest: mark dispatches unattributed failed", "path", path, "dispatch_ids", ids, "error", err)
		}
	}
}

// maybeBackfillDispatchMeta is called from RunOnce's "nothing new to
// read" branch for path (F4): a batch already committed for path may
// have left w.pendingDispatchMeta[path] populated because its sidecar
// was not there yet at the time. This re-reads the sidecar and, once
// found, commits a descriptors-only patch -- MetricsPatch's zero-value
// TokenDelta contributes nothing to any sum, so this never touches, and
// never risks double-counting, the tokens that batch already committed.
// A patch is written per remembered stage-run-id/agentId pair, since a
// backfill entry has never (yet) been re-attributed and RunOnce's own
// per-file loop above builds no new Delta for this call to draw on.
//
// A GetHarvestOffset/CommitHarvestBatch round trip using the *current*
// offset as both expected and new value is deliberate, not a copy-paste
// of the ordinary commit path above: it reuses CommitHarvestBatch's own
// optimistic-concurrency guard (HarvestSink's own doc comment) to detect
// a concurrent harvester racing this same path, while never asking the
// offset to move -- this call has no new bytes to account for, only
// descriptors for tokens a past batch already committed.
//
// Still no sidecar (hasMeta false), a failure encoding or committing the
// patch, or losing the concurrency race all leave path's entry in
// w.pendingDispatchMeta untouched, so a later cycle gets another chance
// -- exactly like every other retry path in this file. Only a successful
// commit clears it.
func (w *Watcher) maybeBackfillDispatchMeta(ctx context.Context, path string) {
	pending, ok := w.pendingDispatchMeta[path]
	if !ok || len(pending) == 0 {
		return
	}

	meta, hasMeta := ReadDispatchMeta(path)
	if !hasMeta {
		// Bounded retry, mirroring resolveSessionTokens' own give-up
		// bookkeeping (F31, pass 7 of this change's own review panel): a
		// sidecar that never arrives must not be re-read forever.
		w.dispatchMetaCycles[path]++
		if w.dispatchMetaCycles[path] < maxDispatchMetaBackfillCycles {
			return // still no sidecar; try again next cycle.
		}
		w.warn("harvest: dispatch-meta sidecar unresolved after the bounded window, giving up",
			"path", path, "cycles", w.dispatchMetaCycles[path])
		w.gaveUpDispatchMeta[path] = true
		delete(w.dispatchMetaCycles, path)
		delete(w.pendingDispatchMeta, path)
		return
	}
	delete(w.dispatchMetaCycles, path)

	patches := make(map[int64]json.RawMessage, len(pending))
	for stageRunID, agentID := range pending {
		depth := strconv.Itoa(meta.SpawnDepth)
		mp := MetricsPatch{
			Dispatches: map[string]DispatchBucket{
				agentID: {
					AgentType:   meta.AgentType,
					Description: meta.Description,
					Model:       meta.Model,
					SpawnDepth:  &depth,
				},
			},
		}
		patch, err := json.Marshal(mp)
		if err != nil {
			w.warn("harvest: encode dispatch-meta backfill failed, will retry", "path", path, "stage_run_id", stageRunID, "error", err)
			continue
		}
		patches[stageRunID] = patch
	}
	if len(patches) == 0 {
		return
	}

	offset, found, err := w.sink.GetHarvestOffset(ctx, path)
	if err != nil {
		w.warn("harvest: get committed offset for dispatch-meta backfill failed, will retry", "path", path, "error", err)
		return
	}

	applied, err := w.sink.CommitHarvestBatch(ctx, path, offset, found, offset, patches)
	if err != nil {
		w.warn("harvest: commit dispatch-meta backfill failed, will retry", "path", path, "error", err)
		return
	}
	if !applied {
		// Lost a race with a concurrent harvester for this file -- benign,
		// exactly like the ordinary commit path above; retry next cycle.
		return
	}

	delete(w.pendingDispatchMeta, path)

	// Price every stage run this backfill just gave a model to (F11, pass
	// 2 of this change's own review panel): the ordinary commit path's
	// own pricing pass (RunOnce, above) already ran for these stage runs
	// while this dispatch's model was still empty, and pricing.go's
	// `if db.Model == "" { continue }` skipped it -- without a pricing
	// pass here too, that dispatch's cost_usd stays permanently absent
	// even though a model is now on record for it. Deliberately after
	// the commit has reported applied, never inside it, and a pricing
	// failure is warned about rather than fatal, matching the ordinary
	// commit path's own discipline exactly (Pricer's own doc comment,
	// and the comment on that call site above).
	if w.pricer != nil {
		for stageRunID := range patches {
			if err := w.pricer.Price(ctx, stageRunID); err != nil {
				w.warn("harvest: price stage run failed, will retry next cycle", "stage_run_id", stageRunID, "error", err)
			}
		}
	}
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
	w.seedPersistedGiveUps(ctx)
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

// seedPersistedGiveUps reads every give-up this Watcher's binder has ever
// persisted, once, on this Watcher's first cycle (task 6, tasks.md, "seed
// the pending set from PersistedGiveUps at start") -- kan-302's recovery
// path: a token an earlier process abandoned is searched for again once
// this process starts, since its transcript may carry the mark by now.
//
// Clearing each returned token from w.gaveUpTokens is defensive rather
// than load-bearing today -- w.gaveUpTokens starts empty on every new
// Watcher, so nothing has set an entry before this first call runs -- but
// it keeps the seeded set's actual membership correct by construction
// rather than by the accident of call order.
//
// w.tokenCycles is never touched here: it too starts empty, so a
// retried token gets a fresh bounded window of exactly
// maxSessionTokenResolutionCycles, the same as the first attempt (task 6,
// tasks.md, "the retry is bounded exactly as the first attempt is").
//
// w.seededGiveUps guards this to exactly one call across this Watcher's
// whole lifetime, successful call included -- a failed read leaves it
// false, so the next cycle tries again, the same "log and retry next
// cycle" discipline every other store call in this file follows, rather
// than abandoning recovery for this process's whole lifetime.
func (w *Watcher) seedPersistedGiveUps(ctx context.Context) {
	if w.seededGiveUps {
		return
	}
	giveUps, err := w.sessionTokens.PersistedGiveUps(ctx)
	if err != nil {
		w.warn("harvest: list persisted give-ups failed, will retry", "error", err)
		return
	}
	for _, g := range giveUps {
		delete(w.gaveUpTokens, g.Token)
		w.retriedTokens[g.Token] = true
	}
	w.seededGiveUps = true
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
			if err := w.sessionTokens.RecordSessionTokenGiveUp(ctx, sessionToken, reasonSessionNeverBound, time.Now()); err != nil {
				w.warn("harvest: persist give-up failed, will retry", "stage_run_ids", stageRunIDs, "error", err)
				continue
			}
			// A stamp failure here must never block the give-up already
			// recorded above (task 6.1, tasks.md, step 4): it is logged
			// and this branch still proceeds to mark the token given up.
			if err := w.sessionTokens.MarkDispatchesUnattributed(ctx, sessionToken, reasonSessionNeverBound, 0); err != nil {
				w.warn("harvest: stamp unattributed dispatches failed", "stage_run_ids", stageRunIDs, "session_token", sessionToken, "error", err)
			}
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
			if err := w.sessionTokens.RecordSessionTokenGiveUp(ctx, sessionToken, reasonSessionAmbiguous, time.Now()); err != nil {
				w.warn("harvest: persist give-up failed, will retry", "stage_run_ids", stageRunIDs, "error", err)
				continue
			}
			// Same never-block guarantee as the case 0 branch above: a
			// stamp failure here is logged, never a reason to undo the
			// give-up already recorded.
			if err := w.sessionTokens.MarkDispatchesUnattributed(ctx, sessionToken, reasonSessionAmbiguous, len(sessions)); err != nil {
				w.warn("harvest: stamp unattributed dispatches failed", "stage_run_ids", stageRunIDs, "session_token", sessionToken, "error", err)
			}
			w.gaveUpTokens[sessionToken] = true
			delete(w.tokenCycles, sessionToken)
		}
	}
}

// encodePatches marshals each stage run's Delta into the json.RawMessage
// shape HarvestSink.CommitHarvestBatch's deltas parameter expects: the
// whole-run total under "tokens" (Delta.Total, unchanged meaning), when
// this batch carried any, the per-model breakdown under "models"
// (Delta.Models, task 22), and, when this batch carried any, the
// per-dispatch breakdown under "dispatches" (Delta.Dispatches, KAN-201)
// -- each omitted entirely rather than encoded as an empty object when a
// batch attributed nothing to it, so it never masks a stage run's own
// key already holding real data from an earlier batch (json.RawMessage(nil)
// participates in no merge at all, where an encoded "models":{} would
// still be a value jsonb_deep_add has to reconcile against).
//
// meta and hasMeta are RunOnce's own single ReadDispatchMeta read of the
// transcript file this batch of deltas came from (its per-file loop,
// above) -- the descriptors every entry in every delta's Dispatches gets,
// since every record in one transcript file shares that file's own
// agentId (KAN-201's tasks.md, "Facts this plan rests on"). hasMeta
// false leaves a DispatchBucket's descriptor fields at their zero value,
// which their own omitempty tags then drop from the JSON entirely --
// never an invented "" or 0.
func encodePatches(deltas map[int64]Delta, meta DispatchMeta, hasMeta bool) (map[int64]json.RawMessage, error) {
	patches := make(map[int64]json.RawMessage, len(deltas))
	for stageRunID, delta := range deltas {
		mp := MetricsPatch{Tokens: delta.Total, Speed: delta.Speed}
		if len(delta.Models) > 0 {
			// A plain nil-check-and-assign, not upsertBucket (F35, pass 7
			// of this change's own review panel): mp.Models starts nil for
			// this stageRunID on every call, so there is never an existing
			// value to fold in -- upsertBucket's own "get-or-zero, mutate,
			// store back" shape reads as accumulation it never actually
			// performs here. attribute.go's two Attribute call sites keep
			// upsertBucket, correctly: those genuinely read and mutate an
			// existing value across multiple records in the same batch.
			mp.Models = make(map[string]ModelBucket, len(delta.Models))
			for model, td := range delta.Models {
				mp.Models[model] = ModelBucket{Tokens: td}
			}
		}
		if len(delta.Dispatches) > 0 {
			mp.Dispatches = make(map[string]DispatchBucket, len(delta.Dispatches))
			for agentID, td := range delta.Dispatches {
				db := DispatchBucket{Tokens: td}
				if hasMeta {
					db.AgentType = meta.AgentType
					db.Description = meta.Description
					db.Model = meta.Model
					// A fresh local, not &meta.SpawnDepth directly (F5):
					// meta is one shared value for every dispatch this
					// call encodes, and while nothing here currently
					// mutates it after this point, giving each entry
					// its own addressable copy keeps that true by
					// construction rather than by the loop body never
					// changing again. strconv.Itoa converts
					// meta.SpawnDepth (a plain int, from
					// ReadDispatchMeta) to DispatchBucket's own wire
					// representation (F23: SpawnDepth is *string, per
					// that field's own doc comment in attribute.go).
					depth := strconv.Itoa(meta.SpawnDepth)
					db.SpawnDepth = &depth
				}
				mp.Dispatches[agentID] = db
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
