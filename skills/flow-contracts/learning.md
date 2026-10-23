# Learning

**This file is canonical for the learning note.** The phase files that compose and print it cite it
and restate none of it.

### When a note runs

A note is composed only on `/flow`'s implementation run — the first run past the plan gate — and on
`/flow-fast`.
Never a fix run, an integrate run, or a creating run that stops at the plan gate.

### Reading the setting

```bash
flow settings get | jq .learning
```

`true` composes a note; `false` composes none. The operator switches it with `/flow-settings`, per
**Workflow** (`skills/flow-settings/SKILL.md`). Anything else — a non-zero exit, or the empty output the pipe
leaves when the store is unreachable — prints no note and never blocks the run.

### Choosing a topic

```bash
flow learning list
```

prints one JSON array of `{topic, count, lastNote}`. A topic at `count` 3 is exhausted and skipped.
Pick, in this order:

1. **Related to the run** — the plan's language, a design approach its tasks use, or an
   agentic-coding, LLM or AI-infrastructure practice the run itself exercises.
2. **Otherwise random** — a programming-language feature or a system-design pattern.

A topic is a lowercase kebab slug (`go-generics`, `advisory-locks`), at most 64 characters.

### The note

- **Theory** — one paragraph of at most about 150 words, tied to the run where the topic is.
- **Puzzle** — a short coding puzzle where the topic admits one, its answer on its own
  `**Answer:**` line after it — never an HTML fold, per **Handoff output**
  (`skills/flow-contracts/pipeline.md`); omitted where it does not.
- **Study** — two or three links, each checked first with
  `curl -sfI --max-time 5 <url>` — credential-free, read-only — and dropped on a non-zero exit.

### Recording it

Recorded as the final message prints, never at composition, so a run that stops earlier spends no
showing of the topic:

```bash
flow learning add <topic> -note '<the topic in under 80 characters>' -change <name> -C <project>
```

`-C` names the main checkout, never the worktree: `/flow-fast`'s cleanup has removed its worktree
before the final message prints, and the project key derives from the git common directory, the
same from either.

Exit 1 with stderr naming `already shown three times` or `rejected` is the store refusing the
topic, and the run picks another topic. Any other non-zero exit — the store unreachable — prints the
note unrecorded and never blocks.

### The printed section

Printed in the run's final message, where the citing phase file places it:

```text
## Learning — <topic>

<the theory paragraph>

**Puzzle:** <the puzzle>

**Answer:** <the answer>

**Study:** <link> · <link> · <link>
```

`**Puzzle:**` and `**Answer:**` are omitted where the note carries none.
