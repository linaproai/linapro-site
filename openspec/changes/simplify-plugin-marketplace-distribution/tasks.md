## 1. 数据模型与迁移

- [x] 1.1 新增幂等 SQL 迁移：插件/Git 源字段（`source_kind`、`repo_url`、`repo_provider`、`credential_ref`、同步状态时间戳等）及既有 upload 行回填
- [x] 1.2 扩展 release/artifact 模型以支持 git `source_ref`、`tar.gz` 产物类型或 content-type 区分
- [x] 1.3 设计并实现私有仓 token 安全存储（加密或宿主密钥引用），确保 DAO/API 不落明文回显
- [x] 1.4 重新生成或更新插件 DAO/DO/Entity（按插件本地 `hack/config.yaml` 与项目代码生成约定）

## 2. 上传通道：zip / tar.gz 与扫描

- [x] 2.1 扩展上传入口接受 `.zip`、`.tar.gz`、`.tgz`，统一解压安全边界（路径穿越、大小、解压炸弹）
- [x] 2.2 复用并扩展源码包扫描器以支持 tar 容器，保持目录规范校验
- [x] 2.3 复用并扩展动态包扫描器以支持 tar 容器，保持 wasm/清单一致性校验
- [x] 2.4 上传路径强制 `source_kind=upload`，拒绝 git 插件混用上传
- [x] 2.5 补充源码/动态 zip 与 tar.gz 扫描与上传单测

## 3. Git 元数据发现

- [x] 3.1 实现 GitHub/Gitee 提供商客户端：列表 tags、读取 raw `plugin.yaml`、可选目录存在性抽样（白名单主机）
- [x] 3.2 实现 `DiscoverGitMetadata`：版本一致性校验、`type=source` 限制、草稿写入、不落全量源码
- [x] 3.3 实现登记 Git 源 API/服务：归属校验、credential 绑定、登记后立即发现
- [x] 3.4 实现定时轮询任务（jobcap/cron）：扫描全部 git 源、增量 tag、失败状态可观测
- [x] 3.5 动态类型/不支持场景诊断信息与错误码（英文源文本）
- [x] 3.6 补充 Git 发现、版本不一致、认证失败、禁止混源单测（可用 fake HTTP 提供商）

## 4. 查询契约：distribution 与可见性

- [x] 4.1 在版本详情/列表项或专用 `GET .../distribution` 返回 `distribution` 投影
- [x] 4.2 `mode=git` 返回 `repoUrl`/`ref`/`provider`/`requiresAuth`，永不回传 token
- [x] 4.3 `mode=https` 返回产物类型、`sha256`、下载会话约定字段
- [x] 4.4 在数据库查询阶段注入可见性/下载权限过滤，无权不泄露坐标
- [x] 4.5 审核通过/下架/同步后刷新读模型与缓存失效点
- [x] 4.6 补充 distribution 与可见性单测

## 5. 发布工作台前端

- [x] 5.1 「我的插件」增加登记 Git 源表单（URL、可选 token、可见性等）
- [x] 5.2 上传流支持 `zip`/`tar.gz` 选择与校验提示
- [x] 5.3 列表展示 `source_kind`、同步状态、最新发现版本摘要
- [x] 5.4 详情/安装引导展示 `distribution` 摘要（git 坐标或 HTTPS 校验信息）
- [x] 5.5 更新插件 `manifest/i18n` 中文/英文 UI 文案与 apidoc 翻译

## 6. CLI / linactl 安装

- [x] 6.1 确定命令命名（`marketplace.install` 或扩展 `plugins.install`）并注册跨平台入口
- [x] 6.2 实现读取市场 API `distribution` 的客户端（含鉴权与错误处理）
- [x] 6.3 `mode=git`：clone/fetch 指定 ref 到 `apps/lina-plugins/<plugin-id>/`，衔接或复用现有 plugins 工作区逻辑
- [x] 6.4 `mode=https`：创建下载会话、下载、sha256 校验、解压/落盘
- [x] 6.5 源码插件提示重建部署；动态包提取 wasm 并说明后续本地上传治理步骤
- [x] 6.6 记录跨平台影响（macOS/Linux/Windows）并补充命令级测试或可重复验证步骤
- [x] 6.7 更新 `linactl` README 中英文说明

