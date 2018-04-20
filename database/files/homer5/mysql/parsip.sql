-- name: create-partition-sip_capture_call
ALTER TABLE sip_capture_call_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));

-- name: create-partition-sip_capture_registration
ALTER TABLE sip_capture_registration_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));
