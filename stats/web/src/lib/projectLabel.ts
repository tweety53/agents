// projectLabel derives a project's display name from its key. A project
// key is `<basename of the main checkout>-<first 8 hex of sha1 of that
// checkout's absolute path>` (State file,
// skills/myflow-contracts/state-file.md); the hash exists to disambiguate
// two same-named checkouts, not to be read by a human. This strips that
// documented suffix -- and only that exact shape, anchored at the end,
// lowercase hex, exactly eight characters -- and returns the key
// unchanged when it does not match. The input is whatever the server put
// in projectKey, so a key that was not derived this way (or a raw
// basename with no suffix at all) must come back untouched rather than
// losing a trailing segment to a pattern that was never about it.
const PROJECT_KEY_SUFFIX = /-[0-9a-f]{8}$/;

export function projectLabel(key: string): string {
  return key.replace(PROJECT_KEY_SUFFIX, "");
}
