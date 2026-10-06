package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/api/apierrors"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/models"
)

func disableAllOAuthProviders(cfg *conf.ProviderConfiguration) {
	cfg.Apple.Enabled = false
	cfg.Azure.Enabled = false
	cfg.Bitbucket.Enabled = false
	cfg.Discord.Enabled = false
	cfg.Facebook.Enabled = false
	cfg.Snapchat.Enabled = false
	cfg.Figma.Enabled = false
	cfg.Fly.Enabled = false
	cfg.Github.Enabled = false
	cfg.Gitlab.Enabled = false
	cfg.Google.Enabled = false
	cfg.Kakao.Enabled = false
	cfg.Notion.Enabled = false
	cfg.Keycloak.Enabled = false
	cfg.Linkedin.Enabled = false
	cfg.LinkedinOIDC.Enabled = false
	cfg.Spotify.Enabled = false
	cfg.Slack.Enabled = false
	cfg.SlackOIDC.Enabled = false
	cfg.Twitter.Enabled = false
	cfg.Twitch.Enabled = false
	cfg.VercelMarketplace.Enabled = false
	cfg.WorkOS.Enabled = false
	cfg.Zoom.Enabled = false
}

func TestAcceptInviteRendersProviderChoices(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)
	defer api.db.Close()

	config.External.Email.Enabled = false
	disableAllOAuthProviders(&config.External)
	config.External.Google.Enabled = true
	config.External.Github.Enabled = true

	projectID, organizationID, _ := InitializeTestDatabase(t, api, config)
	user := createInvitedUserForTest(t, api, organizationID, projectID, "invitee@example.com", "invite-token-hash")

	req := httptest.NewRequest(http.MethodGet, "http://localhost/accept-invite?invite_token="+url.QueryEscape(user.ConfirmationToken)+"&redirect_to="+url.QueryEscape("https://dashboard.example.com/auth/callback"), nil)
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Sign in with Google")
	require.Contains(t, w.Body.String(), "Sign in with GitHub")
	require.Contains(t, w.Body.String(), "provider=google")
	require.Contains(t, w.Body.String(), "provider=github")
	require.Contains(t, w.Body.String(), "invite_token="+user.ConfirmationToken)
}

func TestAcceptInviteRedirectsWhenSingleProvider(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)
	defer api.db.Close()

	config.External.Email.Enabled = false
	disableAllOAuthProviders(&config.External)
	config.External.Google.Enabled = true

	projectID, organizationID, _ := InitializeTestDatabase(t, api, config)
	user := createInvitedUserForTest(t, api, organizationID, projectID, "solo@example.com", "solo-invite-token")

	req := httptest.NewRequest(http.MethodGet, "http://localhost/accept-invite?invite_token="+url.QueryEscape(user.ConfirmationToken), nil)
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	require.Contains(t, location, "/authorize?")
	require.Contains(t, location, "provider=google")
	require.Contains(t, location, "invite_token="+user.ConfirmationToken)
}

func TestAcceptInviteRejectsExpiredAndBannedTokens(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)
	defer api.db.Close()

	config.External.Email.Enabled = false
	disableAllOAuthProviders(&config.External)
	config.External.Google.Enabled = true

	projectID, organizationID, _ := InitializeTestDatabase(t, api, config)

	expired := createInvitedUserForTest(t, api, organizationID, projectID, "expired-invite@example.com", "expired-invite-token")
	sentAt := time.Now().Add(-time.Duration(config.Mailer.OtpExp+60) * time.Second)
	expired.ConfirmationSentAt = &sentAt
	require.NoError(t, api.db.UpdateOnly(expired, "confirmation_sent_at"))

	banned := createInvitedUserForTest(t, api, organizationID, projectID, "banned-invite@example.com", "banned-invite-token")
	until := time.Now().Add(time.Hour)
	banned.BannedUntil = &until
	require.NoError(t, api.db.UpdateOnly(banned, "banned_until"))

	expiredReq := httptest.NewRequest(http.MethodGet, "http://localhost/accept-invite?invite_token="+url.QueryEscape(expired.ConfirmationToken), nil)
	expiredRec := httptest.NewRecorder()
	api.handler.ServeHTTP(expiredRec, expiredReq)
	require.Equal(t, http.StatusForbidden, expiredRec.Code)
	require.Contains(t, expiredRec.Body.String(), apierrors.ErrorCodeOTPExpired)

	bannedReq := httptest.NewRequest(http.MethodGet, "http://localhost/accept-invite?invite_token="+url.QueryEscape(banned.ConfirmationToken), nil)
	bannedRec := httptest.NewRecorder()
	api.handler.ServeHTTP(bannedRec, bannedReq)
	require.Equal(t, http.StatusForbidden, bannedRec.Code)
	require.Contains(t, bannedRec.Body.String(), apierrors.ErrorCodeUserBanned)
}

