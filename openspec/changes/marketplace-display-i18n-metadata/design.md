## Context

市场身份表 `plugin_marketplace_plugin.name/summary` 与读模型同为单语字段；包扫描用 `plugin.yaml` 的 `name`/`description` 写入一次。文档已有 `plugin_marketplace_doc` 文本快照与制品磁盘上的包文件，图片只在包内。用户明确要求：

1. 多语言**名称与摘要**用数据表存储（高频展示）。
2. **文档正文与图片**前期继续磁盘（版本制品），不为文档内容加新表。

## Goals / Non-Goals

**Goals**

- 按 `release_id + locale` 持久化市场展示 `name`/`summary`。
- 扫描时从包内运行时 i18n 键与 `plugin.yaml` 填充快照。
- 列表/详情按请求语言投影；前端切语言重拉。
- 身份表 `name`/`summary` 保留为 fallback / 默认语言投影，便于搜索与无 i18n 插件。

**Non-Goals**

- 不为文档 Markdown、渲染 HTML、图片建立新的内容表或 BLOB 表。
- 不做机器翻译；只使用包内或发布者已提供的文案。
- 不改变插件安装后宿主「插件管理」的 `plugin.<id>.*` 运行时投影机制。
- 不提供历史 SQL 迁移/兼容分支；安装 SQL 仅维护最终态 `001`。

## Decisions

### 1. 展示元数据表（仅短字段）

表名沿用市场既有前缀：`plugin_marketplace_display_i18n`。

| 字段 | 说明 |
|------|------|
| `release_id` | 绑定版本（必填） |
| `plugin_id` / `release_version` | 冗余便于排查与批量查询 |
| `locale` | 如 `en-US`、`zh-CN` |
| `name` | 展示名称 |
| `summary` | 列表摘要（可由 description 键映射） |
| `source` | `package_i18n` / `plugin_yaml` / `publisher` |
| 软删除 | `deleted_at`（与市场其他表一致） |

唯一约束：`(release_id, locale)` 在 `deleted_at IS NULL` 时唯一。

### 2. 写入来源与优先级

对每个扫描到的 locale：

1. 若存在 `plugin.<pluginId>.name` → name；`plugin.<pluginId>.description` → summary（截断至 512）。
2. 否则用 `plugin.yaml` 的 name/description 写入**默认 locale**（`i18n.default`，缺省 `en-US`）一行。
3. 发布者表单手填时更新身份表 fallback，并可写入/覆盖 `source=publisher` 的默认 locale 行（若表单未提供多语言，仅默认语言）。

**上传包**：从 ZIP/动态包内 `manifest/i18n/<locale>/*.json`（排除 `apidoc`）解析扁平/嵌套键并合并。

**Git 同步**：在发现草稿版本时，根据 remote tree 选出运行时 i18n JSON 路径，经 `ReadFile` 拉取正文（有文件数/体积上限），同样解析并合并后写入 `display_i18n`。不得仅依赖 `plugin.yaml` 单语字段，否则 monorepo（如 `official-plugins`）同步后列表无法随工作台语言切换名称/摘要。

替换可变草稿 release 时，先软删或替换该 `release_id` 下全部 display 行再写入，保证与包一致。
### 3. 读路径回退

选定展示版本（目录：最新已发布；我的插件：身份上的 latest 或最新草稿版本）后：

```text
请求 locale → 插件 i18n.default → en-US → plugin 表 name/summary
```

批量列表：一次 `WHERE release_id IN (...)` 拉取候选 locale 行，内存选优，避免 N+1。

### 4. 文档与图片（磁盘权威，删除 doc 表）

- 权威：版本制品（`plugin_marketplace_artifact` + ArtifactStore 磁盘）。
- **删除** `plugin_marketplace_doc` 表及全部 DAO/entity/do 与 `replaceReleaseDocuments` 写入路径。
- 读文档：上传包从主包 ZIP 解压 Markdown 渲染；Git 源在同步时写入 `docs-snapshot/<pluginId>/<version>/` 磁盘快照，读取时再渲染。
- 图片仍随包内 `manifest/docs/**` 路径校验；不入库。

### 5. 前端

- 「我的插件」、市场目录列表：`watch(preferences.app.locale)` 后重新 `query`。
- 名称/摘要仍直接渲染 API 值，前端不做业务文案 `$t` 映射。

## Risks / Trade-offs

- **无包 i18n 的插件**：仅有默认 locale 一行 + fallback，切语言可能不变——符合预期。
- **表前缀**：市场历史表使用 `plugin_marketplace_*`；本表与之一致，避免混用两套前缀。
- **读模型**：`plugin_marketplace_plugin_read_model` 仍可存默认语言快照供搜索；展示列以 display_i18n 投影为准。
- **性能**：列表批量查 display 表；索引 `(release_id, locale)` / `(plugin_id, release_version)`。

## Migration Plan

1. 新增 SQL 迁移建表。
2. 部署后新上传/同步的包写入 display 行；旧数据仅有 fallback 字段，行为与现网一致直至重新扫包或发新版本。
3. 可选后续任务：对已有 release 后台回填（非本变更必须）。

## Open Questions

- 无（用户已确认：元数据进表，文档/图片磁盘）。
