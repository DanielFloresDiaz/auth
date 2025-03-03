#!/bin/bash
set -e

# psql -U postgres -c "ALTER SYSTEM SET log_statement = 'all';"
# psql -U postgres -c "ALTER SYSTEM SET log_min_messages = 'NOTICE';"
# psql -U postgres -c "SELECT pg_reload_conf();"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

    -- Supabase super admin
    CREATE USER supabase_auth_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD '$POSTGRES_PASSWORD';
    CREATE SCHEMA IF NOT EXISTS $DB_NAMESPACE AUTHORIZATION supabase_auth_admin;
    GRANT CREATE ON DATABASE postgres TO supabase_auth_admin;
    ALTER USER supabase_auth_admin SET search_path = '$DB_NAMESPACE';
    GRANT ALL PRIVILEGES ON SCHEMA public TO supabase_auth_admin;

     --zion user
    CREATE USER zion_user NOINHERIT LOGIN NOREPLICATION PASSWORD '$ZION_DB_USER_PASSWORD';
    GRANT USAGE ON SCHEMA $DB_NAMESPACE TO zion_user;

    -- zion admin
    CREATE USER zion_admin NOINHERIT LOGIN NOREPLICATION PASSWORD '$ZION_DB_ADMIN_PASSWORD';
    GRANT USAGE ON SCHEMA $DB_NAMESPACE TO zion_admin;

    -- brawler user
    CREATE USER brawler_user NOINHERIT LOGIN NOREPLICATION PASSWORD '$BRAWLER_DB_USER_PASSWORD';
    GRANT USAGE ON SCHEMA $DB_NAMESPACE TO brawler_user;

    -- brawler admin
    CREATE USER brawler_admin NOINHERIT LOGIN NOREPLICATION PASSWORD '$BRAWLER_DB_ADMIN_PASSWORD';
    GRANT USAGE ON SCHEMA $DB_NAMESPACE TO brawler_admin;
EOSQL