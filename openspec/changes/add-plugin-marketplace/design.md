## Context

`localdocs/plugin-marketplace-technical-design.md`的总体边界成立：插件市场只负责发布、检索、文档展示、审核、可信来源和下载，不替代`lina-core`已有插件发现、安装、启用、禁用、升级、`hostServices`授权和运行时缓存治理。

结合本次补充约束，市场自身必须作为源码插件开发，并使用内置分发模式。现有规范已经定义`distribution=builtin`只能用于`type=source`且需要编译期源码插件注册绑定；普通插件管理列表默认隐藏`builtin`插件，宿主启动阶段会自动收敛内置源码插件。因此本变更不修改`lina-core`分发语义，而是在`apps/lina-plugins/linapro-plugin-marketplace/`内新增一个内置源码插件。

## Goals / Non-Goals

**Goals:**

- 以`linapro-plugin-marketplace`源码插件交付市场能力，`plugin.yaml`声明`type: source`、`distribution: builtin`和`i18n.enabled: true`。
- 支持开发者发布源码插件市场包和动态插件运行时市场包。
- 支持用户分页检索、查看详情、按语言查看`manifest/docs/`文档、识别风险摘要并下载产物。
- 支持审核、发布者归属、版本不可变、短期下载会话、可见性过滤和下载统计快照。
- 支持多语言运行时 UI、菜单、接口文档、错误消息和市场文档语言回退。
- 复用现有本地插件治理：动态插件下载后上传`plugin.wasm`，源码插件下载后进入`apps/lina-plugins/`并重新构建部署。

**Non-Goals:**

- `MVP`不实现在线支付、订阅、发票、退款、收入分成或完整许可证服务。
- `MVP`不让源码插件下载后在生产运行时直接写入源码工作区或一键生效。
- `MVP`不把市场发布状态、审核状态、下载量、评分或付费授权写入被发布插件的`plugin.yaml`。
- `MVP`不引入必需的`marketplace.yaml`。
- `MVP`不修改`distribution=managed | builtin`的既有语义。

## Decisions

### 决策一：市场作为内置源码插件实现

