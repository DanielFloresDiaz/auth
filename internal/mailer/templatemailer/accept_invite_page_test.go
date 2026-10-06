package templatemailer

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	authtemplates "github.com/supabase/auth/templates"
	"gopkg.in/yaml.v3"
)

func TestDefaultAcceptInviteTemplateParses(t *testing.T) {
	_, err := template.New("accept_invite").Funcs(themedTemplateFuncs).Parse(authtemplates.DefaultAcceptInviteHTML)
	require.NoError(t, err)
	require.Contains(t, authtemplates.DefaultAcceptInviteHTML, "{{ range .Providers }}")
	require.Contains(t, authtemplates.DefaultAcceptInviteHTML, "{{ css .Theme.Colors.PageBg }}")
}

func TestBrawlerThemeAcceptInviteRenders(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "templates", "brawler", projectThemeFileName))
	require.NoError(t, err)
	var partial ProjectTheme
	require.NoError(t, yaml.Unmarshal(raw, &partial))

	tpl := defaultAcceptInviteTemplateBody()
	var buf bytes.Buffer
	require.NoError(t, tpl.Execute(&buf, AcceptInviteTemplateData{
		Email: "dev@example.com",
		Theme: mergeProjectTheme(DefaultProjectTheme(), partial),
	}))
	require.Contains(t, buf.String(), "Brawler")
	require.Contains(t, buf.String(), "#ff8000")
}
