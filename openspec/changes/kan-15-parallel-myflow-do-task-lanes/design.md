## Context

Several myflow changes can already run at once: separate sessions, separate worktrees, separate
branches, separate state files. What they cannot do is **run the application**.

Two apply worktrees share one Postgres database. One change's Flyway migrations alter the schema the
other is testing against, and the result looks like a real test failure rather than a collision. That
is the failure KAN-15's own addition names: *"when db schema is different between them it can produce
false positive testing results"*.

**The problem was never the whole stack.** It is that one change's Flyway migrations alter the schema
another change is testing against, and the result looks like a real test failure rather than a
collision. Two adjacent resources fail the same way once changes genuinely run at once — Redis holds
sessions and cache in one index, and MinIO has a single bucket. And the applications bind fixed
ports, so a second `devStart` cannot start at all, whatever the database is called.

This design namespaces the **logical resource inside each shared service**, and leaves the services
themselves alone.

Five facts were read out of the repository rather than assumed.

**`application.yml` already parameterises the datasource** — `url: ${DB_URL:jdbc:postgresql://localhost:5432/gymie}`
— but **`application-local.yml` hardcodes** `jdbc:postgresql://localhost:5432/gymie` with no
placeholder, and `local` is the profile development runs. The local profile therefore overrides the
hook that already exists, and that one line is the pivot of the whole gymie change.

**`dbReset` already takes `-PpgDb`**, defaulting to `gymie`, and already runs
`DROP DATABASE IF EXISTS … WITH (FORCE)` then `CREATE DATABASE "$pgDb"`, correctly quoted. The
per-database machinery exists; it is only ever pointed at one name.

**The gateway has no datasource.** Only the backend touches Postgres, which halves the surface.

**`dbSeed` starts `:app:bootRun --spring.profiles.active=local`**, so it inherits exactly the
hardcoded URL above.

**Gymie's integration tests use Testcontainers**, so the test database is already ephemeral and
per-JVM. The test suite is untouched.

## Goals / Non-Goals

**Goals:**

- Stop one change's migrations from reaching another change's database.
- Let two changes run their applications at the same time.
- Leave the main checkout behaving exactly as it does today.

**Non-Goals:**

- **In-run parallelism.** `/myflow-do` stays sequential. The lane mechanism designed for this change
  was rejected; see the superseded decisions below.
- **Containerizing the applications.** Designed in full and rejected; also superseded below.
- Changing `docker-compose.yml`, the production images, the helm chart, or the test suite.
- Touching the **admin** frontend repository. The KMP frontend repository is touched, in one
  one-file change — see `kmp-frontend-port-rotates`.
- Isolating mailpit.

## Decisions

### In-run lanes are not built; concurrency comes from running several changes at once

**ID:** no-in-run-lanes
**Status:** active
**Chosen:** drop the lane mechanism entirely — no `## Lanes` table, no `**Lane:**` tag, no lane
guard, no dependency graph, no rebase, no join. `/myflow-do` stays sequential, and the wall-clock
saving comes from the operator running several changes concurrently in separate sessions, which the
pipeline already supports once the stack is isolated.
**Considered:**
- *In-run lanes joined by patch* — the design this supersedes. It bought concurrency inside one run
  at the cost of a declaration format, a guard, a dependency graph, a rebase policy, a join with its
  own conflict policy, and two new classes of temporary artifact. Rejected as more complexity than
  the problem warrants.
- *Lanes as separate branches with a PR each* — an intermediate position, also superseded. Rejected
  once it was clear that a lane with its own branch, worktree, PR, review panel, human gate and
  finish run **is** a change, so the lane concept only adds a second name for something the pipeline
  already has.

### The applications stay host processes

**ID:** apps-stay-host-processes
**Status:** active
**Chosen:** leave the applications running as host processes and isolate them by offsetting their
ports, rather than moving them into containers. The problem to solve is a shared database, and a
containerized dev stack is a far larger change than that problem justifies — four dev Dockerfiles,
bind mounts, volume-backed caches, a second run mode, and an unproven Kotlin/JS dev server.
**Considered:** *The dev stack in containers*, designed in full and superseded below. It would have
removed the fork-orphan class outright, which is a genuine benefit, at the cost of rewriting the
daily development loop for four applications across three repositories. The orphan handling stays
where it is instead.

