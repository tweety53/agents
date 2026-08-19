# myflow-stats-views Specification

## Purpose
TBD - created by archiving change kan-16-myflow-stats-app. Update Purpose after archive.
## Requirements
### Requirement: Statistics are queryable over an arbitrary period

The daemon SHALL expose a statistics API in which every view accepts a start and an end instant, and
SHALL perform the period restriction in the query rather than returning unbounded data for a client
to filter.

A period containing no stage runs SHALL return an empty result, never an error.

#### Scenario: A period is requested

- **WHEN** a view is requested for a given period
- **THEN** the response covers only stage runs falling within it
- **AND** the restriction was applied by the query rather than by the client

#### Scenario: An empty period

- **WHEN** a view is requested for a period in which no stage runs exist
- **THEN** the response is an empty result and the request succeeds

#### Scenario: A run straddling a period boundary

- **WHEN** a stage run starts before the period and ends inside it
- **THEN** the view states which convention it applies to that run and applies it consistently
  across every view

### Requirement: Project is a first-class filter

Every view SHALL accept a project filter and SHALL also support aggregation across projects, because
the state directory is already keyed per project and more than one project uses the pipeline.

#### Scenario: One project is selected

- **WHEN** a view is requested with a project filter
- **THEN** only that project's stage runs contribute to the result

#### Scenario: All projects are selected

- **WHEN** a view is requested with no project filter
- **THEN** every project's stage runs contribute, and the result identifies which project each
  contribution came from

### Requirement: The eight views are served

The statistics surface SHALL provide these views, each answering the question named against it:

| View | Question |
|------|----------|
| Live state board | What is every open change's state, across every project, and what command comes next |
| Cost per change | What did one change cost end to end, broken down by command and stage |
| Stage leaderboard | Which stages cost the most across a period, by mean, median and p90 |
| Trend over time | Is the cost per change rising or falling as the pipeline changes |
| Cache efficiency | What is the cache-read to cache-creation ratio per stage |
| Panel economics | How many findings does each review panel roster preset produce per token spent |
| Model comparison | How do cost and subsequent rework differ for the same stage across models |
| Rework rate | How often is a command re-run as a fix, and how often is a stage abandoned |

#### Scenario: The live state board replaces reading state files

- **WHEN** the live state board is requested
- **THEN** it lists every non-archived change with its state, its last update, and the command that
  should run next
- **AND** the information matches what `/myflow-status` reports for the same changes

#### Scenario: Cost per change is broken down

- **WHEN** cost per change is requested for one change
- **THEN** the result attributes tokens, currency and wall-clock time to each command and each stage
  of that change

#### Scenario: Rework is counted from attempts

- **WHEN** the rework-rate view is requested
- **THEN** it derives rework from recorded stage attempt numbers and abandoned outcomes rather than
  inferring it from timing

### Requirement: The interface is served locally and unauthenticated

The daemon SHALL serve its user interface and its API bound to the loopback interface only, and
SHALL NOT require credentials.

The interface SHALL be served by the same process that serves the API, so the two cannot reach
different versions.

#### Scenario: A remote client attempts to connect

- **WHEN** a client on another host attempts to reach the daemon
- **THEN** the connection is not accepted, because the listener is bound to loopback

#### Scenario: The interface and API versions

- **WHEN** the daemon is upgraded
- **THEN** the interface it serves is the one built with that binary, and no separately deployed
  interface can be at a different version

### Requirement: The absence of history is stated rather than drawn as zero

Where a view covers a period before telemetry existed, it SHALL state that no data was recorded for
that span rather than rendering it as measured zero cost. No history precedes the store: it starts
empty, and nothing is imported into it.

#### Scenario: A period before the store held anything

- **WHEN** a view covers a period before stage runs were recorded
- **THEN** the output distinguishes "not recorded" from "recorded as zero"

### Requirement: Every reasonable field is filterable, searchable and sortable

The query surface SHALL allow filtering, sorting and free-text search across the recorded fields of
both changes and stage runs — including the change's project, name, state, branch, linked issue,
planning effort and roster, and the stage run's command, stage, attempt, harness, outcome, timings
and repository — **and** across keys within the metrics structure, addressed by key path.

Results SHALL be paginated, and every ordering SHALL be made total by a unique tiebreaker so that
paging through a result set shows each row exactly once.

A record whose metrics lack the key being sorted on SHALL order as *absent*, distinct from a record
whose metric is recorded as zero.

#### Scenario: Sorting by a metrics key

- **WHEN** results are sorted by a key inside the metrics structure
- **THEN** records are ordered by that key's value
- **AND** records lacking the key are ordered as absent rather than as zero

#### Scenario: Free-text search

- **WHEN** a search term is supplied
- **THEN** records matching it in any identity field are returned

#### Scenario: Paging is stable

- **WHEN** a caller pages through a result set ordered by a field with duplicate values
- **THEN** no record appears on two pages and no record is skipped

### Requirement: Query fields are resolved through a fixed allowlist

A field named in a filter, a sort or a search SHALL be resolved through a fixed mapping from
accepted names to stored fields. A name absent from that mapping SHALL be rejected with an error
naming the field and the accepted alternatives.

