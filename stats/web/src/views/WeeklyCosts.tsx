// Weekly costs -- where does flow spend go, week by week? Weeks run
// Wednesday 00:00 UTC to the next Wednesday. The table is always the last
// 8 weeks, the current partial week included, whatever the period picker
// says; the chart follows the picker in the same weekly buckets. A
// category with no costed run that week is Unavailable in the table and a
// gap in the chart, never $0.
import { useMemo } from "react";
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { TimeSeriesPanel } from "../components/TimeSeriesPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { WeeklyCostPoint } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const DAY_MS = 24 * 60 * 60 * 1000;
const WEEKS = 8;

const categories: { key: keyof Omit<WeeklyCostPoint, "week">; name: string }[] = [
  { key: "overallUsd", name: "Overall" },
  { key: "planningUsd", name: "Planning" },
  { key: "implementationUsd", name: "Implementation" },
  { key: "reviewUsd", name: "Review" },
  { key: "selfReviewUsd", name: "Self-review" },
];

const columns: Column<WeeklyCostPoint>[] = [
  { key: "week", header: "Week", sortable: true, accessor: (r) => r.week },
  ...categories.map(
    ({ key, name }): Column<WeeklyCostPoint> => ({
      key,
      header: name,
      sortable: true,
      accessor: (r) => r[key],
      render: (r) => <Unavailable value={r[key]} format={formatUsd} />,
    }),
  ),
];

/** The last 8 Wednesday-to-Wednesday weeks: from this week's UTC
 * Wednesday 00:00 minus 7 weeks, to now. */
export function lastEightWeeks(now: Date): { from: Date; to: Date } {
  const midnight = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
  const sinceWednesday = (now.getUTCDay() + 4) % 7;
  return { from: new Date(midnight - (sinceWednesday + (WEEKS - 1) * 7) * DAY_MS), to: now };
}

/** Every week of `from`'s 8-week window, in order: the response's row where
 * it carries one, an all-null row where no costed run fell in that week. */
function everyWeek(from: Date, rows: WeeklyCostPoint[]): WeeklyCostPoint[] {
  const byWeek = new Map(rows.map((r) => [r.week, r]));
  return Array.from({ length: WEEKS }, (_, i) => {
    const week = new Date(from.getTime() + i * 7 * DAY_MS).toISOString().slice(0, 10);
    return (
      byWeek.get(week) ?? {
        week,
        overallUsd: null,
        planningUsd: null,
        implementationUsd: null,
        reviewUsd: null,
        selfReviewUsd: null,
      }
    );
  });
}

export function WeeklyCosts({ period, project, model }: ViewProps) {
  const eightWeeks = useMemo(() => lastEightWeeks(new Date()), []);
  const table = useStatsView<WeeklyCostPoint[]>("weekly-costs", { ...eightWeeks, project, model });
  const chart = useStatsView<WeeklyCostPoint[]>("weekly-costs", { from: period.from, to: period.to, project, model });

  return (
    <ViewFrame title="Weekly costs" description="Wednesday-to-Wednesday cost by pipeline phase.">
      <Panel title="Last 8 weeks" description="Overall includes stages outside the four phases." state={table}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={everyWeek(eightWeeks.from, data.rows)}
            rowKey={(r) => r.week}
          />
        )}
      </Panel>
      <Panel title="Weekly cost" description="Each week's cost by phase over the selected period -- a week the period cuts counts only its days inside it." state={chart}>
        {(data) => (
          <TimeSeriesPanel
            series={categories.map(({ key, name }) => ({
              name,
              points: data.rows.map((r) => ({ day: r.week, value: r[key] })),
            }))}
            format={formatUsd}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
