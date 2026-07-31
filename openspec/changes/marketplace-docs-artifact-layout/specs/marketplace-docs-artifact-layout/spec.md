## ADDED Requirements

### Requirement: Git 文档快照按插件版本与 locale 保留原始相对路径

系统 SHALL 将 Git 源版本文档正文存储于 ArtifactStore，路径布局为：

`<plugin-id>/<version>/docs/<locale>/<docPath>`

其中 `<docPath>` MUST 为规范化后的文档相对路径（保留原始文件名与子目录），MUST NOT 使用内容 hash 作为文件名。系统 MUST 将版本级文档清单写入：

`<plugin-id>/<version>/meta/docs-manifest.json`

清单 MUST 至少记录每篇文档的 `locale`、`docPath`、`sourceKind`、`contentKey` 与 `contentHash`。

#### Scenario: 同步后磁盘路径可读

- **WHEN** 系统为插件 `linapro-ai-core` 版本 `v0.1.0` 同步 locale `zh-CN` 下文档 `index.md`
- **THEN** 正文 MUST 存储在 ArtifactStore key `linapro-ai-core/v0.1.0/docs/zh-CN/index.md`
- **AND** 清单 MUST 位于 `linapro-ai-core/v0.1.0/meta/docs-manifest.json`

#### Scenario: 禁止 hash 文件名正文

- **WHEN** 系统写入任意 Git 文档正文快照
- **THEN** content key MUST NOT 采用 `docs-snapshot/.../content/<hash>.md` 形态

### Requirement: 同步必须按 content hash 判定是否更新

系统在 Git 文档同步写入时 MUST 以文档正文 content hash 判定是否覆盖本地文件。仅当本地已存在相同存储 key 时，系统 MUST 比较本地内容 hash（或清单中记录的 hash）与远端正文 hash：hash 相同 MAY 跳过写盘；hash 不同 MUST 覆盖写入。系统 MUST NOT 仅因本地已存在同名文件而跳过更新。

#### Scenario: 同名文件内容变化时覆盖

- **WHEN** 本地已存在 key `demo/v0.1.0/docs/en-US/index.md` 且 content hash 为 H1
- **AND** 同步得到同路径正文 content hash 为 H2 且 H1 ≠ H2
- **THEN** 系统 MUST 用新正文覆盖该 key

#### Scenario: 同名文件内容未变时可不写盘

- **WHEN** 本地已存在相同 key 且本地 content hash 与远端正文 hash 相同
- **THEN** 系统 MAY 跳过对该 key 的 Put
- **AND** 读路径随后仍 MUST 能读到该正文

### Requirement: 同步必须清理远端已删除的文档文件

系统在完成一次版本文档同步后，MUST 删除该版本文档树中不再属于远端集合的本地文档对象，使本地集合与远端文档集合一致。

#### Scenario: 远端删除文档后本地被清理

- **WHEN** 上一轮同步本地存在 `demo/v0.1.0/docs/zh-CN/old.md`
- **AND** 本轮同步远端集合不再包含该文档
- **THEN** 系统 MUST 删除对应本地对象
- **AND** 新清单 MUST NOT 再引用该 content key

### Requirement: ArtifactStore 支持幂等删除

本地 ArtifactStore MUST 提供按 storage key 删除对象的能力。删除不存在的 key MUST 视为成功（幂等）。

#### Scenario: 删除已存在对象

- **WHEN** 调用方对已存在的 key 执行 Delete
- **THEN** 后续 Open 该 key MUST 失败为未找到

#### Scenario: 删除不存在对象

- **WHEN** 调用方对不存在的 key 执行 Delete
- **THEN** 操作 MUST 成功且不返回存储失败
