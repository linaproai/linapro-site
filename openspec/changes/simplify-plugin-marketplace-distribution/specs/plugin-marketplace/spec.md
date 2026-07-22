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

系统 SHALL 对`source_kind=git`的插件通过 GitHub/Gitee 平台 API 发现版本标签或分支引用，并读取远程`plugin.yaml`等元数据。系统 MUST NOT 为发现流程执行完整`git clone`或将仓库全量文件镜像到服务端对象存储。系统 MUST 在登记成功后将插件置为待验证（`pending_verify`）并纳入异步处理流水线（登记请求可为识别插件根而做最小远程探测，但完整版本发现与验证 MUST 由异步定时任务在待验证阶段内推进），并 MUST 以可配置间隔定时轮询已登记 Git 源。新发现的引用 MUST 生成或刷新**未发布**草稿；已发布版本 MUST 保持不可变。系统 MUST NOT 再将「待拉取」（`pending_fetch`）作为用户可见或可写入的处理状态。

版本引用解析 MUST 遵守：

1. 当远程存在一个或多个符合 semver 规则的版本 tag 时，系统 MUST 使用这些 tag（按提供商返回顺序或等价的新到旧顺序）进行发现，且 MUST NOT 因存在 tag 而回退到分支。
2. 当远程不存在任何符合规则的版本 tag 时，系统 MUST 回退读取`main`分支最新内容并据此生成草稿。
3. 当既无符合规则的版本 tag、且`main`分支也不存在或不可读时，系统 MUST 以明确诊断失败，且 MUST NOT 使用其他分支名作为隐式回退。

#### Scenario: 登记后入列并由异步任务发现 tags

- **WHEN** 发布者成功登记可访问的 Git 源
- **AND** 远程存在符合规则的版本标签
- **THEN** 系统在登记后将插件置为待验证（`pending_verify`）
- **AND** 异步定时任务在待验证阶段完成至少一次版本发现与校验
- **AND** 为每个新 tag 生成对应版本草稿或更新既有可变草稿

#### Scenario: 无版本 tag 时回退 main 分支

- **WHEN** 发布者登记可访问的 Git 源
- **AND** 远程不存在符合规则的版本标签
- **AND** `main`分支存在且可读
- **THEN** 系统以`main`为`source_ref`读取`plugin.yaml`并生成或刷新草稿
- **AND** 同时解析并持久化该`main`分支当时的`source_commit`（完整 commit SHA）
- **AND** 不报「repository has no version tags」类阻断错误

#### Scenario: 无 tag 且无 main 时失败

- **WHEN** 发布者登记 Git 源或触发元数据同步
- **AND** 远程不存在符合规则的版本标签
- **AND** `main`分支不存在或不可读
- **THEN** 系统返回元数据发现失败
- **AND** 诊断信息说明缺少版本标签且`main`分支不可用

#### Scenario: 定时轮询发现新 tag

- **WHEN** 定时任务运行且某 Git 源出现新的有效 tag
- **THEN** 系统为该 tag 创建新的版本草稿
- **AND** 不覆盖已审核发布的同版本记录

#### Scenario: 服务端不落全量源码

- **WHEN** 系统对 Git 源执行元数据发现
- **THEN** 服务端仅持久化仓库坐标、插件根相对路径、tag/ref、commit SHA、清单快照与同步状态等元数据
- **AND** 不得将完整 backend/frontend 源码树作为市场 artifact 存储

### Requirement: Git 版本标签必须与 plugin.yaml version 一致

系统 SHALL 在基于**版本 tag**的 Git 元数据发现时，校验该 tag 与对应插件根目录`plugin.yaml`的`version`字段语义一致（允许 tag 带或不带`v`前缀的规范化比较）。不一致时，系统 MUST NOT 允许该版本进入可提交审核的合格草稿，或 MUST 将草稿标记为不可提交并给出明确错误。当发现引用为回退的`main`分支时，系统 MUST 使用`plugin.yaml`的`version`作为草稿版本，且 MUST NOT 要求存在与版本同名的 tag。

#### Scenario: tag 与清单版本一致

- **WHEN** 远程 tag 为`v1.2.0`且对应插件根`plugin.yaml`的`version`规范化后同为`1.2.0`或`v1.2.0`
- **THEN** 系统可保存该版本草稿供后续提交审核

#### Scenario: tag 与清单版本不一致

- **WHEN** 远程 tag 为`v1.2.0`且对应插件根`plugin.yaml`的`version`为`1.0.0`
- **THEN** 系统拒绝将该版本视为可提交审核的合格草稿
- **AND** 同步状态或诊断信息说明版本不一致

