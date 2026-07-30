// This file implements GitHub/Gitee metadata discovery for marketplace Git
// sources. It stores repository coordinates and version drafts only; full
// source trees are never cloned onto the marketplace server.
//
// Discovery rules:
//   - Prefer semver version tags when present.
//   - When no semver tags exist, fall back to the main branch.
//   - Fail only when neither semver tags nor main are available.
//   - Auto-detect single-plugin (repo root) vs multi-plugin (nested plugin roots).

package marketplace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
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
	gitSourceKind           = "git"
	uploadSourceKind        = "upload"
	gitSyncStatusSuccess    = "success"
	gitSyncStatusFailed     = "failed"
	gitSyncStatusAuthFailed = "auth_failed"
	gitSyncStatusPartial    = "partial"
	defaultGitHTTPTimeout   = 20 * time.Second
	maxGitTagsPerDiscovery  = 100
	maxGitPluginsPerRepo    = 50
	maxGitPluginPathDepth   = 6
	// gitManifestFetchWorkers bounds concurrent plugin.yaml reads for monorepos.
	// official-plugins has dozens of roots; sequential reads exceed browser timeouts.
	gitManifestFetchWorkers = 8
	// maxGitDocsIndexed caps remote documentation reads during Git discovery.
	maxGitDocsIndexed = 20
	// maxGitDocBytes rejects oversized remote markdown files during indexing.
	maxGitDocBytes = 256 * 1024
	// maxGitI18nFilesIndexed caps remote runtime i18n JSON reads used for display name/summary.
	maxGitI18nFilesIndexed = 40
	// maxGitI18nFileBytes rejects oversized remote i18n JSON during display catalog loads.
	maxGitI18nFileBytes       = 256 * 1024
	gitMetadataUserAgent      = "LinaPro-Plugin-Marketplace/1.0"
	gitRequiredPathPluginYAML = "plugin.yaml"
	gitRequiredPathBackendGo  = "backend/plugin.go"
	gitRequiredPathEmbedGo    = "plugin_embed.go"
	gitFallbackBranchMain     = "main"
	gitDiscoveryRefKindTag    = "tag"
	gitDiscoveryRefKindBranch = "branch"
	// configKeyGitHubAccessToken is the plugin-scoped config key for a platform
	// GitHub personal access token used when registration omits accessToken.
	configKeyGitHubAccessToken = "github.accessToken"
	// configKeyGiteeAccessToken is the plugin-scoped config key for a platform
	// Gitee personal access token used when registration omits accessToken.
	configKeyGiteeAccessToken = "gitee.accessToken"
)

var (
	gitSemverTagPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)
	gitPluginIDPattern  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	gitIgnoredPathRoots = map[string]struct{}{
		".git":         {},
		"node_modules": {},
		"vendor":       {},
		"dist":         {},
		"build":        {},
		"temp":         {},
		"tmp":          {},
	}
)

// gitRepoRef is a normalized GitHub/Gitee repository coordinate.
type gitRepoRef struct {
	Provider marketv1.MarketplaceRepoProvider
	Owner    string
	Name     string
	CloneURL string
	APIHost  string
}

// gitDiscoveryRef is one candidate tag or branch used for metadata discovery.
type gitDiscoveryRef struct {
	Name string
	Kind string
}

// gitPluginRoot is one discovered source-plugin root inside a repository.
type gitPluginRoot struct {
	Path     string
	Manifest *gitPluginManifest
}

// gitRemoteClient reads tags, trees, and raw files from Git hosting APIs.
type gitRemoteClient interface {
	ListTags(ctx context.Context, repo gitRepoRef, token string) ([]string, error)
	RefExists(ctx context.Context, repo gitRepoRef, ref string, token string) (bool, error)
	// ResolveCommitSHA resolves a branch, tag, or other ref to a full commit SHA.
	ResolveCommitSHA(ctx context.Context, repo gitRepoRef, ref string, token string) (string, error)
	ListTreePaths(ctx context.Context, repo gitRepoRef, ref string, token string) ([]string, error)
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

// RegisterGitSource registers a Git repository, auto-detects plugin roots, and
// enqueues plugins in pending_verify for async discovery/verify/review processing.
// Version discovery is intentionally not completed on the request path so the
// plugin appears in My Plugins immediately with a clear pipeline status.
func (s *serviceImpl) RegisterGitSource(ctx context.Context, in RegisterGitSourceInput) (*RegisterGitSourceResult, error) {
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
	// Only publisher-supplied tokens are encrypted into credential rows.
	// Platform config tokens are shared fallbacks and must not be stored per user.
	userToken := strings.TrimSpace(in.AccessToken)
	token, err := s.resolveGitAccessToken(ctx, repo.Provider, userToken)
	if err != nil {
		return nil, err
	}
	client := s.gitClient()

	// Minimal bootstrap discovery only: identify plugin roots so each plugin can
	// enter My Plugins as pending_verify. Full tag/version import runs async.
	discoveryRefs, err := s.resolveGitDiscoveryRefs(ctx, client, repo, token)
	if err != nil {
		return nil, err
	}
	bootstrapRef := discoveryRefs[0]
	roots, err := s.discoverSourcePluginRoots(ctx, client, repo, bootstrapRef.Name, token)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, bizerr.NewCode(CodeMarketplaceGitDiscoveryFailed, bizerr.P("diagnostic", "repository contains no valid source plugin roots"))
	}

	credentialRef := ""
	if userToken != "" {
		credentialRef, err = s.saveGitCredential(ctx, in.OwnerUserID, repo.Provider.String(), userToken)
		if err != nil {
			return nil, err
		}
	}

	records := make([]*PluginRecord, 0, len(roots))
	for _, root := range roots {
		plugin, upsertErr := s.upsertGitPluginFromRoot(ctx, publisher, repo, root, credentialRef, in)
		if upsertErr != nil {
			return nil, upsertErr
		}
		if setErr := s.setPluginProcessStatus(
			ctx,
			plugin.Id,
			marketv1.MarketplaceProcessStatusPendingVerify,
			"queued for async verification",
		); setErr != nil {
			return nil, setErr
		}
		record, getErr := s.getPluginRecordByID(ctx, plugin.Id)
		if getErr != nil {
			return nil, getErr
		}
		if record != nil {
			records = append(records, record)
		}
	}
	return &RegisterGitSourceResult{Plugins: records}, nil
}

// DiscoverGitMetadata refreshes tags or the main-branch fallback for one Git-backed plugin.
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
	storedToken, err := s.loadGitCredentialToken(ctx, plugin.CredentialRef)
	if err != nil {
		return nil, err
	}
	repo, err := parseGitRepoURL(plugin.RepoUrl)
	if err != nil {
		return nil, err
	}
	token, err := s.resolveGitAccessToken(ctx, repo.Provider, storedToken)
	if err != nil {
		return nil, err
	}
	return s.discoverGitMetadataForPlugin(ctx, plugin, token, s.gitClient())
}

