## ADDED Requirements

### Requirement: Every path citation in an installed file names its root

A file `setup.sh` installs or copies SHALL NOT carry a path citation whose root is unstated. Exactly
three roots are recognised:

| Form | Resolves against |
|------|------------------|
| `skills/…`, `rules/…`, `commands/…`, `commands-claude/…`, `hooks/…` | the installed root, wherever the harness placed it |
| `<agents repo>/…` | this checkout, by the procedure under **Where the agents repository is** (`skills/myflow-contracts/project-configuration.md`) |
| `<project>/…` | the project the running command is working against |

Two of the three are **placeholder roots**, written `<…>/`. The corpus uses further placeholder roots
for locations that are neither this checkout nor the target project, and a citation under one of
those is likewise rooted rather than missing a root. The recognised set is **closed**:

| Placeholder root | Names |
|------------------|-------|
| `<agents repo>/` | this checkout |
| `<project>/` | the project the running command is working against |
| `<abs-worktree>/` | the apply worktree a command is operating in |
| `<changeRoot>/` | the change's own directory, as `openspec status` reports it |
| `<state-dir>/` | the machine-level myflow state directory |
| `<skill-dir>/` | wherever the installed skills root places the currently-relevant skill's own directory |

**The set is closed, and an unrecognised placeholder is a violation, not a root.** A guard that
accepted any `<…>`-shaped first segment would fail open: `<foo>/openspec/specs/x.md` would pass while
naming an unrooted path, and so would a typo of a real placeholder — `<changeroot>/`, `<change-root>/`.
Adding a placeholder root is a deliberate act that extends this table; it is not something an author
does by accident in a citation.

The installed roots keep their bare form because they already resolve correctly in every harness and
because `scripts/check-references.sh` derives its entire coverage from resolving exactly those paths
against the repository root.

A path citation under none of the three roots is a violation. It resolves against whichever
repository the reading agent is standing in, which for an installed file is a tree whose contents a
contributor controls.

#### Scenario: An installed file cites this repository's own README

- **WHEN** an installed file carries the citation `` `README.md` ``
- **THEN** the repository's own lint reports that citation and fails
- **AND** the citation is corrected to `` `<agents repo>/README.md` ``

#### Scenario: An installed file cites the target project's own configuration

- **WHEN** an installed file carries the citation `` `<project>/.myflow/project.md` ``
- **THEN** the lint reports nothing for that citation

#### Scenario: An installed file cites another installed file

- **WHEN** an installed file carries the citation `` `skills/myflow-contracts/pipeline.md` ``
- **THEN** the lint reports nothing for that citation, because the first segment is an installed root

### Requirement: The guard derives the installed set by running the installer

`scripts/check-installed-citations.sh` SHALL determine which files are installed by executing
`setup.sh` into a sandbox and reading back what appeared — each symlink resolved to its
repository-relative source, plus each copied file — rather than by re-implementing the installer's
globs or by carrying a written list of installed paths.

A re-implementation drifts from the installer the first time an install path changes, and a written
list fails open: a file that starts being installed is scanned by neither. Deriving from a real run
cannot drift, which is what KAN-102 asks for.

The set of installed **roots** SHALL be the first path segment of each derived source path.

The guard SHALL refuse to run `setup.sh` unless both `HOME` and the target project directory lie
inside the sandbox directory that run created, adopting `scripts/test-setup.sh`'s existing refusal
rather than restating it.

#### Scenario: A newly installed directory is covered without editing the guard

- **WHEN** `setup.sh` gains an install path for a directory it did not previously ship
- **AND** the guard is run with no edit of its own
- **THEN** files under that directory are scanned, and its name is an installed root

#### Scenario: A rule the installer does not link is out of scope

- **WHEN** a rule declares no `alwaysApply: true` and `setup.sh` therefore never links it
- **THEN** the guard does not scan that file, and no written exception names it

#### Scenario: The sandbox refuses a run that would touch the real home directory

- **WHEN** the guard is about to invoke `setup.sh` with a `HOME` outside the sandbox it created
- **THEN** it refuses, writes its reason to stderr, and writes nothing to stdout

### Requirement: The citation classifier distinguishes a citation from a path-shaped token

The guard SHALL classify each backticked token before judging it, and SHALL NOT treat every token
containing a `/` as a citation. On this repository's corpus that naive test yields 723 candidates,
roughly three quarters of them not citations at all.

The guard SHALL NOT classify as a citation:

