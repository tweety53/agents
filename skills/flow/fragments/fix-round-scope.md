> **FIX-ROUND SCOPE:** this is a fix-round re-review. Its scope is the fix diff you were given
> and the sites of the findings it fixes, plus at most the code neighbouring those hunks — the
> enclosing function or section — and nothing else on the branch. Run only the tests and
> specs that diff touches, never the module, repository or live-spec suite. The fix report at
> `<fix report>` carries each fixed finding's proof: the test run before the fix and after it,
> and the failure with the fix reverted. Check that proof against the diff; reproduce it
> yourself only where it is missing, does not match the diff, or does not show the failure it
> claims — and say which of those it was. Read every hunk of the fix diff, a fix made for another
> slot's finding included, for a defect the fix itself introduces — an evaluation moved out of the
> guard that kept it safe, a branch the fix changed — and report it as a new finding at its line.
> Given a diff, your report file carries one verdict line
> per finding you were sent to re-review, in exactly this shape: `verdict: F<n> — fixed —
> <path>:<line>` or `verdict: F<n> — not fixed — <path>:<line>`, where `<path>:<line>` is a line
> inside the hunk of that diff you read for that finding (`none` in its place, on a `not fixed`
> line, when the diff touched nothing of it).
