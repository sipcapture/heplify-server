-- name: insert-alias
INSERT INTO alias (id, gid, ip, port, capture_id, alias, status, created) VALUES
(1, 10, '192.168.0.30', 0, 'homer01', 'proxy01', 1, '2014-06-12 20:36:50');

-- name: insert-group
INSERT INTO "group" (gid, name) VALUES (10, 'Administrator');

-- name: insert-node
INSERT INTO node (id, host, dbname, dbport, dbusername, dbpassword, dbtables, name, status) VALUES
(1, '127.0.0.1', 'homer_data', '5432', 'homer_user', 'homer_password', 'sip_capture', 'homer01', 1);

-- name: insert-setting
INSERT INTO setting (id, uid, param_name, param_value, valid_param_from, valid_param_to, param_prio, active) VALUES
(1, 1, 'timerange', '{"from":"2018-02-21T18:34:42.654Z","to":"2018-12-21T18:44:42.654Z"}', '2012-01-01 00:00:00', '2032-12-01 00:00:00', 10, 1);

-- name: insert-user
INSERT INTO "user" (uid, gid, grp, username, password, firstname, lastname, email, department, regdate, lastvisit, active) VALUES
(1, 10, 'users,admins', 'admin', crypt('test123', gen_salt('md5')), 'Admin', 'Admin', 'admin@test.com', 'Voice Enginering', '2012-01-19 00:00:00', '2015-05-29 07:17:35', 1),
(2, 10, 'users', 'noc', crypt('123test', gen_salt('md5')), 'NOC', 'NOC', 'noc@test.com', 'Voice NOC', '2012-01-19 00:00:00', '2015-05-29 07:17:35', 1);

-- name: insert-user_menu
INSERT INTO user_menu (id, name, alias, icon, weight, active) VALUES
('_1426001444630', 'SIP Search', 'search', 'fa-search', 10, 1),
('_1427728371642', 'Home', 'home', 'fa-home', 1, 1);