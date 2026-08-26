#!/usr/bin/env python3
"""check-installed-citations.py — fail when a path citation in an installed
file names no root.

A file `setup.sh` installs or copies is read by an agent standing in
whatever tree the harness placed it — never necessarily this checkout. A
bare citation like `` `README.md` `` in such a file is read against THAT
tree, which is exactly how a bug ships: the citation resolves to nothing,
or worse, to some unrelated file the target project happens to have.
Exactly three FORMS of root are recognised (canonical definition:
specs/myflow-citation-roots/spec.md — do not restate the rule here):

  - an installed root's own bare form (`skills/…`, `rules/…`, `commands/…`,
    `commands-claude/…`, `hooks/…`, wherever the harness placed it)
  - a placeholder root — a CLOSED set of five, see PLACEHOLDER_ROOTS below
    for the set itself and why membership rather than bracket shape is
    what a placeholder root requires
  - (nothing else — anything not matching one of the above is a violation)

GUARD CONTRACT (as check-installed-citations.sh, this script's wrapper,
presents it — see that file and this module's main() for why the split):
  0  clean — every citation in every installed file names a root, and
     every scanned member's coverage is either non-zero or declared.
  1  violations found — one or more citations name no root, or a member
     reported zero coverage without a declared reason. Fix the named
     file:line by prefixing the citation; never suppress.
  2  environment — the guard cannot determine anything: the root override
     is set but empty or not a directory, the sandboxed installer run
     failed, or a file it needed to read could not be read or decoded. A
     refusal writes its reason to stderr and NOTHING to stdout, so no
     caller can read an absent report as a clean one.

THIS SCRIPT'S OWN exit codes are narrower than the contract above: 2 for
every refusal (unchanged, and still the whole story on stdout/stderr), 0
otherwise. Whether the run is clean, carries citation violations, or
carries an undeclared-zero-coverage violation is a decision
check-installed-citations.sh makes after this script exits — it is the
one that sources scripts/lib/coverage.sh and owns coverage_declare/
coverage_verdict, per task 2 step 5's own instruction to use that
library's existing mechanism rather than a second one reinvented here in
Python. So on a successful run this script prints a real citation
violation line as-is, then the resolved root and one coverage line per
scanned member behind two sentinel prefixes (CIC_ROOT_PREFIX,
CIC_COVERAGE_PREFIX) for the wrapper to parse, and always exits 0 — never
1 — because "1" is the wrapper's word to say, once it has folded in
coverage_verdict's own answer too.

HOW "INSTALLED" IS DERIVED — never re-implemented, never a written list.
This guard creates its own throwaway sandbox, runs the real `setup.sh`
`global` and `all <sandbox>/proj` into it (refusing first unless both the
HOME and the project directory it is about to hand the installer lie
inside that sandbox — see run_setup/ensure_within_sandbox, adopting
scripts/test-setup.sh's own refusal rather than restating it), and reads
back what appeared: every symlink resolved to its repository-relative
source, every regular file matched by content against a real top-level
repository file. The installed ROOTS are the first path segment of every
derived source that has one — a source with none (a copied top-level file
such as CLAUDE.md or AGENTS.md, whose repository-relative path IS its own
basename) never becomes a root, so a bare citation of CLAUDE.md itself
still needs a prefix; it is exactly this literal single-segment shape a
bare, no-slash citation like `` `README.md` `` would otherwise collide
with. A re-implementation of the installer's globs drifts the moment an
install path changes; deriving from a real run cannot.

THE CLASSIFIER — block structure before tokens, following the ordering
scripts/check-plan-provenance.py's own docstring records as the reason it
is not a regular expression, though far narrower in scope than that
guard's fence/container parser: this one only needs to know whether a line
sits inside a fenced bash/sh/zsh block, tracked the same way
scripts/check-guard-symlinks.sh's own RULE3_AWK/CITATION_AWK already do
(a fence's opening backtick run sets its length and language; closing
requires a run at least as long, with nothing trailing). A line inside
such a fence, if it is not a `#` comment, is a shell argument and is
excluded WHOLESALE — see scan_file_for_citations — matching that guard's
own choice to scan content inside a non-bash/sh/zsh fence not at all,
neither as shell nor as prose. Outside any fence, a citation is a
backtick-delimited span; inside a bash/sh/zsh comment line, it is a bare
whitespace-delimited word — code has no reason to backtick-quote its own
comments' paths.

Per candidate token, excluded before a root is ever judged (see
classify_token): an absolute path, a `~`-rooted path, a URL, a token whose
first segment is a shell variable reference, a git ref shape, and a
regular-expression or glob fragment. A token with no "/" is a citation
only when it names a real file at the agents-repository root — this is
what catches `` `README.md` `` and `` `setup.sh` `` without flagging every
generic mention of `SKILL.md` or `tasks.md`.
"""

