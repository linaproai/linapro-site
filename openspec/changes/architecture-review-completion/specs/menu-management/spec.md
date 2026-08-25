## ADDED Requirements

### Requirement: 菜单写入不强制图标全局唯一

系统 SHALL 允许不同目录或菜单使用相同图标。宿主 MUST NOT 把侧栏图标不重复当作完整性校验。

#### Scenario: 更新菜单为已使用的图标

- **WHEN** 管理员把菜单图标改成另一菜单已使用的图标
- **THEN** 保存成功
- **AND** 不得返回图标重复错误

### Requirement: 角色菜单授权接口不是树选择页面形状

系统 SHALL 为角色授权提供菜单编号列表。宿主 MUST NOT 把`checkedKeys`、半选集合或专用下拉树 DTO 当作角色授权的核心契约。默认工作台 MUST 用通用菜单列表自行拼树和勾选。

#### Scenario: 编辑角色已授权菜单

- **WHEN** 管理员打开角色授权
- **THEN** 宿主提供该角色的`menuIds`和可授权菜单资源
- **AND** 工作台自己生成树选择勾选状态