### `docker-compose.yml` is not touched at all

**ID:** compose-untouched
**Status:** active
**Chosen:** leave the container-runtime configuration byte-identical. The data services stay shared
containers on their current ports and names; what gets namespaced is the database, the cache index
and the bucket **inside** them.
**Considered:** *Parameterizing the compose project name, container names and host ports*, which
every earlier revision of this design did. It duplicates whole services per workspace — a second
Postgres, Redis, MinIO, pgAdmin and mailpit — to solve a problem that lives one level down, inside
one of them.

### Isolate the logical resource, not the service

**ID:** isolate-logical-resources-not-services
**Status:** active
**Chosen:** one database, one cache index and one bucket per workspace, inside the shared services.
This is the smallest thing that fixes the failure, and it costs no extra containers, no extra data
ports and no extra memory.
**Considered:** *A full service stack per workspace*, which isolates more thoroughly — a wedged
Postgres in one workspace cannot affect another — at several times the resource cost, and requires
the compose changes the decision above rejects.

### Isolation is automatic in an apply worktree

**ID:** isolation-is-automatic-in-a-worktree
**Status:** active
**Chosen:** a worktree gets its own database, cache index, bucket and port block without asking; the
main checkout always gets the defaults. **Nothing is opt-in**, because the failure being prevented is
one that forgetting a flag reproduces silently — and a silent wrong answer is exactly what this
change exists to remove.
**Considered:**
- *Explicit request only.* Nothing changes for anyone who does not ask — and forgetting reproduces
  the original bug with no warning.
- *Automatic with an opt-out flag.* Reasonable, and rejected only as unnecessary surface: a worktree
  that wants the shared database can pass the existing `-PpgDb`.

### The cache index is probed, not derived

**ID:** redis-index-probed-not-derived
**Status:** active
**Chosen:** `devStart` probes which Redis indices are in use and claims the lowest free one,
recording it for that run. Every other value in this design is derived from the digest; this one
cannot be, because Redis offers **16** indices and a digest modulo 16 collides at roughly 6% for two
concurrent workspaces — silently sharing sessions and cache, which is the failure class this change
removes.
**Considered:**
- *Derive it from the digest.* Deterministic and stable across sessions, consistent with the database
  name and bucket — and wrong one time in sixteen, silently.
- *Derive it and refuse to start on a collision.* Never silently wrong, but it blocks a legitimate
  second workspace outright, with renaming the change as the only remedy.

The cost is that the index is not stable across restarts. That does not matter: Redis holds sessions
and cache, and both are disposable.

### A new workspace database starts empty

**ID:** database-empty-then-migrate
**Status:** active
**Chosen:** create the database empty, let Flyway migrate it and `dbSeed` seed it — the same path a
fresh machine and CI take, so a workspace database is never subtly different from a clean one.
**Considered:** *`CREATE DATABASE … TEMPLATE gymie`.* Instant, and carries existing data — but
Postgres refuses a template copy while any session is connected to the source, so it fails whenever
anything is running against `gymie`, and it copies that database's migration state, which may not
match the branch.

### The database name is underscored; the bucket takes the id verbatim

**ID:** database-name-underscored
**Status:** active
**Chosen:** `gymie_kan_15_55a6` for the database, `gymie-media-kan-15-55a6` for the bucket. `dbReset`
already quotes its identifier, so a dashed database name would work today — but an unquoted call site
added later would break on it, and underscores cost nothing. Bucket names are not SQL identifiers, so
they take the id as it is spelled.
**Considered:** *Dashes throughout*, matching the id and the branch name exactly, relying on every
present and future call site quoting correctly. *The bare id with no project prefix*, which is
shortest but leaves `psql -l` listing workspace databases with nothing tying them to the project.

### Database tasks default to the workspace's own database

