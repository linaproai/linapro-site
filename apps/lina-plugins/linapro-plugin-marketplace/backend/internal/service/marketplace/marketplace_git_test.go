package marketplace

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestParseGitRepoURLSupportsGitHubAndGitee(t *testing.T) {
	t.Parallel()

	github, err := parseGitRepoURL("https://github.com/linaproai/demo-plugin")
	if err != nil {
		t.Fatalf("parse github: %v", err)
	}
	if github.Provider != marketv1.MarketplaceRepoProviderGitHub || github.Owner != "linaproai" || github.Name != "demo-plugin" {
		t.Fatalf("unexpected github ref: %+v", github)
	}
	if !strings.HasSuffix(github.CloneURL, ".git") {
		t.Fatalf("expected clone url with .git, got %s", github.CloneURL)
	}

	gitee, err := parseGitRepoURL("https://gitee.com/org/plugin.git")
	if err != nil {
		t.Fatalf("parse gitee: %v", err)
	}
	if gitee.Provider != marketv1.MarketplaceRepoProviderGitee || gitee.Owner != "org" || gitee.Name != "plugin" {
		t.Fatalf("unexpected gitee ref: %+v", gitee)
	}
}

func TestParseGitRepoURLRejectsUnsupportedHosts(t *testing.T) {
	t.Parallel()
	if _, err := parseGitRepoURL("https://gitlab.com/org/plugin"); err == nil {
		t.Fatal("expected unsupported host error")
	}
}

func TestVersionsSemanticallyEqual(t *testing.T) {
	t.Parallel()
	cases := []struct {
		tag, version string
		want         bool
	}{
		{"v1.2.0", "1.2.0", true},
		{"v1.2.0", "v1.2.0", true},
		{"1.2.0", "v1.2.0", true},
		{"v1.2.0", "1.0.0", false},
	}
	for _, tc := range cases {
		if got := versionsSemanticallyEqual(tc.tag, tc.version); got != tc.want {
			t.Fatalf("versionsSemanticallyEqual(%q,%q)=%v want %v", tc.tag, tc.version, got, tc.want)
		}
	}
}

func TestValidateGitSourceManifestRejectsDynamic(t *testing.T) {
	t.Parallel()
	err := validateGitSourceManifest(&gitPluginManifest{
		ID:      "demo-plugin",
		Version: "v1.0.0",
		Type:    "dynamic",
	})
	if err == nil {
		t.Fatal("expected dynamic type rejection")
	}
}

func TestDetectArchiveKind(t *testing.T) {
	t.Parallel()
	if detectArchiveKind("a.zip") != archiveKindZip {
		t.Fatal("zip")
	}
	if detectArchiveKind("a.tar.gz") != archiveKindTarGz {
		t.Fatal("tar.gz")
	}
	if detectArchiveKind("a.tgz") != archiveKindTarGz {
		t.Fatal("tgz")
	}
	if detectArchiveKind("a.rar") != "" {
		t.Fatal("unsupported should be empty")
	}
}

func TestEncryptDecryptMarketplaceSecret(t *testing.T) {
	t.Parallel()
	cipherText, err := encryptMarketplaceSecret("token-value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if cipherText == "" || cipherText == "token-value" {
		t.Fatalf("cipher text should not be plaintext: %q", cipherText)
	}
	plain, err := decryptMarketplaceSecret(cipherText)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != "token-value" {
		t.Fatalf("got %q", plain)
	}
}

type stubGitClient struct {
	tags     []string
	branches map[string]bool
	commits  map[string]string // ref -> commit SHA
	tree     []string
	files    map[string][]byte
	err      error
}

func (s stubGitClient) ListTags(ctx context.Context, repo gitRepoRef, token string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tags, nil
}

func (s stubGitClient) RefExists(ctx context.Context, repo gitRepoRef, ref string, token string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	if s.branches != nil {
		return s.branches[ref], nil
	}
	// Default: main exists when not explicitly configured.
	return ref == gitFallbackBranchMain, nil
}

