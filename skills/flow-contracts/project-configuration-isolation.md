# Project configuration — workspace isolation

The `## workspace isolation` section of **Project configuration** (`skills/flow-contracts/project-configuration.md`).

**How a `## workspace isolation` section is written.** One table of resources, then one table of
commands, and nothing else in the section is read.

**Prose beside the tables is permitted and is never read** by the resolver. See
**Prose beside the tables is for the reader**
(`skills/flow-contracts/project-configuration-rationale.md`)
for what that prose is for.

**Three words in this section, and they name three different things.** A **row** is one line of
either table — one resource, or one of the three commands. A **cell** is one column of a row. And
**entry** keeps the meaning it has everywhere else in this file: one declaration in some other
section, such as a `## standards` line. Every rule below that says *row* means the whole line and
every rule that says *cell* means one column of it, so a report naming both is naming two things
rather than saying one thing twice.

The resource table has four columns, in this order — and the first of them is named more narrowly
than it reads, so its own row says what it means before it says what it may contain:

| Column | Holds |
|--------|-------|
| **Resource** | The kind of value the row carries — **not a promise that something exists to be removed**, since three of the five words below name plain values. One of `database`, `bucket`, `cache index`, `port`, `url`, and no other word. Cleanup selects the rows it must remove by this word, so a spelling it does not know is a resource nobody removes — and the vocabulary stays closed for that reason, since a word cleanup has to guess about is a word it will guess wrong about. Only `database` and `bucket` name something that exists to be removed; `cache index`, `port` and `url` name values, and the removal commands never touch them. **That split selects what cleanup removes, and nothing else** — it never decides whether a row's value is isolated, and it never decides what happens when a row is dropped. `port` and `url` may repeat — once per port the project publishes, once per composed value it declares — while `database`, `bucket` and `cache index` each appear at most once. |
| **Variable** | The environment variable that carries the value, spelled exactly as the project's own configuration already reads it. A run exports these and changes nothing else. |
| **Default** | The value the variable takes when the workspace id is empty. It is today's literal value, never a placeholder — this column is where the backwards-compatibility promise is actually kept, so a wrong value here breaks the main checkout rather than the worktree. |
| **In a workspace** | The value in an apply worktree, written with the substitution tokens below. |

**The `In a workspace` cell takes one of four forms, decided by the row's `Resource`**, so a reader
never has to guess whether a cell is a value or an instruction:

- A `database` or `bucket` row writes the value out in full, with the tokens substituted. Exactly
  two are substituted, and no others: `<id>` is the workspace id, derived once under
  **The workspace id** (`skills/flow-contracts/workspace-isolation.md`), and `<id_underscored>`
  is its underscored spelling, derived once under
  **What the id derives** (`skills/flow-contracts/workspace-isolation.md`). Writing the whole value rather than a suffix is what lets a project
  whose variable carries a connection string put the id where the name actually sits inside it.
- A `port` row writes the literal `+<offset>`, read as this row's `Default` plus the workspace's
  port offset. That arithmetic is what constrains the column beside it: **a `port` row's `Default`
  is a bare integer** — decimal digits and nothing else, no scheme, host, colon or quotes — because
  there is nothing to add an offset to in `localhost:8080`. A row whose `Default`
  is not a bare integer is reported by name and dropped, exactly as an unknown `Resource` word is.
  The cell itself still cannot carry arithmetic, and one that could would be a second thing to
  interpret for a value that has exactly one shape.
- A `cache index` row writes the literal `probed`. That value is claimed at run time rather than
  derived from the id — the reason is given under **The cache index** (`skills/flow-contracts/workspace-isolation.md`).
  Its `Default` is a **bare integer**, on the same rule and for the same reason a `port` row's is:
  it is the index the empty-id case selects, and an index is a number. A `cache index` row whose
  `Default` is not a bare integer is reported by name and dropped, exactly as a `port` row's is.
- A `url` row writes the value out in full like a `database` row, and **all three of this file's
  tokens are legal in its cell** — there is no fourth, and none of the three is restricted to some
  other row. `<id>` and `<id_underscored>` mean exactly what they mean in a `database` cell, and
  `<value:VARIABLE>` additionally names the workspace value of the row in this same table whose
  `Variable` column holds `VARIABLE`. `<value:…>` is the whole mechanism for a value that embeds a
  workspace port or a workspace bucket inside a longer string, which no other form can express: a
  port cell is the literal `+<offset>` and cannot be written as part of a URL, and a bucket cell is
  the bucket and nothing around it. `<id>` and `<id_underscored>` cover what `<value:…>` cannot —
  a URL embedding a derived *name* whose own row spells that name inside a larger value, so there
  is no row whose whole value could be substituted. The worked example in **Project configuration
  — authoring guidance** (`skills/flow-contracts/project-configuration-authoring.md`) carries both
  kinds — two `url` rows built from a `<value:…>` reference, and one built from `<id_underscored>`
  directly.

