package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"context"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/crypto"

	"github.com/supabase/auth/internal/mailer/templatemailer"
	"github.com/supabase/auth/internal/models"
)

type InviteTestSuite struct {
	suite.Suite
	API    *API
	Config *conf.GlobalConfiguration

	token          string
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
}

func TestInvite(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)

	ts := &InviteTestSuite{
		API:    api,
		Config: config,
	}
	defer api.db.Close()

	suite.Run(t, ts)
}

func (ts *InviteTestSuite) SetupTest() {
	// Invite mail tests disable email signup. Password invite verification still requires it.
	ts.Config.External.Email.Enabled = true

	// Initialize the database with project, organization, and admin user
	ts.ProjectID, ts.OrganizationID, _ = InitializeTestDatabase(ts.T(), ts.API, ts.Config)

	// Setup response recorder with super admin privileges
	ts.token = ts.makeSuperAdmin("")
}

func (ts *InviteTestSuite) makeSuperAdmin(email string) string {
	// Cleanup existing user, if they already exist
	if u, _ := models.FindUserByEmailAndAudience(ts.API.db, email, ts.Config.JWT.Aud, ts.OrganizationID, uuid.Nil); u != nil {
		require.NoError(ts.T(), ts.API.db.Destroy(u), "Error deleting user")
	}

	u, err := models.NewUser("123456789", email, "test", ts.Config.JWT.Aud, map[string]interface{}{"full_name": "Test User"}, ts.OrganizationID, ts.ProjectID)
	require.NoError(ts.T(), err, "Error making new user")
	require.NoError(ts.T(), ts.API.db.Create(u, "organization_role"))

	u.Role = "supabase_admin"

	var token string

	session, err := models.NewSession(u.ID, nil)
	require.NoError(ts.T(), err)
	require.NoError(ts.T(), ts.API.db.Create(session))

	req := httptest.NewRequest(http.MethodPost, "/invite", nil)
	token, _, err = ts.API.generateAccessToken(req, ts.API.db, u, &session.ID, models.Invite)

	require.NoError(ts.T(), err, "Error generating access token")

	p := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	_, err = p.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(ts.Config.JWT.Secret), nil
	})
	require.NoError(ts.T(), err, "Error parsing token")

	return token
}

func (ts *InviteTestSuite) TestInvite() {
	// Request body
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email": "test@example.com",
		"data": map[string]interface{}{
			"a": 1,
		},
		"organization_id": ts.OrganizationID.String(),
		"project_id":      ts.ProjectID.String(),
	}))

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

	// Setup response recorder
	w := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)
	assert.Equal(ts.T(), http.StatusOK, w.Code)
}

func (ts *InviteTestSuite) TestInviteWithProjectIDOnly() {
	email := "project-only-invite@example.com"
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":      email,
		"project_id": ts.ProjectID.String(),
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	assert.Equal(ts.T(), http.StatusOK, w.Code)

	user, err := models.FindUserByEmailAndAudience(ts.API.db, email, ts.Config.JWT.Aud, uuid.Nil, ts.ProjectID)
	require.NoError(ts.T(), err)
	assert.Equal(ts.T(), ts.ProjectID, user.ProjectID)
	assert.False(ts.T(), user.OrganizationID.Valid)
	assert.NotEmpty(ts.T(), user.ConfirmationToken)

	var body struct {
		InviteToken string `json:"invite_token"`
	}
	require.NoError(ts.T(), json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(ts.T(), user.ConfirmationToken, body.InviteToken)
}

func (ts *InviteTestSuite) TestInviteExists() {
	// To allow us to send signup and invite request in succession
	ts.Config.SMTP.MaxFrequency = 200

	email := uuid.Must(uuid.NewV4()).String() + "@example.com"

	{
		// Request body
		var buffer bytes.Buffer
		require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
			"email": email,
			"data": map[string]interface{}{
				"a": 1,
			},
			"organization_id": ts.OrganizationID.String(),
			"project_id":      ts.ProjectID.String(),
		}))

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

		// Setup response recorder
		w := httptest.NewRecorder()

		ts.API.handler.ServeHTTP(w, req)
		assert.Equal(ts.T(), http.StatusOK, w.Code)
	}

	{
		// Request body
		var buffer bytes.Buffer
		require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
			"email": email,
			"data": map[string]interface{}{
				"a": 1,
			},
			"organization_id": ts.OrganizationID.String(),
			"project_id":      ts.ProjectID.String(),
		}))

		// Setup request
		req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

		// Setup response recorder
		w := httptest.NewRecorder()

		ts.API.handler.ServeHTTP(w, req)
		assert.Equal(ts.T(), http.StatusOK, w.Code)
	}
}

