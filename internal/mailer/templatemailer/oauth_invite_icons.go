package templatemailer

import (
	"html/template"
	"strings"
)

const oauthIconAttrs = ` width="20" height="20" aria-hidden="true" focusable="false"`

// Icons match solomon-frontend-lib LoginForm where applicable (Google, GitHub).
var (
	googleOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 48 48"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/></svg>`)

	githubOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 98 96"><path fill="currentColor" fill-rule="evenodd" clip-rule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.125 7.523 5.125 4.367 7.534 11.464 5.361 14.251 4.067.427-3.068 1.715-5.361 3.114-6.603-10.853-1.147-22.26-5.373-22.26-24.089 0-5.361 1.926-9.689 5.039-13.11-.427-1.223-2.184-6.276.486-13.029 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.753.97 11.806.485 13.029 3.155 3.421 5.039 7.75 5.039 13.11 0 18.715-11.495 22.827-22.443 24.029 1.78 1.548 3.316 4.481 3.316 9.126 0 6.603-.08 11.926-.08 13.533 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z"/></svg>`)

	appleOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="currentColor" d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/></svg>`)

	azureOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#F25022" d="M1 1h10.5v10.5H1z"/><path fill="#7FBA00" d="M12.5 1H23v10.5H12.5z"/><path fill="#00A4EF" d="M1 12.5h10.5V23H1z"/><path fill="#FFB900" d="M12.5 12.5H23V23H12.5z"/></svg>`)

	bitbucketOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#2684FF" d="M.778 1.213a.768.768 0 0 0-.768.892l3.263 19.81c.084.5.515.868 1.022.873h14.957a.772.772 0 0 0 .77-.646l3.27-20.03a.768.768 0 0 0-.768-.891L.778 1.213zm15.014 14.932H8.208l-1.449-8.652 9.033-.008z"/></svg>`)

	discordOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#5865F2" d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.419 0 1.334-.956 2.419-2.157 2.419zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.419 0 1.334-.946 2.419-2.157 2.419z"/></svg>`)

	facebookOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#1877F2" d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>`)

	figmaOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#0ACF83" d="M8 24a4 4 0 0 0 4-4v-4H8a4 4 0 0 0 0 8z"/><path fill="#A259FF" d="M4 12a4 4 0 0 1 4-4h4v8H8a4 4 0 0 1-4-4z"/><path fill="#F24E1E" d="M4 4a4 4 0 0 1 4-4h4v8H8a4 4 0 0 1-4-4z"/><path fill="#FF7262" d="M12 0h4a4 4 0 0 1 0 8h-4V0z"/><path fill="#1ABCFE" d="M20 12a4 4 0 0 1-4 4h-4v-8h4a4 4 0 0 1 4 4z"/></svg>`)

	flyOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#7C3AED" d="M12 2 2 7l10 5 10-5-10-5zm0 7L2 14l10 5 10-5-10-5zm0 7-8 4 8 4 8-4-8-4z"/></svg>`)

	gitlabOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#E24329" d="m23.955 13.587-1.342-4.135-2.664-8.17a.955.955 0 0 0-.91-.664H4.961a.955.955 0 0 0-.909.664L1.388 9.452.046 13.587a.96.96 0 0 0 .331 1.075L12 23.054l11.623-8.392a.96.96 0 0 0 .332-1.075z"/><path fill="#FC6D26" d="M12 23.054 7.873 9.288h8.254L12 23.054z"/><path fill="#FCA326" d="M12 23.054 16.127 9.288H23.62L12 23.054zM.38 9.288 7.873 9.288 12 23.054 .38 9.288z"/></svg>`)

	kakaoOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#FEE500" d="M12 3C6.477 3 2 6.477 2 10.5c0 2.178 1.16 4.122 2.978 5.418l-.768 2.808a.5.5 0 0 0 .728.556L9.6 17.9C10.378 18.033 11.18 18.1 12 18.1c5.523 0 10-3.477 10-7.6S17.523 3 12 3z"/></svg>`)

	keycloakOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#4D4D4D" d="M12 2 4 6v6c0 5 3.4 9.7 8 11 4.6-1.3 8-6 8-11V6l-8-4zm0 2.2 6 3v4.8c0 3.8-2.5 7.4-6 8.6-3.5-1.2-6-4.8-6-8.6V7.2l6-3z"/><path fill="#00B8E3" d="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8z"/></svg>`)

	linkedinOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#0A66C2" d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>`)

	notionOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="currentColor" d="M4.459 4.208c.746-.606 1.026-.56 2.428-.466l13.215.793c.892.047 1.171.23 1.171.793v14.834c0 .84-.093 1.335-.746 1.888l-2.428 1.494c-.746.56-1.494.466-2.428.466H5.832c-.933 0-1.494-.14-2.008-.653L1.416 19.94c-.513-.513-.653-1.075-.653-2.008V5.461c0-.933.14-1.494.653-2.008l2.408-1.494c.513-.513 1.075-.653 2.008-.653h.437z"/><path fill="#fff" d="M7.08 6.22v11.56l2.008 1.075V7.295L7.08 6.22zm4.112 11.56 2.008 1.075V7.295l-2.008-1.075v11.56zm4.112 0 2.008 1.075V7.295l-2.008-1.075v11.56z"/></svg>`)

	snapchatOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#FFFC00" stroke="#111" stroke-width=".5" d="M12 2c2.8 0 5 2.2 5.2 5.1 2.5.3 4.3 2.4 4.3 4.9 0 .9-.2 1.7-.6 2.4.3.2.7.3 1.1.3 1.1 0 2-.9 2-2 0-.4-.1-.8-.3-1.1.6-.4 1.1-1 1.4-1.7-1.2-.5-2.1-1.6-2.3-2.9-.4 2.5-2.5 4.4-5.1 4.4s-4.7-1.9-5.1-4.4c-.2 1.3-1.1 2.4-2.3 2.9.3.7.8 1.3 1.4 1.7-.2.3-.3.7-.3 1.1 0 1.1.9 2 2 2 .4 0 .8-.1 1.1-.3-.4.7-.6 1.5-.6 2.4 0 2.5 1.8 4.6 4.3 4.9.2 2.9 2.4 5.1 5.2 5.1z"/></svg>`)

	spotifyOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#1DB954" d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/></svg>`)

	slackOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#E01E5A" d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z"/></svg>`)

	twitchOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#9146FF" d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0 1.714 4.286v15.428h5.143V24l4.286-4.286h3.428L22.286 12V0zm14.571 11.143-3.428 3.428h-3.429l-3 3v-3H6.857V1.714h13.714Z"/></svg>`)

	twitterOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="currentColor" d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>`)

	vercelOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="currentColor" d="M12 2 2 22h20L12 2z"/></svg>`)

	workosOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#6363F1" d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm-1 5h2v8h-2V7zm0 10h2v2h-2v-2z"/></svg>`)

	zoomOAuthIconSVG = template.HTML(`<svg` + oauthIconAttrs + ` viewBox="0 0 24 24"><path fill="#2D8CFF" d="M4.5 4A2.5 2.5 0 0 0 2 6.5v11A2.5 2.5 0 0 0 4.5 20h11a2.5 2.5 0 0 0 2.5-2.5v-11A2.5 2.5 0 0 0 15.5 4h-11zm11 2.2 5.5-3.2v13.6l-5.5-3.2V6.2z"/></svg>`)
)