**ID:** db-tasks-default-to-workspace
**Status:** active
**Chosen:** `dbReset` and `dbSeed` default to the workspace's database in a worktree, and to `gymie`
in the main checkout. Consistent with everything else the worktree does, and safer:
`dbReset -Pforce` from a worktree destroys that change's data rather than the shared database.
**Considered:** *Always defaulting to `gymie` unless `-PpgDb` is passed.* Unchanged behaviour and
nothing to relearn — but it makes the more damaging mistake the easier one to make.

### A failed removal at finish is skipped, not failed

**ID:** failed-drop-is-skipped-not-failed
**Status:** active
**Chosen:** if Postgres is not running when `/myflow-finish` run 2 reaches cleanup, report the
skipped removal and continue to `FINISHED`. That matches how the existing project-supplied stop check
already treats an unreachable stack, and a stale database costs a few megabytes.
**Considered:** *Treating it as a leftover that blocks the `FINISHED` write*, which is what every
other cleanup row does and is the more consistent answer. Rejected because it would block finishing
an already-merged change over a container that happens to be stopped — a cost out of proportion to
the leftover.

### The KMP frontend is not port-isolated

**ID:** kmp-frontend-not-port-isolated
**Status:** superseded by `kmp-frontend-port-rotates`
**Chosen:** leave the KMP frontend on port 3000 and isolate the other three applications. Its dev
server's port is fixed in `gymie-frontend/webApp/build.gradle.kts` at the webpack config level, with
no environment override to reach from outside that repository — and this change does not touch
either frontend repository. So the isolation table carries no port row for it, and `FRONTEND_URL`
keeps its default rather than following a port block it cannot move.
**Cost, stated rather than glossed:** two changes can run their backends, gateways and admin panels
at once, and only one can run the KMP frontend. That is the same class of stated limitation as
`mailpit-stays-shared`, and it is recorded here and in the project's own declaration so nobody
discovers it by finding the second `devStart` already holding 3000.
**Considered:** *Making `devServer.port` read an environment variable*, which isolates the fourth
application and is a two-line edit — but in a third repository, which means a third worktree, a
third review panel and a third pull request, and rewriting the "neither frontend repository is
touched" constraint stated in `proposal.md`, this file and `tasks.md`. Rejected as a cost out of
proportion to the one application it buys, and reversible later by its own change.
**Discovered:** by task 2.1's re-review, after tasks 3.1 and 3.1b had landed. The plan had listed
four application ports since it was written, and nothing had established that one of the four was
unreachable from the two repositories the change is allowed to touch.

### The KMP frontend's port rotates after all

**ID:** kmp-frontend-port-rotates
**Status:** active
**Supersedes:** `kmp-frontend-not-port-isolated`
**Chosen:** rotate the KMP frontend's port with every other application's, and **delete the
shared-port ownership machinery** built to make the collision survivable. `webpack.config.d/` is
Kotlin/JS's own extension point for exactly this, the frontend repository already carries a file
there that mutates `config.devServer`, and `devStart` already passes its environment to what it
starts — so the override is one small file beside an existing one.
**Why the decision it supersedes was wrong, stated precisely, because the error is the interesting
part:** that decision reads *"its dev server's port is fixed in `gymie-frontend/webApp/build.gradle.kts`
at the webpack config level, **with no environment override to reach from outside that
repository**"*. The first half is true and the second is false. It was written from the build file
alone, and `webpack.config.d/` was never looked for — a claim about the absence of a mechanism, made
without searching for the mechanism.
**What the mistake cost.** 469 lines of `SharedPort.kt` and its tests, and 31 call sites.
<!-- measured: wc -l buildSrc/src/{main,test}/kotlin/com/gymie/gradle/SharedPort*.kt; grep -cE 'sharedPort|notStartedReason|portToSweep|claimServicePort' gradle/dev-lifecycle.gradle.kts @ the gymie apply worktree, staged -->
Plus a state file written **outside every repository** at `~/.gymie/dev-shared-ports/`, and **the
same Critical finding three times** — through the start path, the stop path, and the failure path.
<!-- measured: the F2, F17 and F26 rows of .superpowers/sdd/final-review-panel.md @ the agents apply worktree --> All of it existed to make one unmovable port polite, and none
of it made two changes able to run the frontend at once.
**Considered:** *keeping the machinery as a safety net* for a checkout running older code, which is
the residual the third fix round named as unclosable. Rejected: it keeps every line counted above
and a machine-wide state file guarding a collision that up-to-date checkouts can no longer have, and the
same residual is better answered by the port being genuinely free.
**The cost this accepts, in place of the one it removes:** the change now spans **three**
repositories rather than two — a third worktree, a third review and a third pull request — and
`proposal.md`, `tasks.md` and this file each stated the two-repository constraint by name and are
corrected together with this decision.

