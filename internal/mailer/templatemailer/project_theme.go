package templatemailer

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/observability"
	"gopkg.in/yaml.v3"
)

// ProjectTheme parameterizes shared HTML templates (colors, fonts, branding, copy).
type ProjectTheme struct {
	ColorScheme string             `yaml:"color_scheme"`
	Colors      ProjectThemeColors `yaml:"colors"`
	Fonts       ProjectThemeFonts  `yaml:"fonts"`
	Brand       ProjectThemeBrand  `yaml:"brand"`
	Copy        ProjectThemeCopy   `yaml:"copy"`
}

type ProjectThemeColors struct {
	PageBg         string `yaml:"page_bg"`
	CardBg         string `yaml:"card_bg"`
	CardBorder     string `yaml:"card_border"`
	Text           string `yaml:"text"`
	Muted          string `yaml:"muted"`
	Subtle         string `yaml:"subtle"`
	Accent         string `yaml:"accent"`
	AccentBar      string `yaml:"accent_bar"`
	EmailBg        string `yaml:"email_bg"`
	EmailBorder    string `yaml:"email_border"`
	Eyebrow        string `yaml:"eyebrow"`
	StepAccent     string `yaml:"step_accent"`
	ButtonBg       string `yaml:"button_bg"`
	ButtonText     string `yaml:"button_text"`
	Link           string `yaml:"link"`
	FooterMuted    string `yaml:"footer_muted"`
	DividerBorder  string `yaml:"divider_border"`
	ProviderBg     string `yaml:"provider_bg"`
	ProviderFg     string `yaml:"provider_fg"`
	ProviderBorder string `yaml:"provider_border"`
	CardShadow     string `yaml:"card_shadow"`
	ProviderHover  string `yaml:"provider_hover_shadow"`
}

type ProjectThemeFonts struct {
	Eyebrow string `yaml:"eyebrow"`
	Heading string `yaml:"heading"`
	Body    string `yaml:"body"`
}

type ProjectThemeBrand struct {
	ProductName string `yaml:"product_name"`
	Tagline     string `yaml:"tagline"`
}

type ProjectThemeCopy struct {
	Invite                InviteEmailCopy                `yaml:"invite"`
	AcceptInvite          AcceptInvitePageCopy           `yaml:"accept_invite"`
	WhitelistConfirmation WhitelistConfirmationEmailCopy `yaml:"whitelist_confirmation"`
	WhitelistConfirmed    WhitelistConfirmedPageCopy     `yaml:"whitelist_confirmed"`
}

type InviteEmailCopy struct {
	Title        string `yaml:"title"`
	Preheader    string `yaml:"preheader"`
	Eyebrow      string `yaml:"eyebrow"`
	Headline     string `yaml:"headline"`
	Intro        string `yaml:"intro"`
	EmailLabel   string `yaml:"email_label"`
	StepsHeading string `yaml:"steps_heading"`
	Step1        string `yaml:"step1"`
	Step2        string `yaml:"step2"`
	Step3        string `yaml:"step3"`
	CtaLabel     string `yaml:"cta_label"`
	LinkFallback string `yaml:"link_fallback"`
	Footer       string `yaml:"footer"`
}

type AcceptInvitePageCopy struct {
	Title            string `yaml:"title"`
	Eyebrow          string `yaml:"eyebrow"`
	Headline         string `yaml:"headline"`
	Lead             string `yaml:"lead"`
	EmailLabel       string `yaml:"email_label"`
	Step1            string `yaml:"step1"`
	Step2            string `yaml:"step2"`
	Step3            string `yaml:"step3"`
	ProvidersHeading string `yaml:"providers_heading"`
	Footer           string `yaml:"footer"`
}

type WhitelistConfirmationEmailCopy struct {
	Title        string `yaml:"title"`
	Preheader    string `yaml:"preheader"`
	Eyebrow      string `yaml:"eyebrow"`
	Headline     string `yaml:"headline"`
	Intro        string `yaml:"intro"`
	EmailLabel   string `yaml:"email_label"`
	StepsHeading string `yaml:"steps_heading"`
	Step1        string `yaml:"step1"`
	Step2        string `yaml:"step2"`
	Step3        string `yaml:"step3"`
	CtaLabel     string `yaml:"cta_label"`
	LinkFallback string `yaml:"link_fallback"`
	Footer       string `yaml:"footer"`
}

type WhitelistConfirmedPageCopy struct {
	Title      string `yaml:"title"`
	Eyebrow    string `yaml:"eyebrow"`
	Headline   string `yaml:"headline"`
	Lead       string `yaml:"lead"`
	EmailLabel string `yaml:"email_label"`
	Step1      string `yaml:"step1"`
	Step2      string `yaml:"step2"`
	Step3      string `yaml:"step3"`
	Footer     string `yaml:"footer"`
}

