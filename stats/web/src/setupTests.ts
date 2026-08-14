// Loaded once per test file via vite.config.ts's test.setupFiles. Extends
// vitest's `expect` with the jest-dom matchers (toBeInTheDocument, etc.)
// that views.test.tsx and DataTable.test.tsx use to assert what a reader
// actually sees, rather than only that a component mounted, and unmounts
// each rendered tree after its test -- this project does not opt into
// vitest's `globals: true`, so testing-library's own afterEach(cleanup)
// autorun never fires and every test after the first would otherwise see
// every previous test's DOM still attached to document.body.
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

afterEach(() => {
  cleanup();
});
