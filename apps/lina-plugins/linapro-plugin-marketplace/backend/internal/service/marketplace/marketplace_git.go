// This file implements GitHub/Gitee metadata discovery for marketplace Git
// sources. It stores repository coordinates and version drafts only; full
// source trees are never cloned onto the marketplace server.

package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	gitSourceKind              = "git"
	uploadSourceKind           = "upload"
	gitSyncStatusSuccess       = "success"
	gitSyncStatusFailed        = "failed"
	gitSyncStatusAuthFailed    = "auth_failed"
	gitSyncStatusPartial       = "partial"
	defaultGitHTTPTimeout      = 20 * time.Second
	maxGitTagsPerDiscovery     = 100
	gitMetadataUserAgent       = "LinaPro-Plugin-Marketplace/1.0"
	gitRequiredPathPluginYAML  = "plugin.yaml"
	gitRequiredPathBackendGo   = "backend/plugin.go"
	gitRequiredPathEmbedGo     = "plugin_embed.go"
)

var (
	gitSemverTagPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)
	gitPluginIDPattern  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// gitRepoRef is a normalized GitHub/Gitee repository coordinate.
type gitRepoRef struct {
	Provider marketv1.MarketplaceRepoProvider
	Owner    string
	Name     string
	CloneURL string
	APIHost  string
}

// gitRemoteClient reads tags and raw files from Git hosting APIs.
type gitRemoteClient interface {
	ListTags(ctx context.Context, repo gitRepoRef, token string) ([]string, error)
	ReadFile(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) ([]byte, error)
	PathExists(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) (bool, error)
}

// httpGitRemoteClient is the default GitHub/Gitee REST client.
type httpGitRemoteClient struct {
	httpClient *http.Client
}

// RegisterGitSourceInput carries one publisher Git source registration.
type RegisterGitSourceInput struct {
	PublisherKey string
	OwnerUserID  int64
	RepoURL      string
	AccessToken  string
	Visibility   marketv1.MarketplaceVisibility
	Homepage     string
	License      string
}

// DiscoverGitMetadataInput identifies one Git plugin for metadata discovery.
type DiscoverGitMetadataInput struct {
	PluginID    string
	OwnerUserID int64 // optional ownership check; zero skips owner binding
}

// DiscoverGitMetadataResult summarizes one discovery run.
type DiscoverGitMetadataResult struct {
	Plugin *PluginRecord
	Synced int
}

// GetDistributionInput identifies one release distribution projection.
type GetDistributionInput struct {
	PluginID   string
	Version    string
	Visibility VisibilitySubject
}