import os
import re
import shutil
import subprocess
import sys
import tempfile

GUARD_NAME = "check-installed-citations"

# Protocol this script's stdout speaks to its own wrapper
# (check-installed-citations.sh), which owns the final report and exit
# code — see main()'s own closing comment for why. A real violation line
# (`path:line: message`) is printed as-is; everything else this script
# needs to hand back is prefixed with one of these two sentinels, neither
# of which collides with a real repository path.
CIC_ROOT_PREFIX = "CIC-ROOT\t"
CIC_COVERAGE_PREFIX = "CIC-COVERAGE\t"


class SandboxRefusal(Exception):
    """Raised when an installer invocation would touch something outside
    the sandbox this guard created for itself."""


# ---------------------------------------------------------------------------
# Root resolution
# ---------------------------------------------------------------------------


def resolve_repo_root():
    """REPO_ROOT is always resolved from this script's own physical
    location by default — one level above scripts/, exactly as
    check-plan-provenance.py's resolve_repo_root does, and safe for the
    same reason that guard's own comment gives: this script has exactly
    one home (scripts/), never symlinked into any skill's own directory —
    the farm distributes skills/, not scripts/ — so "one level above the
    RESOLVED physical location" never answers a skill directory by
    accident. CHECK_INSTALLED_CITATIONS_ROOT is an explicit, opt-in
    override honoured only when set, solely so the companion harness
    (test-check-installed-citations.sh) can point this guard at a
    sandboxed fixture tree under TMPDIR without touching this repository —
    never set it for a normal invocation.
    """
    if "CHECK_INSTALLED_CITATIONS_ROOT" in os.environ:
        root = os.environ["CHECK_INSTALLED_CITATIONS_ROOT"]
        if root == "":
            print(
                f"{GUARD_NAME}: CHECK_INSTALLED_CITATIONS_ROOT is set but empty",
                file=sys.stderr,
            )
            sys.exit(2)
        return root

    script_dir = os.path.dirname(os.path.realpath(__file__))
    return os.path.dirname(script_dir)


# ---------------------------------------------------------------------------
# Deriving the installed set by running the installer
# ---------------------------------------------------------------------------


def ensure_within_sandbox(path, sandbox):
    """Refuse (raise SandboxRefusal) unless the resolved, physical <path>
    lies inside <sandbox>. Deliberately a standalone function with no
    subprocess side effect of its own, so a caller — including
    test-check-installed-citations.sh's direct-import case — can prove the
    refusal fires BEFORE any external command runs, never merely that the
    end result looked safe.
    """
    sandbox_real = os.path.realpath(sandbox)
    path_real = os.path.realpath(path)
    if path_real != sandbox_real and not path_real.startswith(sandbox_real + os.sep):
        raise SandboxRefusal(f"{path} is not inside the sandbox {sandbox}")


def run_setup(repo_root, sandbox, home, project_dir, mode):
    """Run `<repo_root>/setup.sh <mode> [project_dir]` with HOME=<home>,
    refusing first — via ensure_within_sandbox — unless both <home> and
    <project_dir> (when given) lie inside <sandbox>. Returns the completed
    subprocess.run result. Every real caller in this file constructs both
    paths AS subdirectories of a sandbox it just created, so the refusal
    never fires in normal operation; it exists to make that invariant a
    checked fact rather than an assumption.
    """
    ensure_within_sandbox(home, sandbox)
    if project_dir is not None:
        ensure_within_sandbox(project_dir, sandbox)

    args = [os.path.join(repo_root, "setup.sh"), mode]
    if project_dir is not None:
        args.append(project_dir)
    env = dict(os.environ)
    env["HOME"] = home
    return subprocess.run(
        args, env=env, cwd=sandbox, capture_output=True, text=True
    )


