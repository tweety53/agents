# Manual test — kan-15-parallel-myflow-do-task-lanes

This change spans three repositories, exercised differently.

**agents** has no runnable application — it is the source of the myflow skills, commands and rules.
Its checks are commands to run:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes
```

**gymie** is the application. Its checks run its stack from the apply worktree:

```bash
cd /Users/tweety53/Projects/gymie-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes
```

**gymie-frontend** contributes one file — the dev server's port override — and is exercised through
gymie's stack rather than on its own:

```bash
cd /Users/tweety53/Projects/gymie-frontend/.worktrees/openspec-kan-15-parallel-myflow-do-task-lanes
```

This worktree resolves to workspace **`kan-15-55a6`** — database `gymie_kan_15_55a6`, bucket
`gymie-media-kan-15-55a6`, cache index 1, backend **11351**, gateway **11350**, admin panel
**6271**, KMP frontend **6270**. All four move. The main checkout keeps `gymie`, index 0,
`gymie-media` and 8080 / 8081 / 3000 / 3001.

**Both stacks can now run at once, frontends included.** The KMP frontend's port was the last
exception; it moves like the others since `gymie-frontend` gained a dev-server port override.

## The workspace contract (agents)

- [ ] check `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
      `scripts/check-plan-provenance.sh` and `scripts/check-task-build-green.sh` — all exit 0
- [ ] check every harness in `.myflow/project.md`'s `## test` list exits 0
- [ ] check `scripts/test-check-cleanup-complete.sh` covers the workspace row — its cases name the
      survivors command, the timeout, the pipeline status and the duplicated-heading skip
- [ ] check the id derivation agrees across all three implementations — run the fenced block in
      `skills/myflow-contracts/workspace-isolation.md` under `LC_ALL=C` and under a UTF-8 locale and
      confirm one answer
- [ ] check a sandboxed install ships the contract:
      `SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global` then look for
      `workspace-isolation.md` under `$SANDBOX/.claude/skills/myflow-contracts/`

## Cleanup verification (agents)

- [ ] check the guard reports this change's live resources as leftovers —
      `scripts/check-cleanup-complete.sh <gymie worktree> kan-15-parallel-myflow-do-task-lanes /Users/tweety53/Agents/myflow/state/gymie-7c1f238a`
- [ ] check a change with nothing to find reports `COMPLETE:` and an empty survivor report
- [ ] check a project declaring no isolation is unaffected — the same guard against
      `/Users/tweety53/Projects/agents` says nothing about a workspace
- [ ] check a service that cannot be reached is reported as **skipped**, not failed, and does not
      block the terminal state

## Workspace resolution (gymie)

- [ ] check `./gradlew devWorkspace` in the worktree prints the id, database, bucket, cache index and
      the port block
- [ ] check `./gradlew devWorkspace -PworkspaceChange=` prints the main checkout's defaults unchanged
- [ ] check the whole port block is discarded together when any one of its ports is held — bind
      11350 and confirm backend and admin move too, rather than only the clash being repaired
- [ ] check all four ports move in a worktree, and none of them does in the main checkout

## Running two changes at once (gymie)

- [ ] check the worktree's stack starts while the main checkout's is running — `./gradlew devStart`
      in the worktree, **pointing `-PfrontendRoot` at the gymie-frontend worktree**, not the main
      frontend checkout, which has no port override and is refused by design
- [ ] check that pointing `-PfrontendRoot` at the **main** frontend checkout is refused with an
      explanation, before anything starts — that is the version-skew guard
- [ ] check both backends answer at once — the main checkout on 8081 and the worktree on 11351
- [ ] check `docker exec gymie-postgres psql -U gymie -d postgres -c '\l'` lists `gymie` **and**
      `gymie_kan_15_55a6`
- [ ] check `docker exec gymie-redis redis-cli CLIENT LIST` shows clients on two different `db=`
      indices
- [ ] check the object store holds `gymie-media` **and** `gymie-media-kan-15-55a6`
- [ ] check `docker ps` lists the **same five containers as before** — no new ones
- [ ] check the worktree's admin panel reaches the worktree's gateway, not the main checkout's
- [ ] check `./gradlew devStop` in the worktree leaves the main checkout's processes alone
- [ ] check **both KMP frontends serve at once** — the main checkout's on 3000 and the worktree's
      on its own port — which is the check that was impossible before this change
- [ ] check `./gradlew devStop` in the worktree leaves the main checkout's frontend on 3000 running
- [ ] check `./gradlew devStop` in the main checkout leaves the worktree's frontend running

## Starting and stopping named services (gymie)