### Mailpit stays shared

**ID:** mailpit-stays-shared
**Status:** active
**Chosen:** leave mailpit alone. It is a development mail catcher with a single inbox and no natural
namespace, so two changes' emails interleave.
**Considered:** *A mailpit instance per workspace*, which would need the per-service duplication
`compose-untouched` rejects, for a resource whose failure mode is confusion rather than corruption.
Recorded as a stated limitation rather than solved.

### The applications run in containers, not as host processes

**ID:** apps-run-in-containers
**Status:** superseded by apps-stay-host-processes
**Chosen:** move all four applications — backend, gateway, KMP webApp and Next.js admin — into the
workspace's own container project, rather than isolating host processes by offsetting their ports.
A substantial part of `gradle/dev-lifecycle.gradle.kts` exists only because host processes fork
children that outlive their parent and squat ports; containers remove that problem class rather than
managing it, and make isolation uniform with the data services that are already containerized.
**Considered:**
- *Host processes with per-workspace port offsets* — the previously approved design, superseded here.
  It parameterized around the collision instead of removing it, and left the orphan-sweeping code
  doing the job it was written to do.
- *Containerizing only the JVM applications.* Halves the work, but the orphan comment in the
  lifecycle file names `npm run dev` explicitly, so the named failure would survive.

### Two run modes, with host as the default

**ID:** two-run-modes-host-default
**Status:** superseded by isolation-is-automatic-in-a-worktree
**Chosen:** keep the existing host-process path as the **default** and add container mode as an
explicit opt-in. The default loop stays fast and familiar, which is what makes it the easy thing to
test against; container mode is there when a change needs to run beside another. Host mode is
deliberately single-workspace and gets **no** per-workspace port offsets — building offset machinery
for host processes is exactly the work container mode exists to replace.
**Considered:**
- *Containers only, no host mode.* Would have allowed deleting the whole process-management layer,
  which was the original attraction. Rejected because it changes the daily loop for every
  invocation, including the ones that have nothing to do with concurrency.
- *Containers by default, host on request.* Same objection in a weaker form: the common case would
  pay for the rare one.
- *An explicit property every time, with no default.* Predictable, but a thing to remember on every
  invocation with no safe answer when it is forgotten.

### Container mode is opt-in, and isolation follows it

**ID:** container-mode-is-opt-in
**Status:** superseded by isolation-is-automatic-in-a-worktree
**Chosen:** a worktree gets an isolated stack only when container mode is explicitly requested —
never inferred from the presence of a worktree, a workspace id, or anything else. The workspace id,
the port block and the suffixed project name are all container-mode concepts.
**Considered:** *Inferring container mode from a worktree*, which is what
`every-apply-worktree-isolated` above assumed. Rejected because it makes the default loop depend on
where you happen to be standing. **The cost is real and is recorded rather than glossed:** two
changes run concurrently only if the operator asks for container mode, and running both in the
default host mode collides exactly as it does today, with no warning.

### Bind-mount the source; keep build state in named volumes

