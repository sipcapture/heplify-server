-- name: alter-logs-table-partition
ALTER TABLE logs_capture_all_TableDate_pPartitionName00 ADD CONSTRAINT chk_logs_capture_all_TableDate_pPartitionName00 CHECK (date < TO_TIMESTAMP('PartitionDate 00:00:00','YYYY-MM-DD HH:MI:SS'));

-- name: alter-report-table-partition
ALTER TABLE report_capture_all_TableDate_pPartitionName00 ADD CONSTRAINT chk_report_capture_all_TableDate_pPartitionName00 CHECK (date < TO_TIMESTAMP('PartitionDate 00:00:00','YYYY-MM-DD HH:MI:SS')); 

-- name: alter-rtcp-table-partition
ALTER TABLE rtcp_capture_all_TableDate_pPartitionName00 ADD CONSTRAINT chk_rtcp_capture_all_TableDate_pPartitionName00 CHECK (date < TO_TIMESTAMP('PartitionDate 00:00:00','YYYY-MM-DD HH:MI:SS')); 

-- name: alter-call-table-partition
ALTER TABLE sip_capture_call_TableDate_pPartitionName00 ADD CONSTRAINT chk_sip_capture_call_TableDate_pPartitionName00 CHECK (date < TO_TIMESTAMP('PartitionDate 00:00:00','YYYY-MM-DD HH:MI:SS')); 

-- name: alter-registration-table-partition
ALTER TABLE sip_capture_registration_TableDate_pPartitionName00 ADD CONSTRAINT chk_sip_capture_registration_TableDate_pPartitionName00 CHECK (date < TO_TIMESTAMP('PartitionDate 00:00:00','YYYY-MM-DD HH:MI:SS')); 
