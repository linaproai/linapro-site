## ADDED Requirements

### Requirement: Manifest文档说明资源路径语义
文档`7200-manifest.md` SHALL 说明`ManifestService`的路径语义：路径相对`manifest/`目录，使用斜杠分隔。

#### Scenario: 路径语义清晰
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解`profile.yaml`、`config/config.example.yaml`、`sql/*.sql`等路径的含义

### Requirement: Manifest文档说明资源管线边界
文档 SHALL 说明`ManifestService`的边界：只读取原始资源，不负责让配置、SQL或语言包生效。

#### Scenario: 管线边界清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解读取原文与资源生效是不同的关注点

### Requirement: Manifest文档包含主要能力概览
文档 SHALL 以表格形式简要列出`Get`、`Exists`、`Scan`三个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