`-Pservices=<comma-separated names>` takes any subset of `backend`, `gateway`, `frontend`,
`admin-frontend`. It works in the worktree **and in the main checkout** — it needs no workspace.

- [ ] check `./gradlew devStart` with no flag still starts all four, exactly as before
- [ ] check several names in one command — `./gradlew devStart -Pservices=backend,gateway` acts on
      both, which is the whole point of the flag
- [ ] check `./gradlew devStart -Pservices=backend` restarts only the backend, and the frontend is
      not recompiled — that is the minute-scale saving the flag exists for
- [ ] check the services it did not name keep the ports they were on: `./gradlew devWorkspace`
      reports the same block before and after the partial start
- [ ] check `./gradlew devStop -Pservices=frontend` stops only the frontend
- [ ] check `./gradlew devStatus -Pservices=backend,frontend` reports only those two
- [ ] check `./gradlew devRestart -Pservices=admin-frontend` does what `devStart` with the same flag
      does
- [ ] check names in any order and in any case resolve the same, and are still acted on in dependency
      order — backend first on the way up, reverse on the way down
- [ ] check `-Pservices=admin` is refused and the refusal names `admin-frontend` and the other three
- [ ] check `-Pservices=` is refused rather than read as all four
- [ ] check `./gradlew devStart -Pservices=backend` does **not** demand a frontend checkout — point
      `-PfrontendRoot` at a path that does not exist and confirm it still starts

## The recorded port block (gymie)

A workspace **writes down** the ports it resolved, in `.dev-stack/port-block`, and a partial start
reads that record rather than working out again where its siblings are. Deleting the file restores
the deterministic block.

- [ ] check a full `./gradlew devStart` writes `.dev-stack/port-block`, one `service=port` line per
      service
- [ ] check a partial start reads it — the services it leaves alone keep exactly the ports the file
      names, and the admin panel still reaches the running gateway
- [ ] check the whole block is still discarded together when the deterministic block is taken, and
      that the record then holds the discovered ports rather than the deterministic ones
- [ ] check deleting `.dev-stack/port-block` puts the next full start back on the deterministic
      block, and that the README says so where an operator would look
- [ ] check a record naming only some of the services is ignored rather than half-applied
- [ ] check `./gradlew devWorkspaceRemove` removes the record along with the database, bucket and
      cache index

## What the stack does with a record it cannot use (gymie)

The dev stack keeps a pid and a port per service under `.dev-stack/`. A run that cannot read one now
**warns and carries on** — it only refuses to *delete* what it could not read. Make a record
unreadable with `chmod 000`, or write text into it, and restore it afterwards.

- [ ] check a corrupt record for one service does not stop `./gradlew devStart` from running — it
      warns and proceeds, which is what it did before this change
- [ ] check every unusable record is named in one message, not just the first
- [ ] check the advice differs by cause: a file that cannot be read points at permissions and disk
      space, a file holding garbage says the directory is fine and tells you to delete it
- [ ] check `./gradlew devStop` does not delete a pid or port record when the process it tried to
      kill is still alive — and that it says which one survived and how to find it
- [ ] check a service whose pid file is briefly unreadable during startup is not reported as having
      exited — the health check retries instead
- [ ] check a second `./gradlew devStart -Pservices=backend` fired while the first is mid-restart is
      **refused**, naming the port and the service, rather than killing the holder
- [ ] check `chmod 000 .dev-stack/port-block` while the stack is up, then `./gradlew devStart` — it
      refuses, names the services it cannot account for, and **does not move the block**
- [ ] check `./gradlew devStatus` prints the block, its shape and any staleness notice, and still
      answers with the data services stopped

## The same flag in the main checkout (gymie)

- [ ] check all of the above again in `/Users/tweety53/Projects/gymie`, where there is no workspace —
      the flag must not depend on `-PworkspaceChange` or on being in a worktree
- [ ] check `./gradlew devStart -Pservices=backend` there rebinds 8081 and leaves 8080, 3000 and 3001
      untouched
- [ ] check the main checkout's ports are still the declared ones, not a discovered block

## Cache index claims (gymie)

- [ ] check two workspaces starting in turn claim **different** cache indices
- [ ] check a second `devStart` for the same workspace reuses its own index rather than claiming
      another
- [ ] check `./gradlew devWorkspaceInventory` lists every workspace's database, bucket and claimed
      index, and names an index holding keys that nothing claims
- [ ] check `./gradlew devWorkspaceRemove -PworkspaceId=<id>` releases the claim as well as
      dropping the database and bucket
- [ ] check `./gradlew devCacheIndexReclaim -PcacheIndex=<n>` refuses without `-Pforce`, and
      refuses a claimed index outright

