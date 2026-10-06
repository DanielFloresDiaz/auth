package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/mailer/mockclient"
	"github.com/supabase/auth/internal/models"
)

func TestWhitelistUnknownProject(t *testing.T) {
	api, _, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	body := whitelistBody(t, uuid.Must(uuid.NewV4()), "ada@acme.com", map[string]any{"name": "Ada"})
	w := postWhitelist(api, body)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.Empty(t, mailer.WhitelistConfirmationMailCalls)
}

func TestWhitelistRejectsBadEmail(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	body := whitelistBody(t, projectID, "not-an-email", map[string]any{})
	w := postWhitelist(api, body)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Empty(t, mailer.WhitelistConfirmationMailCalls)
}

func TestWhitelistCreateStoresRequestOnlyAfterConfirm(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	body := whitelistBody(t, projectID, "Ada@Acme.com", map[string]any{"name": "Ada", "company": "Acme"})
	w := postWhitelist(api, body)
	require.Equal(t, http.StatusOK, w.Code)
	assertWhitelistAck(t, w.Body.Bytes())
	require.NotContains(t, w.Body.String(), "Ada")
	require.NotContains(t, w.Body.String(), "Acme")

	var requestCount int
	require.NoError(t, api.db.RawQuery(
		`SELECT count(*) FROM auth.whitelist_requests WHERE project_id = ? AND lower(email) = ?`,
		projectID,
		"ada@acme.com",
	).First(&requestCount))
	require.Equal(t, 0, requestCount)

	require.Len(t, mailer.WhitelistConfirmationMailCalls, 1)
	require.Equal(t, "ada@acme.com", mailer.WhitelistConfirmationMailCalls[0].Email)

	confirm := httptest.NewRecorder()
	api.handler.ServeHTTP(confirm, whitelistConfirmRequest(t, mailer.WhitelistConfirmationMailCalls[0].ConfirmURL))
	require.Equal(t, http.StatusOK, confirm.Code)
	require.Contains(t, confirm.Body.String(), "Access request recorded")
	require.NotContains(t, confirm.Body.String(), "Ada")

	stored, err := models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.Equal(t, models.WhitelistStatusPending, stored.Status)
	answers, err := stored.AnswersObject()
	require.NoError(t, err)
	require.Equal(t, "Ada", answers["name"])
	require.Equal(t, "Acme", answers["company"])

	confirmation, err := models.FindWhitelistConfirmation(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.Nil(t, confirmation)
}

func TestWhitelistOpenRequestDoesNotLeakOrResend(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	firstBody := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Ada", "phone": "+15551212"})
	first := postWhitelist(api, firstBody)
	require.Equal(t, http.StatusOK, first.Code)
	confirm := httptest.NewRecorder()
	api.handler.ServeHTTP(confirm, whitelistConfirmRequest(t, mailer.WhitelistConfirmationMailCalls[0].ConfirmURL))
	require.Equal(t, http.StatusOK, confirm.Code)

	mailer.Reset()
	secondBody := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "nope"})
	second := postWhitelist(api, secondBody)
	require.Equal(t, http.StatusOK, second.Code)
	assertWhitelistAck(t, second.Body.Bytes())
	require.NotContains(t, second.Body.String(), "+15551212")
	require.NotContains(t, second.Body.String(), "Ada")
	require.Empty(t, mailer.WhitelistConfirmationMailCalls)

	stored, err := models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	answers, err := stored.AnswersObject()
	require.NoError(t, err)
	require.Equal(t, "Ada", answers["name"])
	require.Equal(t, "+15551212", answers["phone"])
}

func TestWhitelistRejectedEmailCanReapplyAfterConfirm(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	firstBody := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Ada"})
	first := postWhitelist(api, firstBody)
	require.Equal(t, http.StatusOK, first.Code)
	confirm := httptest.NewRecorder()
	api.handler.ServeHTTP(confirm, whitelistConfirmRequest(t, mailer.WhitelistConfirmationMailCalls[0].ConfirmURL))
	require.Equal(t, http.StatusOK, confirm.Code)

	require.NoError(t, api.db.RawQuery(
		`UPDATE auth.whitelist_requests SET status = 'rejected', rejected_at = current_timestamp WHERE project_id = ? AND lower(email) = ?`,
		projectID,
		"ada@acme.com",
	).Exec())

	mailer.Reset()
	secondBody := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Grace"})
	second := postWhitelist(api, secondBody)
	require.Equal(t, http.StatusOK, second.Code)
	assertWhitelistAck(t, second.Body.Bytes())
	require.NotContains(t, second.Body.String(), "denied")
	require.NotContains(t, second.Body.String(), "Ada")
	require.Len(t, mailer.WhitelistConfirmationMailCalls, 1)

	reconfirm := httptest.NewRecorder()
	api.handler.ServeHTTP(reconfirm, whitelistConfirmRequest(t, mailer.WhitelistConfirmationMailCalls[0].ConfirmURL))
	require.Equal(t, http.StatusOK, reconfirm.Code)

	stored, err := models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.Equal(t, models.WhitelistStatusPending, stored.Status)
	answers, err := stored.AnswersObject()
	require.NoError(t, err)
	require.Equal(t, "Grace", answers["name"])

	var count int
	require.NoError(t, api.db.RawQuery(
		`SELECT count(*) FROM auth.whitelist_requests WHERE project_id = ? AND lower(email) = ?`,
		projectID,
		"ada@acme.com",
	).First(&count))
	require.Equal(t, 1, count)
}

