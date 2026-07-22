// This file persists and projects marketplace display name/summary metadata
// per release and locale. Document bodies and images stay on artifact disk.

package marketplace

import (
	"archive/zip"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	displaySourcePackageI18N = "package_i18n"
	displaySourcePluginYAML  = "plugin_yaml"
	displaySourcePublisher   = "publisher"
	defaultDisplayLocale     = "en-US"
	maxDisplayNameRunes      = 128
	maxDisplaySummaryRunes   = 512
)

// marketplaceDisplayI18nItem is one locale row ready for persistence.
type marketplaceDisplayI18nItem struct {
	Locale  string
	Name    string
	Summary string
	Source  string
}

// replaceReleaseDisplayI18n replaces all display metadata rows for one release.
func (s *serviceImpl) replaceReleaseDisplayI18n(
	ctx context.Context,
	release *ReleaseRecord,
	items []*marketplaceDisplayI18nItem,
) error {
	if release == nil || release.ID <= 0 {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	normalized := normalizeDisplayI18nItems(items)
	if len(normalized) == 0 {
		return nil
	}

	return dao.PluginMarketplaceDisplayI18n.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		// Final-state replace: delete existing rows for the release, then insert.
		if _, err := dao.PluginMarketplaceDisplayI18n.Ctx(ctx).
			Where(do.PluginMarketplaceDisplayI18n{ReleaseId: release.ID}).
			Delete(); err != nil {
			return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
		}
		for _, item := range normalized {
			if item == nil {
				continue
			}
			if _, err := dao.PluginMarketplaceDisplayI18n.Ctx(ctx).Data(do.PluginMarketplaceDisplayI18n{
				ReleaseId:      release.ID,
				PluginId:       release.PluginID,
				ReleaseVersion: release.Version,
				Locale:         item.Locale,
				Name:           item.Name,
				Summary:        item.Summary,
				Source:         item.Source,
			}).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}
		return nil
	})
}

// batchDisplayI18nByReleaseIDs loads active display rows for the given releases.
func (s *serviceImpl) batchDisplayI18nByReleaseIDs(
	ctx context.Context,
	releaseIDs []int,
) (map[int][]*entity.PluginMarketplaceDisplayI18n, error) {
	out := make(map[int][]*entity.PluginMarketplaceDisplayI18n)
	ids := uniquePositiveInts(releaseIDs)
	if len(ids) == 0 {
		return out, nil
	}
	cols := dao.PluginMarketplaceDisplayI18n.Columns()
	var rows []*entity.PluginMarketplaceDisplayI18n
	if err := dao.PluginMarketplaceDisplayI18n.Ctx(ctx).
		WhereIn(cols.ReleaseId, ids).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	for _, row := range rows {
		if row == nil || row.ReleaseId <= 0 {
			continue
		}
		out[row.ReleaseId] = append(out[row.ReleaseId], row)
	}
	return out, nil
}

// pickDisplayNameSummary selects localized name/summary with fallback chain:
// request locale → defaultLocale → en-US → identity fallback.
func pickDisplayNameSummary(
	rows []*entity.PluginMarketplaceDisplayI18n,
	requestLocale string,
	defaultLocale string,
	fallbackName string,
	fallbackSummary string,
) (name string, summary string) {
	byLocale := make(map[string]*entity.PluginMarketplaceDisplayI18n, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		locale := normalizeDisplayLocale(row.Locale)
		if locale == "" {
			continue
		}
		byLocale[locale] = row
	}
	for _, candidate := range []string{
		normalizeDisplayLocale(requestLocale),
		normalizeDisplayLocale(defaultLocale),
		defaultDisplayLocale,
	} {
		if candidate == "" {
			continue
		}
		if row := byLocale[candidate]; row != nil {
			name = firstNonEmpty(row.Name, fallbackName)
			summary = firstNonEmpty(row.Summary, fallbackSummary)
			return name, summary
		}
	}
	return fallbackName, fallbackSummary
}

