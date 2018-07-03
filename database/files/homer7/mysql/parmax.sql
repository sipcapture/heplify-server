-- name: create-partitionmax-hep_proto_100_default
ALTER TABLE hep_proto_100_default_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_35_default
ALTER TABLE hep_proto_35_default_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_5_default
ALTER TABLE hep_proto_5_default_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_call
ALTER TABLE hep_proto_1_call_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_register
ALTER TABLE hep_proto_1_register_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);

-- name: create-partitionmax-hep_proto_1_default
ALTER TABLE hep_proto_1_default_{{date}} ADD PARTITION (PARTITION pmax VALUES LESS THAN MAXVALUE);