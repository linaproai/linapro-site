// This file verifies marketplace document indexing safety checks, conservative
// rendering, locale fallback selection, and rendered-document cache keys.

package marketplace

import (
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
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

func TestSelectMarketplaceDocumentFallbackUsesDefaultLocale(t *testing.T) {
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{{
		PluginID:   "linapro-demo-source",
		Version:    "v0.1.0",
		Locale:     "zh-CN",
		Path:       "index.md",
		SourceKind: documentSourceKindManifestDocs,
		Title:      "中文文档",
	}}, GetReleaseDocumentInput{
		PluginID:      "linapro-demo-source",
		Version:       "v0.1.0",
		Locale:        "en-US",
		DefaultLocale: "zh-CN",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected fallback document")
	}
	if record.ResolvedLocale != "zh-CN" || !record.FallbackUsed {
		t.Fatalf("unexpected fallback metadata: %#v", record)
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
