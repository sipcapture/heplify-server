-- name: create-partition-hep_proto_35_report
ALTER TABLE hep_proto_35_report_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-hep_proto_5_rtcp
ALTER TABLE hep_proto_5_rtcp_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
