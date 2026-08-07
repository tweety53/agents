# Review panel — kan-15-parallel-myflow-do-task-lanes

## Pass 1 — the full roster

Every optional slot fired: the diff is ~3,900 lines across two repositories, builds SQL and executes
commands read from a tracked file, signals processes, adds concurrency in port and cache-index
claiming, and modifies tests.

| Slot | Model | Verdict |
|---|---|---|
| 0 — primary | Sonnet | spec coverage complete; 1 Major, 1 Minor |
| 1 — defect hunt | Sonnet | 1 Critical, 1 Major |
| 2 — principles (Merged) | Sonnet | clean; 2 Minor |
| 3 — security | Sonnet | no Critical or Major; 2 Minor latent |
| 4 — adversarial | Sonnet | 1 Critical, 2 Important, 1 Minor |
| 5 — principles (Lens B) | Sonnet | 2 Major, 1 Important, 2 Minor |
| 6 — principles (Lens C) | Sonnet | 1 Critical, 2 Major, 3 Minor |

**Slots 1 and 3 were adapted.** This harness exposes no `bugbot` or `security-review` agent type, so
both ran as general-purpose agents against the skill's own `bug-hunter-reviewer-prompt.md` and
`security-reviewer-prompt.md`, on the panel's model. Their model is recorded as Sonnet rather than
`unknown (agent-defined)` because the dispatch chose it.

**Three slots found the same Critical independently** — the cache-index claim. Slot 5 found the
missing lock, slot 1 found that the claim is invisible to any other workspace until a key exists,
and slot 4 established that the collision is the *default* outcome for a burst of fresh workspaces
rather than a rare race, because the algorithm always prefers the lowest free-looking index.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Lens B, defect hunt, adversarial | Critical | `gradle/dev-lifecycle.gradle.kts` `claimCacheIndex` | probe-then-claim with no lock; the claim is written only to a per-worktree file and never to the cache, and `INFO keyspace` lists only indices that already hold a key — so two fresh workspaces reliably claim the same index. This is the silent shared-cache failure probing was chosen to prevent |
| F2 | defect hunt | Critical | `gradle/dev-lifecycle.gradle.kts` `startService`/`freePort` | `freePort` runs for every service including the KMP frontend, which is deliberately not port-isolated, so its resolved port is always the shared 3000. A second worktree's `devStart` silently kills a sibling worktree's frontend — on the very use case the change exists to enable |
| F3 | Lens C | Critical | `skills/myflow-contracts/pipeline.md` run 2 step 6 table | the `COMPLETE:` row says "report the cleanup as verified" and never requires naming an appended `SKIPPED:` note, so a skip can reach the operator as a pass — the outcome the guard's own header argues must be impossible |
| F4 | Lens C | Major | `scripts/check-cleanup-complete.sh` `run_survivors` | the bounded wait kills the host-side process group; a command reaching a service through `docker exec` may leave the in-container process running. Test 28b only exercises a host-native `sleep`, never the shape gymie actually declares |
| F5 | Lens B, Lens C | Major | `gradle/dev-lifecycle.gradle.kts`, `skills/myflow-contracts/pipeline.md` registry | nothing enumerates orphaned workspace databases and buckets, and the cache index has no removal path at all — 15 slots, no registry row, no sweep, no documented recovery |
| F6 | primary | Major | `openspec/changes/.../tasks.md` task 4.1 | the two-stacks-at-once proof has not been run |
| F7 | adversarial | Major | `skills/myflow-contracts/project-configuration.md` resource table | the resource-table validation rules — closed vocabulary, four cell forms, `<value:…>` resolution, bare-integer defaults, refuse-rather-than-fall-back — exist only as prose an agent re-performs each run, with no mechanical check anywhere |
| F8 | principles, adversarial | Major | `buildSrc/src/test/kotlin/com/gymie/gradle/WorkspaceTest.kt` | the Kotlin pin asserts against literals transcribed from the contract's comments, where the guard's equivalent drives the canonical block itself; a contract edit can leave Kotlin pinning the wrong answer with every test still green |
| F9 | security | Minor | `scripts/check-cleanup-complete.sh` change-name gate | the `case` bracket expression is a collation range, so it admits different names under different locales; the fix pattern already exists two functions later in the same file |
| F10 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts` `databaseExists`/`ensureWorkspaceDatabase` | the new creation path skips the `isContainerRunning` pre-check `dbReset` already performs, so a stopped container surfaces a raw docker stderr blob instead of the existing guidance |
| F11 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts` `warnIfGoogleOAuthNotConfigured` | warns when credentials are placeholders, but says nothing when a workspace is isolated *and* real credentials are set — the case where the unregistered redirect URI actually bites |
| F12 | Lens C | Minor | `scripts/check-cleanup-complete.sh` | the SIGTERM→SIGKILL grace is a hardcoded 2s with no override and no stated rationale, in a file that justifies every other constant |
| F13 | principles | Minor | `src/app/src/main/resources/application-local.yml` | five defaults are duplicated from `application.yml` on self-documentation grounds, with nothing asserting the pairs stay in step |
| F14 | adversarial | Minor | `gradle/dev-lifecycle.gradle.kts` `workspaceChangeName` | `-PworkspaceChange` is also reachable as `ORG_GRADLE_PROJECT_workspaceChange`, so a leaked environment variable silently isolates or de-isolates any checkout; the destructive paths still gate on validation, so this is surprising rather than dangerous |
| F15 | Lens B | Minor | `scripts/check-cleanup-complete.sh` | `CHECK_CLEANUP_SURVIVORS_TIMEOUT` is a documented runtime knob that exists so the test suite need not spend a minute per timeout case |

