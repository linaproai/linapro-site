## MODIFIED Requirements

### Requirement: 插件菜单与权限过滤必须使用稳定投影

系统 SHALL 在跨模块过滤导航菜单和权限菜单时使用稳定投影结构，MUST NOT 把`SysMenu`实体传入插件编排层。

#### Scenario: 权限菜单过滤不再接收表实体

- **WHEN** 角色服务按插件启用态过滤权限菜单
- **THEN** 调用`FilterPermissionMenus`时传入`menu.FilterItem`投影
- **AND** MUST NOT 把`[]*entity.SysMenu`作为该跨模块契约的参数类型
