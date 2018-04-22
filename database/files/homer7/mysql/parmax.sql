-- name: create-partitionmax-hep_proto_100_logs
ALTER TABLE hep_proto_100_logs_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_35_report
ALTER TABLE hep_proto_35_report_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_5_rtcp
ALTER TABLE hep_proto_5_rtcp_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_call
ALTER TABLE hep_proto_1_call_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_register
ALTER TABLE hep_proto_1_register_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_default
ALTER TABLE hep_proto_1_default_DayDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);