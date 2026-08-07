# Tasks — kan-15-parallel-myflow-do-task-lanes

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** give each apply worktree its own Postgres database, Redis index, MinIO bucket and
application ports, so one change's migrations cannot reach another change's schema and two changes
can run their applications at the same time.

**Architecture:** a workspace id derived from the change name (`kan-15-55a6`) yields a database
`gymie_kan_15_55a6`, a bucket `gymie-media-kan-15-55a6`, and an offset block for the application
ports it can move — the backend, the gateway and the admin panel. The Redis index is **probed and
claimed**, not derived, because sixteen indices collide too easily. The data service containers,
their host ports and `docker-compose.yml` are **not touched** — only the logical resources inside
them are namespaced.

**All four application ports move, the KMP frontend's included**, per the design decision
`kmp-frontend-port-rotates`. Its dev server's port is set in `gymie-frontend`'s own build file, but
`webApp/webpack.config.d/` is Kotlin/JS's extension point for exactly this and that repository
already carries a snippet there — so the override is one small file beside an existing one. The earlier
decision that the frontend could not move was wrong, and the machinery built around it is deleted by
this change rather than maintained.

**The change name says "lanes" and this plan does not build them.** Two earlier designs — in-run
lanes, and a containerized dev stack — were rejected and are preserved in `design.md` as superseded
decisions. The name is kept because the artifact's source path derives from it and a revision round
must not mint a new URL.

**Repositories:** this change spans **three**. The admin frontend repository is not touched;
`gymie-frontend` is, in one one-file change.

- `/Users/tweety53/Projects/agents` — a documentation-and-guards repository with no runnable
  application. Verification is the guards declared in its `.myflow/project.md`
  (`check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`,
  `check-task-build-green.sh`), the cleanup guard's harness, and targeted `grep` assertions.
- `/Users/tweety53/Projects/gymie` — Kotlin / Spring Boot / Gradle. Verification is
  `./gradlew ktlintCheck detekt`, `./gradlew test`, and running the dev stack in two worktrees.

Group 3's commands run in the **gymie apply worktree**. **Resolve it once from
`git -C /Users/tweety53/Projects/gymie worktree list` and use the absolute path it prints** — never
the main checkout while the worktree holds uncommitted work, and never a relative sibling path.

**Order:** group 1 reduces the contract every later task cites. Group 2 wires the agents-side
consumers and the cleanup row. Group 3 is gymie, and its first task is the configuration placeholder
that everything else there depends on. Group 4 verifies across both repositories and runs last.

**A guard limitation, discovered while implementing an earlier revision of this change.**
`scripts/check-references.sh` is **line-scoped**: it checks a citation only when the bold token and
the backticked path sit on the **same physical line**. A citation split across two lines by prose
wrapping passes unchecked. Every task below that writes a citation must keep the token and path on
one line and **verify by mutating the token and confirming the guard fails** — a passing guard is not
evidence it looked.

**Global constraints, carried into every dispatch:**

- **No suppression markers, ever.** Fix the offending line; never silence a guard, and never weaken a
  guard's configuration or thresholds.
- **Citation over copy**, subject to the line-scoping note above.
- Every fenced block added to any change's `tasks.md`, `design.md` or `proposal.md` carries
  `verified:`/`unverified:` on its info string, and every numeric claim carries a
  `measured:`/`predicted:` comment within two lines, per **The four tags**
  (`skills/myflow-contracts/plan-provenance.md`).
- Every task carries exactly one `**Build:**` tag, per **The build-green tag**
  (`skills/myflow-contracts/build-green.md`).
- **Backwards compatibility is a hard requirement.** The main checkout must keep `gymie`, Redis
  index 0, `gymie-media` and ports 8080 / 8081 / 3000 / 3001. Every placeholder added carries today's
  value as its default.
- **Do not touch** `docker-compose.yml`, the production Dockerfiles, the helm chart, the test suite
  or its Testcontainers setup, or the **admin** frontend repository. `gymie-frontend` is touched in
  exactly one file, `webApp/webpack.config.d/`, and nowhere else.

---

## 1. The workspace isolation contract

### 1.1 Reduce `skills/myflow-contracts/workspace-isolation.md` to this design

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/workspace-isolation.md`

**Interfaces:**
- Produces: the headings later tasks cite — `## The workspace id`, `## What the id derives`,
  `## The cache index`, `## The empty id`, `## Creation and cleanup`.

**This file already exists**, written and guard-verified under an earlier revision. Its **id
derivation is correct and must not change**. Everything written for the containerized design must go.

- [x] Keep `## The workspace id` exactly as it is — prefix plus four-hex SHA-256 of the full change
      name, deterministic, unrecorded, uncoordinated, and no id in the main checkout.

- [x] Replace the container-era sections with `## What the id derives`: a database name, an
      object-store bucket name, and a block of application ports. State that a derived **database
      name must be safe unquoted** — underscores, not dashes — because a call site that omits quoting
      must not be able to break on it, and that a bucket name is not a SQL identifier so it takes the
      id verbatim. State that the **shared service containers, their host ports and the
      container-runtime configuration are not changed**: what is namespaced is the logical resource
      inside each service.

