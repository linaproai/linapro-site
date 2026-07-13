## ADDED Requirements

### Requirement: 市场发布必须支持 Git 源与上传包双通道

系统 SHALL 支持两种插件发布通道：`git`（登记 GitHub 或 Gitee 仓库地址）与`upload`（上传插件压缩包）。每种插件市场主记录 MUST 绑定唯一`source_kind`（`git` 或`upload`），同一`pluginId`后续版本 MUST NOT 混用另一种来源。两种通道产生的版本草稿 MUST 进入既有人工审核状态机，审核通过后方可对具备权限的消费者可见为已发布版本。

#### Scenario: 登记 Git 源创建插件

- **WHEN** 发布者以合法仓库 URL 登记 Git 源并绑定发布者归属
- **THEN** 系统创建或更新`source_kind=git`的市场插件记录
- **AND** 触发一次元数据发现
- **AND** 不在服务端持久化该仓库的完整源码工作树

#### Scenario: 上传包创建版本草稿

- **WHEN** 发布者对`source_kind=upload`的插件上传合法压缩包
- **THEN** 系统解析包内插件信息并保存版本草稿与产物元数据
- **AND** 版本处于可提交审核的草稿或可替换草稿状态

#### Scenario: 禁止同一插件切换发布来源

- **WHEN** 插件已绑定`source_kind=git`
- **AND** 发布者尝试以上传包方式为其创建版本
- **THEN** 系统拒绝该请求
- **AND** 不写入新版本或产物

### Requirement: Git 源只做元数据发现且登记后立即同步并定时轮询

系统 SHALL 对`source_kind=git`的插件通过 GitHub/Gitee 平台 API 发现版本标签并读取远程`plugin.yaml`等元数据。系统 MUST NOT 为发现流程执行完整`git clone`或将仓库全量文件镜像到服务端对象存储。系统 MUST 在登记成功后立即执行一次发现，并 MUST 以可配置间隔定时轮询已登记 Git 源的 tags。新发现的 tag MUST 生成或刷新**未发布**草稿；已发布版本 MUST 保持不可变。

#### Scenario: 登记后立即发现 tags

- **WHEN** 发布者成功登记可访问的 Git 源
- **AND** 远程存在符合规则的版本标签
- **THEN** 系统在登记响应完成前后完成至少一次发现
- **AND** 为每个新 tag 生成对应版本草稿或更新既有可变草稿

#### Scenario: 定时轮询发现新 tag

- **WHEN** 定时任务运行且某 Git 源出现新的有效 tag
- **THEN** 系统为该 tag 创建新的版本草稿
- **AND** 不覆盖已审核发布的同版本记录

#### Scenario: 服务端不落全量源码

- **WHEN** 系统对 Git 源执行元数据发现
- **THEN** 服务端仅持久化仓库坐标、tag/ref、清单快照与同步状态等元数据
- **AND** 不得将完整 backend/frontend 源码树作为市场 artifact 存储

### Requirement: Git 版本标签必须与 plugin.yaml version 一致

系统 SHALL 在 Git 元数据发现时校验版本标签与远程根级`plugin.yaml`的`version`字段语义一致（允许 tag 带或不带`v`前缀的规范化比较）。不一致时，系统 MUST NOT 允许该版本进入可提交审核的合格草稿，或 MUST 将草稿标记为不可提交并给出明确错误。

#### Scenario: tag 与清单版本一致

- **WHEN** 远程 tag 为`v1.2.0`且`plugin.yaml`的`version`规范化后同为`1.2.0`或`v1.2.0`
- **THEN** 系统可保存该版本草稿供后续提交审核

#### Scenario: tag 与清单版本不一致

- **WHEN** 远程 tag 为`v1.2.0`且`plugin.yaml`的`version`为`1.0.0`
- **THEN** 系统拒绝将该版本视为可提交审核的合格草稿
- **AND** 同步状态或诊断信息说明版本不一致

### Requirement: Git 源 MVP 仅服务源码插件且仓库根为插件根

系统 SHALL 将 Git 发布通道的 MVP 范围限定为源码插件：远程`plugin.yaml`的`type` MUST 为`source`。系统 MUST 将仓库根目录视为插件根目录，MUST NOT 在 MVP 中支持 monorepo 子目录配置。动态插件 MUST 通过上传包通道发布。

#### Scenario: 源码插件 Git 草稿可创建

- **WHEN** 远程`plugin.yaml`声明`type: source`且通过版本一致性与最小结构检查
- **THEN** 系统允许创建 Git 来源的源码插件版本草稿

#### Scenario: 动态类型不走 Git 源

- **WHEN** 远程`plugin.yaml`声明`type: dynamic`或无法满足源码 Git 最小检查
- **THEN** 系统不生成可发布的 Git 动态版本
- **AND** 诊断信息提示动态插件应使用上传包通道

