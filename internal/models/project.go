package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	RateLimits  RateLimit `json:"rate_limits" db:"rate_limits"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name used by pop
func (Project) TableName() string {
	tableName := "projects"
	return tableName
}
