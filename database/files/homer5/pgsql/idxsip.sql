-- name: index-call-date
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_date ON sip_capture_call_{{date}}_{{time}} (date);
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_ruri_user ON sip_capture_call_{{date}}_{{time}} (ruri_user);
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_from_user ON sip_capture_call_{{date}}_{{time}} (from_user);
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_to_user ON sip_capture_call_{{date}}_{{time}} (to_user);
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_pid_user ON sip_capture_call_{{date}}_{{time}} (pid_user);
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_auth_user ON sip_capture_call_{{date}}_{{time}} (auth_user);
-- name: index-call-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_callid_aleg ON sip_capture_call_{{date}}_{{time}} (callid_aleg);
-- name: index-call-callid
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_callid ON sip_capture_call_{{date}}_{{time}} (callid);
-- name: index-call-method
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_method ON sip_capture_call_{{date}}_{{time}} (method);
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_source_ip ON sip_capture_call_{{date}}_{{time}} (source_ip);
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_call_{{date}}_{{time}}_destination_ip ON sip_capture_call_{{date}}_{{time}} (destination_ip);

-- name: index-registration-date
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_date ON sip_capture_registration_{{date}}_{{time}} (date);
-- name: index-registration-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_ruri_user ON sip_capture_registration_{{date}}_{{time}} (ruri_user);
-- name: index-registration-from_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_from_user ON sip_capture_registration_{{date}}_{{time}} (from_user);
-- name: index-registration-to_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_to_user ON sip_capture_registration_{{date}}_{{time}} (to_user);
-- name: index-registration-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_pid_user ON sip_capture_registration_{{date}}_{{time}} (pid_user);
-- name: index-registration-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_auth_user ON sip_capture_registration_{{date}}_{{time}} (auth_user);
-- name: index-registration-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_callid_aleg ON sip_capture_registration_{{date}}_{{time}} (callid_aleg);
-- name: index-registration-callid
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_callid ON sip_capture_registration_{{date}}_{{time}} (callid);
-- name: index-registration-method
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_method ON sip_capture_registration_{{date}}_{{time}} (method);
-- name: index-registration-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_source_ip ON sip_capture_registration_{{date}}_{{time}} (source_ip);
-- name: index-registration-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_registration_{{date}}_{{time}}_destination_ip ON sip_capture_registration_{{date}}_{{time}} (destination_ip);

-- name: index-rest-date
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_date ON sip_capture_rest_{{date}}_{{time}} (date);
-- name: index-rest-ruri_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_ruri_user ON sip_capture_rest_{{date}}_{{time}} (ruri_user);
-- name: index-rest-from_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_from_user ON sip_capture_rest_{{date}}_{{time}} (from_user);
-- name: index-rest-to_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_to_user ON sip_capture_rest_{{date}}_{{time}} (to_user);
-- name: index-rest-pid_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_pid_user ON sip_capture_rest_{{date}}_{{time}} (pid_user);
-- name: index-rest-auth_user
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_auth_user ON sip_capture_rest_{{date}}_{{time}} (auth_user);
-- name: index-rest-callid_aleg
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_callid_aleg ON sip_capture_rest_{{date}}_{{time}} (callid_aleg);
-- name: index-rest-callid
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_callid ON sip_capture_rest_{{date}}_{{time}} (callid);
-- name: index-rest-method
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_method ON sip_capture_rest_{{date}}_{{time}} (method);
-- name: index-rest-source_ip
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_source_ip ON sip_capture_rest_{{date}}_{{time}} (source_ip);
-- name: index-rest-destination_ip
CREATE INDEX IF NOT EXISTS sip_capture_rest_{{date}}_{{time}}_destination_ip ON sip_capture_rest_{{date}}_{{time}} (destination_ip);
