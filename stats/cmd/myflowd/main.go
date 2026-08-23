// Command myflowd is the daemon that owns the myflow PostgreSQL pool and
// serves the state API. It binds loopback only -- a non-loopback
// MYFLOWD_HOST is refused before any listener is opened -- and shuts down
// gracefully: in-flight requests complete before the connection pool
// closes.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/sweep"
	"github.com/tweety53/agents/stats/internal/web"
)

// shutdownGrace bounds how long ListenAndServe's graceful shutdown waits
// for in-flight requests to finish before giving up and closing the
// connection pool anyway.
const shutdownGrace = 15 * time.Second

// harvestInterval is how often the transcript watcher scans
// ~/.claude/projects (or MYFLOW_TRANSCRIPTS_DIR) for new bytes. Like
// reconnectPingInterval below, this is off the pipeline's hot path --
// nothing a CLI command does waits on it -- so it favors a short enough
// interval that token figures show up in the statistics views soon after
// a stage runs, without polling the filesystem tree aggressively enough
// to matter.
const harvestInterval = 5 * time.Second

// reconnectPingInterval is how often Reconciler.Watch pings the store to
// detect design.md's "regains a database connection" transition -- the
// trigger for a replay beyond the one this daemon already runs once at
// startup. It is not on the pipeline's own hot path (nothing here blocks a
// CLI request), so it favors an interval short enough that an outage's
// journal entries land in the store again soon after the database comes
// back, without polling aggressively enough to matter against a local
// Postgres instance.
const reconnectPingInterval = 10 * time.Second

// sweepInterval is how often the abandoned-stage sweeper scans for stage
// runs whose session has gone silent. Off the pipeline's hot path, like
// harvestInterval and reconnectPingInterval above, so it favors an
// interval short enough that an abandoned stage shows up in the
// rework-rate view soon after its timeout elapses, without adding
// meaningful load to a local Postgres instance.
const sweepInterval = 1 * time.Minute

