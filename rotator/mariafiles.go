package rotator

var (
	selectlogmaria      = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'logs_capture_all_%' and TABLE_NAME < 'logs_capture_all_{{date}}';"
	selectreportmaria   = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'report_capture_all_%' and TABLE_NAME < 'report_capture_all_{{date}}';"
	selectrtcpmaria     = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'rtcp_capture_all_%' and TABLE_NAME < 'rtcp_capture_all_{{date}}';"
	selectcallmaria     = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'sip_capture_call_%' and TABLE_NAME < 'sip_capture_call_{{date}}';"
	selectregistermaria = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'sip_capture_registration_%' and TABLE_NAME < 'sip_capture_registration_{{date}}';"
	selectdefaultmaria  = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME LIKE 'sip_capture_rest_%' and TABLE_NAME < 'sip_capture_rest_{{date}}';"
)

var (
	droplogmaria      = "DROP TABLE IF EXISTS {{partName}};"
	dropreportmaria   = "DROP TABLE IF EXISTS {{partName}};"
	droprtcpmaria     = "DROP TABLE IF EXISTS {{partName}};"
	dropcallmaria     = "DROP TABLE IF EXISTS {{partName}};"
	dropregistermaria = "DROP TABLE IF EXISTS {{partName}};"
	dropdefaultmaria  = "DROP TABLE IF EXISTS {{partName}};"
)

var insconfmaria = []string{
	`INSERT INTO alias (id, gid, ip, port, capture_id, alias, status, created) VALUES
	(1, 10, '192.168.0.30', 0, 'homer01', 'proxy01', 1, '2014-06-12 20:36:50');`,

	"INSERT INTO `group` (`gid`, `name`) VALUES (10, 'Administrator');",

	`INSERT INTO node (id, host, dbname, dbport, dbusername, dbpassword, dbtables, name, status) VALUES
	(1, '127.0.0.1', 'homer_data', '3306', 'homer_user', 'homer_password', 'sip_capture', 'homer01', 1);`,

	`INSERT INTO setting (id, uid, param_name, param_value, valid_param_from, valid_param_to, param_prio, active) VALUES
	(1, 1, 'timerange', '{"from":"2015-05-26T18:34:42.654Z","to":"2015-05-26T18:44:42.654Z"}', '2012-01-01 00:00:00', '2032-12-01 00:00:00', 10, 1);`,

	`INSERT INTO user (uid, gid, grp, username, password, firstname, lastname, email, department, regdate, lastvisit, active) VALUES
	(1, 10, 'users,admins', 'admin', PASSWORD('test123'), 'Admin', 'Admin', 'admin@test.com', 'Voice Enginering', '2012-01-19 00:00:00', '2015-05-29 07:17:35', 1);`,

	`INSERT INTO user_menu (id, name, alias, icon, weight, active) VALUES
	('_1426001444630', 'SIP Search', 'search', 'fa-search', 10, 1),
	('_1427728371642', 'Home', 'home', 'fa-home', 1, 1);`,
}

var parlogmaria = []string{
	"ALTER TABLE logs_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
}

/* var parmaxmaria = []string{
	"ALTER TABLE logs_capture_all_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
	"ALTER TABLE report_capture_all_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
	"ALTER TABLE rtcp_capture_all_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
	"ALTER TABLE sip_capture_call_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
	"ALTER TABLE sip_capture_registration_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
	"ALTER TABLE sip_capture_rest_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);",
} */

var parqosmaria = []string{
	"ALTER TABLE report_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
	"ALTER TABLE rtcp_capture_all_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
}

var parsipmaria = []string{
	"ALTER TABLE sip_capture_call_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
	"ALTER TABLE sip_capture_registration_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
	"ALTER TABLE sip_capture_rest_{{date}} ADD PARTITION (PARTITION {{date}}_{{time}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') ));",
}