**ID:** source-bind-mount-volumes-for-build-state
**Status:** superseded by apps-stay-host-processes
**Chosen:** mount application source into the container and hold `build/`, `~/.gradle`,
`node_modules`, `.next` and `kotlin-js-store` in named volumes. This was measured rather than
assumed: on this machine, 3000 small-file creates took ~0.201 s natively, ~0.655 s on a bind mount
and ~0.108 s in a named volume — the volume beating the host filesystem, because it never crosses
the host boundary. Compose namespaces volumes by project, so per-workspace build isolation is a
consequence of the layout rather than something engineered.
<!-- measured: the benchmark run in this session against Docker 29.2.1 / Compose v5.0.2 on aarch64 Docker Desktop; a live measurement, re-runnable but machine-specific -->
**Considered:**
- *A thin image carrying a prebuilt jar, with no source in the container.* No bind mount at all, but
  every change costs a host build plus an image build plus a restart, and Spring devtools is lost.
- *Everything on the bind mount, including build output.* Simplest to describe, and ~3× slower on
  exactly the write pattern a build produces most of.

### Application services sit behind a compose profile

**ID:** apps-behind-a-compose-profile
**Status:** superseded by compose-untouched
**Chosen:** gate the four application services behind a profile, so plain `docker compose up -d`
still brings up the data services alone.
**Considered:** *Adding the applications to the file ungated.* Simplest file, and it would silently
change what every existing invocation does — which would have contradicted
`compose-default-identical` rather than coexisting with it. A second compose file was also
considered and rejected as two files to keep in step plus two flags on every command.

### Four dev Dockerfiles, in gymie, installing toolchains only

**ID:** dev-dockerfiles-in-gymie-under-docker
**Status:** superseded by apps-stay-host-processes
**Chosen:** all four development images live in the gymie repository under `docker/`, and install
**only the toolchain** — source arrives by bind mount at runtime, never by `COPY`. Two things follow:
the build context stays tiny, which matters because the repository has no `.dockerignore`; and the
frontend images can live in gymie without either frontend repository being touched, holding this to
two repositories, two worktrees, two review panels and two pull requests. The production
`Dockerfile.backend` and `Dockerfile.gateway` are untouched and keep serving helm.
**Considered:**
- *Stock images with command overrides and no Dockerfiles at all.* Fewest files, but no place to pin
  toolchain versions or add install steps.
- *A dev Dockerfile committed in each frontend repository.* Honest repository boundaries, at the cost
  of making this a four-repository change. The boundary smell of frontend toolchains living in the
  backend repo is accepted knowingly, and relocating them later is mechanical.

### The test suite stays on the host

**ID:** tests-stay-on-the-host
**Status:** active
**Chosen:** `./gradlew test` runs on the host in both modes, so Testcontainers is untouched and
`/myflow-do`'s verification step needs no change. The container compiles and runs the application;
the suite is a separate host invocation.
**Considered:** *Tests in a container with the runtime socket mounted.* Rejected twice over: it hands
the container control of the host's Docker, and the sibling containers Testcontainers starts would
sit outside the compose project — escaping the very isolation this change exists to provide.
*Pointing tests at the workspace's own Postgres instead* would work and rewrites the integration test
setup, which is well outside this change.

### The process-management code is kept, not deleted

**ID:** keep-the-process-management
**Status:** active
**Chosen:** `listPortListenerPids`, `killPidTree`, `freePort`, the pid tracking and the orphan sweep
all stay. Host mode remains the default and still spawns host processes, so that code is
load-bearing; `dev-lifecycle.gradle.kts` **gains** a container path rather than losing its existing
one.
**Considered:** *Deleting the whole process-management layer*, which was the original attraction of
moving to containers and is what a containers-only design would have allowed. Keeping the default
loop fast was judged worth keeping the code that serves it. **The cost is that both paths must be kept in step**, which is
why both start from the same service list and the same declared ports.

### The plan opens with a KMP feasibility spike

**ID:** spike-kmp-first
**Status:** superseded by apps-stay-host-processes
**Chosen:** prove the KMP webApp's Kotlin/JS dev server works in a container — bind-mounted source,
volume-backed caches, and a watcher that may need polling over VirtioFS — before the rest of the
gymie work is built. It is the one part of this design with no supporting evidence, and the only one
of the four applications on untrodden ground.
**Considered:** *Building in plan order and handling failure if it happens*, which would surface the
risk after the surrounding work was finished; and *spiking all four*, which spends time on three
well-understood paths.

### Every apply worktree is isolated automatically

