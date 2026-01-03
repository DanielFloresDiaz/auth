package e2eapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/supabase/auth/internal/api"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/e2e"
	"github.com/supabase/auth/internal/models"
)

type E2EAPITestSuite struct {
	suite.Suite
	Instance       *Instance
	Config         *conf.GlobalConfiguration
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
}

func TestE2EAPI(t *testing.T) {
	globalCfg := e2e.Must(e2e.Config())
	inst, err := New(globalCfg)
	require.NoError(t, err)

	ts := &E2EAPITestSuite{
		Instance: inst,
		Config:   globalCfg,
	}
	defer inst.Close()

	suite.Run(t, ts)
}

func (ts *E2EAPITestSuite) SetupTest() {
	models.TruncateAll(ts.Instance.Conn)

	project_id := uuid.Must(uuid.NewV4())
	ts.ProjectID = project_id
	// Create a project
	if err := ts.Instance.Conn.RawQuery(fmt.Sprintf("INSERT INTO auth.projects (id, name) VALUES ('%s', 'test_project')", project_id)).Exec(); err != nil {
		panic(err)
	}

	// Create the admin of the organization
	user, err := models.NewUser("", "admin@example.com", "test", ts.Config.JWT.Aud, nil, uuid.Nil, project_id)
	require.NoError(ts.T(), err, "Error making new user")
	require.NoError(ts.T(), ts.Instance.Conn.Create(user, "organization_id", "organization_role"), "Error creating user")

	// Create the organization
	organization_id := uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
	ts.OrganizationID = organization_id
	if err := ts.Instance.Conn.RawQuery(fmt.Sprintf("INSERT INTO auth.organizations (id, name, project_id, admin_id) VALUES ('%s', 'test_organization', '%s', '%s')", organization_id, project_id, user.ID)).Exec(); err != nil {
		panic(err)
	}

	// Set the user as the admin of the organization
	if err := ts.Instance.Conn.RawQuery(fmt.Sprintf("UPDATE auth.users SET organization_id = '%s', organization_role='admin' WHERE id = '%s'", organization_id, user.ID)).Exec(); err != nil {
		panic(err)
	}
}

func (ts *E2EAPITestSuite) TestInstance() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	ts.Run("Success", func() {
		email := "e2eapitest_" + uuid.Must(uuid.NewV4()).String() + "@localhost"
		req := &api.SignupParams{
			Email:          email,
			Password:       "password",
			OrganizationID: ts.OrganizationID,
		}
		res := new(models.User)
		err := Do(ctx, http.MethodPost, ts.Instance.APIServer.URL+"/signup", req, res)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), email, res.Email.String())
	})

	ts.Run("DoAdmin", func() {
		email := "e2eapitest_" + uuid.Must(uuid.NewV4()).String() + "@localhost"
		req := &api.InviteParams{
			Email:          email,
			OrganizationID: ts.OrganizationID,
		}
		res := new(models.User)

		body := new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(req)
		require.NoError(ts.T(), err)

		httpReq, err := http.NewRequestWithContext(
			ctx, "POST", "/invite", body)
		require.NoError(ts.T(), err)

		httpRes, err := ts.Instance.DoAdmin(httpReq)
		require.NoError(ts.T(), err)

		err = json.NewDecoder(httpRes.Body).Decode(res)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), email, res.Email.String())
	})

	ts.Run("DoAdminFailure", func() {
		httpReq, err := http.NewRequestWithContext(
			ctx, "POST", "/invite", nil)
		require.NoError(ts.T(), err)

		httpRes, err := ts.Instance.doAdmin(httpReq, new(int))
		require.Error(ts.T(), err)
		require.Nil(ts.T(), httpRes)

	})

	ts.Run("Failure", func() {
		globalCfg := e2e.Must(e2e.Config())
		globalCfg.DB.Driver = ""
		globalCfg.DB.URL = "invalid"

		inst, err := New(globalCfg)
		require.Error(ts.T(), err)
		require.Nil(ts.T(), inst)
	})

	ts.Run("InitURLFailure", func() {
		globalCfg := e2e.Must(e2e.Config())
		inst, err := New(globalCfg)
		require.NoError(ts.T(), err)
		defer inst.Close()

		inst.APIServer.URL = "\x01"
		err = inst.initURL()
		require.Error(ts.T(), err)
	})
}

