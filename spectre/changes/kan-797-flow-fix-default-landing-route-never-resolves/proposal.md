# kan-797-flow-fix-default-landing-route-never-resolves

## Why

`## default landing route` never resolves when its body carries prose after the literal —
`project-configuration.md`'s match rule is byte-for-byte on the whole trimmed body, so a
documented default (`merge and push` plus an explanatory sentence) is a malformed row: reported,
dropped, resolved as if the key were absent, and the landing question is asked on every run.
Two runs tripped (kan-741, kan-577 integrate narratives); this repo's own `.flow/project.md` was
stripped back to the bare literal as the dodge, which silences the incident without fixing the
mechanism.

Reproduced at base `4a278320` before planning:

- a temp project whose body is `merge and push — <prose>` resolves through `project-get.sh` to
  the whole prose-bearing body → byte-for-byte against `merge and push` fails → absent;
- today's clean config resolves `merge and push`, exit 0 — the incident is dormant, the
  mechanism intact;
- no guard validates the key at authoring time (`check-model-keys` covers
  `## self review model` only).

## What changes

- A single-line-literal key's value resolves as **the body's first non-blank line** —
  whitespace-trimmed, surrounding backticks removed, matched byte-for-byte against the key's
  vocabulary; lines below it are documentation for the reader, never read. A documented default
  may carry its explanation and still resolves.
- The rule is stated once, for all four keys sharing the match sentence: `## default landing
  route`, `## self review model`, `## self review`, `## handoff`. `## jira` is untouched.
- `check-model-keys` reads the head for `## self review model` the same way, so the one Go guard
  touching the family agrees with the contract.
- Everything else is unchanged: the extractor stays generic, `project-get.sh` is untouched, and a
  head matching no literal is still a malformed row — reported by name, dropped.
