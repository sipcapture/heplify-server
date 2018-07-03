-- name: create-partition-hep_proto_100_default
ALTER TABLE hep_proto_100_default_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
