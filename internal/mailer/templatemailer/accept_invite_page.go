package templatemailer

import (
	"bytes"
	"context"
	"html/template"
	"sync"

	authtemplates "github.com/supabase/auth/templates"
)

var (
	defaultAcceptInviteTemplate     *template.Template
	defaultAcceptInviteTemplateOnce sync.Once
)

func defaultAcceptInviteTemplateBody() *template.Template {
	defaultAcceptInviteTemplateOnce.Do(func() {
		defaultAcceptInviteTemplate = template.Must(
			template.New("accept_invite_default").Funcs(themedTemplateFuncs).Parse(authtemplates.DefaultAcceptInviteHTML),
		)
	})
	return defaultAcceptInviteTemplate
}

// RenderAcceptInvitePage executes the accept-invite HTML template with project theme.
func (m *Mailer) RenderAcceptInvitePage(ctx context.Context, projectID string, data AcceptInviteTemplateData) (string, error) {
	data.Theme = m.tc.themeForProject(ctx, m.cfg, projectID)
	tpl := defaultAcceptInviteTemplateBody()

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