def derive_installed_set(repo_root, sandbox):
    """Walk <sandbox> (already populated by two setup.sh runs) and return
    the set of repository-relative source paths it reveals: for every
    symlink, its resolved physical target's path relative to <repo_root>
    (whether that target is a file or a whole directory — install_skills
    symlinks an entire skill directory as one unit); for every regular
    file, the top-level repository file its bytes match, if any (a `cp`
    copy, e.g. CLAUDE.md/AGENTS.md). A regular file matching nothing is not
    part of the installed set — it is something the installer synthesised
    (a rendered managed block), not a citation-bearing file this repository
    tracks.
    """
    repo_root_real = os.path.realpath(repo_root)

    top_level_files = {}
    for name in os.listdir(repo_root):
        full = os.path.join(repo_root, name)
        if os.path.isfile(full) and not os.path.islink(full):
            top_level_files[name] = full

    sources = set()

    def record_if_symlink_into_repo(full):
        if not os.path.islink(full):
            return False
        target = os.path.realpath(full)
        if target == repo_root_real or target.startswith(repo_root_real + os.sep):
            sources.add(os.path.relpath(target, repo_root_real))
        return True

    for dirpath, dirnames, filenames in os.walk(sandbox, followlinks=False):
        kept_dirnames = []
        for d in dirnames:
            full = os.path.join(dirpath, d)
            if record_if_symlink_into_repo(full):
                continue  # a symlinked directory: recorded, never descended into
            kept_dirnames.append(d)
        dirnames[:] = kept_dirnames

        for f in filenames:
            full = os.path.join(dirpath, f)
            if record_if_symlink_into_repo(full):
                continue
            try:
                with open(full, "rb") as fh:
                    data = fh.read()
            except OSError:
                continue
            for name, tl_full in top_level_files.items():
                try:
                    with open(tl_full, "rb") as fh2:
                        tl_data = fh2.read()
                except OSError:
                    continue
                if data == tl_data:
                    sources.add(name)
                    break

    return sources


def compute_roots(sources):
    """The installed roots: the first path segment of every derived source
    that HAS a second segment. A source with none — a copied top-level
    file such as CLAUDE.md, whose repository-relative path is its own bare
    basename — never becomes a root: see this file's module docstring for
    why a bare citation must still fail even when the file it names really
    is installed.
    """
    return {s.split("/", 1)[0] for s in sources if "/" in s}


def compute_corpus(repo_root, sources):
    """Every `.md`/`.mdc` file this guard scans: for a source that is a
    directory (a whole installed skill), every such file anywhere beneath
    it in the REAL repository tree; for a source that is itself a `.md`
    file (a rule, a command, a copied CLAUDE.md/AGENTS.md), itself. A
    source that is a file of some other kind (a hook's `.py`) contributes
    its root but nothing to the corpus.
    """
    corpus = set()
    for s in sources:
        full = os.path.join(repo_root, s)
        if os.path.isdir(full):
            for dirpath, _dirnames, filenames in os.walk(full):
                for f in filenames:
                    if f.endswith(".md") or f.endswith(".mdc"):
                        rel = os.path.relpath(os.path.join(dirpath, f), repo_root)
                        corpus.add(rel)
        elif os.path.isfile(full):
            if s.endswith(".md") or s.endswith(".mdc"):
                corpus.add(s)
    return corpus


# ---------------------------------------------------------------------------
# The citation classifier
# ---------------------------------------------------------------------------

