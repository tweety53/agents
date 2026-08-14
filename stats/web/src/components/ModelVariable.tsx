// ModelVariable is the "model" template variable in the dashboard bar. It
// is populated from GET /api/v1/models for the period and project
// currently selected -- never a hard-coded list -- so a model used for
// the first time appears without any change to this build
// (specs/myflow-stats-views/spec.md, "The models offered"; "obtained from
// the server rather than assumed or hard-coded").
//
// It is disabled, with a stated reason, whenever the caller says the
// active dashboard cannot accept a model restriction -- the live state
// board, whose rows are changes rather than stage runs, is the one case
// this build has today (internal/api/stats.go rejects "model" there with
// 400). `disabled` is a prop rather than something this component infers
// from a route itself: which dashboards accept a model restriction is a
// fact about the *view*, and App.tsx already knows which view is active,
// so asking this component to re-derive that from a URL would duplicate
// knowledge that already lives in one place (Single Source of Truth,
// engineering-principles.md).
import { useEffect, useState } from "react";
import { fetchModels } from "../api";
import type { Period } from "./PeriodPicker";

export interface ModelVariableProps {
  period: Period;
  project: string | undefined;
  model: string | undefined;
  onChange: (model: string | undefined) => void;
  disabled: boolean;
  disabledReason?: string;
}

export function ModelVariable({ period, project, model, onChange, disabled, disabledReason }: ModelVariableProps) {
  const [models, setModels] = useState<string[]>([]);
  const fromMs = period.from.getTime();
  const toMs = period.to.getTime();

  useEffect(() => {
    if (disabled) {
      setModels([]);
      return;
    }
    let cancelled = false;
    fetchModels({ from: period.from, to: period.to, project })
      .then((res) => {
        if (!cancelled) setModels(res.models);
      })
      .catch(() => {
        if (!cancelled) setModels([]);
      });
    return () => {
      cancelled = true;
    };
    // period.from/period.to are Date objects -- new identities every
    // render -- so this effect keys off their millisecond values instead,
    // the same convention useStatsView.ts already follows.
  }, [disabled, fromMs, toMs, project]);

  return (
    <label className="dashboard-variable model-variable" title={disabled ? disabledReason : undefined}>
      Model
      <select
        aria-label="Model"
        value={model ?? ""}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value === "" ? undefined : e.target.value)}
      >
        <option value="">All models</option>
        {models.map((m) => (
          <option key={m} value={m}>
            {m}
          </option>
        ))}
      </select>
      {disabled && disabledReason && <span className="dashboard-variable-hint">{disabledReason}</span>}
    </label>
  );
}
