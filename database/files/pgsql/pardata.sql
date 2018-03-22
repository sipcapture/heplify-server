-- name: create-partition-logs_capture
CREATE TABLE IF NOT EXISTS logs_capture_PartitionName_pnr0000 PARTITION OF logs_capture FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-report_capture
CREATE TABLE IF NOT EXISTS report_capture_PartitionName_pnr0000 PARTITION OF report_capture FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-rtcp_capture
CREATE TABLE IF NOT EXISTS rtcp_capture_PartitionName_pnr0000 PARTITION OF rtcp_capture FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-sip_capture_call
CREATE TABLE IF NOT EXISTS sip_capture_call_PartitionName_pnr0000 PARTITION OF sip_capture_call FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-sip_capture_registration
CREATE TABLE IF NOT EXISTS sip_capture_registration_PartitionName_pnr0000 PARTITION OF sip_capture_registration FOR VALUES FROM ('StartTime') TO ('EndTime');
