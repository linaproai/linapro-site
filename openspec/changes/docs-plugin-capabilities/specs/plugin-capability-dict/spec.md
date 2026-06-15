## ADDED Requirements

### Requirement: Dict文档说明字典标签解析能力
文档`6950-dict.md` SHALL 说明`Dict()`用于解析字典值标签，返回`LabelProjection`并支持是否解析当前语言标签。

#### Scenario: 字典读取边界清晰
- **WHEN** 插件开发者阅读字典能力文档
- **THEN** 理解`ResolveLabels`、`ResolveInput`、`LabelKey`和`Label`的职责

### Requirement: Dict文档说明刷新管理命令
文档 SHALL 说明字典刷新通过`Admin().Dict().Refresh`执行，且动态插件没有独立`dict`服务。

#### Scenario: 字典刷新入口清晰
- **WHEN** 插件开发者需要刷新字典投影
- **THEN** 文档引导其使用可信源码插件管理命令
