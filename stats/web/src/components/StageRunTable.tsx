// StageRunTable is the opened run's table of stage runs, laid out as the
// approved prototype (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// no filter controls, stage runs grouped by session under a `Session N ·
// <phase>` row carrying the session's span, duration and cost
// (session-phase-groups), each row's model shown with its version and effort
// (model-version-effort), and a row's dispatches opening beneath it. It is a
// plain table, not DataTable, because DataTable cannot draw group rows under
// one header (grouped-tables-own-markup).
//
// Every reader from metrics.ts is routed through Unavailable, so an
// unmeasured metric on an otherwise-measured run reads as unavailable, not
// zero. Duration is computed from startedAt/endedAt directly; an open run
// runs to now and reads as running, never as zero-length.
import { Fragment, useState } from "react";
import { Unavailable } from "./Unavailable";
import { outcomeClass } from "./StageTimeline";
import type { DispatchRow, StageRunDTO } from "../api";
import {
  asObject,
  readCostUsd,
  readEffort,
  readTotalInputTokens,
  readMainTokens,
  readModels,
  readServedEffort,
  readSidechainTokens,
} from "../metrics";
import { formatClock, formatModel, formatSpan, formatUsd } from "../format";
import { stageLabel } from "../lib/stageLabel";

// --- sessions and phases (design.md's session-phase-groups) ---

const PHASES: Record<string, string> = Object.fromEntries([
  ...["kickoff", "brainstorm", "design-approval", "create-artifacts", "writing-plans", "decide", "document-fix"].map((s) => [s, "planning"]),
  ...["load-context", "isolate-workspace", "sdd-tdd", "implement", "review-panel", "verify", "visual-verify", "run-instructions", "write-in-progress"].map(
    (s) => [s, "implementation"],
  ),
  ...[
    "preflight",
    "unfinished-work-gate",
    "self-review",
    "stage-diff",
    "commit-two",
    "commit-archive",
    "landing-question",
    "landing-routes",
    "integrate",
    "cleanup",
    "sync-archive",
    "verify-merge",
    "verify-cleanup",
    "refresh-main-checkout",
    "preserve-sessions",
    "write-finished",
  ].map((s) => [s, "landing"]),
]);

function phaseOf(stage: string): string {
  return PHASES[stage.replace(/^flow\./, "")] ?? "other";
}

export interface Session {
  label: string;
  runs: StageRunDTO[];
}

/** Stage runs grouped by session id (the command when a run has none), groups and runs in start order. */
export function groupSessions(stageRuns: StageRunDTO[]): Session[] {
  const byStart = [...stageRuns].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
  const groups = new Map<string, StageRunDTO[]>();
  for (const r of byStart) {
    const key = r.sessionId ?? r.command;
    groups.set(key, [...(groups.get(key) ?? []), r]);
  }
  return [...groups.values()].map((runs, i) => ({ label: `Session ${i + 1}`, runs }));
}

export function runEnd(run: StageRunDTO, now: number): number {
  return run.endedAt ? Date.parse(run.endedAt) : now;
}

function sumCost(runs: StageRunDTO[]): number {
  return runs.reduce((a, r) => a + (readCostUsd(r.metrics) ?? 0), 0);
}

function SessionRow({ session, now }: { session: Session; now: number }) {
  const { runs } = session;
  const t0 = Math.min(...runs.map((r) => Date.parse(r.startedAt)));
  const open = runs.some((r) => !r.endedAt);
  const t1 = Math.max(...runs.map((r) => runEnd(r, now)));
  const phases = [...new Set(runs.map((r) => phaseOf(r.stage)))].join(" + ");
  return (
    <tr className="lb-group">
      <td colSpan={7}>
        {session.label} · {phases}
        <span className="lb-cnt">
          {formatClock(t0)}–{open ? "now" : formatClock(t1)} · {formatSpan(t1 - t0)} · {formatUsd(sumCost(runs))}
        </span>
      </td>
    </tr>
  );
}

// --- model, version and effort (design.md's model-version-effort) ---

function modelEffort(models: string[] | null, effort: string | null) {
  if (models === null) return <Unavailable value={null} />;
  const label = models.map(formatModel).join(", ");
  return <span data-testid="measured">{effort === null ? label : `${label} · ${effort}`}</span>;
}

// --- per-dispatch cost (KAN-201) ---

type DispatchLine = DispatchRow & { effort: string | null };

/** A bucket's own token grand total (main + sidechain), or undefined when neither bucket carried a chargeable field. */
function tokenTotal(bag: unknown): number | undefined {
  const main = readMainTokens(bag);
  const sidechain = readSidechainTokens(bag);
  if (main === null && sidechain === null) return undefined;
  return (main ?? 0) + (sidechain ?? 0);
}

/**
 * One stage run's dispatch rows from its metrics bag's `dispatches.<agentId>`
 * key (internal/harvest.DispatchBucket), sorted by cost descending with an
 * absent cost last -- never treated as zero. Per-dispatch cost is that key's
 * own `cost_usd`, which store.Store.Price writes through the same pricing
 * path as every other cost figure; a dispatch with none reads unavailable,
 * never a fabricated or derived cost.
 */
