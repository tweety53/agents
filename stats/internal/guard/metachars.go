package guard

// reproducerMetachars is the shell-metacharacter set a reproducer command
// line must never carry. It is the one copy: its bash twin was deleted
// (KAN-887), and check-panel-reproducers and run-reproducer both read it here.
const reproducerMetachars = "|;&$`<>(){}~*?[]#\\'\""
