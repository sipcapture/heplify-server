-- name: index-logs-date
CREATE INDEX IF NOT EXISTS logs_capture_{{date}}_{{time}}_date ON logs_capture_{{date}}_{{time}} (date);
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS logs_capture_{{date}}_{{time}}_correlation_id ON logs_capture_{{date}}_{{time}} (correlation_id);