## Pass 2 — full roster, escalated automatically

**Not a targeted re-run, and not a judgment call.** Three of the escalation triggers fired at once:
the fixes touched files outside the set the findings named, the fix diff exceeded ~150 changed lines
by more than an order of magnitude, and — decisively — **they altered public contracts**.
`workspace-isolation.md`'s cache-index section was rewritten around a new claim protocol, and
`project-configuration.md` gained a statement of which of its rules a script now enforces. A targeted
re-run would have asked the slots that raised findings to check their own fixes while nothing
re-examined the contracts those fixes changed.

Pass 2 also carries something pass 1 never saw: `scripts/check-workspace-isolation.sh`, a 564-line
guard with a 992-line harness, created **in response to F7** and therefore reviewed by nobody. A
fix that is itself a new capability is not closed by the finding that prompted it.

| Slot | Diff read |
|---|---|
| 0 — primary | both, rebuilt |
| 1 — defect hunt | both, rebuilt |
| 2 — principles (Merged) | both, rebuilt |
| 3 — security | both, rebuilt |
| 4 — adversarial | both, rebuilt |
| 5 + 6 — principles (Lens B, Lens C) | both, rebuilt — one dispatch, two separately reported lenses |

Lenses B and C share one dispatch in this pass because they read the same material and pass 1
established they do not overlap in what they find; the dispatch requires two separate sections and
two separate verdicts, so no finding is merged away.

### Pass 2's findings

**Two of pass 1's fixes were themselves defective, and both were found by attacking the fix rather
than re-reading the finding.** That is the argument for a full re-run over a targeted one: a
targeted pass asks the slot that raised a finding to check its own repair.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F16 | Lens C, adversarial | Critical | `scripts/check-workspace-isolation.sh`, `.myflow/project.md` lint lists | the guard built to close F7 runs only in **this** repository's lint list, and this repository declares no isolation section — so it is a permanent vacuous pass, and nothing routes gymie's declaration through it. `project-configuration.md` meanwhile claims the rules are "mechanically enforced". "The guard passed" and "the guard never ran" are indistinguishable — F3's shape, one layer up |
| F17 | adversarial | Critical | `gradle/dev-lifecycle.gradle.kts` `portToSweep`/`stopService` | F2's fix reopened the sibling-frontend kill through the **stop** path: a refused service writes no pid or port file, so a later unrelated `devStop` falls back to the declared shared port and kills whichever workspace holds it. Fires on a command the operator has no reason to connect to the victim |
| F18 | primary | Major | `gradle/dev-lifecycle.gradle.kts` `claimServicePort` | the refusal branches on `service.portIsolated` and never on `workspace.isolated`, so the **main checkout** lost its frontend self-heal and is told "another checkout's stack" holds a port that can only be its own orphan |
| F19 | security | Major | `scripts/check-cleanup-complete.sh` verdict assembly | survivor output is spliced into the verdict with only `\r` stripped, so a declared command can emit terminal escapes that erase the real `LEFTOVER:` line and paint a fake `COMPLETE:` — against exactly the text F3's fix requires an operator to read |
| F20 | security | Major | `scripts/check-workspace-isolation.sh` readability gate | `[ -r ]` is true for a directory and `grep`'s error is conflated with "no match", so a `.myflow/project.md` symlinked to a directory yields `ISOLATION-OK`, exit 0 — a fail-open in the script written to prevent one |
| F21 | security | Major | `buildSrc/.../CacheIndex.kt` `cacheIndicesHoldingKeys` | any key in a claimable index removes it from the claimable set, and with no claim key to match, no shipped tooling can reclaim it — fifteen writes wedge the pool durably |
| F22 | defect hunt | Major | `buildSrc/.../CacheIndex.kt` `claimCacheIndex` | one stale `indicesHoldingKeys()` snapshot plus `SET NX`, which tests existence and not value, lets **one** workspace claim two indices; `firstOrNull` in release and survivor reporting means only one is ever seen again |
| F23 | Lens C, adversarial | Important | `buildSrc/.../CacheIndex.kt`, `scripts/check-workspace-isolation.sh` | no warning before pool exhaustion; and a bracketed token containing a space does not match the token regex, so the likeliest human typo is read as literal text rather than refused |
| F24 | primary, Lens C | Minor | the harness, `devStatus`, the give-back | a case gated on `id -u` rather than the runtime condition and measurably flaky; a shared-port refusal invisible after the run that printed it; a check-then-act give-back whose safety rests on an unwritten invariant |
| F25 | defect hunt | Minor | `scripts/check-workspace-isolation.sh` `flush_block` | two tables with no separator merge into one block and produce four misleading cell-count violations that never name the real cause |

## What the panel endorsed

Recorded because an endorsement is evidence too, and because two of these were deliberate
decisions taken under review earlier in the run.

- **The defence in depth around the workspace id held under direct attack.** Security attacked both
  layers — the anchored whole-id shape check and the independent SQL escaping — with real injection
  payloads and got past neither.
- **The anti-drift pin is real, not decorative.** Adversarial widened the guard's digest from four
  hex characters to five and watched 23 assertions fail, byte for byte, across driven names and
  locales; it also mutated `PORT_OFFSET_STEP` and confirmed `jar.finalizedBy(test)` fires on an
  ordinary `./gradlew help`.