var tblconfmaria = []string{
	`CREATE TABLE IF NOT EXISTS alias (
		id int(10) NOT NULL AUTO_INCREMENT,
		gid int(5) NOT NULL DEFAULT 0,
		ip varchar(80) NOT NULL DEFAULT '',
		port int(10) NOT NULL DEFAULT '0',
		capture_id varchar(100) NOT NULL DEFAULT '',
		alias varchar(100) NOT NULL DEFAULT '',
		is_stp tinyint(1) NOT NULL DEFAULT 0,
		status tinyint(1) NOT NULL DEFAULT 0,
		created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY id (id),
		UNIQUE KEY host_2 (ip,port,capture_id),
		KEY host (ip)
	  ) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

	"CREATE TABLE IF NOT EXISTS `group` (gid int(10) NOT NULL DEFAULT 0,name varchar(100) NOT NULL DEFAULT '',UNIQUE KEY gid (gid)) ENGINE=InnoDB DEFAULT CHARSET=latin1;",

	`CREATE TABLE IF NOT EXISTS link_share (
		id int(10) NOT NULL AUTO_INCREMENT,
		uid int(10) NOT NULL DEFAULT 0,
		uuid varchar(120) NOT NULL DEFAULT '',
		data text NOT NULL,
		expire datetime NOT NULL DEFAULT '2032-12-31 00:00:00',
		active tinyint(1) NOT NULL DEFAULT '1',
		PRIMARY KEY (id)
	  ) ENGINE=InnoDB  DEFAULT CHARSET=latin1;`,

	`CREATE TABLE IF NOT EXISTS node (
		id int(10) NOT NULL AUTO_INCREMENT,
		host varchar(80) NOT NULL DEFAULT '',
		dbname varchar(100) NOT NULL DEFAULT '',
		dbport varchar(100) NOT NULL DEFAULT '',
		dbusername varchar(100) NOT NULL DEFAULT '',
		dbpassword varchar(100) NOT NULL DEFAULT '',
		dbtables varchar(100) NOT NULL DEFAULT 'sip_capture',
		name varchar(100) NOT NULL DEFAULT '',
		status tinyint(1) NOT NULL DEFAULT 0,
		PRIMARY KEY (id),
		UNIQUE KEY id (id),
		UNIQUE KEY host_2 (host),
		KEY host (host)
	  ) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

	`CREATE TABLE IF NOT EXISTS setting (
		id int(10) NOT NULL AUTO_INCREMENT,
		uid int(10) NOT NULL DEFAULT '0',
		param_name varchar(120) NOT NULL DEFAULT '',
		param_value text NOT NULL,
		valid_param_from datetime NOT NULL DEFAULT '2012-01-01 00:00:00',
		valid_param_to datetime NOT NULL DEFAULT '2032-12-01 00:00:00',
		param_prio int(2) NOT NULL DEFAULT '10',
		active int(1) NOT NULL DEFAULT '1',
		PRIMARY KEY (id),
		UNIQUE KEY uid_2 (uid,param_name),
		KEY param_name (param_name),
		KEY uid (uid)
	  ) ENGINE=InnoDB  DEFAULT CHARSET=latin1;`,

	`CREATE TABLE IF NOT EXISTS user (
		uid int(10) unsigned NOT NULL AUTO_INCREMENT,
		gid int(10) NOT NULL DEFAULT '10',
		grp varchar(200) NOT NULL DEFAULT '',
		username varchar(50) NOT NULL DEFAULT '',
		password varchar(100) NOT NULL DEFAULT '',
		firstname varchar(250) NOT NULL DEFAULT '',
		lastname varchar(250) NOT NULL DEFAULT '',
		email varchar(250) NOT NULL DEFAULT '',
		department varchar(100) NOT NULL DEFAULT '',
		regdate timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		lastvisit datetime NOT NULL,
		active tinyint(1) NOT NULL DEFAULT '1',
		PRIMARY KEY (uid),
		UNIQUE KEY login (username),
		UNIQUE KEY username (username)
	  ) ENGINE=InnoDB  DEFAULT CHARSET=latin1;`,

	`CREATE TABLE IF NOT EXISTS user_menu (
		id varchar(125) NOT NULL DEFAULT '',
		name varchar(100) NOT NULL DEFAULT '',
		alias varchar(200) NOT NULL DEFAULT '',
		icon varchar(100) NOT NULL DEFAULT '',
		weight int(10) NOT NULL DEFAULT '10',
		active int(1) NOT NULL DEFAULT '1',
		UNIQUE KEY id (id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;`,

	`CREATE TABLE IF NOT EXISTS api_auth_key (
		id int(10) NOT NULL AUTO_INCREMENT,
		authkey varchar(200) NOT NULL,
		source_ip varchar(200) NOT NULL DEFAULT '0.0.0.0',
		startdate datetime NOT NULL DEFAULT '2012-01-01 00:00:00',
		stopdate datetime NOT NULL DEFAULT '2032-01-01 00:00:00',
		userobject varchar(250) NOT NULL,
		description varchar(200) NOT NULL DEFAULT '',
		lastvisit datetime NOT NULL DEFAULT '2012-01-01 00:00:00',
		enable int(1) NOT NULL DEFAULT '1',
		PRIMARY KEY (id),
		UNIQUE KEY authkey (authkey)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1 AUTO_INCREMENT=1;`,
}

