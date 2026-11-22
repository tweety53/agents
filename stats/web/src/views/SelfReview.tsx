// SelfReview -- the self-review pass's findings, To review first: a list of the improvements it
// suggests with the selected one's structured detail beside it, each approved or declined here with
// an optional reason. Auto-fixed, Approved and Declined are one chip away. No Period or Project
// filter, and every list pages 25 rows at a time. A theme's recurrence is a count of the rows
// loaded for it -- the newest 100 of each list plus the page on screen -- that share its
// `[theme: <tag>]` note prefix.
import { useEffect, useMemo, useState, type ReactNode } from "react";
import { Chip, ChipGroup } from "../components/Chip";
import { ViewFrame } from "../components/ViewFrame";
import { fetchSelfReviewPage, reviewSelfReviewFinding, type SelfReviewPage, type SelfReviewRow } from "../api";

type Show = "to-review" | "fixed" | "approved" | "declined";
const SHOWS: { key: Show; label: string }[] = [
  { key: "to-review", label: "To review" },
  { key: "fixed", label: "Auto-fixed" },
  { key: "approved", label: "Approved" },
  { key: "declined", label: "Declined" },
];
const PAGE = 25;
const CORPUS = 100;
const WINDOW_FROM = new Date("2000-01-01T00:00:00Z");
const STATUS: Record<Show, string | undefined> = { "to-review": "to-review", approved: "approved", declined: "declined", fixed: undefined };

/** The theme tag a note opens with (`[theme: <tag>] `), and the note without it. */
export function splitTheme(note: string): { theme: string; title: string } {
  const m = /^\[theme: ([a-z0-9-]+)\]\s*/.exec(note);
  return m ? { theme: m[1], title: note.slice(m[0].length) } : { theme: "", title: note };
}

/** A body's `## Heading` sections by heading; null when it has no `## Proposal`, i.e. an older, free-form detail. */
export function parseSections(detail: string): Map<string, string> | null {
  const out = new Map<string, string>();
  let name = "";
  for (const line of detail.split("\n")) {
    const h = /^##\s+(.+?)\s*$/.exec(line);
    if (h) {
      name = h[1];
      out.set(name, "");
    } else if (name) out.set(name, out.get(name) + line + "\n");
  }
  return out.has("Proposal") ? new Map([...out].map(([k, v]) => [k, v.trim()])) : null;
}

function field(block: string | undefined, key: string): string {
  const m = block && new RegExp(`^${key}:\\s*(.+)$`, "m").exec(block);
  return m ? m[1].trim() : "";
}

/** Minimal markdown: fenced blocks, bullets (a `- [ ]` item is a checklist entry) and paragraphs. */
function Body({ text }: { text: string }) {
  const parts: ReactNode[] = [];
  const lines = text.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.startsWith("```")) {
      const code: string[] = [];
      for (i++; i < lines.length && !lines[i].startsWith("```"); i++) code.push(lines[i]);
      parts.push(
        <pre key={i} className="sr-code">
          {code.map((l, j) => (
            <span key={j} className={l.startsWith("+") ? "sr-add" : l.startsWith("-") ? "sr-del" : ""}>
              {l + "\n"}
            </span>
          ))}
        </pre>,
      );
    } else if (/^\s*[-*]\s+/.test(line)) {
      const items: string[] = [];
      for (; i < lines.length && /^\s*[-*]\s+/.test(lines[i]); i++) items.push(lines[i].replace(/^\s*[-*]\s+/, ""));
      i--;
      parts.push(
        <ul key={i} className={items.every((it) => /^\[( |x)\]/.test(it)) ? "sr-list sr-check" : "sr-list"}>
          {items.map((it, j) => (
            <li key={j}>{it.replace(/^\[( |x)\]\s*/, (_m, c) => (c === "x" ? "☑ " : "☐ "))}</li>
          ))}
        </ul>,
      );
    } else if (line.trim()) parts.push(<p key={i}>{line}</p>);
  }
  return <>{parts}</>;
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <>
      <h4 className="sr-h">{title}</h4>
      {children}
    </>
  );
}

