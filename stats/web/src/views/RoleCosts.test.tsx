// RoleCosts draws GET /api/v1/stats/role-costs: one row per allowed-pairs
// role with its spend, agent time, per-dispatch figures, failure rate and
// model · effort split, and the weekly spend per role.
import { render, screen, within } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import type { RoleCostsResult } from "../api";
import { RoleCosts } from "./RoleCosts";

const { fetchStatsViewMock, fetchProjectsMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn(), fetchProjectsMock: vi.fn() }));
vi.mock("../api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api")>()),
  fetchStatsView: fetchStatsViewMock,
  fetchProjects: fetchProjectsMock,
}));

const DATA: RoleCostsResult = {
  roles: [
    { role: "implementer", group: "Implementation", dispatches: 4, priced: 3, failed: 1, costUSD: 6, agentSeconds: 7200, medianSeconds: 600,
      pairs: [{ pair: "opus · low", costUSD: 5, dispatches: 3 }, { pair: "sonnet · low", costUSD: 1, dispatches: 1 }] },
    { role: "panel-reviewer", group: "Review", dispatches: 2, priced: 2, failed: 0, costUSD: 2, agentSeconds: 600, medianSeconds: 300,
      pairs: [{ pair: "opus · medium", costUSD: 2, dispatches: 2 }] },
  ],
  weeks: [{ week: "2026-11-11", roles: { implementer: 6, "panel-reviewer": 2 } }],
};

beforeEach(() => {
  fetchProjectsMock.mockResolvedValue([]);
  fetchStatsViewMock.mockResolvedValue(DATA);
});

it("draws one row per role with spend, agent time, per-dispatch figures and failure rate", async () => {
  render(<RoleCosts />);
  const table = await screen.findByRole("table", { name: "Spend and agent time per role" });
  const rows = within(table).getAllByRole("row").slice(1);
  expect(rows.map((r) => within(r).getAllByRole("cell").map((c) => c.textContent))).toEqual([
    ["implementer", "Implementation", "4", "$6.00", "75%", "2h 00m", "10m", "$2.00", "25%", "opus · low $5.00 ×3sonnet · low $1.00 ×1"],
    ["panel-reviewer", "Review", "2", "$2.00", "25%", "0h 10m", "5m", "$1.00", "0%", "opus · medium $2.00 ×2"],
  ]);
  expect(fetchStatsViewMock).toHaveBeenCalledWith("role-costs", expect.objectContaining({ project: undefined }));
});

it("draws each week's spend as one segment per role", async () => {
  render(<RoleCosts />);
  const week = await screen.findByRole("group", { name: "Week of 2026-11-11, $8.00" });
  expect(within(week).getAllByTestId("role-seg").map((s) => s.getAttribute("title"))).toEqual(["implementer $6.00", "panel-reviewer $2.00"]);
});
