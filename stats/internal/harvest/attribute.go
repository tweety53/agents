package harvest

import (
	"context"
	"fmt"
	"time"
)

// Window is one stage run's identity and attribution interval -- exactly
// as much as attribution needs to know, and nothing about metrics
// already recorded or the owning change's other fields. A message
// belongs to a Window when its SessionID matches and its Timestamp falls
// in the half-open interval [StartedAt, EndedAt) -- EndedAt nil means
// "still open", per design.md's "Harvesting" section.
//
// The interval is deliberately half-open, not closed on both ends: a
// message timestamped exactly at one window's EndedAt, which also equals
// the next window's StartedAt (the ordinary case for two back-to-back
// stages with no gap), must resolve to exactly one window, not both.
// Closed-on-both-ends would let it match either, with the tie falling to
// whatever order WindowsForSession happened to return -- an incidental
// SQL ORDER BY, never a stated rule (task 9's post-commit review, finding
// F4). Half-open assigns that boundary message to the window that is
// *starting*, which matches the convention design.md's statistics API
// already commits to for a different boundary (period edges): "a stage
// run straddling a period boundary is attributed by its start instant".
// Treating a window's own start instant as the one that claims a tie is
// the same rule applied one level down, so the two conventions cannot
// disagree with each other once the statistics API inherits this one.
type Window struct {
	StageRunID int64
	// Attempt breaks a tie when a message's timestamp falls inside more
	// than one *open* window for the same session simultaneously -- see
	// bestWindow's own doc comment for when that can happen (design.md,
	// "A replayed begin mark can open a second attempt"). This is a
	// different situation from the boundary case above: two genuinely
	// concurrent open windows, not two adjacent closed ones, so the
	// half-open interval does not resolve it by itself.
	Attempt   int
	SessionID string
	StartedAt time.Time
	EndedAt   *time.Time
}

// contains reports whether ts falls inside w's half-open attribution
// interval [StartedAt, EndedAt) -- see Window's own doc comment for why
// the end is exclusive.
func (w Window) contains(ts time.Time) bool {
	if ts.Before(w.StartedAt) {
		return false
	}
	if w.EndedAt != nil && !ts.Before(*w.EndedAt) {
		return false
	}
	return true
}

// WindowSource answers which stage windows exist for a session -- both
// still-open ones (EndedAt nil) and ones that have since closed, because
// a harvest batch can lag behind an end mark and a message's timestamp
// can still fall inside a window that closed before this batch ran.
//
// This is a deliberate choice, not an oversight, for task 10's sweeper
// too: a stage run the sweeper has already closed with outcome
// "abandoned" is exactly as eligible a target for a late-arriving harvest
// batch as one closed by a genuine `stage end`, because closing either
// way sets the same EndedAt this interface's contains() check reads --
// SweepAbandoned (internal/store) sets ended_at to the sweep's own
// instant, so any message truly written inside [StartedAt, ended_at)
// really was spent during that stage's window regardless of how the
// window came to close. Excluding abandoned runs here would mean a
// session that crashed after doing real, chargeable work loses that
// spend from the record entirely -- worse than recording it against a
// stage whose outcome says "abandoned", which is the honest fact anyway.
// A message arriving after the sweep's own EndedAt is not attributed
// here or anywhere else, exactly like any other message that misses
// every window: this rule extends no window past what it already covers.
//
// Defined here, at the consumer, per go-interface-design: internal/harvest
// never imports internal/store, so this package -- and
// TestHarvestNeedsNoDatabase in particular -- is testable with nothing
// but a fake satisfying this one method, no PostgreSQL required. The
// daemon wires a real implementation backed by store.Store.QueryStageRuns
// (cmd/myflowd/main.go).
type WindowSource interface {
	WindowsForSession(ctx context.Context, sessionID string) ([]Window, error)
}

