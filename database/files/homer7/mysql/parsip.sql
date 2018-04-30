-- name: create-partition-hep_proto_1_call
ALTER TABLE hep_proto_1_call_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-hep_proto_1_register
ALTER TABLE hep_proto_1_register_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-hep_proto_1_default
ALTER TABLE hep_proto_1_default_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));