No part of a query SHALL be constructed by interpolating caller-supplied text into a query
identifier position. A metrics key path SHALL be validated against a syntax rule and then supplied
as a bound value, never as interpolated text.

#### Scenario: An unknown sort field

- **WHEN** a caller sorts by a field the mapping does not contain
- **THEN** the request is rejected, naming the field and what is accepted
- **AND** no query is executed

#### Scenario: An injection attempt in a sort field

- **WHEN** a caller supplies query syntax in place of a field name
- **THEN** it fails the mapping and is rejected, and no part of it reaches the executed query

#### Scenario: A malformed metrics key path

- **WHEN** a caller supplies a metrics key path that fails the syntax rule
- **THEN** the request is rejected rather than executed with the path as text

### Requirement: A multi-repository change reads as one row

Every view SHALL present a change spanning several repositories as a single row aggregating the
whole unit of work. A per-repository breakdown SHALL be available on request, and SHALL NOT be the
default shape.

#### Scenario: A multi-repository change in a cost view

- **WHEN** a cost view covers a change affecting two repositories
- **THEN** it reports one row whose totals cover both

#### Scenario: A per-repository breakdown

- **WHEN** a breakdown by repository is requested for that change
- **THEN** each repository's contribution is reported separately, and they sum to the unit's total

### Requirement: One change opens on its own dashboard

The interface SHALL provide a route that opens a single change and shows every stage run recorded
against it, so the cost of one unit of work is readable without reading an aggregate that spans
other changes.

That route SHALL show, per stage run, the command and stage it belongs to, its attempt number, its
outcome, its wall-clock duration, and every metric the run's metrics bag actually carries — and
SHALL apply the absence-is-never-zero rule to each of them individually, so an unmeasured metric on
an otherwise measured run reads as unavailable rather than as zero.

The route's summary totals SHALL come from the server's own aggregation over the whole change, not
from summing the stage runs the interface currently holds, so a change whose stage runs exceed one
page is never summarised from a subset of them.

The route SHALL be reachable by selecting a change from the live state board, not by URL entry
alone.

#### Scenario: A change with runs across several commands

- **WHEN** a change whose stages ran under more than one command is opened
- **THEN** every stage run is listed, each attributed to the command it ran under

#### Scenario: A stage run that recorded no token metrics

- **WHEN** a listed stage run carries wall-clock time but no token metrics
- **THEN** its duration is shown and its token and currency figures read as unavailable, not as zero

#### Scenario: A change with more stage runs than one page holds

- **WHEN** the change's stage runs exceed the page the interface requested
- **THEN** the summary totals still cover every run, because they are the server's aggregate rather
  than a sum of the rows on screen

#### Scenario: Reaching one change from the board

- **WHEN** a change is selected on the live state board
- **THEN** that change's own dashboard opens

### Requirement: The interface is a dashboard of panels under shared controls

The interface SHALL present each view as a dashboard composed of panels, where a panel is one
self-contained result with its own heading, rather than as a single undivided table per page.

A time range and the filter controls SHALL live once in the interface's own chrome and scope every
panel shown beneath them, so no two panels on a dashboard can be showing different periods or
different filters at the same moment.

Changing any shared control SHALL re-request the affected panels from the server rather than
narrowing rows already fetched, so what is displayed is always the server's answer for the controls
as they currently read.

#### Scenario: The time range is changed

- **WHEN** the time range is changed while a dashboard is open
- **THEN** every panel on it reloads for the new range, and none continues to show the old one

#### Scenario: A filter is applied

- **WHEN** a shared filter is set
- **THEN** the panels re-request their data with that filter applied server-side, rather than hiding
  rows already received

### Requirement: Views are restrictable to one model

The statistics surface SHALL accept a model restriction on every view that aggregates stage runs,
applied where the query runs rather than by discarding results after they are returned.

A view whose rows are not stage runs SHALL reject a model restriction rather than accept and ignore
it, so a control that cannot apply is never silently inert.

A stage run whose recorded metrics carry no model SHALL NOT match any model restriction, and the
response SHALL report how many runs were excluded on that ground, so an unmeasured run is
distinguishable from a run measured as belonging to another model.

The set of models offered SHALL be the set actually recorded in the period being viewed, obtained
from the server rather than assumed or hard-coded, so a model this build has never heard of still
appears once it has been used.

#### Scenario: A cost view restricted to one model

- **WHEN** a stage-run view is requested with a model restriction
- **THEN** its rows cover only stage runs whose recorded model is that model

#### Scenario: A model restriction on the live state board

- **WHEN** a model restriction is sent to a view whose rows are changes rather than stage runs
- **THEN** the request is rejected, rather than answered as though no restriction had been sent

#### Scenario: Runs that recorded no model

- **WHEN** a period contains stage runs whose metrics recorded no model, and a model restriction is
  applied
- **THEN** those runs are absent from the rows, and the response reports how many were excluded for
  having no recorded model

#### Scenario: The models offered

