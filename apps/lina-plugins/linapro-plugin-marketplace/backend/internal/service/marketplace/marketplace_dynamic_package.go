// This file implements dynamic-runtime marketplace package upload handling. It
// validates ZIP container safety, rejects development source trees, parses
// plugin.wasm custom sections through public pluginbridge protocols, verifies
// root and embedded manifest consistency, and records dynamic ZIP plus extracted
// plugin.wasm artifact metadata.

package marketplace

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
)

const (
	dynamicPackageDefaultContentType = "application/zip"
	dynamicPackageWasmContentType    = "application/wasm"
	dynamicPackageStoragePrefix      = "dynamic"
	dynamicPackageWasmPath           = "plugin.wasm"
	dynamicPackageReadmePath         = "README.md"
	dynamicPackageReadmeCNPath       = "README.zh-CN.md"
)

// UploadDynamicPackage validates and persists one dynamic runtime marketplace draft.
func (s *serviceImpl) UploadDynamicPackage(
	ctx context.Context,
	in UploadDynamicPackageInput,
) (*DynamicPackageUploadResult, error) {
	if in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	scan, err := scanDynamicPackage(in)
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
		marketv1.MarketplacePluginTypeDynamic,
		in.Visibility,
		in.AutoCreate,
	)
	if err != nil {
		return nil, err
	}
	if err = s.storeUploadedPackage(ctx, scan.storageKey, in.PackagePath); err != nil {
		return nil, err
	}
	if err = s.storeDynamicWasmArtifact(ctx, scan); err != nil {
		return nil, err
	}
	var (
		release         *ReleaseRecord
		packageArtifact *ArtifactRecord
		wasmArtifact    *ArtifactRecord
	)
	if err = dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		release, err = s.SaveReleaseDraft(ctx, SaveReleaseDraftInput{
			PublisherKey:       publisherKey,
			OwnerUserID:        in.OwnerUserID,
			PluginID:           scan.manifest.ID,
			Version:            scan.manifest.Version,
			PluginType:         marketv1.MarketplacePluginTypeDynamic,
			Visibility:         normalizeVisibility(in.Visibility),
			MinHostVersion:     scan.minHostVersion,
			MaxHostVersion:     scan.maxHostVersion,
			ManifestSnapshot:   scan.manifestSnapshot,
			DependencySummary:  scan.dependencySummary,
			HostServiceSummary: scan.hostServiceSummary,
			RouteSummary:       scan.routeSummary,
			SQLSummary:         scan.sqlSummary,
			I18NSummary:        scan.i18nSummary,
			DocsSummary:        scan.docsSummary,
			RiskSummary:        scan.riskSummary,
			ReviewMessage:      dynamicPackageReviewMessage(scan.diagnostics),
			ReplaceDraft:       in.ReplaceDraft,
		})
		if err != nil {
			return err
		}

		packageArtifact, err = s.saveMarketplaceArtifact(ctx, release, &marketplaceArtifactWrite{
			artifactType:   dynamicArtifactTypeForName(scan.fileName),
			storageKey:     scan.storageKey,
			fileName:       scan.fileName,
			contentType:    scan.contentType,
			sizeBytes:      scan.sizeBytes,
			sha256:         scan.packageSha256,
			manifestSha256: scan.manifestSha256,
			wasmSha256:     scan.wasmSha256,
		})
		if err != nil {
			return err
		}
		wasmArtifact, err = s.saveMarketplaceArtifact(ctx, release, &marketplaceArtifactWrite{
			artifactType:   marketv1.MarketplaceArtifactTypePluginWasm,
			storageKey:     scan.wasmStorageKey,
			fileName:       dynamicPackageWasmPath,
			contentType:    dynamicPackageWasmContentType,
			sizeBytes:      scan.wasmSizeBytes,
			sha256:         scan.wasmSha256,
			manifestSha256: scan.manifestSha256,
			wasmSha256:     scan.wasmSha256,
		})
		if err != nil {
			return err
		}
		if err = s.replaceReleaseDisplayI18n(ctx, release, scan.displayI18n); err != nil {
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

	return &DynamicPackageUploadResult{
		Release:         release,
		PackageArtifact: packageArtifact,
		WasmArtifact:    wasmArtifact,
		Diagnostics:     scan.diagnostics,
	}, nil
}

// dynamicPackageScan carries scanner output ready for release draft persistence.
type dynamicPackageScan struct {
	manifest           *sourcePackageManifest
	packageSha256      string
	manifestSha256     string
	wasmSha256         string
	storageKey         string
	wasmStorageKey     string
	fileName           string
	contentType        string
	sizeBytes          int64
	wasmSizeBytes      int64
	wasmBytes          []byte
	minHostVersion     string
	maxHostVersion     string
	manifestSnapshot   string
	dependencySummary  string
	hostServiceSummary string
	routeSummary       string
	sqlSummary         string
	i18nSummary        string
	docsSummary        string
	riskSummary        string
	displayI18n        []*marketplaceDisplayI18nItem
	diagnostics        []*PackageDiagnostic
}

// dynamicPackageWasmSpec carries decoded plugin.wasm custom section data.
type dynamicPackageWasmSpec struct {
	embeddedManifest *sourcePackageManifest
	runtimeMetadata  *protocol.RuntimeArtifactMetadata
	hostServices     []*protocol.HostServiceSpec
	routes           []*contract.RouteContract
	bridgeSpec       *contract.BridgeSpec
	installSQL       []*dynamicPackageSQLAsset
	uninstallSQL     []*dynamicPackageSQLAsset
	mockSQL          []*dynamicPackageSQLAsset
	runtimeI18N      []*dynamicPackageLocaleAsset
	apiDocI18N       []*dynamicPackageLocaleAsset
	resources        []*dynamicPackageManifestResource
}

// dynamicPackageRuntimeCounts stores actual custom-section item counts.
type dynamicPackageRuntimeCounts struct {
	frontendAssets    int
	runtimeI18N       int
	apiDocI18N        int
	totalSQL          int
	mockSQL           int
	manifestResources int
	routes            int
}

