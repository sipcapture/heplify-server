-- name: create-partition-hep_proto_35_default
ALTER TABLE hep_proto_35_default_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-hep_proto_5_default
ALTER TABLE hep_proto_5_default_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
