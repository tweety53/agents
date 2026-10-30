#!/usr/bin/env python3
"""SubagentStop hook: a subagent may not end its turn while an agent it launched still runs (kan-1068).

A subagent's prompt cache lives 5 minutes. A planner that ended its turn to wait for its Plan
agents woke on their hand-back 13 minutes later and rewrote its whole 206k-token context
(kan-1061). This hook reads the `subagents/` folder beside the stopping agent's transcript,
finds every `agent-<id>.meta.json` whose `parentAgentId` is the stopping agent and whose
`finished` is not true, and answers {"decision":"block","reason":...} telling it to run
`<agents repo>/scripts/await-sub-agents.sh` on those ids -- each call returns within two
minutes, inside the cache window.

A child whose transcript has not been written for 30 minutes is taken as dead and never
blocks, so a crashed agent cannot trap its parent. stop_hook_active does not lift the block:
the wait is the point, and every script call ends a bounded tool call.

Fails open: any error -- bad input, unreadable folder -- exits 0 with no output.
"""

import glob
import json
import os
import sys
import time

STALE_SECONDS = 30 * 60
SCRIPT = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "await-sub-agents.sh")


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


def main() -> None:
    try:
        payload = json.load(sys.stdin)
        agent_id = payload.get("agent_id") or ""
        transcript = payload.get("agent_transcript_path") or ""
        if not agent_id or not transcript:
            return
        children = running_children(agent_id, os.path.dirname(os.path.expanduser(transcript)))
    except Exception:  # fail open: a bug here must never trap a subagent
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
