-- Create roles with appropriate privileges
CREATE ROLE auth_admin_role NOINHERIT;
CREATE ROLE solomon_role NOINHERIT;

-- Create schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create pgcrypto extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Assign privileges to roles
GRANT CREATE ON DATABASE postgres_auth TO auth_admin_role;
GRANT ALL PRIVILEGES ON SCHEMA public TO auth_admin_role;
GRANT ALL PRIVILEGES ON SCHEMA auth TO auth_admin_role;

-- Grant usage permissions to solomon_role
GRANT USAGE ON SCHEMA auth TO solomon_role;

-- Grant roles to existing users
-- Note: These users must already exist in Google Cloud SQL
GRANT auth_admin_role TO auth_admin;

-- Set search paths for existing users
ALTER ROLE auth_admin SET search_path TO auth;
ALTER ROLE solomon_role SET search_path TO auth;


-- Grant these roles to existing users if needed
-- GRANT solomon_role TO solomon_admin;
-- GRANT solomon_role TO solomon_user;