**ID:** every-apply-worktree-isolated
**Status:** superseded by container-mode-is-opt-in
**Chosen:** any apply worktree gets a workspace id and its own stack without the operator asking,
while the main checkout gets none — so two concurrent changes never collide without anybody doing
anything, and nothing about the main checkout changes.
**Considered:**
- *Opt in per project only.* Simpler, but a project that declares isolation and forgets to use it is
  back to the silent shared-database failure, which is the one failure mode worth engineering
  against.
- *Isolate only once a second stack is detected.* Least disruption, since the first worktree to
  start keeps today's ports — but which change owns 8080 would then depend on start order, so
  nothing about a worktree's ports would be predictable or bookmarkable.

### The workspace id is derived from the full change name

**ID:** workspace-id-derived-from-change-name
**Status:** active
**Chosen:** a readable prefix plus a four-hex-character digest of the **full** change name, e.g.
`kan-15-3f9a`. The prefix keeps `docker ps` legible; the digest keeps two changes sharing a Jira key
or a truncated slug prefix from colliding. Nothing is recorded and no session coordinates with
another, because the same name always yields the same id.
**Considered:**
- *An id assigned per run and recorded in the state file.* Robust to a change being renamed, but it
  adds a state-file field and makes two sessions' ids depend on write order.
- *An id set by the operator per run.* Most control, but it is a value to remember and two sessions
  can pick the same one — reintroducing the collision this change exists to remove.
- *The prefix alone, with no digest.* Readable, but `kan-15-parallel-lanes` and
  `kan-15-parallel-two` collapse to the same prefix under truncation, which is the collision in its
  most likely form rather than its most exotic.

### Ports are deterministic from the same digest, with a whole-block fallback

**ID:** ports-deterministic-from-digest
**Status:** active
**Chosen:** offset each declared port by a value derived from the digest that produced the id, so a
change's ports are stable across sessions and bookmarkable with no registry. Check the block free
before use; if **any** port in it is bound, discard the **entire** block for free-port discovery
rather than patching it port by port, so a workspace's ports are never a mixture nobody can reason
about. Print the bound ports in the manual test guide.
**Considered:**
- *Deterministic offsets with no fallback.* Predictable, but it fails outright against any unrelated
  process holding one of those ports.
- *Always discovering free ports.* Never collides, but a change's ports differ every session, so
  nothing can be bookmarked.

### Named volume declarations are deliberately left unchanged

**ID:** volumes-unchanged-by-design
**Status:** superseded by compose-untouched
**Chosen:** change nothing in the volume declarations. The container runtime prefixes named volumes
with the project name, so parameterizing the project name isolates them as a consequence — and
applying the id to each declaration as well would double-apply it, producing a name carrying the
workspace id twice. State the reason in the configuration itself, so the absence of a change is
distinguishable from an oversight.
**Considered:** *Suffixing each volume declaration with the id.* Looks more explicit and is wrong:
it produces the double-applied name above, and a reviewer comparing it against the container names
would see an inconsistency that is in fact correct.

### The empty id reproduces today exactly

**ID:** compose-default-identical
**Status:** active
**Chosen:** every id-derived value resolves to its current value when the id is empty, so a plain
checkout keeps `gymie-postgres` on 5432 and existing seeded volumes, saved connections and local
configuration keep working.
**Considered:**
- *Always suffixed, with the main checkout taking an id too.* Uniform, with no empty-id special
  case — but every existing local setup breaks once, for nothing anyone asked for.
- *Names suffixed, ports unchanged by default.* A half-measure: two stacks could coexist as
  containers and still collide on ports, which is the failure being removed.

### The stack is started by the run that needs it, and torn down at finish

**ID:** stack-started-by-the-run
**Status:** superseded by two-run-modes-host-default
**Chosen:** `/myflow-do` exports the workspace id when it runs the project's `## run` and `## test`
commands, so whatever those start carries the id; `/myflow-finish` run 2 tears the stack down by
extending the `## stop` call its worktree cleanup already makes, and the cleanup guard reports a
surviving container as a leftover.
**Considered:** *A per-task `**Stack:** required` tag deciding when to start.* That tag belonged to
the lane plan format, which no longer exists; without lanes there is nothing for a per-task
declaration to schedule.

