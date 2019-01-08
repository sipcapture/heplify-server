-- name: index-report-create_date
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_create_date ON hep_proto_35_default_{{date}}_{{time}} (create_date);
-- name: index-report-sid
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_sid ON hep_proto_35_default_{{date}}_{{time}} (sid);

-- name: index-report-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_srcIp ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-report-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_dstIp ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-report-payloadType
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_payloadType ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->'payloadType'));



-- name: index-rtcp-create_date
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_create_date ON hep_proto_5_default_{{date}}_{{time}} (create_date);
-- name: index-rtcp-sid
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_sid ON hep_proto_5_default_{{date}}_{{time}} (sid);

-- name: index-rtcp-srcIp
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_srcIp ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->'srcIp'));
-- name: index-rtcp-dstIp
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_dstIp ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->'dstIp'));
-- name: index-rtcp-payloadType
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_payloadType ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->'payloadType'));
