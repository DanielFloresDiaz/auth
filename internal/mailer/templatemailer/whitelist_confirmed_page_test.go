package templatemailer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
	authtemplates "github.com/supabase/auth/templates"
)

func TestDefaultWhitelistConfirmedTemplateParses(t *testing.T) {
	html, err := RenderDefaultWhitelistConfirmedPage(WhitelistConfirmedTemplateData{
		Email:     "ada@acme.com",
		ProjectID: uuid.Must(uuid.NewV4()).String(),
		SiteURL:   "https://example.netlify.com",
	})
	require.NoError(t, err)
	require.Contains(t, html, "Access request recorded")
	require.Contains(t, html, "ada@acme.com")
	require.Contains(t, authtemplates.DefaultWhitelistConfirmedHTML, "{{ .Theme.Copy.WhitelistConfirmed.Headline }}")
}

func TestWhitelistConfirmedPageProjectTheme(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "iec"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(base, "iec", projectThemeFileName), []byte(`copy:
  whitelist_confirmed:
    lead: Project recorded themed lead.
`), 0o600))
	cfg.Mailer.Templates.ProjectDir = base

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "IEC", nil
	}

	mailer := New(cfg, &recordingMailClient{}, cache)
	html, err := mailer.RenderWhitelistConfirmedPage(t.Context(), projectID.String(), WhitelistConfirmedTemplateData{
		Email:     "override@example.com",
		ProjectID: projectID.String(),
	})
	require.NoError(t, err)
	require.Contains(t, html, "Project recorded themed lead.")
	require.Contains(t, html, "override@example.com")

	plainCache := NewCache()
	plainMailer := New(cfg, &recordingMailClient{}, plainCache)
	other, err := plainMailer.RenderWhitelistConfirmedPage(t.Context(), uuid.Must(uuid.NewV4()).String(), WhitelistConfirmedTemplateData{
		Email: "default@example.com",
	})
	require.NoError(t, err)
	require.Contains(t, other, "Access request recorded")
	require.NotContains(t, other, "Project recorded themed lead.")
}

func TestWhitelistConfirmedPageDefaultThemeWithoutLookup(t *testing.T) {
	cfg := &conf.GlobalConfiguration{}
	mailer := New(cfg, &recordingMailClient{}, NewCache())
	html, err := mailer.RenderWhitelistConfirmedPage(t.Context(), uuid.Must(uuid.NewV4()).String(), WhitelistConfirmedTemplateData{
		Email: "plain@example.com",
	})
	require.NoError(t, err)
	require.Contains(t, html, "plain@example.com")
	require.Contains(t, html, "A project admin reviews it from here.")
}
