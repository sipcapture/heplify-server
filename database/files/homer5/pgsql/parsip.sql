-- name: create-partition-sip_capture_call
CREATE TABLE IF NOT EXISTS sip_capture_call_PartitionName_pnr0000 PARTITION OF sip_capture_call FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-sip_capture_registration
CREATE TABLE IF NOT EXISTS sip_capture_registration_PartitionName_pnr0000 PARTITION OF sip_capture_registration FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-sip_capture_rest
CREATE TABLE IF NOT EXISTS sip_capture_rest_PartitionName_pnr0000 PARTITION OF sip_capture_rest FOR VALUES FROM ('StartTime') TO ('EndTime');