## 7. 测试与治理门禁

- [x] 7.1 扩展或新增市场 E2E（TC 编号遵循 `lina-e2e`：在既有 TC001–TC003 后续分配，如 TC004 Git 登记与发现、TC005 tar.gz 上传与 distribution 展示）
- [x] 7.2 运行市场插件相关单元测试、前端检查、`make plugins.check`（或项目约定等价命令）
- [x] 7.3 运行 `openspec validate simplify-plugin-marketplace-distribution --strict`
- [x] 7.4 在任务记录中写明 i18n、缓存一致性、数据权限、dev-tooling 跨平台与 DI 影响结论（无新增运行期 DI 时显式无影响）
- [x] 7.5 实现完成后调用 `lina-review` 做变更审查

## Feedback

- [x] **FB-1**: 添加插件弹窗中「上传压缩包」应作为表单字段与「分发方式」左标签对齐，右侧为提示与上传区域
- [x] **FB-2**: Git 元数据发现在无 version tags 时不应直接失败；应优先使用最新 semver tags，无 tags 时回退 main，main 也不存在才报错
- [x] **FB-3**: Git 源登记/发现应自动识别单插件仓库与多插件仓库，并为多插件记录 repo_path / distribution.path
- [x] **FB-4**: 添加插件后应立即进入「我的插件」列表，并由异步定时任务推进处理状态：待验证 → 待审核 → 已发布
- [x] **FB-5**: 添加插件抽屉仅保留顶部「添加插件」标题，去掉内容区紧挨顶部的重复分区标题
- [x] **FB-6**: 去掉用户可见的「待拉取」状态；Git/上传添加成功后一律进入「待验证」，异步任务在待验证内完成元数据发现与校验后进入待审核
- [x] **FB-7**: 「我的插件」状态下拉选项与列表展示状态不一致（仍为草稿等 market 值）；应对齐待验证/待审核/已发布等展示态并支持对应筛选
- [x] **FB-8**: 「我的插件」名称列偏窄，需加宽
- [x] **FB-9**: 「我的插件」在最新版本列右侧增加下载量列
- [x] **FB-10**: 无 tag 回退 main 时须在 release 表记录发现时的 commit id；`distribution.ref` 优先使用该 commit，禁止仅以浮动 `main` 分支下载已发布/已登记版本
- [x] **FB-11**: 插件须保留多版本历史记录，用户可查询并选择安装任一已发布历史版本（兼容性回退）
- [x] **FB-12**: 补充历史版本与 Git commit 钉扎 E2E（TC006）及 source pin i18n（避免 vue-i18n `@` 链接语法）
- [x] **FB-13**: 插件市场展示插件文档时按用户当前语言环境自动选择 i18n 文档：仅一种语言则展示该语言；多种语言优先匹配用户语言；无法匹配则回退英文
- [x] **FB-14**: 「我的插件」操作列仅保留三个按钮：详情、新版本、下架；新版本请求服务端自动更新版本信息；详情展示插件信息与文档；下架撤回发布状态并在市场不可见
- [x] **FB-15**: 「我的插件」列表将来源（sourceKind）单独展示为一列，不再作为名称右侧标签
- [x] **FB-16**: 市场需持久化版本文档语言包正文快照，并让详情页一次读取多语言文档集合后本地快速切换
- [x] **FB-17**: 「我的插件」点击详情后文档页不应展示为空，应显示当前版本可用的文档正文
- [x] **FB-18**: 通过 Git 仓库 `https://github.com/linaproai/official-plugins` 添加插件后，「我的插件」详情文档页不应为空，应展示仓库内当前版本可用文档正文
- [x] **FB-19**: 「我的插件」详情文档左侧目录应列出版本下全部可导航 `manifest/docs` Markdown 文件（而非仅 `index.md`），且目录项标题使用各 md 文件首个标题而非文件名
- [x] **FB-20**: 插件市场详情文档页 Markdown 渲染样式不美观；应使用成熟 Markdown 渲染链路（对齐 VS Code/GitHub 预览），支持代码高亮、表格、图片等常用语法，并支持 Mermaid 图表渲染
- [x] **FB-21**: 「我的插件」列表支持按插件标识、状态、下载量、更新时间排序，默认按插件标识升序
- [x] **FB-22**: 「我的插件」详情风险摘要显示「警告/提示」计数，但风险页内容为空；Git 发现路径应与上传路径一致，将 scanner diagnostics 持久化为 plugin_marketplace_risk 明细行
- [x] **FB-23**: 插件详情风险摘要相关展示与风险 Tab 中 scanner 风险说明文案为英文源文本，缺少运行时 i18n；应按 diagnostic code 映射中英文语言包
- [x] **FB-24**: 插件详情「同步信息」字段直接展示 Git 发现写入的英文 lastSyncMessage（如 discovered 0 new draft releases...），缺少运行时 i18n；应按已知诊断模式映射中英文语言包，并覆盖我的插件状态 tooltip 与失败流水线提示
- [x] **FB-25**: 「我的插件」详情风险 Tab 列表未按风险等级排序；应按 high → warning → info 展示，高风险在前
- [x] **FB-26**: 审查「风险摘要」字段 Tag 是否按风险等级排序（高风险在前）；若已正确则记录审查结论，无需改代码
- [x] **FB-27**: 插件详情「同步信息」字段值使用 marketplace-muted 导致字号 12px，与 Descriptions 其它字段默认字号不一致；应与表格/描述列表正文对齐

