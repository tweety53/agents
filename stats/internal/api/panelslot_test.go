package api

import "testing"

// TestPanelSlotFromKey pins the slot a hook-begun panel row is stored under:
// the hook passes no -slot, so the key's slot part names it, a part, repo or
// unread suffix dropped; a key naming no panel slot yields none.
func TestPanelSlotFromKey(t *testing.T) {
	for key, want := range map[string]string{
		"panel-0-mutation":                          "mutation",
		"panel-0-primary+principles":                "primary+principles",
		"panel-1-primary+test-sanity+failure-modes": "primary+test-sanity+failure-modes",
		"panel-9-primary-unread":                    "primary",
		"panel-2-mutation-gymie":                    "mutation",
		"panel-1-rerun":                             "",
		"panel-1-primary+rerun":                     "",
		"fix-9-review":                              "",
		"task-3-reviewer":                           "",
	} {
		if got := panelSlotFromKey(key); got != want {
			t.Errorf("panelSlotFromKey(%q) = %q, want %q", key, got, want)
		}
	}
}
