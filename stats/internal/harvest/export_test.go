package harvest

// AttributeAgentFileRecordsForTest exposes attributeAgentFileRecords to
// harvest_test's black-box tests.
var AttributeAgentFileRecordsForTest = attributeAgentFileRecords

// EncodePatchesForTest exposes encodePatches to harvest_test's black-box
// tests -- every other test of it goes through Watcher.RunOnce, the one
// real entry point, but the signals wire shape is more directly pinned by
// calling it straight over a hand-built Delta.
var EncodePatchesForTest = encodePatches
