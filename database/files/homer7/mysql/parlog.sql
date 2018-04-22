-- name: create-partition-hep_proto_100_logs
ALTER TABLE hep_proto_100_logs_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));
