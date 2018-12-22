-- name: create-user_settings
CREATE TABLE IF NOT EXISTS "public"."user_settings" (
    "id" integer DEFAULT nextval('user_settings_id_seq') NOT NULL,
    "guid" uuid,
    "username" character varying(100) NOT NULL,
    "partid" integer NOT NULL,
    "category" character varying(100) DEFAULT 'settings' NOT NULL,
    "create_date" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "param" character varying(100) DEFAULT 'default' NOT NULL,
    "data" json,
    CONSTRAINT "user_settings_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

-- name: create-users
CREATE TABLE IF NOT EXISTS "public"."users" (
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
) WITH (oids = false);

-- name: create-mapping_schema
CREATE TABLE "public"."mapping_schema" (
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
) WITH (oids = false);

