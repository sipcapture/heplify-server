-- name: index-report-date
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_date ON hep_proto_35_default_{{date}}_{{time}} ((data_header->'create_date'));
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_sid ON hep_proto_35_default_{{date}}_{{time}} ((data_header->'sid'));

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_date ON hep_proto_5_default_{{date}}_{{time}} ((data_header->'create_date'));
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_sid ON hep_proto_5_default_{{date}}_{{time}} ((data_header->'sid'));
