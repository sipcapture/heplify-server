-- name: index-logs-date
CREATE INDEX IF NOT EXISTS hep_proto_100_logs_PartitionName_pnr0000_date ON hep_proto_100_logs_PartitionName_pnr0000 ((data_header->'date'));
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS hep_proto_100_logs_PartitionName_pnr0000_cid ON hep_proto_100_logs_PartitionName_pnr0000 ((data_header->'cid'));

-- name: index-report-date
CREATE INDEX IF NOT EXISTS hep_proto_35_report_PartitionName_pnr0000_date ON hep_proto_35_report_PartitionName_pnr0000 ((data_header->'date'));
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS hep_proto_35_report_PartitionName_pnr0000_cid ON hep_proto_35_report_PartitionName_pnr0000 ((data_header->'cid'));

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS hep_proto_5_rtcp_PartitionName_pnr0000_date ON hep_proto_5_rtcp_PartitionName_pnr0000 ((data_header->'date'));
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS hep_proto_5_rtcp_PartitionName_pnr0000_cid ON hep_proto_5_rtcp_PartitionName_pnr0000 ((data_header->'cid'));

-- name: index-call-date
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_date ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'date'));
-- name: index-call-ruri_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_ruri_user ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'ruri_user'));
-- name: index-call-from_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_from_user ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'from_user'));
-- name: index-call-to_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_to_user ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'to_user'));
-- name: index-call-pid_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_pid_user ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'pid_user'));
-- name: index-call-auth_user
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_auth_user ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'auth_user'));
-- name: index-call-cid
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_cid ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'cid'));
-- name: index-call-method
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_method ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'method'));
-- name: index-call-source_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_source_ip ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'source_ip'));
-- name: index-call-destination_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_call_PartitionName_pnr0000_destination_ip ON hep_proto_1_call_PartitionName_pnr0000 ((data_header->'destination_ip'));

-- name: index-registration-date
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_date ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'date'));
-- name: index-registration-ruri_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_ruri_user ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'ruri_user'));
-- name: index-registration-from_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_from_user ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'from_user'));
-- name: index-registration-to_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_to_user ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'to_user'));
-- name: index-registration-pid_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_pid_user ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'pid_user'));
-- name: index-registration-auth_user
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_auth_user ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'auth_user'));
-- name: index-registration-cid
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_cid ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'cid'));
-- name: index-registration-method
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_method ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'method'));
-- name: index-registration-source_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_source_ip ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'source_ip'));
-- name: index-registration-destination_ip
CREATE INDEX IF NOT EXISTS hep_proto_1_register_PartitionName_pnr0000_destination_ip ON hep_proto_1_register_PartitionName_pnr0000 ((data_header->'destination_ip'));
