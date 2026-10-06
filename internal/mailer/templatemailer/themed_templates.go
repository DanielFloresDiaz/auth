package templatemailer

import (
	"bytes"
	htmltemplate "html/template"
	"sync"
	texttemplate "text/template"

	authtemplates "github.com/supabase/auth/templates"
)

var themedTemplateFuncs = htmltemplate.FuncMap{
	"css": func(v string) htmltemplate.CSS {
		return htmltemplate.CSS(v) // #nosec G203 -- theme.yaml supplies intentional CSS for mail templates
	},
}

var (
	defaultInviteMailTemplate     *htmltemplate.Template
	defaultInviteMailTemplateOnce sync.Once

	defaultWhitelistConfirmationMailTemplate     *htmltemplate.Template
	defaultWhitelistConfirmationMailTemplateOnce sync.Once
)

func defaultInviteMailTemplateBody() *htmltemplate.Template {
	defaultInviteMailTemplateOnce.Do(func() {
		defaultInviteMailTemplate = htmltemplate.Must(
			htmltemplate.New("invite_default").Funcs(themedTemplateFuncs).Parse(authtemplates.DefaultInviteHTML),
		)
	})
	return defaultInviteMailTemplate
}

func defaultWhitelistConfirmationMailTemplateBody() *htmltemplate.Template {
	defaultWhitelistConfirmationMailTemplateOnce.Do(func() {
		defaultWhitelistConfirmationMailTemplate = htmltemplate.Must(
			htmltemplate.New("whitelist_confirmation_default").Funcs(themedTemplateFuncs).Parse(authtemplates.DefaultWhitelistConfirmationHTML),
		)
	})
	return defaultWhitelistConfirmationMailTemplate
}

func renderCopyString(src string, data any) (string, error) {
	if src == "" || !bytes.Contains([]byte(src), []byte("{{")) {
		return src, nil
	}
	tpl, err := texttemplate.New("copy").Parse(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderThemeCopyForData(theme ProjectTheme, data map[string]any) (ProjectTheme, error) {
	var err error
	theme.Copy.Invite, err = renderInviteEmailCopy(theme.Copy.Invite, data)
	if err != nil {
		return theme, err
	}
	theme.Copy.WhitelistConfirmation, err = renderWhitelistConfirmationEmailCopy(theme.Copy.WhitelistConfirmation, data)
	if err != nil {
		return theme, err
	}
	return theme, nil
}

func renderInviteEmailCopy(c InviteEmailCopy, data any) (InviteEmailCopy, error) {
	var err error
	if c.Preheader, err = renderCopyString(c.Preheader, data); err != nil {
		return c, err
	}
	if c.Step1, err = renderCopyString(c.Step1, data); err != nil {
		return c, err
	}
	if c.Step2, err = renderCopyString(c.Step2, data); err != nil {
		return c, err
	}
	if c.Step3, err = renderCopyString(c.Step3, data); err != nil {
		return c, err
	}
	if c.Footer, err = renderCopyString(c.Footer, data); err != nil {
		return c, err
	}
	return c, nil
}

func renderWhitelistConfirmationEmailCopy(c WhitelistConfirmationEmailCopy, data any) (WhitelistConfirmationEmailCopy, error) {
	var err error
	if c.Preheader, err = renderCopyString(c.Preheader, data); err != nil {
		return c, err
	}
	if c.Step1, err = renderCopyString(c.Step1, data); err != nil {
		return c, err
	}
	if c.Step2, err = renderCopyString(c.Step2, data); err != nil {
		return c, err
	}
	if c.Step3, err = renderCopyString(c.Step3, data); err != nil {
		return c, err
	}
	if c.Footer, err = renderCopyString(c.Footer, data); err != nil {
		return c, err
	}
	return c, nil
}

func renderThemeCopyForWhitelistConfirmed(theme ProjectTheme, data WhitelistConfirmedTemplateData) (ProjectTheme, error) {
	var err error
	if theme.Copy.WhitelistConfirmed.Footer, err = renderCopyString(theme.Copy.WhitelistConfirmed.Footer, data); err != nil {
		return theme, err
	}
	return theme, nil
}
