## 1. 后端诊断策略与 payload

- [x] 1.1 扩展 `PackageDiagnostic` 支持证据字段；实现 code→disposition/blocking 策略表与 payload 序列化
- [x] 1.2 源码包扫描写入 SQL/依赖/i18n/docs 证据并生成完整 payload
- [x] 1.3 动态包扫描写入 hostServices/routes/SQL/manifest 证据并生成完整 payload
- [x] 1.4 `replaceReleaseRisks` 与 `riskItemFromEntity` 投影 `disposition`/`blocking`；更新 `MarketplaceRiskItem` API 字段
- [x] 1.5 提交审核时检查 blocking findings 并拒绝

## 2. 前端风险展示

- [x] 2.1 扩展 `utils/risk.ts`：处置排序、结构化文案解析、证据读取
- [x] 2.2 详情页风险 Tab：摘要条、筛选、展开指引与证据列表
- [x] 2.3 审核页风险列表对齐同一展示与排序
- [x] 2.4 「我的插件」提交审核前拦截 blocking 项并提示

## 3. i18n

- [x] 3.1 为全部已知 finding code 补充 zh-CN/en-US 的 reason/remediation/acceptance 与处置/筛选文案

## 4. 测试与验证

- [x] 4.1 后端单测：策略表、payload 证据、blocking 提交门禁、风险排序投影
- [x] 4.2 前端 unit：risk helper 排序与文案键
- [x] 4.3 E2E：风险 Tab 处置标签、展开建议、blocking 相关展示（TC016）
- [x] 4.4 `openspec validate marketplace-risk-remediation-guidance --strict`

## Feedback

- [x] **FB-1**: 框架兼容版本配置（`framework_dependency_missing`）不应作为阻塞提交项；策略表与读路径以 code 为准
- [x] **FB-2**: 风险 Tab 标题 i18n 键与 reason/remediation/acceptance 子键嵌套冲突，导致回退显示英文 summary