- an absolute path, or one rooted at `~`
- a URL
- a token whose first segment is a shell variable reference
- a git ref shape such as `origin/main` or `refs/heads/…`
- a regular-expression or glob fragment
- a token on a non-comment line inside a fenced `bash`, `sh` or `zsh` block, which is a shell
  argument and cannot carry a prefix
- a token beginning `../`, which is relative to something the document does not name, so no correct
  root exists to write. `./` and a bare `..` are **not** excluded — only the parent-relative form
- a token carrying a leading or trailing `'` or `"` **inside its own backtick span**, which is a
  quoted fragment of program output rather than a citation
- a token whose first segment is `origin` **and whose remainder carries neither a file extension nor
  a trailing `/`**, which is a git ref. Those two bounds keep `` `origin/README.md` `` and
  `` `origin/rules/` `` reportable while excluding the corpus's real refs, `origin/$BASE` and
  `origin/openspec/<name>`. `` `origin/README` `` — no extension, no trailing slash — remains
  excluded, and that residue is accepted: a branch genuinely could be named `README`
- a token matching `openspec/` followed by exactly one `<…>` segment and nothing more, which is this
  project's branch-naming template. The "and nothing more" bound is load-bearing:
  `openspec/specs/x.md` and `openspec/changes/<name>/` SHALL remain reportable
- a token ending `:<digits>`, which names a line inside a file rather than a path to cite. **This
  exclusion is deliberately unbounded and is a known hole**: a genuinely unrooted
  `` `.myflow/project.md:42` `` is excluded with it. Narrowing it — stripping the suffix and
  classifying the remainder — was tried and rejected, because it forces a fabricated `file:line`
  illustration in a findings table to be reported, and a findings table's Location column is
  diff-relative by construction
- a token any of whose **non-final** `/`-delimited segments ends in `:`, which no real path segment
  does. The non-final bound is load-bearing: it excludes `measured:/predicted:` while leaving a real
  `` `.myflow/project.md:` `` reportable

**A backtick span whose content begins `<!--` is an HTML comment, and SHALL be skipped whole** —
emitting no tokens at all — rather than word-split with its fragments excluded afterwards. Being a
comment is a property of the span, not of any word inside it, and rescuing fragments after splitting
is what made a later bound on multi-word merging break this rule.

**A comment that does not begin its own span is a known, accepted limitation.** Where prose precedes
`<!--` inside one span and a slash-bearing word sits inside the comment, the span-level skip does not
fire, the merge bound refuses to merge across that word, and the orphaned fragment is judged on its
own merits — producing a spurious violation. This fails **closed**: it over-reports rather than
letting an unrooted citation through, which is the safe direction for a guard. It occurs nowhere in
the corpus. It is recorded here rather than fixed, because narrowing it would mean parsing comment
structure inside a span, and a rule nobody can state in a sentence is worth less than a documented
edge.

Because the decision needs block structure as well as token shape, the classifier SHALL be
implemented in Python behind a thin shell wrapper, following the precedent
`scripts/check-plan-provenance.py` set and `.myflow/project.md` records, rather than as a
hand-rolled shell pattern.

A token carrying no `/` SHALL be classified as a citation only when it resolves to a real file at
the agents-repository root. This is what catches `` `README.md` `` and `` `setup.sh` `` without
flagging every generic mention of `SKILL.md` or `tasks.md`.

#### Scenario: A path inside a fenced command block is not a citation

