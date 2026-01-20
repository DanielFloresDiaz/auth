--liquibase formatted sql

--changeset solomon.auth-index:1 labels:auth context:auth
--comment: create index on projects_tiers tier
CREATE INDEX IF NOT EXISTS projects_tiers_tier_index ON "auth".projects_tiers (tier);
--rollback DROP INDEX "auth".projects_tiers_tier_index;

--changeset solomon.auth-index:2 labels:auth context:auth
--comment: create index on api_keys organization_id
CREATE INDEX IF NOT EXISTS api_keys_key_index ON "auth".api_keys (organization_id);
--rollback DROP INDEX "auth".api_keys_key_index;

--changeset solomon.auth-index:3 labels:auth context:auth
--comment: create index on organizations project_id
CREATE INDEX IF NOT EXISTS organizations_project_id_index ON "auth".organizations (project_id);
--rollback DROP INDEX "auth".organizations_project_id_index;

--changeset solomon.auth-index:4 labels:auth context:auth
--comment: create index on organizations admin_id
CREATE INDEX IF NOT EXISTS organizations_admin_id_index ON "auth".organizations (admin_id);
--rollback DROP INDEX "auth".organizations_admin_id_index;

--changeset solomon.auth-index:5 labels:auth context:auth
--comment: create index on projects name
CREATE UNIQUE INDEX IF NOT EXISTS projects_name_index ON "auth".projects (name);
--rollback DROP INDEX "auth".projects_name_index;

--changeset solomon.auth-index:6 labels:auth context:auth
--comment: create index on smtp_configs_organizations organization_id
CREATE INDEX IF NOT EXISTS smtp_configs_organizations_organization_id_index ON "auth".smtp_configs_organizations (organization_id);
--rollback DROP INDEX "auth".smtp_configs_organizations_organization_id_index;

--changeset solomon.auth-index:7 labels:auth context:auth
--comment: create index on smtp_configs_projects project_id
CREATE INDEX IF NOT EXISTS smtp_configs_projects_project_id_index ON "auth".smtp_configs_projects (project_id);
--rollback DROP INDEX "auth".smtp_configs_projects_project_id_index;

--changeset solomon.auth-index:8 labels:auth context:auth
--comment: create index on users organization_id
CREATE INDEX IF NOT EXISTS users_organization_id_index ON "auth".users (organization_id);
--rollback DROP INDEX "auth".users_organization_id_index;

--changeset solomon.auth-index:9 labels:auth context:auth
--comment: create index on users project_id
CREATE INDEX IF NOT EXISTS users_project_id_index ON "auth".users (project_id);
--rollback DROP INDEX "auth".users_project_id_index;

--changeset solomon.auth-index:10 labels:auth context:auth
--comment: create index on flow_state organization_id
CREATE INDEX IF NOT EXISTS flow_state_organization_id_index ON "auth".flow_state (organization_id);
--rollback DROP INDEX "auth".flow_state_organization_id_index;

--changeset solomon.auth-index:11 labels:auth context:auth
--comment: create index on flow_state project_id
CREATE INDEX IF NOT EXISTS flow_state_project_id_index ON "auth".flow_state (project_id);
--rollback DROP INDEX "auth".flow_state_project_id_index;

--changeset solomon.auth-index:12 labels:auth context:auth
--comment: create index on identities organization_id
CREATE INDEX IF NOT EXISTS identities_organization_id_index ON "auth".identities (organization_id);
--rollback DROP INDEX "auth".identities_organization_id_index;

--changeset solomon.auth-index:13 labels:auth context:auth
--comment: create index on identities project_id
CREATE INDEX IF NOT EXISTS identities_project_id_index ON "auth".identities (project_id);
--rollback DROP INDEX "auth".identities_project_id_index;

--changeset solomon.auth-index:14 labels:auth context:auth
--comment: create unique index on users email, organization_id, project_id
CREATE UNIQUE INDEX IF NOT EXISTS users_email_partial_key ON "auth".users (email, organization_id, project_id) WHERE (is_sso_user = false);
COMMENT ON INDEX "auth".users_email_partial_key IS 'auth: a partial unique index that applies only when is_sso_user is false';
--rollback DROP INDEX "auth".users_email_partial_key;
