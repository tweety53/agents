import { describe, expect, it } from "vitest";
import { formatModel, formatRelative, formatSpan, formatUsd } from "./format";

describe("formatUsd", () => {
  it("formatUsd always renders two decimals", () => {
    expect(formatUsd(2.957)).toBe("$2.96");
    expect(formatUsd(2.75)).toBe("$2.75");
    expect(formatUsd(0.028)).toBe("$0.03");
    expect(formatUsd(3)).toBe("$3.00");
  });
});

describe("formatRelative", () => {
  const now = Date.parse("2026-10-28T21:20:00Z");
  const ago = (min: number) => new Date(now - min * 60000).toISOString();
  it("formatRelative reads an instant as the time since it", () => {
    expect(formatRelative(ago(0.4), now)).toBe("just now");
    expect(formatRelative(ago(1), now)).toBe("1m ago");
    expect(formatRelative(ago(59), now)).toBe("59m ago");
    expect(formatRelative(ago(140), now)).toBe("2h ago");
    expect(formatRelative(ago(1500), now)).toBe("1d ago");
    expect(formatRelative(ago(2900), now)).toBe("2d ago");
  });

  it("formatSpan reads minutes under an hour and hours with padded minutes above", () => {
    expect(formatSpan(23 * 60000)).toBe("23m");
    expect(formatSpan(88 * 60000)).toBe("1h 28m");
    expect(formatSpan(5 * 3600e3 + 5 * 60000)).toBe("5h 05m");
  });
});

describe("formatModel", () => {
  it("formatModel shortens a model id to family and version", () => {
    expect(formatModel("claude-opus-4-7")).toBe("opus-4.7");
    expect(formatModel("claude-sonnet-5")).toBe("sonnet-5");
    expect(formatModel("claude-haiku-4-5-20251001")).toBe("haiku-4-5-20251001");
    expect(formatModel("opus")).toBe("opus");
  });
});
