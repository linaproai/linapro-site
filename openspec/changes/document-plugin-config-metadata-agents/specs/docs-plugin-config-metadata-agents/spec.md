## ADDED Requirements

### Requirement: 插件配置文档

中文官网 SHALL 提供面向插件开发者的插件配置说明，准确介绍插件独立配置文件的职责、文件位置、读取优先级和使用方式。

#### Scenario: 展示插件配置读取顺序

- **WHEN** 开发者阅读插件配置文档
- **THEN** 文档 SHALL 说明插件配置优先读取生产配置根下`plugins/<plugin-id>/config.yaml`
- **AND** 文档 SHALL 说明开发期默认配置来自`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`
- **AND** 文档 SHALL 说明动态插件在缺少外部配置文件时可使用当前有效发布`artifact`携带的`manifest/config/config.yaml`
- **AND** 文档 SHALL 说明`manifest/config/config.example.yaml`只是配置模板，不作为运行时默认值

#### Scenario: 解释解耦配置设计收益

- **WHEN** 开发者阅读插件配置设计说明
- **THEN** 文档 SHALL 说明插件业务配置不应写入宿主主配置文件
- **AND** 文档 SHALL 说明插件配置与宿主配置解耦可以降低插件启停、部署覆盖和个性化配置对主框架的侵入
- **AND** 文档 SHALL 说明插件应通过插件作用域服务读取自己的配置

### Requirement: 插件manifest资源文档

中文官网 SHALL 提供插件`manifest`资源的读取说明，准确介绍路径语义、资源边界和个性化设计用法，并说明具体文件名没有框架级特殊语义。

#### Scenario: 展示manifest资源读取模型

- **WHEN** 开发者阅读插件`manifest`资源文档
- **THEN** 文档 SHALL 说明插件通过`HostServices.Manifest()`读取当前插件`manifest/`下的原始资源
- **AND** 文档 SHALL 使用相对`manifest/`的路径示例，例如`profile.yaml`或`resources/policy.yaml`
- **AND** 文档 SHALL 说明调用路径不应写成`manifest/profile.yaml`
- **AND** 文档 SHALL 说明具体自定义资源文件名只是插件自定义命名示例，不具备框架级特殊语义

#### Scenario: 展示源码插件和动态插件读取边界

- **WHEN** 开发者阅读插件`manifest`资源来源说明
- **THEN** 文档 SHALL 说明源码插件读取当前插件嵌入文件系统或开发目录中的`manifest/`资源
- **AND** 文档 SHALL 说明动态插件读取当前有效发布`artifact`中携带的`manifest/`资源快照
- **AND** 文档 SHALL 说明动态插件必须通过`service: manifest`和`resources.paths`声明并获得授权后才能读取对应路径
- **AND** 文档 SHALL 说明源码插件和动态插件都不能读取宿主、其他插件、绝对路径、URL或路径穿越目标

#### Scenario: 展示manifest资源边界

- **WHEN** 开发者阅读插件`manifest`资源边界说明
- **THEN** 文档 SHALL 说明`manifest/config/`、`manifest/sql/`和`manifest/i18n/`分别属于配置、安装脚本和国际化专用管线
- **AND** 文档 SHALL 说明这些目录下的文件可以作为原始资源通过`Manifest()`读取，但读取它们不会替代或触发对应专用管线的运行时生效逻辑
- **AND** 文档 SHALL 说明插件可以使用普通声明型`YAML`文件表达分类、能力开关、外部系统描述或页面个性化参数

### Requirement: AI项目规范管理文档

中文官网 SHALL 介绍`LinaPro`渐进式`AI Coding`项目规范管理策略，并说明`make agents`如何兼容主流`AI Coding`工具。

#### Scenario: 展示全局规范与渐进式规则

- **WHEN** 开发者阅读`AI`原生设计或开发工具文档
- **THEN** 文档 SHALL 说明项目根目录`AGENTS.md`是主项目规范入口
- **AND** 文档 SHALL 说明`.agents/rules/*.md`用于承载按领域渐进披露的细则
- **AND** 文档 SHALL 说明`AI Agent`在开发过程中按需求读取命中的规则文件

#### Scenario: 展示插件本地规范优先级

- **WHEN** 开发者阅读插件相关`AI Coding`规范说明
- **THEN** 文档 SHALL 说明插件可以在`apps/lina-plugins/<plugin-id>/AGENTS.md`定义本地规范
- **AND** 文档 SHALL 说明开发该插件目录内文件时，本地`AGENTS.md`优先于全局项目规范中冲突的规则
- **AND** 文档 SHALL 说明未被插件本地规范覆盖的部分仍继续遵守全局规范和命中的规则文件

#### Scenario: 展示make agents桥接能力

- **WHEN** 开发者阅读`make agents`相关文档
- **THEN** 文档 SHALL 说明`make agents`会将`.agents/skills`、`.agents/prompts`和根目录`AGENTS.md`桥接到不同`AI Coding`工具约定路径
- **AND** 文档 SHALL 说明原生读取`AGENTS.md`或`.agents/skills`的工具会被识别为原生支持，不需要重复创建对应软链
- **AND** 文档 SHALL 保留常用命令示例，帮助开发者按单个工具执行`link`或`unlink`

### Requirement: 中文主稿优先且不同步英文

本次文档变更 SHALL 只更新中文主稿，避免在中文内容审阅前修改英文或其他语言文档。

#### Scenario: 保持i18n目录不变

- **WHEN** 本次变更完成
- **THEN** `apps/lina-site/i18n/`目录 SHALL 没有因本次变更新增、修改、重命名或删除文件
- **AND** 最终说明 SHALL 明确英文文档尚未同步，等待中文审阅后再处理
