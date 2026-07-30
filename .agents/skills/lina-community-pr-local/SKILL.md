---
name: lina-community-pr-local
description: >-
  从本地工作区分析相对 LinaPro 框架主仓库 linaproai/linapro 的变更，推送并创建主仓 PR。
  必须用户手动触发，禁止自动触发该技能。
---

# Lina Community PR Local

从当前本地仓库分析哪些变更值得同步到`LinaPro`框架主仓库，并以`PR`形式提交到`linaproai/linapro`。本技能只处理本地已有变更的筛选、组装、提交、推送和建`PR`，不主动实现新需求、不修复代码、不处理`apps/lina-plugins`子模块内部`PR`。

## 核心规则

1. **只手动触发**：只有用户明确调用`lina-community-pr-local`、`$lina-community-pr-local`或同义的“用本地变更给主仓开`PR`”请求时才能使用；禁止因看到本地有框架变更而自动调用。
2. 目标主仓库固定为`https://github.com/linaproai/linapro`，目标分支固定为`main`。
3. 执行前读取仓库顶层`AGENTS.md`，并按实际候选路径读取命中的`.agents/rules/*.md`。本技能不豁免文档、前端、后端、接口、数据库、测试、`i18n`或插件规则。
4. 严格保护用户本地改动：禁止`git reset --hard`、`git clean -fd`、静默`stash`、强推、覆盖未选择路径或丢弃任何工作区内容。
5. 不使用`git add -A`。所有暂存、提交和补丁组装都必须使用用户确认后的明确路径清单。
6. 创建`PR`前必须完成：分析候选路径 → 排除禁入路径 → 给出其它变更建议与选项 → 用户选择 → 组装干净`PR`分支 → 验证 → 提交 → `rebase`到最新`main` → 推送 → 创建`PR`。
7. 若存在冲突、远端权限不明、候选范围不清、测试门禁失败或用户未选择其它变更处理方式，停止并报告，不继续推送或创建`PR`。

## 固定排除范围

以下内容不得进入本技能创建的主仓`PR`：

| 范围 | 处理 |
| --- | --- |
| `.github/**` | 永远排除，不暂存、不复制到`PR`工作树。 |
| `go.work` | 永远排除。 |
| `go.work.sum` | 永远排除。 |
| `openspec/**`中与本地插件相关的内容 | 默认排除。凡内容指向`apps/lina-plugins/<plugin-id>`、插件私有实现、插件本地迭代或插件专属规范，均视为本地插件相关。 |

`apps/lina-plugins`子模块内部改动不属于本技能的提交对象。若主仓只显示子模块指针变化，将其作为“其它变更”单独分析；只有该指针指向已合并到子模块远端`main`的提交，且用户明确选择纳入时，才允许放入主仓`PR`。

## 优先范围

优先分析并建议纳入：

- `apps/lina-core/**`
- `apps/lina-vben/**`

这些路径有变更时，默认把它们放入“建议纳入”集合，但仍要检查是否包含调试残留、临时文件、生成错误或与本次框架同步无关的修改。若`apps/lina-core/pkg/plugin`下插件能力发生变化，必须审查`apps/lina-core/pkg/plugin/README.md`是否需要同步更新。

## 其它变更的建议口径

对优先范围外的所有本地变更，必须逐项给出是否值得提交`PR`的建议，并通过选项让用户选择。不要替用户直接纳入。

常见分类：

| 分类 | 建议 |
| --- | --- |
| 支撑`apps/lina-core`或`apps/lina-vben`变更的`api/`、`manifest/`、生成文件、语言包、配置或测试 | 通常建议纳入，并说明依赖关系。 |
| `apps/lina-site`官网文档或示例 | 仅在它解释本次框架能力变化、避免文档失真时建议纳入。 |
| `hack/`、`Makefile`、构建脚本、`CI`相关工具 | 只有与本次框架变更直接相关且通过规则门禁时建议纳入。 |
| `openspec/`非插件内容 | 作为可选项说明；若只是本地治理记录而非主仓框架事实，通常不建议纳入。 |
| `apps/lina-plugins/**`内部文件 | 不纳入本技能主仓`PR`；提示应改走子模块流程。 |
| 临时文件、日志、本地环境文件、排除范围文件 | 不建议纳入，排除。 |

向用户给出选择时使用类似结构：

