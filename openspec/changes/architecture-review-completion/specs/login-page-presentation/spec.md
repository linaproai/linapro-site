## ADDED Requirements

### Requirement: 登录页布局由默认工作台本地偏好承担

默认管理工作台 SHALL 自行决定登录面板左右布局和 slogan 插画。宿主公开配置 MUST NOT 再提供这些字段作为登录页渲染前置。

#### Scenario: 登录页在缺少宿主布局字段时仍可渲染

- **WHEN** `GET /config/public/frontend`不含`panelLayout`和`sloganImage`
- **THEN** 默认工作台登录页仍能展示账密登录
- **AND** 不得注册已关闭的验证码或二维码登录路由
