// This file verifies the unread-count controller surfaces notify-owned
// unauthenticated errors instead of defining controller-local bizerr codes.

package usermsg

import (
	"context"
	"testing"

	v1 "lina-core/api/usermsg/v1"
	"lina-core/internal/service/notify"
	"lina-core/pkg/bizerr"
)

// TestCountRejectsMissingCurrentUser verifies a missing request identity is
// rejected by the notify inbox service with the stable unauthenticated code.
func TestCountRejectsMissingCurrentUser(t *testing.T) {
	controller := &ControllerV1{notifySvc: notify.New(nil)}
	_, err := controller.Count(context.Background(), &v1.CountReq{})
	if !bizerr.Is(err, notify.CodeNotifyNotAuthenticated) {
		t.Fatalf("Count missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
}
