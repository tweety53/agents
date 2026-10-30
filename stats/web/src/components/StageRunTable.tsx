// StageRunTable is the opened run's table of stage runs, laid out as the
// approved prototype (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// no filter controls, stage runs grouped by session under a `Session N ·
// <phase>` row carrying the session's span, duration and cost
// (session-phase-groups), each row's model shown with its version and effort
// (model-version-effort), and a row's dispatches opening beneath it. It is a
// plain table, not DataTable, because DataTable cannot draw group rows under
// one header (grouped-tables-own-markup).
// A /flow-interactive session's table instead leads with a Request column
// and lists its runs newest first, ungrouped (the `requests` prop, KAN-1061).
//
// Every reader from metrics.ts is routed through Unavailable, so an
// unmeasured metric on an otherwise-measured run reads as unavailable, not
// zero. Duration is computed from startedAt/endedAt directly; an open run
// runs to now and reads as running, never as zero-length.
import { Fragment, useState, type ReactNode } from "react";
import { Unavailable } from "./Unavailable";
import { outcomeClass } from "./StageTimeline";
import type { DispatchRow, RecordDispatch, StageRunDTO } from "../api";
import {
  asObject,
  readCostUsd,
  readEffort,
  readMainModels,
  readModels,
  readTotalInputTokens,
  readMainTokens,
  readServedEffort,
  readServedModel,
  readSidechainTokens,
} from "../metrics";
import { formatClock, formatModel, formatSpan, formatUsd } from "../format";
import { stageLabel } from "../lib/stageLabel";
import { STAGE_PHASE } from "../lib/phase";
import { describeDispatch } from "../lib/describeDispatch";

// --- sessions and phases (design.md's session-phase-groups) ---

