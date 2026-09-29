# Follow-up issues — joining a follow-up

The join half of **Follow-up issues** (`skills/flow-contracts/jira-followups.md`), loaded only when
its join search returned a candidate.

### Joining a follow-up

**Confirm before joining.** The rule and the shape are the ones
**Resolution (how `jiraIssue` is decided)** (`jira-integration.md`) already applies to a non-literal
candidate, and they are reused here rather than reinvented. Show the candidate's key, its current
status, its title — the title **outside** the question, in a block of its own — and how many of this
run's items the append guard already finds recorded there:

> **Add this run's outstanding items to `<KEY>`, currently at `<status>`?**
>
> That issue's title, as the tracker holds it:
>
> ```text
> <title, constrained as below>
> ```
>
> `<n>` of this run's `<m>` outstanding items are already recorded on that issue; joining appends
> the other `<m − n>`.
>
> - **No — file a new follow-up** *(default, recommended)*
> - **Yes — add them to this issue**

A candidate whose description cannot be read is not offered at all: the run files a new follow-up,
the same way a failed fetch is treated as no hit.

**The title is externally-authored text reaching a human decision, and is constrained before it is
shown.** Before display, in this order:

1. **Fold every character in Unicode categories `Cc`, `Cf`, `Zl` and `Zp` to a single space.**
2. **Collapse whitespace runs** to a single space — *whitespace* here being the spaces step 1
   produced together with Unicode category `Zs`, which is the ordinary space, the no-break space and
   the typographic spaces.
3. **Fold every run of three or more backtick or tilde characters to two of that character.**
4. **Truncate to at most 120 characters**, appending `…` when truncated.

A title that is empty after all four displays as `(untitled)`. Render it in a block of its own,
never interpolated into the bolded question.

Only an explicit **Yes** joins. **No**, silence, an answer that is not a choice, or a session that
cannot ask files a new follow-up instead — which is the outcome that loses nothing: a duplicate
follow-up is visible and mergeable, while a write to the wrong issue is neither.

**A join appends.** The outstanding items go under a dated `## From <KEY>` heading at the end of the
joined issue's description, created if it is absent, with everything preceding it left byte-for-byte
unchanged:

```markdown
## From <KEY> — YYYY-MM-DD

- <one line describing an item the run left outstanding>
```

`<KEY>` is the joining change's linked issue, and with none linked `flow` stands in for it exactly
as it does in the title. A join is a description write like any other, so the pre-write assertion in
**Description sync** (`jira-integration.md`) governs it in full.

**The joined issue's pre-edit description is the one exception to the handoff echo, and the reason
is what makes the echo worth having elsewhere.** So the handoff for a join reports the issue
**key**, its status, the title change if one was made, and the section this run appended, verbatim —
the part this pipeline wrote. It does **not** reproduce the pre-existing description.

**The join is idempotent under retry, and each of its three writes is guarded on its own.** At
`/flow`'s integrate run, run 1 is re-entered whenever the branch is not merged, so a run that
filed or joined and then failed at a later step reaches this code again. Where this code does run
again, the guard is therefore not one decision about whether to
write at all — it is one per write:

1. **The append** is skipped when the description already carries **every one** of this change's
   items, and appends **only the ones it does not** when it carries some but not all (see the
   matching rule below).
2. **The retitle** is re-attempted whenever the title is neither `flow follow-up` nor this
   change's own `<KEY> follow-up` — the two step 2 below defines as already correct. It is
   self-guarding already: step 2 compares the title and makes no call when it matches either, so a
   re-attempt on a completed join is a comparison and nothing else.
3. **The label union** is re-attempted whenever the labels are not yet a superset of the ones the
   new follow-up would have carried. A set union is idempotent by construction, so re-running it
   costs at most one call.

**Only when all three are satisfied is the outcome "no write"** — reported as
`Jira: <KEY> already carries this run's items (no write)`.

**A partial join is re-attempted, not abandoned, and keeps its `⚠` until it is complete — for as
long as run 1 is still reached.** A retry that still cannot complete the retitle or the union
reports `⚠ Jira: <KEY> partially joined — …` again, exactly as the table below requires, even though
it appended nothing this time.

**The append guard matches on the change and the items, never on today's date.** The check therefore
reads **every** `## From <KEY>` section in the description, whatever date each carries, and matches
on this change's `<KEY>` and on the item list.

