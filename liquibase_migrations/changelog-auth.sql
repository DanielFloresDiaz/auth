--liquibase formatted sql

--------------------------- enums ---------------------------
--changeset solomon.auth:2 labels:auth context:auth
--comment: create organization_roles enum
DO $$ BEGIN
CREATE TYPE "auth"."organization_roles" AS ENUM (
  'admin',
  'client',
  'api_key',
  'project_admin'
);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP TYPE "auth"."organization_roles";

--changeset solomon.auth:3 labels:auth context:auth
--comment: create role permissions enum
DO $$ BEGIN
CREATE TYPE "auth"."role_permissions" AS ENUM (
);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
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

--changeset solomon.auth:12.1 labels:auth context:auth
--comment: create spend_limit_status enum
CREATE TYPE "auth"."spend_limit_status" AS ENUM ('active', 'non_active', 'pending_cancellation');
--rollback DROP TYPE "auth"."spend_limit_status";

--changeset solomon.auth:12.2 labels:auth context:auth
--comment: create period_interval enum
CREATE TYPE "auth"."period_interval" AS ENUM ('monthly', 'quarterly', 'annual');
--rollback DROP TYPE "auth"."period_interval";

--changeset solomon.auth:12.2.1 labels:auth context:auth
--comment: create subscription_kind enum
CREATE TYPE "auth"."subscription_kind" AS ENUM ('prepaid', 'billing');
--rollback DROP TYPE "auth"."subscription_kind";

--changeset solomon.auth:12.3 labels:auth context:auth
--comment: create credit_status enum
CREATE TYPE "auth"."credit_status" AS ENUM ('active', 'disabled');
--rollback DROP TYPE "auth"."credit_status";

--changeset solomon.auth:14 labels:auth context:auth splitStatements:false
--comment: alter table auth.users
ALTER TABLE "auth".users ADD COLUMN IF NOT EXISTS organization_id uuid NULL;
ALTER TABLE "auth".users ADD COLUMN IF NOT EXISTS project_id uuid NOT NULL;
ALTER TABLE "auth".users ADD COLUMN IF NOT EXISTS organization_role "auth".organization_roles NOT NULL DEFAULT 'client';
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_project_id_fkey') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_organization_id_fkey') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_email_organization_id_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_email_organization_id_unique UNIQUE (email, organization_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_email_project_id_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_email_project_id_unique UNIQUE (email, project_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_phone_organization_id_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_phone_organization_id_unique UNIQUE (phone, organization_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_phone_project_id_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_phone_project_id_unique UNIQUE (phone, project_id);
    END IF;
END $$;
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

--changeset solomon.auth:15 labels:auth context:auth splitStatements:false
--comment: alter table auth.flow_state
ALTER TABLE "auth".flow_state ADD COLUMN IF NOT EXISTS organization_id uuid NULL;
ALTER TABLE "auth".flow_state ADD COLUMN IF NOT EXISTS project_id uuid NOT NULL;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'flow_state_organization_id_fkey') THEN
        ALTER TABLE "auth".flow_state ADD CONSTRAINT flow_state_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'flow_state_project_id_fkey') THEN
        ALTER TABLE "auth".flow_state ADD CONSTRAINT flow_state_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id);
    END IF;
END $$;
--rollback ALTER TABLE "auth".flow_state
--rollback DROP COLUMN organization_id
--rollback DROP COLUMN project_id
--rollback DROP CONSTRAINT flow_state_organization_id_fkey
--rollback DROP CONSTRAINT flow_state_project_id_fkey

--changeset solomon.auth:16 labels:auth context:auth splitStatements:false
--comment: alter table auth.identitites
ALTER TABLE "auth".identities ADD COLUMN IF NOT EXISTS organization_id uuid NULL;
ALTER TABLE "auth".identities ADD COLUMN IF NOT EXISTS project_id uuid NOT NULL;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'identities_organization_id_fkey') THEN
        ALTER TABLE "auth".identities ADD CONSTRAINT identities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "auth".organizations (id) ON DELETE CASCADE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'identities_project_id_fkey') THEN
        ALTER TABLE "auth".identities ADD CONSTRAINT identities_project_id_fkey FOREIGN KEY (project_id) REFERENCES "auth".projects (id) ON DELETE CASCADE;
    END IF;
