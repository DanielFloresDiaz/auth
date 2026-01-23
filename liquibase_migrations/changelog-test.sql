--liquibase formatted sql

--changeset solomon.auth:1 labels:auth context:auth
--comment: initial setup
CREATE USER supabase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;
CREATE USER auth_admin CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';
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
--rollback DROP ROLE rl_auth_user_role; DROP ROLE solomon_auth_user_role; DROP ROLE solomon_auth_admin_role; DROP USER auth_admin; DROP USER supabase_admin; DROP SCHEMA IF EXISTS auth CASCADE; DROP EXTENSION IF EXISTS "pgcrypto";