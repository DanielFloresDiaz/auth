package templatemailer

import (
	"github.com/supabase/auth/internal/conf"
)

type oauthInvitePresentation struct {
	name    string
	enabled func(*conf.GlobalConfiguration) bool
	label   string
}

func oauthInvitePresentationEntry(name, label string, enabled func(*conf.GlobalConfiguration) bool) oauthInvitePresentation {
	return oauthInvitePresentation{
		name:    name,
		enabled: enabled,
		label:   label,
	}
}

// oauthInvitePresentationRegistry lists OAuth providers that may appear on the accept-invite page.
// Order is the display order when multiple providers are enabled.
var oauthInvitePresentationRegistry = []oauthInvitePresentation{
	oauthInvitePresentationEntry("google", "Sign in with Google", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Google.Enabled }),
	oauthInvitePresentationEntry("github", "Sign in with GitHub", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Github.Enabled }),
	oauthInvitePresentationEntry("apple", "Sign in with Apple", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Apple.Enabled }),
	oauthInvitePresentationEntry("azure", "Sign in with Azure", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Azure.Enabled }),
	oauthInvitePresentationEntry("bitbucket", "Sign in with Bitbucket", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Bitbucket.Enabled }),
	oauthInvitePresentationEntry("discord", "Sign in with Discord", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Discord.Enabled }),
	oauthInvitePresentationEntry("facebook", "Sign in with Facebook", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Facebook.Enabled }),
	oauthInvitePresentationEntry("figma", "Sign in with Figma", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Figma.Enabled }),
	oauthInvitePresentationEntry("fly", "Sign in with Fly.io", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Fly.Enabled }),
	oauthInvitePresentationEntry("gitlab", "Sign in with GitLab", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Gitlab.Enabled }),
	oauthInvitePresentationEntry("kakao", "Sign in with Kakao", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Kakao.Enabled }),
	oauthInvitePresentationEntry("keycloak", "Sign in with Keycloak", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Keycloak.Enabled }),
	oauthInvitePresentationEntry("linkedin", "Sign in with LinkedIn", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Linkedin.Enabled }),
	oauthInvitePresentationEntry("linkedin_oidc", "Sign in with LinkedIn", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.LinkedinOIDC.Enabled }),
	oauthInvitePresentationEntry("notion", "Sign in with Notion", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Notion.Enabled }),
	oauthInvitePresentationEntry("snapchat", "Sign in with Snapchat", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Snapchat.Enabled }),
	oauthInvitePresentationEntry("spotify", "Sign in with Spotify", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Spotify.Enabled }),
	oauthInvitePresentationEntry("slack", "Sign in with Slack", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Slack.Enabled }),
	oauthInvitePresentationEntry("slack_oidc", "Sign in with Slack", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.SlackOIDC.Enabled }),
	oauthInvitePresentationEntry("twitch", "Sign in with Twitch", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Twitch.Enabled }),
	oauthInvitePresentationEntry("twitter", "Sign in with Twitter", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Twitter.Enabled }),
	oauthInvitePresentationEntry("vercel_marketplace", "Sign in with Vercel", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.VercelMarketplace.Enabled }),
	oauthInvitePresentationEntry("workos", "Sign in with WorkOS", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.WorkOS.Enabled }),
	oauthInvitePresentationEntry("zoom", "Sign in with Zoom", func(cfg *conf.GlobalConfiguration) bool { return cfg.External.Zoom.Enabled }),
}
