-- name: index-alias_id
CREATE UNIQUE INDEX IF NOT EXISTS alias_id ON alias (id);
-- name: index-alias_id_port_captid
CREATE UNIQUE INDEX IF NOT EXISTS alias_id_port_captid ON alias (id,port,capture_id);
-- name: index-alias_ip_idx
CREATE INDEX IF NOT EXISTS alias_ip_idx ON alias (ip);

-- name: index-group_gid
CREATE UNIQUE INDEX IF NOT EXISTS group_gid ON "group" (gid);

-- name: index-node_id
CREATE UNIQUE INDEX IF NOT EXISTS node_id ON node (id);
-- name: index-node_host
CREATE UNIQUE INDEX IF NOT EXISTS node_host ON node (host);
-- name: index-node_host_idx
CREATE INDEX IF NOT EXISTS node_host_idx ON node (host);

-- name: index-setting_id
CREATE UNIQUE INDEX IF NOT EXISTS setting_id ON setting (uid,param_name);
-- name: index-setting_param_name
CREATE INDEX IF NOT EXISTS setting_param_name ON setting (param_name);
-- name: index-setting_uid
CREATE INDEX IF NOT EXISTS setting_uid ON setting (uid);

-- name: index-user_name
CREATE UNIQUE INDEX IF NOT EXISTS user_name ON "user" (username);

-- name: index-user_menu_id
CREATE UNIQUE INDEX IF NOT EXISTS user_menu_id ON "user_menu" (id);

-- name: index-api_auth_key_authkey
CREATE UNIQUE INDEX IF NOT EXISTS api_auth_key_authkey ON "api_auth_key" (authkey);
