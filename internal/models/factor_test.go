package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"auth/internal/conf"
	"auth/internal/storage"
	"auth/internal/storage/test"
)

type FactorTestSuite struct {
	suite.Suite
	db         *storage.Connection
	TestFactor *Factor
}

func TestFactor(t *testing.T) {
	globalConfig, err := conf.LoadGlobal(modelsTestConfig)
	require.NoError(t, err)
	conn, err := test.SetupDBConnection(globalConfig)
	require.NoError(t, err)
	ts := &FactorTestSuite{
		db: conn,
	}
	defer ts.db.Close()
	suite.Run(t, ts)
}

func (ts *FactorTestSuite) SetupTest() {
	TruncateAll(ts.db)
	project_id := uuid.Must(uuid.NewV4())
	// Create a project
	if err := ts.db.RawQuery(fmt.Sprintf("INSERT INTO auth.projects (id, name) VALUES ('%s', 'test_project')", project_id)).Exec(); err != nil {
		panic(err)
	}

	// Create the admin of the organization
	user, err := NewUser("", "admin@example.com", "test", "", nil, uuid.Nil, project_id)
	require.NoError(ts.T(), err, "Error making new user")
	require.NoError(ts.T(), ts.db.Create(user, "organization_id", "organization_role"), "Error creating user")

	// Create the organization
	organization_id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	if err := ts.db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization', '%s', '%s')", organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := ts.db.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	user, err = NewUser("", "agenericemail@gmail.com", "secret", "test", nil, organization_id, uuid.Nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), ts.db.Create(user, "project_id", "organization_role"))

	factor := NewTOTPFactor(user, "asimplename")
	require.NoError(ts.T(), factor.SetSecret("topsecret", false, "", ""))
	require.NoError(ts.T(), ts.db.Create(factor))
	ts.TestFactor = factor
}

func (ts *FactorTestSuite) TestFindFactorByFactorID() {
	n, err := FindFactorByFactorID(ts.db, ts.TestFactor.ID)
	require.NoError(ts.T(), err)
	require.Equal(ts.T(), ts.TestFactor.ID, n.ID)

	_, err = FindFactorByFactorID(ts.db, uuid.Nil)
	require.EqualError(ts.T(), err, FactorNotFoundError{}.Error())
}

func (ts *FactorTestSuite) TestUpdateStatus() {
	newFactorStatus := FactorStateVerified
	require.NoError(ts.T(), ts.TestFactor.UpdateStatus(ts.db, newFactorStatus))
	require.Equal(ts.T(), newFactorStatus.String(), ts.TestFactor.Status)
}

func (ts *FactorTestSuite) TestUpdateFriendlyName() {
	newName := "newfactorname"
	require.NoError(ts.T(), ts.TestFactor.UpdateFriendlyName(ts.db, newName))
	require.Equal(ts.T(), newName, ts.TestFactor.FriendlyName)
}

func (ts *FactorTestSuite) TestEncodedFactorDoesNotLeakSecret() {
	encodedFactor, err := json.Marshal(ts.TestFactor)
	require.NoError(ts.T(), err)

	decodedFactor := Factor{}
	json.Unmarshal(encodedFactor, &decodedFactor)
	require.Equal(ts.T(), decodedFactor.Secret, "")
}
