#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER replicator WITH REPLICATION LOGIN PASSWORD '$REPLICATOR_PASSWORD';
EOSQL

# Allow replication connections from any host, secured by scram-sha-256
echo "host replication replicator 0.0.0.0/0 scram-sha-256" >> "$PGDATA/pg_hba.conf"
