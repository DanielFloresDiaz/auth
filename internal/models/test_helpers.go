package models

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/storage"
)

// InitializeTestDatabase sets up the database with a project, organization, and admin user.
// It returns the project ID, organization ID, and admin user.
func InitializeTestDatabase(t *testing.T, db *storage.Connection, config *conf.GlobalConfiguration) (uuid.UUID, uuid.UUID, *User) {
	// Connect as the postgres superuser to create project, user and organization
	superuserURL := config.DB.URL
	if u, err := url.Parse(config.DB.URL); err == nil {
		u.User = url.UserPassword("postgres", "root")
		superuserURL = u.String()
	}

	superuserDeets := &pop.ConnectionDetails{
		Dialect: config.DB.Driver,
		URL:     superuserURL,
	}

	superuserDB, err := pop.NewConnection(superuserDeets)
	require.NoError(t, err, "Should be able to connect as postgres superuser")
	require.NoError(t, superuserDB.Open())
	defer superuserDB.Close()

	setup_db := &storage.Connection{Connection: superuserDB}

	require.NoError(t, TruncateAll(setup_db))

	project_id := uuid.Must(uuid.NewV4())
	// Create a project with a unique name using the UUID
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.projects (id, name) VALUES ('%s', 'test_project_%s')", project_id, project_id)).Exec(); err != nil {
		panic(err)
	}

	// Create the admin of the organization
	user, err := NewUser("", "admin@example.com", "test", config.JWT.Aud, nil, uuid.Nil, project_id)
	require.NoError(t, err, "Error making new user")
	require.NoError(t, db.Create(user, "organization_id", "organization_role"), "Error creating user")

	// Create the organization if it doesn't exist
	organization_id := uuid.Must(uuid.NewV4())
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization_%s', '%s', '%s')", organization_id, organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Insert organization tier
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations_config (organization_id, project_id, tier) VALUES ('%s', '%s', 'free')", organization_id, project_id)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := setup_db.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	return project_id, organization_id, user
}

// ReseedTestTenant recreates the project and organization after TruncateAll.
// TruncateAll removes tenant rows, and later inserts still reference the ids
// the test suite captured during setup.
func ReseedTestTenant(t *testing.T, db *storage.Connection, organizationID, projectID uuid.UUID) {
	t.Helper()

	superuserURL := db.URL()
	if u, err := url.Parse(db.URL()); err == nil {
		u.User = url.UserPassword("postgres", "root")
		superuserURL = u.String()
	}
	superuserDB, err := pop.NewConnection(&pop.ConnectionDetails{
		Dialect: "postgres",
		URL:     superuserURL,
	})
	require.NoError(t, err)
	require.NoError(t, superuserDB.Open())
	defer superuserDB.Close()
	setupDB := &storage.Connection{Connection: superuserDB}

	require.NoError(t, setupDB.RawQuery(
		"INSERT INTO auth.projects (id, name) VALUES (?, ?) ON CONFLICT (id) DO NOTHING",
		projectID, "test_project_"+projectID.String(),
	).Exec())

	admin, err := NewUser("", "reseed-admin+"+projectID.String()+"@example.com", "test", "authenticated", nil, uuid.Nil, projectID)
	require.NoError(t, err)
	require.NoError(t, db.Create(admin, "organization_id", "organization_role"))

	require.NoError(t, setupDB.RawQuery(
		"INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES (?, ?, ?, ?) ON CONFLICT (id) DO NOTHING",
		organizationID, "test_organization_"+organizationID.String(), projectID, admin.ID,
	).Exec())
	require.NoError(t, setupDB.RawQuery(
		"INSERT INTO auth.organizations_config (organization_id, project_id, tier) VALUES (?, ?, 'free') ON CONFLICT (organization_id, project_id) DO NOTHING",
		organizationID, projectID,
	).Exec())
}
