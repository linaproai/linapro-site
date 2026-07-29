// This file verifies locale normalization, runtime bundle aggregation, and
// context-aware translation behavior for the host i18n service.

package i18n

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gvalid"
	"lina-core/internal/model"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/i18nresource"
	"lina-core/pkg/plugin/pluginhost"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

const testPluginID = "plugin-i18n-test"
const testCacheInvalidatePluginID = "plugin-i18n-cache-invalidate"

// stubConfigService supplies focused i18n config fixtures without requiring a
// full host config service implementation for locale tests.
type stubConfigService struct {
	hostconfig.Service
	cfg *hostconfig.I18nConfig
}

// GetI18n returns the fixture i18n config for locale tests.
func (s stubConfigService) GetI18n(_ context.Context) *hostconfig.I18nConfig {
	return s.cfg
}

// init registers one minimal source plugin fixture with embedded i18n assets.
func init() {
	plugin := pluginhost.NewDeclarations(testPluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(testPluginID, true))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(`{
  "plugin": {
    "plugin-i18n-test": {
      "name": "Runtime Test Plugin"
    }
  }
}`)},
	})
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}

// resetRuntimeBundleCache clears the in-memory runtime bundle cache between tests.
func resetRuntimeBundleCache() {
	invalidateRuntimeBundleCache()
	invalidateRuntimeLocaleCache()
}

// runtimeLocaleDescriptorsForTest returns the enabled runtime locale
// descriptors from the same file-backed config used by production startup.
func runtimeLocaleDescriptorsForTest(t *testing.T) []LocaleDescriptor {
	t.Helper()

	svc, ok := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil)).(*serviceImpl)
	if !ok {
		t.Fatal("expected i18n.New to return *serviceImpl")
	}
	locales := svc.loadEnabledRuntimeLocales(context.Background())
	if len(locales) == 0 {
		t.Fatal("expected at least one configured runtime locale")
	}
	return locales
}

// TestNormalizeLocale verifies that raw locale aliases normalize to canonical locale codes.
func TestNormalizeLocale(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		raw      string
		expected string
	}{
		{name: "zh short tag", raw: "zh", expected: "zh"},
		{name: "zh underscore", raw: "zh_CN", expected: DefaultLocale},
		{name: "english us", raw: "en-US", expected: EnglishLocale},
		{name: "traditional chinese", raw: "zh_tw", expected: "zh-TW"},
		{name: "english gb", raw: "en-gb", expected: "en-GB"},
		{name: "french", raw: "fr-fr", expected: "fr-FR"},
		{name: "script tag", raw: "zh_hans_cn", expected: "zh-Hans-CN"},
		{name: "invalid", raw: "zh-中文", expected: ""},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if actual := normalizeLocale(testCase.raw); actual != testCase.expected {
				t.Fatalf("expected locale %q, got %q", testCase.expected, actual)
			}
		})
	}
}

// TestNormalizeAcceptLanguage verifies that the first valid language tag is normalized.
func TestNormalizeAcceptLanguage(t *testing.T) {
	t.Parallel()

	header := "fr-FR, en-GB;q=0.8, zh-CN;q=0.6"
	if actual := normalizeAcceptLanguage(header); actual != "fr-FR" {
		t.Fatalf("expected accept-language locale %q, got %q", "fr-FR", actual)
	}
}

// TestResolveLocaleFallsBackToDefault verifies that explicit unsupported locales
// fall back to the configured runtime default language.
func TestResolveLocaleFallsBackToDefault(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	if actual := svc.ResolveLocale(context.Background(), "fr-FR"); actual != DefaultLocale {
		t.Fatalf("expected unsupported locale to fall back to %q, got %q", DefaultLocale, actual)
	}
}

// TestParseLocaleJSONSupportsFlatKeys verifies runtime locale resources can be
// maintained with the current flat dotted-key format.
func TestParseLocaleJSONSupportsFlatKeys(t *testing.T) {
	t.Parallel()

	flatCatalog := parseLocaleJSON([]byte(`{
  "menu.dashboard.title": "Workbench",
  "plugin.demo.name": "Demo"
}`))
	if actual := flatCatalog["menu.dashboard.title"]; actual != "Workbench" {
		t.Fatalf("expected flat key translation %q, got %q", "Workbench", actual)
	}
}

