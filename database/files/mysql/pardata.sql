-- name: create-partition-logs_capture
ALTER TABLE logs_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-report_capture
ALTER TABLE report_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-rtcp_capture
ALTER TABLE rtcp_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-sip_capture_call
ALTER TABLE sip_capture_call_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-sip_capture_registration
ALTER TABLE sip_capture_registration_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));