-- name: create-partition-hep_proto_35_report
ALTER TABLE hep_proto_35_report_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));

-- name: create-partition-hep_proto_5_rtcp
ALTER TABLE hep_proto_5_rtcp_TableDate ADD PARTITION (PARTITION pPartitionName_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('StartTime') ));
