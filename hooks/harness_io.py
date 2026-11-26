"""Shared stdin/stdout helpers for dual-harness hooks (Claude Code and Antigravity).

Claude Code sends snake_case payloads (`tool_name`, `tool_input`, `session_id`,
`hook_event_name`) and expects empty stdout on allow or `hookSpecificOutput` on
deny/context injection.

Antigravity sends camelCase protojson payloads (`toolCall`, `conversationId`,
`hookEventName`, `lastUserInput`, `transcriptPath`) and expects a JSON object on
stdout (`{"decision": "allow"|"deny", ...}` for PreToolUse, `{"injectSteps": ...}`
or `{}` for PreInvocation, `{}` for PostToolUse and Stop).

Standard library only — hooks run under `/usr/bin/python3` with no third-party deps.
"""

import json
import sys
from typing import Tuple


def is_antigravity(payload: dict) -> bool:
    """True when payload uses Antigravity's camelCase hook schema."""
    return (
        "hookEventName" in payload
        or "toolCall" in payload
        or "conversationId" in payload
    )


def get_event_name(payload: dict) -> str:
    val = payload.get("hookEventName") or payload.get("hook_event_name")
    return val if isinstance(val, str) else ""


def get_session_id(payload: dict) -> str:
    val = payload.get("conversationId") or payload.get("session_id")
    return val if isinstance(val, str) else ""


def get_transcript_path(payload: dict) -> str:
    val = payload.get("transcriptPath") or payload.get("transcript_path")
    return val if isinstance(val, str) else ""


def get_tool_call(payload: dict) -> Tuple[str, dict, str]:
    """Return (tool_name, tool_args, tool_id) across Claude Code and Antigravity."""
    if is_antigravity(payload):
        tc = payload.get("toolCall")
        if not isinstance(tc, dict):
            return "", {}, ""
        name = tc.get("name")
        args = tc.get("args")
        tid = tc.get("id")
        return (
            name if isinstance(name, str) else "",
            args if isinstance(args, dict) else {},
            tid if isinstance(tid, str) else "",
        )
    name = payload.get("tool_name")
    args = payload.get("tool_input")
    tid = payload.get("tool_use_id")
    return (
        name if isinstance(name, str) else "",
        args if isinstance(args, dict) else {},
        tid if isinstance(tid, str) else "",
    )


def emit_pre_tool_deny(payload: dict, reason: str) -> None:
    if is_antigravity(payload):
        json.dump({"decision": "deny", "reason": reason}, sys.stdout)
    else:
        json.dump(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": reason,
                }
            },
            sys.stdout,
        )
    sys.stdout.flush()


def emit_pre_tool_allow(payload: dict) -> None:
    if is_antigravity(payload):
        json.dump({"decision": "allow"}, sys.stdout)
        sys.stdout.flush()


def emit_empty_if_antigravity(payload: dict) -> None:
    if is_antigravity(payload):
        json.dump({}, sys.stdout)
        sys.stdout.flush()