// TestBuildRuntimeMessagesIncludesHostAndSourcePlugin verifies that the runtime
// message bundle merges host translations with registered source-plugin assets.
func TestBuildRuntimeMessagesIncludesHostAndSourcePlugin(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	messages := svc.BuildRuntimeMessages(context.Background(), EnglishLocale)

	if actual, ok := lookupMessageString(messages, "menu.dashboard.title"); !ok || actual != "Dashboard" {
		t.Fatalf("expected host menu translation %q, got %q (exists=%v)", "Dashboard", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "dict.cron_job_status.name"); !ok || actual != "Scheduled Job Status" {
		t.Fatalf("expected scheduled-job dict translation %q, got %q (exists=%v)", "Scheduled Job Status", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "dict.sys_menu_type.B.label"); !ok || actual != "Button" {
		t.Fatalf("expected built-in menu-type translation %q, got %q (exists=%v)", "Button", actual, ok)
	}
	if actual, ok := lookupMessageString(messages, "plugin.plugin-i18n-test.name"); !ok || actual != "Runtime Test Plugin" {
		t.Fatalf("expected plugin translation %q, got %q (exists=%v)", "Runtime Test Plugin", actual, ok)
	}
}

// TestRuntimeLocalesUsesRequestedDisplayLocale verifies that the runtime
// locale list exposes localized display names and stable native names.
func TestRuntimeLocalesUsesRequestedDisplayLocale(t *testing.T) {
	resetRuntimeBundleCache()

	var (
		svc             = New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
		expectedLocales = runtimeLocaleDescriptorsForTest(t)
		output          = svc.RuntimeLocales(context.Background(), EnglishLocale)
		locales         = output.Items
	)
	if !output.Enabled {
		t.Fatal("expected runtime language switching to be enabled")
	}
	if len(locales) != len(expectedLocales) {
		t.Fatalf("expected %d runtime locales, got %d", len(expectedLocales), len(locales))
	}

	localeMap := make(map[string]LocaleDescriptor, len(locales))
	for _, locale := range locales {
		localeMap[locale.Locale] = locale
	}

	for _, expected := range expectedLocales {
		actual, ok := localeMap[expected.Locale]
		if !ok {
			t.Fatalf("expected locale %q to be returned", expected.Locale)
		}
		if actual.Name == "" {
			t.Fatalf("expected locale %q to have localized display name", expected.Locale)
		}
		expectedNativeName := strings.TrimSpace(expected.NativeName)
		if expectedNativeName == "" {
			expectedNativeName = expected.Locale
		}
		if actual.NativeName != expectedNativeName {
			t.Fatalf("expected locale %q native name %q, got %q", expected.Locale, expectedNativeName, actual.NativeName)
		}
		if actual.Direction != LocaleDirectionLTR.String() {
			t.Fatalf("expected locale %q direction %q, got %q", expected.Locale, LocaleDirectionLTR.String(), actual.Direction)
		}
		if actual.IsDefault != expected.IsDefault {
			t.Fatalf("expected locale %q default marker %v, got %v", expected.Locale, expected.IsDefault, actual.IsDefault)
		}
	}
}

// TestBuildConfiguredRuntimeLocalesUsesConfigLocalesAsWhitelist verifies that
// removing a locale from config i18n.locales disables it even when its JSON
// resource file still exists.
func TestBuildConfiguredRuntimeLocalesUsesConfigLocalesAsWhitelist(t *testing.T) {
	t.Parallel()

	config := &hostconfig.I18nConfig{
		Default: DefaultLocale,
		Enabled: true,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: EnglishLocale, NativeName: "English"},
			{Locale: DefaultLocale, NativeName: "简体中文"},
		},
	}
	locales := normalizeRuntimeLocales(buildConfiguredRuntimeLocales(
		[]string{DefaultLocale, EnglishLocale, "fr-FR"},
		config,
	), config.Default)

	if len(locales) != 2 {
		t.Fatalf("expected 2 enabled locales, got %d: %+v", len(locales), locales)
	}
	for _, locale := range locales {
		if locale.Locale == "fr-FR" {
			t.Fatalf("expected locale absent from config to be disabled: %+v", locales)
		}
	}
}

