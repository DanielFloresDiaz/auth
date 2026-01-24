CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

-- Supabase super admin
CREATE USER auth_admin CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION auth_admin;
GRANT CREATE ON DATABASE postgres TO auth_admin;
GRANT CREATE, USAGE ON SCHEMA public TO auth_admin;
ALTER USER auth_admin SET search_path = 'auth';

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

GRANT solomon_auth_admin_role TO auth_admin;
GRANT solomon_auth_user_role TO solomon_auth_admin_role;
