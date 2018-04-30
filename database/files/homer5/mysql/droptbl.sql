-- name: drop-logs-table
DROP TABLE logs_capture_all_{{date}};

-- name: drop-report-table
DROP TABLE report_capture_all_{{date}};

-- name: drop-rtcp-table
DROP TABLE rtcp_capture_all_{{date}};

-- name: drop-call-table
DROP TABLE sip_capture_call_{{date}};

-- name: drop-registration-table
DROP TABLE sip_capture_registration_{{date}};

-- name: drop-rest-table
DROP TABLE sip_capture_rest_{{date}};