func TestWhitelistExpiredConfirmationDoesNotInsert(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	body := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Ada"})
	w := postWhitelist(api, body)
	require.Equal(t, http.StatusOK, w.Code)

	expiredAt := time.Now().Add(-time.Duration(config.Mailer.OtpExp+60) * time.Second)
	require.NoError(t, api.db.RawQuery(
		`UPDATE auth.whitelist_confirmations SET sent_at = ? WHERE project_id = ? AND lower(email) = ?`,
		expiredAt,
		projectID,
		"ada@acme.com",
	).Exec())

	confirm := httptest.NewRecorder()
	api.handler.ServeHTTP(confirm, whitelistConfirmRequest(t, mailer.WhitelistConfirmationMailCalls[0].ConfirmURL))
	require.Equal(t, http.StatusBadRequest, confirm.Code)
	require.Contains(t, confirm.Body.String(), "invalid or has expired")

	stored, err := models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestWhitelistNewerPostInvalidatesOlderConfirmLink(t *testing.T) {
	api, config, mailer := setupWhitelistAPI(t)
	defer api.db.Close()
	config.SMTP.MaxFrequency = 0

	projectID, _, _ := InitializeTestDatabase(t, api, config)
	body := whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Ada"})
	first := postWhitelist(api, body)
	require.Equal(t, http.StatusOK, first.Code)
	oldURL := mailer.WhitelistConfirmationMailCalls[0].ConfirmURL

	second := postWhitelist(api, whitelistBody(t, projectID, "ada@acme.com", map[string]any{"name": "Grace"}))
	require.Equal(t, http.StatusOK, second.Code)
	require.Len(t, mailer.WhitelistConfirmationMailCalls, 2)
	newURL := mailer.WhitelistConfirmationMailCalls[1].ConfirmURL
	require.NotEqual(t, oldURL, newURL)

	oldConfirm := httptest.NewRecorder()
	api.handler.ServeHTTP(oldConfirm, whitelistConfirmRequest(t, oldURL))
	require.Equal(t, http.StatusBadRequest, oldConfirm.Code)

	stored, err := models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.Nil(t, stored)

	newConfirm := httptest.NewRecorder()
	api.handler.ServeHTTP(newConfirm, whitelistConfirmRequest(t, newURL))
	require.Equal(t, http.StatusOK, newConfirm.Code)

	stored, err = models.FindWhitelistRequest(api.db, projectID, "ada@acme.com")
	require.NoError(t, err)
	require.NotNil(t, stored)
	answers, err := stored.AnswersObject()
	require.NoError(t, err)
	require.Equal(t, "Grace", answers["name"])
}

func setupWhitelistAPI(t *testing.T) (*API, *conf.GlobalConfiguration, *mockclient.MockMailer) {
	t.Helper()
	mailer := &mockclient.MockMailer{}
	api, config, err := setupAPIForTest(WithMailer(mailer))
	require.NoError(t, err)
	return api, config, mailer
}

func assertWhitelistAck(t *testing.T, body []byte) {
	t.Helper()
	var response whitelistAckResponse
	require.NoError(t, json.Unmarshal(body, &response))
	require.Equal(t, whitelistAckMessage, response.Message)
	require.NotContains(t, string(body), "answers")
	require.NotContains(t, string(body), "status")
}

func whitelistBody(t *testing.T, projectID uuid.UUID, email string, answers map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"project_id": projectID,
		"email":      email,
		"answers":    answers,
	})
	require.NoError(t, err)
	return body
}

func postWhitelist(api *API, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "http://localhost/whitelist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)
	return w
}

func whitelistConfirmRequest(t *testing.T, confirmURL string) *http.Request {
	t.Helper()
	parsed, err := url.Parse(confirmURL)
	require.NoError(t, err)
	token := parsed.Query().Get("token")
	require.NotEmpty(t, token)
	return httptest.NewRequest(http.MethodGet, "http://localhost/whitelist/confirm?token="+url.QueryEscape(token), nil)
}