// TestBuildConfiguredRuntimeLocalesDisabledReturnsDefaultOnly verifies that
// i18n.enabled=false suppresses all non-default runtime locales.
func TestBuildConfiguredRuntimeLocalesDisabledReturnsDefaultOnly(t *testing.T) {
	t.Parallel()

	config := &hostconfig.I18nConfig{
		Default: DefaultLocale,
		Enabled: false,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: EnglishLocale, NativeName: "English"},
			{Locale: DefaultLocale, NativeName: "简体中文"},
			{Locale: "fr-FR", NativeName: "Français"},
		},
	}
	locales := normalizeRuntimeLocales(buildConfiguredRuntimeLocales(
		[]string{DefaultLocale, EnglishLocale, "fr-FR"},
		config,
	), config.Default)

	if len(locales) != 1 {
		t.Fatalf("expected only one locale when i18n is disabled, got %d: %+v", len(locales), locales)
	}
	if locales[0].Locale != DefaultLocale || !locales[0].IsDefault {
		t.Fatalf("expected disabled i18n to keep only default locale, got %+v", locales[0])
	}
}

// TestResolveLocaleUsesDefaultWhenI18nDisabled verifies explicit non-default
// locale requests are ignored when runtime language switching is disabled.
func TestResolveLocaleUsesDefaultWhenI18nDisabled(t *testing.T) {
	resetRuntimeBundleCache()
	t.Cleanup(resetRuntimeBundleCache)

	cfg := &hostconfig.I18nConfig{
		Default: DefaultLocale,
		Enabled: false,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: DefaultLocale, NativeName: "简体中文"},
			{Locale: EnglishLocale, NativeName: "English"},
			{Locale: "fr-FR", NativeName: "Français"},
		},
	}
	svc := &serviceImpl{configSvc: stubConfigService{cfg: cfg}}

	if actual := svc.ResolveLocale(context.Background(), EnglishLocale); actual != DefaultLocale {
		t.Fatalf("expected disabled i18n to resolve explicit locale to %q, got %q", DefaultLocale, actual)
	}
	if actual := svc.ResolveLocale(context.Background(), "fr-FR"); actual != DefaultLocale {
		t.Fatalf("expected disabled i18n to resolve non-default locale to %q, got %q", DefaultLocale, actual)
	}
}

// TestFallbackRuntimeLocalesUsesConfiguredDefault verifies the last-resort
// runtime locale list is still driven by i18n.default.
func TestFallbackRuntimeLocalesUsesConfiguredDefault(t *testing.T) {
	t.Parallel()

	locales := fallbackRuntimeLocales(&hostconfig.I18nConfig{Default: EnglishLocale})

	if len(locales) != 1 {
		t.Fatalf("expected one fallback locale, got %d: %+v", len(locales), locales)
	}
	if locales[0].Locale != EnglishLocale || !locales[0].IsDefault {
		t.Fatalf("expected fallback locale to use configured default, got %+v", locales[0])
	}
}

// TestGetDefaultRuntimeLocaleUsesConfiguredDefault verifies default-locale
// resolution does not depend on the package-level test locale constants.
func TestGetDefaultRuntimeLocaleUsesConfiguredDefault(t *testing.T) {
	resetRuntimeBundleCache()

	cfg := &hostconfig.I18nConfig{
		Default: EnglishLocale,
		Enabled: false,
		Locales: []hostconfig.I18nLocaleConfig{
			{Locale: DefaultLocale, NativeName: "简体中文"},
			{Locale: EnglishLocale, NativeName: "English"},
		},
	}
	svc := &serviceImpl{configSvc: stubConfigService{cfg: cfg}}

	if actual := svc.getDefaultRuntimeLocale(context.Background()); actual != EnglishLocale {
		t.Fatalf("expected configured default locale %q, got %q", EnglishLocale, actual)
	}
}

