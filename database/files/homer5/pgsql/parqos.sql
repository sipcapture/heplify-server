-- name: create-partition-report_capture
CREATE TABLE IF NOT EXISTS report_capture_{{date}}_{{time}} PARTITION OF report_capture FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-rtcp_capture
CREATE TABLE IF NOT EXISTS rtcp_capture_{{date}}_{{time}} PARTITION OF rtcp_capture FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
