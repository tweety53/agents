// projectLabel derives a project's display name from its key by stripping
// the documented `-[0-9a-f]{8}` suffix (State file,
// skills/myflow-contracts/state-file.md). The whole risk is the fallback:
// a key that does not end in that exact shape must come back unchanged,
// since the helper's input is whatever the server put in projectKey and it
// must not lose a segment to a pattern that was never about it -- these
// cases are tasks.md's own step-1 enumeration.
import { describe, expect, it } from "vitest";
import { projectLabel } from "./projectLabel";

describe("projectLabel", () => {
  it("strips the documented 8-hex suffix", () => {
    expect(projectLabel("agents-a740d89c")).toBe("agents");
  });

  it("strips the suffix from a second, differently-shaped key", () => {
    expect(projectLabel("gymie-7c1f238a")).toBe("gymie");
  });

  it("returns a key with no suffix unchanged", () => {
    expect(projectLabel("agents")).toBe("agents");
  });

  it("returns a key with a seven-hex-character tail unchanged", () => {
    expect(projectLabel("repo-a740d89")).toBe("repo-a740d89");
  });

  it("returns a key with a nine-hex-character tail unchanged", () => {
    expect(projectLabel("repo-a740d89c1")).toBe("repo-a740d89c1");
  });

  it("returns a key whose tail is the right length but not hex unchanged", () => {
    expect(projectLabel("repo-zzzzzzzz")).toBe("repo-zzzzzzzz");
  });

  it("returns a key with an uppercase hex tail unchanged, since the derivation produces lowercase hex", () => {
    expect(projectLabel("repo-A740D89C")).toBe("repo-A740D89C");
  });

  it("strips only the trailing suffix from a basename that itself contains hyphens", () => {
    expect(projectLabel("my-repo-a740d89c")).toBe("my-repo");
  });

  it("returns the empty string unchanged rather than throwing", () => {
    expect(projectLabel("")).toBe("");
  });
});
