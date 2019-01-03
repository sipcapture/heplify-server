-- name: create-report-table
CREATE TABLE `report_capture_all_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `micro_ts` bigint(18) NOT NULL DEFAULT '0',
  `correlation_id` varchar(256) NOT NULL DEFAULT '',
  `source_ip` varchar(60) NOT NULL DEFAULT '',
  `source_port` int(10) NOT NULL DEFAULT 0,
  `destination_ip` varchar(60) NOT NULL DEFAULT '',
  `destination_port` int(10) NOT NULL DEFAULT 0,
  `proto` int(5) NOT NULL DEFAULT 0,
  `family` int(1) DEFAULT NULL,
  `type` int(5) NOT NULL DEFAULT 0,
  `node` varchar(125) NOT NULL DEFAULT '',
  `msg` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`date`),
  KEY `date` (`date`),
  KEY `correlationid` (`correlation_id`(255))
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);

-- name: create-rtcp-table
CREATE TABLE `rtcp_capture_all_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `micro_ts` bigint(18) NOT NULL DEFAULT '0',
  `correlation_id` varchar(256) NOT NULL DEFAULT '',
  `source_ip` varchar(60) NOT NULL DEFAULT '',
  `source_port` int(10) NOT NULL DEFAULT 0,
  `destination_ip` varchar(60) NOT NULL DEFAULT '',
  `destination_port` int(10) NOT NULL DEFAULT 0,
  `proto` int(5) NOT NULL DEFAULT 0,
  `family` int(1) DEFAULT NULL,
  `type` int(5) NOT NULL DEFAULT 0,
  `node` varchar(125) NOT NULL DEFAULT '',
  `msg` varchar(1500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`date`),
  KEY `date` (`date`),
  KEY `correlationid` (`correlation_id`(255))
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);