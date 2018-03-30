-- name: create-alias
CREATE TABLE IF NOT EXISTS alias (
  id SERIAL NOT NULL,
  gid INTEGER NOT NULL DEFAULT 0,
  ip varchar(80) NOT NULL DEFAULT '',
  port INTEGER NOT NULL DEFAULT '0', 
  capture_id varchar(100) NOT NULL DEFAULT '',
  alias varchar(100) NOT NULL DEFAULT '',
  is_stp smallint NOT NULL DEFAULT 0,  
  status smallint NOT NULL DEFAULT 0,  
  created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

-- name: create-group
CREATE TABLE IF NOT EXISTS "group" (
  gid INTEGER NOT NULL DEFAULT 0,
  name varchar(100) NOT NULL DEFAULT ''
);

-- name: create-link_share
CREATE TABLE IF NOT EXISTS link_share (
  id SERIAL NOT NULL,
  uid INTEGER NOT NULL DEFAULT 0,
  uuid varchar(120) NOT NULL DEFAULT '',
  data text NOT NULL,
  expire timestamp NOT NULL DEFAULT '2032-12-31 00:00:00',
  active smallint NOT NULL DEFAULT '1',
  PRIMARY KEY (id)
);

-- name: create-node
CREATE TABLE IF NOT EXISTS node (
  id SERIAL NOT NULL,
  host varchar(80) NOT NULL DEFAULT '',
  dbname varchar(100) NOT NULL DEFAULT '',
  dbport varchar(100) NOT NULL DEFAULT '',
  dbusername varchar(100) NOT NULL DEFAULT '',
  dbpassword varchar(100) NOT NULL DEFAULT '',
  dbtables varchar(100) NOT NULL DEFAULT 'sip_capture',
  name varchar(100) NOT NULL DEFAULT '',
  status smallint NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
);

-- name: create-setting
CREATE TABLE IF NOT EXISTS setting (
  id SERIAL NOT NULL,
  uid INTEGER NOT NULL DEFAULT '0',
  param_name varchar(120) NOT NULL DEFAULT '',
  param_value text NOT NULL,
  valid_param_from timestamp NOT NULL DEFAULT '2012-01-01 00:00:00',
  valid_param_to timestamp NOT NULL DEFAULT '2032-12-01 00:00:00',
  param_prio integer NOT NULL DEFAULT '10',
  active INTEGER NOT NULL DEFAULT '1',
  PRIMARY KEY (id)
);

-- name: create-user
CREATE TABLE IF NOT EXISTS "user" (
  uid SERIAL NOT NULL,
  gid INTEGER NOT NULL DEFAULT '10',
  grp varchar(200) NOT NULL DEFAULT '',
  username varchar(50) NOT NULL DEFAULT '',
  password varchar(100) NOT NULL DEFAULT '',
  firstname varchar(250) NOT NULL DEFAULT '',
  lastname varchar(250) NOT NULL DEFAULT '',
  email varchar(250) NOT NULL DEFAULT '',
  department varchar(100) NOT NULL DEFAULT '',
  regdate timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  lastvisit timestamp NOT NULL,
  active smallint NOT NULL DEFAULT '1',
  PRIMARY KEY (uid)
);

-- name: create-user_menu
CREATE TABLE IF NOT EXISTS user_menu (
  id varchar(125) NOT NULL DEFAULT '',
  name varchar(100) NOT NULL DEFAULT '',
  alias varchar(200) NOT NULL DEFAULT '',
  icon varchar(100) NOT NULL DEFAULT '',
  weight INTEGER NOT NULL DEFAULT '10',
  active INTEGER NOT NULL DEFAULT '1'
);

-- name: create-api_auth_key
CREATE TABLE IF NOT EXISTS api_auth_key (
  id SERIAL NOT NULL,
  authkey varchar(200) NOT NULL DEFAULT '',
  source_ip varchar(200) NOT NULL DEFAULT '0.0.0.0',
  startdate timestamp NOT NULL DEFAULT '2012-01-01 00:00:00',
  stopdate timestamp NOT NULL DEFAULT '2031-01-01 00:00:00',
  userobject varchar(250) NOT NULL DEFAULT '',
  description varchar(200) NOT NULL DEFAULT '',
  lastvisit timestamp NOT NULL,
  enable smallint NOT NULL DEFAULT '1',
  PRIMARY KEY (id)
);
