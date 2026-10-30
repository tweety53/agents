# Finding

- **Severity:** Important (visual-verify departure)
- **Location:** stats/web/src/views/Experiments.tsx:111-112 (`StatCell`), :148 (`Chart` mean-after colour)
- **Verifier's sentence:** "on the closed experiment, Review rounds / change and Verify reds show a green `▼ vs 0.0 before` when before and after are both 0.0 — `StatCell` picks `d > 0 ? \"▲\" : \"▼\"`, so an unchanged value (d = 0) is drawn as an improvement."
