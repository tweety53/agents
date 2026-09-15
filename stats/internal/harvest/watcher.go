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
// internal/fallback.StateRoot's FLOW_STATE_DIR already uses for the
// state directory.
const DefaultTranscriptsRootEnv = "FLOW_TRANSCRIPTS_DIR"

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
// cmd/flowd/main.go asserts this at compile time).
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
// nothing prevents two Watchers -- two flowd processes, one stale
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
// daemon, as part of the *store.Store it passes as NewWatcher's deps
// (cmd/flowd/main.go).
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
// cmd/flowd's own harvestInterval (5s, main.go), 60 cycles is 5
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
// RecordSessionTokenGiveUp, PersistedGiveUps and MarkDispatchesUnattributed
// methods are written to match this interface exactly -- widened by task 6
// (tasks.md, kan-212-persist-per-dispatch-cost-tokens-model-and-role) to
// carry the give-up half of that change alongside the binding half it
// already carried.
//
// GiveUp is declared here, in internal/harvest, rather than in
// internal/store -- store.Store.PersistedGiveUps returns it directly, with
// no adapter, exactly as store.Store.DispatchWindowsForSession already
// returns DispatchWindow. The dependency runs store -> harvest, never the
// reverse, which is what keeps TestHarvestNeedsNoDatabase true even as
// this interface grows.
//
// Widened again by task 6.1 (tasks.md) to add MarkDispatchesUnattributed,
// the token form: the give-up itself was stamping nothing, so the "session
// never bound" state task 8 renders had no producer. Its
// MarkDispatchesUnattributedByID sibling left the interface with KAN-414:
// a dispatch-grain ambiguity stopped being a stampable failure when it
// became an apportionment.
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
	// MarkDispatchesUnattributed stamps every dispatch recorded under
	// token with the reason its session's cost could not be attributed,
	// and, when positive, how many candidates an ambiguous match could
	// not tell apart -- resolveSessionTokens' own two give-up branches
	// call this immediately after RecordSessionTokenGiveUp, with that
	// branch's own reason (task 6.1, tasks.md, "stamp the dispatches of
	// a session that never bound"). A session that never bound leaves
	// every one of its dispatches
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
//
// The dispatch-grain ambiguity this block once named a third reason for
// ("matched more than one dispatch") is no longer a failure at all:
// KAN-414 made it an apportionment, and the record's spend lands on the
// candidates pro-rata instead of being written off (apportionRecord,
// attribute.go).
const (
	reasonSessionNeverBound = "session never bound"
	reasonSessionAmbiguous  = "matched more than one session"
)

// Source is one transcript source: a root to walk for *.jsonl files and the
// pair of read functions that parse those files' own line shape. The Claude
// transcript format and ZCode's rollout format share the offset, attribution
// and commit machinery but not their line shapes, so the parser rides with
// the root instead of being hardcoded in the read path -- one Watcher runs
// both, and a third harness with a machine-readable transcript is one new
// constructor, not a fork of this file.
type Source struct {
	Root string
	// ReadNew reads path from offset to EOF, exactly ReadNewRecords'
	// contract (transcript.go): records and commands from the complete
	// portion only, the new offset covering no partial trailing line.
	ReadNew func(path string, offset int64, openAgentCalls map[string]bool) (Batch, error)
	// ReadAllCmds reads path whole for commands only, exactly
	// ReadAllCommands' contract (transcript.go): no Records, no offset
	// read or written -- the retried-give-up scan's shape.
	ReadAllCmds func(path string) ([]CommandRecord, error)
}

// NewClaudeSource builds the source over a Claude Code transcripts root
// (~/.claude/projects, or FLOW_TRANSCRIPTS_DIR's override).
func NewClaudeSource(root string) Source {
	return Source{Root: root, ReadNew: ReadNewRecords, ReadAllCmds: ReadAllCommands}
}

// NewRolloutSource builds the source over a ZCode rollout root
// (~/.zcode/cli/rollout, or FLOW_ZCODE_ROLLOUTS_DIR's override).
func NewRolloutSource(root string) Source {
	return Source{Root: root, ReadNew: ReadRolloutNewRecords, ReadAllCmds: ReadRolloutAllCommands}
}

// transcriptSet is one source's discovered file list for a single RunOnce
// pass -- the pairing kept so scanRetriedTokens can read each file through
// the source whose parser produced it.
type transcriptSet struct {
	source Source
	files  []string
}

