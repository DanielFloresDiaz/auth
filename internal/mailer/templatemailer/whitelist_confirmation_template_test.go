package templatemailer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
)

func TestWhitelistConfirmationMailUsesDefaultTemplate(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	cfg.SiteURL = "https://example.netlify.com"

	recorder := &recordingMailClient{}
	req := httptest.NewRequest(http.MethodPost, "/whitelist", nil)
	require.NoError(t, New(cfg, recorder, NewCache()).WhitelistConfirmationMail(
		req,
		uuid.Must(uuid.NewV4()).String(),
		"ada@acme.com",
		"http://localhost:9999/whitelist/confirm?token=hashed-token",
	))
	require.Contains(t, recorder.body, "Confirm your access request")
	require.Contains(t, recorder.body, "ada@acme.com")
	require.Contains(t, recorder.body, "http://localhost:9999/whitelist/confirm?token=hashed-token")
}

func TestWhitelistConfirmationMailProjectTheme(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	cfg.SiteURL = "https://example.netlify.com"

	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "zion"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "zion", projectThemeFileName), []byte(`brand:
  product_name: Zion
colors:
  accent: "#E6A815"
copy:
  whitelist_confirmation:
    intro: Project specific access request copy.
`), 0o644))
	cfg.Mailer.Templates.ProjectDir = base

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "ZION", nil
	}

	req := httptest.NewRequest(http.MethodPost, "/whitelist", nil)
	projectMail := &recordingMailClient{}
	require.NoError(t, New(cfg, projectMail, cache).WhitelistConfirmationMail(
		req, projectID.String(), "override@example.com", "http://localhost:9999/whitelist/confirm?token=override-token",
	))
	require.Contains(t, projectMail.body, "Zion")
	require.Contains(t, projectMail.body, "Project specific access request copy.")
	require.Contains(t, projectMail.body, "override-token")

	defaultMail := &recordingMailClient{}
	require.NoError(t, New(cfg, defaultMail, NewCache()).WhitelistConfirmationMail(
		req, uuid.Must(uuid.NewV4()).String(), "default@example.com", "http://localhost:9999/whitelist/confirm?token=default-token",
	))
	require.Contains(t, defaultMail.body, "Confirm your access request")
	require.NotContains(t, defaultMail.body, "Project specific access request copy.")
}