// Bucket is one accumulation of token figures -- design.md's metrics bag
// table, minus the keys that are not derived from usage (cost_usd,
// duration_ms, and so on, which other writers own).
//
// CacheCreation keeps its original meaning and its name: the collapsed
// total, unchanged from before task 23, that task 24's aggregations
// already read. CacheCreation5m and CacheCreation1h are the same total
// split by which rate actually applied (task 23) -- for any message whose
// usage.cache_creation object was present (Usage.CacheSplitKnown), its
// contribution lands in exactly one of the two, so
// CacheCreation == CacheCreation5m + CacheCreation1h + CacheCreationUnknown
// always holds, and CacheCreation == CacheCreation5m + CacheCreation1h
// holds whenever CacheCreationUnknown is zero -- every message in this
// bucket carried a known split. CacheCreationUnknown accumulates the
// collapsed total from messages whose split was *not* known: real,
// measured spend that store.Store.Price must still record but must never
// price, because guessing which of the two rates applied is a 37.5%-to-60%
// error in whichever direction is wrong (task 23's own plan-provenance
// note). All three cache-creation fields omit themselves from the JSON
// encoding when zero (jsonb_deep_add, 0005_jsonb_deep_add.sql, sums
// numeric leaves at any depth -- an absent key contributes nothing to that
// sum, exactly like a zero would, so omitting it changes no total; it
// only keeps a batch that saw no unknown-split usage from ever writing a
// "cache_creation_unknown" key at all).
type Bucket struct {
	Input                int64 `json:"input"`
	Output               int64 `json:"output"`
	CacheCreation        int64 `json:"cache_creation"`
	CacheCreation5m      int64 `json:"cache_creation_5m,omitempty"`
	CacheCreation1h      int64 `json:"cache_creation_1h,omitempty"`
	CacheCreationUnknown int64 `json:"cache_creation_unknown,omitempty"`
	CacheRead            int64 `json:"cache_read"`
	Thinking             int64 `json:"thinking"`
}

func (b *Bucket) add(u Usage) {
	b.Input += u.InputTokens
	b.Output += u.OutputTokens
	b.CacheCreation += u.CacheCreationInputTokens
	if u.CacheSplitKnown {
		b.CacheCreation5m += u.CacheCreation5mTokens
		b.CacheCreation1h += u.CacheCreation1hTokens
	} else {
		b.CacheCreationUnknown += u.CacheCreationInputTokens
	}
	b.CacheRead += u.CacheReadInputTokens
	b.Thinking += u.ThinkingTokens
}

// TokenDelta is the usage found in *one batch* of newly read records,
// split into main-thread and sidechain buckets -- design.md's
// "tokens.main"/"tokens.sidechain" split, so a dispatching command's own
// cost is distinguishable from what it dispatched.
//
// TokenDelta is always just one batch's contribution, never a running
// total: Watcher hands it to store.Store.CommitHarvestBatch (via the
// HarvestSink interface, watcher.go), which adds it onto whatever
// Postgres already holds for that stage run's metrics -- in the same
// transaction that advances the transcript's consumed byte offset. This
// package itself never needs to know, remember, or recompute a
// cumulative total: Postgres is the only place one exists, and the
// offset advancing alongside it is what makes replaying a batch, or not
// replaying it, always the correct choice (see watcher.go's own doc
// comment on CommitHarvestBatch's atomicity -- task 9's post-commit
// review, findings F1 and its follow-up).
type TokenDelta struct {
	Main      Bucket `json:"main"`
	Sidechain Bucket `json:"sidechain"`
}

func (t *TokenDelta) bucket(sidechain bool) *Bucket {
	if sidechain {
		return &t.Sidechain
	}
	return &t.Main
}

