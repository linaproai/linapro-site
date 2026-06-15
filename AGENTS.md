# 仓库用途

本仓库承载`LinaPro`官方网站（`apps/lina-site/`，基于`Docusaurus 3.10`）以及配套的 `OpenSpec` 治理文件。`LinaPro`本体——也就是官网所描述的 AI 驱动全栈框架 —— 位于单独的只读参考仓库`linapro`；编写站点文案时，应以该路径中的事实内容为依据。

官网发布地址为`https://linapro.ai/`，源仓库地址为`https://github.com/linaproai/linapro`，对应的本地仓库地址为：`/Users/john/Workspace/github/linaproai/linapro`。

`AGENTS.md`是指向本文件的符号链接 —— `Claude Code`、`Codex`以及其他`AI Coding Agent`会读取`AGENTS.md`的智能体应共享同一套说明。

# 架构说明

## 站点结构（`apps/lina-site/`）

- `docusaurus.config.ts`——站点配置入口。`docs.routeBasePath: '/'`表示**文档直接挂在站点根路径**，不存在单独的`/docs`前缀；根路由首页由`src/pages/index.tsx`覆盖。站点级`SEO`文案和`locale`元信息集中在`siteI18n.ts`，配置文件只消费该小型映射；导航、侧边栏、首页组件等普通界面文案遵循`Docusaurus`官方翻译机制，通过`write-translations`生成并维护在`i18n/<locale>/`的`JSON`资源中。
- `docs/`——官网文档的中文主稿目录，也是 `Docusaurus` 内容发现的基准目录；当前标签为`v0.1.x(Latest)`。即便当前`defaultLocale`仍为`en`，也必须先在这里维护中文内容，再同步到英文和其他语言版本。旧版本位于`versioned_docs/`、`versioned_sidebars/`和`versions.json`，由`node version.js`生成。`editUrl`目前指向 GitHub 上的`linaproai/linapro`——后续站点源码预期会回到 `LinaPro` 主 `monorepo` 中。
- `i18n/`——国际化资源目录。当前`defaultLocale: 'en'`，可用 `locale` 为`['en', 'zh-Hans']`。英文文档位于`i18n/en/docusaurus-plugin-content-docs/current/`，用于覆盖默认英文根路径；其他语言资源分别位于`i18n/<locale>/docusaurus-plugin-content-docs/`、`docusaurus-plugin-content-blog/`、`docusaurus-theme-classic/`以及`code.json`。`zh-Hans`站点文档默认复用`docs/`中文主稿，普通界面翻译保留在`code.json`、`docusaurus-theme-classic/*.json`和`docusaurus-plugin-content-docs/current.json`。
- `sidebars.ts`——定义三个显式侧边栏：`quickSidebar`（快速体验）、`mainSidebar`（开发手册）和`communitySidebar`（开源社区）。导航栏中的`Get Started`使用`quickSidebar`，入口为`/quick/preface`，快速开始正文仍位于`docs/quick/quickstart.md`并发布到`/quickstart`。
- `src/pages/index.tsx`——自定义首页；`src/components/HomepageFeatures/`和`src/theme/`中包含对经典主题的 `swizzle` 内容。仓库已启用 `Mermaid`（`themes: ['@docusaurus/theme-mermaid']`、`markdown.mermaid: true`），文档中的架构图和流程图应优先使用 `Mermaid`。
- 当前启用的插件包括：`docusaurus-plugin-image-zoom`（缩放选择器为`.markdown :not(em) > img`）以及`@docusaurus/plugin-ideal-image`（开发环境下禁用）。

## OpenSpec 治理（`openspec/`）

本仓库使用 `OpenSpec` 管理需求提案、设计和落地过程，目录结构如下：

- `openspec/changes/<change-id>/`——进行中的提案目录，包含`proposal.md`、`design.md`、`tasks.md`和`specs/`。
- `openspec/changes/archive/`——已完成并归档的变更。
- `openspec/specs/`——变更归档后沉淀的基线规范。
- `openspec/config.yaml`——项目级语言与生成策略配置。

涉及 `OpenSpec` 工作流时，优先使用`opsx:propose`、`opsx:explore`、`opsx:apply`、`opsx:archive`技能（或其`openspec-*`别名），不要手工拼装整套变更文件。

# 文档规范

- 编写官网文档时，需要严格遵守该文件的规范`linapro-site/.agents/instructions/markdown-format.instructions.md`。
- 编写官网文档时，优先在`apps/lina-site/docs/`维护中文主稿，确认中文内容后再同步到`apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/`中的英文版本。即便当前`defaultLocale`仍为`en`，也必须遵守“中文先行、英文同步”的流程。
- 中文文档内容符合中文习惯的表达方式，同时保持内容的准确性和专业性，不能生硬表达含义。
- 英文版本需要保持为地道的英文表达，不要逐字翻译中文内容。

# 测试规范

使用`E2E`测试工具（如`playwright`）测试时，生成的临时中间内容，如截图、日志等，应该保存在`temp/e2e/`目录下。
