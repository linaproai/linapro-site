## Why

插件市场「我的插件」详情的风险 Tab 目前只展示严重级别、类型与一句摘要。发布者无法区分「必须修复」与「需知情披露」，也拿不到文件路径、宿主服务明细和可执行的修复建议，导致扫描结果无法闭环为可操作清单。

## What Changes

- 为每条扫描 finding 增加**处置分类**（需修复 / 需说明 / 仅提示）与是否**阻塞提交**语义。
- 扩展风险 `payload`：写入稳定 `code` 之外的**证据字段**（SQL 文件路径、host service 摘要、路由 Top N、期望清单路径等），并在 API 中返回。
- 为每个已知 finding code 提供结构化多语言文案：标题、原因/影响、建议操作、验收标准；「我的插件」侧重修复指引，目录/审核侧重影响说明。
- 风险 Tab UI 支持按处置分组/筛选、默认优先展示需修复项、可展开详情与证据列表。
- 提交审核前对 blocking finding 给出明确拦截或确认引导（与现有 submit-review 流程衔接）。
- 补充单元测试与 E2E，覆盖处置排序、结构化文案与证据展示。

## Capabilities

### New Capabilities

- `marketplace-risk-guidance`：市场风险 finding 的处置语义、结构化指引、证据 payload 与工作台展示/提交门禁。

### Modified Capabilities

- （无独立基线能力需改名；市场风险展示能力以本变更新增规范承载。归档时若并入 `plugin-marketplace` 基线，再同步英文基线。）

## Impact

- 代码：`apps/lina-plugins/linapro-plugin-marketplace/` 后端扫描诊断、风险入库 payload、风险列表投影；前端 `detail`/`review` 风险列表与 `utils/risk.ts`；`manifest/i18n` zh-CN/en-US。
- API：`MarketplaceRiskItem.payload` 结构扩展（向后兼容：旧行仅有 `code` 时前端降级）；可选新增处置/blocking 投影字段（由 payload 或服务端派生，不改表结构）。
- 测试：后端诊断/payload 单测；前端 risk helper 单测；E2E 覆盖「我的插件」风险 Tab 展开与处置标签。
- `i18n`：有影响，新增 `riskFinding.<code>.*` 与处置/筛选文案。
- 数据权限 / 缓存：无影响（仅发布者/可见范围内既有风险读路径）。
- 开发工具跨平台：无影响。