// dynamicPackageManifestSnapshot is persisted as the release manifest snapshot.
type dynamicPackageManifestSnapshot struct {
	Root              *sourcePackageManifest            `json:"root"`
	Embedded          *sourcePackageManifest            `json:"embedded"`
	Runtime           *protocol.RuntimeArtifactMetadata `json:"runtime"`
	ManifestResources []*sourcePackageResourceSummary   `json:"manifestResources"`
	Bridge            *contract.BridgeSpec              `json:"bridge,omitempty"`
}

// dynamicPackageSQLAsset mirrors the public JSON shape embedded in plugin.wasm.
type dynamicPackageSQLAsset struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

// dynamicPackageLocaleAsset mirrors locale JSON assets embedded in plugin.wasm.
type dynamicPackageLocaleAsset struct {
	Locale  string `json:"locale"`
	Content string `json:"content"`
}

// dynamicPackageManifestResource mirrors manifest resources embedded in plugin.wasm.
type dynamicPackageManifestResource struct {
	Path          string `json:"path"`
	ContentBase64 string `json:"contentBase64"`
	Content       []byte `json:"-"`
}

// dynamicPackageI18NAssetSummary is a locale asset summary persisted as JSON.
type dynamicPackageI18NAssetSummary struct {
	Locale    string `json:"locale"`
	Kind      string `json:"kind"`
	SizeBytes int    `json:"sizeBytes"`
	Sha256    string `json:"sha256"`
}

// dynamicPackageHostServiceSummary is a review-friendly host service summary.
type dynamicPackageHostServiceSummary struct {
	Service       string   `json:"service"`
	Methods       []string `json:"methods"`
	Paths         []string `json:"paths,omitempty"`
	Tables        []string `json:"tables,omitempty"`
	Keys          []string `json:"keys,omitempty"`
	ResourceCount int      `json:"resourceCount,omitempty"`
}

// dynamicPackageRouteSummary is a review-friendly dynamic route summary.
type dynamicPackageRouteSummary struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Access      string `json:"access"`
	Permission  string `json:"permission,omitempty"`
	RequestType string `json:"requestType"`
}

// marketplaceArtifactWrite carries one artifact upsert payload.
type marketplaceArtifactWrite struct {
	artifactType   marketv1.MarketplaceArtifactType
	storageKey     string
	fileName       string
	contentType    string
	sizeBytes      int64
	sha256         string
	manifestSha256 string
	wasmSha256     string
}

// scanDynamicPackage parses and validates one uploaded dynamic runtime package.
func scanDynamicPackage(in UploadDynamicPackageInput) (scan *dynamicPackageScan, err error) {
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
	rootPrefix, err := detectDynamicPackageRoot(fileIndex)
	if err != nil {
		return nil, err
	}
	if err = validateDynamicPackageStructure(fileIndex, rootPrefix); err != nil {
		return nil, err
	}

	manifestBytes, err := readZipFile(fileIndex[rootPrefix+sourcePackageManifestPath])
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
	}
	rootManifest := &sourcePackageManifest{}
	if err = yaml.Unmarshal(manifestBytes, rootManifest); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.yaml cannot be parsed")
	}
	normalizeSourcePackageManifest(rootManifest)
	if err = validateDynamicPackageManifest(rootManifest, in); err != nil {
		return nil, err
	}

	wasmFile := fileIndex[rootPrefix+dynamicPackageWasmPath]
	wasmBytes, err := readZipFile(wasmFile)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
	}
	wasmSpec, err := parseDynamicWasmSpec(wasmBytes)
	if err != nil {
		return nil, err
	}
	if err = normalizeAndCompareDynamicManifests(rootManifest, wasmSpec); err != nil {
		return nil, err
	}

	resourceSummary := buildDynamicManifestResourceSummary(wasmSpec.resources)
	manifestSnapshot, err := packageJSONString(&dynamicPackageManifestSnapshot{
		Root:              rootManifest,
		Embedded:          wasmSpec.embeddedManifest,
		Runtime:           wasmSpec.runtimeMetadata,
		ManifestResources: resourceSummary,
		Bridge:            wasmSpec.bridgeSpec,
	})
	if err != nil {
		return nil, err
	}
	dependencySummary, err := packageJSONString(buildSourceDependencySummary(rootManifest))
	if err != nil {
		return nil, err
	}
	hostServiceSummary, err := packageJSONString(buildDynamicHostServiceSummary(wasmSpec.hostServices))
	if err != nil {
		return nil, err
	}
	routeSummary, err := packageJSONString(buildDynamicRouteSummary(wasmSpec.routes))
	if err != nil {
		return nil, err
	}
	sqlSummary, err := packageJSONString(buildDynamicSQLSummary(wasmSpec))
	if err != nil {
		return nil, err
	}
	i18nSummary, err := packageJSONString(buildDynamicI18NSummary(wasmSpec))
	if err != nil {
		return nil, err
	}
	docs, err := buildDynamicDocsSummary(fileIndex, rootPrefix, wasmSpec.resources)
	if err != nil {
		return nil, err
	}
	docsSummary, err := packageJSONString(docs)
	if err != nil {
		return nil, err
	}
	defaultLocale := defaultDisplayLocale
	if rootManifest.I18N != nil {
		defaultLocale = defaultLocaleFromManifest(rootManifest.I18N.Default)
	}
	displayI18n := buildDisplayI18nFromPackageYAML(
		rootManifest.ID,
		rootManifest.Name,
		rootManifest.Description,
		defaultLocale,
	)
	displayI18n = mergePackageI18nDisplayItems(
		rootManifest.ID,
		displayI18n,
		extractDynamicPackageDisplayCatalogs(wasmSpec.runtimeI18N),
	)
	diagnostics := dynamicPackageDiagnostics(wasmSpec)
	riskSummary, err := packageJSONString(buildSourceRiskSummary(diagnostics))
	if err != nil {
		return nil, err
	}

	minHostVersion, maxHostVersion := sourcePackageHostBounds(rootManifest)
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
		contentType = dynamicPackageDefaultContentType
	}
	wasmSha := sha256Hex(wasmBytes)
	storageKey := normalizeKey(in.StorageKey)
	if storageKey == "" {
		storageKey = path.Join(dynamicPackageStoragePrefix, rootManifest.ID, rootManifest.Version, packageSha+storageKeyExtension(fileName))
	}
	wasmStorageKey := normalizeKey(in.WasmStorageKey)
	if wasmStorageKey == "" {
		wasmStorageKey = path.Join(dynamicPackageStoragePrefix, rootManifest.ID, rootManifest.Version, wasmSha+".wasm")
	}

	return &dynamicPackageScan{
		manifest:           rootManifest,
		packageSha256:      packageSha,
		manifestSha256:     sha256Hex(manifestBytes),
		wasmSha256:         wasmSha,
		storageKey:         storageKey,
		wasmStorageKey:     wasmStorageKey,
		fileName:           fileName,
		contentType:        contentType,
		sizeBytes:          sizeBytes,
		wasmSizeBytes:      int64(len(wasmBytes)),
		wasmBytes:          wasmBytes,
		minHostVersion:     minHostVersion,
		maxHostVersion:     maxHostVersion,
		manifestSnapshot:   manifestSnapshot,
		dependencySummary:  dependencySummary,
		hostServiceSummary: hostServiceSummary,
		routeSummary:       routeSummary,
		sqlSummary:         sqlSummary,
		i18nSummary:        i18nSummary,
		docsSummary:        docsSummary,
		riskSummary:        riskSummary,
		displayI18n:        displayI18n,
		diagnostics:        diagnostics,
	}, err
}

