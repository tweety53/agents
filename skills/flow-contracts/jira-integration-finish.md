# Jira integration — finish and issue creation

The parts of **Jira integration** (`skills/flow-contracts/jira-integration.md`) only a finish run or an
issue-creating command acts on. Loaded by bare `/flow` run 1 and by every command that creates an issue.

### In Review is not tied to a pull request

bare `/flow`'s In Review transition is **not** conditioned on a pull request existing. It fires
at the end of a successful run 1 whichever route was taken — pull request, merge and push, or
handled manually — because conditioning it on a PR is what let a merge-and-push change reach Done
without ever passing through In Review. A run 1 that stops before its chosen route completes — a
rejected push, a merge conflict, a failed PR creation — transitions nothing.

### The join confirmation

**The join confirmation** is the other carve-out from **Never blocking** (`skills/flow-contracts/jira-integration.md`), stated under
**Joining a follow-up** (`skills/flow-contracts/jira-followups-join.md`), and bounded identically: asked once,
only when a candidate was actually found, with anything but an explicit yes taking the safe course
and the run continuing regardless. Both exist for the same reason — a write aimed at an issue this
pipeline did not choose — and neither ever gates a state write.

### The join echo — the exception to Description sync's pre-edit echo

**A join is the one write this echo does not cover**, because the description it would reproduce
belongs to another change's issue and this pipeline never authored it. What is echoed there instead
is under **Joining a follow-up** (`skills/flow-contracts/jira-followups-join.md`), and why under
**Moved by KAN-859 — jira-followups.md** (`skills/flow-contracts/jira-integration-rationale.md`).

### Labels on issues the pipeline creates

An issue any `/flow*` command creates carries **every label on the change's linked issue, plus
`AI-generated`**. No label is invented: the parent's labels exist by construction, and
`AI-generated` is applied only because the project already uses it. With no linked issue, the
created issue carries `AI-generated` alone. Link the created issue to the change's issue whenever
one exists.

**A self-review finding adds one more label.** An issue filed from a self-review finding carries,
on top of the set above, the label naming the angle that produced it. The angle-to-label table is
canonical in `skills/flow-self-review/SKILL.md`'s step 2, cited here rather than copied, so the
two cannot drift.

**Creation is a Jira write like any other, and fails the same way.** `createJiraIssue` can be
refused for auth, permission, an unknown project key, a label the project does not allow, or a
missing required field, and a session may have no Atlassian tooling at all. Any of those is one
`⚠ Jira: skipped — <reason>` line and the command continues, per
**Never blocking** (`skills/flow-contracts/jira-integration.md`). A command whose operator chose an option that
*includes* filing — run 1's "File or join a Jira follow-up, then continue" is the one that exists today — still
does the rest of what that option promised; the filing failing does not silently convert the answer
into a different one, and it is never left unmentioned.

### Follow-up issues

Follow-up naming, the join search, and the append-only join write are governed by
**Follow-up issues** (`skills/flow-contracts/jira-followups.md`), loaded by bare `/flow`
run 1.
