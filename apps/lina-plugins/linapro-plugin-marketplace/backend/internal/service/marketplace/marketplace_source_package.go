// This file implements source-plugin marketplace package upload handling. It
// validates ZIP container safety, source plugin directory structure, root
// plugin.yaml identity, SQL/i18n/docs/dependency summaries, draft release
// persistence, and source artifact checksum metadata.

package marketplace

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	sourcePackageDefaultContentType = "application/zip"
	sourcePackageStoragePrefix      = "source"
	sourcePackageManifestPath       = "plugin.yaml"
	sourcePackageGoModPath          = "go.mod"
	sourcePackageBackendEntryPath   = "backend/plugin.go"
	sourcePackageEmbedPath          = "plugin_embed.go"
)

var (
	sourcePackagePluginIDPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	sourcePackageSemverPattern       = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)
	sourcePackageVersionRangePattern = regexp.MustCompile(`^(>=|<=|>|<|=)?v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$`)
)

// UploadSourcePackage validates and persists one source-plugin marketplace draft.
func (s *serviceImpl) UploadSourcePackage(
	ctx context.Context,
	in UploadSourcePackageInput,
) (*SourcePackageUploadResult, error) {
	if in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	scan, err := scanSourcePackage(in)
	if err != nil {
		return nil, err
	}
	if plugin, getErr := s.getPluginByID(ctx, scan.manifest.ID); getErr == nil && plugin != nil {
		if normalizeSourceKind(plugin.SourceKind) == gitSourceKind {
			return nil, bizerr.NewCode(CodeMarketplaceSourceKindConflict)
		}
	}
	publisherKey, err := s.resolvePublisherKeyForPackageUpload(
		ctx,
		in.PublisherKey,
		in.OwnerUserID,
		scan.manifest,
		marketv1.MarketplacePluginTypeSource,
		in.Visibility,
		in.AutoCreate,
	)
	if err != nil {
		return nil, err
	}
	if err = s.storeUploadedPackage(ctx, scan.storageKey, in.PackagePath); err != nil {
		return nil, err
	}

	var (
		release  *ReleaseRecord
		artifact *ArtifactRecord
	)
	if err = dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		release, err = s.SaveReleaseDraft(ctx, SaveReleaseDraftInput{
			PublisherKey:      publisherKey,
			OwnerUserID:       in.OwnerUserID,
			PluginID:          scan.manifest.ID,
			Version:           scan.manifest.Version,
			PluginType:        marketv1.MarketplacePluginTypeSource,
			Visibility:        normalizeVisibility(in.Visibility),
			MinHostVersion:    scan.minHostVersion,
			MaxHostVersion:    scan.maxHostVersion,
			ManifestSnapshot:  scan.manifestSnapshot,
			DependencySummary: scan.dependencySummary,
			SQLSummary:        scan.sqlSummary,
			I18NSummary:       scan.i18nSummary,
			DocsSummary:       scan.docsSummary,
			RiskSummary:       scan.riskSummary,
			ReviewMessage:     sourcePackageReviewMessage(scan.diagnostics),
			ReplaceDraft:      in.ReplaceDraft,
		})
		if err != nil {
			return err
		}

		artifact, err = s.saveSourcePackageArtifact(ctx, release, scan)
		if err != nil {
			return err
		}
		if err = s.replaceReleaseDocuments(ctx, release, scan.docsIndex); err != nil {
			return err
		}
		if err = s.replaceReleaseRisks(ctx, release, scan.diagnostics); err != nil {
			return err
		}
		plugin, pluginErr := s.getPluginByID(ctx, scan.manifest.ID)
		if pluginErr != nil {
			return pluginErr
		}
		return s.touchPluginLatestDraft(ctx, plugin, release)
	}); err != nil {
		return nil, err
	}

	return &SourcePackageUploadResult{
		Release:     release,
		Artifact:    artifact,
		Diagnostics: scan.diagnostics,
	}, nil
}

// sourcePackageScan carries scanner output ready for release draft persistence.
type sourcePackageScan struct {
	manifest          *sourcePackageManifest
	packageSha256     string
	manifestSha256    string
	storageKey        string
	fileName          string
	contentType       string
	sizeBytes         int64
	minHostVersion    string
	maxHostVersion    string
	manifestSnapshot  string
	dependencySummary string
	sqlSummary        string
	i18nSummary       string
	docsSummary       string
	riskSummary       string
	docsIndex         []*marketplaceDocumentIndexItem
	diagnostics       []*PackageDiagnostic
}

