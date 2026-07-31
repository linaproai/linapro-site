## Why

插件市场 Git 文档快照当前使用 `docs-snapshot/<plugin>/<version>/content/<hash>.md` 布局：类型顶层目录使数据分散，hash 文件名导致运维无法识别文档内容；若按文件名存在即跳过同步，还会在同版本重同步时漏掉正文更新。需要在保留磁盘权威存储的前提下，改为按插件版本聚合、保留原文件名，并以内容 hash 判定是否覆盖。

## What Changes

- **BREAKING**：Git 文档磁盘快照 key 从 `docs-snapshot/.../content/<hash>.md` 改为 `<plugin-id>/<version>/docs/<locale>/<docPath>`；清单迁至 `<plugin-id>/<version>/meta/docs-manifest.json`。
- 同步时按路径落盘、按 content hash 决定是否 `Put`；同名文件 hash 不同必须覆盖。
- 远端已删除的文档对应本地文件必须清理，避免陈旧文档继续可读。
- `ArtifactStore` 增加 `Delete`，支撑孤儿清理。
- 更新插件 README 中的存储路径说明；无双读兼容（插件规范不考虑兼容性）。

## Capabilities

### New Capabilities

- `marketplace-docs-artifact-layout`：插件市场 Git 文档制品的路径布局、hash 增量同步与孤儿清理契约。

### Modified Capabilities

- （无基线已归档的 marketplace 文档存储需求；本变更新增 capability。活跃变更 `runtime-workspace-data-path-contract` 中对 `docs-snapshot/...` 的示例路径由本变更在实现与文档上取代。）

## Impact

- 代码：`linapro-plugin-marketplace` 的 `ArtifactStore`、Git 文档快照写入/读取、相关单元测试与 README。
- API：无 HTTP 契约变更；文档读接口仍从磁盘快照渲染。
- 运维：已有 `docs-snapshot` 数据需重新 Git 同步后生效；可不迁移旧文件。
- i18n / 数据权限 / DI：无影响。