#### Scenario: main 回退使用清单版本

- **WHEN** 系统因无有效 tag 而从`main`分支发现元数据
- **AND** 对应插件根`plugin.yaml`声明`version: 0.2.0`
- **THEN** 系统保存版本草稿`v0.2.0`（或规范化等价形式）
- **AND** `source_ref`为`main`
- **AND** `source_commit`为发现时`main`指向的 commit SHA

### Requirement: Git 来源版本必须钉扎 source_commit 且安装引用不可浮动

系统 SHALL 在 Git 元数据发现写入或刷新**可变** release 草稿时，解析候选 ref（semver tag 或回退的`main`）对应的完整 commit SHA，并持久化为`source_commit`。`source_ref` MAY 保留逻辑引用名（如`v1.0.0`或`main`）供展示。对已发布或不可变 release，系统 MUST NOT 因后续 main 前进、tag force-push 或再次同步而改写其`source_ref`/`source_commit`。

当返回`distribution.mode=git`时，`distribution.ref` MUST 优先使用该版本的`source_commit`；仅当`source_commit`缺失且无法解析时，才可回退到`source_ref`。系统 MUST NOT 让已登记/已发布版本的安装路径仅依赖浮动分支名`main`，以避免市场展示版本与实际检出内容不一致。

#### Scenario: main 回退版本钉扎 commit 用于安装

- **WHEN** 系统因无有效 tag 从`main`发现版本草稿并完成入库
- **AND** 发现时`main`指向 commit`abc123…`
- **THEN** 该 release 的`source_ref`为`main`
- **AND** `source_commit`为`abc123…`
- **AND** 查询该版本`distribution`时`ref`为`abc123…`（或与之等价的完整 SHA）
- **AND** MUST NOT 仅返回未钉扎的`main`作为安装引用

#### Scenario: 已发布版本不因 main 前进而改变安装引用

- **WHEN** 某 Git 来源版本已发布且`source_commit`为`abc123…`
- **AND** 远程`main`随后前进到新的 commit
- **AND** 定时同步再次运行
- **THEN** 该已发布版本的`source_commit`与`distribution.ref`保持`abc123…`
- **AND** 系统不得覆盖该已发布版本的安装坐标

#### Scenario: tag 发现同样持久化 commit

- **WHEN** 系统基于 semver tag`v1.0.0`发现元数据
- **AND** 该 tag 当时指向 commit`def456…`
- **THEN** release 保存`source_ref=v1.0.0`与`source_commit=def456…`
- **AND** `distribution.ref`优先为`def456…`

### Requirement: 插件必须保留历史版本记录并支持选择安装

系统 SHALL 为同一`pluginId`保留多条版本 release 历史（不同`release_version`）。新版本上架、Git 发现新 tag 或上传新包 MUST NOT 自动删除、覆盖或不可查询既有**已发布**历史版本。系统 MUST 提供按插件列出可见版本的查询能力，并 MUST 允许有权限的消费者按指定版本获取`distribution`或下载会话，以便在兼容性问题时安装历史版本进行回退。

#### Scenario: 多版本并存可查询

- **WHEN** 同一插件存在多个已发布版本（例如`v1.0.0`与`v1.1.0`）
- **AND** 有权限用户查询该插件的版本列表
- **THEN** 响应包含上述历史版本条目
- **AND** 不因存在更新的`v1.1.0`而隐藏`v1.0.0`

#### Scenario: 消费者安装指定历史版本

- **WHEN** 有权限用户请求某已发布历史版本（非最新）的分发信息或下载
- **THEN** 系统返回该指定版本的`distribution`或下载会话
- **AND** 安装/下载内容对应该历史版本（Git 源使用其钉扎的`source_commit`；上传包使用其 artifact）
- **AND** 不静默替换为最新版本

#### Scenario: 新版本发布不抹除历史

- **WHEN** 审核通过并发布新版本`v1.2.0`
- **AND** 同插件先前已发布`v1.1.0`
- **THEN** `v1.1.0`仍可作为历史版本被查询与安装
- **AND** 其`source_commit`或产物校验和保持不变

### Requirement: Git 源仅服务源码插件并自动识别单插件与多插件仓库

系统 SHALL 将 Git 发布通道限定为源码插件：远程`plugin.yaml`的`type` MUST 为`source`（空类型按源码处理时仍须通过源码最小结构检查）。动态插件 MUST 通过上传包通道发布。

