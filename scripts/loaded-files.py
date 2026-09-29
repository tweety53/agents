#!/usr/bin/env python3
"""loaded-files.py [transcript.jsonl | dir ...] -- which flow instruction files each Claude Code
session actually pulled into context, and what they cost (KAN-852). A report, never a verdict:
it always exits 0.

One JSONL transcript is one session (/clear starts a new one). For every Read, Skill load, or
Bash cat/sed/head/tail/awk of a *.md or *.mdc under skills/, rules/ or commands-claude/, it
reports the bytes that entered context, the API turn they entered at, the files read more than
once, and the turn-weighted re-read cost -- bytes/4 x the API turns left in the session, since
every later turn re-reads the whole context as cache.

With no argument it reads every ~/.claude/projects/*/*.jsonl; a directory argument reads the
*.jsonl files directly inside it. Ported from the prototype loaded_files.py in
docs/prompt-audit-2026-09-29/tools.md.
"""
import collections
import json
import pathlib
import re
import sys

FLOW_FILE = re.compile(r"((?:skills|rules|commands-claude)/[\w./-]+\.mdc?)")
BASH_READ = re.compile(r"\b(cat|sed|head|tail|awk)\b")
TOP = 25


def text_of(content):
    """The text a tool result carried: a string, or the text parts of a content list."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return "".join(c.get("text", "") for c in content if isinstance(c, dict))
    return ""


def loaded_name(tool_use):
    """The flow file a tool use reads, or None when it reads none."""
    name, inp = tool_use.get("name"), tool_use.get("input") or {}
    if name == "Read":
        m = FLOW_FILE.search(inp.get("file_path", ""))
        if not m:
            return None
        return m.group(1) + (f" @{inp['offset']}" if inp.get("offset") else "")
    if name == "Skill":
        return f"<Skill {inp.get('skill')}>"
    if name == "Bash":
        command = inp.get("command", "")
        hits = FLOW_FILE.findall(command) if BASH_READ.search(command) else []
        return "bash:" + ",".join(sorted(set(hits))) if hits else None
    return None


def scan(path):
    """(loads, turns): each load as (name, bytes, turn), and the session's API turn count."""
    pending, loads, turn = {}, [], 0
    with open(path, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            try:
                event = json.loads(line)
            except ValueError:
                continue
            msg = event.get("message") or {}
            if event.get("type") == "assistant" and msg.get("usage"):
                turn += 1
            content = msg.get("content")
            for c in content if isinstance(content, list) else []:
                if not isinstance(c, dict):
                    continue
                if c.get("type") == "tool_use":
                    name = loaded_name(c)
                    if name:
                        pending[c.get("id")] = name
                elif c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                    size = len(text_of(c.get("content")).encode())
                    loads.append((pending.pop(c["tool_use_id"]), size, turn))
    return loads, turn


def transcripts(args):
    if not args:
        return sorted(pathlib.Path.home().glob(".claude/projects/*/*.jsonl"))
    out = []
    for a in args:
        p = pathlib.Path(a)
        out += sorted(p.glob("*.jsonl")) if p.is_dir() else [p]
    return out


def report(path):
    try:
        loads, turns = scan(path)
    except OSError as err:
        print(f"== {path}: unreadable ({err})")
        return
    if not loads:
        return
    total = sum(b for _, b, _ in loads)
    weight = sum(b // 4 * (turns - t) for _, b, t in loads)
    print(f"== {path.name}: {len(loads)} loads, {total} bytes (~{total // 4} tok), "
          f"{turns} API turns, ~{weight / 1e6:.1f}M token-turns re-read")
    print(f"   {'bytes':>7}  {'turn':>4}  file")
    for name, size, t in sorted(loads, key=lambda x: -(x[1] * (turns - x[2])))[:TOP]:
        print(f"   {size:7d}  {t:4d}  {name}")
    seen = collections.Counter(n.split(" @")[0] for n, _, _ in loads)
    rereads = {n: k for n, k in seen.items() if k > 1}
    if rereads:
        print(f"   re-read: {rereads}")


def main(args):
    for path in transcripts(args):
        report(path)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