- **No fourth id-derivation divergence exists.** The defect hunt compared all three implementations
  line by line after being told three had already been found.
- **The empty-id path is inert**, verified independently by three slots.
- **Automatic isolation rather than opt-in**, **derive-never-record for the id**, and
  **discard the whole port block rather than patch it** were each examined and endorsed as the right
  call, not merely accepted.

## Verification pass over fix round 2

Pass 1's fixes were caught out by pass 2. **Pass 2's fixes had been reviewed by nobody**, so a
focused verification read fix round 2's diffs alone rather than the whole branch. It found four
more, one of them Critical — and the Critical is the *same defect class for the third time*, each
appearance inside the repair of the one before it.

| ID | Severity | Location | Note |
|---|---|---|---|
| F26 | Critical | `gradle/dev-lifecycle.gradle.kts` `startService` liveness check | the shared-port owner record is deleted **unconditionally** when a launched process dies inside the one-second window. It removes the file rather than this checkout's claim on it, so a losing racer's cleanup unattributes the *winner's* running frontend and the next `devStart` anywhere takes it. F2 through the start path, F17 through the stop path, F26 through the failure path |
| F27 | Major | `skills/myflow-do/SKILL.md`, `project-configuration.md` | the wiring says the installed `skills/` directory is a symlink into the checkout; measured, it is a real directory holding one symlink per skill — what this repository's own guard calls "a farm of symlinked directories". The citation for resolving the agents repo points at a file that never mentions it, and a project-local install has no symlink at all |
| F28 | Major | both guards' `sanitize_display` | the headers argue that escaping `\` is what makes the encoding injective; deleting that line leaves every suite green. Mutating the C0 loop fails immediately, so the harnesses can catch this — they just never build the pair that pins it |
| F29 | Minor | `scripts/check-cleanup-complete.sh` ref checks | `git show-ref --verify --quiet` reports an unreadable ref exactly as it reports an absent one, and git offers no channel that separates them — so the file's "a command's failure is never read as its negative answer" invariant has an exception it does not name |

**Three rounds, three appearances, three different call sites.** That is evidence about the shape of
the code rather than about the care of the people writing it: ownership-checked deletion is a rule
three call sites each have to remember, and each round a new site forgot it. The fix dispatched for
F26 asks for the rule to become a function every caller goes through, not a fourth reminder.


## Pass 3 — scoped, and the scope is recorded rather than assumed

Since pass 2 the branch gained **group 5**: a third repository, the KMP frontend's port rotated, and
**786 lines deleted** — the shared-port ownership machinery. Each of 5.1, 5.2 and 5.3 had its own
per-task review. Pass 3 asks what no per-task review can: whether the three hold **together**, and
whether the branch is still what the spec asked for after so much editing.

Three slots ran, not seven, and the two omissions have reasons rather than a budget:

| Slot | Ran | Why |
|---|---|---|
| 0 — primary | yes | spec coverage and cross-artifact consistency after group 5 |
| 1 — defect hunt | yes | the new three-repository seam and the deletion's edges |
| 5 — principles (Lens B) | yes | a deletion is a simplicity question, and Lens B raised the original cache-index Critical |
| 3 — security | no | group 5 **removes** attack surface — a state file written outside every repository is gone — and adds one file that reads an environment variable |
| 4 — adversarial | no | 5.3's own review was adversarial in method: it traced every deletion to its call sites and proved the restored `freePort`/`portToSweep` algebraically equivalent to the pre-machinery original at `HEAD` |
| 2, 6 — principles Merged, Lens C | no | neither the hard invariants nor the operational surface changed shape; the deletion removed an ops concern rather than adding one |

**What pass 3 found.** Lens B: one Minor, and the judgement that the design is now simpler than the
problem rather than merely smaller — *"it didn't reduce the cost of a hazard, it deleted the
hazard."* Primary: ready for the gate, one Minor. Defect hunt: one Major, below.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F30 | defect hunt | Major | `gradle/dev-lifecycle.gradle.kts`, the `frontend` service | the only one of four with no `healthPath`, so its moved port is never verified at runtime. `FRONTEND_PORT` is a cross-repository, environment-only contract; a `gymie-frontend` checkout predating this change binds 3000 while gymie reports the offset port and calls the service started |
| F31 | primary | Minor | `proposal.md`, `design.md`, `tasks.md`, the test guide | six claims that the override is "four lines"; the delivered file is ten, because the implementer added validation the plan did not predict. Spelled as a word, so `check-plan-provenance` — which matches digit runs — could not see it |
| F32 | Lens B | Minor | `project-configuration.md`, the `url` row grammar | fix round 3 widened a `url` cell to accept `<id>`/`<id_underscored>`; gymie's real declaration uses neither, and the only example exercising it was invented in the same diff |
| F33 | defect hunt | Minor | `buildSrc/.../DevEnvTest.kt` | a test named for the opposite of what it asserts — it says a valid port "exports as the empty string" and asserts `isNotEmpty` |

## Pass 4 — full roster, escalated automatically

Task group 6 added `-Pservices`, so the panel ran again. **Escalated without asking**, on two of the
recorded triggers at once: the round's diff is about a thousand changed lines, and three fix rounds
had already run.

**The agents worktree was not re-reviewed, and that is a scoping decision rather than an omission.**
Group 6 is gymie-only; the agents diff is byte-identical to the one pass 3 cleared — 18 files,
+5198/−50 — so a re-run there could only reproduce a result already on record. The gymie diff was
rewritten in full and every slot read it, with the round's own `fix-round-4.diff` named as the part
to weight.

