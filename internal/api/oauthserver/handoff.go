package oauthserver

import (
	"context"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/supabase/auth/internal/models"
)

const handoffClientName = "frontend-handoff"

// EnsureHandoffClient upserts the public client the product frontends use to
// move a session to another origin. Missing id or redirect URIs leaves the
// table alone so auth still starts when handoff is not configured.
func (s *Server) EnsureHandoffClient(ctx context.Context) error {
	cfg := s.config.OAuthServer
	clientID := strings.TrimSpace(cfg.HandoffClientID)
	redirectURIs := compactRedirectURIs(cfg.HandoffRedirectURIs)
	if clientID == "" || len(redirectURIs) == 0 {
		return nil
	}

	id, err := uuid.FromString(clientID)
	if err != nil {
		return err
	}

	params := &OAuthServerClientRegisterParams{
		RedirectURIs:            redirectURIs,
		ClientType:              models.OAuthServerClientTypePublic,
		TokenEndpointAuthMethod: models.TokenEndpointAuthMethodNone,
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ClientName:              handoffClientName,
		RegistrationType:        "manual",
	}
	if err := params.validate(); err != nil {
		return err
	}

	db := s.db.WithContext(ctx)
	existing, err := models.FindOAuthServerClientByID(db, id)
	if err != nil {
		if !models.IsNotFoundError(err) {
			return err
		}
		client := &models.OAuthServerClient{
			ID:                      id,
			RegistrationType:        "manual",
			ClientType:              models.OAuthServerClientTypePublic,
			TokenEndpointAuthMethod: models.TokenEndpointAuthMethodNone,
			ClientName:              &params.ClientName,
		}
		client.SetRedirectURIs(redirectURIs)
		client.SetGrantTypes(params.GrantTypes)
		return models.CreateOAuthServerClient(db, client)
	}

	existing.RegistrationType = "manual"
	existing.ClientType = models.OAuthServerClientTypePublic
	existing.TokenEndpointAuthMethod = models.TokenEndpointAuthMethodNone
	existing.ClientSecretHash = ""
	existing.ClientName = &params.ClientName
	existing.SetRedirectURIs(redirectURIs)
	existing.SetGrantTypes(params.GrantTypes)
	return models.UpdateOAuthServerClient(db, existing)
}

func compactRedirectURIs(uris []string) []string {
	out := make([]string, 0, len(uris))
	for _, uri := range uris {
		trimmed := strings.TrimSpace(uri)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