/** A stage's phase, from the run view's one stage-to-phase mapping (lib/phase.ts). */
function phaseOf(stage: string): string {
  return STAGE_PHASE[stage] ?? "Other";
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

/** Distinct models compactly, families sharing a version joined: ["claude-opus-5-5", "claude-sonnet-5-5"] reads "opus+sonnet 5.5". */
export function compactModels(models: string[]): string {
  const byVersion = new Map<string, string[]>();
  for (const label of new Set(models.map(formatModel))) {
    const [fam, ver = ""] = label.split(" ");
    byVersion.set(ver, [...(byVersion.get(ver) ?? []), fam]);
  }
  return [...byVersion].map(([ver, fams]) => `${fams.join("+")}${ver ? ` ${ver}` : ""}`).join(", ");
}

function modelEffortText(models: string[], effort: string | null): string {
  const label = compactModels(models);
  return effort === null ? label : `${label} · ${effort}`;
}

// --- per-dispatch cost (KAN-201) ---

type DispatchLine = DispatchRow & { effort: string | null; inputTokens: number | null };

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
      // The exact model -- its recorded version, else the one most served -- before the bare family.
      model: d && typeof d.modelVersion === "string" ? d.modelVersion : (readServedModel(d) ?? (d && typeof d.model === "string" ? d.model : undefined)),
      tokens: tokenTotal(d),
      costUsd: (d ? readCostUsd(d) : null) ?? undefined,
      effort: readServedEffort(d),
      inputTokens: readTotalInputTokens(d),
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

/** The harness description's own words after its leading dispatch key ("<key> minor fixes" reads "minor fixes"). */
function descriptionWords(d: DispatchRow): string | undefined {
  const m = d.description ? /^\S+-\S+\s+(.+)$/.exec(d.description) : null;
  return m ? m[1] : d.description;
}

/** A dispatch's plain words: its recorded key -- else the key its harness description leads with -- through describeDispatch, else the description's own words. */
function dispatchName(d: DispatchRow, rec: RecordDispatch | undefined): string {
  if (rec) return describeDispatch(rec.key || undefined, rec.role);
  const key = d.description?.split(/\s/)[0];
  const named = key ? describeDispatch(key, d.agentType ?? "") : key;
  return named && named !== key ? named : (descriptionWords(d) ?? d.agentType ?? "dispatch");
}

const capitalize = (t: string) => t.charAt(0).toUpperCase() + t.slice(1);

/** A dispatch's plain words with its stage's own prefix dropped ("Self-review: fix round 2" reads "Fix round 2" in a Self-review run). */
function actionName(run: StageRunDTO, d: DispatchRow, rec: RecordDispatch | undefined): string {
  return capitalize(dispatchName(d, rec).replace(`${stageLabel(run.stage)}: `, ""));
}

/** Each action's "<prefix>: " left out where the one before it carries the same: "Code review round 2: primary", "Code review round 2: test sanity" read "Code review round 2: primary", "test sanity". */
export function foldPrefixes(names: string[]): string[] {
  const prefix = (n: string) => /^(.+?): /.exec(n)?.[1];
  return names.map((n, i) => {
    const p = prefix(n);
    return i > 0 && p !== undefined && p === prefix(names[i - 1]) ? n.slice(p.length + 2) : n;
  });
}

/** A row's agents' models, distinct in the order its actions ran, each counted past one: "opus 5.5 · low ×3, sonnet 5.5 · low". */
function agentModels(own: DispatchLine[]): string {
  const counts = new Map<string, number>();
  for (const d of own) {
    const m = d.model ? modelEffortText([d.model], d.effort) : "unknown";
    counts.set(m, (counts.get(m) ?? 0) + 1);
  }
  return [...counts].map(([m, n]) => (n > 1 ? `${m} ×${n}` : m)).join(", ");
}

/** A dispatch that began before its run did: the tail of an earlier run's dispatch, whose last turns landed here -- its cost stays in this run, its row does not. */
function isTail(run: StageRunDTO, rec: RecordDispatch | undefined): boolean {
  return rec !== undefined && Date.parse(rec.startedAt) < Date.parse(run.startedAt) - 1000;
}

/** A dispatch's outcome in the words its work earns: a review's clean or findings, a fix's fixed, else its recorded outcome. */
function dispatchOutcome(rec: RecordDispatch | undefined): string | undefined {
  if (!rec) return undefined;
  if (!rec.endedAt) return "running";
  if (rec.outcome === "fix") return "findings";
  if ((rec.outcome ?? "completed") === "completed" && /fix/.test(rec.key ?? "")) return "fixed";
  return rec.outcome ?? "completed";
}

const outcomeLook = (o: string) => (o === "running" ? "open" : o === "failed" || o === "aborted" ? "failed" : "completed");

/**
 * One stage run: named by what its own dispatches did, each with its model and
 * effort ("Fix round 2 (opus 5.5 · low) → Review round 2 (sonnet 5.5 · low)"),
 * led by its stage unless its session runs that stage alone; its outcome the
 * last dispatch's; its cost the run's whole, split agents + main session on
 * hover. Two or more own dispatches open beneath it in time order, attached
 * to it as one tinted group, sliding in and out in CSS (closing keeps them
 * mounted until their closing animation ends); one dispatch is the row itself.
 */
function StageRunRow({ run, now, records, showStage, requests }: { run: StageRunDTO; now: number; records: Map<string, RecordDispatch>; showStage: boolean; requests: boolean }) {
  const [open, setOpen] = useState(false);
  const [closing, setClosing] = useState(false);
  const toggle = () => {
    setClosing(open);
    setOpen(!open);
  };
  const all = readDispatchRows(run.metrics);
  const at = (d: DispatchRow) => Date.parse(records.get(d.agentId)?.startedAt ?? run.startedAt);
  const own = all.filter((d) => !isTail(run, records.get(d.agentId))).sort((a, b) => at(a) - at(b));
  const stage = stageLabel(run.stage);
  const actions = foldPrefixes(own.map((d) => actionName(run, d, records.get(d.agentId)))).join(" → ");
  // No stage lead where the session runs this stage alone, or where the first action already names it.
  const name = !actions ? stage : showStage && !actions.startsWith(stage) ? `${stage} · ${actions}` : actions;
  const expandable = own.length > 1;
  const stageCost = readCostUsd(run.metrics);
  // The agents' share is every dispatch bucket the run carries, tails included; the rest is the main session's own turns.
  const agents = all.length > 0 && all.every((d) => d.costUsd !== undefined) ? all.reduce((a, d) => a + (d.costUsd ?? 0), 0) : null;
  const split = stageCost !== null && agents !== null ? `agents ${formatUsd(agents)} + main session ${formatUsd(Math.max(0, stageCost - agents))}` : undefined;
  const tokensIn = readTotalInputTokens(run.metrics);
  const agentTokens = all.length > 0 ? all.reduce((a, d) => a + (d.inputTokens ?? 0), 0) : null;
  const tokenSplit =
    tokensIn !== null && agentTokens !== null ? `agents ${formatKTokens(agentTokens)} + main session ${formatKTokens(Math.max(0, tokensIn - agentTokens))}` : undefined;
  const outcome = outcomeClass(run);
  const last = own.length ? dispatchOutcome(records.get(own[own.length - 1].agentId)) : undefined;
  const word = run.outcome === "failed" ? "failed" : (last ?? run.outcome ?? "completed");
  return (
    <>
      <tr className={`lb-row${expandable ? " expandable" : ""}${open ? " open" : ""}`} onClick={() => expandable && toggle()}>
        {requests && <td className="lb-rq">{run.request === undefined ? "" : `R${run.request}`}</td>}
        <td title={name}>
          {expandable ? (
            <button
              type="button"
              className="lb-expander"
              aria-expanded={open}
              aria-label={`${open ? "Hide" : "Show"} dispatches`}
              onClick={(e) => {
                e.stopPropagation();
                toggle();
              }}
            >
              <span className="lb-chev" aria-hidden="true" />
            </button>
          ) : (
            <span className="lb-expander" />
          )}
          {name}
        </td>
        <td>
          {outcome === "open" ? (
            <span className="lb-out open" data-testid="run-open">
              running
            </span>
          ) : (
            <span className={`lb-out ${outcome === "failed" ? "failed" : outcomeLook(word)}`}>{word}</span>
          )}
        </td>
        <td className="r lb-num">{formatClock(Date.parse(run.startedAt))}</td>
        <td className="r lb-num">{formatSpan(runEnd(run, now) - Date.parse(run.startedAt))}</td>
        <td className="r lb-num" title={tokenSplit} data-testid="run-tokens">
          <Unavailable value={tokensIn} format={formatKTokens} />
        </td>
        <td data-testid="run-model">
          {own.length > 0 ? (
            <span title={agentModels(own)}>{agentModels(own)}</span>
          ) : (
            // No agent: the main session's own model.
            (() => {
              const models = readMainModels(run.metrics) ?? readModels(run.metrics);
              return models ? (
                <span className="lb-muted" title="main session">
                  {modelEffortText(models, readEffort(run.metrics))}
                </span>
              ) : (
                <Unavailable value={null} />
              );
            })()
          )}
        </td>
        <td className="r lb-num" title={split} data-testid="run-cost">
          <Unavailable value={stageCost} format={formatUsd} />
        </td>
      </tr>
      {(open || closing) &&
        own.map((d, i) => {
          const rec = records.get(d.agentId);
          const status = dispatchOutcome(rec);
          const names = foldPrefixes(own.map((x) => actionName(run, x, records.get(x.agentId))));
          // Each cell's content sits in a div the open/close animation sizes; a td's own height cannot animate.
          const cell = (content: ReactNode) => <div className="lb-cell">{content}</div>;
          return (
            <tr
              key={d.agentId}
              className={`lb-dispatch ${open ? "in" : "out"}${i === own.length - 1 ? " last" : ""}`}
              data-testid={open ? "dispatch-row" : undefined}
              onAnimationEnd={() => setClosing(false)}
            >
              {requests && <td>{cell(null)}</td>}
              <td className="lb-dispatch-name" title={`agent ${d.agentId}`}>
                {cell(
                  <>
                    <span className="lb-agent-branch" aria-hidden="true">
                      ⎿{" "}
                    </span>
                    {d.description ?? names[i]}
                  </>,
                )}
              </td>
              <td>{cell(status ? <span className={`lb-out ${outcomeLook(status)}`}>{status}</span> : <Unavailable value={null} />)}</td>
              <td className="r lb-num">{cell(rec ? formatClock(Date.parse(rec.startedAt)) : <Unavailable value={null} />)}</td>
              <td className="r lb-num">{cell(rec ? formatSpan((rec.endedAt ? Date.parse(rec.endedAt) : now) - Date.parse(rec.startedAt)) : <Unavailable value={null} />)}</td>
              <td className="r lb-num">{cell(<Unavailable value={d.inputTokens} format={formatKTokens} />)}</td>
              <td>{cell(d.model ? modelEffortText([d.model], d.effort) : <Unavailable value={null} />)}</td>
              <td className="r lb-num">{cell(<Unavailable value={d.costUsd ?? null} format={formatUsd} />)}</td>
            </tr>
          );
        })}
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
  /** The change's recorded dispatches, which give a dispatch row its key, outcome and timing. */
  records?: RecordDispatch[];
  /** A /flow-interactive session's table (KAN-1061): a leading Request column reading R<n>, rows newest first with no session group rows. */
  requests?: boolean;
}

export function StageRunTable({ stageRuns, truncated, totalStageRuns, records = [], requests = false }: StageRunTableProps) {
  const now = Date.now();
  const columns = requests ? 8 : 7;
  const byAgent = new Map(records.filter((d) => d.agentId).map((d) => [d.agentId!, d]));
  return (
    <>
      {truncated && (
        <p role="status" className="stage-run-table-truncated">
          Showing {stageRuns.length} of {totalStageRuns} stage runs recorded for this change — narrow the
          period or view the change directly for the rest.
        </p>
      )}
      <div className="lb-scroll">
        <table className="lb-table lb-runs">
          <colgroup>
            {requests && <col className="c-rq" />}
            <col />
            <col className="c-out" />
            <col className="c-start" />
            <col className="c-dur" />
            <col className="c-tok" />
            <col className="c-model" />
            <col className="c-cost" />
          </colgroup>
          <thead>
            <tr>
              {requests && <th>Request</th>}
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
                <td colSpan={columns} className="lb-muted">
                  No stage runs recorded for this change.
                </td>
              </tr>
            )}
            {requests
              ? [...stageRuns]
                  .sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))
                  .map((r) => <StageRunRow key={r.stageRunId} run={r} now={now} records={byAgent} showStage={false} requests />)
              : groupSessions(stageRuns).map((s) => (
                  <Fragment key={s.label}>
                    <SessionRow session={s} now={now} />
                    {s.runs.map((r) => (
                      <StageRunRow key={r.stageRunId} run={r} now={now} records={byAgent} showStage={new Set(s.runs.map((x) => x.stage)).size > 1} requests={false} />
                    ))}
                  </Fragment>
                ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
