## ADDED Requirements

### Requirement: 插件市场必须以内置源码插件交付

系统 SHALL 将插件市场实现为`apps/lina-plugins/linapro-plugin-marketplace`源码插件。该插件的`plugin.yaml` MUST 声明`type: source`、`distribution: builtin`和`i18n.enabled: true`，并通过源码插件编译期注册绑定同一插件 ID。市场业务后端、前端、SQL、运行时语言包、接口文档翻译、展示文档和 E2E 测试 MUST 闭环在该插件目录内。

#### Scenario: 市场插件使用 builtin 源码分发

- **WHEN** 宿主扫描`linapro-plugin-marketplace`插件清单
- **THEN** 清单校验通过`type=source`和`distribution=builtin`
- **AND** 该插件存在同 ID 的编译期源码插件注册绑定
- **AND** 宿主启动引导按现有 builtin 策略安装、升级并启用该插件

#### Scenario: 普通插件管理不展示市场插件治理动作

- **WHEN** 管理员打开普通插件管理页面
- **THEN** 普通插件管理列表不展示`distribution=builtin`的市场插件
- **AND** 用户不得通过普通插件管理入口安装、禁用、卸载或手动升级市场插件

### Requirement: 市场必须维护发布者、插件、版本和产物归属

系统 SHALL 在市场插件自己的数据表中维护发布者、插件市场主记录、发布版本、发布产物、分类标签、审核状态、文档索引、风险摘要和下载会话。市场 MUST 维护插件 ID 归属；首次发布成功后，后续版本只能由拥有该插件 ID 权限的发布者发布。已发布版本产物 MUST 不可变，同一`pluginId + version`不得静默覆盖。

#### Scenario: 首次发布绑定插件 ID 归属

- **WHEN** 发布者首次发布插件`P`并审核通过
- **THEN** 市场将`P`的插件 ID 归属于该发布者
- **AND** 后续同一插件 ID 的版本发布必须校验发布者归属

#### Scenario: 已发布版本不可覆盖

- **WHEN** 插件`P`的版本`v1.0.0`已经发布
- **AND** 发布者再次上传同一`pluginId + version`的产物
- **THEN** 市场拒绝静默覆盖
- **AND** 修复必须发布新版本或在草稿审核期内替换未发布产物

### Requirement: 市场必须支持源码插件市场包

系统 SHALL 支持上传源码插件市场包。源码插件市场包 MUST 是包含完整源码插件目录的`.zip`容器，根目录 MUST 包含`plugin.yaml`、`backend/`、`frontend/`、`manifest/`和`plugin_embed.go`。市场审核 MUST 校验插件 ID 命名、版本唯一性、源码目录结构、`manifest/sql/`、`manifest/i18n/`、`manifest/docs/`入口和框架版本依赖。

#### Scenario: 源码插件市场包审核通过

- **WHEN** 发布者上传源码插件市场包
- **AND** 包内目录结构、`plugin.yaml`、版本唯一性、文档入口和依赖声明均满足规范
- **THEN** 市场保存草稿版本和源码产物
- **AND** 审核结果包含源码结构、SQL、i18n、文档和依赖摘要

#### Scenario: 源码插件包缺少源码结构

- **WHEN** 发布者上传源码插件市场包
- **AND** 包内缺少`backend/`、`frontend/`、`manifest/`或`plugin_embed.go`
- **THEN** 市场审核失败
- **AND** 错误指出缺失的源码插件目录或文件

### Requirement: 市场必须支持动态插件运行时市场包

系统 SHALL 支持上传动态插件运行时市场包。动态插件运行时市场包 MUST 是`.zip`容器，根目录 MUST 包含`plugin.wasm`和根级`plugin.yaml`，并 MAY 包含`manifest/docs/`和`README`文档。动态运行时市场包 MUST NOT 包含`frontend/`、`backend/`、`hack/`或`main.go`等开发源码目录。市场审核 MUST 校验`plugin.wasm`文件头、ABI、内嵌清单、根级清单一致性、`hostServices`、动态路由、SQL、i18n 资源和文档入口。

