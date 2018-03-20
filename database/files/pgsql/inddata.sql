-- name: index-logs-date
CREATE INDEX IF NOT EXISTS logs_capture_all_PartitionName_StartTime_date ON logs_capture_all_PartitionName_StartTime (date);
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS logs_capture_all_PartitionName_StartTime_correlation_id ON logs_capture_all_PartitionName_StartTime (correlation_id);

-- name: index-report-date
CREATE INDEX IF NOT EXISTS report_capture_all_PartitionName_StartTime_date ON report_capture_all_PartitionName_StartTime (date);
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS report_capture_all_PartitionName_StartTime_correlation_id ON report_capture_all_PartitionName_StartTime (correlation_id);

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS rtcp_capture_all_PartitionName_StartTime_date ON rtcp_capture_all_PartitionName_StartTime (date);
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS rtcp_capture_all_PartitionName_StartTime_correlation_id ON rtcp_capture_all_PartitionName_StartTime (correlation_id);

-- name: index-call-date
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_date ON sip_capture_call_PartitionName_StartTime (date);
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_ruri_user ON sip_capture_call_PartitionName_StartTime (ruri_user);
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_from_user ON sip_capture_call_PartitionName_StartTime (from_user);
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_to_user ON sip_capture_call_PartitionName_StartTime (to_user);
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_pid_user ON sip_capture_call_PartitionName_StartTime (pid_user);
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_auth_user ON sip_capture_call_PartitionName_StartTime (auth_user);
-- name: index-call-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_callid_aleg ON sip_capture_call_PartitionName_StartTime (callid_aleg);
-- name: index-call-callid
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_callid ON sip_capture_call_PartitionName_StartTime (callid);
-- name: index-call-method
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_method ON sip_capture_call_PartitionName_StartTime (method);
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_source_ip ON sip_capture_call_PartitionName_StartTime (source_ip);
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_PartitionName_StartTime_destination_ip ON sip_capture_call_PartitionName_StartTime (destination_ip);

-- name: index-registration-date
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_date ON sip_capture_registration_PartitionName_StartTime (date);
-- name: index-registration-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_ruri_user ON sip_capture_registration_PartitionName_StartTime (ruri_user);
-- name: index-registration-from_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_from_user ON sip_capture_registration_PartitionName_StartTime (from_user);
-- name: index-registration-to_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_to_user ON sip_capture_registration_PartitionName_StartTime (to_user);
-- name: index-registration-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_pid_user ON sip_capture_registration_PartitionName_StartTime (pid_user);
-- name: index-registration-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_auth_user ON sip_capture_registration_PartitionName_StartTime (auth_user);
-- name: index-registration-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_callid_aleg ON sip_capture_registration_PartitionName_StartTime (callid_aleg);
-- name: index-registration-callid
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_callid ON sip_capture_registration_PartitionName_StartTime (callid);
-- name: index-registration-method
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_method ON sip_capture_registration_PartitionName_StartTime (method);
-- name: index-registration-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_source_ip ON sip_capture_registration_PartitionName_StartTime (source_ip);
-- name: index-registration-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_PartitionName_StartTime_destination_ip ON sip_capture_registration_PartitionName_StartTime (destination_ip);
