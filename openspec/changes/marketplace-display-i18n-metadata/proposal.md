## Why

插件市场「我的插件」与目录列表中的名称、摘要当前写入单语业务字段，切换工作台语言后不会本地化。文档与图片已随版本包落在制品磁盘，列表高频需要的是短展示元数据，不应把文档正文或图片二进制再堆进数据表。

## What Changes

- 新增**按版本 × 语言**的市场展示元数据表，仅存储插件 `name` / `summary`（及必要溯源字段），供列表、详情、审核队列等高频路径按请求语言投影。
- 包扫描（源码包 / 动态包 / 自动建身份）时从包内 `plugin.<id>.name`、`plugin.<id>.description` 与 `plugin.yaml` 写入各 locale 展示快照；文档 Markdown、图片只存在版本制品磁盘中。
- **删除**遗留表 `plugin_marketplace_doc` 及其 DAO/写入路径；文档读路径改为从包 ZIP 或 Git 文档磁盘快照按需渲染。
- 列表与详情 API 按请求 locale 回退选择展示名称与摘要：请求语言 → 插件 `i18n.default` → `en-US` → 身份表 fallback 字段。
- 市场前端（至少「我的插件」、管理列表）在切换语言后重新请求列表，避免沿用旧语言响应。

## Capabilities

### New Capabilities

- `marketplace-display-i18n`: 市场插件展示名称与摘要的多语言数据表存储、版本绑定、扫描写入与按请求语言投影。

### Modified Capabilities

- （无独立基线能力需改名；市场能力以本变更新增规范承载。若归档时并入 `plugin-marketplace` 基线，再同步英文基线。）

## Impact

- 代码：`apps/lina-plugins/linapro-plugin-marketplace/` 后端 SQL/DAO、包扫描、列表/详情投影、控制器 locale 解析；前端 mine/catalog 列表语言切换重拉。
- 数据：新增展示元数据表；不新增文档/图片内容表。
- API：列表/详情响应中的 `name`/`summary` 语义变为「当前请求语言下的展示值」；契约字段名不变。
- `i18n`：市场插件自身 UI 文案无强制新增；被发布插件的展示键约定与宿主 `plugin.<id>.*` 对齐。
- 测试：后端单元测试覆盖扫描写入与 locale 回退；前端或 E2E 覆盖「我的插件」语言切换后名称/摘要变化。
