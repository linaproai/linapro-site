## ADDED Requirements

### Requirement: 市场展示名称与摘要按版本与语言落库

系统 SHALL 为每个市场发布版本持久化按语言划分的展示元数据，至少包含 `name` 与 `summary`。展示元数据 MUST 绑定 `release_id` 与 `locale`，并 MUST NOT 将文档 Markdown、渲染 HTML 或图片二进制作为本能力的存储内容。

#### Scenario: 扫描含 i18n 的插件包写入多语言展示行

- **WHEN** 发布者上传或同步的插件包在 `manifest/i18n/<locale>/` 中提供 `plugin.<plugin-id>.name` 与 `plugin.<plugin-id>.description`
- **THEN** 系统为该草稿或可变版本的每个相关 locale 写入或替换展示元数据行
- **AND** `description` 键映射为市场列表使用的 `summary`（长度受摘要字段上限约束）

#### Scenario: 无 i18n 资源时写入默认语言回退行

- **WHEN** 插件包未提供运行时展示键
- **THEN** 系统至少根据 `plugin.yaml` 的 `name` 与 `description` 写入默认 locale（`i18n.default` 或 `en-US`）一行展示元数据
- **AND** 身份表 `name`/`summary` 仍可作为最终 fallback

### Requirement: 列表、审核队列与详情按请求语言投影展示字段

系统 SHALL 在市场目录列表、「我的插件」列表、管理列表、审核队列与插件详情中，按当前请求语言返回已本地化的展示字段。列表与详情 MUST 本地化 `name` 与 `summary`，审核队列 MUST 本地化 `pluginName`。回退顺序 MUST 为：请求 locale → 插件声明的默认 locale → `en-US` → 身份表 fallback 字段。

#### Scenario: 中文请求返回中文名称与摘要

- **WHEN** 操作者在 `zh-CN` 语言下请求「我的插件」列表
- **AND** 目标版本存在 `zh-CN` 展示元数据
- **THEN** 响应中的 `name` 与 `summary` 为中文展示值而非仅 `plugin.yaml` 英文原文

#### Scenario: 缺少请求语言时回退

- **WHEN** 请求 locale 为 `zh-CN` 但该版本仅有 `en-US` 展示行
- **THEN** 系统返回 `en-US` 展示值
- **AND** 若展示表无任何行，则返回身份表 `name`/`summary`

#### Scenario: 中文审核队列返回中文插件名

- **WHEN** 管理者在 `zh-CN` 语言下请求插件审核队列
- **AND** 待审版本存在 `zh-CN` 展示元数据
- **THEN** 响应中的 `pluginName` 为中文展示值
- **AND** 系统对当前分页版本批量装配展示元数据，不得逐行查询

### Requirement: 文档与图片不以数据表存储正文或二进制

系统 SHALL 以版本制品磁盘（或等价 ArtifactStore）作为文档文件与图片二进制的权威存储。系统 MUST NOT 使用 `plugin_marketplace_doc` 或等价表持久化文档全文、渲染 HTML 或图片二进制。文档读取 MUST 从包 ZIP 或 Git 文档磁盘快照加载并按需渲染。

#### Scenario: 展示元数据表不包含文档正文

- **WHEN** 系统写入市场展示元数据
- **THEN** 写入内容仅包含短文本展示字段（名称、摘要及溯源元数据）
- **AND** 不写入 Markdown 全文、渲染 HTML 大字段或图片内容

#### Scenario: 上传包文档从制品包读取

- **WHEN** 操作者请求某已上传版本的文档
- **THEN** 系统从该版本主包 ZIP 中的 `manifest/docs` 或 README 读取 Markdown 并安全渲染
- **AND** 不查询文档正文数据表

#### Scenario: Git 源文档从磁盘快照读取

- **WHEN** Git 同步完成文档 enrichment
- **THEN** 系统将文档正文写入 ArtifactStore 下的 docs-snapshot
- **AND** 后续文档 API 从该快照读取并渲染

### Requirement: 工作台切换语言后刷新市场列表展示

默认管理工作台中的市场「我的插件」「插件列表」与「插件审核」页 MUST 在当前语言变更后重新请求对应列表接口，以使名称、筛选项、列标题与新的请求语言一致。

#### Scenario: 切换到英文后列表展示英文名称

- **WHEN** 操作者在「我的插件」页从 `zh-CN` 切换到 `en-US`
- **AND** 目标版本存在英文展示元数据
- **THEN** 列表中的名称与摘要在不手动整页刷新的情况下更新为英文展示值

#### Scenario: 审核页切换语言后刷新队列名称

- **WHEN** 管理者在「插件审核」页从 `en-US` 切换到 `zh-CN`
- **AND** 待审版本存在中文展示元数据
- **THEN** 队列中的插件名称、筛选项与列标题在不手动整页刷新的情况下更新为中文
