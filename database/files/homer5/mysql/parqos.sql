-- name: create-partition-report_capture
ALTER TABLE report_capture_all_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));

-- name: create-partition-rtcp_capture
ALTER TABLE rtcp_capture_all_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));
