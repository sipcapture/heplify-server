package rotator

var (
	listdroplogpg      = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_100_default_%' and tablename < 'hep_proto_100_default_{{date}}_{{time}}';"}
	listdropreportpg   = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_35_default_%' and tablename < 'hep_proto_35_default_{{date}}_{{time}}';"}
	listdropisuppg     = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_54_default_%' and tablename < 'hep_proto_54_default_{{date}}_{{time}}';"}
	listdroprtcppg     = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_5_default_%' and tablename < 'hep_proto_5_default_{{date}}_{{time}}';"}
	listdropcallpg     = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_call_%' and tablename < 'hep_proto_1_call_{{date}}_{{time}}';"}
	listdropregisterpg = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_registration_%' and tablename < 'hep_proto_1_registration_{{date}}_{{time}}';"}
	listdropdefaultpg  = []string{"SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_default_%' and tablename < 'hep_proto_1_default_{{date}}_{{time}}';"}
)

var (
	droplogpg      = []string{"DROP TABLE IF EXISTS {{partName}};"}
	dropreportpg   = []string{"DROP TABLE IF EXISTS {{partName}};"}
	dropisuppg     = []string{"DROP TABLE IF EXISTS {{partName}};"}
	droprtcppg     = []string{"DROP TABLE IF EXISTS {{partName}};"}
	dropcallpg     = []string{"DROP TABLE IF EXISTS {{partName}};"}
	dropregisterpg = []string{"DROP TABLE IF EXISTS {{partName}};"}
	dropdefaultpg  = []string{"DROP TABLE IF EXISTS {{partName}};"}
)

var idxconfpg = []string{
	"CREATE SEQUENCE IF NOT EXISTS user_settings_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 2 CACHE 1;",
	"CREATE SEQUENCE IF NOT EXISTS users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 2 CACHE 1;",
	"CREATE SEQUENCE IF NOT EXISTS mapping_schema_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 6 CACHE 1;",
	"CREATE SEQUENCE IF NOT EXISTS alias_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;",
	"CREATE SEQUENCE IF NOT EXISTS global_settings_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;",
}

var idxlogpg = []string{
	"CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_create_date ON hep_proto_100_default_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_sid ON hep_proto_100_default_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_srcIp ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_dstIp ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}}_correlation_id ON hep_proto_100_default_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",
}

var idxisuppg = []string{
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_create_date ON hep_proto_54_default_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_sid ON hep_proto_54_default_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_correlation_id ON hep_proto_54_default_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_called_number ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'called_number'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_calling_number ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'calling_number'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_opc ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'opc'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_dpc ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'dpc'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_cic ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'cic'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_msg_name ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'msg_name'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}}_callid ON hep_proto_54_default_{{date}}_{{time}} ((data_header->>'callid'));",
}

var idxqospg = []string{
	"CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_create_date ON hep_proto_35_default_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_sid ON hep_proto_35_default_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_srcIp ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_dstIp ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}}_correlation_id ON hep_proto_35_default_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_create_date ON hep_proto_5_default_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_sid ON hep_proto_5_default_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_srcIp ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_dstIp ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}}_correlation_id ON hep_proto_5_default_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",
}

var idxsippg = []string{
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_create_date ON hep_proto_1_call_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_sid ON hep_proto_1_call_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_srcIp ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_dstIp ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_correlation_id ON hep_proto_1_call_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_ruri_domain ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'ruri_domain'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_ruri_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'ruri_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_from_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'from_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_to_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'to_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_pid_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'pid_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_auth_user ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'auth_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_callid ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'callid'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}}_method ON hep_proto_1_call_{{date}}_{{time}} ((data_header->>'method'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_create_date ON hep_proto_1_registration_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_sid ON hep_proto_1_registration_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_srcIp ON hep_proto_1_registration_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_dstIp ON hep_proto_1_registration_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_correlation_id ON hep_proto_1_registration_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_ruri_domain ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'ruri_domain'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_ruri_user ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'ruri_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_from_user ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'from_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_to_user ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'to_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_pid_user ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'pid_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_auth_user ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'auth_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_callid ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'callid'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}}_method ON hep_proto_1_registration_{{date}}_{{time}} ((data_header->>'method'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_create_date ON hep_proto_1_default_{{date}}_{{time}} (create_date);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_sid ON hep_proto_1_default_{{date}}_{{time}} (sid);",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_srcIp ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->>'srcIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_dstIp ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->>'dstIp'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_correlation_id ON hep_proto_1_default_{{date}}_{{time}} ((protocol_header->>'correlation_id'));",

	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_ruri_domain ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'ruri_domain'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_ruri_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'ruri_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_from_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'from_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_to_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'to_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_pid_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'pid_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_auth_user ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'auth_user'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_callid ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'callid'));",
	"CREATE INDEX IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}}_method ON hep_proto_1_default_{{date}}_{{time}} ((data_header->>'method'));",
}

