package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func (ts *ExternalTestSuite) TestSignupExternalApple() {
	provider := "apple"
	url_path := fmt.Sprintf("http://localhost/authorize?provider=%s&organization_id=%s&project_id=%s", provider, ts.OrganizationID.String(), ts.ProjectID.String())
	req := httptest.NewRequest(http.MethodGet, url_path, nil)
	w := httptest.NewRecorder()
	ts.API.handler.ServeHTTP(w, req)
	ts.Require().Equal(http.StatusFound, w.Code)
	u, err := url.Parse(w.Header().Get("Location"))
	ts.Require().NoError(err, "redirect url parse failed")
	q := u.Query()
	ts.Equal(ts.Config.External.Apple.RedirectURI, q.Get("redirect_uri"))
	ts.Equal(ts.Config.External.Apple.ClientID, []string{q.Get("client_id")})
	ts.Equal("code", q.Get("response_type"))
	ts.Equal("email name", q.Get("scope"))

	assertValidOAuthState(ts, q.Get("state"), "apple")
}
