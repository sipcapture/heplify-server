-- name: create-logs-table-partition
CREATE TABLE logs_capture_pPartitionName00() INHERITS (logs_capture);

-- name: create-report-table-partition
CREATE TABLE report_capture_all_TableDate_pPartitionName00() INHERITS (report_capture_all_TableDate);

-- name: create-rtcp-table-partition
CREATE TABLE rtcp_capture_all_TableDate_pPartitionName00() INHERITS (rtcp_capture_all_TableDate);

-- name: create-call-table-partition
CREATE TABLE sip_capture_call_TableDate_pPartitionName00() INHERITS (sip_capture_call_TableDate);

-- name: create-registration-table-partition
CREATE TABLE sip_capture_registration_TableDate_pPartitionName00() INHERITS (sip_capture_registration_TableDate);
