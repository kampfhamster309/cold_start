package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Role is a grant's role component. Phase 0 defines exactly these three
// (tech-stack §6); Phase 2 doesn't add new roles, only new resource types
// grants can target.
type Role string

const (
	RoleViewer Role = "viewer"
	RoleEditor Role = "editor"
	RoleAdmin  Role = "admin"
)

// ResourceType is a grant's resource_type component (architecture §5).
// ResourceGlobal is the only sentinel Phase 0's UI issues grants against;
// the rest exist in the schema so Phase 2 can populate rows against them
// without a migration.
type ResourceType string

const (
	ResourceGlobal         ResourceType = "global"
	ResourceDocSpace       ResourceType = "doc_space"
	ResourceRepoConnection ResourceType = "repo_connection"
	ResourceOnboardingHire ResourceType = "onboarding_hire"
)

// ErrResourceIDMismatch mirrors the grants table's CHECK constraint:
// ResourceGlobal must carry no resource_id, every other resource_type
// must carry one. Checked here too so callers get this error instead of
// an opaque constraint-violation error from the database.
var ErrResourceIDMismatch = errors.New("auth: resource_id must be nil for ResourceGlobal and set otherwise")

// GrantRole records that userID holds role over the given resource.
// Granting the same (user, role, resource_type, resource_id) twice is a
// no-op, not an error — callers assigning roles don't need to check
// first.
func GrantRole(ctx context.Context, conn *sql.DB, userID string, role Role, resourceType ResourceType, resourceID *string) error {
	if err := validateResourceID(resourceType, resourceID); err != nil {
		return err
	}

	_, err := conn.ExecContext(ctx,
		`INSERT INTO grants (user_id, role, resource_type, resource_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		userID, role, resourceType, resourceID,
	)
	if err != nil {
		return fmt.Errorf("auth: grant role: %w", err)
	}
	return nil
}

// HasGrant reports whether userID holds role over the given resource.
func HasGrant(ctx context.Context, conn *sql.DB, userID string, role Role, resourceType ResourceType, resourceID *string) (bool, error) {
	if err := validateResourceID(resourceType, resourceID); err != nil {
		return false, err
	}

	var exists bool
	err := conn.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM grants
			WHERE user_id = $1 AND role = $2 AND resource_type = $3
			  AND resource_id IS NOT DISTINCT FROM $4
		)`,
		userID, role, resourceType, resourceID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("auth: check grant: %w", err)
	}
	return exists, nil
}

func validateResourceID(resourceType ResourceType, resourceID *string) error {
	if resourceType == ResourceGlobal && resourceID != nil {
		return ErrResourceIDMismatch
	}
	if resourceType != ResourceGlobal && resourceID == nil {
		return ErrResourceIDMismatch
	}
	return nil
}
