package api

import (
	"net/http"

	"github.com/supabase/auth/internal/api/apierrors"
	"github.com/supabase/auth/internal/mailer/templatemailer"
	"github.com/supabase/auth/internal/models"
	"github.com/supabase/auth/internal/utilities"
)

// AcceptInvite lets an invited user pick an OAuth provider before starting /authorize.
func (a *API) AcceptInvite(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	config := a.config

	inviteToken := r.URL.Query().Get("invite_token")
	if inviteToken == "" {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "invite_token is required")
	}

	user, err := models.FindUserByConfirmationToken(db, inviteToken)
	if err != nil {
		if models.IsNotFoundError(err) {
			return apierrors.NewNotFoundError(apierrors.ErrorCodeInviteNotFound, "Invite not found")
		}
		return apierrors.NewInternalServerError("Database error finding user").WithInternalError(err)
	}
	if err := validateInviteAcceptance(user, config.Mailer.OtpExp); err != nil {
		return err
	}

	referrerURL := utilities.GetReferrer(r, config)
	externalURL := getExternalHost(ctx)
	providers := templatemailer.BuildAcceptInviteProviders(config, inviteToken, referrerURL, externalURL)
	if len(providers) == 0 {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "No OAuth providers are enabled for invite acceptance")
	}

	if len(providers) == 1 {
		http.Redirect(w, r, providers[0].URL, http.StatusFound)
		return nil
	}

	mailer, ok := a.mailer.(*templatemailer.Mailer)
	if !ok {
		return apierrors.NewInternalServerError("Accept invite page is not configured")
	}

	projectID := user.ProjectID.String()
	html, err := mailer.RenderAcceptInvitePage(ctx, projectID, templatemailer.AcceptInviteTemplateData{
		Email:     user.GetEmail(),
		SiteURL:   config.SiteURL,
		ProjectID: projectID,
		Providers: providers,
	})
	if err != nil {
		return apierrors.NewInternalServerError("Error rendering invite acceptance page").WithInternalError(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(html))
	return err
}

// validateInviteAcceptance rejects invites that are banned or past Mailer.OtpExp.
// A missing ConfirmationSentAt is expired, matching isOtpValid.
func validateInviteAcceptance(user *models.User, otpExp uint) error {
	if user.IsBanned() {
		return apierrors.NewForbiddenError(apierrors.ErrorCodeUserBanned, "User is banned")
	}
	if user.ConfirmationSentAt == nil || isOtpExpired(user.ConfirmationSentAt, otpExp) {
		return apierrors.NewForbiddenError(apierrors.ErrorCodeOTPExpired, "Email link is invalid or has expired")
	}
	return nil
}
