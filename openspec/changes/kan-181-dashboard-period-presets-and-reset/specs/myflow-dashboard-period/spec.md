## ADDED Requirements

### Requirement: The dashboard opens on a default period

The dashboard SHALL open on a default period of the preceding 30 days ending at the present instant,
and that default SHALL be asserted by a test rather than left as incidental behaviour.

Every view is gated behind the period, so the default decides what a first-time visitor sees. A
default nothing asserts can drift without any test failing.

#### Scenario: A visitor opens the dashboard

- **WHEN** the dashboard is opened with no period specified
- **THEN** the period is the preceding 30 days ending now

### Requirement: Common ranges are chosen without typing

The period control SHALL offer presets for the ranges an operator picks routinely, and selecting one
SHALL set both bounds together.

The control SHALL indicate which preset, if any, matches the current period, so the active range is
legible without comparing two timestamps.

Manual entry of an arbitrary range SHALL remain available, and entering one SHALL clear the
indication that a preset is active, since the range is then no longer that preset's.

**A preset covering all recorded history SHALL resolve to a concrete lower bound**, because the
server requires both bounds and rejects an open range. The bound SHALL be low enough that no
recorded data falls below it.

#### Scenario: A preset is selected

- **WHEN** an operator selects a preset
- **THEN** both bounds are set to that preset's range, and that preset is shown as active

#### Scenario: An arbitrary range is typed

- **WHEN** an operator edits a bound by hand so the range matches no preset
- **THEN** the range takes effect and no preset is shown as active

#### Scenario: The all-history preset is selected

- **WHEN** an operator selects the preset covering all recorded history
- **THEN** the request carries a concrete lower bound the server accepts, and no record is excluded
  by it

### Requirement: The period can be returned to the default

The control SHALL offer a way back to the default period, and that control SHALL be present only
while the current period differs from the default.

Showing it unconditionally would make it furniture; showing it only when it does something makes its
presence the signal that the period has been changed.

#### Scenario: The period differs from the default

- **WHEN** the current period is not the default
- **THEN** a control returning it to the default is available

#### Scenario: The period is the default

- **WHEN** the current period is the default
- **THEN** no reset control is shown

### Requirement: The period survives a reload and can be shared

The period SHALL be carried in the page's URL, so that reloading preserves it and the URL identifies
the range someone else would see.

**Carrying it SHALL NOT change which view the URL selects.** The route is matched against the URL's
path portion alone, decided before the period is read, so adding or removing a period never alters
route resolution.

**A period that cannot be read as a valid range SHALL fall back to the default**, rather than being
partially applied or rendering an empty view. Missing, incomplete, unparsable, or reversed bounds are
all invalid. A URL is an untrusted input, and the failure this capability exists to prevent — an
empty page with no explanation — must not be reachable by crafting one.

Updating the period SHALL NOT add a history entry per change, so returning to the previous page does
not require stepping back through every range the operator tried.

#### Scenario: A period-carrying URL is reloaded

- **WHEN** a URL carrying a period is loaded
- **THEN** that period is applied, and the view named by the URL is the one shown

#### Scenario: A URL carries an invalid period

- **WHEN** a URL's period is missing a bound, unparsable, or ends before it starts
- **THEN** the default period is applied and the view still renders

#### Scenario: The route is unaffected by the period

- **WHEN** a period is added to, or removed from, a URL naming a view
- **THEN** the same view is selected either way

### Requirement: An empty view says whether the period is why

Where a view returns no rows and the period is not the default, the empty state SHALL say the period
may be the reason and SHALL offer the return to the default.

An empty result and a wrong range are indistinguishable on screen, and that ambiguity is what makes a
mistyped bound expensive: the operator concludes there is no data. The distinction is only meaningful
when the period has been changed, so a default period SHALL NOT carry the explanation — under the
default, empty means empty.

#### Scenario: A view is empty under a narrowed period

- **WHEN** a view returns no rows and the period is not the default
- **THEN** the empty state says the period may be why and offers the return to the default

#### Scenario: A view is empty under the default period

- **WHEN** a view returns no rows and the period is the default
- **THEN** the empty state reports no data without blaming the period
