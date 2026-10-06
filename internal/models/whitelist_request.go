package models

import (
	"database/sql"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/supabase/auth/internal/storage"
)

const (
	WhitelistStatusPending  = "pending"
	WhitelistStatusViewed   = "viewed"
	WhitelistStatusAccepted = "accepted"
	WhitelistStatusRejected = "rejected"
)

// WhitelistRequest is a public access request for one project.
type WhitelistRequest struct {
	ID      uuid.UUID `json:"id" db:"id"`
	Email   string    `json:"email" db:"email"`
	Status  string    `json:"status" db:"status"`
	Answers []byte    `json:"-" db:"answers"`
}

// FindWhitelistRequest returns the request for this project and email, including a rejected one.
func FindWhitelistRequest(tx *storage.Connection, projectID uuid.UUID, email string) (*WhitelistRequest, error) {
	request := &WhitelistRequest{}
	err := tx.RawQuery(
		`SELECT id, email, status::text AS status, answers
		 FROM auth.whitelist_requests
		 WHERE project_id = ? AND lower(email) = ?
		 LIMIT 1`,
		projectID,
		email,
	).First(request)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "error finding whitelist request")
	}
	return request, nil
}

// CreateWhitelistRequest inserts a pending request. answersJSON must be a JSON object.
func CreateWhitelistRequest(tx *storage.Connection, id, projectID uuid.UUID, email string, answersJSON []byte) (*WhitelistRequest, error) {
	if err := tx.RawQuery(
		`INSERT INTO auth.whitelist_requests (id, project_id, email, answers, status)
		 VALUES (?, ?, ?, CAST(? AS jsonb), 'pending')`,
		id,
		projectID,
		email,
		string(answersJSON),
	).Exec(); err != nil {
		return nil, errors.Wrap(err, "error creating whitelist request")
	}
	return &WhitelistRequest{
		ID:      id,
		Email:   email,
		Status:  WhitelistStatusPending,
		Answers: answersJSON,
	}, nil
}

// ProjectExists reports whether auth.projects contains the id.
func ProjectExists(tx *storage.Connection, projectID uuid.UUID) (bool, error) {
	var id uuid.UUID
	err := tx.RawQuery(`SELECT id FROM auth.projects WHERE id = ?`, projectID).First(&id)
	if errors.Cause(err) == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "error finding project")
	}
	return true, nil
}

// AnswersObject unmarshals the stored JSON object.
func (r *WhitelistRequest) AnswersObject() (map[string]any, error) {
	if len(r.Answers) == 0 {
		return map[string]any{}, nil
	}
	var answers map[string]any
	if err := json.Unmarshal(r.Answers, &answers); err != nil {
		return nil, err
	}
	if answers == nil {
		return map[string]any{}, nil
	}
	return answers, nil
}