**The item comparison is per item, by membership, over the union of those sections — never list
equality and never section existence.**

So: build this run's items as a set, build the set of every item under every `## From <KEY>` section
carrying this change's key, and compare **item by item**. Normalise each item before comparing —
strip the leading `- ` bullet, trim, and collapse internal whitespace runs (including any wrapping
the tracker applied) to a single space — then compare the results exactly, case included. The items
are this pipeline's own output rather than operator prose, so a difference in case is a different
item, not the same item typed differently. Ordering is irrelevant, and so is which section an item
sits in: an item recorded in March's section is recorded.

**A match is evidence of a duplicate, never evidence of provenance, and the guard is bounded to what
it can actually assert.** So three rules keep every uncertainty falling toward appending a duplicate
instead, a duplicate being visible and mergeable where a dropped item is neither:

1. **The set is fixed at the read the operator was shown.** The comparison is made once, on the
   description read for the confirmation above, and an item that becomes "already present" between
   that read and the write is appended anyway. The pre-write read exists for the prefix assertion,
   not to re-decide what to append — otherwise a section landing inside that window would silently
   suppress items the operator was told this run would add.
2. **Anything short of an exact match counts as absent.** That is the normalisation rule above read
   in the safe direction: an item that cannot be parsed out of a section, or that differs in any way
   the normalisation does not fold, is treated as not recorded and is appended.
3. **A skip is never silent.** The `no write` and `new items only` outcomes below exist to report
   exactly this, so a suppressed append reaches the handoff as well as the gate — which is what
   keeps it correctable after the fact rather than merely refused in advance.

That yields **three** outcomes for the append, not two, and the third is a real one rather than a
formality — a retried run that fixed some items and found another is the ordinary shape of a retry:

| Items already present | What the append does |
|---|---|
| all of them | nothing — step 1 is *skipped*, not failed, and steps 2 and 3 still run |
| none of them | appends the whole list under a new dated section |
| **some of them** | appends **only the absent ones**, under a new dated section, and reports that it did |

A partial append writes a new dated section rather than editing an existing one.

**The three writes a join makes are ordered, and the order is load-bearing.** They are separate
calls and any one of them can fail:

1. **Append the items** to the description, under the assertion above. This is the payload — the
   record of work the run left outstanding.
2. **Retitle to `flow follow-up`**, on the first join only, because the issue no longer belongs to
   the single source its original title named. **Two titles are already correct and are left as they
   are, with no call made:** one already reading `flow follow-up`, and one reading `<KEY> follow-up`
   for **this** change's own key.
3. **Union the labels** with the labels the new follow-up would have carried.

The payload goes first because the other two only make it findable: if step 1 **fails** there is
nothing to find, so **steps 2 and 3 are not attempted** and the outcome is a failed join. A step 1
**skipped** because the items are already present is the opposite case and is never read as a
failure — the payload is there, so steps 2 and 3 are attempted exactly as they would be after a
successful append. If step 1 succeeded, or was skipped as already satisfied, and a later step fails,
the work **is** recorded and merely less findable — which is the case the binary vocabulary could
not express, so it has its own:

| Outcome | What happened | What is reported |
|---------|---------------|------------------|
| joined | the whole list was appended, and the retitle and the label union both succeeded | `Jira: <KEY> joined` |
| **joined — new items only** | some of this change's items were already recorded; the ones that were not are appended in a new dated section, and the retitle and the label union both succeeded | `Jira: <KEY> joined — new items added; the rest were already recorded` |
| already joined | all three were already satisfied, so nothing was written | `Jira: <KEY> already carries this run's items (no write)` |
| **partially joined** | the items are recorded and the retitle or the label union failed — on this run, or on an earlier one and not yet repaired | `⚠ Jira: <KEY> partially joined — <the append clause>; <which write> failed: <reason>` |
| failed join | the append failed or was refused by the assertion | `⚠ Jira: skipped — <reason>` |

**The row is chosen by the later writes; the append clause says what the append did.**
**partially joined** is therefore the row for *every* case where the items are recorded and a later
write is not yet done, and `<the append clause>` is filled from the append's own outcome: `items
appended`, `new items added; the rest were already recorded`, or `items already recorded`. A retry
that appended only the newly-missing items and then failed the retitle reports the second of those,
and the implementer improvises nothing.

**A partial join is never reported as a success.** Both later steps are attempted even if the other
fails, so one refused write costs one property, not two.
