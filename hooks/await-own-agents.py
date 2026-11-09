#!/usr/bin/env python3
"""SubagentStop hook: a subagent may not end its turn while an agent it launched still runs (kan-1068).

A subagent's prompt cache lives 5 minutes. A planner that ended its turn to wait for its Plan
agents woke on their hand-back 13 minutes later and rewrote its whole 206k-token context
(kan-1061). This hook reads the `subagents/` folder beside the stopping agent's transcript,
finds every `agent-<id>.meta.json` whose `parentAgentId` is the stopping agent and whose
`finished` is not true, and answers {"decision":"block","reason":...} telling it to run
`flow-guard await-sub-agents` on those ids -- each call returns within two
minutes, inside the cache window.

A child whose transcript has not been written for 30 minutes is taken as dead and never
blocks, so a crashed agent cannot trap its parent. stop_hook_active does not lift the block:
the wait is the point, and every script call ends a bounded tool call.

A subagent may not stop while a Bash command it moved to the background still runs either: the
tool_result of one of its own Bash tool_use calls that opens with "Command running in background
with ID: <id>", or has a line opening with "Command [did not complete within its Ns timeout and ]was
moved to the background (ID: <id>)" -- partial output may precede it --, launches <id>, and a `<task-id><id></task-id>` task-notification anywhere in
its transcript ends it. Pairing by tool_use_id and anchoring the sentence keep a launch the agent
merely read -- a Read of a transcript, a diff of a test, a grep of a JSONL line -- from trapping it. A child that ended with such a command out
woke again after its parent had moved on -- kan-1085's fixer left a build running in the background
and its parent waited 44 minutes on it.

Fails open: any error -- bad input, unreadable folder -- exits 0 with no output.
"""

import glob
import json
import os
import re
import sys
import time

STALE_SECONDS = 30 * 60
SCRIPT = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "flow-guard") + " await-sub-agents"


def running_children(agent_id: str, folder: str) -> "list[str]":
    out = []
    for meta in sorted(glob.glob(os.path.join(folder, "agent-*.meta.json"))):
        try:
            with open(meta) as f:
                m = json.load(f)
        except (OSError, ValueError):
            continue
        if m.get("parentAgentId") != agent_id or m.get("finished") is True:
            continue
        child = os.path.basename(meta)[len("agent-"):-len(".meta.json")]
        transcript = os.path.join(folder, f"agent-{child}.jsonl")
        if os.path.exists(transcript) and time.time() - os.path.getmtime(transcript) > STALE_SECONDS:
            continue
        out.append(child)
    return out


RUN_IN_BACKGROUND = re.compile(r"Command running in background with ID: ([A-Za-z0-9]+)")
MOVED_TO_BACKGROUND = re.compile(
    r"^Command (?:did not complete within its \d+s timeout and )?was moved to the background \(ID: ([A-Za-z0-9]+)\)",
    re.MULTILINE,
)
TASK_ID = re.compile(r"<task-id>([A-Za-z0-9]+)</task-id>")


def running_background(transcript: str) -> "list[str]":
    launched, done, bash_calls = [], set(), set()
    with open(transcript, encoding="utf-8", errors="replace") as f:
        for line in f:
            done.update(TASK_ID.findall(line))
            try:
                entry = json.loads(line)
            except ValueError:
                continue
            # A writer may escape `<` as \u003c; the decoded entry carries the tag either way.
            done.update(TASK_ID.findall(json.dumps(entry, ensure_ascii=False)))
            kind = entry.get("type") if isinstance(entry, dict) else None
            message = entry.get("message") if kind in ("user", "assistant") else None
            content = message.get("content") if isinstance(message, dict) else None
            if not isinstance(content, list):
                continue
            for block in content:
                if not isinstance(block, dict):
                    continue
                if kind == "assistant" and block.get("type") == "tool_use" and block.get("name") == "Bash":
                    bash_calls.add(block.get("id"))
                if kind != "user" or block.get("type") != "tool_result" or block.get("tool_use_id") not in bash_calls:
                    continue
                text = block.get("content")
                if isinstance(text, list):
                    text = " ".join(b.get("text", "") for b in text if isinstance(b, dict))
                if not isinstance(text, str):
                    continue
                started = RUN_IN_BACKGROUND.match(text) or MOVED_TO_BACKGROUND.search(text)
                if started and started.group(1) not in launched:
                    launched.append(started.group(1))
    return [i for i in launched if i not in done]


def main() -> None:
    try:
        payload = json.load(sys.stdin)
        agent_id = payload.get("agent_id") or ""
        transcript = payload.get("agent_transcript_path") or ""
        if not agent_id or not transcript:
            return
        transcript = os.path.expanduser(transcript)
        children = running_children(agent_id, os.path.dirname(transcript))
        background = running_background(transcript) if os.path.exists(transcript) else []
    except Exception:  # fail open: a bug here must never trap a subagent
        return
    if background:
        print(json.dumps({
            "decision": "block",
            "reason": (
                f"Background commands you launched are still running ({' '.join(background)}). "
                "FOREGROUND BUILDS: wait for each one's task-notification, or stop it with TaskStop, "
                "before you end your turn -- a child that ends with work still out wakes again after "
                "its parent has moved on."
            ),
        }))
        return
    if not children:
        return
    ids = " ".join(children)
    print(json.dumps({
        "decision": "block",
        "reason": (
            f"Agents you launched are still running ({ids}). Ending your turn now loses your "
            f"5-minute prompt cache before their hand-back arrives. Run `{SCRIPT} {ids}` -- it "
            "returns within two minutes, exit 0 once all finished, exit 1 naming those still "
            "running -- and call it again on exit 1; read each hand-back as it arrives."
        ),
    }))


if __name__ == "__main__":
    main()
