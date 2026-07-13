// This file defines marketplace service business error codes with English
// fallback text. Runtime translation resources are maintained under the
// plugin-owned manifest/i18n directories.

package marketplace

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeMarketplaceInvalidInput reports that required marketplace service input is missing or invalid.
	CodeMarketplaceInvalidInput = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_INVALID_INPUT",
		"Marketplace request input is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePublisherAlreadyExists reports a duplicate publisher key.
	CodeMarketplacePublisherAlreadyExists = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PUBLISHER_ALREADY_EXISTS",
		"Marketplace publisher already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePublisherNotFound reports that the requested publisher does not exist.
	CodeMarketplacePublisherNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PUBLISHER_NOT_FOUND",
		"Marketplace publisher does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplacePublisherUnavailable reports that a publisher cannot create or update drafts.
	CodeMarketplacePublisherUnavailable = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PUBLISHER_UNAVAILABLE",
		"Marketplace publisher is not available for publishing",
		gcode.CodeNotAuthorized,
	)
	// CodeMarketplacePluginNotFound reports that the requested marketplace plugin does not exist.
	CodeMarketplacePluginNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PLUGIN_NOT_FOUND",
		"Marketplace plugin does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplacePluginIDOwned reports that another publisher owns the plugin ID.
	CodeMarketplacePluginIDOwned = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PLUGIN_ID_OWNED",
		"Marketplace plugin ID is already owned by another publisher",
		gcode.CodeNotAuthorized,
	)
	// CodeMarketplaceReleaseNotFound reports that the requested release does not exist.
	CodeMarketplaceReleaseNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_RELEASE_NOT_FOUND",
		"Marketplace release does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplaceReleaseImmutable reports that a release can no longer be mutated.
	CodeMarketplaceReleaseImmutable = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_RELEASE_IMMUTABLE",
		"Marketplace release is immutable and cannot be changed",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceReleaseDraftExists reports that a draft exists and replacement was not allowed.
	CodeMarketplaceReleaseDraftExists = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_RELEASE_DRAFT_EXISTS",
		"Marketplace release draft already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePackageInvalid reports that an uploaded marketplace package cannot be parsed.
	CodeMarketplacePackageInvalid = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PACKAGE_INVALID",
		"Marketplace package is invalid: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePackageStructureInvalid reports that a source package is missing required files or directories.
	CodeMarketplacePackageStructureInvalid = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PACKAGE_STRUCTURE_INVALID",
		"Marketplace package structure is invalid: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePackageManifestMismatch reports that package plugin.yaml does not match upload fields.
	CodeMarketplacePackageManifestMismatch = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PACKAGE_MANIFEST_MISMATCH",
		"Marketplace package manifest does not match upload fields: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplacePackageScanFailed reports an unexpected scanner failure.
	CodeMarketplacePackageScanFailed = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_PACKAGE_SCAN_FAILED",
		"Marketplace package scan failed",
		gcode.CodeInternalError,
	)
	// CodeMarketplaceDocumentInvalid reports that marketplace documentation content is unsafe or malformed.
	CodeMarketplaceDocumentInvalid = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOCUMENT_INVALID",
		"Marketplace package documentation is invalid: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceDocumentNotFound reports that no indexed document matches the requested fallback chain.
	CodeMarketplaceDocumentNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOCUMENT_NOT_FOUND",
		"Marketplace release documentation does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplaceDownloadArtifactNotFound reports that a release artifact is unavailable for download.
	CodeMarketplaceDownloadArtifactNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOWNLOAD_ARTIFACT_NOT_FOUND",
		"Marketplace download artifact does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplaceDownloadSessionNotFound reports that a download session cannot be used by the caller.
	CodeMarketplaceDownloadSessionNotFound = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOWNLOAD_SESSION_NOT_FOUND",
		"Marketplace download session does not exist",
		gcode.CodeNotFound,
	)
	// CodeMarketplaceDownloadSessionExpired reports that the download session is no longer active.
	CodeMarketplaceDownloadSessionExpired = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOWNLOAD_SESSION_EXPIRED",
		"Marketplace download session has expired",
		gcode.CodeNotAuthorized,
	)
	// CodeMarketplaceDownloadSessionUnavailable reports that the download session state rejects the action.
	CodeMarketplaceDownloadSessionUnavailable = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_DOWNLOAD_SESSION_UNAVAILABLE",
		"Marketplace download session is not available",
		gcode.CodeNotAuthorized,
	)
	// CodeMarketplaceReviewStateInvalid reports an invalid review state transition.
	CodeMarketplaceReviewStateInvalid = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_REVIEW_STATE_INVALID",
		"Marketplace release review state does not allow this transition",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceStatusInvalid reports an unsupported marketplace status transition.
	CodeMarketplaceStatusInvalid = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_STATUS_INVALID",
		"Marketplace plugin status does not allow this transition",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceStorageFailed reports a lower-level database failure.
	CodeMarketplaceStorageFailed = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_STORAGE_FAILED",
		"Marketplace storage operation failed",
		gcode.CodeInternalError,
	)
	// CodeMarketplaceSourceKindConflict reports mixed git/upload publish sources.
	CodeMarketplaceSourceKindConflict = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_SOURCE_KIND_CONFLICT",
		"Marketplace plugin publish source kind does not allow this operation",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceGitDiscoveryFailed reports remote Git metadata discovery failure.
	CodeMarketplaceGitDiscoveryFailed = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_GIT_DISCOVERY_FAILED",
		"Marketplace Git metadata discovery failed: {diagnostic}",
		gcode.CodeInternalError,
	)
	// CodeMarketplaceGitAuthFailed reports private repository authentication failure.
	CodeMarketplaceGitAuthFailed = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_GIT_AUTH_FAILED",
		"Marketplace Git authentication failed: {diagnostic}",
		gcode.CodeNotAuthorized,
	)
	// CodeMarketplaceGitVersionMismatch reports tag and plugin.yaml version mismatch.
	CodeMarketplaceGitVersionMismatch = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_GIT_VERSION_MISMATCH",
		"Marketplace Git tag does not match plugin.yaml version: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
	// CodeMarketplaceGitDynamicUnsupported reports dynamic plugins on Git sources.
	CodeMarketplaceGitDynamicUnsupported = bizerr.MustDefine(
		"PLUGIN_MARKETPLACE_GIT_DYNAMIC_UNSUPPORTED",
		"Marketplace Git sources support source plugins only: {diagnostic}",
		gcode.CodeInvalidParameter,
	)
)