// buildDisplayI18nFromPackageYAML builds at least the default-locale row from plugin.yaml.
func buildDisplayI18nFromPackageYAML(
	pluginID string,
	name string,
	description string,
	defaultLocale string,
) []*marketplaceDisplayI18nItem {
	locale := normalizeDisplayLocale(defaultLocale)
	if locale == "" {
		locale = defaultDisplayLocale
	}
	displayName := firstNonEmpty(strings.TrimSpace(name), strings.TrimSpace(pluginID))
	summary := strings.TrimSpace(description)
	if summary == "" {
		summary = displayName
	}
	return []*marketplaceDisplayI18nItem{{
		Locale:  locale,
		Name:    truncateRunes(displayName, maxDisplayNameRunes),
		Summary: truncateRunes(summary, maxDisplaySummaryRunes),
		Source:  displaySourcePluginYAML,
	}}
}

// mergePackageI18nDisplayItems merges runtime i18n catalog keys into display rows.
// Keys: plugin.<pluginId>.name and plugin.<pluginId>.description (→ summary).
func mergePackageI18nDisplayItems(
	pluginID string,
	base []*marketplaceDisplayI18nItem,
	localeCatalogs map[string]map[string]string,
) []*marketplaceDisplayI18nItem {
	byLocale := make(map[string]*marketplaceDisplayI18nItem)
	for _, item := range base {
		if item == nil {
			continue
		}
		locale := normalizeDisplayLocale(item.Locale)
		if locale == "" {
			continue
		}
		copyItem := *item
		copyItem.Locale = locale
		byLocale[locale] = &copyItem
	}

	nameKey := "plugin." + strings.TrimSpace(pluginID) + ".name"
	descriptionKey := "plugin." + strings.TrimSpace(pluginID) + ".description"
	for locale, catalog := range localeCatalogs {
		normalizedLocale := normalizeDisplayLocale(locale)
		if normalizedLocale == "" || catalog == nil {
			continue
		}
		name := strings.TrimSpace(catalog[nameKey])
		summary := strings.TrimSpace(catalog[descriptionKey])
		if name == "" && summary == "" {
			continue
		}
		item := byLocale[normalizedLocale]
		if item == nil {
			item = &marketplaceDisplayI18nItem{Locale: normalizedLocale}
			byLocale[normalizedLocale] = item
		}
		if name != "" {
			item.Name = truncateRunes(name, maxDisplayNameRunes)
		}
		if summary != "" {
			item.Summary = truncateRunes(summary, maxDisplaySummaryRunes)
		}
		item.Source = displaySourcePackageI18N
		if item.Name == "" {
			item.Name = firstNonEmpty(baseName(base), pluginID)
			item.Name = truncateRunes(item.Name, maxDisplayNameRunes)
		}
		if item.Summary == "" {
			item.Summary = firstNonEmpty(baseSummary(base), item.Name)
			item.Summary = truncateRunes(item.Summary, maxDisplaySummaryRunes)
		}
	}
	return mapDisplayItemsToSlice(byLocale)
}

// extractSourcePackageDisplayCatalogs reads non-apidoc JSON catalogs under manifest/i18n.
func extractSourcePackageDisplayCatalogs(
	files map[string]*zip.File,
	rootPrefix string,
) (map[string]map[string]string, error) {
	out := make(map[string]map[string]string)
	prefix := rootPrefix + "manifest/i18n/"
	for filePath, file := range files {
		if file == nil || !strings.HasPrefix(filePath, prefix) {
			continue
		}
		if strings.ToLower(filepath.Ext(filePath)) != ".json" {
			continue
		}
		relative := strings.TrimPrefix(filePath, prefix)
		segments := strings.Split(relative, "/")
		if len(segments) < 2 {
			continue
		}
		if len(segments) >= 3 && segments[1] == "apidoc" {
			continue
		}
		locale := normalizeDisplayLocale(segments[0])
		if locale == "" {
			continue
		}
		content, err := readZipFile(file)
		if err != nil {
			return nil, err
		}
		catalog, err := parseFlatI18nJSON(content)
		if err != nil {
			continue
		}
		merged := out[locale]
		if merged == nil {
			merged = make(map[string]string)
			out[locale] = merged
		}
		for key, value := range catalog {
			merged[key] = value
		}
	}
	return out, nil
}

