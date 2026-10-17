import { describe, expect, it } from "vitest";
import { formatUsd } from "./format";

describe("formatUsd", () => {
  it("formatUsd always renders two decimals", () => {
    expect(formatUsd(2.957)).toBe("$2.96");
    expect(formatUsd(2.75)).toBe("$2.75");
    expect(formatUsd(0.028)).toBe("$0.03");
    expect(formatUsd(3)).toBe("$3.00");
  });
});
