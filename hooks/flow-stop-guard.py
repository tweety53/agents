#!/usr/bin/env python3
"""Stop hook: refuse a /flow turn end mid-phase (kan-1024).

`skills/flow-contracts/pipeline.md` says a turn ends only at a gate, a block, or the run's
final handoff; prose alone failed five times (kan-1015, kan-1017, kan-1018, kan-1033,
kan-1024). This hook reads this session's last /flow stage run from the flow store and the
change's record, and answers {"decision":"block","reason":...} when the record is
STARTED and:

- the last /flow stage run is still open (endedAt null) -- the reason names that stage; or
- every task of the plan is ticked (ticked == total > 0) --
  implementation is done and flow.write-in-progress is not yet written.

It allows (exit 0, no output) at an AskUserQuestion or a turn whose last text opens a line with
`## Question` (the stop **What still stops** names), after an operator message saying
"stop", while a background agent this transcript launched has not yet notified its end (the
wait on one-shot subagents, skills/flow-contracts/pipeline.md), while a Bash command it moved to the
background has not either (await-own-agents.py's running_background; the wait
skills/flow/verify-and-handoff.md prescribes for a long check), at IN_PROGRESS / FINISHED, at STARTED with no tasks or none ticked and no open
stage, and when the session has no /flow run. A repeated end at the same open stage is refused
every time: stop_hook_active never lets one through (kan-1099, where a second status report
offering to continue passed a fingerprint that had not moved).

None of those allowances covers a turn whose last tool result is a refusal -- a flow guard's
REFUSED, a hook's "Dispatch blocked", an auto-mode classifier denial -- while the record is
STARTED: a refusal's cure is in the run's reach (**What still stops**,
skills/flow-contracts/operator-prompts-auto-resolution.md), so neither a question nor a pending
background command makes that end legitimate, and the block says to fix the cause. Only an operator
message saying "stop" passes it. The fourth such end since the operator's last message passes,
so a real dead end still surfaces; the count is read from this hook's own feedback in the
transcript. (gymie kan-1091: a merger dispatch refused over an `-unread` report, the reviewer's
report edited to pass it, the edit denied, the turn ended on a report to the operator.)

Before any of that, it reads the turn's last assistant text: a Finished handoff (a `## Finished`
line, skills/flow/cleanup.md) or a /flow-fast landing report (recognised by carrying either line)
that lacks `**Learning:**` or `**Prompt tuning:**` blocks with a reason naming the missing
line(s) -- in any session, with no store lookup. With stop_hook_active true this check is skipped,
so a turn that already got that block once is never trapped. Only text after the operator's last
(non-isMeta) message counts, so a turn ending on tool calls alone is never judged on an earlier
turn's text; and a label counts only at the start of a line, so prose quoting `**Learning:**`
mid-sentence is not a handoff. A handoff split across text blocks around a tool call is judged on
its last text block alone.

Ticked/total counts are read from the change's tasks.md under the stage run's repoRoot:
`flow tasks count` prints only the total and records a plan-growth observation as a side
effect, so it is not used here.

Fails open: any error -- bad input, store down, unreadable plan -- exits 0 with no output.
A bug here must never trap a session's turn.
"""

import importlib.util
import json
import os
import re
import sys
from typing import Optional
from urllib import parse, request

_spec = importlib.util.spec_from_file_location(
    "flow_active_change",
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "flow-active-change.py"),
)
_fac = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_fac)
_aoa_spec = importlib.util.spec_from_file_location(
    "await_own_agents",
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "await-own-agents.py"),
)
_aoa = importlib.util.module_from_spec(_aoa_spec)
_aoa_spec.loader.exec_module(_aoa)

TASK_RE = re.compile(r"^- \[([ xX])\] ")
STOP_RE = re.compile(r"^\s*(stop|pause|wait)\b", re.IGNORECASE)
TRANSCRIPT_TAIL_BYTES = 1 << 20
HANDOFF_LINES = ("**Learning:**", "**Prompt tuning:**")
QUESTION_RE = re.compile(r"^## Question\s*$", re.MULTILINE)


REFUSAL_RE = re.compile(r"\bblocked\b|refused|auto mode classifier", re.IGNORECASE)
REFUSAL_BLOCK = "flow-stop-guard: the turn ends on a refusal"
REFUSAL_CAP = 3


def store_get(path: str, query: dict) -> dict:
    addr = os.environ.get("FLOW_ADDR") or _fac.DEFAULT_STORE_ADDR
    url = f"{addr}{path}?{parse.urlencode(query)}"
    with request.urlopen(url, timeout=_fac.STORE_TIMEOUT_SECONDS) as resp:
        return json.load(resp)


def last_flow_run(session_id: str) -> Optional[dict]:
    body = store_get(
        "/api/v1/stage-runs",
        {"session_id": session_id, "command": "/flow", "sort": "-started_at", "limit": "1"},
    )
    runs = body.get("stageRuns")
    if not isinstance(runs, list) or not runs or not runs[0].get("changeName"):
        return None
    return runs[0]


def change_record(run: dict) -> Optional[dict]:
    records = store_get("/api/v1/changes/find", {"name": run["changeName"]}).get("records") or []
    for rec in records:
        if rec.get("projectKey") == run.get("projectKey"):
            return rec
    return records[0] if records else None


