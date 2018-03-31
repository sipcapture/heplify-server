-- name: index-logs-date
CREATE INDEX IF NOT EXISTS logs_capture_PartitionName_pnr0000_date ON logs_capture_PartitionName_pnr0000 (date);
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS logs_capture_PartitionName_pnr0000_correlation_id ON logs_capture_PartitionName_pnr0000 (correlation_id);
