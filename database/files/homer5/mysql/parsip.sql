-- name: create-partition-sip_capture_call
ALTER TABLE sip_capture_call_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-sip_capture_registration
ALTER TABLE sip_capture_registration_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-sip_capture_rest
ALTER TABLE sip_capture_rest_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
