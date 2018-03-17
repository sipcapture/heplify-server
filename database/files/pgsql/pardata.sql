-- name: create-logs-table-partition
CREATE TABLE logs_capture_PartitionName PARTITION OF logs_capture FOR VALUES FROM ('PartitionDate 00:00:00') TO ('PartitionDate 23:59:59');

-- name: create-report-table-partition
CREATE TABLE report_capture_PartitionName PARTITION OF report_capture FOR VALUES FROM ('PartitionDate 00:00:00') TO ('PartitionDate 23:59:59');

-- name: create-rtcp-table-partition
CREATE TABLE rtcp_capture_PartitionName PARTITION OF rtcp_capture FOR VALUES FROM ('PartitionDate 00:00:00') TO ('PartitionDate 23:59:59');

-- name: create-call-table-partition
CREATE TABLE sip_capture_call_PartitionName PARTITION OF sip_capture_call FOR VALUES FROM ('PartitionDate 00:00:00') TO ('PartitionDate 23:59:59');

-- name: create-registration-table-partition
CREATE TABLE sip_capture_registration_PartitionName PARTITION OF sip_capture_registration FOR VALUES FROM ('PartitionDate 00:00:00') TO ('PartitionDate 23:59:59');
