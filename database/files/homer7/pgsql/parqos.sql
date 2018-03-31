-- name: create-partition-hep_proto_35_report
CREATE TABLE IF NOT EXISTS hep_proto_35_report_PartitionName_pnr0000 PARTITION OF hep_proto_35_report FOR VALUES FROM ('StartTime') TO ('EndTime');

-- name: create-partition-hep_proto_5_rtcp
CREATE TABLE IF NOT EXISTS hep_proto_5_rtcp_PartitionName_pnr0000 PARTITION OF hep_proto_5_rtcp FOR VALUES FROM ('StartTime') TO ('EndTime');
