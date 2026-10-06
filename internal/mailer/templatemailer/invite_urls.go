package templatemailer

import (
	"html/template"
	"net/url"

	"github.com/supabase/auth/internal/conf"
)

// AcceptInviteProvider is one OAuth choice rendered on the accept-invite page.
type AcceptInviteProvider struct {
	Name    string
	Label   string
	URL     string
	IconSVG template.HTML
}

// AcceptInviteTemplateData is passed to accept-invite HTML templates.
type AcceptInviteTemplateData struct {
	Email     string
	SiteURL   string
	ProjectID string
	Providers []AcceptInviteProvider
	Theme     ProjectTheme
}

// OAuthInviteProvider is kept for callers that only need provider names and labels.
type OAuthInviteProvider struct {
	Name  string
	Label string
}

// BuildAcceptInviteProviders returns enabled OAuth providers with authorize URLs and presentation metadata.
func BuildAcceptInviteProviders(cfg *conf.GlobalConfiguration, inviteToken, referrer string, externalURL *url.URL) []AcceptInviteProvider {
	if cfg == nil || cfg.External.Email.Enabled || externalURL == nil || inviteToken == "" {
		return nil
	}

	var providers []AcceptInviteProvider
	for _, entry := range oauthInvitePresentationRegistry {
		if entry.enabled == nil || !entry.enabled(cfg) {
			continue
		}
		authorizeURL, ok := OAuthInviteAuthorizeURL(cfg, entry.name, inviteToken, referrer, externalURL)
		if !ok {
			continue
		}
		providers = append(providers, AcceptInviteProvider{
			Name:    entry.name,
			Label:   entry.label,
			URL:     authorizeURL,
			IconSVG: oauthInviteIconSVG(entry.name),
		})
	}
	return providers
}

// OAuthInviteProviders lists enabled OAuth providers available for invite acceptance.
func OAuthInviteProviders(cfg *conf.GlobalConfiguration) []OAuthInviteProvider {
	if cfg == nil || cfg.External.Email.Enabled {
		return nil
	}
	var out []OAuthInviteProvider
	for _, entry := range oauthInvitePresentationRegistry {
		if entry.enabled != nil && entry.enabled(cfg) {
			out = append(out, OAuthInviteProvider{Name: entry.name, Label: entry.label})
		}
	}
	return out
}

// OAuthInviteAcceptURL returns the provider-neutral invite landing URL when OAuth
// invites are used (email signup disabled and at least one OAuth provider enabled).
func OAuthInviteAcceptURL(cfg *conf.GlobalConfiguration, tokenHash, referrer string, externalURL *url.URL) (string, bool) {
	if !oauthInviteFlowEnabled(cfg, externalURL, tokenHash) {
		return "", false
	}

	query := url.Values{}
	query.Set("invite_token", tokenHash)
	if referrer != "" {
		query.Set("redirect_to", referrer)
	}
	ref := &url.URL{Path: "/accept-invite", RawQuery: query.Encode()}
	return externalURL.ResolveReference(ref).String(), true
}

// OAuthInviteAuthorizeURL builds /authorize for a specific provider and invite token.
func OAuthInviteAuthorizeURL(cfg *conf.GlobalConfiguration, provider, tokenHash, referrer string, externalURL *url.URL) (string, bool) {
	if externalURL == nil || tokenHash == "" || !oauthInviteProviderEnabled(cfg, provider) {
		return "", false
	}

	query := url.Values{}
	query.Set("provider", provider)
	query.Set("invite_token", tokenHash)
	if referrer != "" {
		query.Set("redirect_to", referrer)
	}
	ref := &url.URL{Path: "/authorize", RawQuery: query.Encode()}
	return externalURL.ResolveReference(ref).String(), true
}

func oauthInviteProviderEnabled(cfg *conf.GlobalConfiguration, provider string) bool {
	if cfg == nil || cfg.External.Email.Enabled {
		return false
	}
	for _, entry := range oauthInvitePresentationRegistry {
		if entry.name != provider {
			continue
		}
		return entry.enabled != nil && entry.enabled(cfg)
	}
	return false
}

func oauthInviteFlowEnabled(cfg *conf.GlobalConfiguration, externalURL *url.URL, tokenHash string) bool {
	if cfg == nil || cfg.External.Email.Enabled || externalURL == nil || tokenHash == "" {
		return false
	}
	return len(BuildAcceptInviteProviders(cfg, tokenHash, "", externalURL)) > 0
}