// sweepSilenceTimeout bounds how long a stage run may sit open with no end
// mark before the sweeper closes it with outcome "abandoned" (design.md's
// "Stage marks": "The daemon sweeps stages whose session has been silent
// past a timeout"). Set well above this pipeline's longest ordinary stage
// -- `/myflow-do`'s "SDD + TDD per task" can legitimately run for hours on
// a large task, and the review panel dispatches its own long-running
// subagents -- so the sweeper only ever catches a session that will never
// send an end mark, never one still genuinely working.
const sweepSilenceTimeout = 6 * time.Hour

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := run(logger); err != nil {
		logger.Error("myflowd exiting", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.FromEnv()
	if err != nil {
		// A malformed MYFLOWD_PORT lands here: refused before anything
		// else starts, per config.ErrInvalidPort -- never silently
		// defaulted to a working-but-wrong port.
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.RunMigrations(ctx); err != nil {
		return err
	}

	// Task 23: publish the real pricing rates before anything might call
	// Price -- the reconciler's replay and the watcher's own harvest loop
	// both start below, and either could touch a stage run whose cost
	// depends on these rows already existing. SeedPricing's ON CONFLICT
	// upsert (store.Store.PutPricing) makes this a no-op on every
	// subsequent startup, so seeding unconditionally on every run is
	// correct, not merely convenient.
	if err := st.SeedPricing(ctx); err != nil {
		return err
	}

	// Task 12: embed and serve the SPA at "/". internal/web.FS requires
	// stats/web/vite.config.ts's outDir (internal/web/dist) to already
	// exist and contain a real build -- go:embed itself refuses to
	// compile this binary otherwise (internal/web/embed.go's own doc
	// comment), which is this task's "fails loudly at compile time, not
	// silently at runtime" requirement. cmd/myflowd is therefore the one
	// package in this daemon that requires `vite build` to have run
	// first; internal/api's own tests do not, since api.WithSPA is the
	// only place that dependency is wired in.
	distFS, err := web.FS()
	if err != nil {
		return err
	}
	spaHandler, err := web.Handler(distFS)
	if err != nil {
		return err
	}

	srv, err := api.New(cfg, st, st, st, st, logger, api.WithSPA(spaHandler))
	if err != nil {
		// A non-loopback MYFLOWD_HOST lands here: refused before any
		// listener is opened, per config.Config.Validate.
		return err
	}

	// Task 6: replay the write-ahead journal now, at startup -- store.Open
	// above has already confirmed connectivity via its own ping, so this
	// is exactly the "at startup" half of design.md's "Availability and
	// reconciliation". The "on reconnect" half is Watch, started below.
	reconciler := reconcile.New(st, st, st, fallback.StateRoot(), logger)
	startupResult, startupErr := reconciler.Run(ctx)
	logReconcileResult(logger, "startup", startupResult, startupErr)

	watchCtx, stopWatch := context.WithCancel(ctx)
	defer stopWatch()
	go reconciler.Watch(watchCtx, st.Ping, reconnectPingInterval, func(result reconcile.Result, err error) {
		logReconcileResult(logger, "reconnect", result, err)
	})

	// Task 9: harvest session transcripts and attribute their usage to
	// open stage windows. st (a *store.Store) satisfies both
	// harvest.WindowSource (via the storeWindowSource adapter below,
	// backed by QueryStageRuns' session_id filter) and harvest.HarvestSink
	// directly -- GetHarvestOffset and CommitHarvestBatch's signatures
	// already match the interface harvest.go declares, so no adapter is
	// needed for that half. Their atomicity (one transaction covers both
	// a batch's additive metrics and its offset advance) is what removed
	// this daemon's earlier local harvest-offsets.json file entirely --
	// see watcher.go's own doc comment on HarvestSink for why a purely
	// local, separately-committed offset was unsafe (task 9's post-commit
	// review, findings F1 and its follow-up).
	transcriptsRoot, err := harvest.DefaultTranscriptsRoot()
	if err != nil {
		return err
	}
	attributor := harvest.NewAttributor(storeWindowSource{st})
	// Task 23: price every stage run a batch touches, once
	// CommitHarvestBatch has reported it applied (harvest.Pricer's own
	// doc comment explains why after, never inside, that commit). st
	// satisfies harvest.Pricer directly -- Price's signature already
	// matches it (compile-time check below) -- so, like HarvestSink, no
	// adapter is needed here.
	watcher := newTranscriptWatcher(transcriptsRoot, st, attributor, logger)
	go watcher.Run(watchCtx, harvestInterval)

	// Task 10: close stage runs whose session has gone silent past
	// sweepSilenceTimeout, with outcome "abandoned" -- design.md's "Stage
	// marks". st satisfies sweep.AbandonedSweeper directly (SweepAbandoned,
	// task 3), so no adapter is needed here either. A stage run this closes
	// may still receive a late harvest batch afterwards for messages inside
	// its now-fixed window -- see harvest.WindowSource's own doc comment
	// (attribute.go) for why that is deliberate, not a race left unhandled.
	sweeper := sweep.New(st, sweepSilenceTimeout, logger)
	go sweeper.Run(watchCtx, sweepInterval)

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("myflowd listening", "addr", cfg.Addr())
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("myflowd shutting down", "signal", ctx.Err())
	case err := <-serveErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	// ListenAndServe's goroutine returns nil once Shutdown has drained it;
	// wait for that so the store's pool (deferred above) closes only after
	// every in-flight request has actually finished, not merely after
	// Shutdown decided it was done.
	return <-serveErr
}

// newTranscriptWatcher builds the harvest.Watcher the daemon runs.
// Extracted out of run (KAN-172, task 7) so wiring_test.go can call it
// directly and assert on the *constructed* Watcher -- Watcher.HasPricer,
// Watcher.HasSessionTokenBinder -- rather than on this file's own source
// text, which a refactor could keep unchanged while silently dropping an
// option. This is the exact class of defect task 7 fixes: the daemon
// built this watcher inline with harvest.WithPricer(st) but no
// harvest.WithSessionTokenBinder(st), so pendingSessionTokens always
// returned nil and no stage run was ever bound, despite tasks 1-6 all
// working and 329 Go tests staying green throughout.
func newTranscriptWatcher(root string, st *store.Store, attributor *harvest.Attributor, logger *slog.Logger) *harvest.Watcher {
	// KAN-258: the second, dispatch-grain attribution pass runs beside the
	// first over the same batch (design.md, "Cost attribution"). st
	// satisfies both halves of it directly -- DispatchWindowsForSession
	// returns harvest.DispatchWindow, and MergeDispatchMetrics matches
	// harvest.DispatchMetricsSink -- so, like HarvestSink and Pricer,
	// neither needs an adapter (compile-time checks below).
	return harvest.NewWatcher(root, st, attributor, logger,
		harvest.WithPricer(st),
		harvest.WithSessionTokenBinder(st),
		harvest.WithDispatchAttribution(harvest.NewDispatchAttributor(st), st),
	)
}

// storeWindowSource adapts *store.Store to harvest.WindowSource: the one
// piece of glue internal/harvest needs but cannot build itself, since it
// never imports internal/store (its own package doc explains why --
// TestHarvestNeedsNoDatabase exercises the harvester with no database at
// all). This is the only place in the daemon that turns a session id into
// a store.Query -- QueryStageRuns' session_id filter (added to
// internal/store/query.go's allowlist by this same task) -- so a stage
// run's identity never crosses into internal/harvest as anything but the
// harvest.Window shape attribution actually needs.
type storeWindowSource struct{ st *store.Store }

func (s storeWindowSource) WindowsForSession(ctx context.Context, sessionID string) ([]harvest.Window, error) {
	runs, _, err := s.st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "session_id", Op: store.OpEq, Value: sessionID}},
		Limit:   store.NoLimit,
	})
	if err != nil {
		return nil, err
	}
	windows := make([]harvest.Window, 0, len(runs))
	for _, r := range runs {
		windows = append(windows, harvest.Window{
			StageRunID: r.ID,
			Attempt:    r.Attempt,
			SessionID:  sessionID,
			StartedAt:  r.StartedAt,
			EndedAt:    r.EndedAt,
		})
	}
	return windows, nil
}