function readDispatchRows(bag: unknown): DispatchLine[] {
  const obj = asObject(bag);
  const dispatches = obj ? asObject(obj.dispatches) : null;
  if (!dispatches) return [];

  const rows: DispatchLine[] = Object.entries(dispatches).map(([agentId, raw]) => {
    const d = asObject(raw);
    return {
      agentId,
      agentType: d && typeof d.agent_type === "string" ? d.agent_type : undefined,
      description: d && typeof d.description === "string" ? d.description : undefined,
      model: d && typeof d.model === "string" ? d.model : undefined,
      tokens: tokenTotal(d),
      costUsd: (d ? readCostUsd(d) : null) ?? undefined,
      effort: readServedEffort(d),
    };
  });

  return rows.sort((a, b) => {
    if (a.costUsd === undefined && b.costUsd === undefined) return 0;
    if (a.costUsd === undefined) return 1;
    if (b.costUsd === undefined) return -1;
    return b.costUsd - a.costUsd;
  });
}

function formatKTokens(v: number): string {
  return `${(v / 1e3).toFixed(0)}k`;
}

function StageRunRow({ run, now }: { run: StageRunDTO; now: number }) {
  const [open, setOpen] = useState(false);
  const dispatches = readDispatchRows(run.metrics);
  const outcome = outcomeClass(run);
  return (
    <>
      <tr className="lb-row" onClick={() => dispatches.length > 0 && setOpen(!open)}>
        <td>
          {dispatches.length > 0 ? (
            <button
              type="button"
              className="lb-expander"
              aria-expanded={open}
              aria-label={`${open ? "Hide" : "Show"} dispatches`}
              onClick={(e) => {
                e.stopPropagation();
                setOpen(!open);
              }}
            >
              {open ? "▾" : "▸"}
            </button>
          ) : (
            <span className="lb-expander" />
          )}
          {stageLabel(run.stage)}
          {run.autoOpened && (
            <span className="lb-muted" title="opened automatically: the run marked no stage">
              {" "}
              (unmarked)
            </span>
          )}
          {stageLabel(run.stage) !== run.stage && <div className="lb-muted lb-stage-id">{run.stage}</div>}
        </td>
        <td>
          {outcome === "open" ? (
            <span className="lb-out open" data-testid="run-open">
              running
            </span>
          ) : (
            <span className={`lb-out ${outcome}`}>{run.outcome ?? "completed"}</span>
          )}
        </td>
        <td className="r lb-num">{formatClock(Date.parse(run.startedAt))}</td>
        <td className="r lb-num">{formatSpan(runEnd(run, now) - Date.parse(run.startedAt))}</td>
        <td className="r lb-num">
          <Unavailable value={readTotalInputTokens(run.metrics)} format={formatKTokens} />
        </td>
        <td>{modelEffort(readModels(run.metrics), readEffort(run.metrics))}</td>
        <td className="r lb-num">
          <Unavailable value={readCostUsd(run.metrics)} format={formatUsd} />
        </td>
      </tr>
      {open &&
        dispatches.map((d) => (
          <tr key={d.agentId} className="lb-dispatch" data-testid="dispatch-row">
            <td className="lb-dispatch-name">{d.description ?? d.agentType ?? d.agentId}</td>
            <td colSpan={4} className="lb-muted">
              {d.agentId}
            </td>
            <td>{modelEffort(d.model === undefined ? null : [d.model], d.effort)}</td>
            <td className="r lb-num">
              <Unavailable value={d.costUsd ?? null} format={formatUsd} />
            </td>
          </tr>
        ))}
    </>
  );
}

export interface StageRunTableProps {
  stageRuns: StageRunDTO[];
  /** True when the change has more stage runs than the fetched page held
   * (useRunDetail's own `truncated`) -- task 25, step 4: rather than
   * silently rendering a short list that looks complete, this states the
   * shortfall directly above the table. */
  truncated?: boolean;
  /** The server's own total stage-run count for this change, reported
   * alongside `truncated`. */
  totalStageRuns?: number;
}

export function StageRunTable({ stageRuns, truncated, totalStageRuns }: StageRunTableProps) {
  const now = Date.now();
  return (
    <>
      {truncated && (
        <p role="status" className="stage-run-table-truncated">
          Showing {stageRuns.length} of {totalStageRuns} stage runs recorded for this change — narrow the
          period or view the change directly for the rest.
        </p>
      )}
      <div className="lb-scroll">
        <table className="lb-table">
          <thead>
            <tr>
              <th>Stage</th>
              <th>Outcome</th>
              <th className="r">Started</th>
              <th className="r">Duration</th>
              <th className="r">Tokens in</th>
              <th>Model · effort</th>
              <th className="r">Cost</th>
            </tr>
          </thead>
          <tbody>
            {stageRuns.length === 0 && (
              <tr>
                <td colSpan={7} className="lb-muted">
                  No stage runs recorded for this change.
                </td>
              </tr>
            )}
            {groupSessions(stageRuns).map((s) => (
              <Fragment key={s.label}>
                <SessionRow session={s} now={now} />
                {s.runs.map((r) => (
                  <StageRunRow key={r.stageRunId} run={r} now={now} />
                ))}
              </Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
