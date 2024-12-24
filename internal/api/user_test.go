package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth/internal/conf"
	"auth/internal/crypto"
	"auth/internal/models"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	API    *API
	Config *conf.GlobalConfiguration
}

func TestUser(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)

	ts := &UserTestSuite{
		API:    api,
		Config: config,
	}
	defer api.db.Close()

	suite.Run(t, ts)
}

func (ts *UserTestSuite) SetupTest() {
	models.TruncateAll(ts.API.db)

	project_id := uuid.Must(uuid.NewV4())
	// Create a project
	if err := ts.API.db.RawQuery(fmt.Sprintf("INSERT INTO auth.projects (id, name) VALUES ('%s', 'test_project')", project_id)).Exec(); err != nil {
		panic(err)
	}

	// Create the admin of the organization
	user, err := models.NewUser("", "admin@example.com", "test", ts.Config.JWT.Aud, nil, uuid.Nil, project_id)
	require.NoError(ts.T(), err, "Error making new user")
	require.NoError(ts.T(), ts.API.db.Create(user, "organization_id", "organization_role"), "Error creating user")

	// Create the organization
	organization_id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	if err := ts.API.db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization', '%s', '%s')", organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := ts.API.db.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Create user
	u, err := models.NewUser("123456789", "test@example.com", "password", ts.Config.JWT.Aud, nil, organization_id, uuid.Nil)
	require.NoError(ts.T(), err, "Error creating test user model")
	require.NoError(ts.T(), ts.API.db.Create(u, "project_id", "organization_role"), "Error saving new test user")
}

func (ts *UserTestSuite) generateToken(user *models.User, sessionId *uuid.UUID) string {
	req := httptest.NewRequest(http.MethodPost, "/token?grant_type=password", nil)
	token, _, err := ts.API.generateAccessToken(req, ts.API.db, user, sessionId, models.PasswordGrant)
	require.NoError(ts.T(), err, "Error generating access token")
	return token
}

func (ts *UserTestSuite) generateAccessTokenAndSession(user *models.User) string {
	session, err := models.NewSession(user.ID, nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), ts.API.db.Create(session))

	req := httptest.NewRequest(http.MethodPost, "/token?grant_type=password", nil)
	token, _, err := ts.API.generateAccessToken(req, ts.API.db, user, &session.ID, models.PasswordGrant)
	require.NoError(ts.T(), err, "Error generating access token")
	return token
}

func (ts *UserTestSuite) TestUserGet() {
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err, "Error finding user")
	token := ts.generateAccessTokenAndSession(u)

	require.NoError(ts.T(), err, "Error generating access token")

	req := httptest.NewRequest(http.MethodGet, "http://localhost/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), http.StatusOK, w.Code)
}

