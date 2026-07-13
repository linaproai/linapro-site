## Context

`linapro-plugin-marketplace`已作为内置源码插件交付，并具备发布者、插件草稿、版本上传、提交审核与审核决策等后端能力。管理工作台当前入口为：

| 现有菜单 | 权限 | 问题 |
|----------|------|------|
| Plugin Marketplace（catalog） | `market:plugin:view` | 与“消费侧目录”混同，发布者/管理者职责不清 |
| Plugin Marketplace Detail（隐藏） | `market:plugin:view` | 详情跳转 path 缺少父级前缀导致 404 |
| Marketplace Console | `market:plugin:publish` | 发布与审核塞在同一页；发布者也能看到审核 Tab 文案与结构 |

数据侧已具备发布者归属字段（`plugin_marketplace_publisher.owner_user_id`、插件绑定`publisher_id`），可在查询阶段过滤“我的插件”。

本变更只调整**管理工作台**信息架构、菜单权限与对应列表/审核页面；不改变市场包格式、审核规则与本地插件安装边界。

## Goals / Non-Goals

**Goals:**

- 管理工作台出现独立目录**插件市场**。
- 发布者仅见**我的插件**：归属内列表 + 状态 + 在页内发布新插件/新版本。
- 管理者见**插件列表**（全量状态）与**插件审核**（待审队列与决策）。
- 菜单与权限绑定清晰，服务端强制归属/审核权限过滤。
- 修复详情/返回路由与发布表单校验缺陷，保证主路径可用。

**Non-Goals:**

- 不在本变更实现对外“应用商店消费端”独立产品化（卡片商城、评分等）。
- 不修改`lina-core`插件安装/启用语义。
- 不实现在线支付、组织级多发布者复杂治理、自动认证发布者。
- 不把市场目录挂到租户业务菜单；市场插件保持`platform_only`。

## Decisions

### 决策一：菜单树形态

在`plugin.yaml`中声明：

```text
插件市场 (type: D, 顶级目录, 无 parent_key, sort=8；宿主扩展中心 sort=9)
├── 我的插件     (type: M, perms: market:plugin:publish)
├── 插件列表     (type: M, perms: market:plugin:review)   # 管理者
├── 插件审核     (type: M, perms: market:plugin:review)   # 管理者
└── 插件详情     (type: M, visible: 0, perms: market:plugin:view)
    └── 下载     (type: B, perms: market:plugin:download)
```

选择理由：

- 目录节点`type: D`且**不设置**`parent_key`，挂到侧边栏一级，与工作台、扩展中心同级。
- `sort: 8`小于宿主`extension`目录的`sort: 9`，保证「插件市场」排在「扩展中心」上面。
- 发布者与管理者用不同`perms`驱动菜单裁剪，无需前端硬编码角色名。
- 详情页保持`visible: 0`，从列表进入。

替代方案：

| 方案 | 取舍 |
|------|------|
| 挂在`extension`下 | 与普通插件管理混在扩展中心，产品要求独立一级目录 |
| 继续挂在`extension`扁平 | 无法形成“市场”目录，与普通 Plugins 管理混淆 |
| 单一 Console + 前端按权限藏 Tab | 仍易泄漏入口结构，审核/发布耦合难维护 |

### 决策二：权限与数据范围

| 角色语义 | 权限标签 | 数据范围 | 页面 |
|----------|----------|----------|------|
| 发布者 | `market:plugin:publish` | 仅当前用户拥有的发布者下的插件 | 我的插件 |
| 管理者 | `market:plugin:review` | 全部市场插件与待审版本 | 插件列表、插件审核 |
| 只读浏览 | `market:plugin:view` | 可见性过滤后的详情内容 | 隐藏详情页 |
| 下载 | `market:plugin:download` | 可见性过滤后的已发布产物 | 详情页下载按钮与下载`API` |

约束：

- “我的插件”列表 API MUST 在数据库查询阶段按`owner_user_id`（或等价发布者归属）过滤。
- 发布草稿、上传版本包与提交审核 MUST 传递当前用户 ID，并在任何产物存储或业务状态写入前同时校验`publisherKey`、`pluginId`与当前用户的归属关系。
- 仅有`publish`无`review`的用户 MUST NOT 看到插件列表/审核菜单，也 MUST NOT 通过 API 拉取全量管理列表或执行审核。
- 拥有`view`但不拥有`download`的用户 MUST NOT 看到下载按钮，且下载会话创建、查询与内容读取`API`仍必须拒绝请求。
- 同时拥有`publish`与`review`的管理员：可见全部管理菜单；“我的插件”仍只显示本人归属插件，全量在“插件列表”。

是否单独新增`market:plugin:manage-list`：

