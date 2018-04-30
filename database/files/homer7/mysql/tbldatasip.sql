-- name: create-call-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_call_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` text NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `sid` (`sid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);

-- name: create-register-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_register_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` text NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `sid` (`sid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
);

-- name: create-default-table
CREATE TABLE IF NOT EXISTS `hep_proto_1_default_{{date}}` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `sid` varchar(256) NOT NULL DEFAULT '',
  `create_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `protocol_header` json NOT NULL,
  `data_header` json NOT NULL,
  `raw` text NOT NULL DEFAULT '',
  PRIMARY KEY (`id`,`create_date`),
  KEY `create_date` (`create_date`),
  KEY `sid` (`sid`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
PARTITION BY RANGE ( UNIX_TIMESTAMP(`create_date`) ) (
    PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )

