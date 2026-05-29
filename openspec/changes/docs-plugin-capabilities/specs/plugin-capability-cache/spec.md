## ADDED Requirements

### Requirement: Cache文档说明插件作用域隔离设计
文档`6900-cache.md` SHALL 说明`CacheService`的插件作用域隔离机制：每个插件的缓存自动绑定到当前插件ID和租户范围。

#### Scenario: 隔离机制可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解缓存命名空间自动隔离，插件间不会互相干扰

### Requirement: Cache文档说明缓存值类型和过期策略
文档 SHALL 说明`CacheItem`支持的值类型（字符串和整数）以及过期策略设计。

#### Scenario: 值类型和过期策略清晰
- **WHEN** 插件开发者查看设计约束
- **THEN** 理解`CacheValueKindString`和`CacheValueKindInt`的区别，以及`ttl=0`表示不过期

### Requirement: Cache文档包含主要能力概览
文档 SHALL 以表格形式简要列出`Get`、`Set`、`Delete`、`Incr`、`Expire`五个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
