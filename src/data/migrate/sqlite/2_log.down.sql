-- Drop the log table together with its indexes; sqlite removes the indexes with the table.
DROP INDEX IF EXISTS idx_log_created;
DROP INDEX IF EXISTS idx_log_listener_level_created;
DROP INDEX IF EXISTS idx_log_listener_created;
DROP TABLE IF EXISTS log;
