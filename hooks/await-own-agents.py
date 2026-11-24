#!/usr/bin/env python3
"""SubagentStop hook: a subagent may not end its turn while an agent it launched still runs (kan-1068).

A subagent's prompt cache lives 5 minutes. A planner that ended its turn to wait for its Plan
agents woke on their hand-back 13 minutes later and rewrote its whole 206k-token context
(kan-1061). This hook reads the `subagents/` folder beside the stopping agent's transcript,
finds every `agent-<id>.meta.json` whose `parentAgentId` is the stopping agent and whose
`finished` is not true, and WAITS for them itself, polling every few seconds for up to
AWAIT_CHILDREN_SECONDS (240, inside the cache window; the default hook timeout is 10 minutes). When
they all finish it answers {"decision":"block","reason":...} once, telling the agent to read their
hand-backs; when some still run at the cap it blocks with "still running, stop again" -- each such
re-stop is one cheap warm-cache turn, so the cache stays warm however long a child takes.

A child whose transcript has not been written for 30 minutes is taken as dead and never
blocks, so a crashed agent cannot trap its parent. stop_hook_active does not lift the block:
the wait is the point, and every wait is bounded.

A subagent may not stop while a Bash command it moved to the background still runs either: the
tool_result of one of its own Bash tool_use calls that opens with "Command running in background
with ID: <id>", or has a line opening with "Command [did not complete within its Ns timeout and ]was
moved to the background (ID: <id>)" -- partial output may precede it --, launches <id>, and a `<task-id><id></task-id>` task-notification anywhere in
its transcript ends it, as does a TaskStop result "Successfully stopped task: <id>" paired with one of
the agent's own TaskStop calls. Pairing by tool_use_id and anchoring the sentence keep a launch the agent
merely read -- a Read of a transcript, a diff of a test, a grep of a JSONL line -- from trapping it. A child that ended with such a command out
woke again after its parent had moved on -- kan-1085's fixer left a build running in the background
and its parent waited 44 minutes on it.

A stop that carries a final message (the payload's `last_assistant_message`) is a finished agent: its report
is written, so a background command it still holds is no work anyone waits for. The hook kills it --
every process holding that command's `<id>.output` file open, found with `lsof`, so only that agent's own
task is touched, never the parent's or a sibling's -- and lets the stop through instead of blocking on a
notification that may never come (kan-1092 task 19: a hung build left the agent "waiting on 1 shell" until
its parent stopped it by hand). A stop with no final message blocks only on a background command that is
still alive: one with a process holding its `<id>.output` file open (the same `lsof` lookup). A command whose
output file is gone or has no holder has ended -- a shell that died, or finished with its notification lost --
and never blocks, so no notification is needed. When `lsof` is missing or fails, a command counts as alive
only if its output file exists and was modified within the last 120 seconds.

A stop it lets through -- no child and no background command left -- closes the agent's flow
dispatch row: `flow-guard end-dispatch-row` reads the transcript's `Key:` line and final message
and ends the row the begin hook opened, so the parent ends none itself. Every stop it lets through or blocks then runs `flow-guard
end-stale-dispatch-rows <cwd>`, which ends the change's open rows whose agent finished (meta
`finished`) and went quiet 10 minutes ago -- the rows of agents whose own stop never ended them.

Fails open: any error -- bad input, unreadable folder -- exits 0 with no output.
"""

import glob
import json
import os
import re
import subprocess
import sys
import time

STALE_SECONDS = 30 * 60
FLOW_GUARD = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "flow-guard")
WAIT_SECONDS = int(os.environ.get("AWAIT_CHILDREN_SECONDS", "240"))
POLL_SECONDS = 3


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
STOPPED = re.compile(r"Successfully stopped task: ([A-Za-z0-9]+)")


LSOF_FALLBACK_SECONDS = 120


def output_paths(text: str, task: str) -> "set[str]":
    return set(re.findall(r"(/[^\s\"']*/" + re.escape(task) + r"\.output)", text))


def holders(path: str) -> "list[int] | None":
    """Pids holding `path` open, or None when lsof is missing or fails."""
    try:
        r = subprocess.run(["lsof", "-t", "--", path], capture_output=True, text=True, timeout=10)
    except Exception:
        return None
    if r.returncode > 1:
        return None
    return [int(p) for p in r.stdout.split() if p.isdigit() and int(p) != os.getpid()]


