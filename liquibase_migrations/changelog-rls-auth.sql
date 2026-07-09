--liquibase formatted sql

--changeset solomon.auth-rls:1 labels:auth context:auth
--comment: Enable rls and create policy for api_keys
ALTER TABLE "auth".api_keys ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
    CREATE POLICY api_keys_policy ON "auth".api_keys
    FOR ALL
    USING (
        (
            organization_id = current_setting('app.current_organization', true)::uuid
            AND
            current_setting('app.organization_role', true)::text = 'admin'
        )
        OR
        (
            current_setting('app.organization_role', true)::text = 'project_admin'
            AND
            current_setting('app.current_project', true)::uuid = project_id
        )
    )
    WITH CHECK (
        (
            organization_id = current_setting('app.current_organization', true)::uuid
            AND
            current_setting('app.organization_role', true)::text = 'admin'
        )
        OR
        (
            current_setting('app.organization_role', true)::text = 'project_admin'
            AND
            current_setting('app.current_project', true)::uuid = project_id
        )
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS api_keys_policy ON "auth".api_keys;

--changeset solomon.auth-rls:2 labels:auth context:auth
--comment: Enable row level security on auth.organizations
ALTER TABLE "auth".organizations ENABLE ROW LEVEL SECURITY;
--rollback ALTER TABLE "auth".organizations DISABLE ROW LEVEL SECURITY;

--changeset solomon.auth-rls:3 labels:auth context:auth
--comment: Create SELECT policy for auth.organizations (current owner = admin_id or project_admin)
DO $$ BEGIN
    CREATE POLICY organizations_select_policy ON "auth".organizations
    FOR SELECT
    USING (
        current_setting('app.current_owner', true)::uuid = admin_id
        OR
        (
            current_setting('app.organization_role', true)::text = 'project_admin'
            AND
            current_setting('app.current_project', true)::uuid = project_id
        )
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS organizations_select_policy ON "auth".organizations;

--changeset solomon.auth-rls:4 labels:auth context:auth
--comment: Create UPDATE policy for auth.organizations (current owner = admin_id or project_admin)
DO $$ BEGIN
    CREATE POLICY organizations_update_policy ON "auth".organizations
    FOR UPDATE
    USING (
        current_setting('app.current_owner', true)::uuid = admin_id
        OR
        (
            current_setting('app.organization_role', true)::text = 'project_admin'
            AND
            current_setting('app.current_project', true)::uuid = project_id
        )
    )
    WITH CHECK (
        current_setting('app.current_owner', true)::uuid = admin_id
        OR
        (
            current_setting('app.organization_role', true)::text = 'project_admin'
            AND
            current_setting('app.current_project', true)::uuid = project_id
        )
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS organizations_update_policy ON "auth".organizations;

--changeset solomon.auth-rls:5 labels:auth context:auth
--comment: Create INSERT policy for auth.organizations (current owner = admin_id)
DO $$ BEGIN
    CREATE POLICY organizations_insert_policy ON "auth".organizations
    FOR INSERT
    WITH CHECK (
        current_setting('app.current_owner', true)::uuid = admin_id
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS organizations_insert_policy ON "auth".organizations;

--changeset solomon.auth-rls:6 labels:auth context:auth
--comment: Create DELETE policy for auth.organizations (requires project_admin and matching project)
DO $$ BEGIN
    CREATE POLICY organizations_delete_policy ON "auth".organizations
    FOR DELETE
    USING (
        current_setting('app.organization_role', true)::text = 'project_admin'
        AND current_setting('app.current_project', true)::uuid = project_id
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS organizations_delete_policy ON "auth".organizations;

--changeset solomon.auth-rls:7 labels:auth context:auth
--comment: create rls policy on organization_usage_summary table
ALTER TABLE "auth".organization_usage_summary ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
    CREATE POLICY organization_usage_summary_policy ON "auth".organization_usage_summary
    FOR ALL
    USING (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key'))
    )
    WITH CHECK (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key'))
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS organization_usage_summary_policy ON "auth".organization_usage_summary;

--changeset solomon.auth-rls:8 labels:auth context:auth
--comment: create rls policy on user_usage_summary table
ALTER TABLE "auth".user_usage_summary ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
    CREATE POLICY user_usage_summary_policy ON "auth".user_usage_summary
    FOR ALL
    USING (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key')) OR
        (user_id = current_setting('app.current_owner', true)::uuid)
    )
    WITH CHECK (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key')) OR
        (user_id = current_setting('app.current_owner', true)::uuid)
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS user_usage_summary_policy ON "auth".user_usage_summary;

--changeset solomon.auth-rls:9 labels:auth context:auth
--comment: create rls policy on api_key_usage_summary table
ALTER TABLE "auth".api_key_usage_summary ENABLE ROW LEVEL SECURITY;
DO $$ BEGIN
    CREATE POLICY api_key_usage_summary_policy ON "auth".api_key_usage_summary
    FOR ALL
    USING (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key'))
    )
    WITH CHECK (
        (organization_id = current_setting('app.current_organization', true)::uuid AND
        current_setting('app.organization_role', true)::text IN ('admin', 'api_key'))
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
--rollback DROP POLICY IF EXISTS api_key_usage_summary_policy ON "auth".api_key_usage_summary;
