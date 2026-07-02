---
name: linasite-sync-i18n-dirty
description: >-
  只同步当前 git 工作区中有变化的 apps/lina-site/docs/ 中文 Markdown 文档到
  apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/。
  需要手动调用该技能，禁止自动调用。
---

# linasite-sync-i18n-dirty：只同步当前变更文档

基于 `linasite-sync-i18n` 的翻译质量标准，但只处理当前 git 中有变化的中文主稿文档。不要做全量扫描、全量差异报告或顺手修复无关 i18n 文件。

**核心原则：**

1. **只看当前变更** — 同步范围仅来自 `git status` / `git diff` 中的 `apps/lina-site/docs/**/*.md`
2. **路径镜像** — `apps/lina-site/docs/<path>.md` 对应 `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/<path>.md`
3. **中文主稿为准** — 对新增、修改和重命名后的文档，读取当前工作区版本并同步英文
4. **删除需确认** — 中文主稿删除时，先报告对应英文镜像路径；只有用户确认后才删除英文文件
5. **slug 不变** — 翻译时绝不修改 frontmatter `slug`
6. **标题贴近路由** — frontmatter `title` 应尽可能采用 `slug` 或路径路由名称的自然英文形式
7. **地道英文优先** — frontmatter 标题、描述、关键词和正文标题都要符合英语技术文档习惯，不能机械直译
8. **保护无关改动** — 不修改不在当前 docs 变更清单里的文档，不回滚用户已有改动

---

## 目录路径

| 角色 | 路径 |
|------|------|
| 中文主稿 | `apps/lina-site/docs/` |
| 英文 i18n | `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/` |

---

## 工作流程

### 1. 收集当前 docs 变更

先检查中文主稿目录的 git 状态：

```bash
git status --short -- apps/lina-site/docs
```

再用 diff 补充暂存区和工作区中的重命名、删除等信息：

```bash
git diff --name-status -- apps/lina-site/docs
git diff --cached --name-status -- apps/lina-site/docs
```

把结果合并成**当前变更同步清单**，只保留 `apps/lina-site/docs/**/*.md`。忽略 `_category_.json`、图片、`.DS_Store` 和其他非 Markdown 文件。

状态处理：

| 状态 | 来源示例 | 处理方式 |
|------|----------|----------|
| 新增 | `A` / `??` | 读取中文新文件，创建英文镜像文件 |
| 修改 | `M` / `MM` / `AM` | 读取当前工作区版本，覆盖更新英文镜像文件 |
| 重命名 | `R... old new` | 同步新路径英文镜像；旧英文路径是否删除需要用户确认 |
| 删除 | `D` | 报告对应英文镜像路径；用户确认后删除英文文件 |

如果当前变更同步清单为空，直接说明没有发现需要同步的中文 Markdown 文档，不要启动全量 i18n 审查。

### 2. 输出本次同步计划

执行任何写入前，向用户展示本次范围：

```markdown
## Dirty i18n 同步计划

### 将翻译或更新
- apps/lina-site/docs/docs/1000-concepts/example.md
  -> apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/docs/1000-concepts/example.md

### 将处理重命名
- old: apps/lina-site/docs/docs/old-name.md
- new: apps/lina-site/docs/docs/new-name.md
  -> 更新新路径英文镜像；旧英文镜像待确认是否删除

### 删除待确认
- source deleted: apps/lina-site/docs/docs/obsolete.md
  -> candidate i18n delete: apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/docs/obsolete.md
```

确认规则：

- 对新增、修改、重命名后的新文件：报告后直接同步，不把“是否同步”作为可拒绝选项。
- 对删除和重命名旧路径：必须询问用户是否删除对应英文 i18n 文件。
- 如果英文目标文件已有未提交改动，先报告该路径会被覆盖；只有它属于当前变更镜像文件时才继续。

### 3. 翻译新增、修改和重命名后的文档

对每个需要写入英文镜像的中文文档：