#### Scenario: 不支持 monorepo 子目录

- **WHEN** 发布者尝试登记带插件子目录参数的 Git 源
- **THEN** 系统拒绝该参数或忽略并文档化不支持
- **AND** 仅按仓库根目录解释插件内容

### Requirement: 上传通道必须支持 zip 与 tar.gz 并强制目录规范

系统 SHALL 接受`zip`与`tar.gz`（含`.tgz`）格式的市场上传包，并在解压后按插件类型强制目录规范。源码包 MUST 包含完整源码插件约定目录与`plugin.yaml`。动态运行时包 MUST 包含根级`plugin.yaml`与`plugin.wasm`，且 MUST NOT 包含`frontend/`、`backend/`、`hack/`或`main.go`等开发源码目录。解压过程 MUST 阻断路径穿越与超限膨胀。

#### Scenario: 上传合法 tar.gz 源码包

- **WHEN** 发布者上传符合源码插件目录规范的`.tar.gz`包
- **THEN** 系统解析`plugin.yaml`并保存草稿与产物
- **AND** 产物校验和可供后续 HTTPS 下载校验

#### Scenario: 动态包缺少 plugin.wasm

- **WHEN** 发布者上传声称动态类型但缺少`plugin.wasm`的压缩包
- **THEN** 系统拒绝保存为合格动态版本草稿
- **AND** 错误指出缺失文件

#### Scenario: 压缩包路径穿越被拒绝

- **WHEN** 上传包内条目包含`../`等越界路径
- **THEN** 系统拒绝该包
- **AND** 不写入产物存储

### Requirement: 已发布版本查询必须返回 distribution 分发投影

系统 SHALL 在版本详情或专用分发查询接口中返回`distribution`对象，供 CLI 安装使用。当版本来自 Git 源时，`distribution.mode` MUST 为`git`，并 MUST 包含`repoUrl`与`ref`，MUST NOT 包含平台代持的明文访问令牌。当版本来自上传包时，`distribution.mode` MUST 为`https`，并 MUST 提供创建下载会话所需的产物类型与`sha256`（或等价校验信息）。对调用方不可见或未发布且无发布/审核特权的版本，系统 MUST NOT 泄露完整分发坐标。

#### Scenario: Git 版本返回 git 分发信息

- **WHEN** 有权限用户查询已发布 Git 来源版本的分发信息
- **THEN** 响应`distribution.mode`为`git`
- **AND** 包含可克隆的`repoUrl`与对应`ref`
- **AND** 不包含服务端保存的 token 明文

#### Scenario: 上传版本返回 https 分发信息

- **WHEN** 有权限用户查询已发布上传来源版本的分发信息
- **THEN** 响应`distribution.mode`为`https`
- **AND** 包含产物类型与`sha256`等校验字段
- **AND** 客户端仍须通过既有下载会话获取包体（除非设计明确的等价受控 URL）

#### Scenario: 无权用户不能读取私有版本分发信息

- **WHEN** 用户无权访问某私有插件版本
- **AND** 请求该版本的分发信息
- **THEN** 系统拒绝请求
- **AND** 不返回`repoUrl`、下载会话或校验和

### Requirement: 私有 Git 仓库凭证必须由平台代持且不可回显

系统 SHALL 允许发布者为私有 Git 源配置访问令牌，并由平台安全存储（加密或宿主密钥能力），插件记录仅保存凭证引用。元数据发现任务 MUST 使用该凭证访问私有仓库 API。任何市场查询、审核投影与错误消息 MUST NOT 返回令牌明文。

#### Scenario: 配置私有仓 token 后可发现

- **WHEN** 发布者为私有仓库配置有效 token 并登记 Git 源
- **THEN** 元数据发现可列出 tags 并读取`plugin.yaml`
- **AND** 后续读接口不返回 token 明文

#### Scenario: token 失效时同步失败可观测

- **WHEN** 私有仓 token 失效或被撤销
- **AND** 定时或立即发现运行
- **THEN** 系统记录可展示的同步失败状态（如认证失败）
- **AND** 不在日志或 API 中打印 token

### Requirement: CLI 必须按 distribution 模式拉取插件到本地

系统 SHALL 通过`linactl`（或等价跨平台 CLI 入口）提供基于市场分发信息的安装能力。当`distribution.mode=git`时，CLI MUST 使用 git 将指定`ref`拉取到本地源码插件工作区约定路径。当`distribution.mode=https`时，CLI MUST 经受控下载获取包体并校验`sha256`后落到本地。CLI MUST NOT 绕过市场可见性与下载权限；源码插件 MUST NOT 被实现为生产宿主运行时一键安装。

#### Scenario: CLI 安装 Git 来源源码插件