### The build tool's dependency cache stays shared

**ID:** gradle-cache-shared
**Status:** superseded by apps-stay-host-processes
**Chosen:** `~/.gradle` stays shared across concurrent worktrees. It is designed for concurrent
access across processes, and it is named as the single deliberate exception wherever the isolation is
documented rather than left for a reader to discover.
**Considered:**
- *A separate `GRADLE_USER_HOME` per worktree.* True isolation, at the cost of a full re-resolve and
  re-download per worktree — gigabytes of disk and minutes of wall clock.
- *A copy-on-write clone of the shared cache.* Near-free on APFS, but macOS-specific and needing a
  fallback on every other platform this pipeline runs on.

### Take both repositories in one change

**ID:** reach-two-repositories
**Status:** active
**Chosen:** the agents repository and the gymie repository together, so the isolation contract and
its first real consumer land at once and the contract is proven by something using it.
**Considered:**
- *agents only, with the configuration key defined and unfilled.* Smaller and one review panel, but
  it ships a key nothing consumes — which is how a key drifts from what its first consumer needs.
- *gymie only, with no contract.* The compose parameterization alone would work for gymie and teach
  no other project anything, so the next repository would reinvent it.

### Keep the change name even though it no longer describes the change

**ID:** change-name-kept-despite-mismatch
**Status:** active
**Chosen:** keep `kan-15-parallel-myflow-do-task-lanes`. The proposal artifact's source path derives
from the change name, and a revision round must republish to the same URL rather than mint a new one;
renaming would break that. The mismatch is recorded in the proposal and in the brainstorm design
rather than hidden.
**Considered:** *Renaming to `kan-15-workspace-isolation`.* Accurate, and it would mint a new
artifact URL — the one thing a revision round is forbidden to do — while leaving the old URL
advertised in the previous handoff and pointing at a superseded plan.

---

### Isolate lanes with scratch worktrees at detached HEAD, joined by patch

**ID:** lane-isolation-scratch-worktrees
**Status:** superseded by no-in-run-lanes
**Chosen:** one detached-`HEAD` worktree per lane at the merge base, joined into the canonical
worktree with `git apply --3way`.
**Considered:** one worktree with path-scoped review diffs, rejected because lanes would share a
source tree; lane branches with real commits, rejected because it breaks the NO-COMMITS boundary.

### Declare lanes in the plan, at planning time

**ID:** lane-declared-at-planning-time
**Status:** superseded by no-in-run-lanes
**Chosen:** `superpowers:writing-plans` derives a `## Lanes` table and per-task `**Lane:**` tags
during `/myflow-start`, checked by a guard.
**Considered:** deriving lanes at dispatch time, rejected because independence would be inferred
rather than reviewed.

### A lane is a queue, not an agent

**ID:** lane-is-a-queue-parent-controls
**Status:** superseded by no-in-run-lanes
**Chosen:** the parent session stays the SDD controller and dispatches one implementer per active
lane per round.
**Considered:** handing a whole lane to one long-lived subagent, rejected because it makes two lanes'
progress unobservable to the thing reconciling them.

### Join with git apply --3way, and never auto-resolve

**ID:** join-by-patch-3way
**Status:** superseded by no-in-run-lanes
**Chosen:** stage each lane, diff against the merge base, apply to the canonical worktree with
`--3way --index`; a conflict stops the run.
**Considered:** auto-resolving, or applying in dependency order, both rejected as silently settling a
disagreement between two independent implementations.

### A blocked lane stops; siblings run to their natural end

**ID:** blocked-lane-siblings-finish
**Status:** superseded by no-in-run-lanes
**Chosen:** the blocked lane stops, siblings finish the task in flight and hold, and the operator is
asked once.
**Considered:** stopping everything immediately; proceeding and flagging at the join; collapsing the
lanes automatically.

### Keep lane worktrees out of the state file's worktrees map

