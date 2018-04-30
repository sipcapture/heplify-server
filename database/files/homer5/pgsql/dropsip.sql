-- name: drop-partition-sip_capture_call
DROP TABLE sip_capture_call_{{date}}_{{time}};

-- name: drop-partition-sip_capture_registration
DROP TABLE sip_capture_registration_{{date}}_{{time}};

-- name: drop-partition-sip_capture_rest
DROP TABLE sip_capture_rest_{{date}}_{{time}};
