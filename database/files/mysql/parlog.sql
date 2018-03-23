-- name: create-partition-logs_capture
ALTER TABLE logs_capture_all_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));