```markdown
## 变更分析与选择

建议默认纳入：
- `apps/lina-core/...`：原因
- `apps/lina-vben/...`：原因

建议可选纳入：
- `path/...`：原因、收益、风险

建议排除：
- `path/...`：原因

请选择：
1. 只提交优先目录与必要支撑文件（推荐）
2. 提交优先目录、必要支撑文件和上述“可选纳入”项
3. 自定义路径清单
4. 暂停，不创建`PR`
```

若除优先目录和必要支撑文件外没有其它变更，说明“无其它候选变更”，可继续创建`PR`，无需虚构选项。

## 阶段一：前置检查

确认仓库、远端、认证和本地状态：

```bash
git rev-parse --show-toplevel
git status --short --branch
git branch --show-current
git remote -v
gh auth status
```

解析指向`linaproai/linapro`的本地远端：

1. 优先使用已有远端中`fetch`或`push`地址匹配`github.com:linaproai/linapro.git`、`https://github.com/linaproai/linapro.git`或等价地址的远端。
2. 若没有匹配远端，停止并询问用户要使用哪个远端或是否添加远端；不要自动添加。
3. 若需要从 fork 推送分支，确认 fork 远端和`gh pr create --head <owner>:<branch>`所需的`owner`。

读取最新基线：

```bash
target_remote=<remote-name>
git fetch "$target_remote" main
target_base="$target_remote/main"
```

## 阶段二：收集本地差异

同时收集已提交差异、暂存差异、未暂存差异和未跟踪文件：

```bash
git diff --stat "$target_base"...HEAD
git diff --name-status "$target_base"...HEAD
git diff --cached --stat
git diff --cached --name-status
git diff --stat
git diff --name-status
git ls-files --others --exclude-standard
git submodule status apps/lina-plugins
git -C apps/lina-plugins status --short --branch
```

分析时区分三类来源：

- 已提交但尚未进入`main`的分支差异：`git diff "$target_base"...HEAD`
- 暂存区差异：`git diff --cached`
- 工作区未暂存差异与未跟踪文件：`git diff`和`git ls-files --others --exclude-standard`

不要只看`git status --short`，因为它无法完整说明已提交但未推送或未合入`main`的变更。

## 阶段三：路径分类与用户选择

建立候选路径表，至少记录：

- 路径
- 来源：已提交、暂存、未暂存、未跟踪、子模块指针
- 分类：优先纳入、必要支撑、可选纳入、建议排除、固定排除
- 原因
- 需要读取的规则文件
- 建议验证命令

分类完成后先输出分析和选项，等待用户选择。只有用户选择后才能进入组装、提交、推送和建`PR`步骤。

## 阶段四：隔离组装`PR`分支

默认使用临时`git worktree`从最新`main`组装干净`PR`分支，避免把未选择或固定排除的本地改动带入提交。

选择分支名：

- 用户给出分支名时使用用户分支名。
- 用户未给出时，根据变更主题生成`chore/<slug>`、`fix/<slug>`或`feature/<slug>`。
- 创建前检查本地和远端是否已存在该分支；存在则换名或询问用户，不得重置已有分支。

```bash
branch_name=<branch-name>
git show-ref --verify --quiet "refs/heads/$branch_name"
git ls-remote --exit-code --heads "$target_remote" "$branch_name"
```

创建隔离工作树：

```bash
pr_worktree=$(mktemp -d "${TMPDIR:-/tmp}/linapro-pr-local.XXXXXX")
git worktree add -b "$branch_name" "$pr_worktree" "$target_base"
```

将用户已确认路径的差异按来源复制到隔离工作树：

```bash
selected_paths=(<user-confirmed-paths>)
git diff --binary "$target_base"...HEAD -- "${selected_paths[@]}" > "$pr_worktree/committed.patch"
git diff --cached --binary -- "${selected_paths[@]}" > "$pr_worktree/staged.patch"
git diff --binary -- "${selected_paths[@]}" > "$pr_worktree/unstaged.patch"

test ! -s "$pr_worktree/committed.patch" || git -C "$pr_worktree" apply --index "$pr_worktree/committed.patch"
test ! -s "$pr_worktree/staged.patch" || git -C "$pr_worktree" apply --index "$pr_worktree/staged.patch"
test ! -s "$pr_worktree/unstaged.patch" || git -C "$pr_worktree" apply --index "$pr_worktree/unstaged.patch"
```

未跟踪文件不会出现在`git diff`中。对用户选择纳入的未跟踪文件，复制到`pr_worktree`后再显式`git add -- <path>`；复制前再次确认路径不在固定排除范围。

若补丁应用失败，停止并报告冲突文件与失败命令。不要在隔离工作树里自行改代码“修补”补丁，除非用户明确要求。