// DiscoverAllGitSources scans every Git-backed plugin for new tags or main updates.
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
		storedToken, err := s.loadGitCredentialToken(ctx, row.CredentialRef)
		if err != nil {
			_ = s.updateGitSyncStatus(ctx, row.Id, gitSyncStatusFailed, "credential load failed")
			continue
		}
		repo, parseErr := parseGitRepoURL(row.RepoUrl)
		if parseErr != nil {
			_ = s.updateGitSyncStatus(ctx, row.Id, gitSyncStatusFailed, "repository url is invalid")
			continue
		}
		token, tokenErr := s.resolveGitAccessToken(ctx, repo.Provider, storedToken)
		if tokenErr != nil {
			_ = s.updateGitSyncStatus(ctx, row.Id, gitSyncStatusFailed, "platform git token config failed")
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

// resolveGitAccessToken prefers a publisher-supplied or stored credential, then
// falls back to the plugin-scoped platform token for the hosting provider.
func (s *serviceImpl) resolveGitAccessToken(
	ctx context.Context,
	provider marketv1.MarketplaceRepoProvider,
	preferred string,
) (string, error) {
	preferred = strings.TrimSpace(preferred)
	if preferred != "" {
		return preferred, nil
	}
	if s == nil || s.pluginConfig == nil {
		return "", nil
	}
	key := platformGitAccessTokenConfigKey(provider)
	if key == "" {
		return "", nil
	}
	token, err := s.pluginConfig.String(ctx, key, "")
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceGitDiscoveryFailed, bizerr.P("diagnostic", "read platform git access token config failed"))
	}
	return strings.TrimSpace(token), nil
}

// platformGitAccessTokenConfigKey returns the plugin config key for one host.
func platformGitAccessTokenConfigKey(provider marketv1.MarketplaceRepoProvider) string {
	switch provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		return configKeyGitHubAccessToken
	case marketv1.MarketplaceRepoProviderGitee:
		return configKeyGiteeAccessToken
	default:
		return ""
	}
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

func (s *serviceImpl) resolveGitDiscoveryRefs(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	token string,
) ([]gitDiscoveryRef, error) {
	tags, err := client.ListTags(ctx, repo, token)
	if err != nil {
		return nil, mapGitClientError(err)
	}
	semverTags := filterSemverTags(tags)
	if len(semverTags) > 0 {
		refs := make([]gitDiscoveryRef, 0, len(semverTags))
		for _, tag := range semverTags {
			refs = append(refs, gitDiscoveryRef{Name: tag, Kind: gitDiscoveryRefKindTag})
		}
		return refs, nil
	}
	exists, err := client.RefExists(ctx, repo, gitFallbackBranchMain, token)
	if err != nil {
		return nil, mapGitClientError(err)
	}
	if !exists {
		return nil, bizerr.NewCode(
			CodeMarketplaceGitDiscoveryFailed,
			bizerr.P("diagnostic", "repository has no version tags and main branch does not exist"),
		)
	}
	return []gitDiscoveryRef{{Name: gitFallbackBranchMain, Kind: gitDiscoveryRefKindBranch}}, nil
}

func (s *serviceImpl) discoverSourcePluginRoots(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	token string,
) ([]gitPluginRoot, error) {
	// Always list the remote tree first so monorepos can validate structure from the
	// path set without N extra PathExists round-trips (each ~1s against GitHub).
	paths, err := client.ListTreePaths(ctx, repo, ref, token)
	if err != nil {
		return nil, mapGitClientError(err)
	}
	pathSet := buildGitPathSet(paths)

	// Single-plugin: repository root is a valid source plugin root.
	if gitPathSetHasSourceStructure(pathSet, "") {
		root, ok, rootErr := s.trySourcePluginRoot(ctx, client, repo, ref, "", token, true)
		if rootErr != nil {
			return nil, rootErr
		}
		if ok {
			return []gitPluginRoot{root}, nil
		}
	}

	// Multi-plugin: filter nested candidates with the tree path set, then read
	// plugin.yaml concurrently so large official monorepos finish within the
	// browser request timeout budget.
	candidates := candidatePluginRootsFromTree(paths)
	filtered := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if gitPathSetHasSourceStructure(pathSet, candidate) {
			filtered = append(filtered, candidate)
		}
	}
	if len(filtered) > maxGitPluginsPerRepo {
		filtered = filtered[:maxGitPluginsPerRepo]
	}
	roots, loadErr := s.loadSourcePluginRootsParallel(ctx, client, repo, ref, token, filtered)
	if loadErr != nil {
		return nil, loadErr
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Path < roots[j].Path
	})
	return roots, nil
}

// loadSourcePluginRootsParallel reads plugin.yaml for candidate roots concurrently.
func (s *serviceImpl) loadSourcePluginRootsParallel(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	token string,
	candidates []string,
) ([]gitPluginRoot, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	workerCount := gitManifestFetchWorkers
	if workerCount > len(candidates) {
		workerCount = len(candidates)
	}

	type loadResult struct {
		root gitPluginRoot
		ok   bool
		err  error
	}
	jobs := make(chan string)
	results := make(chan loadResult, len(candidates))
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for candidate := range jobs {
				if ctx.Err() != nil {
					results <- loadResult{err: ctx.Err()}
					return
				}
				root, ok, rootErr := s.trySourcePluginRoot(ctx, client, repo, ref, candidate, token, true)
				results <- loadResult{root: root, ok: ok, err: rootErr}
			}
		}()
	}

	go func() {
		for _, candidate := range candidates {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- candidate:
			}
		}
		close(jobs)
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	roots := make([]gitPluginRoot, 0, len(candidates))
	seenIDs := make(map[string]struct{}, len(candidates))
	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		if !result.ok || result.root.Manifest == nil {
			continue
		}
		if _, exists := seenIDs[result.root.Manifest.ID]; exists {
			continue
		}
		seenIDs[result.root.Manifest.ID] = struct{}{}
		roots = append(roots, result.root)
	}
	if firstErr != nil && len(roots) == 0 {
		return nil, firstErr
	}
	return roots, nil
}

func (s *serviceImpl) trySourcePluginRoot(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	repoPath string,
	token string,
	structureAlreadyChecked bool,
) (gitPluginRoot, bool, error) {
	manifestBytes, err := client.ReadFile(ctx, repo, ref, gitPathJoin(repoPath, gitRequiredPathPluginYAML), token)
	if err != nil {
		if isMarketplaceStructureMissing(err) {
			return gitPluginRoot{}, false, nil
		}
		return gitPluginRoot{}, false, mapGitClientError(err)
	}
	manifest, err := parseGitPluginManifest(manifestBytes)
	if err != nil {
		return gitPluginRoot{}, false, nil
	}
	if err = validateGitSourceManifest(manifest); err != nil {
		// Dynamic plugins and invalid manifests are skipped during multi-root scan.
		return gitPluginRoot{}, false, nil
	}
	if !structureAlreadyChecked {
		if err = validateGitRemoteStructure(ctx, client, repo, ref, repoPath, token); err != nil {
			return gitPluginRoot{}, false, nil
		}
	}
	return gitPluginRoot{Path: normalizeGitRepoPath(repoPath), Manifest: manifest}, true, nil
}

