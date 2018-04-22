-- name: index-call-date
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_date ON sip_capture_call_DayDate_pnr0000 (date);
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_ruri_user ON sip_capture_call_DayDate_pnr0000 (ruri_user);
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_from_user ON sip_capture_call_DayDate_pnr0000 (from_user);
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_to_user ON sip_capture_call_DayDate_pnr0000 (to_user);
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_pid_user ON sip_capture_call_DayDate_pnr0000 (pid_user);
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_auth_user ON sip_capture_call_DayDate_pnr0000 (auth_user);
-- name: index-call-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_callid_aleg ON sip_capture_call_DayDate_pnr0000 (callid_aleg);
-- name: index-call-callid
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_callid ON sip_capture_call_DayDate_pnr0000 (callid);
-- name: index-call-method
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_method ON sip_capture_call_DayDate_pnr0000 (method);
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_source_ip ON sip_capture_call_DayDate_pnr0000 (source_ip);
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_DayDate_pnr0000_destination_ip ON sip_capture_call_DayDate_pnr0000 (destination_ip);

-- name: index-registration-date
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_date ON sip_capture_registration_DayDate_pnr0000 (date);
-- name: index-registration-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_ruri_user ON sip_capture_registration_DayDate_pnr0000 (ruri_user);
-- name: index-registration-from_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_from_user ON sip_capture_registration_DayDate_pnr0000 (from_user);
-- name: index-registration-to_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_to_user ON sip_capture_registration_DayDate_pnr0000 (to_user);
-- name: index-registration-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_pid_user ON sip_capture_registration_DayDate_pnr0000 (pid_user);
-- name: index-registration-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_auth_user ON sip_capture_registration_DayDate_pnr0000 (auth_user);
-- name: index-registration-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_callid_aleg ON sip_capture_registration_DayDate_pnr0000 (callid_aleg);
-- name: index-registration-callid
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_callid ON sip_capture_registration_DayDate_pnr0000 (callid);
-- name: index-registration-method
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_method ON sip_capture_registration_DayDate_pnr0000 (method);
-- name: index-registration-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_source_ip ON sip_capture_registration_DayDate_pnr0000 (source_ip);
-- name: index-registration-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_DayDate_pnr0000_destination_ip ON sip_capture_registration_DayDate_pnr0000 (destination_ip);

-- name: index-rest-date
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_date ON sip_capture_rest_DayDate_pnr0000 (date);
-- name: index-rest-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_ruri_user ON sip_capture_rest_DayDate_pnr0000 (ruri_user);
-- name: index-rest-from_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_from_user ON sip_capture_rest_DayDate_pnr0000 (from_user);
-- name: index-rest-to_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_to_user ON sip_capture_rest_DayDate_pnr0000 (to_user);
-- name: index-rest-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_pid_user ON sip_capture_rest_DayDate_pnr0000 (pid_user);
-- name: index-rest-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_auth_user ON sip_capture_rest_DayDate_pnr0000 (auth_user);
-- name: index-rest-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_callid_aleg ON sip_capture_rest_DayDate_pnr0000 (callid_aleg);
-- name: index-rest-callid
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_callid ON sip_capture_rest_DayDate_pnr0000 (callid);
-- name: index-rest-method
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_method ON sip_capture_rest_DayDate_pnr0000 (method);
-- name: index-rest-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_source_ip ON sip_capture_rest_DayDate_pnr0000 (source_ip);
-- name: index-rest-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_rest_DayDate_pnr0000_destination_ip ON sip_capture_rest_DayDate_pnr0000 (destination_ip);
