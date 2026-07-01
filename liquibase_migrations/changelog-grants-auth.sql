--liquibase formatted sql

--changeset solomon.auth:grant:1 labels:auth context:auth
--comment: grant select, insert, update, delete on all api_keys in schema auth to roles
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys TO solomon_auth_user_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys FROM solomon_auth_user_role;

--changeset solomon.auth:grant:2 labels:auth context:auth
--comment: grant select on projects to solomon_auth_user_role
GRANT SELECT ON "auth".projects TO solomon_auth_user_role;
--rollback REVOKE SELECT ON "auth".projects FROM solomon_auth_user_role;

--changeset solomon.auth:grant:3 labels:auth context:auth
--comment: grant select, insert, delete on project_rate_limits to solomon_auth_user_role
GRANT SELECT, INSERT, DELETE ON "auth".project_rate_limits TO solomon_auth_user_role;
--rollback REVOKE SELECT, INSERT, DELETE ON "auth".project_rate_limits FROM solomon_auth_user_role;

--changeset solomon.auth:grant:4 labels:auth context:auth
--comment: grant select on organizations to solomon_auth_user_role
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".organizations TO solomon_auth_user_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".organizations FROM solomon_auth_user_role;

--changeset solomon.auth:grant:4.1 labels:auth context:auth
--comment: grant select on organizations_tier to solomon_auth_user_role
GRANT SELECT ON "auth".organizations_tier TO solomon_auth_user_role;
--rollback REVOKE SELECT ON "auth".organizations_tier FROM solomon_auth_user_role;

--changeset solomon.auth:grant:6 labels:auth context:auth
--comment: grant UPDATE on users to solomon_auth_admin_role
GRANT SELECT, UPDATE ON "auth".users TO solomon_auth_admin_role;
--rollback REVOKE UPDATE ON "auth".users FROM solomon_auth_admin_role;

--changeset solomon.auth:grant:7 labels:auth context:auth
--comment: grant INSERT, UPDATE on organizations_tier to solomon_auth_admin_role
GRANT INSERT, UPDATE ON "auth".organizations_tier TO solomon_auth_admin_role;
--rollback REVOKE INSERT, UPDATE ON "auth".organizations_tier FROM solomon_auth_admin_role;

--changeset solomon.auth:grant:7.1 labels:auth context:auth
--comment: grant SELECT on organizations_periodic_limit to solomon_auth_user_role
GRANT SELECT ON "auth".organizations_periodic_limit TO solomon_auth_user_role;
--rollback REVOKE SELECT ON "auth".organizations_periodic_limit FROM solomon_auth_user_role;

--changeset solomon.auth:grant:7.2 labels:auth context:auth
--comment: grant SELECT on organizations_spend_credits to solomon_auth_user_role
GRANT SELECT ON "auth".organizations_spend_credits TO solomon_auth_user_role;
--rollback REVOKE SELECT ON "auth".organizations_spend_credits FROM solomon_auth_user_role;

--changeset solomon.auth:grant:7.3 labels:auth context:auth
--comment: grant INSERT, UPDATE on organizations_periodic_limit to solomon_auth_admin_role
GRANT INSERT, UPDATE ON "auth".organizations_periodic_limit TO solomon_auth_admin_role;
--rollback REVOKE INSERT, UPDATE ON "auth".organizations_periodic_limit FROM solomon_auth_admin_role;

--changeset solomon.auth:grant:7.4 labels:auth context:auth
--comment: grant INSERT, UPDATE on organizations_spend_credits to solomon_auth_admin_role
GRANT INSERT, UPDATE ON "auth".organizations_spend_credits TO solomon_auth_admin_role;
--rollback REVOKE INSERT, UPDATE ON "auth".organizations_spend_credits FROM solomon_auth_admin_role;

--changeset solomon.auth:grant:8 labels:auth context:auth
--comment: grant SELECT on projects to rl_auth_user_role
GRANT SELECT ON "auth".projects TO rl_auth_user_role;
--rollback REVOKE SELECT ON "auth".projects FROM rl_auth_user_role;

--changeset solomon.auth:grant:9 labels:auth context:auth
--comment: grant SELECT, INSERT, UPDATE, DELETE on project_rate_limits to rl_auth_user
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".project_rate_limits TO rl_auth_user_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".project_rate_limits
