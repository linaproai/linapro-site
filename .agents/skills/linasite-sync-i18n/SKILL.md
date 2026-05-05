---
name: linasite-sync-i18n
description: >-
  检测 apps/lina-site/docs/ 中文主稿与 apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/
  英文翻译之间的缺漏和内容差异，并进行补全修复。当 docs/ 中新增或更新了文档、或需要全量同步审查时使用。
  翻译采用地道英文表达，绝不逐字翻译，确保英文版本在英语环境下自然流畅。
---

# linasite-sync-i18n：文档 i18n 同步与翻译

对中文主稿进行全面审查，确保每篇文档都有对应的地道英文翻译，并与中文内容保持同步。本技能输出自然、专业的英文文档——而非机械翻译。

**核心原则：**
1. **路径镜像** — `docs/<path>.md` 对应 `i18n/en/docusaurus-plugin-content-docs/current/<path>.md`
2. **slug 是唯一标识** — 每篇文档的 `slug` 字段是规范 URL，翻译时绝不修改
3. **地道英文优先** — 先理解内容，再以英文习惯表达；绝不逐字翻译
4. **中文主稿为准** — `docs/` 是唯一事实来源，英文版跟随中文版更新

---

## 目录路径

| 角色 | 路径 |
|------|------|
| 中文主稿 | `apps/lina-site/docs/` |
| 英文 i18n | `apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/` |

---

## 工作流程

### 1. 扫描中文主稿

列出中文主稿目录中的所有 Markdown 文件：

```bash
find apps/lina-site/docs -name "*.md" | sort
```

跳过非内容文件（`.DS_Store`、`_category_.json` 等）。

对每个 `.md` 文件记录：
- **相对路径**（相对于 `apps/lina-site/docs/`，例如 `docs/1000-concepts/1000-ai-native.md`）
- **期望的 i18n 路径**：`apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current/<relative-path>`
- **slug**：从 frontmatter 的 `slug:` 字段读取

---

### 2. 扫描已有英文翻译

列出 i18n/en 目录中已存在的所有 `.md` 文件：

```bash
find apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current -name "*.md" | sort
```

对每个文件读取其 frontmatter 中的 **slug**，构建索引：`slug → i18n/en 文件路径`。

---

### 3. 对每篇中文文档进行分类

将每篇中文主稿归入以下四个状态之一：

| 状态 | 判断条件 | 处理方式 |
|------|----------|----------|
| **已同步** | 镜像路径存在且内容与中文版基本一致 | ✓ 无需操作 |
| **内容不同步** | 镜像路径存在，但内容与中文版差异较大 | 按翻译规范重新翻译并覆盖 |
| **路径不匹配** | 镜像路径不存在，但 slug 在其他 i18n/en 文件中匹配到 | 提交用户决策——可能是有意重命名 |
| **缺失** | 镜像路径不存在，且 slug 在任何 i18n/en 文件中均未找到 | 需要创建英文翻译 |

#### 内容差异判断标准

当文件存在于镜像路径时，对比中英文内容：

1. **读取中文源文件**的正文内容（去除 frontmatter）
2. **读取对应英文文件**的正文内容
3. 判断为**差异较大**（需要重新翻译）的条件：
   - 英文文件的章节数量与中文明显不符（差异超过 1 个一级标题）
   - 英文文件缺少中文中存在的主要内容段落
   - 英文文件中存在大量未翻译的中文字符（代码块外）
   - 英文文件内容篇幅不足中文的 60%（排除代码块）
4. 判断为**基本一致**（无需操作）的条件：
   - 章节结构对应
   - 核心内容已翻译
   - 少量措辞差异属正常范围

---

### 4. 报告差异分析

在执行任何修改前，先向用户展示结构化摘要：

```
## i18n 同步报告

### 已同步（无需操作）
- docs/quick/1000-overview.md ✓
- docs/community/1000-community.md ✓
...

### 内容不同步（需要重新翻译）
- docs/docs/1000-concepts/1000-ai-native.md
  原因：英文版缺少"规范驱动开发"章节，且篇幅仅为中文版的 45%
  操作：按翻译规范重新翻译并覆盖

### 路径不匹配（slug 匹配到其他路径）
- docs/docs/1000-concepts/1000-ai-native.md
  Slug: /docs/ai-native
  发现于：i18n/en/.../docs/core-concepts/built-in-capabilities.md
  → 请确认：是否有意为之？（y = 跳过，n = 在镜像路径重新创建）

### 缺失（需新建英文翻译）
- docs/docs/2000-architecture/1000-core-host.md  [slug: /docs/architecture/core-host]
- docs/docs/3000-plugin-development/2000-wasm-plugins.md  [slug: /docs/plugin-development/wasm]
...

总计：X 已同步，Y 内容不同步，Z 路径不匹配，W 缺失
```

