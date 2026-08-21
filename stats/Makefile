.PHONY: test web-test web-build build up down restart ui-test-up ui-test-down check-uitest-overrides

# Runs the Go test suite. Tests that need the compose stack's PostgreSQL
# skip themselves with a clear message when it is not running.
#
# internal/web and cmd/myflowd both embed web/'s build output
# (internal/web/dist, via //go:embed) and therefore do not compile at all
# without web-build having run first -- deliberately: task 12's own
# requirement is that a build with no dist/ present fails loudly at
# compile time, never silently at runtime. web-build is a prerequisite
# here so a plain `make test` on a fresh checkout builds the SPA it needs
# rather than hitting that failure; running `go test ./...` directly,
# skipping this Makefile, still hits it exactly as intended.
#
# check-uitest-overrides is a prerequisite here for the same reason
# web-build is: a Makefile edit that drops `override` from a UITEST_*
# variable (see the comment above those variables, below) is exactly as much
# a defect as a missing SPA build, and `make test` is the one command every
# workflow here already runs before calling anything else done. It is ALSO a
# prerequisite of ui-test-up and ui-test-down themselves (below) -- those,
# not `make test`, are the documented entry point that actually runs the
# destructive commands a dropped `override` would expose, so catching the
# defect only when someone happens to run `make test` first is not enough.
test: web-build check-uitest-overrides
	go test ./... -race -count=1

check-uitest-overrides:
	../scripts/check-uitest-overrides.sh

# Runs the SPA's own test suite (vitest).
web-test:
	cd web && npm ci && npm test

# Builds the SPA with Vite, into internal/web/dist -- see
# web/vite.config.ts's own comment for why that directory, not web/dist,
# is the build's outDir.
web-build:
	cd web && npm ci && npm run build

# The whole daemon binary: the Vite build, then `go build`, in that order
# -- never the other way around -- so the embedded assets `go build`
# bakes in are always the ones just produced, never a stale dist/ left
# over from an earlier build.
build: web-build
	go build ./...

# Brings up the dedicated myflow-postgres container (see docker-compose.yml).
# Does not touch any other compose stack running on this machine.
up:
	docker compose up -d

down:
	docker compose down

# Rebuilds and restarts the LIVE stack: the myflow-postgres container and the
# myflowd daemon on its default port. `up`/`down` above manage the container
# alone; this is the whole-stack equivalent, for when the daemon is running
# stale code.
#
# `build` is a PREREQUISITE, not a recipe step, and the ordering is the point:
# Make runs it to completion before the recipe's first line, so a broken SPA
# build or a Go compile error fails with the running daemon still up and
# serving. Stopping first and building second would turn every failed build
# into unplanned downtime.
#
# NO PIDFILE, deliberately. The live daemon is usually started by hand
# (`./bin/myflowd` from a terminal), so a pidfile this target wrote would be
# absent exactly when it is needed. The port is the authority instead: whatever
# holds $(LIVE_PORT) IS the live daemon, which is also why a stale pidfile
# cannot send this `kill` somewhere else. That matters here for the same reason
# the `override` rules below matter for the UI-test stack -- a `kill` driven by
# a caller-nameable file is the hazard; a `kill` driven by "who holds this
# port" has no such handle. $(LIVE_PORT) is still `override` so it cannot be
# repointed at another service's port from the command line.
#
# The daemon is started with nohup and detached: the operator's own foreground
# run dies with their terminal, and this target is invoked from wherever, so a
# child of `make` would die when Make exits. Its output goes to $(LIVE_LOG)
# rather than the terminal for the same reason.
#
# A DETACHED LIVE DAEMON HARVESTS TRANSCRIPTS UNATTENDED. It scans the real
# MYFLOW_TRANSCRIPTS_DIR every 5s (harvestInterval, cmd/myflowd/main.go) and
# writes into the live database -- during this repo's own development an agent
# left running that way harvested 2,961 transcript offsets before anyone
# noticed (see .myflow/project.md's `## run` section). That is the intended
# behaviour of the live stack, not a defect, but it is why this target restarts
# the live daemon rather than silently starting one that was never running:
# see the "was not running" branch below, which says so instead of assuming
# the operator wanted a daemon.
override LIVE_PORT := 4173
override LIVE_BIN := bin/myflowd
override LIVE_LOG := /tmp/myflow-live.log

