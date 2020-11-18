package rotator

var (
	selectlogpg      = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_100_default_%' and tablename < 'hep_proto_100_default_{{date}}_{{time}}';"
	selectreportpg   = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_35_default_%' and tablename < 'hep_proto_35_default_{{date}}_{{time}}';"
	selectisuppg     = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_54_default_%' and tablename < 'hep_proto_54_default_{{date}}_{{time}}';"
	selectrtcppg     = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_5_default_%' and tablename < 'hep_proto_5_default_{{date}}_{{time}}';"
	selectcallpg     = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_call_%' and tablename < 'hep_proto_1_call_{{date}}_{{time}}';"
	selectregisterpg = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_registration_%' and tablename < 'hep_proto_1_registration_{{date}}_{{time}}';"
	selectdefaultpg  = "SELECT tablename FROM pg_tables WHERE tablename LIKE 'hep_proto_1_default_%' and tablename < 'hep_proto_1_default_{{date}}_{{time}}';"
)

var (
	sysDF = "CREATE OR REPLACE FUNCTION sys_df() \nRETURNS SETOF text[]\nLANGUAGE plpgsql \nas\n$$\nBEGIN\n    CREATE TEMP TABLE IF NOT EXISTS tmp_sys_df (content text) ON COMMIT DROP;\n    COPY tmp_sys_df FROM PROGRAM 'df $PGDATA | tail -n +2';\n    RETURN QUERY SELECT regexp_split_to_array(content, '\\s+') FROM tmp_sys_df;\nEND;\n$$;"
)
var (
	droplogpg      = "DROP TABLE IF EXISTS {{partName}};"
	dropreportpg   = "DROP TABLE IF EXISTS {{partName}};"
	dropisuppg     = "DROP TABLE IF EXISTS {{partName}};"
	droprtcppg     = "DROP TABLE IF EXISTS {{partName}};"
	dropcallpg     = "DROP TABLE IF EXISTS {{partName}};"
	dropregisterpg = "DROP TABLE IF EXISTS {{partName}};"
	dropdefaultpg  = "DROP TABLE IF EXISTS {{partName}};"
)

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
