-- name: alter-logs-table-chk
ALTER TABLE logs_capture_all_20110111_p2013082901 ADD CONSTRAINT chk_logs_capture_all_20110111_p2013082901 CHECK (date < to_timestamp(1377734400));

-- name: alter-report-table-chk
ALTER TABLE report_capture_all_20110111_p2013082901 ADD CONSTRAINT chk_report_capture_all_20110111_p2013082901 CHECK (date < to_timestamp(1377734400)); 

-- name: alter-rtcp-table-chk
ALTER TABLE rtcp_capture_all_20110111_p2013082901 ADD CONSTRAINT chk_rtcp_capture_all_20110111_p2013082901 CHECK (date < to_timestamp(1377734400)); 

-- name: alter-call-table-chk
ALTER TABLE sip_capture_call_20110111_p2013082901 ADD CONSTRAINT chk_sip_capture_call_20110111_p2013082901 CHECK (date < to_timestamp(1377734400)); 

-- name: alter-registration-table-chk
ALTER TABLE sip_capture_registration_20110111_p2013082901 ADD CONSTRAINT chk_sip_capture_registration_20110111_p2013082901 CHECK (date < to_timestamp(1377734400)); 