- **WHEN** the set of available models is requested for a period
- **THEN** it lists exactly the distinct models recorded in that period, and a model used for the
  first time appears without any change to the interface

### Requirement: A recorded but unmeasured run is distinguishable from an absent one

The interface SHALL distinguish three states rather than two: a period in which nothing was
recorded, a run that was recorded but received no measurement, and a value measured as zero.

A run that was recorded and never attributed SHALL NOT be presented the same way as a period with no
runs. Where every run in a period is unattributed, the interface SHALL say so, so that a
misconfiguration in recording is distinguishable from a genuinely quiet period.

#### Scenario: A run recorded without measurement

- **WHEN** a stage run exists for the period but received no usage
- **THEN** it is shown as recorded and unmeasured, distinctly from a run measured as zero

#### Scenario: A period whose runs are all unmeasured

- **WHEN** every stage run in the requested period is unattributed
- **THEN** the interface reports that runs were recorded but none was measured, rather than
  reporting that no data was recorded

### Requirement: A project is named on screen, not keyed

Wherever the interface names a project for a human to read, it SHALL show the project's **display
name** — the project key with its derivation suffix removed — and SHALL NOT show the raw key.

A project key is `<basename of the main checkout>-<first 8 hex of sha1 of that checkout's absolute
path>`. The suffix exists so two same-named repositories in different directories address different
records; it identifies, and identity is not a label.

The display name SHALL be derived by removing a trailing `-` followed by exactly eight lowercase
hexadecimal characters. A key not matching that shape SHALL be shown **unchanged** — a key derived
some other way stays readable rather than losing its last segment to a pattern that was never about
it.

The full key SHALL remain reachable from every surface that shows a display name: as the row
identity, as the link target where the surface links, and as a tooltip on the named element. Where
two projects share a display name, what distinguishes them is therefore still available without
being the thing every reader has to parse.

#### Scenario: A project is named in a table

- **WHEN** a view renders a column naming the project a row came from
- **THEN** the cell reads the display name, and the full key is available from that cell

#### Scenario: A key that does not carry the derivation suffix

- **WHEN** a project key does not end in `-` plus eight hexadecimal characters
- **THEN** it is shown exactly as it is, with nothing removed

#### Scenario: Two projects share a display name

- **WHEN** two project keys differ only in their suffix
- **THEN** both are named identically on screen, and each one's full key is still reachable from its
  own surface

### Requirement: The project parameter accepts a display name

Every endpoint accepting a `project` parameter SHALL accept either the full project key or a display
name, so that a value read off the screen is a value the interface accepts.

Resolution SHALL happen on the server, where the set of projects is known. The interface SHALL NOT
reconstruct the mapping from whichever rows it happens to be showing — a project with no rows in the
selected period is absent from the interface and present in the store, so a client-side mapping
would be wrong exactly when a filter is most needed.

A display name matching **exactly one** project SHALL resolve to that project's key. A display name
matching **more than one** SHALL be rejected with a client error that names the ambiguity and lists
the candidate keys; it SHALL NOT be resolved to any one of them, because a filter whose value means
two projects is not a filter, and choosing silently answers a question the caller did not ask. A
value matching **none** SHALL be passed through unchanged, yielding no rows exactly as an unknown key
does today.

The view queries themselves SHALL continue to filter on an exact project key. Resolution sits above
them, so the rule is stated once rather than copied into every view's query.

#### Scenario: A display name is filtered on

- **WHEN** a view is requested with a `project` value that is a display name matching one project
- **THEN** the result is that project's rows, exactly as if its full key had been sent

#### Scenario: A display name matching two projects

- **WHEN** a view is requested with a `project` value matching more than one project
- **THEN** the request is rejected, and the error names the ambiguity and the candidate keys

#### Scenario: A full key is filtered on

- **WHEN** a view is requested with a `project` value that is an exact project key
- **THEN** it is used as-is, with no resolution attempted

### Requirement: A stage run opens onto its own dispatches

The route that shows one change's stage runs SHALL let a stage run be expanded to show the dispatches
it made, one row per dispatch, carrying that dispatch's description, agent type, model, token total
and cost, ordered by cost descending.

Per-dispatch cost SHALL be derived through the same pricing path every other cost figure uses,
applied to the dispatch's own recorded model.

A dispatch the store never measured SHALL be rendered as unavailable rather than as zero, exactly as
an unmeasured stage run already is, so that the absence of a measurement stays distinguishable from a
measurement of nothing.

This requirement SHALL NOT add a view. The count of views this capability serves is unchanged, and
per-dispatch cost is reached only by expanding a stage run on the route that already shows it.

#### Scenario: A panel stage expands to its slots

- **WHEN** a stage run that dispatched three review slots and a fix round is expanded
- **THEN** four rows are shown, the most expensive first, each naming its description, agent type and
  model

#### Scenario: An unmeasured dispatch is not drawn as zero

- **WHEN** a dispatch carries no token figures
- **THEN** its row renders as unavailable rather than as a zero cost

#### Scenario: No view is added

- **WHEN** the interface is loaded
- **THEN** it serves the same set of views it served before, and the per-dispatch rows appear only
  under an expanded stage run