系统 MUST 在元数据发现时自动识别仓库布局，而无需发布者手工填写子目录：

1. **单插件仓库**：若仓库根目录存在合法源码插件`plugin.yaml`且通过最小结构检查，则插件根为仓库根，`repo_path`为空。
2. **多插件仓库**：若仓库根不是合法源码插件根，系统 MUST 通过远程目录树发现子目录中的合法源码插件根（例如`apps/lina-plugins/<plugin-id>/`或一级子目录插件），并为每个合法插件根创建或更新独立的市场插件记录，记录各自的`repo_path`（相对仓库根的插件根路径）。

登记同一仓库 URL 时，系统 MUST 为识别到的每个合法源码插件根各维护一条`source_kind=git`记录（同一发布者归属、同一`repo_url`/凭证），且 MUST 在分发投影中暴露插件根相对路径，供 CLI 从 monorepo 检出后落到正确本地路径。系统 MUST NOT 要求调用方在登记 API 中传入插件子目录参数。

#### Scenario: 源码插件 Git 草稿可创建

- **WHEN** 远程某插件根`plugin.yaml`声明`type: source`且通过版本一致性与最小结构检查
- **THEN** 系统允许创建 Git 来源的源码插件版本草稿

#### Scenario: 动态类型不走 Git 源

- **WHEN** 远程`plugin.yaml`声明`type: dynamic`或无法满足源码 Git 最小检查
- **THEN** 系统不生成可发布的 Git 动态版本
- **AND** 诊断信息提示动态插件应使用上传包通道

#### Scenario: 自动识别单插件仓库

- **WHEN** 发布者登记仓库根即为合法源码插件的 Git 源
- **THEN** 系统创建一条市场插件记录
- **AND** `repo_path`为空
- **AND** 按仓库根读取`plugin.yaml`与结构文件

#### Scenario: 自动识别多插件仓库

- **WHEN** 发布者登记一个在子目录中包含多个合法源码插件根的仓库
- **AND** 仓库根本身不是合法源码插件根
- **THEN** 系统为每个合法插件根创建或更新对应市场插件记录
- **AND** 各记录保存各自的`repo_path`
- **AND** 元数据发现分别读取各插件根下的`plugin.yaml`

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

系统 SHALL 在版本详情或专用分发查询接口中返回`distribution`对象，供 CLI 安装使用。当版本来自 Git 源时，`distribution.mode` MUST 为`git`，并 MUST 包含`repoUrl`与`ref`（`ref`优先为钉扎的`source_commit`），MUST NOT 包含平台代持的明文访问令牌。当版本来自上传包时，`distribution.mode` MUST 为`https`，并 MUST 提供创建下载会话所需的产物类型与`sha256`（或等价校验信息）。对调用方不可见或未发布且无发布/审核特权的版本，系统 MUST NOT 泄露完整分发坐标。分发查询 MUST 按请求路径中的版本定位到对应 release，不得默认改写为最新版本。

#### Scenario: Git 版本返回 git 分发信息

- **WHEN** 有权限用户查询已发布 Git 来源版本的分发信息
- **THEN** 响应`distribution.mode`为`git`
- **AND** 包含可克隆的`repoUrl`与对应`ref`
- **AND** 当该版本存在`source_commit`时，`ref`为该 commit SHA
- **AND** 不包含服务端保存的 token 明文

#### Scenario: 上传版本返回 https 分发信息

- **WHEN** 有权限用户查询已发布上传来源版本的分发信息
- **THEN** 响应`distribution.mode`为`https`
- **AND** 包含产物类型与`sha256`等校验字段
- **AND** 客户端仍须通过既有下载会话获取包体（除非设计明确的等价受控 URL）

#### Scenario: 指定历史版本分发不漂移到最新

- **WHEN** 有权限用户查询版本`v1.0.0`的分发信息
- **AND** 同插件另有更新的已发布版本`v1.1.0`
- **THEN** 返回的`distribution.version`为`v1.0.0`
- **AND** Git 源时`ref`对应该`v1.0.0`记录的钉扎 commit，而非`v1.1.0`或当前`main` tip

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
- **THEN** 系统创建 Git 来源插件并立即出现在“我的插件”列表
- **AND** 插件处理状态为待验证（`pending_verify`）
- **AND** 完整元数据发现、验证与进审由异步定时任务在待验证阶段推进，而非阻塞登记请求完成全部流水线

#### Scenario: 发布者上传压缩包版本

