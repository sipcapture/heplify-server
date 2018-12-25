-- name: create-report-table
CREATE TABLE IF NOT EXISTS `hep_proto_35_default_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` text NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `sid` (`sid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);

-- name: create-rtcp-table
CREATE TABLE IF NOT EXISTS `hep_proto_5_default_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` text NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `sid` (`sid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);