1. 完整阅读中文源文件，理解页面定位、技术概念、结构和术语。
2. 从 frontmatter `slug` 的最后一段推导英文路由名称；没有 slug 时，从文件名去掉排序数字后推导。
3. 必要时读取同目录 1 到 3 篇已有英文文档，保持标题风格和术语一致。
4. 将整篇文档翻译或重新翻译到镜像路径；内容不同步时覆盖整文，不做零散局部 patch。
5. 保持 Markdown 结构、提示块、代码块、表格和 Mermaid 图表语法正确。
6. 确保父目录存在，再写入英文文件。

翻译要求：

| 项目 | 要求 |
|------|------|
| Frontmatter | `slug`、`sidebar_position`、`hide_title` 等结构字段保持不变；`title`、`description`、`keywords` 用自然英文表达 |
| 标题 | `title` 尽可能与 slug/路径表达的英文路由名称一致；`title` 和正文 `##` / `###` 必须准确、自然、专业，符合英语技术文档命名习惯 |
| 正文 | 以英语技术文档作者口吻重写，不逐句硬译 |
| 术语 | LinaPro、API 名称、配置键、命令、插件 ID 保持准确一致 |
| 代码块 | 不翻译代码、命令、配置键和值；除非中文注释本身是文档说明的一部分，否则不改代码块内容 |
| Mermaid | 保留图表类型和语法，翻译节点标签和边标签 |
| slug | 严禁修改 |

标题示例：

```yaml
# 中文
slug: '/docs/reference/openapi'
title: 'OpenAPI接口文档'

# 英文
slug: '/docs/reference/openapi'
title: 'OpenAPI Reference'
```

标题命名规则：

- 优先从 `slug` 推导标题候选：`/docs/architecture/core-host` → `Core Host`，`/docs/development/commands` → `Development Commands`。
- 没有 `slug` 时，从镜像路径文件名推导，先去掉 `1000-` 这类排序前缀和 `.md` 后缀。
- 路由名中的固定术语保持站点通用写法，例如 `OpenAPI`、`WASM`、`i18n`、`LinaPro`。
- 如果中文标题和路由名不完全一致，先阅读正文判断页面定位；在语义一致时优先采用路由名的自然英文形式。
- 不为了逐字对应中文而写出不自然标题，例如用 `OpenAPI Reference`，不要用 `OpenAPI Interface Documentation`。
- 不为了贴合路由而丢失必要限定词；如果正文说明页面范围更具体，可以在路由名基础上补足上下文。

### 4. 处理删除和重命名旧路径

删除中文主稿或重命名后遗留的旧英文镜像文件，不自动删除。

处理步骤：

1. 根据中文旧路径推导英文镜像路径。
2. 检查该英文文件是否存在。
3. 向用户报告候选删除文件。
4. 用户确认后才删除。
5. 如果文件不存在，在最终摘要中说明无需删除。

### 5. 验证

同步后至少执行：

```bash
git status --short -- apps/lina-site/docs apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current
```

对本次写入的每个英文镜像文件确认：

- 文件存在于预期路径
- `slug` 与中文源文件一致
- frontmatter `title` 与 `slug` / 路径的英文路由概念一致，且读起来像自然的英语技术文档标题
- Markdown frontmatter 仍有效
- 没有明显残留的中文正文内容（代码块内允许按原样保留）
- 未触碰当前变更清单之外的 i18n 文档

最终输出简短摘要：

```markdown
## Dirty i18n 同步完成

- 新建英文翻译：N
- 更新英文翻译：M
- 重命名新路径同步：R
- 删除英文镜像：D
- 跳过删除：S

### 本次处理
- apps/lina-site/i18n/en/.../example.md
```

---

## 执行守则

- **不要全量审查** — 本技能只服务当前 git 变更，除非用户明确改口要求全量同步。
- **不要修改无关文件** — 即使发现其他英文文档过期，也只报告为非本技能范围，不顺手修。
- **不要静默删除** — 删除英文 i18n 文件必须获得用户确认。
- **不要机械直译** — 标题和正文要按英语技术文档习惯自然表达；标题优先贴近 slug/路径路由名称的自然英文形式。
- **不要改 slug** — `slug` 是规范 URL，必须与中文主稿保持一致。
- **保护用户改动** — 遇到目标英文文件已有未提交改动时，先判断是否正是当前镜像文件；不要回滚任何用户改动。
