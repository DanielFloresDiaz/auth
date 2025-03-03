--liquibase formatted sql

--------------------------- enums ---------------------------
--changeset solomon.auth:1 labels:auth context:auth
--comment: create tier_organizations enum
CREATE TYPE "auth"."tier_organizations" AS ENUM (
  'low',
  'medium',
  'high'
);
--rollback DROP TYPE "auth"."tier_organizations";

--changeset solomon.auth:2 labels:auth context:auth
--comment: create organization_roles enum
CREATE TYPE "auth"."organization_roles" AS ENUM (
  'admin',
  'client',
  'api_key'
);
--rollback DROP TYPE "auth"."organization_roles";

--changeset solomon.auth:3 labels:auth context:auth
--comment: create projects table
CREATE TABLE IF NOT EXISTS "auth".projects (
	id uuid UNIQUE NOT NULL,
	name varchar(255) NOT NULL UNIQUE,
	description text NULL,
	rate_limits jsonb NULL,
	created_at timestamptz DEFAULT current_timestamp,
	updated_at timestamptz DEFAULT current_timestamp,
	CONSTRAINT projects_pkey PRIMARY KEY (id)
);
--rollback DROP TABLE "auth".projects;

--changeset solomon.auth:4 labels:auth context:auth
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

--changeset solomon.auth:5 labels:auth context:auth
--comment: add tier_model columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_model text DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_model text DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_model;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_model;

--changeset solomon.auth:6 labels:auth context:auth
--comment: add tier_time columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_time text DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_time text DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_time;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_time;

--changeset solomon.auth:7 labels:auth context:auth
--comment: add tier_usages columns to organizations table
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS admin_tier_usage text DEFAULT 'low';
ALTER TABLE "auth".organizations ADD COLUMN IF NOT EXISTS client_tier_usage text DEFAULT 'low';
--rollback ALTER TABLE "auth".organizations DROP COLUMN admin_tier_usage;
--rollback ALTER TABLE "auth".organizations DROP COLUMN client_tier_usage;

--changeset solomon.auth:8 labels:auth context:auth
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

--changeset solomon.auth:9 labels:auth context:auth
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

--changeset solomon.auth:10 labels:auth context:auth
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