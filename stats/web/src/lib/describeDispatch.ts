// Plain words for what a running dispatch does, read from its dispatch key's
// documented shapes (task-, panel-, self-review-, pipeline-fix-, verify,
// visual-verify, explore-); a key matching none falls back to itself, a
// missing one to the role.
export function describeDispatch(key: string | undefined, role: string): string {
  if (!key) return role;
  const after = (m: string | undefined) => (m ? ` after fix ${m}` : "");
  let m: RegExpExecArray | null;
  if ((m = /^tasks?-([\d+]+)-implementer(?:-fix-(\d+))?$/.exec(key)))
    return m[2] ? `Fixing review findings on ${taskWord(m[1])}, round ${m[2]}` : `Implementing ${taskWord(m[1])}`;
  if ((m = /^tasks?-([\d+]+)-reviewer(?:-fix-(\d+))?$/.exec(key))) return `Reviewing ${taskWord(m[1])}${after(m[2])}`;
  if ((m = /^panel-fix-(\d+)$/.exec(key))) return `Fixing code review findings, round ${m[1]}`;
  if ((m = /^panel-(\d+)-(.+)$/.exec(key))) return `Code review round ${m[1]}: ${m[2].replace(/\+/g, ", ").replace(/-/g, " ")}`;
  if ((m = /^full-suite-fix-(\d+)$/.exec(key))) return `Fixing the full test suite, round ${m[1]}`;
  if (/^self-review-.+-pass$/.test(key)) return "Self-review: review pass";
  if ((m = /^self-review-.+-plan$/.exec(key))) return "Self-review: planning fixes";
  if ((m = /^self-review-.+-fix-g(\d+)$/.exec(key))) return `Self-review: fixing the plan, group ${m[1]}`;
  if ((m = /^self-review-.+-fix-v(\d+)$/.exec(key))) return `Self-review: verify fix round ${m[1]}`;
  if ((m = /^self-review-.+-fix(?:-(\d+))?$/.exec(key))) return m[1] ? `Self-review: fix round ${m[1]}` : "Self-review: fixing the plan";
  if ((m = /^self-review-.+-review-(\d+)$/.exec(key))) return `Self-review: review round ${m[1]}`;
  if ((m = /^pipeline-fix-(\d+)-(plan|fix|review)(?:-(\d+))?$/.exec(key)))
    return `Pipeline fix ${m[1]}: ${{ plan: "planning", fix: "fixing", review: "reviewing" }[m[2] as "plan"]}${m[3] ? `, round ${m[3]}` : ""}`;
  if ((m = /^visual-verify(?:-fix-(\d+))?(-full)?(?:-(.+))?$/.exec(key)))
    return `Checking the UI${m[2] ? " in full" : ""}${after(m[1])}${m[3] ? ` · ${m[3]}` : ""}`;
  if ((m = /^verify(?:-fix-(\d+))?(?:-(.+))?$/.exec(key))) return `Running lint and tests${after(m[1])}${m[2] ? ` · ${m[2]}` : ""}`;
  if (/^explore-\d+$/.test(key)) return "Exploring the code";
  if (/^plan-\d+$/.test(key)) return "Planning the change";
  return key;
}

/** "task 3", or "tasks 1, 2, 3" for a `+`-joined task group. */
const taskWord = (ids: string) => (ids.includes("+") ? `tasks ${ids.replace(/\+/g, ", ")}` : `task ${ids}`);
