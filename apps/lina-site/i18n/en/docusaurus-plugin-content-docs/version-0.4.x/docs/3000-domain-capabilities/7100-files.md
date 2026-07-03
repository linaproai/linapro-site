---
slug: '/docs/domain-capability-files'
title: 'Files'
hide_title: true
description: 'The Files capability is the governed entry for source plugins and dynamic plugins to read the host file management system. The host manages the full lifecycle of user-uploaded files through the `sys_file` table. Plugins obtain read-only file views through `services.Files()` and perform controlled deletion through `Admin().Files()`. The capability does not expose physical storage paths, file hashes, or underlying storage backends to plugins, providing only `FileProjection` as a safe domain view.'
keywords:
  - Files capability
  - filecap
  - FileProjection
  - BatchGetFiles
  - EnsureFilesVisible
  - DeleteFiles
  - sys_file
  - host file management
  - file view
  - file upload
  - business scene
  - file deletion
  - hostServices
  - host:files
  - plugin capability
  - LinaPro
---

## Introduction

Source plugins read host-managed file views through `services.Files()`. Dynamic plugins declare `service: files` in `plugin.yaml` and use `BatchGetFiles` and `EnsureFilesVisible`. Trusted source plugins that need to delete host files use `services.Admin().Files().DeleteFiles`.

The Files capability provides read-only access to the host file management system. Plugins cannot upload files or modify file metadata through this capability.

**Capability Phase**: Runtime

**Supported Plugin Types**: Source plugins, Dynamic plugins

## Capability Design

### Host File Management Model

The host file management is centered on the `sys_file` table, managing the full lifecycle of user-uploaded files:

| Field | Description |
|-------|-------------|
| `id` | File primary key |
| `tenant_id` | Tenant ownership |
| `name` | Storage file name |
| `original` | Original file name |
| `suffix` | File extension |
| `scene` | Business scene identifier |
| `size` | File size |
| `hash` | SHA-256 hash for deduplication |
| `url` | Access path |
| `path` | Physical storage path |
| `engine` | Storage engine identifier |
| `created_by` | Uploader |

The upload process automatically computes SHA-256 hashes for file deduplication within the same tenant: files with identical hashes reuse physical storage, creating only new metadata records.

### Plugin-Visible View

Plugins access file information through `FileProjection`, which does not expose physical storage paths, hash values, or underlying storage backends:

| Field | Description |
|-------|-------------|
| `ID` | File identifier |
| `Name` | Display file name |
| `MimeType` | Media type |
| `SizeBytes` | File size |
| `BusinessScene` | Business scene |

```mermaid
graph TB
    Source["Source Plugin"] --> Files["services.Files()"]
    Dynamic["Dynamic Plugin"] --> Files
    Files --> Projection["FileProjection read-only view"]
    Admin["Trusted Source Plugin"] --> AdminFiles["services.Admin().Files()"]
    AdminFiles --> Delete["DeleteFiles controlled deletion"]
    Projection --> SysFile["sys_file table"]
    Delete --> SysFile
```

### Relationship with Other Resource Capabilities

| Capability | Purpose |
|------------|---------|
| `Files()` | Read host file management views, e.g., user-uploaded files, business attachments |
| `Storage()` | Plugin's own object read/write, e.g., export results, temporary artifacts |
| `Manifest()` | Read plugin's read-only `manifest/` resources shipped with the artifact |
| `AI AssetRef` | Reference protected input or output assets in the `AI` capability |

## Interface Definitions

### Source Plugin Interface

| Entry | Method | Description |
|-------|--------|-------------|
| `Files()` | `BatchGet` | Batch-reads visible file views |
| `Files()` | `Search` | Searches visible file candidates by business scene, keyword, and media type |
| `Files()` | `EnsureVisible` | Validates that target file set is visible to the current call context |
| `Admin().Files()` | `Delete` | Deletes visible files; host performs scene and target validation |

### Dynamic Plugin Interface

Dynamic plugins can access two types of services:

**`files` service — Host file views:**

| Dynamic Method | Capability Constant | Description |
|----------------|---------------------|-------------|
| `files.batch_get` | `host:files` | Batch-reads visible file views |
| `files.search` | `host:files` | Searches visible file candidates by business scene, keyword, and media type |
| `files.visible.ensure` | `host:files` | Validates file visibility |

**`storage` service — Plugin-scoped object storage:**

| Dynamic Method | Description |
|----------------|-------------|
| `storage.put` | Write object |
| `storage.get` | Read object |
| `storage.delete` | Delete object |
| `storage.delete_many` | Batch-delete objects |
| `storage.list` | List objects |
| `storage.list_cursor` | Cursor-paginated object listing |
| `storage.stat` | Query object metadata |
| `storage.batch_stat` | Batch-query object metadata |

## Capability Usage

### Source Plugin Usage

Source plugins read host-managed file views through `services.Files()`:

```go
// Batch-read file views
result, err := services.Files().BatchGet(ctx, capabilityCtx, fileIDs)

// Search visible file candidates
page, err := services.Files().Search(ctx, capabilityCtx, filecap.SearchInput{
    BusinessScene: "avatar",
    Keyword:       "profile",
    MimeType:      "image/png",
    Page:          pageRequest,
})

// Validate file visibility
err := services.Files().EnsureVisible(ctx, capabilityCtx, fileIDs)
```

Trusted source plugins deleting files:

```go
err := services.Admin().Files().Delete(ctx, capabilityCtx, fileIDs)
```

### Dynamic Plugin Usage

Dynamic plugins declare the `files` service in `plugin.yaml`:

```yaml
hostServices:
  - service: files
    methods:
      - files.batch_get
      - files.search
      - files.visible.ensure
```

## Design Constraints

- **Read-only view.** Plugins cannot upload or modify files through the Files capability — only read and validate.
- **No physical paths exposed.** `FileProjection` does not contain storage paths, hash values, or access URLs.
- **Deletion is a governance command.** Delete operations must go through `Admin().Files()`, with the host performing scene and target validation.
- **Visibility is controlled by the host.** Whether a file is visible to the current plugin is determined by the host based on tenant and data scope.

## Related Services

- [Storage Capability](/docs/domain-capability-storage)
- [Manifest Resources Capability](/docs/domain-capability-manifest)
- [AI Capability](/docs/domain-capability-ai)
