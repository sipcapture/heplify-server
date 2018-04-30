-- name: index-report-date
CREATE INDEX IF NOT EXISTS report_capture_{{date}}_{{time}}_date ON report_capture_{{date}}_{{time}} (date);
-- name: index-report-correlation
CREATE INDEX IF NOT EXISTS report_capture_{{date}}_{{time}}_correlation_id ON report_capture_{{date}}_{{time}} (correlation_id);

-- name: index-rtcp-date
CREATE INDEX IF NOT EXISTS rtcp_capture_{{date}}_{{time}}_date ON rtcp_capture_{{date}}_{{time}} (date);
-- name: index-rtcp-correlation
CREATE INDEX IF NOT EXISTS rtcp_capture_{{date}}_{{time}}_correlation_id ON rtcp_capture_{{date}}_{{time}} (correlation_id);
