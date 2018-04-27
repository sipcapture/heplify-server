-- name: create-partition-sip_capture_call
ALTER TABLE sip_capture_call_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));

-- name: create-partition-sip_capture_registration
ALTER TABLE sip_capture_registration_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));

-- name: create-partition-sip_capture_rest
ALTER TABLE sip_capture_rest_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));
