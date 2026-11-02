import { describe, expect, it } from "vitest";
import { weekWindow } from "./weekWindow";

describe("weekWindow", () => {
  it("weekWindow starts n whole Wednesday weeks back from a mid-week now", () => {
    const now = Date.parse("2026-06-19T12:00:00Z");
    const w = weekWindow(8, now);
    expect(w.from.toISOString()).toBe("2026-04-29T00:00:00.000Z");
    expect(w.to.getTime()).toBe(now);
    expect(w.weeks).toEqual(["2026-04-29", "2026-05-06", "2026-05-13", "2026-05-20", "2026-05-27", "2026-06-03", "2026-06-10", "2026-06-17"]);
  });

  it("weekWindow counts a Wednesday as the first day of its own week", () =>
    expect(weekWindow(2, Date.parse("2026-06-17T00:30:00Z")).weeks).toEqual(["2026-06-10", "2026-06-17"]));
});
