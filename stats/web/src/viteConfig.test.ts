// @vitest-environment node
import { describe, expect, it } from "vitest";
import config from "../vite.config";

// A killed parent vitest orphans `forks` workers to PID 1 (~180 MB each);
// threads die with the parent.
describe("vite.config test pool", () => {
  it("runs on two worker threads", () => {
    expect(config.test?.pool).toBe("threads");
    expect(config.test?.maxWorkers).toBe(2);
  });
});
