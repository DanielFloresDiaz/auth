-- Create roles with appropriate privileges
CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

-- Create auth_admin user
CREATE USER auth_admin CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';

-- Create schema and grant permissions on schema level
CREATE SCHEMA IF NOT EXISTS auth;
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

-- Grant usage permissions to custom roles
GRANT USAGE ON SCHEMA auth TO solomon_auth_admin_role;
GRANT USAGE ON SCHEMA auth TO solomon_auth_user_role;
GRANT USAGE ON SCHEMA auth TO rl_auth_user_role;

-- Grant these roles to existing users if needed
GRANT solomon_auth_admin_role TO auth_admin;
GRANT solomon_auth_user_role TO solomon_auth_admin_role;
-- GRANT solomon_auth_admin_role TO solomon_admin;