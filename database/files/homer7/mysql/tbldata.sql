-- name: create-logs-table
CREATE TABLE IF NOT EXISTS `hep_proto_100_logs_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);

-- name: create-report-table
CREATE TABLE IF NOT EXISTS `hep_proto_35_report_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);

-- name: create-rtcp-table
CREATE TABLE IF NOT EXISTS `hep_proto_5_rtcp_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);

-- name: create-call-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_call_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);

-- name: create-register-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_register_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);

-- name: create-default-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_default_DayDate` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `gid` int(10) NOT NULL DEFAULT 0,
  `cid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` varchar(3000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `cid` (`cid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION DayDate_pnr0 VALUES LESS THAN ( UNIX_TIMESTAMP('PartitionMin') )
);