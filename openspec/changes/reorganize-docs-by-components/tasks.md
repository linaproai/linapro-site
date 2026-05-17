## 1. 内容盘点与迁移准备

- [x] 1.1 盘点 `2000-architecture/`、`3000-features/`、`4000-plugin-development/` 下所有中文主稿的 front matter、`slug`、标题、站内链接和主要章节
- [x] 1.2 对照 `/Users/john/Workspace/github/linaproai/linapro` 主仓库 README、源码目录和插件清单，确认组件边界与事实口径
- [x] 1.3 制定旧文档到 `2000-components/` 目标文件的迁移映射，标记需要合并、删减或保留引用的章节

## 2. 组件设计目录结构

- [x] 2.1 在 `apps/lina-site/docs/docs/2000-components/` 下建立扁平化目标文件结构，确保存在 `2000-components.md`
- [x] 2.2 使用千位递增文件名前缀组织目标文档，不创建 `2000-lina-core/`、`3000-lina-vben/` 或 `5000-distributed/` 子目录
- [x] 2.3 为 `2000-components.md` 设置稳定 front matter，并使用新的 `/docs/components` 入口 `slug`

## 3. 组件文档整合

- [x] 3.1 编写 `2000-components.md`，直接介绍框架组件、组件地图、核心职责和协作关系
- [x] 3.2 整合 `3000-core-host.md`，覆盖核心宿主、配置管理、接口文档、定时任务、`I18N`、多租户宿主侧能力和扩展接缝
- [x] 3.3 整合 `4000-admin-workspace.md`，说明 `lina-vben` 的定位、技术栈、功能模块和插件菜单动态注入机制
- [x] 3.4 整合 `5000-plugin-system.md`，说明双模式插件系统的设计理念、治理主链、生命周期、隔离机制、多租户字段和宿主边界
- [x] 3.5 整合 `6000-source-plugins.md`，说明源码插件的适用场景、目录结构、注册机制、开发流程和最佳实践
- [x] 3.6 整合 `7000-wasm-plugins.md`，说明 `WASM`动态插件的沙箱模型、桥接协议、构建、安装、运行和运维方式
- [x] 3.7 整合 `8000-plugin-management.md`，说明插件来源、工作区初始化、安装、升级、状态查看、启用、禁用和卸载流程
- [x] 3.8 整合 `9000-distributed-architecture.md`，说明单机/集群部署、`Redis`协调、选主、分布式锁、键值缓存和任务调度防重复机制

## 4. 链接、入口与旧目录收敛

- [x] 4.1 保留已存在组件相关页面的原有 `slug`，避免 `/docs/core-host`、`/docs/plugin-system` 等稳定链接失效
- [x] 4.2 更新中文主稿中的站内链接和相关文档引用，使其指向整合后的组件设计页面或保留的稳定 `slug`
- [x] 4.3 移除或改写旧的 `2000-architecture/`、`3000-features/`、`4000-plugin-development/` 中文入口，确保侧边栏不继续呈现重复分类
- [x] 4.4 确认本次变更没有新增、修改、删除或重命名 `apps/lina-site/i18n/` 下的任何文件

## 5. 文档质量与验证

- [x] 5.1 按 `.agents/instructions/markdown-format.instructions.md` 审查所有新增和整合后的 Markdown 文档格式
- [x] 5.2 审查整合后的中文表达，确保内容自然、通顺、易于理解，并删除机械重复或过时描述
- [x] 5.3 运行站点构建或文档链接检查，验证 Docusaurus 能正常发现 `2000-components/` 文档且无明显断链
- [x] 5.4 运行 `openspec validate reorganize-docs-by-components --strict`，确认本变更规范有效

## Feedback

- [x] **FB-1**: `2000-components.md` 不应说明“为什么按组件组织”等目录导读内容，应直接介绍 `LinaPro` 框架组件信息
- [x] **FB-2**: 旧功能特性目录下的配置管理、接口文档、多租户、定时任务和 `I18N` 国际化内容被过度压缩，需要补回为组件设计专题页
- [x] **FB-3**: `2000-components/` 下不应再出现插件系统等子目录，所有组件设计文档需要保持同层扁平结构
