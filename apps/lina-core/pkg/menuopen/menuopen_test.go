package menuopen

import "testing"

func TestResolveEmbeddedFromJSPath(t *testing.T) {
	mode, target := Resolve("/x-assets/demo/v1/mount.js", 0)
	if mode != Embedded || target != "/x-assets/demo/v1/mount.js" {
		t.Fatalf("expected embedded hosted js, got %s %s", mode, target)
	}
}

func TestResolveDoesNotUseWorkbenchComponentOrQuery(t *testing.T) {
	mode, target := Resolve("/x-assets/demo/v1/index.html", 0)
	if mode != IFrame || target != "/x-assets/demo/v1/index.html" {
		t.Fatalf("expected iframe hosted html, got %s %s", mode, target)
	}
}

func TestSanitizeQueryDropsWorkbenchKeys(t *testing.T) {
	query := SanitizeQuery(map[string]string{
		"pluginAccessMode": "embedded-mount",
		"embeddedSrc":      "/x-assets/demo/v1/mount.js",
		"tab":              "overview",
	})
	if query["tab"] != "overview" || query["pluginAccessMode"] != "" {
		t.Fatalf("expected workbench keys stripped, got %#v", query)
	}
}

func TestStripWorkbenchResource(t *testing.T) {
	if StripWorkbenchResource("system/plugin/dynamic-page") != "" {
		t.Fatal("expected workbench dynamic page path to be stripped")
	}
	if StripWorkbenchResource("system/user/index") != "system/user/index" {
		t.Fatal("expected ordinary resource to stay")
	}
	if StripWorkbenchResource("#/views/system/user/index.vue") != "system/user/index" {
		t.Fatal("expected view prefixes stripped from stored resource")
	}
}
