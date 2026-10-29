// What a running self-review is doing right now, derived from its dispatch keys
// (skills/flow-self-review/SKILL.md: self-review-<name>-pass, -plan, -fix /
// -fix-g<g> / -fix-<r>, -review-<r>, -fix-v<v>) and its plan's ticks. Landing
// and filing run in the main session with no dispatch of their own, so they
// are never "now" -- only done once the stage closes.
import type { PlanTask } from "../api";
import { formatModel, formatSpan } from "../format";

export const SR_STEPS = ["review", "plan", "fix", "re-review", "verify", "land", "file"] as const;
export type SrStep = (typeof SR_STEPS)[number];

export interface SrDispatch {
  key?: string;
  role: string;
  taskId?: string;
  model: string;
  modelVersion?: string;
  startedAt: string;
  endedAt?: string;
}

/** The procedure step a self-review dispatch key names, with its round where the key carries one; null for any other key. */
export function srStep(key: string | undefined): { step: SrStep; round?: number } | null {
  if (!key?.startsWith("self-review-")) return null;
  let m: RegExpExecArray | null;
  if (/-pass$/.test(key)) return { step: "review" };
  if (/-plan(?:-\d+)?$/.test(key)) return { step: "plan" };
  if ((m = /-fix-v(\d+)$/.exec(key))) return { step: "verify", round: Number(m[1]) };
  if ((m = /-review-(\d+)$/.exec(key))) return { step: "re-review", round: Number(m[1]) };
  if ((m = /-fix(?:-g\d+|-(\d+))?$/.exec(key))) return m[1] ? { step: "fix", round: Number(m[1]) } : { step: "fix" };
  return null;
}

export interface SrNow {
  /** The "Now:" line. */
  text: string;
  /** The step the progress line highlights; null before any dispatch and once done. */
  current: SrStep | null;
  done: boolean;
  /** The plan task a running fixer works on. */
  taskN?: number;
}

/** The self-review's current step: its running dispatch, else its latest one with main between steps; "Done" once the stage closed. */
export function selfReviewNow(dispatches: SrDispatch[], tasks: PlanTask[] | undefined, stageOpen: boolean, now: number): SrNow {
  if (!stageOpen) return { text: "Done", current: null, done: true };
  const mine = dispatches.filter((d) => srStep(d.key)).sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
  const running = mine.filter((d) => !d.endedAt).pop();
  const d = running ?? mine[mine.length - 1];
  if (!d) return { text: "Main session is between steps (nothing dispatched yet)", current: null, done: false };
  const info = srStep(d.key)!;
  const taskN =
    info.step === "fix" && info.round === undefined
      ? d.taskId
        ? Number(d.taskId)
        : tasks?.find((t) => t.state !== "done")?.n
      : undefined;
  const task = taskN === undefined ? undefined : tasks?.find((t) => t.n === taskN);
  const label =
    info.step === "review"
      ? "Reviewing the change"
      : info.step === "plan"
        ? "Planning fixes"
        : info.step === "fix"
          ? task
            ? `Fixing task ${task.n} "${task.title}"`
            : info.round !== undefined
              ? `Fixing the review's findings, round ${info.round}`
              : "Fixing"
          : info.step === "re-review"
            ? `Reviewing the fixes, round ${info.round}`
            : `Verifying — red verify fix round ${info.round}`;
  if (running) {
    return { text: `${label} · for ${formatSpan(now - Date.parse(d.startedAt))} · ${formatModel(d.modelVersion ?? d.model)}`, current: info.step, done: false, taskN };
  }
  return { text: `Main session is between steps (last: ${label}, ${formatSpan(now - Date.parse(d.endedAt!))} ago)`, current: info.step, done: false };
}
