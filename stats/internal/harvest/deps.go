package harvest

import (
	"context"
	"encoding/json"
	"time"
)

// Deps is everything a Watcher needs beyond its root, its sink and its
// Attributor. It is one required parameter rather than a set of
// functional options (KAN-173): each of the four interfaces it composes
// is optional in a type signature but mandatory in practice --
// production supplies exactly one real implementation of each, and all
// four come from the same *store.Store -- so an omitted option
// compiled, tested green, and ran inert. Twice (KAN-16, KAN-172).
//
// Composing the four rather than restating their methods is what makes
// a fifth dependency a compile error too: adding a method to any
// constituent breaks every implementation that has not grown it.
type Deps interface {
	Pricer
	SessionTokenBinder
	DispatchMetricsSink
	DispatchWindowSource
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
