// This file verifies publisher write ownership validation and database-side
// owner filtering used by marketplace publishing operations.

package marketplace

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestCreatePublisherRequiresOwnerUserID(t *testing.T) {
	service := &serviceImpl{}
	_, err := service.CreatePublisher(context.Background(), CreatePublisherInput{
		PublisherKey: "publisher-a",
		Name:         "Publisher A",
	})
	if !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
}

func TestUpdatePublisherRequiresOwnerUserID(t *testing.T) {
	service := &serviceImpl{}
	_, err := service.UpdatePublisher(context.Background(), UpdatePublisherInput{
		CurrentPublisherKey: "publisher-a",
		PublisherKey:        "publisher-a",
		Name:                "Publisher A",
	})
	if !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
}

func TestPublisherOwnerFilterAddsOwnerToDatabaseQuery(t *testing.T) {
	captured := []capturedSelect{}
	db := newSelectCaptureDB(t)
	model := applyPublisherOwnerFilter(
		db.Model(dao.PluginMarketplacePublisher.Table()).
			Where(dao.PluginMarketplacePublisher.Columns().PublisherKey, "publisher-a"),
		1001,
	).Hook(selectCaptureHook(&captured))

	var publisher *entity.PluginMarketplacePublisher
	if err := model.Scan(&publisher); err != nil {
		t.Fatalf("execute publisher ownership query: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected one captured query, got %d", len(captured))
	}

	query := captured[0]
	sql := strings.ToLower(query.sql)
	for _, fragment := range []string{"publisher_key", "owner_user_id"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("publisher ownership query missing %q: %s", fragment, query.sql)
		}
	}
	if !containsArgument(query.args, "publisher-a") && !strings.Contains(sql, "publisher-a") {
		t.Fatalf("publisher ownership query missing publisher key: sql=%s args=%#v", query.sql, query.args)
	}
	if !containsArgument(query.args, int64(1001)) && !strings.Contains(sql, "1001") {
		t.Fatalf("publisher ownership query missing owner user ID: sql=%s args=%#v", query.sql, query.args)
	}
}
