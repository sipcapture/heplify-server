-- name: create-partitionmax-hep_proto_100_logs
ALTER TABLE hep_proto_100_logs_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_35_report
ALTER TABLE hep_proto_35_report_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_5_rtcp
ALTER TABLE hep_proto_5_rtcp_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_call
ALTER TABLE hep_proto_1_call_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_register
ALTER TABLE hep_proto_1_register_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_default
ALTER TABLE hep_proto_1_default_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);