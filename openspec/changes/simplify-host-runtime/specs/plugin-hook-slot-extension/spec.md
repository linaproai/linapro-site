## ADDED Requirements

### Requirement: 生产分发不得解释演示钩子动作

系统 SHALL 把`insert`、`sleep`、`error`视为测试或演示动作。生产钩子分发 MUST NOT 按这三项执行插行、休眠或主动报错。真实扩展 MUST 走类型化回调或宿主服务调用。

#### Scenario: 生产路径遇到演示动作

- **WHEN** 已启用插件的清单仍声明`action: sleep`或等价演示动作
- **THEN** 生产分发跳过或拒绝该动作
- **AND** 不得阻塞宿主主链路
- **AND** 测试可以用假注册器覆盖这三项语义