## 阶段五：验证`PR`内容

在隔离工作树中检查最终`PR`差异：

```bash
git -C "$pr_worktree" status --short
git -C "$pr_worktree" diff --stat "$target_base"
git -C "$pr_worktree" diff --name-status "$target_base"
git -C "$pr_worktree" diff --name-only "$target_base" | rg '^(\.github/|go\.work$|go\.work\.sum$)'
```

上述`rg`命令返回`1`且无输出时表示未命中固定排除路径，属于通过状态。

还要人工检查`openspec/`候选是否属于本地插件相关内容；若属于，必须移出`PR`内容后重新验证。

按命中规则运行验证。若无法运行完整验证，运行最小可信验证并在`PR`正文和最终输出中说明。常见口径：

- 涉及`apps/lina-core`后端：按`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/database.md`等命中规则执行对应编译、生成或测试门禁。
- 涉及`apps/lina-vben`前端：按`.agents/rules/frontend-ui.md`、`.agents/rules/i18n.md`、`.agents/rules/testing.md`执行前端检查和用户可观察行为验证。
- 涉及文档：按`.agents/rules/documentation.md`和`.agents/instructions/markdown-format.instructions.md`做格式、链接或事实审查。

若确认无`i18n`影响或无文档治理影响，在执行记录或最终输出中明确写出该判断。

## 阶段六：提交、rebase、推送和创建`PR`

提交前再次确保仅包含用户确认路径：

```bash
git -C "$pr_worktree" status --short
git -C "$pr_worktree" diff --name-status --cached
git -C "$pr_worktree" diff --name-status
```

提交：

```bash
git -C "$pr_worktree" add -- <user-confirmed-paths>
git -C "$pr_worktree" commit -m "<type>(<scope>): <subject>"
```

提交消息从`git log --oneline -20`归纳仓库风格。若已有分支差异由多个语义组成，优先拆分为多个提交；但不要为了拆分而改动用户未选择文件。

`PR`前`rebase`门禁：

```bash
git -C "$pr_worktree" fetch "$target_remote" main
git -C "$pr_worktree" rebase "$target_remote/main"
git -C "$pr_worktree" merge-base --is-ancestor "$target_remote/main" HEAD
git -C "$pr_worktree" log --oneline --decorate "$target_remote/main"..HEAD
```

`rebase`冲突时停止，保留隔离工作树路径并报告；不要自动`rebase --continue`、`--skip`或`--abort`。

推送：

```bash
git -C "$pr_worktree" push -u "$target_remote" "$branch_name"
```

创建主仓`PR`：

```bash
gh pr create \
  --repo linaproai/linapro \
  --base main \
  --head "$branch_name" \
  --title "<title>" \
  --body-file -
```

若使用 fork 远端推送，`--head`使用`<fork-owner>:<branch_name>`。

`PR`正文包含：

- `Summary`：说明纳入的主要目录和原因。
- `Excluded`：列出固定排除和用户选择排除的路径摘要。
- `Tests`：列出已运行验证；未运行的说明原因。
- `Risk`：说明影响范围，特别是`apps/lina-core`、`apps/lina-vben`、接口、数据库、权限、`i18n`或用户可观察行为。
- `Related Issue`：识别到关联`Issue`时使用关闭关键词；没有则写明无。

## 收尾

`PR`创建成功后：

1. 输出`PR`地址、分支名、提交`SHA`、纳入路径摘要、排除路径摘要、验证结果和仍留在原始工作区的本地改动。
2. 如果隔离工作树已经干净且分支已成功推送，可以移除本技能创建的临时工作树：

```bash
git worktree remove "$pr_worktree"
```

3. 不删除原始工作区中的任何文件或改动。
4. 若`PR`创建失败，保留隔离工作树并报告路径，方便用户继续排查。

## 最终输出格式

成功时：

```markdown
## 主仓 PR 已创建

- PR：<url>
- 分支：`<branch_name>`
- 提交：`<sha>`
- 纳入范围：`apps/lina-core`、`apps/lina-vben`、...
- 排除范围：`.github`、`go.work`、`go.work.sum`、本地插件相关`openspec`、...
- 其它变更选择：用户选择了选项`<n>`
- 验证：...
- 原始工作区：未选择改动仍保留，未被提交
```

阻断时：

```markdown
## 主仓 PR 未创建

- 阻断原因：...
- 已完成：...
- 未执行：commit / push / PR
- 需要用户决定：...
- 临时工作树：`<path>`（如已创建）
```