// Watcher periodically scans its sources' roots for *.jsonl files and
// harvests whatever bytes are new since each one's last committed offset,
// attributing them via its Attributor and committing the result -- both
// the token deltas and the advanced offset -- through its HarvestSink in
// one atomic call per file. It never talks to PostgreSQL directly --
// only through WindowSource (via Attributor), HarvestSink, and deps (its
// Pricer, SessionTokenBinder, DispatchMetricsSink and
// DispatchWindowSource).
type Watcher struct {
	sources    []Source
	sink       HarvestSink
	attributor *Attributor
	logger     *slog.Logger

	// deps is every dependency beyond root, sink and attributor -- pricing,
	// session-token binding, dispatch-metrics merging and dispatch-window
	// lookup (KAN-173) -- collapsed into one required parameter rather than
	// the three separate fields (pricer, sessionTokens, dispatchMetrics)
	// this struct used to carry. NewWatcher panics if it is nil, so every
	// method below can call it unconditionally: there is no "was this
	// configured" state left to guard against, only harvest.NoDeps' own
	// deliberate no-op.
	deps Deps

	// dispatchAttributor is the second, dispatch-grain attribution pass
	// (design.md, "Cost attribution"), built inside NewWatcher from deps
	// (which satisfies DispatchWindowSource) rather than supplied by a
	// caller -- there is then nothing for a caller to omit, unlike the
	// WithDispatchAttribution option this replaced.
	dispatchAttributor *DispatchAttributor

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

	// openAgentCalls carries, per transcript path, the agent tool_use ids
	// no tool_result has resolved yet -- the set the next read joins
	// denials against, so a dispatch whose launch never came can still be
	// recognised as denied when its error result lands in a later batch.
	// Bounded like the pending lists: a path whose set outgrows the cap
	// is cleared wholesale rather than grown without end, losing only the
	// ability to join a straddled denial -- never a pairing that already
	// happened.
	openAgentCalls map[string]map[string]bool

	// pendingDenials carries, per transcript path, the denied dispatch
	// attempts whose begin may not have arrived yet -- a panel round
	// records its begins AFTER the launches, so a denial can land a batch
	// before the begin it should retire. Bounded like the other pending
	// lists.
	pendingDenials map[string][]time.Time

	// pendingBegins and pendingLaunches carry, per transcript path, the
	// begins and launches KAN-322's pairing has seen but not yet paired.
	// A batch boundary may split the pair -- the launch result and the
	// begin command that names its row can commit in different harvest
	// batches, in either order (a panel round launches its slots in one
	// message and records their begins in one Bash call afterwards,
	// review-panel.md; a single dispatch records its begin immediately
	// before the launch, implement.md) -- so both sides survive into the
	// next RunOnce until paired or evicted. Both lists are bounded:
	// maxPendingStampEvents per path, oldest dropped, so a run of
	// begin-less launches or never-launched begins cannot grow the state
	// without end. A dropped event simply never stamps -- the ordinary
	// silence, never a wrong pairing.
	pendingBegins   map[string][]dispatchBeginEvent
	pendingLaunches map[string][]AgentLaunch
}

