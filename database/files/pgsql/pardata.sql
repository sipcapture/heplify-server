-- name: create-logs-table-inh
CREATE TABLE logs_capture_all_TableDate_pPartitionName00() INHERITS (logs_capture_all_TableDate);

-- name: create-report-table-inh
CREATE TABLE report_capture_all_TableDate_pPartitionName00() INHERITS (report_capture_all_TableDate);

-- name: create-rtcp-table-inh
CREATE TABLE rtcp_capture_all_TableDate_pPartitionName00() INHERITS (rtcp_capture_all_TableDate);

-- name: create-call-table-inh
CREATE TABLE sip_capture_call_TableDate_pPartitionName00() INHERITS (sip_capture_call_TableDate);

-- name: create-registration-table-inh
CREATE TABLE sip_capture_registration_TableDate_pPartitionName00() INHERITS (sip_capture_registration_TableDate);
