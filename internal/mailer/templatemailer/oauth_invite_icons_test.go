package templatemailer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOAuthInviteIconSVGKnownProviders(t *testing.T) {
	for _, name := range []string{
		"google", "github", "apple", "azure", "bitbucket", "discord", "facebook", "figma",
		"fly", "gitlab", "kakao", "keycloak", "linkedin", "linkedin_oidc", "notion",
		"snapchat", "spotify", "slack", "slack_oidc", "twitch", "twitter",
		"vercel_marketplace", "workos", "zoom",
	} {
		icon := string(oauthInviteIconSVG(name))
		require.Contains(t, icon, "svg", "provider %q should have svg icon", name)
		require.NotContains(t, icon, "provider-icon--generic", "provider %q should not fall back to generic", name)
	}
}

func TestOAuthInviteIconSVGUnknownProviderFallsBack(t *testing.T) {
	icon := string(oauthInviteIconSVG("unknown_provider"))
	require.Contains(t, icon, "provider-icon--generic")
	require.True(t, strings.Contains(icon, "U") || strings.Contains(icon, "?"))
}
