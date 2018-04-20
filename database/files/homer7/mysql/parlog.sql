-- name: create-partition-hep_proto_100_logs
ALTER TABLE hep_proto_100_logs_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));