## Feedback 影响记录（FB-25 ~ FB-27）

- [x] 根因（FB-25）：`ListReleaseRisks` 仅 `OrderDesc(id)`，前端原样渲染 `currentRisks`/`reviewRisks`，未按 severity 排序
- [x] 修复（FB-25）：后端按 severity CASE 排序 + 页内稳定排序；前端 `sortMarketplaceRiskFindingsBySeverity` 在详情/审核页加载后排序
- [x] 审查（FB-26）：风险摘要 Tag 模板已按 `high → warning → info` 顺序渲染（`getRiskCounts().high/warning/info`）；无代码缺陷；E2E TC-15a 固化摘要顺序断言
- [x] 根因（FB-27）：`lastSyncMessage` 包了 `marketplace-muted`（`font-size: 12px`），与 Descriptions 正文默认字号不一致
- [x] 修复（FB-27）：去掉同步信息字段的 `marketplace-muted`，继承 Descriptions 正文样式
- [x] i18n：无新增文案/语言包键；仅展示顺序与样式
- [x] 缓存一致性：无影响
- [x] 数据权限/可见性：无影响
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 外部规则：已读 `plugin.md`、`frontend-ui.md`、`backend-go.md`、`testing.md`、`i18n.md`、`openspec.md`；API 契约无变更
- [x] 截图：`temp/20260731/145625-risk-list-severity-order.png`（风险 Tab 高危在前）、`temp/20260731/145625-risk-summary-severity-order.png`（摘要 高危→警告→提示）、`temp/20260731/145617-last-sync-message-i18n-zh.png`（同步信息字号对齐）
- [x] 测试：后端 `TestSortMarketplaceRiskItemsBySeverity`、前端 unit severity ordering、E2E TC-15a；回归 TC-12/TC-13/TC-14；`make lint.go dir=apps/lina-plugins/linapro-plugin-marketplace plugins=1`、`openspec validate --strict` 通过

## Feedback 影响记录（FB-24）

- [x] 根因：Git 发现与流水线将英文诊断写入 `last_sync_message`；详情「同步信息」、失败流水线提示与我的插件状态 tooltip 原样渲染，未走运行时语言包。用户反馈中的 `discovered 0 new draft releases...` 对应「同步信息」字段（风险摘要标签本身已 i18n）
- [x] 修复：新增 `frontend/utils/sync-message.ts` 的 `formatMarketplaceLastSyncMessage`；按后端已知英文源文本模式映射 `detail.syncMessage.*`；详情页、失败提示与我的插件 tooltip 接入；补齐 en-US/zh-CN 10 个键
- [x] i18n：有影响；插件 `manifest/i18n/zh-CN|en-US/plugin.json` 新增 syncMessage 键；未知自由文本诊断保留英文 fallback
- [x] 缓存一致性：无影响；翻译在前端求值，不改 last_sync_message 权威数据
- [x] 数据权限/可见性：无影响
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 外部规则：已读 `plugin.md`、`frontend-ui.md`、`i18n.md`、`testing.md`、`openspec.md`；后端 Go/API/SQL 无契约变更
- [x] 截图：修复前 `temp/e2e/20260731/03-risk-summary-field.png` 同步信息为英文；修复后 `temp/20260731/143944-last-sync-message-i18n-zh.png` 为「未发现新草稿版本（已有 1 个不可变版本）」
- [x] 测试：`node --test marketplace-frontend.test.mjs`、E2E TC-14a、回归 TC-12/TC-13、`openspec validate --strict` 通过
- [x] 生效说明：需加载新增 i18n 资源与前端代码（重启宿主/刷新前端）；历史 last_sync_message 英文行无需重扫，前端模式匹配即可

