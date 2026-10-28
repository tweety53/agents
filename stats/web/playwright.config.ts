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
    // Pinned to software rasterisation: unpinned, Chromium's GPU and
    // partial raster paths antialias a few edge pixels (select corners)
    // differently from run to run, so three captures of an unchanged tree
    // differed pixelwise; with these flags they are byte-identical
    // (kan-1046 self-review, 8d).
    launchOptions: {
      args: [
        "--disable-gpu-rasterization",
        "--disable-partial-raster",
        "--disable-gpu",
      ],
    },
  },
  reporter: [["list"]],
  timeout: 60_000,
  // No `expect` tolerance: screenshots compare at Playwright's default of
  // zero differing pixels. The rasterisation pin above removes the edge
  // antialiasing noise at its source rather than tolerating it; a future
  // flake is fixed the same way, never by widening a tolerance.
});
