-- name: create-partition-hep_proto_54_default
CREATE TABLE IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}} PARTITION OF hep_proto_54_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