func (ts *InviteTestSuite) TestInviteAfterSignupShouldNotReturnSensitiveFields() {
	// To allow us to send signup and invite request in succession
	ts.Config.SMTP.MaxFrequency = 5
	// Request body
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email": "test@example.com",
		"data": map[string]interface{}{
			"a": 1,
		},
		"organization_id": ts.OrganizationID.String(),
		"project_id":      ts.ProjectID.String(),
	}))

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

	// Setup response recorder
	w := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)
	assert.Equal(ts.T(), http.StatusOK, w.Code)

	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":    "test@example.com",
		"password": "test123",
		"data": map[string]interface{}{
			"a": 1,
		},
		"organization_id": ts.OrganizationID.String(),
		"project_id":      ts.ProjectID.String(),
	}))

	// Setup request
	req = httptest.NewRequest(http.MethodPost, "/signup", &buffer)
	req.Header.Set("Content-Type", "application/json")

	// Setup response recorder
	x := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(x, req)

	require.Equal(ts.T(), http.StatusOK, x.Code)

	data := models.User{}
	require.NoError(ts.T(), json.NewDecoder(x.Body).Decode(&data))
	// Sensitive fields
	require.Equal(ts.T(), 0, len(data.Identities))
	require.Equal(ts.T(), 0, len(data.UserMetaData))
}

func (ts *InviteTestSuite) TestInvite_WithoutAccess() {
	// Request body
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email": "test@example.com",
		"data": map[string]interface{}{
			"a": 1,
		},
		"organization_id": ts.OrganizationID.String(),
		"project_id":      ts.ProjectID.String(),
	}))

	// Setup request
	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")

	// Setup response recorder
	w := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)
	assert.Equal(ts.T(), http.StatusUnauthorized, w.Code) // 401 OK because the invite request above has no Authorization header
}

func (ts *InviteTestSuite) TestVerifyInvite() {
	cases := []struct {
		desc        string
		email       string
		requestBody map[string]interface{}
		expected    int
	}{
		{
			"Verify invite with password",
			"test@example.com",
			map[string]interface{}{
				"email":           "test@example.com",
				"type":            "invite",
				"token":           "asdf",
				"password":        "testing",
				"organization_id": ts.OrganizationID.String(),
				"project_id":      ts.ProjectID.String(),
			},
			http.StatusOK,
		},
		{
			"Verify invite with no password",
			"test1@example.com",
			map[string]interface{}{
				"email":           "test1@example.com",
				"type":            "invite",
				"token":           "asdf",
				"organization_id": ts.OrganizationID.String(),
				"project_id":      ts.ProjectID.String(),
			},
			http.StatusOK,
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			user, err := models.NewUser("", c.email, "", ts.Config.JWT.Aud, nil, ts.OrganizationID, ts.ProjectID)
			now := time.Now()
			user.InvitedAt = &now
			user.ConfirmationSentAt = &now
			user.EncryptedPassword = nil
			user.ConfirmationToken = crypto.GenerateTokenHash(c.email, c.requestBody["token"].(string))
			require.NoError(ts.T(), err)
			require.NoError(ts.T(), ts.API.db.Create(user, "organization_role"))
			require.NoError(ts.T(), models.CreateOneTimeToken(ts.API.db, user.ID, user.GetEmail(), user.ConfirmationToken, models.ConfirmationToken))

			// Find test user
			_, err = models.FindUserByEmailAndAudience(ts.API.db, c.email, ts.Config.JWT.Aud, ts.OrganizationID, uuid.Nil)
			require.NoError(ts.T(), err)

			// Request body
			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(c.requestBody))

			// Setup request
			req := httptest.NewRequest(http.MethodPost, "http://localhost/verify", &buffer)
			req.Header.Set("Content-Type", "application/json")

			// Setup response recorder
			w := httptest.NewRecorder()

			ts.API.handler.ServeHTTP(w, req)

			assert.Equal(ts.T(), c.expected, w.Code, w.Body.String())
		})
	}
}

