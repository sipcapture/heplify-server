-- name: index-logs-date
CREATE INDEX logs_capture_date ON "logs_capture" (date);
-- name: index-logs-correlation
CREATE INDEX logs_capture_correlation ON "logs_capture" (correlation_id);

-- name: index-report-date
CREATE INDEX report_capture_all_TableDate_date ON "report_capture_all_TableDate" (date);
-- name: index-report-correlation
CREATE INDEX report_capture_all_TableDate_correlation ON "report_capture_all_TableDate" (correlation_id);

-- name: index-rtcp-date
CREATE INDEX rtcp_capture_all_TableDate_date ON "rtcp_capture_all_TableDate" (date);
-- name: index-rtcp-correlation
CREATE INDEX rtcp_capture_all_TableDate_correlation ON "rtcp_capture_all_TableDate" (correlation_id);

-- name: index-call-ruri_user
CREATE INDEX sip_capture_call_TableDate_ruri_user ON "sip_capture_call_TableDate" (ruri_user);
-- name: index-call-from_user
CREATE INDEX sip_capture_call_TableDate_from_user ON "sip_capture_call_TableDate" (from_user);
-- name: index-call-to_user
CREATE INDEX sip_capture_call_TableDate_to_user ON "sip_capture_call_TableDate" (to_user);
-- name: index-call-pid_user
CREATE INDEX sip_capture_call_TableDate_pid_user ON "sip_capture_call_TableDate" (pid_user);
-- name: index-call-auth_user
CREATE INDEX sip_capture_call_TableDate_auth_user ON "sip_capture_call_TableDate" (auth_user);
-- name: index-call-callid_aleg
CREATE INDEX sip_capture_call_TableDate_callid_aleg ON "sip_capture_call_TableDate" (callid_aleg);
-- name: index-call-date
CREATE INDEX sip_capture_call_TableDate_date ON "sip_capture_call_TableDate" (date);
-- name: index-call-callid
CREATE INDEX sip_capture_call_TableDate_callid ON "sip_capture_call_TableDate" (callid);

-- name: index-registration-ruri_user
CREATE INDEX sip_capture_registration_TableDate_ruri_user ON "sip_capture_registration_TableDate" (ruri_user);
-- name: index-registration-from_user
CREATE INDEX sip_capture_registration_TableDate_from_user ON "sip_capture_registration_TableDate" (from_user);
-- name: index-registration-to_user
CREATE INDEX sip_capture_registration_TableDate_to_user ON "sip_capture_registration_TableDate" (to_user);
-- name: index-registration-pid_user
CREATE INDEX sip_capture_registration_TableDate_pid_user ON "sip_capture_registration_TableDate" (pid_user);
-- name: index-registration-auth_user
CREATE INDEX sip_capture_registration_TableDate_auth_user ON "sip_capture_registration_TableDate" (auth_user);
-- name: index-registration-callid_aleg
CREATE INDEX sip_capture_registration_TableDate_callid_aleg ON "sip_capture_registration_TableDate" (callid_aleg);
-- name: index-registration-date
CREATE INDEX sip_capture_registration_TableDate_date ON "sip_capture_registration_TableDate" (date);
-- name: index-registration-callid
CREATE INDEX sip_capture_registration_TableDate_callid ON "sip_capture_registration_TableDate" (callid);