## Workspace creation and database tasks (gymie)

- [ ] check the first start creates the database and bucket and says so once
- [ ] check a second start reuses them and prints no creation notice
- [ ] check a newly created database is migrated and seeded by the normal path, not copied
- [ ] check `dbReset` in the worktree targets the workspace's database, and that
      `-PconfirmDb=gymie` — the main checkout's habit — now refuses
- [ ] check `dbSeed` in the worktree seeds the workspace's database and leaves `gymie` untouched
- [ ] check `-PpgDb` still overrides both

## Backwards compatibility (gymie)

- [ ] check `git diff develop -- docker-compose.yml` shows no diff at all
- [ ] check the main checkout still binds 8080, 8081, 3000 and 3001 against `gymie`, index 0 and
      `gymie-media`
- [ ] check `./gradlew test` passes — the Testcontainers suite is untouched
- [ ] check `./gradlew ktlintCheck detekt` exits 0
- [ ] check OAuth sign-in still works in the main checkout, where the redirect URI is unchanged

## Behaviour worth knowing before you run it

- **`devStart` now waits for the frontend to serve**, up to 600s, where it previously returned as
  soon as the process was alive. A cold Kotlin/JS compile plus a yarn install is genuinely slow, so
  a first run reports success later than it used to.
- **A frontend refusal throws mid-run**, after the backend and gateway have started. Re-running
  after fixing `-PfrontendRoot` completes the stack rather than restarting it, since an
  already-running service is skipped.

- **`devWorkspace` now goes through the port-block decision.** With this checkout's own services
  running on its block it reports the deterministic block rather than discovering a new one. That is
  the correct answer and it is a change from today, where any listener made it report discovered
  ports.

## Known incomplete

- **No run has had both full stacks up at the same moment.** Every other check here was performed
  during implementation, and isolation was evidenced by reading Postgres, Redis and MinIO directly —
  two databases and two buckets coexist right now, and a worktree backend was seen answering on its
  own port while the shared one was in use. The blocker was that the worktree's stack used to take
  port 3000; **that is fixed**, so the checks under **Running two changes at once** are now
  genuinely runnable and are the ones this leaves open.
- **The frontend port override has not been exercised through `devStart`.** It was verified directly
  — a socket bound on a spare port, and the generated webpack config read back — but the real path
  crosses one more hop than the probe did: gymie's Gradle, then a `bash -lc ./gradlew` child, then a
  possibly-warm second daemon, then webpack. Environment propagation was measured in an isolated
  project and holds both ways; the end-to-end run needs a live stack. **Start the worktree's stack
  twice with two different ports and read the bound port from the socket.**
- **The main checkout's `devStop` now sweeps port 3000 unconditionally** when nothing is recorded,
  as it already did for 8080, 8081 and 3001. Correct — only the main checkout can have bound it —
  but 3000 is a more common default across unrelated tools than the other three, so a hand-started
  process there is now swept where previously it was not.
- **A partial start refuses when a port it needs is held by something it is not stopping.** An
  earlier revision of this round killed the holder instead, on the argument that a held port always
  belongs to a service the run is about to stop. That argument was wrong — any process can bind any
  port — and the review caught it, so the refusal the plan asked for is what shipped. Worth checking
  the refusal is legible when you meet it.
- **The dev stack now refuses in several new situations, and none has been seen by a human.** A run
  that cannot read a pid or port record refuses instead of guessing; a run that cannot locate a
  running service it is leaving alone refuses rather than publishing a port that service is not on.
  Both are the right answer to a real defect — the previous behaviour silently leaked a process, or
  wrote a wrong port into the admin panel's generated config. But **a refusal that fires too readily
  is its own problem in a dev tool**, and the fix round that added them says plainly that it could
  not measure how often the second one trips during an ordinary slow boot. Watch for a refusal you
  did not deserve, and say so — that is the judgement no reviewer could make.
- **No `dev*` task was run for the `-Pservices` work — not once, by anyone.** Every implementer on
  that group was forbidden from running `devStart`, `devStop`, `devRestart` and `devStatus`, because
  the operator's own stack was live and those commands would have acted on it. The decisions are
  covered by 21 unit tests and the wiring was read rather than executed, so **every check under
  *Starting and stopping named services* and *The same flag in the main checkout* is genuinely
  untried.** They are the ones to run first.
- **The cache claim protocol has not been exercised under genuine contention.** Two review passes
  repaired it — the claim was invisible to other workspaces, then one workspace could hold two
  indices, then a poisoned index could not be reclaimed. All fixed and unit-tested, but the race is
  modelled with a fake; no run has had two workspaces claiming against a real cache at once.
