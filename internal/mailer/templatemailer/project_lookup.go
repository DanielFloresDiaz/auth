package templatemailer

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/supabase/auth/internal/models"
	"github.com/supabase/auth/internal/storage"
)

// ProjectNameLookupFromDB returns a lookup backed by auth.projects.
func ProjectNameLookupFromDB(db *storage.Connection) ProjectNameLookup {
	return func(ctx context.Context, projectID uuid.UUID) (string, error) {
		return models.FindProjectNameByID(db.WithContext(ctx), projectID)
	}
}