## Feedback 影响记录（FB-23）

- [x] 根因：scanner 将英文 `summary` 与稳定 `payload.code` 写入 `plugin_marketplace_risk`；详情/审核页直接渲染 `risk.summary`，未按 code 走运行时语言包；严重级别/类型标签已有 i18n，正文未覆盖
- [x] 修复：新增 `frontend/utils/risk.ts` 的 `formatMarketplaceRiskFindingSummary`；详情与审核风险列表按 `payload.code` 映射 `detail.riskFinding.*`；补齐 en-US/zh-CN 全部已知 diagnostic code 文案
- [x] i18n：有影响；插件 `manifest/i18n/zh-CN|en-US/plugin.json` 新增 10 个 riskFinding 键；运行时以英文源文本为 fallback
- [x] 缓存一致性：无影响；翻译在前端求值，不改风险明细权威数据
- [x] 数据权限/可见性：无影响
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 外部规则：已读 `plugin.md`、`frontend-ui.md`、`i18n.md`、`testing.md`、`openspec.md`；后端 Go/API/SQL 无契约变更
- [x] 测试：`node --test marketplace-frontend.test.mjs`、E2E TC-12b/TC-13a（zh-CN 正文、摘要计数一致）通过；`openspec validate --strict` 通过
- [x] 生效说明：需重启宿主/插件以加载新增 i18n 资源与前端代码；历史风险行已有 `payload.code` 无需重扫

## Feedback 影响记录（FB-22）

- [x] 根因：`discoverOneGitRef` 仅通过 `buildSourceRiskSummary` 写入 `release.risk_summary` 聚合计数，未像源码包/动态包上传路径那样调用 `replaceReleaseRisks` 写入 `plugin_marketplace_risk` 明细行；风险页 API 只读明细表，故摘要有「警告 2 提示 1」而风险 Tab 为空
- [x] 修复：草稿发现路径在 display i18n 之后调用 `replaceReleaseRisks(ctx, release, diagnostics)`；已发布/不可变版本在 re-discovery 时回填风险明细（不改动冻结的 `risk_summary`）
- [x] i18n：无影响；无新增运行时 UI 文案、菜单、API 文档源文本或语言包键
- [x] 缓存一致性：无影响；风险明细仍以 `plugin_marketplace_risk` 为权威数据源，发现时整体替换
- [x] 数据权限/可见性：无影响；风险列表仍走 `requireVisibleRelease`，不扩大可见范围
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 外部规则：已读 `plugin.md`、`backend-go.md`、`testing.md`、`openspec.md`、`i18n.md`；API/SQL/前端 UI/缓存/数据权限/dev-tooling 无契约级变更
- [x] 测试：`GOWORK=off go test ./backend/internal/service/marketplace -count=1`、`make lint.go dir=apps/lina-plugins/linapro-plugin-marketplace plugins=1`、E2E TC-12（复现空列表 + 修复后明细与摘要一致）、回归 TC-9/TC-10、`openspec validate --strict` 通过
- [x] 既有数据：已登记的 Git 插件需等待下次 Git 元数据发现（定时任务或「新版本」同步）回填风险明细行

## Feedback 影响记录（FB-21）

