## Context

插件市场 Git 源在同步时把文档正文写入 `ArtifactStore`（默认 `temp/plugin-marketplace/artifacts`），读路径再按需渲染 Markdown。当前布局为：

```text
docs-snapshot/<pluginId>/<version>/
  manifest.json
  content/<contentHash>.md
```

包体则使用 `source|dynamic/<pluginId>/<version>/...`。文档与包体顶层约定不一致；hash 文件名不可读；同版本重同步时若误用「文件存在即跳过」会漏更新。

插件根 `AGENTS.md` 要求不考虑兼容性，允许直接切换 key 布局。

## Goals / Non-Goals

**Goals:**

- 文档按 `plugin → version → docs → locale → 原始相对路径` 聚合落盘。
- 同步时以 content hash 判定是否覆盖；路径相同且 hash 相同可跳过写盘。
- 远端集合不再包含的文档必须从本地删除。
- 保留版本级 manifest，记录 locale/docPath/sourceKind/contentHash/contentKey。

**Non-Goals:**

- 不重构 `source/`、`dynamic/` 包体 key（可后续独立收敛）。
- 不做旧 `docs-snapshot` 双读或自动迁移。
- 不改变 HTTP 文档 API 契约与 DB 表结构。
- 不把文档正文写回数据库。

## Decisions

### 1. 路径布局

```text
<plugin-id>/<version>/
  docs/
    <locale>/
      index.md
      guide/foo.md
  meta/
    docs-manifest.json
```

- `docPath` 使用已规范化的市场文档相对路径（禁止 `..`、绝对路径）。
- locale 单独一层，避免多语言同名覆盖。
- 根下不再使用 `docs-snapshot` 前缀，便于同一版本树扩展 `package/` 等资源。

**备选**：保留 `docs-snapshot` 顶层 — 否决，与「按插件聚合」目标冲突。

### 2. Hash 增量同步

对每个远端文档：

1. `contentKey = <plugin>/<version>/docs/<locale>/<docPath>`
2. `remoteHash = sha256(body)`（与现有 `ContentHash` 一致）
3. 若本地 Open 成功且本地 body hash == remoteHash → 跳过 Put
4. 否则 Put 覆盖
5. 写入 manifest 项时记录 `contentHash`，便于下次优先比对（也可再次读盘校验）

**备选**：始终全量 Put — 功能正确但多余 IO；仍采用 hash 跳过。

### 3. 孤儿清理

- 同步前加载旧 manifest 的 contentKey 集合（若存在）。
- 同步后对「旧集合 − 新集合」中的 key 调用 `Delete`。
- 对本地磁盘实现：若能解析版本 `docs/` 目录，再 walk 删除不在新集合中的残留文件（覆盖无 manifest 的脏数据）。

`ArtifactStore` 新增：

```go
Delete(ctx context.Context, key string) error
```

缺失对象视为成功（幂等删除）。

### 4. 读取路径

`loadDocumentIndexItemsFromGitSnapshot` 仍读 `meta/docs-manifest.json`，再按 `contentKey` Open 正文；渲染逻辑不变。

### 5. 与 runtime 路径契约的关系

`storage.root` 仍解析到 workspace 锚定的制品根；仅根下相对 key 从 `docs-snapshot/...` 变为 `<plugin>/<version>/docs/...`。

## Risks / Trade-offs

- [旧快照不可读] → 重新 Git 同步；README 说明无需迁移。
- [docPath 含危险段] → 继续走现有 `normalizeMarketplaceDocumentPath` / key 拒绝 `..`。
- [Delete 与测试 mock] → 所有 `ArtifactStore` 实现与 test double 补 `Delete`。
- [Walk 仅本地 store] → 孤儿 walk 在 LocalPath 可用时执行；纯内存 store 靠 manifest 差分 Delete 覆盖单测。

## Migration Plan

1. 部署新代码。
2. 触发 Git 元数据同步，重写文档快照。
3. 可选：手动删除历史 `docs-snapshot/` 目录释放磁盘。

## Open Questions

- 无。