// upsertBucket looks up the current value for key in *m -- lazily
// allocating the map first if *m is nil, exactly like a plain map
// assignment to a nil map would panic without this check -- passes it to
// combine, and stores combine's result back under key. combine receives
// V's zero value when key was not yet present in the map.
//
// Used by Attribute's per-model and per-dispatch token accumulation
// (below), where combine is a genuine accumulator (bucket().add() onto
// whatever TokenDelta was already there for this stage run's batch so
// far) -- nil-check, lazy make, get-or-zero, mutate, store back, done
// once instead of twice, with each call site's own key and mutation left
// entirely to it: this helper merges no two maps together and changes
// what neither of them records.
//
// encodePatches (watcher.go) used to call this too (F24, this change's
// own review panel), for its per-model and per-dispatch metrics-patch
// construction -- but that combine never actually read the zero value it
// was handed: it built a fresh ModelBucket or DispatchBucket from
// scratch every time, since its destination map starts nil on every
// call. That made the code read as accumulation it never performed, so
// F35 (pass 7 of this change's own review panel) replaced those two call
// sites with a plain nil-check-and-assign instead of this helper --
// upsertBucket now serves only the two sites that genuinely accumulate.
func upsertBucket[K comparable, V any](m *map[K]V, key K, combine func(V) V) {
	if *m == nil {
		*m = make(map[K]V)
	}
	(*m)[key] = combine((*m)[key])
}

// Delta is what Attribute computes for one touched stage run: Total is
// the whole-run token delta -- unchanged in meaning from what this type
// used to be on its own, and still the sole source of the metrics bag's
// top-level "tokens" key, which every existing aggregation reads. Models
// is the same messages broken out a second way, by the model each
// message's own transcript line carried, keyed by that model string and
// feeding the metrics bag's "models.<model>.tokens" key (task 22).
//
// A record whose message carried no model contributes to Total but to
// no entry in Models -- never to a fabricated "unknown" or "" key. Doing
// otherwise would let unmeasured usage masquerade as a real model choice
// the moment it reached a UI dropdown, which is the absence-is-never-zero
// rule broken at its source (MetricsPatch's own doc comment says the
// same thing from the encoding side).
//
// Models is always a subset of what Total accounts for: because both are
// filled from the same records in the same pass (Attribute's loop below),
// summing every entry in Models reproduces Total exactly for a run whose
// every record carried a model, and falls short of it by exactly the
// no-model records' contribution otherwise -- never more, never
// mismatched.
type Delta struct {
	Total  TokenDelta
	Models map[string]TokenDelta
	// Dispatches is the same messages broken out a third way, by the
	// agentId each subagent transcript line carries (KAN-201), keyed by
	// that agentId and feeding the metrics bag's "dispatches.<agentId>"
	// key (encodePatches, watcher.go, adds that dispatch's descriptors
	// alongside these token figures -- Attribute itself sees only
	// records, never a transcript path, so it has no descriptors to add).
	//
	// This is Models' own rule (this doc comment, above) applied to a
	// second key: a record whose message carried no agentId -- every
	// parent-session message -- contributes to Total and, when it carries
	// a model, to Models, but to no entry here, never a fabricated ""
	// key. Dispatches is always a subset of what Total accounts for,
	// filled from the same records in the same pass as Total and Models.
	Dispatches map[string]TokenDelta
	// Speed is the last non-empty Usage.Speed seen among this stage run's
	// records in this batch -- last-write-wins, the same rule Effort
	// already documents for MetricsPatch, and correct for the same
	// reason: a stage run's speed setting does not change message to
	// message in the ordinary case, so the latest observation is as good
	// a record of it as any, and never worse than leaving it unset. Empty
	// when no record in this batch carried a speed at all.
	Speed string
}

// ModelBucket is one model's contribution to a stage run's metrics bag,
// nested under "models.<model>" -- currently just the token totals that
// model's own messages carried, mirroring MetricsPatch's own "tokens"
// key one level down (task 22, Delta's own doc comment explains why the
// per-model split exists at all). store.Store.Price (internal/store,
// pricing.go) later adds a "cost_usd" sibling to this same object; the
// harvester itself never writes cost.
type ModelBucket struct {
	Tokens TokenDelta `json:"tokens"`
}

