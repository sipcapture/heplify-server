-- name: create-partition-hep_proto_1_call
ALTER TABLE hep_proto_1_call_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-hep_proto_1_register
ALTER TABLE hep_proto_1_register_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-hep_proto_1_default
ALTER TABLE hep_proto_1_default_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));