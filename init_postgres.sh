#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;
    CREATE USER $AUTH_DB_ADMIN NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD '$AUTH_DB_ADMIN_PASSWORD';
    
    -- Create pgcrypto extension with superuser privileges
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";

    -- Create schema and grant permissions on schema level
    CREATE SCHEMA IF NOT EXISTS $DB_NAMESPACE AUTHORIZATION $AUTH_DB_ADMIN;
    GRANT ALL PRIVILEGES ON SCHEMA public TO $AUTH_DB_ADMIN;

    -- Set search paths for existing users and roles
    ALTER ROLE $AUTH_DB_ADMIN SET search_path TO $DB_NAMESPACE;
    ALTER ROLE supabase_admin SET search_path TO $DB_NAMESPACE;
EOSQL