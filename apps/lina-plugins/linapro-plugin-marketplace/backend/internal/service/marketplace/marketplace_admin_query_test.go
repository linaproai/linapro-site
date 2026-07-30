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

func TestPluginIdentitySortFieldWhitelist(t *testing.T) {
	cols := dao.PluginMarketplacePlugin.Columns()
	cases := []struct {
		orderBy string
		want    string
	}{
		{orderBy: "pluginId", want: cols.PluginId},
		{orderBy: "plugin_id", want: cols.PluginId},
		{orderBy: "marketStatus", want: cols.MarketStatus},
		{orderBy: "status", want: cols.MarketStatus},
		{orderBy: "downloadCount", want: cols.DownloadCount},
		{orderBy: "updatedAt", want: cols.UpdatedAt},
		{orderBy: "name", want: ""},
		{orderBy: "id; drop table", want: ""},
		{orderBy: "", want: ""},
	}
	for _, tc := range cases {
		got := pluginIdentitySortField(tc.orderBy)
		if got != tc.want {
			t.Fatalf("orderBy=%q got %q want %q", tc.orderBy, got, tc.want)
		}
	}
}

func TestApplyPluginIdentityListOrderDefaults(t *testing.T) {
	captured := []capturedSelect{}
	db := newSelectCaptureDB(t)

	// Owned default: plugin_id ASC
	owned := applyPluginIdentityListOrder(
		db.Model(dao.PluginMarketplacePlugin.Table()).Hook(selectCaptureHook(&captured)),
		"",
		"",
		true,
	)
	if _, err := owned.All(); err != nil {
		t.Fatalf("owned default order query: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected one owned query, got %d", len(captured))
	}
	ownedSQL := strings.ToLower(captured[0].sql)
	if !strings.Contains(ownedSQL, "order by") || !strings.Contains(ownedSQL, "plugin_id") {
		t.Fatalf("owned default order missing plugin_id: %s", captured[0].sql)
	}

	captured = nil
	// Managed default: updated_at DESC
	managed := applyPluginIdentityListOrder(
		db.Model(dao.PluginMarketplacePlugin.Table()).Hook(selectCaptureHook(&captured)),
		"",
		"",
		false,
	)
	if _, err := managed.All(); err != nil {
		t.Fatalf("managed default order query: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected one managed query, got %d", len(captured))
	}
	managedSQL := strings.ToLower(captured[0].sql)
	if !strings.Contains(managedSQL, "updated_at") {
		t.Fatalf("managed default order missing updated_at: %s", captured[0].sql)
	}

	captured = nil
	// Explicit downloadCount desc
	sorted := applyPluginIdentityListOrder(
		db.Model(dao.PluginMarketplacePlugin.Table()).Hook(selectCaptureHook(&captured)),
		"downloadCount",
		"desc",
		true,
	)
	if _, err := sorted.All(); err != nil {
		t.Fatalf("explicit sort query: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected one sorted query, got %d", len(captured))
	}
	sortedSQL := strings.ToLower(captured[0].sql)
	if !strings.Contains(sortedSQL, "download_count") {
		t.Fatalf("explicit sort missing download_count: %s", captured[0].sql)
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

func TestReviewQueuePluginNameUsesRequestedLocale(t *testing.T) {
	row := &entity.PluginMarketplaceRelease{Id: 7, PluginRecordId: 5}
	displayByRelease := map[int][]*entity.PluginMarketplaceDisplayI18n{
		7: {
			{Locale: "en-US", Name: "Object Storage - S3"},
			{Locale: "zh-CN", Name: "对象存储-S3"},
		},
	}
	pluginNames := map[int]string{5: "Object Storage - S3"}

	got := reviewQueuePluginName(row, pluginNames, displayByRelease, "zh-CN")
	if got != "对象存储-S3" {
		t.Fatalf("expected localized review queue name, got %q", got)
	}

	got = reviewQueuePluginName(row, pluginNames, nil, "zh-CN")
	if got != pluginNames[5] {
		t.Fatalf("expected identity fallback name, got %q", got)
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