// storeDynamicWasmArtifact persists the extracted plugin.wasm bytes for direct download.
func (s *serviceImpl) storeDynamicWasmArtifact(ctx context.Context, scan *dynamicPackageScan) error {
	if scan == nil || len(scan.wasmBytes) == 0 || normalizeKey(scan.wasmStorageKey) == "" {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	if s.artifacts == nil {
		return bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	return s.artifacts.Put(ctx, scan.wasmStorageKey, bytes.NewReader(scan.wasmBytes))
}

// storeUploadedPackage copies one local uploaded package file into durable storage.
func (s *serviceImpl) storeUploadedPackage(ctx context.Context, storageKey string, packagePath string) error {
	if s.artifacts == nil {
		return bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	if normalizeKey(storageKey) == "" || normalizeKey(packagePath) == "" {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	return s.artifacts.PutFile(ctx, storageKey, packagePath)
}

// saveMarketplaceArtifact creates or replaces one release artifact metadata row.
func (s *serviceImpl) saveMarketplaceArtifact(
	ctx context.Context,
	release *ReleaseRecord,
	write *marketplaceArtifactWrite,
) (*ArtifactRecord, error) {
	if release == nil || write == nil {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	artifactType := write.artifactType.String()
	data := do.PluginMarketplaceArtifact{
		ReleaseId:      release.ID,
		PluginId:       release.PluginID,
		ReleaseVersion: release.Version,
		ArtifactType:   artifactType,
		StorageKey:     normalizeKey(write.storageKey),
		FileName:       normalizeKey(write.fileName),
		ContentType:    normalizeKey(write.contentType),
		SizeBytes:      write.sizeBytes,
		Sha256:         normalizeKey(write.sha256),
		ManifestSha256: normalizeKey(write.manifestSha256),
		WasmSha256:     normalizeKey(write.wasmSha256),
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

// detectDynamicPackageRoot supports either direct-root or single-directory dynamic packages.
func detectDynamicPackageRoot(files map[string]*zip.File) (string, error) {
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
			return "", packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "dynamic package must contain a single plugin root directory")
		}
	}
	if _, ok := files[root+"/"+sourcePackageManifestPath]; !ok {
		return "", packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "dynamic package root is missing plugin.yaml")
	}
	return root + "/", nil
}

// validateDynamicPackageStructure checks required runtime package entries and source exclusions.
func validateDynamicPackageStructure(files map[string]*zip.File, rootPrefix string) error {
	requiredFiles := []string{sourcePackageManifestPath, dynamicPackageWasmPath}
	for _, required := range requiredFiles {
		if _, ok := files[rootPrefix+required]; !ok {
			return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "dynamic package is missing "+required)
		}
	}
	if _, ok := files[rootPrefix+"main.go"]; ok {
		return packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "dynamic runtime package must not contain main.go")
	}
	for _, forbidden := range []string{"backend/", "frontend/", "hack/"} {
		if zipDirHasFile(files, rootPrefix+forbidden) {
			return packageDiagnosticError(
				CodeMarketplacePackageStructureInvalid,
				"dynamic runtime package must not contain "+strings.TrimSuffix(forbidden, "/")+"/",
			)
		}
	}
	return nil
}

// validateDynamicPackageManifest validates root plugin.yaml identity and dependency fields.
func validateDynamicPackageManifest(manifest *sourcePackageManifest, in UploadDynamicPackageInput) error {
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
	if manifest.Type != marketv1.MarketplacePluginTypeDynamic.String() {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml type must be dynamic")
	}
	if manifest.Distribution != "managed" {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "dynamic package distribution must be managed")
	}
	if normalizeKey(in.PluginID) != "" && manifest.ID != normalizeKey(in.PluginID) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml id does not match upload pluginId")
	}
	if normalizeKey(in.Version) != "" && manifest.Version != normalizeKey(in.Version) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.yaml version does not match upload version")
	}
	return validateSourcePackageDependencies(manifest)
}

