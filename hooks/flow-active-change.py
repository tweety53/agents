#!/usr/bin/env python3
"""UserPromptSubmit hook: name the session's active /flow change on every plain prompt.

A problem report typed as a plain message, with no /flow prefix, needs to know which
change it belongs to before the SKILL.md rule can turn it into a fix run. The change name
is not carried in session state a hook can read directly, but every /flow run marks its
own stages with `flow stage begin -command '/flow' ... -session-token mf-<token> <name>`,
and that command is recorded verbatim in the session's transcript. This hook finds the
last such mark and injects the resolved change name as additional context.

Fails open, like enforce-agent-baseline.py: any unexpected input, missing transcript, or
internal fault exits 0 with no output. A hook that supplies context must never be the
reason a prompt is refused.
"""

import json
import os
import re
import sys
from typing import Optional

MARK_RE = re.compile(
    r"flow stage begin -command '/flow' [^;|&\n]*?-session-token mf-[A-Za-z0-9_-]+ ([A-Za-z0-9._-]+)"
)


def find_last_change(transcript_path: str) -> Optional[str]:
    with open(transcript_path, "rb") as f:
        text = f.read().decode("utf-8", errors="replace")
    matches = MARK_RE.findall(text)
    return matches[-1] if matches else None


def resolve_transcript_path(payload: dict) -> Optional[str]:
    transcript_path = payload.get("transcript_path")
    if isinstance(transcript_path, str) and transcript_path:
        return transcript_path

    session_id = payload.get("session_id") or os.environ.get("CLAUDE_SESSION_ID")
    if not session_id:
        return None

    root = os.environ.get("FLOW_ZCODE_ROLLOUTS_DIR") or os.path.expanduser(
        "~/.zcode/cli/rollout"
    )
    return os.path.join(root, f"model-io-sess_{session_id}.jsonl")


def main() -> int:
    payload = json.load(sys.stdin)

    if not isinstance(payload, dict):
        return 0

    prompt = payload.get("prompt")
    if isinstance(prompt, str) and prompt.startswith("/"):
        return 0

    transcript_path = resolve_transcript_path(payload)
    if not transcript_path:
        return 0

    name = find_last_change(transcript_path)
    if not name:
        return 0

    line = (
        f"flow: this session's last /flow run marked change {name}. A plain message "
        "reporting a problem or asking for a change to it is a fix run of "
        f"{name} — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md)."
    )
    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": line,
            }
        },
        sys.stdout,
    )
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block a prompt on a hook fault
        sys.exit(0)
