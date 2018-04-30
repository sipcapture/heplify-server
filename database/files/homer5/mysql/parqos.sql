-- name: create-partition-report_capture
ALTER TABLE report_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));

-- name: create-partition-rtcp_capture
ALTER TABLE rtcp_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