- [x] 根因：我的插件列表未启用远程排序；后端 `listPluginsFromIdentityTable` 固定 `updated_at DESC`，前端列无 `sortable`
- [x] 修复：`MyPluginListReq` 增加 `orderBy`/`orderDirection`；拥有列表默认 `pluginId ASC`；白名单字段排序；前端 `sortConfig` + 四列 `sortable`；mock/E2E 同步
- [x] i18n：有影响；插件 `zh-CN` apidoc 补充 `orderBy`/`orderDirection` 字段说明；无新增运行时 UI 文案键
- [x] 缓存一致性：无影响
- [x] 数据权限/可见性：无影响；仍按 owner 过滤，仅改变排序
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 外部规则：已读 `plugin.md`、`frontend-ui.md`、`api-contract.md`（DTO 变更）、`backend-go.md`、`testing.md`、`openspec.md`、`i18n.md`
- [x] 测试：`GOWORK=off go test ./backend/internal/service/marketplace -run PluginIdentity|ApplyPluginIdentity`、`node --test` 前端/apidoc 单测、E2E TC-11、`openspec validate --strict` 通过

## Feedback 影响记录（FB-20）

- [x] 根因：详情文档仅用裸 `markdown-it` + 简易 CSS，无语法高亮、无 Mermaid、样式与 VS Code/GitHub 预览差距大；`data:image` 被 markdown-it 默认 `validateLink` 拦截
- [x] 修复：插件 `frontend` 引入 `highlight.js` + `mermaid`；`utils/markdown.ts` 增加高亮 fence、Mermaid 围栏、安全图片规则与 `data:image/*` 放行；详情页挂载后 `enhanceMarketplaceMarkdown` 渲染图表；CSS 对齐 VS Code/GitHub Markdown 预览
- [x] i18n：无新增运行时 UI 文案、菜单、按钮、表格列、API 文档源文本或语言包键；渲染对象为插件文档 Markdown 正文本身
- [x] 缓存一致性：无影响；不新增缓存、快照或失效策略
- [x] 数据权限/可见性：无影响；仅前端渲染链路，不改变文档读取或可见性边界
- [x] 开发工具跨平台：无影响；依赖安装仍走插件 `frontend/pnpm install`，未改 Makefile/linactl/CI 默认入口
- [x] DI：无新增运行期宿主 DI；`mermaid` 为前端动态 import
- [x] 外部规则：已读 `plugin.md`、`frontend-ui.md`、`testing.md`、`openspec.md`、`i18n.md`；后端 Go/API/SQL/缓存/数据权限/dev-tooling 无影响
- [x] 测试：`node --test` 插件前端静态与 markdown 渲染单测通过；`E2E TC-7`（表格/图片/Mermaid/代码高亮）、回归 `TC-9`/`TC-10` 通过；`openspec validate simplify-plugin-marketplace-distribution --strict` 通过

## 8. 任务完成影响记录

- [x] 8.1 i18n：有影响；已更新市场插件 zh-CN/en-US UI 文案，API 错误码英文源文本已补充
- [x] 8.2 缓存一致性：有影响；审核通过仍走既有读模型重建；Git 同步写入插件同步状态字段
- [x] 8.3 数据权限/可见性：有影响；distribution 查询复用 resolveAccessiblePlugin；token 不回显
- [x] 8.4 开发工具跨平台：有影响；marketplace.install 使用 stdlib HTTP + 本地 git，已在 README 中英文说明
- [x] 8.5 DI：无新增运行期宿主 DI；市场服务仍 `New(nil)` 自建 artifact store；Jobs 注册使用同一构造路径
- [x] 8.6 验证：`GOWORK=off go test ./backend/...`（marketplace 插件）通过；`linactl` build 通过；`openspec validate ... --strict` 通过

## Feedback 影响记录（FB-13 ~ FB-15）

- [x] i18n：有影响；文档选择按请求语言/Accept-Language；更新 `mine.messages.packageAdded`/`delistOnlyPublished`/`confirm.delist` 中英文文案
- [x] 缓存一致性：无影响；文档读取仍走既有版本索引与按需渲染，未改缓存键策略
- [x] 数据权限/可见性：无新增写路径权限；下架仍复用既有 `OwnerDelistPlugin` 归属校验与公开目录过滤
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖；文档 locale 使用 `gi18n.LanguageFromCtx` 与 `Accept-Language`，未注入新服务
- [x] 测试：文档 fallback 单测更新；前端 composition 单测更新；E2E POM/TC001 覆盖三按钮与来源列；`openspec validate ... --strict` 通过