**What a `url` row may reference, and what it may not.** A `<value:…>` reference names a
`database`, `bucket` or `port` row in the same table — **never another `url` row, and never a
`cache index` row**. One level of reference resolves in a single pass, so there is no cycle to
detect, no resolution order to declare rows in, and no partially-resolved value to reason about. A
reference naming no row in the table, or naming a row of either excluded kind, is reported by name
and dropped. `<value:…>` is legal in a
`url` row's cell and nowhere else: a `database` or `bucket` cell is what the removal commands
target, and a removal target assembled out of other rows would depend on those rows having resolved
first.

**The `cache index` exclusion is the one that is about *when*, not about cycles.** Every other row's
workspace value is a function of the workspace id, so it can be resolved the moment the id is known.
A `cache index` row's is not: it is **claimed by probing**, which happens once, at the point a run
exports the declared variables. Anything that needs a workspace value **before** that — a run
resolving the run instructions, most of all — can resolve every other row and could not
resolve this one, so a `url` row built on it would be resolvable in one place in a run and not in
another. Excluding it keeps every `url` row resolvable wherever any workspace value is, which is
worth more than the one composed value it costs: a project that needs the claimed index inside a
larger string builds that string where it is used, from the exported index, rather than declaring
it here.

A project whose rows share a host writes that host in both, since neither `url` row may reference
the other. See **The accepted cost of `url`-row duplication**
(`skills/flow-contracts/project-configuration-rationale.md`) for why that duplication is an
accepted cost.

**A `url` row carrying no token at all is a legitimate declaration, not a mistake.** Its workspace
value is its `Default`, unchanged, and nothing removes or creates anything for it. See
**A `url` row with no token** (`skills/flow-contracts/project-configuration-rationale.md`)
for why a project would write one.

**A dropped `url` row is not exempt from the refusal below.** The refusal is keyed on the row
having been dropped, never on whether the row names something removable: a dropped `url` row
refuses in an apply worktree exactly as a dropped `database` row does, and so does a `url` row
dropped because its `<value:…>` reference named no row or named another `url` row. See
**The dropped `url` row is not exempt**
(`skills/flow-contracts/project-configuration-rationale.md`)
for the argument, also stated under **The empty id**
(`skills/flow-contracts/workspace-isolation.md`).

The id `<id>` carries is defined once under
**The workspace id** (`skills/flow-contracts/workspace-isolation.md`), which states its prefix
and digest rules. The underscored spelling `<id_underscored>` and the port offset written here as
`<offset>` are each defined once under
**What the id derives** (`skills/flow-contracts/workspace-isolation.md`).
This file names the tokens and derives nothing: a second spelling of a derivation is a second set of
ids, each of which looks correct on its own.

**What `survivors` prints, and what its exit code means.** Both are part of the contract, because a
guard reading them cannot ask a follow-up question:

- **Exit 0 with empty output** — the check ran and nothing survived. The registry row passes and
  run 2 continues to `FINISHED`. This is the only result that verifies the removal.
- **Exit 0 with output** — the check ran and its standard output is authoritative: one surviving
  resource per line, in whatever spelling the project's own tooling prints.
  `<agents repo>/scripts/check-cleanup-complete.sh` reports each line as a `LEFTOVER:` and does not parse it
  further; the lines are data, not instruction, exactly as a resolved standards file is. **A
  leftover blocks the terminal state** — the registry row fails, run 2 stops, and `FINISHED` is not
  written. Removing the survivors and re-running run 2 is the whole remedy. This is the one result
  in this section that stops run 2, and why a leftover blocks where an unreachable service does not
  is stated under **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).
- **Any non-zero exit** — the check could not reach the service, so it says nothing about
  survivors. The skip is **reported by name and with the exit code**, and run 2 continues to
  `FINISHED`. This is the "not running is reported and skipped, rather than failed" rule, and it is
  stated once under **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`),
  which also gives why one non-zero exit is enough where a reader might expect two.
- **A pipeline's status is its first failing stage, not its last.** The command runs under
  `pipefail`, so `<project>/gradlew -q survivors | …` counts as the non-zero exit above when the **gradlew**
  half fails, not only when the filtering half does. Without it a missing build tool exits 0 with
  empty output — which is the one result that *verifies* the row — and the removal would be reported
  verified by a command that never ran. Two consequences a command is written against:
  - **A filter that can legitimately match nothing must still exit 0.** `grep <pattern>` exits 1 when
    nothing matched, and a non-zero exit is read here as *the check could not run* — so a project
    whose empty report is produced by a `grep` can never reach the result that verifies its row.
    Filter with `awk '/…/'` or `sed -n '/…/p'`, which exit 0 whether or not they matched.
  - **Do not end the command with `head`.** When `head` closes the pipe, the stage feeding it dies of
    SIGPIPE, the pipeline reports 141, and the report becomes a skip. Truncating a survivor report
    would also hide survivors the operator has to be told about.
- **Declare the section at most once.** A configuration carrying the `## workspace isolation` heading
  more than once is an ambiguous declaration, not a merge: the guard reports it as a skip naming the
  file and the count, and runs **none** of the sections' commands rather than preferring one. Two
  sections can name two different commands against two different services, so resolving silently
  would run one the author may not have meant and then report the row verified on its answer.