var insconfpg = []string{
	`INSERT INTO "alias" ("id", "guid", "alias", "ip", "port", "mask", "captureID", "status", "create_date") VALUES
	(1,	'84b2d4d5-ff68-4c0e-b5d2-08ab3f4da8e1',	'localhost',	'127.0.0.1',	5060,	32,	'0',	'1',	'2019-03-09 11:20:25.115838+00');`,

	`INSERT INTO "global_settings" ("id", "guid", "partid", "category", "create_date", "param", "data") VALUES
	(1,	'eb45e96d-0d8e-40b3-b10b-2990d5abb2c0',	1,	'search',	'2019-03-09 11:06:40.345+00',	'lokiserver',	'{"host":"http://127.0.0.1:3100"}'),
	(2,	'af47f362-71ff-42b4-a81a-dd974962212e',	1,	'search',	'2019-03-09 11:06:40.345+00',	'promserver',	'{"host":"http://127.0.0.1:9090/api/v1/"}');`,

	`INSERT INTO "user_settings" ("id", "guid", "username", "partid", "category", "create_date", "param", "data") VALUES
	(1,	'0484a281-55d8-4fa9-8fbd-338bc23ddb41',	'admin',	10,	'dashboard',	'2018-12-22 18:29:03.375+00',	'home',	'{"id":"home","name":"Home","alias":"home","selectedItem":"","title":"Home","weight":10.0,"widgets":[{"reload":false,"frameless":false,"title":"World Clock","group":"Tools","name":"clock","description":"Display date and time","templateUrl":"widgets/clock/view.html","controller":"clockController","controllerAs":"clock","sizeX":1,"sizeY":1,"config":{"title":"World Clock","timePattern":"HH:mm:ss","datePattern":"YYYY-MM-DD","location":{"value":-60,"offset":"+1","name":"GMT+1 CET","desc":"Central European Time"},"showseconds":false},"edit":{"reload":true,"immediate":false,"controller":"clockEditController","templateUrl":"widgets/clock/edit.html"},"row":0,"col":0,"api":{},"uuid":"0131d42a-793d-47d6-ad03-7cdc6811fb56"},{"title":"Proto Search","group":"Search","name":"protosearch","description":"Display Search Form component","refresh":false,"sizeX":2,"sizeY":1,"config":{"title":"CALL SIP SEARCH","searchbutton":true,"protocol_id":{"name":"SIP","value":1},"protocol_profile":{"name":"call","value":"call"}},"uuid":"ed426bd0-ff21-40f7-8852-58700abc3762","fields":[{"name":"1:call:sid","selection":"Session ID","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"sid"},{"name":"1:call:protocol_header.srcIp","selection":"Source IP","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"protocol_header.srcIp"},{"name":"1:call:protocol_header.srcPort","selection":"Src Port","form_type":"input","hepid":1,"profile":"call","type":"integer","field_name":"protocol_header.srcPort"},{"name":"1:call:raw","selection":"SIP RAW","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"raw"}],"row":0,"col":1},{"title":"InfluxDB Chart","group":"Charts","name":"influxdbchart","description":"Display SIP Metrics","refresh":true,"sizeX":2,"sizeY":1,"config":{"title":"HEPIC Chart","chart":{"type":{"value":"line"}},"dataquery":{"data":[{"sum":false,"main":{"name":"heplify_method_response","value":"heplify_method_response"},"database":{"name":"homer"},"retention":{"name":"60s"},"type":[{"name":"counter","value":"counter"}],"tag":{},"typetag":{"name":"response","value":"response"}}]},"panel":{"queries":[{"name":"A1","type":{"name":"InfluxDB","alias":"influxdb"},"database":{"name":"homer"},"retention":{"name":"60s"},"value":"query"}]}},"edit":{},"api":{},"uuid":"8c8b4589-426a-4016-b964-d613ab6997b3","row":0,"col":3}],"config":{"margins":[10.0,10.0],"columns":"6","pushing":true,"draggable":{"handle":".box-header"},"resizable":{"enabled":true,"handles":["n","e","s","w","ne","se","sw","nw"]}}}'),
	(2,	'692287e7-e6d7-44de-8125-11312af4f6f3',	'support',	10,	'dashboard',	'2018-12-22 18:29:03.375+00',	'home',	'{"id":"home","name":"Home","alias":"home","selectedItem":"","title":"Home","weight":10.0,"widgets":[{"reload":false,"frameless":false,"title":"World Clock","group":"Tools","name":"clock","description":"Display date and time","templateUrl":"widgets/clock/view.html","controller":"clockController","controllerAs":"clock","sizeX":1,"sizeY":1,"config":{"title":"World Clock","timePattern":"HH:mm:ss","datePattern":"YYYY-MM-DD","location":{"value":-60,"offset":"+1","name":"GMT+1 CET","desc":"Central European Time"},"showseconds":false},"edit":{"reload":true,"immediate":false,"controller":"clockEditController","templateUrl":"widgets/clock/edit.html"},"row":0,"col":0,"api":{},"uuid":"0131d42a-793d-47d6-ad03-7cdc6811fb56"},{"title":"Proto Search","group":"Search","name":"protosearch","description":"Display Search Form component","refresh":false,"sizeX":2,"sizeY":1,"config":{"title":"CALL SIP SEARCH","searchbutton":true,"protocol_id":{"name":"SIP","value":1},"protocol_profile":{"name":"call","value":"call"}},"uuid":"ed426bd0-ff21-40f7-8852-58700abc3762","fields":[{"name":"1:call:sid","selection":"Session ID","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"sid"},{"name":"1:call:protocol_header.srcIp","selection":"Source IP","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"protocol_header.srcIp"},{"name":"1:call:protocol_header.srcPort","selection":"Src Port","form_type":"input","hepid":1,"profile":"call","type":"integer","field_name":"protocol_header.srcPort"},{"name":"1:call:raw","selection":"SIP RAW","form_type":"input","hepid":1,"profile":"call","type":"string","field_name":"raw"}],"row":0,"col":1},{"title":"InfluxDB Chart","group":"Charts","name":"influxdbchart","description":"Display SIP Metrics","refresh":true,"sizeX":2,"sizeY":1,"config":{"title":"HEPIC Chart","chart":{"type":{"value":"line"}},"dataquery":{"data":[{"sum":false,"main":{"name":"heplify_method_response","value":"heplify_method_response"},"database":{"name":"homer"},"retention":{"name":"60s"},"type":[{"name":"counter","value":"counter"}],"tag":{},"typetag":{"name":"response","value":"response"}}]},"panel":{"queries":[{"name":"A1","type":{"name":"InfluxDB","alias":"influxdb"},"database":{"name":"homer"},"retention":{"name":"60s"},"value":"query"}]}},"edit":{},"api":{},"uuid":"8c8b4589-426a-4016-b964-d613ab6997b3","row":0,"col":3}],"config":{"margins":[10.0,10.0],"columns":"6","pushing":true,"draggable":{"handle":".box-header"},"resizable":{"enabled":true,"handles":["n","e","s","w","ne","se","sw","nw"]}}}');`,

	`INSERT INTO "users" ("id", "username", "partid", "email", "firstname", "lastname", "department", "usergroup", "hash", "guid", "created_at") VALUES
	(1,	'admin',	10,	'root@localhost',	'Homer',	'Admin',	'NOC',	'admin',	'$2a$10$hFURHY210kbyE/fPEXsSnOuSs6FTDbVMiss07PacXI/43G.YX.xDi',	'11111111-1111-1111-1111-111111111111',	'2018-12-22 18:29:03.352983+00'),
	(2,	'support',	10,	'root@localhost',	'Homer',	'Support',	'NOC',	'admin',	'$2a$10$hFURHY210kbyE/fPEXsSnOuSs6FTDbVMiss07PacXI/43G.YX.xDi',	'22222222-2222-2222-2222-222222222222',	'2018-12-22 18:29:03.352983+00');`,

	`INSERT INTO "mapping_schema" ("id", "guid", "profile", "hepid", "hep_alias", "partid", "version", "retention", "partition_step", "create_index", "create_table", "correlation_mapping", "fields_mapping", "mapping_settings", "schema_mapping", "schema_settings", "create_date") VALUES
	(1,	'28a666a2-859e-469f-90ac-89ee4be48bfc',	'default',	1,	'SIP',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"data_header.callid","lookup_id":100,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]},{"source_field":"data_header.callid","lookup_id":5,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.protocolFamily","name":"Proto Family","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.protocol","name":"Protocol Type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.srcIp","name":"Source IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.dstIp","name":"Destination IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.srcPort","name":"Src Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.dstPort","name":"Dst Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeSeconds","name":"Timeseconds","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeUseconds","name":"Usecond time","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.payloadType","name":"Payload type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.captureId","name":"Capture ID","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.capturePass","name":"Capture Pass","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.correlation_id","name":"Correlation ID","type":"string","index":"none","form_type":"input"},{"id":"data_header.method","name":"SIP Method","type":"string","index":"none","form_type":"input","form_default":["INVITE","BYE","100","200","183","CANCEL"]},{"id":"data_header.callid","name":"SIP Callid","type":"string","index":"none","form_type":"input"},{"id":"data_header.cseq","name":"SIP Cseq","type":"integer","index":"none","form_type":"input"},{"id":"data_header.to_user","name":"SIP To user","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_tag","name":"SIP From tag","type":"string","index":"none","form_type":"input"},{"id":"data_header.protocol","name":"SIP Protocol","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_user","name":"SIP From user","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"SIP RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(2,	'6560d012-6cca-48e6-bdf4-b9dc93c55f80',	'call',	1,	'SIP',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"data_header.callid","lookup_id":100,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]},{"source_field":"data_header.callid","lookup_id":5,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.protocolFamily","name":"Proto Family","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.protocol","name":"Protocol Type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.srcIp","name":"Source IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.dstIp","name":"Destination IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.srcPort","name":"Src Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.dstPort","name":"Dst Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeSeconds","name":"Timeseconds","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeUseconds","name":"Usecond time","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.payloadType","name":"Payload type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.captureId","name":"Capture ID","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.capturePass","name":"Capture Pass","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.correlation_id","name":"Correlation ID","type":"string","index":"none","form_type":"input"},{"id":"data_header.method","name":"SIP Method","type":"string","index":"none","form_type":"input","form_default":["INVITE","BYE","100","200","183","CANCEL"]},{"id":"data_header.callid","name":"SIP Callid","type":"string","index":"none","form_type":"input"},{"id":"data_header.cseq","name":"SIP Cseq","type":"integer","index":"none","form_type":"input"},{"id":"data_header.to_user","name":"SIP To user","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_tag","name":"SIP From tag","type":"string","index":"none","form_type":"input"},{"id":"data_header.protocol","name":"SIP Protocol","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_user","name":"SIP From user","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"SIP RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(3,	'9632545b-d71c-4808-9097-7337471556cf',	'registration',	1,	'SIP',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"data_header.callid","lookup_id":100,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]},{"source_field":"data_header.callid","lookup_id":5,"lookup_profile":"default","lookup_field":"sid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.protocolFamily","name":"Proto Family","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.protocol","name":"Protocol Type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.srcIp","name":"Source IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.dstIp","name":"Destination IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.srcPort","name":"Src Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.dstPort","name":"Dst Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeSeconds","name":"Timeseconds","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeUseconds","name":"Usecond time","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.payloadType","name":"Payload type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.captureId","name":"Capture ID","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.capturePass","name":"Capture Pass","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.correlation_id","name":"Correlation ID","type":"string","index":"none","form_type":"input"},{"id":"data_header.method","name":"SIP Method","type":"string","index":"none","form_type":"input","form_default":["INVITE","BYE","100","200","183","CANCEL"]},{"id":"data_header.callid","name":"SIP Callid","type":"string","index":"none","form_type":"input"},{"id":"data_header.cseq","name":"SIP Cseq","type":"integer","index":"none","form_type":"input"},{"id":"data_header.to_user","name":"SIP To user","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_tag","name":"SIP From tag","type":"string","index":"none","form_type":"input"},{"id":"data_header.protocol","name":"SIP Protocol","type":"string","index":"none","form_type":"input"},{"id":"data_header.from_user","name":"SIP From user","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"SIP RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(4,	'10c19417-ce04-4c28-984c-801e686e1ce7',	'default',	100,	'LOG',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"sid","lookup_id":1,"lookup_profile":"call","lookup_field":"data_header.callid","lookup_range":[-300,200]},{"source_field":"sid","lookup_id":1,"lookup_profile":"registration","lookup_field":"data_header.callid","lookup_range":[-300,200]},{"source_field":"sid","lookup_id":1,"lookup_profile":"default","lookup_field":"data_header.callid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.protocolFamily","name":"Proto Family","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.protocol","name":"Protocol Type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.srcIp","name":"Source IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.dstIp","name":"Destination IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.srcPort","name":"Src Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.dstPort","name":"Dst Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeSeconds","name":"Timeseconds","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeUseconds","name":"Usecond time","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.payloadType","name":"Payload type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.captureId","name":"Capture ID","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.capturePass","name":"Capture Pass","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.correlation_id","name":"Correlation ID","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(5,	'2696a72d-a267-449b-a61a-3a9406d22685',	'default',	34,	'RTP-FULL-REPORT',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"sid","lookup_id":1,"lookup_profile":"call","lookup_field":"data_header.callid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.protocolFamily","name":"Proto Family","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.protocol","name":"Protocol Type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.srcIp","name":"Source IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.dstIp","name":"Destination IP","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.srcPort","name":"Src Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.dstPort","name":"Dst Port","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeSeconds","name":"Timeseconds","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.timeUseconds","name":"Usecond time","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.payloadType","name":"Payload type","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.captureId","name":"Capture ID","type":"integer","index":"none","form_type":"input"},{"id":"protocol_header.capturePass","name":"Capture Pass","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.correlation_id","name":"Correlation ID","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(6,	'fc9bf543-0747-450a-8b98-f08bdc05f061',	'default',	1000,	'JANUS',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"sid","lookup_id":1,"lookup_profile":"call","lookup_field":"data_header.callid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.address","name":"Proto Address","type":"string","index":"none","form_type":"input"},{"id":"data_header.family","name":"Family","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.port","name":"Protocol port","type":"integer","index":"none","form_type":"input"},{"id":"data_header.type","name":"Data type","type":"integer","index":"none","form_type":"input"},{"id":"data_header.handle","name":"Data Handle","type":"integer","index":"none","form_type":"input"},{"id":"data_header.medium","name":"Data Medium","type":"string","index":"none","form_type":"input"},{"id":"data_header.source","name":"Data Source","type":"string","index":"none","form_type":"input"},{"id":"data_header.session","name":"Data Session","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00'),
	(7,	'465a0b24-f824-42e5-8bba-b00afd7d386c',	'default',	54,	'ISUP-JSON',	10,	1,	10,	10,	'{}',	'CREATE TABLE test(id integer, data text);',	'[{"source_field":"sid","lookup_id":1,"lookup_profile":"call","lookup_field":"data_header.callid","lookup_range":[-300,200]}]',	'[{"id":"sid","type":"string","index":"secondary","name":"Session ID","form_type":"input"},{"id":"protocol_header.address","name":"Proto Address","type":"string","index":"none","form_type":"input"},{"id":"data_header.family","name":"Family","type":"string","index":"none","form_type":"input"},{"id":"protocol_header.port","name":"Protocol port","type":"integer","index":"none","form_type":"input"},{"id":"data_header.type","name":"Data type","type":"integer","index":"none","form_type":"input"},{"id":"data_header.handle","name":"Data Handle","type":"integer","index":"none","form_type":"input"},{"id":"data_header.medium","name":"Data Medium","type":"string","index":"none","form_type":"input"},{"id":"data_header.source","name":"Data Source","type":"string","index":"none","form_type":"input"},{"id":"data_header.session","name":"Data Session","type":"string","index":"none","form_type":"input"},{"id":"raw","name":"RAW","type":"string","index":"none","form_type":"input"}]',	NULL,	'{}',	'{}',	'2018-12-22 19:30:45.996+00');`,
}

