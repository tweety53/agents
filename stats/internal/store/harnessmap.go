package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrDispatchPairInvalid is returned by RecordDispatch when the row's
// recorded model/effort pair is not one the dispatching harness's mapping
// can produce -- KAN-610. It is typed so internal/api answers 400 rather
// than the 500 internal/client reads as the store being unavailable (which
// would journal a write every future replay refuses), and so
// internal/reconcile retires the journalled entry as definitive rather
// than queueing it forever, exactly as it does for ErrAgentIDInvalid.
var ErrDispatchPairInvalid = errors.New("store: dispatch pair the harness mapping cannot produce")

// dispatchPair is one harness mapping row: the model and effort every
// dispatch on that harness runs, whatever was chosen for it.
type dispatchPair struct {
	model  string
	effort string
}

// harnessDispatchPairs is the harness mapping -- **Harness mapping**
// (`skills/flow-contracts/model-policy.md`) is canonical for it, and this
// table cites rather than restates it: on harness `zcode` every model a
// dispatch would be given is `glm-5.3-flash` at effort `high`, recorded as
// dispatched -- `unknown (agent-defined)` and a pre-mapping choice alike
// are pairs the mapping cannot produce, and a ledger carrying one records
// a model the dispatch could not have run. No other harness maps anything,
// so a harness absent here validates nothing: its dispatch rows record
// whatever the dispatcher set.
var harnessDispatchPairs = map[string]dispatchPair{
	"zcode": {model: "glm-5.3-flash", effort: "high"},
}

// validateDispatchPair refuses a dispatch whose recorded model/effort pair
// the session token's harness mapping cannot produce. The mapping speaks
// only where it can: a dispatch with no session token, or one whose token
// has marked no stage run yet, carries no harness, so any pair on it
// passes. The check sits beside validateAgentID in RecordDispatch's path
// rather than in the handler, so the live route and a replayed write
// cannot answer "is this dispatch recordable?" differently.
func (s *Store) validateDispatchPair(ctx context.Context, in records.Dispatch) error {
	if in.SessionToken == "" {
		return nil
	}
	harness, err := s.harnessForSessionToken(ctx, in.SessionToken)
	if err != nil {
		return err
	}
	want, mapped := harnessDispatchPairs[harness]
	if !mapped {
		return nil
	}
	if in.Model == want.model && in.Effort == want.effort {
		return nil
	}
	return fmt.Errorf("%w: recorded (%q, %q); harness %q runs only (%q, %q)",
		ErrDispatchPairInvalid, in.Model, in.Effort, harness, want.model, want.effort)
}

// harnessForSessionToken resolves the harness the stage runs marked under
// a session token carry. A token's marks all carry one harness -- the run
// marks it once and every stage of that run is the same harness -- so the
// latest row answers for the token, and a token with no rows resolves to
// "" (no harness knowable) rather than an error.
func (s *Store) harnessForSessionToken(ctx context.Context, token string) (string, error) {
	const q = `SELECT harness FROM stage_runs WHERE session_token = $1 ORDER BY id DESC LIMIT 1`
	var harness string
	if err := s.pool.QueryRow(ctx, q, token).Scan(&harness); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("store: harness for session token: %w", err)
	}
	return harness, nil
}
