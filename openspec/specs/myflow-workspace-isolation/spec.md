# myflow-workspace-isolation Specification

## Purpose
How a workspace id is derived from a change name, what it derives — a database, an object
store bucket, a block of application ports — what it deliberately does not derive, why the
cache index is claimed by probing, and how the resources are created and removed.

## Requirements

### Requirement: A workspace id is derived from the change name

A workspace id SHALL be derived from the **full** change name: a readable prefix followed by a
four-hexadecimal-character digest of that name.

The derivation SHALL be deterministic — the same change name always yields the same id — so that two
sessions that never communicate cannot produce the same id for different changes, or different ids
for the same change. No workspace id SHALL be recorded in the state file, and no session SHALL
coordinate with another to obtain one.

The **main checkout** SHALL have no workspace id.

#### Scenario: The same change in a later session

- **WHEN** a change derives its workspace id in one session and again a week later
- **THEN** both derivations produce the same id

#### Scenario: Two changes sharing a Jira key

- **WHEN** two changes named `kan-15-parallel-lanes` and `kan-15-parallel-two` each derive an id
- **THEN** the two ids differ, because the digest is taken over the full change name

### Requirement: An apply worktree is isolated automatically

An apply worktree SHALL receive its own isolated resources without the operator requesting them, and
the main checkout SHALL receive the project's declared defaults.

Isolation SHALL NOT be opt-in. The failure it prevents — running against another change's database
schema — is one that forgetting a flag would silently reproduce.

#### Scenario: Starting the stack in a worktree

- **WHEN** an operator starts the applications in an apply worktree without asking for anything
- **THEN** they run against that workspace's own database, cache index, object store and ports

#### Scenario: Starting the stack in the main checkout

- **WHEN** an operator starts the applications in the main checkout
- **THEN** every value is the project's declared default, exactly as before this change

### Requirement: The id derives the database, object store and ports

The workspace id SHALL derive a database name, an object-store bucket name, and a block of
application ports. Each SHALL resolve to the project's declared default value when the id is empty.

A derived database name SHALL be safe to use **unquoted**, because a call site that omits quoting
must not be able to break on it.

The shared service containers, their host ports, and the container-runtime configuration that
declares them SHALL NOT be changed. What is namespaced is the logical resource inside each service,
not the service itself.

#### Scenario: A worktree's derived names

- **WHEN** a worktree derives its resources from the id
- **THEN** it gets a database, a bucket and a port block distinct from every other workspace's

#### Scenario: The container runtime configuration

- **WHEN** this change is applied
- **THEN** the container-runtime configuration file is byte-identical to its previous state, and the
  shared services keep their current names and host ports

### Requirement: The cache index is claimed by probing, not derived

The cache index SHALL be chosen by probing which indices are in use and claiming a free one, and
SHALL be recorded for the duration of that run.

It SHALL NOT be derived from the digest. The index space is small enough — sixteen — that a derived
value collides at roughly six percent for two concurrent workspaces, and a collision silently shares
sessions and cache, which is the failure class this capability exists to remove.

#### Scenario: Two workspaces starting in turn

- **WHEN** a second workspace starts while a first holds an index
- **THEN** the second claims a different free index

#### Scenario: The claimed index within a run

- **WHEN** a workspace has claimed an index
- **THEN** every process it starts for that run uses that same index

### Requirement: A workspace's database and bucket are created on demand and start empty

Starting the applications in a workspace SHALL create that workspace's database and bucket if they
do not exist, and SHALL report having done so the first time, so the side effect is visible rather
than silent.

A newly created database SHALL start **empty** and be brought up to date by the project's normal
migration and seeding path — the same path a fresh machine takes — rather than by copying an
existing database.

#### Scenario: First start in a new worktree

- **WHEN** the applications are started in a worktree whose database does not yet exist
- **THEN** the database and bucket are created, the operator is told, migrations run against the
  empty database, and seeding follows

#### Scenario: Subsequent starts

- **WHEN** the applications are started again in the same worktree
- **THEN** the existing database and bucket are reused and no creation notice is printed

### Requirement: Database maintenance tasks target the workspace's own database

A project's database reset and seed tasks SHALL default to the workspace's own database when run
from an apply worktree, and to the project default when run from the main checkout.

#### Scenario: A destructive reset from a worktree

- **WHEN** a destructive database reset is run from an apply worktree
- **THEN** it targets that workspace's database, not the shared default

### Requirement: Application URLs follow the workspace's ports

Any configuration a run generates for the applications SHALL derive every application URL from that
workspace's port block.

**An application whose port is fixed outside the declaring project's own repository is not
isolated, and SHALL be recorded as a stated limitation rather than left to be discovered.** The
project declares the ports it can move; one it cannot move keeps its default, and every URL derived
from that port keeps its default with it. Isolation is not claimed for what a project cannot
control, since a claimed isolation that does not hold is the silent wrong answer this capability
exists to remove.

