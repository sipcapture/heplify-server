-- name: drop-logs-table
DROP TABLE hep_proto_100_logs_{{date}};

-- name: drop-report-table
DROP TABLE hep_proto_35_report_{{date}};

-- name: drop-rtcp-table
DROP TABLE hep_proto_5_rtcp_{{date}};

-- name: drop-call-table
DROP TABLE hep_proto_1_call_{{date}};

-- name: drop-register-table
DROP TABLE hep_proto_1_register_{{date}};

-- name: drop-default-table
DROP TABLE hep_proto_1_default_{{date}};