// parseDynamicWasmSpec parses plugin.wasm and validates public runtime sections.
func parseDynamicWasmSpec(wasmBytes []byte) (*dynamicPackageWasmSpec, error) {
	sections, err := protocol.ListCustomSections(wasmBytes)
	if err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm must be a valid WASM artifact")
	}
	manifestSection, ok := sections[protocol.WasmSectionManifest]
	if !ok {
		return nil, packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "plugin.wasm is missing embedded manifest section")
	}
	runtimeSection, ok := sections[protocol.WasmSectionRuntime]
	if !ok {
		return nil, packageDiagnosticError(CodeMarketplacePackageStructureInvalid, "plugin.wasm is missing runtime metadata section")
	}

	embeddedManifest := &sourcePackageManifest{}
	if err = json.Unmarshal(manifestSection, embeddedManifest); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm embedded manifest cannot be parsed")
	}
	normalizeSourcePackageManifest(embeddedManifest)
	if err = validateDynamicEmbeddedManifest(embeddedManifest); err != nil {
		return nil, err
	}

	runtimeMetadata := &protocol.RuntimeArtifactMetadata{}
	if err = json.Unmarshal(runtimeSection, runtimeMetadata); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm runtime metadata cannot be parsed")
	}

	hostServices, err := parseDynamicHostServices(embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	routes, err := parseDynamicRoutes(embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	bridgeSpec, err := parseDynamicBridgeSpec(sections)
	if err != nil {
		return nil, err
	}
	installSQL, err := parseDynamicSQLAssets(sections, protocol.WasmSectionInstallSQL, "install SQL")
	if err != nil {
		return nil, err
	}
	uninstallSQL, err := parseDynamicSQLAssets(sections, protocol.WasmSectionUninstallSQL, "uninstall SQL")
	if err != nil {
		return nil, err
	}
	mockSQL, err := parseDynamicSQLAssets(sections, protocol.WasmSectionMockSQL, "mock SQL")
	if err != nil {
		return nil, err
	}
	runtimeI18N, err := parseDynamicLocaleAssets(sections, protocol.WasmSectionI18NAssets, "runtime_i18n")
	if err != nil {
		return nil, err
	}
	apiDocI18N, err := parseDynamicLocaleAssets(sections, protocol.WasmSectionAPIDocI18NAssets, "apidoc_i18n")
	if err != nil {
		return nil, err
	}
	resources, err := parseDynamicManifestResources(sections)
	if err != nil {
		return nil, err
	}
	counts, err := dynamicRuntimeCounts(sections, installSQL, uninstallSQL, mockSQL, runtimeI18N, apiDocI18N, resources, routes)
	if err != nil {
		return nil, err
	}
	if err = validateDynamicRuntimeMetadata(runtimeMetadata, counts); err != nil {
		return nil, err
	}

	embeddedManifest.HostServices = hostServices
	return &dynamicPackageWasmSpec{
		embeddedManifest: embeddedManifest,
		runtimeMetadata:  runtimeMetadata,
		hostServices:     hostServices,
		routes:           routes,
		bridgeSpec:       bridgeSpec,
		installSQL:       installSQL,
		uninstallSQL:     uninstallSQL,
		mockSQL:          mockSQL,
		runtimeI18N:      runtimeI18N,
		apiDocI18N:       apiDocI18N,
		resources:        resources,
	}, nil
}

// validateDynamicEmbeddedManifest validates plugin.wasm embedded manifest fields.
func validateDynamicEmbeddedManifest(manifest *sourcePackageManifest) error {
	if manifest == nil || manifest.ID == "" || manifest.Name == "" || manifest.Version == "" || manifest.Type == "" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm embedded manifest must include id, name, version, and type")
	}
	if !sourcePackagePluginIDPattern.MatchString(manifest.ID) || len(manifest.ID) > 64 {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm embedded manifest id is invalid")
	}
	if !sourcePackageSemverPattern.MatchString(manifest.Version) {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm embedded manifest version must use semver format")
	}
	if manifest.Type != marketv1.MarketplacePluginTypeDynamic.String() {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.wasm embedded manifest type must be dynamic")
	}
	if manifest.Distribution != "managed" {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "plugin.wasm embedded distribution must be managed")
	}
	return validateSourcePackageDependencies(manifest)
}

// normalizeAndCompareDynamicManifests verifies root plugin.yaml and plugin.wasm consistency.
func normalizeAndCompareDynamicManifests(root *sourcePackageManifest, wasmSpec *dynamicPackageWasmSpec) error {
	if root == nil || wasmSpec == nil || wasmSpec.embeddedManifest == nil {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "dynamic package manifests are missing")
	}

	rootHostServices, err := normalizeDynamicHostServices(root.ID, root.HostServices, "root plugin.yaml hostServices")
	if err != nil {
		return err
	}
	root.HostServices = rootHostServices

	embedded := wasmSpec.embeddedManifest
	embedded.HostServices = wasmSpec.hostServices
	for _, field := range []struct {
		name  string
		left  string
		right string
	}{
		{name: "id", left: root.ID, right: embedded.ID},
		{name: "name", left: root.Name, right: embedded.Name},
		{name: "version", left: root.Version, right: embedded.Version},
		{name: "type", left: root.Type, right: embedded.Type},
		{name: "scopeNature", left: root.ScopeNature, right: embedded.ScopeNature},
		{name: "defaultInstallMode", left: root.DefaultInstallMode, right: embedded.DefaultInstallMode},
	} {
		if field.left != field.right {
			return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, field.name+" differs between root plugin.yaml and plugin.wasm")
		}
	}
	if !dynamicBoolPointersEqual(root.SupportsMultiTenant, embedded.SupportsMultiTenant) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "supportsMultiTenant differs between root plugin.yaml and plugin.wasm")
	}
	if !dynamicDependencySummariesEqual(root, embedded) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "dependencies differ between root plugin.yaml and plugin.wasm")
	}
	if !dynamicHostServicesEqual(root.HostServices, embedded.HostServices) {
		return packageDiagnosticError(CodeMarketplacePackageManifestMismatch, "hostServices differ between root plugin.yaml and plugin.wasm")
	}
	return nil
}

