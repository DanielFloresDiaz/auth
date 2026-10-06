package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/supabase/auth/internal/storage"
)

// WhitelistConfirmation is an inbox proof that has not yet become a request.
type WhitelistConfirmation struct {
	ID        uuid.UUID `db:"id"`
	ProjectID uuid.UUID `db:"project_id"`
	Email     string    `db:"email"`
	Answers   []byte    `db:"answers"`
	TokenHash string    `db:"token_hash"`
	SentAt    time.Time `db:"sent_at"`
}

// FindWhitelistConfirmation returns the unconfirmed submission for this project and email.
func FindWhitelistConfirmation(tx *storage.Connection, projectID uuid.UUID, email string) (*WhitelistConfirmation, error) {
	confirmation := &WhitelistConfirmation{}
	err := tx.RawQuery(
		`SELECT id, project_id, email, answers, token_hash, sent_at
		 FROM auth.whitelist_confirmations
		 WHERE project_id = ? AND lower(email) = ?
		 LIMIT 1`,
		projectID,
		email,
	).First(confirmation)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "error finding whitelist confirmation")
	}
	return confirmation, nil
}

// FindWhitelistConfirmationByToken returns the confirmation for this hash.
func FindWhitelistConfirmationByToken(tx *storage.Connection, tokenHash string) (*WhitelistConfirmation, error) {
	confirmation := &WhitelistConfirmation{}
	err := tx.RawQuery(
		`SELECT id, project_id, email, answers, token_hash, sent_at
		 FROM auth.whitelist_confirmations
		 WHERE token_hash = ?
		 LIMIT 1`,
		tokenHash,
	).First(confirmation)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "error finding whitelist confirmation")
	}
	return confirmation, nil
}

// UpsertWhitelistConfirmation replaces any unconfirmed submission for this project and email.
func UpsertWhitelistConfirmation(tx *storage.Connection, id, projectID uuid.UUID, email, tokenHash string, answersJSON []byte, sentAt time.Time) error {
	return tx.Transaction(func(tx *storage.Connection) error {
		if err := tx.RawQuery(
			`DELETE FROM auth.whitelist_confirmations WHERE project_id = ? AND lower(email) = ?`,
			projectID,
			email,
		).Exec(); err != nil {
			return errors.Wrap(err, "error replacing whitelist confirmation")
		}
		if err := tx.RawQuery(
			`INSERT INTO auth.whitelist_confirmations (id, project_id, email, answers, token_hash, sent_at)
			 VALUES (?, ?, ?, CAST(? AS jsonb), ?, ?)`,
			id,
			projectID,
			email,
			string(answersJSON),
			tokenHash,
			sentAt,
		).Exec(); err != nil {
			return errors.Wrap(err, "error creating whitelist confirmation")
		}
		return nil
	})
}

// DeleteWhitelistConfirmation removes a confirmation after it has been applied.
func DeleteWhitelistConfirmation(tx *storage.Connection, id uuid.UUID) error {
	if err := tx.RawQuery(
		`DELETE FROM auth.whitelist_confirmations WHERE id = ?`,
		id,
	).Exec(); err != nil {
		return errors.Wrap(err, "error deleting whitelist confirmation")
	}
	return nil
}

// ReopenRejectedWhitelistRequest replaces a rejected request with a new pending submission.
func ReopenRejectedWhitelistRequest(tx *storage.Connection, projectID uuid.UUID, email string, answersJSON []byte) error {
	if err := tx.RawQuery(
		`UPDATE auth.whitelist_requests
		 SET answers = CAST(? AS jsonb),
		     status = 'pending',
		     rejected_at = NULL,
		     updated_at = current_timestamp
		 WHERE project_id = ? AND lower(email) = ? AND status = 'rejected'`,
		string(answersJSON),
		projectID,
		email,
	).Exec(); err != nil {
		return errors.Wrap(err, "error reopening whitelist request")
	}
	return nil
}