FENCE_OPEN_RE = re.compile(r"^(`{3,})(.*)$")
SHELL_FENCE_LANG_RE = re.compile(r"^(bash|sh|zsh)(\s|$)")
URL_RE = re.compile(r"^[A-Za-z][A-Za-z0-9+.-]*://")
GIT_REF_PATH_RE = re.compile(r"^refs/")
GIT_BRANCH_RE = re.compile(r"^[A-Za-z0-9_.-]+/(main|master|HEAD)$")
# ORIGIN_REF_EXTENSION_RE — the per-task-review bound on the `origin/…`
# exclusion below (finding 1): the ORIGINAL rule excluded EVERY token
# whose first segment was the literal `origin`, unconditionally — so
# `` `origin/README.md` `` sailed through uncounted, never even reaching
# judges_ok. The real corpus's only two `origin/…` sites, `origin/$BASE`
# and `origin/spectre/<name>`, are both genuine remote-tracking refs, so
# the rule is narrowed rather than deleted: a token whose first segment is
# `origin` is a ref only when what follows has no file extension — a real
# path citation in this corpus always ends in one (`.md`, `.sh`, `.py`,
# …), and neither live site does. `origin/README.md` now falls through
# unexcluded, is counted as checked, and is reported (first segment
# `origin` names no recognised root).
ORIGIN_REF_EXTENSION_RE = re.compile(r"\.[A-Za-z0-9]+$")
# GIT_BRANCH_SPECTRE_RE — task 10 first drew this as an exact-literal set,
# `{"openspec/<name>"}`, for the one branch name this corpus wrote that
# GIT_BRANCH_RE's suffix set (main/master/HEAD) does not cover. The second
# per-task review found a second live site of the identical shape —
# `openspec/<change-name>` — that an exact-literal set does not generalise
# to, so it is a SHAPE now: the branch prefix followed by exactly one `<…>`
# placeholder segment and NOTHING ELSE. The prefix became `spectre/` at the
# spectre cutover, moved WITH the contract prose that writes the branch
# name rather than after it: a shape exclusion left pointing at a prefix
# the corpus no longer writes stops matching silently, and every branch
# name it used to exempt is then reported as an unrooted path. Anchored at both ends so it cannot
# widen past that: `spectre/specs/x.md` (no brackets) and
# `spectre/changes/<name>/` (a bracket segment followed by MORE path) are
# both still fully reportable — this is the deliberately narrow bound that
# stops the shape from becoming the same fail-open hole task 9's
# placeholder generalisation was reverted for. `[^/<>]+` inside the
# brackets, not `.+`, so the bracket segment itself cannot smuggle in a
# second `/`.
GIT_BRANCH_SPECTRE_RE = re.compile(r"^spectre/<[^/<>]+>$")
# GIT_BRANCH_CHORE_RE — a sibling of GIT_BRANCH_SPECTRE_RE, added by
# kan-239's task 9 for the second pipeline-created branch shape that
# exemption does not cover: `chore/archive-<name>`, the branch
# `/myflow-finish` run 2 creates and names, backticked, in
# finish-contract.md, and `chore/self-review-<name>`, which run 2 no
# longer creates — kan-239 commits the self-review report onto the
# archive branch instead — but which survives as real branches in this
# repository from before that change. Both shapes are covered because
# both are branch names wherever they appear. Same shape,
# same bounds, for the same reason — a branch name is not a filesystem
# path, so it should never be judged against one: `chore/` followed by a
# fixed branch-kind word (`archive` or `self-review`) and exactly one
# `<…>` placeholder segment and NOTHING ELSE. Anchored at both ends so it
# cannot widen past that: `chore/archive-<name>/spec.md` (a bracket
# segment followed by MORE path) stays fully reportable — the same
# fail-open bound GIT_BRANCH_SPECTRE_RE's own comment describes.
# `[^/<>]+` inside the brackets, not `.+`, so the bracket segment itself
# cannot smuggle in a second `/`.
GIT_BRANCH_CHORE_RE = re.compile(r"^chore/(?:archive|self-review)-<[^/<>]+>$")
# FILE_LINE_RE — a token ending `:<digits>` names a line inside a file (a
# findings-table Location cell, taken verbatim from `git diff` output and
# diff-relative by construction), not a path to cite. Rooting such a token
# teaches a format real findings never use — see skills/myflow-do/SKILL.md's
# example findings row. Matched against the whole token, so a real path
# citation that merely CONTAINS a colon elsewhere is untouched.
#
# PER-TASK-REVIEW FINDING 3 — narrowing was tried and rejected, not
# overlooked. The steer was to strip the `:<digits>` suffix and classify
# the remainder, so a real unrooted path stays reportable. Tried exactly
# that: it forces skills/myflow-do/SKILL.md:485's own findings-table
# example row — `` `src/Foo.kt:42` `` — to be judged as an unrooted
# citation, because "src/Foo.kt" genuinely IS unrooted, and that row is
# deliberately fabricated (no such file exists) to teach the Location
# cell's FORMAT, not to cite a real one. Re-rooting a fabricated example
# would misrepresent it exactly as forcing a root onto task 9's negative
# examples or a quoted git error message would. So the full-token
# exclusion stays: this IS a documented hole (a real `path:line` citation
# in this shape is unreportable) rather than an undocumented one — see
# specs/myflow-citation-roots/spec.md, which the planner amends to say so.
FILE_LINE_RE = re.compile(r":[0-9]+$")
# "$" is deliberately excluded from this set: a leading "$" is the shell-
# variable-reference exclusion's own signal (see classify_token), and
# keeping it here too would make that exclusion untestable in isolation —
# every shell-variable token already contains a "$" and would be caught by
# this set regardless, which is exactly the untested-mutation shape KAN-197
# names (a case that still passes with the line it covers deleted proves
# nothing). A genuine glob/regex fragment is still caught by the other
# metacharacters below.
REGEX_GLOB_CHARS = set("[](){}*+?|^")

