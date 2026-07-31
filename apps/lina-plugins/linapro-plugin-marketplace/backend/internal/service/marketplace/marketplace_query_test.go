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

func TestPluginDetailItemFromEntitiesProjectsSourceFields(t *testing.T) {
	now := time.UnixMilli(1767247200000)
	detail := pluginDetailItemFromEntities(
		&entity.PluginMarketplacePlugin{
			PluginId:        "linapro-demo-source",
			Name:            "Demo Source",
			Summary:         "Demo summary",
			Description:     "Long description",
			PluginType:      marketv1.MarketplacePluginTypeSource.String(),
			MarketStatus:    marketv1.MarketplaceStatusPublished.String(),
			ProcessStatus:   marketv1.MarketplaceProcessStatusCompleted.String(),
			Visibility:      marketv1.MarketplaceVisibilityPublic.String(),
			LatestVersion:   "v0.2.0",
			Repository:      "https://github.com/example/demo-source",
			License:         "Apache-2.0",
			DownloadCount:   3,
			SourceKind:      gitSourceKind,
			RepoUrl:         "https://github.com/example/demo-source.git",
			RepoProvider:    marketv1.MarketplaceRepoProviderGitHub.String(),
			RepoPath:        "apps/lina-plugins/linapro-demo-source",
			CredentialRef:   "cred-ref-1",
			LastSyncStatus:  "success",
			LastSyncMessage: "discovered 2 draft releases",
			LastSyncAt:      &now,
			PublishedAt:     &now,
			UpdatedAt:       &now,
		},
		&entity.PluginMarketplacePublisher{
			PublisherKey: "linapro",
			Name:         "LinaPro",
			Verified:     true,
		},
		nil,
		nil,
		nil,
	)
	if detail == nil {
		t.Fatal("expected detail projection")
	}
	if detail.SourceKind != marketv1.MarketplaceSourceKindGit {
		t.Fatalf("expected git source kind, got %#v", detail.SourceKind)
	}
	if detail.RepoUrl != "https://github.com/example/demo-source.git" {
		t.Fatalf("unexpected repo url: %#v", detail.RepoUrl)
	}
	if detail.RepoProvider != marketv1.MarketplaceRepoProviderGitHub {
		t.Fatalf("unexpected repo provider: %#v", detail.RepoProvider)
	}
	if detail.RepoPath != "apps/lina-plugins/linapro-demo-source" {
		t.Fatalf("unexpected repo path: %#v", detail.RepoPath)
	}
	if !detail.RequiresAuth {
		t.Fatal("expected RequiresAuth when credential ref is present")
	}
	if detail.LastSyncStatus != "success" || detail.LastSyncMessage != "discovered 2 draft releases" {
		t.Fatalf("unexpected last sync fields: status=%q message=%q", detail.LastSyncStatus, detail.LastSyncMessage)
	}
	if detail.LastSyncAt == nil || *detail.LastSyncAt != now.UnixMilli() {
		t.Fatalf("unexpected lastSyncAt: %#v", detail.LastSyncAt)
	}

	uploadDetail := pluginDetailItemFromEntities(
		&entity.PluginMarketplacePlugin{
			PluginId:     "linapro-demo-upload",
			Name:         "Demo Upload",
			PluginType:   marketv1.MarketplacePluginTypeDynamic.String(),
			MarketStatus: marketv1.MarketplaceStatusDraft.String(),
			Visibility:   marketv1.MarketplaceVisibilityPrivate.String(),
			// Empty source kind must normalize to upload, matching list projection.
			SourceKind: "",
		},
		nil,
		nil,
		nil,
		nil,
	)
	if uploadDetail == nil {
		t.Fatal("expected upload detail projection")
	}
	if uploadDetail.SourceKind != marketv1.MarketplaceSourceKindUpload {
		t.Fatalf("expected empty source kind to normalize to upload, got %#v", uploadDetail.SourceKind)
	}
	if uploadDetail.RequiresAuth {
		t.Fatal("expected RequiresAuth false without credential ref")
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
	if item.Disposition != marketv1.MarketplaceRiskDispositionNeedAttention {
		t.Fatalf("expected need_attention disposition, got %s", item.Disposition)
	}
	if item.Blocking {
		t.Fatalf("host services finding must not block submit")
	}
	if item.CreatedAt == nil || *item.CreatedAt != now.UnixMilli() {
		t.Fatalf("unexpected createdAt: %#v", item.CreatedAt)
	}

	framework := riskItemFromEntity(&entity.PluginMarketplaceRisk{
		RiskType: marketv1.MarketplaceRiskTypeDependency.String(),
		Severity: marketv1.MarketplaceRiskSeverityWarning.String(),
		Source:   "plugin.yaml",
		Summary:  "Framework compatibility dependency is not declared.",
		// Legacy payload may still embed blocking=true; policy must win on read.
		Payload: `{"code":"framework_dependency_missing","blocking":true,"disposition":"need_fix"}`,
	})
	if framework.Disposition != marketv1.MarketplaceRiskDispositionNeedAttention || framework.Blocking {
		t.Fatalf("expected non-blocking framework finding, got disposition=%s blocking=%v", framework.Disposition, framework.Blocking)
	}

	blocking := riskItemFromEntity(&entity.PluginMarketplaceRisk{
		RiskType: marketv1.MarketplaceRiskTypeMultiTenant.String(),
		Severity: marketv1.MarketplaceRiskSeverityWarning.String(),
		Source:   "manifest/i18n",
		Summary:  "Plugin declares i18n.enabled but no manifest i18n JSON files were detected.",
		Payload:  `{"code":"i18n_files_missing"}`,
	})
	if blocking.Disposition != marketv1.MarketplaceRiskDispositionNeedFix || !blocking.Blocking {
		t.Fatalf("expected need_fix blocking finding, got disposition=%s blocking=%v", blocking.Disposition, blocking.Blocking)
	}
}

// TestSortMarketplaceRiskItemsBySeverity pins presentation order:
// blocking/need_fix first, then disposition, then severity within the same bucket.
func TestSortMarketplaceRiskItemsBySeverity(t *testing.T) {
	t.Parallel()

	items := []*marketv1.MarketplaceRiskItem{
		{
			Severity: marketv1.MarketplaceRiskSeverityInfo,
			Summary:  "info-a",
			Payload:  map[string]any{"code": "source_docs_indexed"},
		},
		{
			Severity: marketv1.MarketplaceRiskSeverityHigh,
			Summary:  "high-attention",
			Payload:  map[string]any{"code": "dynamic_host_services_present"},
		},
		{
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Summary:  "fix-a",
			Payload:  map[string]any{"code": "i18n_files_missing"},
		},
		{
			Severity: marketv1.MarketplaceRiskSeverityInfo,
			Summary:  "info-b",
			Payload:  map[string]any{"code": "dynamic_runtime_detected"},
		},
		{
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Summary:  "attention-a",
			Payload:  map[string]any{"code": "source_sql_present"},
		},
		{
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Summary:  "framework-attention",
			Payload:  map[string]any{"code": "framework_dependency_missing"},
		},
	}
	sortMarketplaceRiskItemsBySeverity(items)

	// need_fix first, then need_attention by severity, then info_only.
	want := []string{"fix-a", "high-attention", "attention-a", "framework-attention", "info-a", "info-b"}
	got := make([]string, 0, len(items))
	for _, item := range items {
		got = append(got, item.Summary)
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("risk order mismatch at %d: got %v want %v", i, got, want)
		}
	}

	// Empty and single-element slices are no-ops.
	sortMarketplaceRiskItemsBySeverity(nil)
	sortMarketplaceRiskItemsBySeverity([]*marketv1.MarketplaceRiskItem{
		{Severity: marketv1.MarketplaceRiskSeverityInfo, Summary: "only"},
	})
}