// NewWatcher builds a Watcher over sources (each root scanned recursively
// for *.jsonl files, each file parsed by its own source's read functions),
// sink (where offsets are read from and results are committed to) and
// attributor (how records become deltas). deps is
// every other dependency (KAN-173) -- required, not optional: NewWatcher
// panics if it is nil rather than silently building a Watcher that prices
// nothing, binds nothing and charges no dispatch, which is exactly the
// defect KAN-16 and KAN-172 both were. A caller with nothing to wire
// passes harvest.NoDeps{} to opt out explicitly. logger may be nil.
func NewWatcher(sources []Source, sink HarvestSink, attributor *Attributor, deps Deps, logger *slog.Logger) *Watcher {
	if deps == nil {
		panic("harvest.NewWatcher: deps is nil; pass harvest.NoDeps{} to opt out explicitly")
	}
	return &Watcher{
		sources:             sources,
		sink:                sink,
		attributor:          attributor,
		deps:                deps,
		dispatchAttributor:  NewDispatchAttributor(deps),
		logger:              logger,
		tokenCycles:         make(map[string]int),
		gaveUpTokens:        make(map[string]bool),
		retriedTokens:       make(map[string]bool),
		pendingDispatchMeta: make(map[string]map[int64]string),
		dispatchMetaCycles:  make(map[string]int),
		gaveUpDispatchMeta:  make(map[string]bool),
		pendingBegins:       make(map[string][]dispatchBeginEvent),
		pendingLaunches:     make(map[string][]AgentLaunch),
		pendingDenials:      make(map[string][]time.Time),
		openAgentCalls:      make(map[string]map[string]bool),
	}
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
// Withholding a batch that revealed a sessionToken: every file's newly
// read batch is checked for still-pending sessionTokens
// (matchSessionTokens) before it is attributed or committed at all -- a
// Watcher built with harvest.NoDeps (KAN-173) simply finds no pending
// tokens ever, since NoDeps.UnresolvedSessionTokens always reports none.
// A batch that matched one is *not* attributed or
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
	var sets []transcriptSet
	for _, source := range w.sources {
		files, err := discoverTranscripts(source.Root)
		if err != nil {
			return 0, err
		}
		sets = append(sets, transcriptSet{source: source, files: files})
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
	for _, set := range sets {
		for _, path := range set.files {
			if ctx.Err() != nil {
				return touchedFiles, ctx.Err()
			}

			offset, found, err := w.sink.GetHarvestOffset(ctx, path)
			if err != nil {
				w.warn("harvest: get committed offset failed, will retry", "path", path, "error", err)
				continue
			}

			batch, err := set.source.ReadNew(path, offset, w.openAgentCalls[path])
			if err != nil {
				w.warn("harvest: read transcript failed, will retry", "path", path, "error", err)
				continue
			}
			records, commands, launches, denials, newOffset :=
				batch.Records, batch.Commands, batch.Launches, batch.Denials, batch.NewOffset

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

			// KAN-322: the launch tool results this batch just committed name the
			// agent ids the dispatch protocol no longer asks a dispatcher to type.
			// Only reached here, after the batch applied: a batch withheld for
			// session-token resolution or lost to a concurrent harvester is
			// re-read from the same offset next cycle, so its launches are
			// processed exactly once too.
			w.stampDispatchAgents(ctx, path, commands, launches, denials)
			w.rememberOpenAgentCalls(path, batch.OpenCalls)

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
			} else if IsAgentTranscriptPath(path) && !w.gaveUpDispatchMeta[path] {
				// Never re-add an entry for a path this Watcher has already
				// given up backfilling (F31, pass 7 of this change's own review
				// panel) -- mirrors pendingSessionTokens' own gaveUpTokens filter above: a give-up must actually stop this
				// Watcher from looking, not just stop it from logging.
				// Gated on IsAgentTranscriptPath because ReadDispatchMeta
				// looks for its sidecar only beside a subagents/ directory:
				// any other path -- a ZCode rollout file, main-session or
				// subagent alike (KAN-506) -- can never grow one, so
				// remembering it here would buy maxDispatchMetaBackfillCycles
				// of futile re-reads and a give-up warning per dispatch for
				// descriptors no retry could ever produce.
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
			for stageRunID := range deltas {
				if err := w.deps.Price(ctx, stageRunID); err != nil {
					w.warn("harvest: price stage run failed, will retry next cycle", "stage_run_id", stageRunID, "error", err)
				}
			}
		}
	}

	w.scanRetriedTokens(sets, pendingSessionTokens, matchedSessions)
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
func (w *Watcher) scanRetriedTokens(sets []transcriptSet, pending map[int64]string, matched map[string]map[string]bool) {
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

	for _, set := range sets {
		for _, path := range set.files {
			commands, err := set.source.ReadAllCmds(path)
			if err != nil {
				w.warn("harvest: scan transcript for a retried session token failed, will retry", "path", path, "error", err)
				continue
			}
			w.matchSessionTokens(toScan, commands, matched)
		}
	}
}

// dispatchAgentIDForPath names the dispatch a transcript path belongs to --
// a Claude per-agent transcript by its subagents/ directory and file name
// (AgentIDFromTranscriptPath), a ZCode subagent dispatch's rollout file by
// its own file name (AgentIDFromRolloutPath) -- and reports false for a
// path that names none. Those two shapes are the only files that ARE one
// dispatch's usage; every other batch routes to window inference.
func dispatchAgentIDForPath(path string) (string, bool) {
	if id, ok := AgentIDFromTranscriptPath(path); ok {
		return id, true
	}
	return AgentIDFromRolloutPath(path)
}

// attributeDispatches runs the second attribution pass over one batch's
// records and merges each touched dispatch's delta into deps. Every
// Watcher runs this pass unconditionally (KAN-173): dispatchAttributor is
// built inside NewWatcher from deps, which also carries the sink this
// merges into, so there is no longer a configured/unconfigured state to
// branch on the way WithDispatchAttribution's absence used to skip this
// entirely.
//
// A batch read from a per-agent transcript (path under a subagents/
// directory, IsAgentTranscriptPath) never reaches the inference pass at
// all: that file IS one agent's usage, so attributeAgentFile credits the
// dispatch rows the file's own agentId resolves to, directly
// (attributeAgentFileRecords). Routing by source is what kan-357 exists
// to do -- the inference pass is what stamped every resumed agent's rows
// unattributed, because its identity pass matches several rows carrying
// the same agentId and its interval pass refuses to guess between them.
// A ZCode subagent dispatch's rollout file
// (model-io-sess_subagent_agent_<id>.jsonl, AgentIDFromRolloutPath) is the
// same shape of fact and routes the same way (KAN-506): one file IS one
// dispatch's usage, keyed on that dispatch's own per-dispatch session id,
// so concurrent dispatches neither lose nor blend their cost figures.
// Only batches from top-level session transcripts -- embedded sidechains,
// the shape Cursor and Codex and pre-agent-file Claude transcripts
// produce -- still go through DispatchAttributor.
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
// Where a record's agent id or timestamp matched more than one dispatch --
// the ambiguity this pass once wrote off by stamping every candidate
// unattributed -- Attribute now apportions its usage across the candidates
// pro-rata by duration (KAN-414, apportionRecord), and the apportioned
// shares merge through MergeDispatchMetrics beside the outright ones, each
// touched candidate's patch carrying the concurrency-attribution record
// (MetricsPatch.Apportioned) that says part of its figure was shared. The
// "do not stamp a dispatch that attributed" rule (task 6, tasks.md) has no
// work left to do: nothing is stamped, so nothing attributed can be
// contradicted by a stamp.
func (w *Watcher) attributeDispatches(ctx context.Context, records []Record, path string) {
	if agentID, ok := dispatchAgentIDForPath(path); ok {
		w.attributeAgentFile(ctx, records, path, agentID)
		return
	}

	deltas, err := w.dispatchAttributor.Attribute(ctx, records)
	if err != nil {
		w.warn("harvest: attribute dispatch windows failed, this batch's dispatch figures are lost", "path", path, "error", err)
		return
	}

	for dispatchID, dd := range deltas {
		mp := MetricsPatch{Tokens: dd.Tokens, Signals: &SignalsDelta{Sidechain: dd.Signals}}
		if dd.ApportionedRecords > 0 {
			mp.Apportioned = &ApportionedDelta{Records: dd.ApportionedRecords}
		}
		patch, err := json.Marshal(mp)
		if err != nil {
			w.warn("harvest: encode dispatch metrics failed", "path", path, "dispatch_id", dispatchID, "error", err)
			continue
		}
		if err := w.deps.MergeDispatchMetrics(ctx, dispatchID, patch); err != nil {
			w.warn("harvest: merge dispatch metrics failed, this batch's figures for it are lost", "path", path, "dispatch_id", dispatchID, "error", err)
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
// attributeAgentFile credits one agent-file batch's records to the
// dispatch rows agentID resolves to -- kan-357's "record each dispatch's
// own usage at the source", routed here by attributeDispatches for every
// batch dispatchAgentIDForPath names. The windows come from deps'
// DispatchWindowsForAgent, ordered by (started_at, id), and the per-row
// sums merge through MergeDispatchMetrics exactly as the inference
// path's do: additive batch deltas, keyed by the dispatch row's own id.
//
// An agent whose id matches no dispatch row credits nothing and says
// nothing -- an agent the dispatch protocol never recorded (a fork's
// subagents, today) has no row to touch, and the stage grain already
// counts its tokens. The failure posture is attributeDispatches' own:
// log, step over, never return -- the stage pass has already committed
// this batch by the time anything here can fail.
func (w *Watcher) attributeAgentFile(ctx context.Context, records []Record, path string, agentID string) {
	windows, err := w.deps.DispatchWindowsForAgent(ctx, agentID)
	if err != nil {
		w.warn("harvest: dispatch windows for agent failed, this batch's dispatch figures are lost", "path", path, "agent_id", agentID, "error", err)
		return
	}

	for dispatchID, dd := range attributeAgentFileRecords(windows, records) {
		patch, err := json.Marshal(MetricsPatch{Tokens: dd.Tokens, Signals: &SignalsDelta{Sidechain: dd.Signals}})
		if err != nil {
			w.warn("harvest: encode dispatch metrics failed", "path", path, "dispatch_id", dispatchID, "error", err)
			continue
		}
		if err := w.deps.MergeDispatchMetrics(ctx, dispatchID, patch); err != nil {
			w.warn("harvest: merge dispatch metrics failed, this batch's figures for it are lost", "path", path, "dispatch_id", dispatchID, "error", err)
		}
	}
}

// dispatchBeginEvent is the pairing-relevant slice of a
// `flow record dispatch begin` command found in a transcript: the row the
// launch's agent id will be stamped onto is named by (sessionToken, key),
// the same two literals the begin call itself carries, and StartedAt is
// the command line's own transcript instant -- when the dispatcher issued
// the begin, the same write the daemon stamps the row's started_at at
// (KAN-324) -- which is what pairs it with a launch across either
// ordering without any window inference standing in between. A begin
// recorded immediately before its launch sits seconds from the launch's
// tool result; a panel round's begins, recorded after the round's
// launches, sit seconds behind them -- both inside the tolerance, both
// anchored by measurement rather than by a caller's typed claim.
type dispatchBeginEvent struct {
	SessionToken string
	Key          string
	StartedAt    time.Time
}

// maxStampTimeTolerance is how far a begin's own -started-at may sit from
// a launch's timestamp and still pair with it. Both instants describe the
// same launch: the harness writes the tool result when the launch lands,
// the dispatcher types -started-at as that same launch's start. A panel
// round records its begins right after the launches return -- seconds; a
// single dispatch records its begin immediately before -- again seconds.
// What the tolerance exists to EXCLUDE is a stale begin: one whose launch
// never happened (a denied dispatch) or whose launch paired already,
// sitting minutes away from the launch now being paired. Rounds run for
// minutes, so two minutes separates "this launch" from "some earlier
// round's begin" with room on both sides.
const maxStampTimeTolerance = 2 * time.Minute

// maxPendingStampEvents bounds both pending lists, per path. A dispatch
// round is a handful of events; eight absorbs any real round plus the
// cross-batch straddle the pending lists exist for, while a transcript
// that somehow accumulates begin-less launches (a caller recording begins
// the daemon cannot see, or never recording them at all) cannot grow the
// Watcher's state without end.
const maxPendingStampEvents = 8

// beginCommandShape is what a command must contain to pair with a launch:
// the begin verb itself plus the two flags that identify its row. The
// row's time anchor is the command line's own transcript instant
// (CommandRecord.Timestamp) -- the same write the daemon stamps the row's
// started_at at (KAN-324) -- never a caller-typed flag, because there is
// no longer one: KAN-324 removed the instants from the caller's hands. It
// deliberately re-encodes the CLI's own required-flag contract instead of
// importing it: the harvester parses transcript TEXT the CLI never sees,
// and internal/harvest imports nothing from cmd/ (the dependency would run
// the wrong way), so the two lists are kept in step by this note. A
// diagnostic that merely mentions one of them (a grep quoting a token, or
// a mention of the key without the rest) fails the shape and pairs with
// nothing -- the mention-versus-invocation distinction
// matchSessionTokens already draws, applied to the same class of
// transcript text.
var beginCommandShape = []string{
	"flow record dispatch begin",
	"-key",
	"-session-token",
}

// dispatchBeginsFromCommand judges one Bash command's text as zero or
// more dispatch begins and extracts each one's row identity and start
// instant. Line continuations and `&&` joins are folded first, because
// the skills' templates wrap one invocation across several physical lines
// and a panel round records every slot's begin in one Bash call; what
// remains is one candidate invocation per line. Every shape literal must
// be present, and each flag must carry a value -- the CLI itself refuses
// a begin without them, so a fragment missing one is a mention or a
// mistype, never a row opener.
func dispatchBeginsFromCommand(command string) []dispatchBeginEvent {
	joined := strings.ReplaceAll(command, "\\\n", " ")
	fragments := strings.Split(joined, "\n")
	var out []dispatchBeginEvent
	for _, fragment := range fragments {
		for _, piece := range strings.Split(fragment, "&&") {
			ev, ok := dispatchBeginFromInvocation(piece)
			if ok {
				out = append(out, ev)
			}
		}
	}
	return out
}

// dispatchBeginFromCommand judges one invocation's text as a possible
// dispatch begin. Renamed from its singular predecessor's role: it sees
// one candidate invocation, never the whole Bash command.
func dispatchBeginFromInvocation(invocation string) (dispatchBeginEvent, bool) {
	for _, part := range beginCommandShape {
		if !strings.Contains(invocation, part) {
			return dispatchBeginEvent{}, false
		}
	}
	token, ok := commandFlagValue(invocation, "-session-token")
	if !ok {
		return dispatchBeginEvent{}, false
	}
	key, ok := commandFlagValue(invocation, "-key")
	if !ok {
		return dispatchBeginEvent{}, false
	}
	return dispatchBeginEvent{SessionToken: token, Key: key}, true
}

// commandFlagValue reads one flag's value out of an invocation's text:
// the next whitespace-delimited token after the flag, with one layer of
// surrounding single or double quotes stripped. The flag must begin at a
// word boundary -- preceded by whitespace or the start of the text -- so
// a `-change add-keyboard-shortcuts` slug containing the substring `-key`
// is never read as a -key flag. The values a dispatcher types are
// validated literals (never a shell substitution), so this is
// deliberately not a shell parser -- a shape it cannot read yields no
// value, and no value stamps nothing.
func commandFlagValue(command, flag string) (string, bool) {
	offset := 0
	idx := -1
	for {
		next := strings.Index(command[offset:], flag)
		if next < 0 {
			return "", false
		}
		idx = next + offset
		if idx == 0 || command[idx-1] == ' ' || command[idx-1] == '\t' {
			break
		}
		offset = idx + 1
	}
	rest := strings.TrimLeft(command[idx+len(flag):], "	 ")
	if rest == "" {
		return "", false
	}
	end := strings.IndexAny(rest, " 	")
	if end < 0 {
		end = len(rest)
	}
	value := rest[:end]
	value = strings.Trim(value, `'"`)
	if value == "" {
		return "", false
	}
	return value, true
}

// stampDispatchAgents pairs KAN-322's agent launches with the dispatch
// begin commands that name their rows, and stamps each pair's agent id
// onto that row via deps' DispatchAgentStamper.
//
// PAIRING IS BY TIME, NOT BY LINE POSITION, and the reason is the two
// orderings the dispatch protocol itself produces. A single dispatch
// records its begin immediately before the launch (implement.md), so the
// begin sits at a smaller line; a panel round launches its slots in one
// message and records their begins in one Bash call afterwards
// (review-panel.md), so the begins sit at larger lines -- and with more
// than one launch in the round, line order would make the first launch
// steal whichever begin precedes it, attributing one dispatch's identity
// to another row. What both orderings hold constant is the TIME: each
// begin's own -started-at names the launch it belongs to, and a launch's
// tool result carries its own timestamp. A begin pairs with the unpaired
// launch nearest in time, within maxStampTimeTolerance; a begin with no
// launch in reach, or a launch with no begin, stays pending for the next
// batch -- the pair may simply straddle a batch boundary -- and is
// eventually evicted by the bound. Nothing is ever stamped by inference
// from line order or row order alone.
//
// The stamp fills only an empty agent_id (store.StampDispatchAgent's own
// contract), so a hand-typed id always wins and a replayed or duplicated
// stamp cannot corrupt a row that already carries one.
//
// The failure posture is attributeDispatches' own: log, step over, never
// return. A stamp that fails loses that one row's automatic id -- the
// row stays empty, ordinary on the harnesses that expose no id at all --
// and every other row is unaffected.
func (w *Watcher) stampDispatchAgents(ctx context.Context, path string, commands []CommandRecord, launches []AgentLaunch, denials []time.Time) {
	if len(commands) == 0 && len(launches) == 0 && len(denials) == 0 {
		return
	}

	var begins []dispatchBeginEvent
	for _, c := range commands {
		for _, ev := range dispatchBeginsFromCommand(c.Command) {
			// The command line's own instant is the begin's time anchor; a
			// command whose line carries none can never pair.
			if !c.Timestamp.IsZero() {
				ev.StartedAt = c.Timestamp
				begins = append(begins, ev)
			}
		}
	}
	begins = append(w.pendingBegins[path], begins...)
	launches = append(w.pendingLaunches[path], launches...)
	denials = append(w.pendingDenials[path], denials...)

	// Every (begin, launch) pair within tolerance is considered; a pair
	// is taken only while both sides are unclaimed, in the sorted order
	// below. The denial check at claim time is what makes a stale begin
	// lose: a denied dispatch's begin sits seconds from the launch that
	// eventually follows, but the launch's own begin sits closer, and the
	// closer pair is decided first.
	type pairing struct {
		beginIdx, launchIdx int
		delta               time.Duration
	}
	var candidates []pairing
	for bi, begin := range begins {
		for li, launch := range launches {
			delta := launch.Timestamp.Sub(begin.StartedAt)
			if delta < 0 {
				delta = -delta
			}
			if delta <= maxStampTimeTolerance {
				candidates = append(candidates, pairing{bi, li, delta})
			}
		}
	}
	// Candidates sort by LAUNCH time first: a panel round records every
	// begin in one Bash call, so those begins share one transcript instant
	// and only their order against the launches' own order separates them
	// -- the round's k-th launch belongs to the round's k-th begin, both
	// emitted in slot order. Within one launch, the nearest begin wins
	// (delta, then begin start, then key), which is what keeps a stale
	// begin from out-ranking the launch's own begin. Deterministic under
	// any input order.
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !launches[a.launchIdx].Timestamp.Equal(launches[b.launchIdx].Timestamp) {
			return launches[a.launchIdx].Timestamp.Before(launches[b.launchIdx].Timestamp)
		}
		if a.delta != b.delta {
			return a.delta < b.delta
		}
		if !begins[a.beginIdx].StartedAt.Equal(begins[b.beginIdx].StartedAt) {
			return begins[a.beginIdx].StartedAt.Before(begins[b.beginIdx].StartedAt)
		}
		if begins[a.beginIdx].Key != begins[b.beginIdx].Key {
			return begins[a.beginIdx].Key < begins[b.beginIdx].Key
		}
		return a.launchIdx < b.launchIdx
	})

	// A denial is evidence that a begin's dispatch never launched. A begin
	// may stamp only when its launch is STRICTLY nearer than every denial
	// within tolerance: where a denial sits as near or nearer, the begin
	// may be the denied attempt's own (a panel round launches its slots
	// together and records their begins together, so the two are
	// indistinguishable by time), and stamping would attribute one
	// dispatch's identity to another row. Such a begin is retired -- its
	// decision is final, because every later launch is farther still -- and
	// the row keeps no id, the ordinary silence. The retirement is
	// therefore per-begin and evidence-bound, never a blanket window: an
	// innocent concurrent begin whose launch is strictly nearer still
	// stamps.
	takenBegin := make([]bool, len(begins))
	claimed := make([]bool, len(launches))
	for _, c := range candidates {
		if takenBegin[c.beginIdx] || claimed[c.launchIdx] {
			continue
		}
		begin := begins[c.beginIdx]
		ambiguous := false
		for _, at := range denials {
			delta := at.Sub(begin.StartedAt)
			if delta < 0 {
				delta = -delta
			}
			if delta < c.delta {
				ambiguous = true
				break
			}
		}
		if ambiguous {
			takenBegin[c.beginIdx] = true
			continue
		}
		takenBegin[c.beginIdx] = true
		claimed[c.launchIdx] = true
		stamped, err := w.deps.StampDispatchAgent(ctx, begin.SessionToken, begin.Key, launches[c.launchIdx].AgentID)
		if err != nil {
			w.warn("harvest: stamp dispatch agent failed, the row keeps no id from this launch",
				"path", path, "key", begin.Key, "agent_id", launches[c.launchIdx].AgentID, "error", err)
			continue
		}
		if !stamped {
			// No empty row under (session-token, key): the begin fell back
			// to a journal not yet replayed, the row carries a hand-typed
			// id already, or the key names a row this transcript cannot
			// see. All ordinary; nothing to stamp, nothing to log.
			continue
		}
	}

	// Whatever remains unpaired or unclaimed REPLACES the pending lists --
	// the merged lists already contain the carried events, so appending
	// survivors onto them would double-store every event that survives two
	// passes and let a stale begin claim a later launch. Bounded: oldest
	// dropped, so neither list can grow without end. A dropped event never
	// stamps -- the ordinary silence, never a wrong pairing.
	w.pendingBegins[path] = appendPending(nil, survivors(begins, takenBegin))
	w.pendingLaunches[path] = appendPending(nil, unclaimed(launches, claimed))
	w.pendingDenials[path] = appendPending(w.pendingDenials[path], denials)
}

// survivors returns the begins no pass consumed -- neither paired with a
// launch nor retired as ambiguous against a denial. Their launches may
// still arrive in a later batch.
func survivors(begins []dispatchBeginEvent, taken []bool) []dispatchBeginEvent {
	var out []dispatchBeginEvent
	for i, begin := range begins {
		if !taken[i] {
			out = append(out, begin)
		}
	}
	return out
}

// unclaimed returns the launches no begin took this pass.
func unclaimed(launches []AgentLaunch, claimed []bool) []AgentLaunch {
	var out []AgentLaunch
	for i, launch := range launches {
		if !claimed[i] {
			out = append(out, launch)
		}
	}
	return out
}

// appendPending appends new pending events behind the old and enforces
// the bound, dropping from the front.
func appendPending[T any](pending, added []T) []T {
	pending = append(pending, added...)
	if len(pending) > maxPendingStampEvents {
		pending = pending[len(pending)-maxPendingStampEvents:]
	}
	return pending
}

// maxOpenAgentCalls bounds one path's open agent-call set. A dispatch
// round is a handful of calls; the cap absorbs any real round plus a
// straddled one, and a set that somehow outgrows it is cleared wholesale
// rather than grown without end.
const maxOpenAgentCalls = 64

// rememberOpenAgentCalls carries one batch's still-unresolved agent
// tool_use ids into the next read, so a denial landing in a later batch
// can still be joined to its tool call. The set replaces the previous
// one -- the batch's parser already merged what was carried in -- and is
// bounded by wholesale clearing, the lossy-is-safe trade the field's own
// doc comment states.
func (w *Watcher) rememberOpenAgentCalls(path string, ids []string) {
	if len(ids) == 0 {
		delete(w.openAgentCalls, path)
		return
	}
	if len(ids) > maxOpenAgentCalls {
		delete(w.openAgentCalls, path)
		return
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	w.openAgentCalls[path] = set
}

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
	for stageRunID := range patches {
		if err := w.deps.Price(ctx, stageRunID); err != nil {
			w.warn("harvest: price stage run failed, will retry next cycle", "stage_run_id", stageRunID, "error", err)
		}
	}
}

// pendingSessionTokens returns this cycle's stage-run-id -> session-token
// map to search for: whatever w.deps.UnresolvedSessionTokens reports,
// minus any run whose token this Watcher has already given up on
// (gaveUpTokens, keyed by token -- task 4b) -- so a token that has already
// been logged as abandoned is never looked for again by this process,
// which is the whole point of tracking gaveUpTokens at all. A Watcher
// built with harvest.NoDeps (KAN-173) always gets an empty map back --
// UnresolvedSessionTokens' own no-op -- so every other token-related step
// below is a no-op over it, resolving session tokens remaining additive
// exactly as it was when a nil binder meant "not configured".
func (w *Watcher) pendingSessionTokens(ctx context.Context) map[int64]string {
	w.seedPersistedGiveUps(ctx)
	all, err := w.deps.UnresolvedSessionTokens(ctx)
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
	giveUps, err := w.deps.PersistedGiveUps(ctx)
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
// own usage string (`flow stage begin ...` / `flow stage end ...`),
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
//     valid to the flag package cmd/flow/stage.go builds on.
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
// base name is the CLI's -- to reject a command that only PRINTS a
// mark-shaped string, e.g. `echo "flow stage begin ...
// -session-token mf-abc123 ..."`. KAN-174 removes that requirement: real
// marks are routinely emitted after variable assignments
// (`N=kan; T=mf-x; cd /repo`) and on a later line of a multi-statement
// block, not just behind a single leading `cd ... &&`, and every one of
// those shapes was measured rejected by the F5 anchor in production --
// all 14 stage runs from the finish sequence that merged F5 went unbound
// (design.md). A position anchor loose enough to admit those shapes (e.g.
// requiring the CLI name only after a command boundary -- start of text, or
// after `;`, `&&`, `||`, `|`, a newline) was considered and rejected: the
// echoed-example text it is meant to exclude still only needs a newline
// before the CLI name to satisfy a boundary anchor too, which is common in
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
			if err := w.deps.RecordSessionTokenGiveUp(ctx, sessionToken, reasonSessionNeverBound, time.Now()); err != nil {
				w.warn("harvest: persist give-up failed, will retry", "stage_run_ids", stageRunIDs, "error", err)
				continue
			}
			// A stamp failure here must never block the give-up already
			// recorded above (task 6.1, tasks.md, step 4): it is logged
			// and this branch still proceeds to mark the token given up.
			if err := w.deps.MarkDispatchesUnattributed(ctx, sessionToken, reasonSessionNeverBound, 0); err != nil {
				w.warn("harvest: stamp unattributed dispatches failed", "stage_run_ids", stageRunIDs, "session_token", sessionToken, "error", err)
			}
			w.gaveUpTokens[sessionToken] = true
			delete(w.tokenCycles, sessionToken)

		case 1:
			var sessionID string
			for s := range sessions {
				sessionID = s
			}
			bound, err := w.deps.BindSession(ctx, sessionToken, sessionID)
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
			if err := w.deps.RecordSessionTokenGiveUp(ctx, sessionToken, reasonSessionAmbiguous, time.Now()); err != nil {
				w.warn("harvest: persist give-up failed, will retry", "stage_run_ids", stageRunIDs, "error", err)
				continue
			}
			// Same never-block guarantee as the case 0 branch above: a
			// stamp failure here is logged, never a reason to undo the
			// give-up already recorded.
			if err := w.deps.MarkDispatchesUnattributed(ctx, sessionToken, reasonSessionAmbiguous, len(sessions)); err != nil {
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
		sig := delta.Signals
		mp := MetricsPatch{Tokens: delta.Total, Speed: delta.Speed, Signals: &sig}
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
				if s, ok := delta.DispatchSignals[agentID]; ok {
					s := s
					db.Signals = &s
				}
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