Three slots reported clean: primary, principles (Merged) and security. **Two slots reached the same
defect independently** — the defect hunt and Lens C both found `restartDevStack` stopping the
selection before any check that can still fail the run — recorded as one finding, F35, rather than
two. The adversarial slot disproved none of the seven claims it was given and found its Important by
attacking what those claims left unsaid.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F34 | defect hunt | Major | `buildSrc/.../PortBlock.kt`, `gradle/dev-lifecycle.gradle.kts` | a block entry counts as a clash only while the port is *held*, so a service that took a discovered port earlier is silently reported on its deterministic one once that slot frees up. The wrong port then reaches `BACKEND_URL`, the admin panel's `.env.local` and the announcement, with no warning anywhere |
| F35 | defect hunt + Lens C | Major | `gradle/dev-lifecycle.gradle.kts`, `restartDevStack` | the selection is stopped before `resolveWorkingDirs`, `ensureWorkspaceResources` and the port-block decision, each of which can throw. A refused partial start leaves the named service down and prints "Nothing is started" — and on the port-block path, re-running reproduces it exactly |
| F36 | adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `resolveWorkspacePorts` | clash detection spans every declared service rather than the ones this run will bind, so two concurrent partial restarts in one workspace can refuse each other on a transient — and the remedy printed, stop the whole stack, is wrong for one |
| F37 | Lens B | Important | `buildSrc/.../DevLogs.kt`, `gradle/dev-lifecycle.gradle.kts` | a fifth hardcoded copy of the service names and their order, beside a comment added this round claiming there is only one place a command learns them |
| F38 | adversarial | Minor | `gradle/dev-lifecycle.gradle.kts`, `resolveWorkingDirs` | every selected checkout is resolved — each logging as it goes — before any is checked, so a missing frontend now prints the admin panel's auto-detect line before failing. The pre-refactor code stopped at the first |
| F39 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts`, the `TrackedService` build | `.exists()` then `.readText()` with nothing between them; a concurrent stop removing the file turns a readable refusal into a raw Java stack trace |

## Pass 5 — full roster again, and every slot pointed at the repair

Fix round 5 tripped three escalation triggers at once — it altered a public signature (`portBlockDecision` gained a parameter and its `Keep` gained a field), moved a refusal out of `startService`, and added a file none of the findings named. So the roster ran in full again, and **every prompt named the repair as the thing to attack rather than the finding**, because pass 2 of this panel had already found two of its own fixes defective. That instruction is what produced this pass's findings: three slots reported clean, and the four that did not were all reading the repair.

**One root cause underlies three of the eight.** `StateFile.kt`'s `recordedText` — itself fix round 5's answer to F39 — collapses *absent*, *empty* and *unreadable* into one `null`. That is right for a caller asking whether something is there and wrong for a caller that needs the number, and fix round 5 introduced two callers of the second kind without noticing the difference. The adversarial slot stated it exactly: the callers that need *presence* are safe and the ones that need *accuracy* are not, so "I do not know this port" silently became "this service has no recorded port", which the decision then answers with the block value. **A fix for one finding became the cause of three.**

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F40 | security | Important | `buildSrc/.../PortBlock.kt`, `live` | a preserved sibling's port is trusted from a file with nothing checking anything is listening on it — `live` is gated on an alive pid, not on a bound socket. That number is **written** into the admin panel's `.env.local` on a first run and into the OAuth redirect URI, so a stale record is baked into generated config |
| F43 | Lens C + adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `stopService` | the pid and sweep-port reads both go through the swallow, and the two `delete()` calls then run unconditionally. A read that fails for any reason other than a concurrent stop leaves the process alive and bound, logs "not running", and destroys the only record of it. **A regression this round introduced** — the previous code threw, which was ugly but left the breadcrumb |
| F46 | adversarial | Important | `buildSrc/.../PortBlock.kt`, `live` → `preserved` | a live sibling whose port file is unreadable is dropped from `live`, so the resolved map claims it is on its block port. After an earlier `Discover` it is not, and nothing refuses — `heldPorts` only probes the block's own numbers. The wrong port reaches `GATEWAY_PORT`, the admin `.env.local`, the announcement and `devWorkspace`. No test reaches the shape: every `recordedPort = null` case pairs with `starting = ALL`, where `preserved` is always empty |
| F47 | adversarial | Important | `buildSrc/.../PortBlock.kt`, `released` | the same collapse on a **starting** service makes it absent from `released`, so the clash filter fires on the port it is itself holding. Alone, the run discards the block for an ordinary restart; with a sibling up, it refuses and **names that sibling** for a clash it had no part in |
| F41 | Lens B | Minor | `buildSrc/.../PortBlock.kt`, `Keep.preserved` | derivable from `ports` and `block`, which the sole caller already holds. Raised with the reviewer explicitly declining to ask for the change, as the answer to a question this panel put to it |
| F42 | Lens B | Minor | `buildSrc/.../PortBlock.kt` | `describeClashes` and `describePorts` differ only in per-entry wording. Likewise raised with an explicit "not asking for a change" |
| F44 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts`, `restartDevStack` KDoc | claims a run that cannot complete "stops nothing", which holds for the refusal checks and not for a failure found by attempting a start — `freePort`, the liveness check and the health timeout all sit after the stop loop |
| F45 | defect hunt | Minor | `gradle/dev-lifecycle.gradle.kts`, `announceWorkspace` | `portsAreDeterministic` is `preserved.isEmpty()`, so one preserved sibling labels the whole block "not this workspace's block" while three services are exactly on it |

