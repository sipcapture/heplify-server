-- name: create-partition-hep_proto_1_call
ALTER TABLE hep_proto_1_call_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-hep_proto_1_register
ALTER TABLE hep_proto_1_register_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-hep_proto_1_default
ALTER TABLE hep_proto_1_default_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));