type projectThemeCacheEntry struct {
	modUnix int64
	theme   ProjectTheme
	invalid bool
}

// DefaultProjectTheme matches the generic root templates under templates/*.html.
func DefaultProjectTheme() ProjectTheme {
	return ProjectTheme{
		ColorScheme: "light",
		Colors: ProjectThemeColors{
			PageBg:         "#eef0f3",
			CardBg:         "#ffffff",
			CardBorder:     "#d1d5db",
			Text:           "#111827",
			Muted:          "#4b5563",
			Subtle:         "#6b7280",
			Accent:         "#111827",
			AccentBar:      "#111827",
			EmailBg:        "#f9fafb",
			EmailBorder:    "#e5e7eb",
			Eyebrow:        "#6b7280",
			StepAccent:     "#111827",
			ButtonBg:       "#111827",
			ButtonText:     "#ffffff",
			Link:           "#111827",
			FooterMuted:    "#9ca3af",
			DividerBorder:  "#e5e7eb",
			ProviderBg:     "#ffffff",
			ProviderFg:     "#1f2937",
			ProviderBorder: "rgba(0, 0, 0, 0.1)",
			CardShadow:     "0 4px 24px rgba(15, 23, 42, 0.06)",
			ProviderHover:  "0 6px 16px rgba(0, 0, 0, 0.1)",
		},
		Fonts: ProjectThemeFonts{
			Eyebrow: "'Source Sans 3',Segoe UI,Helvetica,Arial,sans-serif",
			Heading: "Archivo,'Source Sans 3',Segoe UI,Helvetica,Arial,sans-serif",
			Body:    "'Source Sans 3',Segoe UI,Helvetica,Arial,sans-serif",
		},
		Copy: ProjectThemeCopy{
			Invite: InviteEmailCopy{
				Title:        "You have been invited to create an account",
				Preheader:    "You were invited to create an account for {{ .Email }}. Open this email to accept the invitation and sign in with Google or GitHub.",
				Eyebrow:      "Account invitation",
				Headline:     "You&rsquo;re invited to join",
				Intro:        "Someone on your team invited you to create an account. Use the button below to continue&mdash;you&rsquo;ll choose Google or GitHub on the next screen.",
				EmailLabel:   "Invited email",
				StepsHeading: "What happens next",
				Step1:        "Open the secure invitation link below.",
				Step2:        "Sign in with Google or GitHub using the same email address shown above.",
				Step3:        "You&rsquo;ll be returned to the application once your account is ready.",
				CtaLabel:     "Accept invitation",
				LinkFallback: "If the button doesn&rsquo;t work, copy and paste this link into your browser:",
				Footer:       "This message was sent because {{ .Email }} was invited to create an account. The link is intended only for that address. If you weren&rsquo;t expecting this invitation, you can safely ignore this email&mdash;no account will be created without your action.",
			},
			AcceptInvite: AcceptInvitePageCopy{
				Title:            "Accept invitation",
				Eyebrow:          "Secure sign-in",
				Headline:         "Finish creating your account",
				Lead:             "You opened a valid invitation link. Choose how you&rsquo;d like to sign in&mdash;your provider account must use the same email address we invited.",
				EmailLabel:       "Invited email",
				Step1:            "Pick Google or GitHub below.",
				Step2:            "Approve access when your provider asks.",
				Step3:            "We&rsquo;ll create your account and send you back to the app.",
				ProvidersHeading: "Continue with",
				Footer:           "This page is only for completing your invitation. If you didn&rsquo;t request access, you can close this window.",
			},
			WhitelistConfirmation: WhitelistConfirmationEmailCopy{
				Title:        "Confirm your access request",
				Preheader:    "Confirm the access request for {{ .Email }}. A project admin reviews it only after you open this link.",
				Eyebrow:      "Access request",
				Headline:     "Confirm your access request",
				Intro:        "Someone used this email address to request access. Use the button below to confirm it so a project admin can review the request.",
				EmailLabel:   "Requested email",
				StepsHeading: "What happens next",
				Step1:        "Open the confirmation link below.",
				Step2:        "Your request is recorded for a project admin to review.",
				Step3:        "If it is accepted, you will get a separate email to continue with Google or GitHub.",
				CtaLabel:     "Confirm your access request",
				LinkFallback: "If the button doesn&rsquo;t work, copy and paste this link into your browser:",
				Footer:       "This message was sent because {{ .Email }} was used to request access. The link is intended only for that address. If you did not ask for this, you can ignore this email&mdash;no request will be recorded without your confirmation.",
			},
			WhitelistConfirmed: WhitelistConfirmedPageCopy{
				Title:      "Access request recorded",
				Eyebrow:    "Access request",
				Headline:   "Access request recorded",
				Lead:       "Your access request was recorded. A project admin reviews it from here.",
				EmailLabel: "Requested email",
				Step1:      "Your request is now waiting for a project admin.",
				Step2:      "If it is accepted, you get an email to continue with Google or GitHub.",
				Step3:      "You can close this page.",
				Footer:     "This page confirms the request for {{ .Email }} only. If you did not ask for access, you can close this window.",
			},
		},
	}
}