**在执行前向用户确认：**
- 内容不同步：确认是否重新翻译并覆盖
- 路径不匹配：确认是跳过还是在镜像路径新建
- 缺失：确认是否继续翻译

---

### 5. 翻译或重新翻译文件

对每个确认需要处理的文件（包括**缺失**和**内容不同步**两类）：

**a. 声明开始：** `## 翻译中：<source-path>`

**b. 完整阅读中文源文件** — 通读全文，理解：
- 文档主题及其在 LinaPro 体系中的定位
- 技术术语及其准确英文对应词
- 文档语气（概念介绍、操作指引、参考手册等）
- 结构元素：标题层级、代码块、提示框、Mermaid 图表

**c. 以地道英文撰写** — 把自己当作原创技术文档作者，而非翻译者：

| 翻译原则 | 说明 |
|----------|------|
| 先理解再动笔 | 通读全文后再写第一个英文单词 |
| 自然表达 | 使用英语技术写作中惯用的句式结构 |
| 保持技术精度 | LinaPro 专属术语、API 名称、配置键名保持原样 |
| 匹配语气 | 概念文档保持概念性，操作指引保持步骤性 |
| 标题与结构 | 保持相同层级结构，标题文字自然翻译 |
| 代码块 | 绝不翻译代码、代码注释或 CLI 命令 |
| Mermaid 图表 | 翻译节点标签和边标签，保留图表类型和语法 |
| Frontmatter | 翻译 `title`、`description`、`keywords`；`slug` 和 `sidebar_position` 保持不变 |

**d. Frontmatter 处理示例：**

```yaml
# 中文源文件
---
slug: '/docs/architecture/core-host'
title: '核心主机'
hide_title: true
description: '本文介绍...'
keywords:
  - 核心主机
  - 插件
---

# 英文输出
---
slug: '/docs/architecture/core-host'      # ← 不变
title: 'Core Host'                         # ← 自然翻译
hide_title: true                           # ← 原样保留
description: 'This page covers...'        # ← 意译（地道英文）
keywords:                                  # ← 翻译为英文术语
  - core host
  - plugin
---
```

**e. 写入文件：**
- 确保父目录存在：`mkdir -p <parent-dir>`
- 将翻译后内容写入镜像路径（缺失时新建，内容不同步时覆盖）

**f. 完成声明：** `✓ 完成：<i18n-path>`

---

### 6. 验证输出

所有翻译完成后：

```bash
# 确认期望的文件均已存在
find apps/lina-site/i18n/en/docusaurus-plugin-content-docs/current -name "*.md" | sort
```

重新执行差异分析，确认缺失和不同步文件均已归零。

输出最终摘要：

```
## 同步完成

- 新建翻译：N 篇
- 重新翻译（内容不同步）：M 篇
- 跳过（已同步）：K 篇
- 用户决策（路径不匹配）：P 篇

### 本次处理的文件
- apps/lina-site/i18n/en/.../docs/2000-architecture/1000-core-host.md（新建）
- apps/lina-site/i18n/en/.../docs/1000-concepts/1000-ai-native.md（重新翻译）
...
```

---

## 翻译质量检查清单

写入每个译文文件前，在脑内逐项确认：

- [ ] 已完整阅读并理解中文文档
- [ ] 英文读起来对母语者自然流畅
- [ ] 技术术语与现有英文文档保持一致
- [ ] 输出中不含中文字符（代码块内的中文除外）
- [ ] `slug` 与中文源文件完全一致
- [ ] 所有代码块内容未被修改
- [ ] 标题层级与中文版保持一致
- [ ] Mermaid 图表内容已适当翻译

---

## 执行守则

- **严禁修改 `slug` 字段** — 这是规范 URL，必须与中文源文件保持一致
- **通读全文再翻译** — 不允许逐段翻译，必须在理解全文后再动笔
- **修改前先确认** — 展示差异报告并获得用户确认后再执行任何写入
- **路径不匹配须人工决策** — 不得静默跳过或覆盖，必须提交用户决策
- **禁止机械直译** — 若某句话用英文说起来不自然，重写它，只要意思相同
- **保护代码完整性** — 不得修改代码块、行内代码或 CLI 示例中的任何内容
- **参考现有英文文档风格** — 翻译前先阅读 2～3 篇已有的 i18n/en 文档，匹配用词和语气
- **逐文件处理** — 声明 → 翻译 → 写入 → 确认，完成后再处理下一篇
- **内容不同步时覆盖整文** — 不做局部 patch，重新翻译整篇以确保一致性