**Two Minors were raised by a reviewer that did not want them acted on**, F41 and F42, and they are recorded anyway. A finding's severity governs whether it blocks the handoff; the reviewer's appetite for the fix does not. What that costs is one judgement each in the fix round and, if it stands, a handback — which is the designed route and is cheaper than a record that quietly drops what a reviewer said.


## Pass 6 — four slots, chosen by the operator, and the pass that ended the approach

**The roster was cut on evidence, by the operator's decision at the pass 5 handback.** Primary and
the two other principles lenses had reported clean twice running while every finding in passes 4 and
5 came from the same four slots, so those four ran: defect hunt, adversarial, Lens C and security.
The reduction is recorded as a decision rather than a saving — it is a bet that a slot's history
predicts its yield, and this pass is the evidence for it: **13 findings, three of them Critical, from
four reviewers.**

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F48 | adversarial | Critical | `gradle/dev-lifecycle.gradle.kts`, `resolveWorkspacePorts` | `offBlockServices` is built from `preserved.keys` alone and drops `unconfirmed`, so `devWorkspace` prints "this run cannot tell which port it is serving on" and two lines later "deterministic block". Deterministic, no race — and unreachable by any test, because the function lives in the untested shell |
| F49 | defect hunt + security + adversarial | Critical | `buildSrc/.../PortBlock.kt`, `publishablePort` | the carve-out returns a recorded port equal to the block port with **no listener check**. `startService` writes the port file a second after spawn, so a service whose health check later timed out leaves a record naming a dead port — republished with no notice. A new test pinned the behaviour as intended |
| F50 | adversarial | Critical | `gradle/dev-lifecycle.gradle.kts`, `stopService` | the records are deleted even when `killPidTree` failed to kill — `destroyForcibly` neither throws nor guarantees termination. The exact catastrophe `StopRecords.kt` was written to prevent, reached through the gap between kill and delete rather than through an unreadable file |
| F51 | defect hunt + adversarial | Major | `buildSrc/.../PortBlock.kt`, the `unconfirmed` → `Refuse` branch | a sibling still compiling — port file written, socket not yet open — makes an unrelated run refuse. An ordinary two-terminal sequence, not a race |
| F52 | defect hunt | Major | `gradle/dev-lifecycle.gradle.kts`, the `devWorkspace` task | the unconfirmed guess is printed as `KEY=VALUE` beside confirmed ones, with the caveat in a differently-shaped line elsewhere. `.myflow/project.md` documents that output as an interface for other tooling to read |
| F53 | Lens C | Important | `buildSrc/.../PortBlock.kt`, `unlocatableMessage` | tells the operator to run `devStop`, which for the unreadable cause hits `refuseUnreadableStopRecords` and refuses in turn. The documented recovery is defeated by another refusal |
| F54 | Lens C | Important | `gradle/dev-lifecycle.gradle.kts`, `waitForHttpHealth` | a third caller of the nullable pid read that the KDoc's safety argument never named: one unreadable read reports "process exited before health check passed" and skips the loop's own retry, sending the operator to look for a crash that did not happen |
| F55 | adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `refuseUnreadableStopRecords` | a plain `devStart` now refuses outright when **any** service's record is corrupt, where every prior round shrugged. Hardening, and a break of this change's own not-breaking promise for its most common command |
| F56 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts` | `probed` is not deduplicated, so the steady-state case probes each block port twice — the concrete cost of four `lsof` calls becoming eight |
| F57 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts`, `refuseUnreadableStopRecords` | throws on the first bad record, so two bad records cost two fix-and-rerun cycles |
| F58 | Lens C + adversarial | Minor | `buildSrc/.../StopRecords.kt` | the message says "could not read" and advises checking permissions and free space, but `Unknown` also covers a file that read fine and held garbage — where the fix is to delete it, which the message never says |
| F59 | defect hunt | Minor | `gradle/dev-lifecycle.gradle.kts`, `stopService` | the surrounding claim that a refused run "stops nothing" holds only for a failure present at the batch check; one appearing inside the window leaves exactly the half-stopped stack the check was added to prevent |
| F60 | adversarial | Minor | `buildSrc/.../PortBlock.kt`, `released` | an `Unknown` port record credits the block port as released, which is wrong for a service that had moved off-block. Absorbed by `freePort`'s blanket kill, so the clash-detection under-run is invisible rather than harmful |

## Pass 7 — the redesign under review

Group 7 deleted the inference — `PortBlock.kt` and its 29-test suite, 885 lines — and replaced it with
a recorded block. Four slots read it: defect hunt, adversarial and Lens C as before, plus **Lens B,
brought back on purpose**. Its earlier verdict in this change was the one that named the pattern
(*"it didn't reduce the cost of a hazard, it deleted the hazard"*), and a design reversal is exactly
the judgement that lens exists to make.

**The design was endorsed and the implementation was not.** Lens B recomputed the line accounting
independently rather than accepting it, confirmed −256 in the port-block region and −12 net, grepped
the tree for every deleted symbol and found no vestige, and judged the remaining complexity linear
where the inference had been combinatorial. The adversarial slot separately established that the
record's write ordering is **self-healing across a crash** — the write precedes any process work, so
an interrupted run either re-discovers cleanly or leaves a half-written file that reads back as
`Unknown` rather than as truth.

