package api

import (
	"net/http"

	"auth/internal/models"
	"auth/internal/storage"
	"github.com/crewjam/saml"
	"github.com/gofrs/uuid"
)

type SingleSignOnParams struct {
	ProviderID          uuid.UUID `json:"provider_id"`
	Domain              string    `json:"domain"`
	RedirectTo          string    `json:"redirect_to"`
	SkipHTTPRedirect    *bool     `json:"skip_http_redirect"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	ProjectID           uuid.UUID `json:"project_id"`
}

type SingleSignOnResponse struct {
	URL string `json:"url"`
}

func (p *SingleSignOnParams) validate() (bool, error) {
	hasProviderID := p.ProviderID != uuid.Nil
	hasDomain := p.Domain != ""

	if hasProviderID && hasDomain {
		return hasProviderID, badRequestError(ErrorCodeValidationFailed, "Only one of provider_id or domain supported")
	} else if !hasProviderID && !hasDomain {
		return hasProviderID, badRequestError(ErrorCodeValidationFailed, "A provider_id or domain needs to be provided")
	}
	if p.OrganizationID == uuid.Nil && p.ProjectID == uuid.Nil {
		return hasProviderID, badRequestError(ErrorCodeValidationFailed, "Organization ID or Project ID must be set")
	}

	return hasProviderID, nil
}

// SingleSignOn handles the single-sign-on flow for a provided SSO domain or provider.
func (a *API) SingleSignOn(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)

	params := &SingleSignOnParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	var err error
	hasProviderID := false

	if hasProviderID, err = params.validate(); err != nil {
		return err
	}
	codeChallengeMethod := params.CodeChallengeMethod
	codeChallenge := params.CodeChallenge

	if err := validatePKCEParams(codeChallengeMethod, codeChallenge); err != nil {
		return err
	}
	flowType := getFlowFromChallenge(params.CodeChallenge)
	var flowStateID *uuid.UUID
	flowStateID = nil
	if isPKCEFlow(flowType) {

		organization_id := params.OrganizationID
		project_id := params.ProjectID
		flowState, err := generateFlowState(db, models.SSOSAML.String(), models.SSOSAML, codeChallengeMethod, codeChallenge, nil, organization_id, project_id)
		if err != nil {
			return err
		}
		flowStateID = &flowState.ID
	}

	var ssoProvider *models.SSOProvider

	if hasProviderID {
		ssoProvider, err = models.FindSSOProviderByID(db, params.ProviderID)
		if models.IsNotFoundError(err) {
			return notFoundError(ErrorCodeSSOProviderNotFound, "No such SSO provider")
		} else if err != nil {
			return internalServerError("Unable to find SSO provider by ID").WithInternalError(err)
		}
	} else {
		ssoProvider, err = models.FindSSOProviderByDomain(db, params.Domain)
		if models.IsNotFoundError(err) {
			return notFoundError(ErrorCodeSSOProviderNotFound, "No SSO provider assigned for this domain")
		} else if err != nil {
			return internalServerError("Unable to find SSO provider by domain").WithInternalError(err)
		}
	}

	entityDescriptor, err := ssoProvider.SAMLProvider.EntityDescriptor()
	if err != nil {
		return internalServerError("Error parsing SAML Metadata for SAML provider").WithInternalError(err)
	}

	serviceProvider := a.getSAMLServiceProvider(entityDescriptor, false /* <- idpInitiated */)

	authnRequest, err := serviceProvider.MakeAuthenticationRequest(
		serviceProvider.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return internalServerError("Error creating SAML Authentication Request").WithInternalError(err)
	}

	// Some IdPs do not support the use of the `persistent` NameID format,
	// and require a different format to be sent to work.
	if ssoProvider.SAMLProvider.NameIDFormat != nil {
		authnRequest.NameIDPolicy.Format = ssoProvider.SAMLProvider.NameIDFormat
	}

	relayState := models.SAMLRelayState{
		SSOProviderID: ssoProvider.ID,
		RequestID:     authnRequest.ID,
		RedirectTo:    params.RedirectTo,
		FlowStateID:   flowStateID,
	}

	if err := db.Transaction(func(tx *storage.Connection) error {
		if terr := tx.Create(&relayState); terr != nil {
			return internalServerError("Error creating SAML relay state from sign up").WithInternalError(err)
		}

		return nil
	}); err != nil {
		return err
	}

	ssoRedirectURL, err := authnRequest.Redirect(relayState.ID.String(), serviceProvider)
	if err != nil {
		return internalServerError("Error creating SAML authentication request redirect URL").WithInternalError(err)
	}

	skipHTTPRedirect := false

	if params.SkipHTTPRedirect != nil {
		skipHTTPRedirect = *params.SkipHTTPRedirect
	}

	if skipHTTPRedirect {
		return sendJSON(w, http.StatusOK, SingleSignOnResponse{
			URL: ssoRedirectURL.String(),
		})
	}

	http.Redirect(w, r, ssoRedirectURL.String(), http.StatusSeeOther)
	return nil
}
