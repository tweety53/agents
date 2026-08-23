# myflow-run-telemetry delta — kan-212-persist-per-dispatch-cost-tokens-model-and-role

## MODIFIED Requirements

### Requirement: A stage run is bound to its session after the mark, within a bounded window

Binding SHALL be permitted to complete after the mark is recorded, because the evidence that
identifies the session may not exist at the moment the mark is made.

Binding SHALL be attempted for a bounded period, after which the stage run SHALL be left
unattributed. An unbounded search SHALL NOT be performed, so that a harness which will never produce
the evidence costs a bounded amount of work rather than a permanent one.

Once bound, usage SHALL be attributed exactly as it is for a stage run whose session was known from
the start; binding SHALL introduce no second attribution path.

**A token the bounded search gave up on SHALL be recorded durably, not held in process memory
alone.** A give-up held only in memory is invisible to every other component, is lost on restart
without being retried, and leaves the only account of it in a log line.

**A persisted give-up SHALL be re-attempted when the search restarts.** A run whose transcript still
holds its marks is recoverable, and abandoning it permanently on the strength of one bounded window
discards data that is still present. Re-attempting SHALL be bounded exactly as the first attempt is,
so a harness that will never produce the evidence still costs a bounded amount of work per restart.

**The re-attempt SHALL NOT depend on new bytes arriving in the transcript.** A finished run's
transcript has been read to its end, so a search that only examines newly-read content can never find
marks that were read past — which is every mark belonging to the runs this recovery exists for. The
re-attempt SHALL therefore search the transcript independently of the position the usage-attribution
pass has reached.

**That search SHALL NOT produce usage records or advance the attribution position.** Binding a
correlator and counting usage are separate concerns: the position exists so usage is counted exactly
once, and a recovery path that re-delivered already-counted records would double-count them.

**Binding recovers a stage run's identity, not its past cost.** A stage run bound after the fact SHALL
carry the session it belongs to, and SHALL NOT thereby gain token figures for usage recorded before it
was bound — that usage lies behind the attribution position the search above must not disturb. Usage
recorded after the binding SHALL attribute normally.

**A stage run left unattributed SHALL be distinguishable from one whose harness writes no
transcript.** Both have no token figures; only one of them is a loss, and a reader that cannot tell
them apart cannot tell a broken run from an ordinary one.

#### Scenario: The evidence appears after the mark

- **WHEN** the evidence identifying a stage run's session becomes available after the mark was
  recorded
- **THEN** the stage run is bound to that session, and usage falling in its window is attributed to
  it

#### Scenario: A harness that produces no such evidence

- **WHEN** a stage is marked on a harness that produces no session evidence at all
- **THEN** binding is attempted for the bounded period and then abandoned, and the stage run remains
  recorded and unattributed

#### Scenario: Binding does not disturb what is already attributed

- **WHEN** a stage run is bound to its session
- **THEN** usage already recorded against other stage runs is unchanged

#### Scenario: A given-up token is retried after a restart

- **WHEN** the bounded search gives up on a token, and the search later restarts while that token's
  transcript still carries its marks
- **THEN** the token is searched for again and binds, and the stage runs carrying it record the
  session they belong to

#### Scenario: The marks were read past before the retry

- **WHEN** a persisted give-up's transcript has already been read to its end, so no new bytes will
  ever carry its marks
- **THEN** the retry still finds those marks and binds the token
- **AND** no usage record is produced and the attribution position is unchanged

#### Scenario: A stage run bound after the fact

- **WHEN** a stage run is bound by the retry, after usage for its window was already read past
- **THEN** it records its session and reports its cost as unattributed rather than as zero
- **AND** usage recorded after the binding is attributed to it normally

#### Scenario: An unattributed run is distinguishable from an unmeasurable one

- **WHEN** a reader examines a stage run with no token figures
- **THEN** it can tell whether the run's session was never bound or whether its harness writes no
  transcript
