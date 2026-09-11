# Review panel — kan-515-visual-verification-tooling-crop-mockup-frames

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | simple-reviewer | Minor | scripts/test-compose-mockup-frames.sh:63 | `PAGE_RGB="10 10 10"` is declared and never referenced; make_frame_png's heredoc hardcodes its own PAGE constant independently. |
| F2 | principles | Minor | scripts/test-compose-mockup-frames.sh:63 | KISS/dead code: `PAGE_RGB` is declared but the heredoc hardcodes its own PAGE constant, so the two could silently drift with no test catching it. |

findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 none — a dead top-level variable assignment has no runtime-observable effect; confirmed by grep returning only the declaration
finding-reproducer: F2 none — a dead variable has no runtime-observable behavior; confirmed by grep returning only the declaration
