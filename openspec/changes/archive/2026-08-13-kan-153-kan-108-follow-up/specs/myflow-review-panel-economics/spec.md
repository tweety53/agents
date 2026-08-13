## MODIFIED Requirements

### Requirement: The auto-escalate trigger set is stated and discriminates

`/myflow-do` SHALL escalate a panel re-run from **Targeted** to **Full** automatically, without
asking and stating the reason in the panel record, when any of the following holds:

- the fix touched a file outside the set named in the findings;
- the fix altered a delta spec, a migration, or a guard's behaviour;
- a targeted re-run surfaced a **new** Critical finding;
- three or more fix rounds have already run;
- the fix diff exceeds approximately 150 changed lines **and** the diff adds a new file.

**Every trigger in that set SHALL discriminate between fixes in the repository it runs in.** A
condition that fires on every fix round of every change in a given repository selects nothing there:
it converts the ladder's default from Targeted to Full while still reading as a trigger. Where a
condition is found not to discriminate, the correct repair is to narrow the condition, not to remove
the escalation.

The clause naming `a guard's behaviour` SHALL replace the earlier clause naming `a public contract`,
which did not discriminate in a repository whose product is contracts. The narrowed clause still fires
on a fix that changes a script's exit codes or its output, which is a real coverage reason and is not
implied by the delta-spec clause: a guard's behaviour can change with no spec edit at all.

**The size clause SHALL be an amplifier and SHALL NOT be an independent trigger.** A diff's length
carries no risk signal on its own — a mechanical rename is large and harmless, and a one-line change
to a guard's exit code is small and dangerous — so size alone selects almost as indiscriminately as
the vacuous clause that preceded this narrowing. Once the earlier clause was narrowed, size became
the dominant path to a full re-run, which is the same defect in a second place, and the same repair
applies: narrow the condition rather than remove the escalation.

**The signal size is paired with SHALL be `adds a new file`, and SHALL NOT restate the other risk
signals in this set.** Each of those — a delta spec, a migration, a guard's behaviour, a file outside
the set named in the findings — already escalates on its own clause above, so a conjunction naming
them reaches no case the ladder does not already reach and is dead text in a trigger list, which is
the same defect as a trigger that fires on everything. `adds a new file` is the one signal with no
clause of its own, and the gap it closes is real: a large body of brand-new, wholly unreviewed code
in a file the findings themselves named, which no other clause sees. A trigger set SHALL be checked
for this on every edit — a condition that can never fire independently is as much a defect as one
that always fires.

Escalation SHALL remain a coverage decision and never a waiver: a targeted re-run SHALL still dispatch
no fewer than two slots, and handoff SHALL still require zero open findings at any severity from every
slot that has run.

#### Scenario: A contract edit no longer forces a full re-run

- **WHEN** a fix round in a repository made of contract documents alters a contract file but no delta
  spec, no migration, and no guard's behaviour
- **THEN** the re-run is Targeted, and the panel record states that no escalation trigger fired

#### Scenario: A fix that changes a guard's behaviour escalates

- **WHEN** a fix round changes a guard script's exit codes or its reported output
- **THEN** the re-run escalates to Full, and the panel record names that trigger

#### Scenario: A trigger that fires on every fix is a defect in the trigger

- **WHEN** a trigger is observed to fire on every fix round of every change in a repository
- **THEN** the condition is narrowed, and the escalation itself is not removed

#### Scenario: A large mechanical fix carrying no risk signal stays Targeted

- **WHEN** a fix round's diff exceeds approximately 150 changed lines and adds no new file, and no
  other trigger in the set has fired
- **THEN** the re-run is Targeted, and the panel record states that no escalation trigger fired

#### Scenario: A large fix adding a new file escalates

- **WHEN** a fix round's diff exceeds approximately 150 changed lines and adds a new file, and no
  other trigger has fired — the new file's path is one the findings named, so the outside-the-set
  clause does not fire