def merge_bracket_placeholder_words(words):
    """Merge a `<…>`-bracketed phrase spanning more than one word back into
    one word, wherever it starts.

    A plain whitespace split — used both to walk a backtick span's own
    words (extract_backtick_tokens) and to tokenize a shell-fence comment
    line (scan_file_for_citations) — breaks any placeholder whose own text
    contains a space (`<agents repo>`, `<the running command's own skill
    directory>`, `<project root>`) into several fragments before either
    ever gets a chance to classify it. Left unmerged, the FIRST fragment
    (`<agents`) is a citation candidate that never reaches classify_token/
    judges_ok at all (a per-task-review finding — the original, narrower
    version of this function existed for exactly `<agents repo>/…`), and —
    once every word in a span is classified rather than only the first,
    per that same review's finding 2 — every OTHER fragment is judged on
    its own, standalone, which turns a single multi-word placeholder
    citation into a garbled, unreadable violation (`` `repo>/…` ``,
    `` `directory>/scripts/<name>` ``) instead of the real one. Generalised
    from a hardcoded two-word case to ANY run of words starting with a `<`
    that has no `>` of its own, consumed until a later word supplies one —
    so the reported token is always the real placeholder text, whether or
    not that placeholder turns out to be a recognised root (judges_ok's
    own, separately closed, business).

    A `<` that never finds a closing `>` before the words run out is left
    exactly as split — nothing here manufactures a placeholder that is not
    actually there. So is a `<` that DOES find one but with a `/` in some
    word strictly between them — see the bound inline below, added after
    panel round 2 found the earlier, unbounded version merging a shell
    redirection (`some-cmd < scripts/input.txt > output.log`) into one
    garbled false citation. Nested brackets closing at the first `>`
    (`` `<a <b> c>` `` merges only through `<b>`) and an unclosed `<` are
    both reviewed and left as-is — neither is a defect this corpus's own
    content ever exercises.
    """
    merged = []
    i = 0
    n = len(words)
    while i < n:
        w = words[i]
        if w.startswith("<") and ">" not in w:
            j = i + 1
            parts = [w]
            closed = False
            while j < n:
                parts.append(words[j])
                if ">" in words[j]:
                    closed = True
                    j += 1
                    break
                j += 1
            # PANEL ROUND 2 — bound added: a word strictly BETWEEN the
            # opening `<` word and the closing `>` word (parts[1:-1]; empty
            # when they are adjacent, as in `<agents repo>/…`) must carry
            # no `/`. A placeholder phrase never has one there — every real
            # one in this corpus is plain English words until the closing
            # bracket. A shell redirection does: `some-cmd < scripts/
            # input.txt > output.log` splits into `<`, `scripts/input.txt`,
            # `>`, and without this bound the greedy merge above joins all
            # three into one garbled "citation" over ordinary prose
            # documenting a command inline — this guard's over-report
            # failure mode, not its silent-pass one, but a real defect in
            # the one function that produced the change's original
            # critical (silent-pass) defect. Rejecting here leaves the
            # words exactly as split, the same as an unclosed `<` already
            # falls through untouched below.
            if closed and not any("/" in p for p in parts[1:-1]):
                merged.append(" ".join(parts))
                i = j
                continue
        merged.append(w)
        i += 1
    return merged


def extract_backtick_tokens(line):
    """Every span between a pair of single backticks on <line>, in order —
    EVERY whitespace-delimited word in the span, not only the first.
    Greedy left-to-right pairing — this file's corpus writes citations in
    plain single-backtick spans, never nested or doubled, so the extra
    machinery check-references.sh's normalized_line carries for a
    soft-wrapped orphan marker has no call site here.

    Per-task-review finding 2: an earlier cut of this function kept only
    the span's LEADING word, on the theory that a multi-word span is an
    inline shell example — `` `agents/setup.sh global` `` — and everything
    past the first word is an argument, not a citation. That theory is
    right about `global` (see below), but the mechanism was wrong: it
    dropped every OTHER word in the span too, including a genuine second
    citation — `` `see .flow/project.md` `` — which then went not merely
    unreported but NEVER SEEN, the exact recognised-versus-never-seen
    collapse this guard's own test suite exists to catch (see
    agents-repo-prefix-bogus-path and placeholder-rooted-is-recognised).
    Classifying every word closes that hole without reopening the
    original one: `global` and `workspaceRemove` carry no `/` and match no
    real file at the agents-repository root, so classify_token's own
    bare-token rule excludes them independently, and a quoted fragment of
    shell output (`` `sed -n '/…/p'` ``'s `'/…/p'`) is excluded by the
    leading/trailing-quote rule task 9 added — neither exclusion is this
    function's job to redo.

    A SPAN whose own content begins with `<!--` is skipped WHOLESALE —
    no words extracted, nothing merged, nothing handed to classify_token
    at all. Panel round 3 first added a TOKEN-level exclusion for this
    (`classify_token` rejecting any word starting `<!--`), then round 4
    found it conflicts with the redirection bound in
    merge_bracket_placeholder_words: an HTML comment illustrating a path
    inline (`` `<!-- measured: ./gradlew test @ c515c42 -->` ``) contains
    a `/`-bearing word (`./gradlew`) strictly between `<!--` and `-->`,
    so that bound correctly refuses to merge it — and once unmerged, the
    fragment `./gradlew` is judged on its own with nothing left to
    protect it. The real fix is this one: being a comment is a property
    of the WHOLE SPAN, not of any word inside it, so the span is excluded
    HERE, before word-splitting or merging ever runs, and the two rules
    stop competing over the same words. classify_token's own former
    `<!--` check is now dead — a token-level rule can never see a comment
    span's own content once the span produces no tokens at all — and is
    deleted rather than kept "just in case": this repository's stated
    policy is a hole that is written down, or a rule that is deleted, not
    a rule nothing can reach.
    """
    tokens = []
    i = 0
    n = len(line)
    while i < n:
        if line[i] == "`":
            j = line.find("`", i + 1)
            if j == -1:
                break
            span = line[i + 1 : j]
            # Panel round 5 (the panel's code-quality reviewer): leading whitespace INSIDE
            # the backticks — `` ` <!-- ... -->` `` — must not defeat this
            # check. Stripped only for the test, never for the content
            # handed onward: split() already ignores leading/trailing
            # whitespace on its own, so nothing here changes what a
            # non-comment span produces. Deliberately NOT widened to
            # "`<!--` appears anywhere in the span" — this corpus
            # discusses comment syntax in PROSE (plan-provenance.md) where
            # a real citation can sit elsewhere in the same span; matching
            # anywhere would skip that citation unseen. A comment that is
            # not span-initial at all (`` `see <!-- … --> for detail` ``)
            # is accepted, documented residue instead — see
            # specs/myflow-citation-roots/spec.md.
            if span.lstrip().startswith("<!--"):
                i = j + 1
                continue  # an HTML comment, illustrating syntax — no tokens at all
            words = merge_bracket_placeholder_words(span.split())
            tokens.extend(words)
            i = j + 1
        else:
            i += 1
    return tokens