// DispatchBucket is one dispatch's contribution to a stage run's metrics
// bag, nested under "dispatches.<agentId>" -- ModelBucket's own shape
// (above) extended with that dispatch's own descriptors (KAN-201), read
// from its transcript's sibling meta sidecar by ReadDispatchMeta
// (transcript.go) and threaded through by encodePatches (watcher.go),
// which alone knows which file a batch of records came from; Attribute
// never sees a transcript path, only records, so it cannot fill these in
// itself.
//
// AgentType, Description, Model and SpawnDepth all omit themselves from
// the JSON encoding, rather than writing an empty string or an absent
// depth, when no sidecar was found for this dispatch's transcript --
// the same absence-is-never-a-value rule DispatchMeta's own doc comment
// states, applied one level down at the metrics-bag encoding. All four
// descriptors resolve last-write-wins under jsonb_deep_add
// (0005_jsonb_deep_add.sql) while Tokens' nested leaves sum -- exactly
// what a dispatch whose token figures grow across batches, but whose
// descriptors never change once its sidecar is written, needs. That
// requires every descriptor to reach jsonb_deep_add as something other
// than a JSON *number*: the function's own rule is "sum two numbers at
// the same key, else b replaces a" (0005_jsonb_deep_add.sql:61-62), so a
// numeric leaf never resolves last-write-wins at all -- it accumulates.
// AgentType, Description and Model are strings on the wire already, so
// this holds for them for free; SpawnDepth needs its own encoding to get
// there (below).
//
// SpawnDepth is a *string in Go, not a *int (F23, this change's own
// review panel: a *int forced the dispatchBucketWire shadow struct plus
// hand-written MarshalJSON/UnmarshalJSON this file used to carry, purely
// to encode one field as a string, with no production reader anywhere in
// this repository that ever needed it back as an int -- every call site
// that sets it already holds a plain int locally, from ReadDispatchMeta's
// DispatchMeta.SpawnDepth, and converts with strconv.Itoa at the point of
// construction). A pointer, not a plain string, because a genuine
// top-level dispatch's sidecar carries "spawnDepth":0, a real, measured
// fact -- indistinguishable from "no sidecar was found" if this field's
// zero value and its absent value collapsed together, which `omitempty`
// on a plain string cannot prevent (Go's encoding/json omits an empty
// string exactly as readily as an unset one). A pointer's own zero value
// is nil, distinct from a pointer to "0", so `omitempty` on a *string
// omits only genuine absence, never a present-and-zero depth. AgentType,
// Description and Model do not share this hazard: none of them has a
// legitimate empty-string value coming out of ReadDispatchMeta
// (transcript.go) -- an agent type, a description and a model name are
// never recorded as "" by a sidecar that exists, so an empty string
// there really does mean "no sidecar", exactly as omitempty already
// treats it.
//
// SpawnDepth is a JSON *string* ("0", "1", ...) rather than a JSON
// number (F12, pass 2 of this change's own review panel) precisely so
// encoding/json's ordinary struct handling suffices, with no custom
// Marshal/Unmarshal needed: a dispatch harvested across N ordinary
// cycles has encodePatches (watcher.go) re-send this same constant value
// on every one of them (its own "if hasMeta" branch runs unconditionally,
// every commit, not once), and a plain JSON number at that key would
// make jsonb_deep_add sum it N times over instead of resolving
// last-write-wins like every other descriptor -- silent, unbounded, and
// specific to this one field. jsonb_deep_add itself is deliberately
// unchanged -- summing numeric leaves is correct and load-bearing for
// every token figure this same function merges. Decoding a corrupted
// spawn_depth (an unquoted number where a string is expected) is still
// rejected, not silently accepted: encoding/json itself refuses to
// unmarshal a JSON number into a Go string field, so that guarantee
// costs this type nothing extra to keep.
type DispatchBucket struct {
	Tokens      TokenDelta `json:"tokens"`
	AgentType   string     `json:"agent_type,omitempty"`
	Description string     `json:"description,omitempty"`
	Model       string     `json:"model,omitempty"`
	SpawnDepth  *string    `json:"spawn_depth,omitempty"`
}

