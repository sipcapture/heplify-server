-- name: index-logs-create_date
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_create_date ON hep_proto_100_default_{{date}}_{{time}} (create_date);
-- name: index-logs-sid
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_sid ON hep_proto_100_default_{{date}}_{{time}} (sid);

-- name: index-logs-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_srcIp ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-logs-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_dstIp ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-logs-payloadType
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_payloadType ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->'payloadType'));