func (ts *E2EAPITestSuite) TestDo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Covers calls to Do with a `req` param type which can't marshaled
	ts.Run("InvalidRequestType", func() {
		req := make(chan string)
		err := Do(ctx, http.MethodPost, "http://localhost", &req, nil)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "json: unsupported type: chan string")
	})

	// Covers calls to Do with a `res` param type which can't marshaled
	ts.Run("InvalidResponseType", func() {
		res := make(chan string)
		err := Do(ctx, http.MethodGet, ts.Instance.APIServer.URL+"/settings", nil, &res)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "json: cannot unmarshal object into Go value of type chan string")
	})

	// Covers status code >= 400 error handling switch statement
	ts.Run("api.HTTPErrorResponse_to_apierrors.HTTPError", func() {
		res := make(chan string)
		err := Do(ctx, http.MethodGet, ts.Instance.APIServer.URL+"/user", nil, &res)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "401: This endpoint requires a valid Bearer token")
	})

	// Covers http.NewRequestWithContext
	ts.Run("InvalidHTTPMethod", func() {
		err := Do(ctx, "\x01", "http://localhost", nil, nil)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "net/http: invalid method")
	})

	// Covers status code >= 400 error handling switch statement json.Unmarshal
	// by hitting the default error handler that returns html
	ts.Run("InvalidResponse", func() {
		err := Do(ctx, http.MethodGet, ts.Instance.APIServer.URL+"/404", nil, nil)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "invalid character")
	})

	// Covers defaultClient.Do failure
	ts.Run("InvalidURL", func() {
		err := Do(ctx, http.MethodPost, "invalid", nil, nil)
		require.Error(ts.T(), err)
		require.ErrorContains(ts.T(), err, "unsupported protocol")
	})

	// Covers http.StatusNoContent handling
	ts.Run("StatusNoContent", func() {
		hr := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		srv := httptest.NewServer(hr)
		defer srv.Close()

		err := Do(ctx, http.MethodPost, srv.URL, nil, nil)
		require.NoError(ts.T(), err)
	})

	// Covers IO errors
	ts.Run("IOError", func() {

		for _, statusCode := range []int{http.StatusBadRequest, http.StatusOK} {

			// Covers IO errors for the sc >= 400 and default status code
			// handling in the switch statement within do.
			testName := fmt.Sprintf("Status=%v", http.StatusText(statusCode))
			ts.Run(testName, func() {

				// We assign a sentinel error to ensure propagation.
				sentinel := errors.New("sentinel")

				// This implementation of the http.RoundTripper is a way to
				// cover the io.ReadAll(io.LimitReader(...)) lines in the switch
				// statements inside do().
				rtFn := roundTripperFunc(func(req *http.Request) (*http.Response, error) {

					// Call the default http.RoundTripper implementation provided
					// by the http.Default client to build a valid http.Response.
					res, err := http.DefaultClient.Do(req)
					if err != nil {
						return nil, err
					}

					// Wrap the res.Body in an io.ErrReader using our sentinel
					// error. This causes the first call to read the response
					// body to return our sentinel error.
					res.Body = io.NopCloser(iotest.ErrReader(sentinel))
					return res, nil
				})

				// We need to swap the defaultClient with a new client which has
				// the (*Client).Transport set to our http.RoundTripper above.
				prev := defaultClient
				defer func() {
					defaultClient = prev
				}()
				defaultClient = new(http.Client)
				defaultClient.Transport = rtFn

				hr := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(statusCode)
				})

				srv := httptest.NewServer(hr)
				defer srv.Close()

				// We send the request and expect back our sentinel error.
				err := Do(ctx, http.MethodPost, srv.URL, nil, nil)
				require.Error(ts.T(), err)
				require.Equal(ts.T(), sentinel, err)
			})
		}
	})
}

// roundTripperFunc is like http.HandlerFunc for a http.RoundTripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper by calling itself.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