// MetricsPatch is the JSON shape a harvest batch's results are added into
// a stage run's metrics bag as, via store.Store.CommitHarvestBatch's
// jsonb_deep_add. Tokens is the whole-run total (unchanged meaning from
// before task 22: every existing aggregation over the metrics bag's
// top-level "tokens" key keeps working); Models is the same batch's
// records broken out by model, additive per model exactly like Tokens is
// additive overall, so two batches touching the same run and the same
// model sum rather than replace with no merge code and no migration
// (0005_jsonb_deep_add.sql sums numeric leaves at any depth). Dispatches
// is the same batch's records broken out a third way, by agentId
// (KAN-201) -- additive per dispatch's Tokens exactly like Models is
// additive per model, with that dispatch's descriptors resolved
// last-write-wins alongside it (DispatchBucket's own doc comment).
// Effort is last-write-wins, like every other non-numeric leaf under
// jsonb_deep_add: the harvester simply records whichever value the most
// recently processed message in this batch carried. There is no scalar
// Model field any more -- a stage run's model is now entirely a property
// of which keys Models carries, never a single last-write-wins string,
// which is what made a mixed-model run (task 22's whole reason for
// existing) impossible to represent correctly in the first place.
//
// Exported (unlike its predecessor, the unexported metricsPatch) because
// Watcher, not Attributor, is now the thing that marshals it into the
// json.RawMessage store.Store.CommitHarvestBatch's deltas parameter
// expects -- Attribute itself returns Delta values, not encoded patches,
// since it no longer talks to any sink at all.
type MetricsPatch struct {
	Tokens     TokenDelta                `json:"tokens"`
	Models     map[string]ModelBucket    `json:"models,omitempty"`
	Dispatches map[string]DispatchBucket `json:"dispatches,omitempty"`
	Effort     string                    `json:"effort,omitempty"`
	// Speed carries Delta.Speed through to the stored metrics bag's
	// top-level "speed" key -- read back by store.Store.Price (task 23)
	// to decide whether a run's models are priced at their fast-mode
	// input/output rate. Last-write-wins through jsonb_deep_add, like
	// Effort: two strings at the same key fall to "b replaces a"
	// (0005_jsonb_deep_add.sql's own doc comment), which is correct here
	// for the same reason it is correct for Effort.
	Speed string `json:"speed,omitempty"`
}

// Attributor assigns parsed transcript records to open stage windows and
// computes the per-stage-run token delta each one contributes. It never
// writes anywhere -- see Watcher (watcher.go) for how a computed delta
// reaches Postgres, atomically alongside the offset that makes it safe
// to compute the same way twice.
type Attributor struct {
	windows WindowSource
}

// NewAttributor builds an Attributor over windows -- a narrow,
// consumer-defined interface (see its own doc comment), so a caller can
// pass a fake with no database at all.
func NewAttributor(windows WindowSource) *Attributor {
	return &Attributor{windows: windows}
}

// Attribute assigns every record in records to the open window (per
// bestWindow) matching its session and containing its timestamp, and
// returns one Delta per touched stage run -- the whole-run total plus
// the same records broken out by model (Delta's own doc comment).
//
// Attribute is a pure computation over records and whatever
// WindowsForSession answers: it carries no state across calls, and two
// calls given the same records return the same deltas. Not
// double-counting those deltas once they reach Postgres is entirely the
// caller's responsibility -- Watcher.RunOnce commits them and this
// batch's consumed offset together, in one transaction, so a delta this
// method computes is either applied exactly once or not applied at all,
// never twice. The same holds per model: two batches touching the same
// run and the same model add rather than replace once committed
// (jsonb_deep_add, 0005_jsonb_deep_add.sql), so Attribute never needs to
// see more than one batch's worth of records to compute a correct delta.
//
// Records whose session matches no window, or whose timestamp falls
// outside every one of that session's windows, are silently not
// attributed -- design.md draws no window for them to belong to, and
// inventing one would be exactly the self-reporting this requirement
// exists to avoid.
func (a *Attributor) Attribute(ctx context.Context, records []Record) (map[int64]Delta, error) {
	deltas := make(map[int64]Delta)
	if len(records) == 0 {
		return deltas, nil
	}

	bySession := make(map[string][]Record)
	var sessionOrder []string
	for _, r := range records {
		if r.SessionID == "" {
			continue
		}
		if _, ok := bySession[r.SessionID]; !ok {
			sessionOrder = append(sessionOrder, r.SessionID)
		}
		bySession[r.SessionID] = append(bySession[r.SessionID], r)
	}

	for _, sessionID := range sessionOrder {
		windows, err := a.windows.WindowsForSession(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("harvest: windows for session %s: %w", sessionID, err)
		}
		for _, r := range bySession[sessionID] {
			w, ok := bestWindow(windows, r.Timestamp)
			if !ok {
				continue
			}
			d := deltas[w.StageRunID]
			d.Total.bucket(r.IsSidechain).add(r.Usage)
			if r.Usage.Speed != "" {
				d.Speed = r.Usage.Speed
			}
			if r.Model != "" {
				upsertBucket(&d.Models, r.Model, func(td TokenDelta) TokenDelta {
					td.bucket(r.IsSidechain).add(r.Usage)
					return td
				})
			}
			if r.AgentID != "" {
				upsertBucket(&d.Dispatches, r.AgentID, func(td TokenDelta) TokenDelta {
					td.bucket(r.IsSidechain).add(r.Usage)
					return td
				})
			}
			deltas[w.StageRunID] = d
		}
	}

	return deltas, nil
}

