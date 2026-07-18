// This file provides shared archive detection and safe tar.gz-to-zip conversion
// so existing ZIP package scanners can validate zip and tar.gz uploads.

package marketplace

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"strings"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
)

const (
	archiveKindZip    = "zip"
	archiveKindTarGz  = "tar.gz"
	maxArchiveEntries = 20000
	maxArchiveBytes   = 512 * 1024 * 1024 // 512 MiB uncompressed budget
)

// detectArchiveKind returns zip, tar.gz, or empty when the name is unsupported.
func detectArchiveKind(fileName string) string {
	lower := strings.ToLower(strings.TrimSpace(fileName))
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return archiveKindTarGz
	case strings.HasSuffix(lower, ".zip"):
		return archiveKindZip
	default:
		return ""
	}
}

// materializeZipPackagePath returns a path suitable for zip.OpenReader. When the
// upload is already a zip, it returns the original path. When the upload is
// tar.gz, it converts entries into a temporary zip file that the caller must
// delete via cleanup.
func materializeZipPackagePath(packagePath string, fileName string) (zipPath string, cleanup func(), err error) {
	kind := detectArchiveKind(fileName)
	if kind == "" {
		kind = detectArchiveKind(packagePath)
	}
	switch kind {
	case archiveKindZip, "":
		return packagePath, func() {}, nil
	case archiveKindTarGz:
		return convertTarGzToTempZip(packagePath)
	default:
		return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package must be .zip or .tar.gz")
	}
}

// convertTarGzToTempZip expands a tar.gz into a temporary zip with the same
// relative paths, enforcing path safety and size limits.
func convertTarGzToTempZip(packagePath string) (zipPath string, cleanup func(), err error) {
	src, err := os.Open(packagePath)
	if err != nil {
		return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package file cannot be read")
	}
	defer src.Close()

	gzReader, err := gzip.NewReader(src)
	if err != nil {
		return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package must be a valid gzip container")
	}
	defer gzReader.Close()

	tempFile, err := os.CreateTemp("", "marketplace-package-*.zip")
	if err != nil {
		return "", nil, packageDiagnosticError(CodeMarketplacePackageScanFailed, "temporary package file cannot be created")
	}
	zipPath = tempFile.Name()
	cleanup = func() {
		_ = os.Remove(zipPath)
	}
	success := false
	defer func() {
		if !success {
			_ = tempFile.Close()
			cleanup()
			cleanup = nil
		}
	}()

	zipWriter := zip.NewWriter(tempFile)
	tarReader := tar.NewReader(gzReader)
	var (
		entries int
		total   int64
	)
	for {
		header, readErr := tarReader.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package must be a valid tar.gz container")
		}
		if header == nil || header.Typeflag != tar.TypeReg {
			continue
		}
		normalized, skip, normErr := normalizeZipEntryName(header.Name)
		if normErr != nil {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, normErr.Error())
		}
		if skip {
			continue
		}
		entries++
		if entries > maxArchiveEntries {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package contains too many files")
		}
		if header.Size > 0 {
			total += header.Size
			if total > maxArchiveBytes {
				return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package uncompressed size exceeds limit")
			}
		}
		writer, createErr := zipWriter.Create(normalized)
		if createErr != nil {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageScanFailed, "temporary package entry cannot be written")
		}
		written, copyErr := io.Copy(writer, io.LimitReader(tarReader, maxArchiveBytes+1))
		if copyErr != nil {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package entry cannot be read")
		}
		if written > maxArchiveBytes {
			return "", nil, packageDiagnosticError(CodeMarketplacePackageInvalid, "package uncompressed size exceeds limit")
		}
	}
	if err = zipWriter.Close(); err != nil {
		return "", nil, packageDiagnosticError(CodeMarketplacePackageScanFailed, "temporary package cannot be finalized")
	}
	if err = tempFile.Close(); err != nil {
		return "", nil, packageDiagnosticError(CodeMarketplacePackageScanFailed, "temporary package cannot be finalized")
	}
	success = true
	return zipPath, cleanup, nil
}

// packageContentTypeForName returns a MIME type for known archive names.
func packageContentTypeForName(fileName string, fallback string) string {
	switch detectArchiveKind(fileName) {
	case archiveKindTarGz:
		return "application/gzip"
	case archiveKindZip:
		return "application/zip"
	default:
		if strings.TrimSpace(fallback) != "" {
			return fallback
		}
		return "application/zip"
	}
}

// sourceArtifactTypeForName maps an archive name to a marketplace artifact type.
func sourceArtifactTypeForName(fileName string) marketv1.MarketplaceArtifactType {
	if detectArchiveKind(fileName) == archiveKindTarGz {
		return marketv1.MarketplaceArtifactTypeSourceTarGz
	}
	return marketv1.MarketplaceArtifactTypeSourceZip
}

// dynamicArtifactTypeForName maps an archive name to a marketplace artifact type.
func dynamicArtifactTypeForName(fileName string) marketv1.MarketplaceArtifactType {
	if detectArchiveKind(fileName) == archiveKindTarGz {
		return marketv1.MarketplaceArtifactTypeDynamicTarGz
	}
	return marketv1.MarketplaceArtifactTypeDynamicZip
}

// storageKeyExtension returns the preferred storage object suffix for an archive.
func storageKeyExtension(fileName string) string {
	switch detectArchiveKind(fileName) {
	case archiveKindTarGz:
		return ".tar.gz"
	default:
		return ".zip"
	}
}

// ensurePackageArchiveSupported rejects unsupported archive file names.
func ensurePackageArchiveSupported(fileName string) error {
	if normalizeKey(fileName) == "" {
		return nil
	}
	if detectArchiveKind(fileName) == "" {
		return packageDiagnosticError(CodeMarketplacePackageInvalid, "package file name must end with .zip, .tar.gz, or .tgz")
	}
	return nil
}
