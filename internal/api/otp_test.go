package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/conf"
	"auth/internal/models"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OtpTestSuite struct {
	suite.Suite
	API    *API
	Config *conf.GlobalConfiguration
}

func TestOtp(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)

	ts := &OtpTestSuite{
		API:    api,
		Config: config,
	}
	defer api.db.Close()

	suite.Run(t, ts)
}

func (ts *OtpTestSuite) SetupTest() {
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

	// Create the organization if it doesn't exist
	organization_id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	if err := ts.API.db.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization', '%s', '%s')", organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := ts.API.db.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

}

func (ts *OtpTestSuite) TestOtpPKCE() {
	ts.Config.External.Phone.Enabled = true
	testCodeChallenge := "testtesttesttesttesttesttestteststeststesttesttesttest"

	var buffer bytes.Buffer
	cases := []struct {
		desc     string
		params   OtpParams
		expected struct {
			code     int
			response map[string]interface{}
		}
	}{
		{
			desc: "Test (PKCE) Success Magiclink Otp",
			params: OtpParams{
				Email:               "test@example.com",
				CreateUser:          true,
				CodeChallengeMethod: "s256",
				CodeChallenge:       testCodeChallenge,
				OrganizationID:      uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusOK,
				make(map[string]interface{}),
			},
		},
		{
			desc: "Test (PKCE) Failure, no code challenge",
			params: OtpParams{
				Email:               "test@example.com",
				CreateUser:          true,
				CodeChallengeMethod: "s256",
				OrganizationID:      uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusBadRequest,
				map[string]interface{}{
					"code":       float64(http.StatusBadRequest),
					"error_code": ErrorCodeValidationFailed,
					"msg":        "PKCE flow requires code_challenge_method and code_challenge",
				},
			},
		},
		{
			desc: "Test (PKCE) Failure, no code challenge method",
			params: OtpParams{
				Email:          "test@example.com",
				CreateUser:     true,
				CodeChallenge:  testCodeChallenge,
				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusBadRequest,
				map[string]interface{}{
					"code":       float64(http.StatusBadRequest),
					"error_code": ErrorCodeValidationFailed,
					"msg":        "PKCE flow requires code_challenge_method and code_challenge",
				},
			},
		},
		{
			desc: "Test (PKCE) Success, phone with valid params",
			params: OtpParams{
				Phone:               "123456789",
				CreateUser:          true,
				CodeChallengeMethod: "s256",
				CodeChallenge:       testCodeChallenge,
				OrganizationID:      uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusInternalServerError,
				map[string]interface{}{
					"code": float64(http.StatusInternalServerError),
					"msg":  "Unable to get SMS provider",
				},
			},
		},
	}
	for _, c := range cases {
		ts.Run(c.desc, func() {
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(c.params))

			req := httptest.NewRequest(http.MethodPost, "/otp", &buffer)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.API.handler.ServeHTTP(w, req)

			require.Equal(ts.T(), c.expected.code, w.Code)
			data := make(map[string]interface{})
			require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))

		})
	}
}

func (ts *OtpTestSuite) TestOtp() {
	// Configured to allow testing of invalid channel params
	ts.Config.External.Phone.Enabled = true
	cases := []struct {
		desc     string
		params   OtpParams
		expected struct {
			code     int
			response map[string]interface{}
		}
	}{
		{
			desc: "Test Success Magiclink Otp",
			params: OtpParams{
				Email:      "test@example.com",
				CreateUser: true,
				Data: map[string]interface{}{
					"somedata": "metadata",
				},
				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusOK,
				make(map[string]interface{}),
			},
		},
		{
			desc: "Test Failure Pass Both Email & Phone",
			params: OtpParams{
				Email:          "test@example.com",
				Phone:          "123456789",
				CreateUser:     true,
				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusBadRequest,
				map[string]interface{}{
					"code":       float64(http.StatusBadRequest),
					"error_code": ErrorCodeValidationFailed,
					"msg":        "Only an email address or phone number should be provided",
				},
			},
		},
		{
			desc: "Test Failure invalid channel param",
			params: OtpParams{
				Phone:          "123456789",
				Channel:        "invalidchannel",
				CreateUser:     true,
				OrganizationID: uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
			},
			expected: struct {
				code     int
				response map[string]interface{}
			}{
				http.StatusBadRequest,
				map[string]interface{}{
					"code":       float64(http.StatusBadRequest),
					"error_code": ErrorCodeValidationFailed,
					"msg":        InvalidChannelError,
				},
			},
		},
	}

	for _, c := range cases {
		ts.Run(c.desc, func() {
			var buffer bytes.Buffer
			require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(c.params))

			req := httptest.NewRequest(http.MethodPost, "/otp", &buffer)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			ts.API.handler.ServeHTTP(w, req)

			require.Equal(ts.T(), c.expected.code, w.Code)

			data := make(map[string]interface{})
			require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))

			// response should be empty
			assert.Equal(ts.T(), data, c.expected.response)
		})
	}
}

func (ts *OtpTestSuite) TestNoSignupsForOtp() {
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":           "newuser@example.com",
		"create_user":     false,
		"organization_id": uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
	}))

	req := httptest.NewRequest(http.MethodPost, "/otp", &buffer)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)

	require.Equal(ts.T(), http.StatusUnprocessableEntity, w.Code)

	data := make(map[string]interface{})
	require.NoError(ts.T(), json.NewDecoder(w.Body).Decode(&data))

	// response should be empty
	assert.Equal(ts.T(), data, map[string]interface{}{
		"code":       float64(http.StatusUnprocessableEntity),
		"error_code": ErrorCodeOTPDisabled,
		"msg":        "Signups not allowed for otp",
	})
}

func (ts *OtpTestSuite) TestSubsequentOtp() {
	ts.Config.SMTP.MaxFrequency = 0
	userEmail := "foo@example.com"
	var buffer bytes.Buffer
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":           userEmail,
		"organization_id": uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
	}))

	req := httptest.NewRequest(http.MethodPost, "/otp", &buffer)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)

	require.Equal(ts.T(), http.StatusOK, w.Code)
	id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	newUser, err := models.FindUserByEmailAndAudience(ts.API.db, userEmail, ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)
	require.NotEmpty(ts.T(), newUser.ConfirmationToken)
	require.NotEmpty(ts.T(), newUser.ConfirmationSentAt)
	require.Empty(ts.T(), newUser.RecoveryToken)
	require.Empty(ts.T(), newUser.RecoverySentAt)
	require.Empty(ts.T(), newUser.EmailConfirmedAt)

	// since the signup process hasn't been completed,
	// subsequent requests for another magiclink should not create a recovery token
	require.NoError(ts.T(), json.NewEncoder(&buffer).Encode(map[string]interface{}{
		"email":           userEmail,
		"organization_id": uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000")),
	}))

	req = httptest.NewRequest(http.MethodPost, "/otp", &buffer)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()

	ts.API.handler.ServeHTTP(w, req)

	require.Equal(ts.T(), http.StatusOK, w.Code)

	user, err := models.FindUserByEmailAndAudience(ts.API.db, userEmail, ts.Config.JWT.Aud, id, uuid.Nil)
	require.NoError(ts.T(), err)
	require.NotEmpty(ts.T(), user.ConfirmationToken)
	require.NotEmpty(ts.T(), user.ConfirmationSentAt)
	require.Empty(ts.T(), user.RecoveryToken)
	require.Empty(ts.T(), user.RecoverySentAt)
	require.Empty(ts.T(), user.EmailConfirmedAt)
}
