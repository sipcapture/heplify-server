-- name: index-report-date
CREATE INDEX IF NOT EXISTS hep_proto_35_report_DayDate_pnr0000_date ON hep_proto_35_report_DayDate_pnr0000 ((data_header->'create_date'));
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS hep_proto_35_report_DayDate_pnr0000_sid ON hep_proto_35_report_DayDate_pnr0000 ((data_header->'sid'));

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS hep_proto_5_rtcp_DayDate_pnr0000_date ON hep_proto_5_rtcp_DayDate_pnr0000 ((data_header->'create_date'));
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS hep_proto_5_rtcp_DayDate_pnr0000_sid ON hep_proto_5_rtcp_DayDate_pnr0000 ((data_header->'sid'));