func (s *serviceImpl) upsertGitPluginFromRoot(
	ctx context.Context,
	publisher *entity.PluginMarketplacePublisher,
	repo gitRepoRef,
	root gitPluginRoot,
	credentialRef string,
	in RegisterGitSourceInput,
) (*entity.PluginMarketplacePlugin, error) {
	if root.Manifest == nil {
		return nil, bizerr.NewCode(CodeMarketplacePackageInvalid)
	}
	manifest := root.Manifest
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
	repoPath := normalizeGitRepoPath(root.Path)
	gitVisibility := marketv1.MarketplaceVisibilityPrivate
	// plugin.yaml already carries version at bootstrap. Persist it on the plugin
	// row so My Plugins can show a version immediately, without waiting for the
	// async release-draft discovery job.
	discoveredVersion := normalizeVersionLabel(manifest.Version)

	if existing == nil {
		id, insertErr := dao.PluginMarketplacePlugin.Ctx(ctx).Data(do.PluginMarketplacePlugin{
			PublisherId:     publisher.Id,
			PluginId:        manifest.ID,
			Name:            name,
			Summary:         summary,
			Description:     normalizeKey(manifest.Description),
			PluginType:      marketv1.MarketplacePluginTypeSource.String(),
			MarketStatus:    marketv1.MarketplaceStatusDraft.String(),
			ProcessStatus:   marketv1.MarketplaceProcessStatusPendingVerify.String(),
			Visibility:      gitVisibility.String(),
			LatestReleaseId: 0,
			LatestVersion:   discoveredVersion,
			Homepage:        firstNonEmpty(in.Homepage, manifest.Homepage),
			Repository:      repo.CloneURL,
			License:         firstNonEmpty(in.License, manifest.License),
			DownloadCount:   0,
			SourceKind:      gitSourceKind,
			RepoUrl:         repo.CloneURL,
			RepoProvider:    repo.Provider.String(),
			RepoPath:        repoPath,
			CredentialRef:   credentialRef,
		}).InsertAndGetId()
		if insertErr != nil {
			return nil, bizerr.WrapCode(insertErr, CodeMarketplaceStorageFailed)
		}
		return s.getPluginEntityByRecordID(ctx, intID(id))
	}

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
		RepoPath:     repoPath,
	}
	if marketv1.MarketplaceStatus(existing.MarketStatus) == marketv1.MarketplaceStatusDraft {
		update.Visibility = gitVisibility.String()
		if discoveredVersion != "" {
			update.LatestVersion = discoveredVersion
		}
	} else if normalizeKey(existing.LatestVersion) == "" && discoveredVersion != "" {
		update.LatestVersion = discoveredVersion
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
	return s.getPluginEntityByRecordID(ctx, existing.Id)
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
	refs, err := s.resolveGitDiscoveryRefs(ctx, client, repo, token)
	if err != nil {
		status := gitSyncStatusFailed
		if isGitAuthError(err) {
			status = gitSyncStatusAuthFailed
		}
		_ = s.updateGitSyncStatus(ctx, plugin.Id, status, gitPublicErrorMessage(err))
		return nil, err
	}

	// Load remote tree once for resource/docs enrichment across version refs.
	// Structure is stable enough across tags for path discovery; file content is
	// still read per ref when indexing documentation.
	treePaths, treeErr := client.ListTreePaths(ctx, repo, refs[0].Name, token)
	if treeErr != nil {
		// Tree failure should not block plugin.yaml-only version import.
		treePaths = nil
	}

	synced := 0
	failures := 0
	skippedImmutable := 0
	var firstFailure string
	for _, ref := range refs {
		// Enrich docs for every discovered ref. Remote body reads stay bounded by
		// maxGitDocsIndexed inside indexGitReleaseDocuments.
		outcome, discoverErr := s.discoverOneGitRef(ctx, plugin, repo, ref, token, client, treePaths, true)
		if discoverErr != nil {
			failures++
			if firstFailure == "" {
				firstFailure = gitPublicErrorMessage(discoverErr)
			}
			continue
		}
		switch outcome {
		case gitDiscoverRefCreatedOrUpdated:
			synced++
		case gitDiscoverRefImmutable:
			// Existing submitted/published versions are a successful no-op.
			// Count them so re-registration is not reported as "discovered 0".
			skippedImmutable++
		}
	}

	status := gitSyncStatusSuccess
	message := fmt.Sprintf("discovered %d draft releases", synced)
	if failures > 0 && synced == 0 {
		status = gitSyncStatusFailed
		message = fmt.Sprintf("failed to import %d refs", failures)
		if firstFailure != "" {
			message = message + ": " + firstFailure
		}
	} else if failures > 0 {
		status = gitSyncStatusPartial
		message = fmt.Sprintf("discovered %d drafts with %d ref failures", synced, failures)
		if firstFailure != "" {
			message = message + ": " + firstFailure
		}
	} else if synced == 0 && skippedImmutable > 0 {
		// Existing immutable versions are a successful no-op discovery, not a failure.
		message = fmt.Sprintf(
			"discovered 0 new draft releases (%d existing immutable version(s))",
			skippedImmutable,
		)
	}
	if err = s.updateGitSyncStatus(ctx, plugin.Id, status, message); err != nil {
		return nil, err
	}
	// New drafts re-enter the async verify path so they can auto-submit for review.
	if synced > 0 {
		processStatus := marketv1.MarketplaceProcessStatus(normalizeProcessStatus(plugin.ProcessStatus))
		if processStatus == marketv1.MarketplaceProcessStatusCompleted ||
			processStatus == marketv1.MarketplaceProcessStatusFailed ||
			processStatus == marketv1.MarketplaceProcessStatusPendingReview {
			if setErr := s.setPluginProcessStatus(
				ctx,
				plugin.Id,
				marketv1.MarketplaceProcessStatusPendingVerify,
				"new git draft queued for verification",
			); setErr != nil {
				return nil, setErr
			}
			_ = s.setLatestMutableReleaseProcessStatus(
				ctx,
				plugin.PluginId,
				marketv1.MarketplaceProcessStatusPendingVerify,
			)
		}
	} else if skippedImmutable > 0 {
		// Re-register / re-sync with only immutable versions: restore the plugin
		// process status from the existing submitted or published release so the
		// owner UI does not stay at failed/pending_verify with zero drafts.
		if _, healErr := s.healProcessStatusFromExistingReleases(ctx, plugin); healErr != nil {
			return nil, healErr
		}
	}
	record, err := s.getPluginRecordByID(ctx, plugin.Id)
	if err != nil {
		return nil, err
	}
	return &DiscoverGitMetadataResult{Plugin: record, Synced: synced}, nil
}

// gitDiscoverRefOutcome classifies one ref import attempt for sync diagnostics.
type gitDiscoverRefOutcome int

const (
	// gitDiscoverRefNone means no draft was written and the ref was not an
	// immutable skip (unexpected empty result; treat as no-op).
	gitDiscoverRefNone gitDiscoverRefOutcome = iota
	// gitDiscoverRefCreatedOrUpdated means a mutable draft was inserted or replaced.
	gitDiscoverRefCreatedOrUpdated
	// gitDiscoverRefImmutable means the version already exists as submitted/published.
	gitDiscoverRefImmutable
)

