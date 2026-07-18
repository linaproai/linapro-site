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

## Git 平台 Token 配置

登记 Git 源时若未填写发布者个人访问令牌，市场会回退读取插件业务配置中的平台级 Token，用于调用 GitHub/Gitee API 做元数据发现（列 tag、读 tree、读 `plugin.yaml`），以规避未认证限流。

| 配置键 | 用途 |
|--------|------|
| `github.accessToken` | GitHub PAT（Classic：`public_repo`；Fine-grained：`Contents: Read`） |
| `gitee.accessToken` | Gitee 个人访问令牌（可选） |

配置来源（独占优先级，不合并）：

1. 宿主 `config.yaml` 的 `plugin.linapro-plugin-marketplace` 段  
2. 生产配置根下 `plugins/linapro-plugin-marketplace/config.yaml`  
3. 开发默认 `manifest/config/config.yaml`  

模板见 `manifest/config/config.example.yaml`。平台 Token **不会**写入发布者凭证表，也不会被后续 API 回显；发布者表单中的 `accessToken` 仍优先于平台配置。

宿主配置示例：

```yaml
plugin:
  linapro-plugin-marketplace:
    github:
      accessToken: "ghp_xxxxxxxx"
    gitee:
      accessToken: ""
```

## 验证

在本目录运行插件包冒烟测试：

```bash
go test ./... -count=1
```
