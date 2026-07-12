// This file verifies marketplace read projection helpers that do not require a
// live database. Database-backed pagination and visibility cases are covered by
// the later integration task once data-permission filters are wired.

package marketplace

import (
	"strings"
	"testing"
	"time"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestNormalizeMarketplacePageBounds(t *testing.T) {
	pageNum, pageSize := normalizeMarketplacePage(0, 0)
	if pageNum != defaultMarketplacePageNum || pageSize != defaultMarketplacePageSize {
		t.Fatalf("unexpected default page bounds: pageNum=%d pageSize=%d", pageNum, pageSize)
	}

	pageNum, pageSize = normalizeMarketplacePage(3, maxMarketplacePageSize+10)
	if pageNum != 3 || pageSize != maxMarketplacePageSize {
		t.Fatalf("unexpected capped page bounds: pageNum=%d pageSize=%d", pageNum, pageSize)
	}
}

func TestPluginListReadModelBaseCriteriaRequiresPublished(t *testing.T) {
	criteria := pluginListReadModelBaseCriteria()
	if criteria.MarketStatus != marketv1.MarketplaceStatusPublished.String() {
		t.Fatalf("unexpected list base criteria: %#v", criteria.MarketStatus)
	}
}

func TestMarketplaceManagementReadAllowed(t *testing.T) {
	tests := []struct {
		name    string
		owned   bool
		subject VisibilitySubject
		want    bool
	}{
		{name: "publisher owner", owned: true, subject: VisibilitySubject{CanPublish: true}, want: true},
		{name: "reviewer", subject: VisibilitySubject{CanReview: true}, want: true},
		{name: "owner without publish permission", owned: true, subject: VisibilitySubject{UserID: 1001}, want: false},
		{name: "unrelated authenticated user", subject: VisibilitySubject{UserID: 1001}, want: false},
		{name: "anonymous", subject: VisibilitySubject{}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := marketplaceManagementReadAllowed(tc.owned, tc.subject); got != tc.want {
				t.Fatalf("expected management access %t, got %t", tc.want, got)
			}
		})
	}
}

func TestMarketplacePublishedReadAllowed(t *testing.T) {
	tests := []struct {
		name    string
		owned   bool
		subject VisibilitySubject
		want    bool
	}{
		{name: "public reader", subject: VisibilitySubject{UserID: 1001}, want: true},
		{name: "publisher owner", owned: true, subject: VisibilitySubject{CanPublish: true}, want: true},
		{name: "publisher non owner", subject: VisibilitySubject{CanPublish: true}, want: false},
		{name: "reviewer", subject: VisibilitySubject{CanReview: true}, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := marketplacePublishedReadAllowed(tc.owned, tc.subject); got != tc.want {
				t.Fatalf("expected published access %t, got %t", tc.want, got)
			}
		})
	}
}

func TestMarketplaceReleaseStatusFilter(t *testing.T) {
	tests := []struct {
		name       string
		requested  string
		management bool
		wantStatus string
		wantOK     bool
	}{
		{name: "public default", wantStatus: "published", wantOK: true},
		{name: "public published", requested: "published", wantStatus: "published", wantOK: true},
		{name: "public draft rejected", requested: "draft", wantOK: false},
		{name: "reviewer draft", requested: "draft", management: true, wantStatus: "draft", wantOK: true},
		{name: "reviewer all", management: true, wantOK: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, ok := marketplaceReleaseStatusFilter(tc.requested, tc.management)
			if status != tc.wantStatus || ok != tc.wantOK {
				t.Fatalf("expected status=%q ok=%t, got status=%q ok=%t", tc.wantStatus, tc.wantOK, status, ok)
			}
		})
	}
}

func TestMarketplaceReleaseManagementReadDoesNotExpandDownload(t *testing.T) {
	subject := VisibilitySubject{CanReview: true}
	if !marketplaceReleaseManagementReadAllowed(false, subject, marketplaceVisibilityPermissionView) {
		t.Fatal("expected reviewer to inspect unpublished release details")
	}
	if marketplaceReleaseManagementReadAllowed(false, subject, marketplaceVisibilityPermissionDownload) {
		t.Fatal("expected reviewer state not to expand unpublished download access")
	}
	publisher := VisibilitySubject{CanPublish: true}
	if marketplaceReleaseManagementReadAllowed(true, publisher, marketplaceVisibilityPermissionDownload) {
		t.Fatal("expected publisher state not to expand unpublished download access")
	}
}

