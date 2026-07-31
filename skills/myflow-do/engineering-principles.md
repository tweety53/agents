# Engineering Principles

The single source of truth for the engineering principles every myflow implementation
must satisfy. Implementer dispatches in `/myflow-do` require this file as reading; the panel's principles reviewer checks
the diff against it via [principles-reviewer-prompt.md](principles-reviewer-prompt.md).

**Do not restate this list anywhere else.** Skills, prompts, and rules reference this
path — a second copy is itself a Single Source of Truth violation.

Each entry gives the principle's name, what it means in one sentence, and one concrete
cue for spotting a violation. The cue is a starting point for judgment, not a lint rule:
several of these principles pull against each other (DRY against WET, KISS against
extensibility), and naming the tradeoff you took is a valid answer to a finding.

The three groups below double as the reviewer's lenses: **Structure** is always covered by the
required principles slot, **Simplicity & state** is lens B, and **Robustness & ops** is lens C.
They carry no path citation on purpose — they are the next three headings of this file, and a
citation pointing at the file the reader is already in reads as pointing somewhere else.

---

## Structure

### SOLID

Single responsibility, open/closed, Liskov substitution, interface segregation, and
dependency inversion — five rules that together keep a type's reasons to change few and
its substitutions safe.

**Violation looks like:** a class whose name contains "and" or "Manager", a subclass that
throws `UnsupportedOperationException` for an inherited method, an interface whose
implementers stub out half its methods, or a service that constructs its own database
client instead of receiving an abstraction.

### Separation of Concerns

Each module addresses one concern, so that changing how something is done does not
require touching what is being done.

**Violation looks like:** SQL strings, HTTP status codes, or JSON field names appearing
inside business logic; a controller computing pricing rules; a domain model annotated for
both persistence and serialization.

### Low coupling, high cohesion

Modules depend on as few others as possible, and everything inside a module belongs
together.

**Violation looks like:** a change to one class forcing edits in five unrelated files; a
`utils` or `helpers` package whose members share nothing but their location; an import
block reaching into four different feature packages.

### Composition over inheritance

Assemble behavior from collaborating parts rather than inheriting it, so behavior can
vary at runtime instead of being fixed at compile time.

**Violation looks like:** an inheritance chain three or more levels deep, a base class
added purely to share two helper methods, or a subclass that overrides a parent method to
disable it.

### Program to an interface, not an implementation

Callers depend on the contract they need, not the concrete type that happens to satisfy
it today.

**Violation looks like:** a field or parameter typed `ArrayList` instead of `List`, a
service field typed as a concrete `…Impl`, or an `instanceof`/`is` chain switching on
concrete subtypes.

### Principle of Least Knowledge (Law of Demeter)

An object talks only to its immediate collaborators, never reaching through them to their
internals.

**Violation looks like:** a train-wreck call chain such as
`order.getCustomer().getAddress().getCity().toUpperCase()`, or a test that must build
three levels of nested objects just to exercise one method.

### Information hiding / encapsulation

A module exposes the smallest surface that serves its callers and keeps every decision
likely to change behind it.

**Violation looks like:** public mutable fields, a getter returning the internal
collection by reference, `internal`/`private` relaxed to `public` "so the test can see
it", or a data class whose invariants are enforced by its callers rather than itself.

---

## Simplicity & state

### DRY — Don't Repeat Yourself

Every piece of knowledge has one authoritative representation; duplicated *knowledge*
drifts even when the duplicated *code* does not.

**Violation looks like:** the same validation rule, magic number, status-code mapping, or
regex written in two places, so a fix has to be applied twice to be correct.

### WET — the deliberate counterweight to DRY

Some duplication is correct: when two code paths merely look alike but change for
different reasons, keeping them separate is cheaper than a shared abstraction that must
be parameterised for both.

**Violation looks like:** an abstraction extracted from two coincidentally-similar call
sites, now carrying boolean flags or a mode enum to serve them both — the "wrong
abstraction" is more expensive than the duplication it removed. Also a violation in the
other direction: unexplained duplication with no note of why it was kept.

### KISS — Keep It Simple

Choose the simplest construction that solves the problem actually in front of you.

**Violation looks like:** a strategy registry, plugin hook, or config switch with exactly
one implementation; generics or reflection where a direct call would do; a comment
explaining why the clever version is clever.

### CQS / CQRS

A method either changes state or answers a question, never both; at system scale, the
read path and write path may be separate models entirely.

**Violation looks like:** a getter that lazily writes to a cache or a database, a
`validate…()` that also mutates its argument, or a save method returning a freshly
computed report the caller then depends on.

### Single Source of Truth

Each fact — a value, a schema, a piece of configuration — is stored and defined in
exactly one place; everything else derives from it.

**Violation looks like:** an enum duplicated between backend and frontend, a default
value hardcoded in both code and config, or a computed total persisted alongside the rows
it is computed from with nothing keeping them consistent.

### Idempotency

Performing the same operation twice produces the same result as performing it once — the
property that makes retries safe.

**Violation looks like:** a POST handler or message consumer with no idempotency key or
uniqueness constraint, a migration that fails on re-run, or a counter incremented on
every delivery of an at-least-once event.

---

## Robustness & ops

### Robustness and failure principles

Be strict about what you emit and defensive about what you accept, fail fast on
programmer error, degrade gracefully on environmental error, and make every failure
observable.

**Violation looks like:** an empty `catch` block, a swallowed exception logged at debug,
an external call with no timeout or bounded retry, a fallback that silently returns empty
data, or an error message that omits the identifier needed to investigate it.

### Principle of Least Astonishment / Surprise

A component behaves the way its name, signature, and neighbours lead a reader to expect.

**Violation looks like:** a `get…` that performs network I/O, a `delete` that soft-deletes
while its siblings hard-delete, a parameter whose meaning inverts an established
convention, or a function whose behavior contradicts its own KDoc.

### Principle of Least Privilege

Every actor — user, service, token, process — gets the narrowest permission and shortest
lifetime that lets it do its job.

**Violation looks like:** a wildcard role or scope, an endpoint added without an
authorization check because "the gateway handles it", a long-lived or non-expiring token,
a database user with DDL rights for a read path, or a secret in a file readable by
everything.

### Testing principles

Tests are written first, assert observable behavior rather than implementation, run fast
and deterministically, and fail for exactly one reason.

**Violation looks like:** a test asserting that a mock was called rather than what
changed, a sleep used for synchronisation, a test whose result depends on execution order
or the wall clock, an assertion weakened rather than a bug fixed, or a test deleted or
`@Ignore`d in the same diff that changed the behavior it covered.

### Twelve-Factor App

Config comes from the environment, backing services are attached resources, processes are
stateless and disposable, and dev/prod stay as close as possible.

**Violation looks like:** an environment name branched on in code (`if (env == "prod")`),
a credential or URL committed to the repository, state kept in process memory across
requests, or logs written to a file the app manages instead of the event stream.
