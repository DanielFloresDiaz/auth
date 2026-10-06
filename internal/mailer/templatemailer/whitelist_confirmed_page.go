package templatemailer

import (
	"bytes"
	"context"
	"html/template"
	"sync"

	authtemplates "github.com/supabase/auth/templates"
)

// WhitelistConfirmedTemplateData is passed to the page shown after a confirmation link is opened.
type WhitelistConfirmedTemplateData struct {
	Email     string
	ProjectID string
	SiteURL   string
	Theme     ProjectTheme
}

var (
	defaultWhitelistConfirmedTemplate     *template.Template
	defaultWhitelistConfirmedTemplateOnce sync.Once
)

func defaultWhitelistConfirmedTemplateBody() *template.Template {
	defaultWhitelistConfirmedTemplateOnce.Do(func() {
		defaultWhitelistConfirmedTemplate = template.Must(
			template.New("whitelist_confirmed_default").Funcs(themedTemplateFuncs).Parse(authtemplates.DefaultWhitelistConfirmedHTML),
		)
	})
	return defaultWhitelistConfirmedTemplate
}

// RenderDefaultWhitelistConfirmedPage executes the embedded confirmation page with the default theme.
func RenderDefaultWhitelistConfirmedPage(data WhitelistConfirmedTemplateData) (string, error) {
	data.Theme = DefaultProjectTheme()
	rendered, err := renderThemeCopyForWhitelistConfirmed(data.Theme, data)
	if err != nil {
		return "", err
	}
	data.Theme = rendered
	return executeWhitelistConfirmed(defaultWhitelistConfirmedTemplateBody(), data)
}

// RenderWhitelistConfirmedPage executes the confirmation page with project theme.
func (m *Mailer) RenderWhitelistConfirmedPage(ctx context.Context, projectID string, data WhitelistConfirmedTemplateData) (string, error) {
	data.Theme = m.tc.themeForProject(ctx, m.cfg, projectID)
	rendered, err := renderThemeCopyForWhitelistConfirmed(data.Theme, data)
	if err != nil {
		return "", err
	}
	data.Theme = rendered
	return executeWhitelistConfirmed(defaultWhitelistConfirmedTemplateBody(), data)
}

func executeWhitelistConfirmed(tpl *template.Template, data WhitelistConfirmedTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
