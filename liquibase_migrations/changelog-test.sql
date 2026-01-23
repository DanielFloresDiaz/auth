--liquibase formatted sql

--changeset solomon.auth:1 labels:auth context:auth
--comment: initial setup
-- Set search paths for existing users and roles
-- Create schema and grant permissions on schema level
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION auth_admin;
GRANT CREATE ON DATABASE postgres TO auth_admin;
GRANT CREATE, USAGE ON SCHEMA public TO auth_admin;
ALTER ROLE auth_admin SET search_path TO auth;
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE ROLE solomon_auth_admin_role;
CREATE ROLE solomon_auth_user_role;
CREATE ROLE rl_auth_user_role;
GRANT USAGE ON SCHEMA auth TO solomon_auth_admin_role;
GRANT USAGE ON SCHEMA auth TO solomon_auth_user_role;
GRANT USAGE ON SCHEMA auth TO rl_auth_user_role;
GRANT solomon_auth_admin_role TO auth_admin;
GRANT solomon_auth_user_role TO solomon_auth_admin_role;
--rollback DROP ROLE solomon_auth_admin_role;
--rollback DROP ROLE solomon_auth_user_role;
--rollback DROP ROLE rl_auth_user_role;