restart: build
	@pid="$$(lsof -ti tcp:$(LIVE_PORT) -sTCP:LISTEN 2>/dev/null | head -1)"; \
	if [ -n "$$pid" ]; then \
		echo "restart: stopping myflowd $$pid on :$(LIVE_PORT)"; \
		kill "$$pid" 2>/dev/null; \
		for i in $$(seq 1 30); do \
			kill -0 "$$pid" 2>/dev/null || break; \
			sleep 1; \
		done; \
		if kill -0 "$$pid" 2>/dev/null; then \
			echo "restart: myflowd $$pid did not exit within 30s -- still holding :$(LIVE_PORT)" >&2; \
			exit 1; \
		fi; \
	else \
		echo "restart: nothing was listening on :$(LIVE_PORT) -- starting a daemon that was not previously running"; \
	fi
	# Both, and in this order, on purpose: `up -d` is what creates the
	# container when it is absent (`restart` alone errors there), and
	# `restart` is what guarantees a fresh process when it was already up
	# (`up -d` alone is a no-op there). Neither covers both cases.
	docker compose up -d
	docker compose restart
	go build -o $(LIVE_BIN) ./cmd/myflowd
	@nohup ./$(LIVE_BIN) >> $(LIVE_LOG) 2>&1 & \
	echo "restart: started myflowd (pid $$!), logging to $(LIVE_LOG)"
	@echo "restart: waiting for myflowd to answer on :$(LIVE_PORT)"
	@for i in $$(seq 1 60); do \
		curl -sf http://127.0.0.1:$(LIVE_PORT)/api/v1/changes >/dev/null 2>&1 && \
			{ echo "restart: ready on http://127.0.0.1:$(LIVE_PORT)"; exit 0; }; \
		sleep 1; \
	done; \
	echo "restart: myflowd did not answer on :$(LIVE_PORT) within 60s -- see $(LIVE_LOG)" >&2; \
	exit 1

# The UI-test stack: a database and a daemon instance separate from the
# live ones, for ad-hoc UI testing from the main checkout (see
# openspec/changes/kan-180-separate-db-and-app-for-ui-testing/specs/myflow-ui-test-stack/spec.md).
# Same myflow-postgres container as the live stack (host port 5433), a
# distinct database (myflow_uitest) and a distinct daemon port (4174), so
# both stacks can run at the same time without interfering.
# `override` refuses any command-line or environment reassignment of
# UITEST_DB and UITEST_ID outright -- a command-line assignment overrides a
# simple Makefile assignment (`:=`) regardless of operator, so without
# `override` `make ui-test-up UITEST_DB=myflow` (or `UITEST_ID=<anything>`)
# would resolve that variable to the caller's value despite the `:=` below.
# `override` is the one directive that wins back that precedence, and both
# variables need it, independently, because each guards a different
# destructive path:
#
# - `override UITEST_ID` protects ui-test-up's force-drop and ui-test-down's
#   drop, both of which are driven by $(UITEST_ID) via
#   `../scripts/workspace.sh remove $(UITEST_ID) [--force]` (workspace.sh
#   derives the database name from the id itself). Confirmed with
#   `make ui-test-up UITEST_ID=kan-999-fake -n`: without `override` here the
#   dry run prints `../scripts/workspace.sh remove kan-999-fake --force`,
#   which would force-drop a real `myflow_kan_999_fake` workspace database --
#   possibly one belonging to an active `/myflow-do` worktree -- since
#   workspace.sh's own `myflow_` prefix check only ever protects the shared
#   "myflow" database, not other workspace databases.
#
# - `override UITEST_DB` protects UITEST_DSN below, which MYFLOWD_DSN points
#   a second `myflowd` (and `uitest-seed`) at. cmd/myflowd/main.go runs
#   RunMigrations and SeedPricing, and cmd/uitest-seed seeds its fixture,
#   against whatever DSN they are handed, with no guard of their own -- so an
#   unprotected UITEST_DB would let `make ui-test-up UITEST_DB=myflow` start
#   a second daemon that migrates, seeds and serves the live database on
#   port 4174, live data included.
#
# Do not remove or weaken either `override`, and if this recipe is ever
# copied into another target, copy both `override`s with it -- without them,
# the defects above reach the live database again, silently.
#
# RULE, not a list: every UITEST_* variable declared below -- present ones
# and any added later -- MUST be `override`. ui-test-up and ui-test-down
# take no parameters; none of these names is meant to be operator-tunable.
# Without `override`, a command-line or environment assignment
# (`make ui-test-up UITEST_LOG=/etc/passwd`, or `make -e` with the same
# variable exported) silently wins over the `:=` below, and several of
# these variables are read straight into a `dropdb --force`, an unqualified
# `rm -rf`, or a `kill` on a pid read from a caller-named file. Naming and
# fixing the exposed variable one at a time is what let this recur across
# three prior review rounds (UITEST_DB, then UITEST_ID, then UITEST_DSN /
# UITEST_TRANSCRIPTS_DIR / UITEST_STATE_DIR / UITEST_LOG / UITEST_BIN /
# UITEST_PIDFILE) -- `scripts/check-uitest-overrides.sh` (invoked by
# `make check-uitest-overrides`, and by `make test`) enforces the rule
# itself so it cannot recur a fourth time.
override UITEST_DB := myflow_uitest
# The workspace id scripts/workspace.sh derives myflow_uitest from (its
# `myflow_<id_underscored>` rule, id "uitest" has no "-" to replace) --
# kept alongside UITEST_DB, not computed from it, so the two Makefile
# callers below name the id directly rather than reversing the DSN back
# into one.
override UITEST_ID := uitest
override UITEST_PORT := 4174
override UITEST_DSN := postgres://myflow:myflow@localhost:5433/$(UITEST_DB)?sslmode=disable
override UITEST_PIDFILE := /tmp/myflow-uitest.pid
override UITEST_BIN := /tmp/myflow-uitest-myflowd
override UITEST_TRANSCRIPTS_DIR := /tmp/myflow-uitest-transcripts
override UITEST_STATE_DIR := /tmp/myflow-uitest-state
override UITEST_LOG := /tmp/myflow-uitest.log