function Detail({ row, same }: { row: SelfReviewRow; same: SelfReviewRow[] }) {
  const s = parseSections(row.detail);
  if (!s) return <pre className="sr-raw">{row.detail}</pre>;
  const er = s.get("Effort and risk");
  const gap = s.get("Current text");
  const change = s.get("Proposed change");
  return (
    <>
      <div className="sr-prop">
        <div className="sr-k">Proposal</div>
        {s.get("Proposal")}
      </div>
      <div className="sr-strip">
        {[
          ["Effort", field(er, "Effort")],
          ["Touches", field(er, "Touches")],
          ["Seen", same.length ? `${same.length + 1}× (theme)` : "first time"],
          ["Risk", field(er, "Risk")],
        ].map(([k, v]) => (
          <div key={k} className="sr-stat">
            <div className="sr-k">{k}</div>
            <div>{v || "—"}</div>
          </div>
        ))}
      </div>
      <Section title="What happened">
        <Body text={s.get("What happened") ?? ""} />
      </Section>
      {s.get("Examples") && (
        <Section title="Examples">
          <Body text={s.get("Examples")!} />
        </Section>
      )}
      {gap && (
        <Section title="Current text that shows the gap">
          <blockquote className="sr-quote">{gap}</blockquote>
        </Section>
      )}
      {change && (
        <Section title="Proposed change">
          <Body text={change} />
        </Section>
      )}
      {s.get("Benefit") && (
        <Section title="Benefit">
          <Body text={s.get("Benefit")!} />
        </Section>
      )}
      {same.length > 0 && (
        <Section title="Same theme before">
          <ul className="sr-list">
            {same.slice(0, 5).map((r) => (
              <li key={r.id}>
                {splitTheme(r.note).title} <span className="lb-muted">· {r.change} · {r.disposition === "fixed" ? "fixed" : r.reviewStatus}</span>
              </li>
            ))}
          </ul>
        </Section>
      )}
      {s.get("Done when") && (
        <Section title="Done when">
          <Body text={s.get("Done when")!} />
        </Section>
      )}
      {s.get("Trace") && (
        <Section title="Trace">
          <span className="lb-muted">{s.get("Trace")}</span>
        </Section>
      )}
    </>
  );
}