// RegisterGitSource registers a Git repository and immediately discovers tags.
func (s *serviceImpl) RegisterGitSource(ctx context.Context, in RegisterGitSourceInput) (*PluginRecord, error) {
	if in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	publisher, err := s.requirePublisherOwnedByUser(ctx, in.PublisherKey, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	repo, err := parseGitRepoURL(in.RepoURL)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(in.AccessToken)
	client := s.gitClient()
	// Probe plugin.yaml at default branch HEAD via tags list first later; probe main/master files.
	tags, err := client.ListTags(ctx, repo, token)
	if err != nil {
		return nil, mapGitClientError(err)
	}
	if len(tags) == 0 {
		return nil, bizerr.NewCode(CodeMarketplaceGitDiscoveryFailed, bizerr.P("diagnostic", "repository has no version tags"))
	}

	// Bootstrap identity from the newest tag's plugin.yaml.
	newest := tags[0]
	manifestBytes, err := client.ReadFile(ctx, repo, newest, gitRequiredPathPluginYAML, token)
	if err != nil {
		return nil, mapGitClientError(err)
	}
	manifest, err := parseGitPluginManifest(manifestBytes)
	if err != nil {
		return nil, err
	}
	if err = validateGitSourceManifest(manifest); err != nil {
		return nil, err
	}

	existing, err := s.getPluginByID(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.PublisherId != publisher.Id {
			return nil, bizerr.NewCode(CodeMarketplacePluginIDOwned)
		}
		if normalizeSourceKind(existing.SourceKind) != gitSourceKind {
			return nil, bizerr.NewCode(CodeMarketplaceSourceKindConflict)
		}
	}

	credentialRef := ""
	if token != "" {
		credentialRef, err = s.saveGitCredential(ctx, in.OwnerUserID, repo.Provider.String(), token)
		if err != nil {
			return nil, err
		}
	}

	name := normalizeKey(manifest.Name)
	if name == "" {
		name = manifest.ID
	}
	summary := normalizeKey(manifest.Description)
	if summary == "" {
		summary = name
	}
	if len(summary) > 512 {
		summary = summary[:512]
	}

	// Git add keeps plugins private/draft until an explicit publish review passes.
	gitVisibility := marketv1.MarketplaceVisibilityPrivate
	if existing == nil {
		id, insertErr := dao.PluginMarketplacePlugin.Ctx(ctx).Data(do.PluginMarketplacePlugin{
			PublisherId:     publisher.Id,
			PluginId:        manifest.ID,
			Name:            name,
			Summary:         summary,
			Description:     normalizeKey(manifest.Description),
			PluginType:      marketv1.MarketplacePluginTypeSource.String(),
			MarketStatus:    marketv1.MarketplaceStatusDraft.String(),
			Visibility:      gitVisibility.String(),
			LatestReleaseId: 0,
			LatestVersion:   "",
			Homepage:        firstNonEmpty(in.Homepage, manifest.Homepage),
			Repository:      repo.CloneURL,
			License:         firstNonEmpty(in.License, manifest.License),
			DownloadCount:   0,
			SourceKind:      gitSourceKind,
			RepoUrl:         repo.CloneURL,
			RepoProvider:    repo.Provider.String(),
			CredentialRef:   credentialRef,
		}).InsertAndGetId()
		if insertErr != nil {
			return nil, bizerr.WrapCode(insertErr, CodeMarketplaceStorageFailed)
		}
		existing, err = s.getPluginEntityByRecordID(ctx, intID(id))
		if err != nil {
			return nil, err
		}
	} else {
		update := do.PluginMarketplacePlugin{
			Name:         name,
			Summary:      summary,
			Description:  normalizeKey(manifest.Description),
			Homepage:     firstNonEmpty(in.Homepage, manifest.Homepage, existing.Homepage),
			Repository:   repo.CloneURL,
			License:      firstNonEmpty(in.License, manifest.License, existing.License),
			SourceKind:   gitSourceKind,
			RepoUrl:      repo.CloneURL,
			RepoProvider: repo.Provider.String(),
		}
		// Preserve marketplace visibility for already-published plugins; only
		// force private while the plugin is still a draft add.
		if marketv1.MarketplaceStatus(existing.MarketStatus) == marketv1.MarketplaceStatusDraft {
			update.Visibility = gitVisibility.String()
		}
		if credentialRef != "" {
			update.CredentialRef = credentialRef
		}
		if _, updateErr := dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(do.PluginMarketplacePlugin{Id: existing.Id}).
			Data(update).
			Update(); updateErr != nil {
			return nil, bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		existing, err = s.getPluginEntityByRecordID(ctx, existing.Id)
		if err != nil {
			return nil, err
		}
	}

	result, err := s.discoverGitMetadataForPlugin(ctx, existing, token, client)
	if err != nil {
		return nil, err
	}
	return result.Plugin, nil
}

// DiscoverGitMetadata refreshes tags for one Git-backed marketplace plugin.
func (s *serviceImpl) DiscoverGitMetadata(ctx context.Context, in DiscoverGitMetadataInput) (*DiscoverGitMetadataResult, error) {
	plugin, err := s.getPluginByID(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if normalizeSourceKind(plugin.SourceKind) != gitSourceKind {
		return nil, bizerr.NewCode(CodeMarketplaceSourceKindConflict)
	}
	if in.OwnerUserID > 0 {
		if _, err = s.requirePublisherIDOwnedByUser(ctx, plugin.PublisherId, in.OwnerUserID); err != nil {
			return nil, err
		}
	}
	token, err := s.loadGitCredentialToken(ctx, plugin.CredentialRef)
	if err != nil {
		return nil, err
	}
	return s.discoverGitMetadataForPlugin(ctx, plugin, token, s.gitClient())
}

// DiscoverAllGitSources scans every Git-backed plugin for new tags.
func (s *serviceImpl) DiscoverAllGitSources(ctx context.Context) (int, error) {
	var rows []*entity.PluginMarketplacePlugin
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{SourceKind: gitSourceKind}).
		OrderAsc(dao.PluginMarketplacePlugin.Columns().Id).
		Limit(500).
		Scan(&rows); err != nil {
		return 0, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	synced := 0
	client := s.gitClient()
	for _, row := range rows {
		if row == nil {
			continue
		}
		token, err := s.loadGitCredentialToken(ctx, row.CredentialRef)
		if err != nil {
			_ = s.updateGitSyncStatus(ctx, row.Id, gitSyncStatusFailed, "credential load failed")
			continue
		}
		result, err := s.discoverGitMetadataForPlugin(ctx, row, token, client)
		if err != nil {
			continue
		}
		synced += result.Synced
	}
	return synced, nil
}

// GetDistribution returns the CLI install projection for one visible release.
func (s *serviceImpl) GetDistribution(ctx context.Context, in GetDistributionInput) (*marketv1.MarketplaceDistributionItem, error) {
	plugin, owned, err := s.resolveAccessiblePlugin(ctx, in.PluginID, in.Visibility)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	release, err := s.requireRelease(ctx, plugin.PluginId, in.Version)
	if err != nil {
		return nil, err
	}
	if !releaseDistributionAllowed(release, owned, in.Visibility) {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	return s.distributionFromEntities(ctx, plugin, release)
}

func (s *serviceImpl) discoverGitMetadataForPlugin(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	token string,
	client gitRemoteClient,
) (*DiscoverGitMetadataResult, error) {
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	repo, err := parseGitRepoURL(plugin.RepoUrl)
	if err != nil {
		_ = s.updateGitSyncStatus(ctx, plugin.Id, gitSyncStatusFailed, err.Error())
		return nil, err
	}
	tags, err := client.ListTags(ctx, repo, token)
	if err != nil {
		status := gitSyncStatusFailed
		if isGitAuthError(err) {
			status = gitSyncStatusAuthFailed
		}
		_ = s.updateGitSyncStatus(ctx, plugin.Id, status, gitPublicErrorMessage(err))
		return nil, mapGitClientError(err)
	}

	synced := 0
	failures := 0
	for _, tag := range tags {
		if !gitSemverTagPattern.MatchString(tag) {
			continue
		}
		created, discoverErr := s.discoverOneGitTag(ctx, plugin, repo, tag, token, client)
		if discoverErr != nil {
			failures++
			continue
		}
		if created {
			synced++
		}
	}

	status := gitSyncStatusSuccess
	message := fmt.Sprintf("discovered %d draft releases", synced)
	if failures > 0 && synced == 0 {
		status = gitSyncStatusFailed
		message = fmt.Sprintf("failed to import %d tags", failures)
	} else if failures > 0 {
		status = gitSyncStatusPartial
		message = fmt.Sprintf("discovered %d drafts with %d tag failures", synced, failures)
	}
	if err = s.updateGitSyncStatus(ctx, plugin.Id, status, message); err != nil {
		return nil, err
	}
	record, err := s.getPluginRecordByID(ctx, plugin.Id)
	if err != nil {
		return nil, err
	}
	return &DiscoverGitMetadataResult{Plugin: record, Synced: synced}, nil
}

func (s *serviceImpl) discoverOneGitTag(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	repo gitRepoRef,
	tag string,
	token string,
	client gitRemoteClient,
) (createdOrUpdated bool, err error) {
	manifestBytes, err := client.ReadFile(ctx, repo, tag, gitRequiredPathPluginYAML, token)
	if err != nil {
		return false, err
	}
	manifest, err := parseGitPluginManifest(manifestBytes)
	if err != nil {
		return false, err
	}
	if err = validateGitSourceManifest(manifest); err != nil {
		return false, err
	}
	if normalizeKey(manifest.ID) != normalizeKey(plugin.PluginId) {
		return false, packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml id does not match marketplace plugin id")
	}
	if !versionsSemanticallyEqual(tag, manifest.Version) {
		return false, bizerr.NewCode(CodeMarketplaceGitVersionMismatch, bizerr.P("diagnostic", "tag "+tag+" does not match plugin.yaml version "+manifest.Version))
	}
	if err = validateGitRemoteStructure(ctx, client, repo, tag, token); err != nil {
		return false, err
	}

	existing, err := s.getReleaseByPluginVersion(ctx, plugin.PluginId, manifest.Version)
	if err != nil {
		return false, err
	}
	if existing != nil && immutableRelease(existing) {
		return false, nil
	}

	snapshot, err := packageJSONString(manifest)
	if err != nil {
		return false, err
	}
	deps, err := packageJSONString(buildSourceDependencySummary(manifestAsSource(manifest)))
	if err != nil {
		return false, err
	}
	in := SaveReleaseDraftInput{
		PublisherKey:      "", // filled via requirePluginForPublisher path below
		OwnerUserID:       0,
		PluginID:          plugin.PluginId,
		Version:           normalizeVersionLabel(manifest.Version),
		PluginType:        marketv1.MarketplacePluginTypeSource,
		Visibility:        marketv1.MarketplaceVisibility(plugin.Visibility),
		MinHostVersion:    sourcePackageHostBoundsFromManifest(manifest),
		ManifestSnapshot:  snapshot,
		DependencySummary: deps,
		SQLSummary:        "[]",
		I18NSummary:       "[]",
		DocsSummary:       "[]",
		RiskSummary:       `{"info":0,"warning":0,"high":0}`,
		ReviewMessage:     "Imported from Git tag " + tag,
		ReplaceDraft:      true,
		SourceRef:         tag,
	}

	// Save release draft with source_ref without full publisher ownership re-check
	// because discovery runs as a trusted system/owner path after plugin load.
	release, err := s.saveGitReleaseDraft(ctx, plugin, in)
	if err != nil {
		return false, err
	}
	return release != nil, nil
}

// saveGitReleaseDraft persists a mutable Git-sourced release draft.
func (s *serviceImpl) saveGitReleaseDraft(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	in SaveReleaseDraftInput,
) (*ReleaseRecord, error) {
	if plugin == nil || normalizeKey(in.Version) == "" {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	existing, err := s.getReleaseByPluginVersion(ctx, plugin.PluginId, in.Version)
	if err != nil {
		return nil, err
	}
	data := s.releaseDraftData(plugin, in)
	data.SourceRef = normalizeKey(in.SourceRef)
	if existing != nil {
		if immutableRelease(existing) {
			return nil, bizerr.NewCode(CodeMarketplaceReleaseImmutable)
		}
		if err = dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
			if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
				Where(do.PluginMarketplaceRelease{Id: existing.Id}).
				Data(data).
				Update(); updateErr != nil {
				return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
			}
			cols := dao.PluginMarketplaceRelease.Columns()
			if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
				Where(do.PluginMarketplaceRelease{Id: existing.Id}).
				Data(cols.SubmittedAt, nil, cols.ReviewedAt, nil, cols.PublishedAt, nil).
				Update(); updateErr != nil {
				return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return s.getReleaseRecordByID(ctx, existing.Id)
	}
	id, err := dao.PluginMarketplaceRelease.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getReleaseRecordByID(ctx, intID(id))
}

func (s *serviceImpl) distributionFromEntities(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	release *entity.PluginMarketplaceRelease,
) (*marketv1.MarketplaceDistributionItem, error) {
	item := &marketv1.MarketplaceDistributionItem{
		PluginId:   plugin.PluginId,
		Version:    release.ReleaseVersion,
		PluginType: marketv1.MarketplacePluginType(release.PluginType),
	}
	if normalizeSourceKind(plugin.SourceKind) == gitSourceKind {
		item.Mode = marketv1.MarketplaceDistributionModeGit
		item.RepoUrl = firstNonEmpty(plugin.RepoUrl, plugin.Repository)
		item.Ref = firstNonEmpty(release.SourceRef, release.ReleaseVersion)
		item.Provider = marketv1.MarketplaceRepoProvider(plugin.RepoProvider)
		item.RequiresAuth = normalizeKey(plugin.CredentialRef) != ""
		return item, nil
	}
	item.Mode = marketv1.MarketplaceDistributionModeHTTPS
	item.DownloadSessionRequired = true
	artifact, err := s.primaryArtifactForRelease(ctx, release)
	if err != nil {
		return nil, err
	}
	if artifact != nil {
		item.ArtifactType = marketv1.MarketplaceArtifactType(artifact.ArtifactType)
		item.Sha256 = artifact.Sha256
		item.SizeBytes = artifact.SizeBytes
	}
	return item, nil
}

func (s *serviceImpl) primaryArtifactForRelease(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
) (*entity.PluginMarketplaceArtifact, error) {
	if release == nil {
		return nil, nil
	}
	types := []string{
		marketv1.MarketplaceArtifactTypeSourceZip.String(),
		marketv1.MarketplaceArtifactTypeSourceTarGz.String(),
		marketv1.MarketplaceArtifactTypeDynamicZip.String(),
		marketv1.MarketplaceArtifactTypeDynamicTarGz.String(),
		marketv1.MarketplaceArtifactTypePluginWasm.String(),
	}
	var artifact *entity.PluginMarketplaceArtifact
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Where(do.PluginMarketplaceArtifact{ReleaseId: release.Id}).
		WhereIn(dao.PluginMarketplaceArtifact.Columns().ArtifactType, types).
		OrderAsc(dao.PluginMarketplaceArtifact.Columns().Id).
		Limit(1).
		Scan(&artifact); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return artifact, nil
}

func (s *serviceImpl) updateGitSyncStatus(ctx context.Context, pluginRecordID int, status string, message string) error {
	now := time.Now()
	if len(message) > 1024 {
		message = message[:1024]
	}
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: pluginRecordID}).
		Data(do.PluginMarketplacePlugin{
			LastSyncAt:      &now,
			LastSyncStatus:  status,
			LastSyncMessage: message,
		}).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

func (s *serviceImpl) getPluginEntityByRecordID(ctx context.Context, id int) (*entity.PluginMarketplacePlugin, error) {
	var plugin *entity.PluginMarketplacePlugin
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: id}).
		Scan(&plugin); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	return plugin, nil
}

func (s *serviceImpl) gitClient() gitRemoteClient {
	if s.gitRemote != nil {
		return s.gitRemote
	}
	return &httpGitRemoteClient{httpClient: &http.Client{Timeout: defaultGitHTTPTimeout}}
}

// requireUploadSourcePlugin ensures upload package mutations stay on upload plugins.
func (s *serviceImpl) requireUploadSourcePlugin(
	ctx context.Context,
	pluginID string,
	ownerUserID int64,
) (*entity.PluginMarketplacePlugin, error) {
	plugin, err := s.requirePluginForPublisher(ctx, "", pluginID, ownerUserID)
	if err != nil {
		return nil, err
	}
	kind := normalizeSourceKind(plugin.SourceKind)
	if kind == "" {
		kind = uploadSourceKind
	}
	if kind != uploadSourceKind {
		return nil, bizerr.NewCode(CodeMarketplaceSourceKindConflict)
	}
	return plugin, nil
}

func parseGitRepoURL(raw string) (gitRepoRef, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return gitRepoRef{}, bizerr.NewCode(CodeMarketplaceInvalidInput, bizerr.P("diagnostic", "repository URL is required"))
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return gitRepoRef{}, bizerr.NewCode(CodeMarketplaceInvalidInput, bizerr.P("diagnostic", "repository URL is invalid"))
	}
	host := strings.ToLower(parsed.Hostname())
	provider := marketv1.MarketplaceRepoProvider("")
	apiHost := ""
	switch host {
	case "github.com", "www.github.com":
		provider = marketv1.MarketplaceRepoProviderGitHub
		apiHost = "https://api.github.com"
	case "gitee.com", "www.gitee.com":
		provider = marketv1.MarketplaceRepoProviderGitee
		apiHost = "https://gitee.com/api/v5"
	default:
		return gitRepoRef{}, bizerr.NewCode(CodeMarketplaceInvalidInput, bizerr.P("diagnostic", "only github.com and gitee.com repositories are supported"))
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return gitRepoRef{}, bizerr.NewCode(CodeMarketplaceInvalidInput, bizerr.P("diagnostic", "repository URL must include owner and name"))
	}
	return gitRepoRef{
		Provider: provider,
		Owner:    parts[0],
		Name:     parts[1],
		CloneURL: "https://" + host + "/" + parts[0] + "/" + parts[1] + ".git",
		APIHost:  apiHost,
	}, nil
}

type gitPluginManifest struct {
	ID           string                       `json:"id" yaml:"id"`
	Name         string                       `json:"name" yaml:"name"`
	Version      string                       `json:"version" yaml:"version"`
	Type         string                       `json:"type" yaml:"type"`
	Description  string                       `json:"description,omitempty" yaml:"description"`
	Homepage     string                       `json:"homepage,omitempty" yaml:"homepage"`
	License      string                       `json:"license,omitempty" yaml:"license"`
	Dependencies *sourcePackageDependencySpec `json:"dependencies,omitempty" yaml:"dependencies"`
}

func parseGitPluginManifest(raw []byte) (*gitPluginManifest, error) {
	manifest := &gitPluginManifest{}
	if err := yaml.Unmarshal(raw, manifest); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml cannot be parsed")
	}
	manifest.ID = normalizeKey(manifest.ID)
	manifest.Name = normalizeKey(manifest.Name)
	manifest.Version = normalizeKey(manifest.Version)
	manifest.Type = strings.ToLower(normalizeKey(manifest.Type))
	return manifest, nil
}

func validateGitSourceManifest(manifest *gitPluginManifest) error {
	if manifest == nil {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml is required")
	}
	if !gitPluginIDPattern.MatchString(manifest.ID) {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml id is invalid")
	}
	if manifest.Version == "" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml version is required")
	}
	if manifest.Type != "" && manifest.Type != string(marketv1.MarketplacePluginTypeSource) {
		return bizerr.NewCode(CodeMarketplaceGitDynamicUnsupported, bizerr.P("diagnostic", "Git sources support source plugins only; publish dynamic plugins via upload packages"))
	}
	return nil
}