- **WHEN** 用户使用 CLI 安装某已发布 Git 来源源码插件版本
- **THEN** CLI 读取`distribution`中的`repoUrl`与`ref`并完成 git 拉取
- **AND** 内容落入`apps/lina-plugins/<plugin-id>/`或文档约定的等价工作区路径
- **AND** 提示需要重新构建部署后由本地治理发现

#### Scenario: CLI 安装上传来源插件包

- **WHEN** 用户使用 CLI 安装某已发布`https`分发版本
- **THEN** CLI 创建或使用下载会话获取包体
- **AND** 校验`sha256`成功后解压或保存到本地约定位置

#### Scenario: 分发信息缺失时安装失败

- **WHEN** 目标版本无有效`distribution`或用户无权获取
- **THEN** CLI 以非零退出码失败
- **AND** 不在本地写入不完整插件目录作为成功结果

### Requirement: 我的插件发布入口必须覆盖 Git 登记与包上传

系统 SHALL 在“我的插件”发布流中提供登记 Git 源与上传`zip`/`tar.gz`两种明确动作。发布相关写操作 MUST 继续校验发布者归属，并在任何草稿、版本或产物写入前拒绝越权请求。

#### Scenario: 发布者登记 Git 源

- **WHEN** 发布者在“我的插件”选择登记仓库并提交合法表单
- **THEN** 系统创建 Git 来源插件并触发元数据发现
- **AND** 列表可展示同步状态或已发现版本摘要

#### Scenario: 发布者上传压缩包版本

- **WHEN** 发布者在“我的插件”为上传型插件选择合法`zip`或`tar.gz`并上传
- **THEN** 系统保存版本草稿
- **AND** 发布者可继续提交审核

## MODIFIED Requirements

### Requirement: 市场必须支持源码插件市场包

系统 SHALL 支持上传源码插件市场包。源码插件市场包 MUST 是包含完整源码插件目录的`.zip`或`.tar.gz`容器，根目录 MUST 包含`plugin.yaml`、`backend/`、`frontend/`、`manifest/`和`plugin_embed.go`。市场审核 MUST 校验插件 ID 命名、版本唯一性、源码目录结构、`manifest/sql/`、`manifest/i18n/`、`manifest/docs/`入口和框架版本依赖。

#### Scenario: 源码插件市场包审核通过

- **WHEN** 发布者上传源码插件市场包
- **AND** 包内目录结构、`plugin.yaml`、版本唯一性、文档入口和依赖声明均满足规范
- **THEN** 市场保存草稿版本和源码产物
- **AND** 审核结果包含源码结构、SQL、i18n、文档和依赖摘要

#### Scenario: 源码插件包缺少源码结构

- **WHEN** 发布者上传源码插件市场包
- **AND** 包内缺少`backend/`、`frontend/`、`manifest/`或`plugin_embed.go`
- **THEN** 市场审核失败或拒绝合格草稿
- **AND** 错误指出缺失的源码插件目录或文件

#### Scenario: 源码插件包支持 tar.gz 容器

- **WHEN** 发布者上传符合目录规范的源码插件`.tar.gz`包
- **THEN** 系统按与 zip 等价的校验语义处理
- **AND** 保存草稿版本与产物校验信息

### Requirement: 市场必须支持动态插件运行时市场包

系统 SHALL 支持上传动态插件运行时市场包。动态插件运行时市场包 MUST 是`.zip`或`.tar.gz`容器，根目录 MUST 包含`plugin.wasm`和根级`plugin.yaml`，并 MAY 包含`manifest/docs/`和`README`文档。动态运行时市场包 MUST NOT 包含`frontend/`、`backend/`、`hack/`或`main.go`等开发源码目录。市场审核 MUST 校验`plugin.wasm`文件头、ABI、内嵌清单、根级清单一致性、`hostServices`、动态路由、SQL、i18n 资源和文档入口。动态插件 MVP MUST 通过上传通道发布，MUST NOT 依赖 Git 全量拉取`plugin.wasm`。

#### Scenario: 动态运行时包清单一致

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 根级`plugin.yaml`与`plugin.wasm`内嵌清单的关键字段一致
- **THEN** 市场保存草稿版本、运行时产物和审核摘要
- **AND** 审核摘要包含 ABI、`hostServices`、路由、SQL、i18n 和文档信息

#### Scenario: 动态运行时包清单不一致

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 根级`plugin.yaml`与`plugin.wasm`内嵌清单的`id`、`version`、`type`、`dependencies`、`hostServices`或多租户字段不一致
- **THEN** 市场审核失败或拒绝合格草稿
- **AND** 市场不得发布该版本

#### Scenario: 动态运行时包包含开发源码目录

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 包内包含`frontend/`、`backend/`、`hack/`或`main.go`
- **THEN** 市场拒绝合格动态草稿
- **AND** 错误说明动态运行时包不得携带开发源码目录
