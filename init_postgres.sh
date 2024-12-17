#!/bin/bash
set -e

psql -U postgres -c "ALTER SYSTEM SET log_statement = 'all';"
psql -U postgres -c "ALTER SYSTEM SET log_min_messages = 'NOTICE';"
psql -U postgres -c "SELECT pg_reload_conf();"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

    -- Supabase super admin
    CREATE USER supabase_auth_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD '$POSTGRES_PASSWORD';
    CREATE SCHEMA IF NOT EXISTS $DB_NAMESPACE AUTHORIZATION supabase_auth_admin;
    GRANT CREATE ON DATABASE postgres TO supabase_auth_admin;
    ALTER USER supabase_auth_admin SET search_path = '$DB_NAMESPACE';
EOSQL