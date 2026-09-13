package harvest

import (
	"context"
	"encoding/json"
	"time"
)

// DispatchAgentStamper fills an empty agent_id on one dispatch row,
// named by the (sessionToken, key) pair its own begin command carried --
// KAN-322's automatic capture of the id the harness reports at launch.
// The store implementation fills only a row whose agent_id is empty and
// reports whether it stamped, so a hand-typed id always wins and a
// launch whose row is absent or already named changes nothing.
//
// Defined here, at the consumer, per go-interface-design, like every
// other interface in this file: internal/harvest never imports
// internal/store, and *store.Store satisfies this with no adapter once
// store.Store.StampDispatchAgent exists.
type DispatchAgentStamper interface {
	StampDispatchAgent(ctx context.Context, sessionToken, key, agentID string) (stamped bool, err error)
}

// Deps is everything a Watcher needs beyond its root, its sink and its
// Attributor. It is one required parameter rather than a set of
// functional options (KAN-173): each of the interfaces it composes
// is optional in a type signature but mandatory in practice --
// production supplies exactly one real implementation of each, and all
// of them come from the same *store.Store -- so an omitted option
// compiled, tested green, and ran inert. Twice (KAN-16, KAN-172).
//
// Composing the interfaces rather than restating their methods is what makes
// a missing dependency a compile error too: adding a method to any
// constituent breaks every implementation that has not grown it.
type Deps interface {
	Pricer
	SessionTokenBinder
	DispatchMetricsSink
	DispatchWindowSource
	AgentWindowSource
	DispatchAgentStamper
}

// AgentWindowSource answers which dispatch rows carry an agentId -- the
// agent-file attribution pass's source (attributeAgentFileRecords). One
// resumed agent shares its agentId across several dispatch rows, and the
// split that separates their usage reads the rows back ordered by
// (started_at, id); the store query carrying that contract is
// store.Store.DispatchWindowsForAgent (internal/store/records.go).
//
// Returned directly as harvest.DispatchWindow for the same reason
// DispatchWindowSource's answer is: *store.Store then satisfies this
// interface with no adapter, and internal/harvest keeps importing nothing
// from internal/store.
type AgentWindowSource interface {
	DispatchWindowsForAgent(ctx context.Context, agentID string) ([]DispatchWindow, error)
}

// NoDeps satisfies Deps with a no-op for every method: zero values,
// nil errors, nothing recorded. It is exported for tests -- a test
// that needs no dependency passes NoDeps{}, and a test that needs one
// embeds NoDeps and overrides that single method.
//
// A daemon must never wire this. cmd/flowd/wiring_test.go's
// TestDaemonWiresTheRealStore is what says so.
type NoDeps struct{}

// Price is a no-op: it prices nothing and never fails.
func (NoDeps) Price(ctx context.Context, stageRunID int64) error {
	return nil
}

// UnresolvedSessionTokens always reports no unresolved tokens.
func (NoDeps) UnresolvedSessionTokens(ctx context.Context) (map[int64]string, error) {
	return nil, nil
}

// BindSession never binds anything.
func (NoDeps) BindSession(ctx context.Context, sessionToken string, sessionID string) (bound int64, err error) {
	return 0, nil
}

// RecordSessionTokenGiveUp records nothing and never fails.
func (NoDeps) RecordSessionTokenGiveUp(ctx context.Context, token, reason string, at time.Time) error {
	return nil
}

// PersistedGiveUps always reports no persisted give-ups.
func (NoDeps) PersistedGiveUps(ctx context.Context) ([]GiveUp, error) {
	return nil, nil
}

// MarkDispatchesUnattributedByID stamps nothing and never fails.
func (NoDeps) MarkDispatchesUnattributedByID(ctx context.Context, ids []int64, reason string, candidates int) error {
	return nil
}

// MarkDispatchesUnattributed stamps nothing and never fails.
func (NoDeps) MarkDispatchesUnattributed(ctx context.Context, token, reason string, candidates int) error {
	return nil
}

// MergeDispatchMetrics merges nothing and never fails.
func (NoDeps) MergeDispatchMetrics(ctx context.Context, dispatchID int64, patch json.RawMessage) error {
	return nil
}

// DispatchWindowsForSession always reports no dispatch windows.
func (NoDeps) DispatchWindowsForSession(ctx context.Context, sessionID string) ([]DispatchWindow, error) {
	return nil, nil
}

// DispatchWindowsForAgent always reports no dispatch windows.
func (NoDeps) DispatchWindowsForAgent(ctx context.Context, agentID string) ([]DispatchWindow, error) {
	return nil, nil
}

// StampDispatchAgent stamps nothing and reports false.
func (NoDeps) StampDispatchAgent(ctx context.Context, sessionToken, key, agentID string) (stamped bool, err error) {
	return false, nil
}