func (ts *UserTestSuite) TestUserUpdateEmail() {
	cases := []struct {
		desc                       string
		userData                   map[string]interface{}
		isSecureEmailChangeEnabled bool
		isMailerAutoconfirmEnabled bool
		expectedCode               int
	}{
		{
			desc: "User doesn't have an existing email",
			userData: map[string]interface{}{
				"email":           "",
				"phone":           "",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
			},
			isSecureEmailChangeEnabled: false,
			isMailerAutoconfirmEnabled: false,
			expectedCode:               http.StatusOK,
		},
		{
			desc: "User doesn't have an existing email and double email confirmation required",
			userData: map[string]interface{}{
				"email":           "",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
				"phone":           "234567890",
			},
			isSecureEmailChangeEnabled: true,
			isMailerAutoconfirmEnabled: false,
			expectedCode:               http.StatusOK,
		},
		{
			desc: "User has an existing email",
			userData: map[string]interface{}{
				"email":           "foo@example.com",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
				"phone":           "",
			},
			isSecureEmailChangeEnabled: false,
			isMailerAutoconfirmEnabled: false,
			expectedCode:               http.StatusOK,
		},
		{
			desc: "User has an existing email and double email confirmation required",
			userData: map[string]interface{}{
				"email":           "bar@example.com",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
				"phone":           "",
			},
			isSecureEmailChangeEnabled: true,
			isMailerAutoconfirmEnabled: false,
			expectedCode:               http.StatusOK,
		},
		{
			desc: "Update email with mailer autoconfirm enabled",
			userData: map[string]interface{}{
				"email":           "bar@example.com",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
				"phone":           "",
			},
			isSecureEmailChangeEnabled: true,
			isMailerAutoconfirmEnabled: true,
			expectedCode:               http.StatusOK,
		},
		{
			desc: "Update email with mailer autoconfirm enabled and anonymous user",
			userData: map[string]interface{}{
				"email":           "bar@example.com",
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
				"phone":           "",
				"is_anonymous":    true,
			},
			isSecureEmailChangeEnabled: true,
			isMailerAutoconfirmEnabled: true,
			expectedCode:               http.StatusOK,
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
			u, err := models.NewUser("", "", "", ts.Config.JWT.Aud, nil, id, uuid.Nil)
			require.NoError(ts.T(), err, "Error creating test user model")
			require.NoError(ts.T(), u.SetEmail(ts.API.db, c.userData["email"].(string)), "Error setting user email")
			require.NoError(ts.T(), u.SetPhone(ts.API.db, c.userData["phone"].(string)), "Error setting user phone")
			if isAnonymous, ok := c.userData["is_anonymous"]; ok {
				u.IsAnonymous = isAnonymous.(bool)
			}
			require.NoError(ts.T(), ts.API.db.Create(u, "project_id", "organization_role"), "Error saving test user")

			token := ts.generateAccessTokenAndSession(u)

			require.NoError(ts.T(), err, "Error generating access token")

			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
				"email":           "new@example.com",
				"organization_id": c.userData["organization_id"],
			}))
			req := httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			ts.Config.Mailer.SecureEmailChangeEnabled = c.isSecureEmailChangeEnabled
			ts.Config.Mailer.Autoconfirm = c.isMailerAutoconfirmEnabled
			ts.API.handler.ServeHTTP(w, req)
			require.Equal(ts.T(), c.expectedCode, w.Code)

			var data models.User
			require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))

			if c.isMailerAutoconfirmEnabled && u.IsAnonymous {
				require.Empty(ts.T(), data.EmailChange)
				require.Equal(ts.T(), "new@example.com", data.GetEmail())
				require.Len(ts.T(), data.Identities, 1)
			} else {
				require.Equal(ts.T(), "new@example.com", data.EmailChange)
				require.Len(ts.T(), data.Identities, 0)
			}

			// remove user after each case
			require.NoError(ts.T(), ts.API.db.Destroy(u))
		})
	}

}
func (ts *UserTestSuite) TestUserUpdatePhoneAutoconfirmEnabled() {
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	existingUser, err := models.NewUser("22222222", "", "", ts.Config.JWT.Aud, nil, id, uuid.Nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), ts.API.db.Create(existingUser, "project_id", "organization_role"), "Error saving new test user")

	cases := []struct {
		desc         string
		userData     map[string]string
		expectedCode int
	}{
		{
			desc: "New phone number is the same as current phone number",
			userData: map[string]string{
				"phone": "123456789",
			},
			expectedCode: http.StatusOK,
		},
		{
			desc: "New phone number exists already",
			userData: map[string]string{
				"phone": "22222222",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			desc: "New phone number is different from current phone number",
			userData: map[string]string{
				"phone": "234567890",
			},
			expectedCode: http.StatusOK,
		},
	}

	ts.Config.Sms.Autoconfirm = true

	for _, c := range cases {
		ts.Run(c.desc, func() {
			token := ts.generateAccessTokenAndSession(u)
			require.NoError(ts.T(), err, "Error generating access token")

			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
				"phone":           c.userData["phone"],
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
			}))
			req := httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			ts.API.handler.ServeHTTP(w, req)
			require.Equal(ts.T(), c.expectedCode, w.Code)

			if c.expectedCode == http.StatusOK {
				// check that the user response returned contains the updated phone field
				data := &models.User{}
				require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))
				require.Equal(ts.T(), data.GetPhone(), c.userData["phone"])
			}
		})
	}

}

