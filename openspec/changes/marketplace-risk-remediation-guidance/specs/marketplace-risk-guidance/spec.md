## ADDED Requirements

### Requirement: Risk findings carry disposition and blocking semantics

系统 SHALL 为每条市场扫描 risk finding 派生稳定处置分类与是否阻塞提交语义。处置分类 MUST 为以下之一：`need_fix`（需修复）、`need_attention`（需说明/审核关注）、`info_only`（仅提示）。系统 MUST 将 `disposition` 与 `blocking` 投影到风险列表 API 响应中。派生 MUST 基于稳定扫描 `code` 的策略表，不得依赖发布者自报。

#### Scenario: Missing framework dependency is attention-only and non-blocking

- **WHEN** 包扫描产生 code 为 `framework_dependency_missing` 的 finding
- **THEN** 风险列表项的 `disposition` MUST 为 `need_attention` 且 `blocking` MUST 为 `false`

#### Scenario: SQL presence is attention-only and non-blocking

- **WHEN** 包扫描产生 code 为 `source_sql_present` 或 `dynamic_sql_present` 的 finding
- **THEN** 风险列表项的 `disposition` MUST 为 `need_attention` 且 `blocking` MUST 为 `false`

#### Scenario: Documentation index is informational

- **WHEN** 包扫描产生 code 为 `source_docs_indexed` 或 `dynamic_runtime_detected` 的 finding
- **THEN** 风险列表项的 `disposition` MUST 为 `info_only` 且 `blocking` MUST 为 `false`

### Requirement: Risk payload includes bounded evidence

系统 SHALL 在写入 release 风险行时，将结构化证据放入 finding `payload`（JSON）。payload MUST 包含稳定 `code`，并在可得时包含有界证据字段（如 `files`、`services`、`routes`、`expectedPath`、`expectedField`、`example`）。列表类证据 MUST 截断到固定上限，并在截断时标记 `truncated` 与总数。

#### Scenario: SQL finding lists package SQL paths

- **WHEN** 源码包扫描检测到 `manifest/sql` 下的 SQL 资源并产生 `source_sql_present`
- **THEN** 对应风险项 payload MUST 包含 `files` 数组，元素为包内 SQL 相对路径（不超过上限）

#### Scenario: Host service finding lists requested services

- **WHEN** 动态包扫描检测到 host services 并产生 `dynamic_host_services_present`
- **THEN** 对应风险项 payload MUST 包含 `services` 摘要列表（服务名及可选 methods/tables/paths）

#### Scenario: Legacy risk rows without evidence still resolve disposition

- **WHEN** 历史风险行 payload 仅含 `code` 且无证据字段
- **THEN** 风险列表 API MUST 仍能根据 `code` 返回 `disposition` 与 `blocking`，且不得因缺少证据而失败

### Requirement: Risk UI presents remediation guidance and evidence

管理工作台插件详情与审核检查中的风险列表 MUST 展示处置标签，并支持展开每条 finding 的结构化指引：标题、原因/影响、建议操作、验收标准。指引文案 MUST 通过插件 i18n 按 `code` 本地化。当 payload 含证据时，UI MUST 展示文件路径、宿主服务或路由等证据列表。列表默认排序 MUST 优先 blocking 与 `need_fix`，再按严重级别。

#### Scenario: Publisher opens risks tab on my plugin detail

- **WHEN** 发布者在「我的插件」详情选择版本并打开风险 Tab
- **THEN** 系统 MUST 显示各 finding 的处置标签与可展开的建议操作，且 `need_fix` 项排在 `info_only` 之前

#### Scenario: User expands a risk finding with files evidence

- **WHEN** 用户展开带 `files` 证据的风险项
- **THEN** UI MUST 列出证据路径，并显示该 code 的原因与建议操作文案

#### Scenario: Empty risks still shows completion empty state

- **WHEN** 所选版本扫描完成且无 finding
- **THEN** UI MUST 显示既有「未发现风险条目」空态，不得展示错误

### Requirement: Blocking findings gate review submission

系统 SHALL 在发布者提交版本审核时检查该版本是否存在 `blocking=true` 的 risk finding。若存在，系统 MUST 拒绝提交并返回可识别的业务错误；前端 MUST 提示用户先处理阻塞项。非 blocking 的 `need_attention` finding MUST NOT 单独阻止提交。

#### Scenario: Submit blocked when fixable findings remain

- **WHEN** 发布者对仍含 `i18n_files_missing` 或 `dynamic_manifest_resources_missing` 的版本调用提交审核
- **THEN** 系统 MUST 拒绝提交且不改变审核状态

#### Scenario: Submit allowed with non-blocking attention findings

- **WHEN** 发布者对仅含 `source_sql_present` 或 `framework_dependency_missing`（无 blocking 项）的版本提交审核
- **THEN** 系统 MUST 允许进入既有审核提交流程

### Requirement: Risk finding titles use non-conflicting i18n keys

每条 finding 的标题、原因、建议操作、验收标准 MUST 使用互不冲突的运行时 i18n 键。标题键 MUST 使用 `detail.riskFinding.<code>.title`，不得使用与子键共享同一路径前缀且自身又是叶子值的 `detail.riskFinding.<code>`（嵌套消息树会覆盖叶子标题，导致 UI 回退到英文 `summary`）。

#### Scenario: zh-CN risk tab localizes SQL finding title

- **WHEN** 用户界面语言为 zh-CN，且 finding 的 `payload.code` 为 `source_sql_present`
- **THEN** 风险列表标题 MUST 显示中文翻译，不得显示英文源文 `Source package contains SQL resources that require reviewer inspection.`
