**Severity:** Important
**Slot:** gated per-task reviewer (task 4)
**Location:** stats/internal/guard/livestorecopy.go:811

`stats/internal/guard/livestorecopy.go:811-828` (`copyStore`) — when `psql` exits before `pg_dump` has written everything (its `ON_ERROR_STOP=1` on the first failing statement, a refused connection, any early exit), `live-store-copy up` hangs forever instead of failing with exit 1. `dump.StdoutPipe()` leaves the read end of the pipe open in the guard process for as long as `dump.Wait()` has not run; once `restore` has exited nothing reads that pipe, so `pg_dump` (the `docker exec` client) blocks on a full pipe buffer and never gets the SIGPIPE the bash pipeline gave it, and `dump.Wait()` waits on it for ever. The bash it replaces exits 1 with "copy into flow_<name>_live failed" in the same situation. Any real `flow` dump is far larger than the 64 KB pipe buffer.