Then both of them found a Critical in what the code actually does, and neither is about the idea.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F63 | defect hunt | Critical | `buildSrc/.../ResolvedBlock.kt`, `resolveFromRecord` | a partial run's path has **no `heldPorts` parameter at all**, so nothing asks whether a port it is about to bind is occupied — `startService`'s `freePort` simply kills the holder. Group 7's own plan requires a refusal there, and the two deleted tests that pinned it were not replaced |
| F64 | adversarial | Critical | `buildSrc/.../ResolvedBlock.kt` `resolveWholeBlock`, with `gradle/dev-lifecycle.gradle.kts` `startService` | the decision assumes a full run rebinds every service; `startService` skips any service already running. So an ordinary repeat `devStart` against an unreadable or incomplete record reads its own healthy services' ports as clashes, invents four ports, **writes them to the record as truth**, skips the running services, and puts a wrong gateway URL in the admin panel's `.env.local` |
| F61 | Lens B | Important | `buildSrc/.../ResolvedBlock.kt` | the "is this a full run" predicate — the hinge the file turns on — is written twice as two independent copies |
| F65 | Lens C + defect hunt + adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `killPidTree` | the boolean is untrustworthy two ways: a dead or absent handle returns `true` without ever examining descendants, and the tree is snapshotted once so a process forked during the kill window is never signalled and never counted. `stopService` then deletes the pid record on a false "freed" — the record `StopRecords` exists to protect. The port half is already honest, because it re-verifies live |
| F66 | adversarial + defect hunt | Important | `gradle/dev-lifecycle.gradle.kts`, `removeWorkspacePortBlockRecord` | dead code on the documented path: run 2 invokes the removal **from the main checkout**, where the guard's `workspace().id` is empty, so the file is never removed. It logs nothing, unlike all three sibling removals, and `workspaceSurvivors` never looks for it — so the cleanup guard reports `COMPLETE` with it stranded |
| F67 | adversarial | Important | `buildSrc/.../ResolvedBlock.kt`, `resolveFromRecord` | `Recorded.None` returns the deterministic block with **no caveat**, conflating "confirmed nothing moved" with "never checked". A workspace predating this change has exactly that shape, with its services on discovered ports |
| F68 | adversarial | Important | `buildSrc/.../ResolvedBlock.kt`, `usableRecord` | all-or-nothing: a record missing one newly-added service discards the entries it had right, and `devWorkspace` — which documents itself as reporting the port a service is actually on — reports numbers never bound |
| F69 | Lens C | Important | `buildSrc/.../ResolvedBlock.kt`, the three recovery messages | they name the bare filename, not a path, so an operator is told what to delete and not where. `StopRecords.describe()` in this same round prints the absolute path, with a docstring saying "the path is what the operator opens" |
| F62 | Lens B | Minor | `buildSrc/.../ResolvedBlock.kt`, `Resolved.record` | equals `ports` when non-null and null otherwise, so the only fact it carries is "may this be persisted" — F61's predicate again, restated as data |
| F70 | Lens C | Minor | `gradle/dev-lifecycle.gradle.kts`, `statusService` | the block's shape and staleness notice print only from `announceWorkspace`, reachable only from a mutating command, so an operator cannot check the record before deciding to restart |

**What separates this pass from the two before it.** Passes 5 and 6 found the *idea* failing — an
inference over three signals that do not carry the fact. Pass 7 found a sound idea implemented with
cases missing, and said so: one lens endorsed the design outright, and the adversarial slot's crash
analysis came back clean. F64 is the one worth remembering, because it is not a defect in either
half — the decision is right about what it decides, and `startService`'s skip-if-running is right
about what it skips. It is a defect in the **seam**, which is why a pure unit suite over the decision
and a shell with no test harness could both be green while the two disagree about what will happen.


### What this pass decided

**The approach was ending, and the count is how it showed.** Round 5 closed six findings and its
repair produced eight. Round 6 closed eight and its repair produced thirteen. Not one finding in
either round was against `-Pservices`, the feature the operator asked for — every one was against the
machinery built to answer *which port is a service this run is not touching actually serving on*, by
inference over a pid file, a per-service port file and `lsof`.

**None of those three carries the fact.** A pid file says a process exists, not what it bound. A port
file says what a process *intended* to bind, written about a second after spawn and left in place
when the health check later times out — which is precisely F49's trigger and F51's. `lsof` says a
port is busy, not whose it is. Each round made the inference more precise over inputs that cannot
answer the question, and F48 is the reduction to absurdity: three rounds of increasing precision, and
the untested shell discarded the result on the way out.

The operator was given the count, the trend and the alternative, and chose to **record the resolved
block and delete the inference** — group 7. Most of these thirteen are closed by deletion rather than
by repair. What survives the redesign is F50, F54, F55, F57 and F58, none of which is about
inference: they are about not destroying a record, not misreporting a crash, not refusing a run that
should proceed, and telling an operator which of two different problems they have.

**Pass 3 recorded this same shape once already, in this same change** — *"it didn't reduce the cost of
a hazard, it deleted the hazard"* — and it took three rounds to recognise the second instance. The
first was found by the operator asking a question. So was this one.

## Pass 8 — the verification pass

Two slots, the two that found pass 7's Criticals, each pointed at the repair. **Both confirmed pass
7's Criticals genuinely closed** — by tracing the new call graph by hand rather than by re-reading the
findings, which is the check this panel has learned to insist on: three fix rounds in this change have
been defective in their own repair.