// Compile-time check that *store.Store already satisfies harvest.HarvestSink
// with no adapter -- GetHarvestOffset and CommitHarvestBatch are written to
// match harvest.HarvestSink exactly (watcher.go's own doc comment on that
// interface explains why), so this is a guard against that drifting
// silently, not a type this file otherwise needs.
var (
	_ harvest.HarvestSink  = (*store.Store)(nil)
	_ harvest.WindowSource = storeWindowSource{}
	_ harvest.Pricer       = (*store.Store)(nil)
	// Unlike storeWindowSource above, the dispatch-grain window source
	// needs no adapter: store.DispatchWindowsForSession already answers in
	// harvest.DispatchWindow, its own doc comment explaining why the type
	// crosses that way round.
	_ harvest.DispatchWindowSource = (*store.Store)(nil)
	_ harvest.DispatchMetricsSink  = (*store.Store)(nil)
	_ sweep.AbandonedSweeper       = (*store.Store)(nil)
	// *store.Store satisfies the widened harvest.SessionTokenBinder (task
	// 6, kan-212-persist-per-dispatch-cost-tokens-model-and-role) with no
	// adapter either: RecordSessionTokenGiveUp, PersistedGiveUps and
	// MarkDispatchesUnattributedByID are written to match it exactly,
	// PersistedGiveUps returning harvest.GiveUp directly for the same
	// reason DispatchWindowsForSession returns harvest.DispatchWindow
	// directly, above.
	_ harvest.SessionTokenBinder = (*store.Store)(nil)
)

// logReconcileResult reports one Reconciler.Run outcome. A replay failure
// is deliberately not fatal to the daemon -- reconcile.Reconciler.Run
// already leaves anything it could not resolve safely in the journal for
// the next trigger (startup, reconnect, or `myflow journal flush`) to pick
// up, so logging and continuing is correct here, not a swallowed error.
func logReconcileResult(logger *slog.Logger, trigger string, result reconcile.Result, err error) {
	if err != nil {
		logger.Error("myflowd journal replay failed", "trigger", trigger, "error", err,
			"journals", result.Journals, "applied", result.Applied, "refused", result.Refused)
		return
	}
	if result.Applied > 0 || result.Refused > 0 {
		logger.Info("myflowd journal replay", "trigger", trigger,
			"journals", result.Journals, "applied", result.Applied, "refused", result.Refused)
	}
}