// parseDynamicHostServices decodes and normalizes embedded host service declarations.
func parseDynamicHostServices(pluginID string, sections map[string][]byte) ([]*protocol.HostServiceSpec, error) {
	content, ok := sections[protocol.WasmSectionBackendHostServices]
	if !ok {
		return []*protocol.HostServiceSpec{}, nil
	}

	items := make([]*protocol.HostServiceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm hostServices section cannot be parsed")
	}
	return normalizeDynamicHostServices(pluginID, items, "plugin.wasm hostServices")
}

// normalizeDynamicHostServices validates and sorts host service declarations.
func normalizeDynamicHostServices(
	pluginID string,
	items []*protocol.HostServiceSpec,
	source string,
) ([]*protocol.HostServiceSpec, error) {
	normalized, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, items)
	if err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, source+" is invalid")
	}
	sort.Slice(normalized, func(left int, right int) bool {
		return normalized[left].Service < normalized[right].Service
	})
	for _, item := range normalized {
		if item == nil {
			continue
		}
		sort.Slice(item.Resources, func(left int, right int) bool {
			return item.Resources[left].Ref < item.Resources[right].Ref
		})
	}
	return normalized, nil
}

// parseDynamicRoutes decodes and validates embedded dynamic route contracts.
func parseDynamicRoutes(pluginID string, sections map[string][]byte) ([]*contract.RouteContract, error) {
	content, ok := sections[protocol.WasmSectionBackendRoutes]
	if !ok {
		return []*contract.RouteContract{}, nil
	}

	routes := make([]*contract.RouteContract, 0)
	if err := json.Unmarshal(content, &routes); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm routes section cannot be parsed")
	}
	if err := contract.ValidateRouteContracts(pluginID, routes); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm routes section is invalid")
	}
	sort.Slice(routes, func(left int, right int) bool {
		leftKey := routes[left].Method + " " + routes[left].Path
		rightKey := routes[right].Method + " " + routes[right].Path
		return leftKey < rightKey
	})
	return routes, nil
}

// parseDynamicBridgeSpec decodes and validates the optional bridge contract.
func parseDynamicBridgeSpec(sections map[string][]byte) (*contract.BridgeSpec, error) {
	content, ok := sections[protocol.WasmSectionBackendBridge]
	if !ok {
		return nil, nil
	}

	spec := &contract.BridgeSpec{}
	if err := json.Unmarshal(content, spec); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm bridge section cannot be parsed")
	}
	if err := contract.ValidateBridgeSpec(spec); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm bridge section is invalid")
	}
	return spec, nil
}

// parseDynamicSQLAssets decodes one embedded SQL section.
func parseDynamicSQLAssets(
	sections map[string][]byte,
	sectionName string,
	source string,
) ([]*dynamicPackageSQLAsset, error) {
	content, ok := sections[sectionName]
	if !ok {
		return []*dynamicPackageSQLAsset{}, nil
	}

	assets := make([]*dynamicPackageSQLAsset, 0)
	if err := json.Unmarshal(content, &assets); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+source+" section cannot be parsed")
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+source+" contains a null item")
		}
		asset.Key = normalizeKey(asset.Key)
		asset.Content = normalizeKey(asset.Content)
		if asset.Key == "" || asset.Content == "" {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+source+" is missing key or content")
		}
		if strings.Contains(asset.Key, "/") || strings.Contains(asset.Key, "\\") || strings.ToLower(filepath.Ext(asset.Key)) != ".sql" {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+source+" key must be a SQL file name")
		}
	}
	sort.Slice(assets, func(left int, right int) bool {
		return assets[left].Key < assets[right].Key
	})
	return assets, nil
}

// parseDynamicLocaleAssets decodes one embedded locale JSON section.
func parseDynamicLocaleAssets(
	sections map[string][]byte,
	sectionName string,
	kind string,
) ([]*dynamicPackageLocaleAsset, error) {
	content, ok := sections[sectionName]
	if !ok {
		return []*dynamicPackageLocaleAsset{}, nil
	}

	assets := make([]*dynamicPackageLocaleAsset, 0)
	if err := json.Unmarshal(content, &assets); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+kind+" section cannot be parsed")
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+kind+" contains a null item")
		}
		asset.Locale = normalizeKey(asset.Locale)
		asset.Content = normalizeKey(asset.Content)
		if asset.Locale == "" || asset.Content == "" {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+kind+" is missing locale or content")
		}
		if !dynamicJSONContentIsObject(asset.Content) {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+kind+" content must be a JSON object")
		}
	}
	sort.Slice(assets, func(left int, right int) bool {
		return assets[left].Locale < assets[right].Locale
	})
	return assets, nil
}

// parseDynamicManifestResources decodes embedded manifest resource payloads.
func parseDynamicManifestResources(sections map[string][]byte) ([]*dynamicPackageManifestResource, error) {
	content, ok := sections[protocol.WasmSectionManifestResources]
	if !ok {
		return []*dynamicPackageManifestResource{}, nil
	}

	resources := make([]*dynamicPackageManifestResource, 0)
	if err := json.Unmarshal(content, &resources); err != nil {
		return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm manifest resources section cannot be parsed")
	}
	seen := make(map[string]struct{}, len(resources))
	for _, resource := range resources {
		if resource == nil {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm manifest resources contain a null item")
		}
		normalizedPath, err := normalizeDynamicManifestResourcePath(resource.Path)
		if err != nil {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm manifest resource path is invalid")
		}
		if _, ok := seen[normalizedPath]; ok {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm manifest resource path is duplicated")
		}
		seen[normalizedPath] = struct{}{}
		decoded, err := base64.StdEncoding.DecodeString(normalizeKey(resource.ContentBase64))
		if err != nil || len(decoded) == 0 {
			return nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm manifest resource content is invalid")
		}
		resource.Path = normalizedPath
		resource.Content = decoded
	}
	sort.Slice(resources, func(left int, right int) bool {
		return resources[left].Path < resources[right].Path
	})
	return resources, nil
}

