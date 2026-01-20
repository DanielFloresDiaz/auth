package models

import (
	"database/sql"
	"time"

	"github.com/supabase/auth/internal/storage"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	AdminID   uuid.UUID `json:"admin_id" db:"admin_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationTier struct {
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	Tier            string    `json:"tier" db:"tier"`
	AdminTierModel  string    `json:"admin_tier_model" db:"admin_tier_model"`
	ClientTierModel string    `json:"client_tier_model" db:"client_tier_model"`
	AdminTierTime   string    `json:"admin_tier_time" db:"admin_tier_time"`
	ClientTierTime  string    `json:"client_tier_time" db:"client_tier_time"`
	AdminTierUsage  string    `json:"admin_tier_usage" db:"admin_tier_usage"`
	ClientTierUsage string    `json:"client_tier_usage" db:"client_tier_usage"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name used by pop
func (Organization) TableName() string {
	tableName := "organizations"
	return tableName
}

func (OrganizationTier) TableName() string {
	return "organizations_tier"
}

func findOrganizationTier(tx *storage.Connection, query string, args ...interface{}) (*OrganizationTier, error) {
	obj := &OrganizationTier{}
	if err := tx.Eager().Q().Where(query, args...).First(obj); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil // Or handle as needed
		}
		return nil, errors.Wrap(err, "error finding organization tier")
	}

	return obj, nil
}

func FindTiersByOrganizationIDAndOrganizationRole(tx *storage.Connection, organization_id uuid.UUID, organization_role string) (string, string, string, error) {

	var tier_model string = "free"
	var tier_time string = "free"
	var tier_usage string = "free"
	var query string
	var args []interface{}

	if organization_id != uuid.Nil {
		query = "organization_id = ?"
		args = append(args, organization_id)
		organizationTier, err := findOrganizationTier(tx, query, args...)

		if err != nil {
			return "", "", "", err
		}

		if organizationTier == nil {
			return tier_model, tier_time, tier_usage, nil
		}

		if organization_role == "admin" {
			tier_model = organizationTier.AdminTierModel
			tier_time = organizationTier.AdminTierTime
			tier_usage = organizationTier.AdminTierUsage
		} else {
			tier_model = organizationTier.ClientTierModel
			tier_time = organizationTier.ClientTierTime
			tier_usage = organizationTier.ClientTierUsage
		}
	}
	return tier_model, tier_time, tier_usage, nil
}