One new Critical, and it is the change's oldest defect class wearing a new shape. The per-service
partial-record accounting added for F68 composes a record over the deterministic block **without
checking the result is injective**, so a record naming one service at another's deterministic port
puts two services on one port — and `freePort` then kills a live sibling. `parseResolvedBlock`
validates a record's internal coherence and cannot see this, because the collision is between the
record and the block, not within the record.

**And one finding that is not a defect at all but is worth more than most of them.** The test the
previous round wrote to cover the sibling case cannot fail: it marks the sibling as running, but the
port it holds is never queried, because the probe only ever covers the binding set. The test passes
identically with the sibling not running — so it cannot distinguish the correct behaviour from the
regressed design it exists to guard against, and it would stay green through that regression. Counted
as coverage, worth none.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F71 | adversarial + defect hunt | Critical | `buildSrc/.../ResolvedBlock.kt`, `placedByRecord` | a placed port is never checked against the deterministic port of a service the record does not name, so composition can yield two services on one port and `freePort` kills the sibling that legitimately holds it |
| F72 | defect hunt | Important | `gradle/dev-lifecycle.gradle.kts`, `resolveWorkspaceBlock` | the record is committed before the stop loop, so a failed kill leaves it naming a port nothing is on. Loud for the run that fails, silent for every run after — and the stuck service stays "placed" by its own wrong entry, so a retried full run does not self-heal |
| F73 | adversarial + defect hunt | Important | `gradle/dev-lifecycle.gradle.kts`, `isServiceRunning` | the KDoc claims one predicate read by both sides; `startService` calls `readPidIfRunning` directly and never calls it. Identical text today, so nothing is wrong yet — but the guarantee the previous round's seam argument rests on is not in the code |
| F74 | adversarial | Important | `buildSrc/.../ResolvedBlockTest.kt` | the sibling test cannot fail: the sibling's port is never probed, so the assertion holds whether or not the sibling is running, and a regression to whole-block scoping would leave it green |
| F75 | adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `workspaceSurvivors` | run 2 invokes the survivors report from the main checkout, where `holdsPortBlockRecordOf` is never true — so a record left behind by a skipped worktree removal is reported as zero survivors and the cleanup guard goes green over a real leak |
| F76 | adversarial + defect hunt | Minor | `gradle/dev-lifecycle.gradle.kts`, `readPidIfRunning` KDoc | still describes the "already running" shortcut and its worst case; the shortcut was replaced by a throw in the round that did not update it |
| F77 | adversarial | Minor | `gradle/dev-lifecycle.gradle.kts`, `killPidTree` | a process forked in the trailing forced-kill window is counted but never signalled, where the comment claims both |
| F78 | adversarial | Minor | `gradle/dev-lifecycle.gradle.kts`, `stopService` | "loud on every path" is untrue when the port record is unreadable *and* the kill mis-reports success — `portToSweep` returns null by design, so the compound case is silent |

**Three of the eight are a KDoc claiming what the code does not do** — F73, F76, F77 — and that is
worth naming as this change's signature defect rather than filing three times and moving on. Every
one was written by a fix round explaining why its repair was sound. An explanation is not an
enforcement, and a reviewer reading the explanation instead of the code is how F73 survived a full
pass: the previous round's entire seam argument rests on a sharing that was described and never
implemented.

## Pass 9 — the pass that came back clean

The same two slots, on the round that closed pass 8. **Neither found a Critical**, and both earned
that by checking rather than accepting: the defect hunt ran the new porcelain parser against this
machine's real `git worktree list --porcelain` output, and the adversarial slot reproduced the
previous round's mutation analysis analytically — confirming both that the rewritten test now fails
under the scope-widening mutant and that the old one stayed green under it.

Six of the eight claims the round made were tested and held, including the two that had been claimed
falsely in earlier rounds: `livePid` really does unify the two readers now, where round 8's
equivalent was described and never implemented, and the git-failure path really does map onto
`check-cleanup-complete.sh`'s non-zero-means-nothing-said branch, which the reviewer verified by
reading that script rather than the comment about it.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F79 | adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `restartDevStack` | the per-service record write is reached only from inside the start loop, and the stop loop before it has already stopped everything — so a full run failing on its **first** service leaves the previous record intact, describing four services this run just stopped. Worse than stale: something taking a vacated port is read as *ours* by the next full run's `ours` set, masking a real clash. It contradicts the file's own header invariant |
| F80 | defect hunt | Minor | `buildSrc/.../ResolvedBlock.kt`, the collision check | the refusal is unscoped by `binding`, unlike every other check beside it. That is correct — `workspaceEnvironment` exports every service's port into every started process, so a collision anywhere leaks — but the reasoning is in neither the KDoc nor the message, and no test covers a binding disjoint from the colliding pair. Scoping it "for consistency" would silently reopen what it closed |
| F81 | defect hunt | Minor | `gradle/dev-lifecycle.gradle.kts`, `startService` | computes `pidFile(service.name)`, then calls `livePid`, which computes it again |
| F82 | adversarial | Minor | `buildSrc/.../CheckoutScan.kt` | the "what it does not answer" section names one bound and omits another: a checkout ever run with `-PworkspaceChange` diverges from what its branch or directory derives. `portBlockRecordsOf` already works around that divergence for the reporting checkout, so the codebase knows about it two files over |

**Three of these four are about writing down what is true**, which is the same signature defect pass 8
named — and the shape has now inverted in a useful way. F80 and F82 are not code that lies; they are
correct code whose reasoning is undocumented and, in F80's case, untested. That is the difference
between a round that is failing and a round that is finishing.

## Pass 10 — the closing verification

