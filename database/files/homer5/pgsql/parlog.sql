-- name: create-partition-logs_capture
CREATE TABLE IF NOT EXISTS logs_capture_DayDate_pnr0000 PARTITION OF logs_capture FOR VALUES FROM ('StartTime') TO ('EndTime');
