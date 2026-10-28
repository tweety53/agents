// Command flowd is the daemon that owns the flow PostgreSQL pool and
// serves the state API. It binds loopback only -- a non-loopback
// FLOWD_HOST is refused before any listener is opened -- and shuts down
// gracefully: in-flight requests complete before the connection pool
// closes.
//
// It claims its port before its database. acquireStartup below validates
// the configuration, checks the pidfile named after the resolved port,
// opens the listener and only then records that pidfile -- all before
// store.Open is reached, so a start refused because another daemon
// already holds that port or that pidfile runs no migration, seeds no
// pricing and harvests no transcript
// (the single-instance requirement: "a refused start touches
// no database").
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/pidfile"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/sweep"
	"github.com/tweety53/agents/stats/internal/web"
)

// shutdownGrace bounds how long Serve's graceful shutdown waits for
// in-flight requests to finish before giving up and closing the
// connection pool anyway.
const shutdownGrace = 15 * time.Second

// harvestInterval is how often the transcript watcher scans
// ~/.claude/projects (or FLOW_TRANSCRIPTS_DIR) for new bytes. Like
// reconnectPingInterval below, this is off the pipeline's hot path --
// nothing a CLI command does waits on it -- so it favors a short enough
// interval that token figures show up in the statistics views soon after
// a stage runs, without polling the filesystem tree aggressively enough
// to matter.
const harvestInterval = 5 * time.Second

