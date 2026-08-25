## ADDED Requirements

### Requirement: 宿主公开契约不得绑定默认工作台展示形状

系统 SHALL 保持`lina-core`公开 HTTP 契约与默认 Vben 工作台的页面结构解耦。宿主 MUST NOT 把 Vben 布局词、Ant Design 标签样式类、树选择/`checkedKeys`表单形状或默认壳组件路径写进通用契约。

#### Scenario: 公开前端配置不含布局词

- **WHEN** 未登录客户端请求`GET /config/public/frontend`
- **THEN** 载荷不含`panel-left`、`sidebar-mixed-nav`、`sloganImage`

#### Scenario: 字典核心不含标签样式

- **WHEN** 客户端读写字典数据
- **THEN** 核心字段是值与标签
- **AND** 不得把`tagStyle`或`cssClass`当作宿主核心契约字段

#### Scenario: 角色授权不是树勾选 DTO

- **WHEN** 客户端读取角色已授权菜单
- **THEN** 宿主返回菜单编号列表
- **AND** 不得要求调用方消费`checkedKeys`或专用树选择 DTO 才能完成授权
