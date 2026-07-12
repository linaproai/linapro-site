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

## Verification

Run the plugin package smoke test from this directory:

```bash
go test ./... -count=1
```