# Defence in depth for the `rm -rf` sites below: `override` (above) is what
# actually stops an operator-controlled value from reaching them, but an
# unqualified `rm -rf $(VAR)` is still a hazard if a variable were ever
# empty -- `rm -rf ""` with a stray leading space in the recipe expands to
# `rm -rf` on the current directory's contents by way of shell word
# splitting -- or a future edit accidentally overwrote one of the `:=`
# assignments above without `override`. Every path this Makefile deletes
# falls under this one prefix; refuse anything that does not.
#
# WHAT THIS ACTUALLY CHECKS, NO MORE: the `case` below is a LEXICAL prefix
# match on the string, not a resolved-path check -- it does not shell out to
# `realpath`/`readlink -f`, so a value such as
# `/tmp/myflow-uitest/../../etc` matches the pattern while naming `/etc`.
# That is not reachable today: every caller of this macro passes one of the
# `override`-protected literals declared above, none of which contains `..`.
# But the guarantee this macro provides is exactly "the literal string
# starts with `/tmp/myflow-uitest`", not "the path resolves under
# `/tmp/myflow-uitest`" -- if a future edit ever passes this macro a path
# built from something other than those literals, re-derive that guarantee
# rather than assume it still holds.
#
# `override define` (not plain `define`): a `define`d variable is an
# ordinary recursive variable to Make, so without `override` a
# command-line assignment replaces this entire macro body, `case` and all
# -- confirmed with
# `make -n ui-test-down 'UITEST_SAFE_RMRF=@echo INJECTED $(1)'`, which
# without `override` prints the injected text in place of the safety
# `case` block. `override` on a `define` is valid GNU Make and is the
# only thing that stops that substitution; `scripts/check-uitest-overrides.sh`
# checks `define`-declared UITEST_* names for it the same way it checks
# `:=` assignments.
override define UITEST_SAFE_RMRF
for p in $(1); do \
	case "$$p" in \
		/tmp/myflow-uitest*) rm -rf "$$p" ;; \
		*) echo "refusing to rm -rf unexpected path: $$p" >&2; exit 1 ;; \
	esac; \
done
endef