// TestRegisterSourcePluginInvalidatesRuntimeBundleCache verifies that source
// plugin registrations clear the cached runtime bundle so new translations are visible.
func TestRegisterSourcePluginInvalidatesRuntimeBundleCache(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	messages := svc.BuildRuntimeMessages(context.Background(), EnglishLocale)
	if _, ok := lookupMessageString(messages, "plugin."+testCacheInvalidatePluginID+".name"); ok {
		t.Fatalf("expected plugin %q translation to be absent before registration", testCacheInvalidatePluginID)
	}

	plugin := pluginhost.NewDeclarations(testCacheInvalidatePluginID)
	plugin.Assets().UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml": &fstest.MapFile{Data: []byte(sourcePluginI18NManifestFixture(testCacheInvalidatePluginID, true))},
		"manifest/i18n/en-US/plugin.json": &fstest.MapFile{Data: []byte(`{
  "plugin": {
    "plugin-i18n-cache-invalidate": {
      "name": "Cache Invalidation Plugin"
    }
  }
}`)},
	})
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		t.Fatalf("failed to register source plugin fixture: %v", err)
	}

	messages = svc.BuildRuntimeMessages(context.Background(), EnglishLocale)
	if actual, ok := lookupMessageString(messages, "plugin."+testCacheInvalidatePluginID+".name"); !ok || actual != "Cache Invalidation Plugin" {
		t.Fatalf("expected cache-invalidated plugin translation %q, got %q (exists=%v)", "Cache Invalidation Plugin", actual, ok)
	}
}

// TestTranslateUsesContextLocaleAndFallback verifies that Translate resolves the
// locale from business context and falls back to the provided literal when needed.
func TestTranslateUsesContextLocaleAndFallback(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	if actual := svc.Translate(ctx, "framework.description", "fallback"); actual == "fallback" {
		t.Fatal("expected translated framework description, got fallback")
	}
	if actual := svc.Translate(ctx, "missing.translation.key", "fallback"); actual != "fallback" {
		t.Fatalf("expected fallback value %q, got %q", "fallback", actual)
	}
	if actual := svc.Translate(ctx, "job.handler.host.session-cleanup.name", "Online Session Cleanup"); actual != "Online Session Cleanup" {
		t.Fatalf("expected source text fallback %q, got %q", "Online Session Cleanup", actual)
	}
}

// TestLocalizeErrorSupportsFormattedBusinessKeys verifies that backend error
// keys can be formatted after translation using gerror text arguments.
func TestLocalizeErrorSupportsFormattedBusinessKeys(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	actual := svc.LocalizeError(ctx, gerror.Newf("error.upload.fileTooLarge", 20))
	if actual != "File size must not exceed 20MB" {
		t.Fatalf("expected localized formatted error %q, got %q", "File size must not exceed 20MB", actual)
	}
}

// TestLocalizeErrorSupportsValidationKeys verifies that flat validation keys
// are translated after validation when they were stored as message IDs.
func TestLocalizeErrorSupportsValidationKeys(t *testing.T) {
	resetRuntimeBundleCache()

	svc := New(bizctx.New(), hostconfig.New(), cachecoord.Default(nil))
	ctx := context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: EnglishLocale})

	err := gvalid.New().
		Data("").
		Rules("required").
		Messages("validation.auth.login.username.required").
		Run(ctx)
	if err == nil {
		t.Fatal("expected validation error")
	}

	actual := svc.LocalizeError(ctx, err)
	if actual != "Please enter a username" {
		t.Fatalf("expected localized validation error %q, got %q", "Please enter a username", actual)
	}
}

// TestHostBizerrMessageKeysCoveredByErrorJSON ensures caller-visible host error
// codes resolve to a real runtime i18n key under manifest/i18n/<locale>/error.json.
func TestHostBizerrMessageKeysCoveredByErrorJSON(t *testing.T) {
	moduleRoot := resolveLinaCoreModuleRoot(t)
	codes := collectHostBizerrDefinitions(t, moduleRoot)
	if len(codes) == 0 {
		t.Fatal("expected host bizerr definitions under apps/lina-core")
	}

	locales := []string{DefaultLocale, EnglishLocale}
	for _, locale := range locales {
		keys := loadFlattenedErrorJSON(t, filepath.Join(moduleRoot, "manifest", "i18n", locale, "error.json"))
		var missing []string
		for _, item := range codes {
			if _, ok := keys[item.messageKey]; !ok {
				missing = append(missing, item.errorCode+" -> "+item.messageKey)
			}
		}
		if len(missing) > 0 {
			t.Fatalf(
				"locale %s missing %d bizerr messageKey(s) in error.json:\n  %s",
				locale,
				len(missing),
				strings.Join(missing, "\n  "),
			)
		}
	}
}

