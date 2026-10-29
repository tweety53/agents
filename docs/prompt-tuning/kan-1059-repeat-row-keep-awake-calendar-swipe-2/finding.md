Held-back interaction source drops Hover/Focus, so the day cell lost its hover and focus layers: `clickable(interactionSource = presses, indication = null)` routes every interaction into `rememberHeldBackPresses`, whose `when` forwards only Press, Release and Cancel.

Severity: Important
Location: gymie-frontend:shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/HomeCalendarGrid.kt:559
