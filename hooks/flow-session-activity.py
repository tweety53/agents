#!/usr/bin/env python3
"""Claude Code hook: tell flowd what this session is doing now.

The Live state board's Running now cards say whether a change's session is working or
waiting on the operator. This hook reports each transition as one
PUT /api/v1/sessions/<session_id>/activity {"state", "detail"}:

  PreToolUse on AskUserQuestion  -> question, the first question's text
  Notification naming permission -> permission, the notification's message
  PreToolUse, PostToolUse        -> working, the tool and its short label
  UserPromptSubmit               -> working
  Stop                           -> idle

Any other event sends nothing. detail is cut to 500 bytes, the most flowd accepts.

Fails open, like flow-active-change.py: malformed input, a payload without a session id
or an unreachable store exits 0 with no output. A hook that reports state must never be
the reason a tool call or a prompt is held up.
"""

import json
import os
import sys
from typing import Optional, Tuple
from urllib import parse, request

DEFAULT_STORE_ADDR = "http://127.0.0.1:4173"
STORE_TIMEOUT_SECONDS = 2
MAX_DETAIL_BYTES = 500


def activity(payload: dict) -> Optional[Tuple[str, str]]:
    event = payload.get("hook_event_name")
    # A subagent's tool events fire under the parent's session_id and carry
    # agent_id (https://code.claude.com/docs/en/hooks); they must not overwrite
    # the main thread's state, e.g. turn an open question back to working
    # (kan-1048 self-review, 1-J). Its Notification still reaches the board.
    if event in ("PreToolUse", "PostToolUse") and payload.get("agent_id"):
        return None
    if event == "PreToolUse" and payload.get("tool_name") == "AskUserQuestion":
        questions = (payload.get("tool_input") or {}).get("questions") or [{}]
        return "question", str(questions[0].get("question") or "")
    if event == "Notification":
        message = str(payload.get("message") or "")
        if "permission" in message.lower():
            return "permission", message
        return None
    # Every tool call reports working, before it runs and after: a Stop the
    # stop guard refuses leaves "idle" behind, and a long call (a poll, a
    # build) must not read as a turn that ended while it runs.
    if event in ("PreToolUse", "PostToolUse"):
        return "working", action(payload)
    if event == "UserPromptSubmit":
        return "working", ""
    if event == "Stop":
        return "idle", ""
    return None


def action(payload: dict) -> str:
    """What the main thread just did, in a few words, for the board: the tool and
    its own short label -- a Bash call's description (else its first command
    line), an edited or read file's name, a dispatched agent's or skill's name."""
    tool = str(payload.get("tool_name") or "")
    inp = payload.get("tool_input") or {}
    if not isinstance(inp, dict):
        return tool
    if tool == "Bash":
        label = str(inp.get("description") or "")
        if not label:
            lines = str(inp.get("command") or "").splitlines()
            label = lines[0] if lines else ""
        return f"Bash: {label}" if label else tool
    if tool in ("Edit", "Write", "Read", "NotebookEdit"):
        path = str(inp.get("file_path") or inp.get("notebook_path") or "")
        return f"{tool} {os.path.basename(path)}" if path else tool
    if tool in ("Agent", "Task"):
        label = str(inp.get("description") or inp.get("subagent_type") or "")
        return f"Agent: {label}" if label else tool
    if tool == "Skill":
        return f"Skill {inp.get('skill')}" if inp.get("skill") else tool
    return tool


def truncate(text: str) -> str:
    # Postgres text refuses U+0000, so a NUL would make the store reject the row.
    text = text.replace("\x00", "")
    return text.encode("utf-8")[:MAX_DETAIL_BYTES].decode("utf-8", errors="ignore")


def main() -> None:
    try:
        payload = json.load(sys.stdin)
        session_id = payload.get("session_id")
        if not isinstance(session_id, str) or not session_id:
            return
        found = activity(payload)
        if found is None:
            return
        state, detail = found
        addr = os.environ.get("FLOW_ADDR") or DEFAULT_STORE_ADDR
        body = json.dumps({"state": state, "detail": truncate(detail)}).encode("utf-8")
        req = request.Request(
            f"{addr}/api/v1/sessions/{parse.quote(session_id, safe='')}/activity",
            data=body,
            method="PUT",
            headers={"Content-Type": "application/json"},
        )
        with request.urlopen(req, timeout=STORE_TIMEOUT_SECONDS):
            pass
    except Exception:
        return


if __name__ == "__main__":
    main()
    sys.exit(0)
