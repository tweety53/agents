#!/usr/bin/env bash
# workspace.sh — create, remove and report this repository's per-workspace
# database, the `create`/`remove`/`survivors` triple named in
# .flow/project.md's `## workspace isolation` section per
# skills/flow-contracts/project-configuration.md (canonical for what each
# command runs, what `survivors` prints, and what its exit code means).
#
# Usage:
#   scripts/workspace.sh create <id>
#   scripts/workspace.sh remove <id> [--force]
#   scripts/workspace.sh survivors <id>
#
# `remove`'s `--force` is opt-in and additive: omitted, `remove` runs exactly
# as it always has (`dropdb --if-exists`, no --force). Passed, it adds
# Postgres' WITH (FORCE) to the drop, which succeeds against a database with
# a live connection this process did not start. It changes nothing about id
# validation or the refusal to operate on the shared "flow" database below
# — both run identically whether or not --force is passed, so an id that
# would be rejected today is rejected exactly the same way with --force.
#
# `--force` also refuses any id whose database name does not end in
# "_uitest" -- the same suffix rule cmd/uitest-seed/guard.go enforces on its
# own destructive path (specs/myflow-ui-test-stack/spec.md's "Destructive
# test-stack paths refuse to act on any other database"). This is defence
# in depth, not the only guard: the caller (stats/Makefile's ui-test-up) is
# still responsible for passing the right id, since this script has no way
# to know a caller's *intent* -- only a second, independent check on the
# one attribute that distinguishes the UI-test database from every other
# workspace database this script can otherwise remove. As of this writing
# every `--force` caller in this repository targets the UI-test stack, so
# this refuses nothing anyone currently needs; a future caller that
# legitimately needs a forced drop of a non-uitest workspace would need
# this rule revisited, not silently bypassed.
#
# The database service is the flow-postgres container (superuser flow,
# database flow, host port 5433). Every subcommand reaches it through
# `docker exec`, since this machine has no host-installed psql/createdb/
# dropdb client — the container's own copies are used instead, for `create`
# and `remove` as well as `survivors`, so there is exactly one mechanism for
# reaching the service rather than a host path for two commands and a
# container path for the third.
#
# A workspace database is named flow_<id_underscored> (every "-" in the id
# replaced by "_"), per skills/flow-contracts/workspace-isolation.md's
# "What the id derives". <id> is validated against that section's own output
# alphabet ([a-z0-9-]) before it is ever interpolated into SQL or a database
# name — not because a malformed id is expected, but because both create and
# remove pass it into a query pipefail cannot audit after the fact.
#
# `create` is a no-op when the database already exists, so a re-run of the
# pipeline in an existing worktree is not an error. `create` and `remove`
# both refuse an empty or missing id rather than operating on the shared
# "flow" database.
#
# `survivors` is read by scripts/check-cleanup-complete.sh under the rules in
# project-configuration.md's "What `survivors` prints, and what its exit
# code means":
#   - exit 0, empty stdout  -> nothing survived (the only verifying result)
#   - exit 0, non-empty     -> one surviving database name per line
#   - any non-zero exit     -> the check could not run; reported as a skip
# The filter is `awk`, never `grep`: `grep` exits 1 on no match, and that
# exit is read here as "the check could not run" rather than "nothing
# survived", so a grep-filtered empty report could never reach the one
# result that verifies cleanup. The command runs under `pipefail` so a
# failing `docker exec` is reported as that non-zero skip rather than
# swallowed into an empty-and-therefore-verified report, and the pipeline
# does not end in `head`, which would turn a real survivor list into a
# SIGPIPE-truncated report.
#
# `survivors` also carries its own timeout INSIDE the container: the
# caller's 60-second bound (project-configuration.md's "The one-minute
# bound") terminates a process group on the machine running the guard, and
# `docker exec`'s real work happens inside the container, outside that
# group — so a hung query there would outlive the bound undetected. A
# 5-second connect timeout and a 5-second statement timeout on the client
# running inside the container end it instead.

set -o pipefail

CONTAINER="flow-postgres"
SUPERUSER="flow"

usage() {
  echo "usage: $0 {create|remove|survivors} <id> [--force]" >&2
  echo "       --force is accepted only by remove" >&2
  exit 2
}

cmd="${1:-}"
id="${2:-}"
force_flag="${3:-}"

[ -n "$cmd" ] || usage
if [ -z "$id" ]; then
  echo "ERROR: a workspace id is required (refusing to operate on the shared $SUPERUSER database)" >&2
  exit 2
fi
case "$id" in
  *[!a-z0-9-]*)
    echo "ERROR: workspace id '$id' is not [a-z0-9-]; refusing to build a database name from it" >&2
    exit 2
    ;;
esac

force=0
if [ -n "$force_flag" ]; then
  if [ "$cmd" = "remove" ] && [ "$force_flag" = "--force" ]; then
    force=1
  else
    echo "ERROR: unrecognized argument '$force_flag'" >&2
    usage
  fi
fi

db="flow_${id//-/_}"

if [ "$force" -eq 1 ]; then
  case "$db" in
    *_uitest) ;;
    *)
      echo "ERROR: --force refused: database '$db' does not end in _uitest" >&2
      exit 2
      ;;
  esac
fi

case "$cmd" in
  create)
    exists="$(docker exec "$CONTAINER" psql -U "$SUPERUSER" -d "$SUPERUSER" -tAc \
      "select 1 from pg_database where datname = '$db'")"
    rc=$?
    if [ "$rc" -ne 0 ]; then
      echo "ERROR: could not reach $CONTAINER to check for $db" >&2
      exit "$rc"
    fi
    if [ "$exists" = "1" ]; then
      exit 0
    fi
    docker exec "$CONTAINER" createdb -U "$SUPERUSER" "$db" || {
      rc=$?
      echo "ERROR: createdb failed for $db" >&2
      exit "$rc"
    }
    echo "created $db"
    ;;
  remove)
    if [ "$force" -eq 1 ]; then
      docker exec "$CONTAINER" dropdb -U "$SUPERUSER" --if-exists --force "$db" || {
        rc=$?
        echo "ERROR: dropdb --force failed for $db" >&2
        exit "$rc"
      }
    else
      docker exec "$CONTAINER" dropdb -U "$SUPERUSER" --if-exists "$db" || {
        rc=$?
        echo "ERROR: dropdb failed for $db" >&2
        exit "$rc"
      }
    fi
    ;;
  survivors)
    docker exec -e PGCONNECT_TIMEOUT=5 -e PGOPTIONS="-c statement_timeout=5000" \
      "$CONTAINER" psql -U "$SUPERUSER" -d "$SUPERUSER" -tAc \
      "select datname from pg_database where datname = '$db'" \
      | awk 'NF { print $0 }'
    ;;
  *)
    usage
    ;;
esac