func TestBuildPackageDiagnosticRiskPayloadIncludesEvidence(t *testing.T) {
	t.Parallel()

	payload := buildPackageDiagnosticRiskPayload(&PackageDiagnostic{
		Code:     "source_sql_present",
		Severity: marketv1.MarketplaceRiskSeverityWarning,
		Source:   "manifest/sql",
		Message:  "SQL present",
		Evidence: &PackageDiagnosticEvidence{
			Files:      []string{"manifest/sql/001-init.sql"},
			TotalCount: 1,
		},
	})
	if payload.Code != "source_sql_present" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if payload.Disposition != marketv1.MarketplaceRiskDispositionNeedAttention.String() {
		t.Fatalf("unexpected disposition: %s", payload.Disposition)
	}
	if payload.Blocking {
		t.Fatalf("SQL finding must not be blocking")
	}
	if len(payload.Files) != 1 || payload.Files[0] != "manifest/sql/001-init.sql" {
		t.Fatalf("unexpected files: %#v", payload.Files)
	}

	frameworkPayload := buildPackageDiagnosticRiskPayload(&PackageDiagnostic{
		Code: "framework_dependency_missing",
		Evidence: &PackageDiagnosticEvidence{
			ExpectedPath:  "plugin.yaml",
			ExpectedField: "dependencies.framework.version",
			Example:       ">=1.0.0 <2.0.0",
		},
	})
	if frameworkPayload.Blocking || frameworkPayload.Disposition != marketv1.MarketplaceRiskDispositionNeedAttention.String() {
		t.Fatalf("unexpected framework payload: %#v", frameworkPayload)
	}

	fixPayload := buildPackageDiagnosticRiskPayload(&PackageDiagnostic{
		Code: "i18n_files_missing",
		Evidence: &PackageDiagnosticEvidence{
			ExpectedPath:  "manifest/i18n",
			ExpectedField: "i18n.enabled / locale JSON bundles",
			Example:       "manifest/i18n/zh-CN/plugin.json",
		},
	})
	if !fixPayload.Blocking || fixPayload.Disposition != marketv1.MarketplaceRiskDispositionNeedFix.String() {
		t.Fatalf("unexpected fix payload: %#v", fixPayload)
	}
}

