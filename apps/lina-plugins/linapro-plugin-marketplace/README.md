# LinaPro Plugin Marketplace

`linapro-plugin-marketplace` is the built-in source plugin for private plugin marketplace workflows in `LinaPro`.

The plugin owns marketplace-specific backend code, frontend resources, SQL, runtime i18n resources, API-documentation i18n resources, marketplace display documents, and plugin-owned tests under this directory. It does not change `lina-core` plugin installation, enablement, dynamic runtime validation, or source plugin rebuild boundaries.

## Current Scope

| Area | Boundary |
|------|----------|
| Distribution | `type: source` with `distribution: builtin` |
| Runtime language | `i18n.enabled: true` with `en-US` source content and `zh-CN` translations |
| Source delivery | Downloaded source plugins must be placed under `apps/lina-plugins/<plugin-id>` and deployed through a host rebuild |
| Dynamic delivery | Downloaded dynamic packages must reuse the existing local dynamic plugin upload governance |

## Local Storage Path Configuration

Uploaded packages, Git documentation snapshots, and controlled download bytes are stored on the local filesystem under `storage.root`.

| Config key | Purpose | Default |
|------------|---------|---------|
| `storage.root` | Artifact and documentation snapshot root | `temp/plugin-marketplace/artifacts` |

Relative paths resolve against the host **workspace root** via the shared `lina-core/pkg/runtimepath` contract (same anchoring as upload and dynamic-plugin storage). They do **not** depend on the process working directory. `make dev` injects `LINAPRO_WORKSPACE_ROOT` (repo root) and `LINAPRO_DATA_ROOT` (`<repo>/temp`), so the default location is:

```text
<repo>/temp/plugin-marketplace/artifacts/
  <plugin-id>/
    <version>/
      docs/
        <locale>/
          index.md
          ...
      meta/
        docs-manifest.json
```

Git documentation snapshots keep original relative file names under `docs/<locale>/`. On each sync the service compares content hashes and overwrites only when the body changes; documents removed upstream are deleted locally. Package bytes still use `source/` and `dynamic/` key prefixes under the same root.

not `apps/lina-core/temp/...`. Absolute configured paths are used as-is. Production should use an absolute path outside the repository.

If you still have snapshots under `apps/lina-core/temp/plugin-marketplace/` or the legacy `docs-snapshot/.../content/<hash>.md` layout, re-run Git metadata sync (or delete the old tree). Dual-read compatibility is not provided.

## Git Platform Token Configuration

When a publisher registers a Git source without a personal access token, the marketplace falls back to plugin-scoped platform tokens for GitHub/Gitee metadata discovery (list tags, read tree, read `plugin.yaml`). This avoids unauthenticated API rate limits on shared egress IPs.

| Config key | Purpose |
|------------|---------|
| `github.accessToken` | GitHub PAT (classic `public_repo`, or fine-grained Contents: Read) |
| `gitee.accessToken` | Optional Gitee personal access token |

These settings are **owned by this plugin** and must not be placed in the host framework `config.yaml`. Sources (exclusive priority, no merge):

1. Development default: `apps/lina-plugins/linapro-plugin-marketplace/manifest/config/config.yaml`
2. Production: `plugins/linapro-plugin-marketplace/config.yaml` under the host config root

See `manifest/config/config.example.yaml` for the full template. Platform tokens are **not** stored in publisher credential rows and are never returned by marketplace APIs. A publisher-supplied form `accessToken` always wins over the platform config.

Plugin config example (`manifest/config/config.yaml` or the production plugin config file):

```yaml
storage:
  root: "temp/plugin-marketplace/artifacts"
github:
  accessToken: "ghp_xxxxxxxx"
gitee:
  accessToken: ""
```

## Verification

Run the plugin package smoke test from this directory:

```bash
go test ./... -count=1
```