- **本变更不新增**，管理者列表与审核共用`market:plugin:review`，降低权限矩阵复杂度。
- 若未来需要“可看全量但不能审核”，再拆权限。

### 决策三：页面拆分与交互

| 页面 path（示意） | 职责 |
|-------------------|------|
| `plugin-marketplace-mine` | 我的插件：表格展示状态；工具栏“发布插件”；抽屉/分步向导完成发布与上传 |
| `plugin-marketplace-admin-list` | 插件列表：全量状态、筛选、进详情；无发布表单 |
| `plugin-marketplace-review` | 插件审核：默认待审队列；选中版本后展示审计项与通过/拒绝 |
| `plugin-marketplace-detail` | 详情（隐藏菜单）：版本、文档、风险、下载（按权限） |

发布向导原则（相对现 Console）：

1. 绑定/创建发布者（若当前用户尚无发布者则引导创建）。
2. 创建插件草稿或选择已有插件发新版本。
3. 上传 ZIP → 提交审核。
4. 表单使用`rules`校验；成功后刷新“我的插件”列表。

审核页原则：

- 默认加载`reviewStatus in (submitted, reviewing)`的版本队列，禁止“只能手输 pluginId”。
- 支持按`pluginId`/状态筛选；页面主体使用待审队列表格，选中版本后通过标准审核抽屉展示清单、文档、风险明细与通过/拒绝决策，保证窄视口下列表与审核内容不会互相挤压。

### 决策四：API 调整策略

优先复用现有资源，扩展查询参数与权限：

| 能力 | 方案 |
|------|------|
| 我的插件列表 | `GET /market/plugins` 增加`scope=mine`（或独立 path），`permission: market:plugin:publish`，服务端强制归属过滤 |
| 管理插件列表 | 同列表接口`scope=admin`或默认 review 权限返回含草稿/下架等状态投影，`permission: market:plugin:review` |
| 审核队列 | `GET /market/releases`（新建）或在现有 releases 上提供跨插件待审查询，`permission: market:plugin:review` |
| 发布写接口 | 保持现有`publish`路径与标签；控制器传递当前用户 ID，服务在草稿、上传和提交审核前校验发布者与插件归属，且上传必须在产物存储前完成校验 |
| 审核写接口 | 保持现有`review`路径与标签，仅允许`market:plugin:review`权限执行 |

列表投影 MUST 包含：pluginId、名称、类型、发布者、市场状态、最新版本、最新审核状态、可见性、更新时间、风险摘要等状态字段。

路由跳转：前端统一使用宿主解析后的顶级目录绝对`path`（如`/plugin-marketplace/plugin-marketplace-mine`），或通过路由`name`导航，禁止写死缺少`/plugin-marketplace`目录段的无效`path`。

### 决策五：旧入口迁移

| 旧 path | 迁移 |
|---------|------|
| `plugin-marketplace-catalog` | 管理侧移除或降级；若仍需“可下载目录”可后续单独做消费页，本变更不保留为默认管理菜单 |
| `plugin-marketplace-console` | 拆入“我的插件”发布流与“插件审核”页后删除菜单 |

菜单 key 变更后依赖插件菜单同步/升级流程刷新`sys_menu`；实现时提升插件版本并验证安装升级幂等。

## Risks / Trade-offs

- **[Risk] 仅有 view 权限的用户失去管理侧入口** → 按产品要求管理侧以 publish/review 为主；view/download 保留 API 给下载与详情，不强制菜单。
- **[Risk] 管理员同时发布自己的插件时菜单较多** → 接受；“我的插件”与“插件列表”语义不同。
- **[Risk] 待审跨插件列表缺少现成索引** → 在`plugin_marketplace_release`上按`review_status + updated_at`查询并补索引；注意 N+1，一次查出版本再批量装插件名。
- **[Risk] 旧书签/E2E 失效** → tasks 中统一更新 E2E 与文档；菜单升级后回归。
- **[Trade-off] 不拆 manage-list 权限** → 简化授权，未来可加。

## Migration Plan

1. 更新`plugin.yaml`菜单与 i18n；提升插件版本。
2. 新增/改造前端页面与路由；删除或架空旧 console/catalog 管理菜单。
3. 扩展列表/审核查询 API 与归属过滤。
4. 重启/同步插件菜单，验证发布者角色与管理员角色菜单差集。
5. 更新 E2E；修复路由与表单校验。
6. 回滚：恢复旧菜单声明与页面 path（插件版本回退），数据模型无破坏性迁移。

## Open Questions

- 消费侧“已发布插件浏览/下载”是否仍需要独立菜单：本变更默认**不**作为管理目录子菜单；若产品需要，可在后续变更用`market:plugin:view`挂只读目录页。
- 发布者是否允许一人多`publisher_key`：沿用现有模型；“我的插件”聚合当前用户名下全部发布者。