市场插件目录为`apps/lina-plugins/linapro-plugin-marketplace/`，并维护源码插件标准结构：`plugin.yaml`、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/sql/`、`manifest/i18n/`、`manifest/docs/`和`hack/tests/e2e/`。

选择理由：

- 市场是业务能力，不是`lina-core`插件治理通用领域契约。
- `distribution=builtin`已经提供启动自动安装、自动启用和普通插件管理隐藏策略，正好适配“随框架内置交付”的市场入口。
- 源码插件可以自带后端、前端、SQL、i18n、文档和 E2E，业务闭环清晰。

替代方案：

| 方案 | 取舍 |
|------|------|
| 放入`lina-core`系统模块 | 会扩大核心宿主职责，容易把市场业务和通用插件治理耦合 |
| 独立云服务优先 | 不满足本次“源码插件、内置分发、可私有化部署”的约束 |
| 动态插件实现市场 | `distribution=builtin`不允许动态插件，且市场需要较深的后端、SQL 和工作台集成 |

### 决策二：市场包使用统一压缩容器，运行时权威仍来自既有治理

`MVP`使用`.zip`作为市场上传和下载容器：

| 包类型 | 必需内容 | 用途 |
|--------|----------|------|
| 源码插件市场包 | 完整源码插件目录、`plugin.yaml`、`manifest/docs/` | 下载到源码工作区后随宿主重新构建 |
| 动态插件运行时市场包 | 根级`plugin.yaml`、`plugin.wasm`、`manifest/docs/`、`README` | 下载后提取`plugin.wasm`并复用现有动态上传治理 |

动态插件的根级`plugin.yaml`只用于市场快速索引和审核；本地运行时安装校验仍以`plugin.wasm`内嵌清单为准。市场审核必须阻断根级清单与内嵌清单在`id`、`name`、`version`、`type`、`dependencies`、`hostServices`和多租户字段上的不一致。

### 决策三：市场数据模型归属插件自身，并优先使用发布读模型

市场表位于插件自己的`manifest/sql/`，正式实现时以单迭代 SQL 文件维护。核心表包括发布者、插件、版本、产物、文档索引、风险摘要、下载会话、分类标签关系和列表读模型。

设计约束：

- 已发布版本不可变，同一`pluginId + version`不得静默覆盖。
- 下架、废弃、审核状态通过状态字段表达，不物理删除发布历史。
- 用户可管理的草稿、发布者和插件记录保留`deleted_at`用于软删除和审计恢复；已发布版本优先用状态流转保留历史。
- 列表接口读取读模型或搜索索引，不逐行读取包存储、解析`plugin.yaml`或加载完整`Markdown`。
- 高频查询索引覆盖`plugin_id`唯一查询、发布者列表、版本唯一约束、状态分页、类型分类过滤、文档语言读取和下载会话过期校验。

### 决策四：API 使用市场插件自有 REST 契约

市场插件通过源码插件路由注册器接入宿主，并在插件专属命名空间下注册`/market/plugins`等 REST 资源路径。公开访问路径以`/x/linapro-plugin-marketplace`为前缀，例如`/x/linapro-plugin-marketplace/market/plugins`。查询类接口使用`GET`并分页，副作用接口使用`POST`或`PUT`。公开响应中的时间点字段使用 Unix 毫秒时间戳。

核心接口分组：

| 分组 | 示例 |
|------|------|
| 目录查询 | `GET /x/linapro-plugin-marketplace/market/plugins`、`GET /x/linapro-plugin-marketplace/market/plugins/{pluginId}`、`GET /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases` |
| 文档风险 | `GET /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases/{version}/docs`、`GET /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases/{version}/risks` |
| 发布审核 | `POST /x/linapro-plugin-marketplace/market/plugins`、`POST /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases`、`POST /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases/{version}/submit-review` |
| 下载 | `POST /x/linapro-plugin-marketplace/market/plugins/{pluginId}/releases/{version}/downloads`、`GET /x/linapro-plugin-marketplace/market/download-sessions/{sessionId}` |

权限标签使用市场插件自己的权限命名空间，例如`market:plugin:view`、`market:plugin:publish`、`market:plugin:review`和`market:plugin:download`。

### 决策五：可见性、下载和统计必须在数据库查询阶段过滤

市场读取类接口按公开、私有和未来授权快照注入可见性过滤。详情、文档、风险和下载会话创建都必须先确认当前用户可见目标插件和版本；聚合统计不得泄露无权访问的私有插件存在性。

下载会话是短期授权资源，创建时校验可见性和下载权限，返回短期地址或受控会话 ID。下载事件异步聚合为统计快照，普通列表和文档查询不得产生业务写入。

### 决策六：多语言分成插件运行时语言和市场文档语言

市场插件自身启用`i18n.enabled: true`，因此：

- API DTO、`g.Meta`和错误 fallback 使用英文源文本。
- `manifest/i18n/zh-CN/`维护运行时 UI、菜单、错误和`apidoc`翻译资源。
- 前端静态文案使用运行时语言包，不在模块顶层调用`$t()`求值。
- 市场文档按`manifest/docs/<locale>/`读取；被发布插件是否启用运行时`i18n`不影响市场文档多语言展示。

文档读取顺序为当前语言、被发布插件`plugin.yaml`中的`i18n.default`、`zh-CN`、`README.zh-CN.md`、`README.md`。

### 决策七：缓存以发布不可变和读模型为基础

| 缓存对象 | 权威数据源 | 失效触发 | 集群策略 | 最大陈旧时间 |
|----------|------------|----------|----------|--------------|
| 市场列表读模型 | 市场业务表和审核结果 | 发布、下架、审核、分类标签、统计快照刷新 | 数据库读模型为准，必要时通过插件缓存能力按 scope 失效 | 不超过统计快照刷新周期 |
| 文档渲染缓存 | 版本产物、`doc_hash` | 新版本发布或草稿产物替换 | 缓存键含`release_id + locale + path + doc_hash` | 已发布版本可长期缓存 |
| 下载统计快照 | 下载事件流水 | 异步聚合任务 | 聚合写入数据库快照，列表只读快照 | 允许短暂延迟，不作为权限依据 |

缓存不可用时，读取路径回源到数据库或产物存储；权限和可见性不得依赖可丢失缓存放行。

### 决策八：前端使用现有 Vben 工作台模式

市场前端放在插件`frontend/`目录，并接入当前`apps/lina-vben/apps/web-antd`工作台的插件页面、菜单和权限机制。页面实现使用`Page`、`useVbenVxeGrid`、`useVbenForm`、`useVbenModal`、`Upload.Dragger`、`requestClient.download`和`IconifyIcon`等现有模式。

核心页面包括：

- 市场目录列表和筛选。
- 插件详情、版本、文档和风险摘要。
- 发布者后台：上传草稿、提交审核、查看审核结果。
- 审核后台：清单一致性、`hostServices`、`SQL`、外部网络、文档入口和路由风险。
- 下载确认和动态插件导入衔接。

## Risks / Trade-offs

| 风险 | 缓解策略 |
|------|----------|
| 源码插件下载被误解为运行时安装 | 下载确认和文档明确源码插件必须进入`apps/lina-plugins/`并重新构建部署 |
| 市场业务污染`lina-core` | 默认闭环在`linapro-plugin-marketplace`，只有缺失稳定插件能力时才单独评估宿主扩展 |
| 列表接口读取成本过高 | 发布成功后生成读模型，列表只读最小投影和统计快照 |
| 文档渲染引入脚本风险 | 禁止脚本执行，图片路径限制在版本文档资源范围内，外链做安全展示 |
| 动态包清单不一致 | 审核阻断根级`plugin.yaml`和`plugin.wasm`内嵌清单关键字段差异 |
| 私有插件存在性泄露 | 查询、详情、文档、风险、下载和统计都在数据库查询阶段注入可见性过滤 |
| `builtin`插件生命周期失败影响启动 | 启动失败保留诊断；回滚通过移除或修复内置源码插件后重新部署 |

## Migration Plan

1. 创建`linapro-plugin-marketplace`源码插件骨架，先检查插件根目录是否存在本地`AGENTS.md`。
2. 维护`plugin.yaml`、`plugin_embed.go`、源码插件注册入口和`distribution: builtin`。
3. 新增插件 SQL、DAO、后端 API、service、controller、i18n 和错误码。
4. 新增市场前端页面、API 适配、菜单权限和运行时语言包。
5. 增加单元测试、插件 E2E、i18n 检查、SQL 静态检查、Go 编译门禁和`openspec validate add-plugin-marketplace --strict`。
6. 回滚时移除或修复该内置源码插件并重新部署；业务表保留，不通过回滚脚本物理删除市场历史数据。

## Open Questions

- `MVP`确认使用`.zip`作为市场包容器；如后续需要支持`OCI artifact`、远程`Git`来源或离线签名包，应另立变更。
- 付费授权、私有企业市场源、离线许可证和企业审批流仅保留数据边界，不进入本次实现范围。