func (ts *UserTestSuite) TestUserUpdatePassword() {
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	r, err := models.GrantAuthenticatedUser(ts.API.db, u, models.GrantParams{})
	require.NoError(ts.T(), err)

	r2, err := models.GrantAuthenticatedUser(ts.API.db, u, models.GrantParams{})
	require.NoError(ts.T(), err)

	// create a session and modify it's created_at time to simulate a session that is not recently logged in
	notRecentlyLoggedIn, err := models.FindSessionByID(ts.API.db, *r2.SessionId, true)
	require.NoError(ts.T(), err)

	// cannot use Update here because Update doesn't removes the created_at field
	require.NoError(ts.T(), ts.API.db.RawQuery(
		"update "+notRecentlyLoggedIn.TableName()+" set created_at = ? where id = ?",
		time.Now().Add(-24*time.Hour),
		notRecentlyLoggedIn.ID).Exec(),
	)

	type expected struct {
		code            int
		isAuthenticated bool
	}

	var cases = []struct {
		desc                    string
		newPassword             string
		nonce                   string
		requireReauthentication bool
		sessionId               *uuid.UUID
		expected                expected
	}{
		{
			desc:                    "Need reauthentication because outside of recently logged in window",
			newPassword:             "newpassword123",
			nonce:                   "",
			requireReauthentication: true,
			sessionId:               &notRecentlyLoggedIn.ID,
			expected:                expected{code: http.StatusBadRequest, isAuthenticated: false},
		},
		{
			desc:                    "No nonce provided",
			newPassword:             "newpassword123",
			nonce:                   "",
			sessionId:               &notRecentlyLoggedIn.ID,
			requireReauthentication: true,
			expected:                expected{code: http.StatusBadRequest, isAuthenticated: false},
		},
		{
			desc:                    "Invalid nonce",
			newPassword:             "newpassword1234",
			nonce:                   "123456",
			sessionId:               &notRecentlyLoggedIn.ID,
			requireReauthentication: true,
			expected:                expected{code: http.StatusUnprocessableEntity, isAuthenticated: false},
		},
		{
			desc:                    "No need reauthentication because recently logged in",
			newPassword:             "newpassword123",
			nonce:                   "",
			requireReauthentication: true,
			sessionId:               r.SessionId,
			expected:                expected{code: http.StatusOK, isAuthenticated: true},
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			ts.Config.Security.UpdatePasswordRequireReauthentication = c.requireReauthentication
			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]string{
				"password":        c.newPassword,
				"nonce":           c.nonce,
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
			}))

			req := httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
			req.Header.Set("Content-Type", "application/json")
			token := ts.generateToken(u, c.sessionId)

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			// Setup response recorder
			w := httptest.NewRecorder()
			ts.API.handler.ServeHTTP(w, req)
			require.Equal(ts.T(), c.expected.code, w.Code)

			// Request body
			u, err = models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
			require.NoError(ts.T(), err)

			isAuthenticated, _, err := u.Authenticate(context.Background(), ts.API.db, c.newPassword, ts.API.config.Security.DBEncryption.DecryptionKeys, ts.API.config.Security.DBEncryption.Encrypt, ts.API.config.Security.DBEncryption.EncryptionKeyID)
			require.NoError(ts.T(), err)

			require.Equal(ts.T(), c.expected.isAuthenticated, isAuthenticated)
		})
	}
}

func (ts *UserTestSuite) TestUserUpdatePasswordNoReauthenticationRequired() {
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	type expected struct {
		code            int
		isAuthenticated bool
	}

	var cases = []struct {
		desc                    string
		newPassword             string
		nonce                   string
		requireReauthentication bool
		expected                expected
	}{
		{
			desc:                    "Invalid password length",
			newPassword:             "",
			nonce:                   "",
			requireReauthentication: false,
			expected:                expected{code: http.StatusUnprocessableEntity, isAuthenticated: false},
		},

		{
			desc:                    "Valid password length",
			newPassword:             "newpassword",
			nonce:                   "",
			requireReauthentication: false,
			expected:                expected{code: http.StatusOK, isAuthenticated: true},
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			ts.Config.Security.UpdatePasswordRequireReauthentication = c.requireReauthentication
			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]string{
				"password":        c.newPassword,
				"nonce":           c.nonce,
				"organization_id": "123e4567-e89b-12d3-a456-426655440000",
			}))

			req := httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
			req.Header.Set("Content-Type", "application/json")
			token := ts.generateAccessTokenAndSession(u)

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			// Setup response recorder
			w := httptest.NewRecorder()
			ts.API.handler.ServeHTTP(w, req)
			require.Equal(ts.T(), c.expected.code, w.Code)

			// Request body
			u, err = models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
			require.NoError(ts.T(), err)

			isAuthenticated, _, err := u.Authenticate(context.Background(), ts.API.db, c.newPassword, ts.API.config.Security.DBEncryption.DecryptionKeys, ts.API.config.Security.DBEncryption.Encrypt, ts.API.config.Security.DBEncryption.EncryptionKeyID)
			require.NoError(ts.T(), err)

			require.Equal(ts.T(), c.expected.isAuthenticated, isAuthenticated)
		})
	}
}

