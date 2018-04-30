-- name: create-partition-sip_capture_call
CREATE TABLE IF NOT EXISTS sip_capture_call_{{date}}_{{time}} PARTITION OF sip_capture_call FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-sip_capture_registration
CREATE TABLE IF NOT EXISTS sip_capture_registration_{{date}}_{{time}} PARTITION OF sip_capture_registration FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-sip_capture_rest
CREATE TABLE IF NOT EXISTS sip_capture_rest_{{date}}_{{time}} PARTITION OF sip_capture_rest FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