#### Scenario: 动态运行时包清单一致

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 根级`plugin.yaml`与`plugin.wasm`内嵌清单的关键字段一致
- **THEN** 市场保存草稿版本、运行时产物和审核摘要
- **AND** 审核摘要包含 ABI、`hostServices`、路由、SQL、i18n 和文档信息

#### Scenario: 动态运行时包清单不一致

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 根级`plugin.yaml`与`plugin.wasm`内嵌清单的`id`、`version`、`type`、`dependencies`、`hostServices`或多租户字段不一致
- **THEN** 市场审核失败
- **AND** 市场不得发布该版本

#### Scenario: 动态运行时包包含开发源码目录

- **WHEN** 发布者上传动态插件运行时市场包
- **AND** 包内包含`frontend/`、`backend/`、`hack/`或`main.go`
- **THEN** 市场审核失败
- **AND** 错误说明动态运行时包不得携带开发源码目录

### Requirement: 市场文档必须来自版本随行资源并支持语言回退

系统 SHALL 将`manifest/docs/`视为市场展示文档资源。市场文档 MUST 随插件版本交付，且不得改变插件运行时配置、安装、启用或运行时行为。读取文档时，系统 MUST 按当前语言、被发布插件`plugin.yaml`的`i18n.default`、`zh-CN`、`README.zh-CN.md`、`README.md`顺序回退。被发布插件未启用运行时`i18n`时，市场仍 MUST 能展示其`manifest/docs/<locale>/`或`README`文档。

#### Scenario: 按当前语言读取文档

- **WHEN** 用户以`en-US`语言查看插件`P`版本`v1.0.0`详情
- **AND** 该版本包存在`manifest/docs/en-US/index.md`
- **THEN** 市场返回该英文文档正文和目录
- **AND** 不读取完整包源码或其他无关语言正文

#### Scenario: 当前语言缺失时回退

- **WHEN** 用户以`en-US`语言查看插件`P`版本`v1.0.0`详情
- **AND** 该版本缺少`manifest/docs/en-US/index.md`
- **THEN** 市场按默认语言、`zh-CN`、`README.zh-CN.md`、`README.md`顺序回退
- **AND** 响应标记实际使用的文档语言和来源

#### Scenario: 文档资源不影响运行时 i18n

- **WHEN** 被发布插件未声明`i18n.enabled: true`
- **AND** 包内提供`manifest/docs/zh-CN/index.md`
- **THEN** 市场可以展示该文档
- **AND** 宿主不得因此把该插件视为已启用运行时多语言治理

### Requirement: 市场文档渲染必须安全

系统 SHALL 在渲染市场文档时阻断脚本执行和跨版本资源访问。图片和附件路径 MUST 限制在当前版本包的`manifest/docs/assets/`或文档相对目录下。外链 MAY 展示但 MUST 标记为外部链接。文档正文进入搜索索引前 MUST 移除危险标签、大型二进制内容和不可索引资源。

#### Scenario: 文档脚本被阻断

- **WHEN** 市场文档包含脚本标签或事件处理属性
- **THEN** 渲染结果不得执行脚本
- **AND** 管理工作台页面上下文不得被注入任意脚本

#### Scenario: 图片路径越界被拒绝

- **WHEN** 市场文档引用`../../secret.png`或其他越界路径
- **THEN** 市场拒绝读取该资源
- **AND** 响应不得暴露包外文件内容

### Requirement: 市场列表必须分页并读取最小投影

系统 SHALL 为市场列表提供分页查询和数量上限。列表响应 MUST 返回最小必要投影，包括插件 ID、名称、摘要、发布者、分类标签、类型、最新版本、兼容范围、风险摘要、下载统计快照和更新时间。列表响应 MUST NOT 返回完整`Markdown`正文、完整产物内容或需要前端逐项补查才能展示的基础字段。

