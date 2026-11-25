-- Log records produced by every source the application listens to.
--
-- One table holds them all: a listener is a column rather than a table, so a new source is a row value instead of
-- a migration. The clear mark of a listener is therefore a filter on the column, which is what keeps each one's
-- history independent without a second table per source.
--
-- level stores the slog.Level value as is, so a record reaches the table without a translation step and a floor of
-- "warn" becomes "level >= 4". The four standard levels are debug -4, info 0, warn 4 and error 8; the wider range
-- leaves room for a custom level, which slog allows between the same bounds.

CREATE TABLE IF NOT EXISTS log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    listener   TEXT    NOT NULL CHECK (length(trim(listener)) > 0), -- source name, for example "app"
    level      INTEGER NOT NULL CHECK (level BETWEEN -128 AND 127), -- slog.Level: debug -4, info 0, warn 4, error 8
    msg        TEXT    NOT NULL CHECK (length(trim(msg)) > 0),
    detail     TEXT    NOT NULL DEFAULT '',                         -- key = value pairs, possibly several lines
    created_at INTEGER NOT NULL CHECK (created_at > 0)              -- milliseconds since the epoch
);

-- A request always names one listener and a start instant, so the pair is the access path. Ordering by created_at
-- inside the index lets the range scan come out sorted and the tail query stop after topk rows.
CREATE INDEX IF NOT EXISTS idx_log_listener_created ON log (listener, created_at);

-- The level floor narrows the same range, so it sits between the listener and the instant. Keeping it after the
-- listener means the index still serves a request that only filters by time.
CREATE INDEX IF NOT EXISTS idx_log_listener_level_created ON log (listener, level, created_at);

-- An exact level is also common, and it reads better from the middle column than from a scan of the range above.
CREATE INDEX IF NOT EXISTS idx_log_created ON log (created_at);