def alive(text: str, task: str) -> bool:
    """A background command lives while a process holds its output file; without lsof, while the file is fresh."""
    for path in output_paths(text, task):
        if not os.path.exists(path):
            continue
        pids = holders(path)
        if pids is None:
            if time.time() - os.path.getmtime(path) <= LSOF_FALLBACK_SECONDS:
                return True
        elif pids:
            return True
    return False


def kill_background(transcript: str, ids: "list[str]") -> None:
    """SIGTERM every process holding an `<id>.output` file of the given background commands open."""
    with open(transcript, encoding="utf-8", errors="replace") as f:
        text = f.read()
    for task in ids:
        for path in output_paths(text, task):
            if os.path.exists(path):
                for pid in holders(path) or []:
                    try:
                        os.kill(pid, 15)
                    except OSError:
                        pass


def running_background(transcript: str) -> "list[str]":
    launched, done, bash_calls, stop_calls = [], set(), set(), set()
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
                if kind == "assistant" and block.get("type") == "tool_use" and block.get("name") == "TaskStop":
                    stop_calls.add(block.get("id"))
                use_id = block.get("tool_use_id")
                if kind != "user" or block.get("type") != "tool_result" or use_id not in bash_calls | stop_calls:
                    continue
                text = block.get("content")
                if isinstance(text, list):
                    text = " ".join(b.get("text", "") for b in text if isinstance(b, dict))
                if not isinstance(text, str):
                    continue
                if use_id in stop_calls:
                    done.update(STOPPED.findall(text))
                    continue
                started = RUN_IN_BACKGROUND.match(text) or MOVED_TO_BACKGROUND.search(text)
                if started and started.group(1) not in launched:
                    launched.append(started.group(1))
    return [i for i in launched if i not in done]


def end_row(transcript: str, payload: dict) -> None:
    """Close the stopping agent's flow dispatch row; best-effort, output discarded."""
    session_id, cwd = payload.get("session_id") or "", payload.get("cwd") or ""
    if not session_id or not cwd:
        return
    try:
        subprocess.run([FLOW_GUARD, "end-dispatch-row", transcript, session_id, cwd,
                        payload.get("last_assistant_message") or ""],
                       capture_output=True, timeout=30, cwd=cwd)
    except Exception:
        pass
    # A stop is the moment to close rows whose own stop never did (flow-guard sweep-dispatch-rows).
    try:
        subprocess.run([FLOW_GUARD, "sweep-dispatch-rows", os.path.dirname(transcript), session_id, cwd],
                       capture_output=True, timeout=60, cwd=cwd)
    except Exception:
        pass


def sweep_stale(payload: dict) -> None:
    """End the change's finished-agent rows (flow-guard end-stale-dispatch-rows); best-effort, never blocking."""
    cwd = payload.get("cwd") or ""
    if not cwd:
        return
    try:
        subprocess.run([FLOW_GUARD, "end-stale-dispatch-rows", cwd], capture_output=True, timeout=60, cwd=cwd)
    except Exception:
        pass


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
        if background:
            with open(transcript, encoding="utf-8", errors="replace") as f:
                text = f.read()
            background = [t for t in background if alive(text, t)]
    except Exception:  # fail open: a bug here must never trap a subagent
        return
    sweep_stale(payload)
    if background and (payload.get("last_assistant_message") or "").strip():
        try:
            kill_background(transcript, background)
        except Exception:
            pass
        background = []
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
        end_row(transcript, payload)
        return
    ids = " ".join(children)
    deadline = time.time() + WAIT_SECONDS
    while children and time.time() < deadline:
        time.sleep(POLL_SECONDS)
        try:
            children = running_children(agent_id, os.path.dirname(transcript))
        except Exception:  # fail open
            return
    if children:
        reason = (f"Agents you launched are still running ({' '.join(children)}). Stop again to keep waiting "
                  "for them -- ending your turn on your own would lose your 5-minute prompt cache.")
    else:
        reason = (f"Agents you launched have finished ({ids}). Read each hand-back and continue; "
                  "say so when one returned nothing.")
    print(json.dumps({"decision": "block", "reason": reason}))


if __name__ == "__main__":
    main()
