#!/usr/bin/env python3
"""cache-misses.py — report every zero-cache-read turn in one or more Claude
Code transcripts. A turn whose `cache_read_input_tokens` is 0 (or absent)
while `input_tokens` exceeds 5000 is a full re-price of the whole context at
full input rate instead of the cheap cache-read rate (KAN-473). This script
lists each such turn with its gap since the previous surviving turn, what
preceded it, and the tools used since the last usage-bearing turn, so a
future `/flow` run's transcripts can be checked for the same pattern.

This is a report, not a guard: it always exits 0, and an unreadable or
non-JSONL file is named on stderr and skipped rather than failing the run.
"""

import argparse
import datetime
import json
import sys
from pathlib import Path

MIN_INPUT_TOKENS = 5000


def label_for(path: Path) -> str:
    meta_path = path.with_suffix("").with_suffix(".meta.json")
    if meta_path.exists():
        try:
            meta = json.loads(meta_path.read_text())
            description = meta.get("description")
            if description:
                return description
        except (OSError, json.JSONDecodeError):
            pass
    return path.name


def expand_paths(args):
    for arg in args:
        path = Path(arg)
        if path.is_dir():
            for child in sorted(path.glob("*.jsonl")):
                yield child
        else:
            yield path


def parse_timestamp(ts: str) -> datetime.datetime:
    if ts.endswith("Z"):
        ts = ts[:-1] + "+00:00"
    return datetime.datetime.fromisoformat(ts)


def load_entries(path: Path):
    with path.open() as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            yield json.loads(line)


def content_kind(message_content) -> str:
    if isinstance(message_content, str):
        return "text"
    if isinstance(message_content, list) and message_content:
        first = message_content[0]
        if isinstance(first, dict) and first.get("type") == "text":
            return "text"
    return "tool_result"


def report(path: Path) -> None:
    print(f"===== {label_for(path)}")

    tools_since_last_usage = []
    prev_user_kind = "text"
    prev_turn_time = None
    seen_usage_keys = set()
    turn_number = 0

    for entry in load_entries(path):
        entry_type = entry.get("type")

        if entry_type == "user":
            message = entry.get("message", {})
            prev_user_kind = content_kind(message.get("content"))
            continue

        if entry_type != "assistant":
            continue

        message = entry.get("message", {})
        for block in message.get("content", []) or []:
            if isinstance(block, dict) and block.get("type") == "tool_use":
                name = block.get("name")
                if name:
                    tools_since_last_usage.append(name)

        usage = message.get("usage")
        if not usage:
            continue

        timestamp = entry.get("timestamp")
        input_tokens = usage.get("input_tokens", 0)
        cache_read = usage.get("cache_read_input_tokens", 0)

        key = (timestamp[:19] if timestamp else None, input_tokens, cache_read)
        if key in seen_usage_keys:
            continue
        seen_usage_keys.add(key)
        turn_number += 1

        turn_time = parse_timestamp(timestamp) if timestamp else None
        if turn_time is not None and prev_turn_time is not None:
            gap = (turn_time - prev_turn_time).total_seconds()
        else:
            gap = 0.0

        if not cache_read and input_tokens > MIN_INPUT_TOKENS:
            hhmmss = turn_time.strftime("%H:%M:%S") if turn_time else "??:??:??"
            print(
                f"turn#{turn_number} {hhmmss} gap={gap:.1f}s in={input_tokens} "
                f"prev_user={prev_user_kind} "
                f"tools_since_last_usage={tools_since_last_usage}"
            )

        tools_since_last_usage = []
        if turn_time is not None:
            prev_turn_time = turn_time


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Report every zero-cache-read turn over 5000 input tokens "
        "in one or more transcripts."
    )
    parser.add_argument(
        "paths", nargs="+", help="transcript .jsonl file(s) or subagents/ directory(ies)"
    )
    args = parser.parse_args()

    for path in expand_paths(args.paths):
        try:
            report(path)
        except (OSError, json.JSONDecodeError) as exc:
            print(f"cache-misses.py: skipping {path}: {exc}", file=sys.stderr)

    return 0


if __name__ == "__main__":
    sys.exit(main())
