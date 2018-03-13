-- name: create-logs-table-inh
CREATE TABLE logs_capture_all_20110111_p2013082901() INHERITS (logs_capture_all_20110111);

-- name: create-report-table-inh
CREATE TABLE report_capture_all_20110111_p2013082901() INHERITS (report_capture_all_20110111);

-- name: create-rtcp-table-inh
CREATE TABLE rtcp_capture_all_20110111_p2013082901() INHERITS (rtcp_capture_all_20110111);

-- name: create-call-table-inh
CREATE TABLE sip_capture_call_20110111_p2013082901() INHERITS (sip_capture_call_20110111);

-- name: create-registration-table-inh
CREATE TABLE sip_capture_registration_20110111_p2013082901() INHERITS (sip_capture_registration_20110111);
