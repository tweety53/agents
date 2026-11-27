// RoleCosts draws GET /api/v1/stats/role-costs as the approved prototype
// (.superpowers/sdd/<change>/role-costs-prototype.html): roles grouped under
// their pipeline stage, one line each with a model · effort mix, the full
// pair list on click, and the weekly spend split by stage.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
      pairs: [{ pair: "opus · low", costUSD: 4.5, dispatches: 2 }, { pair: "sonnet · low", costUSD: 1.5, dispatches: 1 }, { pair: "opus · default", costUSD: 0, dispatches: 1 }] },
    { role: "no-progress-fixer", group: "Implementation", dispatches: 1, priced: 1, failed: 0, costUSD: 1, agentSeconds: 60, medianSeconds: 60,
      pairs: [{ pair: "opus · low", costUSD: 1, dispatches: 1 }] },
    { role: "panel-reviewer", group: "Review", dispatches: 2, priced: 2, failed: 0, costUSD: 3, agentSeconds: 600, medianSeconds: 300,
      pairs: [{ pair: "opus · medium", costUSD: 3, dispatches: 2 }] },
  ],
  weeks: [
    { week: "2026-11-11", roles: { implementer: 5, "panel-reviewer": 3 } },
    { week: "2026-11-18", roles: { implementer: 1, "no-progress-fixer": 1 } },
  ],
};

beforeEach(() => {
  fetchProjectsMock.mockResolvedValue([]);
  fetchStatsViewMock.mockResolvedValue(DATA);
});

const cells = (r: HTMLElement) => within(r).getAllByRole("cell").map((c) => c.textContent);

it("opens on the 7d chip and requests its 7-day window", async () => {
  render(<RoleCosts />);
  await screen.findByRole("table", { name: "Spend and agent time per role" });
  const req = fetchStatsViewMock.mock.calls.find((c) => c[0] === "role-costs")![1] as { from: Date; to: Date };
  expect(Math.round((req.to.getTime() - req.from.getTime()) / 86400000)).toBe(7);
  const periodGroup = within(screen.getByRole("group", { name: "Period" }));
  expect(periodGroup.getByRole("button", { name: "7 days" })).toHaveAttribute("aria-pressed", "true");
});

it("groups roles under a stage row carrying the stage's totals", async () => {
  render(<RoleCosts />);
  const table = await screen.findByRole("table", { name: "Spend and agent time per role" });
  const rows = within(table).getAllByRole("row").slice(1).filter((r) => !r.hidden);
  expect(rows.map(cells)).toEqual([
    ["Implementation", "5", "$7.00", "70%", "2h 01m", "", "", "", ""],
    ["Writes the codeimplementer", "4", "$6.00", "60%", "2h 00m", "10m", "$2.00", "25%", "opus low · 75%+2"],
    ["Takes over a stuck fixno-progress-fixer", "1", "$1.00", "10%", "0h 01m", "1m", "$1.00", "0%", "opus low · 100%"],
    ["Review", "2", "$3.00", "30%", "0h 10m", "", "", "", ""],
    ["Reviews the branchpanel-reviewer", "2", "$3.00", "30%", "0h 10m", "5m", "$1.50", "0%", "opus medium · 100%"],
  ]);
});

it("opens every pair of a role under its mix on click, unpriced runs on one line", async () => {
  render(<RoleCosts />);
  const table = await screen.findByRole("table", { name: "Spend and agent time per role" });
  const row = within(table).getByRole("row", { name: /Writes the code/ });
  expect(screen.queryByTestId("role-pairs")).toBeNull();
  await userEvent.click(row);
  const list = screen.getByTestId("role-pairs");
  expect(within(list).getAllByTestId("role-pair").map((p) => p.textContent)).toEqual(["opuslow$4.50×2", "sonnetlow$1.50×1"]);
  expect(list).toHaveTextContent("+ 1 unpriced run (opus · default ×1)");
  await userEvent.click(row);
  expect(screen.queryByTestId("role-pairs")).toBeNull();
});

it("splits each week's spend by stage, in a bar and a stage-by-week table", async () => {
  render(<RoleCosts />);
  const bar = await screen.findByRole("group", { name: "Week of 2026-11-11, $8.00" });
  expect(within(bar).getAllByTestId("stage-seg").map((s) => s.getAttribute("title"))).toEqual(["Implementation $5.00", "Review $3.00"]);
  const table = screen.getByRole("table", { name: "Spend per week by stage" });
  expect(within(table).getAllByRole("row").slice(1).map(cells)).toEqual([
    ["Implementation", "$5.00", "$2.00", "$7.00"],
    ["Writes the codeimplementer", "$5.00", "$1.00", "$6.00"],
    ["Takes over a stuck fixno-progress-fixer", "—", "$1.00", "$1.00"],
    ["Review", "$3.00", "—", "$3.00"],
    ["Reviews the branchpanel-reviewer", "$3.00", "—", "$3.00"],
    ["Total", "$8.00", "$2.00", "$10.00"],
  ]);
});

it("shows a priced amount under a cent as <$0.01, never $0.00", async () => {
  fetchStatsViewMock.mockResolvedValue({
    roles: [{ role: "pair-judge", group: "Pair choice", dispatches: 1, priced: 1, failed: 0, costUSD: 0.004, agentSeconds: 3, medianSeconds: 3,
      pairs: [{ pair: "sonnet · low", costUSD: 0.004, dispatches: 1 }] }],
    weeks: [{ week: "2026-11-18", roles: { "pair-judge": 0.004 } }],
  });
  render(<RoleCosts />);
  const table = await screen.findByRole("table", { name: "Spend and agent time per role" });
  const row = within(table).getByRole("row", { name: /pair-judge/ });
  expect(cells(row).slice(1, 3)).toEqual(["1", "<$0.01"]);
});