// sourcePackageManifest is the marketplace scanner's local plugin.yaml subset.
type sourcePackageManifest struct {
	ID                  string                       `json:"id" yaml:"id"`
	Name                string                       `json:"name" yaml:"name"`
	Version             string                       `json:"version" yaml:"version"`
	Type                string                       `json:"type" yaml:"type"`
	Distribution        string                       `json:"distribution" yaml:"distribution"`
	ScopeNature         string                       `json:"scopeNature" yaml:"scope_nature"`
	SupportsMultiTenant *bool                        `json:"supportsMultiTenant,omitempty" yaml:"supports_multi_tenant"`
	DefaultInstallMode  string                       `json:"defaultInstallMode" yaml:"default_install_mode"`
	Description         string                       `json:"description,omitempty" yaml:"description"`
	Author              string                       `json:"author,omitempty" yaml:"author"`
	Homepage            string                       `json:"homepage,omitempty" yaml:"homepage"`
	License             string                       `json:"license,omitempty" yaml:"license"`
	I18N                *sourcePackageI18N           `json:"i18n,omitempty" yaml:"i18n"`
	Dependencies        *sourcePackageDependencySpec `json:"dependencies,omitempty" yaml:"dependencies"`
	HostServices        []*protocol.HostServiceSpec  `json:"hostServices,omitempty" yaml:"hostServices"`
}

// sourcePackageI18N is the plugin.yaml i18n subset used by marketplace scanning.
type sourcePackageI18N struct {
	Enabled bool                       `json:"enabled" yaml:"enabled"`
	Default string                     `json:"default,omitempty" yaml:"default"`
	Locales []*sourcePackageI18NLocale `json:"locales,omitempty" yaml:"locales"`
}

// sourcePackageI18NLocale is one declared plugin locale.
type sourcePackageI18NLocale struct {
	Locale     string `json:"locale" yaml:"locale"`
	NativeName string `json:"nativeName,omitempty" yaml:"nativeName"`
}

// sourcePackageDependencySpec is the plugin.yaml dependency subset.
type sourcePackageDependencySpec struct {
	Framework *sourcePackageFrameworkDependency `json:"framework,omitempty" yaml:"framework"`
	Plugins   []*sourcePackagePluginDependency  `json:"plugins,omitempty" yaml:"plugins"`
}

// sourcePackageFrameworkDependency declares the framework version range.
type sourcePackageFrameworkDependency struct {
	Version string `json:"version,omitempty" yaml:"version"`
}

// sourcePackagePluginDependency declares one plugin dependency edge.
type sourcePackagePluginDependency struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version,omitempty" yaml:"version"`
}

// sourcePackageResourceSummary is a file-level scanner summary persisted as JSON.
type sourcePackageResourceSummary struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	SizeBytes uint64 `json:"sizeBytes"`
	Sha256    string `json:"sha256"`
}

// sourcePackageI18NSummary is a locale-level i18n scanner summary.
type sourcePackageI18NSummary struct {
	Locale           string   `json:"locale"`
	RuntimeFiles     []string `json:"runtimeFiles"`
	APIDocFiles      []string `json:"apidocFiles"`
	DeclaredInPlugin bool     `json:"declaredInPlugin"`
	DefaultLocale    bool     `json:"defaultLocale"`
}

