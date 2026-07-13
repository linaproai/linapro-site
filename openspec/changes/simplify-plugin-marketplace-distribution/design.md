## Context

`linapro-plugin-marketplace` 已作为内置源码插件交付，具备发布者、插件草稿、ZIP 上传、人工审核、目录检索、文档、风险摘要和 HTTPS 下载会话。发布链路偏重「先本地打包再上传」，消费侧统一走对象存储流，缺少 Git 仓库作为一等发布源，也缺少按来源分支的 CLI 安装契约。

本变更在市场插件内简化分发模型，不改 `lina-core` 插件治理主链路。用户确认的约束：

| 决策 | 取值 |
|------|------|
| Git 存储 | 只存坐标 + 元数据 |
| 服务端是否拉全量代码 | 否，只做元数据发现 |
| 同步时机 | 登记后立即发现 + 定时轮询 tags |
| 插件类型 | 源码 + 动态（动态以上传包为主） |
| 审核 | 保持人工审核 |
| 消费执行方 | CLI / `linactl` |
| 私有仓 | 支持平台代持 token |
| monorepo | MVP 不做 |
| Tag 与 yaml version | 必须一致 |

## Goals / Non-Goals

**Goals:**

- 发布双通道：Git 元数据发现、上传 `zip`/`tar.gz`。
- 两种通道均强制插件目录规范，平台可自动识别内容。
- 版本草稿仍进入既有审核状态机；已发布版本不可变。
- 对外查询暴露统一 `distribution` 投影，CLI 可决定 git 或 HTTPS。
- Git 轮询与首次同步不落全量源码树到服务端磁盘。
- 凭证安全：token 加密或受控密钥存储，API/日志不回传明文。

**Non-Goals:**

- 不实现 monorepo `subdir`。
- 不通过 Git 源自动发布动态 `plugin.wasm`（动态包走上传）。
- 不引入 webhook（可后续增强）。
- 不取消人工审核、不取消发布者归属。
- 不让市场或生产宿主运行时一键写入 `apps/lina-plugins`。
- 不实现支付/订单/许可证服务。
- 不修改 `plugin.yaml` 的 `distribution=managed|builtin` 语义。

## Decisions

### 决策一：发布来源 `source_kind` 成为版本/插件级一等字段

| `source_kind` | 含义 | 服务端产物 | 消费 `distribution.mode` |
|---------------|------|------------|---------------------------|
| `git` | 绑定 GitHub/Gitee 仓库 | 元数据快照 + 版本行，无全量包体 | `git` |
| `upload` | 上传压缩包 | artifact 对象 + 校验和 | `https` |

插件首次发布绑定 `source_kind`；同一插件后续版本默认同源。若产品上需要「同一 pluginId 先 git 后改上传」，MVP **禁止混用**，降低校验与 CLI 分支复杂度。实现上在插件主记录固定 `source_kind`，新版本校验一致性。

替代方案：每版本可不同来源 → 灵活但列表/安装提示混乱，拒绝。

### 决策二：Git 元数据发现只走平台 API，不 clone

```
登记 repo_url + provider + credential_ref(可选)
        │
        ▼
规范化 URL（仅允许 github.com / gitee.com 等白名单主机）
        │
        ▼
List tags（GitHub/Gitee REST）
        │
        ▼
对每个候选 tag：
  GET raw plugin.yaml @ tag
  解析 id/name/version/type/deps/...
  校验 tag 与 version 一致、type=source、目录约定的远程可读文件（至少 plugin.yaml）
        │
        ▼
写入/更新 release 草稿 + manifest_snapshot
（不下载 backend/frontend 全树，不落 git working tree）
```

说明：

- 「目录规范」在 Git 模式下以**可远程验证的最小集合**为准：至少根级 `plugin.yaml`；能通过 Contents/Raw API 抽样检查 `backend/plugin.go`、`plugin_embed.go`、`manifest/` 是否存在则做存在性校验。完整源码结构的最终权威仍在 CLI `git checkout` 后由本地治理/构建暴露。
- 动态插件 tag 若 `type=dynamic` 或仅含 wasm 线索：MVP **跳过或标为不支持**，提示改用上传包。
- 定时任务：可配置间隔（默认建议 15–30 分钟），扫描所有 `source_kind=git` 且未删除的插件；增量以「已知 tag 集合」去重；已发布 tag 不覆盖；新 tag 生成新草稿。
- 首次登记：同步路径与定时任务共用同一 `DiscoverGitMetadata` 服务方法。

替代方案：shallow clone 到 temp 再扫 → 违背「不同步完整代码到服务端」；拒绝。

### 决策三：上传通道扩展 `tar.gz`，复用现有扫描语义

- 容器：`application/zip`、`application/gzip`（`.tar.gz` / `.tgz`）。
- 安全：路径穿越拒绝、单文件/总大小上限、解压炸弹防护，与现有 zip 策略对齐。
- 源码包 / 动态包目录规范沿用既有市场规范（`add-plugin-marketplace`）：
  - 源码：`plugin.yaml`、`backend/`、`frontend/`、`manifest/`、`plugin_embed.go` 等
  - 动态：`plugin.yaml` + `plugin.wasm`，禁止开发源码目录；根级与内嵌清单关键字段一致
- 上传成功 → 草稿 + artifact + 文档/风险索引 → 人工审核。

### 决策四：`distribution` 查询契约（CLI 唯一安装输入）

版本详情与「安装元数据」接口返回：

```json
{
  "pluginId": "example-plugin",
  "version": "v1.0.0",
  "pluginType": "source",
  "distribution": {
    "mode": "git",
    "repoUrl": "https://github.com/org/example-plugin.git",
    "ref": "v1.0.0",
    "provider": "github",
    "requiresAuth": true
  }
}
```

