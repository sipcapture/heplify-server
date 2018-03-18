-- name: index-logs-date
CREATE INDEX IF NOT EXISTS logs_capture_all_PartitionName_date ON logs_capture_all_PartitionName (date);
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS logs_capture_all_PartitionName_correlation_id ON logs_capture_all_PartitionName (correlation_id);

-- name: index-report-date
CREATE INDEX IF NOT EXISTS report_capture_all_PartitionName_date ON report_capture_all_PartitionName (date);
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS report_capture_all_PartitionName_correlation_id ON report_capture_all_PartitionName (correlation_id);

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS rtcp_capture_all_PartitionName_date ON rtcp_capture_all_PartitionName (date);
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS rtcp_capture_all_PartitionName_correlation_id ON rtcp_capture_all_PartitionName (correlation_id);

-- name: index-call-date
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_date ON sip_capture_call_PartitionName (date);
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_ruri_user ON sip_capture_call_PartitionName (ruri_user);
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_from_user ON sip_capture_call_PartitionName (from_user);
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_to_user ON sip_capture_call_PartitionName (to_user);
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_pid_user ON sip_capture_call_PartitionName (pid_user);
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_auth_user ON sip_capture_call_PartitionName (auth_user);
-- name: index-call-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_callid_aleg ON sip_capture_call_PartitionName (callid_aleg);
-- name: index-call-callid
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_callid ON sip_capture_call_PartitionName (callid);
-- name: index-call-method
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_method ON sip_capture_call_PartitionName (method);
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_source_ip ON sip_capture_call_PartitionName (source_ip);
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_destination_ip ON sip_capture_call_PartitionName (destination_ip);

-- name: index-registration-date
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_date ON sip_capture_registration_PartitionName (date);
-- name: index-registration-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_ruri_user ON sip_capture_registration_PartitionName (ruri_user);
-- name: index-registration-from_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_from_user ON sip_capture_registration_PartitionName (from_user);
-- name: index-registration-to_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_to_user ON sip_capture_registration_PartitionName (to_user);
-- name: index-registration-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_pid_user ON sip_capture_registration_PartitionName (pid_user);
-- name: index-registration-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_auth_user ON sip_capture_registration_PartitionName (auth_user);
-- name: index-registration-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_callid_aleg ON sip_capture_registration_PartitionName (callid_aleg);
-- name: index-registration-callid
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_callid ON sip_capture_registration_PartitionName (callid);
-- name: index-registration-method
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_method ON sip_capture_registration_PartitionName (method);
-- name: index-registration-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_source_ip ON sip_capture_registration_PartitionName (source_ip);
-- name: index-registration-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_destination_ip ON sip_capture_registration_PartitionName (destination_ip);
