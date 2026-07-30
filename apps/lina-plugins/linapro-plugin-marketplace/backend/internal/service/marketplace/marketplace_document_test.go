// This file verifies marketplace document indexing safety checks, conservative
// rendering, locale fallback selection, and rendered-document cache keys.

package marketplace

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
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

func TestOpenPackageDocumentArchiveReadsTarGzDocs(t *testing.T) {
	tarPath := writeDocumentPackageTarGz(t, "linapro-demo-source/", map[string]string{
		"manifest/docs/en-US/index.md": "# Tar Docs\n\nVisible from uploaded tar package.\n",
	})

	reader, cleanup, err := openPackageDocumentArchive(tarPath, "linapro-demo-source.tar.gz")
	if err != nil {
		t.Fatalf("openPackageDocumentArchive returned error: %v", err)
	}
	defer func() {
		if closeErr := cleanup(); closeErr != nil {
			t.Fatalf("cleanup package document archive: %v", closeErr)
		}
	}()

	fileIndex, err := indexSourceZipFiles(reader.File)
	if err != nil {
		t.Fatalf("indexSourceZipFiles returned error: %v", err)
	}
	items, err := buildZipDocumentIndex(fallbackEnUSLocale, fileIndex, "linapro-demo-source/")
	if err != nil {
		t.Fatalf("buildZipDocumentIndex returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one document item, got %d", len(items))
	}
	if items[0].DocPath != defaultDocumentPath || !strings.Contains(items[0].Markdown, "Tar Docs") {
		t.Fatalf("unexpected tar.gz document item: %#v", items[0])
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

func TestSelectMarketplaceDocumentFallbackUsesReadmeWhenOnlyReadmeExists(t *testing.T) {
	t.Parallel()
	record, err := selectMarketplaceDocumentFallback([]*DocumentRecord{
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "en-US",
			Path:       readmeDocumentPath,
			SourceKind: documentSourceKindReadme,
			Title:      "Readme",
		},
		{
			PluginID:   "linapro-demo-source",
			Version:    "v0.1.0",
			Locale:     "zh-CN",
			Path:       readmeCNDocumentPath,
			SourceKind: documentSourceKindReadme,
			Title:      "中文 README",
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
		t.Fatal("expected README fallback document")
	}
	if record.SourceKind != documentSourceKindReadme || record.Path != defaultDocumentPath || record.ResolvedLocale != fallbackZhCNLocale {
		t.Fatalf("unexpected README fallback metadata: %#v", record)
	}
}

func TestSelectMarketplaceDocumentFallbackIgnoresReadmeWhenManifestDocsExist(t *testing.T) {
	t.Parallel()
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
			Path:       "index.md",
			SourceKind: documentSourceKindManifestDocs,
			Title:      "Guide",
		},
	}, GetReleaseDocumentInput{
		PluginID: "linapro-demo-source",
		Version:  "v0.1.0",
		Locale:   "ja-JP",
		Path:     "index.md",
	})
	if err != nil {
		t.Fatalf("selectMarketplaceDocumentFallback returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected manifest docs document")
	}
	if record.Path != "index.md" || record.SourceKind != documentSourceKindManifestDocs {
		t.Fatalf("unexpected selected document: %#v", record)
	}
}

