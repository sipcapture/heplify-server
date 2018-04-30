-- name: create-partition-hep_proto_1_call
CREATE TABLE IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}} PARTITION OF hep_proto_1_call FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-hep_proto_1_register
CREATE TABLE IF NOT EXISTS hep_proto_1_register_{{date}}_{{time}} PARTITION OF hep_proto_1_register FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-hep_proto_1_default
CREATE TABLE IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}} PARTITION OF hep_proto_1_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