**ID:** lane-worktrees-not-in-state-map
**Status:** superseded by no-in-run-lanes
**Chosen:** lane worktrees as temporary artifacts with their own registry rows.
**Considered:** recording lanes in the map, rejected because every consumer of that map would change
meaning at once.

### Cap lanes at four

**ID:** lane-cap-four
**Status:** superseded by no-in-run-lanes
**Chosen:** at most four declared lanes, enforced by the guard.
**Considered:** two fixed, rejected as forcing a wide plan into an artificial split; unbounded,
rejected on resource contention.

### Lanes ship as separate branches, one PR each

**ID:** lanes-ship-as-separate-branches
**Status:** superseded by no-in-run-lanes
**Chosen:** each lane gets branch `openspec/<name>/lane-<id>`, its own worktree, its own review
panel, its own human gate and its own PR targeting develop, with a declared `Depends on` column
producing stacked lanes and a parent-performed rebase on a discovered dependency.
**Considered:** the patch join it replaced, and one PR containing every lane, rejected as
contradicting shipping separately. Superseded in turn once it was clear that a lane shaped this way
is a change.

### Start a lane's stack on demand, via a per-task tag

**ID:** stack-on-demand
**Status:** superseded by stack-started-by-the-run
**Chosen:** a task declares `**Stack:** required`; the lane brings its stack up before the first such
task and tears it down at lane end.
**Considered:** eagerly for every lane; eagerly for lanes whose scope touches application code.

### Offset ports by lane index

**ID:** ports-offset-then-discovery
**Status:** superseded by ports-deterministic-from-digest
**Chosen:** each declared port offset by `index × 100` for the lane at that index, with whole-block
discovery on collision.
**Considered:** fixed offsets with no fallback; always discovering. The whole-block fallback survives
into the superseding decision; the index-based derivation does not, because without lanes there is no
index — a change has a name, not a position.

## Risks / Trade-offs

**The Redis probe is racy between two simultaneous starts** → two `devStart`s launched in the same
second could both observe an index as free and claim it. The window is small and the recovery is to
restart one, but it is a real gap and no lock is proposed: a lock would be more machinery than the
failure justifies, and the failure is a shared cache rather than a corrupted database.

**A four-hex digest is not collision-proof** → two concurrently running changes whose full names
collide in four hex characters would share a database. A machine has tens of change names rather
than thousands, so this is unlikely rather than impossible, and widening the digest is a one-line
change if it ever bites.

**Mailpit stays shared** → two changes' emails interleave in one inbox. Confusing when testing email
flows concurrently, not corrupting, and mailpit has no natural namespace to use.

**A worktree's ports differ from the project's documented ones** → which is why the manual test guide
names the worktree's own URLs rather than the project's defaults. An operator who opens the
documented URL out of habit reaches whichever workspace holds 8080; the guide is the mitigation, and
it is why that requirement is in the spec rather than left implicit.

**Two app stacks cost real machine resources** → two JVMs, two Node processes and two Gradle daemons.
Nothing here schedules or limits them; a change that never runs the apps never pays.

**The port block is bounded by what is free** → the block is checked before use and falls back to
discovery as a whole, per the existing rule. A machine running many workspaces will eventually
exhaust the predictable blocks and fall back for all of them.

**`application-local.yml` gains placeholders in a file every developer uses** → but every placeholder
carries today's value as its default, so a checkout with nothing set behaves exactly as before. This
is the one shared file this change edits, and it is edited in the least invasive way available.

**Isolation is automatic, which means it also applies when it is not wanted** → an operator who
deliberately wants a worktree pointed at the shared database has to pass `-PpgDb=gymie`. That is a
smaller cost than the reverse default, where forgetting a flag silently reproduces the original bug.

## Migration Plan

**Nothing changes in the main checkout**: the same database, the same ports, the same bucket, the
same Redis index, and `docker-compose.yml` byte-identical. A worktree becomes isolated automatically
the first time `devStart` runs there — creating its database and bucket, and saying so. There is
nothing to roll back beyond reverting the change itself.

## Open questions

None. Every question raised across four brainstorming sessions was put to the operator and answered
before this design was written.
