// describeDispatch turns each documented dispatch-key shape into plain words;
// a key matching none comes back unchanged so nothing reads as invented.
import { describe, expect, it } from "vitest";
import { describeDispatch } from "./describeDispatch";

describe("describeDispatch", () => {
  it.each([
    ["task-3-implementer", "Implementing task 3"],
    ["task-3-implementer-fix-2", "Fixing review findings on task 3, round 2"],
    ["task-3+4-reviewer", "Reviewing task 3, 4"],
    ["task-3-reviewer-fix-1", "Reviewing task 3 after fix 1"],
    ["panel-1-primary+mutation", "Code review round 1: primary, mutation"],
    ["panel-2-test-sanity+call-efficiency", "Code review round 2: test sanity, call efficiency"],
    ["panel-fix-1", "Fixing code review findings, round 1"],
    ["full-suite-fix-2", "Fixing the full test suite, round 2"],
    ["self-review-kan-1048-rework-pass", "Self-review: reviewing the change"],
    ["self-review-kan-1048-rework-plan", "Self-review: planning fixes"],
    ["self-review-kan-1048-rework-fix-g2", "Self-review: fixing group 2"],
    ["self-review-kan-1048-rework-fix-v1", "Self-review: fixing a red verify, round 1"],
    ["self-review-kan-1048-rework-fix", "Self-review: fixing"],
    ["self-review-kan-1048-rework-review-1", "Self-review: reviewing fixes, round 1"],
    ["pipeline-fix-3-review-2", "Pipeline fix 3: reviewing, round 2"],
    ["visual-verify-gymie", "Checking the UI · gymie"],
    ["visual-verify-fix-1-full", "Checking the UI in full after fix 1"],
    ["verify-fix-2-gymie-frontend", "Running lint and tests after fix 2 · gymie-frontend"],
    ["verify", "Running lint and tests"],
    ["explore-1", "Exploring the code"],
    ["something-else", "something-else"],
  ])("%s", (key, want) => {
    expect(describeDispatch(key, "implementer")).toBe(want);
  });

  it("falls back to the role without a key", () => {
    expect(describeDispatch(undefined, "explorer")).toBe("explorer");
  });
});
