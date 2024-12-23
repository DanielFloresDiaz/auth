package api

import (
	"net/http"

	"auth/internal/api/provider"
	"auth/internal/models"
	"auth/internal/storage"

	"github.com/fatih/structs"
	"github.com/gofrs/uuid"
)

// InviteParams are the parameters the Signup endpoint accepts
type InviteParams struct {
	Email          string                 `json:"email"`
	Data           map[string]interface{} `json:"data"`
	OrganizationID uuid.UUID              `json:"organization_id"`
}

// Invite is the endpoint for inviting a new user
func (a *API) Invite(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)
	adminUser := getAdminUser(ctx)
	params := &InviteParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}

	var err error
	params.Email, err = a.validateEmail(params.Email)
	if err != nil {
		return err
	}

	aud := a.requestAud(ctx, r)
	user, err := models.FindUserByEmailAndAudience(db, params.Email, aud, params.OrganizationID, uuid.Nil)
	if err != nil && !models.IsNotFoundError(err) {
		return internalServerError("Database error finding user").WithInternalError(err)
	}

	err = db.Transaction(func(tx *storage.Connection) error {
		if user != nil {
			if user.IsConfirmed() {
				return unprocessableEntityError(ErrorCodeEmailExists, DuplicateEmailMsg)
			}
		} else {
			signupParams := SignupParams{
				Email:          params.Email,
				Data:           params.Data,
				Aud:            aud,
				Provider:       "email",
				OrganizationID: params.OrganizationID,
			}

			// because params above sets no password, this method
			// is not computationally hard so it can be used within
			// a database transaction
			user, err = signupParams.ToUserModel(false /* <- isSSOUser */)
			if err != nil {
				return err
			}

			var excludeColumns []string
			excludeColumns = append(excludeColumns, "organization_role")
			excludeColumns = append(excludeColumns, "project_id")
			if user.OrganizationID.UUID == uuid.Nil {
				excludeColumns = append(excludeColumns, "organization_id")
			}

			user, err = a.signupNewUser(tx, user, excludeColumns...)
			if err != nil {
				return err
			}
			identity, err := a.createNewIdentity(tx, user, "email", structs.Map(provider.Claims{
				Subject: user.ID.String(),
				Email:   user.GetEmail(),
			}),
				excludeColumns...)
			if err != nil {
				return err
			}
			user.Identities = []models.Identity{*identity}
		}

		if terr := models.NewAuditLogEntry(r, tx, adminUser, models.UserInvitedAction, "", map[string]interface{}{
			"user_id":    user.ID,
			"user_email": user.Email,
		}); terr != nil {
			return terr
		}

		if err := a.sendInvite(r, tx, user); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return sendJSON(w, http.StatusOK, user)
}
