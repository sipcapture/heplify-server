-- name: create-partition-report_capture
ALTER TABLE report_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-rtcp_capture
ALTER TABLE rtcp_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));
