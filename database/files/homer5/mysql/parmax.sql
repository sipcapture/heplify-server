-- name: create-partitionmax-logs_capture
ALTER TABLE logs_capture_all_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-report_capture
ALTER TABLE report_capture_all_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-rtcp_capture
ALTER TABLE rtcp_capture_all_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-sip_capture_call
ALTER TABLE sip_capture_call_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-sip_capture_registration
ALTER TABLE sip_capture_registration_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-sip_capture_rest
ALTER TABLE sip_capture_rest_TableDate ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);