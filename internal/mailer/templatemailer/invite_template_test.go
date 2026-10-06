package templatemailer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/models"
	"gopkg.in/yaml.v3"
)

func TestInviteMailUsesAcceptInviteURL(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	cfg.SiteURL = "https://example.netlify.com"
	cfg.External.Email.Enabled = false
	cfg.External.Google.Enabled = true
	cfg.Mailer.URLPaths.Invite = "/verify"

	orgID := uuid.Must(uuid.NewV4())
	projectID := uuid.Must(uuid.NewV4())
	user, err := models.NewUser("", "google-invite@example.com", "", "authenticated", nil, orgID, projectID)
	require.NoError(t, err)
	user.ConfirmationToken = "hashed-invite-token"

	recorder := &recordingMailClient{}
	mailer := New(cfg, recorder, NewCache())
	external, err := url.Parse("http://localhost:9999")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/invite", nil)
	require.NoError(t, mailer.InviteMail(req, user, "otp", "https://dashboard.insidec.org/auth/callback", external))
	require.Contains(t, recorder.body, "/accept-invite?")
	require.Contains(t, recorder.body, "invite_token=hashed-invite-token")
	require.NotContains(t, recorder.body, "provider=google")
	require.NotContains(t, recorder.body, "provider=github")
	require.NotContains(t, recorder.body, "organization_id=")
	require.NotContains(t, recorder.body, "project_id=")
	require.Contains(t, recorder.body, "redirect_to=")
	require.Contains(t, recorder.body, "Accept invitation")
	require.NotContains(t, recorder.body, "/verify")
}

func TestInviteMailProjectTheme(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	cfg.SiteURL = "https://example.netlify.com"
	cfg.External.Email.Enabled = false
	cfg.External.Google.Enabled = true

	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "brawler"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "brawler", projectThemeFileName), []byte(`brand:
  product_name: Themed Product
copy:
  invite:
    intro: Project specific invite body copy.
`), 0o644))
	cfg.Mailer.Templates.ProjectDir = base

	user, err := models.NewUser("", "override@example.com", "", "authenticated", nil, uuid.Must(uuid.NewV4()), projectID)
	require.NoError(t, err)
	user.ConfirmationToken = "override-token"

	otherProject := uuid.Must(uuid.NewV4())
	other, err := models.NewUser("", "default@example.com", "", "authenticated", nil, uuid.Must(uuid.NewV4()), otherProject)
	require.NoError(t, err)
	other.ConfirmationToken = "default-token"

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "BRAWLER", nil
	}

	external, err := url.Parse("http://localhost:9999")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/invite", nil)

	projectMail := &recordingMailClient{}
	require.NoError(t, New(cfg, projectMail, cache).InviteMail(req, user, "otp", "https://dashboard.insidec.org/auth/callback", external))
	require.Contains(t, projectMail.body, "Themed Product")
	require.Contains(t, projectMail.body, "Project specific invite body copy.")
	require.Contains(t, projectMail.body, "invite_token=override-token")

	defaultMail := &recordingMailClient{}
	require.NoError(t, New(cfg, defaultMail, NewCache()).InviteMail(req, other, "otp", "https://dashboard.insidec.org/auth/callback", external))
	require.Contains(t, defaultMail.body, "Accept invitation")
	require.NotContains(t, defaultMail.body, "Themed Product")
}

func TestEmbeddedProductThemesLoadFromRepo(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "templates")
	for _, slug := range []string{"brawler", "zion", "iec"} {
		raw, err := os.ReadFile(filepath.Join(dir, slug, projectThemeFileName))
		require.NoError(t, err, slug)
		var partial ProjectTheme
		require.NoError(t, yaml.Unmarshal(raw, &partial), slug)
		require.NotEmpty(t, partial.Brand.ProductName, slug)
	}
}

func TestOAuthInviteAcceptURL(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	cfg.External.Email.Enabled = false
	cfg.External.Google.Enabled = true
	cfg.External.Github.Enabled = true

	external, err := url.Parse("http://localhost:9999")
	require.NoError(t, err)

	acceptURL, ok := OAuthInviteAcceptURL(cfg, "token-hash", "https://dashboard.example.com/auth/callback", external)
	require.True(t, ok)
	require.Contains(t, acceptURL, "/accept-invite?")
	require.Contains(t, acceptURL, "invite_token=token-hash")
	require.Contains(t, acceptURL, "redirect_to=")
}