func TestBuildMarketplaceDocumentCatalogListsAllManifestPathsAndOmitsReadme(t *testing.T) {
	t.Parallel()
	catalog := buildMarketplaceDocumentCatalog([]*DocumentRecord{
		{
			Path:       "index.md",
			Title:      "智能中心",
			Locale:     "zh-CN",
			SourceKind: documentSourceKindManifestDocs,
		},
		{
			Path:       "index.md",
			Title:      "AI Hub",
			Locale:     "en-US",
			SourceKind: documentSourceKindManifestDocs,
		},
		{
			Path:       "configuration.md",
			Title:      "配置说明",
			Locale:     "zh-CN",
			SourceKind: documentSourceKindManifestDocs,
		},
		{
			Path:       "changelog.md",
			Title:      "更新日志",
			Locale:     "zh-CN",
			SourceKind: documentSourceKindManifestDocs,
		},
		{
			Path:       "README.md",
			Title:      "README",
			Locale:     "en-US",
			SourceKind: documentSourceKindReadme,
		},
		{
			Path:       readmeCNDocumentPath,
			Title:      "中文 README",
			Locale:     "zh-CN",
			SourceKind: documentSourceKindReadme,
		},
	}, "zh-CN")
	if len(catalog) != 3 {
		t.Fatalf("expected 3 catalog entries without README, got %#v", catalog)
	}
	if catalog[0].Path != "index.md" || catalog[0].Title != "智能中心" {
		t.Fatalf("expected preferred-locale index first, got %#v", catalog[0])
	}
	paths := []string{catalog[0].Path, catalog[1].Path, catalog[2].Path}
	if paths[1] != "changelog.md" || paths[2] != "configuration.md" {
		t.Fatalf("unexpected catalog path order: %#v", paths)
	}
	if len(catalog[0].Locales) != 2 {
		t.Fatalf("expected index locales zh-CN and en-US, got %#v", catalog[0].Locales)
	}
}

func TestBuildMarketplaceDocumentCatalogUsesReadmeWhenOnlyReadmeExists(t *testing.T) {
	t.Parallel()
	catalog := buildMarketplaceDocumentCatalog([]*DocumentRecord{
		{
			Path:       readmeDocumentPath,
			Title:      "README",
			Locale:     "en-US",
			SourceKind: documentSourceKindReadme,
		},
		{
			Path:       readmeCNDocumentPath,
			Title:      "中文 README",
			Locale:     "zh-CN",
			SourceKind: documentSourceKindReadme,
		},
	}, "zh-CN")
	if len(catalog) != 1 {
		t.Fatalf("expected one README fallback catalog item, got %#v", catalog)
	}
	if catalog[0].Path != defaultDocumentPath || catalog[0].Title != "中文 README" || catalog[0].SourceKind != documentSourceKindReadme {
		t.Fatalf("unexpected README fallback catalog item: %#v", catalog[0])
	}
	if len(catalog[0].Locales) != 2 {
		t.Fatalf("expected README locales zh-CN and en-US, got %#v", catalog[0].Locales)
	}
}

func TestFilterDocumentRecordsByPathMapsReadmeOnlyReleaseToIndex(t *testing.T) {
	t.Parallel()
	records := []*DocumentRecord{
		{
			Path:       readmeDocumentPath,
			Locale:     "en-US",
			SourceKind: documentSourceKindReadme,
		},
		{
			Path:       readmeCNDocumentPath,
			Locale:     "zh-CN",
			SourceKind: documentSourceKindReadme,
		},
	}
	filtered := filterDocumentRecordsByPath(records, defaultDocumentPath)
	if len(filtered) != 2 {
		t.Fatalf("expected README records to map to index.md, got %#v", filtered)
	}
}

func TestFilterDocumentRecordsByPathOmitsReadmeWhenManifestDocsExist(t *testing.T) {
	t.Parallel()
	records := []*DocumentRecord{
		{
			Path:       "configuration.md",
			Locale:     "en-US",
			SourceKind: documentSourceKindManifestDocs,
		},
		{
			Path:       readmeDocumentPath,
			Locale:     "en-US",
			SourceKind: documentSourceKindReadme,
		},
	}
	filtered := filterDocumentRecordsByPath(records, defaultDocumentPath)
	if len(filtered) != 0 {
		t.Fatalf("expected README records to stay hidden when manifest docs exist, got %#v", filtered)
	}
}

func writeDocumentPackageTarGz(t *testing.T, rootPrefix string, files map[string]string) string {
	t.Helper()

	tarPath := filepath.Join(t.TempDir(), "document-package.tar.gz")
	file, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("create tar.gz file: %v", err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		payload := []byte(content)
		header := &tar.Header{
			Name: rootPrefix + name,
			Mode: 0o600,
			Size: int64(len(payload)),
		}
		if err = tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header %s: %v", name, err)
		}
		if _, err = tarWriter.Write(payload); err != nil {
			t.Fatalf("write tar entry %s: %v", name, err)
		}
	}
	if err = tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err = gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("close tar.gz file: %v", err)
	}
	return tarPath
}