func mergeProjectTheme(base, override ProjectTheme) ProjectTheme {
	out := base
	if override.ColorScheme != "" {
		out.ColorScheme = override.ColorScheme
	}
	out.Colors = mergeThemeColors(base.Colors, override.Colors)
	out.Fonts = mergeThemeFonts(base.Fonts, override.Fonts)
	out.Brand = mergeThemeBrand(base.Brand, override.Brand)
	out.Copy = mergeThemeCopy(base.Copy, override.Copy)
	return out
}

func mergeThemeColors(base, o ProjectThemeColors) ProjectThemeColors {
	return ProjectThemeColors{
		PageBg:         pickString(o.PageBg, base.PageBg),
		CardBg:         pickString(o.CardBg, base.CardBg),
		CardBorder:     pickString(o.CardBorder, base.CardBorder),
		Text:           pickString(o.Text, base.Text),
		Muted:          pickString(o.Muted, base.Muted),
		Subtle:         pickString(o.Subtle, base.Subtle),
		Accent:         pickString(o.Accent, base.Accent),
		AccentBar:      pickString(o.AccentBar, base.AccentBar),
		EmailBg:        pickString(o.EmailBg, base.EmailBg),
		EmailBorder:    pickString(o.EmailBorder, base.EmailBorder),
		Eyebrow:        pickString(o.Eyebrow, base.Eyebrow),
		StepAccent:     pickString(o.StepAccent, base.StepAccent),
		ButtonBg:       pickString(o.ButtonBg, base.ButtonBg),
		ButtonText:     pickString(o.ButtonText, base.ButtonText),
		Link:           pickString(o.Link, base.Link),
		FooterMuted:    pickString(o.FooterMuted, base.FooterMuted),
		DividerBorder:  pickString(o.DividerBorder, base.DividerBorder),
		ProviderBg:     pickString(o.ProviderBg, base.ProviderBg),
		ProviderFg:     pickString(o.ProviderFg, base.ProviderFg),
		ProviderBorder: pickString(o.ProviderBorder, base.ProviderBorder),
		CardShadow:     pickString(o.CardShadow, base.CardShadow),
		ProviderHover:  pickString(o.ProviderHover, base.ProviderHover),
	}
}

func mergeThemeFonts(base, o ProjectThemeFonts) ProjectThemeFonts {
	return ProjectThemeFonts{
		Eyebrow: pickString(o.Eyebrow, base.Eyebrow),
		Heading: pickString(o.Heading, base.Heading),
		Body:    pickString(o.Body, base.Body),
	}
}

func mergeThemeBrand(base, o ProjectThemeBrand) ProjectThemeBrand {
	return ProjectThemeBrand{
		ProductName: pickString(o.ProductName, base.ProductName),
		Tagline:     pickString(o.Tagline, base.Tagline),
	}
}

func mergeThemeCopy(base, o ProjectThemeCopy) ProjectThemeCopy {
	return ProjectThemeCopy{
		Invite:                mergeInviteCopy(base.Invite, o.Invite),
		AcceptInvite:          mergeAcceptInviteCopy(base.AcceptInvite, o.AcceptInvite),
		WhitelistConfirmation: mergeWhitelistConfirmationCopy(base.WhitelistConfirmation, o.WhitelistConfirmation),
		WhitelistConfirmed:    mergeWhitelistConfirmedCopy(base.WhitelistConfirmed, o.WhitelistConfirmed),
	}
}

func mergeInviteCopy(base, o InviteEmailCopy) InviteEmailCopy {
	return InviteEmailCopy{
		Title:        pickString(o.Title, base.Title),
		Preheader:    pickString(o.Preheader, base.Preheader),
		Eyebrow:      pickString(o.Eyebrow, base.Eyebrow),
		Headline:     pickString(o.Headline, base.Headline),
		Intro:        pickString(o.Intro, base.Intro),
		EmailLabel:   pickString(o.EmailLabel, base.EmailLabel),
		StepsHeading: pickString(o.StepsHeading, base.StepsHeading),
		Step1:        pickString(o.Step1, base.Step1),
		Step2:        pickString(o.Step2, base.Step2),
		Step3:        pickString(o.Step3, base.Step3),
		CtaLabel:     pickString(o.CtaLabel, base.CtaLabel),
		LinkFallback: pickString(o.LinkFallback, base.LinkFallback),
		Footer:       pickString(o.Footer, base.Footer),
	}
}