// dynamicRuntimeCounts computes actual custom-section counts for metadata checks.
func dynamicRuntimeCounts(
	sections map[string][]byte,
	installSQL []*dynamicPackageSQLAsset,
	uninstallSQL []*dynamicPackageSQLAsset,
	mockSQL []*dynamicPackageSQLAsset,
	runtimeI18N []*dynamicPackageLocaleAsset,
	apiDocI18N []*dynamicPackageLocaleAsset,
	resources []*dynamicPackageManifestResource,
	routes []*contract.RouteContract,
) (dynamicPackageRuntimeCounts, error) {
	frontendCount, err := dynamicSectionArrayCount(sections, protocol.WasmSectionFrontendAssets, "frontend assets")
	if err != nil {
		return dynamicPackageRuntimeCounts{}, err
	}
	return dynamicPackageRuntimeCounts{
		frontendAssets:    frontendCount,
		runtimeI18N:       len(runtimeI18N),
		apiDocI18N:        len(apiDocI18N),
		totalSQL:          len(installSQL) + len(uninstallSQL) + len(mockSQL),
		mockSQL:           len(mockSQL),
		manifestResources: len(resources),
		routes:            len(routes),
	}, nil
}

// dynamicSectionArrayCount returns the array length for optional sections.
func dynamicSectionArrayCount(sections map[string][]byte, sectionName string, source string) (int, error) {
	content, ok := sections[sectionName]
	if !ok {
		return 0, nil
	}
	items := make([]json.RawMessage, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return 0, packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+source+" section cannot be parsed")
	}
	return len(items), nil
}

// validateDynamicRuntimeMetadata verifies runtime kind, ABI, and section counts.
func validateDynamicRuntimeMetadata(
	metadata *protocol.RuntimeArtifactMetadata,
	counts dynamicPackageRuntimeCounts,
) error {
	if metadata == nil {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm runtime metadata is missing")
	}
	runtimeKind := strings.ToLower(normalizeKey(metadata.RuntimeKind))
	if runtimeKind == "" {
		runtimeKind = contract.RuntimeKindWasm
	}
	if runtimeKind != contract.RuntimeKindWasm {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm runtime kind must be wasm")
	}
	abiVersion := strings.ToLower(normalizeKey(metadata.ABIVersion))
	if abiVersion == "" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm ABI version is missing")
	}
	if abiVersion != contract.SupportedABIVersion {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm ABI version is not supported")
	}
	metadata.RuntimeKind = runtimeKind
	metadata.ABIVersion = abiVersion
	if err := dynamicCountMatches(metadata.FrontendAssetCount, counts.frontendAssets, "frontend asset"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.I18NAssetCount, counts.runtimeI18N, "runtime i18n asset"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.APIDocI18NAssetCount, counts.apiDocI18N, "apidoc i18n asset"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.SQLAssetCount, counts.totalSQL, "SQL asset"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.MockSQLAssetCount, counts.mockSQL, "mock SQL asset"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.ManifestResourceCount, counts.manifestResources, "manifest resource"); err != nil {
		return err
	}
	if err := dynamicCountMatches(metadata.RouteCount, counts.routes, "route"); err != nil {
		return err
	}
	metadata.FrontendAssetCount = maxPositive(metadata.FrontendAssetCount, counts.frontendAssets)
	metadata.I18NAssetCount = maxPositive(metadata.I18NAssetCount, counts.runtimeI18N)
	metadata.APIDocI18NAssetCount = maxPositive(metadata.APIDocI18NAssetCount, counts.apiDocI18N)
	metadata.SQLAssetCount = maxPositive(metadata.SQLAssetCount, counts.totalSQL)
	metadata.MockSQLAssetCount = maxPositive(metadata.MockSQLAssetCount, counts.mockSQL)
	metadata.ManifestResourceCount = maxPositive(metadata.ManifestResourceCount, counts.manifestResources)
	metadata.RouteCount = maxPositive(metadata.RouteCount, counts.routes)
	return nil
}

// dynamicCountMatches validates positive metadata counts against actual section counts.
func dynamicCountMatches(expected int, actual int, label string) error {
	if expected > 0 && expected != actual {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "plugin.wasm "+label+" count does not match runtime metadata")
	}
	return nil
}

// maxPositive returns actual when metadata omitted the count.
func maxPositive(expected int, actual int) int {
	if expected > 0 {
		return expected
	}
	return actual
}

// buildDynamicManifestResourceSummary lists embedded manifest resources.
func buildDynamicManifestResourceSummary(resources []*dynamicPackageManifestResource) []*sourcePackageResourceSummary {
	items := make([]*sourcePackageResourceSummary, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		items = append(items, &sourcePackageResourceSummary{
			Kind:      dynamicManifestResourceKind(resource.Path),
			Path:      resource.Path,
			SizeBytes: uint64(len(resource.Content)),
			Sha256:    sha256Hex(resource.Content),
		})
	}
	return items
}

// buildDynamicSQLSummary lists embedded SQL resources.
func buildDynamicSQLSummary(spec *dynamicPackageWasmSpec) []*sourcePackageResourceSummary {
	items := make([]*sourcePackageResourceSummary, 0)
	items = append(items, dynamicSQLAssetSummaries(spec.installSQL, "install_sql")...)
	items = append(items, dynamicSQLAssetSummaries(spec.uninstallSQL, "uninstall_sql")...)
	items = append(items, dynamicSQLAssetSummaries(spec.mockSQL, "mock_sql")...)
	sort.Slice(items, func(left int, right int) bool {
		if items[left].Kind == items[right].Kind {
			return items[left].Path < items[right].Path
		}
		return items[left].Kind < items[right].Kind
	})
	return items
}

// dynamicSQLAssetSummaries converts SQL assets to persisted summaries.
func dynamicSQLAssetSummaries(assets []*dynamicPackageSQLAsset, kind string) []*sourcePackageResourceSummary {
	items := make([]*sourcePackageResourceSummary, 0, len(assets))
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		content := []byte(asset.Content)
		items = append(items, &sourcePackageResourceSummary{
			Kind:      kind,
			Path:      asset.Key,
			SizeBytes: uint64(len(content)),
			Sha256:    sha256Hex(content),
		})
	}
	return items
}

