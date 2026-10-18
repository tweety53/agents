#!/usr/bin/env python3
"""PreToolUse hook: block subagent dispatches that omit the agent-baseline pointer.

Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT
inherited by dispatched agents, nor by agents those agents dispatch. The prompt is the
only channel, so every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md
plus the sentence that makes the receiving agent pass it on in turn.

The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the
line itself. It denies the call instead and returns the exact text to add; the model then
re-dispatches with it.

It also pins the agent type and model of the /flow roles whose contracts fix them
(PINNED_ROLES): a dispatch whose prompt carries a role's marker on any other
subagent_type/model is denied with the values it must carry. Prose in
skills/flow/visual-verify.md did not stop a hand-written verifier dispatch on flow-high.

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

# (prompt marker, subagent_type, model, the contract that fixes the pair). First match
# wins: the tooling analyst's prompt also names visual-verify-verifier.md, so its own
# marker, its report file, is checked first.
PINNED_ROLES = (
    ("tooling-analysis-", "flow-medium", "opus", "skills/flow/visual-verify-tooling-analysis.md"),
    ("visual-verify-verifier.md", "flow-low", "opus", "skills/flow/visual-verify.md"),
)

# Harness mapping (skills/flow-contracts/model-policy.md): on zcode every pinned pair
# is dispatched as flow-high / glm-5.3-flash. The zcode registration runs the hook from
# ~/.zcode/hooks, which is how the hook tells the harness.
ZCODE_PAIR = ("flow-high", "glm-5.3-flash")

# flow-high is dispatched in exactly two cases (skills/flow/implement.md): the no-progress
# escalation -- marked by its opening line, or a self-review loop key whose rows
# scripts/check-loop-dispatch.sh pins -- and a panel dispatch the operator approved at high. Anything else on flow-high was
# hand-picked: a parent ran scoped in-run fixes on flow-high where the decision said inline.
FLOW_HIGH_MARKERS = ("No-progress escalation", "Key: self-review-", "Operator-approved high")


def flow_high_reason(prompt: str, tool_input: dict) -> "str | None":
    if tool_input.get("subagent_type") != "flow-high" or "/.zcode/" in __file__:
        return None
    if any(marker in prompt for marker in FLOW_HIGH_MARKERS):
        return None
    return (
        "Dispatch blocked: flow-high is reserved for the no-progress escalation and a panel "
        "dispatch the operator approved at high (skills/flow/implement.md).\n\n"
        "Read the recorded decision (`flow record decisions -change <name>`) and dispatch on "
        "the pair it names -- or, when its execution is `inline`, make the change yourself "
        "with no dispatch. A no-progress escalation opens its prompt with the line "
        "`No-progress escalation`; an approved panel dispatch with `Operator-approved high`."
    )


def pinned_role_reason(prompt: str, tool_input: dict) -> "str | None":
    for marker, subagent_type, model, contract in PINNED_ROLES:
        if marker not in prompt:
            continue
        if "/.zcode/" in __file__:
            subagent_type, model = ZCODE_PAIR
        got = (tool_input.get("subagent_type"), tool_input.get("model"))
        if got == (subagent_type, model):
            return None
        return (
            f"Dispatch blocked: a prompt carrying `{marker}` is a role {contract} pins.\n\n"
            f"Got subagent_type={got[0]!r}, model={got[1]!r}.\n"
            f"Re-send this call with subagent_type: \"{subagent_type}\" and "
            f"model: \"{model}\", nothing else changed. No rule or instruction overrides it."
        )
    return None


def deny(reason: str) -> None:
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
    if not all(marker in lowered for marker in REQUIRED_MARKERS):
        deny(DENY_REASON)
        return 0

    reason = pinned_role_reason(prompt, tool_input) or flow_high_reason(prompt, tool_input)
    if reason:
        deny(reason)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block work on a hook fault
        sys.exit(0)
