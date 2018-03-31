-- name: create-partition-report_capture
CREATE TABLE IF NOT EXISTS report_capture_PartitionName_pnr0000 PARTITION OF report_capture FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-rtcp_capture
CREATE TABLE IF NOT EXISTS rtcp_capture_PartitionName_pnr0000 PARTITION OF rtcp_capture FOR VALUES FROM ('StartTime') TO ('EndTime');
