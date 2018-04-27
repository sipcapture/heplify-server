-- name: create-partition-hep_proto_35_report
ALTER TABLE hep_proto_35_report_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));

-- name: create-partition-hep_proto_5_rtcp
ALTER TABLE hep_proto_5_rtcp_DayDate ADD PARTITION (PARTITION DayDate_pnr0000 VALUES LESS THAN ( UNIX_TIMESTAMP('EndTime') ));
