-- name: index-call-date
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_date ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'create_date'));
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
-- name: index-call-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_sid ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'sid'));
-- name: index-call-method
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_method ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'method'));
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_source_ip ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'source_ip'));
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_destination_ip ON hep_proto_1_call_{{date}}_{{time}} ((data_header->'destination_ip'));

-- name: index-register-date
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_date ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'create_date'));
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
-- name: index-register-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_sid ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'sid'));
-- name: index-register-method
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_method ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'method'));
-- name: index-register-source_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_source_ip ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'source_ip'));
-- name: index-register-destination_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}}_destination_ip ON hep_proto_1_register_{{date}}_{{time}} ((data_header->'destination_ip'));

-- name: index-default-date
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_date ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'create_date'));
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
-- name: index-default-sid
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_sid ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'sid'));
-- name: index-default-method
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_method ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'method'));
-- name: index-default-source_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_source_ip ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'source_ip'));
-- name: index-default-destination_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_destination_ip ON hep_proto_1_default_{{date}}_{{time}} ((data_header->'destination_ip'));