- **WHEN** 发布者在“我的插件”为上传型插件选择合法`zip`或`tar.gz`并上传
- **THEN** 系统保存版本草稿并立即出现在“我的插件”列表
- **AND** 插件处理状态进入待验证（`pending_verify`）
- **AND** 验证成功后由异步任务进入待审核，而不是要求发布者在验证前手工点“提交审核”

### Requirement: 市场文档必须按用户语言环境选择并回退到英文

系统 SHALL 在插件市场详情与文档读取路径中，按调用方当前语言环境选择被发布插件版本随行的`manifest/docs`文档。文档选择 MUST 遵守：

1. 当该版本对目标文档路径仅有一种语言资源时，系统 MUST 直接返回该语言文档，不得因请求语言不同而拒绝展示。
2. 当存在多种语言资源时，系统 MUST 优先返回与用户当前语言环境匹配的文档。
3. 当无法匹配用户当前语言时，系统 MUST 优先回退到英文（`en-US` 或等价 `en`）文档。
4. 当英文也不存在时，系统 MAY 再回退到`README`入口文档；响应 MUST 标记`resolvedLocale`与`fallbackUsed`。

请求未显式传`locale`时，系统 MUST 使用请求上下文语言（如`Accept-Language`/宿主解析后的请求语言）作为首选语言。被发布插件未启用运行时`i18n`时，市场仍 MUST 能展示其`manifest/docs/<locale>/`文档。

#### Scenario: 仅一种语言文档时直接展示

- **WHEN** 用户以`en-US`查看插件版本详情文档
- **AND** 该版本仅索引到`manifest/docs/zh-CN/index.md`
- **THEN** 市场返回中文文档正文
- **AND** `resolvedLocale`为`zh-CN`
- **AND** `fallbackUsed`为`true`

#### Scenario: 多语言时匹配用户语言

- **WHEN** 用户以`zh-CN`查看插件版本详情文档
- **AND** 该版本同时存在`manifest/docs/zh-CN/index.md`与`manifest/docs/en-US/index.md`
- **THEN** 市场返回中文文档正文
- **AND** `resolvedLocale`为`zh-CN`

#### Scenario: 多语言无法匹配时回退英文

- **WHEN** 用户以`ja-JP`查看插件版本详情文档
- **AND** 该版本存在`manifest/docs/zh-CN/index.md`与`manifest/docs/en-US/index.md`，但不存在日文文档
- **THEN** 市场返回英文文档正文
- **AND** `resolvedLocale`为`en-US`
- **AND** `fallbackUsed`为`true`

### Requirement: 市场必须持久化版本文档语言包内容并支持快速切换

系统 SHALL 在版本发布包解析或 Git 元数据发现时，将每个可展示的版本文档语言内容保存为版本级文档快照。文档快照 MUST 至少包含插件 ID、版本、文档路径、语言、来源类型、标题、摘要、内容哈希和经过安全渲染后的正文内容。详情页读取文档时，系统 MUST 能在一次请求中返回目标文档路径的可用语言文档集合，使前端可以在本地快速切换语言，而不需要按语言逐个重新请求服务端。

文档语言包读取 MUST 复用既有版本可见性与发布者/审核者边界。对调用方不可见的版本，系统 MUST NOT 返回任意语言的文档快照或语言列表。Git 来源文档快照 MUST 来自发现时读取的钉扎引用内容，不得在普通详情读取路径中重新访问 Git 提供商。上传来源文档快照 MUST 来自包解析时已验证的文档内容，不得要求详情读取路径重新解压产物来恢复正文。

#### Scenario: 详情页一次读取多语言文档包

- **WHEN** 用户查看某版本的`index.md`文档
- **AND** 该版本存在`zh-CN`和`en-US`两种文档快照
- **THEN** 文档读取响应包含按当前语言规则选中的`document`
- **AND** 同时包含该路径下所有可用语言的`documents`集合
- **AND** 每个集合项包含对应语言的安全渲染正文内容

#### Scenario: 前端本地切换文档语言

- **WHEN** 详情页已经加载包含多语言内容的文档响应
- **AND** 用户切换到响应内已有的另一种语言
- **THEN** 页面使用已下载的文档快照立即切换正文
- **AND** 不为同一插件、版本、文档路径和语言再次发起文档读取请求

#### Scenario: Git 文档详情读取不回源

- **WHEN** 用户读取 Git 来源版本的文档详情
- **AND** 该版本已经在元数据发现时保存文档快照
- **THEN** 系统直接从版本文档快照返回正文
- **AND** 不在详情请求路径中调用 GitHub/Gitee 文件读取接口