var tbldatalogmaria = []string{
	`CREATE TABLE logs_capture_all_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		type int(5) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg text NOT NULL,
		PRIMARY KEY (id,date),
		KEY date (date),
		KEY correlationid (correlation_id(255))
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,
}

var tbldataqosmaria = []string{
	`CREATE TABLE report_capture_all_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		type int(5) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg varchar(3000) NOT NULL DEFAULT '',
		PRIMARY KEY (id,date),
		KEY date (date),
		KEY correlationid (correlation_id(255))
	  ) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,

	`CREATE TABLE rtcp_capture_all_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		type int(5) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg varchar(1500) NOT NULL DEFAULT '',
		PRIMARY KEY (id,date),
		KEY date (date),
		KEY correlationid (correlation_id(255))
	  ) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,
}

var tbldatasipmaria = []string{
	`CREATE TABLE sip_capture_call_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		method varchar(50) NOT NULL DEFAULT '',
		reply_reason varchar(100) NOT NULL DEFAULT '',
		ruri varchar(200) NOT NULL DEFAULT '',
		ruri_user varchar(100) NOT NULL DEFAULT '',
		ruri_domain varchar(150) NOT NULL DEFAULT '',
		from_user varchar(100) NOT NULL DEFAULT '',
		from_domain varchar(150) NOT NULL DEFAULT '',
		from_tag varchar(64) NOT NULL DEFAULT '',
		to_user varchar(100) NOT NULL DEFAULT '',
		to_domain varchar(150) NOT NULL DEFAULT '',
		to_tag varchar(64) NOT NULL DEFAULT '',
		pid_user varchar(100) NOT NULL DEFAULT '',
		contact_user varchar(120) NOT NULL DEFAULT '',
		auth_user varchar(120) NOT NULL DEFAULT '',
		callid varchar(120) NOT NULL DEFAULT '',
		callid_aleg varchar(120) NOT NULL DEFAULT '',
		via_1 varchar(256) NOT NULL DEFAULT '',
		via_1_branch varchar(80) NOT NULL DEFAULT '',
		cseq varchar(25) NOT NULL DEFAULT '',
		diversion varchar(256) NOT NULL DEFAULT '',
		reason varchar(200) NOT NULL DEFAULT '',
		content_type varchar(256) NOT NULL DEFAULT '',
		auth varchar(256) NOT NULL DEFAULT '',
		user_agent varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		contact_ip varchar(60) NOT NULL DEFAULT '',
		contact_port int(10) NOT NULL DEFAULT 0,
		originator_ip varchar(60) NOT NULL DEFAULT '',
		originator_port int(10) NOT NULL DEFAULT 0,
		expires int(5) NOT NULL DEFAULT '-1',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		custom_field1 varchar(120) NOT NULL DEFAULT '',
		custom_field2 varchar(120) NOT NULL DEFAULT '',
		custom_field3 varchar(120) NOT NULL DEFAULT '',
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		rtp_stat varchar(256) NOT NULL DEFAULT '',
		type int(2) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg varchar(3000) NOT NULL DEFAULT '',
		PRIMARY KEY (id,date),
		KEY ruri_domain (ruri_domain),
		KEY ruri_user (ruri_user),
		KEY from_domain (from_domain),
		KEY from_user (from_user),
		KEY to_domain (to_domain),
		KEY to_user (to_user),
		KEY pid_user (pid_user),
		KEY auth_user (auth_user),
		KEY callid_aleg (callid_aleg),
		KEY date (date),
		KEY callid (callid),
		KEY method (method),
		KEY source_ip (source_ip),
		KEY destination_ip (destination_ip),
		KEY user_agent (user_agent)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,

	`CREATE TABLE sip_capture_registration_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		method varchar(50) NOT NULL DEFAULT '',
		reply_reason varchar(100) NOT NULL DEFAULT '',
		ruri varchar(200) NOT NULL DEFAULT '',
		ruri_user varchar(100) NOT NULL DEFAULT '',
		ruri_domain varchar(150) NOT NULL DEFAULT '',
		from_user varchar(100) NOT NULL DEFAULT '',
		from_domain varchar(150) NOT NULL DEFAULT '',
		from_tag varchar(64) NOT NULL DEFAULT '',
		to_user varchar(100) NOT NULL DEFAULT '',
		to_domain varchar(150) NOT NULL DEFAULT '',
		to_tag varchar(64) NOT NULL DEFAULT '',
		pid_user varchar(100) NOT NULL DEFAULT '',
		contact_user varchar(120) NOT NULL DEFAULT '',
		auth_user varchar(120) NOT NULL DEFAULT '',
		callid varchar(120) NOT NULL DEFAULT '',
		callid_aleg varchar(120) NOT NULL DEFAULT '',
		via_1 varchar(256) NOT NULL DEFAULT '',
		via_1_branch varchar(80) NOT NULL DEFAULT '',
		cseq varchar(25) NOT NULL DEFAULT '',
		diversion varchar(256) NOT NULL DEFAULT '',
		reason varchar(200) NOT NULL DEFAULT '',
		content_type varchar(256) NOT NULL DEFAULT '',
		auth varchar(256) NOT NULL DEFAULT '',
		user_agent varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		contact_ip varchar(60) NOT NULL DEFAULT '',
		contact_port int(10) NOT NULL DEFAULT 0,
		originator_ip varchar(60) NOT NULL DEFAULT '',
		originator_port int(10) NOT NULL DEFAULT 0,
		expires int(5) NOT NULL DEFAULT '-1',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		custom_field1 varchar(120) NOT NULL DEFAULT '',
		custom_field2 varchar(120) NOT NULL DEFAULT '',
		custom_field3 varchar(120) NOT NULL DEFAULT '',
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		rtp_stat varchar(256) NOT NULL DEFAULT '',
		type int(2) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg varchar(3000) NOT NULL DEFAULT '',
		PRIMARY KEY (id,date),
		KEY ruri_domain (ruri_domain),
		KEY from_user (from_user),
		KEY to_user (to_user),
		KEY auth_user (auth_user),
		KEY date (date),
		KEY callid (callid),
		KEY method (method),
		KEY source_ip (source_ip),
		KEY destination_ip (destination_ip),
		KEY user_agent (user_agent)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,

	`CREATE TABLE sip_capture_rest_{{date}} (
		id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		micro_ts bigint(18) NOT NULL DEFAULT '0',
		method varchar(50) NOT NULL DEFAULT '',
		reply_reason varchar(100) NOT NULL DEFAULT '',
		ruri varchar(200) NOT NULL DEFAULT '',
		ruri_user varchar(100) NOT NULL DEFAULT '',
		ruri_domain varchar(150) NOT NULL DEFAULT '',
		from_user varchar(100) NOT NULL DEFAULT '',
		from_domain varchar(150) NOT NULL DEFAULT '',
		from_tag varchar(64) NOT NULL DEFAULT '',
		to_user varchar(100) NOT NULL DEFAULT '',
		to_domain varchar(150) NOT NULL DEFAULT '',
		to_tag varchar(64) NOT NULL DEFAULT '',
		pid_user varchar(100) NOT NULL DEFAULT '',
		contact_user varchar(120) NOT NULL DEFAULT '',
		auth_user varchar(120) NOT NULL DEFAULT '',
		callid varchar(120) NOT NULL DEFAULT '',
		callid_aleg varchar(120) NOT NULL DEFAULT '',
		via_1 varchar(256) NOT NULL DEFAULT '',
		via_1_branch varchar(80) NOT NULL DEFAULT '',
		cseq varchar(25) NOT NULL DEFAULT '',
		diversion varchar(256) NOT NULL DEFAULT '',
		reason varchar(200) NOT NULL DEFAULT '',
		content_type varchar(256) NOT NULL DEFAULT '',
		auth varchar(256) NOT NULL DEFAULT '',
		user_agent varchar(256) NOT NULL DEFAULT '',
		source_ip varchar(60) NOT NULL DEFAULT '',
		source_port int(10) NOT NULL DEFAULT 0,
		destination_ip varchar(60) NOT NULL DEFAULT '',
		destination_port int(10) NOT NULL DEFAULT 0,
		contact_ip varchar(60) NOT NULL DEFAULT '',
		contact_port int(10) NOT NULL DEFAULT 0,
		originator_ip varchar(60) NOT NULL DEFAULT '',
		originator_port int(10) NOT NULL DEFAULT 0,
		expires int(5) NOT NULL DEFAULT '-1',
		correlation_id varchar(256) NOT NULL DEFAULT '',
		custom_field1 varchar(120) NOT NULL DEFAULT '',
		custom_field2 varchar(120) NOT NULL DEFAULT '',
		custom_field3 varchar(120) NOT NULL DEFAULT '',
		proto int(5) NOT NULL DEFAULT 0,
		family int(1) DEFAULT NULL,
		rtp_stat varchar(256) NOT NULL DEFAULT '',
		type int(2) NOT NULL DEFAULT 0,
		node varchar(125) NOT NULL DEFAULT '',
		msg varchar(3000) NOT NULL DEFAULT '',
		PRIMARY KEY (id,date),
		KEY ruri_user (ruri_user),
		KEY from_user (from_user),
		KEY to_user (to_user),
		KEY date (date),
		KEY callid (callid),
		KEY method (method),
		KEY source_ip (source_ip),
		KEY destination_ip (destination_ip),
		KEY user_agent (user_agent)
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8
	  PARTITION BY RANGE ( UNIX_TIMESTAMP(date) ) (
		  PARTITION {{date}}_{{minTime}} VALUES LESS THAN ( UNIX_TIMESTAMP('{{endTime}}') )
	  );`,
}
