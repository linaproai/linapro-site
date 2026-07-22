// This file implements marketplace document lookup from version artifact disk
// storage. Markdown is rendered on read from package ZIP or Git docs snapshots;
// document bodies are never persisted in database tables.

package marketplace

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"html"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	documentSourceKindManifestDocs = "manifest_docs"
	documentSourceKindReadme       = "readme"
	defaultDocumentPath            = "index.md"
	defaultDocumentLocale          = "default"
	fallbackZhCNLocale             = "zh-CN"
	fallbackEnUSLocale             = "en-US"
	marketplaceDocsPrefix          = "manifest/docs/"
	readmeDocumentPath             = "README.md"
	readmeCNDocumentPath           = "README.zh-CN.md"
	maxDocumentSummaryRunes        = 240
	maxDocumentSearchTextRunes     = 4096
)

var (
	marketplaceMarkdownImagePattern = regexp.MustCompile(`!\[[^\]]*]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	marketplaceHTMLImagePattern     = regexp.MustCompile(`(?i)<img[^>]+src\s*=\s*["']([^"']+)["']`)
	marketplaceHTMLTagPattern       = regexp.MustCompile(`(?s)<[^>]+>`)
	marketplaceMarkdownLinkPattern  = regexp.MustCompile(`!?\[([^\]]*)]\([^)]+\)`)
)

// marketplaceDocumentIndexItem carries one package document ready for indexing.
type marketplaceDocumentIndexItem struct {
	Locale          string
	DocPath         string
	SourceKind      string
	Title           string
	Summary         string
	ContentHash     string
	SearchText      string
	RenderedContent string
}

// ResolveReleaseDocumentIndex returns one document selected by fallback rules.
func (s *serviceImpl) ResolveReleaseDocumentIndex(
	ctx context.Context,
	in GetReleaseDocumentInput,
) (*DocumentRecord, error) {
	records, err := s.loadVisibleReleaseDocumentRecords(ctx, in)
	if err != nil {
		return nil, err
	}
	selected, err := selectMarketplaceDocumentFallback(records, in)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return nil, bizerr.NewCode(CodeMarketplaceDocumentNotFound)
	}
	return selected, nil
}

// loadVisibleReleaseDocumentRecords loads candidate documents for one visible
// release from package ZIP artifacts or Git docs disk snapshots.
func (s *serviceImpl) loadVisibleReleaseDocumentRecords(
	ctx context.Context,
	in GetReleaseDocumentInput,
) ([]*DocumentRecord, error) {
	release, err := s.requireVisibleRelease(
		ctx,
		in.PluginID,
		in.Version,
		in.Visibility,
		marketplaceVisibilityPermissionView,
	)
	if err != nil {
		return nil, err
	}
	requestedPath, err := normalizeMarketplaceDocumentPath(in.Path)
	if err != nil {
		return nil, err
	}
	candidatePaths := documentCandidatePaths(requestedPath)

	items, err := s.loadReleaseDocumentIndexItems(ctx, release)
	if err != nil {
		return nil, err
	}
	records := make([]*DocumentRecord, 0, len(items))
	allowedPaths := make(map[string]struct{}, len(candidatePaths))
	for _, candidatePath := range candidatePaths {
		allowedPaths[candidatePath] = struct{}{}
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := allowedPaths[item.DocPath]; !ok {
			continue
		}
		records = append(records, documentRecordFromIndexItem(release, item))
	}
	return records, nil
}

// documentCandidatePaths returns doc paths considered for one request path.
func documentCandidatePaths(requestedPath string) []string {
	candidatePaths := []string{requestedPath}
	if isIndexDocumentPath(requestedPath) {
		candidatePaths = append(candidatePaths, readmeCNDocumentPath, readmeDocumentPath)
	}
	seenPath := make(map[string]struct{}, len(candidatePaths))
	filteredPaths := make([]string, 0, len(candidatePaths))
	for _, candidatePath := range candidatePaths {
		if _, ok := seenPath[candidatePath]; ok {
			continue
		}
		seenPath[candidatePath] = struct{}{}
		filteredPaths = append(filteredPaths, candidatePath)
	}
	return filteredPaths
}

// loadReleaseDocumentIndexItems reads all package/Git docs for one release.
func (s *serviceImpl) loadReleaseDocumentIndexItems(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
) ([]*marketplaceDocumentIndexItem, error) {
	if release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	items, err := s.loadDocumentIndexItemsFromPackageArtifact(ctx, release)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}
	return s.loadDocumentIndexItemsFromGitSnapshot(ctx, release)
}

// loadDocumentIndexItemsFromPackageArtifact opens the primary package ZIP and
// extracts Markdown docs for rendering.
func (s *serviceImpl) loadDocumentIndexItemsFromPackageArtifact(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
) ([]*marketplaceDocumentIndexItem, error) {
	if s == nil || s.artifacts == nil || release == nil {
		return nil, nil
	}
	artifact, err := s.selectPackageArtifactForDocuments(ctx, release)
	if err != nil {
		return nil, err
	}
	if artifact == nil || strings.TrimSpace(artifact.StorageKey) == "" {
		return nil, nil
	}
	localPath, err := s.artifacts.LocalPath(ctx, artifact.StorageKey)
	if err != nil {
		return nil, err
	}
	reader, err := zip.OpenReader(localPath)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	defer func() {
		_ = reader.Close()
	}()
	fileIndex, err := indexSourceZipFiles(reader.File)
	if err != nil {
		return nil, err
	}
	rootPrefix, err := detectSourcePackageRoot(fileIndex)
	if err != nil {
		// Dynamic packages may place plugin.yaml at root without source layout.
		rootPrefix = ""
		if _, ok := fileIndex[sourcePackageManifestPath]; !ok {
			return nil, nil
		}
	}
	defaultLocale := fallbackEnUSLocale
	if manifestFile := fileIndex[rootPrefix+sourcePackageManifestPath]; manifestFile != nil {
		if content, readErr := readZipFile(manifestFile); readErr == nil {
			manifest := &sourcePackageManifest{}
			if yaml.Unmarshal(content, manifest) == nil {
				normalizeSourcePackageManifest(manifest)
				defaultLocale = manifestDefaultLocale(manifest)
			}
		}
	}
	return buildZipDocumentIndex(defaultLocale, fileIndex, rootPrefix)
}

// selectPackageArtifactForDocuments returns the package ZIP preferred for docs.
func (s *serviceImpl) selectPackageArtifactForDocuments(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
) (*entity.PluginMarketplaceArtifact, error) {
	if release == nil {
		return nil, nil
	}
	var rows []*entity.PluginMarketplaceArtifact
	cols := dao.PluginMarketplaceArtifact.Columns()
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Fields(
			cols.Id,
			cols.ReleaseId,
			cols.ArtifactType,
			cols.StorageKey,
			cols.FileName,
		).
		Where(do.PluginMarketplaceArtifact{ReleaseId: release.Id}).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	pluginType := marketv1.MarketplacePluginType(release.PluginType)
	var selected *entity.PluginMarketplaceArtifact
	for _, row := range rows {
		if row == nil {
			continue
		}
		if !isPackageDocumentArtifactType(row.ArtifactType) {
			continue
		}
		if selected == nil || artifactPriority(row, pluginType) < artifactPriority(selected, pluginType) {
			selected = row
		}
	}
	return selected, nil
}

// isPackageDocumentArtifactType reports package archives that may contain docs.
func isPackageDocumentArtifactType(artifactType string) bool {
	switch marketv1.MarketplaceArtifactType(artifactType) {
	case marketv1.MarketplaceArtifactTypeSourceZip,
		marketv1.MarketplaceArtifactTypeSourceTarGz,
		marketv1.MarketplaceArtifactTypeDynamicZip,
		marketv1.MarketplaceArtifactTypeDynamicTarGz:
		return true
	default:
		return false
	}
}

// docsSnapshotManifest is the on-disk index for Git-sourced documentation.
type docsSnapshotManifest struct {
	Items []*docsSnapshotManifestItem `json:"items"`
}

// docsSnapshotManifestItem points at one Markdown body stored under ArtifactStore.
type docsSnapshotManifestItem struct {
	Locale     string `json:"locale"`
	DocPath    string `json:"docPath"`
	SourceKind string `json:"sourceKind"`
	ContentKey string `json:"contentKey"`
}

// replaceReleaseGitDocumentSnapshot writes rendered Git docs to artifact disk.
func (s *serviceImpl) replaceReleaseGitDocumentSnapshot(
	ctx context.Context,
	release *ReleaseRecord,
	items []*marketplaceDocumentIndexItem,
	rawBodies map[string][]byte,
) error {
	if s == nil || s.artifacts == nil || release == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	root := docsSnapshotRoot(release.PluginID, release.Version)
	manifest := &docsSnapshotManifest{Items: make([]*docsSnapshotManifestItem, 0, len(items))}
	for _, item := range items {
		if item == nil {
			continue
		}
		key := documentIndexKey(item.Locale, item.DocPath)
		body := rawBodies[key]
		if len(body) == 0 {
			continue
		}
		contentKey := path.Join(root, "content", item.ContentHash+".md")
		if err := s.artifacts.Put(ctx, contentKey, bytes.NewReader(body)); err != nil {
			return err
		}
		manifest.Items = append(manifest.Items, &docsSnapshotManifestItem{
			Locale:     item.Locale,
			DocPath:    item.DocPath,
			SourceKind: item.SourceKind,
			ContentKey: contentKey,
		})
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.artifacts.Put(ctx, docsSnapshotManifestKey(release.PluginID, release.Version), bytes.NewReader(payload))
}

// loadDocumentIndexItemsFromGitSnapshot loads Git docs previously written to disk.
func (s *serviceImpl) loadDocumentIndexItemsFromGitSnapshot(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
) ([]*marketplaceDocumentIndexItem, error) {
	if s == nil || s.artifacts == nil || release == nil {
		return nil, nil
	}
	reader, err := s.artifacts.Open(ctx, docsSnapshotManifestKey(release.PluginId, release.ReleaseVersion))
	if err != nil {
		// Missing snapshot is normal for releases without Git docs enrichment.
		return nil, nil
	}
	defer func() {
		_ = reader.Close()
	}()
	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var manifest docsSnapshotManifest
	if err = json.Unmarshal(payload, &manifest); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	items := make([]*marketplaceDocumentIndexItem, 0, len(manifest.Items))
	for _, entry := range manifest.Items {
		if entry == nil || strings.TrimSpace(entry.ContentKey) == "" {
			continue
		}
		bodyReader, openErr := s.artifacts.Open(ctx, entry.ContentKey)
		if openErr != nil {
			continue
		}
		body, readErr := io.ReadAll(bodyReader)
		_ = bodyReader.Close()
		if readErr != nil || len(body) == 0 {
			continue
		}
		item, indexErr := indexMarketplaceDocument(entry.Locale, entry.DocPath, entry.SourceKind, string(body))
		if indexErr != nil {
			continue
		}
		items = append(items, item)
	}
	sortMarketplaceDocumentIndex(items)
	return dedupeMarketplaceDocumentIndex(items), nil
}

func docsSnapshotRoot(pluginID string, version string) string {
	return path.Join("docs-snapshot", normalizeKey(pluginID), normalizeKey(version))
}

func docsSnapshotManifestKey(pluginID string, version string) string {
	return path.Join(docsSnapshotRoot(pluginID, version), "manifest.json")
}

// documentRecordFromIndexItem projects an in-memory index item for one release.
func documentRecordFromIndexItem(
	release *entity.PluginMarketplaceRelease,
	item *marketplaceDocumentIndexItem,
) *DocumentRecord {
	if release == nil || item == nil {
		return nil
	}
	return &DocumentRecord{
		ReleaseID:       release.Id,
		PluginID:        release.PluginId,
		Version:         release.ReleaseVersion,
		Locale:          item.Locale,
		ResolvedLocale:  item.Locale,
		Path:            item.DocPath,
		SourceKind:      item.SourceKind,
		Title:           item.Title,
		Summary:         item.Summary,
		ContentHash:     item.ContentHash,
		SearchText:      item.SearchText,
		RenderedContent: item.RenderedContent,
	}
}

// buildZipDocumentIndex extracts manifest/docs and README Markdown from ZIP files.
func buildZipDocumentIndex(
	defaultLocale string,
	files map[string]*zip.File,
	rootPrefix string,
) ([]*marketplaceDocumentIndexItem, error) {
	paths := make([]string, 0, len(files))
	for filePath := range files {
		paths = append(paths, filePath)
	}
	sort.Strings(paths)

	items := make([]*marketplaceDocumentIndexItem, 0)
	for _, filePath := range paths {
		file := files[filePath]
		relative := strings.TrimPrefix(filePath, rootPrefix)
		var (
			sourceKind string
			locale     string
			docPath    string
		)
		switch {
		case strings.HasPrefix(relative, marketplaceDocsPrefix) && strings.ToLower(filepath.Ext(relative)) == ".md":
			parsedLocale, parsedPath, err := parseManifestDocsPath(relative, defaultLocale)
			if err != nil {
				return nil, err
			}
			locale = parsedLocale
			docPath = parsedPath
			sourceKind = documentSourceKindManifestDocs
		case relative == readmeDocumentPath:
			locale = fallbackEnUSLocale
			docPath = readmeDocumentPath
			sourceKind = documentSourceKindReadme
		case relative == readmeCNDocumentPath:
			locale = fallbackZhCNLocale
			docPath = readmeCNDocumentPath
			sourceKind = documentSourceKindReadme
		default:
			continue
		}

		content, err := readZipFile(file)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
		}
		item, err := indexMarketplaceDocument(locale, docPath, sourceKind, string(content))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sortMarketplaceDocumentIndex(items)
	return dedupeMarketplaceDocumentIndex(items), nil
}

// indexMarketplaceDocument validates, renders, and summarizes one Markdown document.
func indexMarketplaceDocument(
	locale string,
	docPath string,
	sourceKind string,
	content string,
) (*marketplaceDocumentIndexItem, error) {
	normalizedLocale := normalizeDocumentLocale(locale)
	normalizedPath, err := normalizeMarketplaceDocumentPath(docPath)
	if err != nil {
		return nil, err
	}
	rendered, plain, err := renderMarketplaceMarkdown(normalizedLocale, normalizedPath, sourceKind, content)
	if err != nil {
		return nil, err
	}
	searchText := truncateRunes(normalizeDocumentPlainText(plain), maxDocumentSearchTextRunes)
	title := extractMarketplaceDocumentTitle(content, normalizedPath)
	return &marketplaceDocumentIndexItem{
		Locale:          normalizedLocale,
		DocPath:         normalizedPath,
		SourceKind:      sourceKind,
		Title:           truncateRunes(title, 255),
		Summary:         truncateRunes(searchText, maxDocumentSummaryRunes),
		ContentHash:     sha256Hex([]byte(content)),
		SearchText:      searchText,
		RenderedContent: rendered,
	}, nil
}

// renderMarketplaceMarkdown produces conservative safe HTML and plain text.
func renderMarketplaceMarkdown(
	locale string,
	docPath string,
	sourceKind string,
	content string,
) (rendered string, plain string, err error) {
	if err = validateMarketplaceDocumentAssetRefs(locale, docPath, sourceKind, content); err != nil {
		return "", "", err
	}

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var (
		htmlBuilder  strings.Builder
		plainBuilder strings.Builder
		paragraph    []string
	)
	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		text := strings.Join(paragraph, " ")
		htmlBuilder.WriteString("<p>")
		htmlBuilder.WriteString(html.EscapeString(text))
		htmlBuilder.WriteString("</p>")
		htmlBuilder.WriteByte('\n')
		plainBuilder.WriteString(stripMarkdownText(text))
		plainBuilder.WriteByte('\n')
		paragraph = paragraph[:0]
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flushParagraph()
			continue
		}
		level, heading := parseMarkdownHeading(trimmed)
		if level > 0 {
			flushParagraph()
			htmlBuilder.WriteString("<h")
			htmlBuilder.WriteByte(byte('0' + level))
			htmlBuilder.WriteByte('>')
			htmlBuilder.WriteString(html.EscapeString(heading))
			htmlBuilder.WriteString("</h")
			htmlBuilder.WriteByte(byte('0' + level))
			htmlBuilder.WriteString(">\n")
			plainBuilder.WriteString(stripMarkdownText(heading))
			plainBuilder.WriteByte('\n')
			continue
		}
		paragraph = append(paragraph, trimmed)
	}
	flushParagraph()
	return htmlBuilder.String(), plainBuilder.String(), nil
}

// validateMarketplaceDocumentAssetRefs rejects unsafe image references.
func validateMarketplaceDocumentAssetRefs(
	locale string,
	docPath string,
	sourceKind string,
	content string,
) error {
	for _, match := range marketplaceMarkdownImagePattern.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		if err := validateMarketplaceDocumentAssetPath(locale, docPath, sourceKind, match[1]); err != nil {
			return err
		}
	}
	for _, match := range marketplaceHTMLImagePattern.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		if err := validateMarketplaceDocumentAssetPath(locale, docPath, sourceKind, match[1]); err != nil {
			return err
		}
	}
	return nil
}

// validateMarketplaceDocumentAssetPath keeps images inside the current documentation boundary.
func validateMarketplaceDocumentAssetPath(
	locale string,
	docPath string,
	sourceKind string,
	target string,
) error {
	cleanTarget := strings.TrimSpace(strings.Split(strings.Split(target, "#")[0], "?")[0])
	cleanTarget = strings.ReplaceAll(cleanTarget, "\\", "/")
	if cleanTarget == "" {
		return documentDiagnosticError("document image path cannot be empty")
	}
	if strings.Contains(cleanTarget, "://") || strings.HasPrefix(cleanTarget, "//") ||
		strings.HasPrefix(cleanTarget, "/") || strings.HasPrefix(strings.ToLower(cleanTarget), "data:") {
		return documentDiagnosticError("document image path must be a package-relative path")
	}
	if len(cleanTarget) >= 2 && cleanTarget[1] == ':' {
		return documentDiagnosticError("document image path must not contain drive prefix")
	}
	if sourceKind == documentSourceKindReadme {
		normalized := path.Clean(cleanTarget)
		if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
			return documentDiagnosticError("document image path must not traverse parent directories")
		}
		return nil
	}

	localeRoot := path.Join("manifest/docs", normalizeDocumentLocale(locale))
	docFullPath := path.Join(localeRoot, docPath)
	docDir := path.Dir(docFullPath)
	targetPath := path.Clean(path.Join(docDir, cleanTarget))
	if targetPath == "." || targetPath == ".." || strings.HasPrefix(targetPath, "../") ||
		!strings.HasPrefix(targetPath, "manifest/docs/") {
		return documentDiagnosticError("document image path escapes manifest/docs")
	}
	globalAssets := "manifest/docs/assets/"
	localeAssets := path.Join(localeRoot, "assets") + "/"
	docDirPrefix := docDir + "/"
	if strings.HasPrefix(targetPath, globalAssets) ||
		strings.HasPrefix(targetPath, localeAssets) ||
		strings.HasPrefix(targetPath, docDirPrefix) {
		return nil
	}
	return documentDiagnosticError("document image path must stay under manifest/docs/assets or the document directory")
}

// selectMarketplaceDocumentFallback applies locale and README fallback rules to index records.
// Selection order for version-scoped marketplace docs:
//  1. When only one manifest_docs language exists for the path, return that language.
//  2. When multiple languages exist, prefer an exact match to the user locale.
//  3. Otherwise prefer English (en-US / en).
//  4. For the index path only, apply the same requested-locale / English
//     fallback order to README.zh-CN.md / README.md.
//  5. As a last resort, return any remaining manifest_docs candidate for the path.
func selectMarketplaceDocumentFallback(
	records []*DocumentRecord,
	in GetReleaseDocumentInput,
) (*DocumentRecord, error) {
	requestedPath, err := normalizeMarketplaceDocumentPath(in.Path)
	if err != nil {
		return nil, err
	}
	requestedLocale := normalizeDocumentLocale(in.Locale)

	byKey := make(map[string]*DocumentRecord, len(records))
	pathLocales := make([]string, 0, 4)
	seenPathLocales := make(map[string]struct{}, 4)
	for _, record := range records {
		if record == nil {
			continue
		}
		byKey[documentLookupKey(record.SourceKind, record.Locale, record.Path)] = record
		if record.SourceKind != documentSourceKindManifestDocs || record.Path != requestedPath {
			continue
		}
		locale := normalizeDocumentLocale(record.Locale)
		if _, ok := seenPathLocales[locale]; ok {
			continue
		}
		seenPathLocales[locale] = struct{}{}
		pathLocales = append(pathLocales, locale)
	}

	// Single available language: always show it regardless of the request locale.
	if len(pathLocales) == 1 {
		if record := byKey[documentLookupKey(documentSourceKindManifestDocs, pathLocales[0], requestedPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}

	// Multiple languages: prefer the user locale, then English.
	candidateLocales := uniqueDocumentLocales(
		requestedLocale,
		fallbackEnUSLocale,
		"en",
	)
	// Drop the synthetic "default" marker so empty request locale does not pin
	// an artificial locale before English.
	filteredCandidates := make([]string, 0, len(candidateLocales))
	for _, locale := range candidateLocales {
		if locale == defaultDocumentLocale {
			continue
		}
		filteredCandidates = append(filteredCandidates, locale)
	}
	for _, locale := range filteredCandidates {
		if record := byKey[documentLookupKey(documentSourceKindManifestDocs, locale, requestedPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}
	if isIndexDocumentPath(requestedPath) {
		if record := selectReadmeDocumentFallback(byKey, filteredCandidates); record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}
	// Last resort: any remaining manifest_docs language for the path.
	for _, locale := range pathLocales {
		if record := byKey[documentLookupKey(documentSourceKindManifestDocs, locale, requestedPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}
	return nil, nil
}

// selectReadmeDocumentFallback applies locale fallback to README entry docs.
func selectReadmeDocumentFallback(
	byKey map[string]*DocumentRecord,
	candidateLocales []string,
) *DocumentRecord {
	if len(byKey) == 0 {
		return nil
	}
	readmeRecords := []*DocumentRecord{
		byKey[documentLookupKey(documentSourceKindReadme, fallbackZhCNLocale, readmeCNDocumentPath)],
		byKey[documentLookupKey(documentSourceKindReadme, fallbackEnUSLocale, readmeDocumentPath)],
	}
	available := make([]*DocumentRecord, 0, len(readmeRecords))
	for _, record := range readmeRecords {
		if record != nil {
			available = append(available, record)
		}
	}
	if len(available) == 1 {
		return available[0]
	}
	for _, locale := range candidateLocales {
		for _, record := range available {
			if normalizeDocumentLocale(record.Locale) == locale {
				return record
			}
		}
	}
	return nil
}

// collectMarketplaceDocumentBundle returns same-path language alternatives for the selected document.
func collectMarketplaceDocumentBundle(
	records []*DocumentRecord,
	selected *DocumentRecord,
	in GetReleaseDocumentInput,
) ([]*DocumentRecord, error) {
	if selected == nil {
		return nil, nil
	}
	requestedPath, err := normalizeMarketplaceDocumentPath(in.Path)
	if err != nil {
		return nil, err
	}
	requestedLocale := normalizeDocumentLocale(in.Locale)
	bundle := make([]*DocumentRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		switch selected.SourceKind {
		case documentSourceKindManifestDocs:
			if record.SourceKind != documentSourceKindManifestDocs || record.Path != requestedPath {
				continue
			}
		case documentSourceKindReadme:
			if record.SourceKind != documentSourceKindReadme || !isIndexDocumentPath(requestedPath) {
				continue
			}
		default:
			continue
		}
		cloned := *record
		cloned.RequestedLocale = requestedLocale
		cloned.ResolvedLocale = record.Locale
		cloned.FallbackUsed = false
		bundle = append(bundle, &cloned)
	}
	return bundle, nil
}

// markDocumentFallback clones one record and annotates fallback metadata.
func markDocumentFallback(record *DocumentRecord, requestedLocale string, requestedPath string) *DocumentRecord {
	cloned := *record
	cloned.RequestedLocale = requestedLocale
	cloned.ResolvedLocale = record.Locale
	cloned.FallbackUsed = record.SourceKind != documentSourceKindManifestDocs ||
		record.Locale != requestedLocale ||
		record.Path != requestedPath
	return &cloned
}

// parseManifestDocsPath normalizes manifest/docs/<locale>/<path> entries.
func parseManifestDocsPath(relativePath string, defaultLocale string) (locale string, docPath string, err error) {
	rest := strings.TrimPrefix(relativePath, marketplaceDocsPrefix)
	segments := strings.Split(rest, "/")
	if len(segments) >= 2 && looksLikeLocale(segments[0]) {
		locale = normalizeDocumentLocale(segments[0])
		docPath = strings.Join(segments[1:], "/")
	} else {
		locale = normalizeDocumentLocale(defaultLocale)
		docPath = rest
	}
	docPath, err = normalizeMarketplaceDocumentPath(docPath)
	return locale, docPath, err
}

// normalizeMarketplaceDocumentPath validates a document path inside docs or README fallback.
func normalizeMarketplaceDocumentPath(value string) (string, error) {
	trimmed := strings.ReplaceAll(normalizeKey(value), "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, marketplaceDocsPrefix)
	if trimmed == "" {
		trimmed = defaultDocumentPath
	}
	if strings.Contains(trimmed, "://") || strings.HasPrefix(trimmed, "/") {
		return "", documentDiagnosticError("document path must be relative")
	}
	if len(trimmed) >= 2 && trimmed[1] == ':' {
		return "", documentDiagnosticError("document path must not contain drive prefix")
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", documentDiagnosticError("document path must not traverse parent directories")
	}
	return normalized, nil
}

// manifestDefaultLocale returns the declared documentation fallback locale.
func manifestDefaultLocale(manifest *sourcePackageManifest) string {
	if manifest != nil && manifest.I18N != nil && normalizeKey(manifest.I18N.Default) != "" {
		return normalizeDocumentLocale(manifest.I18N.Default)
	}
	return defaultDocumentLocale
}

// normalizeDocumentLocale trims locale values and applies the default marker.
func normalizeDocumentLocale(value string) string {
	trimmed := normalizeKey(value)
	if trimmed == "" {
		return defaultDocumentLocale
	}
	return trimmed
}

// uniqueDocumentLocales returns non-empty locales in first-seen order.
func uniqueDocumentLocales(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	locales := make([]string, 0, len(values))
	for _, value := range values {
		locale := normalizeDocumentLocale(value)
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		locales = append(locales, locale)
	}
	return locales
}

// looksLikeLocale reports whether one path segment is a locale marker.
func looksLikeLocale(value string) bool {
	trimmed := normalizeKey(value)
	return strings.Contains(trimmed, "-") || len(trimmed) == 2
}

// isIndexDocumentPath reports whether README fallback should apply.
func isIndexDocumentPath(value string) bool {
	return normalizeKey(value) == "" || value == defaultDocumentPath
}

// documentIndexKey builds a release-local index key.
func documentIndexKey(locale string, docPath string) string {
	return normalizeDocumentLocale(locale) + "\x00" + docPath
}

// documentLookupKey builds a lookup key including the source kind.
func documentLookupKey(sourceKind string, locale string, docPath string) string {
	return sourceKind + "\x00" + documentIndexKey(locale, docPath)
}

// sortMarketplaceDocumentIndex keeps document indexing deterministic.
func sortMarketplaceDocumentIndex(items []*marketplaceDocumentIndexItem) {
	sort.Slice(items, func(left int, right int) bool {
		if items[left].Locale == items[right].Locale {
			if items[left].SourceKind == items[right].SourceKind {
				return items[left].DocPath < items[right].DocPath
			}
			return items[left].SourceKind < items[right].SourceKind
		}
		return items[left].Locale < items[right].Locale
	})
}

// dedupeMarketplaceDocumentIndex keeps the first document for each locale/path pair.
func dedupeMarketplaceDocumentIndex(items []*marketplaceDocumentIndexItem) []*marketplaceDocumentIndexItem {
	seen := make(map[string]struct{}, len(items))
	result := make([]*marketplaceDocumentIndexItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		key := documentIndexKey(item.Locale, item.DocPath)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

// parseMarkdownHeading extracts Markdown ATX heading level and text.
func parseMarkdownHeading(line string) (level int, text string) {
	count := 0
	for count < len(line) && line[count] == '#' && count < 6 {
		count++
	}
	if count == 0 || count >= len(line) || line[count] != ' ' {
		return 0, ""
	}
	return count, normalizeKey(line[count+1:])
}

// extractMarketplaceDocumentTitle returns the first Markdown heading or file name.
func extractMarketplaceDocumentTitle(content string, docPath string) string {
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		_, heading := parseMarkdownHeading(strings.TrimSpace(line))
		if heading != "" {
			return heading
		}
	}
	base := path.Base(docPath)
	ext := path.Ext(base)
	if ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return strings.ReplaceAll(base, "-", " ")
}

// stripMarkdownText converts a small Markdown subset to plain text for indexing.
func stripMarkdownText(value string) string {
	withoutTags := marketplaceHTMLTagPattern.ReplaceAllString(value, " ")
	withoutLinks := marketplaceMarkdownLinkPattern.ReplaceAllString(withoutTags, "$1")
	replacer := strings.NewReplacer("#", " ", "*", " ", "`", " ", "_", " ", ">", " ", "|", " ")
	return replacer.Replace(withoutLinks)
}

// normalizeDocumentPlainText collapses whitespace for summary and search indexes.
func normalizeDocumentPlainText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// truncateRunes truncates text by rune count.
func truncateRunes(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}

// documentDiagnosticError creates a document validation business error.
func documentDiagnosticError(diagnostic string) error {
	return bizerr.NewCode(CodeMarketplaceDocumentInvalid, bizerr.P("diagnostic", diagnostic))
}
