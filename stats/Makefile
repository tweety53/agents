.PHONY: test web-test web-build build up down

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
test: web-build
	go test ./... -race -count=1

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
