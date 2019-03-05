-- name: index-default-create_date
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_create_date ON hep_proto_54_default_{{date}}_{{time}} (create_date);
-- name: index-default-sid
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_sid ON hep_proto_54_default_{{date}}_{{time}} (sid);

-- name: index-default-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_srcIp ON hep_proto_54_default_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-default-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_dstIp ON hep_proto_54_default_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-default-correlation_id
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_correlation_id ON hep_proto_54_default_{{date}}_{{time}} ((protocol_header->'correlation_id'));

-- name: index-default-called_number
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_called_number ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'called_number'->>'num'));
-- name: index-default-calling_number
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_calling_number ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'calling_number' ->> 'num'));
-- name: index-default-calling_party
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_calling_party ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'calling_party' ->> 'num'));
-- name: index-default-opc
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_opc ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'opc'));
-- name: index-default-dpc
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_dpc ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'dpc'));
-- name: index-default-cic
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_cic ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'cic'));
-- name: index-default-method
CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_msg_name ON hep_proto_54_default_{{date}}_{{time}} ((data_header->'msg_name'));
