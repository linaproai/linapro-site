## Why

当前插件市场发布链路只支持上传 ZIP、多步手填草稿与统一 HTTPS 下载会话，发布者必须先本地打包，消费侧也无法按来源选择 git 拉取。需要把「发布来源」与「消费分发」收成清晰的双通道模型：Git 仓库元数据发现与上传包解析并存，同时保持既有人工审核与本地插件治理边界。

## What Changes

- 发布入口扩展为两种方式：
  - **Git 源**：登记 GitHub/Gitee 仓库地址（公开仓或平台代持 token 的私有仓）；服务端**只做元数据发现**（tags、`plugin.yaml` 等），**不**将完整源码 clone/mirror 到服务端。
  - **上传包**：支持 `zip` / `tar.gz`；解压后按插件目录规范识别并读取插件信息；源码包与动态运行时包均支持。
- Git 同步策略：**登记后立即发现一次 + 后台定时轮询 tags**；发现的新版本进入草稿，仍须人工审核后上架。
- 版本一致性：Git tag 与检出路径上的 `plugin.yaml` version **必须一致**，否则该版本不可提交审核。
- **MVP 不做 monorepo 子目录**；仓库根目录即插件根目录。
- **Git 源服务源码插件形态**；动态插件以上传包为主。
- 查询/详情/版本 API 增加统一 `distribution` 投影，供 CLI 按模式安装：
  - `mode=git`：返回 `repoUrl` + `ref`（及必要鉴权提示，不回传明文 token）
  - `mode=https`：返回受控下载信息 + `sha256`
- 消费侧以 **CLI / `linactl`** 为安装执行方：按 `distribution.mode` 执行 git 拉取或 HTTPS 下载到本地工作区，再复用既有本地插件治理；市场服务与 `lina-core` 不负责在生产宿主运行时一键写入源码工作区。
- 保留现有发布者归属、人工审核状态机、已发布版本不可变、可见性过滤与下载会话（HTTPS 通道继续使用）。
- 发布工作台增加「登记 Git 源」与「上传 zip/tar.gz」入口，减少与包内清单重复的手填字段依赖（能从元数据推导的优先推导）。

## Capabilities

### New Capabilities

- 无（分发简化落在既有 `plugin-marketplace` 能力上演进）。

### Modified Capabilities

- `plugin-marketplace`：扩展发布来源（Git 元数据发现 / 上传包）、产物与版本模型、`distribution` 查询契约、定时元数据同步，以及 CLI 消费分发约定；目录规范与人工审核要求保持并细化。

## Impact

- **插件代码**：`apps/lina-plugins/linapro-plugin-marketplace/` 后端 API/服务、SQL、前端「我的插件」发布流、i18n/apidoc、单元与 E2E。
- **API**：新增 Git 源登记与元数据刷新相关接口；发布上传支持 `tar.gz`；版本/详情响应增加 `distribution`；下载会话仅用于 `mode=https` 版本。
- **数据库**：插件或版本表增加来源类型、仓库 URL、凭证引用、最近同步状态等字段；token 不得明文落库日志。
- **定时任务**：市场插件内（或宿主 job 能力）登记「Git 元数据轮询」任务，只打平台 API/元数据读取，不落全量源码树。
- **开发工具**：`hack/tools/linactl` 扩展市场安装路径（可复用/衔接现有 `plugins.install` 思想）：根据市场 API 的 `distribution` 执行 git 或 HTTPS。
- **不改**：`lina-core` 插件发现/安装/启用/禁用/升级主契约；`distribution=managed|builtin` 语义；支付/订单；monorepo subdir；Git 源上的动态 wasm 全量拉取。
- **影响分析**：
  - `i18n`：有影响（新菜单/表单/错误码/apidoc）。
  - 缓存一致性：有影响（同步发现新版本、审核通过后读模型失效）。
  - 数据权限/可见性：有影响（私有仓元数据与私有插件可见性、token 仅发布者/平台配置可读）。
  - 开发工具跨平台：有影响（`linactl` git/HTTPS 路径需 macOS/Linux/Windows 验证）。
  - 测试：需补充 Git 元数据发现、tar.gz 上传、distribution 契约与 CLI 分支的单测/E2E 或命令级测试。