或：

```json
{
  "distribution": {
    "mode": "https",
    "artifactType": "source_zip",
    "sha256": "...",
    "sizeBytes": 12345,
    "downloadSessionRequired": true
  }
}
```

规则：

- 仅对调用方**可见且已发布（或具备下载权限）**的版本返回完整分发信息。
- `mode=git`：**永不**在 API 中返回平台代持的 token；`requiresAuth=true` 时 CLI 使用用户本地 git credential / 环境变量。
- `mode=https`：继续走短期下载会话 + content 流；CLI 创建会话后下载并校验 `sha256`。
- 源码插件：CLI 落到 `apps/lina-plugins/<plugin-id>/`，提示重建部署；不调用运行时一键安装。
- 动态插件：CLI 下载包后提取 `plugin.wasm`，引导既有动态上传治理（可先落本地文件 + 打印下一步，或调用宿主 API，以实现阶段选型为准，优先复用现有上传入口）。

### 决策五：凭证与私有仓

- 发布者在登记 Git 源时可选提交 token；服务端写入**加密凭证表或宿主密钥能力**，插件表只存 `credential_ref`。
- 元数据发现任务使用 `credential_ref` 访问私有仓 API。
- 轮换：发布者可更新/清除 token；清除后私有仓发现失败并记录 `sync_status=auth_failed`。
- 审核员可见同步错误摘要，不可见 token。

### 决策六：审核与发布流保持，入口变薄

```
Git：登记仓库 → 自动发现版本草稿 → 发布者提交审核 → 通过上架
上传：选包上传 → 自动解析草稿 → 发布者提交审核 → 通过上架
```

- 插件展示名/摘要等：优先 `plugin.yaml` 与 docs 推导，表单仅保留可见性、分类等市场字段。
- 状态机不改：`draft/submitted/reviewing/approved/rejected` 与 `draft/published/delisted/deprecated`。

### 决策七：CLI 落点

- 优先新增或扩展 `linactl` 市场相关子命令（命名实现期确定，例如 `marketplace.install` 或扩展 `plugins.install` 增加 `market` 源）。
- 复用现有 `plugins.install` 的「clone/fetch + 拷贝到 `apps/lina-plugins` + lock 文件」模式处理 `mode=git`。
- `mode=https` 新增下载与解压路径，跨平台使用 Go 标准库。
- 开发工具规则：`dev-tooling` 命中，任务需记录跨平台验证。

### 决策八：数据模型增量（示意）

在插件 SQL 中新增幂等迁移（PostgreSQL-only），建议字段：

**plugin 或独立 git_source 表**

- `source_kind`：`git` | `upload`
- `repo_url`、`repo_provider`（`github`|`gitee`）
- `credential_ref`、`last_sync_at`、`last_sync_status`、`last_sync_message`

**release**

- `source_ref`（git tag/ref；upload 可为空）
- `distribution_mode` 冗余或运行时由 `source_kind` 推导

**artifact**

- `artifact_type` 扩展允许 `source_tar_gz` / `dynamic_tar_gz`（或统一 `package` + `content_type`）

已有 upload 数据默认 `source_kind=upload`、`distribution_mode=https`。

### 决策九：API 增量（示意）

| 能力 | 方法/路径方向 |
|------|----------------|
| 登记 Git 源 / 创建 git 插件 | `POST .../market/plugins/git-sources` 或扩展 `POST .../market/plugins` 带 `sourceKind` |
| 手动触发元数据发现（可选运维） | `POST .../plugins/{id}/git-sync`（发布者）；定时任务为主 |
| 上传 | 既有 releases 上传支持 tar.gz |
| 安装元数据 | `GET .../plugins/{id}/releases/{version}/distribution` 或嵌入详情/版本项 |

权限标签沿用 `market:plugin:publish` / `view` / `download` / `review`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Git 模式无法在服务端做完整源码树扫描 | 远程存在性抽样 + 审核人工抽查；完整校验在 CLI 拉取后/构建时 |
| 平台 API 限流 | 轮询退避、按插件错峰、缓存 ETag/条件请求 |
| Tag 被 force-push | 已发布版本不可变；新发现若 sha 变化仅影响未发布草稿或记录警告 |
| Token 泄露 | 加密存储、权限隔离、审计、禁止响应回显 |
| 双通道增加前端复杂度 | 「我的插件」用两个清晰动作：登记仓库 / 上传包 |
| 与未归档 `add-plugin-marketplace` / 菜单重构变更并行 | 本变更基于当前插件代码演进；归档顺序由维护者协调，增量 SQL 序号接续 |

## Migration Plan

1. 部署含新 SQL 的市场插件版本；既有 upload 行回填 `source_kind=upload`。
2. 启用 Git 发现任务（可先配置关闭，验证后再开）。
3. 发布带 `distribution` 的 API；旧客户端忽略未知字段可兼容。
4. 发布 `linactl` 市场安装命令；文档说明 git/https 两种路径。
5. 回滚：停用 Git 登记入口与定时任务；upload 路径保持可用；新列可保留。

## Open Questions

1. Git 提供商 MVP 是否严格白名单 `github.com` + `gitee.com`（建议是）。
2. 定时轮询默认间隔与是否允许发布者自定义（建议全局配置，不按插件定制）。
3. `linactl` 命令最终命名：`marketplace.install` vs 扩展 `plugins.install`（实现前在 dev-tooling 任务中敲定）。
4. 动态包 HTTPS 安装是「只下载到本地」还是「CLI 直接调宿主动态上传 API」（建议：先下载+校验+打印/可选上传参数）。