// extractDynamicPackageDisplayCatalogs maps embedded runtime i18n assets to catalogs.
func extractDynamicPackageDisplayCatalogs(
	assets []*dynamicPackageLocaleAsset,
) map[string]map[string]string {
	out := make(map[string]map[string]string)
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		locale := normalizeDisplayLocale(asset.Locale)
		if locale == "" || strings.TrimSpace(asset.Content) == "" {
			continue
		}
		catalog, err := parseFlatI18nJSON([]byte(asset.Content))
		if err != nil {
			continue
		}
		out[locale] = catalog
	}
	return out
}

// parseFlatI18nJSON accepts flat dotted keys or one-level nested objects.
func parseFlatI18nJSON(content []byte) (map[string]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	flattenI18nValue("", raw, out)
	return out, nil
}

// flattenI18nValue writes nested JSON string leaves as dotted keys.
func flattenI18nValue(prefix string, value interface{}, out map[string]string) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			next := key
			if prefix != "" {
				next = prefix + "." + key
			}
			flattenI18nValue(next, nested, out)
		}
	case string:
		if prefix != "" {
			out[prefix] = typed
		}
	}
}

// normalizeDisplayI18nItems trims and drops empty locale rows.
func normalizeDisplayI18nItems(items []*marketplaceDisplayI18nItem) []*marketplaceDisplayI18nItem {
	out := make([]*marketplaceDisplayI18nItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		locale := normalizeDisplayLocale(item.Locale)
		name := truncateRunes(strings.TrimSpace(item.Name), maxDisplayNameRunes)
		summary := truncateRunes(strings.TrimSpace(item.Summary), maxDisplaySummaryRunes)
		if locale == "" || name == "" {
			continue
		}
		if summary == "" {
			summary = name
		}
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = displaySourcePluginYAML
		}
		out = append(out, &marketplaceDisplayI18nItem{
			Locale:  locale,
			Name:    name,
			Summary: summary,
			Source:  source,
		})
	}
	return out
}

// normalizeDisplayLocale normalizes locale tags used by marketplace display rows.
func normalizeDisplayLocale(locale string) string {
	trimmed := strings.TrimSpace(locale)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "_", "-")
	parts := strings.Split(trimmed, "-")
	if len(parts) == 1 {
		return strings.ToLower(parts[0])
	}
	// Keep BCP47-ish language-REGION form used by the host (zh-CN, en-US).
	language := strings.ToLower(parts[0])
	region := strings.ToUpper(parts[len(parts)-1])
	if language == "" || region == "" {
		return ""
	}
	return language + "-" + region
}

func baseName(items []*marketplaceDisplayI18nItem) string {
	for _, item := range items {
		if item != nil && strings.TrimSpace(item.Name) != "" {
			return item.Name
		}
	}
	return ""
}

func baseSummary(items []*marketplaceDisplayI18nItem) string {
	for _, item := range items {
		if item != nil && strings.TrimSpace(item.Summary) != "" {
			return item.Summary
		}
	}
	return ""
}

func mapDisplayItemsToSlice(byLocale map[string]*marketplaceDisplayI18nItem) []*marketplaceDisplayI18nItem {
	out := make([]*marketplaceDisplayI18nItem, 0, len(byLocale))
	for _, item := range byLocale {
		if item == nil {
			continue
		}
		out = append(out, item)
	}
	return normalizeDisplayI18nItems(out)
}

func uniquePositiveInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// defaultLocaleFromManifest returns plugin.yaml i18n.default or en-US.
func defaultLocaleFromManifest(defaultLocale string) string {
	if locale := normalizeDisplayLocale(defaultLocale); locale != "" {
		return locale
	}
	return defaultDisplayLocale
}
