## ADDED Requirements

### Requirement: I18n文档说明运行时翻译设计
文档`7100-i18n.md` SHALL 说明`I18nService`的运行时翻译机制：基于请求Locale动态解析翻译键。

#### Scenario: 翻译机制可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解翻译键从请求上下文获取Locale，按插件语言包查找翻译值

### Requirement: I18n文档说明翻译键命名约定
文档 SHALL 说明翻译键的命名约定和与插件语言包的关系。

#### Scenario: 命名约定清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解翻译键的层级结构和`fallback`参数的作用

### Requirement: I18n文档包含主要能力概览
文档 SHALL 以表格形式简要列出`GetLocale`、`Translate`、`FindMessageKeys`三个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
