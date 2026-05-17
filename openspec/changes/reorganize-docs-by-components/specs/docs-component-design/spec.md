## ADDED Requirements

### Requirement: 组件设计目录必须统一承载组件视角文档
系统 SHALL 使用 `apps/lina-site/docs/docs/2000-components/` 作为中文主稿中组件设计内容的统一目录，替代按“架构设计、功能特性、扩展开发”分散组织同一组件内容的方式。

#### Scenario: 用户浏览中文文档侧边栏
- **WHEN** 用户查看中文文档中的组件设计相关章节
- **THEN** 系统在 `2000-components` 目录下展示组件设计内容
- **AND** 同一组件的设计理念、核心架构、功能特性、扩展开发接缝和使用方式不会分散在旧的架构、功能和扩展开发目录中

### Requirement: 组件设计目录必须提供组件概览文档
系统 SHALL 在 `apps/lina-site/docs/docs/2000-components/2000-components.md` 提供“组件设计”概览文档，用于直接介绍 `LinaPro` 框架组件、核心职责和协作关系。

#### Scenario: 用户进入组件设计目录
- **WHEN** 用户打开组件设计目录入口
- **THEN** 系统展示 `2000-components.md` 的正文内容
- **AND** 页面直接说明 `LinaPro` 的主要框架组件、组件职责和组件之间的协作关系
- **AND** 页面不介绍为什么按组件组织文档、推荐阅读路径或与其他文档目录的关系

### Requirement: 组件设计文件必须按千位递增排序
系统 MUST 使用千位递增的文件名前缀组织 `2000-components/` 下的 Markdown 文档，确保 Docusaurus 自动侧边栏顺序稳定且易于维护。

#### Scenario: 开发者新增或重命名组件设计文档
- **WHEN** 开发者在 `2000-components/` 下维护 Markdown 文件
- **THEN** 文件名使用 `2000`、`3000`、`4000` 等千位递增前缀
- **AND** 不使用 `2100`、`2200` 等百位细分前缀作为排序编号

### Requirement: 组件设计目录不得为指定组件创建子目录
系统 MUST 将 `lina-core`、`lina-vben` 和分布式架构相关文档直接放在 `2000-components/` 下，不得创建 `2000-lina-core/`、`3000-lina-vben/` 或 `5000-distributed/` 等子目录。

#### Scenario: 文档重组后检查组件设计目录结构
- **WHEN** 开发者检查 `apps/lina-site/docs/docs/2000-components/`
- **THEN** `lina-core`、`lina-vben` 和分布式架构相关文档直接位于该目录下
- **AND** 目录中不存在 `2000-lina-core/`、`3000-lina-vben/` 或 `5000-distributed/` 子目录

### Requirement: 组件文档内容必须经过整合重写
系统 SHALL 将迁移到组件设计目录的中文内容按组件主线整合重写，使文档表达符合中文技术文档习惯，并避免重复搬运旧章节。

#### Scenario: 用户阅读某个组件设计页面
- **WHEN** 用户打开核心宿主、管理工作台、插件系统、源码插件、`WASM`动态插件、插件管理或分布式架构页面
- **THEN** 页面围绕该组件本身组织内容
- **AND** 正文以自然、通顺、易于理解的中文说明组件的设计理念、实现细节、能力边界和使用方式
- **AND** 页面不以机械拼接旧文档段落的方式重复呈现相同信息

### Requirement: 组件文档必须保持 URL 稳定
系统 SHALL 在重组中文主稿目录时保留已有组件相关页面的稳定 `slug`，降低站内链接、外部链接和搜索索引失效风险。

#### Scenario: 用户访问旧组件相关链接
- **WHEN** 用户访问重组前已经存在的组件相关 `slug`
- **THEN** 系统仍能展示对应主题的文档内容
- **AND** 文档目录结构变化不会导致这些稳定链接直接失效

### Requirement: 本次组件重组不得修改 i18n 内容
系统 MUST NOT 在本次组件设计重组中修改 `apps/lina-site/i18n/` 目录下的任何文件。

#### Scenario: 开发者完成组件设计重组
- **WHEN** 开发者检查本次变更的文件清单
- **THEN** 变更只包含中文主稿、OpenSpec 文档或必要的中文站点结构调整
- **AND** `apps/lina-site/i18n/` 目录下没有文件被新增、修改、删除或重命名