- **THEN** the re-run escalates to Full, and the panel record names both the size and the new file

#### Scenario: A condition that can never fire independently is a defect in the trigger

- **WHEN** a trigger's condition is found to be reachable only when some other trigger has already
  fired
- **THEN** the condition is narrowed to the case it alone reaches, or removed, rather than left in
  the set as text that selects nothing

#### Scenario: Targeting is never a coverage waiver

- **WHEN** a Targeted re-run runs because no trigger fired
- **THEN** at least two slots are dispatched, and handoff still requires zero open findings at any
  severity

### Requirement: A fix instruction is verified by a failing reproducer before dispatch

Every slot the panel dispatches SHALL be required, in its dispatch prompt, to supply for each finding
it raises either a **runnable command** that demonstrates the defect, or the literal exemption form
`none — <reason>`. Slots dispatched by `subagent_type` SHALL receive this requirement as prompt text,
the way the mutation-testing brief already reaches Bugbot; no agent definition is edited to carry it.

The panel record SHALL carry the reproducers in a marker block of their own — one
`reproducers-total: <n>` count line and one `finding-reproducer: F<n> <command | none — reason>` line
per finding. That block SHALL be separate from the `finding-status:` block, whose lines are required
to occupy one unbroken span; interleaving the two SHALL NOT be done.

