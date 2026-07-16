## 1. 数据模型与迁移

- [x] 1.1 新增幂等 SQL 迁移：插件/Git 源字段（`source_kind`、`repo_url`、`repo_provider`、`credential_ref`、同步状态时间戳等）及既有 upload 行回填
- [x] 1.2 扩展 release/artifact 模型以支持 git `source_ref`、`tar.gz` 产物类型或 content-type 区分
- [x] 1.3 设计并实现私有仓 token 安全存储（加密或宿主密钥引用），确保 DAO/API 不落明文回显
- [x] 1.4 重新生成或更新插件 DAO/DO/Entity（按插件本地 `hack/config.yaml` 与项目代码生成约定）

## 2. 上传通道：zip / tar.gz 与扫描

- [x] 2.1 扩展上传入口接受 `.zip`、`.tar.gz`、`.tgz`，统一解压安全边界（路径穿越、大小、解压炸弹）
- [x] 2.2 复用并扩展源码包扫描器以支持 tar 容器，保持目录规范校验
- [x] 2.3 复用并扩展动态包扫描器以支持 tar 容器，保持 wasm/清单一致性校验
- [x] 2.4 上传路径强制 `source_kind=upload`，拒绝 git 插件混用上传
- [x] 2.5 补充源码/动态 zip 与 tar.gz 扫描与上传单测

## 3. Git 元数据发现

- [x] 3.1 实现 GitHub/Gitee 提供商客户端：列表 tags、读取 raw `plugin.yaml`、可选目录存在性抽样（白名单主机）
- [x] 3.2 实现 `DiscoverGitMetadata`：版本一致性校验、`type=source` 限制、草稿写入、不落全量源码
- [x] 3.3 实现登记 Git 源 API/服务：归属校验、credential 绑定、登记后立即发现
- [x] 3.4 实现定时轮询任务（jobcap/cron）：扫描全部 git 源、增量 tag、失败状态可观测
- [x] 3.5 动态类型/不支持场景诊断信息与错误码（英文源文本）
- [x] 3.6 补充 Git 发现、版本不一致、认证失败、禁止混源单测（可用 fake HTTP 提供商）

## 4. 查询契约：distribution 与可见性

- [x] 4.1 在版本详情/列表项或专用 `GET .../distribution` 返回 `distribution` 投影
- [x] 4.2 `mode=git` 返回 `repoUrl`/`ref`/`provider`/`requiresAuth`，永不回传 token
- [x] 4.3 `mode=https` 返回产物类型、`sha256`、下载会话约定字段
- [x] 4.4 在数据库查询阶段注入可见性/下载权限过滤，无权不泄露坐标
- [x] 4.5 审核通过/下架/同步后刷新读模型与缓存失效点
- [x] 4.6 补充 distribution 与可见性单测

## 5. 发布工作台前端

- [x] 5.1 「我的插件」增加登记 Git 源表单（URL、可选 token、可见性等）
- [x] 5.2 上传流支持 `zip`/`tar.gz` 选择与校验提示
- [x] 5.3 列表展示 `source_kind`、同步状态、最新发现版本摘要
- [x] 5.4 详情/安装引导展示 `distribution` 摘要（git 坐标或 HTTPS 校验信息）
- [x] 5.5 更新插件 `manifest/i18n` 中文/英文 UI 文案与 apidoc 翻译

## 6. CLI / linactl 安装

- [x] 6.1 确定命令命名（`marketplace.install` 或扩展 `plugins.install`）并注册跨平台入口
- [x] 6.2 实现读取市场 API `distribution` 的客户端（含鉴权与错误处理）
- [x] 6.3 `mode=git`：clone/fetch 指定 ref 到 `apps/lina-plugins/<plugin-id>/`，衔接或复用现有 plugins 工作区逻辑
- [x] 6.4 `mode=https`：创建下载会话、下载、sha256 校验、解压/落盘
- [x] 6.5 源码插件提示重建部署；动态包提取 wasm 并说明后续本地上传治理步骤
- [x] 6.6 记录跨平台影响（macOS/Linux/Windows）并补充命令级测试或可重复验证步骤
- [x] 6.7 更新 `linactl` README 中英文说明

## 7. 测试与治理门禁

- [x] 7.1 扩展或新增市场 E2E（TC 编号遵循 `lina-e2e`：在既有 TC001–TC003 后续分配，如 TC004 Git 登记与发现、TC005 tar.gz 上传与 distribution 展示）
- [x] 7.2 运行市场插件相关单元测试、前端检查、`make plugins.check`（或项目约定等价命令）
- [x] 7.3 运行 `openspec validate simplify-plugin-marketplace-distribution --strict`
- [x] 7.4 在任务记录中写明 i18n、缓存一致性、数据权限、dev-tooling 跨平台与 DI 影响结论（无新增运行期 DI 时显式无影响）
- [ ] 7.5 实现完成后调用 `lina-review` 做变更审查

## Feedback

- [x] **FB-1**: 添加插件弹窗中「上传压缩包」应作为表单字段与「分发方式」左标签对齐，右侧为提示与上传区域

## 8. 任务完成影响记录

- [x] 8.1 i18n：有影响；已更新市场插件 zh-CN/en-US UI 文案，API 错误码英文源文本已补充
- [x] 8.2 缓存一致性：有影响；审核通过仍走既有读模型重建；Git 同步写入插件同步状态字段
- [x] 8.3 数据权限/可见性：有影响；distribution 查询复用 resolveAccessiblePlugin；token 不回显
- [x] 8.4 开发工具跨平台：有影响；marketplace.install 使用 stdlib HTTP + 本地 git，已在 README 中英文说明
- [x] 8.5 DI：无新增运行期宿主 DI；市场服务仍 `New(nil)` 自建 artifact store；Jobs 注册使用同一构造路径
- [x] 8.6 验证：`GOWORK=off go test ./backend/...`（marketplace 插件）通过；`linactl` build 通过；`openspec validate ... --strict` 通过