var oauthInviteIcons = map[string]template.HTML{
	"google":             googleOAuthIconSVG,
	"github":             githubOAuthIconSVG,
	"apple":              appleOAuthIconSVG,
	"azure":              azureOAuthIconSVG,
	"bitbucket":          bitbucketOAuthIconSVG,
	"discord":            discordOAuthIconSVG,
	"facebook":           facebookOAuthIconSVG,
	"figma":              figmaOAuthIconSVG,
	"fly":                flyOAuthIconSVG,
	"gitlab":             gitlabOAuthIconSVG,
	"kakao":              kakaoOAuthIconSVG,
	"keycloak":           keycloakOAuthIconSVG,
	"linkedin":           linkedinOAuthIconSVG,
	"linkedin_oidc":      linkedinOAuthIconSVG,
	"notion":             notionOAuthIconSVG,
	"snapchat":           snapchatOAuthIconSVG,
	"spotify":            spotifyOAuthIconSVG,
	"slack":              slackOAuthIconSVG,
	"slack_oidc":         slackOAuthIconSVG,
	"twitch":             twitchOAuthIconSVG,
	"twitter":            twitterOAuthIconSVG,
	"vercel_marketplace": vercelOAuthIconSVG,
	"vercel":             vercelOAuthIconSVG,
	"workos":             workosOAuthIconSVG,
	"zoom":               zoomOAuthIconSVG,
}

// oauthInviteIconSVG returns brand artwork for a provider id, or a generic initial badge.
func oauthInviteIconSVG(provider string) template.HTML {
	if icon, ok := oauthInviteIcons[provider]; ok {
		return icon
	}
	return genericOAuthIconSVG(provider)
}

func genericOAuthIconSVG(provider string) template.HTML {
	initial := "?"
	if provider != "" {
		initial = strings.ToUpper(string([]rune(provider)[0]))
	}
	return template.HTML(`<span class="provider-icon provider-icon--generic" aria-hidden="true">` + initial + `</span>`)
}
