-- name: insert-alias-table
INSERT INTO alias (id, gid, ip, port, capture_id, alias, status, created) VALUES
(1, 10, '192.168.0.30', 0, 'homer01', 'proxy01', 1, '2014-06-12 20:36:50'),
(2, 10, '192.168.0.4', 0, 'homer01', 'acme-234', 1, '2014-06-12 20:37:01'),
(22, 10, '127.0.0.1:5060', 0, 'homer01', 'sip.local.net', 1, '2014-06-12 20:37:01');

-- name: insert-group-table
INSERT INTO "group" (gid, name) VALUES (10, 'Administrator');

-- name: insert-node-table
INSERT INTO node (id, host, dbname, dbport, dbusername, dbpassword, dbtables, name, status) VALUES
(1, '127.0.0.1', 'homer_data', '3306', 'homer_user', 'mysql_password', 'sip_capture', 'homer01', 1),
(21, '10.1.0.7', 'homer_data', '3306', 'homer_user', 'mysql_password', 'sip_capture', 'external', 1);

-- name: insert-setting-table
INSERT INTO setting (id, uid, param_name, param_value, valid_param_from, valid_param_to, param_prio, active) VALUES
(1, 1, 'timerange', '{"from":"2015-05-26T18:34:42.654Z","to":"2015-05-26T18:44:42.654Z"}', '2012-01-01 00:00:00', '2032-12-01 00:00:00', 10, 1);

-- name: insert-user-table
INSERT INTO "user" (uid, gid, grp, username, password, firstname, lastname, email, department, regdate, lastvisit, active) VALUES
(1, 10, 'users,admins', 'admin', crypt('test123', gen_salt('md5')), 'Admin', 'Admin', 'admin@test.com', 'Voice Enginering', '2012-01-19 00:00:00', '2015-05-29 07:17:35', 1),
(2, 10, 'users', 'noc', crypt('123test', gen_salt('md5')), 'NOC', 'NOC', 'noc@test.com', 'Voice NOC', '2012-01-19 00:00:00', '2015-05-29 07:17:35', 1);

-- name: insert-user_menu-table
INSERT INTO user_menu (id, name, alias, icon, weight, active) VALUES
('_1426001444630', 'SIP Search', 'search', 'fa-search', 10, 1),
('_1427728371642', 'Home', 'home', 'fa-home', 1, 1),
('_1431721484444', 'Alarms', 'alarms', 'fa-warning', 20, 1);