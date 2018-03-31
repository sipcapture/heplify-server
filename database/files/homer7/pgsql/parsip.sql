-- name: create-partition-hep_proto_1_call
CREATE TABLE IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000 PARTITION OF hep_proto_1_call FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-hep_proto_1_register
CREATE TABLE IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000 PARTITION OF hep_proto_1_register FOR VALUES FROM ('StartTime') TO ('EndTime');
