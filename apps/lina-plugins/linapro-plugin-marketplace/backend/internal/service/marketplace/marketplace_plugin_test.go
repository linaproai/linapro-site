// This file verifies marketplace plugin identity writes require an explicit
// publisher owner before any plugin lookup or mutation.

package marketplace

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
)

func TestSavePluginDraftRequiresOwnerUserID(t *testing.T) {
	service := &serviceImpl{}
	_, err := service.SavePluginDraft(context.Background(), SavePluginDraftInput{
		PublisherKey: "publisher-a",
		PluginID:     "plugin-a",
		Name:         "Plugin A",
		Summary:      "Plugin summary",
	})
	if !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
}