### Requirement: 我的插件列表操作列仅包含详情、新版本与下架

系统 SHALL 在“我的插件”列表每一行的操作列中仅提供三个操作：

1. **详情**：打开该插件详情，展示插件基本信息、版本与文档内容。
2. **新版本**：请求服务端自动刷新/更新该插件的版本信息。对`source_kind=git`的插件，系统 MUST 触发 Git 元数据发现并刷新草稿版本；对`source_kind=upload`的插件，系统 MUST 引导发布者上传新版本包并由服务端解析后进入处理流水线。
3. **下架**：撤回已发布状态。下架成功后插件`market_status`变为已下架，公开插件市场目录 MUST NOT 再展示该插件；插件仍可在“我的插件”列表中查看。

系统 MUST NOT 在该操作列中继续提供“更多”溢出菜单、手工“发布”按钮或其他并列主操作。下架操作仅对当前已发布插件可执行。

#### Scenario: 操作列仅三个动作

- **WHEN** 发布者打开“我的插件”列表
- **THEN** 每一行操作列仅展示详情、新版本、下架
- **AND** 不展示更多菜单或发布按钮

#### Scenario: 新版本触发服务端版本信息更新

- **WHEN** 发布者对`source_kind=git`的插件点击“新版本”
- **THEN** 系统请求服务端执行 Git 元数据同步
- **AND** 同步成功后列表刷新展示最新发现版本摘要

#### Scenario: 下架后市场不可见

- **WHEN** 发布者对已发布插件点击“下架”并确认
- **THEN** 插件状态变为已下架
- **AND** 公开市场目录查询不再返回该插件

### Requirement: 我的插件列表必须单独展示来源列

系统 SHALL 在“我的插件”列表中将来源（`sourceKind`，如 Git 仓库或上传包）作为独立列展示，MUST NOT 仅以插件名称右侧的标签形式承载来源信息。

#### Scenario: 来源独立列

- **WHEN** 发布者打开“我的插件”列表
- **THEN** 表格存在“来源”列
- **AND** 插件名称列不再附带来源标签

### Requirement: 添加插件后必须进入异步处理状态机

系统 SHALL 在插件添加（Git 登记或包上传）成功后，立即将插件纳入发布者“我的插件”列表，并通过异步定时任务推进处理流水线。处理状态 MUST 覆盖并按序切换：

1. **待验证**（`pending_verify`）：添加成功后的统一入列状态。对 Git 源，本阶段内由异步任务完成元数据发现与结构/清单校验；对上传包，本阶段内校验已入库的包体与清单。
2. **待审核**（`pending_review`）：验证成功，进入人工审核队列
3. **已完成**（`completed`）：审核通过后对具备权限的消费者可见（可与`market_status=published`并存）
4. **处理失败**（`failed`）：验证或发现失败，保留诊断信息

系统 MUST NOT 再暴露或写入「待拉取」（`pending_fetch`）处理状态。系统 MUST 将上述状态投影到“我的插件”列表与详情。验证失败时 MUST 进入可观测失败状态并保留诊断信息，且 MUST NOT 静默进入已发布。审核拒绝 MUST NOT 将插件标记为已发布。公开目录仍仅展示审核通过后的已发布版本。

#### Scenario: 添加后立即可见且为待验证

- **WHEN** 发布者成功添加一个 Git 来源或上传包插件
- **THEN** “我的插件”列表立即包含该插件
- **AND** 展示状态为待验证（`pending_verify`）
- **AND** 公开市场目录尚不可见该未发布插件

#### Scenario: 待验证阶段完成发现与校验后进入待审核

- **WHEN** 异步定时任务处理一条`pending_verify`插件
- **AND**（Git 源）元数据发现成功且（两种来源）验证通过
- **THEN** 插件处理状态变为待审核（`pending_review`）
- **AND** 对应版本进入人工审核队列（`review_status=submitted` 或等价）

#### Scenario: 审核通过后已发布

- **WHEN** 审核人员批准处于待审核的版本
- **THEN** 插件/版本市场状态变为已发布（`published`）
- **AND** 处理状态标记为完成（`completed` 或等价）
- **AND** 具备权限的消费者可在公开目录看到该版本

#### Scenario: 上传包与 Git 源共享同一入列状态

- **WHEN** 发布者通过上传包或 Git 登记添加插件
- **THEN** 初始处理状态均为待验证（`pending_verify`）
- **AND** 验证成功后进入待审核，审核通过后才已发布

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
