// SelfReview -- what the self-review pass did, in two read views: Fixed (every
// finding it landed a fix for, with its full detail, read only) and To review
// (the improvements it suggests, each approved or declined here; a Status chip
// keeps decided ones out of the way). Its Period and Project chips are its own,
// like the other dashboards'. Tasks are the operator's to create and plan.
import { useEffect, useState } from "react";
import { Chip, ChipGroup } from "../components/Chip";
import { ViewFrame } from "../components/ViewFrame";
import { fetchProjects, fetchSelfReviewPage, reviewSelfReviewFinding, type SelfReviewPage, type SelfReviewRow } from "../api";
import { PERIOD_DAYS, windowFor, type PeriodChip } from "../format";
import { projectLabel } from "../lib/projectLabel";

type Show = "fixed" | "to-review";
type Status = "to-review" | "approved" | "declined";
const PERIOD_LABELS: Record<PeriodChip, string> = { "7d": "7 days", "30d": "30 days", "90d": "90 days" };
const STATUS_LABEL: Record<Status, string> = { "to-review": "To review", approved: "Approved", declined: "Declined" };

const PAGE_SIZE = 25;

/** Prev / Next under a list, with the range it shows. */
function Pager({ page, onPage }: { page: SelfReviewPage; onPage: (offset: number) => void }) {
  if (page.total <= page.limit) return null;
  const last = Math.min(page.offset + page.rows.length, page.total);
  return (
    <div className="sr-pager">
      <button disabled={page.offset === 0} onClick={() => onPage(Math.max(0, page.offset - page.limit))}>
        Previous
      </button>
      <span className="lb-muted">
        {page.offset + 1}–{last} of {page.total}
      </span>
      <button disabled={last >= page.total} onClick={() => onPage(page.offset + page.limit)}>
        Next
      </button>
    </div>
  );
}

export function SelfReview() {
  const [win, setWin] = useState(() => windowFor("30d"));
  const [project, setProject] = useState<string | undefined>(undefined);
  const [projects, setProjects] = useState<string[]>([]);
  const [show, setShow] = useState<Show>("fixed");
  const [status, setStatus] = useState<Status>("to-review");
  const [offset, setOffset] = useState(0);
  const [reload, setReload] = useState(0);
  const [page, setPage] = useState<SelfReviewPage | null>(null);
  const [error, setError] = useState("");
  const [failed, setFailed] = useState("");

  useEffect(() => {
    fetchProjects().then(setProjects, () => setProjects([]));
  }, []);

  // One page at a time: the filters, the page and a decision made here each refetch just that page.
  useEffect(() => {
    let live = true;
    setError("");
    fetchSelfReviewPage({ from: win.from, to: win.to, project, show, status: show === "to-review" ? status : undefined, limit: PAGE_SIZE, offset }).then(
      (p) => live && setPage(p),
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [win, project, show, status, offset, reload]);

  const filter = <T,>(set: (v: T) => void) => (v: T) => {
    setOffset(0);
    set(v);
  };

  const decide = (row: SelfReviewRow, to: Status) =>
    reviewSelfReviewFinding(row.project, row.id, to).then(
      () => {
        setFailed("");
        setReload((n) => n + 1);
      },
      (err: unknown) => setFailed(err instanceof Error ? err.message : String(err)),
    );

  return (
    <ViewFrame title="Self-review" description="What each self-review pass fixed, and the improvements it suggests for you to approve or decline.">
      <div className="lb-fbar">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={win.period === p} onClick={() => filter(setWin)(windowFor(p))}>
              {PERIOD_LABELS[p]}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={project === undefined} onClick={() => filter(setProject)(undefined)}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p} on={project === p} onClick={() => filter(setProject)(p)} title={p}>
              {projectLabel(p)}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Show">
          <Chip on={show === "fixed"} onClick={() => filter(setShow)("fixed")}>
            Fixed
          </Chip>
          <Chip on={show === "to-review"} onClick={() => filter(setShow)("to-review")}>
            To review
          </Chip>
        </ChipGroup>
        {show === "to-review" && page && (
          <ChipGroup name="Status">
            {(Object.keys(STATUS_LABEL) as Status[]).map((s) => (
              <Chip key={s} on={status === s} onClick={() => filter(setStatus)(s)}>
                {STATUS_LABEL[s]} {s === "to-review" ? page.toReview : s === "approved" ? page.approved : page.declined}
              </Chip>
            ))}
          </ChipGroup>
        )}
      </div>
      <section className="lb-panel" aria-label={show === "fixed" ? "Fixed" : "To review"}>
        <div className="lb-panel-h">
          <span>{show === "fixed" ? "Fixed" : "To review"}</span>
        </div>
        <div className="lb-panel-b">
          {error && <p className="sr-err">{error}</p>}
          {failed && <p className="sr-err">Could not save: {failed}</p>}
          {!page && !error && <p className="lb-muted">Loading…</p>}
          {page && page.rows.length === 0 && (
            <p className="lb-muted">{show === "fixed" ? "No fixed findings in this period." : `Nothing ${STATUS_LABEL[status].toLowerCase()}.`}</p>
          )}
          {page && page.rows.length > 0 && show === "fixed" && (
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
                  {page.rows.map((r) => (
                    <tr key={r.id}>
                      <td>{r.recordedAt.slice(0, 10)}</td>
                      <td className="sr-wrap sr-narrow">{r.change}</td>
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
          )}
          {page &&
            show === "to-review" &&
            page.rows.map((row) => (
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
          {page && <Pager page={page} onPage={setOffset} />}
        </div>
      </section>
    </ViewFrame>
  );
}