// buildDynamicI18NSummary lists embedded runtime and API-doc i18n resources.
func buildDynamicI18NSummary(spec *dynamicPackageWasmSpec) []*dynamicPackageI18NAssetSummary {
	items := make([]*dynamicPackageI18NAssetSummary, 0, len(spec.runtimeI18N)+len(spec.apiDocI18N))
	items = append(items, dynamicLocaleAssetSummaries(spec.runtimeI18N, "runtime_i18n")...)
	items = append(items, dynamicLocaleAssetSummaries(spec.apiDocI18N, "apidoc_i18n")...)
	sort.Slice(items, func(left int, right int) bool {
		if items[left].Locale == items[right].Locale {
			return items[left].Kind < items[right].Kind
		}
		return items[left].Locale < items[right].Locale
	})
	return items
}

// dynamicLocaleAssetSummaries converts locale assets to persisted summaries.
func dynamicLocaleAssetSummaries(assets []*dynamicPackageLocaleAsset, kind string) []*dynamicPackageI18NAssetSummary {
	items := make([]*dynamicPackageI18NAssetSummary, 0, len(assets))
	for _, asset := range assets {
		if asset == nil {
			continue
		}
		content := []byte(asset.Content)
		items = append(items, &dynamicPackageI18NAssetSummary{
			Locale:    asset.Locale,
			Kind:      kind,
			SizeBytes: len(content),
			Sha256:    sha256Hex(content),
		})
	}
	return items
}

// buildDynamicDocsSummary lists ZIP and embedded manifest documentation entries.
func buildDynamicDocsSummary(
	files map[string]*zip.File,
	rootPrefix string,
	resources []*dynamicPackageManifestResource,
) ([]*sourcePackageResourceSummary, error) {
	items := make([]*sourcePackageResourceSummary, 0)
	for filePath, file := range files {
		relative := strings.TrimPrefix(filePath, rootPrefix)
		if strings.HasPrefix(relative, "manifest/docs/") && strings.ToLower(filepath.Ext(relative)) == ".md" {
			checksum, err := zipFileSHA256(file)
			if err != nil {
				return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
			}
			items = append(items, &sourcePackageResourceSummary{
				Kind:      "marketplace_doc",
				Path:      relative,
				SizeBytes: file.UncompressedSize64,
				Sha256:    checksum,
			})
			continue
		}
		if relative == dynamicPackageReadmePath || relative == dynamicPackageReadmeCNPath {
			checksum, err := zipFileSHA256(file)
			if err != nil {
				return nil, bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
			}
			items = append(items, &sourcePackageResourceSummary{
				Kind:      "marketplace_readme",
				Path:      relative,
				SizeBytes: file.UncompressedSize64,
				Sha256:    checksum,
			})
		}
	}
	for _, resource := range resources {
		if resource == nil || !strings.HasPrefix(resource.Path, "manifest/docs/") ||
			strings.ToLower(filepath.Ext(resource.Path)) != ".md" {
			continue
		}
		items = append(items, &sourcePackageResourceSummary{
			Kind:      "embedded_marketplace_doc",
			Path:      resource.Path,
			SizeBytes: uint64(len(resource.Content)),
			Sha256:    sha256Hex(resource.Content),
		})
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Path < items[right].Path
	})
	return items, nil
}

// buildDynamicHostServiceSummary builds reviewer-facing host service summaries.
func buildDynamicHostServiceSummary(items []*protocol.HostServiceSpec) []*dynamicPackageHostServiceSummary {
	summaries := make([]*dynamicPackageHostServiceSummary, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		summaries = append(summaries, &dynamicPackageHostServiceSummary{
			Service:       item.Service,
			Methods:       append([]string{}, item.Methods...),
			Paths:         append([]string{}, item.Paths...),
			Tables:        append([]string{}, item.Tables...),
			Keys:          append([]string{}, item.Keys...),
			ResourceCount: len(item.Resources),
		})
	}
	return summaries
}

// buildDynamicRouteSummary builds reviewer-facing dynamic route summaries.
func buildDynamicRouteSummary(routes []*contract.RouteContract) []*dynamicPackageRouteSummary {
	summaries := make([]*dynamicPackageRouteSummary, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		summaries = append(summaries, &dynamicPackageRouteSummary{
			Method:      route.Method,
			Path:        route.Path,
			Access:      route.Access,
			Permission:  route.Permission,
			RequestType: route.RequestType,
		})
	}
	return summaries
}

