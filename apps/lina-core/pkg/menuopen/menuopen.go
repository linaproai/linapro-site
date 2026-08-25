// Package menuopen defines the generic menu open-mode codes returned by host
// navigation resources. Workbenches compile these modes into shell-specific
// routes. Stored workbench component paths and query keys are stripped from
// host resources instead of being used to infer open mode.
package menuopen

import (
	"path"
	"strings"

	"lina-core/pkg/plugin/pluginhost"
)

// Mode identifies how a navigable menu should be opened.
type Mode string

// Canonical open-mode values.
const (
	// Page opens an in-app page resource owned by the current workbench.
	Page Mode = "page"
	// Embedded mounts a hosted frontend asset inside the workbench shell.
	Embedded Mode = "embedded"
	// IFrame opens a hosted or remote URL inside an iframe.
	IFrame Mode = "iframe"
	// External opens a hosted or remote URL in a new window.
	External Mode = "external"
)

// Legacy workbench-local query keys and component paths still recognized when
// deriving open mode from stored menu paths. Host navigation no longer emits
// these values.
const (
	// workbenchAccessModeQueryKey is the former query key for plugin access mode.
	workbenchAccessModeQueryKey = "pluginAccessMode"
	// workbenchDynamicPageResource is the former hosted dynamic-page component path.
	workbenchDynamicPageResource = "system/plugin/dynamic-page"
	// workbenchEmbeddedSourceQueryKey is the former query key for embedded asset URLs.
	workbenchEmbeddedSourceQueryKey = "embeddedSrc"
)

// String returns the canonical persisted open-mode value.
func (mode Mode) String() string {
	return string(mode)
}

// IsValid reports whether the mode is a known non-empty value.
func (mode Mode) IsValid() bool {
	switch mode {
	case Page, Embedded, IFrame, External:
		return true
	default:
		return false
	}
}

// Resolve derives the generic open mode and target address from stored menu
// path and external-link flag. Target is empty for ordinary in-app pages.
func Resolve(menuPath string, isFrame int) (Mode, string) {
	target := LinkTarget(menuPath)
	if isFrame == 1 && target != "" {
		return External, target
	}
	if target != "" {
		if isHostedScript(menuPath) {
			return Embedded, target
		}
		return IFrame, target
	}
	return Page, ""
}

// LinkTarget returns the hosted-asset or remote URL stored on a menu path.
func LinkTarget(menuPath string) string {
	trimmedPath := strings.TrimSpace(menuPath)
	if trimmedPath == "" {
		return ""
	}
	lowerPath := strings.ToLower(trimmedPath)
	if strings.HasPrefix(lowerPath, "http://") || strings.HasPrefix(lowerPath, "https://") {
		return trimmedPath
	}
	normalizedHostedPath := "/" + strings.TrimLeft(trimmedPath, "/")
	if strings.HasPrefix(normalizedHostedPath, pluginhost.HostedAssetURLPrefix) {
		return normalizedHostedPath
	}
	return ""
}

// SanitizeQuery removes workbench-only query keys from a host navigation
// resource so the public contract stays shell-agnostic.
func SanitizeQuery(query map[string]string) map[string]string {
	if len(query) == 0 {
		return nil
	}
	sanitized := make(map[string]string, len(query))
	for key, value := range query {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" || trimmedKey == workbenchAccessModeQueryKey || trimmedKey == workbenchEmbeddedSourceQueryKey {
			continue
		}
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue == "" {
			continue
		}
		sanitized[trimmedKey] = trimmedValue
	}
	if len(sanitized) == 0 {
		return nil
	}
	return sanitized
}

// StripWorkbenchResource clears stored workbench component paths so plugin
// menu sync writes opaque resources instead of default-shell Vue paths.
func StripWorkbenchResource(component string) string {
	normalized := NormalizeResource(component)
	if normalized == workbenchDynamicPageResource {
		return ""
	}
	return normalized
}

// StripWorkbenchQuery removes workbench-only keys from a plugin menu query map.
func StripWorkbenchQuery(query map[string]interface{}) map[string]interface{} {
	if len(query) == 0 {
		return query
	}
	cleaned := make(map[string]interface{}, len(query))
	for key, value := range query {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" || trimmedKey == workbenchAccessModeQueryKey || trimmedKey == workbenchEmbeddedSourceQueryKey {
			continue
		}
		cleaned[trimmedKey] = value
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

// NormalizeResource strips workbench view prefixes from a stored component
// path so callers receive an opaque resource address.
func NormalizeResource(component string) string {
	normalized := strings.TrimSpace(component)
	normalized = strings.TrimPrefix(normalized, "#")
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "views/")
	normalized = strings.TrimPrefix(normalized, "views\\")
	normalized = strings.TrimSuffix(normalized, ".vue")
	return strings.ReplaceAll(normalized, "\\", "/")
}

// isHostedScript reports whether menuPath points at a hosted JavaScript bundle
// rather than an in-app page component.
func isHostedScript(menuPath string) bool {
	extension := strings.ToLower(path.Ext(strings.TrimSpace(menuPath)))
	return extension == ".js" || extension == ".mjs"
}
