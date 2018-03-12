#!/bin/bash
set -e
mkdir -pv "/var/lib/postgresql/data/homer"
for tablespace in homer1 homer2 homer3; do
	mkdir -pv "/var/lib/postgresql/data/$tablespace"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "CREATE TABLESPACE dbspace LOCATION '/data/$tablespace'"
done
