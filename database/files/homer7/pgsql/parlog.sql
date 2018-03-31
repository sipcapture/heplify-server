-- name: create-partition-hep_proto_100_logs
CREATE TABLE IF NOT EXISTS hep_proto_100_logs_PartitionName_pnr0000 PARTITION OF hep_proto_100_logs FOR VALUES FROM ('StartTime') TO ('EndTime');