END $$;
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
ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_email_organization_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_email_project_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_phone_organization_id_unique;
ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_phone_project_id_unique;
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_email_organization_id_unique UNIQUE (email, organization_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_email_project_id_unique UNIQUE (email, project_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_phone_organization_id_unique UNIQUE (phone, organization_id);
--rollback ALTER TABLE "auth".users ADD CONSTRAINT users_phone_project_id_unique UNIQUE (phone, project_id);

--changeset solomon.auth:19 labels:auth context:auth splitStatements:false
--comment: Add uniqueness constraint for users -> email, project_id, organization_id and phone, project_id, organization_id
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_email_project_id_org_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_email_project_id_org_unique UNIQUE (email, project_id, organization_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_phone_project_id_org_unique') THEN
        ALTER TABLE "auth".users ADD CONSTRAINT users_phone_project_id_org_unique UNIQUE (phone, project_id, organization_id);
    END IF;
END $$;
--rollback ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_email_project_id_org_unique;
--rollback ALTER TABLE "auth".users DROP CONSTRAINT IF EXISTS users_phone_project_id_org_unique;

--changeset solomon.auth:20 labels:auth context:auth
--comment: Add uniqueness constraint for users with NULL organization_id and same email, project_id and phone, project_id
CREATE UNIQUE INDEX IF NOT EXISTS users_email_project_id_org_null_unique 
ON "auth".users (email, project_id) 
WHERE organization_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_project_id_org_null_unique
ON "auth".users (phone, project_id)
WHERE organization_id IS NULL;
--rollback DROP INDEX IF EXISTS "auth".users_email_project_id_org_null_unique;

--changeset solomon.auth:21 labels:auth context:auth
--comment: Replace global identities unique constraint with project-scoped one to allow the same external identity to sign in to multiple projects
ALTER TABLE "auth".identities DROP CONSTRAINT IF EXISTS identities_provider_id_provider_unique;
ALTER TABLE "auth".identities ADD CONSTRAINT identities_provider_id_provider_project_id_unique UNIQUE (provider_id, provider, project_id);
--rollback ALTER TABLE "auth".identities DROP CONSTRAINT IF EXISTS identities_provider_id_provider_project_id_unique;
--rollback ALTER TABLE "auth".identities ADD CONSTRAINT identities_provider_id_provider_unique UNIQUE (provider_id, provider);

--changeset solomon.auth:22 labels:auth context:auth
--comment: create organizations_periodic_limit table
CREATE TABLE IF NOT EXISTS "auth".organizations_periodic_limit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL,
    subscription_kind "auth".subscription_kind NOT NULL DEFAULT 'prepaid',
    amount real NOT NULL CHECK (amount > 0),
    period_interval "auth".period_interval NOT NULL DEFAULT 'monthly',
    period_start timestamptz NOT NULL,
    status "auth".spend_limit_status NOT NULL DEFAULT 'non_active',
    cancel_at_period_end boolean NOT NULL DEFAULT false,
    created_at timestamptz DEFAULT current_timestamp,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT organizations_periodic_limit_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
    CONSTRAINT organizations_periodic_limit_org_kind_unique UNIQUE (organization_id, subscription_kind)
);
CREATE INDEX IF NOT EXISTS organizations_periodic_limit_organization_id_idx
    ON "auth".organizations_periodic_limit (organization_id);
--rollback DROP TABLE "auth".organizations_periodic_limit;

--changeset solomon.auth:23 labels:auth context:auth
--comment: create organizations_spend_credits table
CREATE TABLE IF NOT EXISTS "auth".organizations_spend_credits (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL UNIQUE,
    amount real NOT NULL CHECK (amount >= 0),
    expires_at timestamptz NOT NULL,
    status "auth".credit_status NOT NULL DEFAULT 'active',
    last_added_at timestamptz NOT NULL DEFAULT current_timestamp,
    credits_disabled boolean NOT NULL DEFAULT false,
    created_at timestamptz DEFAULT current_timestamp,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT organizations_spend_credits_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".organizations_spend_credits;

--changeset solomon.auth:24 labels:auth context:auth
--comment: normalize organizations_periodic_limit period_start to 1st of month
UPDATE auth.organizations_periodic_limit
SET period_start = date_trunc('month', period_start AT TIME ZONE 'UTC') AT TIME ZONE 'UTC'
WHERE period_start IS NOT NULL;
--rollback SELECT 1;

--changeset solomon.auth:25.1 labels:auth context:auth
--comment: create organizations_credits_spent table
CREATE TABLE IF NOT EXISTS "auth".organizations_credits_spent (
    organization_id uuid PRIMARY KEY,
    amount_spent real NOT NULL DEFAULT 0,
    created_at timestamptz DEFAULT current_timestamp,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT organizations_credits_spent_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".organizations_credits_spent;

--changeset solomon.auth:25.2 labels:auth context:auth
--comment: create organizations_periodic_spent table
CREATE TABLE IF NOT EXISTS "auth".organizations_periodic_spent (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL,
    subscription_kind "auth".subscription_kind NOT NULL,
    period tstzrange NOT NULL,
    amount_spent real NOT NULL DEFAULT 0,
    created_at timestamptz DEFAULT current_timestamp,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT organizations_periodic_spent_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
    CONSTRAINT organizations_periodic_spent_org_kind_period_unique
        UNIQUE (organization_id, subscription_kind, period)
);
--rollback DROP TABLE "auth".organizations_periodic_spent;

--changeset solomon.auth:26 labels:auth context:auth
--comment: create organization_usage_summary table
CREATE TABLE IF NOT EXISTS "auth".organization_usage_summary (
    organization_id uuid PRIMARY KEY,
    project_id uuid NOT NULL,
    prompt_tokens bigint NOT NULL DEFAULT 0,
    completion_tokens bigint NOT NULL DEFAULT 0,
    cached_tokens bigint NOT NULL DEFAULT 0,
    total_tokens integer GENERATED ALWAYS AS (prompt_tokens + completion_tokens) STORED,
    total_price real NOT NULL DEFAULT 0,
    total_seconds real NOT NULL DEFAULT 0,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT organization_usage_summary_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
    CONSTRAINT organization_usage_summary_project_id_fkey
        FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".organization_usage_summary;

--changeset solomon.auth:27 labels:auth context:auth
--comment: create project_usage_summary table
CREATE TABLE IF NOT EXISTS "auth".project_usage_summary (
    project_id uuid PRIMARY KEY,
    prompt_tokens bigint NOT NULL DEFAULT 0,
    completion_tokens bigint NOT NULL DEFAULT 0,
    cached_tokens bigint NOT NULL DEFAULT 0,
    total_tokens integer GENERATED ALWAYS AS (prompt_tokens + completion_tokens) STORED,
    total_price real NOT NULL DEFAULT 0,
    total_seconds real NOT NULL DEFAULT 0,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT project_usage_summary_project_id_fkey
        FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".project_usage_summary;

--changeset solomon.auth:28 labels:auth context:auth
--comment: create user_usage_summary table
CREATE TABLE IF NOT EXISTS "auth".user_usage_summary (
    user_id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    project_id uuid NOT NULL,
    prompt_tokens bigint NOT NULL DEFAULT 0,
    completion_tokens bigint NOT NULL DEFAULT 0,
    cached_tokens bigint NOT NULL DEFAULT 0,
    total_tokens integer GENERATED ALWAYS AS (prompt_tokens + completion_tokens) STORED,
    total_price real NOT NULL DEFAULT 0,
    total_seconds real NOT NULL DEFAULT 0,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT user_usage_summary_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES "auth".users(id) ON DELETE CASCADE,
    CONSTRAINT user_usage_summary_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
    CONSTRAINT user_usage_summary_project_id_fkey
        FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".user_usage_summary;

--changeset solomon.auth:29 labels:auth context:auth
--comment: create api_key_usage_summary table
CREATE TABLE IF NOT EXISTS "auth".api_key_usage_summary (
    api_key_id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    project_id uuid NOT NULL,
    prompt_tokens bigint NOT NULL DEFAULT 0,
    completion_tokens bigint NOT NULL DEFAULT 0,
    cached_tokens bigint NOT NULL DEFAULT 0,
    total_tokens integer GENERATED ALWAYS AS (prompt_tokens + completion_tokens) STORED,
    total_price real NOT NULL DEFAULT 0,
    total_seconds real NOT NULL DEFAULT 0,
    updated_at timestamptz DEFAULT current_timestamp,
    CONSTRAINT api_key_usage_summary_api_key_id_fkey
        FOREIGN KEY (api_key_id) REFERENCES "auth".api_keys(id) ON DELETE CASCADE,
    CONSTRAINT api_key_usage_summary_organization_id_fkey
        FOREIGN KEY (organization_id) REFERENCES "auth".organizations(id) ON DELETE CASCADE,
    CONSTRAINT api_key_usage_summary_project_id_fkey
        FOREIGN KEY (project_id) REFERENCES "auth".projects(id) ON DELETE CASCADE
);
--rollback DROP TABLE "auth".api_key_usage_summary;
