// AgentRoles draws every role's allowed pairs in force, grouped by the
// pipeline stage the role runs in, one row per role, as
// .superpowers/sdd/roles-prototype.html draws it. Each role's base and
// opt-in pairs are edited here and saved to flowd, their one source
// (kan-1103).
import { Fragment, useEffect, useRef, useState } from "react";
import { PAIR_SHAPE, fetchRoles, postRolesEdit, type RoleEntry } from "../api";

// Brief plain-English role names for the Does column; the full sentence is the hover.
// ponytail: a role missing here shows its sentence's first four words.
export const SHORT_DOES: Record<string, string> = {
  main: "Runs the pipeline",
  explorer: "Reads the codebase",
  "plan-designer": "Proposes a plan",
  designer: "Draws the undrawn",
  planner: "Writes the plan",
  implementer: "Writes the code",
  "panel-fix": "Fixes review findings",
  "operator-asked": "Does what you asked",
  "gated-reviewer": "Reviews risky tasks",
  "gated-re-reviewer": "Re-checks risky tasks",
  "panel-reviewer": "Reviews the branch",
  decider: "Checks the plan",
  "minor-fix": "Fixes review Minors",
  "no-progress-fixer": "Takes over a stuck fix",
  "panel-rerun": "Re-reviews the branch",
  "panel-merger": "Merges findings",
  "visual-verifier": "Compares UI to mockup",
  runner: "Runs suites, returns a verdict",
  general: "Any task no other role fits",
  "self-review-pass": "Reviews the run",
  "self-review-plan": "Plans review fixes",
  "self-review-fix": "Makes a review fix",
  "self-review-review-1": "Reviews that fix",
  "self-review-review-r": "Re-reviews that fix",
  "pipeline-fix-plan": "Plans a tooling fix",
  "pipeline-fix-fix": "Makes a tooling fix",
  "pipeline-fix-review-1": "Reviews that fix",
  "pipeline-fix-review-r": "Re-reviews that fix",
  "pair-judge": "Picks each dispatch's pair",
};
const shortDoes = (r: RoleEntry) => SHORT_DOES[r.role] ?? r.does.split(" ").slice(0, 4).join(" ");

// Roles run by a guard's headless `claude -p`, never dispatched as an agent.
const HEADLESS = new Set(["pair-judge", "decider"]);

const MAIN_PAIR = "session's own pair";

// Tokens are drawn in thousands, the label's (k).
// The daemon writes instants in its own zone; dates are read in UTC.

// A pairs list wraps between its pairs, never inside one at a pair's hyphen.
function PairList({ pairs, className }: { pairs: string[]; className?: string }) {
  return (
    <code className={className}>
      {pairs.map((p, i) => (
        <Fragment key={p}>
          {i > 0 && ", "}
          <span className="pair">{p}</span>
        </Fragment>
      ))}
    </code>
  );
}

