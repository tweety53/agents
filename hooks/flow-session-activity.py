#!/usr/bin/env python3
"""Claude Code hook: tell flowd what this session is doing now.

The Live state board's Running now cards say whether a change's session is working or
waiting on the operator. This hook reports each transition as one
PUT /api/v1/sessions/<session_id>/activity {"state", "detail"}:

  PreToolUse on AskUserQuestion  -> question, the first question's text
  Notification naming permission -> permission, the notification's message
  UserPromptSubmit, PostToolUse  -> working
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
    if event == "PreToolUse" and payload.get("tool_name") == "AskUserQuestion":
        questions = (payload.get("tool_input") or {}).get("questions") or [{}]
        return "question", str(questions[0].get("question") or "")
    if event == "Notification":
        message = str(payload.get("message") or "")
        if "permission" in message.lower():
            return "permission", message
        return None
    if event in ("UserPromptSubmit", "PostToolUse"):
        return "working", ""
    if event == "Stop":
        return "idle", ""
    return None


def truncate(text: str) -> str:
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
