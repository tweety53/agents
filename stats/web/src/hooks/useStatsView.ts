// useStatsView is the one fetch path every one of the eight views uses: it
// calls fetchStatsView (api.ts) with exactly the period and project the
// caller passed -- never a locally computed or cached period -- so the
// period/project restriction task 13's own non-negotiable requirement
// insists on always happens on the server, once, per request. Sharing this
// hook rather than duplicating its effect in eight components is what
// keeps that guarantee in one place (DRY, engineering-principles.md).
import { useEffect, useState } from "react";
import { fetchStatsView, type StatsResponse, type StatsViewParams, type ViewName } from "../api";

export type StatsViewState<Row> =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: StatsResponse<Row> };

export function useStatsView<Row>(view: ViewName, params: StatsViewParams): StatsViewState<Row> {
  const [state, setState] = useState<StatsViewState<Row>>({ status: "loading" });

  const fromMs = params.from.getTime();
  const toMs = params.to.getTime();

  useEffect(() => {
    let cancelled = false;
    setState({ status: "loading" });
    fetchStatsView<Row>(view, params)
      .then((data) => {
        if (!cancelled) setState({ status: "ready", data });
      })
      .catch((err: unknown) => {
        if (!cancelled) setState({ status: "error", message: err instanceof Error ? err.message : String(err) });
      });
    return () => {
      cancelled = true;
    };
    // params.from/params.to are Date objects -- new identities every render
    // even when the instant is unchanged -- so the effect keys off their
    // millisecond values (fromMs/toMs) instead of the Date objects
    // themselves; params.project, params.breakdown, params.change and
    // params.model are primitives, stable across renders that don't change
    // them. params.model is listed explicitly (task 19) so that changing
    // the dashboard bar's model variable re-requests this view -- omitting
    // it here would leave the view showing the previous model's rows until
    // some other dependency happened to change too.
  }, [view, fromMs, toMs, params.project, params.breakdown, params.change, params.model]);

  return state;
}
