#!/usr/bin/env python3
"""PreToolUse hook: block subagent dispatches that omit the agent-baseline pointer.

Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT
inherited by dispatched agents, nor by agents those agents dispatch. The prompt is the
only channel, so every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md
plus the sentence that makes the receiving agent pass it on in turn.

The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the
line itself. It denies the call instead and returns the exact text to add; the model then
re-dispatches with it.

Fails open: any unexpected input, parse error, or internal fault exits 0 and lets the
dispatch through. A guardrail on a convention must never be the reason work stops.
"""

import json
import sys

# The tool that spawns a subagent. Current builds name it "Agent"; older ones "Task".
DISPATCH_TOOLS = {"Agent", "Task"}

# A fork inherits the parent's full context, CLAUDE.md included, so it needs no pointer.
EXEMPT_SUBAGENT_TYPES = {"fork"}

# Both must appear in the prompt: the file being pointed at, and the propagation clause
# that makes the receiving agent a carrier. Either alone is not the rule.
REQUIRED_MARKERS = ("agent-baseline.md", "verbatim")

REQUIRED_TEXT = (
    "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for "
    "this whole task. Include this instruction verbatim in any prompt you write for "
    "another agent."
)

DENY_REASON = (
    "Dispatch blocked: the subagent prompt does not carry the agent-baseline pointer.\n\n"
    "A dispatched agent starts from an empty context. It has never read ~/.claude/CLAUDE.md, "
    "and neither will anything it dispatches. Without the pointer it operates under no rules "
    "at all — production access, lint suppression, pushes to main, dependency versions.\n\n"
    "Re-send this call with the following two sentences included verbatim in `prompt`:\n\n"
    f"{REQUIRED_TEXT}\n\n"
    "The second sentence is not optional: it is what carries the rules past depth 1.\n"
    "Exempt: subagent_type \"fork\", which inherits this session's context already."
)


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    if not isinstance(payload, dict):
        return 0

    if payload.get("tool_name") not in DISPATCH_TOOLS:
        return 0

    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        return 0

    if tool_input.get("subagent_type") in EXEMPT_SUBAGENT_TYPES:
        return 0

    prompt = tool_input.get("prompt")
    if not isinstance(prompt, str):
        return 0

    lowered = prompt.lower()
    if all(marker in lowered for marker in REQUIRED_MARKERS):
        return 0

    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": DENY_REASON,
            }
        },
        sys.stdout,
    )
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # noqa: BLE001 - fail open, never block work on a hook fault
        sys.exit(0)
