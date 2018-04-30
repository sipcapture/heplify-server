-- name: create-partition-hep_proto_35_report
CREATE TABLE IF NOT EXISTS hep_proto_35_report_{{date}}_{{time}} PARTITION OF hep_proto_35_report FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');

-- name: create-partition-hep_proto_5_rtcp
CREATE TABLE IF NOT EXISTS hep_proto_5_rtcp_{{date}}_{{time}} PARTITION OF hep_proto_5_rtcp FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');