// PairEditor is one tier's pairs as a multiselect field: the chips are its
// trigger, which opens a filterable checklist over every pair the roles use.
// Typing a well-formed pair (PAIR_SHAPE) not yet listed offers to add it.
// Escape, a click outside or Done closes it.
function PairEditor({ role, tier, pairs, catalog, onChange }: { role: string; tier: string; pairs: string[]; catalog: string[]; onChange: (p: string[]) => void }) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState("");
  const box = useRef<HTMLSpanElement>(null);
  // The menu is position: fixed, so the roles panel's scroll box never clips it;
  // it sits under the trigger, or above when the viewport has no room below.
  const [at, setAt] = useState<{ left: number; top?: number; bottom?: number }>({ left: 0 });
  useEffect(() => {
    if (!open) return;
    const r = box.current!.getBoundingClientRect();
    const up = window.innerHeight - r.bottom < 300 && r.top > window.innerHeight - r.bottom;
    setAt(up ? { left: r.left, bottom: window.innerHeight - r.top + 4 } : { left: r.left, top: r.bottom + 4 });
    const shut = () => setOpen(false);
    // A scroll inside the checklist itself keeps it open.
    const scrolled = (ev: Event) => !box.current?.contains(ev.target as Node) && shut();
    window.addEventListener("scroll", scrolled, true);
    window.addEventListener("resize", shut);
    const away = (ev: MouseEvent) => box.current && !box.current.contains(ev.target as Node) && setOpen(false);
    document.addEventListener("mousedown", away);
    return () => {
      document.removeEventListener("mousedown", away);
      window.removeEventListener("scroll", scrolled, true);
      window.removeEventListener("resize", shut);
    };
  }, [open]);
  const q = draft.trim();
  const canAdd = PAIR_SHAPE.test(q) && !pairs.includes(q) && !catalog.includes(q);
  const options = [...new Set([...catalog, ...pairs])].sort().filter((o) => o.toLowerCase().includes(q.toLowerCase()));
  const add = () => {
    onChange([...pairs, q]);
    setDraft("");
  };
  return (
    <span
      className={`ms ${tier}${open ? " open" : ""}`}
      ref={box}
      onKeyDown={(ev) => {
        if (ev.key === "Escape") setOpen(false);
      }}
    >
      <button type="button" className="trig" aria-haspopup="listbox" aria-expanded={open} title={pairs.join(", ")} aria-label={`Choose ${tier} pairs for ${role}`} onClick={() => setOpen(!open)}>
        <span className="chips">
          {pairs.length === 0 ? (
            <span className="none">none</span>
          ) : (
            pairs.map((p) => (
              <span className="chip" key={p}>
                <span className="pair">{p}</span>
              </span>
            ))
          )}
        </span>
        {pairs.length > 1 && <span className="cnt">{pairs.length}</span>}
      </button>
      {open && (
        <div className="menu" style={at}>
          <input
            type="search"
            aria-label={`Add ${tier} pair to ${role}`}
            placeholder="Filter or type a new pair…"
            autoFocus
            value={draft}
            onChange={(ev) => setDraft(ev.target.value)}
            onKeyDown={(ev) => {
              if (ev.key === "Enter") {
                ev.preventDefault();
                if (canAdd) add();
              }
            }}
          />
          <div className="opts">
            {canAdd && (
              <label>
                <input type="checkbox" checked={false} onChange={add} />
                <span>Add “{q}”</span>
              </label>
            )}
            {options.map((o) => (
              <label key={o}>
                <input type="checkbox" checked={pairs.includes(o)} onChange={(ev) => onChange(ev.target.checked ? [...pairs, o] : pairs.filter((x) => x !== o))} />
                <span className="pair">{o}</span>
              </label>
            ))}
          </div>
          <div className="foot">
            <span>{`${pairs.length} selected`}</span>
            <button type="button" onClick={() => setOpen(false)}>
              Done
            </button>
          </div>
        </div>
      )}
    </span>
  );
}

// The agent a role's pairs run, effort stripped, and how that effort is written
// ("@" or a "-" suffix); null when the role has no pair to read it from.
function agentOf(r: RoleEntry): { agent: string; sep: string } | null {
  const p = [...r.base, ...r.optin].find((x) => x.includes("/"));
  if (!p) return null;
  const rest = p.split("/")[1];
  const at = rest.match(/^(.*)@[a-z]+$/);
  if (at) return { agent: at[1], sep: "@" };
  const dash = rest.match(/^(.*)-(low|medium|high|xhigh)$/);
  return dash ? { agent: dash[1], sep: "-" } : { agent: rest, sep: "" };
}

const MODELS = ["opus", "sonnet", "fable", "haiku"];
const EFFORTS = ["low", "medium", "high", "xhigh"];
// Every model x effort the role's agent can run, so a pair nobody has used yet can be ticked.
const optionsFor = (r: RoleEntry): string[] => {
  const a = agentOf(r);
  return a && a.sep ? MODELS.flatMap((m) => EFFORTS.map((e) => `${m}/${a.agent}${a.sep}${e}`)) : [];
};

type Draft = { base: string[]; optin: string[] };

