package api

import (
	"net/http"

	"auth/internal/metering"
	"auth/internal/models"
	"auth/internal/storage"
	"github.com/gofrs/uuid"
)

func (a *API) SignupAnonymously(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	config := a.config
	db := a.db.WithContext(ctx)
	aud := a.requestAud(ctx, r)

	if config.DisableSignup {
		return unprocessableEntityError(ErrorCodeSignupDisabled, "Signups not allowed for this instance")
	}

	params := &SignupParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	if params.OrganizationID == uuid.Nil {
		return badRequestError(ErrorCodeValidationFailed, "Organization ID is required")
	}

	params.Aud = aud
	params.Provider = "anonymous"

	newUser, err := params.ToUserModel(false /* <- isSSOUser */)
	if err != nil {
		return err
	}

	var grantParams models.GrantParams
	grantParams.FillGrantParams(r)

	var token *AccessTokenResponse
	err = db.Transaction(func(tx *storage.Connection) error {
		var terr error
		var excludeColumns []string
		excludeColumns = append(excludeColumns, "organization_role")
		excludeColumns = append(excludeColumns, "project_id")

		newUser, terr = a.signupNewUser(tx, newUser, excludeColumns...)
		if terr != nil {
			return terr
		}
		token, terr = a.issueRefreshToken(r, tx, newUser, models.Anonymous, grantParams)
		if terr != nil {
			return terr
		}
		return nil
	})
	if err != nil {
		return internalServerError("Database error creating anonymous user").WithInternalError(err)
	}

	metering.RecordLogin("anonymous", newUser.ID)
	return sendJSON(w, http.StatusOK, token)
}
