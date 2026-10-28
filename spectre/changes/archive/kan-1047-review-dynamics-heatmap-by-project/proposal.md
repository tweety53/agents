# kan-1047-review-dynamics-heatmap-by-project

## Why

- The Review dynamics page mixes every project, every plan size and long-retired reviewer slots into one view, so a reader cannot tell how gymie's big changes are reviewed today.
- The only project control is the header's free-text field, which wants the full key (`gymie-7c1f238a`).

## What changes

- **Projects** bar on Review dynamics: `All` plus one button per project with a reviewer dispatch in the range (label drops the `-<8 hex>` suffix). A click sets the shared project filter, so the header field, every tile, chart and the heatmap follow.
- **Task size** bar: `All · Big · Regular · Small · Micro`. Filters the whole page to changes whose latest recorded decision has that plan class (`class` query parameter, review-dynamics only).
- **Reviewers** bar: `Current` (default) hides retired slots from the per-slot chart and the heatmap; `All` shows them.
- The page opens on **Last 4 weeks** (was Last 8 weeks). Defaults: Last 4 weeks, Family, All, All, Current.
- Fix run 1: every Review dynamics y-axis label names the value its tick sits at — ticks run `0, s, 2s` on a 1·2·5 step, whole numbers only for whole-number data (peak 3 drew `2` at 1.5; an empty chart left its axis line unlabelled).
