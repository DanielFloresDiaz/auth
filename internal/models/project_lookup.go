package models

import (
	"database/sql"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/supabase/auth/internal/storage"
)

// FindProjectByID returns the project row for id.
func FindProjectByID(tx *storage.Connection, id uuid.UUID) (*Project, error) {
	if id == uuid.Nil {
		return nil, errors.New("project id is required")
	}
	project := &Project{}
	if err := tx.Q().Where("id = ?", id).First(project); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, errors.Wrap(err, "project not found")
		}
		return nil, errors.Wrap(err, "error finding project")
	}
	return project, nil
}

// FindProjectNameByID returns the project name without loading rate_limits JSON,
// which may not match auth's RateLimit shape for rows seeded by solomon data-init.
func FindProjectNameByID(tx *storage.Connection, id uuid.UUID) (string, error) {
	if id == uuid.Nil {
		return "", errors.New("project id is required")
	}
	var name string
	if err := tx.RawQuery(`SELECT name FROM projects WHERE id = ?`, id).First(&name); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return "", errors.Wrap(err, "project not found")
		}
		return "", errors.Wrap(err, "error finding project name")
	}
	return name, nil
}
