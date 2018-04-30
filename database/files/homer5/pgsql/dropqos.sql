-- name: drop-partition-report_capture
DROP TABLE report_capture_{{date}}_{{time}};

-- name: drop-partition-rtcp_capture
DROP TABLE rtcp_capture_{{date}}_{{time}};
