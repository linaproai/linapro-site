## ADDED Requirements

### Requirement: 官网内容架构必须围绕源码事实组织
系统 SHALL 以 LinaPro 主仓库中的主 README、插件 README 和 OpenSpec 规范为官网一期的优先内容来源，确保站点文案与源码事实一致。

#### Scenario: 编写或更新官网核心文案
- **WHEN** 开发者编写或更新首页、文档首页或插件介绍文案
- **THEN** 内容优先依据 LinaPro 主仓库的官方说明文档和规范整理
- **AND** 不允许脱离主仓库事实臆造不存在的模块、流程或能力

### Requirement: 文档信息架构必须覆盖官网一期核心主题
系统 SHALL 提供清晰的文档信息架构，使用户能从项目介绍逐步进入架构、工作台、插件和 OpenSpec 工作流。

#### Scenario: 用户进入文档站
- **WHEN** 用户打开文档首页或侧边栏
- **THEN** 文档信息架构至少覆盖 `Introduction`、`Getting Started`、`Architecture`、`Core Host`、`Web Workspace`、`Plugins`、`OpenSpec Workflow`
- **AND** 文档层级允许后续继续扩展部署、治理或 FAQ 等补充板块

### Requirement: 博客一期内容范围必须聚焦项目演进与实践
系统 SHALL 将博客一期内容聚焦于 LinaPro 的版本演进、架构设计、插件机制和 OpenSpec 实践，不将博客退化为与项目无关的泛技术资讯流。

#### Scenario: 初始化博客内容
- **WHEN** 站点准备博客栏目或首批文章
- **THEN** 首批内容主题来自 LinaPro 的真实架构主题、归档变更或实践总结
- **AND** 文章主题与 LinaPro 的框架定位、插件体系或研发流程直接相关

### Requirement: About 页面必须遵守素材真实性边界
系统 SHALL 为 About 页面预留团队与联系方式结构，但在真实资料未补齐前 MUST 使用明确占位内容而非虚构成员和联系信息。

#### Scenario: 真实团队资料尚未提供
- **WHEN** About 页面需要展示团队成员、联系方式或社区入口
- **THEN** 页面可以展示结构占位、待补充提示或已知真实外链
- **AND** 页面不得伪造成员姓名、头像、邮箱、电话或社交账号
