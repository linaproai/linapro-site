## ADDED Requirements

### Requirement: 字典核心模型是值与标签

系统 SHALL 把字典数据的核心契约定义为值与标签。排序、状态、备注和租户覆盖可以保留。公开创建、更新、查询、导入和导出 MUST NOT 把`tagStyle`、`cssClass`或等价前端标签样式字段当作核心模型。默认工作台若需要彩色标签，MUST 在工作台侧映射。

#### Scenario: 创建字典数据

- **WHEN** 管理员创建一条字典数据
- **THEN** 请求和响应以`value`与`label`为核心
- **AND** 宿主不得要求`tagStyle`或`cssClass`

#### Scenario: 导入导出模板

- **WHEN** 管理员下载或导入字典数据模板
- **THEN** 模板核心列是值与标签
- **AND** 不得把标签样式列当作必填核心列
