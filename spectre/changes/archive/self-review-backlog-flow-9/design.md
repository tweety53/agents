# self-review-backlog-flow-9

## Context

- Five Backlog Flow findings — the next five live ones not already fixed, oldest first: KAN-886,
  KAN-889, KAN-890, KAN-896, KAN-898 — planned as one change per item 1 of **3. Fix every finding
  that is not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-19-self-review-backlog-flow-9-design.md` in this worktree. Every
  question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **Not in this change.** KAN-764 is a design ask, a frame only the operator can draw, and stays
  in Backlog Flow. Two candidates were found not live on `origin/main` while verifying, for the
  parent to move to Done:
  - **KAN-894** — KNOWN-BUGS.md's shared tail conflicts on every rebase: the premise is gone. The
    KNOWN-BUGS route is retired — a found defect "is never logged to a `<project>/KNOWN-BUGS.md`
    file" (`skills/flow-contracts/known-bugs.md:15`) — and no skill appends to the file any more,
    so no rebase meets an appended tail. Its last append is `70f641a0` (2026-10-09), the day the
    route was retired.
    <!-- measured: git log -3 --format='%h %ad %s' --date=short -- KNOWN-BUGS.md @ 373988df → 70f641a0 2026-10-09 newest; git grep -n KNOWN-BUGS -- skills rules scripts stats @ 373988df → known-bugs.md:15's prohibition and two SKILL-rationale.md history lines, no writer -->
  - **KAN-897** — names truncated mechanically at 48 characters: disproved. The rule asks "at most
    48 characters truncated at a word boundary"
    (`skills/flow-contracts/jira-integration.md:71`), and both cited names meet it — kan-841's
    slug `agents-port-the-next-ten-slowest-bash-scripts-to` is 48 characters ending on the whole
    word `to`, kan-816's `flow-fix-the-archive-step-s-set-pair-loop-does` 46 ending on `does`.
    No code derives the slug; a trailing function word is a whole word, not a cut one.
    <!-- measured: printf %s <slug> | wc -c for both slugs, a measurement of two literal names, not re-runnable at a ref → 48 and 46; git grep -n -E '48' -- stats/cmd stats/internal scripts skills @ 373988df → no slug truncation code -->
- **KAN-886 — Task 1.** Source: kan-778's panel F6, deferred as "the four refusals are correct at
  HEAD; only their tests are missing". The four are `panelreproducers.go`'s `state == nil` (a
  `flow state get` printing `null` at exit 0 decodes into a nil map with no error),
  `uwOpenFindings`' string-status check in `unfinishedwork.go` (a null status decodes into `""`),
  and the `FLOW_GUARD_REPO_ROOT` unset refusals of `references.go` and `installedcitations.go`.
  Blast radius: four test files.
  <!-- measured: each refusal removed in turn, its guard's test run (go test ./internal/guard -run TestCheckPanelReproducers|TestCheckUnfinishedWork|TestCheckReferences|TestCheckInstalledCitations) in a scratch clone @ 373988df → ok for all four -->
- **KAN-889 — Task 2.** Source: kan-786's panel F1 and F2, deferred minors. The map parse
  (`stats/internal/guard/taskcommitfields.go:134-147`) refuses an entry that is not
  `<worktree>=<sha>` and a worktree named twice; `tcfMapRepos` builds the satellite with no change
  directory at all, so `changePlanDir`'s link branch (`stats/internal/guard/changeplan.go`) is
  never reached under the map. Blast radius: one test file.
  <!-- measured: the malformed-entry check, the duplicate check and changePlanDir's canonical link lookup each removed in turn, go test ./internal/guard -run TestCheckTaskCommitFields in a scratch clone @ 373988df → ok for the first two (no case reaches them) -->
- **KAN-890 — Task 3.** Source: `docs/self-review/kan-789-flow-improvement-mutation-kill-every-fix-self-review.md`.
  `pfdPanelFixRows` (`stats/internal/guard/panelfixsingledispatch.go`) keeps only rows of the
  invoking token, so the retry-alone rule fires on a resumed round's `-retry`. Round numbers are
  change-unique — **A run's first round is the number `flow record next-round -change <name>`
  prints** (`skills/flow/review-panel.md`) — so another token's original under the same key is the
  same round's. Blast radius: three files.
- **KAN-896 — Task 4.** Source:
  `docs/self-review/kan-813-flow-improvement-re-verify-against-a-fresh-head-self-review.md`.
  `check-verbatim-moves` already compares sentence multisets, so the "positional chunking" is
  `vmBlocks` (`stats/internal/guard/verbatimmoves.go`): a list item's first line stands alone and
  its wrapped continuation lines form a separate block, cutting the sentence that spans them.
  Blast radius: two files.
  <!-- measured: the current guard over kan-813's own commits (FLOW_GUARD_REPO_ROOT=<a worktree at f603fe9b> flow-guard check-verbatim-moves d4b55fa8^) in a scratch clone @ 373988df → 5 FAIL lines, two "deleted or reworded" on the untouched Run **Verify** sentence's two halves; with Task 4's diff → 2 FAIL lines, both genuinely new sentences -->
- **KAN-898 — Task 5.** Source:
  `docs/self-review/kan-817-flow-fix-the-done-when-named-gitignored-self-review.md` and
  `docs/self-review/kan-838-flow-fast-record-guard-self-review.md`. The store refuses any pair but
  `glm-5.3-flash`/`high` on a `zcode` session token (`stats/internal/store/harnessmap.go`,
  KAN-610). **Harness mapping** (`skills/flow-contracts/model-policy.md`) and the dispatch
  paragraphs state the mapped pair, but the reviewer row's recording paragraph in
  `skills/flow/review-panel.md` — the rows both runs had refused, panel rows and fix-round re-run
  rows at `low` — names the decision and re-run pairs with no harness exception. Blast radius:
  one file.
- **No red check on `origin/main`.** `go test ./internal/guard`, `go vet ./...`, `gofmt -l .` and
  the prose guards exit 0 at `373988df`, so no task repairs one.
  <!-- measured: cd stats && go test ./internal/guard; go vet ./internal/guard ./internal/store; gofmt -l .; scripts/check-references.sh, check-installed-citations.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-normative-inventory.sh, check-ask-silence.sh in a scratch clone @ 373988df → all exit 0 -->

## Decisions

### Pin each untested refusal in its own guard's test

**ID:** pin-ported-refusals
**Status:** active
**Chosen:** one subtest per refusal, in the test file of the guard that carries it, each asserting exit 2 and the refusal's own stderr line — the needle that tells the refusal from a later one the guard would reach without it (`not a directory:` for an empty root).
**Considered:** asserting exit 2 alone — with the reference guard's refusal removed, an empty root still exits 2 on `not a directory: `, so the test would stay green; one shared table across the four guards — their harnesses build different `Env`s and fixtures, and a shared one is a new helper for four rows.

### Cover the commit map's refusals and link route as cases 160–162

**ID:** commit-map-cases
**Status:** active
**Chosen:** three cases after case 158 on `tcfMapRepos(t, "147")`: `<b1>=` refused naming the entry, `<b1>=<sha>,<b1>=<sha>` refused naming the worktree, and a first repository carrying `spectre/changes/x-repo-change/link.md` that passes under the map — the link branch of `changePlanDir`, which the absent-dir fixture never reaches.
**Considered:** a new fixture builder for the link-route satellite — one `writeFile` and one commit on the existing builder's repository suffice; a new plan fixture — plan `147` already declares both repositories' files, and `TestTaskFieldParseMatchesPython` would parse a second copy for nothing.

### Another token's original answers only whether the original exists

**ID:** foreign-token-original
**Status:** active
**Chosen:** `pfdPanelFixRows` returns the other session tokens' panel-fix rows apart; each canonical non-retry key among them counts as `foreign` on its base, which satisfies the retry-alone rule, a chunked round's bare-key rule and chunk contiguity; every dispatch count stays the invoking token's, and only rounds the invoking token touched are judged.
**Considered:** scoping every rule to the change, as KAN-890 proposed — case 6 pins that two runs each dispatching a round once are not one round dispatched twice, and an earlier run's rows carry `-commit` shas a rebase since rewrote, which the commit check would flag; exempting only the retry-alone rule — a round resumed mid-chunking (`panel-fix-1`, `-2` under the old token, `-3` under the new) would still fail the bare-key and contiguity rules on the same cause.

### A list item's indented continuation joins its block

**ID:** list-item-continuation-block
**Status:** active
**Chosen:** in `vmBlocks`, a list item opens a block its indented continuation lines join — indentation read past any blockquote markers — and an unindented line after it opens a block of its own; headings, table rows and rules still stand alone. Base and head are split by the same code, so no existing sentence's verdict moves.
**Considered:** anchoring a moved sentence on its matched neighbours, as KAN-896 proposed — the comparison is already a sentence multiset with no positional step, and the cut is the block boundary, so anchoring would treat a symptom; joining any following line, indented or not — `TestVMSentences` pins an unindented quote after a numbered item as its own sentence.

### The reviewer-row paragraph states the mapped pair

**ID:** reviewer-row-mapped-pair
**Status:** active
**Chosen:** one sentence after the paragraph's last, stating that on `zcode` every pair it names is recorded as `glm-5.3-flash` / `high` — the one pair the store accepts there — and on `muse` as the session's own, citing **Harness mapping**.
**Considered:** the store mapping every refused pair to the harness's own, KAN-898's alternative — KAN-610's refusal is the signal that a dispatcher recorded a pair its harness could not have run, and a silent rewrite would also keep the pre-mapping value out of every reply's handshake comparison; restating the mapping at each recording site — the dispatch paragraphs of implement.md, review-panel.md's roster and the fix round already state it, and the reviewer-row paragraph is the one site both refused runs recorded from.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — Tasks 3 and 4 change guard behaviour every panel close and every prose change's lint run go through.
**Considered:** flow-low — Tasks 1, 2 and 5 are tests or one sentence, but the two guard changes alter verdicts a low-effort pass could under-read.

### One fixer group

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, Tasks 1–5 in order, one fixer dispatch (`self-review-backlog-flow-9-fix`): every task is a `git apply` of a diff verified in a scratch clone, and the files are disjoint.
**Considered:** one group per task — five dispatches for work one fixer finishes in series; a Go group and a prose group — the prose task is one sentence.

## Open questions

## Live check

No running service or persistent state is touched: Tasks 3 and 4 change `flow-guard` subcommands
built per call from the checkout by `scripts/lib/flow-guard.sh`, Tasks 1 and 2 add tests, and
Task 5 is prose.
