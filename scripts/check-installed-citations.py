#!/usr/bin/env python3
"""check-installed-citations.py — fail when a path citation in an installed
file names no root.

A file `setup.sh` installs or copies is read by an agent standing in
whatever tree the harness placed it — never necessarily this checkout. A
bare citation like `` `README.md` `` in such a file is read against THAT
tree, which is exactly how a bug ships: the citation resolves to nothing,
or worse, to some unrelated file the target project happens to have.
Exactly three forms are recognised (canonical definition:
specs/myflow-citation-roots/spec.md — do not restate the rule here):

  - an installed root's own bare form (`skills/…`, `rules/…`, `commands/…`,
    `commands-claude/…`, `hooks/…`, wherever the harness placed it)
  - `<agents repo>/…` — this checkout
  - `<project>/…` — the project the running command is working against

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
# "$" is deliberately excluded from this set: a leading "$" is the shell-
# variable-reference exclusion's own signal (see classify_token), and
# keeping it here too would make that exclusion untestable in isolation —
# every shell-variable token already contains a "$" and would be caught by
# this set regardless, which is exactly the untested-mutation shape KAN-197
# names (a case that still passes with the line it covers deleted proves
# nothing). A genuine glob/regex fragment is still caught by the other
# metacharacters below.
REGEX_GLOB_CHARS = set("[](){}*+?|^")

AGENTS_REPO_WORD1 = "<agents"
AGENTS_REPO_WORD2_PREFIX = "repo>/"


def merge_agents_repo_prefix(words):
    """Merge a literal `<agents repo>/…` prefix back into one word.

    The placeholder itself contains a space ("<agents repo>"), so a plain
    whitespace split — used both to collapse a multi-word backtick span to
    its leading word (extract_backtick_tokens) and to tokenize a shell-
    fence comment line (scan_file_for_citations) — breaks it into two
    fragments, `<agents` and `repo>/…`, before either ever gets a chance to
    classify it. Left unmerged, the first fragment is a citation candidate
    that never reaches classify_token/judges_ok at all (silently dropped —
    finding 1), and the second is judged on its own, standalone, as a bogus
    unrooted citation `` `repo>/…` `` (finding 2). Recognised only in this
    exact two-word shape, case-sensitive, since that is the only shape this
    repository's own convention writes; `<project>/…` has no embedded
    space and never needed this.
    """
    merged = []
    i = 0
    n = len(words)
    while i < n:
        if (
            words[i] == AGENTS_REPO_WORD1
            and i + 1 < n
            and words[i + 1].startswith(AGENTS_REPO_WORD2_PREFIX)
        ):
            merged.append(words[i] + " " + words[i + 1])
            i += 2
        else:
            merged.append(words[i])
            i += 1
    return merged


def extract_backtick_tokens(line):
    """Every span between a pair of single backticks on <line>, in order —
    only its FIRST whitespace-delimited word, never the whole span.
    Greedy left-to-right pairing — this file's corpus writes citations in
    plain single-backtick spans, never nested or doubled, so the extra
    machinery check-references.sh's normalized_line carries for a
    soft-wrapped orphan marker has no call site here.

    A single-backtick span carrying more than one word is an inline shell
    example — `` `agents/setup.sh global` ``, `` `sed -n '/…/p'` `` — not a
    bare citation; check-guard-symlinks.sh's own CITATION_AWK already
    treats a backtick span this way (its scan_prose extracts only the
    leading `[A-Za-z0-9._-]+` run as the base token, everything after as
    context). This guard needs slashes to survive where that one does not,
    so the cut is at the first whitespace rather than the first non-word
    character, but the shape of the decision — only the LEADING word of a
    multi-word span is ever a citation candidate — is the same one.
    """
    tokens = []
    i = 0
    n = len(line)
    while i < n:
        if line[i] == "`":
            j = line.find("`", i + 1)
            if j == -1:
                break
            words = merge_agents_repo_prefix(line[i + 1 : j].split())
            if words:
                tokens.append(words[0])
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
                for tok in merge_agents_repo_prefix(raw.split()):
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
    """
    tok = token.strip()
    if not tok:
        return False
    if tok.startswith("/"):
        return False  # absolute path
    if tok.startswith("~"):
        return False  # home-rooted path
    if URL_RE.match(tok):
        return False  # URL

    first_seg = tok.split("/", 1)[0]
    if first_seg.startswith("$"):
        return False  # shell variable reference
    if first_seg == "origin" or GIT_REF_PATH_RE.match(tok) or GIT_BRANCH_RE.match(tok):
        return False  # git ref shape
    if any(c in REGEX_GLOB_CHARS for c in tok):
        return False  # regular-expression or glob fragment

    if "/" not in tok:
        return tok in repo_root_files
    return True


def judges_ok(token, roots):
    """True when <token> — already classified as a citation — names a
    recognised root: an installed root's own first segment, or the
    literal `<agents repo>/` or `<project>/` prefix.
    """
    if token.startswith("<agents repo>/") or token.startswith("<project>/"):
        return True
    first_seg = token.split("/", 1)[0]
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
