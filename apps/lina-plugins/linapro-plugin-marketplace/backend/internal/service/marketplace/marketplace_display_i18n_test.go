package marketplace

import (
	"testing"

	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestPickDisplayNameSummaryFallbackChain(t *testing.T) {
	t.Parallel()

	rows := []*entity.PluginMarketplaceDisplayI18n{
		{Locale: "en-US", Name: "English Name", Summary: "English summary"},
		{Locale: "zh-CN", Name: "中文名称", Summary: "中文摘要"},
	}

	name, summary := pickDisplayNameSummary(rows, "zh-CN", "en-US", "fallback", "fallback-summary")
	if name != "中文名称" || summary != "中文摘要" {
		t.Fatalf("expected zh-CN display, got name=%q summary=%q", name, summary)
	}

	name, summary = pickDisplayNameSummary(rows, "ja-JP", "en-US", "fallback", "fallback-summary")
	if name != "English Name" || summary != "English summary" {
		t.Fatalf("expected en-US fallback, got name=%q summary=%q", name, summary)
	}

	name, summary = pickDisplayNameSummary(nil, "zh-CN", "en-US", "identity", "identity-summary")
	if name != "identity" || summary != "identity-summary" {
		t.Fatalf("expected identity fallback, got name=%q summary=%q", name, summary)
	}
}

func TestMergePackageI18nDisplayItemsUsesNamespacedKeys(t *testing.T) {
	t.Parallel()

	base := buildDisplayI18nFromPackageYAML(
		"demo-plugin",
		"Demo Plugin",
		"Demo description from yaml",
		"en-US",
	)
	catalogs := map[string]map[string]string{
		"en-US": {
			"plugin.demo-plugin.name":        "Demo Plugin EN",
			"plugin.demo-plugin.description": "English package summary",
		},
		"zh-CN": {
			"plugin.demo-plugin.name":        "演示插件",
			"plugin.demo-plugin.description": "中文包摘要",
		},
	}
	items := mergePackageI18nDisplayItems("demo-plugin", base, catalogs)
	byLocale := map[string]*marketplaceDisplayI18nItem{}
	for _, item := range items {
		byLocale[item.Locale] = item
	}
	if byLocale["en-US"] == nil || byLocale["en-US"].Name != "Demo Plugin EN" {
		t.Fatalf("unexpected en-US item: %#v", byLocale["en-US"])
	}
	if byLocale["zh-CN"] == nil || byLocale["zh-CN"].Summary != "中文包摘要" {
		t.Fatalf("unexpected zh-CN item: %#v", byLocale["zh-CN"])
	}
	if byLocale["zh-CN"].Source != displaySourcePackageI18N {
		t.Fatalf("expected package_i18n source, got %q", byLocale["zh-CN"].Source)
	}
}

func TestBuildDisplayI18nFromPackageYAMLDefaultLocale(t *testing.T) {
	t.Parallel()

	items := buildDisplayI18nFromPackageYAML("demo", "Name", "Summary text", "zh-CN")
	if len(items) != 1 {
		t.Fatalf("expected one default row, got %d", len(items))
	}
	if items[0].Locale != "zh-CN" || items[0].Source != displaySourcePluginYAML {
		t.Fatalf("unexpected item: %#v", items[0])
	}
	if items[0].Name != "Name" || items[0].Summary != "Summary text" {
		t.Fatalf("unexpected name/summary: %#v", items[0])
	}
}

func TestParseFlatI18nJSONSupportsNestedAndFlat(t *testing.T) {
	t.Parallel()

	flat, err := parseFlatI18nJSON([]byte(`{"plugin.demo.name":"Flat"}`))
	if err != nil {
		t.Fatal(err)
	}
	if flat["plugin.demo.name"] != "Flat" {
		t.Fatalf("flat parse failed: %#v", flat)
	}

	nested, err := parseFlatI18nJSON([]byte(`{"plugin":{"demo":{"name":"Nested"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if nested["plugin.demo.name"] != "Nested" {
		t.Fatalf("nested parse failed: %#v", nested)
	}
}