- [x] Add `## The cache index`: chosen by **probing** which indices are in use and claiming a free
      one, recorded for the run, and **never derived from the digest**. Give the reason — sixteen
      indices means a derived value collides at roughly six percent for two concurrent workspaces, and
      a collision silently shares sessions and cache. State the accepted cost: the index is not stable
      across restarts, which does not matter for sessions and cache.

- [x] Rewrite `## The empty id`: every derived value resolves to the project's declared default when
      the id is empty, and **isolation is automatic in an apply worktree** rather than opt-in, because
      forgetting a flag would silently reproduce the failure being prevented.

- [x] Add `## Creation and cleanup`: the resources are created on demand and reported the first time;
      a new database starts **empty** and is brought up to date by the project's normal migration and
      seeding path, never by copying an existing database; and at finish the resources are removed,
      with **a service that is not running reported and skipped rather than failed**.

- [x] **Delete** everything describing run modes, containers, dev images, bind mounts, volume-backed
      caches and compose profiles. None of it survives this design.

- [x] Verify, including a mutation test on every citation:

```bash unverified:run after this task's edits are in place
cd /Users/tweety53/Projects/agents
grep -n '^## ' skills/myflow-contracts/workspace-isolation.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
```

Expected: five headings listed, all four guards exit 0. Then mutate each citation's bold token in a
scratch copy and confirm `check-references.sh` **fails** for each.

**Build:** green

---

## 2. The agents-side consumers

### 2.1 The `## workspace isolation` project configuration key

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/project-configuration.md`

- [x] Add `## workspace isolation` to the optional-keys table: the resources a workspace id derives,
      the default value of each, and the commands that create and remove them.

- [x] State that a project declaring no such section behaves exactly as today **everywhere, including
      in apply worktrees**, and that this is **not** a misconfiguration — it is correct for this
      repository, which declares no runnable application.

- [x] State that entries resolve under the same containment rules the file already applies to its
      other entries, and that an entry failing them is reported by name and dropped rather than
      repaired.

**Build:** green

### 2.2 `/myflow-do` computes and exports the id

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-do/SKILL.md` — section **2. Isolate the
  workspace** and section **7. Verify, stage, and hand off**

- [x] In section 2, after the existing worktree instructions: compute the worktree's workspace id
      from the change name, per **The workspace id**
      (`skills/myflow-contracts/workspace-isolation.md`). Cite the contract; do not restate the
      derivation.

- [x] In section 7, before running the project's `## lint` and `## test` commands, export the
      variables the project's `## workspace isolation` section declares. A project declaring none
      exports nothing and behaves exactly as today. **The test command is unchanged** — this task
      changes what is exported, never how tests run.

**Build:** green

### 2.3 The manual test guide names the worktree's own URLs

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-do/SKILL.md` — section **6. Write the manual
  test guide**

- [x] Extend the existing absolute-path rule: every **URL** in the guide is the one this worktree
      resolved, not the project's declared base. State the reason — an operator opening the documented
      URL out of habit reaches whichever workspace holds the default port.

- [x] State that a project declaring no isolation writes the project's declared URLs unchanged, so
      nothing about an existing guide's shape changes.

**Build:** green

### 2.4 The cleanup registry row and its guard marker, in one task

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/pipeline.md` — **Temporary
  artifacts registry**
- Modify: `/Users/tweety53/Projects/agents/scripts/check-cleanup-complete.sh`

- [x] Add one registry row: **Workspace database and bucket** — created on first start in a worktree,
      living in the project's shared data services, removed at finish run 2.

- [x] Add the matching marker line to `check-cleanup-complete.sh`, in the existing block:

```text verified:the marker vocabulary is read from scripts/check-cleanup-complete.sh lines 51-59 at branch main
# registry-row-checked: Workspace database and bucket
```

- [x] Implement the check **against the project's declared survivor-reporting command**, whose
      output and exit-code contract `skills/myflow-contracts/project-configuration.md` defines —
      read it there rather than inventing a convention. A survivor it reports is `LEFTOVER:`;
      **a non-zero exit means the service could not be reached, which is reported and skipped, not
      failed**, so run 2 continues to `FINISHED`. A project declaring no such command has the
      verification reported as skipped, never as passed, and one declaring no isolation at all
      passes, because a step whose artifact is already absent is a success.

- [x] **Never infer a survivor from the removal command's exit code.** A removal that reported
      success against a stale connection leaves the registry row's promise broken with nothing
      having failed, which is why the guard asks rather than assumes.

- [x] **These edits are one task on purpose.** The guard's harness fails when the registry table and
      the marker lines disagree in either direction, so splitting them would leave the repository red
      between two tasks.

- [x] Verify:

```bash unverified:run after this task's edits are in place
cd /Users/tweety53/Projects/agents
scripts/test-check-cleanup-complete.sh
scripts/check-references.sh
```

Expected: both exit 0.

**Build:** green

### 2.5 Contract index rows

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/SKILL.md` — the `## Index` table
- Modify: `/Users/tweety53/Projects/agents/CLAUDE.md` and
  `/Users/tweety53/Projects/agents/AGENTS.md` — the narrower-contracts table
- Modify: `/Users/tweety53/Projects/agents/rules/myflow-manual-review.mdc` — the same table
- Modify: `/Users/tweety53/Projects/agents/skills/README.md` — a **fifth** contract enumeration, not
  in this list when the plan was written. Added deliberately during the run: the task's success
  criterion is that the contract is reachable from **every** entry point, and this file is one.

