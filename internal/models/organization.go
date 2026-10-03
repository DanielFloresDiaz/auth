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
	ZDR       bool      `json:"zdr" db:"zdr"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationTier struct {
	OrganizationID  uuid.UUID `json:"organization_id" db:"organization_id"`
	ProjectID       uuid.UUID `json:"project_id" db:"project_id"`
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

func findOrganization(tx *storage.Connection, query string, args ...interface{}) (*Organization, error) {
	obj := &Organization{}
	if err := tx.Eager().Q().Where(query, args...).First(obj); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error finding organization")
	}
	return obj, nil
}

// FindOrganizationZDR returns the zero-data-retention flag for an organization.
// Missing organizations default to false.
func FindOrganizationZDR(tx *storage.Connection, organization_id uuid.UUID) (bool, error) {
	if organization_id == uuid.Nil {
		return false, nil
	}
	org, err := findOrganization(tx, "id = ?", organization_id)
	if err != nil {
		return false, err
	}
	if org == nil {
		return false, nil
	}
	return org.ZDR, nil
}

func tiersFromOrganizationTier(organizationTier *OrganizationTier, organization_role string) (string, string, string) {
	if organization_role == "admin" || organization_role == "project_admin" {
		return organizationTier.AdminTierModel, organizationTier.AdminTierTime, organizationTier.AdminTierUsage
	}
	return organizationTier.ClientTierModel, organizationTier.ClientTierTime, organizationTier.ClientTierUsage
}

func FindTiersByOrganizationIDAndOrganizationRole(
	tx *storage.Connection,
	organization_id uuid.UUID,
	project_id uuid.UUID,
	organization_role string,
) (string, string, string, error) {
	tier_model := "free"
	tier_time := "free"
	tier_usage := "free"

	if organization_id == uuid.Nil {
		return tier_model, tier_time, tier_usage, nil
	}

	organizationTier, err := findOrganizationTier(tx, "organization_id = ? AND project_id = ?", organization_id, project_id)
	if err != nil {
		return "", "", "", err
	}
	if organizationTier == nil {
		organizationTier, err = findOrganizationTier(tx, "organization_id = ? AND project_id = ?", uuid.Nil, project_id)
		if err != nil {
			return "", "", "", err
		}
	}
	if organizationTier == nil {
		return tier_model, tier_time, tier_usage, nil
	}

	tier_model, tier_time, tier_usage = tiersFromOrganizationTier(organizationTier, organization_role)
	return tier_model, tier_time, tier_usage, nil
}