# ui-test-up resets the UI-test stack: drop and recreate myflow_uitest in
# the running myflow-postgres container, start myflowd against it on its
# own port, wait for the daemon to answer, then seed a fixed fixture
# (cmd/uitest-seed). Resetting on every bring-up -- rather than reusing
# whatever the database already holds -- is the requirement itself, not
# an optimisation: a test stack whose contents persist between sessions
# accumulates the same way the live one did.
#
# MYFLOW_TRANSCRIPTS_DIR and MYFLOW_STATE_DIR are pointed at freshly
# emptied temporary directories -- load-bearing, not tidiness. myflowd's
# transcript harvester scans MYFLOW_TRANSCRIPTS_DIR every 5 seconds
# (harvestInterval, cmd/myflowd/main.go); left on the real transcripts
# root, myflow_uitest would refill itself with the operator's genuine
# session data within seconds of starting, defeating the entire point of
# a stack separate from the live one.
#
# Two separate mechanisms below, each covering a different failure --
# neither is redundant with the other:
#
# - `dropdb --force` (Postgres' WITH (FORCE)) is what makes the DROP
#   itself succeed against any live connection on myflow_uitest,
#   including one this Makefile did not start -- a stray psql, another
#   shell, an operator's own client. Confirmed by dropping the database
#   out from under a running daemon with --force alone: the drop
#   succeeded immediately. internal/store/testsupport_test.go's own
#   per-test-database teardown uses the same escape hatch.
#
# - The previous daemon is still killed BEFORE the drop, but for the
#   PORT, not the drop: a daemon whose database was just force-dropped
#   out from under it does not exit on its own -- confirmed by watching
#   it keep running, still holding :4174 and answering 500s, for two
#   minutes after its database was dropped. A second myflowd then fails
#   to bind :4174 ("address already in use") and this target aborts
#   with the first daemon orphaned. Killing before the drop also means
#   the drop is not racing a daemon that is still writing.
ui-test-up: web-build check-uitest-overrides
	@if [ -f $(UITEST_PIDFILE) ] && kill -0 "$$(cat $(UITEST_PIDFILE))" 2>/dev/null; then \
		pid="$$(cat $(UITEST_PIDFILE))"; \
		echo "ui-test-up: stopping previous daemon $$pid"; \
		kill "$$pid" 2>/dev/null; \
		for i in $$(seq 1 30); do \
			kill -0 "$$pid" 2>/dev/null || break; \
			sleep 1; \
		done; \
		if kill -0 "$$pid" 2>/dev/null; then \
			echo "ui-test-up: previous daemon $$pid did not exit within 30s -- still holding :$(UITEST_PORT)" >&2; \
			exit 1; \
		fi; \
	fi
	rm -f $(UITEST_PIDFILE)
	@echo "ui-test-up: dropping and recreating $(UITEST_DB)"
	# Both halves reuse scripts/workspace.sh: `remove --force` for the
	# drop (--force is the whole point here, see the comment above this
	# target -- it is an opt-in flag on the shared script, not an inline
	# duplicate of it) and `create` for the recreate. Same container, same
	# superuser, one place each fact lives, and by the time create runs
	# the database has just been dropped, so its own "no-op if it already
	# exists" idempotency never engages -- it behaves as a plain createdb.
	../scripts/workspace.sh remove $(UITEST_ID) --force
	../scripts/workspace.sh create $(UITEST_ID)
	@$(call UITEST_SAFE_RMRF,$(UITEST_TRANSCRIPTS_DIR) $(UITEST_STATE_DIR))
	mkdir -p $(UITEST_TRANSCRIPTS_DIR) $(UITEST_STATE_DIR)
	go build -o $(UITEST_BIN) ./cmd/myflowd
	MYFLOWD_PORT=$(UITEST_PORT) MYFLOWD_DSN="$(UITEST_DSN)" \
		MYFLOW_TRANSCRIPTS_DIR=$(UITEST_TRANSCRIPTS_DIR) \
		MYFLOW_STATE_DIR=$(UITEST_STATE_DIR) \
		$(UITEST_BIN) > $(UITEST_LOG) 2>&1 & echo $$! > $(UITEST_PIDFILE)
	@echo "ui-test-up: waiting for myflowd to answer on :$(UITEST_PORT)"
	@for i in $$(seq 1 60); do \
		curl -sf http://127.0.0.1:$(UITEST_PORT)/api/v1/changes >/dev/null 2>&1 && exit 0; \
		sleep 1; \
	done; \
	echo "ui-test-up: myflowd did not answer on :$(UITEST_PORT) within 60s -- see $(UITEST_LOG)" >&2; \
	exit 1
	MYFLOWD_DSN="$(UITEST_DSN)" go run ./cmd/uitest-seed
	@echo "ui-test-up: ready -- export MYFLOW_ADDR=http://127.0.0.1:$(UITEST_PORT) to point a session at it"

# ui-test-down stops the daemon and drops myflow_uitest, both
# idempotently: a second `make ui-test-down` exits 0 rather than erroring
# on an absent pidfile or an already-dropped database.
#
# The drop reuses `../scripts/workspace.sh remove` rather than an inline
# dropdb: unlike ui-test-up's drop, this one carries no --force -- it
# relies on the daemon actually being gone by the time it runs, which is
# why the SIGTERM below is followed by the same bounded wait ui-test-up's
# stop-previous-daemon step uses (myflowd's shutdown grace is 15s;
# 30 one-second polls covers that with margin). Without the wait, an
# unforced `dropdb` can race a daemon that is still holding its
# connection open and fail intermittently.
ui-test-down: check-uitest-overrides
	@if [ -f $(UITEST_PIDFILE) ]; then \
		pid="$$(cat $(UITEST_PIDFILE))"; \
		if kill -0 "$$pid" 2>/dev/null; then \
			kill "$$pid" 2>/dev/null; \
			for i in $$(seq 1 30); do \
				kill -0 "$$pid" 2>/dev/null || break; \
				sleep 1; \
			done; \
			if kill -0 "$$pid" 2>/dev/null; then \
				echo "ui-test-down: daemon $$pid did not exit within 30s -- dropping myflow_uitest anyway" >&2; \
			fi; \
		fi; \
		rm -f $(UITEST_PIDFILE); \
	fi
	../scripts/workspace.sh remove $(UITEST_ID)
	@$(call UITEST_SAFE_RMRF,$(UITEST_TRANSCRIPTS_DIR) $(UITEST_STATE_DIR) $(UITEST_LOG) $(UITEST_BIN))