## Feedback 影响记录（FB-16）

- [x] i18n：有影响；已维护插件 `zh-CN`/`en-US` 运行时资源与 `zh-CN` apidoc 翻译，插件前端/i18n 静态测试通过；全局`make i18n.check`失败于既有宿主 `linapro-content-notice` 缺失键，非本次 marketplace 变更引入
- [x] 缓存一致性：有影响；文档正文快照以 `plugin_marketplace_doc` 为权威数据源，发布包解析或 Git 发现时整体替换，详情读取不新增运行时缓存
- [x] 数据权限/可见性：有影响；文档包读取复用既有 `requireVisibleRelease`，不可见版本不得返回任何语言内容
- [x] 开发工具跨平台：无影响；不新增或修改默认开发工具入口
- [x] DI：无新增运行期依赖；复用市场服务、文档索引与现有 artifact/Git 发现流程
- [x] 测试：补充后端文档快照/多语言包单测、前端本地切换静态测试与 E2E 子断言；`GOWORK=off go test ./backend/... -count=1`、插件前端/apidoc i18n 静态测试、`make lint.go dir=apps/lina-plugins/linapro-plugin-marketplace plugins=1`、`make plugins.check`、`openspec validate ... --strict`通过；`db.init`/`dao`因本地数据库既有 schema 缺失和插件表未初始化阻断，浏览器 E2E 未运行

## Feedback 影响记录（FB-17）

- [x] 根因：上传包可接收 `.tar.gz`/`.tgz`，但详情文档读取存储产物时只按 ZIP 打开，导致压缩包内 `manifest/docs` 未被读取；同时前端详情页只使用 `markdown` 渲染结果，后端或测试夹具仅返回安全 HTML `content` 时正文会为空
- [x] i18n：无新增运行时文案、菜单、按钮、表格列、API 文档源文本或语言包资源；插件市场自身前端 i18n 静态测试通过，全局`make i18n.check`失败于既有宿主 `linapro-content-notice` 缺失键，非 FB-17 引入
- [x] 缓存一致性：无影响；不新增缓存、快照或失效策略，文档详情仍从当前可见版本的存储产物或既有文档快照读取
- [x] 数据权限/可见性：无新增读写边界；「我的插件」详情仍复用既有可见版本文档接口与 `requireVisibleRelease` 可见性约束，不扩大跨用户或公开市场可见范围
- [x] 开发工具跨平台：无影响；不修改 `Makefile`、`linactl`、CI 或脚本入口，`.tar.gz` 转 ZIP 读取复用既有标准库转换路径
- [x] DI：无新增运行期依赖、构造函数参数或启动装配路径
- [x] 测试：`GOWORK=off go test ./backend/internal/service/marketplace -count=1`、`node --test hack/tests/unit/marketplace-frontend.test.mjs`、`make lint.go dir=apps/lina-plugins/linapro-plugin-marketplace plugins=1`、`E2E_BASE_URL=http://127.0.0.1:5666 E2E_BACKEND_BASE_URL=http://127.0.0.1:9120 pnpm test:module -- plugin:linapro-plugin-marketplace --grep "TC-9"`、`E2E_BASE_URL=http://127.0.0.1:5666 E2E_BACKEND_BASE_URL=http://127.0.0.1:9120 pnpm test:module -- plugin:linapro-plugin-marketplace --grep "TC-7"`、`openspec validate simplify-plugin-marketplace-distribution --strict`通过

## Feedback 影响记录（FB-18）