type hostBizerrDef struct {
	errorCode  string
	messageKey string
}

var (
	bizerrDefineCallPattern = regexp.MustCompile(
		`(?s)bizerr\.MustDefine\((.*?)\)`,
	)
	bizerrStringLiteralPattern = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

// collectHostBizerrDefinitions walks host Go sources (excluding tests) and
// derives messageKey for each MustDefine call via bizerr.MessageKey.
func collectHostBizerrDefinitions(t *testing.T, moduleRoot string) []hostBizerrDef {
	t.Helper()

	byCode := make(map[string]hostBizerrDef)
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if !strings.Contains(string(content), "bizerr.MustDefine") {
			return nil
		}
		for _, match := range bizerrDefineCallPattern.FindAllStringSubmatch(string(content), -1) {
			body := match[1]
			literals := bizerrStringLiteralPattern.FindAllStringSubmatch(body, -1)
			if len(literals) < 2 {
				continue
			}
			errorCode := literals[0][1]
			byCode[errorCode] = hostBizerrDef{
				errorCode:  errorCode,
				messageKey: bizerr.MessageKey(errorCode),
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk host module for bizerr definitions: %v", err)
	}

	result := make([]hostBizerrDef, 0, len(byCode))
	for _, item := range byCode {
		result = append(result, item)
	}
	return result
}

// loadFlattenedErrorJSON loads and flattens one host error.json catalog.
func loadFlattenedErrorJSON(t *testing.T, path string) map[string]string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var payload any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return flattenJSONStrings(payload, "")
}

// flattenJSONStrings flattens nested JSON objects into dotted string keys.
func flattenJSONStrings(value any, prefix string) map[string]string {
	result := make(map[string]string)
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			keyText := strings.TrimSpace(key)
			if keyText == "" {
				continue
			}
			next := keyText
			if prefix != "" {
				next = prefix + "." + keyText
			}
			for nestedKey, nestedValue := range flattenJSONStrings(nested, next) {
				result[nestedKey] = nestedValue
			}
		}
	case string:
		if prefix != "" {
			result[prefix] = typed
		}
	}
	return result
}

// resolveLinaCoreModuleRoot returns the apps/lina-core directory containing go.mod.
func resolveLinaCoreModuleRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path for lina-core module root")
	}
	dir := filepath.Dir(currentFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate apps/lina-core go.mod")
		}
		dir = parent
	}
}

// normalizeAcceptLanguage converts an Accept-Language header into the first valid locale tag.
func normalizeAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		languageTag := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale := normalizeLocale(languageTag); locale != "" {
			return locale
		}
	}
	return ""
}

// invalidateRuntimeBundleCache clears all runtime i18n bundle state for tests.
func invalidateRuntimeBundleCache() {
	runtimeBundleCache.invalidate(InvalidateScope{})
	resetRuntimeLocaleCache()
}

// invalidateRuntimeLocaleCache clears the cached locale descriptors for tests.
func invalidateRuntimeLocaleCache() {
	resetRuntimeLocaleCache()
}

// parseLocaleJSON unmarshals one locale JSON file into a flat message catalog for tests.
func parseLocaleJSON(content []byte) map[string]string {
	bundle, err := i18nresource.ParseCatalog(content, i18nresource.ValueModeStringifyScalars)
	if err != nil {
		return map[string]string{}
	}
	return bundle
}

// lookupMessageString retrieves one string message by dotted key path for tests.
func lookupMessageString(messages map[string]interface{}, key string) (string, bool) {
	if len(messages) == 0 {
		return "", false
	}

	current := interface{}(messages)
	for _, segment := range strings.Split(strings.TrimSpace(key), ".") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return "", false
		}
		nested, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		current, ok = nested[segment]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}