- [x] Add a `workspace-isolation.md` row to every contract index above, describing when to load it.
      Describe the load condition without claiming the cache index is **derived** — it is probed and
      claimed, and a trigger row is read by agents who will not open the file.

- [x] Add the contracts already missing from these same lists — `plan-provenance.md` and
      `build-green.md` — and correct any count word beside them. **Raised by this task's own
      implementer:** every enumeration outside `skills/myflow-contracts/SKILL.md` was already short
      by two before this change. Extending a list while leaving known-missing siblings out makes it
      wrong in a new way, in the line this task just touched.

- [x] Verify the contract is reachable from every entry point:

```bash unverified:run after 2.5's edits are in place
cd /Users/tweety53/Projects/agents
grep -rn "workspace-isolation" --include="*.md" --include="*.mdc" . | grep -v openspec/changes/
scripts/check-references.sh
```

Expected: the contract appears in all four indexes, and the guard exits 0.

**Build:** green

### 2.6 `/myflow-finish` run 2 actually removes the workspace's resources

**Files:**
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-finish/SKILL.md` — run 2's steps
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/pipeline.md` — **Run 2 — the
  branch is merged**

**Why this task exists, and why it was not in the plan as written.** Task 2.4's review found that
**nothing in the pipeline calls the project's `remove` command.** The registry row task 2.4 added
says run 2 removes these resources, and **Project configuration** says "`/myflow-finish` run 2 calls
it, and nothing else does" — but run 2's own step list never did, and `skills/myflow-finish/SKILL.md`
does not contain the word *workspace* at all. The guard task 2.4 built would therefore report the
database and bucket as leftovers on **every** run 2 of an isolated project, and the change would
never reach `FINISHED`. The plan built the verification and the declaration without the step they
verify.

- [x] Add the removal to run 2's steps in both files, **before** the cleanup verification that
      checks it — a verification that runs before the thing it verifies can only ever fail.

- [x] Place it relative to the worktree removal deliberately, and say why: the resources live in the
      project's shared services rather than in the worktree, so removing the worktree neither removes
      them nor makes them unreachable. State the order chosen and the reason.

- [x] A project declaring no isolation, or no `remove` command, has this step **skipped, not failed**
      — the same rule the surrounding steps already follow, and the same one the verification applies.

- [x] **A failed removal does not stop run 2 here.** It is reported, and the verification step
      immediately after is what decides the verdict — the survivors report is the authority on what
      is left, per **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`), and
      inferring the outcome from the removal command's exit code is what that contract forbids.

- [x] Verify the wiring is real, not just described:

```bash unverified:run after 2.6's edits are in place
cd /Users/tweety53/Projects/agents
grep -n 'workspace' skills/myflow-finish/SKILL.md
scripts/check-references.sh
scripts/test-check-cleanup-complete.sh
```

Expected: run 2 names the removal step, and both guards exit 0.

**Build:** green

### 2.7 A guard for the `## workspace isolation` declaration

**Files:**
- Create: `/Users/tweety53/Projects/agents/scripts/check-workspace-isolation.sh`
- Create: `/Users/tweety53/Projects/agents/scripts/test-check-workspace-isolation.sh`
- Modify: `/Users/tweety53/Projects/agents/.myflow/project.md` — the `## lint` and `## test` lists
- Modify: `/Users/tweety53/Projects/agents/skills/myflow-contracts/project-configuration.md` — say
  what enforces the rules

**Why this task exists, and why it was not in the plan as written.** The review panel's adversarial
slot found that the resource-table rules — the closed `Resource` vocabulary, the four
`In a workspace` cell forms, `<value:…>` resolution and its exclusions, bare-integer defaults, the
refuse-rather-than-fall-back rule — exist **only as prose an agent re-performs each run**, with no
mechanical check anywhere. The command table is well covered by
`scripts/test-check-cleanup-complete.sh`; the resource table is covered by nothing. The operator
chose to build the guard rather than record it as a follow-up or narrow the contract.

- [x] Write the guard. It reads a project's `.myflow/project.md`, and for a declaration it finds it
      enforces what **project-configuration.md** states — read the rules from that file rather than
      from this checkbox, which would be a second copy of them.

- [x] **A project declaring no section passes**, silently. That is the overwhelmingly common case and
      includes this repository.

- [x] Report every violation **by name**, with the row and the rule it broke — the contract's own
      remedy is that a failing row is reported by name and dropped, so a guard that says only *this
      file is invalid* is not enough to act on.

- [x] Write the harness first, case by case, and make each case fail before the rule exists. Cover at
      minimum: an unknown `Resource` word; a `port` and a `cache index` whose `Default` is not a bare
      integer; each of the four cell forms violated; a `<value:…>` naming no row, naming a `url` row,
      and naming a `cache index` row; a duplicated heading; a missing column; and a valid declaration
      — gymie's, which must pass.

- [x] Add it to this repository's `## lint` and `## test` lists so it runs with the others.

- [x] Say in **project-configuration.md** which rules are now mechanically enforced and which remain
      the agent's to apply, so the contract stops implying that all of them are checked.