func TestPluginListItemFromReadModelDecodesSnapshots(t *testing.T) {
	now := time.UnixMilli(1767247200000)
	item := pluginListItemFromReadModel(&entity.PluginMarketplacePluginReadModel{
		PluginId:          "linapro-demo-source",
		Name:              "Demo Source",
		Summary:           "Demo summary",
		PublisherName:     "LinaPro",
		PublisherVerified: true,
		PluginType:        marketv1.MarketplacePluginTypeSource.String(),
		MarketStatus:      marketv1.MarketplaceStatusPublished.String(),
		Visibility:        marketv1.MarketplaceVisibilityPublic.String(),
		LatestVersion:     "v0.1.0",
		PrimaryTag:        "observability",
		TagCodes:          `["observability","audit"]`,
		RiskCounts:        `{"info":2,"warning":1,"high":0}`,
		DownloadCount:     12,
		PublishedAt:       &now,
		UpdatedAt:         &now,
	}, nil)

	if item.Publisher.Name != "LinaPro" || !item.Publisher.Verified {
		t.Fatalf("expected publisher fallback snapshot, got %#v", item.Publisher)
	}
	if len(item.TagCodes) != 2 || item.TagCodes[1] != "audit" {
		t.Fatalf("unexpected tag codes: %#v", item.TagCodes)
	}
	if item.RiskCounts.Info != 2 || item.RiskCounts.Warning != 1 || item.RiskCounts.High != 0 {
		t.Fatalf("unexpected risk counts: %#v", item.RiskCounts)
	}
	if item.PublishedAt == nil || *item.PublishedAt != now.UnixMilli() {
		t.Fatalf("unexpected publishedAt: %#v", item.PublishedAt)
	}
}

func TestArtifactPriorityPrefersPrimaryPackage(t *testing.T) {
	dynamicZip := &entity.PluginMarketplaceArtifact{ArtifactType: marketv1.MarketplaceArtifactTypeDynamicZip.String()}
	pluginWasm := &entity.PluginMarketplaceArtifact{ArtifactType: marketv1.MarketplaceArtifactTypePluginWasm.String()}
	if artifactPriority(dynamicZip, marketv1.MarketplacePluginTypeDynamic) >=
		artifactPriority(pluginWasm, marketv1.MarketplacePluginTypeDynamic) {
		t.Fatal("expected dynamic ZIP to be the primary dynamic artifact")
	}

	sourceZip := &entity.PluginMarketplaceArtifact{ArtifactType: marketv1.MarketplaceArtifactTypeSourceZip.String()}
	if artifactPriority(sourceZip, marketv1.MarketplacePluginTypeSource) != 1 {
		t.Fatal("expected source ZIP to be the primary source artifact")
	}
}

func TestRiskItemFromEntityParsesPayload(t *testing.T) {
	now := time.UnixMilli(1767240000000)
	item := riskItemFromEntity(&entity.PluginMarketplaceRisk{
		RiskType:  marketv1.MarketplaceRiskTypeHostService.String(),
		Severity:  marketv1.MarketplaceRiskSeverityHigh.String(),
		Source:    "hostServices",
		Summary:   "Requests data access",
		Payload:   `{"code":"dynamic_host_services_present"}`,
		CreatedAt: &now,
	})

	if item.Type != marketv1.MarketplaceRiskTypeHostService || item.Severity != marketv1.MarketplaceRiskSeverityHigh {
		t.Fatalf("unexpected risk item enum projection: %#v", item)
	}
	if item.Payload["code"] != "dynamic_host_services_present" {
		t.Fatalf("unexpected risk payload: %#v", item.Payload)
	}
	if item.CreatedAt == nil || *item.CreatedAt != now.UnixMilli() {
		t.Fatalf("unexpected createdAt: %#v", item.CreatedAt)
	}
}

func TestDocumentItemFromRecordEscapesIndexedText(t *testing.T) {
	item := documentItemFromRecord(&DocumentRecord{
		PluginID:        "linapro-demo-source",
		Version:         "v0.1.0",
		RequestedLocale: "en-US",
		ResolvedLocale:  "zh-CN",
		Path:            "index.md",
		SourceKind:      documentSourceKindManifestDocs,
		Title:           "Demo",
		Summary:         "Summary",
		SearchText:      `<script>alert("x")</script>`,
		ContentHash:     strings.Repeat("a", 64),
		FallbackUsed:    true,
	})

	if strings.Contains(item.Content, "<script>") {
		t.Fatalf("expected indexed content to be escaped, got %s", item.Content)
	}
	if !strings.Contains(item.Content, "&lt;script&gt;") {
		t.Fatalf("expected escaped script marker, got %s", item.Content)
	}
	if !item.FallbackUsed || item.ResolvedLocale != "zh-CN" {
		t.Fatalf("unexpected document fallback metadata: %#v", item)
	}
}

func TestDiagnosticRiskTypeClassification(t *testing.T) {
	cases := []struct {
		name     string
		input    *PackageDiagnostic
		expected marketv1.MarketplaceRiskType
	}{
		{
			name:     "host services",
			input:    &PackageDiagnostic{Code: "dynamic_host_services_present"},
			expected: marketv1.MarketplaceRiskTypeHostService,
		},
		{
			name:     "dynamic routes",
			input:    &PackageDiagnostic{Code: "dynamic_routes_present"},
			expected: marketv1.MarketplaceRiskTypeDynamicRoute,
		},
		{
			name:     "mock sql",
			input:    &PackageDiagnostic{Code: "dynamic_mock_sql_present"},
			expected: marketv1.MarketplaceRiskTypeMockSQL,
		},
		{
			name:     "dependency",
			input:    &PackageDiagnostic{Code: "framework_dependency_missing"},
			expected: marketv1.MarketplaceRiskTypeDependency,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := diagnosticRiskType(tc.input); got != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}
