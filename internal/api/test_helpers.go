package api

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/models"
	"github.com/supabase/auth/internal/storage"
)

// InitializeTestDatabase sets up the database with a project, organization, and admin user.
// It returns the project ID, organization ID, and admin user.
func InitializeTestDatabase(t *testing.T, api *API, config *conf.GlobalConfiguration) (uuid.UUID, uuid.UUID, *models.User) {
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

	require.NoError(t, models.TruncateAll(setup_db))

	project_id := uuid.Must(uuid.NewV4())
	// Create a project with a unique name using the UUID
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.projects (id, name) VALUES ('%s', 'test_project_%s')", project_id, project_id)).Exec(); err != nil {
		panic(err)
	}

	// Create the admin of the organization
	user, err := models.NewUser("", "admin@example.com", "test", config.JWT.Aud, nil, uuid.Nil, project_id)
	require.NoError(t, err, "Error making new user")
	require.NoError(t, api.db.Create(user, "organization_id", "organization_role"), "Error creating user")

	// Create the organization if it doesn't exist
	organization_id := uuid.Must(uuid.NewV4())
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization_%s', '%s', '%s')", organization_id, organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Insert organization tier
	if err := setup_db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations_tier (organization_id, tier) VALUES ('%s', 'free')", organization_id)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := setup_db.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	return project_id, organization_id, user
}