#### Scenario: A worktree's frontends

- **WHEN** a worktree's frontends are started
- **THEN** they reach that worktree's backend, not whichever workspace holds the default port

#### Scenario: An application whose port the project cannot move

- **WHEN** an application's port is fixed in a repository the change does not touch
- **THEN** that application keeps the default port, the limitation is stated in the project's own
  declaration, and no isolation is claimed for it

### Requirement: The test suite is unchanged

The project's test command SHALL be unaffected by this capability. Where a project's tests already
obtain an isolated database by other means, that mechanism SHALL be left exactly as it is, and the
pipeline's verification step SHALL NOT change.

#### Scenario: Running the test suite in a worktree

- **WHEN** the project's test command is run in an isolated worktree
- **THEN** it behaves exactly as it does in the main checkout

### Requirement: A workspace's database and bucket are removed at finish, and a missing service is skipped

The workspace's database and bucket SHALL each have a row in the pipeline's temporary artifacts
registry, and `/myflow-finish` run 2 SHALL remove them.
`scripts/check-cleanup-complete.sh` SHALL carry a marker line for the row and report a survivor as a
leftover.

**A survivor SHALL be established by asking, never inferred from the removal's exit code.** A
project therefore declares a command that reports which of its workspace resources still exist,
alongside the commands that create and remove them, and the cleanup guard reads that command's
answer. A removal that reported success against a stale connection, and a resource a policy refused
to delete, both leave the registry row's promise broken with nothing having failed — so *ran the
removal* is not *verified gone*, and the guard SHALL NOT treat it as such. A project that declares
no such command has its verification reported as **skipped**, never as passed.

#### Scenario: A removal that succeeded and a resource that survived it

- **WHEN** run 2's removal command exits 0 but a workspace resource still exists
- **THEN** the survivor is reported as a leftover, and the terminal state is not written

**A service that is not running SHALL be reported and skipped, not failed.** An absent service means
there is nothing to remove at that moment, and the run SHALL continue to `FINISHED` — matching how
the existing project-supplied stop check already treats an unreachable stack.

#### Scenario: Finishing with the database service running

- **WHEN** run 2 reaches cleanup and the database service is running
- **THEN** the workspace's database and bucket are dropped, and the cleanup verification passes

#### Scenario: Finishing with the database service stopped

- **WHEN** run 2 reaches cleanup and the database service is not running
- **THEN** the removal is reported as skipped, and the run continues to `FINISHED`

#### Scenario: A change that was never isolated

- **WHEN** run 2 reaches cleanup for a change whose project declares no isolation
- **THEN** the check passes, because a step whose artifact is already absent is a success

### Requirement: Projects declare their isolation through an optional configuration section

`.myflow/project.md` MAY carry a `## workspace isolation` section naming the resources a workspace id
derives, the default value of each, and the commands that create and remove them.

A project that declares no such section SHALL behave exactly as it does today, everywhere, including
in apply worktrees. This SHALL NOT be reported as a misconfiguration.

An entry in that section SHALL be resolved under the same rule the project configuration contract
already applies to the same kind of entry, and an entry that fails SHALL be reported by name and
dropped rather than repaired.

**Which rule applies follows what the pipeline does with the entry, not which section it sits in.**
An entry the pipeline **reads as a file** SHALL be contained exactly as the contract's existing
path entries are. An entry the pipeline **executes** SHALL be treated as the project's other
declared commands are, which carry no path containment: isolating a path inside an arbitrary
command means parsing a shell command, and a parser that is wrong about one entry reports the wrong
thing about all of them. An entry whose value the pipeline **substitutes into** SHALL have its
shape validated — the tokens it may carry, and no others.

Containment exists to bound an arbitrary-file-read primitive, which executing a declared command
does not create: a project that declares a command to start its applications can already run
anything through that declaration. Extending path containment to a command would therefore be a
rule that reads as a safeguard while bounding nothing, which is worse than the narrower rule stated
here.

#### Scenario: A declared command that names no path

- **WHEN** a project's create, remove or survivor-reporting command contains no filesystem path
- **THEN** it is executed as the project's other declared commands are, and no containment check is
  attempted on it

#### Scenario: An entry carrying a token the contract does not name

- **WHEN** a declared value or command carries a substitution token the contract does not define
- **THEN** it is reported by name and dropped, rather than passed through with the token unresolved

#### Scenario: A repository with no runnable application

- **WHEN** a project declares no `## workspace isolation` section
- **THEN** every command runs exactly as it does today, and nothing is reported as misconfigured

#### Scenario: A malformed isolation entry

- **WHEN** an entry in `## workspace isolation` fails the project configuration containment rules
- **THEN** it is reported by name and dropped, and no value derived from it is used
