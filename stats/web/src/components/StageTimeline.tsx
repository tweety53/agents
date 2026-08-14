// StageTimeline draws each stage run as a bar positioned by startedAt and
// endedAt across the change's own span, grouped by command -- one lane per
// command, following the dataviz skill's form heuristic for a duration
// across an ordered set of categories. It is presentation over the same
// rows StageRunTable already has: it issues no request of its own, so the
// timeline and the table can never disagree (task 18's own requirement).
//
// An open run (no endedAt) draws to the right edge of the chart and is
// marked as open (a hatched bar, a distinct data-testid) rather than being
// given a fabricated end -- the same absence-is-never-zero rule the table
// applies to its own metrics, applied here to "when did this run finish".
import type { StageRunDTO } from "../api";
import { formatDurationOrOpen } from "../format";

const LABEL_WIDTH = 140;
const LANE_WIDTH = 480;
const ROW_HEIGHT = 22;
const ROW_GAP = 10;
const BAR_HEIGHT = 14;

/** Commands in first-seen order, so lanes are stable and match the table's own row order. */
function commandOrder(stageRuns: StageRunDTO[]): string[] {
  const seen = new Set<string>();
  const order: string[] = [];
  for (const r of stageRuns) {
    if (!seen.has(r.command)) {
      seen.add(r.command);
      order.push(r.command);
    }
  }
  return order;
}

export function StageTimeline({ stageRuns }: { stageRuns: StageRunDTO[] }) {
  if (stageRuns.length === 0) return null;

  const now = Date.now();
  const starts = stageRuns.map((r) => new Date(r.startedAt).getTime());
  const ends = stageRuns.map((r) => (r.endedAt ? new Date(r.endedAt).getTime() : now));
  const spanStart = Math.min(...starts);
  // spanEnd is never <= spanStart: a single instantaneous run still gets a
  // visible, non-zero-width span to draw its bar against.
  const spanEnd = Math.max(spanStart + 1, ...ends);
  const spanMs = spanEnd - spanStart;

  const commands = commandOrder(stageRuns);
  const chartHeight = commands.length * (ROW_HEIGHT + ROW_GAP);
  const totalWidth = LABEL_WIDTH + LANE_WIDTH;

  function laneX(ms: number): number {
    return LABEL_WIDTH + ((ms - spanStart) / spanMs) * LANE_WIDTH;
  }

  return (
    <svg
      className="stage-timeline"
      role="img"
      aria-label="Stage runs across the change's span, grouped by command"
      viewBox={`0 0 ${totalWidth} ${chartHeight}`}
      width={totalWidth}
      height={chartHeight}
    >
      {commands.map((command, rowIndex) => {
        const y = rowIndex * (ROW_HEIGHT + ROW_GAP);
        const runsForCommand = stageRuns.filter((r) => r.command === command);
        return (
          <g key={command}>
            <text x={0} y={y + BAR_HEIGHT} className="stage-timeline-label">
              {command}
            </text>
            {runsForCommand.map((r) => {
              const open = !r.endedAt;
              const startX = laneX(new Date(r.startedAt).getTime());
              const endX = open ? LABEL_WIDTH + LANE_WIDTH : laneX(new Date(r.endedAt as string).getTime());
              const width = Math.max(2, endX - startX);
              return (
                <rect
                  key={r.stageRunId}
                  x={startX}
                  y={y}
                  width={width}
                  height={BAR_HEIGHT}
                  rx={3}
                  className={open ? "stage-timeline-bar-open" : "stage-timeline-bar"}
                  data-testid={open ? "stage-timeline-open" : "stage-timeline-bar"}
                >
                  <title>{`${r.stage} (attempt ${r.attempt}): ${formatDurationOrOpen(r.startedAt, r.endedAt)}${
                    open ? "" : ` -- ${r.outcome ?? "unknown outcome"}`
                  }`}</title>
                </rect>
              );
            })}
          </g>
        );
      })}
    </svg>
  );
}
