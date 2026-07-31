## Context

市场包扫描已产出稳定 finding code（如 `framework_dependency_missing`、`source_sql_present`、`dynamic_host_services_present`），并经 `replaceReleaseRisks` 写入 `plugin_marketplace_risk`。UI 仅渲染 severity、type、source 与一句话 summary。

现状约束：

- `PackageDiagnostic` 仅有 Code/Severity/Source/Message；payload 仅 `{"code":"..."}`。
- finding 语义混杂：缺依赖应修复；存在 SQL/hostServices 多为披露；文档索引为提示。
- 「我的插件」、目录详情、审核页共用同一套 risk list 渲染 helper。

## Goals / Non-Goals

**Goals:**

1. 每条 finding 可判定处置分类（`need_fix` / `need_attention` / `info_only`）与 `blocking`。
2. 扫描入库 payload 携带有界证据（路径/服务/路由），API 原样返回。
3. 前端按处置展示可执行指引（原因、建议、验收）与证据列表。
4. 「我的插件」默认优先展示需修复项；提交审核时对 blocking 给出明确反馈。
5. 保持无表结构迁移：处置与 blocking 由 code 策略表派生，证据写入既有 JSON payload。

**Non-Goals:**

- 不引入独立「修复工作流」状态机或人工确认落库。
- 不把存在 SQL/hostServices 默认升级为 hard fail（除非策略显式 blocking）。
- 不做跨版本 finding diff UI（可后续迭代）。
- 不修改 `lina-core` 宿主插件安装/信任边界。

## Decisions

### D1：处置与 blocking 由稳定 code 策略表派生

- **选择**：后端与前端共享同一 code→disposition/blocking 映射（后端在投影时写入 payload 派生字段或独立响应字段；前端也可本地兜底）。
- **字段**：
  - `disposition`: `need_fix` | `need_attention` | `info_only`
  - `blocking`: bool（当前默认仅对明确错误类 code 为 true；首批将 `i18n_files_missing`、`dynamic_manifest_resources_missing` 标为 need_fix 且 blocking=true）
- **备选**：独立 DB 列 → 拒绝，避免迁移且策略可热改 i18n/代码表。
- **披露类**（SQL、hostServices、routes、mock SQL、**framework 兼容版本未声明**）：`need_attention`、`blocking=false`。框架兼容版本配置属于披露/配置信息，不阻塞提交审核。
- **提示类**（docs indexed、runtime detected）：`info_only`、`blocking=false`。
- **读路径权威性**：列表投影时 `disposition`/`blocking` MUST 始终按 code 策略表派生，不得被历史 payload 中的旧值锁死（策略表变更后已入库风险行立即生效）。

### D2：扩展 `PackageDiagnostic.Payload map[string]any` 并统一入库

- 扫描函数构建诊断时附带 `Evidence` 结构，序列化进 payload：

```json
{
  "code": "source_sql_present",
  "disposition": "need_attention",
  "blocking": false,
  "files": ["manifest/sql/001-xxx.sql"],
  "services": [{"service":"data","methods":["Query"],"tables":["t_x"]}],
  "routes": [{"method":"GET","path":"/api/...","permission":"..."}],
  "expectedPath": "plugin.yaml",
  "expectedField": "dependencies.framework.version",
  "example": ">=1.0.0 <2.0.0"
}
```

- 列表上限：files/routes/services 各最多 20 条，超出写 `truncated: true` 与 `totalCount`。
- `replaceReleaseRisks` 写入完整 payload JSON；旧数据仅有 code 时投影侧补 disposition/blocking。

### D3：API 响应显式投影处置字段

- 在 `MarketplaceRiskItem` 增加可选字段：`disposition`、`blocking`（从 payload 解析或服务端填充），便于前端与 OpenAPI 文档，避免前端硬解析未知 JSON。
- **不 BREAKING**：旧客户端忽略新字段仍可读 summary。

### D4：i18n 结构化键

每个 code（**标题必须带 `.title` 后缀**，避免与子键在嵌套消息树中冲突）：

- `detail.riskFinding.<code>.title`：标题
- `detail.riskFinding.<code>.reason`
- `detail.riskFinding.<code>.remediation`
- `detail.riskFinding.<code>.acceptance`

说明：运行时 i18n 会把 flat key 嵌套为对象树。若同时存在叶子键 `riskFinding.<code>` 与子键 `riskFinding.<code>.reason`，嵌套后叶子标题会被 map 覆盖，前端 `$t(titleKey)` 失败并回退到英文 `summary`。

处置标签：

- `detail.riskDisposition.need_fix|need_attention|info_only`
- 筛选与引导文案：`detail.riskGuide.*`

角色差异：publisher 与 consumer 共用同一 remediation 文案即可（首期）；后续可拆 `.publisher`/`.consumer` 前缀而不改 code。

### D5：前端风险 Tab

- 列表排序：blocking 优先 → disposition 序（need_fix → need_attention → info_only）→ severity 序。
- 筛选 chips：全部 / 需修复 / 需说明 / 仅提示。
- 每条可展开：原因、建议、验收、证据（文件列表、服务表、路由表、期望字段示例）。
- 顶部摘要条：统计三类数量 + blocking 警告。
- 审核页与详情页复用同一 risk 展示组件/helper。

### D6：提交审核门禁

- 在前端 submit-review 前检查当前版本 risks：若存在 `blocking===true`，拦截并提示跳转风险 Tab。
- 后端 submit-review：对 blocking findings 返回业务错误（与前端一致），避免绕过 UI。
- 非 blocking 的 need_attention：不拦截，可在提交确认文案中提示「仍有需说明项」。

## Risks / Trade-offs

- **[Risk] 策略过严导致合法包无法提交** → 首批 blocking 仅限明确缺失类；SQL/hostServices 不 blocking。
- **[Risk] payload 过大** → 硬上限 20 + truncated 标记。
- **[Risk] 旧风险行无证据** → 重新同步/上传后刷新；展示层证据缺失时仅显示指引文案。
- **[Trade-off] 前后端各有一份 disposition 映射** → 以后端投影为准，前端 mapping 作降级与单测；保持 code 列表同步注释。

## Migration Plan

1. 部署新代码后，新扫描/同步写入 enriched payload。
2. 已有 release 风险行在下次 replace 时更新；只读历史草稿可手动重新同步。
3. 回滚：前端忽略新字段即可；旧 payload 仍可显示标题摘要。

## Open Questions

- 无。首批 blocking 集合与证据上限按本设计固定；若产品后续要调整，仅改策略表。