One slot, scoped to fix round 10's 318 lines and told plainly that finding nothing was the expected
outcome and that a manufactured finding would cost more than it was worth. **No Critical, and no
functional defect of any severity.** It confirmed the main-checkout no-op by tracing
`recordsBlock = false` through to the `?: return` that precedes the delete, confirmed the
empty-versus-null discipline at every caller, confirmed the new collision test goes red when the check
is scoped, and confirmed the round touched only the four files its four findings named.

It found two documentation defects, and **the Important one is this change's signature defect caught
one last time, by the reviewer doing the two-file read the comment's author had not.** The KDoc
justified deleting the record rather than truncating it by asserting that an empty file arrives as
`Recorded.Unknown`. It does not: `recordedValue` classifies on the file's *trimmed* text, so an empty
file and an absent one are both `Recorded.None` and either would read correctly. The code was right
and its stated reason was false — which is precisely the shape that has produced a finding in four
separate passes of this panel.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F83 | adversarial | Important | `gradle/dev-lifecycle.gradle.kts`, `recordWorkspaceBlock` KDoc | the delete-versus-truncate reasoning is factually wrong; both forms are `Recorded.None`. No functional consequence, and a false reason left standing is how the next maintainer inherits a wrong model |
| F84 | adversarial | Minor | same KDoc | "names exactly the services that are up" overclaims across the stop loop, where the record already says nothing while the previous run's services are still being killed. The `ResolvedBlock.kt` copy of the same invariant was already phrased accurately |

Both were fixed by the parent rather than a fix round — two comment edits, no code touched. The
correction **names the removed claim rather than quietly replacing it**, so the next reader of that
paragraph learns the reason was checked and found wrong, which is the only durable defence against a
comment that sounds authoritative.


## Findings — the marker block

**One block, both passes, one numbering.** The guard reads only the lines below, requires them
consecutive, and requires exactly one total — so a block per pass is two totals and an interrupted
run. That is what a first attempt at this record produced, and `scripts/check-unfinished-work.sh`
reported it immediately. Pass 2's findings therefore continue pass 1's sequence at F16 rather than
carrying a scheme of their own, because the guard matches the `F<n>` shape specifically and anything
else is not a marker at all.

**This paragraph deliberately does not spell the marker prefix.** The guard reports any line that
names it and is not one, which is how a quoted example inside a record can stand in for a marker
that was never written — and that rule caught this very paragraph on the previous attempt. Three
things about this record were learned by writing it wrongly and being told, which is the guard
doing what its own design argues for.

findings-total: 84
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 withdrawn — not fixed and not fixable by a run: it is a verification that needs two live stacks, and the operator chose to drive the stacks themselves rather than have a run bind their ports unattended. It is recorded in the manual test guide's `## Known incomplete` and left as unticked checks under **Running two changes at once**, which is the gate built for exactly this. The operator delegated the remaining decisions of this run, and this is the one they had already answered directly.
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 fixed
finding-status: F29 fixed
finding-status: F30 fixed
finding-status: F31 fixed
finding-status: F32 withdrawn — kept rather than reverted, by the operator's standing delegation of this run's remaining decisions. The reviewer named it as unexercised and explicitly declined to ask for a revert; the capability answers a real shape (a URL embedding a derived name whose own row holds a whole connection string, so no row's entire value can be substituted); removing it would reopen the ambiguity finding that prompted it in fix round 3; and its cost is a few sentences and their tests, not the 786 lines and recurring Critical that made the frontend-port generality worth deleting. Revisit if a second project adopts the contract without needing it.
finding-status: F33 fixed
finding-status: F34 fixed
finding-status: F35 fixed
finding-status: F36 fixed
finding-status: F37 fixed
finding-status: F38 fixed
finding-status: F39 fixed
finding-status: F40 fixed
finding-status: F41 withdrawn — the operator's decision at the pass 5 handback. The reviewer that raised it said explicitly it was not asking for the change, and the fix round read the argument and agreed; two independent judgements that the current shape is the right one is what the withdrawal records.
finding-status: F42 withdrawn — the operator's decision at the pass 5 handback. The reviewer that raised it said explicitly it was not asking for the change, and the fix round read the argument and agreed; two independent judgements that the current shape is the right one is what the withdrawal records.
finding-status: F43 fixed
finding-status: F44 fixed
finding-status: F45 fixed
finding-status: F46 fixed
finding-status: F47 fixed
finding-status: F48 fixed
finding-status: F49 fixed
finding-status: F50 fixed
finding-status: F51 fixed
finding-status: F52 fixed
finding-status: F53 fixed
finding-status: F54 fixed
finding-status: F55 fixed
finding-status: F56 fixed
finding-status: F57 fixed
finding-status: F58 fixed
finding-status: F59 fixed
finding-status: F60 fixed
finding-status: F61 fixed
finding-status: F62 fixed
finding-status: F63 fixed
finding-status: F64 fixed
finding-status: F65 fixed
finding-status: F66 fixed
finding-status: F67 fixed
finding-status: F68 fixed
finding-status: F69 fixed
finding-status: F70 fixed
finding-status: F71 fixed
finding-status: F72 fixed
finding-status: F73 fixed
finding-status: F74 fixed
finding-status: F75 fixed
finding-status: F76 fixed
finding-status: F77 fixed
finding-status: F78 fixed
finding-status: F79 fixed
finding-status: F80 fixed
finding-status: F81 fixed
finding-status: F82 fixed
finding-status: F83 fixed
finding-status: F84 fixed