func (ts *InviteTestSuite) TestInviteExternalGitlab() {
	tokenCount, userCount := 0, 0
	code := "authcode"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			tokenCount++
			ts.Equal(code, r.FormValue("code"))
			ts.Equal("authorization_code", r.FormValue("grant_type"))
			ts.Equal(ts.Config.External.Gitlab.RedirectURI, r.FormValue("redirect_uri"))

			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"gitlab_token","expires_in":100000}`)
		case "/api/v4/user":
			userCount++
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"Gitlab Test","email":"gitlab@example.com","avatar_url":"http://example.com/avatar","confirmed_at": "2020-01-01T00:00:00.000Z"}`)
		case "/api/v4/user/emails":
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `[]`)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			ts.Fail("unknown gitlab oauth call %s", r.URL.Path)
		}
	}))
	defer server.Close()
	ts.Config.External.Gitlab.URL = server.URL

	// invite user
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(InviteParams{
		Email:          "gitlab@example.com",
		OrganizationID: ts.OrganizationID,
		ProjectID:      ts.ProjectID,
	}))
	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusOK, w.Code)

	// Find test user
	user, err := models.FindUserByEmailAndAudience(ts.API.db, "gitlab@example.com", ts.Config.JWT.Aud, ts.OrganizationID, uuid.Nil)
	require.NoError(ts.T(), err)

	// get redirect url w/ state
	provider := "gitlab"
	url_path := fmt.Sprintf("http://localhost/authorize?provider=%s&invite_token=%s", provider, user.ConfirmationToken)
	req = httptest.NewRequest(http.MethodGet, url_path, nil)
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusFound, w.Code)
	u, err := url.Parse(w.Header().Get("Location"))
	ts.Require().NoError(err, "redirect url parse failed")
	q := u.Query()
	state := q.Get("state")

	// auth server callback
	testURL, err := url.Parse("http://localhost/callback")
	ts.Require().NoError(err)
	v := testURL.Query()
	v.Set("code", code)
	v.Set("state", state)
	testURL.RawQuery = v.Encode()
	req = httptest.NewRequest(http.MethodGet, testURL.String(), nil)
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusFound, w.Code)
	u, err = url.Parse(w.Header().Get("Location"))
	ts.Require().NoError(err, "redirect url parse failed")

	// ensure redirect has #access_token=...
	v, err = url.ParseQuery(u.Fragment)
	ts.Require().NoError(err)
	ts.Require().Empty(v.Get("error_description"))
	ts.Require().Empty(v.Get("error"))

	ts.NotEmpty(v.Get("access_token"))
	ts.NotEmpty(v.Get("refresh_token"))
	ts.NotEmpty(v.Get("expires_in"))
	ts.Equal("bearer", v.Get("token_type"))

	ts.Equal(1, tokenCount)
	ts.Equal(1, userCount)

	// ensure user has been created with metadata
	user, err = models.FindUserByEmailAndAudience(ts.API.db, "gitlab@example.com", ts.Config.JWT.Aud, ts.OrganizationID, uuid.Nil)
	ts.Require().NoError(err)
	ts.Equal("Gitlab Test", user.UserMetaData["full_name"])
	ts.Equal("http://example.com/avatar", user.UserMetaData["avatar_url"])
	ts.Equal("gitlab", user.AppMetaData["provider"])
	ts.Equal([]interface{}{"gitlab"}, user.AppMetaData["providers"])
}