**A runnable `finding-reproducer:` command is constrained, and is dispatched as a direct exec with
an argument vector — without shell interpretation.** `scripts/run-reproducer.sh <worktree>
<reproducer-command-line>` is that argv-exec runner: it resolves the path token to a file that
already exists inside the worktree, executes it with its arguments passed as a real argument
vector — never through a shell — and no reproducer line is ever handed to a shell for
interpretation, regardless of what it contains. It SHALL be a bare path to a file that already
exists inside the worktree, naming a script or test target, optionally followed by plain arguments.
**Every token — the path and every argument — SHALL be resolved against the worktree and required
to pass the same containment test** the `## standards` entry resolution already applies: normalize
`..`, `.` and symlinks, then require the normalized result to stay inside the worktree, so that a
relative path with no lexical `..` segment that escapes only through a symlink is still caught; an
argument that fails containment (for example `../../../etc/passwd`) SHALL make the reproducer
unverifiable exactly as a failing path does. A reproducer line carrying a shell metacharacter
anywhere on it (`` | ; & $ ` < > ( ) { } ~ * ? [ ] # \ ' " ``), a leading `-` on the path token
specifically (not on every token — `script.sh --strict` SHALL NOT be rejected on that account), or a
URL SHALL NOT be run. A reproducer that fails any of these checks SHALL be recorded as unverifiable
and put to the operator, never run — the line naming it is subagent-authored text, and the reviewed
diff itself can carry the injection that reaches it. **The guard script that validates the panel
record SHALL itself reject a runnable reproducer line that is not this bare-path shape**, at exit 1,
naming the violated shape — the constraint SHALL NOT live in prose alone. This mechanical rejection
SHALL extend to a path token or argument that is **absolute** or that carries a **`..` path
segment**: the guard SHALL check this lexically, on the text of the line, and SHALL NOT resolve the
token against a filesystem to do it — the record may be read in a context where the worktree that
wrote it is absent — so it rejects only the shapes that cannot be contained inside any worktree; the
dispatch-time containment test above, run by `run-reproducer.sh` against the worktree that actually
exists at dispatch time, resolves the shapes that can be. A rejected reproducer line SHALL NOT be
silently rewritten to satisfy the guard: it is a refusal, not a defect to fix, since the line may
itself be the injection the guard exists to catch.

**Before dispatching the fix subagent**, `/myflow-do` SHALL run each supplied, constrained
reproducer command through `scripts/run-reproducer.sh <worktree> <reproducer-command-line>`, with
the worktree as the reproducer's working directory, under a bounded wait of **20 seconds** plus a
**2-second** SIGTERM-to-SIGKILL grace — the grace matching `scripts/check-cleanup-complete.sh`'s own
`SURVIVORS_KILL_GRACE`, the 20-second bound being this rule's own number rather than that script's
60-second `SURVIVORS_TIMEOUT`, since a reproducer is one short-lived check under test rather than a
wait for a whole cleanup pass. `run-reproducer.sh`'s own exit code SHALL be read as its documented
contract: **0** demonstrates the defect and clears the finding for dispatch; **1** means the
reproducer exited 0 and demonstrates nothing, so the instruction built on it cannot be verified as a
fix and the finding SHALL NOT be dispatched to the fixer on that pass; **2** is a refusal — the
command line failed a containment or shape check and was never executed at all — recorded
**unverifiable** and put to the operator; **3** is a timeout or a detached survivor — the reproducer
was killed at the bound and SHALL be recorded **unverifiable** and put to the operator, never read as
either a pass or a fail, since a timeout is not an exit; **4** means the run cannot answer at all and
is handled like an unreadable record. A reproducer still running at the bound SHALL be killed. A
reproducer whose process double-forks or detaches SHALL, before being recorded unverifiable, have the
same SIGTERM-then-grace-then-SIGKILL sequence retried against it, found by `run-reproducer.sh`'s own
measured mechanism rather than by a `ps` keyword this platform does not have (`ps -o sid=` does not
exist here, and `ps -o sess=` was measured to report `0` for every process tried); a child still
alive after that attempt SHALL be
named to the operator, by `run-reproducer.sh`'s own exit-3 report, as a **surviving process**, with
its pid, so it can be found and killed manually, rather than merely recorded unverifiable. A
reproducer killed at the bound, or one whose child survives that retry, may already have written to
the worktree; the worktree SHALL be re-checked (`git status`) before the run continues, and, for a
surviving process, SHALL be re-checked again when the operator resumes, since nothing between the
pause and the resume observed what the child wrote.

**`run-reproducer.sh` SHALL launch the reproducer in a process group of its own**, so that a
descendant which re-parents away from it remains findable and killable by group membership rather
than by parentage. Parentage is not a sufficient mechanism: a grandchild whose intermediate parent
exits within milliseconds re-parents to the init process before the first poll, and no
parentage-based lookup can see it afterwards. The group SHALL be established without a shell ever
seeing the reproducer line — the argv vector SHALL pass through untouched — so the direct-exec
guarantee above is preserved. Survivor detection SHALL consult the process group alongside the
descendant walk, and cleanup SHALL signal the group before falling back to that walk. Where the
mechanism that establishes the group is unavailable at run time, the run SHALL refuse — recorded
**unverifiable** and put to the operator — and SHALL NOT fall back to an ungrouped exec, since an
ungrouped exec is the exact condition in which the survivor goes unseen. A reproducer that leaves
the group deliberately, by calling `setsid` itself, is a stated residual limit documented beside the
implementation, and no clean result SHALL claim to have covered it.

A finding whose reproducer passed SHALL be **bounced once** to the slot that raised it, carrying the
reproducer's passing output. If that slot's second reproducer also passes, the run SHALL stop and put
the finding to the operator through the existing handback prompt — take another round, withdraw it
with a reason, or stop the run — and SHALL NOT dispatch it silently.

**Once the fix subagent reports, /myflow-do SHALL re-run every dispatched finding's reproducer**
under the same constraints and SHALL require it now to exit **0**. A reproducer that still exits
non-zero means the fix did not fix it, and the finding SHALL NOT be closed on that pass. **The
passing re-run alone SHALL NOT close a finding: the fix's diff SHALL also touch at least one path
the finding named.** A fix whose diff, for a given finding, touches only the reproducer's own target
and no path the finding named SHALL NOT be treated as a fix; the finding stays open and goes to the
operator through the handback prompt, carrying that fact as the reason.

**The disqualifying condition is "no path the finding named", and the two clauses SHALL be read
together rather than the first alone.** Where the reproducer's target *is* a path the finding named
— the ordinary shape for a finding about a guard script, whose reproducer is that script's own test
harness or the script itself — the fix is material, and a reading that disqualifies it inverts the
condition into one that fails the commonest correct case. Every restatement of this condition in a
skill or contract SHALL carry both clauses; a restatement that carries only the first is a defect in
the restatement, not a narrowing of the requirement.

A finding recorded `none — <reason>` SHALL be dispatched without a run: the rule binds findings
claiming a mechanical defect, and a principles, prose or naming finding has no runnable check to
demand. Where a slot **dispatched by `subagent_type`** supplies nothing at all, the parent SHALL
record `none — not supplied by <slot>` and SHALL dispatch the finding unverified, so the omission is
visible in the record rather than indistinguishable from a verified instruction. This exemption
SHALL NOT apply to a **general-purpose** slot: a general-purpose slot that supplies nothing SHALL be
recorded as its own open finding rather than as `none — not supplied by <slot>`, since the pipeline
fully controls that slot's prompt and it has no third-party excuse for the omission.

The presence of the field SHALL be enforced by a script taking the worktree, exiting **0** when every
`F<n>` named in the `finding-status:` block has exactly one well-formed `finding-reproducer:` line and
`reproducers-total` equals the number of those lines, **1** when violations are found, each reported
with its line, and **2** when it cannot answer at all. Exit **1** SHALL cover two distinct classes,
handled differently by the caller: a **missing or malformed field** (no `finding-reproducer:` line
for some `F<n>`, a malformed `reproducers-total:` count) is a record-completeness defect and the
missing field is added before any fix is dispatched; a **rejected reproducer shape** (a command
carrying a shell metacharacter, an absolute path, a `..` segment, a leading `-` on its path token, a
URL, or a NUL byte) is a refusal, recorded **unverifiable** and put to the operator exactly like a
failed dispatch-time containment check, and SHALL NOT be silently rewritten to satisfy the guard.
It SHALL read the marker block through anchored patterns and SHALL NOT parse the findings table. It
SHALL NOT be a lint step — it needs a worktree for a change in flight — and SHALL be covered by its
own test harness.

#### Scenario: A failing reproducer clears the finding for dispatch

- **WHEN** a finding's reproducer command exits non-zero in the worktree
- **THEN** the finding is dispatched to the fix subagent, and the record carries its reproducer line

#### Scenario: A passing reproducer bounces rather than dispatching

- **WHEN** a finding's reproducer command exits 0
- **THEN** the finding returns to the slot that raised it, carrying the passing output, and reaches no
  fix subagent on that pass

#### Scenario: A second passing reproducer reaches the operator

- **WHEN** a bounced finding's replacement reproducer also exits 0
- **THEN** the run stops at the operator handback with that finding's text and named options

#### Scenario: A prose finding is exempt with its reason stated

- **WHEN** a slot raises a finding about naming, prose, or a principle and records `none — <reason>`
- **THEN** the finding is dispatched without a reproducer run, and the record states the reason

#### Scenario: A slot supplying nothing is recorded as such

- **WHEN** a slot dispatched by `subagent_type` reports a finding with no reproducer and no exemption
- **THEN** the record reads `none — not supplied by <slot>` and the finding is dispatched unverified

#### Scenario: A general-purpose slot supplying nothing is its own finding

- **WHEN** a general-purpose slot reports a finding with no reproducer and no exemption
- **THEN** the omission is recorded as its own open finding, not as `none — not supplied by <slot>`

#### Scenario: An unconstrained reproducer line is never run

- **WHEN** a `finding-reproducer:` line is not a bare path to an existing worktree file, or carries a
  shell metacharacter, an absolute path token, a `..` path segment, a leading `-` on its path token,
  or a URL
- **THEN** the reproducer is never run, and the finding is recorded unverifiable and put to the
  operator

#### Scenario: The record guard rejects a non-bare-path reproducer line mechanically

- **WHEN** a `finding-reproducer:` line's command carries a shell metacharacter, an absolute path
  token, a `..` path segment, a leading `-` on its path token, or a URL
- **THEN** `scripts/check-panel-reproducers.sh` exits 1 against that record, naming the shape it
  violated, rather than accepting it at exit 0, without resolving the token against a filesystem to
  decide it

#### Scenario: A rejected reproducer shape is a refusal, not a defect to fix

- **WHEN** a `finding-reproducer:` line is rejected by the guard's mechanical shape check
- **THEN** the line is recorded unverifiable and put to the operator, and is never silently rewritten
  to satisfy the guard, since the line may itself be the injection the guard exists to catch

#### Scenario: An argument outside the worktree makes a reproducer unverifiable

- **WHEN** a reproducer command's path token is contained but an argument after it, once normalized,
  resolves outside the worktree
- **THEN** the reproducer is never run, and the finding is recorded unverifiable and put to the
  operator

#### Scenario: A reproducer killed at the bound is unverifiable, not a pass or a fail

- **WHEN** a reproducer command is still running when the 20-second bound plus its 2-second grace
  elapses
- **THEN** it is killed, and the finding is recorded unverifiable and put to the operator — its exit
  status is read as neither a pass nor a fail — and the worktree is re-checked (`git status`) before
  the run continues, since the reproducer may already have written to it before being killed

#### Scenario: A detached reproducer child that survives is named to the operator, not merely logged unverifiable

- **WHEN** a reproducer double-forks or detaches a child, and that process is not confirmed gone once
  the SIGTERM-to-SIGKILL grace elapses
- **THEN** the same SIGTERM-then-grace-then-SIGKILL sequence is retried against the reproducer's
  process group, and a child still alive after that attempt is named to the operator as a surviving
  process with its pid rather than merely recorded unverifiable, and the worktree is re-checked
  (`git status`) both now and again when the operator resumes

#### Scenario: A fast double fork is detected rather than missed

- **WHEN** a reproducer double-forks a grandchild whose intermediate parent exits before the first
  poll, so the grandchild has already re-parented and no parentage-based lookup can find it
- **THEN** the process-group lookup finds it anyway, and the finding is recorded unverifiable with the
  surviving process named, rather than reported as "defect not demonstrated"

#### Scenario: A reproducer that leaves the group itself is a stated limit, not a silent miss

- **WHEN** a reproducer calls `setsid` itself and thereby leaves the process group it was launched in
- **THEN** the limitation is documented beside the implementation, and no clean result claims to have
  covered it

#### Scenario: A fixed finding's reproducer must now pass

- **WHEN** the fix subagent reports and a dispatched finding's reproducer is re-run under the same
  constraints
- **THEN** an exit of 0 closes the finding for that pass, and a non-zero exit means the fix did not
  fix it

#### Scenario: A fix that only edits the reproducer does not close the finding

- **WHEN** a dispatched finding's reproducer flips from failing to passing, but the fix subagent's
  diff touches only the reproducer's own target and no path the finding named
- **THEN** the finding is not closed; it stays open and is put to the operator through the handback
  prompt, carrying that fact as the reason

#### Scenario: A fix to a path that is both the finding's target and the reproducer's is material

- **WHEN** a dispatched finding names a path, the finding's reproducer targets that same path — the
  ordinary shape for a finding about a guard script — and the fix's diff makes a non-comment,
  non-whitespace change to it
- **THEN** the fix is material and the finding closes on the reproducer's flip, since the materiality
  condition disqualifies only a diff that touches **no** path the finding named

#### Scenario: A missing reproducer line fails the guard

- **WHEN** the panel record names `F3` in the `finding-status:` block and carries no
  `finding-reproducer:` line for it
- **THEN** the guard exits 1 and names the finding