func TestSourcePackageDiagnosticsCarryEvidence(t *testing.T) {
	t.Parallel()

	diagnostics := sourcePackageDiagnostics(
		&sourcePackageManifest{ID: "demo", Name: "Demo", Version: "0.1.0"},
		[]*sourcePackageResourceSummary{{Path: "manifest/sql/001.sql", Kind: "sql"}},
		nil,
		[]*sourcePackageResourceSummary{{Path: "manifest/docs/zh-CN/index.md", Kind: "marketplace_doc"}},
	)
	byCode := map[string]*PackageDiagnostic{}
	for _, item := range diagnostics {
		byCode[item.Code] = item
	}
	if byCode["source_sql_present"] == nil || byCode["source_sql_present"].Evidence == nil {
		t.Fatalf("expected SQL evidence, got %#v", byCode["source_sql_present"])
	}
	if len(byCode["source_sql_present"].Evidence.Files) != 1 {
		t.Fatalf("expected one SQL path, got %#v", byCode["source_sql_present"].Evidence)
	}
	if byCode["framework_dependency_missing"] == nil {
		t.Fatalf("expected framework dependency finding")
	}
	if resolveRiskDispositionPolicy("framework_dependency_missing").Blocking {
		t.Fatalf("framework dependency finding must not block submit")
	}
	if resolveRiskDispositionPolicy("framework_dependency_missing").Disposition != marketv1.MarketplaceRiskDispositionNeedAttention {
		t.Fatalf("framework dependency finding must be need_attention")
	}
	if byCode["source_docs_indexed"] == nil || byCode["source_docs_indexed"].Evidence == nil {
		t.Fatalf("expected docs evidence")
	}
}

func TestDocumentItemFromRecordUsesRenderedContentSnapshot(t *testing.T) {
	item := documentItemFromRecord(&DocumentRecord{
		PluginID:        "linapro-demo-source",
		Version:         "v0.1.0",
		RequestedLocale: "en-US",
		ResolvedLocale:  "zh-CN",
		Path:            "index.md",
		SourceKind:      documentSourceKindManifestDocs,
		Title:           "Demo",
		Summary:         "Summary",
		RenderedContent: `<p>&lt;script&gt;alert("x")&lt;/script&gt;</p>`,
		ContentHash:     strings.Repeat("a", 64),
		FallbackUsed:    true,
	})

	if strings.Contains(item.Content, "<script>") {
		t.Fatalf("expected rendered snapshot to remain safe, got %s", item.Content)
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

// TestBuildSourceRiskSummaryMatchesDiagnosticSeverityCount guards the invariant
// that the aggregated risk_summary counts mirror the diagnostic rows that
// replaceReleaseRisks persists to plugin_marketplace_risk. The Git discovery
// path previously wrote only the summary via buildSourceRiskSummary and skipped
// the detail rows, so owners saw a populated summary ("警告 2 提示 1") alongside
// an empty risk page. Both projections derive from the same diagnostics slice,
// so the summary total must equal the number of risk rows replaceReleaseRisks
// writes; this test pins that invariant against future divergence.
func TestBuildSourceRiskSummaryMatchesDiagnosticSeverityCount(t *testing.T) {
	t.Parallel()

	diagnostics := []*PackageDiagnostic{
		{Code: "dynamic_host_services_present", Severity: marketv1.MarketplaceRiskSeverityHigh},
		{Code: "dynamic_routes_present", Severity: marketv1.MarketplaceRiskSeverityWarning},
		{Code: "framework_dependency_missing", Severity: marketv1.MarketplaceRiskSeverityWarning},
		{Code: "dynamic_mock_sql_present", Severity: marketv1.MarketplaceRiskSeverityInfo},
	}

	summary := buildSourceRiskSummary(diagnostics)
	if summary.High != 1 || summary.Warning != 2 || summary.Info != 1 {
		t.Fatalf("unexpected severity counts: high=%d warning=%d info=%d", summary.High, summary.Warning, summary.Info)
	}
	if total := summary.High + summary.Warning + summary.Info; total != len(diagnostics) {
		t.Fatalf("summary total %d must equal diagnostic count %d so risk page rows never diverge from the summary", total, len(diagnostics))
	}

	// Nil diagnostics and nil entries must not inflate counts; replaceReleaseRisks
	// mirrors this by skipping nil diagnostics when persisting risk rows.
	empty := buildSourceRiskSummary(nil)
	if empty.High+empty.Warning+empty.Info != 0 {
		t.Fatalf("nil diagnostics must yield zero counts, got %#v", empty)
	}
	withNil := buildSourceRiskSummary([]*PackageDiagnostic{
		nil,
		{Code: "dynamic_routes_present", Severity: marketv1.MarketplaceRiskSeverityWarning},
	})
	if withNil.High+withNil.Warning+withNil.Info != 1 {
		t.Fatalf("nil diagnostic entry must be skipped, got %#v", withNil)
	}
}