// bestWindow returns the window among windows whose interval contains ts,
// preferring the one with the latest StartedAt when more than one
// matches; where two such windows start at the same instant, the higher
// Attempt breaks the tie.
//
// More than one open window for the same session is not the ordinary
// case, but it is a real one. design.md's "stale open stage run" incident
// is what the latest-start rule itself exists for: an orphan left open by
// a dropped end mark and the stage actually running can both sit at
// attempt 1, so an attempt-only tie-break has nothing to say and used to
// fall back to iteration order -- storeWindowSource.WindowsForSession's
// unsorted query, oldest first, forever. Of two windows that both contain
// a message, the one that started later is the one the session is
// actually in: a stale open window must not outrank it.
//
// The same-instant Attempt tiebreak is *not* the ordinary replay path,
// despite once having been described as one. A replayed begin mark does
// carry the original attempt's own StartedAt unchanged -- the CLI
// captures it once, before the RPC attempt (cmd/myflow/stage.go's
// runStageBegin), and internal/reconcile.go's applyStageMarkEntry replays
// that same instant (StartedAt: req.StartedAt) -- but
// store.insertStageRunAndSupersede then closes the earlier attempt with
// ended_at set to that same instant, and Window.contains rejects an empty
// [T, T) interval: a window whose StartedAt equals its EndedAt contains
// no timestamp at all, not even T itself. By the time a harvest cycle
// reads WindowsForSession, the ordinary replay therefore presents exactly
// one open window at that instant, never two, and bestWindow is never
// asked to arbitrate between them.
//
// What the same-instant tiebreak actually defends against is windows the
// store no longer produces on that ordinary path: a stage run recorded
// before this change existed, one carrying no session_token (the
// supersede matches on session_token alone, so a NULL token is never
// closed by it), or the narrow instant before a concurrent supersede's
// own transaction commits. Attempt is kept for those, not for the replay
// case its own history once named.
//
// A window that is neither later nor higher never replaces the current
// best, so the result is deterministic under any input order -- the
// first match standing until a strictly later start, or a same-instant
// strictly higher attempt, displaces it.
//
// This is distinct from the exact-boundary case Window.contains' own doc
// comment covers: two *adjacent, non-overlapping* windows never reach
// this tie-break at all, because the half-open interval already resolves
// which one a boundary message belongs to before bestWindow is asked to
// choose between candidates.
func bestWindow(windows []Window, ts time.Time) (Window, bool) {
	var (
		best  Window
		found bool
	)
	for _, w := range windows {
		if !w.contains(ts) {
			continue
		}
		if !found || w.StartedAt.After(best.StartedAt) ||
			(w.StartedAt.Equal(best.StartedAt) && w.Attempt > best.Attempt) {
			best = w
			found = true
		}
	}
	return best, found
}
