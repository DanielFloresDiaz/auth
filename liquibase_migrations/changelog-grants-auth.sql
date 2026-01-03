--liquibase formatted sql

--changeset solomon.auth:grant:1 labels:auth context:auth
--comment: grant select, insert, update, delete on all api_keys in schema auth to roles
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys TO solomon_user_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys FROM solomon_user_role;

--changeset solomon.auth:grant:2 labels:auth context:auth
--comment: grant select on projects to solomon_user_role
GRANT SELECT ON "auth".projects TO solomon_user_role;
--rollback REVOKE SELECT ON "auth".projects FROM solomon_user_role;

--changeset solomon.auth:grant:3 labels:auth context:auth
--comment: grant select, insert, delete on project_rate_limits to solomon_user_role
GRANT SELECT, INSERT, DELETE ON "auth".project_rate_limits TO solomon_user_role;
--rollback REVOKE SELECT, INSERT, DELETE ON "auth".project_rate_limits FROM solomon_user_role;

--changeset solomon.auth:grant:4 labels:auth context:auth
--comment: grant select on organizations to solomon_user_role
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".organizations TO solomon_user_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".organizations FROM solomon_user_role;

--changeset solomon.auth:grant:5 labels:auth context:auth
--comment: grant select on tier_organizations_tiers to solomon_user_role
GRANT SELECT ON "auth".tier_organizations_tiers TO solomon_user_role;
--rollback REVOKE SELECT ON "auth".tier_organizations_tiers FROM solomon_user_role;

--changeset solomon.auth:grant:6 labels:auth context:auth
--comment: grant update on tier_organizations_tiers to solomon_admin_role
GRANT UPDATE ON "auth".tier_organizations_tiers TO solomon_admin_role;
--rollback REVOKE UPDATE ON "auth".tier_organizations_tiers FROM solomon_admin_role;

