-- name: index-report-date
CREATE INDEX IF NOT EXISTS report_capture_PartitionName_pnr0000_date ON report_capture_PartitionName_pnr0000 (date);
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS report_capture_PartitionName_pnr0000_correlation_id ON report_capture_PartitionName_pnr0000 (correlation_id);

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS rtcp_capture_PartitionName_pnr0000_date ON rtcp_capture_PartitionName_pnr0000 (date);
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS rtcp_capture_PartitionName_pnr0000_correlation_id ON rtcp_capture_PartitionName_pnr0000 (correlation_id);
