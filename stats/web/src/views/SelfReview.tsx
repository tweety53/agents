// SelfReview -- what the self-review pass did, in two read views: Fixed (every
// finding it landed a fix for, with its full detail, read only) and To review
// (the improvements it suggests, each approved or declined here; a Status chip
// keeps decided ones out of the way). Its Period and Project chips are its own,
// like the other dashboards'. Tasks are the operator's to create and plan.
import { useEffect, useState } from "react";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { ViewFrame } from "../components/ViewFrame";
import { fetchProjects, reviewSelfReviewFinding, type SelfReviewRow } from "../api";
import { PERIOD_DAYS, windowFor, type PeriodChip } from "../format";
import { projectLabel } from "../lib/projectLabel";
import { useStatsView } from "../hooks/useStatsView";

type Show = "fixed" | "to-review";
type Status = "to-review" | "approved" | "declined";
const PERIOD_LABELS: Record<PeriodChip, string> = { "7d": "7 days", "30d": "30 days", "90d": "90 days" };
const STATUS_LABEL: Record<Status, string> = { "to-review": "To review", approved: "Approved", declined: "Declined" };

export function SelfReview() {
  const [win, setWin] = useState(() => windowFor("30d"));
  const [project, setProject] = useState<string | undefined>(undefined);
  const [projects, setProjects] = useState<string[]>([]);
  const [show, setShow] = useState<Show>("fixed");
  const [status, setStatus] = useState<Status>("to-review");
  // A decision made here shows at once; the next load reads it from the store.
  const [decided, setDecided] = useState<Record<number, Status>>({});
  const [failed, setFailed] = useState("");
  const state = useStatsView<SelfReviewRow[]>("self-review", { from: win.from, to: win.to, project });

  useEffect(() => {
    fetchProjects().then(setProjects, () => setProjects([]));
  }, []);

  const decide = (row: SelfReviewRow, to: Status) =>
    reviewSelfReviewFinding(row.project, row.id, to).then(
      () => {
        setFailed("");
        setDecided((d) => ({ ...d, [row.id]: to }));
      },
      (err: unknown) => setFailed(err instanceof Error ? err.message : String(err)),
    );

  const suggestions = (rows: SelfReviewRow[]) =>
    rows.filter((r) => r.disposition === "suggested").map((r) => ({ row: r, status: decided[r.id] ?? ((r.reviewStatus || "to-review") as Status) }));

  return (
    <ViewFrame title="Self-review" description="What each self-review pass fixed, and the improvements it suggests for you to approve or decline.">
      <div className="lb-fbar">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={win.period === p} onClick={() => setWin(windowFor(p))}>
              {PERIOD_LABELS[p]}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={project === undefined} onClick={() => setProject(undefined)}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p} on={project === p} onClick={() => setProject(p)} title={p}>
              {projectLabel(p)}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Show">
          <Chip on={show === "fixed"} onClick={() => setShow("fixed")}>
            Fixed
          </Chip>
          <Chip on={show === "to-review"} onClick={() => setShow("to-review")}>
            To review
          </Chip>
        </ChipGroup>
        {show === "to-review" && state.status === "ready" && (
          <ChipGroup name="Status">
            {(Object.keys(STATUS_LABEL) as Status[]).map((s) => (
              <Chip key={s} on={status === s} onClick={() => setStatus(s)}>
                {STATUS_LABEL[s]} {suggestions(state.data.rows).filter((x) => x.status === s).length}
              </Chip>
            ))}
          </ChipGroup>
        )}
      </div>
      {show === "fixed" ? (
        <Panel title="Fixed" state={state}>
          {(data) => {
            const fixed = data.rows.filter((r) => r.disposition === "fixed");
            if (fixed.length === 0) return <p className="lb-muted">No fixed findings in this period.</p>;
            return (
              <div className="lb-scroll">
              <table className="lb-table sr-table">
                <thead>
                  <tr>
                    <th>Recorded</th>
                    <th>Change</th>
                    <th>Angle</th>
                    <th>Severity</th>
                    <th>Finding</th>
                    <th>Fix</th>
                  </tr>
                </thead>
                <tbody>
                  {fixed.map((r) => (
                    <tr key={r.id}>
                      <td>{r.recordedAt.slice(0, 10)}</td>
                      <td>{r.change}</td>
                      <td>{r.angle}</td>
                      <td>{r.severity}</td>
                      <td className="sr-wrap">
                        <div>{r.note}</div>
                        {r.detail && (
                          <details className="sr-detail">
                            <summary>Details</summary>
                            <pre>{r.detail}</pre>
                          </details>
                        )}
                      </td>
                      <td>{r.commitUrl ? <a href={r.commitUrl}>{r.ref}</a> : r.ref}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              </div>
            );
          }}
        </Panel>
      ) : (
        <Panel title="To review" state={state}>
          {(data) => {
            const shown = suggestions(data.rows).filter((s) => s.status === status);
            return (
              <>
                {failed && <p className="sr-err">Could not save: {failed}</p>}
                {shown.length === 0 && <p className="lb-muted">Nothing {STATUS_LABEL[status].toLowerCase()}.</p>}
                {shown.map(({ row }) => (
                  <article key={row.id} className="sr-card" data-testid="suggestion">
                    <header>
                      <strong>{row.note}</strong>
                      <span className="lb-muted">
                        {row.change} · {row.angle}
                        {row.severity && ` · ${row.severity}`} · {row.recordedAt.slice(0, 10)}
                      </span>
                    </header>
                    <pre>{row.detail}</pre>
                    {status === "to-review" && (
                      <footer>
                        <button onClick={() => void decide(row, "approved")}>Approve</button>
                        <button onClick={() => void decide(row, "declined")}>Decline</button>
                      </footer>
                    )}
                  </article>
                ))}
              </>
            );
          }}
        </Panel>
      )}
    </ViewFrame>
  );
}
