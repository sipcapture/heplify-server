-- name: create-partition-hep_proto_100_logs
ALTER TABLE hep_proto_100_logs_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