func (ts *InviteTestSuite) TestInviteExternalGitlab_MismatchedEmails() {
	tokenCount, userCount := 0, 0
	code := "authcode"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			tokenCount++
			ts.Equal(code, r.FormValue("code"))
			ts.Equal("authorization_code", r.FormValue("grant_type"))
			ts.Equal(ts.Config.External.Gitlab.RedirectURI, r.FormValue("redirect_uri"))

			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"gitlab_token","expires_in":100000}`)
		case "/api/v4/user":
			userCount++
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"Gitlab Test","email":"gitlab+mismatch@example.com","avatar_url":"http://example.com/avatar","confirmed_at": "2020-01-01T00:00:00.000Z"}`)
		case "/api/v4/user/emails":
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprint(w, `[]`)
		default:
			w.WriteHeader(500)
			ts.Fail("unknown gitlab oauth call %s", r.URL.Path)
		}
	}))
	defer server.Close()
	ts.Config.External.Gitlab.URL = server.URL

	// invite user
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(InviteParams{
		Email:          "gitlab@example.com",
		OrganizationID: ts.OrganizationID,
		ProjectID:      ts.ProjectID,
	}))
	req := httptest.NewRequest(http.MethodPost, "http://localhost/invite", &buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.token))

	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusOK, w.Code)

	// Find test user
	user, err := models.FindUserByEmailAndAudience(ts.API.db, "gitlab@example.com", ts.Config.JWT.Aud, ts.OrganizationID, uuid.Nil)
	require.NoError(ts.T(), err)

	// get redirect url w/ state
	provider := "gitlab"
	url_path := fmt.Sprintf("http://localhost/authorize?provider=%s&invite_token=%s", provider, user.ConfirmationToken)
	req = httptest.NewRequest(http.MethodGet, url_path, nil)
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusFound, w.Code)
	u, err := url.Parse(w.Header().Get("Location"))
	ts.Require().NoError(err, "redirect url parse failed")
	q := u.Query()
	state := q.Get("state")

	// auth server callback
	testURL, err := url.Parse("http://localhost/callback")
	ts.Require().NoError(err)
	v := testURL.Query()
	v.Set("code", code)
	v.Set("state", state)
	testURL.RawQuery = v.Encode()
	req = httptest.NewRequest(http.MethodGet, testURL.String(), nil)
	w = httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusFound, w.Code)
	u, err = url.Parse(w.Header().Get("Location"))
	ts.Require().NoError(err, "redirect url parse failed")

	// ensure redirect has #access_token=...
	v, err = url.ParseQuery(u.RawQuery)
	ts.Require().NoError(err, u.RawQuery)
	ts.Require().NotEmpty(v.Get("error_description"))
	ts.Require().Equal("invalid_request", v.Get("error"))
}

type recordingMailClient struct {
	body string
}

func (m *recordingMailClient) Mail(_ context.Context, _, _, body string, _ map[string][]string, _ string) error {
	m.body = body
	return nil
}