- **WHEN** an installed file carries `git -C <abs-worktree> reset -- openspec/` on a non-comment line
  inside a ```bash fence
- **THEN** the guard reports nothing for that line

#### Scenario: A generic filename is not a citation

- **WHEN** an installed file mentions `` `tasks.md` `` with no directory part
- **AND** no file named `tasks.md` exists at the agents-repository root
- **THEN** the guard reports nothing for that token

#### Scenario: An HTML comment span emits no tokens

- **WHEN** an installed file carries `` `<!-- measured: ./gradlew test @ c515c42 -->` ``
- **THEN** the guard emits no token from that span at all, rather than splitting it and excluding the
  fragments individually

#### Scenario: A multi-word span is not merged across a path

- **WHEN** an installed file carries `` `some-cmd < data/input.txt > output.log` ``
- **THEN** the guard classifies each word independently rather than merging `< data/input.txt >` into
  one token, and reports `data/input.txt` as naming no root

#### Scenario: A git ref naming a file is still a citation

- **WHEN** an installed file carries the citation `` `origin/README.md` ``
- **THEN** the guard reports it, because the remainder carries a file extension and so names a file
  rather than a ref

#### Scenario: The branch-shape exclusion does not swallow a real path

- **WHEN** an installed file carries `` `openspec/specs/x.md` `` or `` `openspec/changes/<name>/` ``
- **THEN** the guard reports both, because neither is `openspec/` plus one bracketed segment alone

#### Scenario: Every word of a multi-word backtick span is classified

- **WHEN** an installed file carries `` `see .myflow/project.md` ``
- **THEN** the guard reports `.myflow/project.md`, rather than discarding every word after the first

#### Scenario: An unrecognised placeholder root is a violation

- **WHEN** an installed file carries the citation `` `<foo>/openspec/specs/x.md` ``
- **AND** `<foo>/` is not in the closed set of placeholder roots above
- **THEN** the guard reports it, because a bracket shape is not by itself a root

#### Scenario: A recognised placeholder root is rooted

- **WHEN** an installed file carries the citation `` `<abs-worktree>/.superpowers/sdd/final-review.diff` ``
- **THEN** the guard reports nothing for it, and counts it as a checked citation rather than
  discarding it unseen

#### Scenario: A bare filename that names a repository-root file is a citation

- **WHEN** an installed file mentions `` `setup.sh` `` with no directory part
- **AND** `setup.sh` exists at the agents-repository root
- **THEN** the guard reports it, and the citation is corrected to `` `<agents repo>/setup.sh` ``

### Requirement: The guard reports per member and refuses rather than reporting a clean run

`scripts/check-installed-citations.sh` SHALL follow the reporting discipline
`agents-repo-verification` already requires of a corpus-scanning guard: one violation line per
finding as `path:line: message`, a verdict line naming the root and the count, and a per-member
coverage fragment from `scripts/lib/coverage.sh`.

Exit 0 SHALL mean clean, 1 SHALL mean violations were found, and 2 SHALL mean the guard cannot
answer at all — an unreadable root, an installer run that failed, or a file it needed to read and
could not. A refusal SHALL write its reason to stderr and nothing to stdout, so no caller can read an
absent report as a clean one.

#### Scenario: The installer run fails

- **WHEN** the sandboxed `setup.sh` invocation exits non-zero
- **THEN** the guard exits 2, explains on stderr, and prints no report

#### Scenario: A clean corpus reports what it checked

- **WHEN** every citation in every installed file names a root
- **THEN** the guard exits 0 and its verdict line names how many files it scanned

### Requirement: A sibling-file citation resolves against the running command's own skill

A skill that names a file **beside itself** — a reviewer prompt, a principles list, anything a
dispatch hands to a subagent — SHALL resolve it against **the running command's own skill
directory**, exactly as **Guard resolution** (`skills/myflow-contracts/pipeline.md`) already resolves
a named guard.

It SHALL NOT be stated as "the directory you are reading this file from", and SHALL NOT hardcode one
harness's installed path. Both are unrooted citations in the sense this capability already governs:
one command skill routinely runs another's section text verbatim — `/myflow-fast` runs
`skills/myflow-do/SKILL.md`'s stages — so "the file you are reading" names the **cited** skill while
the dispatch belongs to the **running** one, and a hardcoded `~/.claude/skills/…` path names Claude
Code's install root in a file that installs into `~/.cursor/skills/` and `~/.codex/skills/` too.

**Every skill that dispatches with such a file SHALL carry it in its own directory**, as a relative
symlink to the one canonical copy — the rule `skills/*/scripts/` already follows for guards, for the
same reason: one file, reachable by a stable path, with no second copy to drift.

The failure this prevents is silent. A dispatching command whose own directory lacks the file either
reads another skill's copy by luck of a shared install root, or hands a subagent a path that does not
resolve — and the subagent proceeds without the reading it was told was required.

#### Scenario: A composite command dispatches with a sibling file

- **WHEN** `/myflow-fast` dispatches an implementer whose REQUIRED READING is
  `engineering-principles.md`
- **THEN** that file resolves inside `skills/myflow-fast/`, not inside the skill whose section text
  it happens to be running

#### Scenario: The canonical copy is not duplicated

- **WHEN** a second skill needs a sibling file another skill already carries
- **THEN** it carries a relative symlink to that one copy rather than a second copy of the content

#### Scenario: A sibling file no skill text names is not propagated

- **WHEN** a skill directory carries a file that no `SKILL.md` names for a dispatch
- **THEN** no other skill links it, because the rule binds files a skill dispatches with rather than
  every file that happens to sit beside one