export function SelfReview() {
  const [show, setShow] = useState<Show>("to-review");
  const [offset, setOffset] = useState(0);
  const [selId, setSelId] = useState<number | null>(null);
  const [reason, setReason] = useState("");
  const [reload, setReload] = useState(0);
  const [page, setPage] = useState<SelfReviewPage | null>(null);
  const [fixedTotal, setFixedTotal] = useState(0);
  const [corpus, setCorpus] = useState<SelfReviewRow[]>([]);
  const [error, setError] = useState("");
  const [failed, setFailed] = useState("");

  const q = () => ({ from: WINDOW_FROM, to: new Date(Date.now() + 86400000) });

  // The list on screen, one page; a decision made here refetches it.
  useEffect(() => {
    let live = true;
    setError("");
    fetchSelfReviewPage({ ...q(), show: show === "fixed" ? "fixed" : "to-review", status: STATUS[show], limit: PAGE, offset }).then(
      (p) => {
        if (!live) return;
        if (p.rows.length === 0 && offset > 0) setOffset(Math.max(0, offset - PAGE));
        else setPage(p);
      },
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [show, offset, reload]);

  // What the chips count and the themes compare against, whichever list is on screen.
  useEffect(() => {
    let live = true;
    const ask = (status: string) => fetchSelfReviewPage({ ...q(), show: "to-review", status, limit: CORPUS, offset: 0 });
    Promise.all([ask("to-review"), ask("approved"), ask("declined"), fetchSelfReviewPage({ ...q(), show: "fixed", limit: CORPUS, offset: 0 })]).then(
      ([a, b, c, f]) => {
        if (!live) return;
        setCorpus([...a.rows, ...b.rows, ...c.rows, ...f.rows]);
        setFixedTotal(f.total);
      },
      () => undefined,
    );
    return () => {
      live = false;
    };
  }, [reload]);

  const rows = page?.rows ?? [];
  const sel = rows.find((r) => r.id === selId) ?? rows[0];
  const all = useMemo(() => [...corpus, ...rows], [corpus, rows]);
  const sameTheme = (r: SelfReviewRow) => {
    const t = splitTheme(r.note).theme;
    return t ? all.filter((o, i) => o.id !== r.id && all.findIndex((x) => x.id === o.id) === i && splitTheme(o.note).theme === t) : [];
  };
  const counts: Record<Show, number> = { "to-review": page?.toReview ?? 0, approved: page?.approved ?? 0, declined: page?.declined ?? 0, fixed: fixedTotal };

  const decide = (row: SelfReviewRow, to: "to-review" | "approved" | "declined") =>
    reviewSelfReviewFinding(row.project, row.id, to, to === "to-review" ? "" : reason.trim()).then(
      () => {
        setFailed("");
        setReason("");
        setSelId(null);
        setReload((n) => n + 1);
      },
      (err: unknown) => setFailed(err instanceof Error ? err.message : String(err)),
    );

  const last = page ? Math.min(page.offset + rows.length, page.total) : 0;

  return (
    <ViewFrame title="Self-review" description="Improvements the self-review pass suggests for you to approve or decline, and what it fixed on its own.">
      <div className="lb-fbar">
        <ChipGroup name="Show">
          {SHOWS.map(({ key, label }) => (
            <Chip
              key={key}
              on={show === key}
              onClick={() => {
                setShow(key);
                setOffset(0);
                setSelId(null);
                setPage(null);
              }}
            >
              {label} {counts[key]}
            </Chip>
          ))}
        </ChipGroup>
      </div>
      <section className="lb-panel" aria-label={SHOWS.find((s) => s.key === show)!.label}>
        <div className="lb-panel-b">
          {error && <p className="sr-err">{error}</p>}
          {failed && <p className="sr-err">Could not save: {failed}</p>}
          {!page && !error && <p className="lb-muted">Loading…</p>}
          {page && rows.length === 0 && <p className="lb-muted">Nothing here.</p>}
          {page && sel && (
            <div className="sr-split">
              <div className="sr-listcol">
                {rows.map((r) => {
                  const { theme, title } = splitTheme(r.note);
                  const seen = sameTheme(r).length + 1;
                  return (
                    <button key={r.id} className={`sr-item${r.id === sel.id ? " on" : ""}`} data-testid="suggestion" onClick={() => setSelId(r.id)}>
                      <strong>{title}</strong>
                      <span className="sr-meta">
                        {theme && <code className="sr-theme">{theme}</code>}
                        {seen > 1 && <span className="sr-seen">seen {seen}×</span>}
                        {r.severity && <span className="sr-pill">{r.severity}</span>}
                        <span className="lb-muted">
                          {r.angle} · {r.recordedAt.slice(0, 10)}
                        </span>
                      </span>
                    </button>
                  );
                })}
                {page.total > page.limit && (
                  <div className="sr-pager">
                    <button type="button" disabled={page.offset === 0} onClick={() => setOffset(Math.max(0, page.offset - page.limit))}>
                      Previous
                    </button>
                    <span className="lb-muted">
                      {page.offset + 1}–{last} of {page.total}
                    </span>
                    <button type="button" disabled={last >= page.total} onClick={() => setOffset(page.offset + page.limit)}>
                      Next
                    </button>
                  </div>
                )}
              </div>
              <article className="sr-pane" aria-label="Finding">
                <h3>{splitTheme(sel.note).title}</h3>
                <div className="sr-meta">
                  {splitTheme(sel.note).theme && <code className="sr-theme">{splitTheme(sel.note).theme}</code>}
                  {sel.severity && <span className="sr-pill">{sel.severity}</span>}
                  <span className="sr-pill">{sel.angle}</span>
                  <span className="lb-muted">
                    {sel.change} · {sel.recordedAt.slice(0, 10)}
                  </span>
                  {sel.disposition === "fixed" && (sel.commitUrl ? <a href={sel.commitUrl}>{sel.ref}</a> : <span>{sel.ref}</span>)}
                </div>
                <Detail row={sel} same={sameTheme(sel)} />
                {show === "to-review" && (
                  <footer className="sr-actions">
                    <button type="button" className="sr-approve" onClick={() => void decide(sel, "approved")}>
                      Approve
                    </button>
                    <button type="button" className="sr-decline" onClick={() => void decide(sel, "declined")}>
                      Decline
                    </button>
                    <input aria-label="Reason" placeholder="Reason (optional, stored with the decision)" value={reason} onChange={(e) => setReason(e.target.value)} />
                  </footer>
                )}
                {(show === "approved" || show === "declined") && (
                  <footer className="sr-actions">
                    <span>{sel.reason ? `Reason: ${sel.reason}` : "No reason given."}</span>
                    <button type="button" onClick={() => void decide(sel, "to-review")}>Move back to review</button>
                  </footer>
                )}
              </article>
            </div>
          )}
        </div>
      </section>
    </ViewFrame>
  );
}