```bash unverified:run after 2.7's edits are in place
cd /Users/tweety53/Projects/agents
scripts/test-check-workspace-isolation.sh
scripts/check-workspace-isolation.sh /Users/tweety53/Projects/gymie-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes
scripts/check-workspace-isolation.sh /Users/tweety53/Projects/agents
```

Expected: the harness exits 0, gymie's real declaration passes, and this repository — which declares
no section — passes silently.

**Build:** green

---

## 3. The gymie repository

### 3.1 Make the local profile's datasource overridable

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/src/app/src/main/resources/application-local.yml`

**Interfaces:**
- Produces: the `DB_URL` override path that tasks 3.2 and 3.3 depend on.

**This is the pivot of the whole gymie change.** `application.yml` already carries
`url: ${DB_URL:jdbc:postgresql://localhost:5432/gymie}`, but `application-local.yml` hardcodes the
same URL with no placeholder and `local` is the profile development runs — so it overrides the hook
that already exists.

- [x] Replace the hardcoded datasource URL with the same defaulted placeholder `application.yml`
      uses, so an unset `DB_URL` yields byte-identical behaviour.

- [x] Do the same for the Redis index, so an unset value yields index 0.

- [x] Do the same for the server port, so an unset value yields today's port.

- [x] **Do not touch** `application-test.yml`. The test suite uses Testcontainers and is out of scope.

- [x] Verify the defaults are unchanged by starting the backend with nothing set and confirming it
      connects to `gymie` on 5432 and binds its usual port.

**Build:** green

### 3.1b Remove the local profile's three remaining shadows

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/src/app/src/main/resources/application-local.yml`
- Modify: `/Users/tweety53/Projects/gymie/src/gateway/src/main/resources/application.yml`
- Modify: `/Users/tweety53/Projects/gymie/README.md`

**Interfaces:**
- Produces: the `STORAGE_ENDPOINT`, `STORAGE_BUCKET`, `STORAGE_PUBLIC_BASE_URL`, `FRONTEND_URL`
  and `ADMIN_FRONTEND_URL` override paths task 3.2 exports, and a gateway that honours
  `REDIS_DATABASE`.

**Why this task exists, and why it was not in the plan as written.** Task 3.1 fixed the datasource,
and its review found the datasource was not the only value `application-local.yml` shadows. Three
more shadow a placeholder `application.yml` already carries, and each one blocks a requirement of
this change's delta spec rather than merely being untidy: without them the bucket task 3.2 derives
is created and never read, the backend builds its links against whichever workspace holds the
default frontend ports, and the gateway keeps sessions on cache index 0 while the backend moves to
its own. **This is the same defect class 3.1 exists to remove, on three more properties**, and the
plan reached implementation naming only one of the four.

- [x] Replace the hardcoded `gymie.storage` bucket, endpoint and public base URL with the same
      defaulted placeholders `application.yml` already carries, so an unset value yields today's
      literal bucket and endpoint unchanged.

- [x] Replace the hardcoded `app.frontend-url` and `admin.frontend-url` with the defaulted
      placeholders `application.yml` carries, so an unset value yields today's literal URLs
      unchanged.

- [x] Add the Redis index key to the gateway's configuration, which today sets host, port and
      password and no index, defaulting to the same value the backend defaults to.

- [x] **Do not touch** `application-test.yml`, and change nothing in the gateway beyond the one
      key above.

- [x] Document the environment variables this task and 3.1 introduce in the README's
      environment-variable table, which lists the other Redis and storage variables already.

- [x] Replace the remaining hardcoded Redis host, port and password in the local profile with the
      same defaulted placeholders, so the README's claim that those three are overridable becomes
      true. **Raised by this task's review**, which found the README already documenting them as
      environment variables while the local profile silently ignored all three — the same defect
      class on the last values in that block. None of them is workspace-derived; what is fixed here
      is a false claim, not an isolation gap.

- [x] Verify each placeholder **both ways** — unset yields the literal value the main checkout has
      today, and a set value is honoured by the running application, read from the service it
      reaches rather than from the application's own logs:

```bash unverified:run in the gymie apply worktree after 3.1b's edits are in place
cd /Users/tweety53/Projects/gymie-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes
./gradlew ktlintCheck detekt
```

Expected: exit 0, and the unset case indistinguishable from the main checkout's behaviour.

**Build:** green

### 3.2 Derive the workspace resources in the dev lifecycle

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`

- [x] Derive the workspace id from the change name when running in an apply worktree, and use the
      project defaults in the main checkout, per **The workspace id**
      (`skills/myflow-contracts/workspace-isolation.md`) — resolve that file's absolute installed
      path rather than assuming a repo-relative one.

- [x] Derive the database name, the bucket name and the application ports from it — the backend,
      the gateway and the admin panel, and **not** the KMP frontend, per
      `kmp-frontend-not-port-isolated` in `design.md`. Offset the ports as a block and check the
      block free before use, falling back to discovery **as a whole block** rather than port by
      port.

- [x] The admin panel's port is fixed by its own `package.json` dev script, so move it by changing
      the command this file runs rather than by editing the admin frontend repository, which is out
      of scope. Introduce the port variable it needs; none exists today.

- [x] **Probe Redis for a free index and claim the lowest one**, recording it for the run. Do not
      derive it from the digest — sixteen indices collide too readily, and the failure is silent.

- [x] Create the database and bucket if absent, and **report having done so the first time**. Reuse
      the `psql` invocation shape `dbReset` already uses, which quotes its identifier correctly.

- [x] Pass every derived value into the processes it starts, using the placeholder names task 3.1
      and task 3.1b established — the database URL, the cache index, the bucket, the storage public
      base URL, the two frontend URLs and each application's port. A value derived and never
      exported is the failure 3.1b was added to remove, one level further out.

- [x] Leave the pid tracking, `listPortListenerPids`, `killPidTree`, `freePort` and the orphan sweep
      **in place**. The applications remain host processes and those still orphan; this task adds
      derivation, it deletes nothing.

- [x] **Bring the code this task adds under the lint gate, and disclose what stays outside it.**
      `ktlintCheck detekt` is configured inside `subprojects {}` only, so it scans neither `buildSrc/`
      nor the root's `gradle/*.gradle.kts` — proven by this task's review, which inserted a badly
      formatted declaration into both and watched lint still exit 0. Apply the project's existing
      ktlint and detekt configuration to `buildSrc/`, where the derivation lives, and **fix whatever
      it then reports** rather than suppressing it. `gradle/dev-lifecycle.gradle.kts` stays outside
      the gate, by the operator's decision: it carries baseline code this change never touched, and
      widening the gate onto it is its own change. **Say so here rather than letting `Build: green`
      be read as "the project's lint passed on this file"** — what verifies it is compilation plus
      `WorkspaceTest`, which `jar.finalizedBy(test)` runs on every Gradle invocation.

- [x] **Make the id derivation locale-independent, in the contract and in the port together.** This
      task's review ran the canonical block under two locales and got two different ids for the same
      name, and a third from the Kotlin port — so the contract's own promise, that one change name
      always yields one id, does not hold for a name outside `[a-z0-9-]`. Pin the normalisation to
      one defined behaviour, state it in **The workspace id**
      (`skills/myflow-contracts/workspace-isolation.md`), and make the Kotlin match it. Add the
      diverging name to the pinned cases, so the agreement is asserted rather than assumed.

- [x] Verify both checkouts:

```bash unverified:run in the gymie main checkout and then in the apply worktree
./gradlew devStart && ./gradlew devStatus && ./gradlew devStop
```

Expected: the main checkout binds 8080, 8081, 3000 and 3001 against `gymie`; the worktree creates and
binds its own database, index, bucket and port block, and says so on first start.

**Build:** green

### 3.3 Point the generated .env and the database tasks at the workspace

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/src/app/src/main/resources/application-local.yml` — the
  OAuth redirect URI, added after task 3.2's implementer found it
- Modify/create under `/Users/tweety53/Projects/gymie/buildSrc/` — the pure, testable half of this
  task's logic, following task 3.2's precedent. Recorded here rather than left as an undisclosed
  footprint: `dev-lifecycle.gradle.kts` has **no test harness and is outside the lint gate**, so
  logic placed there can be neither unit-tested nor style-checked, and `buildSrc/` is inside both.

- [x] Add the **`remove` and `survivors` commands the project must declare**, alongside the `create`
      command task 3.2 added. Task 3.4 declares all three in `.myflow/project.md`, and
      `/myflow-finish` run 2 calls `remove` and then the cleanup guard calls `survivors` — so
      without them 3.4 declares commands that do not exist and run 2 can never verify a removal.
      Their contract — what `survivors` prints, what each exit code means, and the bound it is given
      — is **The `survivors` output and exit-code contract** in
      `skills/myflow-contracts/project-configuration.md`, read from the agents apply worktree.

- [x] Derive the backend's OAuth **redirect URI** from the workspace's gateway port. It is hardcoded
      to `http://localhost:8080/login/oauth2/code/google`, so a worktree whose gateway moved sends
      Google's callback to whichever workspace holds 8080. **Found by task 3.2's implementer and
      named in no task before this one** — it is the same shadow class as 3.1b, on the one value
      whose failure lands in an external service's redirect.

- [x] Derive every application URL the generated `.env` writes from the workspace's port block. It
      currently pins `ADMIN_API_URL=http://localhost:8080`, which would send a worktree's frontends to
      whichever workspace holds that port. **`ADMIN_API_URL` and `NEXT_PUBLIC_ADMIN_API_URL` are both
      written by `ensureAdminFrontendEnv` and both need deriving** — task 2.1's re-review found the
      second one named nowhere in this plan.

- [x] Default `pgDb` in `dbReset` and `dbSeed` to the workspace's database in a worktree, and to
      `gymie` in the main checkout. `-PpgDb` still overrides both.

- [x] Confirm the existing `dbReset` confirmation guard (`-PconfirmDb` / `-Pforce`) still compares
      against whatever database is actually targeted, so the safety check cannot be bypassed by the
      default changing.

- [x] Verify:

```bash unverified:run in the gymie apply worktree after 3.1 through 3.3
./gradlew dbReset -PconfirmDb=<the workspace database name>
./gradlew dbSeed
./gradlew ktlintCheck detekt
./gradlew test
```

Expected: the reset targets the workspace database, seeding succeeds against it, and both checks
exit 0 with the test suite unaffected.

**Build:** green

### 3.4 Declare gymie's workspace isolation

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/.myflow/project.md`

- [x] Add a `## workspace isolation` section in the shape task 2.1 defines, naming the database,
      bucket, cache index and the application ports this change can move, with their defaults, and
      the create, remove **and survivor-reporting** commands. Read the shape from
      `skills/myflow-contracts/project-configuration.md` as it now stands rather than from this
      checkbox — task 2.1's fix round added a resource kind and a third command verb after this plan
      was written.

- [x] Declare the composed URL values too — the storage public base URL and every application URL
      that embeds a port this workspace moves. The shape has a row kind for exactly this; a URL left
      undeclared is a value nothing exports, which is the failure task 3.1b was added to remove.

- [x] State that the integration tests use Testcontainers and are deliberately untouched, that
      **mailpit is deliberately shared** — one inbox, no natural namespace, so two changes' emails
      interleave — and that the **KMP frontend keeps port 3000 and is not isolated**, per
      `kmp-frontend-not-port-isolated` in `design.md`. All three are stated limitations, and the last
      one is the reason the port list is shorter than the application list.

- [x] Update `## apps` and `## run` to say that a worktree gets its own database, index, bucket and
      ports automatically, and that the main checkout keeps the documented defaults.

**Build:** green

---

## 5. Rotate the KMP frontend's port and retire the shared-port machinery

**Why this group exists.** The review panel found the same Critical **three times** — a workspace
killing another workspace's frontend, through the start path, the stop path and the failure path.
Every one of them lived in machinery that exists only because port 3000 could not move. It can:
`webApp/webpack.config.d/` is Kotlin/JS's own extension point, and `gymie-frontend` already carries
a snippet there that mutates `config.devServer`. See `kmp-frontend-port-rotates` in `design.md`.

**Order:** 5.1 makes the port movable, 5.2 makes gymie move it, 5.3 deletes what is then dead. 5.3
last, so the machinery is never removed before its replacement works.

### 5.1 Give the frontend dev server a port override

**Files:**
- Create: `/Users/tweety53/Projects/gymie-frontend/webApp/webpack.config.d/devServerPort.js`

**Interfaces:**
- Produces: the `FRONTEND_PORT` override path tasks 5.2 and 5.3 depend on.

- [x] Add the snippet, beside the existing `historyApiFallback.js` and in the same shape: read the
      port from the environment, and apply it only when set so an unset environment yields today's
      port unchanged.

- [x] **Do not edit `webApp/build.gradle.kts`.** The hardcoded `port = 3000` stays as the default;
      what this task adds is an override layered over it, which is why the file it touches is the
      extension point rather than the build script.

- [x] Verify both ways — unset yields 3000, set yields the given port — reading the result from the
      **listening socket**, not from a log line.

**Build:** green

### 5.2 Move the frontend with the rest of the port block

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/.myflow/project.md`

- [x] Make the frontend port-isolated like the other three: give its `ManagedService` the same
      treatment, export `FRONTEND_PORT`, and let it join the offset block and the whole-block free
      check.

- [x] Add the `port` row for it to `## workspace isolation`, and change `FRONTEND_URL` from a
      token-free `url` row to one that follows the port, as the other application URLs do.

- [x] Remove the three statements that the KMP frontend is an exception — in `## apps`, in `## run`
      and in the isolation section's limitations — since it no longer is.

- [x] Verify the worktree resolves four moved ports, and that the main checkout still resolves
      8080 / 8081 / 3000 / 3001.

**Build:** green

### 5.3 Delete the shared-port ownership machinery

**Files:**
- Delete: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/SharedPort.kt`
- Delete: `/Users/tweety53/Projects/gymie/buildSrc/src/test/kotlin/com/gymie/gradle/SharedPortTest.kt`
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/README.md` — its dev-lifecycle section still documents the
  ownership record, `~/.gymie/dev-shared-ports/` and "`devStop` removes the record it owns". Task
  5.2's review found it: neither 5.2's nor 5.3's original file list named it, so deleting the code
  without it would leave the repository's front page describing a mechanism that no longer exists.
- Modify: `/Users/tweety53/Projects/gymie/buildSrc/src/test/kotlin/com/gymie/gradle/DevEnvTest.kt` —
  pin the port bounds as literals

- [x] Remove the ownership record, its directory under the user's home, the refusal path, the
      not-started reason files and every call site — **once 5.2 proves nothing needs them**. A port
      every workspace can move has no shared owner to record.

- [x] Restore the straightforward behaviour the machinery replaced: a port this run resolved is
      this run's to free, for every service alike, with no per-service exception.

- [x] Confirm the deletion is total: no `sharedPort`, no `not-started`, no `dev-shared-ports`
      remains anywhere in the repository.

- [x] Verify the full suite still passes, and that `devStop` in one worktree leaves another
      worktree's processes alone — the property the deleted machinery existed to provide, now
      provided by the ports being genuinely distinct.

- [x] **Pin the port bounds as literals.** `DevEnvTest` derives its expectations from
      `MIN_BINDABLE_PORT`/`MAX_BINDABLE_PORT`, so 5.2's review mutated `MAX_BINDABLE_PORT` to 70000
      and the suite still passed — the boundary *logic* is tested, the *values* are not. Assert the
      literal 1 and 65535, which is also what the frontend override checks.

**Build:** green


## 6. Start and stop a chosen subset of the services

**Why this group exists.** The dev stack is all-or-nothing: `devStart`, `devStop` and `devStatus`
each act on all four services. Restarting only the backend after a code change means restarting the
frontend too, which costs a Kotlin/JS compile measured in minutes. An operator asked for a way to
name the services a command acts on, several at a time.

**Why it belongs to this change rather than a follow-up.** Partial starts and workspace port
isolation are not independent: the block-free check in `resolveWorkspacePorts` treats *any* listener
on a block port as a clash and discards the whole block. Under a partial start the holders are this
workspace's own already-running services, so the check would fire on itself and hand the started
service a discovered port while its siblings stayed on the block. The feature cannot be added
correctly without touching the isolation this change introduces, so it is added here.

**Order:** 6.1 and 6.2 are independent pure decisions; 6.3 wires both into the build.

### 6.1 The selection parser

**Files:**
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/ServiceSelection.kt`
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/test/kotlin/com/gymie/gradle/ServiceSelectionTest.kt`

**Interfaces:**
- Produces: the selection function task 6.3 calls once per dev task.

- [x] Parse the raw `-Pservices` value against the list of known service names. Absent means every
      service, which is what keeps the four commands' current behaviour intact for an operator who
      passes nothing.

- [x] Accept several names in one value, comma-separated — that is the whole point of the flag, and
      a one-name-per-invocation surface would not answer the request.

- [x] Return the subset in the **known** order rather than the order given, since the caller derives
      its start order from that list and its stop order from the reverse.

- [x] Refuse a value that is blank or names something unknown, listing the known names. A blank
      value is a typo, not a request for all four; treating it as all four would silently restart a
      frontend the operator was trying to leave alone.

**Build:** green

### 6.2 The port block under a partial start

**Files:**
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/PortBlock.kt`
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/test/kotlin/com/gymie/gradle/PortBlockTest.kt`

**Interfaces:**
- Consumes: nothing — it is handed the block, who holds it and which services this checkout has
  running.
- Produces: the decision task 6.3 substitutes for the inline block-free check.

- [x] Decide between three outcomes rather than today's two: keep the deterministic block, discard
      it and discover free ports, or refuse.

- [x] A block port held by **this checkout's own running service** is not a clash. That is the case
      a partial start creates, and without it the block is discarded by the very stack it belongs to.

- [x] Refuse when a **foreign** holder is found while this checkout still has services running.
      Discovering a new block there would put the started service on ports its running siblings are
      not on — a split stack that reports one set of ports and serves on another. A refusal naming
      the holder and telling the operator to stop the stack and start it whole is the honest answer.

- [x] Keep the existing behaviour exactly where it applies: a foreign holder with nothing of ours
      running still discards and discovers, which is what a full `devStart` does today.

**Build:** green

### 6.3 The four dev tasks honour the selection

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/README.md`

- [x] Read `-Pservices` once per task and pass the resolved selection to `devStart`, `devRestart`,
      `devStop` and `devStatus`. `devRestart` depends on `devStart`, and a Gradle project property
      reaches it unchanged, so it needs no separate surface.

- [x] Start in the selection's order and stop in its reverse, as the four already do today.

- [x] Scope the frontend and admin-frontend checkout-existence refusals to whether that service is
      in the selection. `devStart -Pservices=backend` must not demand a frontend checkout it will
      never enter.

- [x] Leave the workspace resources unconditional — the database, the bucket and the cache index are
      resolved and created whatever the selection is, so two runs against one workspace cannot
      disagree about what it owns.

- [x] Substitute task 6.2's decision for the inline block-free check, and refuse on its refusal with
      the message it produces.

- [x] Name the flag in each of the four task descriptions and in the README's dev-lifecycle section,
      with an example naming more than one service.

- [ ] Verify the four commands against a real stack: all four services with no flag, a subset with
      the flag, an unknown name refused, and a partial start leaving the running siblings on their
      block ports.

**Build:** green

## 7. Record the resolved block, and delete the inference built to replace it

**Why this group exists.** Group 6 made a partial start possible, which raised a question the design
had never had to answer: *which port is a service this run is not touching actually serving on?* Three
fix rounds answered it by inferring — from a pid file, a port file, and `lsof` — and each round's
repair produced more findings than it closed. Round 5 fixed six and created eight. Round 6 fixed
eight and created thirteen, three of them Critical. Not one finding in any of those rounds was
against `-Pservices` itself, which has been clean since the first panel pass that saw it.

**The question is unanswerable from those inputs, which is why the answers kept failing.** A pid file
says a process exists, not what it bound. A port file says what a process *intended* to bind, written
a second after spawn and long before anything is listening. `lsof` says a port is busy, not whose it
is. Every round added precision to an inference over three signals, none of which carries the fact
being inferred.

**So stop inferring it and write it down.** When a run resolves the whole block, it records the
ports it resolved. A later partial start reads that record instead of recomputing the block and
guessing which services have moved off it. The recorded value can go stale — but a stale record is
visible, nameable and deletable, where a wrong inference announces itself as an observation.

**What this deletes:** `preserved`, `unconfirmed`, `publishablePort`, the widened `lsof` probe, and
the parts of `Recorded` that existed to feed them. The three Criticals, both port-related Majors and
three of the Minors go with them — not fixed, gone.

**What it keeps:** the `Recorded` tri-state where it guards *deletion* rather than inference. That
half fixed a genuine defect — `stopService` destroying the only record of a process it failed to
kill — and it is not part of the inference being removed.

**Order:** 7.1 is the record and the decision that reads it, pure and tested. 7.2 wires it in and
performs the deletions. 7.3 closes what survives the redesign.

### 7.1 The recorded block

**Files:**
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/ResolvedBlock.kt`
- Create: `/Users/tweety53/Projects/gymie/buildSrc/src/test/kotlin/com/gymie/gradle/ResolvedBlockTest.kt`

- [x] Decide a workspace's ports from three inputs and nothing else: the deterministic block, the
      recorded block if one exists, and which ports are held. No pid file, no per-service port file,
      no inference about who holds what.

- [x] A run that binds **every** service resolves the block as it does today — deterministic when
      free, discovered when foreign-held — and records what it resolved.

- [x] A run that binds a **subset** uses the recorded block verbatim where one exists, and the
      deterministic block where none does. It never re-derives a sibling's port and never discovers.

- [x] Refuse, rather than discover, when a port the run will bind is held by something that is not a
      service this run is about to stop. Naming the port and the service is the whole of the message;
      no claim is made about who holds it, because nothing here knows.

- [x] State in the KDoc that a recorded block can be stale, that deleting the record restores the
      deterministic block, and why a stale record is preferable to an inference: it can be named.

**Build:** green

### 7.2 Wire it in, and delete what it replaces

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/PortBlock.kt`
- Delete: the inference machinery named below, and its tests

- [x] Substitute `ResolvedBlock` for `portBlockDecision`'s `preserved`/`unconfirmed`/`publishablePort`
      path, and delete those three along with the tests that pinned them. A test asserting a deleted
      behaviour is deleted with it; a test asserting a surviving one is kept.

- [x] Narrow the `lsof` probe back to the ports the run will bind, and deduplicate it.

- [x] Write the record when a full resolution happens; read it when a partial one does; remove it
      when the workspace is removed, alongside the database, bucket and cache index.

- [x] Verify the announcement can no longer contradict itself — the shape it prints and the ports it
      prints come from one value, so there is nothing to drop on the way out.

**Build:** green

### 7.3 What survives the redesign

**Files:**
- Modify: `/Users/tweety53/Projects/gymie/gradle/dev-lifecycle.gradle.kts`
- Modify: `/Users/tweety53/Projects/gymie/buildSrc/src/main/kotlin/com/gymie/gradle/StopRecords.kt`

- [x] `stopService` must not delete a record when the kill it performed did not free the process.
      Deleting the only thing that could find it again is the failure `StopRecords` exists to
      prevent, and an unreadable file was only one way to reach it.

- [x] A plain `devStart` must not refuse because some unrelated service's state file is corrupt.
      **Skip the delete, say so, and carry on** — refusing the whole run to protect a record is a
      cure worse than the disease, and it broke the change's own not-breaking promise for the most
      common command there is.

- [x] `waitForHttpHealth` must not report "the process exited" when it merely failed to read the pid
      file. Retry on the next iteration; the loop already has one.

- [x] Report every service with an unreadable record in one message rather than the first, and say
      what to check for each cause — a file that cannot be read and a file holding garbage need
      different advice, and the second is not fixed by checking permissions.

**Build:** green

## 4. Cross-file verification

### 4.1 Verify the whole change

**Files:**
- Modify: none — this task only reads and reports.

- [x] Run every guard and harness in the agents repository:

```bash unverified:run once every prior task's edits are in place
cd /Users/tweety53/Projects/agents
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-task-build-green.sh
```

Expected: every one exits 0.

- [ ] Prove the isolation, which is the one claim no guard checks. With the main checkout's stack
      already running, start the worktree's stack and confirm both serve at once:

```bash unverified:run with the main checkout's stack up, then in the gymie apply worktree
docker exec gymie-postgres psql -U gymie -d postgres -c '\l' | grep gymie
docker exec gymie-redis redis-cli INFO keyspace
docker ps --format '{{.Names}}' | sort
lsof -nP -iTCP -sTCP:LISTEN | grep -E '8080|8081|3000|3001'
```

Expected: two databases listed, two Redis keyspaces populated, **the same shared containers as
before** — no new ones — and two distinct sets of application ports bound.

- [ ] Prove the backwards-compatibility promise, which is the constraint most easily broken by an
      edit elsewhere:

```bash unverified:run in the gymie main checkout with no worktree stack running
git -C /Users/tweety53/Projects/gymie diff --stat -- docker-compose.yml
./gradlew devStart && ./gradlew devStatus && ./gradlew devStop
```

Expected: `docker-compose.yml` shows **no diff at all**, and `devStart` binds 8080, 8081, 3000 and
3001 against `gymie` exactly as before.

- [x] Confirm the sandboxed installer still works, since a contract file ships with the skills:

```bash verified:read from .myflow/project.md's `## run` section at branch main
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Expected: the installer completes and `skills/myflow-contracts/workspace-isolation.md` is present
under the sandbox's installed skills.

**Build:** green
