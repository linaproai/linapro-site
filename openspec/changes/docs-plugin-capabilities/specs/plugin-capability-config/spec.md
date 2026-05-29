## ADDED Requirements

### Requirement: Config文档说明插件配置读取优先级
文档`7000-config.md` SHALL 说明`ConfigService`的配置读取优先级：生产覆盖 > 开发默认 > 动态插件产物默认。

#### Scenario: 优先级清晰
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解三层配置覆盖关系和`manifest/config/config.example.yaml`不参与运行时默认读取

### Requirement: Config文档说明宿主配置白名单边界
文档 SHALL 说明`HostConfigService`的白名单设计：插件只能读取宿主公开的配置键，不能扫描完整配置树。

#### Scenario: 白名单边界清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解`HostConfig`与`Config`的职责划分，以及为什么插件不应使用`g.Cfg()`

### Requirement: Config文档包含两个服务的方法概览
文档 SHALL 分别以表格形式列出`ConfigService`和`HostConfigService`的主要方法。

#### Scenario: 两个服务的方法可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能区分插件配置和宿主配置各自的可用方法
