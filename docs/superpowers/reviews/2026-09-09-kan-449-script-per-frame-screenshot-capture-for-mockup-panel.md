# Review panel — kan-449-script-per-frame-screenshot-capture-for-mockup

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | important | /Users/tweety53/Projects/gymie/.flow/project.md:1 | Task 6's gymie commit adds three unrelated new sections (## execution mode, ## implementer model, ## review panel, all dynamic) not in its declared scope of just the mockups row and one sentence |
| F2 | principles | important | scripts/compose-mockup-frames.py:198 | A corrupt/unreadable mockup frame file crashes with an uncaught PIL.UnidentifiedImageError traceback and exit code 1, violating the script's own documented cannot-answer exit-2 contract |
| F3 | principles | minor | scripts/compose-mockup-frames.py:110 | stem = name[:-len('.png')] is computed independently in match_capture and in main's per-line loop rather than factored into one helper |
| F4 | code-review-low | important | scripts/compose-mockup-frames.py:145 | Same asymmetric handling: unreadable mockup frame crashes uncaught with exit 1 instead of the documented exit 2, unlike the defensively-validated stdin capture path |
| F5 | mutation | important | scripts/compose-mockup-frames.py:170-173 | Deleting the duplicate-screenshot-name guard leaves scripts/test-compose-mockup-frames.sh fully green — no test supplies a map with two lines for the same screenshot name |
| F6 | mutation | important | scripts/compose-mockup-frames.py:175-177 | Deleting the not-endswith('.png') guard leaves the harness fully green — no test supplies a map line whose screenshot name lacks the .png suffix |
| F7 | mutation | minor | scripts/compose-mockup-frames.py:32 | Changing BACKGROUND from the documented (40,40,40) to (0,0,0) leaves every test green — no test asserts a pixel value, only composite dimensions |
| F8 | mutation | minor | scripts/compose-mockup-frames.py:154 | Bottom-aligning the mockup paste instead of top-aligning it leaves every test green — no test checks paste offset or pixel content, only overall dimensions |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed

reproducers-total: 8
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F3 none — a readability observation, not a runtime-verifiable defect
finding-reproducer: F4 .superpowers/sdd/reproducers/0-code-review-low-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-mutation-2.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-mutation-3.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-4.sh
