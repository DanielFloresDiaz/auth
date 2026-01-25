-- Create roles with appropriate privileges
CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

-- Create auth_admin user
CREATE USER auth_admin CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';

-- Grant auth_admin to postgres user
GRANT auth_admin TO postgres; 

-- Create schema and grant permissions on schema level
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION auth_admin;
GRANT CREATE ON DATABASE $POSTGRES_DB TO auth_admin;
GRANT CREATE, USAGE ON SCHEMA public TO auth_admin;

-- Set search paths for existing users and roles
ALTER ROLE auth_admin SET search_path TO auth;

-- Create pgcrypto extension with superuser privileges
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Define custom roles for application use
CREATE ROLE solomon_auth_admin_role;
CREATE ROLE solomon_auth_user_role;
CREATE ROLE rl_auth_user_role;

-- Grant custom roles to users
GRANT solomon_auth_admin_role TO solomon_admin;
GRANT rl_auth_user_role TO rl_user;
GRANT solomon_auth_user_role TO zion_admin;
GRANT solomon_auth_user_role TO zion_user;
GRANT solomon_auth_user_role TO brawler_admin;
GRANT solomon_auth_user_role TO brawler_user;

-- Grant usage permissions to all users
GRANT USAGE ON SCHEMA auth TO solomon_admin;
GRANT USAGE ON SCHEMA auth TO rl_user;
GRANT USAGE ON SCHEMA auth TO zion_admin;
GRANT USAGE ON SCHEMA auth TO zion_user;
GRANT USAGE ON SCHEMA auth TO brawler_admin;
GRANT USAGE ON SCHEMA auth TO brawler_user;

-- Grant these roles to existing users if needed
GRANT solomon_auth_admin_role TO auth_admin;
GRANT solomon_auth_user_role TO solomon_auth_admin_role;
-- GRANT solomon_auth_admin_role TO solomon_admin;