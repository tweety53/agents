#!/usr/bin/env python3
"""Stop hook: refuse a handoff that drops its never-omitted lines.

It reads the turn's last assistant text: a Finished handoff (a `## Finished` line,
skills/flow/cleanup.md) or a /flow-fast landing report (recognised by carrying either line) that
lacks `**Learning:**` or `**Prompt tuning:**` blocks with a reason naming the missing line(s) -- in
any session, with no store lookup. With stop_hook_active true this check is skipped, so a turn that
already got that block once is never trapped. Only text after the operator's last (non-isMeta)
message counts, so a turn ending on tool calls alone is never judged on an earlier turn's text; and
a label counts only at the start of a line, so prose quoting `**Learning:**` mid-sentence is not a
handoff. A handoff split across text blocks around a tool call is judged on its last text block
alone.

Fails open: any error -- bad input, unreadable transcript -- exits 0 with no output.
A bug here must never trap a session's turn.
"""

import json
import os
import re
import sys

TRANSCRIPT_TAIL_BYTES = 1 << 20
HANDOFF_LINES = ("**Learning:**", "**Prompt tuning:**")


def _texts(content) -> list:
    if isinstance(content, str):
        return [content]
    if isinstance(content, list):
        return [b.get("text", "") for b in content if isinstance(b, dict) and b.get("type") == "text"]
    return []


def transcript_entries(path) -> list:
    if not isinstance(path, str) or not os.path.isfile(path):
        return []
    with open(path, "rb") as fh:
        fh.seek(max(0, os.path.getsize(path) - TRANSCRIPT_TAIL_BYTES))
        return fh.read().decode("utf-8", "replace").splitlines()


def last_text(lines: list) -> str:
    """The turn's last assistant text: only text after the operator's last message counts."""
    last = ""
    for line in lines:
        try:
            entry = json.loads(line)
        except ValueError:
            continue
        if not isinstance(entry, dict):
            continue
        if entry.get("type") == "user" and not entry.get("isMeta"):
            if _texts((entry.get("message") or {}).get("content")):  # tool_result-only is not the operator
                last = ""
        elif entry.get("type") == "assistant":
            texts = _texts((entry.get("message") or {}).get("content"))
            if any(t.strip() for t in texts):
                last = "\n".join(texts)
    return last


def handoff_missing(last: str) -> list:
    """The never-omitted handoff lines the turn's last assistant text drops, if it is a handoff."""
    present = [h for h in HANDOFF_LINES if re.search(r"^" + re.escape(h), last, re.MULTILINE)]
    if not present and not re.search(r"^## Finished\s*$", last, re.MULTILINE):
        return []
    return [h for h in HANDOFF_LINES if h not in present]


def main() -> int:
    payload = json.load(sys.stdin)
    if not isinstance(payload, dict) or payload.get("stop_hook_active"):
        return 0
    missing = handoff_missing(last_text(transcript_entries(payload.get("transcript_path"))))
    if missing:
        json.dump(
            {
                "decision": "block",
                "reason": (
                    f"flow-stop-guard: the handoff lacks its {' and '.join(missing)} line(s), which are "
                    f"never omitted (skills/flow/cleanup.md). Reprint the handoff with them."
                ),
            },
            sys.stdout,
        )
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never trap a turn on a hook fault
        sys.exit(0)