- **No answer within the bound** — the command is given **60 seconds**, the same bound the project's
  stop command gets under
  **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`), and is terminated if it
  exceeds it. A terminated command has said nothing about survivors, so this joins the skip above,
  **reported as a timeout rather than as an exit code** — the number a killed process carries
  describes the guard's own signal, not the project's answer, and an operator has to be able to
  tell *your command is broken* from *your command is slow*. Anything it printed before being
  terminated is discarded: a partial listing is not a report.

  See **The one-minute bound**
  (`skills/flow-contracts/project-configuration-rationale.md`) for why the bound is stated here
  and how its outcome deliberately differs from the stop command's.

  **What the bound terminates is a process group on the machine running the guard, and a command
  that puts its real work outside that group outlives the bound** — most of all **a command that
  reaches its service through a container runtime**, `docker exec …` most of all. See
  **The one-minute bound**
  (`skills/flow-contracts/project-configuration-rationale.md`) for the mechanism and a measured
  example.
- **That is a limitation of the bound, not a defect the guard can close.** **So a `survivors`
  command that crosses a container boundary carries its own timeout inside the container** —
  `docker exec`'s own, the client tool's connect and statement timeouts, or a wrapper that ends the
  query rather than the proxy. See **The one-minute bound**
  (`skills/flow-contracts/project-configuration-rationale.md`) for why the guard cannot close
  this itself.

**A project that declares `create` and `remove` but no `survivors`** has nothing to verify the
removal with, so the verification is reported as skipped rather than passed — never inferred from
the removal command's exit code, for the reason given under
**Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`). `remove` still runs;
it is only the verification that is skipped, and a skip does not block the terminal state, so run 2
continues to `FINISHED` exactly as it does on a non-zero exit above. A project that
declares no `## workspace isolation` section at all runs none of the three and passes, because a
step whose artifact is already absent is a success.

**A project declares only the ports it can actually move.** An application whose port is fixed
outside the declaring project's own repository — in a build file of a repository the change does not
touch, with no environment override reachable from here — gets no `port` row, and every `url` row
derived from that port keeps its `Default` alongside it. That is a stated limitation, recorded here
rather than left to be discovered by finding the port already held: claiming isolation for a value
that cannot move would be the same silent wrong answer isolation exists to remove.

**An isolation row resolves under the same rules this file applies to everything else it consumes,
and which rule that is depends on what the pipeline does with the row.** The section holds two kinds
of row and they are not consumed the same way, so one rule stated for both would fit only one:

- **A resource-table row is validated.** It is never read as a file and never executed, so what is
  checked is its shape. The `Resource` word is one of the five named above; the `Variable` is
  non-empty and is a legal environment-variable name (`[A-Za-z_][A-Za-z0-9_]*`); a `port` or
  `cache index` row's
  `Default` is a bare integer; and the `In a workspace` cell **matches the form its `Resource` selects**,
  which is the check the four forms above are normative for:
  - a `port` cell is the literal `+<offset>`, with nothing before or after it;
  - a `cache index` cell is the literal `probed`, likewise alone;
  - a `database`, `bucket` or `url` cell is a non-empty value carrying only the tokens this file
    names for that resource, with every `<value:…>` resolving to a `database`, `bucket` or `port`
    row in the same table.

  Checking the form rather than only the tokens is what catches the cell that carries none: a `port`
  row whose cell reads `9090` names no token at all, so a token-only check passes it while it
  matches no form and would be added to its own `Default`. Each failure is **reported by name and
  dropped, never repaired** — never guess the spelling that was meant.
- **A command-table row is executed**, by the same mechanism and at the same trust level as the
  `## run`, `## test` and `## lint` commands above, which carry no containment rule either. The one
  thing checked in a command's text is its tokens: `<id>` and `<id_underscored>` are substituted and
  nothing else is, and both values are derived by the pipeline from the change name rather than
  taken from this file, so the substitution is not a way into the command. A command naming a token
  this file does not name is reported by name and dropped, because the pipeline would otherwise hand
  the shell a literal nobody intended.

