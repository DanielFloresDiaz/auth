--liquibase formatted sql

--changeset solomon.auth:1 labels:auth context:auth
--comment: initial setup
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