- [x] 根因：`official-plugins`中`linapro-tenant-core`当前版本只有仓库根`README.md`/`README.zh-CN.md`，没有`manifest/docs`；Git 发现已索引 README 文档记录，但详情文档接口和前端目录将`sourceKind=readme`全部过滤，导致目录为空并返回`PLUGIN_MARKETPLACE_DOCUMENT_NOT_FOUND`
- [x] 修复：后端在同一发布版本没有任何`manifest/docs`文档时，将 README 文档作为`index.md`目录项和正文 fallback；同一版本存在`manifest/docs`时继续隐藏 README，保持 manifest 文档优先
- [x] i18n：有影响；API DTO 的`sourceKind`文档说明新增`readme`返回语义，并同步更新插件`zh-CN`apidoc 翻译；无新增运行时 UI 文案键
- [x] 缓存一致性：无新增缓存、快照失效或分布式同步策略；详情读取仍以发布版本已有包文档或 Git 文档快照为权威来源，未在 GET 文档接口中引入写入副作用
- [x] 数据权限/可见性：无新增读写边界；文档详情仍先执行`requireVisibleRelease`，仅对当前可见版本返回 README fallback，不扩大跨用户、跨租户或公开市场可见范围
- [x] 开发工具跨平台：无影响；不修改`Makefile`、`linactl`、CI、脚本入口或平台相关路径
- [x] DI：无新增运行期依赖、构造函数参数、启动装配或共享实例路径
- [x] 截图审查：修复前真实截图`temp/e2e/20260730/20260730-092935-linapro-tenant-core-docs-tab.png`显示“所选版本暂无可用文档”；修复后 E2E 截图`temp/20260730/175937-mine-detail-readme-docs.png`和真实服务截图`temp/e2e/20260730/175500-real-linapro-tenant-core-docs-final.png`均显示 README 文档目录与正文
- [x] 测试：`GOWORK=off go test ./backend/internal/service/marketplace -count=1`、`node --test hack/tests/unit/marketplace-frontend.test.mjs`、`node --test hack/tests/unit/marketplace-apidoc-i18n.test.mjs`、`make lint.go dir=apps/lina-plugins/linapro-plugin-marketplace plugins=1`、`E2E_BASE_URL=http://127.0.0.1:5666 E2E_BACKEND_BASE_URL=http://127.0.0.1:9120 pnpm --pm-on-fail=ignore -C hack/tests test:module -- plugin:linapro-plugin-marketplace --grep "TC-10"`、`E2E_BASE_URL=http://127.0.0.1:5666 E2E_BACKEND_BASE_URL=http://127.0.0.1:9120 pnpm --pm-on-fail=ignore -C hack/tests test:module -- plugin:linapro-plugin-marketplace --grep "TC-7|TC-9"`、`openspec validate simplify-plugin-marketplace-distribution --strict`通过；`make i18n.check`失败于既有宿主`linapro-content-notice`缺失键，非 FB-18 引入

## Feedback 影响记录（FB-19）

- [x] 根因：Git 文档重索引对已发布版本使用钉扎`source_commit`读取正文，而候选树可能来自较新的发现引用；当`manifest/docs`在钉扎之后才加入仓库时，读正文失败并只剩 README，目录被折叠为单一`index.md`。用户看到的是 README 回退而非完整 md 目录
- [x] 修复：1）`indexGitReleaseDocuments`按内容引用重新`ListTreePaths`，保证候选与可读 blob 一致；2）已发布版本文档富化改为按当前发现引用（tag/`main`）重索引，安装仍使用`source_commit`钉扎；3）每个发现引用都执行文档富化（不再仅限第一个 ref）
- [x] i18n：无新增运行时 UI 文案、菜单、按钮或语言包键；目录标题取自各 md 文件首个 ATX 标题文本；无 apidoc 源文本变更
- [x] 缓存一致性：有影响；Git 同步会替换版本文档磁盘快照（`docs-snapshot/.../manifest.json`与 content），详情 GET 仍只读快照不回源
- [x] 数据权限/可见性：无新增读写边界；文档仍经`requireVisibleRelease`，不扩大跨用户/公开可见范围
- [x] 开发工具跨平台：无影响
- [x] DI：无新增运行期依赖
- [x] 真实验证：`git-sync`后`linapro-tenant-core`/`linapro-ai-core`/`linapro-demo-source`的`catalog`均返回 3 项且标题为「多租户/智能中心/示例插件-源码插件」「更新日志」「配置说明」
- [x] 测试：`GOWORK=off go test ./backend/internal/service/marketplace -count=1`、`node --test hack/tests/unit/marketplace-frontend.test.mjs`、`E2E_BASE_URL=... pnpm -C hack/tests test:module -- plugin:linapro-plugin-marketplace --grep "TC-7|TC-9"`、`openspec validate simplify-plugin-marketplace-distribution --strict`通过