func (ts *InviteTestSuite) TestInviteMailUsesAcceptInviteURL() {
	ts.Config.External.Email.Enabled = false
	ts.Config.External.Google.Enabled = true

	user, err := models.NewUser("", "google-invite@example.com", "", ts.Config.JWT.Aud, nil, ts.OrganizationID, ts.ProjectID)
	ts.Require().NoError(err)
	user.ConfirmationToken = "hashed-invite-token"

	recorder := &recordingMailClient{}
	mailer := templatemailer.New(ts.Config, recorder, templatemailer.NewCache())
	external, err := url.Parse("http://localhost:9999")
	ts.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/invite", nil)
	req.Header.Set("redirect_to", "https://dashboard.org/auth/callback")
	ts.Require().NoError(mailer.InviteMail(req, user, "otp", "https://dashboard.org/auth/callback", external))

	ts.Contains(recorder.body, "/accept-invite?")
	ts.Contains(recorder.body, "invite_token=hashed-invite-token")
	ts.NotContains(recorder.body, "provider=google")
	ts.NotContains(recorder.body, "organization_id=")
	ts.NotContains(recorder.body, "project_id=")
	ts.Contains(recorder.body, "redirect_to=")
	ts.NotContains(recorder.body, "/verify")
}

func (ts *InviteTestSuite) TestInviteMailProjectTheme() {
	ts.Config.External.Email.Enabled = false
	ts.Config.External.Google.Enabled = true

	base := ts.T().TempDir()
	ts.Require().NoError(os.MkdirAll(filepath.Join(base, "brawler"), 0o755))
	ts.Require().NoError(os.WriteFile(filepath.Join(base, "brawler", "theme.yaml"), []byte(`copy:
  invite:
    intro: Project specific invite
`), 0o644))
	ts.Config.Mailer.Templates.ProjectDir = base

	user, err := models.NewUser("", "override@example.com", "", ts.Config.JWT.Aud, nil, ts.OrganizationID, ts.ProjectID)
	ts.Require().NoError(err)
	user.ConfirmationToken = "override-token"

	other, err := models.NewUser("", "default@example.com", "", ts.Config.JWT.Aud, nil, ts.OrganizationID, uuid.Must(uuid.NewV4()))
	ts.Require().NoError(err)
	other.ConfirmationToken = "default-token"

	external, err := url.Parse("http://localhost:9999")
	ts.Require().NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/invite", nil)

	cache := templatemailer.NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "BRAWLER", nil
	}

	projectMail := &recordingMailClient{}
	mailer := templatemailer.New(ts.Config, projectMail, cache)
	ts.Require().NoError(mailer.InviteMail(req, user, "otp", "https://dashboard.insidec.org/auth/callback", external))
	ts.Contains(projectMail.body, "Project specific invite")
	ts.Contains(projectMail.body, "invite_token=override-token")

	defaultMail := &recordingMailClient{}
	mailer = templatemailer.New(ts.Config, defaultMail, templatemailer.NewCache())
	ts.Require().NoError(mailer.InviteMail(req, other, "otp", "https://dashboard.insidec.org/auth/callback", external))
	ts.Contains(defaultMail.body, "Accept invitation")
	ts.NotContains(defaultMail.body, "Project specific invite")
}

func (ts *InviteTestSuite) TestInviteMailInvalidProjectThemeUsesDefault() {
	ts.Config.External.Email.Enabled = false
	ts.Config.External.Google.Enabled = true

	base := ts.T().TempDir()
	ts.Require().NoError(os.MkdirAll(filepath.Join(base, "brawler"), 0o755))
	ts.Require().NoError(os.WriteFile(filepath.Join(base, "brawler", "theme.yaml"), []byte(":\n- bad"), 0o644))
	ts.Config.Mailer.Templates.ProjectDir = base

	user, err := models.NewUser("", "broken@example.com", "", ts.Config.JWT.Aud, nil, ts.OrganizationID, ts.ProjectID)
	ts.Require().NoError(err)
	user.ConfirmationToken = "broken-token"

	cache := templatemailer.NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "BRAWLER", nil
	}

	recorder := &recordingMailClient{}
	mailer := templatemailer.New(ts.Config, recorder, cache)
	external, err := url.Parse("http://localhost:9999")
	ts.Require().NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/invite", nil)
	ts.Require().NoError(mailer.InviteMail(req, user, "otp", "https://dashboard.insidec.org/auth/callback", external))
	ts.Contains(recorder.body, "Accept invitation")
}
