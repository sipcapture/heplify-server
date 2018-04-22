-- name: index-logs-date
CREATE INDEX IF NOT EXISTS hep_proto_100_logs_DayDate_pnr0000_date ON hep_proto_100_logs_DayDate_pnr0000 ((data_header->'create_date'));
-- name: index-logs-correlation
CREATE INDEX IF NOT EXISTS hep_proto_100_logs_DayDate_pnr0000_cid ON hep_proto_100_logs_DayDate_pnr0000 ((data_header->'cid'));