function RoleRow({ r, catalog, draft, onDraft }: { r: RoleEntry; catalog: string[]; draft: Draft; onDraft: (d: Draft) => void }) {
  const main = r.role === "main";
  return (
    <tr className="role">
      <td className="role" title={r.does}>
        <code>{r.role}</code>
        {HEADLESS.has(r.role) && <span className="tag">headless</span>}
        <span className="does">{shortDoes(r)}</span>
      </td>
      <td className="pairs">
        {main ? (
          <PairList pairs={[MAIN_PAIR]} />
        ) : (
          <PairEditor role={r.role} tier="main" pairs={draft.base} catalog={optionsFor(r).length ? optionsFor(r) : catalog} onChange={(base) => onDraft({ ...draft, base })} />
        )}
      </td>
      <td className="pairs">
        {!main && <PairEditor role={r.role} tier="explicit allow" pairs={draft.optin} catalog={optionsFor(r).length ? optionsFor(r) : catalog} onChange={(optin) => onDraft({ ...draft, optin })} />}
      </td>
    </tr>
  );
}

// RolesTable edits every role in place; one Save writes all the changed roles.
function RolesTable({ roles, onSaved }: { roles: RoleEntry[]; onSaved: () => void }) {
  const catalog = [...new Set(roles.flatMap((r) => [...r.base, ...r.optin]))];
  const [drafts, setDrafts] = useState<Record<string, Draft>>({});
  const [saveError, setSaveError] = useState<string | null>(null);
  const draftOf = (r: RoleEntry): Draft => drafts[r.role] ?? { base: r.base, optin: r.optin };
  const changed = roles.filter((r) => r.role !== "main" && (draftOf(r).base.join() !== r.base.join() || draftOf(r).optin.join() !== r.optin.join()));
  // One row per role, groups in order of first appearance.
  const groups: { name: string; roles: RoleEntry[] }[] = [];
  for (const r of roles) {
    let g = groups.find((x) => x.name === r.group);
    if (!g) groups.push((g = { name: r.group, roles: [] }));
    g.roles.push(r);
  }
  return (
    <>
      {changed.length > 0 && (
        <div className="savebar">
          <span>
            {changed.length} {changed.length === 1 ? "role" : "roles"} changed: {changed.map((r) => r.role).join(", ")}
          </span>
          <button
            type="button"
            onClick={() => {
              setSaveError(null);
              postRolesEdit(Object.fromEntries(changed.map((r) => [r.role, draftOf(r)])))
                .then(() => {
                  setDrafts({});
                  onSaved();
                })
                .catch((err: Error) => setSaveError(err.message));
            }}
          >
            Save
          </button>
          <button type="button" onClick={() => setDrafts({})}>
            Discard
          </button>
          {saveError && (
            <span role="alert" className="panel-error">
              {saveError}
            </span>
          )}
        </div>
      )}
      <table>
        <thead>
          <tr>
            <th>Role</th>
            <th>Main</th>
            <th>Explicit allow</th>
          </tr>
        </thead>
        <tbody>
          {groups.map((g) => (
            <Fragment key={g.name}>
              <tr className="grp">
                <th colSpan={3}>{g.name}</th>
              </tr>
              {g.roles.map((r) => (
                <RoleRow key={r.role} r={r} catalog={catalog} draft={draftOf(r)} onDraft={(d) => setDrafts({ ...drafts, [r.role]: d })} />
              ))}
            </Fragment>
          ))}
        </tbody>
      </table>
    </>
  );
}

export function AgentRoles() {
  const [roles, setRoles] = useState<RoleEntry[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  // Bumped after a save: reloads the roles.
  const [reload, setReload] = useState(0);

  useEffect(() => {
    let live = true;
    fetchRoles()
      .then((r) => live && setRoles(r))
      .catch((err: Error) => live && setError(err.message));
    return () => {
      live = false;
    };
  }, [reload]);

  if (error) {
    return (
      <p role="alert" className="panel-error">
        Failed to load: {error}
      </p>
    );
  }
  if (!roles) return <p role="status">Loading…</p>;
  return (
    <div className="agent-roles">
      <section className="panel roles" aria-label="Current roles">
        <h2 className="ttl">Current roles — allowed pairs in force</h2>
        <div className="wrap" id="roles">
          <RolesTable roles={roles} onSaved={() => setReload((n) => n + 1)} />
        </div>
      </section>
    </div>
  );
}