#### Scenario: 分页查询市场列表

- **WHEN** 用户查询市场插件列表
- **THEN** API 使用分页参数并限制最大`pageSize`
- **AND** 响应包含列表项和总数
- **AND** 每个列表项不包含完整文档正文

#### Scenario: 列表读取不触发逐项包解析

- **WHEN** 市场列表返回一页插件
- **THEN** 服务端从读模型或搜索索引读取列表投影
- **AND** 不为每个列表项实时读取包存储、解析`plugin.yaml`或加载`Markdown`

### Requirement: 市场详情、文档、风险和下载必须执行可见性过滤

系统 SHALL 对公开、私有和未来付费插件执行可见性过滤。列表、详情、版本、文档、风险、下载会话和聚合统计 MUST 在数据库查询阶段注入可见性和下载权限过滤，不得先查出范围外数据再在内存中过滤。用户无权访问私有插件时，响应不得泄露该插件存在性、版本数量、下载量或审核状态。

#### Scenario: 无权用户看不到私有插件

- **WHEN** 用户查询市场列表
- **AND** 插件`P`仅对特定组织或租户可见
- **AND** 当前用户不在可见范围内
- **THEN** 列表和总数均不包含`P`
- **AND** 聚合统计不得暗示`P`存在

#### Scenario: 下载前校验可见性

- **WHEN** 用户请求创建插件`P`版本`v1.0.0`的下载会话
- **AND** 当前用户无权访问或下载该版本
- **THEN** 市场拒绝创建下载会话
- **AND** 不返回短期下载地址或产物元数据

### Requirement: 市场下载必须通过短期会话和校验和保护

系统 SHALL 通过`POST`创建下载会话。下载会话 MUST 绑定版本、产物、请求人、过期时间、授权结果和产物`sha256`校验和。短期下载地址或会话读取 MUST 在过期后失效。下载统计 MUST 通过下载事件异步聚合到快照字段，不得让普通列表、详情或文档查询产生业务写入。

#### Scenario: 创建有效下载会话

- **WHEN** 有权限用户请求下载插件`P`版本`v1.0.0`
- **THEN** 市场创建绑定请求人和产物的短期下载会话
- **AND** 响应包含会话 ID、过期时间、产物类型、大小和`sha256`

#### Scenario: 过期下载会话不可用

- **WHEN** 用户访问已过期的下载会话
- **THEN** 市场拒绝下载
- **AND** 不返回产物内容

### Requirement: 动态插件下载后必须复用本地动态上传治理

系统 SHALL 在用户下载动态插件运行时市场包后，引导扩展中心提取根目录`plugin.wasm`并调用现有动态插件上传入口进入本地插件治理。市场检索结果、审核摘要和下载包信息 MUST NOT 作为本地安装权威。本地宿主仍 MUST 校验`plugin.wasm`文件头、ABI、内嵌清单、资源、SQL、依赖和`hostServices`授权。

#### Scenario: 动态市场包导入本地治理

- **WHEN** 用户下载动态插件运行时市场包
- **AND** 扩展中心提取根目录`plugin.wasm`
- **THEN** 扩展中心调用现有动态插件上传入口
- **AND** 本地宿主返回发现版本、运行时状态和授权摘要

#### Scenario: 市场审核通过不绕过本地校验

- **WHEN** 动态插件市场版本已审核通过
- **AND** 用户将下载的`plugin.wasm`上传到本地宿主
- **THEN** 本地宿主仍执行动态插件上传校验
- **AND** 安装时仍要求依赖检查和`hostServices`确认

### Requirement: 源码插件下载不得绕过构建部署边界

