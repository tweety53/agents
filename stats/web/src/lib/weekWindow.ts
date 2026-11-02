import { DAY_MS } from "../format";

const WEEK_MS = 7 * DAY_MS;

// The n Wednesday-00:00-UTC weeks ending with the current one, from that first
// Wednesday to now; `weeks` names each week by its Wednesday's date.
export function weekWindow(n: number, now: number): { from: Date; to: Date; weeks: string[] } {
  const day = Math.floor(now / DAY_MS) * DAY_MS;
  const current = day - ((new Date(day).getUTCDay() + 4) % 7) * DAY_MS;
  const weeks = Array.from({ length: n }, (_, i) => new Date(current - (n - 1 - i) * WEEK_MS).toISOString().slice(0, 10));
  return { from: new Date(current - (n - 1) * WEEK_MS), to: new Date(now), weeks };
}
