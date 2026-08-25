## MODIFIED Requirements

### Requirement: 当前用户消息详情必须由 notify 读取

系统 SHALL 把当前用户收件箱详情、列表、已读和删除的数据访问放在`notify`。用户消息 HTTP 控制器 MAY 做当前用户身份改写和分类 i18n，MUST NOT 再经独立`usermsg`服务直接查投递表。

#### Scenario: 预览一条收件箱消息

- **WHEN** 已登录用户请求自己的一条消息详情
- **THEN** `notify`在查询阶段按投递用户和租户约束读取
- **AND** 找不到或越权时返回与现有收件箱一致的未找到语义
- **AND** 控制器只负责当前用户 ID 与分类文案改写