func mergeAcceptInviteCopy(base, o AcceptInvitePageCopy) AcceptInvitePageCopy {
	return AcceptInvitePageCopy{
		Title:            pickString(o.Title, base.Title),
		Eyebrow:          pickString(o.Eyebrow, base.Eyebrow),
		Headline:         pickString(o.Headline, base.Headline),
		Lead:             pickString(o.Lead, base.Lead),
		EmailLabel:       pickString(o.EmailLabel, base.EmailLabel),
		Step1:            pickString(o.Step1, base.Step1),
		Step2:            pickString(o.Step2, base.Step2),
		Step3:            pickString(o.Step3, base.Step3),
		ProvidersHeading: pickString(o.ProvidersHeading, base.ProvidersHeading),
		Footer:           pickString(o.Footer, base.Footer),
	}
}

func mergeWhitelistConfirmationCopy(base, o WhitelistConfirmationEmailCopy) WhitelistConfirmationEmailCopy {
	return WhitelistConfirmationEmailCopy{
		Title:        pickString(o.Title, base.Title),
		Preheader:    pickString(o.Preheader, base.Preheader),
		Eyebrow:      pickString(o.Eyebrow, base.Eyebrow),
		Headline:     pickString(o.Headline, base.Headline),
		Intro:        pickString(o.Intro, base.Intro),
		EmailLabel:   pickString(o.EmailLabel, base.EmailLabel),
		StepsHeading: pickString(o.StepsHeading, base.StepsHeading),
		Step1:        pickString(o.Step1, base.Step1),
		Step2:        pickString(o.Step2, base.Step2),
		Step3:        pickString(o.Step3, base.Step3),
		CtaLabel:     pickString(o.CtaLabel, base.CtaLabel),
		LinkFallback: pickString(o.LinkFallback, base.LinkFallback),
		Footer:       pickString(o.Footer, base.Footer),
	}
}

func mergeWhitelistConfirmedCopy(base, o WhitelistConfirmedPageCopy) WhitelistConfirmedPageCopy {
	return WhitelistConfirmedPageCopy{
		Title:      pickString(o.Title, base.Title),
		Eyebrow:    pickString(o.Eyebrow, base.Eyebrow),
		Headline:   pickString(o.Headline, base.Headline),
		Lead:       pickString(o.Lead, base.Lead),
		EmailLabel: pickString(o.EmailLabel, base.EmailLabel),
		Step1:      pickString(o.Step1, base.Step1),
		Step2:      pickString(o.Step2, base.Step2),
		Step3:      pickString(o.Step3, base.Step3),
		Footer:     pickString(o.Footer, base.Footer),
	}
}

func pickString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func (o *Cache) themeForProject(ctx context.Context, cfg *conf.GlobalConfiguration, projectID string) ProjectTheme {
	def := DefaultProjectTheme()
	if invalidProjectID(projectID) {
		return def
	}

	projectDir := strings.TrimSpace(cfg.Mailer.Templates.ProjectDir)
	if projectDir == "" {
		return def
	}

	slug := o.projectSlug(ctx, projectID)
	if slug == "" {
		return def
	}

	path, ok := projectTemplatePath(projectDir, slug, projectThemeFileName)
	if !ok {
		return def
	}

	info, err := os.Stat(path)
	if err != nil {
		return def
	}
	modUnix := info.ModTime().UnixNano()

	o.rw.RLock()
	cached, hit := o.projectThemes[projectID]
	o.rw.RUnlock()
	if hit && cached.modUnix == modUnix {
		if cached.invalid {
			return def
		}
		return cached.theme
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		o.storeProjectTheme(projectID, &projectThemeCacheEntry{modUnix: modUnix, invalid: true})
		logProjectThemeError(ctx, projectID, err)
		return def
	}

	var partial ProjectTheme
	if err := yaml.Unmarshal(raw, &partial); err != nil {
		o.storeProjectTheme(projectID, &projectThemeCacheEntry{modUnix: modUnix, invalid: true})
		logProjectThemeError(ctx, projectID, err)
		return def
	}

	merged := mergeProjectTheme(def, partial)
	o.storeProjectTheme(projectID, &projectThemeCacheEntry{modUnix: modUnix, theme: merged})
	return merged
}

func (o *Cache) storeProjectTheme(projectID string, entry *projectThemeCacheEntry) {
	o.rw.Lock()
	defer o.rw.Unlock()
	if o.projectThemes == nil {
		o.projectThemes = make(map[string]*projectThemeCacheEntry)
	}
	o.projectThemes[projectID] = entry
}

func logProjectThemeError(ctx context.Context, projectID string, err error) {
	logger := logrus.NewEntry(logrus.StandardLogger())
	if ctx != nil {
		logger = observability.GetLogEntryFromContext(ctx).Entry
	}
	logger.WithError(err).WithField("project_id", projectID).Warn("project theme load failed; using default theme")
}
