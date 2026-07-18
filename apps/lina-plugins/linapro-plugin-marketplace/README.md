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

## Git Platform Token Configuration

When a publisher registers a Git source without a personal access token, the marketplace falls back to plugin-scoped platform tokens for GitHub/Gitee metadata discovery (list tags, read tree, read `plugin.yaml`). This avoids unauthenticated API rate limits on shared egress IPs.

| Config key | Purpose |
|------------|---------|
| `github.accessToken` | GitHub PAT (classic `public_repo`, or fine-grained Contents: Read) |
| `gitee.accessToken` | Optional Gitee personal access token |

Config sources (exclusive priority, no merge):

1. Host `config.yaml` section `plugin.linapro-plugin-marketplace`
2. Production file `plugins/linapro-plugin-marketplace/config.yaml` under the host config root
3. Development default `manifest/config/config.yaml`

See `manifest/config/config.example.yaml` for the full template. Platform tokens are **not** stored in publisher credential rows and are never returned by marketplace APIs. A publisher-supplied form `accessToken` always wins over the platform config.

Host config example:

```yaml
plugin:
  linapro-plugin-marketplace:
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
