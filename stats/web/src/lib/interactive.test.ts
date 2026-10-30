import { describe, expect, it } from "vitest";
import type { PlanTask, StateBoardRow } from "../api";
import { currentRequest, interactiveWaiting, isInteractive, lastEndedRequest } from "./interactive";

const tasks: PlanTask[] = [
  { n: 1, title: "a", state: "done", outcome: "landed", commit: "a41c9e2" },
  { n: 3, title: "c", state: "done", outcome: "answered" },
  { n: 2, title: "b", state: "in progress" },
  { n: 4, title: "d", state: "pending" },
];
const row = (patch: Partial<StateBoardRow>) => ({ command: "/flow-interactive", tasks, ...patch }) as StateBoardRow;

describe("interactive helpers", () => {
  it("isInteractive is true only for /flow-interactive", () => {
    expect(isInteractive(row({}))).toBe(true);
    expect(isInteractive(row({ command: "/flow" }))).toBe(false);
    expect(isInteractive(row({ command: undefined }))).toBe(false);
  });
  it("currentRequest is the in-progress request", () => {
    expect(currentRequest(row({}))?.n).toBe(2);
    expect(currentRequest(row({ tasks: tasks.filter((t) => t.state !== "in progress") }))).toBeUndefined();
  });
  it("lastEndedRequest is the highest-numbered done request", () => {
    expect(lastEndedRequest(row({}))?.n).toBe(3);
    expect(lastEndedRequest(row({ tasks: [] }))).toBeUndefined();
  });
  it("interactiveWaiting holds with no request in progress and no stage open", () => {
    const idle = tasks.filter((t) => t.state !== "in progress");
    expect(interactiveWaiting(row({ tasks: idle }))).toBe(true);
    expect(interactiveWaiting(row({ tasks: idle, currentStage: "flow.verify" }))).toBe(false);
    expect(interactiveWaiting(row({}))).toBe(false);
  });
});
