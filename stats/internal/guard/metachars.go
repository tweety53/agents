package guard

// reproducerMetachars is scripts/reproducer-metachars.sh's
// REPRODUCER_METACHARS, the shell-metacharacter set a reproducer command line
// must never carry. The bash file stays canonical while bash guards still
// source it; TestMetacharsMatchBashSource fails when the two differ.
const reproducerMetachars = "|;&$`<>(){}~*?[]#\\'\""
