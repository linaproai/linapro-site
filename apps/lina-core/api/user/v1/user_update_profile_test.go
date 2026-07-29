// This file verifies current-user UpdateProfileReq patch validation boundaries.

package v1

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/util/gvalid"
)

// TestUpdateProfileAllowsPasswordOnlyPatch verifies password-only profile updates pass validation.
func TestUpdateProfileAllowsPasswordOnlyPatch(t *testing.T) {
	t.Parallel()

	password := "newpass123"
	req := UpdateProfileReq{Password: &password}
	if err := gvalid.New().Data(req).Run(context.Background()); err != nil {
		t.Fatalf("expected password-only profile update to pass validation, got %v", err)
	}
}
