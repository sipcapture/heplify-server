-- name: create-partition-logs_capture
ALTER TABLE logs_capture_all_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));
