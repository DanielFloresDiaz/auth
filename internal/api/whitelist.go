package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid"
	"github.com/supabase/auth/internal/api/apierrors"
	"github.com/supabase/auth/internal/crypto"
	"github.com/supabase/auth/internal/mailer/templatemailer"
	"github.com/supabase/auth/internal/models"
	"github.com/supabase/auth/internal/storage"
)

const maxWhitelistAnswersBytes = 16 * 1024

const whitelistAckMessage = "Check your email to confirm this access request."

// WhitelistParams is the public body for POST /whitelist.
type WhitelistParams struct {
	ProjectID uuid.UUID      `json:"project_id"`
	Email     string         `json:"email"`
	Answers   map[string]any `json:"answers"`
}

type whitelistAckResponse struct {
	Message string `json:"message"`
}

// CreateWhitelistRequest emails a confirmation link. The request is stored only after that link is opened.
func (a *API) CreateWhitelistRequest(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	config := a.config

	params := &WhitelistParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	if params.ProjectID == uuid.Nil {
		return apierrors.NewNotFoundError(apierrors.ErrorCodeValidationFailed, "Project not found")
	}

	email, err := normalizeWhitelistEmail(params.Email)
	if err != nil {
		return err
	}

	answers := params.Answers
	if answers == nil {
		answers = map[string]any{}
	}
	encoded, err := json.Marshal(answers)
	if err != nil {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "answers must be a JSON object")
	}
	if len(encoded) > maxWhitelistAnswersBytes {
		return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "answers are too large")
	}

	exists, err := models.ProjectExists(db, params.ProjectID)
	if err != nil {
		return apierrors.NewInternalServerError("Database error finding project").WithInternalError(err)
	}
	if !exists {
		return apierrors.NewNotFoundError(apierrors.ErrorCodeValidationFailed, "Project not found")
	}

	existing, err := models.FindWhitelistRequest(db, params.ProjectID, email)
	if err != nil {
		return apierrors.NewInternalServerError("Database error finding whitelist request").WithInternalError(err)
	}
	if existing != nil && whitelistRequestIsOpen(existing.Status) {
		return sendWhitelistAck(w)
	}

	confirmation, err := models.FindWhitelistConfirmation(db, params.ProjectID, email)
	if err != nil {
		return apierrors.NewInternalServerError("Database error finding whitelist confirmation").WithInternalError(err)
	}
	if confirmation != nil {
		sentAt := confirmation.SentAt
		if err := validateSentWithinFrequencyLimit(&sentAt, config.SMTP.MaxFrequency); err != nil {
			return sendWhitelistAck(w)
		}
	}

	externalURL := getExternalHost(ctx)
	if externalURL == nil {
		return apierrors.NewInternalServerError("External host is not configured")
	}

	tokenHash := crypto.GenerateTokenHash(email, crypto.GenerateOtp(config.Mailer.OtpLength))
	query := url.Values{}
	query.Set("token", tokenHash)
	confirmURL := externalURL.ResolveReference(&url.URL{Path: "/whitelist/confirm", RawQuery: query.Encode()}).String()

	if err := a.mailer.WhitelistConfirmationMail(r, params.ProjectID.String(), email, confirmURL); err != nil {
		return apierrors.NewInternalServerError("Error sending confirmation email").WithInternalError(err)
	}

	if err := models.UpsertWhitelistConfirmation(db, uuid.Must(uuid.NewV4()), params.ProjectID, email, tokenHash, encoded, a.Now()); err != nil {
		return apierrors.NewInternalServerError("Database error storing whitelist confirmation").WithInternalError(err)
	}
	return sendWhitelistAck(w)
}

// ConfirmWhitelistRequest turns a confirmation link into a pending request.
func (a *API) ConfirmWhitelistRequest(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)

	token := r.URL.Query().Get("token")
	if token == "" {
		return invalidWhitelistConfirmation()
	}

	confirmation, err := models.FindWhitelistConfirmationByToken(db, token)
	if err != nil {
		return apierrors.NewInternalServerError("Database error finding whitelist confirmation").WithInternalError(err)
	}
	if confirmation == nil || isOtpExpired(&confirmation.SentAt, a.config.Mailer.OtpExp) {
		return invalidWhitelistConfirmation()
	}

	err = db.Transaction(func(tx *storage.Connection) error {
		existing, terr := models.FindWhitelistRequest(tx, confirmation.ProjectID, confirmation.Email)
		if terr != nil {
			return terr
		}
		switch {
		case existing == nil:
			_, terr = models.CreateWhitelistRequest(tx, uuid.Must(uuid.NewV4()), confirmation.ProjectID, confirmation.Email, confirmation.Answers)
		case existing.Status == models.WhitelistStatusRejected:
			terr = models.ReopenRejectedWhitelistRequest(tx, confirmation.ProjectID, confirmation.Email, confirmation.Answers)
		}
		if terr != nil {
			return terr
		}
		return models.DeleteWhitelistConfirmation(tx, confirmation.ID)
	})
	if err != nil {
		return apierrors.NewInternalServerError("Database error confirming whitelist request").WithInternalError(err)
	}

	projectID := confirmation.ProjectID.String()
	pageData := templatemailer.WhitelistConfirmedTemplateData{
		Email:     confirmation.Email,
		ProjectID: projectID,
		SiteURL:   a.config.SiteURL,
	}
	var html string
	if mailer, ok := a.mailer.(*templatemailer.Mailer); ok {
		html, err = mailer.RenderWhitelistConfirmedPage(ctx, projectID, pageData)
	} else {
		html, err = templatemailer.RenderDefaultWhitelistConfirmedPage(pageData)
	}
	if err != nil {
		return apierrors.NewInternalServerError("Error rendering whitelist confirmation page").WithInternalError(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(html))
	return err
}

func invalidWhitelistConfirmation() error {
	return apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Confirmation link is invalid or has expired")
}

func whitelistRequestIsOpen(status string) bool {
	switch status {
	case models.WhitelistStatusPending, models.WhitelistStatusViewed, models.WhitelistStatusAccepted:
		return true
	default:
		return false
	}
}

func sendWhitelistAck(w http.ResponseWriter) error {
	return sendJSON(w, http.StatusOK, &whitelistAckResponse{Message: whitelistAckMessage})
}

func normalizeWhitelistEmail(email string) (string, error) {
	cleaned := strings.ToLower(strings.TrimSpace(email))
	if cleaned == "" || utf8.RuneCountInString(cleaned) > 320 ||
		!strings.Contains(cleaned, "@") ||
		strings.HasPrefix(cleaned, "@") ||
		strings.HasSuffix(cleaned, "@") ||
		strings.Contains(cleaned, " ") {
		return "", apierrors.NewBadRequestError(apierrors.ErrorCodeValidationFailed, "Invalid email address")
	}
	return cleaned, nil
}
