package templates

import _ "embed"

// DefaultInviteHTML is the invite email used when a project has no override.
//
//go:embed invite.html
var DefaultInviteHTML string

// DefaultAcceptInviteHTML is the accept-invite page when a project has no override.
//
//go:embed accept-invite.html
var DefaultAcceptInviteHTML string

// DefaultWhitelistConfirmationHTML is the access-request confirmation email when a project has no override.
//
//go:embed whitelist-confirmation.html
var DefaultWhitelistConfirmationHTML string

// DefaultWhitelistConfirmedHTML is the page shown after a confirmation link is opened.
//
//go:embed whitelist-confirmed.html
var DefaultWhitelistConfirmedHTML string