def scan_file_for_citations(text):
    """Return a list of (lineno, token) candidate citations from <text>,
    1-based line numbers. See this module's docstring, "THE CLASSIFIER",
    for the fence-tracking rule this implements.
    """
    candidates = []
    fence_len = 0
    fence_lang = ""

    for lineno, raw in enumerate(text.split("\n"), start=1):
        stripped = raw.strip()

        if fence_len == 0:
            m = FENCE_OPEN_RE.match(stripped)
            if m:
                fence_len = len(m.group(1))
                fence_lang = m.group(2).strip()
                continue
            for tok in extract_backtick_tokens(raw):
                candidates.append((lineno, tok))
            continue

        # Inside a fence: does this line close it? A run of backticks at
        # least as long as the opener, with nothing trailing.
        j = 0
        while j < len(stripped) and stripped[j] == "`":
            j += 1
        trail = stripped[j:].rstrip()
        if j >= fence_len and trail == "":
            fence_len = 0
            fence_lang = ""
            continue

        if SHELL_FENCE_LANG_RE.match(fence_lang):
            if stripped.startswith("#"):
                for tok in merge_bracket_placeholder_words(raw.split()):
                    candidates.append((lineno, tok))
            # A non-comment line inside a bash/sh/zsh fence is a shell
            # argument and is excluded wholesale — no candidates at all.
        # A fence tagged with any other language (or none) is left alone
        # entirely, matching check-guard-symlinks.sh's own RULE3_AWK/
        # CITATION_AWK precedent: never scanned as shell, never as prose.

    return candidates


