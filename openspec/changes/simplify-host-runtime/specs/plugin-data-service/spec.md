## ADDED Requirements

### Requirement: RecordStore 必须标明为实验能力

系统 SHALL 将动态插件记录存储标为可选实验能力。官方插件主路径文档 MUST NOT 把它写成动态插件访问数据的默认方式。在没有第一方插件真正按表访问之前，MUST NOT 继续扩展查询计划器作为内核主路径。

#### Scenario: 阅读动态插件数据访问文档

- **WHEN** 开发者阅读`pkg/plugin`公开说明
- **THEN** 文档标明`RecordStore`是实验能力
- **AND** 主路径仍以授权 host service 与插件自有存储为准
