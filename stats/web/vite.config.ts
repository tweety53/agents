import { defineConfig } from "vite";
import { configDefaults } from "vitest/config";
import react from "@vitejs/plugin-react";

// The build output lands inside internal/web/dist -- a sibling Go package
// directory, not stats/web/dist -- because go:embed patterns cannot climb
// above the directory holding the //go:embed directive (no "../" in an
// embed pattern). Pointing outDir there is what lets
// stats/internal/web/embed.go embed the build with a single, ordinary
// `//go:embed all:dist` right beside it, and it is also the mechanism
// behind this task's "no dist/, no build" requirement: until `vite build`
// has run at least once, internal/web/dist does not exist, and the
// package that embeds it refuses to compile -- loudly, at compile time,
// never as an empty page served at runtime.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../internal/web/dist",
    emptyOutDir: true,
  },
  // Task 13 adds component tests (DataTable, the eight views) that render
  // real DOM via @testing-library/react, which needs a DOM to render into
  // -- task 12's api.test.ts needed none of this, being pure functions.
  test: {
    // The default `forks` pool spawns one process per CPU; a killed parent
    // (Bash timeout, stopped agent) orphans them to PID 1, ~2.5 GB held.
    // Threads die with the parent; two cap the footprint.
    pool: "threads",
    maxWorkers: 2,
    environment: "jsdom",
    setupFiles: ["./src/setupTests.ts"],
    // tests/visual/baseline.spec.ts (KAN-171) is a Playwright suite, run
    // through its own `npm run test:visual`, never through vitest -- its
    // name still matches vitest's default `*.spec.ts` include glob, so it
    // must be excluded explicitly or vitest tries to import it directly
    // and fails on Playwright's own test.describe() guard against being
    // called outside the Playwright test runner.
    exclude: [...configDefaults.exclude, "tests/visual/**"],
  },
});
