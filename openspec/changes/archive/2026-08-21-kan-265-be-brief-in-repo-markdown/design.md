# Design — be-brief in repo Markdown

The approved design is
`docs/superpowers/specs/2026-08-21-kan-265-be-brief-in-repo-markdown-design.md`. This file carries
the decisions and the open questions.

## Decisions

### Scope is the whole owned corpus

**ID:** scope-whole-corpus
**Status:** active
**Chosen:** every owned `.md` / `.mdc` — `README.md` included — with `node_modules/`, `.superpowers/`, `openspec/changes/archive/` and
`docs/superpowers/` excluded structurally rather than by a list that goes stale.
**Considered:** runtime-loaded text only, which is where a byte cut is paid for on every run but
leaves the specs — the largest unratcheted block — untouched; runtime text plus specs, which stops
short of the root files for no principled reason once specs are in.

### The standard is a clause in `rules/be-brief.mdc`

**ID:** standard-in-be-brief
**Status:** active
**Chosen:** a clause in the always-on rule, so the standard reaches every session and every
dispatched agent without a second authority to keep in sync.
**Considered:** extending `myflow-contract-economy`, which already owns the byte ratchet but is
scoped to `skills/myflow-contracts/` and would have to grow past its own purpose; a new spec, which
buys separation at the cost of a second place for "brief" to be defined differently.

### The clause names two subjects, and this repository is one of them by name

**ID:** clause-two-subjects
**Status:** active
**Chosen:** the clause governs (1) this repository's own Markdown, named explicitly, and (2) the
artifacts a `/myflow-*` run generates in any project — `proposal.md`, `design.md`, `tasks.md`, delta
specs, panel records, self-review reports. Other projects' pre-existing documentation is untouched.
**Considered:** a general rule about "a repository's documentation", which reads more naturally and
matches how every other rule in that file is written, but `be-brief.mdc` installs globally into
every project's `CLAUDE.md` — a general clause would silently govern projects nobody evaluated it
against. Naming this repository is a carve-out inside a global rule, accepted as the narrower harm.

### Completeness and non-repetition are separated in the rule's own text

**ID:** full-versus-brief
**Status:** active
**Chosen:** reword `be-brief.mdc`'s "code, commits, docs and specs stay full" so it reads as a
completeness requirement, and state the new clause as non-repetition. A file is required to be both.
**Considered:** leaving the sentence as it stands, which makes the rule self-contradicting the
moment it governs generated artifacts — `tasks.md` must carry every task **and** stop saying the
same thing three ways, and today's phrasing cannot express that.

### Cuts only, never paraphrase

**ID:** cuts-not-paraphrase
**Status:** active
**Chosen:** a passage may be deleted; it may not be reworded to say the same thing in fewer words.
These files are the pipeline's runtime source, and a requirement re-said slightly differently is a
requirement changed with nothing to detect it. A deletion is visible in a diff and provable against
the normative inventory.
**Considered:** moving rationale into appendices, extending the existing core/rationale split to
files that never got one — a larger win on runtime bytes, but a partition is a far more error-prone
edit than a cut; rewriting for density, which is the only option where a normative sentence can be
lost by paraphrase.

### The guard is the existing ratchet, widened and hand-maintained

**ID:** ratchet-widened
**Status:** active
**Chosen:** grow `check-contract-budget.sh`'s per-file table from 34 rows to one row per owned file
and widen its scan root. Hand-maintained, undeclared still a violation. A byte budget cannot see
restatement and is not trying to: the failure mode is regrowth, and a ratchet makes each instance a
deliberate, reviewable edit.
**Considered:** a cross-file duplicate-passage detector, which measures the real target but needs an
allowlist because these files legitimately quote one another's requirement titles and citation
paths — and the allowlist is what rots; both guards, doubling the maintenance for one signal;
generating the table, which makes "just regenerate it" the path of least resistance and stops the
ratchet ratcheting; directory totals, under which one file can balloon while another shrinks and the
guard says nothing.

### Generated artifacts carry the clause but no guard

**ID:** artifacts-clause-only
**Status:** active
**Chosen:** artifacts a run generates in another repository land at a new path every change, so
there is nothing stable to ratchet. The clause applies to them; no guard enforces it.
**Considered:** a guard in the consuming project, which would put this repository's standard into
every project's lint configuration and fail on any project that had not adopted it.

### The proof is a normative-sentence inventory

**ID:** normative-inventory
**Status:** active
**Chosen:** a script that extracts every sentence containing SHALL, SHALL NOT, MUST or MUST NOT,
normalises whitespace and prints them sorted. The trim must leave that set byte-identical, checked
before the first edit and after the last. None of the guards already in `## lint` would notice a
deleted requirement — `check-references.sh` catches a citation that stops resolving,
`check-markdown-integrity.py` catches structural damage, and a `SHALL` sentence can be deleted with
every one of them green.
**Considered:** relying on the existing lint and test suites alone, which is exactly the gap above;
per-file operator review of all 120 diffs, the only option that also catches a subtly reworded
requirement, rejected on the cost of reading 1.4 MB of before and after.

### No reduction target

**ID:** no-target
**Status:** active
**Chosen:** cut what is genuinely restatement and stop. The ratchet locks in whatever was honestly
cut.
**Considered:** a 15–20% target on runtime-loaded files, which makes the win reportable but pressures
an implementer to keep cutting once the genuine restatement runs out — and the next thing to cut is
a normative sentence; a target on the three largest files only, which has the same pressure in a
smaller blast radius.

### The trim runs largest-first, in groups, each with its own check

**ID:** groups-largest-first
**Status:** active
**Chosen:** file-group by file-group, largest-first, each group one task carrying its own inventory
check and its own full lint run. `rules/be-brief.mdc` is edited first, so every later group is cut
under a standard already written down.
**Considered:** one inventory check at the end, under which a regression cannot be attributed to a
file; alphabetical or per-directory order, which would leave the runtime-loaded text — the files
this change exists for — to the end of a long run.

## Open questions

None. Every question this stage opened was answered before it ended.
