calendar-swipe-month-grid: DEPARTURE — month-grid press NOT fixed. Under a finger-speed drag (30 moves 16ms apart, 150px left, from day 10) frames 002–019 draw the pressed grey fill on day 10 while the pointer has moved away. The delivered probe (12 instantaneous moves, 300ms wait) passes over the defect.

Severity: Important (visual-verify departure)
Location: gymie-playwright:tests/kan-1059-repeat-keep-awake-swipe.spec.ts — "the day under the finger draws no press mid-drag"