func (s *serviceImpl) discoverOneGitRef(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	repo gitRepoRef,
	ref gitDiscoveryRef,
	token string,
	client gitRemoteClient,
	treePaths []string,
	enrichDocs bool,
) (outcome gitDiscoverRefOutcome, err error) {
	repoPath := normalizeGitRepoPath(plugin.RepoPath)
	manifestBytes, err := client.ReadFile(ctx, repo, ref.Name, gitPathJoin(repoPath, gitRequiredPathPluginYAML), token)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	manifest, err := parseGitPluginManifest(manifestBytes)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	if err = validateGitSourceManifest(manifest); err != nil {
		return gitDiscoverRefNone, err
	}
	if normalizeKey(manifest.ID) != normalizeKey(plugin.PluginId) {
		return gitDiscoverRefNone, packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml id does not match marketplace plugin id")
	}
	if ref.Kind == gitDiscoveryRefKindTag && !versionsSemanticallyEqual(ref.Name, manifest.Version) {
		return gitDiscoverRefNone, bizerr.NewCode(
			CodeMarketplaceGitVersionMismatch,
			bizerr.P("diagnostic", "tag "+ref.Name+" does not match plugin.yaml version "+manifest.Version),
		)
	}
	// Prefer tree path set when available to avoid extra PathExists round-trips.
	if len(treePaths) > 0 {
		if !gitPathSetHasSourceStructure(buildGitPathSet(treePaths), repoPath) {
			return gitDiscoverRefNone, packageDiagnosticError(
				CodeMarketplacePackageStructureInvalid,
				"remote repository is missing required source plugin files",
			)
		}
	} else if err = validateGitRemoteStructure(ctx, client, repo, ref.Name, repoPath, token); err != nil {
		return gitDiscoverRefNone, err
	}

	// Pin the resolved commit so installs do not float with branch tips.
	sourceCommit, err := client.ResolveCommitSHA(ctx, repo, ref.Name, token)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	sourceCommit = normalizeKey(sourceCommit)
	if sourceCommit == "" {
		return gitDiscoverRefNone, packageDiagnosticError(
			CodeMarketplaceGitDiscoveryFailed,
			"failed to resolve commit SHA for ref "+ref.Name,
		)
	}

	versionLabel := normalizeVersionLabel(manifest.Version)
	existing, err := s.getReleaseByPluginVersion(ctx, plugin.PluginId, versionLabel)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	if existing != nil && immutableRelease(existing) {
		// Submitted/published releases must not rewrite draft metadata, but
		// documentation snapshots are disk-only and may be incomplete after a
		// monorepo import, storage wipe, or when manifest/docs was added after
		// the install pin was frozen. Re-index docs against the current
		// discovery ref (tag/main) so marketplace catalogs pick up all Markdown
		// files and first-heading titles, while install still uses source_commit.
		if enrichDocs {
			release, releaseErr := s.getReleaseRecordByID(ctx, existing.Id)
			if releaseErr != nil {
				return gitDiscoverRefNone, releaseErr
			}
			docsRef := strings.TrimSpace(ref.Name)
			if docsRef == "" {
				docsRef = firstNonEmpty(normalizeKey(existing.SourceCommit), gitFallbackBranchMain)
			}
			if docsErr := s.indexGitReleaseDocuments(ctx, client, repo, docsRef, token, repoPath, release, treePaths); docsErr != nil {
				_ = s.updateGitSyncStatus(
					ctx,
					plugin.Id,
					gitSyncStatusPartial,
					"immutable release kept; documentation indexing incomplete: "+gitPublicErrorMessage(docsErr),
				)
			}
		}
		return gitDiscoverRefImmutable, nil
	}

	snapshot, err := packageJSONString(manifest)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	deps, err := packageJSONString(buildSourceDependencySummary(manifestAsSource(manifest)))
	if err != nil {
		return gitDiscoverRefNone, err
	}
	sqlSummary, i18nSummary, docsSummary := buildGitResourceSummariesFromTree(treePaths, repoPath)
	sqlJSON, err := packageJSONString(sqlSummary)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	i18nJSON, err := packageJSONString(i18nSummary)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	docsJSON, err := packageJSONString(docsSummary)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	diagnostics := sourcePackageDiagnostics(manifestAsSource(manifest), sqlSummary, i18nSummary, docsSummary)
	riskJSON, err := packageJSONString(buildSourceRiskSummary(diagnostics))
	if err != nil {
		return gitDiscoverRefNone, err
	}

	reviewMessage := "Imported from Git tag " + ref.Name
	if ref.Kind == gitDiscoveryRefKindBranch {
		reviewMessage = "Imported from Git branch " + ref.Name
	}
	in := SaveReleaseDraftInput{
		PublisherKey:      "",
		OwnerUserID:       0,
		PluginID:          plugin.PluginId,
		Version:           versionLabel,
		PluginType:        marketv1.MarketplacePluginTypeSource,
		Visibility:        marketv1.MarketplaceVisibility(plugin.Visibility),
		MinHostVersion:    sourcePackageHostBoundsFromManifest(manifest),
		ManifestSnapshot:  snapshot,
		DependencySummary: deps,
		SQLSummary:        sqlJSON,
		I18NSummary:       i18nJSON,
		DocsSummary:       docsJSON,
		RiskSummary:       riskJSON,
		ReviewMessage:     reviewMessage,
		ReplaceDraft:      true,
		SourceRef:         ref.Name,
		SourceCommit:      sourceCommit,
	}

	release, err := s.saveGitReleaseDraft(ctx, plugin, in)
	if err != nil {
		return gitDiscoverRefNone, err
	}
	if release == nil {
		return gitDiscoverRefNone, nil
	}

	// Keep publisher-facing plugin identity aligned with the discovered plugin.yaml.
	if applyErr := s.applyGitManifestToPlugin(ctx, plugin, manifest, versionLabel, release.ID); applyErr != nil {
		return gitDiscoverRefNone, applyErr
	}
	// Persist display name/summary per locale. Base row comes from plugin.yaml;
	// remote runtime i18n catalogs (manifest/i18n/<locale>/*.json, non-apidoc)
	// override with plugin.<id>.name / plugin.<id>.description when present.
	defaultLocale := defaultDisplayLocale
	if manifest.I18N != nil {
		defaultLocale = defaultLocaleFromManifest(manifest.I18N.Default)
	}
	displayItems := buildDisplayI18nFromPackageYAML(
		plugin.PluginId,
		firstNonEmpty(manifest.Name, plugin.Name, plugin.PluginId),
		firstNonEmpty(manifest.Description, plugin.Summary),
		defaultLocale,
	)
	localeCatalogs := loadGitDisplayI18nCatalogs(ctx, client, repo, ref.Name, token, repoPath, treePaths)
	displayItems = mergePackageI18nDisplayItems(plugin.PluginId, displayItems, localeCatalogs)
	if displayErr := s.replaceReleaseDisplayI18n(ctx, release, displayItems); displayErr != nil {
		return gitDiscoverRefNone, displayErr
	}

	if enrichDocs {
		if docsErr := s.indexGitReleaseDocuments(ctx, client, repo, ref.Name, token, repoPath, release, treePaths); docsErr != nil {
			// Documentation enrichment is best-effort; release metadata still stands.
			_ = s.updateGitSyncStatus(ctx, plugin.Id, gitSyncStatusPartial, "release imported; documentation indexing incomplete: "+gitPublicErrorMessage(docsErr))
		}
	}
	return gitDiscoverRefCreatedOrUpdated, nil
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
	data.SourceCommit = normalizeKey(in.SourceCommit)
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

// applyGitManifestToPlugin updates owner-visible plugin identity fields and draft
// latest-version anchors from the discovered plugin.yaml.
func (s *serviceImpl) applyGitManifestToPlugin(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	manifest *gitPluginManifest,
	versionLabel string,
	releaseID int,
) error {
	if plugin == nil || manifest == nil {
		return nil
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
	data := do.PluginMarketplacePlugin{
		Name:        name,
		Summary:     summary,
		Description: normalizeKey(manifest.Description),
		Homepage:    firstNonEmpty(manifest.Homepage, plugin.Homepage),
		License:     firstNonEmpty(manifest.License, plugin.License),
	}
	// For unpublished plugins, keep list/detail version anchors on the draft.
	if marketv1.MarketplaceStatus(plugin.MarketStatus) == marketv1.MarketplaceStatusDraft {
		data.LatestVersion = normalizeKey(versionLabel)
		if releaseID > 0 {
			data.LatestReleaseId = releaseID
		}
	} else if normalizeKey(plugin.LatestVersion) == "" {
		data.LatestVersion = normalizeKey(versionLabel)
	}
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
		Data(data).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	// Keep the in-memory plugin projection current for subsequent refs.
	plugin.Name = name
	plugin.Summary = summary
	plugin.Description = normalizeKey(manifest.Description)
	if marketv1.MarketplaceStatus(plugin.MarketStatus) == marketv1.MarketplaceStatusDraft {
		plugin.LatestVersion = normalizeKey(versionLabel)
		if releaseID > 0 {
			plugin.LatestReleaseId = releaseID
		}
	}
	return nil
}

// buildGitResourceSummariesFromTree derives SQL/i18n/docs path summaries from one
// remote tree listing without downloading file bodies.
func buildGitResourceSummariesFromTree(
	treePaths []string,
	repoPath string,
) (
	sqlItems []*sourcePackageResourceSummary,
	i18nItems []*sourcePackageI18NSummary,
	docsItems []*sourcePackageResourceSummary,
) {
	repoPath = normalizeGitRepoPath(repoPath)
	sqlItems = make([]*sourcePackageResourceSummary, 0)
	docsItems = make([]*sourcePackageResourceSummary, 0)
	i18nByLocale := make(map[string]*sourcePackageI18NSummary)

	for _, raw := range treePaths {
		rel, ok := gitRelativePluginPath(raw, repoPath)
		if !ok {
			continue
		}
		lower := strings.ToLower(rel)
		switch {
		case strings.HasPrefix(lower, "manifest/sql/") && strings.HasSuffix(lower, ".sql"):
			sqlItems = append(sqlItems, &sourcePackageResourceSummary{
				Kind: "sql",
				Path: rel,
			})
		case strings.HasPrefix(lower, "manifest/docs/") && strings.HasSuffix(lower, ".md"):
			docsItems = append(docsItems, &sourcePackageResourceSummary{
				Kind: "marketplace_doc",
				Path: rel,
			})
		case strings.HasPrefix(lower, "manifest/i18n/"):
			locale, kind := classifyGitI18NPath(rel)
			if locale == "" {
				continue
			}
			item := i18nByLocale[locale]
			if item == nil {
				item = &sourcePackageI18NSummary{Locale: locale}
				i18nByLocale[locale] = item
			}
			switch kind {
			case "apidoc":
				item.APIDocFiles = append(item.APIDocFiles, rel)
			default:
				item.RuntimeFiles = append(item.RuntimeFiles, rel)
			}
		case lower == "readme.md" || lower == "readme.zh-cn.md":
			docsItems = append(docsItems, &sourcePackageResourceSummary{
				Kind: "readme",
				Path: rel,
			})
		}
	}

	i18nItems = make([]*sourcePackageI18NSummary, 0, len(i18nByLocale))
	for _, item := range i18nByLocale {
		i18nItems = append(i18nItems, item)
	}
	sort.Slice(sqlItems, func(i, j int) bool { return sqlItems[i].Path < sqlItems[j].Path })
	sort.Slice(docsItems, func(i, j int) bool { return docsItems[i].Path < docsItems[j].Path })
	sort.Slice(i18nItems, func(i, j int) bool { return i18nItems[i].Locale < i18nItems[j].Locale })
	return sqlItems, i18nItems, docsItems
}

func gitRelativePluginPath(treePath string, repoPath string) (string, bool) {
	normalized := strings.Trim(strings.ReplaceAll(treePath, "\\", "/"), "/")
	if normalized == "" {
		return "", false
	}
	if repoPath == "" {
		return normalized, true
	}
	prefix := repoPath + "/"
	if !strings.HasPrefix(normalized, prefix) {
		return "", false
	}
	return strings.TrimPrefix(normalized, prefix), true
}

func classifyGitI18NPath(relPath string) (locale string, kind string) {
	// Expected shapes:
	//   manifest/i18n/<locale>/plugin.json
	//   manifest/i18n/<locale>/apidoc/*.json
	parts := strings.Split(strings.Trim(relPath, "/"), "/")
	if len(parts) < 4 || parts[0] != "manifest" || parts[1] != "i18n" {
		return "", ""
	}
	locale = parts[2]
	if locale == "" {
		return "", ""
	}
	if len(parts) >= 5 && parts[3] == "apidoc" {
		return locale, "apidoc"
	}
	return locale, "runtime"
}

// selectGitRuntimeI18nPaths returns relative runtime i18n JSON paths under the
// plugin root. API doc catalogs under apidoc/ are excluded because they do not
// carry display name/summary keys.
func selectGitRuntimeI18nPaths(treePaths []string, repoPath string) []string {
	out := make([]string, 0)
	for _, raw := range treePaths {
		rel, ok := gitRelativePluginPath(raw, repoPath)
		if !ok {
			continue
		}
		if strings.ToLower(path.Ext(rel)) != ".json" {
			continue
		}
		locale, kind := classifyGitI18NPath(rel)
		if locale == "" || kind != "runtime" {
			continue
		}
		out = append(out, rel)
	}
	sort.Strings(out)
	if len(out) > maxGitI18nFilesIndexed {
		out = out[:maxGitI18nFilesIndexed]
	}
	return out
}

// loadGitDisplayI18nCatalogs reads remote runtime i18n JSON files and builds
// locale catalogs for display name/summary projection. Failed or oversized
// reads are skipped so plugin.yaml fallback rows still persist.
func loadGitDisplayI18nCatalogs(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	token string,
	repoPath string,
	treePaths []string,
) map[string]map[string]string {
	out := make(map[string]map[string]string)
	if client == nil {
		return out
	}
	candidates := selectGitRuntimeI18nPaths(treePaths, repoPath)
	for _, relPath := range candidates {
		body, readErr := client.ReadFile(ctx, repo, ref, gitPathJoin(repoPath, relPath), token)
		if readErr != nil {
			continue
		}
		if len(body) == 0 || len(body) > maxGitI18nFileBytes {
			continue
		}
		locale, _ := classifyGitI18NPath(relPath)
		normalizedLocale := normalizeDisplayLocale(locale)
		if normalizedLocale == "" {
			continue
		}
		catalog, parseErr := parseFlatI18nJSON(body)
		if parseErr != nil || len(catalog) == 0 {
			continue
		}
		merged := out[normalizedLocale]
		if merged == nil {
			merged = make(map[string]string)
			out[normalizedLocale] = merged
		}
		for key, value := range catalog {
			merged[key] = value
		}
	}
	return out
}

// indexGitReleaseDocuments reads remote Markdown docs under the plugin root and
// writes a disk snapshot under ArtifactStore for later language-aware reads.
// Tree paths are refreshed from the same content ref whenever possible so
// candidates always match blobs that ReadFile can resolve.
func (s *serviceImpl) indexGitReleaseDocuments(
	ctx context.Context,
	client gitRemoteClient,
	repo gitRepoRef,
	ref string,
	token string,
	repoPath string,
	release *ReleaseRecord,
	treePaths []string,
) error {
	if release == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	// Prefer a live tree at the content ref. Callers may pass a shared monorepo
	// tree from another ref; using a mismatched tree would either miss docs or
	// list candidates that fail to read and collapse the catalog to README-only.
	if client != nil && strings.TrimSpace(ref) != "" {
		if liveTree, listErr := client.ListTreePaths(ctx, repo, ref, token); listErr == nil && len(liveTree) > 0 {
			treePaths = liveTree
		}
	}
	candidates := selectGitDocPathsForIndexing(treePaths, repoPath)
	if len(candidates) == 0 {
		return nil
	}
	items := make([]*marketplaceDocumentIndexItem, 0, len(candidates))
	rawBodies := make(map[string][]byte, len(candidates))
	var firstIndexFailure string
	for _, relPath := range candidates {
		body, readErr := client.ReadFile(ctx, repo, ref, gitPathJoin(repoPath, relPath), token)
		if readErr != nil {
			if firstIndexFailure == "" {
				firstIndexFailure = gitPublicErrorMessage(readErr)
			}
			continue
		}
		if len(body) == 0 || len(body) > maxGitDocBytes {
			if firstIndexFailure == "" {
				firstIndexFailure = "documentation body empty or exceeds size limit"
			}
			continue
		}
		sourceKind, locale, docPath := resolveGitDocumentIdentity(relPath)
		item, indexErr := indexMarketplaceDocument(locale, docPath, sourceKind, string(body))
		if indexErr != nil {
			if firstIndexFailure == "" {
				firstIndexFailure = gitPublicErrorMessage(indexErr)
			}
			continue
		}
		items = append(items, item)
		rawBodies[documentIndexKey(item.Locale, item.DocPath)] = body
	}
	if len(items) == 0 {
		// Surface total indexing failure so callers can mark partial sync instead
		// of treating a docs-bearing tree as a successful empty snapshot.
		if len(candidates) > 0 {
			diagnostic := "documentation candidates were found but none could be indexed"
			if firstIndexFailure != "" {
				diagnostic = diagnostic + ": " + firstIndexFailure
			}
			return bizerr.NewCode(CodeMarketplaceGitDiscoveryFailed, bizerr.P("diagnostic", diagnostic))
		}
		return nil
	}
	return s.replaceReleaseGitDocumentSnapshot(ctx, release, items, rawBodies)
}

// resolveGitDocumentIdentity maps one repo-relative markdown path to marketplace
// document identity fields used when indexing Git release documentation.
// README case labels must be lowercased because the switch compares
// strings.ToLower(relPath); English README locale matches ZIP indexing (en-US).
func resolveGitDocumentIdentity(relPath string) (sourceKind string, locale string, docPath string) {
	sourceKind = documentSourceKindManifestDocs
	locale = defaultDocumentLocale
	docPath = relPath
	switch strings.ToLower(relPath) {
	case "readme.md":
		return documentSourceKindReadme, fallbackEnUSLocale, readmeDocumentPath
	case "readme.zh-cn.md":
		return documentSourceKindReadme, fallbackZhCNLocale, readmeCNDocumentPath
	default:
		// manifest/docs/<locale>/... or manifest/docs/*.md
		if strings.HasPrefix(relPath, marketplaceDocsPrefix) {
			docPath = strings.TrimPrefix(relPath, marketplaceDocsPrefix)
			parts := strings.Split(docPath, "/")
			if len(parts) >= 2 {
				// Prefer first directory segment as locale when it looks like zh-CN/en-US.
				if strings.Contains(parts[0], "-") || parts[0] == "zh" || parts[0] == "en" {
					locale = parts[0]
					docPath = strings.Join(parts[1:], "/")
				}
			}
		}
		return sourceKind, locale, docPath
	}
}

func selectGitDocPathsForIndexing(treePaths []string, repoPath string) []string {
	_, _, docs := buildGitResourceSummariesFromTree(treePaths, repoPath)
	if len(docs) == 0 {
		return nil
	}
	// Prefer marketplace manifest/docs over package-root README so the index
	// budget is spent on navigable docs. README is still indexed when budget
	// remains and becomes the docs entry only when no manifest docs exist.
	priority := func(pathValue string) int {
		lower := strings.ToLower(pathValue)
		switch {
		case strings.HasPrefix(lower, "manifest/docs/") &&
			(strings.HasSuffix(lower, "/index.md") || lower == "manifest/docs/index.md"):
			return 0
		case strings.HasPrefix(lower, "manifest/docs/") && strings.HasSuffix(lower, ".md"):
			return 1
		case lower == "readme.zh-cn.md" || lower == "readme.md" || strings.Contains(lower, "readme"):
			return 3
		default:
			return 2
		}
	}
	sort.SliceStable(docs, func(i, j int) bool {
		pi, pj := priority(docs[i].Path), priority(docs[j].Path)
		if pi != pj {
			return pi < pj
		}
		return docs[i].Path < docs[j].Path
	})
	limit := maxGitDocsIndexed
	if limit > len(docs) {
		limit = len(docs)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, docs[i].Path)
	}
	return out
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
		// Prefer the pinned commit so installs reproduce the discovered content
		// even when source_ref is a floating branch such as main.
		item.Ref = firstNonEmpty(release.SourceCommit, release.SourceRef, release.ReleaseVersion)
		item.Path = normalizeGitRepoPath(plugin.RepoPath)
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
	var (
		provider marketv1.MarketplaceRepoProvider
		apiHost  string
	)
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
	I18N         *sourcePackageI18N           `json:"i18n,omitempty" yaml:"i18n"`
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
	repoPath string,
	token string,
) error {
	for _, required := range []string{gitRequiredPathBackendGo, gitRequiredPathEmbedGo} {
		ok, err := client.PathExists(ctx, repo, ref, gitPathJoin(repoPath, required), token)
		if err != nil {
			return err
		}
		if !ok {
			return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "remote repository is missing "+gitPathJoin(repoPath, required))
		}
	}
	return nil
}

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

func filterSemverTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		name := normalizeKey(tag)
		if name == "" {
			continue
		}
		if gitSemverTagPattern.MatchString(name) {
			out = append(out, name)
		}
	}
	return out
}

func normalizeGitRepoPath(value string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" || trimmed == "." {
		return ""
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func gitPathJoin(repoPath string, filePath string) string {
	repoPath = normalizeGitRepoPath(repoPath)
	filePath = strings.Trim(strings.ReplaceAll(filePath, "\\", "/"), "/")
	if repoPath == "" {
		return filePath
	}
	if filePath == "" {
		return repoPath
	}
	return path.Join(repoPath, filePath)
}

// buildGitPathSet indexes remote blob paths for O(1) structure checks.
func buildGitPathSet(paths []string) map[string]struct{} {
	pathSet := make(map[string]struct{}, len(paths))
	for _, raw := range paths {
		filePath := strings.Trim(strings.ReplaceAll(raw, "\\", "/"), "/")
		if filePath == "" {
			continue
		}
		pathSet[filePath] = struct{}{}
	}
	return pathSet
}

// gitPathSetHasSourceStructure reports whether the remote tree already contains
// the minimal source-plugin files for one plugin root.
func gitPathSetHasSourceStructure(pathSet map[string]struct{}, repoPath string) bool {
	if pathSet == nil {
		return false
	}
	for _, required := range []string{
		gitRequiredPathPluginYAML,
		gitRequiredPathBackendGo,
		gitRequiredPathEmbedGo,
	} {
		if _, ok := pathSet[gitPathJoin(repoPath, required)]; !ok {
			return false
		}
	}
	return true
}

// candidatePluginRootsFromTree returns directories that contain plugin.yaml.
// Root plugin.yaml is excluded because single-plugin roots are handled first.
func candidatePluginRootsFromTree(paths []string) []string {
	candidates := make([]string, 0)
	seen := make(map[string]struct{})
	for _, raw := range paths {
		filePath := strings.Trim(strings.ReplaceAll(raw, "\\", "/"), "/")
		if filePath == "" || filePath == gitRequiredPathPluginYAML {
			continue
		}
		if !strings.HasSuffix(filePath, "/"+gitRequiredPathPluginYAML) {
			continue
		}
		dir := strings.TrimSuffix(filePath, "/"+gitRequiredPathPluginYAML)
		dir = normalizeGitRepoPath(dir)
		if dir == "" {
			continue
		}
		if !isAllowedGitPluginPath(dir) {
			continue
		}
		if _, exists := seen[dir]; exists {
			continue
		}
		seen[dir] = struct{}{}
		candidates = append(candidates, dir)
	}
	sort.Strings(candidates)
	return candidates
}

func isAllowedGitPluginPath(repoPath string) bool {
	parts := strings.Split(normalizeGitRepoPath(repoPath), "/")
	if len(parts) == 0 || parts[0] == "" {
		return false
	}
	if len(parts) > maxGitPluginPathDepth {
		return false
	}
	for _, part := range parts {
		if part == ".." || part == "." {
			return false
		}
		if _, ignored := gitIgnoredPathRoots[part]; ignored {
			return false
		}
	}
	return true
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
	if err = gitHostingHTTPError(status, body, token, "list tags"); err != nil {
		return nil, err
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

func (c *httpGitRemoteClient) RefExists(ctx context.Context, repo gitRepoRef, ref string, token string) (bool, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false, nil
	}
	var endpoint string
	switch repo.Provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		endpoint = fmt.Sprintf("%s/repos/%s/%s/branches/%s", repo.APIHost, url.PathEscape(repo.Owner), url.PathEscape(repo.Name), url.PathEscape(ref))
	case marketv1.MarketplaceRepoProviderGitee:
		endpoint = fmt.Sprintf("%s/repos/%s/%s/branches/%s", repo.APIHost, url.PathEscape(repo.Owner), url.PathEscape(repo.Name), url.PathEscape(ref))
	default:
		return false, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	body, status, err := c.doGET(ctx, endpoint, token, repo.Provider)
	if err != nil {
		return false, err
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if err = gitHostingHTTPError(status, body, token, "check branch"); err != nil {
		return false, err
	}
	return true, nil
}

// ResolveCommitSHA resolves a branch, tag, or other ref to a full commit SHA.
func (c *httpGitRemoteClient) ResolveCommitSHA(ctx context.Context, repo gitRepoRef, ref string, token string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("ref is required")
	}
	// Both GitHub and Gitee expose /commits/{ref} that accepts branch, tag, or SHA.
	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/commits/%s",
		repo.APIHost,
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
		url.PathEscape(ref),
	)
	body, status, err := c.doGET(ctx, endpoint, token, repo.Provider)
	if err != nil {
		return "", err
	}
	if err = gitHostingHTTPError(status, body, token, "resolve commit"); err != nil {
		return "", err
	}
	var payload struct {
		SHA string `json:"sha"`
	}
	if err = json.Unmarshal(body, &payload); err != nil || strings.TrimSpace(payload.SHA) == "" {
		return "", fmt.Errorf("commit response is invalid")
	}
	return strings.TrimSpace(payload.SHA), nil
}

func (c *httpGitRemoteClient) ListTreePaths(ctx context.Context, repo gitRepoRef, ref string, token string) ([]string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	switch repo.Provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		return c.listGitHubTreePaths(ctx, repo, ref, token)
	case marketv1.MarketplaceRepoProviderGitee:
		return c.listGiteeTreePaths(ctx, repo, ref, token)
	default:
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
}

func (c *httpGitRemoteClient) listGitHubTreePaths(ctx context.Context, repo gitRepoRef, ref string, token string) ([]string, error) {
	commitEndpoint := fmt.Sprintf(
		"%s/repos/%s/%s/commits/%s",
		repo.APIHost,
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
		url.PathEscape(ref),
	)
	commitBody, status, err := c.doGET(ctx, commitEndpoint, token, repo.Provider)
	if err != nil {
		return nil, err
	}
	if err = gitHostingHTTPError(status, commitBody, token, "read commit"); err != nil {
		return nil, err
	}
	var commit struct {
		Commit struct {
			Tree struct {
				SHA string `json:"sha"`
			} `json:"tree"`
		} `json:"commit"`
	}
	if err = json.Unmarshal(commitBody, &commit); err != nil || commit.Commit.Tree.SHA == "" {
		return nil, fmt.Errorf("commit response is invalid")
	}
	treeEndpoint := fmt.Sprintf(
		"%s/repos/%s/%s/git/trees/%s?recursive=1",
		repo.APIHost,
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
		url.PathEscape(commit.Commit.Tree.SHA),
	)
	treeBody, status, err := c.doGET(ctx, treeEndpoint, token, repo.Provider)
	if err != nil {
		return nil, err
	}
	if err = gitHostingHTTPError(status, treeBody, token, "list tree"); err != nil {
		return nil, err
	}
	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err = json.Unmarshal(treeBody, &tree); err != nil {
		return nil, fmt.Errorf("tree response is invalid")
	}
	paths := make([]string, 0, len(tree.Tree))
	for _, item := range tree.Tree {
		if item.Type != "blob" || item.Path == "" {
			continue
		}
		paths = append(paths, item.Path)
	}
	return paths, nil
}

func (c *httpGitRemoteClient) listGiteeTreePaths(ctx context.Context, repo gitRepoRef, ref string, token string) ([]string, error) {
	// Gitee recursive tree API uses the git/trees endpoint with the branch/tag name.
	endpoint := fmt.Sprintf(
		"%s/repos/%s/%s/git/trees/%s?recursive=1",
		repo.APIHost,
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
		url.PathEscape(ref),
	)
	body, status, err := c.doGET(ctx, endpoint, token, repo.Provider)
	if err != nil {
		return nil, err
	}
	if err = gitHostingHTTPError(status, body, token, "list tree"); err != nil {
		return nil, err
	}
	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err = json.Unmarshal(body, &tree); err != nil {
		return nil, fmt.Errorf("tree response is invalid")
	}
	paths := make([]string, 0, len(tree.Tree))
	for _, item := range tree.Tree {
		if item.Type != "blob" || item.Path == "" {
			continue
		}
		paths = append(paths, item.Path)
	}
	return paths, nil
}

func (c *httpGitRemoteClient) ReadFile(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) ([]byte, error) {
	var endpoint string
	switch repo.Provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		// Prefer Contents API over raw.githubusercontent.com so Authorization
		// tokens apply consistently for private repositories and rate limits.
		endpoint = fmt.Sprintf(
			"%s/repos/%s/%s/contents/%s?ref=%s",
			repo.APIHost,
			url.PathEscape(repo.Owner),
			url.PathEscape(repo.Name),
			encodeGitContentPath(filePath),
			url.QueryEscape(ref),
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
	if status == http.StatusNotFound {
		return nil, packageDiagnosticError(CodeMarketplacePackageStructureInvalid, filePath+" not found at ref "+ref)
	}
	if err = gitHostingHTTPError(status, body, token, "read file"); err != nil {
		return nil, err
	}
	if repo.Provider == marketv1.MarketplaceRepoProviderGitHub {
		return decodeGitHubContentsFile(body)
	}
	return body, nil
}

// encodeGitContentPath encodes each path segment for GitHub Contents API.
func encodeGitContentPath(filePath string) string {
	cleaned := path.Clean(strings.Trim(strings.ReplaceAll(filePath, "\\", "/"), "/"))
	if cleaned == "." || cleaned == "" {
		return ""
	}
	parts := strings.Split(cleaned, "/")
	encoded := make([]string, 0, len(parts))
	for _, part := range parts {
		encoded = append(encoded, url.PathEscape(part))
	}
	return strings.Join(encoded, "/")
}

// decodeGitHubContentsFile extracts file bytes from a GitHub Contents API response.
func decodeGitHubContentsFile(body []byte) ([]byte, error) {
	var payload struct {
		Encoding string `json:"encoding"`
		Content  string `json:"content"`
		Message  string `json:"message"`
		Type     string `json:"type"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		// Some gateways may still proxy raw body; accept plain content.
		if len(body) > 0 && body[0] != '{' {
			return body, nil
		}
		return nil, fmt.Errorf("github contents response is invalid")
	}
	if payload.Type != "" && payload.Type != "file" {
		return nil, fmt.Errorf("github contents path is not a file")
	}
	if strings.EqualFold(payload.Encoding, "base64") {
		// GitHub inserts newlines in base64 payloads.
		cleaned := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, payload.Content)
		decoded, err := decodeBase64Std(cleaned)
		if err != nil {
			return nil, fmt.Errorf("github contents base64 decode failed")
		}
		return decoded, nil
	}
	if payload.Content != "" {
		return []byte(payload.Content), nil
	}
	if payload.Message != "" {
		return nil, fmt.Errorf("github contents error: %s", payload.Message)
	}
	return nil, fmt.Errorf("github contents response is empty")
}

func (c *httpGitRemoteClient) PathExists(ctx context.Context, repo gitRepoRef, ref string, filePath string, token string) (bool, error) {
	_, err := c.ReadFile(ctx, repo, ref, filePath, token)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
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
	switch provider {
	case marketv1.MarketplaceRepoProviderGitHub:
		// GitHub recommends the versioned media type for REST API stability.
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-Github-Api-Version", "2022-11-28")
	default:
		req.Header.Set("Accept", "application/json")
	}
	if token = strings.TrimSpace(token); token != "" {
		switch provider {
		case marketv1.MarketplaceRepoProviderGitHub:
			req.Header.Set("Authorization", "Bearer "+token)
		case marketv1.MarketplaceRepoProviderGitee:
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

// gitHostingHTTPError maps hosting HTTP statuses to actionable errors.
// Important: GitHub unauthenticated rate limits return HTTP 403, which must NOT
// be reported as credential/authentication failure for public repositories.
func gitHostingHTTPError(status int, body []byte, token string, operation string) error {
	if status == http.StatusOK {
		return nil
	}
	message := strings.ToLower(string(body))
	switch status {
	case http.StatusUnauthorized:
		if strings.TrimSpace(token) == "" {
			return fmt.Errorf("%s failed with HTTP 401; repository may be private and requires an access token", operation)
		}
		return newGitAuthError("repository authentication failed: invalid or expired access token")
	case http.StatusForbidden:
		if strings.Contains(message, "rate limit") ||
			strings.Contains(message, "secondary rate limit") ||
			strings.Contains(message, "api rate limit exceeded") {
			if strings.TrimSpace(token) == "" {
				return fmt.Errorf("%s failed: GitHub API rate limit exceeded for unauthenticated requests; provide a personal access token (public_repo or fine-grained contents:read) and retry", operation)
			}
			return fmt.Errorf("%s failed: GitHub API rate limit exceeded for the provided token; wait and retry", operation)
		}
		if strings.TrimSpace(token) == "" {
			// Public repos usually do not 403 unless rate-limited or blocked by network policy.
			return fmt.Errorf("%s denied with HTTP 403; if this is a public repository the host is often rate-limited or blocked—provide a personal access token or retry later", operation)
		}
		return newGitAuthError("repository authentication failed: token lacks access to this repository")
	default:
		if status >= 200 && status < 300 {
			return nil
		}
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 180 {
			snippet = snippet[:180] + "..."
		}
		if snippet == "" {
			return fmt.Errorf("%s failed with status %d", operation, status)
		}
		return fmt.Errorf("%s failed with status %d: %s", operation, status, snippet)
	}
}

// gitAuthError is a typed auth failure from Git hosting APIs.
type gitAuthError struct{ message string }

func (e gitAuthError) Error() string { return e.message }

func newGitAuthError(message string) error { return gitAuthError{message: message} }

func isGitAuthError(err error) bool {
	var target gitAuthError
	return errors.As(err, &target)
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

// decodeBase64Std decodes standard base64 without padding edge cases.
func decodeBase64Std(value string) ([]byte, error) {
	// Prefer StdEncoding; fall back to RawStdEncoding for unpadded payloads.
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
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
