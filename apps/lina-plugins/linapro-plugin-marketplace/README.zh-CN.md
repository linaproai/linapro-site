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

## 本地存储路径配置

市场上传包、Git 文档快照与受控下载内容落在本地磁盘目录，由 `storage.root` 控制。

| 配置键 | 用途 | 默认值 |
|--------|------|--------|
| `storage.root` | 制品与文档快照根目录 | `temp/plugin-marketplace/artifacts` |

相对路径按宿主 **工作区根（WorkspaceRoot）** 解析，与上传目录、动态插件产物等可写路径共用 `lina-core/pkg/runtimepath` 契约，**不**依赖进程 CWD。`make dev` 会注入 `LINAPRO_WORKSPACE_ROOT`（仓库根）与 `LINAPRO_DATA_ROOT`（`<仓库根>/temp`），因此默认落点为：

```text
<repo>/temp/plugin-marketplace/artifacts/
  <plugin-id>/
    <version>/
      docs/
        <locale>/
          index.md
          ...
      meta/
        docs-manifest.json
```

Git 文档快照在 `docs/<locale>/` 下保留原始相对文件名。每次同步会对比 content hash：内容变化才覆盖写盘；远端已删除的文档会清理本地文件。包体仍使用同一根下的 `source/`、`dynamic/` 前缀。

而不是 `apps/lina-core/temp/...`。绝对路径配置保持原样。生产环境建议使用仓库外的绝对路径。

若本地仍有旧版快照在 `apps/lina-core/temp/plugin-marketplace/`，或仍是 `docs-snapshot/.../content/<hash>.md` 布局，请重新执行 Git 元数据同步（或删除旧目录）。不提供双读兼容。

## Git 平台 Token 配置

登记 Git 源时若未填写发布者个人访问令牌，市场会回退读取插件业务配置中的平台级 Token，用于调用 GitHub/Gitee API 做元数据发现（列 tag、读 tree、读 `plugin.yaml`），以规避未认证限流。

| 配置键 | 用途 |
|--------|------|
| `github.accessToken` | GitHub PAT（Classic：`public_repo`；Fine-grained：`Contents: Read`） |
| `gitee.accessToken` | Gitee 个人访问令牌（可选） |

配置由**本插件维护**，不要写入宿主框架 `config.yaml`。来源（独占优先级，不合并）：

1. 开发默认：`apps/lina-plugins/linapro-plugin-marketplace/manifest/config/config.yaml`  
2. 生产：配置根下的 `plugins/linapro-plugin-marketplace/config.yaml`  

模板见 `manifest/config/config.example.yaml`。平台 Token **不会**写入发布者凭证表，也不会被后续 API 回显；发布者表单中的 `accessToken` 仍优先于平台配置。

插件配置示例（`manifest/config/config.yaml` 或生产插件配置文件）：

```yaml
storage:
  root: "temp/plugin-marketplace/artifacts"
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