def classify_token(token, repo_root_files):
    """True when <token> is a citation at all — before its root is ever
    judged. False for every excluded shape this module's docstring names.

    Two exclusions below were added by task 9, after the real corpus
    proved they were needed (each closed exactly one of four permanent
    false positives waves A-C could not rewrite without writing something
    false):

      - A `../`-rooted token: relative to something the document never
        names, so there is no root that could correctly prefix it.
      - A token carrying a leading or trailing quote character (`'`
        or `"`): a real citation in this corpus is never wrapped in one
        inside its own backtick span — a quoted fragment of program
        output (a git error message, a shell one-liner's own quoting) is.

    Two more were added by task 10, after wave C's review found sites
    that satisfy a root judgment while remaining wrong — a git branch
    name told an agent to create a branch of that literal name, and a
    findings-table Location example taught a fabricated citation shape:

      - A GIT_BRANCH_SPECTRE_RE match: `spectre/<…>` names a branch, not
        a filesystem path — see that pattern's own comment for the exact,
        deliberately narrow bound (a shape now, not an exact-literal set,
        after the second per-task review found a second live site).
      - A GIT_BRANCH_CHORE_RE match: `chore/archive-<…>` or
        `chore/self-review-<…>` — kan-239's task 9, a sibling of the
        GIT_BRANCH_SPECTRE_RE exclusion above for the pipeline's other
        branch-name shapes. Run 2 creates the first; the second predates
        kan-239 and survives only as existing branches. See that
        pattern's own comment for the same deliberately narrow bound.
      - A token matching FILE_LINE_RE: a `file:line` location, diff-
        relative by construction, not a path to cite. Narrowing this one
        was tried and rejected — see FILE_LINE_RE's own comment for why it
        stays a full-token exclusion, a documented hole rather than an
        undocumented one.

    Two more were added by the second per-task review, closing the
    remaining real-corpus sites finding 2 (classify every word, not only a
    span's first) newly made visible without ALSO forcing a root onto
    fabricated illustrative content — the HTML-comment case that round
    first drew as a token-level rule here now lives one layer up, in
    extract_backtick_tokens: see that function's own docstring for why a
    comment is a property of the whole span and round 4 moved it there:

      - A token with any NON-FINAL `/`-delimited segment ending in `:`:
        not a path segment — this is what excludes a fragment of
        check-plan-provenance.py's own error-message sentence
        (`measured:/predicted:`), quoted verbatim rather than reworded,
        since rewording it would falsify the quote. Narrowed to non-final
        (round 3) after the first cut silently dropped a real citation
        ending its own span in a bare colon.
    """
    tok = token.strip()
    if not tok:
        return False
    if tok.startswith("/"):
        return False  # absolute path
    if tok.startswith("~"):
        return False  # home-rooted path
    if tok.startswith("../"):
        return False  # parent-relative path — no document-relative root exists to write
    if URL_RE.match(tok):
        return False  # URL
    if tok[0] in "'\"" or tok[-1] in "'\"":
        return False  # quoted program output, not a citation
    if GIT_BRANCH_SPECTRE_RE.match(tok):
        return False  # spectre/<…> — a branch name, not a path
    if GIT_BRANCH_CHORE_RE.match(tok):
        return False  # chore/archive-<…> or chore/self-review-<…> — a branch name, not a path
    if FILE_LINE_RE.search(tok):
        return False  # file:line location, not a path to cite
    # Panel round 3: narrowed to a NON-FINAL segment. The original,
    # unbounded version tested every segment, including the last — which
    # silently dropped a real citation carrying a trailing colon inside
    # its own span (`.flow/project.md:`, e.g. from prose reading "see
    # `.flow/project.md:` for the list"). `measured:/predicted:` still
    # excludes correctly: its non-final segment `measured:` still ends in
    # `:`. Only a colon on a segment that is NOT the path's own last
    # component signals "this is not a path" — a real citation's own
    # final segment ending in `:` is just trailing punctuation.
    segs = tok.split("/")
    if any(seg.endswith(":") for seg in segs[:-1]):
        return False  # a non-final segment ending in ':' is not a path segment

    first_seg = tok.split("/", 1)[0]
    if first_seg.startswith("$"):
        return False  # shell variable reference
    if GIT_REF_PATH_RE.match(tok) or GIT_BRANCH_RE.match(tok):
        return False  # git ref shape
    if first_seg == "origin":
        remainder = tok[len("origin/"):] if tok.startswith("origin/") else tok
        # Panel round 3: a trailing "/" names a DIRECTORY, and a directory
        # is never a git ref, so `origin/rules/` must not take this
        # exclusion (it silently vanished before this bound: a real,
        # reportable unrooted citation, gone). `origin/README` — no
        # extension, no trailing slash — stays excluded: a branch
        # genuinely could be named README, and that residue is accepted
        # rather than closed (documented here the way FILE_LINE_RE's own
        # hole is documented; the planner mirrors it into the spec).
        if not remainder.endswith("/") and not ORIGIN_REF_EXTENSION_RE.search(remainder):
            return False  # origin/<ref> — ref-shaped remainder, not a path
    if any(c in REGEX_GLOB_CHARS for c in tok):
        return False  # regular-expression or glob fragment

    if "/" not in tok:
        return tok in repo_root_files
    return True


# PLACEHOLDER_ROOTS — the closed set specs/myflow-citation-roots/spec.md
# enumerates. Task 9's first cut accepted ANY first segment shaped
# `<…>` — bracket-shaped alone — which fails open: `<foo>/spectre/specs/
# x.md` passed while naming an unrooted path, and so did a typo of a real
# placeholder (`<changeroot>/`, `<change-root>/`). A guard that reports
# clean while checking nothing is worse than no guard — the same
# principle wave A's own `~/` review finding rested on: a site leaves the
# report by acquiring a root, never by ceasing to look like a citation.
# Adding a placeholder root here is a deliberate act that extends the
# spec's own table; it is not something this guard does by pattern-
# matching brackets.
#
# `<skill-dir>` — the second per-task review's sixth root, added after
# finding six real corpus sites (pipeline.md's own "Guard resolution"
# section and handoff-blocks.md) all naming the SAME concept under three
# different English-phrase wordings ("the running command's own skill
# directory", "the producing command's own skill directory", "its own
# skill directory"). It resolves at runtime to the INSTALLED skills root —
# `~/.claude/skills/<skill>/`, `.cursor/skills/<skill>/`, and so on —
# which is neither this checkout nor the target project, the same
# criterion that earned `<abs-worktree>`, `<changeRoot>` and `<state-dir>`
# their places. All three corpus wordings are collapsed to this one short
# form; the surrounding prose (pipeline.md's "Guard resolution" section is
# canonical for the concept) already explains what it means, so the short
# form loses nothing.
PLACEHOLDER_ROOTS = frozenset(
    {
        "<agents repo>",
        "<project>",
        "<abs-worktree>",
        "<changeRoot>",
        "<state-dir>",
        "<skill-dir>",
    }
)


