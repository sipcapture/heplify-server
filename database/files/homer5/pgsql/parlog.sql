-- name: create-partition-logs_capture
CREATE TABLE IF NOT EXISTS logs_capture_{{date}}_{{time}} PARTITION OF logs_capture FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