func validateGitRemoteStructure(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	token string,
) error {
	for _, required := range []string{gitRequiredPathBackendGo, gitRequiredPathEmbedGo} {
		ok, err := client.PathExists(ctx, clientRepo(repo), ref, required, token)
		if err != nil {
			return err
		}
		if !ok {
			return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "remote repository is missing "+required)
		}
	}
	return nil
}

func clientRepo(repo gitRepoRef) gitRepoRef { return repo }

func manifestAsSource(manifest *gitPluginManifest) *sourcePackageManifest {
	if manifest == nil {
		return &sourcePackageManifest{}
	}
	return &sourcePackageManifest{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		Description:  manifest.Description,
		Homepage:     manifest.Homepage,
		License:      manifest.License,
		Dependencies: manifest.Dependencies,
	}
}

func sourcePackageHostBoundsFromManifest(manifest *gitPluginManifest) string {
	if manifest == nil || manifest.Dependencies == nil || manifest.Dependencies.Framework == nil {
		return ""
	}
	return normalizeKey(manifest.Dependencies.Framework.Version)
}

func versionsSemanticallyEqual(tag string, version string) bool {
	return normalizeVersionLabel(tag) == normalizeVersionLabel(version)
}

func normalizeVersionLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "v") || strings.HasPrefix(trimmed, "V") {
		return "v" + strings.TrimPrefix(strings.TrimPrefix(trimmed, "v"), "V")
	}
	return "v" + trimmed
}

func normalizeSourceKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case gitSourceKind:
		return gitSourceKind
	case uploadSourceKind, "":
		return uploadSourceKind
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if normalizeKey(value) != "" {
			return normalizeKey(value)
		}
	}
	return ""
}

func releaseDistributionAllowed(
	release *entity.PluginMarketplaceRelease,
	owned bool,
	subject VisibilitySubject,
) bool {
	if release == nil {
		return false
	}
	if marketplaceManagementReadAllowed(owned, subject) {
		return true
	}
	return marketv1.MarketplaceStatus(release.ReleaseStatus) == marketv1.MarketplaceStatusPublished
}

func (c *httpGitRemoteClient) ListTags(ctx context.Context, repo gitRepoRef, token string) ([]string, error) {
	var endpoint string
	switch repo.Provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		endpoint = fmt.Sprintf("%s/repos/%s/%s/tags?per_page=%d", repo.APIHost, url.PathEscape(repo.Owner), url.PathEscape(repo.Name), maxGitTagsPerDiscovery)
	case marketv1.MarketplaceRepoProviderGitee:
		endpoint = fmt.Sprintf("%s/repos/%s/%s/tags", repo.APIHost, url.PathEscape(repo.Owner), url.PathEscape(repo.Name))
	default:
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	body, status, err := c.doGET(ctx, endpoint, token, repo.Provider)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, gitAuthError("repository authentication failed")
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("list tags failed with status %d", status)
	}
	var payload []struct {
		Name string `json:"name"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("list tags response is invalid")
	}
	tags := make([]string, 0, len(payload))
	for _, item := range payload {
		if name := normalizeKey(item.Name); name != "" {
			tags = append(tags, name)
		}
	}
	return tags, nil
}

func (c *httpGitRemoteClient) ReadFile(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) ([]byte, error) {
	var endpoint string
	switch repo.Provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		endpoint = fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/%s/%s",
			url.PathEscape(repo.Owner),
			url.PathEscape(repo.Name),
			url.PathEscape(ref),
			path.Clean(filePath),
		)
	case marketv1.MarketplaceRepoProviderGitee:
		endpoint = fmt.Sprintf(
			"https://gitee.com/%s/%s/raw/%s/%s",
			url.PathEscape(repo.Owner),
			url.PathEscape(repo.Name),
			url.PathEscape(ref),
			path.Clean(filePath),
		)
	default:
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	body, status, err := c.doGET(ctx, endpoint, token, repo.Provider)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, gitAuthError("repository authentication failed")
	}
	if status == http.StatusNotFound {
		return nil, packageDiagnosticError(CodeMarketplacePackageStructureInvalid, filePath+" not found at ref "+ref)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("read file failed with status %d", status)
	}
	return body, nil
}

func (c *httpGitRemoteClient) PathExists(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) (bool, error) {
	_, err := c.ReadFile(ctx, repo, ref, filePath, token)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	// packageDiagnosticError wraps bizerr - check message
	if isMarketplaceStructureMissing(err) {
		return false, nil
	}
	return false, err
}

func (c *httpGitRemoteClient) doGET(
	ctx context.Context,
	endpoint string,
	token string,
	provider marketv1.MarketplaceRepoProvider,
) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", gitMetadataUserAgent)
	req.Header.Set("Accept", "application/json")
	if token = strings.TrimSpace(token); token != "" {
		switch provider {
		case marketv1.MarketplaceRepoProviderGitHub:
			req.Header.Set("Authorization", "Bearer "+token)
		case marketv1.MarketplaceRepoProviderGitee:
			// Gitee commonly accepts access_token query; also send header for private raw.
			q := req.URL.Query()
			q.Set("access_token", token)
			req.URL.RawQuery = q.Encode()
		}
	}
	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: defaultGitHTTPTimeout}
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 2*1024*1024))
	if err != nil {
		return nil, res.StatusCode, err
	}
	return body, res.StatusCode, nil
}

type gitAuthErrorType struct{ message string }

func (e gitAuthErrorType) Error() string { return e.message }

func gitAuthError(message string) error { return gitAuthErrorType{message: message} }

func isGitAuthError(err error) bool {
	_, ok := err.(gitAuthErrorType)
	return ok
}

func mapGitClientError(err error) error {
	if err == nil {
		return nil
	}
	if isGitAuthError(err) {
		return bizerr.NewCode(CodeMarketplaceGitAuthFailed, bizerr.P("diagnostic", err.Error()))
	}
	if isMarketplaceBusinessError(err) {
		return err
	}
	return bizerr.NewCode(CodeMarketplaceGitDiscoveryFailed, bizerr.P("diagnostic", err.Error()))
}

func isMarketplaceStructureMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "missing")
}

func isMarketplaceBusinessError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Business errors from this package carry stable codes in message or are package diagnostics.
	for _, marker := range []string{
		"PLUGIN_MARKETPLACE_",
		"Marketplace package",
		"Marketplace Git",
		"Marketplace request input is invalid",
		"Marketplace plugin publish source kind",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

func gitPublicErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 512 {
		return msg[:512]
	}
	return msg
}