// dynamicPackageDiagnostics creates reviewer-facing findings for a valid dynamic package.
func dynamicPackageDiagnostics(spec *dynamicPackageWasmSpec) []*PackageDiagnostic {
	diagnostics := []*PackageDiagnostic{{
		Code:     "dynamic_runtime_detected",
		Severity: marketv1.MarketplaceRiskSeverityInfo,
		Source:   dynamicPackageWasmPath,
		Message:  "Dynamic runtime artifact was parsed successfully.",
	}}
	if len(spec.hostServices) > 0 {
		severity := marketv1.MarketplaceRiskSeverityWarning
		if dynamicHostServicesIncludeHighRisk(spec.hostServices) {
			severity = marketv1.MarketplaceRiskSeverityHigh
		}
		services := make([]DiagnosticServiceEvidence, 0, len(spec.hostServices))
		for _, item := range spec.hostServices {
			if item == nil {
				continue
			}
			services = append(services, DiagnosticServiceEvidence{
				Service: item.Service,
				Methods: append([]string{}, item.Methods...),
				Tables:  append([]string{}, item.Tables...),
				Paths:   append([]string{}, item.Paths...),
				Keys:    append([]string{}, item.Keys...),
			})
		}
		bounded, total, truncated := boundServiceEvidence(services)
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "dynamic_host_services_present",
			Severity: severity,
			Source:   protocol.WasmSectionBackendHostServices,
			Message:  "Dynamic package requests host service authorization.",
			Evidence: &PackageDiagnosticEvidence{
				Services:   bounded,
				TotalCount: total,
				Truncated:  truncated,
			},
		})
	}
	if len(spec.routes) > 0 {
		routes := make([]DiagnosticRouteEvidence, 0, len(spec.routes))
		for _, route := range spec.routes {
			if route == nil {
				continue
			}
			routes = append(routes, DiagnosticRouteEvidence{
				Method:     route.Method,
				Path:       route.Path,
				Permission: route.Permission,
				Access:     route.Access,
			})
		}
		bounded, total, truncated := boundRouteEvidence(routes)
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "dynamic_routes_present",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   protocol.WasmSectionBackendRoutes,
			Message:  "Dynamic package exposes runtime routes.",
			Evidence: &PackageDiagnosticEvidence{
				Routes:     bounded,
				TotalCount: total,
				Truncated:  truncated,
			},
		})
	}
	if len(spec.installSQL)+len(spec.uninstallSQL)+len(spec.mockSQL) > 0 {
		files := make([]string, 0, 3)
		if len(spec.installSQL) > 0 {
			files = append(files, "wasm:"+protocol.WasmSectionInstallSQL)
		}
		if len(spec.uninstallSQL) > 0 {
			files = append(files, "wasm:"+protocol.WasmSectionUninstallSQL)
		}
		if len(spec.mockSQL) > 0 {
			files = append(files, "wasm:"+protocol.WasmSectionMockSQL)
		}
		bounded, total, truncated := boundStringEvidence(files)
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "dynamic_sql_present",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   "plugin.wasm sql sections",
			Message:  "Dynamic package contains SQL resources that require reviewer inspection.",
			Evidence: &PackageDiagnosticEvidence{
				Files:      bounded,
				TotalCount: total,
				Truncated:  truncated,
			},
		})
	}
	if len(spec.mockSQL) > 0 {
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "dynamic_mock_sql_present",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   protocol.WasmSectionMockSQL,
			Message:  "Dynamic package includes optional mock SQL resources.",
			Evidence: &PackageDiagnosticEvidence{
				Files:      []string{"wasm:" + protocol.WasmSectionMockSQL},
				TotalCount: 1,
			},
		})
	}
	if len(spec.resources) == 0 {
		diagnostics = append(diagnostics, &PackageDiagnostic{
			Code:     "dynamic_manifest_resources_missing",
			Severity: marketv1.MarketplaceRiskSeverityWarning,
			Source:   protocol.WasmSectionManifestResources,
			Message:  "Dynamic package does not embed manifest resources.",
			Evidence: &PackageDiagnosticEvidence{
				ExpectedPath:  "plugin.wasm",
				ExpectedField: protocol.WasmSectionManifestResources,
			},
		})
	}
	return diagnostics
}

// dynamicHostServicesIncludeHighRisk reports whether host services request data or network access.
func dynamicHostServicesIncludeHighRisk(items []*protocol.HostServiceSpec) bool {
	for _, item := range items {
		if item == nil {
			continue
		}
		switch item.Service {
		case protocol.HostServiceData, protocol.HostServiceNetwork:
			return true
		}
	}
	return false
}

// dynamicPackageReviewMessage returns a compact scanner message for release records.
func dynamicPackageReviewMessage(diagnostics []*PackageDiagnostic) string {
	if len(diagnostics) == 0 {
		return "Dynamic package scan completed."
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

// dynamicManifestResourceKind classifies embedded manifest resources.
func dynamicManifestResourceKind(resourcePath string) string {
	switch {
	case strings.HasPrefix(resourcePath, "manifest/docs/"):
		return "marketplace_doc"
	case strings.HasPrefix(resourcePath, "manifest/i18n/"):
		return "manifest_i18n"
	case strings.HasPrefix(resourcePath, "manifest/sql/uninstall/"):
		return "uninstall_sql"
	case strings.HasPrefix(resourcePath, "manifest/sql/mock-data/"):
		return "mock_sql"
	case strings.HasPrefix(resourcePath, "manifest/sql/"):
		return "install_sql"
	default:
		return "manifest_resource"
	}
}

// normalizeDynamicManifestResourcePath validates source-layout manifest resource paths.
func normalizeDynamicManifestResourcePath(resourcePath string) (string, error) {
	raw := strings.ReplaceAll(normalizeKey(resourcePath), "\\", "/")
	if raw == "" || raw == "." {
		return "", gerror.New("manifest resource path cannot be empty")
	}
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "/") {
		return "", gerror.New("manifest resource path must be relative")
	}
	if len(raw) >= 2 && raw[1] == ':' {
		return "", gerror.New("manifest resource path must not contain drive prefix")
	}
	normalized := path.Clean(raw)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.New("manifest resource path must not traverse parent directories")
	}
	if !strings.HasPrefix(normalized, "manifest/") {
		return "", gerror.New("manifest resource path must use manifest source layout")
	}
	return normalized, nil
}

// dynamicJSONContentIsObject reports whether content is valid JSON object text.
func dynamicJSONContentIsObject(content string) bool {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "{") {
		return false
	}
	var raw json.RawMessage
	return json.Unmarshal([]byte(trimmed), &raw) == nil
}

// dynamicBoolPointersEqual compares optional bool fields exactly.
func dynamicBoolPointersEqual(left *bool, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

// dynamicDependencySummariesEqual compares normalized dependency declarations.
func dynamicDependencySummariesEqual(left *sourcePackageManifest, right *sourcePackageManifest) bool {
	leftContent, leftErr := packageJSONString(buildSourceDependencySummary(left))
	rightContent, rightErr := packageJSONString(buildSourceDependencySummary(right))
	return leftErr == nil && rightErr == nil && leftContent == rightContent
}

// dynamicHostServicesEqual compares canonical normalized host service declarations.
func dynamicHostServicesEqual(left []*protocol.HostServiceSpec, right []*protocol.HostServiceSpec) bool {
	leftContent, leftErr := packageJSONString(left)
	rightContent, rightErr := packageJSONString(right)
	return leftErr == nil && rightErr == nil && leftContent == rightContent
}