def judges_ok(token, roots):
    """True when <token> — already classified as a citation — names a
    recognised root: an installed root's own first segment, or a first
    segment that is a MEMBER of the closed PLACEHOLDER_ROOTS set above —
    never merely bracket-shaped. See that set's own comment for why this
    is a closed membership test rather than a `startswith("<")` pattern.
    """
    first_seg = token.split("/", 1)[0]
    if first_seg in PLACEHOLDER_ROOTS:
        return True
    return first_seg in roots


# ---------------------------------------------------------------------------
# Reporting
# ---------------------------------------------------------------------------


def main():
    try:
        sys.stdout.reconfigure(line_buffering=True)
    except AttributeError:
        pass

    repo_root = resolve_repo_root()
    if not os.path.isdir(repo_root):
        print(f"{GUARD_NAME}: {repo_root} is not a directory — cannot scan", file=sys.stderr)
        sys.exit(2)
    repo_root = os.path.realpath(repo_root)

    sandbox = tempfile.mkdtemp(prefix="check-installed-citations.")
    try:
        home = sandbox
        proj = os.path.join(sandbox, "proj")
        os.makedirs(proj, exist_ok=True)

        for mode, project_dir in (("global", None), ("all", proj)):
            try:
                result = run_setup(repo_root, sandbox, home, project_dir, mode)
            except SandboxRefusal as exc:
                print(f"{GUARD_NAME}: refusing to run setup.sh — {exc}", file=sys.stderr)
                sys.exit(2)
            if result.returncode != 0:
                print(
                    f"{GUARD_NAME}: sandboxed `setup.sh {mode}` exited "
                    f"{result.returncode} — cannot derive the installed set",
                    file=sys.stderr,
                )
                if result.stdout:
                    print(result.stdout, file=sys.stderr)
                if result.stderr:
                    print(result.stderr, file=sys.stderr)
                sys.exit(2)

        sources = derive_installed_set(repo_root, sandbox)
    finally:
        shutil.rmtree(sandbox, ignore_errors=True)

    roots = compute_roots(sources)
    corpus = compute_corpus(repo_root, sources)

    repo_root_files = {
        name
        for name in os.listdir(repo_root)
        if os.path.isfile(os.path.join(repo_root, name))
        and not os.path.islink(os.path.join(repo_root, name))
    }

    violations = []
    coverage = []

    for relpath in sorted(corpus):
        full = os.path.join(repo_root, relpath)
        try:
            with open(full, "r", encoding="utf-8") as fh:
                text = fh.read()
        except OSError as exc:
            print(f"{GUARD_NAME}: cannot read {relpath} — {exc}", file=sys.stderr)
            sys.exit(2)
        except UnicodeDecodeError as exc:
            print(f"{GUARD_NAME}: cannot decode {relpath} as UTF-8 — {exc}", file=sys.stderr)
            sys.exit(2)

        checked = 0
        for lineno, tok in scan_file_for_citations(text):
            if not classify_token(tok, repo_root_files):
                continue
            checked += 1
            if not judges_ok(tok, roots):
                violations.append((relpath, lineno, tok))
        coverage.append((relpath, checked))

    # This script does not decide clean-vs-violations, print a verdict
    # line, or choose an exit code beyond refusal (2) — its wrapper
    # (check-installed-citations.sh) owns all of that, because the wrapper
    # is also where scripts/lib/coverage.sh's coverage_declare/
    # coverage_verdict decide whether an undeclared zero is itself a
    # violation (finding 3 — that decision has to be made through the
    # real Bash library, per task 2 step 5 and this repository's other
    # corpus-scanning guards, not reinvented here in Python). So this
    # script hands back exactly two kinds of line and nothing else: a
    # real citation violation as-is, and everything the wrapper needs to
    # finish the job (the resolved root, once, and one coverage line per
    # scanned member) behind the sentinels above.
    for relpath, lineno, tok in violations:
        print(
            f"{relpath}:{lineno}: citation `{tok}` names no root — "
            "prefix with <agents repo>/ or <project>/"
        )
    print(f"{CIC_ROOT_PREFIX}{repo_root}")
    for relpath, count in coverage:
        print(f"{CIC_COVERAGE_PREFIX}{relpath}\t{count}")
    sys.exit(0)


if __name__ == "__main__":
    main()
