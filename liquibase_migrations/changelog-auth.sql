--liquibase formatted sql

--------------------------- enums ---------------------------
--changeset solomon.auth:1 labels:auth context:auth
--comment: create api_key_permissions enum
CREATE TYPE "auth"."tier_organizations" AS ENUM (
  'low',
  'medium',
  'high'
);
--rollback DROP TYPE "auth"."api_key_permissions";

--changeset solomon.auth:2 labels:auth context:auth
--comment: create organization_roles enum
CREATE TYPE "auth"."organization_roles" AS ENUM (
  'admin',
  'client',
  'api_key',
  'project_admin'
);
--rollback DROP TYPE "auth"."organization_roles";

--changeset solomon.auth:3 labels:auth context:auth
--comment: create role permissions enum
CREATE TYPE "auth"."role_permissions" AS ENUM (
);
--rollback DROP TYPE "auth"."role_permissions";

--changeset solomon.auth:4 labels:auth context:auth
--comment: create projects table
CREATE TABLE IF NOT EXISTS "auth".projects (
	id uuid UNIQUE NOT NULL,
	name varchar(255) NOT NULL UNIQUE,
	description text NULL,
	rate_limits jsonb NULL,
	admin_id uuid NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT projects_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".projects;

--changeset solomon.auth:5 labels:auth context:auth
--comment: create organizations table
CREATE TABLE IF NOT EXISTS "auth".organizations (
	id uuid UNIQUE NOT NULL,
	project_id uuid NOT NULL,
	admin_id uuid UNIQUE NOT NULL,
	name varchar(255) NULL,
	description text NULL,
	tier "auth".tier_organizations DEFAULT 'low',
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT organizations_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE,
	CONSTRAINT organizations_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES "auth".users(id),
	CONSTRAINT organizations_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".organizations;

--changeset solomon.auth:6 labels:auth context:auth
--comment: create smtp_configs_organizations table
CREATE TABLE IF NOT EXISTS "auth".smtp_configs_organizations (
	id serial UNIQUE NOT NULL,
	organization_id uuid NOT NULL,
	domain varchar(255) NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT smtp_configs_organizations_pkey PRIMARY KEY (id),
	CONSTRAINT smtp_configs_organizations_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
	CONSTRAINT smtp_configs_domain_organization_id_unique UNIQUE (domain, organization_id)
);
--rollback DROP TABLE "auth".smtp_configs_organizations;

--changeset solomon.auth:7 labels:auth context:auth
--comment: create smtp_configs_projects table
CREATE TABLE IF NOT EXISTS "auth".smtp_configs_projects (
	id bigserial UNIQUE NOT NULL,
	project_id uuid UNIQUE NOT NULL,
	domain varchar(255) NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT smtp_configs_projects_pkey PRIMARY KEY (id),
	CONSTRAINT smtp_configs_projects_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE,
	CONSTRAINT smtp_configs_domain_project_id_unique UNIQUE (domain, project_id)
);
--rollback DROP TABLE "auth".smtp_configs_projects;

--changeset solomon.auth:8 labels:auth context:auth
--comment: create organization_roles_permissions table
CREATE TABLE IF NOT EXISTS "auth".organization_roles_permissions (
	id serial unique NOT NULL,
	organization_role "auth".organization_roles NOT NULL,
	permissions "auth".role_permissions NOT NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT organization_roles_permissions_organization_role_permission_unique UNIQUE (organization_role, permissions),
	CONSTRAINT organization_roles_permissions_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".organization_roles_permissions;

--changeset solomon.auth:9 labels:auth context:auth
--comment: add tier_model columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_model "public".tier_models DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_model "public".tier_models DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_model;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_model;

--changeset solomon.auth:10 labels:auth context:auth
--comment: add tier_time columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_time "public".tier_times DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_time "public".tier_times DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_time;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_time;

--changeset solomon.auth:11 labels:auth context:auth
--comment: add tier_usage columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_usage "public".tier_usages DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_usage "public".tier_usages DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_usage;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_usage;

--changeset solomon.auth:12 labels:auth context:auth
--comment: create api_keys table
CREATE TABLE IF NOT EXISTS "auth".api_keys (
	id uuid UNIQUE NOT NULL,
	organization_id uuid NOT NULL,
	project_id uuid NOT NULL,
	name text NOT NULL,
	description text,
	tier_model "public".tier_models NOT NULL DEFAULT 'low',
	tier_time "public".tier_times NOT NULL DEFAULT 'low',
	tier_usage "public".tier_usages NOT NULL DEFAULT 'low',
	"key" text UNIQUE NOT NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
	CONSTRAINT api_keys_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".api_keys;

--changeset solomon.auth:13 labels:auth context:auth
--comment: create table tier_organizations_tiers
CREATE TABLE IF NOT EXISTS "auth".tier_organizations_tiers (
	id serial UNIQUE NOT NULL,
	tier "auth".tier_organizations NOT NULL,
	tier_model "public".tier_models NOT NULL,
	tier_time "public".tier_times NOT NULL,
	tier_usage "public".tier_usages NOT NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT tier_organizations_tiers_pkey PRIMARY KEY (id),
	CONSTRAINT tier_organizations_tiers_tier_unique UNIQUE (tier)
);
--rollback DROP TABLE "auth".tier_organizations_tiers;

--changeset solomon.auth:14 labels:auth context:auth
--comment: alter table auth.users
ALTER TABLE "auth".users
        ADD COLUMN organization_id uuid NULL,
        ADD COLUMN project_id uuid NULL,
        ADD COLUMN organization_role "auth".organization_roles NOT NULL DEFAULT 'client',
        ADD CONSTRAINT users_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id) ON DELETE CASCADE,
        ADD CONSTRAINT users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id) ON DELETE CASCADE,
        ADD CONSTRAINT users_email_organization_id_unique UNIQUE (email, organization_id),
        ADD CONSTRAINT users_email_project_id_unique UNIQUE (email, project_id),
        ADD CONSTRAINT users_phone_organization_id_unique UNIQUE (phone, organization_id),
        ADD CONSTRAINT users_phone_project_id_unique UNIQUE (phone, project_id);
--rollback ALTER TABLE "auth".users
--rollback DROP COLUMN organization_id
--rollback DROP COLUMN project_id
--rollback DROP COLUMN organization_role
--rollback DROP CONSTRAINT users_project_id_fkey
--rollback DROP CONSTRAINT users_organization_id_fkey
--rollback DROP CONSTRAINT users_email_organization_id_unique
--rollback DROP CONSTRAINT users_email_project_id_unique
--rollback DROP CONSTRAINT users_phone_organization_id_unique
--rollback DROP CONSTRAINT users_phone_project_id_unique

--changeset solomon.auth:15 labels:auth context:auth
--comment: alter table auth.flow_state
ALTER TABLE "auth".flow_state
ADD COLUMN organization_id uuid NULL,
ADD COLUMN project_id uuid NULL,
ADD CONSTRAINT flow_state_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id),
ADD CONSTRAINT flow_state_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id);
--rollback ALTER TABLE "auth".flow_state
--rollback DROP COLUMN organization_id
--rollback DROP COLUMN project_id
--rollback DROP CONSTRAINT flow_state_organization_id_fkey
--rollback DROP CONSTRAINT flow_state_project_id_fkey