func TestAuthorizeRejectsExpiredAndBannedInviteTokens(t *testing.T) {
	api, config, err := setupAPIForTest()
	require.NoError(t, err)
	defer api.db.Close()

	config.External.Github.Enabled = true
	config.External.Github.ClientID = []string{"test-client"}
	config.External.Github.Secret = "test-secret"
	config.External.Github.RedirectURI = "http://localhost/callback"

	projectID, organizationID, _ := InitializeTestDatabase(t, api, config)

	expired := createInvitedUserForTest(t, api, organizationID, projectID, "expired-authorize@example.com", "expired-authorize-token")
	sentAt := time.Now().Add(-time.Duration(config.Mailer.OtpExp+60) * time.Second)
	expired.ConfirmationSentAt = &sentAt
	require.NoError(t, api.db.UpdateOnly(expired, "confirmation_sent_at"))

	banned := createInvitedUserForTest(t, api, organizationID, projectID, "banned-authorize@example.com", "banned-authorize-token")
	until := time.Now().Add(time.Hour)
	banned.BannedUntil = &until
	require.NoError(t, api.db.UpdateOnly(banned, "banned_until"))

	expiredReq := httptest.NewRequest(http.MethodGet, "http://localhost/authorize?provider=github&invite_token="+url.QueryEscape(expired.ConfirmationToken), nil)
	expiredRec := httptest.NewRecorder()
	api.handler.ServeHTTP(expiredRec, expiredReq)
	require.Equal(t, http.StatusForbidden, expiredRec.Code)
	require.NotContains(t, expiredRec.Header().Get("Location"), "github.com")
	require.Contains(t, expiredRec.Body.String(), apierrors.ErrorCodeOTPExpired)

	bannedReq := httptest.NewRequest(http.MethodGet, "http://localhost/authorize?provider=github&invite_token="+url.QueryEscape(banned.ConfirmationToken), nil)
	bannedRec := httptest.NewRecorder()
	api.handler.ServeHTTP(bannedRec, bannedReq)
	require.Equal(t, http.StatusForbidden, bannedRec.Code)
	require.NotContains(t, bannedRec.Header().Get("Location"), "github.com")
	require.Contains(t, bannedRec.Body.String(), apierrors.ErrorCodeUserBanned)
}

func createInvitedUserForTest(t *testing.T, api *API, organizationID, projectID uuid.UUID, email, confirmationToken string) *models.User {
	t.Helper()

	if u, _ := models.FindUserByEmailAndAudience(api.db, email, api.config.JWT.Aud, organizationID, projectID); u != nil {
		require.NoError(t, api.db.Destroy(u))
	}

	user, err := models.NewUser("", email, "test", api.config.JWT.Aud, nil, organizationID, projectID)
	require.NoError(t, err)
	now := time.Now()
	user.ConfirmationToken = confirmationToken
	user.ConfirmationSentAt = &now
	require.NoError(t, api.db.Create(user))
	require.NoError(t, models.CreateOneTimeToken(api.db, user.ID, email, user.ConfirmationToken, models.ConfirmationToken))

	identity, err := models.NewIdentity(user, "email", map[string]interface{}{
		"sub":   user.ID.String(),
		"email": email,
	})
	require.NoError(t, err)
	require.NoError(t, api.db.Create(identity))

	return user
}
