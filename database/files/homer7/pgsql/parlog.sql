-- name: create-partition-hep_proto_100_logs
CREATE TABLE IF NOT EXISTS hep_proto_100_logs_{{date}}_{{time}} PARTITION OF hep_proto_100_logs FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
