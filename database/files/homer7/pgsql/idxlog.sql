-- name: index-logs-date
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_date ON hep_proto_100_default_{{date}}_{{time}} ((data_header->'create_date'));
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_sid ON hep_proto_100_default_{{date}}_{{time}} ((data_header->'sid'));