--changeset solomon.auth:16 labels:auth context:auth
--comment: alter table auth.identitites
ALTER TABLE "auth".identities
ADD COLUMN organization_id uuid NULL,
ADD COLUMN project_id uuid NULL,
ADD CONSTRAINT identities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id) ON DELETE CASCADE,
ADD CONSTRAINT identities_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id) ON DELETE CASCADE;
--rollback ALTER TABLE "auth".identities
--rollback DROP COLUMN organization_id
--rollback DROP COLUMN project_id
--rollback DROP CONSTRAINT identities_organization_id_fkey
--rollback DROP CONSTRAINT identities_project_id_fkey
--rollback DROP TRIGGER prevent_set_both_organization_and_project

--changeset solomon.auth:17 labels:auth context:auth
--comment: create table project_rate_limits
CREATE TABLE IF NOT EXISTS "auth".project_rate_limits (
	project_id uuid NOT NULL,
	user_id text NOT NULL,
	request_time timestamptz NOT NULL,
	CONSTRAINT project_rate_limits_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".project_rate_limits

--changeset solomon.auth:18 labels:auth context:auth
--comment: Drop unique constraint for users
ALTER TABLE "auth".users DROP CONSTRAINT users_email_organization_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT users_email_project_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT users_phone_organization_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT users_phone_project_id_unique;
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_email_organization_id_unique UNIQUE (email, organization_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_email_project_id_unique UNIQUE (email, project_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_phone_organization_id_unique UNIQUE (phone, organization_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_phone_project_id_unique UNIQUE (phone, project_id);

--changeset solomon.auth:19 labels:auth context:auth
--comment: Add uniqueness constraint for users -> email, project_id, organization_id and phone, project_id, organization_id
ALTER TABLE "auth".users ADD CONSTRAINT users_email_project_id_org_unique UNIQUE (email, project_id, organization_id);
ALTER TABLE "auth".users ADD CONSTRAINT users_phone_project_id_org_unique UNIQUE (phone, project_id, organization_id);
--rollback ALTER TABLE "auth".users DROP CONSTRAINT users_email_project_id_org_unique;
--rollback ALTER TABLE "auth".users DROP CONSTRAINT users_phone_project_id_org_unique;

--changeset solomon.auth:20 labels:auth context:auth
--comment: Add uniqueness constraint for users with NULL organization_id and same email, project_id and phone, project_id
CREATE UNIQUE INDEX IF NOT EXISTS users_email_project_id_org_null_unique 
ON "auth".users (email, project_id) 
WHERE organization_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_project_id_org_null_unique
ON "auth".users (phone, project_id)
WHERE organization_id IS NULL;
--rollback DROP INDEX IF EXISTS "auth".users_email_project_id_org_null_unique;
--rollback DROP INDEX IF EXISTS "auth".users_phone_project_id_org_null_unique;

--changeset solomon.auth-index:1 labels:auth context:auth
--comment: create index on tier_organizations_tiers tier
CREATE UNIQUE INDEX IF NOT EXISTS tier_organizations_tiers_tier_index ON "auth".tier_organizations_tiers (tier);
--rollback DROP INDEX "auth".tier_organizations_tiers_tier_index;

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

--changeset solomon.auth-rls:1 labels:auth context:auth
--comment: Enable rls and create policy for api_keys
ALTER TABLE "auth".api_keys ENABLE ROW LEVEL SECURITY;
CREATE POLICY api_keys_policy ON "auth".api_keys
FOR ALL
USING (
	organization_id = current_setting('app.current_organization_id')::uuid
	AND
	current_setting('app.organization_role')::text IN ('admin', 'project_admin')
)
WITH CHECK (
	organization_id = current_setting('app.current_organization_id')::uuid
	AND
	current_setting('app.organization_role')::text IN ('admin', 'project_admin')
);
--rollback DROP POLICY api_keys_policy ON "auth".api_keys;

--changeset solomon.auth:grant:1 labels:auth context:auth
--comment: grant select, insert, update, delete on all api_keys in schema auth to roles
GRANT SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys TO solomon_role;
--rollback REVOKE SELECT, INSERT, UPDATE, DELETE ON "auth".api_keys FROM solomon_role;

--changeset solomon.auth:grant:2 labels:auth context:auth
--comment: grant select on projects to solomon_role
GRANT SELECT ON "auth".projects TO solomon_role;
--rollback REVOKE SELECT ON "auth".projects FROM solomon_role;

--changeset solomon.auth:grant:3 labels:auth context:auth
--comment: grant select, insert on project_rate_limits to solomon_role
GRANT SELECT, INSERT ON "auth".project_rate_limits TO solomon_role;
--rollback REVOKE SELECT, INSERT ON "auth".project_rate_limits FROM solomon_role;

--changeset solomon.auth:grant:4 labels:auth context:auth
--comment: grant select on organizations to solomon_role
GRANT SELECT ON "auth".organizations TO solomon_role;
--rollback REVOKE SELECT ON "auth".organizations FROM solomon_role;

--changeset solomon.auth:grant:5 labels:auth context:auth
--comment: grant select on tier_organizations_tiers to solomon_role
GRANT SELECT ON "auth".tier_organizations_tiers TO solomon_role;
--rollback REVOKE SELECT ON "auth".tier_organizations_tiers FROM solomon_role;

--changeset solomon.auth:grant:6 labels:auth context:auth
--comment: grant delete on project_rate_limits to solomon_role
GRANT DELETE ON "auth".project_rate_limits TO solomon_role;
--rollback REVOKE DELETE ON "auth".project_rate_limits FROM solomon_role;