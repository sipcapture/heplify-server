-- name: create-partition-logs_capture
ALTER TABLE logs_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));
