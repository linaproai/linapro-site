// This file implements marketplace document indexing and lookup. It extracts
// version-scoped Markdown metadata from source or dynamic packages, validates
// image paths before indexing, renders conservative safe HTML, writes bounded
// document index rows, and applies locale fallback without loading package
// content from storage.

package marketplace

import (
	"archive/zip"
	"context"
	"html"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
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

// ResolveReleaseDocumentIndex returns one indexed document selected by fallback rules.
func (s *serviceImpl) ResolveReleaseDocumentIndex(
	ctx context.Context,
	in GetReleaseDocumentInput,
) (*DocumentRecord, error) {
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

	var rows []*entity.PluginMarketplaceDoc
	if err = dao.PluginMarketplaceDoc.Ctx(ctx).
		Where(do.PluginMarketplaceDoc{ReleaseId: release.Id}).
		OrderAsc(dao.PluginMarketplaceDoc.Columns().SourceKind).
		OrderAsc(dao.PluginMarketplaceDoc.Columns().Locale).
		OrderAsc(dao.PluginMarketplaceDoc.Columns().DocPath).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	records := make([]*DocumentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, documentRecordFromEntity(row))
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

// replaceReleaseDocuments replaces indexed document rows for one mutable release.
func (s *serviceImpl) replaceReleaseDocuments(
	ctx context.Context,
	release *ReleaseRecord,
	items []*marketplaceDocumentIndexItem,
) error {
	if release == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	var existingRows []*entity.PluginMarketplaceDoc
	if err := dao.PluginMarketplaceDoc.Ctx(ctx).
		Where(do.PluginMarketplaceDoc{ReleaseId: release.ID}).
		Scan(&existingRows); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	existingByKey := make(map[string]*entity.PluginMarketplaceDoc, len(existingRows))
	for _, row := range existingRows {
		if row == nil {
			continue
		}
		existingByKey[documentIndexKey(row.Locale, row.DocPath)] = row
	}

	desired := make(map[string]struct{}, len(items))
	return dao.PluginMarketplaceDoc.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, item := range items {
			if item == nil {
				continue
			}
			key := documentIndexKey(item.Locale, item.DocPath)
			desired[key] = struct{}{}
			data := marketplaceDocumentData(release, item)
			existing := existingByKey[key]
			if existing == nil {
				if _, err := dao.PluginMarketplaceDoc.Ctx(ctx).Data(data).Insert(); err != nil {
					return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
				}
				continue
			}
			if _, err := dao.PluginMarketplaceDoc.Ctx(ctx).
				Where(do.PluginMarketplaceDoc{Id: existing.Id}).
				Data(data).
				Update(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}

		for _, row := range existingRows {
			if row == nil {
				continue
			}
			if _, ok := desired[documentIndexKey(row.Locale, row.DocPath)]; ok {
				continue
			}
			if _, err := dao.PluginMarketplaceDoc.Ctx(ctx).
				Where(do.PluginMarketplaceDoc{Id: row.Id}).
				Delete(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}
		return nil
	})
}

// marketplaceDocumentData builds one DO payload for inserting or updating document indexes.
func marketplaceDocumentData(
	release *ReleaseRecord,
	item *marketplaceDocumentIndexItem,
) do.PluginMarketplaceDoc {
	return do.PluginMarketplaceDoc{
		ReleaseId:      release.ID,
		PluginId:       release.PluginID,
		ReleaseVersion: release.Version,
		Locale:         item.Locale,
		DocPath:        item.DocPath,
		SourceKind:     item.SourceKind,
		Title:          item.Title,
		Summary:        item.Summary,
		ContentHash:    item.ContentHash,
		SearchText:     item.SearchText,
	}
}

// buildSourcePackageDocumentIndex extracts Markdown docs from a source ZIP package.
func buildSourcePackageDocumentIndex(
	manifest *sourcePackageManifest,
	files map[string]*zip.File,
	rootPrefix string,
) ([]*marketplaceDocumentIndexItem, error) {
	return buildZipDocumentIndex(manifestDefaultLocale(manifest), files, rootPrefix)
}

// buildDynamicPackageDocumentIndex extracts Markdown docs from dynamic ZIP and embedded resources.
func buildDynamicPackageDocumentIndex(
	manifest *sourcePackageManifest,
	files map[string]*zip.File,
	rootPrefix string,
	resources []*dynamicPackageManifestResource,
) ([]*marketplaceDocumentIndexItem, error) {
	items, err := buildZipDocumentIndex(manifestDefaultLocale(manifest), files, rootPrefix)
	if err != nil {
		return nil, err
	}
	for _, resource := range resources {
		if resource == nil || !strings.HasPrefix(resource.Path, marketplaceDocsPrefix) ||
			strings.ToLower(filepath.Ext(resource.Path)) != ".md" {
			continue
		}
		locale, docPath, parseErr := parseManifestDocsPath(resource.Path, manifestDefaultLocale(manifest))
		if parseErr != nil {
			return nil, parseErr
		}
		item, indexErr := indexMarketplaceDocument(
			locale,
			docPath,
			documentSourceKindManifestDocs,
			string(resource.Content),
		)
		if indexErr != nil {
			return nil, indexErr
		}
		items = append(items, item)
	}
	sortMarketplaceDocumentIndex(items)
	return dedupeMarketplaceDocumentIndex(items), nil
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
func selectMarketplaceDocumentFallback(
	records []*DocumentRecord,
	in GetReleaseDocumentInput,
) (*DocumentRecord, error) {
	requestedPath, err := normalizeMarketplaceDocumentPath(in.Path)
	if err != nil {
		return nil, err
	}
	requestedLocale := normalizeDocumentLocale(in.Locale)
	candidateLocales := uniqueDocumentLocales(
		requestedLocale,
		normalizeDocumentLocale(in.DefaultLocale),
		fallbackZhCNLocale,
		fallbackEnUSLocale,
		defaultDocumentLocale,
	)

	byKey := make(map[string]*DocumentRecord, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		byKey[documentLookupKey(record.SourceKind, record.Locale, record.Path)] = record
	}
	for _, locale := range candidateLocales {
		if record := byKey[documentLookupKey(documentSourceKindManifestDocs, locale, requestedPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}
	if isIndexDocumentPath(requestedPath) {
		if record := byKey[documentLookupKey(documentSourceKindReadme, fallbackZhCNLocale, readmeCNDocumentPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
		if record := byKey[documentLookupKey(documentSourceKindReadme, fallbackEnUSLocale, readmeDocumentPath)]; record != nil {
			return markDocumentFallback(record, requestedLocale, requestedPath), nil
		}
	}
	return nil, nil
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
