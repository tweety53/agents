// Command uitest-seed populates the UI-test stack's database with a fixed,
// committed fixture: two projects, changes spanning STARTED, IN_PROGRESS
// and FINISHED, and stage runs carrying token usage, so the UI-test
// stack's interface renders a populated view rather than an empty one
// (specs/myflow-ui-test-stack/spec.md, "The test stack starts
// populated").
//
// It refuses to run against any database whose name does not end in
// "_uitest" -- checked before any statement is issued -- so a misaimed
// run cannot write fixture data into the live database
// (specs/myflow-ui-test-stack/spec.md, "Destructive test-stack paths
// refuse to act on any other database").
//
// It does not seed pricing: cmd/flowd/main.go already calls
// store.SeedPricing at startup, and make ui-test-up (task 6) starts the
// daemon before running this seeder, so the pricing table is already
// populated by the time this runs. Duplicating that call here would be a
// second place the same seed data is written from.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// openTimeout bounds how long connecting to and pinging the target
// database may take before this command gives up.
const openTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := run(logger); err != nil {
		logger.Error("uitest-seed exiting", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.FromEnv()
	if err != nil {
		return fmt.Errorf("uitest-seed: resolve config: %w", err)
	}

	// The guard runs before any statement is issued against cfg.DSN's
	// target -- before store.Open even dials it.
	if err := requireUitestDatabase(cfg.DSN); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	st, err := store.Open(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("uitest-seed: open store: %w", err)
	}
	defer st.Close()

	if err := st.RunMigrations(ctx); err != nil {
		return fmt.Errorf("uitest-seed: run migrations: %w", err)
	}

	if err := seedFixture(ctx, st); err != nil {
		return err
	}

	logger.Info("uitest-seed: fixture seeded", "dsn", cfg.DSN)
	return nil
}
