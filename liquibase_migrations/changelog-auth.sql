--liquibase formatted sql

--------------------------- enums ---------------------------
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
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT organizations_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE,
	CONSTRAINT organizations_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES "auth".users(id),
	CONSTRAINT organizations_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".organizations;

--changeset solomon.auth:5.1 labels:auth context:auth
--comment: create organizations_tier table
CREATE TABLE IF NOT EXISTS "auth".organizations_tier (
	organization_id uuid PRIMARY KEY,
	tier text DEFAULT 'free',
	admin_tier_model "public".tier_models DEFAULT 'low',
	admin_tier_time "public".tier_times DEFAULT 'low',
	admin_tier_usage "public".tier_usages DEFAULT 'low',
	client_tier_model "public".tier_models DEFAULT 'low',
	client_tier_time "public".tier_times DEFAULT 'low',
	client_tier_usage "public".tier_usages DEFAULT 'low',
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT organizations_tier_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".organizations_tier;

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
--comment: create table projects_tiers
CREATE TABLE IF NOT EXISTS "auth".projects_tiers (
	id serial UNIQUE NOT NULL,
	project_id uuid NOT NULL,
	tier text NOT NULL,
	tier_model "public".tier_models NOT NULL,
	tier_time "public".tier_times NOT NULL,
	tier_usage "public".tier_usages NOT NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT projects_tiers_pkey PRIMARY KEY (id),
	CONSTRAINT projects_tiers_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE,
	CONSTRAINT projects_tiers_project_id_tier_unique UNIQUE (project_id, tier)
);
--rollback DROP TABLE "auth".projects_tiers;

--changeset solomon.auth:14 labels:auth context:auth
--comment: alter table auth.users
ALTER TABLE "auth".users
        ADD COLUMN organization_id uuid NULL,
        ADD COLUMN project_id uuid NOT NULL,
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
ADD COLUMN project_id uuid NOT NULL,
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
ADD COLUMN project_id uuid NOT NULL,
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
