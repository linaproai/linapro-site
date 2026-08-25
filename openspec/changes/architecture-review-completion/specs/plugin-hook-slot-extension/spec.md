## ADDED Requirements

### Requirement: 生产校验拒绝演示钩子动作

系统 SHALL 在生产清单校验和分发中拒绝`insert`、`sleep`、`error`演示钩子动作。测试可以用假注册器覆盖这些动作。官方文档 MUST NOT 再把这些动作教成正式扩展点。

#### Scenario: 插件清单声明 action insert

- **WHEN** 生产路径校验该钩子规格
- **THEN** 校验失败
- **AND** 不得进入生产分发解释器