func (s stubGitClient) ResolveCommitSHA(ctx context.Context, repo gitRepoRef, ref string, token string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if s.commits != nil {
		if sha, ok := s.commits[ref]; ok {
			return sha, nil
		}
	}
	// Deterministic default for unit tests when commits map is omitted.
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

func (s stubGitClient) ListTreePaths(ctx context.Context, repo gitRepoRef, ref string, token string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tree, nil
}

func (s stubGitClient) ReadFile(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	key := ref + ":" + filePath
	if body, ok := s.files[key]; ok {
		return body, nil
	}
	return nil, errors.New(filePath + " not found at ref " + ref)
}

func (s stubGitClient) PathExists(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) (bool, error) {
	_, err := s.ReadFile(ctx, repo, ref, filePath, token)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	return false, err
}

func TestMapGitClientErrorAuth(t *testing.T) {
	t.Parallel()
	err := mapGitClientError(newGitAuthError("repository authentication failed"))
	if err == nil || !strings.Contains(err.Error(), "authentication") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGitHostingHTTPErrorDistinguishesRateLimitFromAuth(t *testing.T) {
	t.Parallel()

	rateLimitBody := []byte(`{"message":"API rate limit exceeded for 1.2.3.4."}`)
	err := gitHostingHTTPError(http.StatusForbidden, rateLimitBody, "", "list tags")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if isGitAuthError(err) {
		t.Fatalf("rate limit must not be auth error: %v", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "rate limit") {
		t.Fatalf("expected rate limit message, got %v", err)
	}

	authErr := gitHostingHTTPError(http.StatusUnauthorized, []byte(`{"message":"Bad credentials"}`), "token", "list tags")
	if !isGitAuthError(authErr) {
		t.Fatalf("expected auth error, got %v", authErr)
	}

	if err = gitHostingHTTPError(http.StatusOK, nil, "", "list tags"); err != nil {
		t.Fatalf("status 200 should pass: %v", err)
	}
}

type stubPluginConfig struct {
	values map[string]string
	err    error
}

func (s stubPluginConfig) Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if s.err != nil {
		return nil, s.err
	}
	if value, ok := s.values[key]; ok {
		return gvar.New(value), nil
	}
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

func (s stubPluginConfig) Exists(ctx context.Context, key string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	_, ok := s.values[key]
	return ok, nil
}

func (s stubPluginConfig) Scan(ctx context.Context, key string, target any) error {
	return nil
}

func (s stubPluginConfig) String(ctx context.Context, key string, defaultValue string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if value, ok := s.values[key]; ok && strings.TrimSpace(value) != "" {
		return value, nil
	}
	return defaultValue, nil
}

func (s stubPluginConfig) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}

func (s stubPluginConfig) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	return defaultValue, nil
}

func (s stubPluginConfig) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

func TestResolveGitAccessTokenPrefersUserTokenThenPlatformConfig(t *testing.T) {
	t.Parallel()

	svc := &serviceImpl{pluginConfig: stubPluginConfig{
		values: map[string]string{
			configKeyGitHubAccessToken: "platform-github-token",
			configKeyGiteeAccessToken:  "platform-gitee-token",
		},
	}}
	ctx := context.Background()

	got, err := svc.resolveGitAccessToken(ctx, marketv1.MarketplaceRepoProviderGitHub, "user-token")
	if err != nil {
		t.Fatalf("user token: %v", err)
	}
	if got != "user-token" {
		t.Fatalf("expected user token, got %q", got)
	}

	got, err = svc.resolveGitAccessToken(ctx, marketv1.MarketplaceRepoProviderGitHub, "  ")
	if err != nil {
		t.Fatalf("github platform fallback: %v", err)
	}
	if got != "platform-github-token" {
		t.Fatalf("expected github platform token, got %q", got)
	}

	got, err = svc.resolveGitAccessToken(ctx, marketv1.MarketplaceRepoProviderGitee, "")
	if err != nil {
		t.Fatalf("gitee platform fallback: %v", err)
	}
	if got != "platform-gitee-token" {
		t.Fatalf("expected gitee platform token, got %q", got)
	}

	emptySvc := &serviceImpl{}
	got, err = emptySvc.resolveGitAccessToken(ctx, marketv1.MarketplaceRepoProviderGitHub, "")
	if err != nil {
		t.Fatalf("nil config: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty token without config, got %q", got)
	}
}

func TestPlatformGitAccessTokenConfigKey(t *testing.T) {
	t.Parallel()
	if platformGitAccessTokenConfigKey(marketv1.MarketplaceRepoProviderGitHub) != configKeyGitHubAccessToken {
		t.Fatal("github key")
	}
	if platformGitAccessTokenConfigKey(marketv1.MarketplaceRepoProviderGitee) != configKeyGiteeAccessToken {
		t.Fatal("gitee key")
	}
	if platformGitAccessTokenConfigKey("") != "" {
		t.Fatal("unknown provider should be empty")
	}
}

func TestDecodeGitHubContentsFileBase64(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"type":"file","encoding":"base64","content":"aWQ6IGRlbW8K\n"}`)
	body, err := decodeGitHubContentsFile(payload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(body) != "id: demo\n" {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestFilterSemverTags(t *testing.T) {
	t.Parallel()
	got := filterSemverTags([]string{"v1.0.0", "release", "v2.1.0", "not-a-version"})
	if len(got) != 2 || got[0] != "v1.0.0" || got[1] != "v2.1.0" {
		t.Fatalf("unexpected tags: %#v", got)
	}
}

func TestResolveGitDiscoveryRefsPrefersSemverTags(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{gitRemote: stubGitClient{
		tags:     []string{"v1.2.0", "docs"},
		branches: map[string]bool{"main": true},
	}}
	refs, err := svc.resolveGitDiscoveryRefs(context.Background(), svc.gitClient(), gitRepoRef{}, "")
	if err != nil {
		t.Fatalf("resolve refs: %v", err)
	}
	if len(refs) != 1 || refs[0].Name != "v1.2.0" || refs[0].Kind != gitDiscoveryRefKindTag {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}

func TestResolveGitDiscoveryRefsFallsBackToMain(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{gitRemote: stubGitClient{
		tags:     []string{"docs"},
		branches: map[string]bool{"main": true},
	}}
	refs, err := svc.resolveGitDiscoveryRefs(context.Background(), svc.gitClient(), gitRepoRef{}, "")
	if err != nil {
		t.Fatalf("resolve refs: %v", err)
	}
	if len(refs) != 1 || refs[0].Name != "main" || refs[0].Kind != gitDiscoveryRefKindBranch {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}

func TestDistributionFromEntitiesPrefersPinnedSourceCommit(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{}
	plugin := &entity.PluginMarketplacePlugin{
		PluginId:     "demo-plugin",
		SourceKind:   gitSourceKind,
		RepoUrl:      "https://github.com/org/demo-plugin.git",
		RepoProvider: marketv1.MarketplaceRepoProviderGitHub.String(),
		RepoPath:     "",
	}
	const pinned = "abc123def4567890abc123def4567890abc123de"
	release := &entity.PluginMarketplaceRelease{
		PluginId:       "demo-plugin",
		ReleaseVersion: "v0.2.0",
		PluginType:     marketv1.MarketplacePluginTypeSource.String(),
		SourceRef:      gitFallbackBranchMain,
		SourceCommit:   pinned,
	}
	item, err := svc.distributionFromEntities(context.Background(), plugin, release)
	if err != nil {
		t.Fatalf("distribution: %v", err)
	}
	if item.Mode != marketv1.MarketplaceDistributionModeGit {
		t.Fatalf("mode=%s", item.Mode)
	}
	if item.Ref != pinned {
		t.Fatalf("ref must prefer pinned commit, got %q want %q", item.Ref, pinned)
	}
	if item.Version != "v0.2.0" {
		t.Fatalf("version drift: %q", item.Version)
	}
}

func TestDistributionFromEntitiesFallsBackToSourceRefWithoutCommit(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{}
	plugin := &entity.PluginMarketplacePlugin{
		PluginId:     "demo-plugin",
		SourceKind:   gitSourceKind,
		RepoUrl:      "https://github.com/org/demo-plugin.git",
		RepoProvider: marketv1.MarketplaceRepoProviderGitHub.String(),
	}
	release := &entity.PluginMarketplaceRelease{
		PluginId:       "demo-plugin",
		ReleaseVersion: "v1.0.0",
		PluginType:     marketv1.MarketplacePluginTypeSource.String(),
		SourceRef:      "v1.0.0",
		SourceCommit:   "",
	}
	item, err := svc.distributionFromEntities(context.Background(), plugin, release)
	if err != nil {
		t.Fatalf("distribution: %v", err)
	}
	if item.Ref != "v1.0.0" {
		t.Fatalf("ref=%q", item.Ref)
	}
}

func TestStubGitClientResolveCommitSHA(t *testing.T) {
	t.Parallel()
	client := stubGitClient{
		commits: map[string]string{
			"main": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
	}
	sha, err := client.ResolveCommitSHA(context.Background(), gitRepoRef{}, "main", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if sha != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("sha=%q", sha)
	}
}

func TestResolveGitDiscoveryRefsFailsWithoutTagsAndMain(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{gitRemote: stubGitClient{
		tags:     nil,
		branches: map[string]bool{},
	}}
	_, err := svc.resolveGitDiscoveryRefs(context.Background(), svc.gitClient(), gitRepoRef{}, "")
	if err == nil {
		t.Fatal("expected failure when no tags and no main")
	}
	if !strings.Contains(err.Error(), "main branch does not exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCandidatePluginRootsFromTree(t *testing.T) {
	t.Parallel()
	paths := []string{
		"plugin.yaml",
		"apps/lina-plugins/alpha/plugin.yaml",
		"apps/lina-plugins/alpha/backend/plugin.go",
		"apps/lina-plugins/beta/plugin.yaml",
		"node_modules/x/plugin.yaml",
		"vendor/y/plugin.yaml",
	}
	got := candidatePluginRootsFromTree(paths)
	if len(got) != 2 || got[0] != "apps/lina-plugins/alpha" || got[1] != "apps/lina-plugins/beta" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}

func TestGitPathSetHasSourceStructure(t *testing.T) {
	t.Parallel()
	pathSet := buildGitPathSet([]string{
		"apps/lina-plugins/alpha/plugin.yaml",
		"apps/lina-plugins/alpha/backend/plugin.go",
		"apps/lina-plugins/alpha/plugin_embed.go",
		"apps/lina-plugins/beta/plugin.yaml",
	})
	if !gitPathSetHasSourceStructure(pathSet, "apps/lina-plugins/alpha") {
		t.Fatal("alpha should have source structure")
	}
	if gitPathSetHasSourceStructure(pathSet, "apps/lina-plugins/beta") {
		t.Fatal("beta is missing backend/plugin.go and plugin_embed.go")
	}
	if gitPathSetHasSourceStructure(pathSet, "") {
		t.Fatal("repository root should not look like a source plugin")
	}
}

func TestBuildGitResourceSummariesFromTree(t *testing.T) {
	t.Parallel()
	tree := []string{
		"linapro-ai-core/plugin.yaml",
		"linapro-ai-core/backend/plugin.go",
		"linapro-ai-core/plugin_embed.go",
		"linapro-ai-core/manifest/sql/001-init.sql",
		"linapro-ai-core/manifest/i18n/zh-CN/plugin.json",
		"linapro-ai-core/manifest/i18n/en-US/apidoc/ai.json",
		"linapro-ai-core/manifest/docs/zh-CN/index.md",
		"linapro-ai-core/README.md",
		"other-plugin/manifest/docs/index.md",
	}
	sqlItems, i18nItems, docsItems := buildGitResourceSummariesFromTree(tree, "linapro-ai-core")
	if len(sqlItems) != 1 || sqlItems[0].Path != "manifest/sql/001-init.sql" {
		t.Fatalf("unexpected sql summary: %#v", sqlItems)
	}
	if len(docsItems) != 2 {
		t.Fatalf("expected readme + docs entry, got %#v", docsItems)
	}
	if len(i18nItems) != 2 {
		t.Fatalf("expected two i18n locales, got %#v", i18nItems)
	}
	selected := selectGitDocPathsForIndexing(tree, "linapro-ai-core")
	if len(selected) == 0 {
		t.Fatal("expected doc paths for indexing")
	}
	if selected[0] != "README.md" {
		t.Fatalf("expected README.md first, got %#v", selected)
	}
}

func TestResolveGitDocumentIdentityReadmeUsesEnUS(t *testing.T) {
	t.Parallel()
	sourceKind, locale, docPath := resolveGitDocumentIdentity("README.md")
	if sourceKind != documentSourceKindReadme || locale != fallbackEnUSLocale || docPath != readmeDocumentPath {
		t.Fatalf("unexpected README.md identity: kind=%q locale=%q path=%q", sourceKind, locale, docPath)
	}
	sourceKind, locale, docPath = resolveGitDocumentIdentity("readme.md")
	if sourceKind != documentSourceKindReadme || locale != fallbackEnUSLocale || docPath != readmeDocumentPath {
		t.Fatalf("unexpected lowercased README identity: kind=%q locale=%q path=%q", sourceKind, locale, docPath)
	}
	sourceKind, locale, docPath = resolveGitDocumentIdentity("README.zh-CN.md")
	if sourceKind != documentSourceKindReadme || locale != fallbackZhCNLocale || docPath != readmeCNDocumentPath {
		t.Fatalf("unexpected README.zh-CN.md identity: kind=%q locale=%q path=%q", sourceKind, locale, docPath)
	}
	sourceKind, locale, docPath = resolveGitDocumentIdentity("manifest/docs/zh-CN/index.md")
	if sourceKind != documentSourceKindManifestDocs || locale != fallbackZhCNLocale || docPath != defaultDocumentPath {
		t.Fatalf("unexpected manifest docs identity: kind=%q locale=%q path=%q", sourceKind, locale, docPath)
	}
}

func TestSelectGitRuntimeI18nPathsExcludesApidoc(t *testing.T) {
	t.Parallel()
	tree := []string{
		"demo/manifest/i18n/en-US/plugin.json",
		"demo/manifest/i18n/en-US/apidoc/api.json",
		"demo/manifest/i18n/zh-CN/plugin.json",
		"demo/manifest/i18n/zh-CN/menu.json",
		"demo/manifest/docs/zh-CN/index.md",
		"other/manifest/i18n/en-US/plugin.json",
	}
	selected := selectGitRuntimeI18nPaths(tree, "demo")
	if len(selected) != 3 {
		t.Fatalf("expected 3 runtime i18n files, got %#v", selected)
	}
	for _, path := range selected {
		if strings.Contains(path, "apidoc") {
			t.Fatalf("apidoc path should be excluded: %s", path)
		}
	}
}

func TestLoadGitDisplayI18nCatalogsMergesNestedPluginKeys(t *testing.T) {
	t.Parallel()
	tree := []string{
		"linapro-ai-core/manifest/i18n/en-US/plugin.json",
		"linapro-ai-core/manifest/i18n/en-US/apidoc/ai.json",
		"linapro-ai-core/manifest/i18n/zh-CN/plugin.json",
	}
	client := stubGitClient{
		files: map[string][]byte{
			"main:linapro-ai-core/manifest/i18n/en-US/plugin.json": []byte(`{
  "plugin": {
    "linapro-ai-core": {
      "name": "AI Hub",
      "description": "English summary"
    }
  }
}`),
			"main:linapro-ai-core/manifest/i18n/en-US/apidoc/ai.json": []byte(`{"api.x":"should not load"}`),
			"main:linapro-ai-core/manifest/i18n/zh-CN/plugin.json": []byte(`{
  "plugin": {
    "linapro-ai-core": {
      "name": "智能中心",
      "description": "中文摘要"
    }
  }
}`),
		},
	}
	catalogs := loadGitDisplayI18nCatalogs(
		context.Background(),
		client,
		gitRepoRef{},
		"main",
		"",
		"linapro-ai-core",
		tree,
	)
	if catalogs["en-US"]["plugin.linapro-ai-core.name"] != "AI Hub" {
		t.Fatalf("unexpected en-US catalog: %#v", catalogs["en-US"])
	}
	if catalogs["zh-CN"]["plugin.linapro-ai-core.description"] != "中文摘要" {
		t.Fatalf("unexpected zh-CN catalog: %#v", catalogs["zh-CN"])
	}
	if _, ok := catalogs["en-US"]["api.x"]; ok {
		t.Fatal("apidoc keys must not be loaded into display catalogs")
	}

	base := buildDisplayI18nFromPackageYAML(
		"linapro-ai-core",
		"AI Hub",
		"Official source plugin",
		"en-US",
	)
	items := mergePackageI18nDisplayItems("linapro-ai-core", base, catalogs)
	byLocale := map[string]*marketplaceDisplayI18nItem{}
	for _, item := range items {
		byLocale[item.Locale] = item
	}
	if byLocale["zh-CN"] == nil || byLocale["zh-CN"].Name != "智能中心" || byLocale["zh-CN"].Summary != "中文摘要" {
		t.Fatalf("expected zh-CN display from git i18n, got %#v", byLocale["zh-CN"])
	}
	if byLocale["en-US"] == nil || byLocale["en-US"].Name != "AI Hub" {
		t.Fatalf("expected en-US display from git i18n, got %#v", byLocale["en-US"])
	}
}

func TestDiscoverSourcePluginRootsSinglePlugin(t *testing.T) {
	t.Parallel()
	manifest := []byte("id: demo-plugin\nname: Demo\nversion: 1.0.0\ntype: source\n")
	client := stubGitClient{
		tree: []string{
			"plugin.yaml",
			"backend/plugin.go",
			"plugin_embed.go",
		},
		files: map[string][]byte{
			"main:plugin.yaml":       manifest,
			"main:backend/plugin.go": []byte("package backend"),
			"main:plugin_embed.go":   []byte("package demo"),
		},
	}
	svc := &serviceImpl{gitRemote: client}
	roots, err := svc.discoverSourcePluginRoots(context.Background(), client, gitRepoRef{}, "main", "")
	if err != nil {
		t.Fatalf("discover roots: %v", err)
	}
	if len(roots) != 1 || roots[0].Path != "" || roots[0].Manifest.ID != "demo-plugin" {
		t.Fatalf("unexpected roots: %#v", roots)
	}
}

func TestDiscoverSourcePluginRootsMultiPlugin(t *testing.T) {
	t.Parallel()
	client := stubGitClient{
		tree: []string{
			"apps/lina-plugins/alpha/plugin.yaml",
			"apps/lina-plugins/alpha/backend/plugin.go",
			"apps/lina-plugins/alpha/plugin_embed.go",
			"apps/lina-plugins/beta/plugin.yaml",
			"apps/lina-plugins/beta/backend/plugin.go",
			"apps/lina-plugins/beta/plugin_embed.go",
		},
		files: map[string][]byte{
			"main:apps/lina-plugins/alpha/plugin.yaml":       []byte("id: alpha-plugin\nname: Alpha\nversion: 0.1.0\ntype: source\n"),
			"main:apps/lina-plugins/alpha/backend/plugin.go": []byte("package backend"),
			"main:apps/lina-plugins/alpha/plugin_embed.go":   []byte("package alpha"),
			"main:apps/lina-plugins/beta/plugin.yaml":        []byte("id: beta-plugin\nname: Beta\nversion: 0.2.0\ntype: source\n"),
			"main:apps/lina-plugins/beta/backend/plugin.go":  []byte("package backend"),
			"main:apps/lina-plugins/beta/plugin_embed.go":    []byte("package beta"),
		},
	}
	svc := &serviceImpl{gitRemote: client}
	roots, err := svc.discoverSourcePluginRoots(context.Background(), client, gitRepoRef{}, "main", "")
	if err != nil {
		t.Fatalf("discover roots: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %#v", roots)
	}
	if roots[0].Path != "apps/lina-plugins/alpha" || roots[0].Manifest.ID != "alpha-plugin" {
		t.Fatalf("unexpected first root: %#v", roots[0])
	}
	if roots[1].Path != "apps/lina-plugins/beta" || roots[1].Manifest.ID != "beta-plugin" {
		t.Fatalf("unexpected second root: %#v", roots[1])
	}
}

func TestNormalizeGitRepoPathAndJoin(t *testing.T) {
	t.Parallel()
	if got := normalizeGitRepoPath("/apps/lina-plugins/demo/"); got != "apps/lina-plugins/demo" {
		t.Fatalf("normalize path: %q", got)
	}
	if got := gitPathJoin("apps/demo", "plugin.yaml"); got != "apps/demo/plugin.yaml" {
		t.Fatalf("join nested: %q", got)
	}
	if got := gitPathJoin("", "plugin.yaml"); got != "plugin.yaml" {
		t.Fatalf("join root: %q", got)
	}
}
