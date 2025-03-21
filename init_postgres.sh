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

    -- Create roles first
    CREATE ROLE solomon_role;
    
    -- Grant permissions on schema level
    GRANT USAGE ON SCHEMA $DB_NAMESPACE TO solomon_role;
    
    -- Create users and assign to roles
    CREATE USER zion_user INHERIT LOGIN NOREPLICATION PASSWORD '$ZION_DB_USER_PASSWORD';
    CREATE USER zion_admin INHERIT LOGIN NOREPLICATION PASSWORD '$ZION_DB_ADMIN_PASSWORD';
    CREATE USER brawler_user INHERIT LOGIN NOREPLICATION PASSWORD '$BRAWLER_DB_USER_PASSWORD';
    CREATE USER brawler_admin INHERIT LOGIN NOREPLICATION PASSWORD '$BRAWLER_DB_ADMIN_PASSWORD';
    
    -- Assign users to roles
    GRANT solomon_role TO zion_user, brawler_user, zion_admin, brawler_admin;
EOSQL