var parlogpg = []string{
	"CREATE TABLE IF NOT EXISTS hep_proto_100_default_{{date}}_{{time}} PARTITION OF hep_proto_100_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
}

var parisuppg = []string{
	"CREATE TABLE IF NOT EXISTS hep_proto_54_default_{{date}}_{{time}} PARTITION OF hep_proto_54_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
}

var parqospg = []string{
	"CREATE TABLE IF NOT EXISTS hep_proto_35_default_{{date}}_{{time}} PARTITION OF hep_proto_35_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
	"CREATE TABLE IF NOT EXISTS hep_proto_5_default_{{date}}_{{time}} PARTITION OF hep_proto_5_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
}

var parsippg = []string{
	"CREATE TABLE IF NOT EXISTS hep_proto_1_call_{{date}}_{{time}} PARTITION OF hep_proto_1_call FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
	"CREATE TABLE IF NOT EXISTS hep_proto_1_registration_{{date}}_{{time}} PARTITION OF hep_proto_1_registration FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
	"CREATE TABLE IF NOT EXISTS hep_proto_1_default_{{date}}_{{time}} PARTITION OF hep_proto_1_default FOR VALUES FROM ('{{startTime}}') TO ('{{endTime}}');",
}

var tblconfpg = []string{
	`CREATE TABLE IF NOT EXISTS "public"."user_settings" (
		"id" integer DEFAULT nextval('user_settings_id_seq') NOT NULL,
		"guid" uuid,
		"username" character varying(100) NOT NULL,
		"partid" integer NOT NULL,
		"category" character varying(100) DEFAULT 'settings' NOT NULL,
		"create_date" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
		"param" character varying(100) DEFAULT 'default' NOT NULL,
		"data" json,
		CONSTRAINT "user_settings_pkey" PRIMARY KEY ("id")
	) WITH (oids = false);`,

	`CREATE TABLE IF NOT EXISTS "public"."users" (
		"id" integer DEFAULT nextval('users_id_seq') NOT NULL,
		"username" character varying(50) NOT NULL,
		"partid" integer DEFAULT '10' NOT NULL,
		"email" character varying(250) NOT NULL,
		"firstname" character varying(50) NOT NULL,
		"lastname" character varying(50) NOT NULL,
		"department" character varying(50) DEFAULT 'NOC' NOT NULL,
		"usergroup" character varying(250) NOT NULL,
		"hash" character varying(128) NOT NULL,
		"guid" character varying(50) NOT NULL,
		"created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
		CONSTRAINT "users_guid_unique" UNIQUE ("guid"),
		CONSTRAINT "users_pkey" PRIMARY KEY ("id"),
		CONSTRAINT "users_username_unique" UNIQUE ("username")
	) WITH (oids = false);`,

	`CREATE TABLE "public"."mapping_schema" (
		"id" integer DEFAULT nextval('mapping_schema_id_seq') NOT NULL,
		"guid" uuid,
		"profile" character varying(100) DEFAULT 'default' NOT NULL,
		"hepid" integer NOT NULL,
		"hep_alias" character varying(100),
		"partid" integer DEFAULT '10' NOT NULL,
		"version" integer NOT NULL,
		"retention" integer DEFAULT '14' NOT NULL,
		"partition_step" integer DEFAULT '3600' NOT NULL,
		"create_index" json,
		"create_table" text,
		"correlation_mapping" json,
		"fields_mapping" json,
		"mapping_settings" json,
		"schema_mapping" json,
		"schema_settings" json,
		"create_date" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
		CONSTRAINT "mapping_schema_pkey" PRIMARY KEY ("id")
	) WITH (oids = false);`,

	`CREATE TABLE "public"."alias" (
		"id" integer DEFAULT nextval('alias_id_seq') NOT NULL,
		"guid" uuid,
		"alias" character varying(40),
		"ip" character varying(60),
		"port" integer,
		"mask" integer,
		"captureID" character varying(20),
		"status" boolean,
		"create_date" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
		CONSTRAINT "alias_pkey" PRIMARY KEY ("id")
	) WITH (oids = false);`,

	`CREATE TABLE "public"."global_settings" (
		"id" integer DEFAULT nextval('global_settings_id_seq') NOT NULL,
		"guid" uuid,
		"partid" integer NOT NULL,
		"category" character varying(100) DEFAULT 'settings' NOT NULL,
		"create_date" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
		"param" character varying(100) DEFAULT 'default' NOT NULL,
		"data" json,
		CONSTRAINT "global_settings_pkey" PRIMARY KEY ("id")
	) WITH (oids = false);`,
}

var tbldatapg = []string{
	`CREATE TABLE IF NOT EXISTS hep_proto_100_default (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_35_default (
  		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_5_default (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_1_call (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_1_registration (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_1_default (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,

	`CREATE TABLE IF NOT EXISTS hep_proto_54_default (
		id BIGSERIAL NOT NULL,
		sid varchar NOT NULL,
		create_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
		protocol_header jsonb NOT NULL,
		data_header jsonb NOT NULL,
		raw varchar NOT NULL
	) PARTITION BY RANGE (create_date);`,
}
