-- name: create-partition-hep_proto_35_default
CREATE TABLE IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}} PARTITION OF hep_proto_35_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-hep_proto_5_default
CREATE TABLE IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}} PARTITION OF hep_proto_5_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
