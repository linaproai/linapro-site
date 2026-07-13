package marketplace

import (
	"context"
	"errors"
	"strings"
	"testing"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
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
	tags  []string
	files map[string][]byte
	err   error
}

func (s stubGitClient) ListTags(ctx context.Context, repo gitRepoRef, token string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tags, nil
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
	err := mapGitClientError(gitAuthError("repository authentication failed"))
	if err == nil || !strings.Contains(err.Error(), "authentication") {
		t.Fatalf("unexpected error: %v", err)
	}
}
