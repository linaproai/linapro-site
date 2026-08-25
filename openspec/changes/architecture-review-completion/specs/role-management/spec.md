## ADDED Requirements

### Requirement: 角色菜单授权只暴露菜单编号列表

系统 SHALL 在角色创建、更新和详情中使用菜单编号列表表达授权。宿主 MUST NOT 要求调用方先消费树选择或已勾选集合页面形状。

#### Scenario: 读取角色详情

- **WHEN** 客户端获取一个角色
- **THEN** 响应包含`menuIds`
- **AND** 不包含`checkedKeys`树勾选结构