func (ts *UserTestSuite) TestUserUpdatePasswordReauthentication() {
	ts.Config.Security.UpdatePasswordRequireReauthentication = true
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	// Confirm the test user
	now := time.Now()
	u.EmailConfirmedAt = &now
	require.NoError(ts.T(), ts.API.db.Update(u), "Error updating new test user")

	token := ts.generateAccessTokenAndSession(u)

	// request for reauthentication nonce
	req := httptest.NewRequest(http.MethodGet, "http://localhost/reauthenticate", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), w.Code, http.StatusOK)

	u, err = models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)
	require.NotEmpty(ts.T(), u.ReauthenticationToken)
	require.NotEmpty(ts.T(), u.ReauthenticationSentAt)

	// update reauthentication token to a known token
	u.ReauthenticationToken = crypto.GenerateTokenHash(u.GetEmail(), "123456")
	require.NoError(ts.T(), ts.API.db.Update(u))

	// update password with reauthentication token
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"password":        "newpass",
		"nonce":           "123456",
		"organization_id": "123e4567-e89b-12d3-a456-426655440000",
	}))

	req = httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), w.Code, http.StatusOK)

	// Request body
	u, err = models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	isAuthenticated, _, err := u.Authenticate(context.Background(), ts.API.db, "newpass", ts.Config.Security.DBEncryption.DecryptionKeys, ts.Config.Security.DBEncryption.Encrypt, ts.Config.Security.DBEncryption.EncryptionKeyID)
	require.NoError(ts.T(), err)

	require.True(ts.T(), isAuthenticated)
	require.Empty(ts.T(), u.ReauthenticationToken)
	require.Nil(ts.T(), u.ReauthenticationSentAt)
}

func (ts *UserTestSuite) TestUserUpdatePasswordLogoutOtherSessions() {
	ts.Config.Security.UpdatePasswordRequireReauthentication = false
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	u, err := models.FindUserByEmailAndAudience(ts.API.db, "test@example.com", ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)

	// Confirm the test user
	now := time.Now()
	u.EmailConfirmedAt = &now
	require.NoError(ts.T(), ts.API.db.Update(u), "Error updating new test user")

	// Login the test user to get first session
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":           u.GetEmail(),
		"password":        "password",
		"organization_id": "123e4567-e89b-12d3-a456-426655440000",
	}))
	req := httptest.NewRequest(http.MethodPost, "http://localhost/token?grant_type=password", &buffer)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), http.StatusOK, w.Code)

	session1 := AccessTokenResponse{}
	require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&session1))

	// Login test user to get second session
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":           u.GetEmail(),
		"password":        "password",
		"organization_id": "123e4567-e89b-12d3-a456-426655440000",
	}))
	req = httptest.NewRequest(http.MethodPost, "http://localhost/token?grant_type=password", &buffer)
	req.Header.Set("Content-Type", "application/json")

	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), http.StatusOK, w.Code)

	session2 := AccessTokenResponse{}
	require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&session2))

	// Update user's password using first session
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"password":        "newpass",
		"organization_id": "123e4567-e89b-12d3-a456-426655440000",
	}))

	req = httptest.NewRequest(http.MethodPut, "http://localhost/user", &buffer)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session1.Token))

	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), http.StatusOK, w.Code)

	// Attempt to refresh session1 should pass
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"refresh_token": session1.RefreshToken,
	}))

	req = httptest.NewRequest(http.MethodPost, "http://localhost/token?grant_type=refresh_token", &buffer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.Equal(ts.T(), http.StatusOK, w.Code)

	// Attempt to refresh session2 should fail
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"refresh_token": session2.RefreshToken,
	}))

	req = httptest.NewRequest(http.MethodPost, "http://localhost/token?grant_type=refresh_token", &buffer)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	require.NotEqual(ts.T(), http.StatusOK, w.Code)
}
