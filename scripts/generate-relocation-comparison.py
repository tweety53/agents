#!/usr/bin/env python3
"""generate-relocation-comparison.py — for a plan that declares itself a
relocation, generate a before/after passage comparison across every file
the plan's own tasks scope, so a review panel can audit a move the same way
`myflow-contract-economy`'s own per-move ledger does (canonical spec:
`openspec/specs/myflow-review-panel-economics/spec.md`'s "A
relocation-declaring plan gets a mechanical passage comparison" requirement
— do not restate its rules here, only implement them).

Reuses `plan-dispatch-bundles.py`'s own `parse_tasks` (and, through it, its
`**Files:**` field grammar) rather than reimplementing the task/field
parser — the same import discipline `check-plan-shape.py` uses to load
`check-task-commit-fields.py`: `importlib.util.spec_from_file_location`
against this file's own real directory, with the loaded module REGISTERED
IN `sys.modules` before `exec_module` runs. `plan-dispatch-bundles.py`
carries no `@dataclass` needing that registration to avoid the Python 3.14
`AttributeError` check-plan-shape.py's own docstring documents, but the
registration is done anyway for the same reason: it is what makes the
loaded module resolvable when Python next imports anything by that name.

Invocation: `<worktree>` `<change-root>` `<merge-base>` `<tasks.md path>`.
`generate-relocation-comparison.sh` is the thin wrapper that resolves the
first three positionally and appends the fourth
(`<changeRoot>/tasks.md`) itself. `<change-root>` is accepted but unused by
this script's own logic — the comparison's output path is always
`<worktree>/.superpowers/sdd/relocation-comparison.md`, independent of
where the plan itself lives.

Exit codes:
  0  written (a table naming every moved / repointed / added / removed
     passage across the plan's scoped files), OR the no-op case — a
     missing, `no`, or malformed `**Relocation:**` header, or an explicit
     `yes` with an empty file scope. Both are success: this script never
     blocks the panel it feeds (see the "never blocks" global constraint
     this change's tasks.md states).
  2  cannot answer: `tasks.md` cannot be read, `<merge-base>` does not
     resolve to a real commit, a file the plan scopes is missing from the
     worktree on disk, a scoped path that is a tree (directory) rather
     than a blob at `<merge-base>`, a `git show <merge-base>:<path>`
     failure once `git cat-file -t <merge-base>:<path>` has already said
     the path is a blob there, or a scoped file (worktree copy or
     merge-base blob) that is not valid UTF-8.

Algorithm
---------

1. Scan `tasks.md`'s lines for the first line matching
   `^\\s*>\\s*\\*\\*Relocation:\\*\\*\\s+(yes|no)\\b`. Anything else — no
   match at all (header missing or malformed) or an explicit `no` — is the
   no-op case: exit 0, write nothing.
2. On `yes`, parse every task in the plan (via `parse_tasks`, checked and
   unchecked alike — the header is plan-level, not scoped to open work) and
   union every task's `**Files:**` paths, plan order preserved, duplicates
   dropped. An empty union is the same no-op case as step 1.
3. For each scoped path, extract PASSAGES from two texts: the blob at
   `<merge-base>:<path>` (via `git show`, gated by `git cat-file -t
   <merge-base>:<path>` first; a path `cat-file -t` says is absent at that
   revision is read as empty text without even calling `git show` — the
   ordinary shape of a file this plan creates fresh; a path that IS
   present there but as a `tree`, not a `blob` — a directory at the
   merge-base ref — is the exit-2 "cannot answer" case (F18), never
   silently read as `git show`'s tree-listing text; any OTHER `git show`
   failure once `cat-file -t` says the path IS a blob, or a blob that is
   not valid UTF-8, is likewise the exit-2 "cannot answer" case, not a
   silent "added" row) and the file's current on-disk content in the
   worktree (read directly; missing here is the exit-2 "scoped file
   missing" case, since a plan task can only declare a path the fix is
   supposed to leave in place or remove — this generation step cannot tell
   those apart from "wrongly scoped", so it reports rather than guesses;
   a worktree file that is not valid UTF-8 is the same exit-2 case).

   A passage is: a run of non-blank lines between two blank lines or
   `##`/`###` headings; OR a single `- `/`* ` bullet line; OR a single
   Markdown table row (`| ... |`). Each passage carries the nearest
   preceding `##`/`###` heading text as its location's second component.
4. Build the before-multiset and after-multiset of `(text, file, heading)`
   tuples across every scoped file. For passages sharing one text:
     - a `(file, heading)` pair present in both sets is UNCHANGED — not a
       row;
     - a leftover before occurrence and a leftover after occurrence of the
       same text (necessarily at different locations, since equal
       locations were already removed as unchanged) pair up as **moved**.
   Remaining before-only texts and after-only texts are then matched by
   REPOINT_SIGNATURE — a normalisation that collapses every backtick-quoted
   span, and the bold (`**...**`) span immediately beside it (the one
   citation shape `myflow-contract-economy`'s spec permits to change), to a
   placeholder, and strips a trailing `above`/`below` word — the same two
   edits, and no others. A bold span anywhere else in the passage is left
   alone. Trailing sentence-ending punctuation is then stripped from BOTH
   sides unconditionally, so a trailing `above`/`below` word's own
   punctuation (removed together with the word on whichever side had it)
   never by itself diverges the two signatures (F16). Two texts sharing a
   signature but not their raw text are **repointed**. Whatever is left is
   **added** (after-only) or **removed** (before-only).
5. Write `<worktree>/.superpowers/sdd/relocation-comparison.md`: a Markdown
   table, columns Passage (the passage's own text, first eight words) /
   Source / Destination / Class, one row per moved/repointed/added/removed
   passage — never one for a passage unchanged in both text and location.
   A literal `|` in a Passage/Source/Destination cell is escaped as `\\|`
   so it cannot be misread as a column separator.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import argparse
import importlib.util
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

SCRIPT_DIR = Path(os.path.dirname(os.path.realpath(__file__)))

# --- Load plan-dispatch-bundles.py for its parse_tasks / Task grammar ------
#
# Named and checked explicitly (not just imported): a Python `import` is
# invisible to check-guard-symlinks.sh's rule 2, which derives a guard's
# required siblings by grepping its SOURCE for `$SCRIPT_DIR/<name>` — this
# check here is defense in depth for anyone invoking this .py directly, the
# same defensive pattern check-plan-shape.py's own module docstring uses
# for check-task-commit-fields.py.
_PDB_PATH = SCRIPT_DIR / "plan-dispatch-bundles.py"
if not _PDB_PATH.is_file():
    print(
        f"generate-relocation-comparison.py: required sibling module not found: {_PDB_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)

_pdb_spec = importlib.util.spec_from_file_location("plan_dispatch_bundles", str(_PDB_PATH))
if _pdb_spec is None or _pdb_spec.loader is None:
    print(f"generate-relocation-comparison.py: cannot load module: {_PDB_PATH}", file=sys.stderr)
    sys.exit(2)
_pdb = importlib.util.module_from_spec(_pdb_spec)
# Registered before exec_module, matching check-plan-shape.py's own
# precedent for loading a hyphenated sibling by path.
sys.modules["plan_dispatch_bundles"] = _pdb
_pdb_spec.loader.exec_module(_pdb)

parse_tasks = _pdb.parse_tasks

# RELOCATION_RE — the plan-level header line. Scanned line by line; the
# first match wins. Anything that never matches (missing header, `no`, or a
# malformed value) is the no-op case.
RELOCATION_RE = re.compile(r"^\s*>\s*\*\*Relocation:\*\*\s+(yes|no)\b")

# HEADING_RE — the `##`/`###` heading that both ends a passage run and
# names the location of every passage that follows it, until the next one.
HEADING_RE = re.compile(r"^#{2,3}\s+(.*)$")

# BULLET_RE — a single `- `/`* ` bullet line, its own one-line passage.
BULLET_RE = re.compile(r"^\s*[-*]\s+.*\S.*$")

# TABLE_ROW_RE — a single Markdown table row, its own one-line passage.
TABLE_ROW_RE = re.compile(r"^\s*\|.*\|\s*$")

# Passage location: the scoped file path, plus the nearest preceding
# heading text (None above the first heading in the file).
Location = Tuple[str, Optional[str]]


def find_relocation(lines: List[str]) -> Optional[bool]:
    """Return True for an explicit `yes`, False for an explicit `no`, or
    None when no line in `lines` matches RELOCATION_RE at all — the
    "missing" case, treated identically to an explicit `no`."""
    for line in lines:
        match = RELOCATION_RE.match(line)
        if match:
            return match.group(1) == "yes"
    return None


def scoped_files(lines: List[str]) -> List[str]:
    """Union of every task's **Files:** paths across the whole plan (task
    checked/unchecked status is irrelevant here — the header is
    plan-level), plan order preserved, duplicates dropped."""
    seen: Dict[str, None] = {}
    for task in parse_tasks(lines):
        for path in task.files:
            if path not in seen:
                seen[path] = None
    return list(seen.keys())


def extract_passages(text: str, file_path: str) -> List[Tuple[str, Location]]:
    """Split `text` (the content of `file_path`, either the merge-base blob
    or the worktree's current content) into passages, each tagged with its
    `(file_path, heading)` location. See module docstring step 3 for the
    passage grammar."""
    passages: List[Tuple[str, Location]] = []
    heading: Optional[str] = None
    buffer: List[str] = []

    def flush() -> None:
        if buffer:
            passage_text = "\n".join(buffer).strip()
            if passage_text:
                passages.append((passage_text, (file_path, heading)))
            buffer.clear()

    for line in text.splitlines():
        heading_match = HEADING_RE.match(line)
        if heading_match:
            flush()
            heading = heading_match.group(1).strip()
            continue
        if line.strip() == "":
            flush()
            continue
        if BULLET_RE.match(line) or TABLE_ROW_RE.match(line):
            flush()
            passages.append((line.strip(), (file_path, heading)))
            continue
        buffer.append(line)
    flush()
    return passages


# BACKTICK_SPAN_RE — every backtick-quoted span. Applied with `.finditer`
# to get an ORDERED, NON-OVERLAPPING list of exact `(start, end)` spans —
# each span's boundaries are unambiguous, since `[^`]*` cannot cross a
# backtick it does not own. This is what a single combined regex (F13's
# root cause) could not guarantee: a citation-adjacency pattern built as
# one `` `[^`]*`...\*\*[^*]*\*\* `` regex can start its "backtick span"
# match at the CLOSING tick of one real span, run through unrelated text,
# and end at the OPENING tick of a second, unrelated span — silently
# absorbing a genuine content edit inside the first span as if it were
# "the separator" before a citation's bold token. Finding every span
# first, independently, and pairing by exact position below, cannot
# reproduce that failure mode: a span's own end position is fixed before
# any pairing decision is made.
BACKTICK_SPAN_RE = re.compile(r"`[^`]*`")

# BOLD_SPAN_RE — every bold (`**...**`) span, found independently of
# backtick spans.
BOLD_SPAN_RE = re.compile(r"\*\*[^*]*\*\*")

# CITATION_SEPARATOR_RE — the exact separator styles the corpus's own
# citations use between a backtick-quoted path and the bold section token
# beside it: nothing but whitespace (e.g. `` `engineering-principles.md`
# **beside this file** ``), a possessive apostrophe-s (`` `skills/flow/
# SKILL.md`'s **Model resolution** ``), or a comma-and-space (`` `old.md`,
# **Old Section** ``) — never anything wider. Matched with `.match` at a
# backtick span's exact end position, so it can only ever consume the text
# immediately following THAT span — never skip over another backtick span,
# which contains a `` ` `` this pattern cannot match through.
CITATION_SEPARATOR_RE = re.compile(r"(?:'s|,)?[ \t]{0,3}")

# TRAILING_POSITION_WORD_RE — a trailing `above`/`below` word (the second
# permitted edit), with a RUN of trailing sentence-ending punctuation
# (including `!`/`?`, not just `.,;:`), stripped before comparison (F20: a
# `?`/`*` single-character class only ever stripped one punctuation mark
# from a narrow set, so "...above..." or "...above!?" still left trailing
# punctuation behind after the word itself was dropped, diverging the two
# signatures even after a genuine repoint). The `\s*` between the word and
# its punctuation run (F22: the run previously had to sit IMMEDIATELY after
# the word, so "...look above  ." — word, whitespace, then punctuation —
# left the word's trailing space behind after stripping, diverging from
# "...look ." even though the only real edit was the permitted one) lets
# whitespace separate the word from a punctuation run that still follows it.
TRAILING_POSITION_WORD_RE = re.compile(r"\b(above|below)\b\s*[.,;:!?]*\s*$")

# TRAILING_PUNCT_RE — a RUN of trailing sentence-ending punctuation
# (including `!`/`?`), stripped as its own, always-applied normalisation
# step (F16, widened by F20). TRAILING_POSITION_WORD_RE above consumes a
# trailing `above`/`below` word TOGETHER WITH any punctuation run
# immediately following it, but only on the side of the comparison that
# actually has that trailing word — the other side (which never had the
# word) keeps its own trailing punctuation untouched. That made a genuine
# repoint (dropping a stale `above`/`below` after a move) misclassify as
# removed+added: "...look above." strips to "...look", while "...look."
# keeps its period, so the two signatures differ by one character even
# though the only real edit was the permitted one. Stripping a trailing
# punctuation RUN from BOTH sides, unconditionally, after the above/below
# strip, makes punctuation presence — one mark or several, `.,;:` or
# `!`/`?` — never by itself a source of signature mismatch.
TRAILING_PUNCT_RE = re.compile(r"[.,;:!?]*\s*$")


def repoint_signature(text: str) -> str:
    """Normalise `text` the way `myflow-contract-economy`'s two permitted
    moved-passage edits allow, via a tokenizer pass rather than a single
    combined regex substitution (F13):

    1. Find every backtick span and every bold span independently, as
       ordered, non-overlapping `(start, end)` positions.
    2. Pair a backtick span with a bold span only when the bold span
       starts at EXACTLY the position CITATION_SEPARATOR_RE's match ends,
       measured from that backtick span's own end — i.e. the bold span is
       genuinely the next thing after that specific backtick span, with
       nothing (not even another backtick span) in between.
    3. Walk `text` left to right emitting it unchanged except: a paired
       (backtick+adjacent-bold) region collapses to one placeholder; an
       unpaired backtick span collapses to its own placeholder; an
       unpaired bold span is left untouched (never collapsed) — so a bold
       edit elsewhere in the passage still falls through to removed+added.
    4. Drop a trailing `above`/`below` word.

    Two passages sharing a signature, but not their raw text, differ by
    exactly those edits — REPOINTED, not ADDED/REMOVED."""
    backtick_spans = [(m.start(), m.end()) for m in BACKTICK_SPAN_RE.finditer(text)]
    bold_span_by_start: Dict[int, int] = {m.start(): m.end() for m in BOLD_SPAN_RE.finditer(text)}

    # events: (start, end, kind) — "citation" collapses to the two-token
    # placeholder, "backtick" to the single-token one. An unpaired bold
    # span gets no event at all, so it is copied through untouched.
    events: List[Tuple[int, int, str]] = []
    for start, end in backtick_spans:
        sep_match = CITATION_SEPARATOR_RE.match(text, end)
        bold_start = sep_match.end() if sep_match else end
        bold_end = bold_span_by_start.get(bold_start)
        if bold_end is not None:
            events.append((start, bold_end, "citation"))
        else:
            events.append((start, end, "backtick"))

    events.sort(key=lambda e: e[0])

    pieces: List[str] = []
    pos = 0
    for start, end, kind in events:
        if start < pos:
            # A backtick span consumed by an earlier citation pairing (its
            # bold portion overlaps this span's start) — nothing left to
            # emit for it on its own.
            continue
        pieces.append(text[pos:start])
        pieces.append("\x00PATH\x00\x00BOLD\x00" if kind == "citation" else "\x00PATH\x00")
        pos = end
    pieces.append(text[pos:])

    normalized = "".join(pieces)
    normalized = TRAILING_POSITION_WORD_RE.sub("", normalized).rstrip()
    normalized = TRAILING_PUNCT_RE.sub("", normalized).rstrip()
    return normalized


def classify(
    before: List[Tuple[str, Location]], after: List[Tuple[str, Location]]
) -> List[Tuple[str, Optional[Location], Optional[Location], str]]:
    """Return rows as (passage_text, source, destination, class) — never one
    for a passage unchanged in both text and location. See module
    docstring step 4 for the classification rule."""
    before_by_text: Dict[str, List[Location]] = {}
    for text, loc in before:
        before_by_text.setdefault(text, []).append(loc)
    after_by_text: Dict[str, List[Location]] = {}
    for text, loc in after:
        after_by_text.setdefault(text, []).append(loc)

    rows: List[Tuple[str, Optional[Location], Optional[Location], str]] = []
    before_only: List[Tuple[str, Location]] = []
    after_only: List[Tuple[str, Location]] = []

    for text, before_locs in before_by_text.items():
        after_locs = after_by_text.get(text)
        if after_locs is None:
            for loc in before_locs:
                before_only.append((text, loc))
            continue
        remaining_before = list(before_locs)
        remaining_after = list(after_locs)
        # Unchanged: an exact (file, heading) pair present in both — remove
        # one occurrence from each side per shared location.
        for loc in list(remaining_before):
            if loc in remaining_after:
                remaining_before.remove(loc)
                remaining_after.remove(loc)
        # Whatever is left shares text but not location: MOVED.
        for before_loc, after_loc in zip(remaining_before, remaining_after):
            rows.append((text, before_loc, after_loc, "moved"))
        leftover_before = remaining_before[len(remaining_after):]
        leftover_after = remaining_after[len(remaining_before):]
        for loc in leftover_before:
            before_only.append((text, loc))
        for loc in leftover_after:
            after_only.append((text, loc))

    for text, after_locs in after_by_text.items():
        if text not in before_by_text:
            for loc in after_locs:
                after_only.append((text, loc))

    # REPOINTED: match remaining before-only/after-only passages by
    # normalised signature.
    before_by_sig: Dict[str, List[Tuple[str, Location]]] = {}
    for text, loc in before_only:
        before_by_sig.setdefault(repoint_signature(text), []).append((text, loc))
    matched_after_ids = set()
    remaining_before_only: List[Tuple[str, Location]] = []
    for text, loc in before_only:
        sig = repoint_signature(text)
        candidates = [
            (index, after_text, after_loc)
            for index, (after_text, after_loc) in enumerate(after_only)
            if index not in matched_after_ids and repoint_signature(after_text) == sig
        ]
        if candidates:
            index, after_text, after_loc = candidates[0]
            matched_after_ids.add(index)
            rows.append((text, loc, after_loc, "repointed"))
        else:
            remaining_before_only.append((text, loc))

    for text, loc in remaining_before_only:
        rows.append((text, loc, None, "removed"))
    for index, (text, loc) in enumerate(after_only):
        if index in matched_after_ids:
            continue
        rows.append((text, None, loc, "added"))

    return rows


def _escape_cell(text: str) -> str:
    """Escape a literal `|` so it cannot be read as a Markdown table column
    separator — applied to every Passage/Source/Destination cell (never to
    Class, which is script-controlled and never contains one)."""
    return text.replace("|", "\\|")


def _first_eight_words(text: str) -> str:
    words = text.split()
    truncated = _escape_cell(" ".join(words[:8]))
    return truncated + ("…" if len(words) > 8 else "")


def _format_location(loc: Optional[Location]) -> str:
    if loc is None:
        return "—"
    path, heading = loc
    path = _escape_cell(path)
    return f"`{path}` § {_escape_cell(heading)}" if heading else f"`{path}`"


class GitShowError(Exception):
    """Raised by `_git_show` for any `git show` failure that is NOT the
    ordinary "path absent at that ref" case — a corrupted object, a
    malformed ref, or a non-UTF-8 blob that failed to decode."""


def _object_type_at_ref(worktree: str, ref: str, path: str) -> Optional[str]:
    """Return the object type (`"blob"`, `"tree"`, ...) that `<ref>:<path>`
    names in `worktree`'s repo, or `None` when nothing exists at that
    ref:path at all. Two-step, deliberately not a single `git cat-file -t`
    call: existence is checked FIRST via `git cat-file -e <ref>:<path>` —
    which, as case (l)'s corrupt-loose-object fixture demonstrates, exits 0
    even for an object git cannot actually inflate, since existence alone
    does not require decompressing it — and only once that says the path
    IS present does `git cat-file -t <ref>:<path>` run to learn its type.
    That second call reads (inflates) the object, so a corrupted one fails
    it even though `-e` already said it exists; that failure is a genuine
    "cannot answer" case and is raised as GitShowError here rather than
    folded into the "absent" `None` return — collapsing the two would
    regress case (l) back to silently reading a corrupt blob as an
    ordinary fresh-file "added" row instead of exiting 2. `None` here is
    reserved for the one case `-e` itself reports as absent; no `git show`
    stderr text is parsed for it, matching F14's rationale directly.
    Deliberately NOT deciding "exists" from `-t`'s exit code alone, which
    cannot distinguish "absent" from "present but unreadable" — and NOT
    deciding blob-vs-tree from `-e` alone (F18), which only asks "does some
    object exist at this path" and returns true for a tree (directory)
    object exactly as readily as for a blob: a scoped path that was a
    directory at `ref` would then have `_git_show` read `git show
    <ref>:<path>`'s tree-listing output — a plain listing of entry names,
    not file content — silently as text and feed it into
    `extract_passages`. Returning the type itself lets the caller reject
    anything but `"blob"` explicitly, the same "cannot answer" exit-2
    disposition `_git_show` already uses for a non-UTF-8 blob."""
    exists = subprocess.run(
        ["git", "-C", worktree, "cat-file", "-e", f"{ref}:{path}"],
        capture_output=True,
        text=True,
    )
    if exists.returncode != 0:
        return None
    result = subprocess.run(
        ["git", "-C", worktree, "cat-file", "-t", f"{ref}:{path}"],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise GitShowError(f"git cat-file -t {ref}:{path} failed: {result.stderr.strip()}")
    return result.stdout.strip()


def _is_symlink_at_ref(worktree: str, ref: str, path: str) -> bool:
    """Return True when the tree entry at `<ref>:<path>` is a symlink (mode
    `120000`), via `git ls-tree <ref> -- <path>` — deliberately NOT via
    `git cat-file -t`, which reports `"blob"` for a symlink object exactly
    as it does for ordinary file content (F21): the blob a symlink object
    stores IS its target path string, so type alone cannot tell "this path
    is a symlink" from "this path is a regular file". `git ls-tree`
    reports the tree entry's own mode instead, which does distinguish them
    (`120000` symlink vs. `100644`/`100755` regular file) — the same kind
    of information `_object_type_at_ref`'s type check already uses to
    reject a `tree` (F18), applied one level down to reject a symlink
    entry too."""
    result = subprocess.run(
        ["git", "-C", worktree, "ls-tree", ref, "--", path],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0 or not result.stdout.strip():
        return False
    mode = result.stdout.split(None, 1)[0]
    return mode == "120000"


def _git_show(worktree: str, ref: str, path: str) -> str:
    """Return the blob at `<ref>:<path>` in `worktree`'s repo, or "" when
    `_object_type_at_ref` says nothing exists at `<ref>:<path>` at all (the
    ordinary shape of a file this plan creates fresh, not an error) —
    `git show` is then skipped entirely, since there is nothing to read.
    When `_object_type_at_ref` says something DOES exist there but it is
    not a `"blob"` (a `"tree"` — a directory at that ref — or any other
    object type), that is the same "cannot answer" case as a `git show`
    failure (F18): raises GitShowError rather than letting a tree's entry
    listing be silently read as file content. A path that IS a blob but
    whose tree entry is a SYMLINK (F21) is the same "cannot answer" case —
    `git show <ref>:<path>` on a symlink returns the literal target path
    string, not resolved content, which would silently diverge from
    whatever the worktree side reads (Python's `open()` follows the
    symlink and reads the real target's content) even for a completely
    unchanged symlinked file. When it IS an ordinary blob but `git show`
    still fails (a corrupted object, a malformed ref, or a blob that is
    not valid UTF-8), that is likewise a genuine unexpected failure:
    raises GitShowError so the caller can report it and exit 2 rather than
    silently misreading it as "added" or crashing with a traceback."""
    object_type = _object_type_at_ref(worktree, ref, path)
    if object_type is None:
        return ""
    if object_type != "blob":
        raise GitShowError(f"{ref}:{path} is a {object_type}, not a file (blob)")
    if _is_symlink_at_ref(worktree, ref, path):
        raise GitShowError(f"{ref}:{path} is a symlink, not a regular file")
    try:
        result = subprocess.run(
            ["git", "-C", worktree, "show", f"{ref}:{path}"],
            capture_output=True,
            text=True,
        )
    except UnicodeDecodeError as exc:
        raise GitShowError(f"{ref}:{path} is not valid UTF-8: {exc}") from exc
    if result.returncode != 0:
        raise GitShowError(f"git show {ref}:{path} failed: {result.stderr.strip()}")
    return result.stdout


def _merge_base_resolves(worktree: str, ref: str) -> bool:
    result = subprocess.run(
        ["git", "-C", worktree, "cat-file", "-e", f"{ref}^{{commit}}"],
        capture_output=True,
        text=True,
    )
    return result.returncode == 0


def generate(worktree: str, merge_base: str, tasks_path: str) -> int:
    try:
        with open(tasks_path, "r", encoding="utf-8") as handle:
            tasks_text = handle.read()
    except OSError as exc:
        print(f"generate-relocation-comparison.py: cannot read {tasks_path}: {exc}", file=sys.stderr)
        return 2

    lines = tasks_text.splitlines()
    relocation = find_relocation(lines)
    if relocation is not True:
        return 0

    paths = scoped_files(lines)
    if not paths:
        return 0

    if not _merge_base_resolves(worktree, merge_base):
        print(
            f"generate-relocation-comparison.py: merge-base {merge_base!r} does not resolve to a commit in {worktree}",
            file=sys.stderr,
        )
        return 2

    before: List[Tuple[str, Location]] = []
    after: List[Tuple[str, Location]] = []
    for path in paths:
        worktree_path = Path(worktree) / path
        if worktree_path.is_symlink():
            # F21: a worktree-side symlink is read by Python's `open()`,
            # which FOLLOWS it to the real target's content — silently
            # different from what the merge-base side would read if IT
            # were also a symlink (the literal target path string, per
            # `_is_symlink_at_ref` above). Rejecting here, the same
            # "cannot answer" exit-2 disposition, means neither side ever
            # gets to silently resolve a symlink the other side would not.
            print(
                f"generate-relocation-comparison.py: scoped file is a symlink in the worktree, cannot compare: {worktree_path}",
                file=sys.stderr,
            )
            return 2
        try:
            with open(worktree_path, "r", encoding="utf-8") as handle:
                after_text = handle.read()
        except OSError as exc:
            print(
                f"generate-relocation-comparison.py: scoped file missing from worktree: {worktree_path}: {exc}",
                file=sys.stderr,
            )
            return 2
        except UnicodeDecodeError as exc:
            print(
                f"generate-relocation-comparison.py: {worktree_path} is not valid UTF-8: {exc}",
                file=sys.stderr,
            )
            return 2
        try:
            before_text = _git_show(worktree, merge_base, path)
        except GitShowError as exc:
            print(f"generate-relocation-comparison.py: {exc}", file=sys.stderr)
            return 2
        before.extend(extract_passages(before_text, path))
        after.extend(extract_passages(after_text, path))

    rows = classify(before, after)

    output_path = Path(worktree) / ".superpowers" / "sdd" / "relocation-comparison.md"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    lines_out = ["| Passage | Source | Destination | Class |", "| --- | --- | --- | --- |"]
    for text, source, destination, cls in rows:
        lines_out.append(
            f"| {_first_eight_words(text)} | {_format_location(source)} | {_format_location(destination)} | {cls} |"
        )
    output_path.write_text("\n".join(lines_out) + "\n", encoding="utf-8")
    return 0


def main(argv: List[str]) -> int:
    parser = argparse.ArgumentParser(prog=argv[0] if argv else "generate-relocation-comparison.py")
    parser.add_argument("worktree")
    parser.add_argument("change_root")
    parser.add_argument("merge_base")
    parser.add_argument("tasks_path")
    try:
        args = parser.parse_args(argv[1:])
    except SystemExit as exc:
        return exc.code if isinstance(exc.code, int) else 2

    return generate(args.worktree, args.merge_base, args.tasks_path)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
