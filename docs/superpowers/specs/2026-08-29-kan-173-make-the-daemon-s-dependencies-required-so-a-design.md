# KAN-173 — the daemon's dependencies, required rather than optional

The approved design for this change is its own artifact, and is canonical there rather than copied
here:

- `spectre/changes/kan-173-make-the-daemon-s-dependencies-required-so-a/design.md` — the problem,
  what the code holds today, the decisions and what each one ruled out, and the honest limit of
  what this change closes.
- `spectre/changes/kan-173-make-the-daemon-s-dependencies-required-so-a/proposal.md` — why the
  change exists and what it changes.

The change is confined to `stats/`: `internal/harvest` (the `Deps` interface, `NoDeps`, and
`NewWatcher`'s signature) and `cmd/flowd` (the daemon's own wiring and its wiring test).