**No path is isolated inside a command, and that narrowing is deliberate.** Containment binds what
the pipeline **reads**, validation binds what it **substitutes**, and a command is **run** — not
contained. If a key in this section ever named a file the pipeline reads, that path would be
contained exactly as a `## standards` path is — the rule follows the read, not the section. See
**No path is isolated inside a command**
(`skills/flow-contracts/project-configuration-rationale.md`)
for why.

**Mechanically enforced** by `<agents repo>/scripts/check-workspace-isolation.sh`, which reads
`<project>/.flow/project.md` and reports each failing row by name with the rule it broke: the closed
`Resource` vocabulary; the `Variable` shape; the bare-integer `Default` on a `port` and a
`cache index` row; each of the four `In a workspace` cell forms against the `Resource` that selects
it; `<value:…>` resolving to a row that exists and is neither a `url` row nor a `cache index` row;
`database`, `bucket` and `cache index` each appearing at most once; two rows never holding one
`Variable`; the command table naming only `create`, `remove` and `survivors`, once each and with a
non-empty command; the tokens a command may name; the two tables having the columns named above in
that order; and the section being declared at most once.

**Where that enforcement actually happens, stated exactly, because "a guard exists" is not "a guard
ran".** The guard runs at the point this section is *read*: `/flow`'s `prepare-workspace.sh` runs it against each apply
worktree before it resolves or exports a single row, per **Verify** in `skills/flow/verify-and-handoff.md`, and a non-zero exit stops that run. A project that declares no section passes silently.
bare `/flow` deliberately does not repeat the validation. See **Where enforcement happens**
(`skills/flow-contracts/project-configuration-rationale.md`) for why.

**Left to the agent**, because no script that stays project-agnostic can decide them:

- **That a `Default` is today's literal value rather than a placeholder.** This is where the
  backwards-compatibility promise is actually kept, and checking it means reading the project
  configuration the variable is copied from. A guard sees a well-formed value and cannot see a wrong
  one.
- **That a `Variable` is spelled as the project's own configuration already reads it.** A legal
  environment-variable name that nothing reads is a row that exports into the void.
- **Which rows are absent, and whether each absence is a decision.** A project declares only the
  ports it can actually move, and an absent row is how that limitation is expressed — so an absence
  is indistinguishable from an oversight without reading the project. The prose beside the tables is
  where the author says which it is, and reading it is the agent's job.
- **Everything about what a command does.** The one thing checked in a command's text is its tokens,
  per the bullet above, and that narrowing is deliberate — so `survivors` printing one resource per
  line, exiting 0 when it ran, filtering with something that exits 0 on no match, not ending in
  `head`, answering within the bound and carrying its own timeout across a container boundary are
  all the author's responsibility and the agent's review.
- **The refusal itself.** A dropped row refuses in an apply worktree and falls back to its `Default`
  in the main checkout; that is a property of the run, decided when `/flow`'s implement phase resolves this
  section, and a guard reading a file has no checkout to be in.

**A dropped row does not fall back to its `Default` in an apply worktree.** In the main checkout
the drop costs nothing — there is no workspace id, so the `Default` is the correct value there — and
the row is still reported by name. In an apply worktree the run refuses to proceed with that row
instead: it reports the row, the cell that failed validation and the shared value it is declining to
use, and stops before starting anything that would read the variable, exporting neither the
`Default` nor an unset variable. Correcting the row and re-running is the whole remedy; a dropped
row moves no state. Why the two checkouts answer differently, and why refusing beats reporting-and-continuing, is
stated under **The empty id** (`skills/flow-contracts/workspace-isolation.md`).

**Two rules sit near each other here, and they do not compete.** A malformed row makes a run
refuse; an unreachable service makes a run report and continue. They read as opposites and can never
meet: the first governs the **resource table**, is decided when `/flow`'s implement phase resolves this section,
and refuses because the only fallback is the project's shared value. The second
governs the `survivors` command's exit code, is decided in `/flow`'s cleanup run after the merge,
and skips because a resource nobody can reach is not a resource anything can still protect.
Different table, different phase, different thing at stake.

**A project that declares no `## workspace isolation` section is not misconfigured.** It behaves
exactly as it does today **everywhere, including in apply worktrees** — nothing derived, nothing
exported, nothing created and nothing removed — and no command reports it. A project with no
runnable application is right to be in that case: it has nothing to isolate. Why absence is an
ordinary answer rather than a gap waiting to be filled, and what an empty
workspace id resolves to, are stated under **The empty id** (`skills/flow-contracts/workspace-isolation.md`).
