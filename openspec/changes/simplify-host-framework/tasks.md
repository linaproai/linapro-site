## 1. 纲领索引

- [x] 1.1 确认三个子变更目录存在：`workbench-host-contract`、`startup-di-backends`、`plugin-hostservice-catalog`
- [x] 1.2 确认子变更 proposal 的范围与本纲领 D3 排除项一致
- [x] 1.3 本变更不得修改`apps/lina-core`或`apps/lina-vben`生产代码
- [x] 1.4 运行`openspec validate simplify-host-framework --strict`
- [x] 1.5 确认第二波目录存在：`simplify-host-runtime`，且其 proposal 不把第一波排除的高风险包合并写进实现任务

### 任务记录

- **DI 来源检查**：无新增运行期依赖。
- **i18n**：无影响。
- **缓存一致性**：无影响。
- **数据权限**：无影响。
- **开发工具跨平台**：无影响。
- **测试策略**：仅 OpenSpec 严格校验。
