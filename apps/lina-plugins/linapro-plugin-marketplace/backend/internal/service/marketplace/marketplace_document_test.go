// This file verifies marketplace document indexing safety checks, conservative
// rendering, locale fallback selection, and rendered-document cache keys.

package marketplace

import (
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestIndexMarketplaceDocumentEscapesScript(t *testing.T) {
	item, err := indexMarketplaceDocument(
		"en-US",
		"index.md",
		documentSourceKindManifestDocs,
		"# Safe Title\n<script>alert('x')</script>\n",
	)
	if err != nil {
		t.Fatalf("indexMarketplaceDocument returned error: %v", err)
	}
	if item.Title != "Safe Title" {
		t.Fatalf("expected title Safe Title, got %q", item.Title)
	}
	if strings.Contains(item.RenderedContent, "<script") {
		t.Fatalf("rendered content should not contain executable script tag: %s", item.RenderedContent)
	}
	if !strings.Contains(item.RenderedContent, "&lt;script&gt;") {
		t.Fatalf("expected script tag to be escaped, got %s", item.RenderedContent)
	}
}

func TestIndexMarketplaceDocumentRejectsTraversalImage(t *testing.T) {
	_, err := indexMarketplaceDocument(
		"en-US",
		"index.md",
		documentSourceKindManifestDocs,
		"![bad](../../secret.png)\n",
	)
	if !bizerr.Is(err, CodeMarketplaceDocumentInvalid) {
		t.Fatalf("expected document invalid error, got %v", err)
	}
}

func TestDocumentRecordFromIndexItemMapsReleaseAndRender(t *testing.T) {
	record := documentRecordFromIndexItem(
		&entity.PluginMarketplaceRelease{
			Id:             7,
			PluginId:       "linapro-demo-source",
			ReleaseVersion: "v0.1.0",
		},
		&marketplaceDocumentIndexItem{
			Locale:          "en-US",
			DocPath:         "index.md",
			SourceKind:      documentSourceKindManifestDocs,
			Title:           "English docs",
			Summary:         "Summary",
			ContentHash:     strings.Repeat("a", 64),
			SearchText:      "Plain text",
			RenderedContent: "<h1>English docs</h1>\n",
		},
	)
	if record == nil {
		t.Fatal("expected document record")
	}
	if record.RenderedContent != "<h1>English docs</h1>\n" {
		t.Fatalf("expected rendered content from disk index item, got %#v", record.RenderedContent)
	}
	if record.PluginID != "linapro-demo-source" || record.Path != "index.md" {
		t.Fatalf("unexpected record projection: %#v", record)
	}
}

func TestSelectMarketplaceDocumentFallbackSingleLocaleAlwaysUsed(t *testing.T) {
	// Only one language is indexed: show it even when the user asks for English.
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{{
		PluginID:   "linapro-demo-source",
		Version:    "v0.1.0",
		Locale:     "zh-CN",
		Path:       "index.md",
		SourceKind: documentSourceKindManifestDocs,
		Title:      "中文文档",
	}}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "en-US",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected single-locale document")
	}
	if record.ResolvedLocale != "zh-CN" || !record.FallbackUsed {
		t.Fatalf("unexpected single-locale metadata: %#v", record)
	}
}

func TestCollectMarketplaceDocumentBundleReturnsSamePathLanguages(t *testing.T) {
	records := []*DocumentRecord{
		{
			PluginID:        "linapro-demo-source",
			Version:         "v0.1.0",
			Locale:          "en-US",
			Path:            "index.md",
			SourceKind:      documentSourceKindManifestDocs,
			Title:           "English docs",
			RenderedContent: "<h1>English docs</h1>\n",
		},
		{
			PluginID:        "linapro-demo-source",
			Version:         "v0.1.0",
			Locale:          "zh-CN",
			Path:            "index.md",
			SourceKind:      documentSourceKindManifestDocs,
			Title:           "中文文档",
			RenderedContent: "<h1>中文文档</h1>\n",
		},
		{
			PluginID:        "linapro-demo-source",
			Version:         "v0.1.0",
			Locale:          "zh-CN",
			Path:            "guide.md",
			SourceKind:      documentSourceKindManifestDocs,
			Title:           "指南",
			RenderedContent: "<h1>指南</h1>\n",
		},
	}
	input := GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "zh-CN",
		Path:     "index.md",
	}
	selected, err := selectMarketplaceDocumentFallback(records, input)
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	bundle, err := collectMarketplaceDocumentBundle(records, selected, input)
	if err != nil {
		t.Fatalf("collectMarketplaceDocumentBundle returned error: %v", err)
	}
	items := documentItemsFromRecords(bundle)

	if len(items) != 2 {
		t.Fatalf("expected two same-path language documents, got %d", len(items))
	}
	if items[0].Path != "index.md" || items[1].Path != "index.md" {
		t.Fatalf("expected only same-path documents, got %#v", items)
	}
	if items[0].Content == "" || items[1].Content == "" {
		t.Fatalf("expected rendered content snapshots in bundle: %#v", items)
	}
}

func TestSelectMarketplaceDocumentFallbackPrefersUserLocaleWhenMultiple(t *testing.T) {
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "zh-CN",
			Path:       "index.md",
			SourceKind: documentSourceKindManifestDocs,
			Title:      "中文文档",
		},
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "en-US",
			Path:       "index.md",
			SourceKind: documentSourceKindManifestDocs,
			Title:      "English docs",
		},
	}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "zh-CN",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected user-locale document")
	}
	if record.ResolvedLocale != "zh-CN" || record.FallbackUsed {
		t.Fatalf("unexpected user-locale metadata: %#v", record)
	}
}

func TestSelectMarketplaceDocumentFallbackUsesEnglishWhenUnmatched(t *testing.T) {
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "zh-CN",
			Path:       "index.md",
			SourceKind: documentSourceKindManifestDocs,
			Title:      "中文文档",
		},
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "en-US",
			Path:       "index.md",
			SourceKind: documentSourceKindManifestDocs,
			Title:      "English docs",
		},
	}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "ja-JP",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected English fallback document")
	}
	if record.ResolvedLocale != "en-US" || !record.FallbackUsed {
		t.Fatalf("unexpected English fallback metadata: %#v", record)
	}
}

func TestSelectMarketplaceDocumentFallbackUsesReadme(t *testing.T) {
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{{
		PluginID:   "linapro-demo-source",
		Version:    "v0.1.0",
		Locale:     "en-US",
		Path:       "README.md",
		SourceKind: documentSourceKindReadme,
		Title:      "Readme",
	}}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "fr-FR",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected README fallback document")
	}
	if record.SourceKind != documentSourceKindReadme || !record.FallbackUsed {
		t.Fatalf("unexpected README fallback metadata: %#v", record)
	}
}

func TestSelectMarketplaceDocumentFallbackReadmeUsesEnglishWhenUnmatched(t *testing.T) {
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "zh-CN",
			Path:       readmeCNDocumentPath,
			SourceKind: documentSourceKindReadme,
			Title:      "中文 README",
		},
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "en-US",
			Path:       readmeDocumentPath,
			SourceKind: documentSourceKindReadme,
			Title:      "English README",
		},
	}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "ja-JP",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected English README fallback document")
	}
	if record.Path != readmeDocumentPath || record.ResolvedLocale != fallbackEnUSLocale || !record.FallbackUsed {
		t.Fatalf("unexpected README English fallback metadata: %#v", record)
	}
}
