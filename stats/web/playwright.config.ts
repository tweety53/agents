import { defineConfig } from "@playwright/test";

// Visual baseline for the stats SPA's dashboard and run-detail views
// (KAN-171). The target is the disposable UI-test stack on 127.0.0.1:4174
// (`make ui-test-up` / `make ui-test-down` in stats/Makefile) -- never the
// dev workspace's daemon on 4173, whose data changes on every run and so
// could never hold a stable baseline, and which CLAUDE.md forbids this
// suite from touching at all. `make ui-test-up` drops and reseeds
// `flow_uitest` with a fixed fixture every time, which is what makes a
// pixel baseline over it stable rather than flaking on live data drift.
export default defineConfig({
  testDir: "./tests/visual",
  use: {
    baseURL: "http://127.0.0.1:4174",
    // Pinned so every capture -- baseline and comparison alike -- renders
    // at the same pixel size. An unpinned viewport would make a screenshot
    // diff meaningless: two runs could differ only because the window did.
    viewport: { width: 1280, height: 1024 },
    deviceScaleFactor: 1,
    // Explicit, not left to the OS default: this is exactly the axis
    // KAN-16's second defect broke on -- a Grafana-style dark interface was
    // asked for, and a browser without an explicit `prefers-color-scheme:
    // dark` signal got a plain white document instead, because the dark
    // palette sat behind that media query with no working default.
    // theme.css now declares dark unconditionally on bare `:root`, with
    // light served only for an explicit light preference or
    // `data-theme="light"` (a deliberate feature, not this defect's
    // subject -- see theme.css's own header comment). Pinning "dark" here
    // is what makes the assertion below meaningful: a browser that DOES
    // prefer dark must actually receive the dark tokens, which is exactly
    // what a regression back to a broken or misnamed selector would break.
    // (Chromium has no true "no-preference" state to test against --
    // it resolves that setting to "light", which is this app's other,
    // intentionally-supported theme, not a way to observe an absent
    // preference.)
    colorScheme: "dark",
  },
  reporter: [["list"]],
  timeout: 60_000,
  // A handful of pixels differ run-to-run from font antialiasing/subpixel
  // rendering even on an unchanged screen -- tolerate a small, fixed pixel
  // count so the baseline doesn't flake, while still catching real drift.
  expect: {
    toHaveScreenshot: { maxDiffPixels: 40 },
  },
});
