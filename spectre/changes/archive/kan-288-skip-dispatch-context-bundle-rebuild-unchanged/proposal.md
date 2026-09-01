# kan-288-skip-dispatch-context-bundle-rebuild-unchanged

## Why

Self-review finding from `kan-258-store-native-run-record`, angle 2 (token/time cost).
`gather-dispatch-context.sh` is rebuilt at the start of every dispatching stage
(`skills/flow/implement.md`'s SDD/TDD stage, `skills/flow/review-panel.md`'s stage start, and its
panel-fix round) — by design, since a fix documented mid-run must never leave a later dispatch
reading a stale plan. In practice the bundle is regenerated even when nothing in it changed: two
runs of one `/flow` pass produced ~150KB each, byte-identical, with no edit in between.

## What changes

`gather-dispatch-context.sh` takes a new required 5th argument, `<output-path>`, and writes the
bundle there itself instead of the caller redirecting its stdout. It hashes the bundle body (the
census line plus every `## <label>` section — everything but the `generated:`/`head:` header
lines) and compares it to a sidecar `<output-path>.hash` from the previous call. When they match,
the write is skipped and the existing file is left untouched; when they differ (including no prior
hash at all), the bundle and the hash are (re)written. Either outcome is reported to stderr, and
each caller surfaces that line so the run stays observable rather than silently reusing a bundle.

The invariant the existing "always rebuild" rule protects is unchanged: every dispatching stage
still *runs* the script every time, so a fix that edits `proposal.md`/`tasks.md` between two
dispatches is always picked up on the very next call — only the cost of an unchanged rebuild is
removed, never the check itself.