系统 SHALL 将源码插件市场包下载视为源码交付，不得把源码插件下载直接接入运行时安装。源码插件下载入口 MUST 明确返回源码包或受控来源配置，并要求用户将源码放入`apps/lina-plugins/<plugin-id>/`、重新构建部署宿主，再通过现有插件治理完成发现、安装、启用或升级。

#### Scenario: 下载源码插件返回源码交付信息

- **WHEN** 用户下载源码插件市场包
- **THEN** 市场返回源码包下载信息或来源配置
- **AND** 响应说明源码插件需要进入`apps/lina-plugins/<plugin-id>/`并重新构建部署

#### Scenario: 源码插件不执行运行时一键安装

- **WHEN** 用户下载源码插件市场包
- **THEN** 市场不得调用运行时插件安装接口直接安装该源码插件
- **AND** 运行中的生产宿主不得被市场下载流程写入源码工作区

### Requirement: 市场必须展示审核风险摘要

系统 SHALL 在插件版本详情和审核后台展示风险摘要。风险摘要 MUST 覆盖`hostServices`资源授权、动态路由、菜单权限、外部网络、数据表访问、安装 SQL、卸载 SQL、mock SQL、插件依赖、多租户字段和文档入口。风险摘要 MUST 来自审核扫描结果或本地治理解析结果，不得由发布者在插件包内自行声明为权威结果。

#### Scenario: 动态插件风险摘要展示

- **WHEN** 用户查看动态插件版本详情
- **THEN** 市场展示`hostServices`、动态路由、SQL、外部网络和数据访问风险摘要
- **AND** 用户可以在下载前识别需要本地安装确认的授权范围

#### Scenario: 发布者声明不能覆盖审核结果

- **WHEN** 插件包内文档声称没有数据访问风险
- **AND** 审核扫描发现`hostServices`请求数据表访问
- **THEN** 市场以审核扫描结果作为风险摘要权威
- **AND** 详情页不得隐藏该数据访问风险

### Requirement: 市场插件必须完整支持多语言治理

系统 SHALL 对`linapro-plugin-marketplace`按启用多语言的插件治理。市场插件 API 文档源文本和业务错误 fallback MUST 使用英文，非英文翻译 MUST 位于该插件自己的`manifest/i18n/<locale>/apidoc/`和运行时语言包中。前端菜单、路由、按钮、表格、表单、提示和状态文案 MUST 使用运行时翻译资源，且不得在模块顶层直接调用`$t()`求值。

#### Scenario: API 文档翻译资源完整

- **WHEN** 市场插件新增 API DTO 或`g.Meta`文档文本
- **THEN** 英文源文本写入 API 定义
- **AND** 中文接口文档翻译写入市场插件自己的`manifest/i18n/zh-CN/apidoc/`

#### Scenario: 前端不显示原始 i18n key

- **WHEN** 用户打开市场列表、详情、发布或审核页面
- **THEN** 页面展示当前语言的翻译文本
- **AND** 关键菜单、按钮、表格列、表单项和提示信息不得显示原始翻译 key

### Requirement: 市场缓存和统计必须有明确一致性边界

系统 SHALL 明确市场列表读模型、文档渲染缓存和下载统计快照的权威数据源、失效触发点、跨实例同步策略、最大可接受陈旧时间和故障降级。权限和可见性 MUST NOT 依赖可丢失缓存放行。发布、下架、审核状态变化、分类标签变化和草稿产物替换 MUST 使受影响的读模型或文档缓存失效。

#### Scenario: 发布后列表读模型刷新

- **WHEN** 插件版本审核通过并发布
- **THEN** 市场刷新该插件的列表读模型或搜索索引
- **AND** 后续列表查询可以看到最新版本、风险摘要和文档摘要

#### Scenario: 文档缓存绑定 doc hash

- **WHEN** 市场读取版本文档
- **THEN** 文档缓存键包含`release_id`、`locale`、`path`和`doc_hash`
- **AND** 新版本发布或草稿产物替换后不得继续返回旧文档正文
