-- name: index-call-create_date
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_create_date ON hep_proto_1_call_{{date}}_{{time}} (create_date);
-- name: index-call-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_sid ON hep_proto_1_call_{{date}}_{{time}} (sid);

-- name: index-call-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_srcIp ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-call-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_dstIp ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-call-correlation_id
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_correlation_id ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->'correlation_id'));

-- name: index-call-ruri_domain
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_ruri_domain ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'ruri_domain'));
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_ruri_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'ruri_user'));
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_from_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'from_user'));
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_to_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'to_user'));
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_pid_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'pid_user'));
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_auth_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'auth_user'));
-- name: index-call-callid
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_callid ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'callid'));
-- name: index-call-method
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_method ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'method'));



-- name: index-register-create_date
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_create_date ON hep_proto_1_register_{{date}}_{{time}} (create_date);
-- name: index-register-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_sid ON hep_proto_1_register_{{date}}_{{time}} (sid);

-- name: index-register-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_srcIp ON hep_proto_1_register_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-register-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_dstIp ON hep_proto_1_register_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-register-correlation_id
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_correlation_id ON hep_proto_1_register_{{date}}_{{time}} ((protocol_header->'correlation_id'));

-- name: index-register-ruri_domain
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_ruri_domain ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'ruri_domain'));
-- name: index-register-ruri_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_ruri_user ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'ruri_user'));
-- name: index-register-from_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_from_user ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'from_user'));
-- name: index-register-to_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_to_user ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'to_user'));
-- name: index-register-pid_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_pid_user ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'pid_user'));
-- name: index-register-auth_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_auth_user ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'auth_user'));
-- name: index-register-callid
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_callid ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'callid'));
-- name: index-register-method
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_method ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'method'));



-- name: index-default-create_date
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_create_date ON hep_proto_1_default_{{date}}_{{time}} (create_date);
-- name: index-default-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_sid ON hep_proto_1_default_{{date}}_{{time}} (sid);

-- name: index-default-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_srcIp ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-default-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_dstIp ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-default-correlation_id
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_correlation_id ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->'correlation_id'));

-- name: index-default-ruri_domain
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_ruri_domain ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'ruri_domain'));
-- name: index-default-ruri_user
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_ruri_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'ruri_user'));
-- name: index-default-from_user
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_from_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'from_user'));
-- name: index-default-to_user
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_to_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'to_user'));
-- name: index-default-pid_user
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_pid_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'pid_user'));
-- name: index-default-auth_user
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_auth_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'auth_user'));
-- name: index-default-callid
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_callid ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'callid'));
-- name: index-default-method
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_method ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'method'));