// reconnectPingInterval is how often Reconciler.Watch pings the store and,
// on every ping that succeeds, replays the journal -- the replay beyond the
// one this daemon already runs once at startup. It is not on the pipeline's
// own hot path (nothing here blocks a CLI request), so it favors an interval
// short enough that an outage's journal entries land in the store soon
// after the database or the daemon's HTTP surface comes back, without
// polling aggressively enough to matter against a local Postgres instance.
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
// -- `/flow`'s "SDD + TDD per task" can legitimately run for hours on
// a large task, and the review panel dispatches its own long-running
// subagents -- so the sweeper only ever catches a session that will never
// send an end mark, never one still genuinely working.
const sweepSilenceTimeout = 6 * time.Hour

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if len(os.Args) > 1 && os.Args[1] == "backfill-models" {
		if err := backfillModels(os.Args[2:]); err != nil {
			logger.Error("flowd backfill-models failed", "error", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "restamp-dispatch" {
		if err := restampDispatch(os.Args[2:]); err != nil {
			logger.Error("flowd restamp-dispatch failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "restamp-ends" {
		if err := restampEnds(os.Args[2:]); err != nil {
			logger.Error("flowd restamp-ends failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "retime-dispatch" {
		if err := retimeDispatch(os.Args[2:]); err != nil {
			logger.Error("flowd retime-dispatch failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "reoutcome-dispatch" {
		if err := reoutcomeDispatch(os.Args[2:]); err != nil {
			logger.Error("flowd reoutcome-dispatch failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "rekey-dispatch" {
		if err := rekeyDispatch(os.Args[2:]); err != nil {
			logger.Error("flowd rekey-dispatch failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "drop-pass" {
		if err := dropPass(os.Args[2:]); err != nil {
			logger.Error("flowd drop-pass failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "drop-mutation" {
		if err := dropMutation(os.Args[2:]); err != nil {
			logger.Error("flowd drop-mutation failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "fold-dispatch" {
		if err := foldDispatch(os.Args[2:]); err != nil {
			logger.Error("flowd fold-dispatch failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "clip-abandoned" {
		if err := clipAbandoned(); err != nil {
			logger.Error("flowd clip-abandoned failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "retoken-stage-run" {
		if err := retokenStageRun(os.Args[2:]); err != nil {
			logger.Error("flowd retoken-stage-run failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "unflag-verdict" {
		if err := unflagVerdict(os.Args[2:]); err != nil {
			logger.Error("flowd unflag-verdict failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if err := run(logger); err != nil {
		logger.Error("flowd exiting", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.FromEnv()
	if err != nil {
		// A malformed FLOWD_PORT lands here: refused before anything
		// else starts, per config.ErrInvalidPort -- never silently
		// defaulted to a working-but-wrong port.
		return err
	}

	// The prelude, before anything opens the database: the loopback rule,
	// the pidfile check, the listener, then the pidfile write. Everything
	// below this point runs only once this daemon holds the port it is
	// about to serve.
	lock, ln, err := acquireStartup(cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := lock.Release(logger); err != nil {
			logger.Error("flowd could not remove its pidfile", "error", err)
		}
	}()
	// This covers the window between the listener opening and srv.Serve
	// taking ownership of it: store.Open, the migrations, the seeding,
	// web.FS and api.New below each return with the port still held
	// otherwise, and every other resource in this function is released by
	// an explicit defer. Serve closes the listener itself on shutdown, so
	// on the ordinary path this call finds it already closed and returns
	// that error -- discarded here deliberately, since a close that
	// reports the work already done is exactly what is wanted.
	defer func() { _ = ln.Close() }()

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

	// KAN-914: a workspace daemon's fresh flow_<id> database has no
	// projects row, and every record write keyed on the project refuses
	// until one exists. Seed it from the working directory -- the same
	// checkout originRemote already reads -- with the key every flow CLI
	// write derives. A failure is logged, never fatal: the daemon then
	// starts exactly as it did before this seed existed.
	if seedsWorkspaceProject(cfg.DSN) {
		if key, mainCheckout, err := fallback.ProjectKey("."); err != nil {
			logger.Warn("flowd did not seed the workspace project row", "error", err)
		} else if err := st.EnsureProject(ctx, key, mainCheckout); err != nil {
			logger.Warn("flowd did not seed the workspace project row", "project_key", key, "error", err)
		}
	}

	// Task 12: embed and serve the SPA at "/". internal/web.FS requires
	// stats/web/vite.config.ts's outDir (internal/web/dist) to already
	// exist and contain a real build -- go:embed itself refuses to
	// compile this binary otherwise (internal/web/embed.go's own doc
	// comment), which is this task's "fails loudly at compile time, not
	// silently at runtime" requirement. cmd/flowd is therefore the one
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

	transcriptsRoot, err := harvest.DefaultTranscriptsRoot()
	if err != nil {
		return err
	}
	agentEnd := func(agentID string) (time.Time, bool) {
		at, found, err := harvest.AgentTranscriptEnd(transcriptsRoot, agentID)
		if err != nil {
			logger.Warn("flowd: read agent transcript end failed, the row ends at the daemon's clock", "agent_id", agentID, "error", err)
		}
		return at, found
	}
	srv, err := api.New(cfg, st, st, st, st, st, logger, api.WithSPA(spaHandler), api.WithGuardRuns(st), api.WithRepricer(st),
		api.WithCommitBase(commitBaseFromRemote(originRemote())), api.WithBuild(buildIdentity()),
		api.WithAgentEnds(agentEnd))
	if err != nil {
		// api.New calls config.Config.Validate itself, which keeps
		// internal/api correct for any other caller, but the loopback
		// rule is no longer enforced here for this daemon: acquireStartup
		// above already refused a non-loopback FLOWD_HOST, before the
		// listener it hands to Serve below was opened.
		return err
	}

	// Task 6: replay the write-ahead journal now, at startup -- store.Open
	// above has already confirmed connectivity via its own ping, so this
	// is exactly the "at startup" half of design.md's "Availability and
	// reconciliation". The "on reconnect" half is Watch, started below, which
	// replays on every healthy ping.
	// A workspace daemon skips both halves: the journal is shared, and an
	// entry replayed into its throwaway database leaves the journal and is
	// dropped with that database at cleanup (KAN-913).
	watchCtx, stopWatch := context.WithCancel(ctx)
	defer stopWatch()
	if replaysSharedJournal(cfg.DSN) {
		reconciler := reconcile.New(st, st, st, fallback.StateRoot(), logger)
		reconciler.AgentEnd = agentEnd
		startupResult, startupErr := reconciler.Run(ctx)
		logReconcileResult(logger, "startup", startupResult, startupErr)
		go reconciler.Watch(watchCtx, st.Ping, reconnectPingInterval, func(result reconcile.Result, err error) {
			logReconcileResult(logger, "watch", result, err)
		})
	} else {
		logger.Info("flowd journal replay skipped: not the dev workspace's flow database")
	}

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
	rolloutRoot, err := harvest.DefaultZcodeRolloutRoot()
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
	watcher := newTranscriptWatcher(transcriptsRoot, rolloutRoot, st, attributor, logger)
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
		logger.Info("flowd listening", "addr", cfg.Addr())
		serveErr <- srv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		logger.Info("flowd shutting down", "signal", ctx.Err())
	case err := <-serveErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	// Serve's goroutine returns nil once Shutdown has drained it;
	// wait for that so the store's pool (deferred above) closes only after
	// every in-flight request has actually finished, not merely after
	// Shutdown decided it was done.
	return <-serveErr
}

// acquireStartup claims the daemon's port, then records the pidfile,
// before anything touches the database. On any failure it releases
// whatever it already claimed and returns no listener.
//
// Its steps run in one order and only one: cfg.Validate, pidfile.Check,
// net.Listen, pidfile.Write. **The write comes after the bind, and that
// is the point.** The pidfile race and the bind race are independent, so
// a daemon that checked and wrote before binding could be the one that
// then loses the bind -- and its own pid-matched Release would delete the
// pidfile of the daemon actually holding the port, disarming the refusal
// for every later start. Writing only after the bind removes the case
// outright: a bind-loser has written nothing and has nothing to release,
// and the daemon named in the file is by construction the one holding the
// port.
//
// The ordering is also KAN-170's second defect: a start refused because
// another daemon holds the port, or holds the pidfile named after it,
// must run no migration, seed no pricing and harvest no transcript into a
// database a live daemon is already writing
// (the single-instance requirement: "a refused start touches
// no database"). Extracting the prelude out of run is what makes that
// testable: wiring_test.go calls this directly, with no database
// anywhere, exactly as it calls newTranscriptWatcher below.
//
// cfg.Validate is called here explicitly even though api.New calls it
// too, because api.New now runs after the listener is open: without this
// call the loopback-only rule would be enforced only once a non-loopback
// listener already existed.
func acquireStartup(cfg config.Config, logger *slog.Logger) (*pidfile.Lock, net.Listener, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	path := pidfile.Path(cfg.Port)
	if err := pidfile.Check(path, logger); err != nil {
		return nil, nil, err
	}

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		// Nothing to undo: Check wrote nothing, so a daemon that loses
		// this bind leaves the winner's pidfile exactly as it found it.
		return nil, nil, fmt.Errorf("flowd: listen on %s: %w", cfg.Addr(), err)
	}

	lock, err := pidfile.Write(path)
	if err != nil {
		// The listener goes back before the error does: this process is
		// about to exit, and a held port it never serves would refuse
		// the next start for no reason.
		if closeErr := ln.Close(); closeErr != nil {
			logger.Error("flowd could not close its listener after failing to write its pidfile", "error", closeErr)
		}
		return nil, nil, err
	}

	return lock, ln, nil
}

// newTranscriptWatcher builds the harvest.Watcher the daemon runs, over
// both transcript sources this daemon harvests: the Claude Code transcripts
// root and the ZCode rollout root (kan-479).
// Extracted out of run (KAN-172, task 7) so wiring_test.go can call it
// directly and assert on the *constructed* Watcher -- originally through
// Watcher.HasPricer and Watcher.HasSessionTokenBinder, now (KAN-173)
// through TestDaemonWiresTheRealStore's reflection over Watcher's deps
// field -- rather than on this file's own source text, which a refactor
// could keep unchanged while silently dropping a dependency.
//
// This is the wiring that KAN-172 task 7 once got wrong: the daemon built
// this watcher with harvest.WithPricer(st) but no
// harvest.WithSessionTokenBinder(st), so pendingSessionTokens always
// returned nil and no stage run was ever bound, despite tasks 1-6 all
// working and 329 Go tests staying green throughout. KAN-173 closes the
// class of defect, not just this instance: harvest.Deps collapses every
// dependency into one required NewWatcher parameter, so dropping one here
// is now a compile error rather than a silently inert option. st (a
// *store.Store) satisfies harvest.Deps whole, with no adapter --
// DispatchWindowsForSession returns harvest.DispatchWindow and
// MergeDispatchMetrics matches harvest.DispatchMetricsSink, exactly like
// HarvestSink and Pricer already did (compile-time checks below).
func newTranscriptWatcher(transcriptsRoot, rolloutRoot string, st *store.Store, attributor *harvest.Attributor, logger *slog.Logger) *harvest.Watcher {
	return harvest.NewWatcher([]harvest.Source{
		harvest.NewClaudeSource(transcriptsRoot),
		harvest.NewRolloutSource(rolloutRoot),
	}, st, attributor, st, logger)
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
	// *store.Store satisfies harvest.SessionTokenBinder (task 6,
	// kan-212-persist-per-dispatch-cost-tokens-model-and-role) with no
	// adapter either: RecordSessionTokenGiveUp, PersistedGiveUps and
	// MarkDispatchesUnattributed are written to match it exactly,
	// PersistedGiveUps returning harvest.GiveUp directly for the same
	// reason DispatchWindowsForSession returns harvest.DispatchWindow
	// directly, above. (Its MarkDispatchesUnattributedByID sibling left
	// the interface with KAN-414, when the dispatch-grain ambiguity it
	// stamped became an apportionment; the store method went with it.)
	_ harvest.SessionTokenBinder = (*store.Store)(nil)
	// harvest.Deps composes Pricer, SessionTokenBinder,
	// DispatchMetricsSink and DispatchWindowSource -- the four checks
	// above -- into the one required NewWatcher parameter (KAN-173).
	// This guard is what proves *store.Store satisfies it whole, with
	// no adapter, before task 2 makes that the daemon's only option.
	_ harvest.Deps = (*store.Store)(nil)
)

// replaysSharedJournal reports whether dsn names the dev workspace's own
// `flow` database, the one store the shared journal under
// fallback.StateRoot() is replayed into. A workspace database
// (flow_<id>, .flow/project.md's ## workspace isolation) never is.
func replaysSharedJournal(dsn string) bool {
	u, err := url.Parse(dsn)
	return err == nil && u.Path == "/flow"
}

// seedsWorkspaceProject reports whether dsn names a workspace database
// (flow_<id>, .flow/project.md's ## workspace isolation) whose daemon seeds
// its own project row at startup. The dev workspace's flow database never
// does -- its writers bootstrap rows -- nor does a flow_<id>_uitest
// database, whose rows cmd/uitest-seed owns.
func seedsWorkspaceProject(dsn string) bool {
	u, err := url.Parse(dsn)
	return err == nil && strings.HasPrefix(u.Path, "/flow_") && !strings.HasSuffix(u.Path, "_uitest")
}

// logReconcileResult reports one Reconciler.Run outcome. A replay failure
// is deliberately not fatal to the daemon -- reconcile.Reconciler.Run
// already leaves anything it could not resolve safely in the journal for
// the next trigger (startup, a Watch tick, or `flow journal flush`) to pick
// up, so logging and continuing is correct here, not a swallowed error.
func logReconcileResult(logger *slog.Logger, trigger string, result reconcile.Result, err error) {
	if err != nil {
		logger.Error("flowd journal replay failed", "trigger", trigger, "error", err,
			"journals", result.Journals, "applied", result.Applied, "refused", result.Refused)
		return
	}
	if result.Applied > 0 || result.Refused > 0 {
		logger.Info("flowd journal replay", "trigger", trigger,
			"journals", result.Journals, "applied", result.Applied, "refused", result.Refused)
	}
}

// originRemote is the daemon working directory's `git remote get-url
// origin`, or "" when git, the checkout or the remote is missing -- the
// self-review view then renders refs as plain text, never fails to start.
func originRemote() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// commitBaseFromRemote maps a GitHub origin remote, in its scp-like ssh,
// ssh:// or https form, to its https commit base. Any other remote maps to "".
func commitBaseFromRemote(remote string) string {
	var repo string
	for _, prefix := range []string{"git@github.com:", "ssh://git@github.com/", "https://github.com/"} {
		if rest, ok := strings.CutPrefix(remote, prefix); ok {
			repo = rest
			break
		}
	}
	if repo == "" {
		return ""
	}
	return "https://github.com/" + strings.TrimSuffix(strings.TrimSuffix(repo, "/"), ".git") + "/commit"
}

// buildIdentity names this binary's build from the VCS stamp `go build`
// writes into it, for the self-review bundle's build note (KAN-771); ""
// when the binary carries none.
func buildIdentity() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return formatBuild(info.Settings)
}

// formatBuild renders vcs.revision, vcs.time and vcs.modified as
// "build <rev12> (<time>)", ", modified" inside the parentheses for a dirty
// tree; "" with no vcs.revision.
func formatBuild(settings []debug.BuildSetting) string {
	var rev, at, modified string
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			at = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if rev == "" {
		return ""
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	out := "build " + rev + " (" + at
	if modified == "true" {
		out += ", modified"
	}
	return out + ")"
}

// unflagVerdict implements `flowd unflag-verdict <verdict-id>`: it clears
// that verdict's false-positive flag (store.Store.UnflagVerdict) -- the repair
// for a flag that landed on the wrong row. It talks to the database
// directly, so it needs no running daemon.
func unflagVerdict(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: flowd unflag-verdict <verdict-id>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("verdict id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.UnflagVerdict(ctx, id)
}

// restampDispatch implements `flowd restamp-dispatch <dispatch-id>
// <agent-transcript | parent-transcript>`. Given an agent transcript, it
// names the row's agent from the file name and replaces the row's figures
// with that whole transcript's usage (store.Store.RestampDispatch) -- the
// repair for a row the harvester stamped with the wrong agent. Given a
// parent transcript, it first marks an agentless row inline
// (store.Store.MarkDispatchInline), then replaces the inline row's figures
// with the share the dispatch attributor gives it over the whole transcript,
// zero tokens when it has none (harvest.ClosedInlinePatch,
// store.Store.RestampInlineDispatch) -- the repair for a back-filled inline
// row or one begun without -agent-id inline; a row naming another agent is
// refused. It talks to the database directly, so it needs no running
// daemon.
func restampDispatch(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: flowd restamp-dispatch <dispatch-id> <agent-transcript.jsonl | parent-transcript.jsonl (marks an agentless row inline)>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("dispatch id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	agentID, ok := harvest.AgentIDFromTranscriptPath(args[1])
	if !ok {
		raw, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		if _, err := st.MarkDispatchInline(ctx, id); err != nil {
			return err
		}
		patch, err := harvest.ClosedInlinePatch(ctx, st, raw, id)
		if err != nil {
			return err
		}
		return st.RestampInlineDispatch(ctx, id, patch)
	}
	patch, err := harvest.AgentFileDispatchPatch(args[1])
	if err != nil {
		return err
	}
	return st.RestampDispatch(ctx, id, agentID, patch)
}

// backfillModels implements `flowd backfill-models`: it stamps every
// dispatch that carries an agent id and no modelVersion with the model its
// agent transcript names (model-version-split), printing how many rows it
// stamped. Like restamp-dispatch it talks to the database directly, so it
// needs no running daemon.
func backfillModels(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: flowd backfill-models")
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	root, err := harvest.DefaultTranscriptsRoot()
	if err != nil {
		return err
	}
	versions, err := harvest.AgentModelVersions(root)
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	n, err := st.BackfillModelVersions(ctx, func(agentID string) (string, bool) {
		v, ok := versions[agentID]
		return v, ok
	})
	if err != nil {
		return err
	}
	fmt.Printf("stamped: %d\n", n)
	return nil
}

// foldDispatch implements `flowd fold-dispatch <keep-id> <drop-id>`
// (store.Store.FoldDispatch): the repair for a verifier row whose
// -diff-base a second, agentless row was back-filled to correct. It talks
// to the database directly, so it needs no running daemon.
func foldDispatch(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: flowd fold-dispatch <keep-id> <drop-id>")
	}
	var ids [2]int64
	for i, a := range args {
		id, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			return fmt.Errorf("dispatch id %q: %w", a, err)
		}
		ids[i] = id
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.FoldDispatch(ctx, ids[0], ids[1])
}

// clipAbandoned implements `flowd clip-abandoned`
// (store.Store.ClipAbandonedStageRuns): the repair for abandoned stage runs
// the sweep ended past the change's next stage (kan-1024's self-review). It
// talks to the database directly, so it needs no running daemon.
func clipAbandoned() error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	n, err := st.ClipAbandonedStageRuns(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("clipped: %d\n", n)
	return nil
}

// retokenStageRun implements `flowd retoken-stage-run <stage-run-id>
// <session-token>` (store.Store.RetokenStageRun): the repair for a stage run
// begun under a placeholder token (kan-1029's self-review). It talks to the
// database directly, so it needs no running daemon.
func retokenStageRun(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: flowd retoken-stage-run <stage-run-id> <session-token>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("stage run id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.RetokenStageRun(ctx, id, args[1])
}

// rekeyDispatch implements `flowd rekey-dispatch <dispatch-id> <key>`
// (store.Store.RekeyDispatch): the repair for a row whose key's suffix
// expanded to nothing (kan-1009's self-review). It talks to the database
// directly, so it needs no running daemon.
func rekeyDispatch(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: flowd rekey-dispatch <dispatch-id> <key>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("dispatch id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.RekeyDispatch(ctx, id, args[1])
}

// dropPass implements `flowd drop-pass <pass-id>` (store.Store.DropPass):
// the repair for a pass-log line recorded twice (kan-1040). It talks to the
// database directly, so it needs no running daemon.
func dropPass(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: flowd drop-pass <pass-id>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("pass id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.DropPass(ctx, id)
}

// dropMutation implements `flowd drop-mutation <mutation-id>`
// (store.Store.DropMutation): the repair for a mutation proof recorded for
// a fix that never ran (kan-1046 self-review, 1d). It talks to the database
// directly, so it needs no running daemon.
func dropMutation(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: flowd drop-mutation <mutation-id>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("mutation id %q: %w", args[0], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.DropMutation(ctx, id)
}

// reoutcomeDispatch implements `flowd reoutcome-dispatch <dispatch-id>
// <outcome> [<cause>]` (store.Store.SetDispatchOutcome): the repair for a
// row a hand-run end closed with the wrong outcome (kan-1017). The outcome
// is one of flow record dispatch end's closed set, and a cause is accepted
// with blocked only.
func reoutcomeDispatch(args []string) error {
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("usage: flowd reoutcome-dispatch <dispatch-id> <outcome> [<cause>]")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("dispatch id %q: %w", args[0], err)
	}
	outcome, cause := args[1], ""
	if len(args) == 3 {
		cause = args[2]
	}
	if !slices.Contains([]string{"completed", "clean", "fix", "blocked", "aborted", "fallback", "timed-out", "skipped"}, outcome) {
		return fmt.Errorf("outcome %q is not a dispatch outcome", outcome)
	}
	if cause != "" && outcome != "blocked" {
		return fmt.Errorf("a cause is recorded with outcome blocked only, not %q", outcome)
	}
	if cause != "" && !slices.Contains([]string{"environment", "test-failure", "missing-fixture"}, cause) {
		return fmt.Errorf("cause %q is not a dispatch cause", cause)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.SetDispatchOutcome(ctx, id, outcome, cause)
}

// restampEnds implements `flowd restamp-ends <since>`
// (store.Store.RestampStampedEnds): the repair for agent rows whose daemon
// stamp a closed row kept instead of the agent's last transcript write
// (kan-1044 self-review, F1-a). It talks to the database directly.
func restampEnds(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: flowd restamp-ends <since RFC3339>")
	}
	since, err := time.Parse(time.RFC3339, args[0])
	if err != nil {
		return fmt.Errorf("since %q: %w", args[0], err)
	}
	root, err := harvest.DefaultTranscriptsRoot()
	if err != nil {
		return err
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	n, err := st.RestampStampedEnds(ctx, since, func(agentID string) (time.Time, bool) {
		at, found, err := harvest.AgentTranscriptEnd(root, agentID)
		return at, err == nil && found
	})
	fmt.Printf("restamped: %d\n", n)
	return err
}

// retimeDispatch implements `flowd retime-dispatch <dispatch-id>
// <started-at> <ended-at> <commit>` (store.Store.RetimeDispatch): the repair
// for an inline row whose begin and end were recorded after the work
// (kan-1003). It talks to the database directly, so it needs no running
// daemon.
func retimeDispatch(args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: flowd retime-dispatch <dispatch-id> <started-at> <ended-at> <commit>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("dispatch id %q: %w", args[0], err)
	}
	started, err := time.Parse(time.RFC3339, args[1])
	if err != nil {
		return fmt.Errorf("started-at %q: %w", args[1], err)
	}
	ended, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		return fmt.Errorf("ended-at %q: %w", args[2], err)
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	ctx := context.Background()
	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer st.Close()
	return st.RetimeDispatch(ctx, id, started, ended, args[3])
}
