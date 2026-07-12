# LinaPro 插件市场

`linapro-plugin-marketplace`是`LinaPro`私有化插件市场流程的内置源码插件。

该插件在本目录内维护市场专属后端代码、前端资源、`SQL`、运行时多语言资源、接口文档多语言资源、市场展示文档和插件自有测试。它不改变`lina-core`已有插件安装、启用、动态运行时校验或源码插件重新构建边界。

## 当前范围

| 范围 | 边界 |
|------|------|
| 分发模式 | `type: source`和`distribution: builtin` |
| 运行时语言 | `i18n.enabled: true`，以`en-US`作为源内容并维护`zh-CN`翻译 |
| 源码交付 | 下载的源码插件必须放入`apps/lina-plugins/<plugin-id>`并通过宿主重新构建部署 |
| 动态交付 | 下载的动态插件包必须复用现有本地动态插件上传治理 |

## 验证

在本目录运行插件包冒烟测试：

```bash
go test ./... -count=1
```
