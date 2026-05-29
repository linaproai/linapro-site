## ADDED Requirements

### Requirement: Notify文档说明通知发布模型
文档`7300-notify.md` SHALL 说明`NotifyService`的通知发布模型：插件通过`SendNoticePublication`将通知扇入宿主收件箱管线。

#### Scenario: 发布模型可理解
- **WHEN** 插件开发者阅读设计思路
- **THEN** 理解通知从插件业务源扇入宿主统一收件箱的设计

### Requirement: Notify文档说明SourceType和CategoryCode设计
文档 SHALL 说明`SourceType`和`CategoryCode`的分类设计及其对通知路由的影响。

#### Scenario: 分类设计清晰
- **WHEN** 插件开发者阅读设计约束
- **THEN** 理解`SourceTypeNotice`和`SourceTypePlugin`的区别，以及`CategoryCode`的收件箱分类作用

### Requirement: Notify文档包含主要能力概览
文档 SHALL 以表格形式简要列出`SendNoticePublication`和`DeleteBySource`两个方法的职责。

#### Scenario: 方法概览可查阅
- **WHEN** 插件开发者查看主要能力章节
- **THEN** 能快速了解每个方法的用途