// sourcePackageDependencySummary is a normalized dependency scanner summary.
type sourcePackageDependencySummary struct {
	Kind    string `json:"kind"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
}

// sourcePackageRiskSummary is the aggregated scanner risk count payload.
type sourcePackageRiskSummary struct {
	Info    int `json:"info"`
	Warning int `json:"warning"`
	High    int `json:"high"`
}

// scanSourcePackage parses and validates one uploaded source package.
func scanSourcePackage(in UploadSourcePackageInput) (scan *sourcePackageScan, err error) {
	packagePath := normalizeKey(in.PackagePath)
	if packagePath == "" {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package path is required")
	}
	if err = ensurePackageArchiveSupported(in.FileName); err != nil {
		return nil, err
	}

	packageSha, sizeBytes, err := fileSHA256AndSize(packagePath)
	if err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package file cannot be read")
	}

	scanPath, cleanup, err := materializeZipPackagePath(packagePath, in.FileName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	zipReader, err := zip.OpenReader(scanPath)
	if err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package must be a valid ZIP or tar.gz container")
	}
	defer func() {
		if closeErr := zipReader.Close(); err == nil && closeErr != nil {
			err = bizerr.WrapCode(closeErr, CodeMarketplacePackageScanFailed)
		}
	}()

	fileIndex, err := indexSourceZipFiles(zipReader.File)
	if err != nil {
		return nil, err
	}
	rootPrefix, err := detectSourcePackageRoot(fileIndex)
	if err != nil {
		return nil, err
	}
	if err = validateSourcePackageStructure(fileIndex, rootPrefix); err != nil {
		return nil, err
	}

	manifestFile := fileIndex[rootPrefix+sourcePackageManifestPath]
	manifestBytes, err := readZipFile(manifestFile)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
	}
	manifest := &sourcePackageManifest{}
	if err = yaml.Unmarshal(manifestBytes, manifest); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml cannot be parsed")
	}
	normalizeSourcePackageManifest(manifest)
	if err = validateSourcePackageManifest(manifest, in); err != nil {
		return nil, err
	}

	sqlResources, err := buildSourceResourceSummaries(fileIndex, rootPrefix, "manifest/sql/", "sql")
	if err != nil {
		return nil, err
	}
	i18nResources := buildSourceI18NSummary(manifest, fileIndex, rootPrefix)
	docsResources, err := buildSourceDocsSummary(fileIndex, rootPrefix)
	if err != nil {
		return nil, err
	}
	docsIndex, err := buildSourcePackageDocumentIndex(manifest, fileIndex, rootPrefix)
	if err != nil {
		return nil, err
	}
	diagnostics := sourcePackageDiagnostics(manifest, sqlResources, i18nResources, docsResources)
	manifestSnapshot, err := packageJSONString(manifest)
	if err != nil {
		return nil, err
	}
	dependencySummary, err := packageJSONString(buildSourceDependencySummary(manifest))
	if err != nil {
		return nil, err
	}
	sqlSummary, err := packageJSONString(sqlResources)
	if err != nil {
		return nil, err
	}
	i18nSummary, err := packageJSONString(i18nResources)
	if err != nil {
		return nil, err
	}
	docsSummary, err := packageJSONString(docsResources)
	if err != nil {
		return nil, err
	}
	riskSummary, err := packageJSONString(buildSourceRiskSummary(diagnostics))
	if err != nil {
		return nil, err
	}

	minHostVersion, maxHostVersion := sourcePackageHostBounds(manifest)
	if override := normalizeKey(in.MinHostVersion); override != "" {
		minHostVersion = override
	}
	if override := normalizeKey(in.MaxHostVersion); override != "" {
		maxHostVersion = override
	}

	fileName := normalizeKey(in.FileName)
	if fileName == "" {
		fileName = filepath.Base(packagePath)
	}
	contentType := packageContentTypeForName(fileName, normalizeKey(in.ContentType))
	if contentType == "" {
		contentType = sourcePackageDefaultContentType
	}
	storageKey := normalizeKey(in.StorageKey)
	if storageKey == "" {
		storageKey = path.Join(sourcePackageStoragePrefix, manifest.ID, manifest.Version, packageSha+storageKeyExtension(fileName))
	}

	return &sourcePackageScan{
		manifest:          manifest,
		packageSha256:     packageSha,
		manifestSha256:    sha256Hex(manifestBytes),
		storageKey:        storageKey,
		fileName:          fileName,
		contentType:       contentType,
		sizeBytes:         sizeBytes,
		minHostVersion:    minHostVersion,
		maxHostVersion:    maxHostVersion,
		manifestSnapshot:  manifestSnapshot,
		dependencySummary: dependencySummary,
		sqlSummary:        sqlSummary,
		i18nSummary:       i18nSummary,
		docsSummary:       docsSummary,
		riskSummary:       riskSummary,
		docsIndex:         docsIndex,
		diagnostics:       diagnostics,
	}, err
}

// saveSourcePackageArtifact creates or replaces source ZIP artifact metadata.
func (s *serviceImpl) saveSourcePackageArtifact(
	ctx context.Context,
	release *ReleaseRecord,
	scan *sourcePackageScan,
) (*ArtifactRecord, error) {
	if release == nil || scan == nil {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	artifactType := sourceArtifactTypeForName(scan.fileName).String()
	data := do.PluginMarketplaceArtifact{
		ReleaseId:      release.ID,
		PluginId:       release.PluginID,
		ReleaseVersion: release.Version,
		ArtifactType:   artifactType,
		StorageKey:     scan.storageKey,
		FileName:       scan.fileName,
		ContentType:    scan.contentType,
		SizeBytes:      scan.sizeBytes,
		Sha256:         scan.packageSha256,
		ManifestSha256: scan.manifestSha256,
		WasmSha256:     "",
	}

	existing, err := s.getArtifactByReleaseType(ctx, release.ID, artifactType)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		id, insertErr := dao.PluginMarketplaceArtifact.Ctx(ctx).Data(data).InsertAndGetId()
		if insertErr != nil {
			return nil, bizerr.WrapCode(insertErr, CodeMarketplaceStorageFailed)
		}
		return s.getArtifactRecordByID(ctx, intID(id))
	}
	if _, updateErr := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Where(do.PluginMarketplaceArtifact{Id: existing.Id}).
		Data(data).
		Update(); updateErr != nil {
		return nil, bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
	}
	return s.getArtifactRecordByID(ctx, existing.Id)
}

// getArtifactByReleaseType loads one artifact by release and artifact type.
func (s *serviceImpl) getArtifactByReleaseType(
	ctx context.Context,
	releaseID int,
	artifactType string,
) (*entity.PluginMarketplaceArtifact, error) {
	var artifact *entity.PluginMarketplaceArtifact
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Where(do.PluginMarketplaceArtifact{ReleaseId: releaseID, ArtifactType: artifactType}).
		Scan(&artifact); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return artifact, nil
}

// getArtifactRecordByID loads one artifact and projects it to a service record.
func (s *serviceImpl) getArtifactRecordByID(ctx context.Context, id int) (*ArtifactRecord, error) {
	var artifact *entity.PluginMarketplaceArtifact
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Where(do.PluginMarketplaceArtifact{Id: id}).
		Scan(&artifact); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if artifact == nil {
		return nil, bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	return artifactRecordFromEntity(artifact), nil
}

// fileSHA256AndSize returns the checksum and size of the uploaded package file.
func fileSHA256AndSize(filePath string) (checksum string, sizeBytes int64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	hasher := sha256.New()
	sizeBytes, err = io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), sizeBytes, nil
}

// indexSourceZipFiles validates ZIP entry paths and returns normalized file entries.
func indexSourceZipFiles(files []*zip.File) (map[string]*zip.File, error) {
	index := make(map[string]*zip.File, len(files))
	for _, file := range files {
		if file == nil || file.FileInfo().IsDir() {
			continue
		}
		normalized, skip, err := normalizeZipEntryName(file.Name)
		if err != nil {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, err.Error())
		}
		if skip {
			continue
		}
		if _, ok := index[normalized]; ok {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package contains duplicate file path: "+normalized)
		}
		index[normalized] = file
	}
	if len(index) == 0 {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package does not contain files")
	}
	return index, nil
}

// normalizeZipEntryName normalizes one ZIP path and rejects traversal or host paths.
func normalizeZipEntryName(value string) (normalized string, skip bool, err error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true, nil
	}
	if strings.Contains(trimmed, "\\") {
		return "", false, gerror.New("package paths must use forward slashes")
	}
	if strings.HasPrefix(trimmed, "/") {
		return "", false, gerror.New("package paths must be relative")
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." {
		return "", true, nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", false, gerror.New("package paths must not traverse parent directories")
	}
	if cleaned == ".DS_Store" || strings.HasPrefix(cleaned, "__MACOSX/") || strings.HasSuffix(cleaned, "/.DS_Store") {
		return "", true, nil
	}
	return cleaned, false, nil
}

// detectSourcePackageRoot supports either direct-root or single-directory source packages.
func detectSourcePackageRoot(files map[string]*zip.File) (string, error) {
	if _, ok := files[sourcePackageManifestPath]; ok {
		return "", nil
	}

	root := ""
	for filePath := range files {
		segments := strings.Split(filePath, "/")
		if len(segments) < 2 {
			return "", packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "plugin.yaml must be at package root or under one root directory")
		}
		if root == "" {
			root = segments[0]
			continue
		}
		if root != segments[0] {
			return "", packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "source package must contain a single plugin root directory")
		}
	}
	if _, ok := files[root+"/"+sourcePackageManifestPath]; !ok {
		return "", packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "source package root is missing plugin.yaml")
	}
	return root + "/", nil
}

// validateSourcePackageStructure checks required source plugin package entries.
func validateSourcePackageStructure(files map[string]*zip.File, rootPrefix string) error {
	requiredFiles := []string{
		sourcePackageManifestPath,
		sourcePackageGoModPath,
		sourcePackageBackendEntryPath,
		sourcePackageEmbedPath,
	}
	for _, required := range requiredFiles {
		if _, ok := files[rootPrefix+required]; !ok {
			return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "source package is missing "+required)
		}
	}

	requiredDirs := []string{"backend/", "frontend/", "manifest/"}
	for _, required := range requiredDirs {
		if !zipDirHasFile(files, rootPrefix+required) {
			return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "source package is missing "+strings.TrimSuffix(required, "/")+"/")
		}
	}
	if !zipDirHasMarkdown(files, rootPrefix+"manifest/docs/") {
		return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "source package is missing manifest/docs markdown entry")
	}
	return nil
}

// zipDirHasFile reports whether at least one file exists under prefix.
func zipDirHasFile(files map[string]*zip.File, prefix string) bool {
	for filePath := range files {
		if strings.HasPrefix(filePath, prefix) {
			return true
		}
	}
	return false
}

// zipDirHasMarkdown reports whether prefix contains at least one Markdown file.
func zipDirHasMarkdown(files map[string]*zip.File, prefix string) bool {
	for filePath := range files {
		if strings.HasPrefix(filePath, prefix) && strings.ToLower(filepath.Ext(filePath)) == ".md" {
			return true
		}
	}
	return false
}

// readZipFile reads one ZIP file entry and closes its reader.
func readZipFile(file *zip.File) (content []byte, err error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := reader.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	return io.ReadAll(reader)
}

// normalizeSourcePackageManifest trims and defaults local manifest fields.
func normalizeSourcePackageManifest(manifest *sourcePackageManifest) {
	if manifest == nil {
		return
	}
	manifest.ID = normalizeKey(manifest.ID)
	manifest.Name = normalizeKey(manifest.Name)
	manifest.Version = normalizeKey(manifest.Version)
	manifest.Type = strings.ToLower(normalizeKey(manifest.Type))
	if manifest.Type == "" {
		manifest.Type = marketv1.MarketplacePluginTypeSource.String()
	}
	manifest.Distribution = strings.ToLower(normalizeKey(manifest.Distribution))
	if manifest.Distribution == "" {
		manifest.Distribution = "managed"
	}
	manifest.ScopeNature = strings.ToLower(normalizeKey(manifest.ScopeNature))
	manifest.DefaultInstallMode = strings.ToLower(normalizeKey(manifest.DefaultInstallMode))
	manifest.Description = normalizeKey(manifest.Description)
	manifest.Author = normalizeKey(manifest.Author)
	manifest.Homepage = normalizeKey(manifest.Homepage)
	manifest.License = normalizeKey(manifest.License)
	if manifest.I18N != nil {
		manifest.I18N.Default = normalizeKey(manifest.I18N.Default)
		for _, locale := range manifest.I18N.Locales {
			if locale == nil {
				continue
			}
			locale.Locale = normalizeKey(locale.Locale)
			locale.NativeName = normalizeKey(locale.NativeName)
		}
	}
	normalizeSourcePackageDependencies(manifest.Dependencies)
}

// normalizeSourcePackageDependencies trims dependency fields in-place.
func normalizeSourcePackageDependencies(spec *sourcePackageDependencySpec) {
	if spec == nil {
		return
	}
	if spec.Framework != nil {
		spec.Framework.Version = normalizeKey(spec.Framework.Version)
	}
	for _, dependency := range spec.Plugins {
		if dependency == nil {
			continue
		}
		dependency.ID = normalizeKey(dependency.ID)
		dependency.Version = normalizeKey(dependency.Version)
	}
}

// validateSourcePackageManifest validates source package manifest identity and dependencies.
func validateSourcePackageManifest(manifest *sourcePackageManifest, in UploadSourcePackageInput) error {
	if manifest == nil {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml is empty")
	}
	if manifest.ID == "" || manifest.Name == "" || manifest.Version == "" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml must include id, name, and version")
	}
	if !sourcePackagePluginIDPattern.MatchString(manifest.ID) || len(manifest.ID) > 64 {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml id must use kebab-case lowercase letters and digits")
	}
	if !sourcePackageSemverPattern.MatchString(manifest.Version) {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml version must use semver format")
	}
	if manifest.Type != marketv1.MarketplacePluginTypeSource.String() {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml type must be source")
	}
	if manifest.Distribution != "managed" && manifest.Distribution != "builtin" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml distribution must be managed or builtin")
	}
	if normalizeKey(in.PluginID) != "" && manifest.ID != normalizeKey(in.PluginID) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml id does not match upload pluginId")
	}
	if normalizeKey(in.Version) != "" && manifest.Version != normalizeKey(in.Version) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml version does not match upload version")
	}
	return validateSourcePackageDependencies(manifest)
}

// validateSourcePackageDependencies validates framework and plugin dependency declarations.
func validateSourcePackageDependencies(manifest *sourcePackageManifest) error {
	if manifest.Dependencies == nil {
		return nil
	}
	if manifest.Dependencies.Framework != nil && manifest.Dependencies.Framework.Version != "" {
		if err := validateSourceVersionRange(manifest.Dependencies.Framework.Version); err != nil {
			return packageDiagnosticError(CodeMarketplacePackageInvalid, "framework dependency version is invalid")
		}
	}

	seen := make(map[string]struct{}, len(manifest.Dependencies.Plugins))
	for _, dependency := range manifest.Dependencies.Plugins {
		if dependency == nil || dependency.ID == "" {
			return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin dependency is missing id")
		}
		if !sourcePackagePluginIDPattern.MatchString(dependency.ID) || len(dependency.ID) > 64 {
			return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin dependency id is invalid")
		}
		if dependency.ID == manifest.ID {
			return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin cannot depend on itself")
		}
		if _, ok := seen[dependency.ID]; ok {
			return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin declares duplicate dependency: "+dependency.ID)
		}
		seen[dependency.ID] = struct{}{}
		if dependency.Version != "" {
			if err := validateSourceVersionRange(dependency.Version); err != nil {
				return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin dependency version is invalid: "+dependency.ID)
			}
		}
	}
	return nil
}

// validateSourceVersionRange validates a whitespace-separated semver range.
func validateSourceVersionRange(value string) error {
	tokens := strings.Fields(normalizeKey(value))
	if len(tokens) == 0 {
		return gerror.New("version range cannot be empty")
	}
	for _, token := range tokens {
		if !sourcePackageVersionRangePattern.MatchString(token) {
			return gerror.New("version range token must use semver comparison format")
		}
	}
	return nil
}

// buildSourceDependencySummary builds normalized dependency summary entries.
func buildSourceDependencySummary(manifest *sourcePackageManifest) []*sourcePackageDependencySummary {
	if manifest == nil || manifest.Dependencies == nil {
		return []*sourcePackageDependencySummary{}
	}
	items := make([]*sourcePackageDependencySummary, 0, len(manifest.Dependencies.Plugins)+1)
	if manifest.Dependencies.Framework != nil && manifest.Dependencies.Framework.Version != "" {
		items = append(items, &sourcePackageDependencySummary{
			Kind:    "framework",
			Version: manifest.Dependencies.Framework.Version,
		})
	}
	for _, dependency := range manifest.Dependencies.Plugins {
		if dependency == nil {
			continue
		}
		items = append(items, &sourcePackageDependencySummary{
			Kind:    "plugin",
			ID:      dependency.ID,
			Version: dependency.Version,
		})
	}
	return items
}

// buildSourceResourceSummaries lists matching package resources with checksums.
func buildSourceResourceSummaries(
	files map[string]*zip.File,
	rootPrefix string,
	resourcePrefix string,
	kind string,
) ([]*sourcePackageResourceSummary, error) {
	items := make([]*sourcePackageResourceSummary, 0)
	fullPrefix := rootPrefix + resourcePrefix
	for filePath, file := range files {
		if !strings.HasPrefix(filePath, fullPrefix) {
			continue
		}
		if strings.HasSuffix(filePath, "/") {
			continue
		}
		if kind == "sql" && strings.ToLower(filepath.Ext(filePath)) != ".sql" {
			continue
		}
		checksum, err := zipFileSHA256(file)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
		}
		items = append(items, &sourcePackageResourceSummary{
			Kind:      sourcePackageResourceKind(filePath, rootPrefix),
			Path:      strings.TrimPrefix(filePath, rootPrefix),
			SizeBytes: file.UncompressedSize64,
			Sha256:    checksum,
		})
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Path < items[right].Path
	})
	return items, nil
}

// sourcePackageResourceKind classifies SQL resources into install/uninstall/mock kinds.
func sourcePackageResourceKind(filePath string, rootPrefix string) string {
	relative := strings.TrimPrefix(filePath, rootPrefix)
	switch {
	case strings.HasPrefix(relative, "manifest/sql/uninstall/"):
		return "uninstall_sql"
	case strings.HasPrefix(relative, "manifest/sql/mock-data/"):
		return "mock_sql"
	default:
		return "install_sql"
	}
}

// buildSourceI18NSummary lists runtime and apidoc i18n files by locale.
func buildSourceI18NSummary(
	manifest *sourcePackageManifest,
	files map[string]*zip.File,
	rootPrefix string,
) []*sourcePackageI18NSummary {
	byLocale := make(map[string]*sourcePackageI18NSummary)
	if manifest != nil && manifest.I18N != nil {
		for _, locale := range manifest.I18N.Locales {
			if locale == nil || locale.Locale == "" {
				continue
			}
			item := ensureSourceI18NSummary(byLocale, locale.Locale)
			item.DeclaredInPlugin = true
			item.DefaultLocale = locale.Locale == manifest.I18N.Default
		}
	}

	prefix := rootPrefix + "manifest/i18n/"
	for filePath := range files {
		if !strings.HasPrefix(filePath, prefix) || strings.ToLower(filepath.Ext(filePath)) != ".json" {
			continue
		}
		relative := strings.TrimPrefix(filePath, prefix)
		segments := strings.Split(relative, "/")
		if len(segments) < 2 {
			continue
		}
		locale := segments[0]
		item := ensureSourceI18NSummary(byLocale, locale)
		cleanRelative := strings.TrimPrefix(filePath, rootPrefix)
		if len(segments) >= 3 && segments[1] == "apidoc" {
			item.APIDocFiles = append(item.APIDocFiles, cleanRelative)
		} else {
			item.RuntimeFiles = append(item.RuntimeFiles, cleanRelative)
		}
	}

	items := make([]*sourcePackageI18NSummary, 0, len(byLocale))
	for _, item := range byLocale {
		sort.Strings(item.RuntimeFiles)
		sort.Strings(item.APIDocFiles)
		items = append(items, item)
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Locale < items[right].Locale
	})
	return items
}

// ensureSourceI18NSummary returns or creates a locale summary entry.
func ensureSourceI18NSummary(items map[string]*sourcePackageI18NSummary, locale string) *sourcePackageI18NSummary {
	item, ok := items[locale]
	if ok {
		return item
	}
	item = &sourcePackageI18NSummary{Locale: locale}
	items[locale] = item
	return item
}

// buildSourceDocsSummary lists marketplace docs markdown entries.
func buildSourceDocsSummary(files map[string]*zip.File, rootPrefix string) ([]*sourcePackageResourceSummary, error) {
	items, err := buildSourceResourceSummaries(files, rootPrefix, "manifest/docs/", "docs")
	if err != nil {
		return nil, err
	}
	docs := make([]*sourcePackageResourceSummary, 0, len(items))
	for _, item := range items {
		if strings.ToLower(filepath.Ext(item.Path)) != ".md" {
			continue
		}
		item.Kind = "marketplace_doc"
		docs = append(docs, item)
	}
	return docs, nil
}

// sourcePackageDiagnostics creates reviewer-facing findings for a valid source package.
func sourcePackageDiagnostics(
	manifest *sourcePackageManifest,
	sqlFiles []*sourcePackageResourceSummary,
	i18nFiles []*sourcePackageI18NSummary,
	docs []*sourcePackageResourceSummary,
) []*PackageDiagnostic {
	diagnostics := make([]*PackageDiagnostic, 0)
	if len(sqlFiles) > 0 {
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "source_sql_present",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   "manifest/sql",
			Message:  "Source package contains SQL resources that require reviewer inspection.",
		})
	}
	diagnostics = append(diagnostics, &PackageDiagnostic{
		Code:     "source_docs_indexed",
		Severity: marketv1.MarketplaceRiskSeverityInfo,
		Source:   "manifest/docs",
		Message:  "Marketplace documentation entries were detected.",
	})
	if manifest.Dependencies == nil || manifest.Dependencies.Framework == nil || manifest.Dependencies.Framework.Version == "" {
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "framework_dependency_missing",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   "plugin.yaml",
			Message:  "Framework compatibility dependency is not declared.",
		})
	}
	if manifest.I18N != nil && manifest.I18N.Enabled && len(i18nFiles) == 0 {
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "i18n_files_missing",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   "manifest/i18n",
			Message:  "Plugin declares i18n.enabled but no manifest i18n JSON files were detected.",
		})
	}
	return diagnostics
}

// buildSourceRiskSummary aggregates diagnostic severities.
func buildSourceRiskSummary(diagnostics []*PackageDiagnostic) *sourcePackageRiskSummary {
	summary := &sourcePackageRiskSummary{}
	for _, diagnostic := range diagnostics {
		if diagnostic == nil {
			continue
		}
		switch diagnostic.Severity {
		case marketv1.MarketplaceRiskSeverityHigh:
			summary.High++
		case marketv1.MarketplaceRiskSeverityWarning:
			summary.Warning++
		default:
			summary.Info++
		}
	}
	return summary
}

// sourcePackageHostBounds extracts best-effort framework compatibility bounds.
func sourcePackageHostBounds(manifest *sourcePackageManifest) (minVersion string, maxVersion string) {
	if manifest == nil || manifest.Dependencies == nil || manifest.Dependencies.Framework == nil {
		return "", ""
	}
	for _, token := range strings.Fields(manifest.Dependencies.Framework.Version) {
		switch {
		case strings.HasPrefix(token, ">="):
			if minVersion == "" {
				minVersion = strings.TrimPrefix(token, ">=")
			}
		case strings.HasPrefix(token, ">"):
			if minVersion == "" {
				minVersion = strings.TrimPrefix(token, ">")
			}
		case strings.HasPrefix(token, "<="):
			if maxVersion == "" {
				maxVersion = strings.TrimPrefix(token, "<=")
			}
		case strings.HasPrefix(token, "<"):
			if maxVersion == "" {
				maxVersion = strings.TrimPrefix(token, "<")
			}
		case strings.HasPrefix(token, "="):
			version := strings.TrimPrefix(token, "=")
			if minVersion == "" {
				minVersion = version
			}
			if maxVersion == "" {
				maxVersion = version
			}
		default:
			if minVersion == "" {
				minVersion = token
			}
			if maxVersion == "" {
				maxVersion = token
			}
		}
	}
	return minVersion, maxVersion
}

// sourcePackageReviewMessage returns a compact scanner message for release records.
func sourcePackageReviewMessage(diagnostics []*PackageDiagnostic) string {
	if len(diagnostics) == 0 {
		return "Source package scan completed."
	}
	messages := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if diagnostic == nil || diagnostic.Message == "" {
			continue
		}
		messages = append(messages, diagnostic.Message)
	}
	joined := strings.Join(messages, " ")
	if len(joined) > 1024 {
		return joined[:1024]
	}
	return joined
}

// packageJSONString marshals scanner payloads into JSONB-ready strings.
func packageJSONString(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
	}
	return string(content), nil
}

// zipFileSHA256 returns a SHA-256 checksum for one ZIP entry.
func zipFileSHA256(file *zip.File) (string, error) {
	content, err := readZipFile(file)
	if err != nil {
		return "", err
	}
	return sha256Hex(content), nil
}

// sha256Hex returns the SHA-256 checksum for content.
func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// packageDiagnosticError creates a package validation business error.
func packageDiagnosticError(code *bizerr.Code, diagnostic string) error {
	return bizerr.NewCode(code, bizerr.P("diagnostic", diagnostic))
}
