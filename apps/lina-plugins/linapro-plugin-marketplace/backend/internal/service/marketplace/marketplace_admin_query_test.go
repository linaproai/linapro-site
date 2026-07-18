// This file verifies marketplace management-query projection helpers, including
// page-level ID collection and batched plugin-tag grouping behavior.

package marketplace

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

type capturedSelect struct {
	args []any
	sql  string
}

func TestListOwnedPluginsFiltersOwnerInDatabaseQuery(t *testing.T) {
	captured := []capturedSelect{}
	db := newSelectCaptureDB(t)
	model := applyOwnerPublisherFilter(
		db.Model(dao.PluginMarketplacePlugin.Table()).Hook(selectCaptureHook(&captured)),
		db.Model(dao.PluginMarketplacePublisher.Table()),
		1001,
	)
	if _, err := model.Count(); err != nil {
		t.Fatalf("execute owned plugin query: %v", err)
	}

	if len(captured) != 1 {
		t.Fatalf("expected one captured query, got %d", len(captured))
	}
	for _, query := range captured {
		sql := strings.ToLower(query.sql)
		for _, fragment := range []string{
			"publisher_id",
			"plugin_marketplace_publisher",
			"owner_user_id",
		} {
			if !strings.Contains(sql, fragment) {
				t.Fatalf("owned query missing %q filter: %s", fragment, query.sql)
			}
		}
		if !containsArgument(query.args, int64(1001)) {
			t.Fatalf("owned query missing owner user argument: %#v", query.args)
		}
	}
}

func TestListReviewQueueDefaultsToSubmittedAndReviewing(t *testing.T) {
	got := reviewQueueStatuses("")
	want := []string{
		marketv1.MarketplaceReviewStatusSubmitted.String(),
		marketv1.MarketplaceReviewStatusReviewing.String(),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected default review statuses: %#v", got)
	}

	got = reviewQueueStatuses(marketv1.MarketplaceReviewStatusRejected)
	if !reflect.DeepEqual(got, []string{marketv1.MarketplaceReviewStatusRejected.String()}) {
		t.Fatalf("unexpected explicit review status: %#v", got)
	}
}

func TestClassifyPluginListStatusFilter(t *testing.T) {
	cases := []struct {
		name       string
		status     string
		wantColumn string
		wantValue  string
		wantOK     bool
	}{
		{
			name:       "pending verify",
			status:     marketv1.MarketplaceProcessStatusPendingVerify.String(),
			wantColumn: "process_status",
			wantValue:  marketv1.MarketplaceProcessStatusPendingVerify.String(),
			wantOK:     true,
		},
		{
			name:       "pending review",
			status:     marketv1.MarketplaceProcessStatusPendingReview.String(),
			wantColumn: "process_status",
			wantValue:  marketv1.MarketplaceProcessStatusPendingReview.String(),
			wantOK:     true,
		},
		{
			name:       "failed",
			status:     marketv1.MarketplaceProcessStatusFailed.String(),
			wantColumn: "process_status",
			wantValue:  marketv1.MarketplaceProcessStatusFailed.String(),
			wantOK:     true,
		},
		{
			name:       "published",
			status:     marketv1.MarketplaceStatusPublished.String(),
			wantColumn: "market_status",
			wantValue:  marketv1.MarketplaceStatusPublished.String(),
			wantOK:     true,
		},
		{
			name:       "delisted",
			status:     marketv1.MarketplaceStatusDelisted.String(),
			wantColumn: "market_status",
			wantValue:  marketv1.MarketplaceStatusDelisted.String(),
			wantOK:     true,
		},
		{
			name:   "unknown ignored",
			status: "not-a-status",
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			column, value, ok := classifyPluginListStatusFilter(tc.status)
			if ok != tc.wantOK || column != tc.wantColumn || value != tc.wantValue {
				t.Fatalf(
					"got column=%q value=%q ok=%t, want column=%q value=%q ok=%t",
					column,
					value,
					ok,
					tc.wantColumn,
					tc.wantValue,
					tc.wantOK,
				)
			}
		})
	}
}

func TestPublisherIDsFromPluginsDeduplicates(t *testing.T) {
	ids := publisherIDsFromPlugins([]*entity.PluginMarketplacePlugin{
		{PublisherId: 1},
		{PublisherId: 1},
		{PublisherId: 2},
		nil,
		{PublisherId: 0},
	})
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("unexpected publisher ids: %#v", ids)
	}
}

func TestLatestReleaseIDsFromPluginsSkipsEmpty(t *testing.T) {
	ids := latestReleaseIDsFromPlugins([]*entity.PluginMarketplacePlugin{
		{LatestReleaseId: 10},
		{LatestReleaseId: 0},
		{LatestReleaseId: 10},
		{LatestReleaseId: 11},
	})
	if len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
		t.Fatalf("unexpected release ids: %#v", ids)
	}
}

func TestPluginRecordIDsFromReleasesDeduplicates(t *testing.T) {
	ids := pluginRecordIDsFromReleases([]*entity.PluginMarketplaceRelease{
		{PluginRecordId: 5, ReviewStatus: marketv1.MarketplaceReviewStatusSubmitted.String()},
		{PluginRecordId: 5},
		{PluginRecordId: 6},
	})
	if len(ids) != 2 || ids[0] != 5 || ids[1] != 6 {
		t.Fatalf("unexpected plugin record ids: %#v", ids)
	}
}

func TestTagCodesByPluginRecordGroupsAndDeduplicatesRows(t *testing.T) {
	rows := []*entity.PluginMarketplacePluginTag{
		{PluginRecordId: 10, TagCode: "observability"},
		{PluginRecordId: 20, TagCode: "workflow"},
		{PluginRecordId: 10, TagCode: " audit "},
		{PluginRecordId: 10, TagCode: "observability"},
		{PluginRecordId: 0, TagCode: "invalid"},
		nil,
	}

	got := tagCodesByPluginRecord(rows)
	if !reflect.DeepEqual(got[10], []string{"observability", "audit"}) {
		t.Fatalf("unexpected plugin 10 tag codes: %#v", got[10])
	}
	if !reflect.DeepEqual(got[20], []string{"workflow"}) {
		t.Fatalf("unexpected plugin 20 tag codes: %#v", got[20])
	}
	if _, ok := got[0]; ok {
		t.Fatalf("unexpected invalid plugin record group: %#v", got[0])
	}
}

func newSelectCaptureDB(t *testing.T) gdb.DB {
	t.Helper()
	db, err := gdb.New(gdb.ConfigNode{Type: "default"})
	if err != nil {
		t.Fatalf("create select capture database: %v", err)
	}
	db.SetDryRun(true)
	return db
}

func selectCaptureHook(captured *[]capturedSelect) gdb.HookHandler {
	return gdb.HookHandler{
		Select: func(_ context.Context, in *gdb.HookSelectInput) (gdb.Result, error) {
			*captured = append(*captured, capturedSelect{
				args: append([]any(nil), in.Args...),
				sql:  in.Sql,
			})
			return gdb.Result{}, nil
		},
	}
}

func containsArgument(args []any, expected any) bool {
	for _, value := range args {
		if reflect.DeepEqual(value, expected) {
			return true
		}
	}
	return false
}
