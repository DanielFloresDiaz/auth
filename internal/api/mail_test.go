package api

import (
	"fmt"
	"testing"

	"auth/internal/conf"
	"auth/internal/models"

	"github.com/gobwas/glob"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MailTestSuite struct {
	suite.Suite
	API    *API
	Config *conf.GlobalConfiguration
}

func TestMail(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)

	ts := &MailTestSuite{
		API:    api,
		Config: config,
	}
	defer api.db.Close()

	suite.Run(t, ts)
}

func (ts *MailTestSuite) SetupTest() {
	models.TruncateAll(ts.API.db)

	ts.Config.Mailer.SecureEmailChangeEnabled = true

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

	// Create User
	u, err := models.NewUser("12345678", "test@example.com", "password", ts.Config.JWT.Aud, nil, organization_id, uuid.Nil)
	require.NoError(ts.T(), err, "Error creating new user model")
	require.NoError(ts.T(), ts.API.db.Create(u, "project_id", "organization_role"), "Error saving new user")
}

func (ts *MailTestSuite) TestValidateEmail() {
	cases := []struct {
		desc          string
		email         string
		expectedEmail string
		expectedError error
	}{
		{
			desc:          "valid email",
			email:         "test@example.com",
			expectedEmail: "test@example.com",
			expectedError: nil,
		},
		{
			desc:          "email should be normalized",
			email:         "TEST@EXAMPLE.COM",
			expectedEmail: "test@example.com",
			expectedError: nil,
		},
		{
			desc:          "empty email should return error",
			email:         "",
			expectedEmail: "",
			expectedError: badRequestError(ErrorCodeValidationFailed, "An email address is required"),
		},
		{
			desc: "email length exceeds 255 characters",
			// email has 256 characters
			email:         "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest@example.com",
			expectedEmail: "",
			expectedError: badRequestError(ErrorCodeValidationFailed, "An email address is too long"),
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			email, err := ts.API.validateEmail(c.email)
			require.Equal(ts.T(), c.expectedError, err)
			require.Equal(ts.T(), c.expectedEmail, email)
		})
	}
}

// func (ts *MailTestSuite) TestGenerateLink() {
// 	// create admin jwt
// 	claims := &AccessTokenClaims{
// 		Role: "supabase_admin",
// 	}
// 	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(ts.Config.JWT.Secret))
// 	require.NoError(ts.T(), err, "Error generating admin jwt")

// 	ts.setURIAllowListMap("http://localhost:8000/**")
// 	// create test cases
// 	cases := []struct {
// 		Desc             string
// 		Body             GenerateLinkParams
// 		ExpectedCode     int
// 		ExpectedResponse map[string]interface{}
// 	}{
// 		{
// 			Desc: "Generate signup link for new user",
// 			Body: GenerateLinkParams{
// 				Email:          "new_user@example.com",
// 				Password:       "secret123",
// 				Type:           "signup",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate signup link for existing user",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				Password:       "secret123",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "signup",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate signup link with custom redirect url",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				Password:       "secret123",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "signup",
// 				RedirectTo:     "http://localhost:8000/welcome",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": "http://localhost:8000/welcome",
// 			},
// 		},
// 		{
// 			Desc: "Generate magic link",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "magiclink",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate invite link",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "invite",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate recovery link",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "recovery",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate email change link",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				NewEmail:       "new@example.com",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				Type:           "email_change_current",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 		{
// 			Desc: "Generate email change link",
// 			Body: GenerateLinkParams{
// 				Email:          "test@example.com",
// 				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
// 				NewEmail:       "new@example.com",
// 				Type:           "email_change_new",
// 			},
// 			ExpectedCode: http.StatusOK,
// 			ExpectedResponse: map[string]interface{}{
// 				"redirect_to": ts.Config.SiteURL,
// 			},
// 		},
// 	}

// 	customDomainUrl, err := url.ParseRequestURI("https://example.gotrue.com")
// 	require.NoError(ts.T(), err)

// 	originalHosts := ts.API.config.Mailer.ExternalHosts
// 	ts.API.config.Mailer.ExternalHosts = []string{
// 		"example.gotrue.com",
// 	}

// 	for _, c := range cases {
// 		ts.Run(c.Desc, func() {
// 			var buffer bytes.Buffer
// 			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(c.Body))
// 			req := httptest.NewRequest(http.MethodPost, customDomainUrl.String()+"/admin/generate_link", &buffer)
// 			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
// 			w := httptest.NewRecorder()

// 			ts.API.handler.ServeHTTP(w, req)

// 			require.Equal(ts.T(), c.ExpectedCode, w.Code)

// 			data := make(map[string]interface{})
// 			require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))

// 			require.Contains(ts.T(), data, "action_link")
// 			require.Contains(ts.T(), data, "email_otp")
// 			require.Contains(ts.T(), data, "hashed_token")
// 			require.Contains(ts.T(), data, "redirect_to")
// 			require.Equal(ts.T(), c.Body.Type, data["verification_type"])

// 			// check if redirect_to is correct
// 			require.Equal(ts.T(), c.ExpectedResponse["redirect_to"], data["redirect_to"])

// 			// check if hashed_token matches hash function of email and the raw otp
// 			require.Equal(ts.T(), crypto.GenerateTokenHash(c.Body.Email, data["email_otp"].(string)), data["hashed_token"])

// 			// check if the host used in the email link matches the initial request host
// 			u, err := url.ParseRequestURI(data["action_link"].(string))
// 			require.NoError(ts.T(), err)
// 			require.Equal(ts.T(), req.Host, u.Host)
// 		})
// 	}

// 	ts.API.config.Mailer.ExternalHosts = originalHosts
// }

func (ts *MailTestSuite) setURIAllowListMap(uris ...string) {
	for _, uri := range uris {
		g := glob.MustCompile(uri, '.', '/')
		ts.Config.URIAllowListMap[uri] = g
	}
}
