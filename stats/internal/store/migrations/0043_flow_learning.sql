-- 0043_flow_learning.sql: add the harness-wide learning setting and the
-- learning-note topic history.
--
-- flow_settings.learning is the switch /flow and /flow-fast read before
-- composing a learning note; it defaults on (learning-default-on), so a
-- row written before this column reads on. flow_learning_topics records
-- every note shown, one row per showing, so RecordLearningTopic can refuse
-- a topic already shown three times (cap-in-store). The cap is
-- harness-wide -- counted over every project -- since the operator reads
-- one stream of notes.
ALTER TABLE flow_settings ADD COLUMN learning BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE flow_learning_topics (
	id          BIGSERIAL PRIMARY KEY,
	topic       TEXT NOT NULL,
	note        TEXT NOT NULL,
	change      TEXT NOT NULL,
	project_key TEXT NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX flow_learning_topics_topic ON flow_learning_topics (topic);
