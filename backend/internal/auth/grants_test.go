package auth_test

import (
	"context"
	"testing"

	"cold_start/backend/internal/auth"
)

func TestGrantRoleAndHasGrant_Global(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	has, err := auth.HasGrant(ctx, conn, user.ID, auth.RoleAdmin, auth.ResourceGlobal, nil)
	if err != nil {
		t.Fatalf("HasGrant before granting: %v", err)
	}
	if has {
		t.Fatal("expected no grant before GrantRole")
	}

	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleAdmin, auth.ResourceGlobal, nil); err != nil {
		t.Fatalf("GrantRole: %v", err)
	}

	has, err = auth.HasGrant(ctx, conn, user.ID, auth.RoleAdmin, auth.ResourceGlobal, nil)
	if err != nil {
		t.Fatalf("HasGrant after granting: %v", err)
	}
	if !has {
		t.Fatal("expected a grant after GrantRole")
	}

	// A different role for the same user/resource is still ungranted.
	has, err = auth.HasGrant(ctx, conn, user.ID, auth.RoleViewer, auth.ResourceGlobal, nil)
	if err != nil {
		t.Fatalf("HasGrant for a different role: %v", err)
	}
	if has {
		t.Fatal("expected RoleViewer to be ungranted when only RoleAdmin was granted")
	}
}

func TestGrantRole_IsIdempotent(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleEditor, auth.ResourceGlobal, nil); err != nil {
		t.Fatalf("first GrantRole: %v", err)
	}
	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleEditor, auth.ResourceGlobal, nil); err != nil {
		t.Fatalf("second GrantRole (expected no-op, not an error): %v", err)
	}
}

func TestGrantRole_RejectsResourceIDMismatch(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resourceID := "not-nil"
	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleViewer, auth.ResourceGlobal, &resourceID); err != auth.ErrResourceIDMismatch {
		t.Fatalf("expected ErrResourceIDMismatch for a global grant with a resource_id, got %v", err)
	}

	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleViewer, auth.ResourceDocSpace, nil); err != auth.ErrResourceIDMismatch {
		t.Fatalf("expected ErrResourceIDMismatch for a scoped grant with no resource_id, got %v", err)
	}
}