def task_counts(run: dict) -> tuple:
    root = run.get("repoRoot")
    if not root:
        return 0, 0
    for leaf in ("spectre", "openspec"):
        path = os.path.join(root, leaf, "changes", run["changeName"], "tasks.md")
        if os.path.isfile(path):
            marks = [m.group(1) for m in map(TASK_RE.match, open(path, encoding="utf-8")) if m]
            return sum(1 for c in marks if c != " "), len(marks)
    return 0, 0


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


def transcript_allows(lines: list) -> bool:
    """True when the turn ends on an AskUserQuestion or the operator's last message starts with stop/pause/wait."""
    last_user, last_tool = "", None
    for line in lines:
        try:
            entry = json.loads(line)
        except ValueError:
            continue
        if not isinstance(entry, dict):
            continue
        content = (entry.get("message") or {}).get("content")
        if entry.get("type") == "user" and not entry.get("isMeta"):  # isMeta: hook feedback, skill text, hand-backs
            texts = _texts(content)
            if texts:  # a tool_result-only entry is not an operator message
                last_user, last_tool = " ".join(texts), None
        elif entry.get("type") == "assistant" and isinstance(content, list):
            for block in content:
                if isinstance(block, dict) and block.get("type") == "tool_use":
                    last_tool = block.get("name")
    return last_tool == "AskUserQuestion" or bool(STOP_RE.search(last_user))


def refusal_stop(lines: list) -> bool:
    """True when the turn's last tool result is a refusal the operator did not stop on, under the cap."""
    last_user, refused, blocks = "", False, 0
    for line in lines:
        try:
            entry = json.loads(line)
        except ValueError:
            continue
        if not isinstance(entry, dict):
            continue
        content = (entry.get("message") or {}).get("content")
        if entry.get("type") == "user":
            texts = _texts(content)
            if entry.get("isMeta"):
                blocks += any(REFUSAL_BLOCK in t for t in texts)
            elif texts:
                last_user, refused, blocks = " ".join(texts), False, 0
            for block in content if isinstance(content, list) else []:
                if isinstance(block, dict) and block.get("type") == "tool_result":
                    body = block.get("content")
                    refused = bool(block.get("is_error")) and bool(REFUSAL_RE.search(" ".join(_texts(body))))
        elif entry.get("type") == "assistant" and isinstance(content, list):
            if any(isinstance(b, dict) and b.get("type") == "tool_use" for b in content):
                refused = False  # a later call: the refusal is no longer the last result
    return refused and blocks < REFUSAL_CAP and not STOP_RE.search(last_user)


LAUNCH_RE = re.compile(r'"status":\s*"async_launched",\s*"agentId":\s*"([^"]+)"')
NOTIFIED_RE = re.compile(r"<task-notification>(?:\\n|\s)*<task-id>([^<]+)</task-id>")


def agents_pending(lines: list) -> bool:
    """True when a background agent launched in the transcript tail has not notified its end."""
    launched, notified = set(), set()
    for line in lines:
        launched.update(LAUNCH_RE.findall(line))
        notified.update(NOTIFIED_RE.findall(line))
    return bool(launched - notified)


def bash_pending(path) -> bool:
    """True when a Bash command this transcript moved to the background has not notified its end."""
    return isinstance(path, str) and os.path.isfile(path) and bool(_aoa.running_background(path))


def main() -> int:
    payload = json.load(sys.stdin)
    if not isinstance(payload, dict):
        return 0
    lines = transcript_entries(payload.get("transcript_path"))
    last = last_text(lines)
    missing = [] if payload.get("stop_hook_active") else handoff_missing(last)
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
    refused = refusal_stop(lines)
    if not refused and (transcript_allows(lines) or QUESTION_RE.search(last) or agents_pending(lines) or bash_pending(payload.get("transcript_path"))):
        return 0
    session_id = _fac.resolve_session_id(payload)
    if not session_id:
        return 0
    run = last_flow_run(session_id)
    if not run:
        return 0

    rec = change_record(run)
    if not rec or rec.get("state") != "STARTED":
        return 0
    if refused:
        json.dump(
            {
                "decision": "block",
                "reason": (
                    f"{REFUSAL_BLOCK} -- a guard refusal, a blocked dispatch or a denied action -- and "
                    f"/flow change {run['changeName']} has no handoff yet. That is not a stop **What still "
                    f"stops** names (skills/flow-contracts/operator-prompts-auto-resolution.md): fix the "
                    f"cause and continue: re-dispatch under a new key, re-run the guard, never edit a report "
                    f"or a ledger row to pass a gate."
                ),
            },
            sys.stdout,
        )
        return 0
    ticked, total = task_counts(run)
    if not run.get("endedAt"):
        reason = f"stage {run.get('stage')} is still open"
    elif total > 0 and ticked == total:
        reason = f"all {total} tasks are ticked at STARTED"
    else:
        return 0

    nxt = (
        "test list, review panel, verify, then flow.write-in-progress"
        if reason.startswith("all")
        else f"finish stage {run.get('stage')}"
    )
    json.dump(
        {
            "decision": "block",
            "reason": (
                f"flow-stop-guard: /flow change {run['changeName']} is mid-phase ({reason}). "
                f"A turn ends only at a gate, a block, or the run's final handoff "
                f"(skills/flow-contracts/pipeline.md). Next: {nxt}."
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
